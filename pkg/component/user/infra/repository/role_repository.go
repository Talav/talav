package repository

import (
	"context"

	"github.com/talav/talav/pkg/component/orm"
	"github.com/talav/talav/pkg/component/user/domain"
	"gorm.io/gorm"
)

// RoleRepository provides role-specific database operations.
type RoleRepository struct {
	*orm.BaseRepository[domain.Role]
}

// NewRoleRepository creates a new role repository instance.
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		BaseRepository: orm.NewBaseRepository[domain.Role](db),
	}
}

// FindOneByName finds a role by name.
func (r *RoleRepository) FindOneByName(ctx context.Context, name string) (*domain.Role, error) {
	return r.FindOne(ctx, "name", name)
}

// EntityName returns the entity name for validation registry.
func (r *RoleRepository) EntityName() string {
	return "Role"
}
