# HTTP Server Component

> Opinionated HTTP server component based on Chi router with integrated logging and Zorya API framework

## Overview

The `httpserver` component provides a complete, production-ready HTTP server solution with:

- **Chi Router** - Fast, lightweight HTTP router
- **Request ID Middleware** - Automatic request ID generation for tracing
- **Structured Logging** - HTTP request logging via `go-chi/httplog`
- **Zorya Integration** - Type-safe API framework with OpenAPI generation
- **Configuration-Driven** - YAML-based configuration
- **Fx Module Support** - Ready-to-use Fx dependency injection module

## Installation

```bash
go get github.com/talav/talav/pkg/component/httpserver
```

## Quick Start

### Using Fx Module (Recommended)

```go
package main

import (
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxhttpserver"
	"github.com/talav/talav/pkg/fx/fxlogger"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fxconfig.FxConfigModule,
		fxlogger.FxLoggerModule,
		fxhttpserver.FxHTTPServerModule,
		fx.Invoke(RegisterRoutes),
	).Run()
}

func RegisterRoutes(api zorya.API) {
	// Register your routes here
}
```

### Manual Usage

```go
package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/talav/talav/pkg/component/httpserver"
	"github.com/talav/talav/pkg/component/logger"
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/component/zorya/adapters"
)

func main() {
	// Create logger
	logFactory := logger.NewDefaultLoggerFactory()
	log, _ := logFactory.Create(logger.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})

	// Create server config
	cfg := httpserver.Config{
		Server: httpserver.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Logging: httpserver.LoggingConfig{
			Enabled:      true,
			Level:        "info",
			Schema:       "standard",
			RecoverPanics: true,
		},
		OpenAPI: httpserver.OpenAPIConfig{
			Title:   "My API",
			Version: "1.0.0",
		},
	}

	// Create router and Zorya API
	router := chi.NewRouter()
	adapter := adapters.NewChi(router)
	api := zorya.NewAPI(
		adapter,
		zorya.WithOpenAPI(cfg.ToZoryaOpenAPI()),
		zorya.WithConfig(cfg.ToZoryaConfig()),
	)

	// Register routes with Zorya
	zorya.Register(api, zorya.BaseRoute{
		Method: http.MethodGet,
		Path:   "/health",
		Operation: &zorya.Operation{
			Summary: "Health check",
		},
	}, func(ctx context.Context, input *struct{}) (*struct {
		Status int `json:"-"`
		Body   struct {
			Status string `json:"status"`
		}
	}, error) {
		return &struct {
			Status int `json:"-"`
			Body   struct {
				Status string `json:"status"`
			}
		}{
			Status: http.StatusOK,
			Body: struct {
				Status string `json:"status"`
			}{
				Status: "ok",
			},
		}, nil
	})

	// Create server
	server, _ := httpserver.NewServer(cfg.Server, api, log)

	// Start server
	ctx := context.Background()
	server.Start(ctx)
}
```

## Configuration

### Configuration Structure

```yaml
# configs/config.yaml
httpserver:
  server:
    host: "localhost"
    port: 8080
    readTimeout: "15s"
    writeTimeout: "15s"
    idleTimeout: "60s"
    readHeaderTimeout: ""
    maxHeaderBytes: 0
    shutdownTimeout: "10s"

  logging:
    enabled: true
    level: info          # debug, info, warn, error
    schema: standard      # standard, ecs, otel, gcp
    recoverPanics: true
    logRequestBody: false
    logResponseBody: false
    skipPaths:
      - "/health"
      - "/metrics"

  api:
    specPath: "/openapi"
    docsPath: "/docs"
    schemasPath: "/schemas"
    defaultFormat: "application/json"
    noFormatFallback: false

  openapi:
    title: "My API"
    description: "API description"
    version: "1.0.0"
    termsOfService: "https://example.com/terms"
    
    contact:
      name: "Support Team"
      email: "support@example.com"
      url: "https://example.com/support"
    
    license:
      name: "Apache 2.0"
      url: "https://www.apache.org/licenses/LICENSE-2.0.html"
      identifier: "Apache-2.0"  # SPDX identifier
    
    tags:
      - name: "users"
        description: "User management"
      - name: "posts"
        description: "Post management"
    
    externalDocs:
      description: "Additional documentation"
      url: "https://example.com/docs"
    
    security:
      - "bearerAuth"
      - "apiKey"
```

### Server Configuration

- **host** (string, default: "localhost") - Server host address
- **port** (int, default: 8080) - Server port number
- **readTimeout** (string, default: "15s") - Maximum duration for reading the entire request
- **writeTimeout** (string, default: "15s") - Maximum duration before timing out writes of the response
- **idleTimeout** (string, default: "60s") - Maximum amount of time to wait for the next request
- **readHeaderTimeout** (string, default: "") - Amount of time allowed to read request headers
- **maxHeaderBytes** (int, default: 0) - Maximum number of bytes the server will read parsing request headers
- **shutdownTimeout** (string, default: "10s") - Maximum duration for graceful shutdown

### Logging Configuration

- **enabled** (bool, default: false) - Enable HTTP request logging
- **level** (string, default: "info") - Log verbosity: `debug`, `info`, `warn`, `error`
- **schema** (string, default: "standard") - Logging schema format:
  - `standard` - Simple key-value format
  - `ecs` - Elastic Common Schema (for Elasticsearch)
  - `otel` - OpenTelemetry format
  - `gcp` - Google Cloud Platform format
- **recoverPanics** (bool, default: false) - Recover from panics and return HTTP 500
- **logRequestBody** (bool, default: false) - Log request bodies (useful for debugging)
- **logResponseBody** (bool, default: false) - Log response bodies (useful for debugging)
- **skipPaths** ([]string) - Paths to exclude from logging (e.g., health checks, metrics)

### API Configuration

The `api` section contains Zorya behavioral configuration:

- **specPath** (string, default: "/openapi") - Path to the OpenAPI spec without extension (serves `/openapi.json` and `/openapi.yaml`)
- **docsPath** (string, default: "/docs") - Path to the API documentation UI
- **schemasPath** (string, default: "/schemas") - Path to the API schemas
- **defaultFormat** (string, default: "application/json") - Default content type
- **noFormatFallback** (bool, default: false) - Disable fallback to application/json

### OpenAPI Configuration

The `openapi` section contains metadata for the OpenAPI specification:

- **title** (string, required) - Title of the API
- **description** (string) - Description of the API
- **version** (string, required) - Version of the API
- **termsOfService** (string) - Terms of service URL
- **contact** (object) - Contact information:
  - **name** (string) - Contact name
  - **email** (string) - Contact email
  - **url** (string) - Contact URL
- **license** (object) - License information:
  - **name** (string) - License name
  - **url** (string) - License URL
  - **identifier** (string) - SPDX license identifier
- **tags** (array) - Tags for grouping operations:
  - **name** (string) - Tag name
  - **description** (string) - Tag description
- **externalDocs** (object) - External documentation reference:
  - **description** (string) - Description
  - **url** (string) - Documentation URL
- **security** (array of strings) - Security scheme names

## Features

### Request ID

Every request automatically receives a unique request ID via Chi's `RequestID` middleware:

- Generated automatically for each request
- Included in HTTP response headers as `X-Request-Id`
- Logged with all request/response logs for distributed tracing
- Always enabled (cannot be disabled)

### Structured Logging

HTTP requests are logged using `go-chi/httplog` with structured logging:

- Request method, path, and status code
- Response time in milliseconds
- Request ID for correlation
- Configurable log levels
- Support for multiple logging schemas (ECS, OTEL, GCP)

Example log output:

```json
{
  "time": "2024-01-01T12:00:00Z",
  "level": "INFO",
  "msg": "request completed",
  "method": "POST",
  "path": "/users",
  "status": 201,
  "latency_ms": 45,
  "request_id": "abc123"
}
```

### Zorya Integration

The server integrates seamlessly with Zorya for:

- Type-safe request/response handling
- Automatic request validation
- Content negotiation (JSON, CBOR)
- OpenAPI schema generation
- RFC 9457 compliant error responses

See the [Zorya documentation](../zorya/README.md) for more details.

## Middleware Stack

The server applies middleware in the following order:

1. **RequestID** - Adds unique request ID (always enabled)
2. **HTTPLog** - Structured request/response logging (if enabled)
3. **Application Middleware** - Your custom middleware via router
4. **Zorya** - Request validation, content negotiation, error handling

## Examples

See the [example server](../../fx/fxhttpserver/examples/api-server/) for a complete working example with:

- Fx dependency injection
- Multiple API endpoints
- Request validation
- OpenAPI documentation
- Health checks

## API Reference

### Server

```go
type Server struct {
    // ... internal fields ...
}

// NewServer creates a new HTTP server that wraps the provided Zorya API.
// The router and middleware should be configured in the Zorya API before passing it here.
func NewServer(cfg ServerConfig, api zorya.API, logger *slog.Logger) (*Server, error)

// Start starts the HTTP server and blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) error
```

### Config

```go
type Config struct {
    Server  ServerConfig
    Logging LoggingConfig
    API     APIConfig
    OpenAPI OpenAPIConfig
}

// ToZoryaOpenAPI converts the OpenAPI config to a Zorya OpenAPI spec
func (c *Config) ToZoryaOpenAPI() *zorya.OpenAPI

// ToZoryaConfig converts the API config to a Zorya Config
func (c *Config) ToZoryaConfig() *zorya.Config
```

## Dependencies

- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/go-chi/httplog/v3` - HTTP request logging
- `github.com/talav/talav/pkg/component/zorya` - API framework

## See Also

- [Zorya API Framework](../zorya/README.md) - Type-safe API framework
- [Logger Component](../logger/README.md) - Structured logging
- [Fx HTTPServer Module](../../fx/fxhttpserver/) - Fx integration module
