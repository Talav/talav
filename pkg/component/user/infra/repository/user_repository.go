package repository

import (
	"context"

	"github.com/talav/talav/pkg/component/orm"
	"github.com/talav/talav/pkg/component/user/domain"
	"gorm.io/gorm"
)

// UserRepository interface defines all user repository methods.
// It embeds the base repository interface and adds user-specific methods.
type UserRepository interface {
	orm.BaseRepositoryInterface[domain.User] // Embed base methods

	// User-specific methods
	FindOneByEmail(ctx context.Context, email string) (*domain.User, error)
	FindOneByEmailWithPreloads(ctx context.Context, email string, preloads ...string) (*domain.User, error)

	// Flexible listing with specification pattern
	ListWithSpec(ctx context.Context, spec func(*gorm.DB) *gorm.DB, limit int, cursor string) ([]*domain.User, error)
}

// userRepository implements UserRepository with embedded base repository.
type userRepository struct {
	*orm.BaseRepository[domain.User]
}

// NewUserRepository creates a new user repository instance.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		BaseRepository: orm.NewBaseRepository[domain.User](db),
	}
}

// FindOneByEmail finds a user by email (automatically preloads roles).
func (r *userRepository) FindOneByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.FindOneWithPreloads(ctx, "email", email, "Roles")
}

// FindByID finds a user by ID (automatically preloads roles).
func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return r.FindByIDWithPreloads(ctx, id, "Roles")
}

// FindOneByEmailWithPreloads finds a user by email with specified preloads.
func (r *userRepository) FindOneByEmailWithPreloads(ctx context.Context, email string, preloads ...string) (*domain.User, error) {
	return r.FindOneWithPreloads(ctx, "email", email, preloads...)
}

// ListWithSpec retrieves users with dynamic filtering and cursor-based pagination.
func (r *userRepository) ListWithSpec(ctx context.Context, spec func(*gorm.DB) *gorm.DB, limit int, cursor string) ([]*domain.User, error) {
	var users []*domain.User
	query := r.GetDB().WithContext(ctx)

	// Apply cursor-based pagination
	if cursor != "" {
		query = query.Where("id > ?", cursor)
	}

	// Apply dynamic filters using specification
	query = spec(query)

	// Apply ordering and limit
	query = query.Order("id ASC").Limit(limit + 1) // +1 to check if there are more pages

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// EntityName returns the entity name for validation registry.
func (r *userRepository) EntityName() string {
	return "User"
}
