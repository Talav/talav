package security

// SecurityFactory is the interface for security component factories.
type SecurityFactory interface {
	CreatePasswordHasher(cfg SecurityConfig) PasswordHasher
	CreateJWTService(cfg JWTConfig) (JWTService, error)
	CreateRefreshTokenService(jwtService JWTService, store RefreshTokenStore) RefreshTokenService
}

// DefaultSecurityFactory is the default SecurityFactory implementation.
type DefaultSecurityFactory struct{}

// NewDefaultSecurityFactory returns a DefaultSecurityFactory, implementing SecurityFactory.
func NewDefaultSecurityFactory() SecurityFactory {
	return &DefaultSecurityFactory{}
}

// CreatePasswordHasher creates a new PasswordHasher with the given configuration.
func (f *DefaultSecurityFactory) CreatePasswordHasher(cfg SecurityConfig) PasswordHasher {
	return NewPasswordHasher(cfg)
}

// CreateJWTService creates a new JWTService with the given configuration.
func (f *DefaultSecurityFactory) CreateJWTService(cfg JWTConfig) (JWTService, error) {
	return NewJWTService(cfg)
}

// CreateRefreshTokenService creates a new RefreshTokenService.
func (f *DefaultSecurityFactory) CreateRefreshTokenService(jwtService JWTService, store RefreshTokenStore) RefreshTokenService {
	return NewRefreshTokenService(jwtService, store)
}

