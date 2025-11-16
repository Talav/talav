# Fx Config Module

> [Fx](https://uber-go.github.io/fx/) module for [config](../component/config).

## Overview

The `fxconfig` module provides seamless integration between the Talav configuration system and [Uber's Fx dependency injection framework](https://github.com/uber-go/fx). It handles configuration loading during application startup and makes configuration available throughout your application via dependency injection.

## Features

- **Automatic Configuration Loading**: Loads configuration during Fx application startup
- **Dependency Injection**: Makes `*config.Config` available for injection into any component
- **Custom Sources**: Register additional configuration sources using `AsConfigSource()`
- **Typed Configuration**: Extract typed configuration structs using `AsConfig[T]()`
- **Error Handling**: Fails fast on configuration errors during application construction
- **Factory Override**: Support for custom configuration factories via `fx.Decorate()`

## Installation

```bash
go get github.com/talav/talav/pkg/fx/fxconfig
```

## Quick Start

### Basic Usage

```go
package main

import (
	"fmt"

	"github.com/talav/talav/pkg/component/config"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fxconfig.FxConfigModule,                    // Load the config module
		fx.Invoke(func(cfg *config.Config) {        // Inject config into any component
			var appCfg struct {
				Name    string `config:"name"`
				Version string `config:"version"`
			}
			cfg.UnmarshalKey("app", &appCfg)
			fmt.Printf("App: %s v%s\n", appCfg.Name, appCfg.Version)
		}),
	).Run()
}
```

## Configuration Files

By default, the module loads configuration from:

**YAML Files** (from `./configs/`):
- `config.yaml` - Base configuration
- `config_{env}.yaml` - Environment-specific (e.g., `config_dev.yaml`)
- `config_{env}_local.yaml` - Local overrides (e.g., `config_dev_local.yaml`)

**Environment Files** (from `.`):
- `.env` - Base environment variables
- `.env.{env}` - Environment-specific (e.g., `.env.dev`)
- `.env.local` - Local overrides
- `.env.{env}.local` - Environment-specific local overrides

**Environment Variables** (always loaded last, highest priority):
- `APP_NAME` → `app.name`
- `DATABASE_HOST` → `database.host`

The environment is determined by the `APP_ENV` environment variable (defaults to `dev`).

For detailed configuration file documentation, see the [config package README](../component/config/README.md).

## Custom Configuration Sources

### Adding Custom Sources

Use `AsConfigSource()` to register additional configuration sources:

```go
package main

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/talav/talav/pkg/component/config"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fxconfig.FxConfigModule,
		// Add custom config source
		fxconfig.AsConfigSource(config.ConfigSource{
			Path:     "./custom-configs",
			Patterns: []string{"app.yaml", "app_{env}.yaml"},
			Parser:   yaml.Parser(),
		}),
		fx.Invoke(func(cfg *config.Config) {
			// Config loaded from both default and custom sources
		}),
	).Run()
}
```

### Multiple Custom Sources

Register multiple sources to load configuration from different locations:

```go
fx.New(
	fxconfig.FxConfigModule,
	fxconfig.AsConfigSource(config.ConfigSource{
		Path:     "./config/app",
		Patterns: []string{"app.yaml", "app_{env}.yaml"},
		Parser:   yaml.Parser(),
	}),
	fxconfig.AsConfigSource(config.ConfigSource{
		Path:     "./config/db",
		Patterns: []string{"database.yaml", "database_{env}.yaml"},
		Parser:   yaml.Parser(),
	}),
	// ...
).Run()
```

**Note**: When custom sources are provided, they replace the default sources entirely.

## Typed Configuration Extraction

### Using AsConfig[T]()

Extract typed configuration structs and make them available for injection:

```go
package main

import (
	"fmt"

	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

type DatabaseConfig struct {
	Host     string `config:"host"`
	Port     int    `config:"port"`
	User     string `config:"user"`
	Password string `config:"password"`
	Name     string `config:"name"`
}

type AppConfig struct {
	Name    string `config:"name"`
	Version string `config:"version"`
	Debug   bool   `config:"debug"`
}

func main() {
	fx.New(
		fxconfig.FxConfigModule,
		// Register typed configs
		fxconfig.AsConfig("database", DatabaseConfig{}),
		fxconfig.AsConfig("app", AppConfig{}),
		// Inject typed configs
		fx.Invoke(func(dbCfg DatabaseConfig, appCfg AppConfig) {
			fmt.Printf("Connecting to %s@%s:%d\n", dbCfg.User, dbCfg.Host, dbCfg.Port)
			fmt.Printf("App: %s v%s (debug: %v)\n", appCfg.Name, appCfg.Version, appCfg.Debug)
		}),
	).Run()
}
```

### Module-Specific Configuration

Modules can declare their own typed configuration:

```go
package logger

import (
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

type LoggerConfig struct {
	Level      string `config:"level"`
	Format     string `config:"format"`
	OutputPath string `config:"output_path"`
}

var Module = fx.Module(
	"logger",
	fxconfig.AsConfig("logger", LoggerConfig{}),
	fx.Provide(NewLogger),
)

func NewLogger(cfg LoggerConfig) *Logger {
	// Create logger with config
	return &Logger{
		level:  cfg.Level,
		format: cfg.Format,
		output: cfg.OutputPath,
	}
}
```

Then in your configuration file:

```yaml
# configs/config.yaml
logger:
  level: info
  format: json
  output_path: /var/log/app.log
```

## Custom Configuration Factory

Override the default factory with a custom implementation:

```go
package main

import (
	"github.com/talav/talav/pkg/component/config"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

type CustomConfigFactory struct{}

func NewCustomConfigFactory() config.ConfigFactory {
	return &CustomConfigFactory{}
}

func (f *CustomConfigFactory) Create(sources ...config.ConfigSource) (*config.Config, error) {
	// Custom configuration loading logic
	return &config.Config{}, nil
}

func main() {
	fx.New(
		fxconfig.FxConfigModule,
		fx.Decorate(NewCustomConfigFactory),        // Override the factory
		fx.Invoke(func(cfg *config.Config) {
			// Uses custom factory
		}),
	).Run()
}
```

## Error Handling

Configuration errors cause the application to fail during construction:

```go
fx.New(
	fxconfig.FxConfigModule,
	fxconfig.AsConfigSource(config.ConfigSource{
		Path:     "./configs",
		Patterns: []string{"invalid.yaml"},  // Invalid YAML
		Parser:   yaml.Parser(),
	}),
).Run()
// Application will fail to start with configuration error
```

This fail-fast behavior ensures configuration issues are caught early rather than causing runtime errors.

## Module Structure

### Provided Dependencies

The module provides the following for injection:

- `*config.Config` - Main configuration instance
- `config.ConfigFactory` - Configuration factory (can be decorated)

### Required Dependencies

The module automatically collects:

- `[]config.ConfigSource` (group: `"config-sources"`) - Custom configuration sources

### Module Name

The module is registered with the name `"config"`.

## API Reference

### FxConfigModule

```go
var FxConfigModule = fx.Module(
	"config",
	fx.Provide(
		config.NewDefaultConfigFactory,
		NewFxConfig,
	),
)
```

The main Fx module. Include this in your `fx.New()` call.

### AsConfigSource

```go
func AsConfigSource(source config.ConfigSource) fx.Option
```

Registers an additional configuration source. Multiple sources can be registered.

**Parameters:**
- `source` - Configuration source with path, patterns, and parser

**Example:**
```go
fxconfig.AsConfigSource(config.ConfigSource{
	Path:     "./configs",
	Patterns: []string{"app.yaml", "app_{env}.yaml"},
	Parser:   yaml.Parser(),
})
```

### AsConfig[T]

```go
func AsConfig[T any](key string, _ T) fx.Option
```

Registers a typed configuration provider that extracts configuration at the given key.

**Parameters:**
- `key` - Configuration key path (e.g., `"database"`, `"app.logger"`)
- `_` - Zero value of type T (used for type inference)

**Returns:** Fx option that provides type T for injection

**Example:**
```go
type DatabaseConfig struct {
	Host string `config:"host"`
	Port int    `config:"port"`
}

fxconfig.AsConfig("database", DatabaseConfig{})
```

## Best Practices

### 1. Use Typed Configuration

Prefer `AsConfig[T]()` over direct `config.Config` injection for better type safety:

```go
// Good
fxconfig.AsConfig("database", DatabaseConfig{})
fx.Invoke(func(cfg DatabaseConfig) { ... })

// Less ideal
fx.Invoke(func(cfg *config.Config) {
	var dbCfg DatabaseConfig
	cfg.UnmarshalKey("database", &dbCfg)
	// ...
})
```

### 2. Module-Level Configuration

Each module should declare its configuration requirements:

```go
var MyModule = fx.Module(
	"mymodule",
	fxconfig.AsConfig("mymodule", MyModuleConfig{}),
	fx.Provide(NewMyService),
)
```

### 3. Environment-Specific Configuration

Use the `{env}` placeholder for environment-specific files:

```go
fxconfig.AsConfigSource(config.ConfigSource{
	Path:     "./configs",
	Patterns: []string{
		"base.yaml",
		"base_{env}.yaml",        // e.g., base_dev.yaml
		"base_{env}_local.yaml",  // e.g., base_dev_local.yaml
	},
	Parser: yaml.Parser(),
})
```

### 4. Fail Fast

Let configuration errors fail during application construction rather than catching and handling them later.

### 5. Document Configuration

Document required configuration keys in your module's README:

```markdown
## Configuration

This module requires the following configuration:

\`\`\`yaml
mymodule:
  enabled: true
  timeout: 30s
  retry_count: 3
\`\`\`
```

## Testing

When testing components that depend on configuration:

```go
func TestMyComponent(t *testing.T) {
	fxtest.New(
		t,
		fx.NopLogger,
		fxconfig.FxConfigModule,
		fxconfig.AsConfigSource(config.ConfigSource{
			Path:     "./testdata",
			Patterns: []string{"test.yaml"},
			Parser:   yaml.Parser(),
		}),
		fx.Invoke(func(cfg *config.Config) {
			// Test with injected config
		}),
	).RequireStart().RequireStop()
}
```

## Dependencies

- [go.uber.org/fx](https://github.com/uber-go/fx) v1.24.0+
- [github.com/talav/talav/pkg/component/config](../component/config)

## See Also

- [Config Package Documentation](../component/config/README.md) - Detailed configuration system documentation
- [Fx Documentation](https://uber-go.github.io/fx/) - Uber's Fx dependency injection framework
- [Koanf Documentation](https://github.com/knadh/koanf) - Underlying configuration library

