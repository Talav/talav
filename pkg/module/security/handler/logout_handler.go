package handler

import (
	"context"

	"github.com/talav/talav/pkg/component/security"
)

// LogoutHandler handles HTTP requests for logout.
type LogoutHandler struct {
	refreshService security.RefreshTokenService
}

// NewLogoutHandler creates a new LogoutHandler instance.
func NewLogoutHandler(refreshService security.RefreshTokenService) *LogoutHandler {
	return &LogoutHandler{
		refreshService: refreshService,
	}
}

// LogoutResponse represents an empty logout response.
type LogoutResponse struct{}

// Handle revokes the user's refresh token and logs them out.
// Requires authentication.
func (h *LogoutHandler) Handle(ctx context.Context, req *struct{}) (*LogoutResponse, error) {
	// Get authenticated user from context
	user := security.GetAuthUserFromContext(ctx)
	if user == nil {
		// This should never happen if Secure() middleware is applied
		return &LogoutResponse{}, nil
	}

	// Revoke all refresh tokens for this user
	// Ignore errors - even if revocation fails, we've removed client-side tokens
	_ = h.refreshService.RevokeAllRefreshTokens(ctx, user.ID)

	return &LogoutResponse{}, nil
}
