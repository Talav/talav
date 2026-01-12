package query

import (
	"context"

	"github.com/talav/talav/pkg/component/media/domain"
	"github.com/talav/talav/pkg/component/media/infra/repo"
	"gorm.io/gorm"
)

// ListMediaQuery represents a query to list media entries.
type ListMediaQuery struct {
	Limit  int
	Cursor string
	Preset string // Optional filter by preset
	Type   string // Optional filter by type (image, video, file)
}

// ListMediaResult represents the result of listing media entries.
type ListMediaResult struct {
	Media      []*domain.Media
	NextCursor string
	HasMore    bool
}

// ListMediaQueryHandler handles list media queries.
type ListMediaQueryHandler struct {
	mediaRepo repo.MediaRepository
}

// NewListMediaQueryHandler creates a new ListMediaQueryHandler.
func NewListMediaQueryHandler(mediaRepo repo.MediaRepository) *ListMediaQueryHandler {
	return &ListMediaQueryHandler{
		mediaRepo: mediaRepo,
	}
}

// Handle processes the ListMediaQuery and returns a list of media entries.
func (h *ListMediaQueryHandler) Handle(ctx context.Context, query *ListMediaQuery) (*ListMediaResult, error) {
	// Build query specification for filters
	spec := func(db *gorm.DB) *gorm.DB {
		queryDB := db

		// Apply filters
		if query.Preset != "" {
			queryDB = queryDB.Where("preset = ?", query.Preset)
		}
		if query.Type != "" {
			// Filter by type column
			queryDB = queryDB.Where("type = ?", query.Type)
		}

		return queryDB
	}

	// Use repository's Find method with limit
	limit := query.Limit
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	// Build query using repository's DB
	queryDB := h.mediaRepo.GetDB().WithContext(ctx)
	queryDB = spec(queryDB)

	// Apply cursor-based pagination
	if query.Cursor != "" {
		queryDB = queryDB.Where("id > ?", query.Cursor)
	}

	// Fetch one extra to check if there are more
	var mediaList []*domain.Media
	err := queryDB.Order("id ASC").Limit(limit + 1).Find(&mediaList).Error
	if err != nil {
		return nil, err
	}

	hasMore := len(mediaList) > limit
	if hasMore {
		// Remove the extra record used to check if there are more pages
		mediaList = mediaList[:limit]
	}

	var nextCursor string
	if hasMore && len(mediaList) > 0 {
		// Use the ID of the last media as the next cursor
		nextCursor = mediaList[len(mediaList)-1].ID
	}

	return &ListMediaResult{
		Media:      mediaList,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
