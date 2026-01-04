# fxhttpserver

Fx module for HTTP server with Zorya API integration.

## Overview

The `fxhttpserver` module provides a complete HTTP server setup with:
- Chi router integration
- Zorya API framework for type-safe endpoints
- Middleware registration system with priority-based ordering
- Request ID and HTTP logging middleware
- OpenAPI documentation generation

## Middleware Registration

The module supports registering custom router-level middlewares with explicit priority control.

### Priority System

Middleware execution order is determined by numeric priority values. Lower numbers execute first. Built-in middlewares have fixed priorities:

- **RequestID**: priority 100 (always first infrastructure middleware)
- **HTTPLog**: priority 200 (after RequestID, if enabled)

### Recommended Priority Ranges

- **< 100**: Before RequestID (rare, usually not needed)
- **100-199**: Between RequestID and HTTPLog
- **200-299**: Between HTTPLog and Zorya (recommended default)
- **>= 300**: Right before Zorya adapter

### Priority Constants

```go
const (
    PriorityRequestID   = 100  // RequestID middleware
    PriorityHTTPLog     = 200  // HTTPLog middleware
    PriorityBeforeZorya = 250  // Default for user middlewares
)
```

### Registering Middleware

Use `AsMiddleware` to register a middleware with a priority:

```go
import (
    "net/http"
    "github.com/talav/talav/pkg/fx/fxhttpserver"
    "go.uber.org/fx"
)

func main() {
    fx.New(
        fxhttpserver.FxHTTPServerModule,
        fxhttpserver.AsMiddleware(
            corsMiddleware,
            fxhttpserver.PriorityBeforeZorya,
            "cors",
        ),
        fxhttpserver.AsMiddleware(
            rateLimitMiddleware,
            240,
            "rate-limit",
        ),
    )
}

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        next.ServeHTTP(w, r)
    })
}
```

### Middleware Execution Order

The final middleware stack executes in this order:

1. User middlewares with priority < 100 (sorted by priority)
2. RequestID (built-in, priority 100)
3. User middlewares with priority 100-199 (sorted by priority)
4. HTTPLog (built-in, priority 200, if enabled)
5. User middlewares with priority 200-299 (sorted by priority)
6. User middlewares with priority >= 300 (sorted by priority)
7. Zorya adapter (wraps router)

Within each priority range, middlewares execute in registration order (fx group order).

## Router-Level vs Group-Level Middleware

### Router-Level Middleware (fxhttpserver)

- Applied globally to all routes via Chi router (`router.Use()`)
- Registered via `fxhttpserver.AsMiddleware()`
- Executes before Zorya processes the request
- Use for: CORS, rate limiting, global auth checks, request tracing

### Group-Level Middleware (Zorya)

- Applied only to routes in a specific Zorya group
- Registered via `group.UseMiddleware()` on Zorya groups
- Handled internally by Zorya
- Use for: group-specific auth, version-specific logic

These work together:

```
Router middleware (global) → Zorya adapter → Group middleware → Route handler
```

### Example: Combining Both

```go
// Router-level: CORS for all routes
fxhttpserver.AsMiddleware(
    corsMiddleware,
    fxhttpserver.PriorityBeforeZorya,
    "cors",
)

// Group-level: Auth for specific routes
group := zorya.NewGroup(api, "/v1")
group.UseMiddleware(authMiddleware)

zorya.Get(group, "/users", getUserHandler)
```

## Examples

### Registering CORS Middleware

```go
fxhttpserver.AsMiddleware(
    func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    },
    fxhttpserver.PriorityBeforeZorya,
    "cors",
)
```

### Registering Rate Limiting Middleware

```go
fxhttpserver.AsMiddleware(
    rateLimitMiddleware,
    240, // Between HTTPLog and Zorya
    "rate-limit",
)
```

### Registering Auth Middleware

```go
fxhttpserver.AsMiddleware(
    authMiddleware,
    230, // Before rate limiting
    "auth",
)
```

### Middleware with Dependencies

You can create middleware constructors that accept dependencies:

```go
func NewAuthMiddleware(cfg *config.Config, logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Use cfg and logger
            token := r.Header.Get("Authorization")
            if !validateToken(token, cfg) {
                logger.Warn("unauthorized request")
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// In your fx setup:
fx.Provide(NewAuthMiddleware),
fxhttpserver.AsMiddleware(
    NewAuthMiddleware, // fx will inject dependencies
    fxhttpserver.PriorityBeforeZorya,
    "auth",
)
```

## Configuration

The module uses `httpserver.Config` for configuration. See [httpserver component documentation](../../component/httpserver/README.md) for details.

## Dependencies

- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/go-chi/httplog/v3` - HTTP request logging
- `github.com/talav/talav/pkg/component/zorya` - API framework
- `go.uber.org/fx` - Dependency injection

## See Also

- [Zorya API Framework](../../component/zorya/README.md) - Type-safe API framework
- [HTTPServer Component](../../component/httpserver/README.md) - HTTP server component
- [Fx Config Module](../fxconfig/README.md) - Configuration management


