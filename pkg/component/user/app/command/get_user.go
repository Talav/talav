package command

import (
	"context"
	"log/slog"

	"github.com/talav/talav/pkg/component/user/domain"
	"github.com/talav/talav/pkg/component/user/infra/repository"
)

// GetUserQuery represents the query to get a user by ID.
type GetUserQuery struct {
	ID string
}

// GetUserResult represents the result of getting a user.
type GetUserResult struct {
	User *domain.User
}

// GetUserQueryHandler handles user retrieval queries.
type GetUserQueryHandler struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
}

// NewGetUserQueryHandler creates a new GetUserQueryHandler.
func NewGetUserQueryHandler(userRepo repository.UserRepository, logger *slog.Logger) *GetUserQueryHandler {
	return &GetUserQueryHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Handle processes the GetUserQuery and returns the user.
func (h *GetUserQueryHandler) Handle(ctx context.Context, query *GetUserQuery) (*GetUserResult, error) {
	user, err := h.userRepo.FindByID(ctx, query.ID)
	if err != nil {
		return nil, err
	}

	return &GetUserResult{User: user}, nil
}
