package fxsecurity

import (
	"time"

	"github.com/talav/talav/pkg/component/security"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxhttpserver"
	"go.uber.org/fx"
)

// ModuleName is the module name.
const ModuleName = "security"

// FxSecurityModule is the Fx security module.
var FxSecurityModule = fx.Module(
	ModuleName,
	fxconfig.AsConfigWithDefaults("security", security.DefaultSecurityConfig(), security.SecurityConfig{}),
	fx.Provide(
		NewSecurityFactory,
		NewPasswordHasher,
		NewJWTService,
		NewRefreshTokenService,
		NewLoginHandlerProvider,
		NewRefreshHandlerProvider,
		NewLogoutHandlerProvider,
	),
)

// NewSecurityFactory creates a new security factory.
func NewSecurityFactory() security.SecurityFactory {
	return security.NewDefaultSecurityFactory()
}

// NewPasswordHasher creates a new password hasher.
func NewPasswordHasher(cfg security.SecurityConfig, factory security.SecurityFactory) security.PasswordHasher {
	return factory.CreatePasswordHasher(cfg)
}

// NewJWTService creates a new JWT service.
func NewJWTService(cfg security.SecurityConfig, factory security.SecurityFactory) (security.JWTService, error) {
	return factory.CreateJWTService(cfg.JWT)
}

// NewRefreshTokenService creates a new refresh token service.
func NewRefreshTokenService(jwtService security.JWTService) security.RefreshTokenService {
	factory := security.NewDefaultSecurityFactory()
	// RefreshTokenStore is optional - pass nil if not using token rotation
	return factory.CreateRefreshTokenService(jwtService, nil)
}

// NewLoginHandlerProvider creates a new login handler provider for Fx.
func NewLoginHandlerProvider(
	userProvider security.UserProvider,
	hasher security.PasswordHasher,
	jwtService security.JWTService,
	refreshService security.RefreshTokenService,
	cfg security.SecurityConfig,
) *LoginHandler {
	expiry := cfg.JWT.AccessTokenExpiry
	if expiry == 0 {
		expiry = 15 * time.Minute
	}
	return NewLoginHandler(userProvider, hasher, jwtService, refreshService, cfg.Cookie, expiry)
}

// NewRefreshHandlerProvider creates a new refresh handler provider for Fx.
func NewRefreshHandlerProvider(
	jwtService security.JWTService,
	refreshService security.RefreshTokenService,
	cfg security.SecurityConfig,
) *RefreshHandler {
	expiry := cfg.JWT.AccessTokenExpiry
	if expiry == 0 {
		expiry = 15 * time.Minute
	}
	return NewRefreshHandler(jwtService, refreshService, cfg.Cookie, expiry)
}

// NewLogoutHandlerProvider creates a new logout handler provider for Fx.
func NewLogoutHandlerProvider(
	refreshService security.RefreshTokenService,
	cfg security.SecurityConfig,
) *LogoutHandler {
	return NewLogoutHandler(refreshService, cfg.Cookie)
}


