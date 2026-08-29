package storage

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/quanttide/qtcloud-asset/provider/internal/share"
)

func testShareEncryptionKey() []byte {
	return []byte("01234567890123456789012345678901")
}

type encryptedTokenArgument struct {
	key       []byte
	plaintext string
}

func (a encryptedTokenArgument) Match(value driver.Value) bool {
	encrypted, ok := value.([]byte)
	if !ok || string(encrypted) == a.plaintext {
		return false
	}
	decrypted, err := share.DecryptToken(a.key, encrypted)
	return err == nil && decrypted == a.plaintext
}

func TestSQLShareStoreCreateStoresHashedAndEncryptedToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	store, err := NewSQLShareStore(db, testShareEncryptionKey())
	if err != nil {
		t.Fatalf("create share store: %v", err)
	}
	createdAt := time.Date(2026, time.August, 27, 10, 0, 0, 0, time.UTC)
	token := "fixed-share-token"
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO folder_shares
			(token_hash, token_ciphertext, title, bucket, prefixes, created_by, created_at, revoked_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs(
			share.TokenHash(token),
			encryptedTokenArgument{key: testShareEncryptionKey(), plaintext: token},
			"Design",
			"qtcloud-asset-studio",
			[]byte(`["design/"]`),
			"owner-1",
			createdAt,
			nil,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	created, err := store.Create(share.Record{
		Token:     token,
		Title:     "Design",
		Bucket:    "qtcloud-asset-studio",
		Prefixes:  []string{"design/"},
		CreatedBy: "owner-1",
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("create share: %v", err)
	}
	if created.Token != token {
		t.Fatalf("expected returned token %q, got %q", token, created.Token)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestPostgresShareStoreCreateUsesPostgresPlaceholdersAndJSONText(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	store, err := NewSQLShareStoreWithDriver(db, testShareEncryptionKey(), "postgres")
	if err != nil {
		t.Fatalf("create postgres share store: %v", err)
	}
	createdAt := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	token := "postgres-share-token"
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO folder_shares
			(token_hash, token_ciphertext, title, bucket, prefixes, created_by, created_at, revoked_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)).
		WithArgs(
			share.TokenHash(token),
			encryptedTokenArgument{key: testShareEncryptionKey(), plaintext: token},
			"Postgres share",
			"qtcloud-asset-studio",
			`["assets/"]`,
			"owner-1",
			createdAt,
			nil,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = store.Create(share.Record{
		Token:     token,
		Title:     "Postgres share",
		Bucket:    "qtcloud-asset-studio",
		Prefixes:  []string{"assets/"},
		CreatedBy: "owner-1",
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("create postgres share: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestShareTargetEncodingKeepsLegacyFoldersAndSupportsFileKeys(t *testing.T) {
	legacy, err := encodeShareTargets([]string{"docs/"}, nil)
	if err != nil {
		t.Fatalf("encode legacy targets: %v", err)
	}
	if string(legacy) != `["docs/"]` {
		t.Fatalf("unexpected legacy target encoding: %s", legacy)
	}

	encoded, err := encodeShareTargets(nil, []string{"docs/readme.md"})
	if err != nil {
		t.Fatalf("encode file targets: %v", err)
	}
	var targetMap map[string][]string
	if err := json.Unmarshal(encoded, &targetMap); err != nil {
		t.Fatalf("decode file target encoding: %v", err)
	}
	if got := targetMap["keys"]; len(got) != 1 || got[0] != "docs/readme.md" {
		t.Fatalf("unexpected file target encoding: %s", encoded)
	}

	var prefixes, keys []string
	if err := decodeShareTargets(legacy, &prefixes, &keys); err != nil {
		t.Fatalf("decode legacy targets: %v", err)
	}
	if len(prefixes) != 1 || prefixes[0] != "docs/" || len(keys) != 0 {
		t.Fatalf("unexpected decoded legacy targets: prefixes=%v keys=%v", prefixes, keys)
	}
	if err := decodeShareTargets(encoded, &prefixes, &keys); err != nil {
		t.Fatalf("decode file targets: %v", err)
	}
	if len(prefixes) != 0 || len(keys) != 1 || keys[0] != "docs/readme.md" {
		t.Fatalf("unexpected decoded file targets: prefixes=%v keys=%v", prefixes, keys)
	}
}

func TestNormalizeShareStoreDriverSupportsProductionPostgresAliases(t *testing.T) {
	tests := map[string]string{
		"postgres":   "pgx",
		"postgresql": "pgx",
		"pgx":        "pgx",
	}
	for input, wantDriver := range tests {
		t.Run(input, func(t *testing.T) {
			driver, dialect, err := normalizeShareStoreDriver(input)
			if err != nil {
				t.Fatalf("normalize driver: %v", err)
			}
			if driver != wantDriver || dialect != shareDatabaseDialectPostgres {
				t.Fatalf("unexpected normalized driver=%q dialect=%v", driver, dialect)
			}
		})
	}

	if _, _, err := normalizeShareStoreDriver("sqlite"); err == nil {
		t.Fatal("expected unsupported share database driver to fail closed")
	}
}

func TestPostgresShareMigrationStatementsAreIdempotent(t *testing.T) {
	statements := postgresShareMigrationStatements()
	if len(statements) != 3 {
		t.Fatalf("expected three PostgreSQL migration statements, got %d", len(statements))
	}
	for _, statement := range statements {
		if !regexp.MustCompile(`(?i)CREATE (TABLE|INDEX) IF NOT EXISTS`).MatchString(statement) {
			t.Fatalf("expected idempotent DDL statement, got %q", statement)
		}
	}
}

func TestSQLShareStoreGetDecryptsTokenAndDecodesPrefixes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	store, err := NewSQLShareStore(db, testShareEncryptionKey())
	if err != nil {
		t.Fatalf("create share store: %v", err)
	}
	token := "persisted-share-token"
	encrypted, err := share.EncryptToken(testShareEncryptionKey(), token)
	if err != nil {
		t.Fatalf("encrypt token fixture: %v", err)
	}
	createdAt := time.Date(2026, time.August, 27, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT token_ciphertext, title, bucket, prefixes, created_by, created_at, revoked_at
		 FROM folder_shares
		 WHERE token_hash = ?`)).
		WithArgs(share.TokenHash(token)).
		WillReturnRows(sqlmock.NewRows([]string{
			"token_ciphertext", "title", "bucket", "prefixes", "created_by", "created_at", "revoked_at",
		}).AddRow(
			encrypted,
			"Design",
			"qtcloud-asset-studio",
			[]byte(`["design/","brand/"]`),
			"owner-1",
			createdAt,
			nil,
		))

	loaded, ok, err := store.Get(token)
	if err != nil {
		t.Fatalf("get share: %v", err)
	}
	if !ok {
		t.Fatal("expected persisted share")
	}
	if loaded.Token != token || len(loaded.Prefixes) != 2 || loaded.Prefixes[1] != "brand/" {
		t.Fatalf("unexpected loaded share: %+v", loaded)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLShareStoreListByOwnerReturnsRevokedRecordsForAudit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	store, err := NewSQLShareStore(db, testShareEncryptionKey())
	if err != nil {
		t.Fatalf("create share store: %v", err)
	}
	firstToken := "first-token"
	secondToken := "second-token"
	firstEncrypted, err := share.EncryptToken(testShareEncryptionKey(), firstToken)
	if err != nil {
		t.Fatalf("encrypt first token: %v", err)
	}
	secondEncrypted, err := share.EncryptToken(testShareEncryptionKey(), secondToken)
	if err != nil {
		t.Fatalf("encrypt second token: %v", err)
	}
	revokedAt := time.Date(2026, time.August, 27, 11, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT token_ciphertext, title, bucket, prefixes, created_by, created_at, revoked_at
		 FROM folder_shares
		 WHERE created_by = ?
		 ORDER BY created_at DESC, token_hash ASC`)).
		WithArgs("owner-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"token_ciphertext", "title", "bucket", "prefixes", "created_by", "created_at", "revoked_at",
		}).
			AddRow(secondEncrypted, "Second", "bucket", []byte(`["b/"]`), "owner-1", revokedAt, nil).
			AddRow(firstEncrypted, "First", "bucket", []byte(`["a/"]`), "owner-1", revokedAt.Add(-time.Hour), revokedAt))

	records, err := store.ListByOwner("owner-1")
	if err != nil {
		t.Fatalf("list shares: %v", err)
	}
	if len(records) != 2 || records[0].Token != secondToken || records[1].RevokedAt == nil {
		t.Fatalf("unexpected owner shares: %+v", records)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLShareStoreRevokeRetainsRecord(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	store, err := NewSQLShareStore(db, testShareEncryptionKey())
	if err != nil {
		t.Fatalf("create share store: %v", err)
	}
	token := "revoke-token"
	revokedAt := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE folder_shares
		 SET revoked_at = COALESCE(revoked_at, ?)
		 WHERE token_hash = ?`)).
		WithArgs(revokedAt, share.TokenHash(token)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	encrypted, err := share.EncryptToken(testShareEncryptionKey(), token)
	if err != nil {
		t.Fatalf("encrypt token fixture: %v", err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT token_ciphertext, title, bucket, prefixes, created_by, created_at, revoked_at
		 FROM folder_shares
		 WHERE token_hash = ?`)).
		WithArgs(share.TokenHash(token)).
		WillReturnRows(sqlmock.NewRows([]string{
			"token_ciphertext", "title", "bucket", "prefixes", "created_by", "created_at", "revoked_at",
		}).AddRow(
			encrypted,
			"Revoke",
			"bucket",
			[]byte(`["a/"]`),
			"owner-1",
			revokedAt.Add(-time.Hour),
			revokedAt,
		))

	record, ok, err := store.Revoke(token, revokedAt)
	if err != nil {
		t.Fatalf("revoke share: %v", err)
	}
	if !ok || record.RevokedAt == nil || !record.RevokedAt.Equal(revokedAt) {
		t.Fatalf("unexpected revoked share: %+v", record)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestNewSQLShareStoreRequiresAES256Key(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	if _, err := NewSQLShareStore(db, []byte("short")); err == nil {
		t.Fatal("expected invalid encryption key to be rejected")
	}
}

func TestSQLShareStoreGetReturnsNotFoundWithoutError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	store, err := NewSQLShareStore(db, testShareEncryptionKey())
	if err != nil {
		t.Fatalf("create share store: %v", err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT token_ciphertext, title, bucket, prefixes, created_by, created_at, revoked_at
		 FROM folder_shares
		 WHERE token_hash = ?`)).
		WithArgs(share.TokenHash("missing")).
		WillReturnError(sql.ErrNoRows)

	_, ok, err := store.Get("missing")
	if err != nil || ok {
		t.Fatalf("expected clean not-found result, got ok=%v err=%v", ok, err)
	}
}
