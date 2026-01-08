# Zorya Security Adapter

This package provides middleware for enforcing security requirements on [Zorya](../../zorya/) framework routes.

## Overview

The adapter connects two independent components:
- **Security Component** (`pkg/component/security`): Generic authentication and authorization primitives
- **Zorya Framework** (`pkg/component/zorya`): Type-safe HTTP routing framework

This follows the principle of separation of concerns - neither component depends on the other, and the adapter acts as glue code.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Zorya Routes (with Secure() metadata)                  │
│  - Defines what needs protection                        │
└────────────────┬────────────────────────────────────────┘
                 │
                 │ RouteSecurityContext in context
                 ▼
┌─────────────────────────────────────────────────────────┐
│  Adapter Middleware (this package)                      │
│  - Reads Zorya metadata                                 │
│  - Converts to generic SecurityRequirements             │
│  - Calls SecurityEnforcer                               │
└────────────────┬────────────────────────────────────────┘
                 │
                 │ SecurityRequirements
                 ▼
┌─────────────────────────────────────────────────────────┐
│  Security Enforcer                                       │
│  - Generic authorization logic                           │
│  - No knowledge of HTTP or Zorya                         │
└─────────────────────────────────────────────────────────┘
```

## Installation

```bash
go get github.com/talav/talav/pkg/component/security/adapter/zorya
```

## Usage

### Basic Setup

```go
import (
    "github.com/talav/talav/pkg/component/security"
    securityzorya "github.com/talav/talav/pkg/component/security/adapter/zorya"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

// Create API
adapter := adapters.NewChiAdapter(chi.NewRouter())
api := zorya.NewAPI(adapter)

// Register JWT middleware (authentication)
jwtService := security.NewJWTService(jwtConfig)
api.UseMiddleware(security.NewJWTMiddleware(jwtService))

// Register enforcement middleware (authorization)
enforcer := security.NewSimpleEnforcer()
api.UseMiddleware(securityzorya.NewEnforcementMiddleware(enforcer))

// Register protected routes
zorya.Get(api, "/admin/users", getUsersHandler,
    zorya.Secure(
        zorya.Roles("admin"),
    ),
)

zorya.Get(api, "/orgs/{orgId}/projects", getProjectsHandler,
    zorya.Secure(
        zorya.Roles("member"),
        zorya.ResourceFromParams(func(params map[string]string) string {
            return "organizations/" + params["orgId"]
        }),
        zorya.Action("view"),
    ),
)
```

### Custom Error Handlers

```go
api.UseMiddleware(securityzorya.NewEnforcementMiddleware(
    enforcer,
    securityzorya.WithUnauthorizedHandler(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(401)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Authentication required",
            "code":  "UNAUTHORIZED",
        })
    }),
    securityzorya.WithForbiddenHandler(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(403)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Insufficient permissions",
            "code":  "FORBIDDEN",
        })
    }),
))
```

### FX Integration

For applications using Uber FX:

```go
import (
    "github.com/talav/talav/pkg/fx/fxsecurity"
    "go.uber.org/fx"
)

fx.New(
    fxsecurity.FxSecurityModule,
    fxsecurity.AsJWTMiddleware(),         // JWT authentication
    fxsecurity.AsEnforcementMiddleware(), // Security enforcement (uses adapter)
    // ... rest of application
)
```

## How It Works

### 1. Route Registration

When you use `zorya.Secure()` on a route:

```go
zorya.Get(api, "/orgs/{orgId}", handler,
    zorya.Secure(
        zorya.Roles("admin"),
        zorya.ResourceFromParams(func(params map[string]string) string {
            return "orgs/" + params["orgId"]
        }),
    ),
)
```

Zorya injects metadata middleware that:
- Extracts path parameters
- Resolves resource templates (e.g., `orgs/{orgId}` → `orgs/123`)
- Stores `RouteSecurityContext` in request context

### 2. Enforcement

The adapter middleware reads the metadata and:

```go
// 1. Read Zorya metadata
zoryaMeta := zorya.GetRouteSecurityContext(r)

// 2. Get authenticated user
user := security.GetAuthUser(r)

// 3. Convert to generic requirements
requirements := &security.SecurityRequirements{
    Roles:       zoryaMeta.Roles,
    Permissions: zoryaMeta.Permissions,
    Resource:    zoryaMeta.Resource,  // Already resolved!
    Action:      zoryaMeta.Action,
}

// 4. Enforce
ok, err := enforcer.Enforce(r.Context(), user, requirements)
```

### 3. Benefits

- **Clean separation**: Security and Zorya are independent
- **Testable**: Mock either side easily
- **Flexible**: Swap enforcers or routing frameworks
- **Go-idiomatic**: Standard middleware pattern

## Configuration Options

### `WithUnauthorizedHandler`

Sets custom handler for 401 Unauthorized responses (missing authentication).

### `WithForbiddenHandler`

Sets custom handler for 403 Forbidden responses (insufficient permissions).

## Related Packages

- **[security](../../)**: Generic security component (JWT, enforcers, etc.)
- **[zorya](../../zorya/)**: Type-safe HTTP routing framework
- **[fxsecurity](../../../fx/fxsecurity/)**: FX module for auto-wiring

## Examples

See [examples directory](../../../../examples/) for complete working examples.

## License

See root LICENSE file.






