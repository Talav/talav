# orm

GORM + golang-migrate wrapper for PostgreSQL. Provides a `*gorm.DB` factory and a `Migration` type for running SQL migrations from the `./migrations` directory.

## Configuration

```go
type ORMConfig struct {
    Host     string `config:"host"`
    User     string `config:"user"`
    Password string `config:"password"`
    Name     string `config:"name"`
    Port     int    `config:"port"`
    SSLMode  string `config:"sslmode"`
}
```

```yaml
database:
  host: localhost
  user: myapp
  password: secret
  name: myapp_db
  port: 5432
  sslmode: disable
```

## Usage

```go
factory := orm.NewDefaultORMFactory()

db, err := factory.Create(orm.ORMConfig{
    Host:     "localhost",
    User:     "myapp",
    Password: "secret",
    Name:     "myapp_db",
    Port:     5432,
    SSLMode:  "disable",
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

- Only PostgreSQL is supported currently. The `driver` config field is present but unused.
- Slow query threshold is hardcoded at 200ms with slog-based GORM logger.
