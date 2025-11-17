package orm

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

// Migration handles database schema migrations
type Migration struct {
	migrator *migrate.Migrate
	db       *gorm.DB
}

// NewMigration creates a new Migration instance
func NewMigration(migrator *migrate.Migrate, db *gorm.DB) *Migration {
	return &Migration{
		migrator: migrator,
		db:       db,
	}
}

// Create creates a new migration file with the given name
func (s *Migration) Create(name string) error {
	// Use timestamp-based naming to avoid conflicts between developers
	// Format: YYYYMMDDHHMMSS (20060102150405 in Go time format)
	timestamp := time.Now().Format("20060102150405")

	migrationsDir := "migrations"
	upFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.up.sql", timestamp, name))
	downFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.down.sql", timestamp, name))

	// Create empty migration files
	if err := os.WriteFile(upFile, []byte(""), 0o644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}
	if err := os.WriteFile(downFile, []byte(""), 0o644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	return nil
}

// Up applies all pending migrations
func (s *Migration) Up() error {
	if err := s.migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations up: %w", err)
	}
	return nil
}

// Down rolls back the last migration
func (s *Migration) Down() error {
	if err := s.migrator.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations down: %w", err)
	}
	return nil
}

// Reset drops all tables and recreates the schema_migrations table
func (s *Migration) Reset() error {
	if err := s.migrator.Drop(); err != nil {
		return fmt.Errorf("failed to reset database: %w", err)
	}

	// After dropping all tables, we need to recreate the schema_migrations table
	// so that future migrations can run. This is a simple, clean approach.
	if err := s.ensureSchemaMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	return nil
}

func (s *Migration) ensureSchemaMigrationsTable() error {
	// Create schema_migrations table if it doesn't exist
	// This matches the structure expected by golang-migrate
	if err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT NOT NULL PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	return nil
}
