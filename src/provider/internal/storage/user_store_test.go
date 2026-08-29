package storage

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
)

func TestSQLUserStorePersistsManagedUsersAcrossStoreInstances(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	firstStore, err := NewSQLUserStoreWithDriver(db, "postgres")
	if err != nil {
		t.Fatalf("create first user store: %v", err)
	}
	secondStore, err := NewSQLUserStoreWithDriver(db, "postgres")
	if err != nil {
		t.Fatalf("create second user store: %v", err)
	}

	createdAt := time.Date(2026, time.August, 29, 10, 0, 0, 0, time.UTC)
	passwordHash := "pbkdf2_sha256$1000$salt$hash"
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users
			(id, external_id, account, email, name, role, status, password_hash, created_at, last_login_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (external_id) DO UPDATE SET
				account = EXCLUDED.account,
				email = EXCLUDED.email,
				name = EXCLUDED.name,
				role = EXCLUDED.role,
				status = EXCLUDED.status,
				password_hash = COALESCE(NULLIF(EXCLUDED.password_hash, ''), users.password_hash),
				last_login_at = COALESCE(EXCLUDED.last_login_at, users.last_login_at)`)).
		WithArgs(
			"user-1",
			"managed:viewer",
			"viewer",
			nil,
			"Viewer",
			auth.RoleViewer,
			auth.UserStatusActive,
			passwordHash,
			createdAt,
			nil,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_id, account, email, name, role, status, password_hash, created_at, last_login_at
			FROM users
			WHERE external_id = $1`)).
		WithArgs("managed:viewer").
		WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(
			"user-1",
			"managed:viewer",
			"viewer",
			nil,
			"Viewer",
			auth.RoleViewer,
			auth.UserStatusActive,
			passwordHash,
			createdAt,
			nil,
		))

	saved, err := firstStore.UpsertManaged(auth.User{
		ID:           "user-1",
		ExternalID:   "managed:viewer",
		Account:      "viewer",
		Name:         "Viewer",
		Role:         auth.RoleViewer,
		Status:       auth.UserStatusActive,
		PasswordHash: passwordHash,
	}, createdAt)
	if err != nil {
		t.Fatalf("persist managed user: %v", err)
	}
	if saved.ID != "user-1" || saved.Account != "viewer" {
		t.Fatalf("unexpected saved managed user: %+v", saved)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_id, account, email, name, role, status, password_hash, created_at, last_login_at
			FROM users
			ORDER BY account ASC, id ASC`)).
		WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(
			"user-1",
			"managed:viewer",
			"viewer",
			nil,
			"Viewer",
			auth.RoleViewer,
			auth.UserStatusActive,
			passwordHash,
			createdAt,
			nil,
		))

	users, err := secondStore.ListWithError()
	if err != nil {
		t.Fatalf("read users from second store: %v", err)
	}
	if len(users) != 1 || users[0].ID != "user-1" || users[0].PasswordHash != passwordHash {
		t.Fatalf("unexpected users loaded from database: %+v", users)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLUserStoreIdentityUpsertPreservesExistingRoleAndStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	store, err := NewSQLUserStoreWithDriver(db, "postgres")
	if err != nil {
		t.Fatalf("create postgres user store: %v", err)
	}
	now := time.Date(2026, time.August, 29, 11, 0, 0, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta(userPostgresIdentityUpsertQuery)).
		WithArgs(
			"user-1",
			"lark-user-1",
			"lixiang",
			nil,
			"Li Xiang",
			auth.RoleViewer,
			auth.UserStatusActive,
			nil,
			now,
			now,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_id, account, email, name, role, status, password_hash, created_at, last_login_at
			FROM users
			WHERE external_id = $1`)).
		WithArgs("lark-user-1").
		WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(
			"user-1",
			"lark-user-1",
			"lixiang",
			nil,
			"Li Xiang",
			auth.RoleAdmin,
			auth.UserStatusDisabled,
			"existing-hash",
			now,
			now,
		))

	saved, err := store.UpsertFromIdentity(auth.User{
		ID:         "user-1",
		ExternalID: "lark-user-1",
		Account:    "lixiang",
		Name:       "Li Xiang",
		Role:       auth.RoleViewer,
		Status:     auth.UserStatusActive,
	}, now)
	if err != nil {
		t.Fatalf("upsert identity user: %v", err)
	}
	if saved.Role != auth.RoleAdmin || saved.Status != auth.UserStatusDisabled || saved.PasswordHash != "existing-hash" {
		t.Fatalf("identity upsert must return persisted authorization fields, got %+v", saved)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLUserStorePreservesUnavailableDatabaseErrors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	store, err := NewSQLUserStoreWithDriver(db, "postgres")
	if err != nil {
		t.Fatalf("create postgres user store: %v", err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_id, account, email, name, role, status, password_hash, created_at, last_login_at
			FROM users
			ORDER BY account ASC, id ASC`)).
		WillReturnError(errors.New("database is down"))

	_, err = store.ListWithError()
	if !errors.Is(err, auth.ErrUserStoreUnavailable) {
		t.Fatalf("expected unavailable user store error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestOpenUserStoreUsesMemoryOnlyWhenRDSIsNotConfigured(t *testing.T) {
	store, closeStore, err := OpenUserStore(&config.Config{UserStoreMode: "memory"})
	if err != nil {
		t.Fatalf("open memory user store: %v", err)
	}
	if closeStore != nil {
		t.Fatal("memory user store should not have a close function")
	}
	if _, ok := store.(*auth.MemoryUserStore); !ok {
		t.Fatalf("expected memory user store, got %T", store)
	}
}

func TestOpenUserStoreFailsClosedWithoutRDSConfiguration(t *testing.T) {
	store, closeStore, err := OpenUserStore(&config.Config{UserStoreMode: "rds"})
	if err == nil {
		t.Fatal("expected missing RDS configuration to return an error")
	}
	if closeStore != nil {
		t.Fatal("unavailable user store should not have a close function")
	}
	if _, ok := store.(*auth.UnavailableUserStore); !ok {
		t.Fatalf("expected unavailable user store, got %T", store)
	}
}

func TestPostgresUserMigrationStatementsAreIdempotent(t *testing.T) {
	statements := postgresUserMigrationStatements()
	if len(statements) != 2 {
		t.Fatalf("expected two PostgreSQL user migration statements, got %d", len(statements))
	}
	for _, statement := range statements {
		if !regexp.MustCompile(`(?i)CREATE (TABLE|INDEX) IF NOT EXISTS`).MatchString(statement) {
			t.Fatalf("expected idempotent DDL statement, got %q", statement)
		}
	}
}

func userColumns() []string {
	return []string{
		"id",
		"external_id",
		"account",
		"email",
		"name",
		"role",
		"status",
		"password_hash",
		"created_at",
		"last_login_at",
	}
}
