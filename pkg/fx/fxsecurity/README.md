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

func (p *MyUserProvider) GetUserByIdentifier(ctx context.Context, identifier string) (*security.SecurityUser, error) {
	// Lookup and return user
}

func main() {
	fx.New(
		fxsecurity.FxSecurityModule,
		fxsecurity.AsJWTMiddleware(jwtService, cfg), // Register JWT middleware
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
```

## Middleware Registration

JWT middleware must be registered explicitly at the app level:

```go
fxsecurity.AsJWTMiddleware(jwtService, cfg)
```

RBAC middleware can be registered per-route:

```go
authzMiddleware := fxsecurity.NewAuthZMiddleware(enforcer)
router.Use(authzMiddleware.EnforceAccess("resource123", "read"))
```

