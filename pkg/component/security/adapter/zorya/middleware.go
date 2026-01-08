package zorya

import (
	"net/http"

	"github.com/talav/talav/pkg/component/security"
	zoryapkg "github.com/talav/talav/pkg/component/zorya"
)

// Option configures the enforcement middleware.
type Option func(*config)

type config struct {
	onUnauthorized func(w http.ResponseWriter, r *http.Request)
	onForbidden    func(w http.ResponseWriter, r *http.Request)
}

// WithUnauthorizedHandler sets the handler for unauthorized requests (401).
// Called when a protected route requires authentication but no user is present.
func WithUnauthorizedHandler(handler func(w http.ResponseWriter, r *http.Request)) Option {
	return func(cfg *config) {
		cfg.onUnauthorized = handler
	}
}

// WithForbiddenHandler sets the handler for forbidden requests (403).
// Called when an authenticated user doesn't meet the security requirements.
func WithForbiddenHandler(handler func(w http.ResponseWriter, r *http.Request)) Option {
	return func(cfg *config) {
		cfg.onForbidden = handler
	}
}

// NewEnforcementMiddleware creates middleware that enforces security requirements
// defined on Zorya routes using the Secure() wrapper.
//
// This adapter connects Zorya's routing framework with the generic security component:
// 1. Reads RouteSecurityContext metadata from context (set by Zorya)
// 2. Converts it to generic SecurityRequirements
// 3. Calls the SecurityEnforcer to make authorization decisions
//
// Usage:
//
//	// Manual setup
//	enforcer := security.NewSimpleEnforcer()
//	api.UseMiddleware(zorya.NewEnforcementMiddleware(enforcer))
//
//	// With custom handlers
//	api.UseMiddleware(zorya.NewEnforcementMiddleware(
//	    enforcer,
//	    zorya.WithUnauthorizedHandler(func(w, r) {
//	        w.Header().Set("Content-Type", "application/json")
//	        w.WriteHeader(401)
//	        json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
//	    }),
//	    zorya.WithForbiddenHandler(func(w, r) {
//	        w.Header().Set("Content-Type", "application/json")
//	        w.WriteHeader(403)
//	        json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
//	    }),
//	))
//
//	// Register protected routes
//	zorya.Get(api, "/admin/dashboard", handler,
//	    zorya.Secure(
//	        zorya.Roles("admin"),
//	    ),
//	)
func NewEnforcementMiddleware(
	enforcer security.SecurityEnforcer,
	opts ...Option,
) func(http.Handler) http.Handler {
	// Configure with defaults
	cfg := &config{
		onUnauthorized: func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		},
		onForbidden: func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read security metadata set by Zorya
			zoryaMeta := zoryapkg.GetRouteSecurityContext(r)

			// No security requirements = public route
			if zoryaMeta == nil {
				next.ServeHTTP(w, r)

				return
			}

			// Get authenticated user
			user := security.GetAuthUser(r)

			// Authentication check: no user + requirements exist = unauthorized (401)
			// If zoryaMeta exists, it is guaranteed to have at least one requirement (validated by Secure()).
			if user == nil {
				cfg.onUnauthorized(w, r)

				return
			}

			// Convert Zorya metadata to generic security requirements
			requirements := &security.SecurityRequirements{
				Roles:       zoryaMeta.Roles,
				Permissions: zoryaMeta.Permissions,
				Resource:    zoryaMeta.Resource,
				Action:      zoryaMeta.Action,
			}

			// Authorization check: user exists, check permissions via enforcer
			ok, err := enforcer.Enforce(r.Context(), user, requirements)
			if err != nil {
				// Enforcer error = internal error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}
			if !ok {
				// User authenticated but doesn't have permission = forbidden (403)
				cfg.onForbidden(w, r)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
