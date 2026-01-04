package fxsecurity

import (
	"net/http"

	"github.com/talav/talav/pkg/component/security"
	"github.com/talav/talav/pkg/fx/fxhttpserver"
	"go.uber.org/fx"
)

// AsJWTMiddleware registers the JWT authentication middleware.
// Call this at the app level to register the middleware.
func AsJWTMiddleware(jwtService security.JWTService, cfg security.SecurityConfig) fx.Option {
	return fxhttpserver.AsMiddleware(
		JWTAuthMiddleware(jwtService, cfg),
		fxhttpserver.PriorityBeforeZorya-20, // 230, after HTTPLog, before business logic
		"jwt-auth",
	)
}

// AsAuthZMiddleware registers an RBAC authorization middleware.
func AsAuthZMiddleware(middleware func(http.Handler) http.Handler, priority int, name string) fx.Option {
	return fxhttpserver.AsMiddleware(middleware, priority, name)
}
