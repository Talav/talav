# Talav Framework

A lightweight Go application framework built on top of Uber FX, providing a clean entry point for building CLI applications with lazy dependency injection.

## Features

- **Lazy Initialization**: Commands like `--help` and `version` work instantly without booting dependencies
- **Automatic FX Cleanup**: FX container is properly shut down after commands complete
- **Component Autonomy**: Components manage their own lifecycle via self-sufficient commands
- **Testing Utilities**: Built-in helpers for testing commands
- **Environment Detection**: Automatic environment detection via `APP_ENV`

## Quick Start

### Simple Application

```go
package main

import (
    "log"
    
    "github.com/talav/talav/pkg/module/framework"
    "github.com/talav/talav/pkg/fx/fxconfig"
    "github.com/talav/talav/pkg/fx/fxhttpserver"
    "github.com/talav/talav/pkg/fx/fxlogger"
)

func main() {
    if err := framework.RunDefault(
        fxconfig.FxConfigModule,
        fxlogger.FxLoggerModule,
        fxhttpserver.FxHTTPServerModule,
    ); err != nil {
        log.Fatal(err)
    }
}
```

### Run Commands

```bash
# Start HTTP server
go run main.go serve-http

# Show version
go run main.go version

# Show help
go run main.go --help
```

## Architecture

### Core Components

#### Application

The `Application` is the main entry point that:
- Creates the root Cobra command
- Manages lazy FX kernel initialization
- Automatically cleans up FX container on exit
- Provides bootstrap helpers

```go
app := framework.NewApplication(
    framework.WithName("myapp"),
    framework.WithVersion("1.0.0"),
    framework.WithEnvironment("production"),
    framework.WithModules(
        fxlogger.FxLoggerModule,
        fxhttpserver.FxHTTPServerModule,
    ),
)

if err := app.Execute(); err != nil {
    log.Fatal(err)
}
```

#### FX Container

The `Application` directly manages the FX dependency injection container:
- Singleton FX app instance
- Eager initialization during application construction
- Automatic cleanup on shutdown
- Commands collected via `fx.Populate` during initialization

## Configuration

### Application Options

```go
app := framework.NewApplication(
    framework.WithName("myapp"),           // Application name
    framework.WithVersion("1.0.0"),        // Application version
    framework.WithEnvironment("prod"),     // Environment (dev, prod, test)
    framework.WithModules(                 // FX modules to load
        fxlogger.FxLoggerModule,
        fxhttpserver.FxHTTPServerModule,
    ),
)
```

### Environment Detection

The framework automatically detects the environment from `APP_ENV`:

```bash
# Development (default)
go run main.go serve-http

# Production
APP_ENV=prod go run main.go serve-http

# Test
APP_ENV=test go run main.go serve-http
```

## Bootstrap Helpers

### RunDefault

Simplest bootstrap - just provide modules:

```go
framework.RunDefault(
    fxlogger.FxLoggerModule,
    fxhttpserver.FxHTTPServerModule,
)
```

### Run

Bootstrap with explicit configuration:

```go
framework.Run(
    framework.WithName("myapp"),
    framework.WithVersion("1.0.0"),
    framework.WithModules(
        fxlogger.FxLoggerModule,
        fxhttpserver.FxHTTPServerModule,
    ),
)
```

### RunWithContext

Bootstrap with custom context:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

framework.RunWithContext(ctx,
    framework.WithEnvironment("production"),
    framework.WithModules(/* ... */),
)
```

## Component Commands

Components define their own commands with no framework dependencies. The framework provides signal handling automatically via context:

```go
// pkg/component/httpserver/cmd/serve.go
package cmd

import (
    "log/slog"

    "github.com/spf13/cobra"
    "github.com/talav/talav/pkg/component/httpserver"
)

func NewServeHTTPCommand(
    server *httpserver.Server,
    logger *slog.Logger,
) *cobra.Command {
    return &cobra.Command{
        Use:   "serve-http",
        Short: "Start the HTTP server",
        RunE: func(cmd *cobra.Command, args []string) error {
            logger.Info("starting HTTP server...")

            // Use context from framework (already has signal handling)
            if err := server.Start(cmd.Context()); err != nil {
                return err
            }

            logger.Info("HTTP server shutdown complete")
            return nil
        },
    }
}
```

Commands are registered via FX modules:

```go
// pkg/fx/fxhttpserver/module.go
var FxHTTPServerModule = fx.Module(
    "httpserver",
    fx.Provide(httpserver.NewServer),
    fxcore.AsRootCommand(cmd.NewServeHTTPCommand),
)
```

## Testing

### Testing Commands

```go
func TestServeHTTPCommand(t *testing.T) {
    app := framework.NewApplication(
        framework.WithModules(
            fxlogger.FxLoggerModule,
            fxhttpserver.FxHTTPServerModule,
        ),
    )

    // Commands are available after kernel boots
    kernel, err := app.GetKernel(context.Background())
    require.NoError(t, err)

    // Test command execution
    // ...
}
```

## Design Principles

### Component Autonomy

Components are self-sufficient and manage their own lifecycle:
- ✅ Components define their own commands
- ✅ Commands handle their own signal management
- ✅ No framework dependencies in components
- ✅ FX just wires dependencies

### Layered Architecture

```
┌─────────────────────────────────────┐
│ Application Layer                   │  ← Can import everything
│ (main.go, framework module)         │
└─────────────────────────────────────┘
              ↑ (imports)
┌─────────────────────────────────────┐
│ FX Layer                            │  ← Can import components only
│ (pkg/fx/*)                          │  ← NO framework imports
└─────────────────────────────────────┘
              ↑ (imports)
┌─────────────────────────────────────┐
│ Component Layer                     │  ← Can import other components only
│ (pkg/component/*)                   │  ← NO framework, NO FX imports
└─────────────────────────────────────┘
```

### Automatic Cleanup

The framework ensures FX container cleanup:
- FX starts when first command runs (lazy)
- FX stops when command completes (automatic)
- No manual cleanup needed
- No resource leaks

## Examples

See the `examples/` directory for complete working examples:
- `simple-http-app/` - Basic HTTP server
- `configured-http-app/` - HTTP server with configuration

## License

MIT
