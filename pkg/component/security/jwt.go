package security

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims embeds RegisteredClaims and adds authentication-specific fields.
type Claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

// RefreshClaims represents refresh token claims (minimal, only user ID).
type RefreshClaims struct {
	jwt.RegisteredClaims
}

// JWTService provides JWT token creation and validation.
type JWTService interface {
	CreateAccessToken(userID string, roles []string) (string, error)
	CreateRefreshToken(userID string) (string, error)
	ValidateAccessToken(token string) (*Claims, error)
	ValidateRefreshToken(token string) (*RefreshClaims, error)
}

// DefaultJWTService is the default implementation of JWTService.
type DefaultJWTService struct {
	cfg      JWTConfig
	strategy SigningStrategy
}

// SigningStrategy defines the interface for JWT signing and verification strategies.
type SigningStrategy interface {
	// Sign creates a token string from JWT claims.
	Sign(claims jwt.Claims) (string, error)
	// Verify parses and validates a token string, returning the claims.
	Verify(tokenString string, claims jwt.Claims) error
	// Algorithm returns the algorithm name.
	Algorithm() string
}

// HS256Strategy implements JWT signing with HMAC-SHA256 (symmetric key).
type HS256Strategy struct {
	secret []byte
}

// RS256Strategy implements JWT signing with RSA-SHA256 (asymmetric keys).
type RS256Strategy struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewJWTService creates a new JWTService with the given configuration.
func NewJWTService(cfg JWTConfig) (JWTService, error) {
	var strategy SigningStrategy
	var err error

	switch cfg.Algorithm {
	case "HS256":
		strategy, err = NewHS256Strategy(cfg.Secret)
	case "RS256":
		strategy, err = NewRS256Strategy(cfg.PrivateKeyPath, cfg.PublicKeyPath)
	default:
		err = fmt.Errorf("unsupported algorithm: %s (supported: HS256, RS256)", cfg.Algorithm)
	}

	if err != nil {
		return nil, err
	}

	return &DefaultJWTService{
		cfg:      cfg,
		strategy: strategy,
	}, nil
}

// NewHS256Strategy creates a new HS256 signing strategy.
func NewHS256Strategy(secret string) (*HS256Strategy, error) {
	if secret == "" {
		return nil, fmt.Errorf("HS256 algorithm requires secret to be set")
	}

	return &HS256Strategy{
		secret: []byte(secret),
	}, nil
}

// NewRS256Strategy creates a new RS256 signing strategy.
func NewRS256Strategy(privateKeyPath, publicKeyPath string) (*RS256Strategy, error) {
	if privateKeyPath == "" || publicKeyPath == "" {
		return nil, fmt.Errorf("RS256 algorithm requires both private_key_path and public_key_path to be set")
	}

	privateKey, err := loadRSAPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey, err := loadRSAPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &RS256Strategy{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// CreateAccessToken creates a new access token for the given user.
func (s *DefaultJWTService) CreateAccessToken(userID string, roles []string) (string, error) {
	expiry := s.cfg.AccessTokenExpiry
	now := time.Now()
	claims := &Claims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	return s.strategy.Sign(claims)
}

// CreateRefreshToken creates a new refresh token for the given user.
func (s *DefaultJWTService) CreateRefreshToken(userID string) (string, error) {
	expiry := s.cfg.RefreshTokenExpiry
	// Generate unique token ID for rotation support
	tokenID := generateTokenID()

	now := time.Now()
	claims := &RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	return s.strategy.Sign(claims)
}

// ValidateAccessToken validates an access token and returns its claims.
func (s *DefaultJWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	if err := s.strategy.Verify(tokenString, claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns its claims.
func (s *DefaultJWTService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	if err := s.strategy.Verify(tokenString, claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// Sign creates a token string using HMAC-SHA256.
func (s *HS256Strategy) Sign(claims jwt.Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

// Verify parses and validates a token using HMAC-SHA256.
func (s *HS256Strategy) Verify(tokenString string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.secret, nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

// Algorithm returns the algorithm name.
func (s *HS256Strategy) Algorithm() string {
	return "HS256"
}

// Sign creates a token string using RSA-SHA256.
func (s *RS256Strategy) Sign(claims jwt.Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.privateKey)
}

// Verify parses and validates a token using RSA-SHA256.
func (s *RS256Strategy) Verify(tokenString string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.publicKey, nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

// Algorithm returns the algorithm name.
func (s *RS256Strategy) Algorithm() string {
	return "RS256"
}

// generateTokenID generates a unique token ID for refresh tokens.
func generateTokenID() string {
	bytes := make([]byte, 16) // 128 bits
	_, _ = rand.Read(bytes)

	return base64.URLEncoding.EncodeToString(bytes)
}

// loadRSAPrivateKey loads an RSA private key from a PEM file.
func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path) //nolint:gosec // path is from trusted config
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}

// loadRSAPublicKey loads an RSA public key from a PEM file.
func loadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path) //nolint:gosec // path is from trusted config
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return publicKey, nil
}
