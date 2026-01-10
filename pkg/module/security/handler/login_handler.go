package handler

import (
	"context"

	"github.com/talav/talav/pkg/component/security"
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/module/security/dto"
)

// LoginHandler handles HTTP requests for authentication.
type LoginHandler struct {
	userProvider security.UserProvider
	hasher       security.PasswordHasher
	jwtService   security.JWTService
}

// NewLoginHandler creates a new LoginHandler instance.
func NewLoginHandler(
	userProvider security.UserProvider,
	hasher security.PasswordHasher,
	jwtService security.JWTService,
) *LoginHandler {
	return &LoginHandler{
		userProvider: userProvider,
		hasher:       hasher,
		jwtService:   jwtService,
	}
}

// Handle authenticates a user and returns a JWT token.
// Returns 401 Unauthorized for invalid credentials.
func (h *LoginHandler) Handle(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Verify user credentials
	securityUser, err := h.userProvider.GetUserByIdentifier(ctx, req.Email)
	if err != nil {
		// Return unauthorized for any lookup error (user not found or other errors).
		// This prevents user enumeration by not distinguishing between different error types.
		return nil, zorya.Error401Unauthorized("Invalid credentials")
	}

	if err := h.hasher.ComparePassword(securityUser.PasswordHash(), req.Password, securityUser.Salt()); err != nil {
		return nil, zorya.Error401Unauthorized("Invalid credentials")
	}

	// Generate JWT token
	token, err := h.jwtService.CreateAccessToken(securityUser.ID(), securityUser.Roles())
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{Token: token}, nil
}
