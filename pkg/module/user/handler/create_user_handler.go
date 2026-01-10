package handler

import (
	"context"

	"github.com/talav/talav/pkg/component/user/app/command"
	"github.com/talav/talav/pkg/module/user/dto"
)

// CreateUserHandler handles HTTP requests for user creation.
type CreateUserHandler struct {
	createUserService *command.CreateUserHandler
}

// NewCreateUserHandler creates a new CreateUserHandler instance.
func NewCreateUserHandler(createUserService *command.CreateUserHandler) *CreateUserHandler {
	return &CreateUserHandler{
		createUserService: createUserService,
	}
}

// Handle creates a new user.
func (h *CreateUserHandler) Handle(ctx context.Context, req *dto.CreateUserRequest) (*dto.CreateUserOutput, error) {
	// Convert DTO to Command
	cmd := &command.CreateUserCommand{
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		RoleNames: req.Roles,
	}

	// Execute command
	result, err := h.createUserService.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Convert result to response
	response := dto.CreateUserOutput{
		Status: 201,
		Body: dto.UserResponse{
			ID:      result.User.UserID,
			Name:    result.User.Name,
			Email:   result.User.Email,
			Roles:   result.User.GetRoleNames(),
			IsAdmin: result.User.IsAdmin(),
		},
	}

	return &response, nil
}
