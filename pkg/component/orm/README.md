# orm

GORM + golang-migrate integration for PostgreSQL, MySQL, and SQLite. It provides a configured `*gorm.DB` factory and a `Migration` type that runs SQL migrations from `./migrations`.

## Configuration

```go
type ORMConfig struct {
	Driver                 string `config:"driver"`
	DSN                    string `config:"dsn"`
	PrepareStmt            bool   `config:"prepare_stmt"`
	SkipDefaultTransaction bool   `config:"skip_default_transaction"`
}
```

Each driver receives its native DSN unchanged:

```yaml
database:
  driver: postgres
  dsn: host=localhost user=myapp password=secret dbname=myapp_db port=5432 sslmode=disable TimeZone=UTC
```

```yaml
database:
  driver: mysql
  dsn: myapp:secret@tcp(localhost:3306)/myapp?multiStatements=true
```

```yaml
database:
  driver: sqlite
  dsn: ./var/myapp.sqlite
```

MySQL requires `multiStatements=true` because the migration runner executes SQL migration files through the same connection.

## Usage

```go
factory := orm.NewDefaultORMFactory()

db, err := factory.Create(orm.ORMConfig{
	Driver: orm.DriverPostgres,
	DSN:    "host=localhost user=myapp password=secret dbname=myapp_db port=5432 sslmode=disable TimeZone=UTC",
}, logger)
```

## Migrations

Migrations are SQL files in `./migrations/` named `{timestamp}_{name}.up.sql` and `{timestamp}_{name}.down.sql`.

```go
migFactory := orm.NewDefaultMigrationFactory()
migration, err := migFactory.Create(db)

migration.Up()     // apply all pending migrations
migration.Down()   // roll back last migration
migration.Reset()  // drop all tables, recreate schema_migrations
migration.Create("add_users_table") // scaffold new migration files
```

`fxorm` registers `migrate up/down/reset/create` as CLI subcommands automatically.

## Repository pattern

Repositories should implement `orm.ExistsChecker` to participate in the unique validator registry. Use `fxorm.AsRepository[T]` to register them in the FX graph.

## Notes

- Supported driver values are exactly `postgres`, `mysql`, and `sqlite`.
- `PrepareStmt` and `SkipDefaultTransaction` map directly to the corresponding GORM connection settings and default to `false`.
- Slow query threshold is hardcoded at 200ms with slog-based GORM logger.
