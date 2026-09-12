package orm

import (
	"fmt"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

const (
	// DriverPostgres selects PostgreSQL.
	DriverPostgres = "postgres"
	// DriverMySQL selects MySQL.
	DriverMySQL = "mysql"
	// DriverSQLite selects SQLite.
	DriverSQLite = "sqlite"
)

type ORMConfig struct {
	// Driver selects the database implementation: postgres, mysql, or sqlite.
	Driver string `config:"driver"`

	// DSN is the native connection string for the selected driver.
	DSN string `config:"dsn"`

	// PrepareStmt caches prepared statements for repeated queries.
	PrepareStmt bool `config:"prepare_stmt"`

	// SkipDefaultTransaction avoids wrapping each individual write in its own transaction.
	SkipDefaultTransaction bool `config:"skip_default_transaction"`
}

// Validate checks that the selected driver has enough information to open a
// connection used by both GORM and the migration runner.
func (c *ORMConfig) Validate() error {
	switch c.Driver {
	case DriverPostgres, DriverMySQL, DriverSQLite:
	case "":
		return fmt.Errorf("database driver is required")
	default:
		return fmt.Errorf("unsupported database driver %q", c.Driver)
	}
	if c.DSN == "" {
		return fmt.Errorf("database dsn is required")
	}
	if c.Driver != DriverMySQL {
		return nil
	}

	dsn, err := mysqlDriver.ParseDSN(c.DSN)
	if err != nil {
		return fmt.Errorf("invalid mysql dsn: %w", err)
	}
	if !dsn.MultiStatements {
		return fmt.Errorf("mysql dsn must enable multiStatements for migrations")
	}

	return nil
}
