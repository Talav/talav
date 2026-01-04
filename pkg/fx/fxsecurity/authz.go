package fxsecurity

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/talav/talav/pkg/component/security"
)

// AuthZMiddleware provides RBAC authorization middleware.
type AuthZMiddleware struct {
	enforcer *casbin.Enforcer
}

// NewAuthZMiddleware creates a new AuthZMiddleware.
func NewAuthZMiddleware(enforcer *casbin.Enforcer) *AuthZMiddleware {
	return &AuthZMiddleware{enforcer: enforcer}
}

// EnforceAccess creates middleware that enforces access to a specific resource.
func (m *AuthZMiddleware) EnforceAccess(resourceID string, action string) func(http.Handler) http.Handler {
	return m.enforce(action, func(*http.Request) string { return resourceID })
}

// EnforceAccessFromPath creates middleware that enforces access using a path parameter as resource ID.
func (m *AuthZMiddleware) EnforceAccessFromPath(pathParam string, action string) func(http.Handler) http.Handler {
	return m.enforce(action, func(r *http.Request) string {
		// Chi router stores path params in context
		return r.PathValue(pathParam)
	})
}

// enforce creates middleware that checks permissions via Casbin.
func (m *AuthZMiddleware) enforce(action string, getResourceID func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resourceID := getResourceID(r)
			if resourceID == "" {
				http.Error(w, "missing resource identifier", http.StatusBadRequest)
				return
			}

			user := security.GetAuthUser(r)
			if user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check permissions for each role
			for _, role := range user.Roles {
				ok, err := m.enforcer.Enforce(user.ID, role, resourceID, action)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				if ok {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}

