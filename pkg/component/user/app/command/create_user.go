package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/talav/talav/pkg/component/security"
	"github.com/talav/talav/pkg/component/user/domain"
	"github.com/talav/talav/pkg/component/user/repository"
)

// CreateUserCommand represents the command to create a new user.
type CreateUserCommand struct {
	Name      string
	Email     string
	Password  string
	RoleNames []string
}

// CreateUserResult represents the result of creating a user.
type CreateUserResult struct {
	User *domain.User
}

// CreateUserHandler handles user creation commands.
type CreateUserHandler struct {
	userRepo repository.UserRepository
	roleRepo *repository.RoleRepository
	hasher   security.PasswordHasher
	logger   *slog.Logger
}

// NewCreateUserHandler creates a new CreateUserHandler.
func NewCreateUserHandler(
	userRepo repository.UserRepository,
	roleRepo *repository.RoleRepository,
	hasher security.PasswordHasher,
	logger *slog.Logger,
) *CreateUserHandler {
	return &CreateUserHandler{
		userRepo: userRepo,
		roleRepo: roleRepo,
		hasher:   hasher,
		logger:   logger,
	}
}

// Handle processes the CreateUserCommand and returns the created user.
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) (*CreateUserResult, error) {
	user, err := domain.NewUser(cmd.Name, cmd.Email, cmd.Password, h.hasher)
	if err != nil {
		h.logger.Error("Failed to create user domain entity", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	roleNames := cmd.RoleNames
	if len(roleNames) == 0 {
		roleNames = []string{domain.RoleIDUser}
	}

	roles := make([]domain.Role, 0, len(roleNames))
	for _, roleName := range roleNames {
		role, err := h.roleRepo.FindOneByName(ctx, roleName)
		if err != nil {
			h.logger.Error("Failed to get role by name", "error", err, "roleName", roleName)
			return nil, fmt.Errorf("failed to assign role %s: %w", roleName, err)
		}
		roles = append(roles, *role)
	}

	user.UserRoles = roles

	err = h.userRepo.Create(ctx, user)
	if err != nil {
		h.logger.Error("Failed to save user to repository", "error", err)
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return &CreateUserResult{User: user}, nil
}
