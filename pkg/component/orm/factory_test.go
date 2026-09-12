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
	requireNoError(t, "create migrations directory", os.Mkdir(migrationsDir, 0o700))
	requireNoError(t, "write up migration", os.WriteFile(
		filepath.Join(migrationsDir, "20260911000000_create_records.up.sql"),
		[]byte("CREATE TABLE records (id INTEGER PRIMARY KEY);"),
		0o600,
	))
	requireNoError(t, "write down migration", os.WriteFile(
		filepath.Join(migrationsDir, "20260911000000_create_records.down.sql"),
		[]byte("DROP TABLE records;"),
		0o600,
	))
	t.Chdir(root)

	db, err := gorm.Open(sqlite.Open(filepath.Join(root, "database.sqlite")), &gorm.Config{})
	requireNoError(t, "open SQLite database", err)
	sqlDB, err := db.DB()
	requireNoError(t, "get sql.DB", err)
	t.Cleanup(func() {
		requireNoError(t, "close SQLite database", sqlDB.Close())
	})

	migration, err := NewDefaultMigrationFactory().Create(db)
	requireNoError(t, "create SQLite migration", err)
	requireNoError(t, "migrate up", migration.Up())
	requireTablePresence(t, db, "records", true)
	requireNoError(t, "migrate down", migration.Down())
	requireTablePresence(t, db, "records", false)
	requireNoError(t, "migrate up after down", migration.Up())
	requireNoError(t, "reset migrations", migration.Reset())
	requireTablePresence(t, db, "records", false)
	requireTablePresence(t, db, "schema_migrations", true)
	requireNoError(t, "migrate up after reset", migration.Up())
}

func requireNoError(t *testing.T, operation string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", operation, err)
	}
}

func requireTablePresence(t *testing.T, db *gorm.DB, table string, expected bool) {
	t.Helper()
	if db.Migrator().HasTable(table) != expected {
		t.Fatalf("table %q presence: got %t, want %t", table, !expected, expected)
	}
}
