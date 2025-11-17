package orm

import (
	"context"

	"gorm.io/gorm"
)

// BaseRepositoryInterface defines the common CRUD operations interface
type BaseRepositoryInterface[T any] interface {
	// Base repository methods
	FindByID(ctx context.Context, id string) (*T, error)
	FindByIDWithPreloads(ctx context.Context, id string, preloads ...string) (*T, error)
	FindOne(ctx context.Context, field string, value any) (*T, error)
	FindOneWithPreloads(ctx context.Context, field string, value any, preloads ...string) (*T, error)
	Find(ctx context.Context, limit, offset int) ([]*T, error)
	FindWithPreloads(ctx context.Context, limit, offset int, preloads ...string) ([]*T, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, conditions map[string]any) (bool, error)
	GetDB() *gorm.DB
	EntityName() string // Returns entity name for registry (e.g., "User")
}

// see https://github.com/aklinkert/go-gorm-repository/blob/master/repository.go for more deailts about this approach
// BaseRepository provides common CRUD operations for any entity type
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository instance
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// FindByID retrieves an entity by its string ID
func (r *BaseRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	return r.FindByIDWithPreloads(ctx, id)
}

// FindByIDWithPreloads retrieves an entity by its string ID with specified preloads
func (r *BaseRepository[T]) FindByIDWithPreloads(ctx context.Context, id string, preloads ...string) (*T, error) {
	return r.FindOneWithPreloads(ctx, "id", id, preloads...)
}

// FindOne retrieves an entity by a specific field
func (r *BaseRepository[T]) FindOne(ctx context.Context, field string, value any) (*T, error) {
	return r.FindOneWithPreloads(ctx, field, value)
}

// FindOneWithPreloads retrieves an entity by a specific field with specified preloads
func (r *BaseRepository[T]) FindOneWithPreloads(ctx context.Context, field string, value any, preloads ...string) (*T, error) {
	var entity T
	query := r.db.WithContext(ctx)

	// Apply preloads
	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.Where(field+" = ?", value).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// Find retrieves all entities (with optional limit/offset)
func (r *BaseRepository[T]) Find(ctx context.Context, limit, offset int) ([]*T, error) {
	return r.FindWithPreloads(ctx, limit, offset)
}

// FindWithPreloads retrieves all entities with specified preloads (with optional limit/offset)
func (r *BaseRepository[T]) FindWithPreloads(ctx context.Context, limit, offset int, preloads ...string) ([]*T, error) {
	var entities []*T
	query := r.db.WithContext(ctx)

	// Apply preloads
	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// Exists checks if an entity exists with given conditions (optimized with LIMIT 1)
func (r *BaseRepository[T]) Exists(ctx context.Context, conditions map[string]any) (bool, error) {
	var exists bool
	query := r.db.WithContext(ctx).Model(new(T)).Select("1").Limit(1)

	for field, value := range conditions {
		query = query.Where(field+" = ?", value)
	}

	err := query.Scan(&exists).Error
	return exists, err
}

// GetDB returns the underlying GORM database instance
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}

// Create inserts a new entity
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// Update saves changes to an existing entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete removes an entity by its string ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}
