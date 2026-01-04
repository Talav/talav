# API Server Example

This example demonstrates how to use the `fxhttpserver` module to create a complete HTTP API server with:

- Chi router with request ID middleware
- Structured HTTP request logging via `go-chi/httplog`
- Zorya API framework for type-safe endpoints
- OpenAPI specification generation
- Fx dependency injection

## Features

- **Health Check Endpoint** (`GET /health`) - System health status
- **Create User** (`POST /users`) - Create a new user with validation
- **Get User** (`GET /users/{id}`) - Retrieve a user by ID
- **OpenAPI Documentation** (`/docs`) - Interactive API documentation
- **OpenAPI Spec** (`/openapi.json`) - Machine-readable API specification

## Running the Example

1. **Install dependencies:**

```bash
go mod tidy
```

2. **Run the server:**

```bash
go run main.go
```

3. **Test the endpoints:**

```bash
# Health check
curl http://localhost:8080/health

# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Get a user
curl http://localhost:8080/users/user-123

# View OpenAPI spec
curl http://localhost:8080/openapi.json

# View API documentation
open http://localhost:8080/docs
```

## Configuration

The server configuration is defined in `config.yaml`:

```yaml
httpserver:
  server:
    host: "localhost"
    port: 8080

  logging:
    enabled: true
    level: info
    schema: standard
    recoverPanics: true
    skipPaths:
      - "/health"

  api:
    title: "API Server Example"
    description: "A comprehensive example API server"
    version: "1.0.0"
```

## Request ID

All requests automatically receive a unique request ID via Chi's `RequestID` middleware. The request ID is:

- Generated automatically for each request
- Included in HTTP response headers as `X-Request-Id`
- Logged with all request/response logs for tracing

## Structured Logging

HTTP requests are logged using `go-chi/httplog` with structured logging:

- Request method, path, and status code
- Response time
- Request ID for correlation
- Configurable log levels (debug, info, warn, error)
- Support for ECS, OTEL, and GCP logging schemas

## Zorya Integration

This example uses Zorya for:

- Type-safe request/response handling
- Automatic request validation
- Content negotiation (JSON, CBOR)
- OpenAPI schema generation
- RFC 9457 compliant error responses

## Architecture

```
┌─────────────────┐
│   Fx App        │
│                 │
│  ┌───────────┐  │
│  │  Config   │  │
│  └───────────┘  │
│       │         │
│  ┌───────────┐  │
│  │  Logger   │  │
│  └───────────┘  │
│       │         │
│  ┌───────────┐  │
│  │HTTPServer │  │
│  │  (Chi)    │  │
│  └───────────┘  │
│       │         │
│  ┌───────────┐  │
│  │  Zorya    │  │
│  │    API    │  │
│  └───────────┘  │
└─────────────────┘
```

## Middleware Stack

1. **RequestID** - Adds unique request ID to each request
2. **HTTPLog** - Structured request/response logging
3. **Zorya** - Request validation, content negotiation, error handling

## Next Steps

- Add database integration
- Implement authentication/authorization
- Add more endpoints
- Configure custom middleware
- Set up production logging (ECS, OTEL, etc.)























