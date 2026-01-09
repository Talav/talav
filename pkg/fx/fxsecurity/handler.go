package fxsecurity

import (
	"context"
	"net/http"
	"time"

	"github.com/talav/talav/pkg/component/security"
	"github.com/talav/talav/pkg/component/zorya"
)

// LoginRequest represents the HTTP request to authenticate a user.
type LoginRequest struct {
	Body struct {
		User     string `json:"user" validate:"required"`     // Email or username
		Password string `json:"password" validate:"required"` // User's password
	} `body:"structured"`
}

// LoginResponse represents the login response with JWT token.
type LoginResponse struct {
	Body struct {
		Token     string `json:"access_token"`
		ExpiresIn int    `json:"expires_in"` // seconds
	} `body:"structured"`
}

// RefreshRequest represents the HTTP request to refresh tokens.
type RefreshRequest struct{}

// RefreshResponse represents the refresh response with new tokens.
type RefreshResponse struct {
	Body struct {
		Token     string `json:"access_token"`
		ExpiresIn int    `json:"expires_in"` // seconds
	} `body:"structured"`
}

// LogoutRequest represents the HTTP request to logout.
type LogoutRequest struct{}

// LogoutResponse represents the logout response.
type LogoutResponse struct {
	Body struct{} `body:"structured"`
}

// LoginHandler handles HTTP requests for authentication.
type LoginHandler struct {
	userProvider      security.UserProvider
	hasher            security.PasswordHasher
	jwtService        security.JWTService
	refreshService    security.RefreshTokenService
	cookieCfg         security.CookieConfig
	accessTokenExpiry time.Duration
}

// NewLoginHandler creates a new LoginHandler instance.
func NewLoginHandler(
	userProvider security.UserProvider,
	hasher security.PasswordHasher,
	jwtService security.JWTService,
	refreshService security.RefreshTokenService,
	cookieCfg security.CookieConfig,
	accessTokenExpiry time.Duration,
) *LoginHandler {
	return &LoginHandler{
		userProvider:      userProvider,
		hasher:            hasher,
		jwtService:        jwtService,
		refreshService:    refreshService,
		cookieCfg:         cookieCfg,
		accessTokenExpiry: accessTokenExpiry,
	}
}

// Handle handles HTTP POST requests to authenticate a user.
func (h *LoginHandler) Handle(ctx context.Context, input *LoginRequest) (*LoginResponse, error) {
	// Verify user credentials
	securityUser, err := h.userProvider.GetUserByIdentifier(ctx, input.Body.User)
	if err != nil {
		// Return unauthorized for any lookup error (prevents user enumeration)
		return nil, zorya.Error401Unauthorized("invalid credentials")
	}

	if err := h.hasher.ComparePassword(securityUser.PasswordHash(), input.Body.Password, securityUser.Salt()); err != nil {
		return nil, zorya.Error401Unauthorized("invalid credentials")
	}

	// Create access token
	accessToken, err := h.jwtService.CreateAccessToken(securityUser.ID(), securityUser.Roles())
	if err != nil {
		return nil, zorya.Error500InternalServerError("failed to create access token", err)
	}

	// Create refresh token
	// Note: Refresh token is created but not yet used.
	// Cookie setting should be done in middleware/response writer wrapper when implemented.
	_, err = h.jwtService.CreateRefreshToken(securityUser.ID())
	if err != nil {
		return nil, zorya.Error500InternalServerError("failed to create refresh token", err)
	}

	expiresIn := int(h.accessTokenExpiry.Seconds())
	if expiresIn == 0 {
		expiresIn = 900 // 15 minutes default
	}

	return &LoginResponse{
		Body: struct {
			Token     string `json:"access_token"`
			ExpiresIn int    `json:"expires_in"`
		}{
			Token:     accessToken,
			ExpiresIn: expiresIn,
		},
	}, nil
}

// RefreshHandler handles HTTP POST requests to refresh access tokens.
type RefreshHandler struct {
	jwtService        security.JWTService
	refreshService    security.RefreshTokenService
	cookieCfg         security.CookieConfig
	accessTokenExpiry time.Duration
}

// NewRefreshHandler creates a new RefreshHandler instance.
func NewRefreshHandler(
	jwtService security.JWTService,
	refreshService security.RefreshTokenService,
	cookieCfg security.CookieConfig,
	accessTokenExpiry time.Duration,
) *RefreshHandler {
	return &RefreshHandler{
		jwtService:        jwtService,
		refreshService:    refreshService,
		cookieCfg:         cookieCfg,
		accessTokenExpiry: accessTokenExpiry,
	}
}

// Handle handles HTTP POST requests to refresh tokens.
func (h *RefreshHandler) Handle(ctx context.Context, input *RefreshRequest) (*RefreshResponse, error) {
	// Extract refresh token from cookie
	// Note: In a real implementation, you'd extract the request from context
	// For now, this is a placeholder that needs to be adapted to extract cookies from request
	// The refresh token should be extracted from the httpOnly cookie set during login

	// This requires access to *http.Request which should be passed via context or handler signature
	// For Zorya, we may need to use a custom handler wrapper or middleware

	return nil, zorya.Error401Unauthorized("refresh token required - implementation needs request access")
}

// LogoutHandler handles HTTP POST requests to logout.
type LogoutHandler struct {
	refreshService security.RefreshTokenService
	cookieCfg      security.CookieConfig
}

// NewLogoutHandler creates a new LogoutHandler instance.
func NewLogoutHandler(
	refreshService security.RefreshTokenService,
	cookieCfg security.CookieConfig,
) *LogoutHandler {
	return &LogoutHandler{
		refreshService: refreshService,
		cookieCfg:      cookieCfg,
	}
}

// Handle handles HTTP POST requests to logout.
func (h *LogoutHandler) Handle(ctx context.Context, input *LogoutRequest) (*LogoutResponse, error) {
	// Clear cookies and optionally revoke refresh token
	// Cookie clearing should be done in response writer wrapper
	return &LogoutResponse{}, nil
}

// SetCookie sets a cookie on the response writer.
func SetCookie(w http.ResponseWriter, name, value string, cfg security.CookieConfig, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		Secure:   cfg.Secure,
		HttpOnly: cfg.HTTPOnly,
		MaxAge:   maxAge,
	}

	switch cfg.SameSite {
	case "Strict":
		cookie.SameSite = http.SameSiteStrictMode
	case "Lax":
		cookie.SameSite = http.SameSiteLaxMode
	case "None":
		cookie.SameSite = http.SameSiteNoneMode
	default:
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)
}

// ClearCookie clears a cookie by setting it to expire immediately.
func ClearCookie(w http.ResponseWriter, name string, cfg security.CookieConfig) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		Secure:   cfg.Secure,
		HttpOnly: cfg.HTTPOnly,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}
