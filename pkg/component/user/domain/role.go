package domain

import (
	"github.com/talav/talav/pkg/component/user"
)

const (
	RoleIDPrefix = "role"

	RoleIDAdmin = RoleIDPrefix + "_admin"
	RoleIDUser  = RoleIDPrefix + "_user"
)

// Role represents a role that can be assigned to users.
type Role struct {
	ID          string `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"uniqueIndex;not null"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewRole creates a new Role entity.
func NewRole(name, description string) *Role {
	role := &Role{
		ID:          user.GenerateID(RoleIDPrefix),
		Name:        name,
		Description: description,
	}

	return role
}
