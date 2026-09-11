//go:build cgo

package orm

import (
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDefaultMigrationFactory_Create_SQLite(t *testing.T) {
	root := t.TempDir()
	migrationsDir := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrationsDir, 0o700); err != nil {
		t.Fatalf("create migrations directory: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(migrationsDir, "20260911000000_create_records.up.sql"),
		[]byte("CREATE TABLE records (id INTEGER PRIMARY KEY);"),
		0o600,
	); err != nil {
		t.Fatalf("write up migration: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(migrationsDir, "20260911000000_create_records.down.sql"),
		[]byte("DROP TABLE records;"),
		0o600,
	); err != nil {
		t.Fatalf("write down migration: %v", err)
	}
	t.Chdir(root)

	db, err := gorm.Open(sqlite.Open(filepath.Join(root, "database.sqlite")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open SQLite database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close SQLite database: %v", err)
		}
	})

	migration, err := NewDefaultMigrationFactory().Create(db)
	if err != nil {
		t.Fatalf("create SQLite migration: %v", err)
	}
	if err := migration.Up(); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	if !db.Migrator().HasTable("records") {
		t.Fatal("migrate up did not create records table")
	}
	if err := migration.Down(); err != nil {
		t.Fatalf("migrate down: %v", err)
	}
	if db.Migrator().HasTable("records") {
		t.Fatal("migrate down did not drop records table")
	}
	if err := migration.Up(); err != nil {
		t.Fatalf("migrate up after down: %v", err)
	}
	if err := migration.Reset(); err != nil {
		t.Fatalf("reset migrations: %v", err)
	}
	if db.Migrator().HasTable("records") {
		t.Fatal("reset did not drop records table")
	}
	if !db.Migrator().HasTable("schema_migrations") {
		t.Fatal("reset did not recreate schema migrations table")
	}
	if err := migration.Up(); err != nil {
		t.Fatalf("migrate up after reset: %v", err)
	}
}
