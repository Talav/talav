package fxsecurity

import (
	"fmt"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
	"github.com/talav/talav/pkg/component/security"
	"github.com/talav/talav/pkg/component/security/adapter/zorya"
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
	fxconfig.AsConfigWithDefaults("casbin", CasbinConfig{}, CasbinConfig{}),
	fx.Provide(
		NewPasswordHasher,
		NewJWTService,
		NewRefreshTokenService,
		NewLoginHandlerProvider,
		NewRefreshHandlerProvider,
		NewLogoutHandlerProvider,
		NewCasbinEnforcerFromAdapter, // Provides *casbin.Enforcer when adapter is provided
		NewSecurityEnforcer,
		RegisterJWTMiddleware,
		RegisterSecurityMiddleware,
	),
)

// EnforcerParams holds dependencies for enforcer creation.
type EnforcerParams struct {
	fx.In
	Config         security.SecurityConfig
	CasbinEnforcer *casbin.Enforcer          `optional:"true"` // Only needed if type is "casbin"
	CustomEnforcer security.SecurityEnforcer `optional:"true"` // Only needed if type is "custom"
}

// NewPasswordHasher creates a new password hasher.
func NewPasswordHasher(cfg security.SecurityConfig) security.PasswordHasher {
	return security.NewPasswordHasher(cfg.Hasher)
}

// NewJWTService creates a new JWT service.
func NewJWTService(cfg security.SecurityConfig) (security.JWTService, error) {
	return security.NewJWTService(cfg.JWT)
}

// NewRefreshTokenService creates a new refresh token service.
// RefreshTokenStore is optional - pass nil if not using token rotation.
func NewRefreshTokenService(jwtService security.JWTService) security.RefreshTokenService {
	return security.NewRefreshTokenService(jwtService, nil)
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

// AsJWTMiddleware registers the JWT authentication middleware with dependency injection.
// Dependencies (jwtService, cfg) are automatically injected by Fx.
//
// Example:
//
//	fx.New(
//		fxsecurity.FxSecurityModule,
//		fxsecurity.AsJWTMiddleware(),
//	)
func RegisterJWTMiddleware() fx.Option {
	return fxhttpserver.AsMiddlewareConstructor(
		security.NewJWTAuthMiddleware,
		fxhttpserver.PriorityBeforeZorya-20,
		"jwt-middleware",
	)
}

// AsSecurityMiddleware registers security enforcement middleware.
// Uses SecurityEnforcer provided by NewSecurityEnforcer (config-based).
// Enforcer type is determined by security.enforcer.type config:
// - "simple": Uses SimpleEnforcer (default)
// - "custom": Uses user-provided SecurityEnforcer.
func RegisterSecurityMiddleware() fx.Option {
	return fxhttpserver.AsMiddlewareConstructor(
		zorya.NewEnforcementMiddleware,
		fxhttpserver.PriorityBeforeZorya-10,
		"security-middleware",
	)
}

// NewSecurityEnforcer creates a SecurityEnforcer based on config.
//
// The enforcer type is determined by security.enforcer.type configuration value.
// This function is automatically called by FxSecurityModule and provides a SecurityEnforcer
// that is injected into AsSecurityMiddleware().
//
// How it works:
//  1. Reads security.enforcer.type from SecurityConfig (provided by fxconfig)
//  2. Uses Fx's optional dependency injection to conditionally require dependencies:
//     - CasbinEnforcer is optional and only required when type is "casbin"
//     - CustomEnforcer is optional and only required when type is "custom"
//  3. Returns the appropriate enforcer based on config type:
//     - "simple" (default): Returns SimpleEnforcer - no dependencies needed
//     - "casbin": Returns CasbinEnforcer - requires Casbin enforcer (provided by NewCasbinEnforcerFromAdapter)
//     - "custom": Returns user-provided SecurityEnforcer - user must provide via fx.Provide()
//
// Usage examples:
//
//	// Simple enforcer (default, no config needed)
//	fx.New(
//		fxsecurity.FxSecurityModule,
//		fxsecurity.AsSecurityMiddleware(),
//	)
//
//	// Casbin enforcer
//	// import "github.com/casbin/gorm-adapter/v3"
//	// import "github.com/casbin/casbin/v2/persist"
//	fx.New(
//		fxsecurity.FxSecurityModule,
//		fx.Provide(func(db *gorm.DB) (persist.Adapter, error) {
//			return gormadapter.NewAdapterByDB(db)
//		}),
//		fxsecurity.AsSecurityMiddleware(),
//	)
//	// config.yaml:
//	//   security.enforcer.type: "casbin"
//	//   casbin.model_path: "configs/casbin.conf"
//
//	// Custom enforcer (user provides implementation)
//	fx.New(
//		fxsecurity.FxSecurityModule,
//		fx.Provide(func() security.SecurityEnforcer {
//			return &MyCustomEnforcer{}
//		}),
//		fxsecurity.AsSecurityMiddleware(),
//	)
//	// config.yaml: security.enforcer.type: "custom"
func NewSecurityEnforcer(p EnforcerParams) (security.SecurityEnforcer, error) {
	switch p.Config.Enforcer.Type {
	case "casbin":
		if p.CasbinEnforcer == nil {
			return nil, fmt.Errorf("casbin enforcer required when enforcer.type is 'casbin' - provide persist.Adapter and casbin config")
		}

		return NewCasbinEnforcer(p.CasbinEnforcer), nil

	case "custom":
		if p.CustomEnforcer == nil {
			return nil, fmt.Errorf("custom enforcer required when enforcer.type is 'custom' - provide SecurityEnforcer")
		}

		return p.CustomEnforcer, nil

	default: // "simple" or empty
		return security.NewSimpleEnforcer(), nil
	}
}

// NewCasbinEnforcerFromAdapter creates a new Casbin enforcer with the provided adapter.
// This is used when enforcer.type is "casbin".
// Requires both persist.Adapter and casbin.model_path configuration.
func NewCasbinEnforcerFromAdapter(adapter persist.Adapter, cfg CasbinConfig) (*casbin.Enforcer, error) {
	if cfg.ModelPath == "" {
		return nil, fmt.Errorf("casbin.model_path is required when security.enforcer.type is 'casbin'")
	}

	return casbin.NewEnforcer(cfg.ModelPath, adapter)
}
