# Config Package

A flexible, framework-agnostic configuration management package built on top of [koanf](https://github.com/knadh/koanf). Supports YAML and `.env` files with environment-specific overrides, environment variable expansion, and consistent key normalization.

## Features

- **Multiple Config Sources**: Support for YAML and `.env` files with custom parsers
- **Environment-Specific Configs**: Automatic loading based on `APP_ENV` with `{env}` placeholder support
- **Environment Variable Expansion**: `${VAR}` placeholder expansion in config values
- **Key Normalization**: Consistent key naming (underscores → dots, uppercase → lowercase)
- **Priority System**: Clear override order (YAML → .env → environment variables)
- **Type-Safe Unmarshaling**: Struct-based configuration with `config` struct tags
- **Framework Agnostic**: Use any `koanf.Parser` and define your own file patterns

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/talav/talav/pkg/component/config"
)

func main() {
    factory := config.NewDefaultConfigFactory()
    
    // Use default sources (./configs/*.yaml and ./.env*)
    cfg, err := factory.Create()
    if err != nil {
        panic(err)
    }
    
    // Access values directly
    appName := cfg.k.String("app.name")
    dbHost := cfg.k.String("database.host")
    
    // Or unmarshal into structs
    type AppConfig struct {
        Name    string `config:"name"`
        Version string `config:"version"`
    }
    var appCfg AppConfig
    cfg.UnmarshalKey("app", &appCfg)
    
    fmt.Println(appCfg.Name)
}
```

## Configuration File Structure

### Default File Patterns

When using `factory.Create()` without arguments, the following files are loaded in order:

**YAML Files** (from `./configs/`):
- `config.yaml` - Base configuration
- `config_{env}.yaml` - Environment-specific (e.g., `config_dev.yaml`)
- `config_{env}_local.yaml` - Local overrides (e.g., `config_dev_local.yaml`)

**Environment Files** (from `.`):
- `.env` - Base environment variables
- `.env.{env}` - Environment-specific (e.g., `.env.dev`)
- `.env.local` - Local overrides
- `.env.{env}.local` - Environment-specific local overrides (e.g., `.env.dev.local`)

**Environment Variables** (always loaded last, highest priority):
- `APP_NAME` → `app.name`
- `DATABASE_HOST` → `database.host`
- `APP_DATABASE_PORT` → `app.database.port`

### Example Configuration Files

**`configs/config.yaml`**:
```yaml
app:
  name: myapp
  version: 1.0.0
  data_dir: ${HOME}/data

database:
  host: localhost
  port: 5432
  user: admin
  name: mydb
```

**`configs/config_dev.yaml`**:
```yaml
database:
  host: localhost
  port: 5433  # Overrides base config
```

**`.env.dev.local`**:
```env
APP_NAME=myapp-dev
DATABASE_PASSWORD=secret123
DATABASE_HOST=dev-db.example.com
```

## Environment Variable Replacement

The config package supports environment variable expansion using `${VAR}` syntax in configuration values. This happens after all config files are loaded.

### Example

**`config.yaml`**:
```yaml
app:
  data_dir: ${HOME}/data
  config_path: ${CONFIG_ROOT}/settings

database:
  url: postgresql://${DB_USER}:${DB_PASS}@${DB_HOST}:5432/mydb
```

**Environment Variables**:
```bash
export HOME=/home/user
export CONFIG_ROOT=/etc/myapp
export DB_USER=admin
export DB_PASS=secret
export DB_HOST=db.example.com
```

**Result**:
- `app.data_dir` → `/home/user/data`
- `app.config_path` → `/etc/myapp/settings`
- `database.url` → `postgresql://admin:secret@db.example.com:5432/mydb`

## Key Normalization

All environment variable keys are normalized for consistency:

- **Underscores → Dots**: `APP_NAME` → `app.name`
- **Uppercase → Lowercase**: `DATABASE_HOST` → `database.host`
- **Nested Keys**: `APP_DATABASE_PORT` → `app.database.port`

This normalization applies to:
- `.env` file keys
- Environment variables loaded via `env.Provider`
- Ensures consistent key naming across all sources

### Example

**`.env` file**:
```env
APP_NAME=myapp
DATABASE_HOST=localhost
APP_DATABASE_PORT=5432
```

**Resulting keys**:
- `app.name` = `"myapp"`
- `database.host` = `"localhost"`
- `app.database.port` = `"5432"`

## Priority and Override Order

Configuration values are loaded in the following order (later sources override earlier):

1. **YAML Files** (in pattern order):
   - `config.yaml`
   - `config_{env}.yaml`
   - `config_{env}_local.yaml`

2. **Environment Files** (in pattern order):
   - `.env`
   - `.env.{env}`
   - `.env.local`
   - `.env.{env}.local`

3. **Environment Variables** (always last, highest priority):
   - `APP_NAME`, `DATABASE_HOST`, etc.

### Example Override Chain

Given:
- `config.yaml`: `app.name = "base"`
- `config_dev.yaml`: `app.name = "dev"`
- `.env.dev.local`: `APP_NAME=local`
- Environment: `APP_NAME=env-override`

**Result**: `app.name = "env-override"` (environment variables win)

## Custom Configuration Sources

You can define custom sources with full control over file patterns, paths, and parsers:

```go
factory := config.NewDefaultConfigFactory()

cfg, err := factory.Create(
    config.ConfigSource{
        Path:     "./config",
        Patterns: []string{"app.yaml", "app_{env}.yaml", "app_{env}_local.yaml"},
        Parser:   yaml.Parser(),
    },
    config.ConfigSource{
        Path:     ".",
        Patterns: []string{".env", ".env.{env}", ".env.local", ".env.{env}.local"},
        Parser:   config.NewDotenvParser(),
    },
)
```

### Custom Parsers

Use any `koanf.Parser` implementation:

```go
import (
    "github.com/knadh/koanf/parsers/json"
    "github.com/knadh/koanf/parsers/toml"
)

// JSON parser
config.ConfigSource{
    Path:     "./config",
    Patterns: []string{"config.json", "config_{env}.json"},
    Parser:   json.Parser(),
}

// TOML parser
config.ConfigSource{
    Path:     "./config",
    Patterns: []string{"config.toml", "config_{env}.toml"},
    Parser:   toml.Parser(),
}
```

### Multiple Source Directories

Load from multiple independent directories:

```go
cfg, err := factory.Create(
    config.ConfigSource{
        Path:     "./config/app",
        Patterns: []string{"app.yaml", "app_{env}.yaml"},
        Parser:   yaml.Parser(),
    },
    config.ConfigSource{
        Path:     "./config/db",
        Patterns: []string{"db.yaml", "db_{env}.yaml"},
        Parser:   yaml.Parser(),
    },
)
```

## Struct Unmarshaling

Unmarshal configuration into type-safe structs using `config` struct tags:

```go
type DatabaseConfig struct {
    Host     string `config:"host"`
    Port     int    `config:"port"`
    User     string `config:"user"`
    Password string `config:"password"`
    Name     string `config:"name"`
}

var dbCfg DatabaseConfig
err := cfg.UnmarshalKey("database", &dbCfg)
if err != nil {
    panic(err)
}

fmt.Println(dbCfg.Host)     // e.g., "localhost"
fmt.Println(dbCfg.Port)      // e.g., 5432
```

### Nested Structures

```go
type AppConfig struct {
    Name    string `config:"name"`
    Version string `config:"version"`
    Database struct {
        Host string `config:"host"`
        Port int    `config:"port"`
    } `config:"database"`
}

var appCfg AppConfig
cfg.UnmarshalKey("app", &appCfg)
```

**Note**: The package uses `config` struct tags (not `koanf`) by default for better abstraction.

## Environment Detection

The environment is determined by the `APP_ENV` environment variable:

- `APP_ENV=dev` → loads `config_dev.yaml`, `.env.dev`, etc.
- `APP_ENV=prod` → loads `config_prod.yaml`, `.env.prod`, etc.
- `APP_ENV=test` → loads `config_test.yaml`, `.env.test`, etc.
- `APP_ENV=""` → defaults to `dev`

The `{env}` placeholder in file patterns is replaced with the current environment value.

## Direct Value Access

Access configuration values directly without unmarshaling:

```go
// String values
appName := cfg.k.String("app.name")
dbHost := cfg.k.String("database.host")

// Integer values
dbPort := cfg.k.Int64("database.port")

// Boolean values
debug := cfg.k.Bool("app.debug")

// Float values
timeout := cfg.k.Float64("app.timeout")
```

## Error Handling

Missing files are handled gracefully - they are skipped without error. Only parsing errors cause failures:

```go
cfg, err := factory.Create()
if err != nil {
    // Only fails on parsing errors, not missing files
    panic(err)
}

// Missing files result in empty/default values
value := cfg.k.String("nonexistent.key") // Returns ""
```

## Best Practices

- **Use Environment Variables for Secrets**: Never commit secrets to YAML files. Use `.env.local` (gitignored) or environment variables.

- **Use Struct Unmarshaling**: Prefer `UnmarshalKey` over direct access for type safety.

- **Environment-Specific Overrides**: Use `{env}` patterns for environment-specific values while keeping base configs shared.

- **Local Development**: Use `.env.local` or `.env.{env}.local` files for local overrides (gitignored).

## Examples

See `factory_test.go` for comprehensive examples covering:
- YAML-only configurations
- Environment file configurations
- Mixed YAML and `.env` sources
- Environment variable expansion
- Multiple source directories
- Custom parsers

## API Reference

### Types

- `Config`: Main configuration container
- `ConfigFactory`: Interface for creating configurations
- `DefaultConfigFactory`: Default factory implementation
- `ConfigSource`: Defines a configuration source (path, patterns, parser)

### Functions

- `NewDefaultConfigFactory()`: Creates a new default factory
- `NewDotenvParser()`: Creates a dotenv parser with key normalization
- `EnvTransformFunc()`: Returns transform function for environment variables

### Methods

- `Config.UnmarshalKey(key string, dest any) error`: Unmarshals config into struct
- `ConfigFactory.Create(sources ...ConfigSource) (*Config, error)`: Creates configuration

