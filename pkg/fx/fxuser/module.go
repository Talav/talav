package fxuser

import (
	"log/slog"

	"github.com/talav/talav/pkg/component/user"
	"github.com/talav/talav/pkg/component/user/app/command"
	"github.com/talav/talav/pkg/component/user/infra/repository"
	"github.com/talav/talav/pkg/component/user/infra/security"
	uservalidator "github.com/talav/talav/pkg/component/user/infra/validator"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxorm"
	"github.com/talav/talav/pkg/fx/fxvalidator"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

const ModuleName = "user"

// Module provides user component functionality within the FX ecosystem.
// It registers repositories, service handlers, validators, and the security adapter.
//
// Dependencies:
//   - fxorm.Module (for database and repository registry)
//   - fxvalidator.Module (for custom validator registration)
//   - fxsecurity.Module (for PasswordHasher)
//
// Usage:
//
//	fx.New(
//	    fxorm.Module,
//	    fxvalidator.Module,
//	    fxsecurity.Module,
//	    fxuser.Module,
//	    // ... application modules
//	)
var Module = fx.Module(
	ModuleName,
	// Register configuration
	fxconfig.AsConfigWithDefaults(
		"auth.password_reset",
		user.PasswordResetConfig{},
		user.PasswordResetConfig{
			TokenExpirationMinutes: 60,
			ResetURLTemplate:       "http://localhost:3000/reset-password?token=%s",
		},
	),
	fx.Provide(
		// Command handlers
		command.NewCreateUserHandler,
		command.NewGetUserQueryHandler,
		command.NewListUsersQueryHandler,
		// Security adapter
		security.NewUserProviderAdapter,
	),
	// Register repositories to the ORM registry
	fxorm.AsRepository[repository.UserRepository](NewUserRepository),
	fxorm.AsRepository[*repository.RoleRepository](NewRoleRepository),
	fx.Provide(NewPasswordResetTokenRepository),
	// Register validators
	fxvalidator.AsValidatorConstructor(uservalidator.NewPasswordValidator),
	fxvalidator.AsTranslationConstructor(uservalidator.NewPasswordTranslation),
)

// NewUserRepository creates a UserRepository and ensures it's registered.
func NewUserRepository(db *gorm.DB, logger *slog.Logger) repository.UserRepository {
	repo := repository.NewUserRepository(db)
	logger.Info("User repository initialized", "entity", repo.EntityName())
	return repo
}

// NewRoleRepository creates a RoleRepository and ensures it's registered.
func NewRoleRepository(db *gorm.DB, logger *slog.Logger) *repository.RoleRepository {
	repo := repository.NewRoleRepository(db)
	logger.Info("Role repository initialized", "entity", repo.EntityName())
	return repo
}

// NewPasswordResetTokenRepository creates a PasswordResetTokenRepository.
func NewPasswordResetTokenRepository(db *gorm.DB, logger *slog.Logger) repository.PasswordResetTokenRepository {
	repo := repository.NewPasswordResetTokenRepository(db)
	logger.Info("Password reset token repository initialized")
	return repo
}
