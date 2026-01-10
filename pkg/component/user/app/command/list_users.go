package command

import (
	"context"
	"log/slog"

	"github.com/talav/talav/pkg/component/user/domain"
	"github.com/talav/talav/pkg/component/user/infra/repository"
	"gorm.io/gorm"
)

// ListUsersQuery represents the query to list users with pagination and filtering.
type ListUsersQuery struct {
	Cursor      string
	Limit       int
	EmailFilter string
	NameFilter  string
}

// ListUsersResult represents the result of listing users.
type ListUsersResult struct {
	Users      []*domain.User
	NextCursor string
	HasMore    bool
}

// ListUsersQueryHandler handles user listing queries.
type ListUsersQueryHandler struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
}

// NewListUsersQueryHandler creates a new ListUsersQueryHandler.
func NewListUsersQueryHandler(userRepo repository.UserRepository, logger *slog.Logger) *ListUsersQueryHandler {
	return &ListUsersQueryHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Handle processes the ListUsersQuery and returns the paginated user list.
func (h *ListUsersQueryHandler) Handle(ctx context.Context, query *ListUsersQuery) (*ListUsersResult, error) {
	// Build dynamic filter specification
	spec := func(db *gorm.DB) *gorm.DB {
		if query.EmailFilter != "" {
			db = db.Where("LOWER(email) LIKE LOWER(?)", "%"+query.EmailFilter+"%")
		}
		if query.NameFilter != "" {
			db = db.Where("LOWER(name) LIKE LOWER(?)", "%"+query.NameFilter+"%")
		}
		return db
	}

	// Query users with specification and cursor pagination
	users, err := h.userRepo.ListWithSpec(ctx, spec, query.Limit, query.Cursor)
	if err != nil {
		h.logger.Error("Failed to list users", "error", err)
		return nil, err
	}

	// Determine pagination metadata
	hasMore := len(users) > query.Limit
	if hasMore {
		// Remove the extra record used to check if there are more pages
		users = users[:query.Limit]
	}

	var nextCursor string
	if hasMore && len(users) > 0 {
		// Use the ID of the last user as the next cursor
		nextCursor = users[len(users)-1].UserID
	}

	return &ListUsersResult{
		Users:      users,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
