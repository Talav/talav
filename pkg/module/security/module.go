package security

import (
	"github.com/talav/talav/pkg/module/security/handler"
	"go.uber.org/fx"
)

// ModuleName is the module name.
const ModuleName = "security"

// Module provides HTTP endpoints for authentication.
//
// Dependencies:
//   - security.UserProvider - for user lookup
//   - security.PasswordHasher - for password verification
//   - security.JWTService - for token generation
//   - security.RefreshTokenService - for token revocation
//   - zorya.API - for route registration
//
// Routes:
//   - POST /auth/login - Authenticate user
//   - POST /auth/logout - Revoke tokens (requires auth)
//
// Usage:
//
//	fx.New(
//		fxconfig.FxConfigModule,
//		fxhttpserver.FxHTTPServerModule,
//		fxsecurity.FxSecurityModule,
//		securityhttp.Module,
//	)
var Module = fx.Module(
	ModuleName,
	fx.Provide(
		handler.NewLoginHandler,
		handler.NewLogoutHandler,
	),
	fx.Invoke(RegisterRoutes),
)
