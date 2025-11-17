# Fx Logger Module

> [Fx](https://uber-go.github.io/fx/) module for [logger](../component/logger).

## Overview

The `fxlogger` module provides seamless integration between the Talav logger system and [Uber's Fx dependency injection framework](https://github.com/uber-go/fx). It handles logger creation during application startup and makes a configured `*slog.Logger` available throughout your application via dependency injection.

## Features

- **Automatic Logger Creation**: Creates logger during Fx application startup
- **Dependency Injection**: Makes `*slog.Logger` available for injection into any component
- **Configuration-Driven**: Logger configuration loaded from application config
- **Standard Library Based**: Built on Go's native `log/slog` package
- **Factory Override**: Support for custom logger factories via `fx.Decorate()`

## Installation

```bash
go get github.com/talav/talav/pkg/fx/fxlogger
```

## Quick Start

### Basic Usage

```go
package main

import (
	"log/slog"

	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxlogger"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fxconfig.FxConfigModule,              // Required: Load config first
		fxlogger.FxLoggerModule,              // Load the logger module
		fx.Invoke(func(log *slog.Logger) {    // Inject logger into any component
			log.Info("application started", "version", "1.0.0")
		}),
	).Run()
}
```

## Configuration

The logger module requires configuration to be available via `fxconfig.FxConfigModule`. The logger configuration is extracted from the `logger` key in your configuration.

### Configuration Structure

```yaml
# configs/config.yaml
logger:
  level: info      # debug, info, warn, error
  format: json     # json, text
  output: stdout   # stdout, or file path (e.g., /var/log/app.log)
  no_color: false  # disable color output (optional)
```

### Configuration Options

**Level**:
- `debug` - Most verbose, includes all logs
- `info` - Default, includes info, warn, and error
- `warn` - Includes warnings and errors
- `error` - Only errors

**Format**:
- `json` - Structured JSON output (default)
- `text` - Human-readable text format

**Output**:
- `stdout` - Console output (default)
- `/path/to/file.log` - File path (writes to both file and stdout)

**NoColor** (optional):
- `true` - Disable color output
- `false` - Enable color output (default)

### Environment-Specific Configuration

Use environment-specific configuration files to customize logging per environment:

```yaml
# configs/config_dev.yaml
logger:
  level: debug
  format: text
  output: stdout

# configs/config_prod.yaml
logger:
  level: info
  format: json
  output: /var/log/app.log
```

## Usage Examples

### Injecting Logger into Components

```go
package service

import (
	"log/slog"
)

type UserService struct {
	log *slog.Logger
}

func NewUserService(log *slog.Logger) *UserService {
	return &UserService{log: log}
}

func (s *UserService) CreateUser(name string) error {
	s.log.Info("creating user", "name", name)
	// ... implementation
	return nil
}
```

```go
package main

import (
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxlogger"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fxconfig.FxConfigModule,
		fxlogger.FxLoggerModule,
		fx.Provide(NewUserService),
		fx.Invoke(func(svc *UserService) {
			svc.CreateUser("John Doe")
		}),
	).Run()
}
```

### Using Logger in HTTP Handlers

```go
package handler

import (
	"log/slog"
	"net/http"
)

type APIHandler struct {
	log *slog.Logger
}

func NewAPIHandler(log *slog.Logger) *APIHandler {
	return &APIHandler{log: log}
}

func (h *APIHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	h.log.Info("request received",
		"method", r.Method,
		"path", r.URL.Path,
		"ip", r.RemoteAddr,
	)
	// ... handle request
}
```

### Structured Logging

The logger supports structured logging with key-value pairs:

```go
func (s *Service) ProcessOrder(orderID string, amount float64) {
	s.log.Info("order processed",
		"order_id", orderID,
		"amount", amount,
		"status", "completed",
	)
}
```

### Error Logging

```go
func (s *Service) HandleError(err error) {
	s.log.Error("operation failed",
		"error", err,
		"operation", "processPayment",
	)
}
```

## Custom Logger Factory

Override the default factory with a custom implementation:

```go
package main

import (
	"log/slog"

	"github.com/talav/talav/pkg/component/logger"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxlogger"
	"go.uber.org/fx"
)

type CustomLoggerFactory struct{}

func NewCustomLoggerFactory() logger.LoggerFactory {
	return &CustomLoggerFactory{}
}

func (f *CustomLoggerFactory) Create(cfg logger.LoggerConfig) (*slog.Logger, error) {
	// Custom logger creation logic
	return slog.Default(), nil
}

func main() {
	fx.New(
		fxconfig.FxConfigModule,
		fxlogger.FxLoggerModule,
		fx.Decorate(NewCustomLoggerFactory),  // Override the factory
		fx.Invoke(func(log *slog.Logger) {
			// Uses custom factory
		}),
	).Run()
}
```

## Error Handling

Logger creation errors cause the application to fail during construction:

```go
fx.New(
	fxconfig.FxConfigModule,
	fxconfig.AsConfigSource(config.ConfigSource{
		Path:     "./configs",
		Patterns: []string{"invalid.yaml"},  // Invalid YAML
		Parser:   yaml.Parser(),
	}),
	fxlogger.FxLoggerModule,
).Run()
// Application will fail to start with configuration error
```

This fail-fast behavior ensures logger configuration issues are caught early rather than causing runtime errors.

## Module Structure

### Provided Dependencies

The module provides the following for injection:

- `*slog.Logger` - Configured logger instance
- `logger.LoggerFactory` - Logger factory (can be decorated)

### Required Dependencies

The module requires:

- `*config.Config` - Provided by `fxconfig.FxConfigModule`
- `logger.LoggerConfig` - Extracted from config at key `"logger"`

### Module Name

The module is registered with the name `"logger"`.

## API Reference

### FxLoggerModule

```go
var FxLoggerModule = fx.Module(
	"logger",
	fxconfig.AsConfig("logger", logger.LoggerConfig{}),
	fx.Provide(
		logger.NewDefaultLoggerFactory,
		func(cfg logger.LoggerConfig, factory logger.LoggerFactory) (*slog.Logger, error) {
			return factory.Create(cfg)
		},
	),
)
```

The main Fx module. Include this in your `fx.New()` call after `fxconfig.FxConfigModule`.

## Best Practices

### 1. Always Include Config Module First

The logger module depends on configuration, so always include `fxconfig.FxConfigModule` before `fxlogger.FxLoggerModule`:

```go
fx.New(
	fxconfig.FxConfigModule,  // Must come first
	fxlogger.FxLoggerModule,
	// ... other modules
)
```

### 2. Use Structured Logging

Prefer structured logging with key-value pairs over formatted strings:

```go
// Good
log.Info("user created", "user_id", userID, "email", email)

// Less ideal
log.Info(fmt.Sprintf("user created: id=%s, email=%s", userID, email))
```

### 3. Use Appropriate Log Levels

- `Debug` - Detailed information for debugging
- `Info` - General informational messages
- `Warn` - Warning messages for potentially harmful situations
- `Error` - Error messages for failures

### 4. Include Context in Logs

Always include relevant context in log messages:

```go
log.Info("request processed",
	"method", r.Method,
	"path", r.URL.Path,
	"status_code", statusCode,
	"duration_ms", duration,
)
```

### 5. Environment-Specific Configuration

Use different log levels and formats per environment:

```yaml
# Development: verbose, human-readable
logger:
  level: debug
  format: text
  output: stdout

# Production: concise, structured
logger:
  level: info
  format: json
  output: /var/log/app.log
```

## Testing

When testing components that depend on the logger:

```go
func TestMyComponent(t *testing.T) {
	var log *slog.Logger

	fxtest.New(
		t,
		fx.NopLogger,
		fxconfig.FxConfigModule,
		fxconfig.AsConfigSource(config.ConfigSource{
			Path:     "./testdata",
			Patterns: []string{"test.yaml"},
			Parser:   yaml.Parser(),
		}),
		fxlogger.FxLoggerModule,
		fx.Populate(&log),
	).RequireStart().RequireStop()

	require.NotNil(t, log)
	log.Info("test message")
}
```

Or use a test logger directly:

```go
func TestMyComponent(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	service := NewMyService(log)
	// ... test service
}
```

## Dependencies

- [go.uber.org/fx](https://github.com/uber-go/fx) v1.24.0+
- [github.com/talav/talav/pkg/component/logger](../component/logger)
- [github.com/talav/talav/pkg/fx/fxconfig](../fxconfig)

## See Also

- [Logger Package Documentation](../component/logger/README.md) - Detailed logger component documentation
- [Fx Documentation](https://uber-go.github.io/fx/) - Uber's Fx dependency injection framework
- [Go slog Package](https://pkg.go.dev/log/slog) - Standard library structured logging






