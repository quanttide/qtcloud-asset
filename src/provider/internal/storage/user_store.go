package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
)

// SQLUserStore persists Provider users in the platform-shared RDS database.
// Passwords are never stored here, only the configured password hash.
type SQLUserStore struct {
	db      *sql.DB
	dialect shareDatabaseDialect
}

const userSelectColumns = `id, external_id, account, email, name, role, status, password_hash, created_at, last_login_at`

const userUpsertQuery = `INSERT INTO users
			(id, external_id, account, email, name, role, status, password_hash, created_at, last_login_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				account = VALUES(account),
				email = VALUES(email),
				name = VALUES(name),
				role = VALUES(role),
				status = VALUES(status),
				password_hash = IF(COALESCE(VALUES(password_hash), '') = '', password_hash, VALUES(password_hash)),
				last_login_at = COALESCE(VALUES(last_login_at), last_login_at)`

const userPostgresUpsertQuery = `INSERT INTO users
			(id, external_id, account, email, name, role, status, password_hash, created_at, last_login_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (external_id) DO UPDATE SET
				account = EXCLUDED.account,
				email = EXCLUDED.email,
				name = EXCLUDED.name,
				role = EXCLUDED.role,
				status = EXCLUDED.status,
				password_hash = COALESCE(NULLIF(EXCLUDED.password_hash, ''), users.password_hash),
				last_login_at = COALESCE(EXCLUDED.last_login_at, users.last_login_at)`

const userIdentityUpsertQuery = `INSERT INTO users
			(id, external_id, account, email, name, role, status, password_hash, created_at, last_login_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				account = VALUES(account),
				email = VALUES(email),
				name = VALUES(name),
				password_hash = IF(COALESCE(VALUES(password_hash), '') = '', password_hash, VALUES(password_hash)),
				last_login_at = COALESCE(VALUES(last_login_at), last_login_at)`

const userPostgresIdentityUpsertQuery = `INSERT INTO users
			(id, external_id, account, email, name, role, status, password_hash, created_at, last_login_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (external_id) DO UPDATE SET
				account = EXCLUDED.account,
				email = EXCLUDED.email,
				name = EXCLUDED.name,
				password_hash = COALESCE(NULLIF(EXCLUDED.password_hash, ''), users.password_hash),
				last_login_at = COALESCE(EXCLUDED.last_login_at, users.last_login_at)`

// NewSQLUserStore creates a durable MySQL-compatible user store.
func NewSQLUserStore(db *sql.DB) (*SQLUserStore, error) {
	return NewSQLUserStoreWithDriver(db, "mysql")
}

// NewSQLUserStoreWithDriver creates a durable user store for a supported RDS driver.
func NewSQLUserStoreWithDriver(db *sql.DB, driver string) (*SQLUserStore, error) {
	if db == nil {
		return nil, errors.New("user database is required")
	}
	_, dialect, err := normalizeShareStoreDriver(driver)
	if err != nil {
		return nil, err
	}
	return &SQLUserStore{db: db, dialect: dialect}, nil
}

// UpsertFromIdentity creates or updates a user from an external identity.
func (s *SQLUserStore) UpsertFromIdentity(user auth.User, now time.Time) (auth.User, error) {
	return s.upsert(user, now, true)
}

// UpsertManaged creates or updates a user from an admin-managed invite.
func (s *SQLUserStore) UpsertManaged(user auth.User, now time.Time) (auth.User, error) {
	return s.upsert(user, now, false)
}

func (s *SQLUserStore) upsert(user auth.User, now time.Time, markLogin bool) (auth.User, error) {
	if s == nil || s.db == nil {
		return auth.User{}, auth.ErrUserStoreUnavailable
	}

	user, err := normalizeUser(user, now, markLogin)
	if err != nil {
		return auth.User{}, err
	}

	var lastLoginAt *time.Time
	if !user.LastLoginAt.IsZero() {
		lastLoginAt = &user.LastLoginAt
	}
	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	_, err = s.db.ExecContext(
		ctx,
		s.upsertStatement(markLogin),
		user.ID,
		user.ExternalID,
		user.Account,
		nullableString(user.Email),
		user.Name,
		string(user.Role),
		string(user.Status),
		nullableString(user.PasswordHash),
		user.CreatedAt.UTC(),
		nullableTime(lastLoginAt),
	)
	if err != nil {
		return auth.User{}, userStoreUnavailable("upsert user", err)
	}

	saved, found, err := s.GetByExternalIDWithError(user.ExternalID)
	if err != nil {
		return auth.User{}, err
	}
	if !found {
		return auth.User{}, userStoreUnavailable("read saved user", sql.ErrNoRows)
	}
	return saved, nil
}

// List returns a stable snapshot of all users sorted by account then ID.
func (s *SQLUserStore) List() []auth.User {
	users, err := s.ListWithError()
	if err != nil {
		log.Printf("list users from database: %v", err)
		return nil
	}
	return users
}

// ListWithError returns all users and preserves database failures for handlers.
func (s *SQLUserStore) ListWithError() ([]auth.User, error) {
	if s == nil || s.db == nil {
		return nil, auth.ErrUserStoreUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	rows, err := s.db.QueryContext(
		ctx,
		s.statement(`SELECT `+userSelectColumns+`
			FROM users
			ORDER BY account ASC, id ASC`),
	)
	if err != nil {
		return nil, userStoreUnavailable("list users", err)
	}
	defer rows.Close()

	users := make([]auth.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, userStoreUnavailable("scan user", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, userStoreUnavailable("iterate users", err)
	}
	return users, nil
}

// GetByID returns a user by internal ID.
func (s *SQLUserStore) GetByID(id string) (auth.User, bool) {
	user, found, err := s.GetByIDWithError(id)
	if err != nil {
		log.Printf("get user by ID from database: %v", err)
		return auth.User{}, false
	}
	return user, found
}

// GetByIDWithError returns a user by internal ID and preserves database failures.
func (s *SQLUserStore) GetByIDWithError(id string) (auth.User, bool, error) {
	if strings.TrimSpace(id) == "" {
		return auth.User{}, false, nil
	}
	return s.getOne(`SELECT `+userSelectColumns+`
			FROM users
			WHERE id = ?`, id)
}

// GetByAccount returns a user by normalized account.
func (s *SQLUserStore) GetByAccount(account string) (auth.User, bool) {
	user, found, err := s.GetByAccountWithError(account)
	if err != nil {
		log.Printf("get user by account from database: %v", err)
		return auth.User{}, false
	}
	return user, found
}

// GetByAccountWithError returns a user by normalized account and preserves database failures.
func (s *SQLUserStore) GetByAccountWithError(account string) (auth.User, bool, error) {
	account = auth.NormalizeAccount(account)
	if account == "" {
		return auth.User{}, false, nil
	}
	return s.getOne(`SELECT `+userSelectColumns+`
			FROM users
			WHERE account = ?`, account)
}

// GetByExternalIDWithError returns a user by the stable identity key.
func (s *SQLUserStore) GetByExternalIDWithError(externalID string) (auth.User, bool, error) {
	if strings.TrimSpace(externalID) == "" {
		return auth.User{}, false, nil
	}
	return s.getOne(`SELECT `+userSelectColumns+`
			FROM users
			WHERE external_id = ?`, externalID)
}

func (s *SQLUserStore) getOne(query, argument string) (auth.User, bool, error) {
	if s == nil || s.db == nil {
		return auth.User{}, false, auth.ErrUserStoreUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	row := s.db.QueryRowContext(ctx, s.statement(query), argument)
	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, false, nil
	}
	if err != nil {
		return auth.User{}, false, userStoreUnavailable("read user", err)
	}
	return user, true, nil
}

// UpdateRole changes a user's role.
func (s *SQLUserStore) UpdateRole(id string, role auth.Role) (auth.User, bool) {
	user, found, err := s.UpdateRoleWithError(id, role)
	if err != nil {
		log.Printf("update user role in database: %v", err)
		return auth.User{}, false
	}
	return user, found
}

// UpdateRoleWithError changes a user's role and preserves database failures.
func (s *SQLUserStore) UpdateRoleWithError(id string, role auth.Role) (auth.User, bool, error) {
	if s == nil || s.db == nil {
		return auth.User{}, false, auth.ErrUserStoreUnavailable
	}
	if strings.TrimSpace(id) == "" {
		return auth.User{}, false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	result, err := s.db.ExecContext(
		ctx,
		s.statement(`UPDATE users SET role = ? WHERE id = ?`),
		string(role),
		id,
	)
	if err != nil {
		return auth.User{}, false, userStoreUnavailable("update user role", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return auth.User{}, false, userStoreUnavailable("inspect updated user", err)
	}
	if affected == 0 {
		return auth.User{}, false, nil
	}
	return s.GetByIDWithError(id)
}

// Disable marks a user as disabled.
func (s *SQLUserStore) Disable(id string, disabledAt time.Time) bool {
	found, err := s.DisableWithError(id, disabledAt)
	if err != nil {
		log.Printf("disable user in database: %v", err)
		return false
	}
	return found
}

// DisableWithError marks a user as disabled and preserves database failures.
func (s *SQLUserStore) DisableWithError(id string, _ time.Time) (bool, error) {
	if s == nil || s.db == nil {
		return false, auth.ErrUserStoreUnavailable
	}
	if strings.TrimSpace(id) == "" {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	result, err := s.db.ExecContext(
		ctx,
		s.statement(`UPDATE users SET status = ? WHERE id = ?`),
		string(auth.UserStatusDisabled),
		id,
	)
	if err != nil {
		return false, userStoreUnavailable("disable user", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, userStoreUnavailable("inspect disabled user", err)
	}
	return affected > 0, nil
}

func (s *SQLUserStore) upsertStatement(markLogin bool) string {
	if s.dialect == shareDatabaseDialectPostgres {
		if markLogin {
			return userPostgresIdentityUpsertQuery
		}
		return userPostgresUpsertQuery
	}
	if markLogin {
		return userIdentityUpsertQuery
	}
	return userUpsertQuery
}

func (s *SQLUserStore) statement(query string) string {
	if s == nil || s.dialect != shareDatabaseDialectPostgres {
		return query
	}

	var builder strings.Builder
	placeholder := 0
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

func normalizeUser(user auth.User, now time.Time, markLogin bool) (auth.User, error) {
	user.Account = auth.NormalizeAccount(user.Account)
	user.Email = auth.NormalizeAccount(user.Email)
	if user.Account == "" {
		user.Account = user.Email
	}
	if user.Account == "" {
		return auth.User{}, errors.New("user account is required")
	}
	if user.ExternalID == "" {
		user.ExternalID = "identity:" + user.Account
	}
	if user.ID == "" {
		id, err := newUserID()
		if err != nil {
			return auth.User{}, err
		}
		user.ID = id
	}
	if strings.TrimSpace(user.Name) == "" {
		user.Name = user.Account
	} else {
		user.Name = strings.TrimSpace(user.Name)
	}
	if user.Role == "" {
		user.Role = auth.RoleViewer
	}
	if user.Status == "" {
		user.Status = auth.UserStatusActive
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if markLogin {
		user.LastLoginAt = now
	}
	return user, nil
}

func scanUser(scanner interface{ Scan(dest ...any) error }) (auth.User, error) {
	var (
		user         auth.User
		email        sql.NullString
		passwordHash sql.NullString
		lastLoginAt  sql.NullTime
		role         string
		status       string
	)
	if err := scanner.Scan(
		&user.ID,
		&user.ExternalID,
		&user.Account,
		&email,
		&user.Name,
		&role,
		&status,
		&passwordHash,
		&user.CreatedAt,
		&lastLoginAt,
	); err != nil {
		return auth.User{}, err
	}
	if email.Valid {
		user.Email = email.String
	}
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}
	user.Role = auth.Role(role)
	user.Status = auth.UserStatus(status)
	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}
	return user, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func newUserID() (string, error) {
	raw := make([]byte, 18)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate user ID: %w", err)
	}
	return "usr_" + base64.RawURLEncoding.EncodeToString(raw), nil
}

func userStoreUnavailable(operation string, err error) error {
	return fmt.Errorf("%w: %s: %v", auth.ErrUserStoreUnavailable, operation, err)
}

// OpenUserStore opens the configured durable user store. Migrations are
// deliberately external so the Provider never changes shared infrastructure
// during startup.
func OpenUserStore(cfg *config.Config) (auth.UserStore, func() error, error) {
	if cfg == nil {
		err := errors.New("provider config is missing")
		return auth.NewUnavailableUserStore(err), nil, err
	}
	if cfg.UserStoreMode == "memory" {
		return newBuiltInMemoryUserStore(), nil, nil
	}
	if cfg.RDSConnectionString == "" {
		err := errors.New("RDS_CONNECTION_STRING is missing")
		return newBuiltInMemoryUserStore(), nil, err
	}

	driver, _, err := normalizeShareStoreDriver(cfg.RDSDriver)
	if err != nil {
		return auth.NewUnavailableUserStore(err), nil, err
	}
	db, err := sql.Open(driver, cfg.RDSConnectionString)
	if err != nil {
		err = fmt.Errorf("open user database: %w", err)
		return auth.NewUnavailableUserStore(err), nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), shareDatabaseTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		err = fmt.Errorf("ping user database: %w", err)
		return auth.NewUnavailableUserStore(err), nil, err
	}
	store, err := NewSQLUserStoreWithDriver(db, driver)
	if err != nil {
		_ = db.Close()
		return auth.NewUnavailableUserStore(err), nil, err
	}
	return store, db.Close, nil
}

func newBuiltInMemoryUserStore() *auth.MemoryUserStore {
	store := auth.NewMemoryUserStore()
	now := time.Now().UTC()
	for _, user := range auth.BuiltInLocalUsers() {
		if _, err := store.UpsertManaged(user, now); err != nil {
			log.Printf("seed built-in local user %s: %v", user.Account, err)
		}
	}
	return store
}
