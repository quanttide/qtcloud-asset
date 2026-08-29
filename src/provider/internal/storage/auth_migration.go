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

// PostgresUserMigrationVersion identifies the controlled PostgreSQL users migration.
const PostgresUserMigrationVersion = "users-postgres-v1"

//go:embed auth_users_postgres.sql
var authUsersPostgresSchema string

// ApplyUserMigration applies the fixed, idempotent PostgreSQL users schema.
// It is intentionally available only for the explicitly named migration.
func ApplyUserMigration(ctx context.Context, cfg *config.Config) error {
	if cfg == nil {
		return errors.New("provider config is missing")
	}
	if cfg.UserMigration != PostgresUserMigrationVersion {
		return fmt.Errorf("unsupported user migration %q", cfg.UserMigration)
	}
	if cfg.RDSConnectionString == "" {
		return errors.New("RDS_CONNECTION_STRING is missing")
	}

	driver, dialect, err := normalizeShareStoreDriver(cfg.RDSDriver)
	if err != nil {
		return err
	}
	if dialect != shareDatabaseDialectPostgres {
		return fmt.Errorf("user migration %q requires a PostgreSQL driver", cfg.UserMigration)
	}

	db, err := sql.Open(driver, cfg.RDSConnectionString)
	if err != nil {
		return fmt.Errorf("open user migration database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, shareDatabaseTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping user migration database: %w", err)
	}
	for _, statement := range postgresUserMigrationStatements() {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply user migration: %w", err)
		}
	}
	return nil
}

func postgresUserMigrationStatements() []string {
	rawStatements := strings.Split(authUsersPostgresSchema, ";")
	statements := make([]string, 0, len(rawStatements))
	for _, statement := range rawStatements {
		statement = strings.TrimSpace(statement)
		if statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}
