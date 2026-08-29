package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/share"
)

// SQLShareStore persists folder shares in the platform-shared RDS database.
// The database stores only a hash of the bearer token for lookups and an
// encrypted copy for owner-facing share management.
type SQLShareStore struct {
	db            *sql.DB
	encryptionKey []byte
	dialect       shareDatabaseDialect
}

const shareDatabaseTimeout = 5 * time.Second

type shareDatabaseDialect uint8

const (
	shareDatabaseDialectMySQL shareDatabaseDialect = iota
	shareDatabaseDialectPostgres
)

// NewSQLShareStore creates a durable MySQL-compatible share store.
func NewSQLShareStore(db *sql.DB, encryptionKey []byte) (*SQLShareStore, error) {
	return NewSQLShareStoreWithDriver(db, encryptionKey, "mysql")
}

// NewSQLShareStoreWithDriver creates a durable share store for a supported RDS driver.
func NewSQLShareStoreWithDriver(db *sql.DB, encryptionKey []byte, driver string) (*SQLShareStore, error) {
	if db == nil {
		return nil, errors.New("share database is required")
	}
	if len(encryptionKey) != 32 {
		return nil, errors.New("share token encryption key must be 32 bytes")
	}
	_, dialect, err := normalizeShareStoreDriver(driver)
	if err != nil {
		return nil, err
	}
	return &SQLShareStore{
		db:            db,
		encryptionKey: append([]byte(nil), encryptionKey...),
		dialect:       dialect,
	}, nil
}

func (s *SQLShareStore) Create(record share.Record) (share.Record, error) {
	if s == nil || s.db == nil {
		return share.Record{}, share.ErrStoreUnavailable
	}
	if record.Bucket == "" ||
		(len(record.Prefixes) == 0 && len(record.Keys) == 0) {
		return share.Record{}, errors.New("share bucket and prefixes or keys are required")
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	if record.Token == "" {
		token, err := newShareToken()
		if err != nil {
			return share.Record{}, err
		}
		record.Token = token
	}

	targets, err := encodeShareTargets(record.Prefixes, record.Keys)
	if err != nil {
		return share.Record{}, fmt.Errorf("encode share targets: %w", err)
	}
	encryptedToken, err := shareTokenCiphertext(s.encryptionKey, record.Token)
	if err != nil {
		return share.Record{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	prefixValue := any(targets)
	if s.dialect == shareDatabaseDialectPostgres {
		prefixValue = string(targets)
	}
	_, err = s.db.ExecContext(
		ctx,
		s.statement(`INSERT INTO folder_shares
			(token_hash, token_ciphertext, title, bucket, prefixes, created_by, created_at, revoked_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`),
		shareTokenHash(record.Token),
		encryptedToken,
		record.Title,
		record.Bucket,
		prefixValue,
		record.CreatedBy,
		record.CreatedAt.UTC(),
		nullableTime(record.RevokedAt),
	)
	if err != nil {
		return share.Record{}, fmt.Errorf("%w: insert share: %v", share.ErrStoreUnavailable, err)
	}
	return cloneRecord(record), nil
}

func (s *SQLShareStore) Get(token string) (share.Record, bool, error) {
	if s == nil || s.db == nil {
		return share.Record{}, false, share.ErrStoreUnavailable
	}
	if token == "" {
		return share.Record{}, false, nil
	}

	var (
		encryptedToken []byte
		record         share.Record
		prefixes       []byte
		revokedAt      sql.NullTime
	)
	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	err := s.db.QueryRowContext(
		ctx,
		s.statement(`SELECT token_ciphertext, title, bucket, prefixes, created_by, created_at, revoked_at
		 FROM folder_shares
		 WHERE token_hash = ?`),
		shareTokenHash(token),
	).Scan(
		&encryptedToken,
		&record.Title,
		&record.Bucket,
		&prefixes,
		&record.CreatedBy,
		&record.CreatedAt,
		&revokedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return share.Record{}, false, nil
	}
	if err != nil {
		return share.Record{}, false, fmt.Errorf("%w: read share: %v", share.ErrStoreUnavailable, err)
	}

	record.Token, err = shareTokenPlaintext(s.encryptionKey, encryptedToken)
	if err != nil {
		return share.Record{}, false, fmt.Errorf("%w: decode share token: %v", share.ErrStoreUnavailable, err)
	}
	if err := decodeShareTargets(prefixes, &record.Prefixes, &record.Keys); err != nil {
		return share.Record{}, false, fmt.Errorf("%w: decode share targets: %v", share.ErrStoreUnavailable, err)
	}
	if revokedAt.Valid {
		revoked := revokedAt.Time
		record.RevokedAt = &revoked
	}
	return cloneRecord(record), true, nil
}

func (s *SQLShareStore) ListByOwner(ownerID string) ([]share.Record, error) {
	if s == nil || s.db == nil {
		return nil, share.ErrStoreUnavailable
	}

	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	rows, err := s.db.QueryContext(
		ctx,
		s.statement(`SELECT token_ciphertext, title, bucket, prefixes, created_by, created_at, revoked_at
		 FROM folder_shares
		 WHERE created_by = ?
		 ORDER BY created_at DESC, token_hash ASC`),
		ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: list shares: %v", share.ErrStoreUnavailable, err)
	}
	defer rows.Close()

	records := make([]share.Record, 0)
	for rows.Next() {
		var (
			encryptedToken []byte
			record         share.Record
			prefixes       []byte
			revokedAt      sql.NullTime
		)
		if err := rows.Scan(
			&encryptedToken,
			&record.Title,
			&record.Bucket,
			&prefixes,
			&record.CreatedBy,
			&record.CreatedAt,
			&revokedAt,
		); err != nil {
			return nil, fmt.Errorf("%w: scan share: %v", share.ErrStoreUnavailable, err)
		}
		record.Token, err = shareTokenPlaintext(s.encryptionKey, encryptedToken)
		if err != nil {
			return nil, fmt.Errorf("%w: decode share token: %v", share.ErrStoreUnavailable, err)
		}
		if err := decodeShareTargets(prefixes, &record.Prefixes, &record.Keys); err != nil {
			return nil, fmt.Errorf("%w: decode share targets: %v", share.ErrStoreUnavailable, err)
		}
		if revokedAt.Valid {
			revoked := revokedAt.Time
			record.RevokedAt = &revoked
		}
		records = append(records, cloneRecord(record))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate shares: %v", share.ErrStoreUnavailable, err)
	}
	return records, nil
}

func (s *SQLShareStore) Revoke(token string, revokedAt time.Time) (share.Record, bool, error) {
	if s == nil || s.db == nil {
		return share.Record{}, false, share.ErrStoreUnavailable
	}
	if token == "" {
		return share.Record{}, false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	_, err := s.db.ExecContext(
		ctx,
		s.statement(`UPDATE folder_shares
		 SET revoked_at = COALESCE(revoked_at, ?)
		 WHERE token_hash = ?`),
		revokedAt.UTC(),
		shareTokenHash(token),
	)
	if err != nil {
		return share.Record{}, false, fmt.Errorf("%w: revoke share: %v", share.ErrStoreUnavailable, err)
	}
	record, ok, err := s.Get(token)
	if err != nil || !ok {
		return record, ok, err
	}
	return record, true, nil
}

func (s *SQLShareStore) statement(query string) string {
	if s == nil || s.dialect != shareDatabaseDialectPostgres {
		return query
	}

	var (
		builder     strings.Builder
		placeholder int
	)
	builder.Grow(len(query) + 8)
	for _, character := range query {
		if character != '?' {
			builder.WriteRune(character)
			continue
		}
		placeholder++
		fmt.Fprintf(&builder, "$%d", placeholder)
	}
	return builder.String()
}

func normalizeShareStoreDriver(driver string) (string, shareDatabaseDialect, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "mysql":
		return "mysql", shareDatabaseDialectMySQL, nil
	case "postgres", "postgresql", "pgx":
		return "pgx", shareDatabaseDialectPostgres, nil
	default:
		return "", 0, fmt.Errorf("unsupported share database driver %q", driver)
	}
}

func newShareToken() (string, error) {
	return share.NewToken()
}

func shareTokenHash(token string) string {
	return share.TokenHash(token)
}

func shareTokenCiphertext(key []byte, token string) ([]byte, error) {
	return share.EncryptToken(key, token)
}

func shareTokenPlaintext(key, encrypted []byte) (string, error) {
	return share.DecryptToken(key, encrypted)
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}

func cloneRecord(record share.Record) share.Record {
	record.Prefixes = append([]string(nil), record.Prefixes...)
	record.Keys = append([]string(nil), record.Keys...)
	if record.RevokedAt != nil {
		revoked := *record.RevokedAt
		record.RevokedAt = &revoked
	}
	return record
}

type shareTargets struct {
	Prefixes []string `json:"prefixes"`
	Keys     []string `json:"keys"`
}

func encodeShareTargets(prefixes, keys []string) ([]byte, error) {
	if len(keys) == 0 {
		return json.Marshal(prefixes)
	}
	return json.Marshal(shareTargets{
		Prefixes: prefixes,
		Keys:     keys,
	})
}

func decodeShareTargets(raw []byte, prefixes, keys *[]string) error {
	var legacy []string
	if err := json.Unmarshal(raw, &legacy); err == nil {
		*prefixes = legacy
		*keys = nil
		return nil
	}

	var targets shareTargets
	if err := json.Unmarshal(raw, &targets); err != nil {
		return err
	}
	*prefixes = targets.Prefixes
	*keys = targets.Keys
	return nil
}

// OpenShareStore opens the configured durable share store. Migrations are
// deliberately external so the Provider never changes shared infrastructure
// during startup.
func OpenShareStore(cfg *config.Config) (share.Store, func() error, error) {
	if cfg == nil {
		err := errors.New("provider config is missing")
		return share.NewUnavailableStore(err), nil, err
	}
	if cfg.ShareStoreMode == "memory" {
		return share.NewMemoryStore(), nil, nil
	}
	if cfg.RDSConnectionString == "" {
		err := errors.New("RDS_CONNECTION_STRING is missing")
		return share.NewUnavailableStore(err), nil, err
	}
	key, err := share.ParseTokenEncryptionKey(cfg.ShareTokenEncryptionKey)
	if err != nil {
		return share.NewUnavailableStore(err), nil, err
	}

	driver, _, err := normalizeShareStoreDriver(cfg.RDSDriver)
	if err != nil {
		return share.NewUnavailableStore(err), nil, err
	}
	db, err := sql.Open(driver, cfg.RDSConnectionString)
	if err != nil {
		return share.NewUnavailableStore(err), nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return share.NewUnavailableStore(err), nil, err
	}
	store, err := NewSQLShareStoreWithDriver(db, key, driver)
	if err != nil {
		_ = db.Close()
		return share.NewUnavailableStore(err), nil, err
	}
	return store, db.Close, nil
}
