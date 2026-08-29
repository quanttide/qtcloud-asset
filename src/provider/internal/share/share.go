// Package share stores public folder-share records.
//
// The Provider supports an in-memory adapter for tests and a durable RDS
// adapter for production. The Store interface keeps the API independent from
// that persistence choice.
package share

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	maxPrefixes     = 32
	maxPrefixLength = 1024
	maxKeys         = 128
	maxKeyLength    = 1024
	maxTitleLength  = 120
)

// ErrStoreUnavailable indicates that durable share storage is not ready.
var ErrStoreUnavailable = errors.New("share store is unavailable")

// Record describes one public, read-only content share.
type Record struct {
	Token     string
	Title     string
	Bucket    string
	Prefixes  []string
	Keys      []string
	CreatedBy string
	CreatedAt time.Time
	RevokedAt *time.Time
}

// Store persists share records.
type Store interface {
	Create(record Record) (Record, error)
	Get(token string) (Record, bool, error)
	ListByOwner(ownerID string) ([]Record, error)
	Revoke(token string, revokedAt time.Time) (Record, bool, error)
}

// NormalizePrefixes validates and canonicalizes folder prefixes.
func NormalizePrefixes(prefixes []string) ([]string, error) {
	if len(prefixes) == 0 || len(prefixes) > maxPrefixes {
		return nil, errors.New("between 1 and 32 folder prefixes are required")
	}

	seen := make(map[string]struct{}, len(prefixes))
	normalized := make([]string, 0, len(prefixes))
	for _, raw := range prefixes {
		prefix := strings.TrimSpace(raw)
		if prefix == "" || len(prefix) > maxPrefixLength || !strings.HasSuffix(prefix, "/") {
			return nil, errors.New("folder prefixes must be non-empty paths ending with /")
		}
		if strings.HasPrefix(prefix, "/") || strings.ContainsRune(prefix, '\x00') {
			return nil, errors.New("folder prefixes must be relative object paths")
		}
		for _, segment := range strings.Split(strings.TrimSuffix(prefix, "/"), "/") {
			if segment == "." || segment == ".." {
				return nil, errors.New("folder prefixes cannot contain . or .. path segments")
			}
		}
		if _, exists := seen[prefix]; exists {
			continue
		}
		seen[prefix] = struct{}{}
		normalized = append(normalized, prefix)
	}
	if len(normalized) == 0 {
		return nil, errors.New("between 1 and 32 folder prefixes are required")
	}
	sort.Strings(normalized)
	return normalized, nil
}

// NormalizeKeys validates and canonicalizes explicitly shared object keys.
func NormalizeKeys(keys []string) ([]string, error) {
	if len(keys) > maxKeys {
		return nil, errors.New("at most 128 file keys are allowed")
	}

	seen := make(map[string]struct{}, len(keys))
	normalized := make([]string, 0, len(keys))
	for _, raw := range keys {
		key := strings.TrimSpace(raw)
		if key == "" || len(key) > maxKeyLength {
			return nil, errors.New("file keys must be non-empty relative paths")
		}
		if strings.HasPrefix(key, "/") || strings.ContainsRune(key, '\x00') {
			return nil, errors.New("file keys must be relative object paths")
		}
		for _, segment := range strings.Split(key, "/") {
			if segment == "." || segment == ".." {
				return nil, errors.New("file keys cannot contain . or .. path segments")
			}
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	sort.Strings(normalized)
	return normalized, nil
}

// AllowsKey reports whether an object key belongs to one of the shared folders.
func AllowsKey(prefixes []string, key string) bool {
	if key == "" {
		return false
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// AllowsObject reports whether an object is covered by a shared folder or key.
func AllowsObject(prefixes, keys []string, key string) bool {
	if AllowsKey(prefixes, key) {
		return true
	}
	for _, allowed := range keys {
		if allowed == key {
			return true
		}
	}
	return false
}

// AllowsPrefix reports whether listing a prefix can contain shared objects.
func AllowsPrefix(prefixes, keys []string, requested string) bool {
	if requested == "" {
		return true
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(prefix, requested) || strings.HasPrefix(requested, prefix) {
			return true
		}
	}
	for _, key := range keys {
		if strings.HasPrefix(key, requested) {
			return true
		}
	}
	return false
}

// MemoryStore is the in-process share store used by tests and local development.
type MemoryStore struct {
	mu      sync.RWMutex
	records map[string]Record
}

// UnavailableStore fails closed when production persistence is not configured.
type UnavailableStore struct {
	err error
}

// NewUnavailableStore creates a store that reports a durable-storage outage.
func NewUnavailableStore(err error) *UnavailableStore {
	if err == nil {
		err = ErrStoreUnavailable
	}
	return &UnavailableStore{err: fmt.Errorf("%w: %v", ErrStoreUnavailable, err)}
}

func (s *UnavailableStore) Create(Record) (Record, error) {
	return Record{}, s.storeError()
}

func (s *UnavailableStore) Get(string) (Record, bool, error) {
	return Record{}, false, s.storeError()
}

func (s *UnavailableStore) ListByOwner(string) ([]Record, error) {
	return nil, s.storeError()
}

func (s *UnavailableStore) Revoke(string, time.Time) (Record, bool, error) {
	return Record{}, false, s.storeError()
}

func (s *UnavailableStore) storeError() error {
	if s == nil || s.err == nil {
		return ErrStoreUnavailable
	}
	return s.err
}

// NewMemoryStore creates an empty in-memory share store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: make(map[string]Record)}
}

// Create saves a share and generates an opaque token when needed.
func (s *MemoryStore) Create(record Record) (Record, error) {
	if s == nil {
		return Record{}, errors.New("share store is not configured")
	}
	if record.Bucket == "" {
		return Record{}, errors.New("share bucket is required")
	}
	if len(record.Prefixes) == 0 && len(record.Keys) == 0 {
		return Record{}, errors.New("share bucket and prefixes or keys are required")
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	if record.Token == "" {
		token, err := newToken()
		if err != nil {
			return Record{}, err
		}
		record.Token = token
	}

	record.Prefixes = append([]string(nil), record.Prefixes...)
	record.Keys = append([]string(nil), record.Keys...)

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.records[record.Token]; exists {
		return Record{}, fmt.Errorf("share token already exists")
	}
	s.records[record.Token] = record
	return clone(record), nil
}

// Get returns a share snapshot.
func (s *MemoryStore) Get(token string) (Record, bool, error) {
	if s == nil {
		return Record{}, false, errors.New("share store is not configured")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[token]
	if !ok {
		return Record{}, false, nil
	}
	return clone(record), true, nil
}

// ListByOwner returns active and revoked shares created by one user.
func (s *MemoryStore) ListByOwner(ownerID string) ([]Record, error) {
	if s == nil {
		return nil, errors.New("share store is not configured")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := make([]Record, 0)
	for _, record := range s.records {
		if record.CreatedBy == ownerID {
			records = append(records, clone(record))
		}
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].CreatedAt.Equal(records[j].CreatedAt) {
			return records[i].Token < records[j].Token
		}
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records, nil
}

// Revoke marks a share inactive while retaining the record for auditability.
func (s *MemoryStore) Revoke(token string, revokedAt time.Time) (Record, bool, error) {
	if s == nil {
		return Record{}, false, errors.New("share store is not configured")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.records[token]
	if !ok {
		return Record{}, false, nil
	}
	if record.RevokedAt == nil {
		record.RevokedAt = &revokedAt
		s.records[token] = record
	}
	return clone(record), true, nil
}

func clone(record Record) Record {
	record.Prefixes = append([]string(nil), record.Prefixes...)
	record.Keys = append([]string(nil), record.Keys...)
	if record.RevokedAt != nil {
		revokedAt := *record.RevokedAt
		record.RevokedAt = &revokedAt
	}
	return record
}

func newToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate share token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
