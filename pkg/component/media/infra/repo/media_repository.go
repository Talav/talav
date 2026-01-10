package repo

import (
	"github.com/talav/talav/pkg/component/media/domain"
	"github.com/talav/talav/pkg/component/orm"
	"gorm.io/gorm"
)

// MediaRepository interface defines all media repository methods
// It embeds the base repository interface.
type MediaRepository interface {
	orm.BaseRepositoryInterface[domain.Media] // Embed base methods
}

// mediaRepository implements MediaRepository with embedded base repository.
type mediaRepository struct {
	*orm.BaseRepository[domain.Media]
}

// NewMediaRepository creates a new media repository instance.
func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{
		BaseRepository: orm.NewBaseRepository[domain.Media](db),
	}
}

// EntityName returns the entity name for validation registry.
func (r *mediaRepository) EntityName() string {
	return "Media"
}
