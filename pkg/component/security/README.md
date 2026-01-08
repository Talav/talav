# Security Component

Framework-agnostic security component providing JWT authentication, password hashing, authorization, and user abstraction.

## Features

- **JWT Token Management**: Create and validate access tokens and refresh tokens
- **Password Hashing**: Bcrypt-based password hashing with salt
- **Authorization**: Pluggable SecurityEnforcer interface for RBAC/ABAC
- **User Abstraction**: Interface-based user provider for decoupling authentication from domain
- **Context Helpers**: Request context integration for authenticated users
- **Algorithm Support**: HS256 (symmetric) and RS256 (asymmetric) JWT algorithms
- **Framework Adapters**: Clean integration with routing frameworks via adapter pattern

## Installation

```bash
go get github.com/talav/talav/pkg/component/security
```

## Quick Start

### Password Hashing

```go
import "github.com/talav/talav/pkg/component/security"

cfg := security.HasherConfig{
    BcryptCost: 10,
    SaltLength: 32,
}
hasher := security.NewPasswordHasher(cfg)

salt, _ := hasher.GenerateSalt()
hash, _ := hasher.HashPassword("password123", salt)
err := hasher.ComparePassword(hash, "password123", salt)
```

### JWT Service

```go
jwtCfg := security.JWTConfig{
    Algorithm: "HS256",
    Secret: "your-secret-key",
    AccessTokenExpiry: 15 * time.Minute,
    RefreshTokenExpiry: 168 * time.Hour,
}

jwtService, _ := security.NewJWTService(jwtCfg)

// Create tokens
accessToken, _ := jwtService.CreateAccessToken("user123", []string{"admin"})
refreshToken, _ := jwtService.CreateRefreshToken("user123")

// Validate tokens
claims, _ := jwtService.ValidateAccessToken(accessToken)
```

### User Provider

Implement the `UserProvider` interface to integrate with your user domain:

```go
type MyUserProvider struct {
    // your dependencies
}

func (p *MyUserProvider) GetUserByIdentifier(ctx context.Context, identifier string) (security.SecurityUser, error) {
    // Lookup user by email/username
    // Return a type that implements SecurityUser interface
    user := &MyUser{
        id:           "user123",
        passwordHash: "...",
        salt:         "...",
        roles:        []string{"user"},
    }
    return user, nil
}

// MyUser implements SecurityUser interface
type MyUser struct {
    id           string
    passwordHash string
    salt         string
    roles        []string
}

func (u *MyUser) ID() string           { return u.id }
func (u *MyUser) PasswordHash() string { return u.passwordHash }
func (u *MyUser) Salt() string         { return u.salt }
func (u *MyUser) Roles() []string      { return u.roles }
```

## Configuration

```yaml
security:
  bcrypt_cost: 10
  salt_length: 32
  jwt:
    algorithm: "HS256"  # or "RS256"
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

## Refresh Token Rotation

The refresh token service supports token rotation for enhanced security:

```go
refreshService := security.NewRefreshTokenService(jwtService, store)

// Rotate refresh token (invalidates old, creates new)
newToken, userID, err := refreshService.RotateRefreshToken(ctx, oldToken)

// Revoke a token
err := refreshService.RevokeRefreshToken(ctx, token)

// Revoke all tokens for a user
err := refreshService.RevokeAllRefreshTokens(ctx, userID)
```

## Context Integration

Use context helpers to store and retrieve authenticated users:

```go
import "github.com/talav/talav/pkg/component/security"

// In middleware
user := &security.AuthUser{
    ID: "user123",
    Roles: []string{"admin"},
}
r = security.SetAuthUser(r, user)

// In handler
user := security.GetAuthUser(r)
if user == nil {
    // Not authenticated
}
```

## Authorization

### SecurityEnforcer Interface

The `SecurityEnforcer` interface provides pluggable authorization logic:

```go
type SecurityEnforcer interface {
    Enforce(ctx context.Context, user *AuthUser, requirements *SecurityRequirements) (bool, error)
}

type SecurityRequirements struct {
    Roles       []string // User needs at least one
    Permissions []string // User needs all
    Resource    string   // Resource identifier (e.g., "organizations/123")
    Action      string   // Action (e.g., "view", "edit", "POST")
    RequireAuth bool     // Whether authentication is required
}
```

### Built-in Enforcers

**SimpleEnforcer**: Role-based authorization using user roles.

```go
enforcer := security.NewSimpleEnforcer()
requirements := &security.SecurityRequirements{
    Roles:       []string{"admin", "manager"},
    Permissions: []string{"users.edit"},
    RequireAuth: true,
}

ok, err := enforcer.Enforce(ctx, user, requirements)
```

### Custom Enforcers

Implement your own authorization logic:

```go
type CasbinEnforcer struct {
    enforcer *casbin.Enforcer
}

func (e *CasbinEnforcer) Enforce(ctx context.Context, user *AuthUser, req *SecurityRequirements) (bool, error) {
    if user == nil {
        return false, nil
    }
    
    // Check resource-based permissions
    if req.Resource != "" {
        return e.enforcer.Enforce(user.ID, req.Resource, req.Action)
    }
    
    // Check roles
    for _, role := range req.Roles {
        if e.enforcer.HasRoleForUser(user.ID, role) {
            return true, nil
        }
    }
    
    return false, nil
}
```

## Framework Integration

The security component is framework-agnostic. Use adapters to integrate with specific frameworks:

### Zorya Framework

Use the [Zorya adapter](./adapter/zorya/) for type-safe HTTP routing:

```go
import (
    "github.com/talav/talav/pkg/component/security"
    securityzorya "github.com/talav/talav/pkg/component/security/adapter/zorya"
    "github.com/talav/talav/pkg/component/zorya"
)

// Setup
api := zorya.NewAPI(adapter)
api.UseMiddleware(security.NewJWTMiddleware(jwtService))
api.UseMiddleware(securityzorya.NewEnforcementMiddleware(enforcer))

// Protected routes
zorya.Get(api, "/admin/users", handler,
    zorya.Secure(
        zorya.Roles("admin"),
    ),
)

zorya.Get(api, "/orgs/{orgId}/projects", handler,
    zorya.Secure(
        zorya.Roles("member"),
        zorya.ResourceFromParams(func(params map[string]string) string {
            return "organizations/" + params["orgId"]
        }),
    ),
)
```

See [adapter/zorya/README.md](./adapter/zorya/README.md) for full documentation.

### Manual Integration

You can integrate with any HTTP framework manually:

```go
// 1. Authentication middleware
func jwtAuthMiddleware(jwtService *security.JWTService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            claims, err := jwtService.ValidateAccessToken(token)
            if err == nil {
                user := &security.AuthUser{
                    ID:    claims.Subject,
                    Roles: claims.Roles,
                }
                r = security.SetAuthUser(r, user)
            }
            next.ServeHTTP(w, r)
        })
    }
}

// 2. Authorization middleware
func authzMiddleware(enforcer security.SecurityEnforcer) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := security.GetAuthUser(r)
            requirements := &security.SecurityRequirements{
                Roles:       []string{"admin"},
                RequireAuth: true,
            }
            
            ok, _ := enforcer.Enforce(r.Context(), user, requirements)
            if !ok {
                http.Error(w, "Forbidden", 403)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```
