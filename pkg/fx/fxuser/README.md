# fxuser

FX module for [pkg/component/user](../../component/user). Wires user and role repositories, CQRS command/query handlers, password validation, and the security user provider adapter.

## Quick start

```go
fx.New(
    fxconfig.FxConfigModule,
    fxlogger.FxLoggerModule,
    fxorm.FxORMModule,
    fxvalidator.FxValidatorModule,
    fxsecurity.FxSecurityModule,
    fxuser.Module,
)
```

## Configuration

```yaml
auth:
  password_reset:
    token_expiration_minutes: 60
    reset_url_template: "https://myapp.com/reset-password?token=%s"
```

## What's wired

| Type | Description |
|---|---|
| `*command.CreateUserHandler` | Create a user |
| `*command.GetUserQueryHandler` | Fetch a user by ID |
| `*command.ListUsersQueryHandler` | List users |
| `repository.UserRepository` | User data access |
| `*repository.RoleRepository` | Role data access |
| `repository.PasswordResetTokenRepository` | Password reset tokens |
| `*security.UserProviderAdapter` | Bridge between user repo and security enforcer |

Repositories are automatically registered into the `fxorm` unique-validator registry.

## Dependencies

- `fxorm.FxORMModule`
- `fxvalidator.FxValidatorModule`
- `fxsecurity.FxSecurityModule`
