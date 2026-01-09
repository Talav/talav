# fxsecurity

Fx module for JWT authentication and RBAC authorization.

## Overview

The `fxsecurity` module provides:
- JWT authentication middleware
- RBAC authorization middleware (via Casbin)
- Login/logout/refresh token handlers
- Password hashing utilities
- User abstraction interface

## Usage

```go
import (
	"github.com/talav/talav/pkg/fx/fxsecurity"
	"github.com/talav/talav/pkg/fx/fxhttpserver"
	"go.uber.org/fx"
)

// Implement UserProvider interface
type MyUserProvider struct {
	// your dependencies
}

func (p *MyUserProvider) GetUserByIdentifier(ctx context.Context, identifier string) (security.SecurityUser, error) {
	// Lookup and return a type that implements security.SecurityUser interface
}

func main() {
	fx.New(
		fxsecurity.FxSecurityModule,
		fxsecurity.AsJWTMiddleware(), // Register JWT middleware (automatically uses jwtService and cfg from Fx)
		fx.Provide(func() security.UserProvider {
			return &MyUserProvider{}
		}),
		fx.Invoke(func(api zorya.API, loginHandler *fxsecurity.LoginHandler) {
			// Register routes
			zorya.Post(api, "/auth/login", loginHandler.Handle)
		}),
	).Run()
}
```

## Configuration

```yaml
security:
  bcrypt_cost: 10
  salt_length: 32
  jwt:
    algorithm: "HS256"
    secret: "${JWT_SECRET}"
    access_token_expiry: "15m"
    refresh_token_expiry: "168h"
  cookie:
    access_token_name: "access_token"
    refresh_token_name: "refresh_token"
    secure: true
    http_only: true
    same_site: "Lax"
  token_source:
    sources: ["header", "cookie"]
    header_name: "Authorization"
  enforcer:
    type: "simple"  # "simple" (default) or "custom"
```

## Security Enforcer Configuration

The security enforcer is selected based on the `security.enforcer.type` configuration value. The enforcer is automatically provided by `FxSecurityModule` and used by `AsSecurityMiddleware()`.

### Enforcer Types

#### Simple Enforcer (Default)

The simple enforcer performs basic role-based checks using roles from the authenticated user. No additional configuration or modules required.

```yaml
security:
  enforcer:
    type: "simple"  # or omit (defaults to "simple")
```

```go
fx.New(
	fxsecurity.FxSecurityModule,
	fxsecurity.AsSecurityMiddleware(), // Uses SimpleEnforcer
)
```

#### Custom Enforcer

When using a custom enforcer, you provide your own `SecurityEnforcer` implementation. The framework only validates that one exists when `type: "custom"` is set.

```yaml
security:
  enforcer:
    type: "custom"
```

```go
// Implement your custom enforcer
type MyCustomEnforcer struct {
	// ... your implementation
}

func (e *MyCustomEnforcer) Enforce(ctx context.Context, subject string, roles []string, resource string, action string) (bool, error) {
	// Your custom logic
}

func (e *MyCustomEnforcer) HasRole(ctx context.Context, subject string, roles []string) (bool, error) {
	// Your custom logic
}

func (e *MyCustomEnforcer) HasPermission(ctx context.Context, subject string, permissions []string) (bool, error) {
	// Your custom logic
}

fx.New(
	fxsecurity.FxSecurityModule,
	fx.Provide(func() security.SecurityEnforcer {
		return &MyCustomEnforcer{}
	}), // Provide your custom enforcer
	fxsecurity.AsSecurityMiddleware(), // Uses your custom enforcer
)
```

**Requirements:**
- Must provide `SecurityEnforcer` via `fx.Provide()`
- Application will fail to start if `type: "custom"` but no `SecurityEnforcer` is provided

**Example: Using Casbin enforcer**

You can use Casbin by setting the enforcer type to "casbin" and providing a persist.Adapter:

```go
import (
	"github.com/casbin/gorm-adapter/v3"
	"github.com/talav/talav/pkg/fx/fxsecurity"
	"go.uber.org/fx"
)

fx.New(
	fxsecurity.FxSecurityModule,
	fx.Provide(func(db *gorm.DB) (persist.Adapter, error) {
		return gormadapter.NewAdapterByDB(db)
	}),
	fxsecurity.AsSecurityMiddleware(),
)
// config.yaml:
//   security.enforcer.type: "casbin"
//   casbin.model_path: "configs/casbin.conf"
```

## Middleware Registration

### JWT Authentication Middleware

JWT middleware must be registered explicitly at the app level:

```go
fxsecurity.AsJWTMiddleware() // Dependencies (jwtService, cfg) are automatically injected by Fx
```

### Security Enforcement Middleware

Security enforcement middleware uses the enforcer configured via `security.enforcer.type`. The enforcer is automatically provided by `FxSecurityModule` based on your configuration.

```go
fxsecurity.AsSecurityMiddleware() // Uses SecurityEnforcer from config
```

The middleware automatically:
- Checks authentication requirements (`RequireAuth()`)
- Validates user roles (`WithRoles()`)
- Enforces permissions (`WithPermissions()`)
- Performs resource-based policy checks (`WithResource()`)

See the [Zorya README](../../component/zorya/README.md#route-security-and-authorization) for details on declarative route security.
