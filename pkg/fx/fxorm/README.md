# fxorm

FX module for [pkg/component/orm](../../component/orm). Provides a `*gorm.DB` connection, a `*orm.Migration` runner, CLI migration subcommands, and a unique-field validator registered into the `fxvalidator` group.

## Quick start

```go
fx.New(
    fxconfig.FxConfigModule,
    fxlogger.FxLoggerModule,
    fxvalidator.FxValidatorModule,
    fxorm.FxORMModule,
)
```

## Configuration

```yaml
database:
  host: localhost
  user: myapp
  password: secret
  name: myapp_db
  port: 5432
  sslmode: disable
```

## CLI commands registered

`fxorm` adds a `migrate` command with four subcommands:

```
app migrate create <name>   # scaffold new up/down SQL files
app migrate up              # apply all pending migrations
app migrate down            # roll back last migration
app migrate reset           # drop all tables, re-create schema_migrations
```

## Registering repositories

```go
fxorm.AsRepository[MyRepository](NewMyRepository)
```

This registers the repository:
1. As its concrete type for direct injection.
2. As `orm.ExistsChecker` in the `repository-checkers` group, so the unique validator can check for existing records.

## Injected types

- `*gorm.DB` — ready-to-use GORM connection
- `*orm.Migration` — migration runner

## Notes

- Only PostgreSQL is supported.
- Slow query threshold is 200ms (hardcoded). Queries above this log at WARN level via slog.
