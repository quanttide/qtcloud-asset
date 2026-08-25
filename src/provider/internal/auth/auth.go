// Package auth provides Provider-side authentication and session primitives.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

const (
	// SessionCookieName is the browser cookie that carries the opaque session ID.
	SessionCookieName = "qtcloud_asset_session"
	// LoginStateCookieName carries the SSO state nonce during the login roundtrip.
	LoginStateCookieName = "qtcloud_asset_login_state"
)

const defaultSessionTTL = 12 * time.Hour

// Role is the first-pass access role used by Plan A.
type Role string

const (
	// RoleViewer can read ordinary asset metadata.
	RoleViewer Role = "viewer"
	// RoleAdmin can manage users and access authorized private resources.
	RoleAdmin Role = "admin"
)

// UserStatus controls whether an account can hold active sessions.
type UserStatus string

const (
	// UserStatusActive can authenticate and use authorized routes.
	UserStatusActive UserStatus = "active"
	// UserStatusDisabled is blocked from future authentication.
	UserStatusDisabled UserStatus = "disabled"
)

// User is the authenticated principal exposed to API handlers.
type User struct {
	ID          string     `json:"id"`
	ExternalID  string     `json:"external_id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Role        Role       `json:"role"`
	Status      UserStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt time.Time  `json:"last_login_at"`
}

// Session stores the server-side login state for one browser session.
type Session struct {
	ID        string
	UserID    string
	User      User
	ExpiresAt time.Time
	RevokedAt *time.Time
	IP        string
	UserAgent string
	CreatedAt time.Time
}

// AuditAction identifies a security-relevant Provider action.
type AuditAction string

const (
	AuditActionLogin          AuditAction = "login"
	AuditActionLogout         AuditAction = "logout"
	AuditActionAuthFailed     AuditAction = "auth_failed"
	AuditActionListBuckets    AuditAction = "list_buckets"
	AuditActionListObjects    AuditAction = "list_objects"
	AuditActionObjectURL      AuditAction = "object_url"
	AuditActionListUsers      AuditAction = "list_users"
	AuditActionInviteUser     AuditAction = "invite_user"
	AuditActionUpdateUserRole AuditAction = "update_user_role"
	AuditActionDisableUser    AuditAction = "disable_user"
	AuditActionRevokeSessions AuditAction = "revoke_sessions"
)

// AuditResult records whether the action succeeded or was denied.
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultDenied  AuditResult = "denied"
	AuditResultFailure AuditResult = "failure"
)

// AuditLog records who did what against which Provider resource.
type AuditLog struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id,omitempty"`
	Action    AuditAction `json:"action"`
	Target    string      `json:"target"`
	Result    AuditResult `json:"result"`
	IP        string      `json:"ip,omitempty"`
	UserAgent string      `json:"user_agent,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

// UserStore persists Provider users.
type UserStore interface {
	UpsertFromIdentity(user User, now time.Time) (User, error)
	UpsertManaged(user User, now time.Time) (User, error)
	List() []User
	GetByID(id string) (User, bool)
	UpdateRole(id string, role Role) (User, bool)
	Disable(id string, disabledAt time.Time) bool
}

// AuditLogStore persists security-relevant events.
type AuditLogStore interface {
	Record(entry AuditLog) error
}

// IdentityProvider exchanges SSO callback details for an authenticated user.
type IdentityProvider interface {
	LoginURL(state string) (string, error)
	Exchange(ctx context.Context, code, state string) (User, error)
}

// ErrIdentityProviderNotConfigured is returned before platform SSO is wired.
var ErrIdentityProviderNotConfigured = errors.New("identity provider is not configured")

// NotConfiguredIdentityProvider keeps auth routes explicit until SSO is wired.
type NotConfiguredIdentityProvider struct{}

// LoginURL returns a setup error when platform SSO is not configured.
func (NotConfiguredIdentityProvider) LoginURL(string) (string, error) {
	return "", ErrIdentityProviderNotConfigured
}

// Exchange returns a setup error when platform SSO is not configured.
func (NotConfiguredIdentityProvider) Exchange(context.Context, string, string) (User, error) {
	return User{}, ErrIdentityProviderNotConfigured
}

// SessionStore persists server-side sessions.
type SessionStore interface {
	Create(session Session) error
	Get(id string) (Session, bool)
	Revoke(id string, revokedAt time.Time) bool
	RevokeUserSessions(userID string, revokedAt time.Time) int
}

// MemoryUserStore is an in-process user store used until RDS is wired.
type MemoryUserStore struct {
	mu      sync.RWMutex
	users   map[string]User
	byEmail map[string]string
}

// NewMemoryUserStore creates an empty memory user store.
func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		users:   make(map[string]User),
		byEmail: make(map[string]string),
	}
}

// UpsertFromIdentity creates or updates a user from an external identity.
func (s *MemoryUserStore) UpsertFromIdentity(user User, now time.Time) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.upsertLocked(user, now, true)
}

// UpsertManaged creates or updates a user from an admin-managed invite.
func (s *MemoryUserStore) UpsertManaged(user User, now time.Time) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.upsertLocked(user, now, false)
}

func (s *MemoryUserStore) upsertLocked(user User, now time.Time, markLogin bool) (User, error) {

	id := user.ID
	if id == "" && user.Email != "" {
		id = s.byEmail[user.Email]
	}
	if id == "" {
		generatedID, err := randomToken()
		if err != nil {
			return User{}, err
		}
		id = generatedID
	}

	existing := s.users[id]
	if existing.CreatedAt.IsZero() {
		existing.CreatedAt = now
	}
	if existing.Status == "" {
		existing.Status = UserStatusActive
	}
	if user.Status != "" {
		existing.Status = user.Status
	}
	if user.Role != "" {
		existing.Role = user.Role
	} else if existing.Role == "" {
		existing.Role = RoleViewer
	}
	existing.ID = id
	existing.ExternalID = user.ExternalID
	existing.Email = user.Email
	existing.Name = user.Name
	if markLogin {
		existing.LastLoginAt = now
	}

	s.users[id] = existing
	if existing.Email != "" {
		s.byEmail[existing.Email] = id
	}
	return existing, nil
}

// List returns a stable snapshot of all users sorted by email then ID.
func (s *MemoryUserStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool {
		if users[i].Email == users[j].Email {
			return users[i].ID < users[j].ID
		}
		return users[i].Email < users[j].Email
	})
	return users
}

// GetByID returns a user by internal ID.
func (s *MemoryUserStore) GetByID(id string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	return user, ok
}

// UpdateRole changes a user's role.
func (s *MemoryUserStore) UpdateRole(id string, role Role) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return User{}, false
	}
	user.Role = role
	s.users[id] = user
	return user, true
}

// Disable marks a user as disabled.
func (s *MemoryUserStore) Disable(id string, _ time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return false
	}
	user.Status = UserStatusDisabled
	s.users[id] = user
	return true
}

// MemorySessionStore is a small in-process session store for tests and the
// pre-RDS auth skeleton. RDS-backed storage replaces this in Day 3.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

// NewMemorySessionStore creates an empty memory session store.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{sessions: make(map[string]Session)}
}

// Create saves a session.
func (s *MemorySessionStore) Create(session Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	return nil
}

// Get returns a session by ID.
func (s *MemorySessionStore) Get(id string) (Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	return session, ok
}

// Revoke marks a session as revoked.
func (s *MemorySessionStore) Revoke(id string, revokedAt time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return false
	}
	session.RevokedAt = &revokedAt
	s.sessions[id] = session
	return true
}

// RevokeUserSessions marks every active session for a user as revoked.
func (s *MemorySessionStore) RevokeUserSessions(userID string, revokedAt time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	revoked := 0
	for id, session := range s.sessions {
		if session.UserID == userID && session.RevokedAt == nil {
			session.RevokedAt = &revokedAt
			s.sessions[id] = session
			revoked++
		}
	}
	return revoked
}

// MemoryAuditLogStore is an in-process audit log store used until RDS is wired.
type MemoryAuditLogStore struct {
	mu      sync.RWMutex
	entries []AuditLog
}

// NewMemoryAuditLogStore creates an empty memory audit log store.
func NewMemoryAuditLogStore() *MemoryAuditLogStore {
	return &MemoryAuditLogStore{}
}

// Record appends an audit log entry.
func (s *MemoryAuditLogStore) Record(entry AuditLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry.ID == "" {
		id, err := randomToken()
		if err != nil {
			return err
		}
		entry.ID = id
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	s.entries = append(s.entries, entry)
	return nil
}

// List returns a snapshot of audit log entries for tests and diagnostics.
func (s *MemoryAuditLogStore) List() []AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := make([]AuditLog, len(s.entries))
	copy(entries, s.entries)
	return entries
}

// ManagerOptions controls cookie and session behavior.
type ManagerOptions struct {
	Store        SessionStore
	SessionTTL   time.Duration
	CookieSecure bool
	SameSite     http.SameSite
}

// Manager creates and validates server-side sessions.
type Manager struct {
	store        SessionStore
	sessionTTL   time.Duration
	cookieSecure bool
	sameSite     http.SameSite
}

// NewManager creates a session manager.
func NewManager(options ManagerOptions) *Manager {
	ttl := options.SessionTTL
	if ttl == 0 {
		ttl = defaultSessionTTL
	}
	store := options.Store
	if store == nil {
		store = NewMemorySessionStore()
	}
	sameSite := options.SameSite
	if sameSite == 0 {
		sameSite = http.SameSiteLaxMode
	}
	return &Manager{
		store:        store,
		sessionTTL:   ttl,
		cookieSecure: options.CookieSecure,
		sameSite:     sameSite,
	}
}

// NewState creates a login CSRF nonce.
func (m *Manager) NewState() (string, error) {
	return randomToken()
}

// LoginStateCookie builds the short-lived state cookie for an SSO roundtrip.
func (m *Manager) LoginStateCookie(state string) *http.Cookie {
	return &http.Cookie{
		Name:     LoginStateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   m.cookieSecure,
		SameSite: m.sameSite,
	}
}

// ClearLoginStateCookie expires the temporary login state cookie.
func (m *Manager) ClearLoginStateCookie() *http.Cookie {
	return clearCookie(LoginStateCookieName, m.cookieSecure, m.sameSite)
}

// CreateSession creates a server-side session and returns its browser cookie.
func (m *Manager) CreateSession(user User, now time.Time) (*http.Cookie, error) {
	return m.CreateSessionWithMetadata(user, now, "", "")
}

// CreateSessionWithMetadata creates a session and records request metadata.
func (m *Manager) CreateSessionWithMetadata(user User, now time.Time, ip, userAgent string) (*http.Cookie, error) {
	id, err := randomToken()
	if err != nil {
		return nil, err
	}
	session := Session{
		ID:        id,
		UserID:    user.ID,
		User:      user,
		ExpiresAt: now.Add(m.sessionTTL),
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: now,
	}
	if err := m.store.Create(session); err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   int(m.sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   m.cookieSecure,
		SameSite: m.sameSite,
	}, nil
}

// Authenticate resolves the current request's session.
func (m *Manager) Authenticate(r *http.Request, now time.Time) (User, string, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return User{}, "", false
	}
	session, ok := m.store.Get(cookie.Value)
	if !ok || session.RevokedAt != nil || !now.Before(session.ExpiresAt) {
		return User{}, "", false
	}
	return session.User, session.ID, true
}

// RevokeFromRequest revokes the current request's session when present.
func (m *Manager) RevokeFromRequest(r *http.Request, now time.Time) bool {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	return m.store.Revoke(cookie.Value, now)
}

// RevokeUserSessions revokes every active session held by a user.
func (m *Manager) RevokeUserSessions(userID string, now time.Time) int {
	return m.store.RevokeUserSessions(userID, now)
}

// ClearSessionCookie expires the browser session cookie.
func (m *Manager) ClearSessionCookie() *http.Cookie {
	return clearCookie(SessionCookieName, m.cookieSecure, m.sameSite)
}

// ValidateState checks the callback state against the HttpOnly state cookie.
func ValidateState(r *http.Request, expected string) bool {
	if expected == "" {
		return false
	}
	cookie, err := r.Cookie(LoginStateCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(expected)) == 1
}

func clearCookie(name string, secure bool, sameSite http.SameSite) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	}
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate auth token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
