package domain

import (
	"fmt"
	"slices"

	"github.com/talav/talav/pkg/component/security"
	"github.com/talav/talav/pkg/component/user"
)

const UserIDPrefix = "usr"

// User represents a user in the domain layer.
type User struct {
	UserID         string `json:"id" gorm:"column:id;primaryKey"`
	Name           string `json:"name" validate:"required,min=2"`
	Email          string `json:"email" validate:"required,email" gorm:"uniqueIndex"`
	HashedPassword string `json:"password" gorm:"column:password"`
	UserSalt       string `json:"salt" gorm:"column:salt"`
	UserRoles      []Role `json:"roles,omitempty" gorm:"many2many:user_roles;foreignKey:UserID;joinForeignKey:UserID;References:ID;joinReferences:RoleID"`
	CreatedAt      int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      int64  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      *int64 `json:"deleted_at,omitempty" gorm:"index"`
}

// NewUser creates a new User entity with generated ID, salt, and hashed password.
// This is the proper way to create users in the domain.
func NewUser(name, email, plainPassword string, hasher security.PasswordHasher) (*User, error) {
	// Business validation
	if name == "" || email == "" || plainPassword == "" {
		return nil, fmt.Errorf("name, email, and password are required")
	}

	// Generate salt
	salt, err := hasher.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password with salt
	hashedPassword, err := hasher.HashPassword(plainPassword, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &User{
		UserID:         user.GenerateID(UserIDPrefix),
		Name:           name,
		Email:          email,
		HashedPassword: hashedPassword,
		UserSalt:       salt,
	}

	return newUser, nil
}

// GetRoleNames returns array of role names for JWT claims.
func (u *User) GetRoleNames() []string {
	roleNames := make([]string, len(u.UserRoles))
	for i, role := range u.UserRoles {
		roleNames[i] = role.Name
	}

	return roleNames
}

// IsAdmin returns true if the user has the admin role.
func (u *User) IsAdmin() bool {
	return slices.Contains(u.GetRoleNames(), RoleIDAdmin)
}

// GetOwnerID implements an owned resource interface.
// Users own themselves.
func (u *User) GetOwnerID() string {
	return u.UserID
}

// Implement security.SecurityUser interface for authentication
// These methods allow User to be used directly as a SecurityUser

func (u *User) ID() string {
	return u.UserID
}

func (u *User) PasswordHash() string {
	return u.HashedPassword
}

func (u *User) Salt() string {
	return u.UserSalt
}

func (u *User) Roles() []string {
	return u.GetRoleNames()
}
