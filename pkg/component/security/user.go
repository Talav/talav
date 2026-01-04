package security

import (
	"context"
	"errors"
)

// ErrUserNotFound is returned when a user is not found by identifier.
var ErrUserNotFound = errors.New("user not found")

// SecurityUser represents a user for authentication purposes.
// This is a security module abstraction that decouples authentication from user domain.
type SecurityUser interface {
	ID() string
	PasswordHash() string
	Salt() string
	Roles() []string
}

// UserProvider provides user lookup for authentication.
// Implementations should convert domain users to SecurityUser.
type UserProvider interface {
	// GetUserByIdentifier retrieves a user by identifier (e.g., email) for authentication.
	// Returns ErrUserNotFound if user doesn't exist.
	GetUserByIdentifier(ctx context.Context, identifier string) (SecurityUser, error)
}
