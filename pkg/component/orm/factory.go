package orm

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
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
//
// Example:
//
//	var factory = NewDefaultORMFactory()
//	var db, _ = factory.Create(ORMConfig{
//		Host:     "localhost",
//		User:     "user",
//		Password: "password",
//		Name:     "dbname",
//		Port:     5432,
//		SSLMode:  "disable",
//	}, logger)
func (f *DefaultORMFactory) Create(cfg ORMConfig, logger *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)

	gormConfig := &gorm.Config{
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

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
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

	// Create PostgreSQL driver
	driver, err := migratePostgres.WithInstance(sqlDB, &migratePostgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create database driver: %w", err)
	}

	// Initialize migrate instance
	migrationsDir := "migrations"
	migrator, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return NewMigration(migrator, db), nil
}
