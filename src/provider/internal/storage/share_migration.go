package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/quanttide/qtcloud-asset/provider/internal/config"
)

// PostgresShareMigrationVersion is the only migration identifier that the
// Provider will execute during the controlled production bootstrap.
const PostgresShareMigrationVersion = "folder-shares-postgres-v1"

//go:embed folder_shares_postgres.sql
var folderSharesPostgresSchema string

// ApplyShareMigration applies the fixed, idempotent PostgreSQL share schema.
// It is intentionally available only for the explicitly named migration.
func ApplyShareMigration(ctx context.Context, cfg *config.Config) error {
	if cfg == nil {
		return errors.New("provider config is missing")
	}
	if cfg.ShareMigration != PostgresShareMigrationVersion {
		return fmt.Errorf("unsupported share migration %q", cfg.ShareMigration)
	}
	if cfg.RDSConnectionString == "" {
		return errors.New("RDS_CONNECTION_STRING is missing")
	}

	driver, dialect, err := normalizeShareStoreDriver(cfg.RDSDriver)
	if err != nil {
		return err
	}
	if dialect != shareDatabaseDialectPostgres {
		return fmt.Errorf("share migration %q requires a PostgreSQL driver", cfg.ShareMigration)
	}

	db, err := sql.Open(driver, cfg.RDSConnectionString)
	if err != nil {
		return fmt.Errorf("open share migration database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, shareDatabaseTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping share migration database: %w", err)
	}
	for _, statement := range postgresShareMigrationStatements() {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply share migration: %w", err)
		}
	}
	return nil
}

func postgresShareMigrationStatements() []string {
	rawStatements := strings.Split(folderSharesPostgresSchema, ";")
	statements := make([]string, 0, len(rawStatements))
	for _, statement := range rawStatements {
		statement = strings.TrimSpace(statement)
		if statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}
