# Security Component

Framework-agnostic security component providing JWT authentication, password hashing, and user abstraction.

## Features

- **JWT Token Management**: Create and validate access tokens and refresh tokens
- **Password Hashing**: Bcrypt-based password hashing with salt
- **User Abstraction**: Interface-based user provider for decoupling authentication from domain
- **Context Helpers**: Request context integration for authenticated users
- **Algorithm Support**: HS256 (symmetric) and RS256 (asymmetric) JWT algorithms

## Installation

```bash
go get github.com/talav/talav/pkg/component/security
```

## Quick Start

### Password Hashing

```go
import "github.com/talav/talav/pkg/component/security"

factory := security.NewDefaultSecurityFactory()
cfg := security.DefaultSecurityConfig()
hasher := factory.CreatePasswordHasher(cfg)

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

jwtService, _ := factory.CreateJWTService(jwtCfg)

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

func (p *MyUserProvider) GetUserByIdentifier(ctx context.Context, identifier string) (*security.SecurityUser, error) {
    // Lookup user by email/username
    // Convert to SecurityUser
    return &security.SecurityUser{
        ID: "user123",
        PasswordHash: "...",
        Salt: "...",
        Roles: []string{"user"},
    }, nil
}
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
refreshService := factory.CreateRefreshTokenService(jwtService, store)

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

