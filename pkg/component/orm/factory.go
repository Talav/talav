package orm

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	migrateMySQL "github.com/golang-migrate/migrate/v4/database/mysql"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	migrateSQLite "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ORMFactory is the interface for [gorm.DB] factories.
type ORMFactory interface {
	Create(cfg ORMConfig, logger *slog.Logger) (*gorm.DB, error)
}

// DefaultORMFactory is the default [ORMFactory] implementation.
type DefaultORMFactory struct{}

// NewDefaultORMFactory returns a [DefaultORMFactory], implementing [ORMFactory].
func NewDefaultORMFactory() ORMFactory {
	return &DefaultORMFactory{}
}

// Create returns a new [gorm.DB] from the given [ORMConfig].
func (f *DefaultORMFactory) Create(cfg ORMConfig, logger *slog.Logger) (*gorm.DB, error) {
	dialector, err := gormDialector(cfg)
	if err != nil {
		return nil, err
	}

	gormConfig := &gorm.Config{
		PrepareStmt:            cfg.PrepareStmt,
		SkipDefaultTransaction: cfg.SkipDefaultTransaction,
		Logger: gormlogger.NewSlogLogger(
			logger,
			gormlogger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  gormlogger.Info,
				IgnoreRecordNotFoundError: false,
				ParameterizedQueries:      false,
			},
		),
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %w", cfg.Driver, err)
	}

	return db, nil
}

func gormDialector(cfg ORMConfig) (gorm.Dialector, error) {
	switch cfg.Driver {
	case DriverPostgres:
		return postgres.Open(cfg.DSN), nil
	case DriverMySQL:
		return mysql.Open(cfg.DSN), nil
	case DriverSQLite:
		return sqlite.Open(cfg.DSN), nil
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}
}

// MigrationFactory is the interface for [Migration] factories.
type MigrationFactory interface {
	Create(db *gorm.DB) (*Migration, error)
}

// DefaultMigrationFactory is the default [MigrationFactory] implementation.
type DefaultMigrationFactory struct{}

// NewDefaultMigrationFactory returns a [DefaultMigrationFactory], implementing [MigrationFactory].
func NewDefaultMigrationFactory() MigrationFactory {
	return &DefaultMigrationFactory{}
}

// Create returns a new [Migration] instance from the given [gorm.DB].
func (f *DefaultMigrationFactory) Create(db *gorm.DB) (*Migration, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	var driver database.Driver
	dialect := db.Name()
	switch dialect {
	case DriverPostgres:
		driver, err = migratePostgres.WithInstance(sqlDB, &migratePostgres.Config{})
	case DriverMySQL:
		driver, err = migrateMySQL.WithInstance(sqlDB, &migrateMySQL.Config{})
	case DriverSQLite:
		driver, err = migrateSQLite.WithInstance(sqlDB, &migrateSQLite.Config{})
	default:
		return nil, fmt.Errorf("unsupported migration database %q", dialect)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create %s migration driver: %w", db.Name(), err)
	}

	migrationsDir := "migrations"
	migrator, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir, dialect, driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return NewMigration(migrator, db), nil
}
