package securityhttp

import (
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/module/securityhttp/handler"
)

// RegisterRoutes registers all authentication-related routes.
// Called automatically by the FX module during application startup.
func RegisterRoutes(
	api zorya.API,
	loginHandler *handler.LoginHandler,
	logoutHandler *handler.LogoutHandler,
) {
	// Auth group for authentication-related endpoints
	authGroup := zorya.NewGroup(api, "/auth")

	// POST /auth/login - User login
	zorya.Post(authGroup, "/login", loginHandler.Handle,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "User Login",
				Description: "Authenticate a user with email and password to receive a JWT token",
				Tags:        []string{"Authentication"},
				OperationID: "login",
			}
		},
	)

	// POST /auth/logout - User logout (requires authentication)
	zorya.Post(authGroup, "/logout", logoutHandler.Handle,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "User Logout",
				Description: "Revoke refresh tokens and clear authentication",
				Tags:        []string{"Authentication"},
				OperationID: "logout",
			}
		},
		zorya.Secure(
			func(s *zorya.RouteSecurity) {
				// Require any authenticated user (no specific roles needed)
				s.Roles = []string{} // Empty roles means just authentication required
			},
		),
	)
}
