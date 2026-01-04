package security

import (
	"context"
	"net/http"
)

type contextKey string

const authUserKey contextKey = "auth_user"

// AuthUser represents authenticated user with roles from JWT.
type AuthUser struct {
	ID    string
	Roles []string
}

// SetAuthUser stores the authenticated user in the request context.
func SetAuthUser(r *http.Request, user *AuthUser) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), authUserKey, user))
}

// GetAuthUser retrieves authenticated user from request context.
// Returns nil if not authenticated.
func GetAuthUser(r *http.Request) *AuthUser {
	return GetAuthUserFromContext(r.Context())
}

// GetAuthUserFromContext retrieves authenticated user from context.
// Returns nil if not authenticated.
func GetAuthUserFromContext(ctx context.Context) *AuthUser {
	user, ok := ctx.Value(authUserKey).(*AuthUser)
	if !ok {
		return nil
	}

	return user
}
