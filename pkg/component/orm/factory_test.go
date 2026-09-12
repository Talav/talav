//go:build cgo

package orm

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/gorm"
)

func TestGORMDialector(t *testing.T) {
	tests := []struct {
		name     string
		config   ORMConfig
		expected string
	}{
		{
			name: "PostgreSQL",
			config: ORMConfig{
				Driver: DriverPostgres,
				DSN:    "host=localhost user=app dbname=app port=5432 sslmode=disable",
			},
			expected: DriverPostgres,
		},
		{
			name: "MySQL",
			config: ORMConfig{
				Driver: DriverMySQL,
				DSN:    "app:secret@tcp(localhost:3306)/app?multiStatements=true",
			},
			expected: DriverMySQL,
		},
		{
			name: "SQLite",
			config: ORMConfig{
				Driver: DriverSQLite,
				DSN:    "database.sqlite",
			},
			expected: DriverSQLite,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dialector, err := gormDialector(test.config)
			requireNoError(t, "create GORM dialector", err)
			if dialector.Name() != test.expected {
				t.Fatalf("dialector name: got %q, want %q", dialector.Name(), test.expected)
			}
		})
	}
}

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

	db, err := NewDefaultORMFactory().Create(ORMConfig{
		Driver:                 DriverSQLite,
		DSN:                    filepath.Join(root, "database.sqlite"),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	requireNoError(t, "open SQLite database", err)
	if !db.PrepareStmt {
		t.Fatal("prepared statements are disabled")
	}
	if !db.SkipDefaultTransaction {
		t.Fatal("default transactions are enabled")
	}
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
