# Security HTTP Module

Opinionated HTTP endpoints for authentication using Zorya framework.

## Overview

Provides ready-to-use authentication endpoints:
- **Login** - Email/password authentication with JWT tokens
- **Logout** - Token revocation

## Installation

```bash
go get github.com/talav/talav/pkg/module/securityhttp
```

## Usage

### With FX Module (Recommended)

```go
package main

import (
    "github.com/talav/talav/pkg/fx/fxconfig"
    "github.com/talav/talav/pkg/fx/fxhttpserver"
    "github.com/talav/talav/pkg/fx/fxsecurity"
    "github.com/talav/talav/pkg/module/securityhttp"
    "go.uber.org/fx"
)

func main() {
    fx.New(
        fxconfig.FxConfigModule,
        fxhttpserver.FxHTTPServerModule,
        fxsecurity.FxSecurityModule,
        securityhttp.Module,
        
        // Provide UserProvider implementation
        fx.Provide(NewMyUserProvider),
    ).Run()
}
```

### Manual Registration

```go
package main

import (
    "github.com/talav/talav/pkg/component/security"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/module/securityhttp"
    "github.com/talav/talav/pkg/module/securityhttp/handler"
)

func main() {
    api := zorya.NewAPI(adapter)
    
    // Create handlers
    loginHandler := handler.NewLoginHandler(userProvider, hasher, jwtService)
    logoutHandler := handler.NewLogoutHandler(refreshService)
    
    // Register routes
    securityhttp.RegisterRoutes(api, loginHandler, logoutHandler)
}
```

## Endpoints

### POST /auth/login

Authenticate user with email and password.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Errors:**
- `400` - Validation error (invalid email format, missing fields)
- `401` - Invalid credentials

### POST /auth/logout

Revoke user's refresh tokens. Requires authentication.

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200):**
```json
{}
```

**Errors:**
- `401` - Not authenticated

## Security

- Passwords verified using bcrypt
- JWT tokens for authentication
- User enumeration prevention (same error for invalid email/password)
- Refresh token revocation on logout

## Dependencies

Requires:
- `pkg/component/security` - Authentication primitives
- `pkg/component/zorya` - HTTP routing framework
- `pkg/fx/fxsecurity` - Security FX module (if using FX)

## Architecture

```
securityhttp/
├── dto/              # Request/response types
│   └── login.go
├── handler/          # HTTP handlers (thin layer)
│   ├── login_handler.go
│   └── logout_handler.go
├── routes.go         # Route registration
├── module.go         # FX module
└── README.md
```

**Design principles:**
- Thin handlers - delegate to security component
- DTOs separate from domain models
- Framework-specific (Zorya)
- Opinionated defaults
- FX-friendly but not FX-dependent

## Customization

### Custom Error Responses

Override error handling in handlers:

```go
type CustomLoginHandler struct {
    *handler.LoginHandler
}

func (h *CustomLoginHandler) Handle(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
    resp, err := h.LoginHandler.Handle(ctx, req)
    if err != nil {
        // Custom error logging, metrics, etc.
        return nil, customError(err)
    }
    return resp, nil
}
```

### Additional Routes

Add custom auth routes alongside the module:

```go
fx.New(
    securityhttp.Module,
    fx.Invoke(func(api zorya.API) {
        zorya.Post(api, "/auth/refresh", refreshHandler.Handle,
            zorya.Summary("Refresh Token"),
        )
    }),
)
```

## Testing

```go
func TestLoginHandler(t *testing.T) {
    userProvider := &mockUserProvider{}
    hasher := security.NewPasswordHasher(security.DefaultHasherConfig())
    jwtService, _ := security.NewJWTService(security.DefaultJWTConfig())
    
    handler := handler.NewLoginHandler(userProvider, hasher, jwtService)
    
    req := &dto.LoginRequest{
        Email:    "test@example.com",
        Password: "password123",
    }
    
    resp, err := handler.Handle(context.Background(), req)
    if err != nil {
        t.Fatal(err)
    }
    
    if resp.Token == "" {
        t.Error("expected token")
    }
}
```

## See Also

- [security component](../../component/security/README.md) - Core authentication primitives
- [zorya component](../../component/zorya/README.md) - HTTP routing framework
- [fxsecurity module](../../fx/fxsecurity/README.md) - Security FX wiring
