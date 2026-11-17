package security

import "time"

// SecurityConfig represents the configuration for the security module.
type SecurityConfig struct {
	Hasher      HasherConfig      `config:"hasher"`
	JWT         JWTConfig         `config:"jwt"`
	Cookie      CookieConfig      `config:"cookie"`
	TokenSource TokenSourceConfig `config:"token_source"`
	Enforcer    EnforcerConfig    `config:"enforcer"`
}

// JWTConfig represents JWT configuration.
type JWTConfig struct {
	Secret             string        `config:"secret"`
	PrivateKeyPath     string        `config:"private_key_path"`
	PublicKeyPath      string        `config:"public_key_path"`
	Algorithm          string        `config:"algorithm"`
	AccessTokenExpiry  time.Duration `config:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `config:"refresh_token_expiry"`
}

// CookieConfig represents cookie configuration for tokens.
type CookieConfig struct {
	AccessTokenName  string `config:"access_token_name"`
	RefreshTokenName string `config:"refresh_token_name"`
	Domain           string `config:"domain"`
	Path             string `config:"path"`
	Secure           bool   `config:"secure"`
	HTTPOnly         bool   `config:"http_only"`
	SameSite         string `config:"same_site"`
}

// TokenSourceConfig represents token extraction configuration.
type TokenSourceConfig struct {
	Sources    []string `config:"sources"`
	HeaderName string   `config:"header_name"`
	CookieName string   `config:"cookie_name"`
}

// HasherConfig represents password hasher configuration.
type HasherConfig struct {
	BcryptCost int `config:"bcrypt_cost"`
	SaltLength int `config:"salt_length"`
}

// EnforcerConfig represents enforcer configuration.
type EnforcerConfig struct {
	Type string `config:"type"`
}

const (
	// algorithmHS256 names the HMAC-SHA256 signing algorithm, the default.
	algorithmHS256 = "HS256"

	// defaultAccessTokenName is the default name of the access token cookie.
	defaultAccessTokenName = "access_token"

	// tokenSourceHeader reads the token from a request header.
	tokenSourceHeader = "header"

	// tokenSourceCookie reads the token from a request cookie.
	tokenSourceCookie = "cookie"
)

// DefaultSecurityConfig returns a SecurityConfig with all default values set.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Hasher: HasherConfig{
			BcryptCost: 10,
			SaltLength: 32,
		},
		JWT: JWTConfig{
			Algorithm:          algorithmHS256,
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 168 * time.Hour, // 7 days
		},
		Cookie: CookieConfig{
			AccessTokenName:  defaultAccessTokenName,
			RefreshTokenName: "refresh_token",
			Path:             "/",
			Secure:           true,
			HTTPOnly:         true,
			SameSite:         "Lax",
		},
		TokenSource: TokenSourceConfig{
			Sources:    []string{tokenSourceHeader, tokenSourceCookie},
			HeaderName: "Authorization",
			CookieName: defaultAccessTokenName,
		},
		Enforcer: EnforcerConfig{
			Type: "simple",
		},
	}
}
