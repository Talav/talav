package handler

import (
	"context"

	"github.com/talav/talav/pkg/component/user/app/command"
	"github.com/talav/talav/pkg/module/user/dto"
)

// ListUsersHandler handles HTTP requests for listing users.
type ListUsersHandler struct {
	listUsersService *command.ListUsersQueryHandler
}

// NewListUsersHandler creates a new ListUsersHandler instance.
func NewListUsersHandler(listUsersService *command.ListUsersQueryHandler) *ListUsersHandler {
	return &ListUsersHandler{
		listUsersService: listUsersService,
	}
}

// Handle lists users with pagination and filtering.
func (h *ListUsersHandler) Handle(ctx context.Context, input *dto.ListUsersInput) (*dto.ListUsersOutput, error) {
	// Set default limit if not provided
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Convert DTO to Query
	query := &command.ListUsersQuery{
		Cursor:      input.Cursor,
		Limit:       limit,
		EmailFilter: input.Email,
		NameFilter:  input.Name,
	}

	// Execute query
	result, err := h.listUsersService.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	// Convert result to response
	users := make([]dto.UserResponse, len(result.Users))
	for i, user := range result.Users {
		users[i] = dto.UserResponse{
			ID:      user.UserID,
			Name:    user.Name,
			Email:   user.Email,
			Roles:   user.GetRoleNames(),
			IsAdmin: user.IsAdmin(),
		}
	}

	response := &dto.ListUsersOutput{
		Body: dto.ListUsersResponse{
			Users:      users,
			NextCursor: result.NextCursor,
			HasMore:    result.HasMore,
		},
	}

	return response, nil
}
