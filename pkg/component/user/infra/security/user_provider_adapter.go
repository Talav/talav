package security

import (
	"context"

	"github.com/talav/talav/pkg/component/security"
	"github.com/talav/talav/pkg/component/user/infra/repository"
)

// UserProviderAdapter adapts UserRepository to security.UserProvider.
// This decouples the security module from the user domain.
type UserProviderAdapter struct {
	userRepo repository.UserRepository
}

// NewUserProviderAdapter creates a new adapter.
func NewUserProviderAdapter(userRepo repository.UserRepository) security.UserProvider {
	return &UserProviderAdapter{
		userRepo: userRepo,
	}
}

// GetUserByIdentifier implements security.UserProvider.
func (a *UserProviderAdapter) GetUserByIdentifier(ctx context.Context, identifier string) (security.SecurityUser, error) {
	user, err := a.userRepo.FindOneByEmail(ctx, identifier)
	if err != nil {
		return nil, err
	}

	// domain.User implements security.SecurityUser interface
	return user, nil
}
