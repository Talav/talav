# System Map

## What is Actually Built & Running

> Living map of the Talav framework architecture, components, and deployment

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│ Application Layer (main.go, pkg/module/*)                      │
│ - Bootstrap application with framework.NewApplication()         │
│ - Compose FX modules                                            │
│ - Define CLI commands                                           │
└─────────────────────────────────────────────────────────────────┘
                            ↓ (imports)
┌─────────────────────────────────────────────────────────────────┐
│ FX Layer (pkg/fx/*)                                             │
│ - Wire components via Uber FX modules                           │
│ - Register commands to application                              │
│ - Provide configuration to components                           │
└─────────────────────────────────────────────────────────────────┘
                            ↓ (imports)
┌─────────────────────────────────────────────────────────────────┐
│ Component Layer (pkg/component/*)                               │
│ - Pure business logic (no framework/FX dependencies)            │
│ - Define domain models, services, repositories                  │
│ - Provide Cobra commands for self-management                    │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### Framework Layer (`pkg/module/framework`)

**Purpose**: Application bootstrapping and CLI management

**Technology**: Go 1.25+, Uber FX, Cobra

**Key Files**:
- `application.go`: Main `Application` struct, manages FX lifecycle and Cobra root command
- `options.go`: Builder pattern options for `Application` configuration
- `bootstrap.go`: Convenience functions (`RunDefault`, `Run`, `RunWithContext`)

**Responsibilities**:
- Create Cobra root command and inject module commands
- Initialize FX container during app construction (fail-fast on DI errors)
- Provide signal handling context to all commands (`SIGINT`, `SIGTERM`)
- Ensure FX cleanup on shutdown

### API Framework (`pkg/component/zorya`)

**Purpose**: Type-safe HTTP API framework with OpenAPI generation

**Technology**: Generic Go handlers, Chi/Fiber adapters, content negotiation

**Entry Points**:
- `zorya.NewAPI(adapter Adapter, opts ...Option)` - Create API instance
- `zorya.Get/Post/Put/Delete[I, O any](api, path, handler, opts)` - Register routes

**Key Features**:
- Type-safe request/response handling via generics
- Automatic request parameter decoding (path, query, header, cookie, body)
- Content negotiation (JSON, CBOR, + custom formats)
- Request validation via go-playground/validator
- RFC 9457 error responses with structured details
- OpenAPI 3.1 schema generation
- Route security declarations (Auth, Roles, Permissions, Resources)
- Streaming responses (SSE, chunked transfer)
- Conditional requests (If-Match, If-None-Match, ETag)

**Architecture**:
- Router adapters: `adapters.NewChi()`, `adapters.NewFiber()`, `adapters.NewStdlib()`
- Schema codec: `schema.Codec` for request/response marshaling
- Validator: Pluggable `Validator` interface
- Middleware: `api.UseMiddleware()`, `route.Middlewares`
- Transformers: `api.UseTransformer()` for response modification

### HTTP Server (`pkg/component/httpserver`)

**Purpose**: Production-ready HTTP server with Chi, logging, and Zorya

**Technology**: Chi v5 router, go-chi/httplog, slog

**Configuration**:
- YAML-based via `pkg/component/config`
- Server settings (host, port, timeouts, shutdown)
- Logging (level, schema, skip paths, panic recovery)
- OpenAPI metadata (title, version, contact, license)

**Middleware Stack** (applied in order):
1. `middleware.RequestID` - Unique request ID (always enabled)
2. `httplog.RequestLogger` - Structured HTTP logging (if enabled)
3. Custom user middleware (via Chi router or Zorya API)
4. Zorya request processing (decode, validate, handle, encode)

**Entry Point**:
- `httpserver.NewServer(cfg, api, logger)` - Create server
- `server.Start(ctx)` - Start server (blocks until context cancelled)

### Configuration (`pkg/component/config`)

**Purpose**: YAML/environment variable configuration with validation

**Technology**: Viper, struct tags

**Key Features**:
- Load from YAML files or environment variables
- Type-safe config structs
- Default value support
- Validation via struct tags
- Multi-environment support (dev, staging, prod)

**Usage**:
```go
type MyConfig struct {
    Host string `mapstructure:"host" default:"localhost"`
    Port int    `mapstructure:"port" default:"8080"`
}

factory := config.NewFactory()
cfg := MyConfig{}
err := factory.Parse("config.yaml", &cfg)
```

### Logger (`pkg/component/logger`)

**Purpose**: Structured logging with slog

**Technology**: Go 1.21+ slog, configurable handlers

**Key Features**:
- JSON/text output formats
- Configurable log levels (debug, info, warn, error)
- Multiple outputs (stdout, file, syslog)
- Contextual logging with request IDs

### Validator (`pkg/component/validator`)

**Purpose**: Request validation with go-playground/validator

**Technology**: go-playground/validator v10

**Key Features**:
- Struct tag-based validation
- Custom validators via registry
- Localization support for error messages
- Nested struct validation
- Cross-field validation

**Validation Tags**: `required`, `email`, `min`, `max`, `oneof`, `pattern`, etc.

### Schema (`pkg/component/schema`)

**Purpose**: HTTP request parameter decoding and response encoding

**Technology**: Reflection, struct tags, mapstructure

**Key Features**:
- Decode path, query, header, cookie, body parameters
- Content type detection (JSON, XML, URL-encoded, multipart)
- Style support (form, simple, label, matrix)
- Explode parameter handling
- Default value application
- Field metadata caching for performance

**Struct Tags**: `schema`, `body`, `status`, `header`, `default`

### ORM (`pkg/component/orm`)

**Purpose**: Database abstraction with GORM

**Technology**: GORM v2, PostgreSQL/MySQL/SQLite drivers

**Key Features**:
- Connection pool management
- Migration runner
- Transaction support
- Soft deletes
- Hooks and callbacks

### Security (`pkg/component/security`)

**Purpose**: Authentication and authorization enforcement

**Technology**: Simple role-based or Casbin policy-based

**Key Features**:
- `SecurityMiddleware` enforces route requirements
- `SecurityEnforcer` interface for custom logic
- Built-in enforcers: `SimpleEnforcer`, `CasbinEnforcer`
- Context-based user injection (`security.GetAuthUser(ctx)`)
- Declarative route security via Zorya's `Secure()` wrapper

**Integration**:
```go
api.UseMiddleware(security.NewSecurityMiddleware(enforcer,
    security.OnUnauthorized(handler),
    security.OnForbidden(handler),
))

zorya.Get(api, "/admin/users", handler,
    zorya.Secure(
        zorya.Roles("admin"),
        zorya.Permissions("users:read"),
    ),
)
```

### User (`pkg/component/user`)

**Purpose**: User management with roles and permissions

**Technology**: GORM repositories, service layer

**Key Features**:
- User CRUD operations
- Role and permission management
- Password hashing (bcrypt)
- Repositories: `UserRepository`, `RoleRepository`, `PermissionRepository`
- Service layer for business logic

### Email (`pkg/component/email`)

**Purpose**: Email sending via SMTP

**Technology**: go-mail library

**Key Features**:
- Template-based emails (HTML/plain text)
- SMTP configuration
- TLS support
- Attachment handling

### Blob (`pkg/component/blob`)

**Purpose**: Cloud storage abstraction

**Technology**: gocloud.dev/blob

**Key Features**:
- Unified API for S3, GCS, Azure Blob, local filesystem
- Factory pattern for provider selection
- Registry for custom providers

### Media (`pkg/component/media`)

**Purpose**: Media file management (images, videos, documents)

**Technology**: Blob storage, image processing

**Key Features** (prototype stage):
- Media upload and storage
- Image resizing and cropping
- File type validation
- Storage backend abstraction

### Seeder (`pkg/component/seeder`)

**Purpose**: Database seeding for testing/development

**Technology**: GORM, fake data generation

**Key Features**:
- Seed data insertion
- Data factory pattern
- Foreign key handling
- Idempotent seeding

## FX Modules

Each FX module (in `pkg/fx/*`) wraps a component and provides:
- Uber FX module definition (`fx.Module`)
- Component constructor providers (`fx.Provide`)
- Command registration (`fxcore.AsRootCommand`)
- Configuration loading (`fxconfig.AsConfig`)

**Available Modules**:
- `fxcore` - Core utilities (command groups)
- `fxconfig` - Configuration loading
- `fxlogger` - Logger setup
- `fxhttpserver` - HTTP server with Zorya
- `fxvalidator` - Validator setup
- `fxorm` - Database connection
- `fxsecurity` - Security middleware
- `fxuser` - User system
- `fxemail` - Email sender
- `fxblob` - Blob storage
- `fxseeder` - Database seeder
- `fxmedia` - Media management

## Data Flow

### HTTP Request Flow

1. **Request Arrival**: Client sends HTTP request to server
2. **Middleware Chain**:
   - `RequestID` → Generates unique ID, adds to context
   - `HTTPLog` → Logs request start
   - Security middleware → Checks authentication/authorization
3. **Zorya Processing**:
   - Content negotiation → Determines request/response formats
   - Schema decoding → Parses request into input struct (`I`)
   - Validation → Validates input struct against `validate` tags
   - Handler execution → Calls `func(ctx context.Context, input *I) (*O, error)`
   - Response encoding → Marshals output struct (`O`) to response format
   - Error handling → Converts errors to RFC 9457 format
4. **Middleware Chain (response)**:
   - `HTTPLog` → Logs response (status, latency)
5. **Response Delivery**: Client receives typed response

### Configuration Loading Flow

1. **FX Initialization**: `fxconfig.FxConfigModule` starts
2. **File Discovery**: Looks for `configs/config.yaml` (or `APP_CONFIG_PATH`)
3. **Parsing**: Viper loads YAML, merges with environment variables
4. **Type Conversion**: Maps to config structs via `mapstructure`
5. **Validation**: Validates config against struct tags
6. **Injection**: FX provides config to components that need it

### Database Query Flow

1. **Repository Call**: Service calls repository method (e.g., `userRepo.FindByID(ctx, id)`)
2. **GORM Translation**: Repository translates to GORM query
3. **SQL Execution**: GORM generates SQL, executes against database
4. **Row Mapping**: GORM maps rows to struct
5. **Return**: Repository returns domain model to service

## Deployment

### Environments

| Environment | Purpose | Configuration |
|-------------|---------|---------------|
| Development | Local development | `APP_ENV=dev`, SQLite/local PostgreSQL |
| Staging | Pre-production testing | `APP_ENV=staging`, Cloud database |
| Production | Live users | `APP_ENV=prod`, Cloud database, monitoring |

### Build & Deploy Process

1. **Build Binary**:
   ```bash
   go build -o bin/myapp main.go
   ```

2. **Configuration**:
   - Copy `configs/config.yaml` to deployment target
   - Set environment variables (`APP_ENV`, `DB_HOST`, etc.)

3. **Run Application**:
   ```bash
   ./bin/myapp serve-http
   ```

4. **Health Check**:
   - Endpoint: `GET /health`
   - Returns: `200 OK` with `{"status": "ok"}`

### Docker Deployment

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/myapp main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/myapp /myapp
COPY configs/config.yaml /configs/config.yaml
EXPOSE 8080
CMD ["/myapp", "serve-http"]
```

## Configuration

### Environment Variables

| Variable | Purpose | Required | Default |
|----------|---------|----------|---------|
| `APP_ENV` | Environment name | No | `dev` |
| `APP_CONFIG_PATH` | Config file path | No | `configs/config.yaml` |
| `DB_HOST` | Database host | Yes | - |
| `DB_PORT` | Database port | Yes | `5432` |
| `DB_USER` | Database user | Yes | - |
| `DB_PASSWORD` | Database password | Yes | - |
| `DB_NAME` | Database name | Yes | - |
| `LOG_LEVEL` | Log level | No | `info` |

### Feature Flags

| Flag | Purpose | Status | Controls |
|------|---------|--------|----------|
| `ENABLE_OPENAPI_DOCS` | Enable OpenAPI docs endpoint | On | `/docs` endpoint |
| `ENABLE_CORS` | Enable CORS middleware | Off | CORS headers |

## Dependencies

### Runtime Dependencies

**Core**:
- `uber.go/fx` v1.20+ - Dependency injection
- `github.com/spf13/cobra` v1.8+ - CLI framework
- `github.com/spf13/viper` v1.18+ - Configuration
- `log/slog` (stdlib) - Structured logging

**HTTP/API**:
- `github.com/go-chi/chi/v5` v5.0+ - HTTP router
- `github.com/go-chi/httplog/v3` v3.0+ - HTTP logging
- `github.com/gofiber/fiber/v2` v2.52+ - Alternative router (optional)

**Validation/Serialization**:
- `github.com/go-playground/validator/v10` v10.19+ - Request validation
- `github.com/fxamacker/cbor/v2` v2.6+ - CBOR encoding
- `encoding/json` (stdlib) - JSON encoding

**Database**:
- `gorm.io/gorm` v1.25+ - ORM
- `gorm.io/driver/postgres` v1.5+ - PostgreSQL driver
- `gorm.io/driver/sqlite` v1.5+ - SQLite driver

**Storage/Email**:
- `gocloud.dev/blob` v0.37+ - Cloud storage abstraction
- `github.com/wneessen/go-mail` v0.4+ - Email sending

**Security**:
- `github.com/casbin/casbin/v2` v2.82+ - RBAC/ABAC enforcement
- `golang.org/x/crypto` - Password hashing

### Build Dependencies

- Go 1.25+ - Compiler
- `golangci-lint` v1.55+ - Linter

## Monitoring & Observability

### Logs

- **Location**: `stdout` (JSON format in production)
- **Key Events**:
  - Application startup/shutdown
  - HTTP request/response (method, path, status, latency, request ID)
  - Database queries (if slow query log enabled)
  - Errors with stack traces

### Metrics

- **Tool**: Prometheus (future integration)
- **Key Metrics** (planned):
  - HTTP request count by path/status
  - Request latency histogram
  - Active database connections
  - Error rate

### Alerts

- **Tool**: Alertmanager (future integration)
- **Critical Alerts** (planned):
  - HTTP 5xx rate > 1%
  - Request latency p99 > 1s
  - Database connection pool exhausted

## Known Issues & Debt

- [ ] **Media component incomplete**: Storage works, but image processing not implemented
- [ ] **No WebSocket support**: Zorya only supports HTTP/REST, not real-time protocols
- [ ] **OpenAPI security schemes**: Schema generation works, but security scheme definitions need improvement
- [ ] **Test coverage gaps**: Some components <80% coverage (validator, email, seeder)
- [ ] **Performance benchmarks missing**: Need to establish baseline metrics for Zorya overhead
- [ ] **FxHTTPServer RegisterLifecycle unused**: Dead code, should be removed (see [module.go:120-138](/workspace/pkg/fx/fxhttpserver/module.go:120-138))

---

**Last Updated**: 2026-01-14
**Updated By**: AI Assistant
