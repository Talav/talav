# Logger Package

A simple, configurable wrapper around Go's standard library [slog](https://pkg.go.dev/log/slog) package. Provides a factory pattern for creating structured loggers with flexible configuration.

## Features

- **Standard Library Based**: Built on Go's native `log/slog` package
- **Configurable Log Levels**: `debug`, `info`, `warn`, `error`
- **Multiple Formats**: JSON (default) and text output
- **Flexible Output**: Console (stdout) or file-based logging
- **Factory Pattern**: Easy logger creation via `LoggerFactory` interface

## Installation

```bash
go get github.com/talav/talav/pkg/component/logger
```

## Quick Start

```go
package main

import (
	"github.com/talav/talav/pkg/component/logger"
)

func main() {
	factory := logger.NewDefaultLoggerFactory()
	
	log, err := factory.Create(logger.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
	if err != nil {
		panic(err)
	}
	
	log.Info("application started", "version", "1.0.0")
	log.Debug("this won't be logged at info level")
	log.Error("something went wrong", "error", err)
}
```

## Configuration

### LoggerConfig

```go
type LoggerConfig struct {
	Level   string `config:"level"`    // debug, info, warn, error
	Format  string `config:"format"`   // json, text
	Output  string `config:"output"`   // stdout, or file path
	NoColor bool   `config:"no_color"` // disable color output
}
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

## Usage with Config Package

When using with the [config package](../config):

```yaml
# config.yaml
logger:
  level: info
  format: json
  output: /var/log/app.log
```

```go
import (
	"github.com/talav/talav/pkg/component/config"
	"github.com/talav/talav/pkg/component/logger"
)

// Load configuration
cfgFactory := config.NewDefaultConfigFactory()
cfg, _ := cfgFactory.Create()

// Extract logger config
var logCfg logger.LoggerConfig
cfg.UnmarshalKey("logger", &logCfg)

// Create logger
logFactory := logger.NewDefaultLoggerFactory()
log, _ := logFactory.Create(logCfg)

log.Info("configured from file")
```

## Examples

### JSON Format (Default)

```go
log, _ := factory.Create(logger.LoggerConfig{
	Level:  "info",
	Format: "json",
	Output: "stdout",
})

log.Info("user login", "user_id", 123, "ip", "192.168.1.1")
// Output: {"time":"2024-01-01T12:00:00Z","level":"INFO","msg":"user login","user_id":123,"ip":"192.168.1.1"}
```

### Text Format

```go
log, _ := factory.Create(logger.LoggerConfig{
	Level:  "debug",
	Format: "text",
	Output: "stdout",
})

log.Debug("processing request", "path", "/api/users")
// Output: time=2024-01-01T12:00:00.000Z level=DEBUG msg="processing request" path=/api/users
```

### File Output

```go
log, _ := factory.Create(logger.LoggerConfig{
	Level:  "info",
	Format: "json",
	Output: "/var/log/app.log",
})

log.Info("server started", "port", 8080)
// Writes to both /var/log/app.log and stdout
```

## API Reference

### NewDefaultLoggerFactory

```go
func NewDefaultLoggerFactory() LoggerFactory
```

Returns a new `DefaultLoggerFactory` instance.

### LoggerFactory.Create

```go
func Create(cfg LoggerConfig) (*slog.Logger, error)
```

Creates a new `*slog.Logger` instance with the given configuration.

**Parameters:**
- `cfg` - Logger configuration

**Returns:**
- `*slog.Logger` - Configured logger instance
- `error` - Error if logger creation fails

## Notes

- When specifying a file path for `Output`, the logger writes to both the file and stdout
- If file creation fails, the logger automatically falls back to stdout
- Default level is `info` if not specified or invalid
- Default format is `json` if not specified

## Dependencies

- Go 1.21+ (for `log/slog` support)

## See Also

- [Go slog Package](https://pkg.go.dev/log/slog) - Standard library structured logging
- [Config Package](../config/README.md) - Configuration management

