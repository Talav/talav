package handler

import (
	"context"
	"errors"

	"github.com/talav/talav/pkg/component/security"
	"github.com/talav/talav/pkg/component/user/app/command"
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/module/user/dto"
	"gorm.io/gorm"
)

// GetUserHandler handles HTTP requests for retrieving user data.
type GetUserHandler struct {
	getUserService *command.GetUserQueryHandler
}

// NewGetUserHandler creates a new GetUserHandler instance.
func NewGetUserHandler(getUserService *command.GetUserQueryHandler) *GetUserHandler {
	return &GetUserHandler{
		getUserService: getUserService,
	}
}

// Handle retrieves a user by ID.
func (h *GetUserHandler) Handle(ctx context.Context, input *dto.GetUserInput) (*dto.GetUserOutput, error) {
	return h.getUserByID(ctx, input.ID)
}

// HandleCurrentUser retrieves the current authenticated user.
func (h *GetUserHandler) HandleCurrentUser(ctx context.Context, _ *struct{}) (*dto.GetUserOutput, error) {
	// Get authenticated user from context
	authUser := security.GetAuthUserFromContext(ctx)
	if authUser == nil {
		return nil, zorya.Error401Unauthorized("Authentication required")
	}

	return h.getUserByID(ctx, authUser.ID)
}

// getUserByID retrieves a user by ID and handles common error cases.
func (h *GetUserHandler) getUserByID(ctx context.Context, userID string) (*dto.GetUserOutput, error) {
	query := &command.GetUserQuery{ID: userID}
	result, err := h.getUserService.Handle(ctx, query)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, zorya.Error404NotFound("User not found")
		}

		return nil, err
	}

	response := &dto.GetUserOutput{
		Body: dto.UserResponse{
			ID:      result.User.UserID,
			Name:    result.User.Name,
			Email:   result.User.Email,
			Roles:   result.User.GetRoleNames(),
			IsAdmin: result.User.IsAdmin(),
		},
	}

	return response, nil
}
