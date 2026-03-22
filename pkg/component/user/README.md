# user

User management component. Provides CQRS-style command/query handlers for user and role operations, GORM-backed repositories, bcrypt password hashing, and a security adapter for integration with `pkg/component/security`.

## What's included

- **Repositories**: `UserRepository`, `RoleRepository`, `PasswordResetTokenRepository` (PostgreSQL via GORM)
- **Command handlers**: `CreateUserHandler`, `GetUserQueryHandler`, `ListUsersQueryHandler`
- **Validators**: `PasswordValidator` (strength rules with i18n translations)
- **Security adapter**: `UserProviderAdapter` — bridges the user repo to the security enforcer interface

## Configuration

```go
type PasswordResetConfig struct {
    TokenExpirationMinutes int    `config:"token_expiration_minutes"` // default: 60
    ResetURLTemplate       string `config:"reset_url_template"`       // default: http://localhost:3000/reset-password?token=%s
}
```

```yaml
auth:
  password_reset:
    token_expiration_minutes: 60
    reset_url_template: "https://myapp.com/reset-password?token=%s"
```

## Dependencies

Requires `fxorm` (database), `fxvalidator` (custom validators), and `fxsecurity` (password hasher) to be in the FX graph. Use `fxuser.Module` to wire everything:

```go
fx.New(
    fxorm.FxORMModule,
    fxvalidator.FxValidatorModule,
    fxsecurity.FxSecurityModule,
    fxuser.Module,
)
```

## Notes

- The component is early-stage. HTTP handlers for user endpoints live in `pkg/module/user`, not here.
- Role and permission assignment is not yet implemented at the command handler level.
