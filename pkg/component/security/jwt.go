package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
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
	cfg            JWTConfig
	signingKey     interface{}
	verificationKey interface{}
}

// NewJWTService creates a new JWTService with the given configuration.
func NewJWTService(cfg JWTConfig) (JWTService, error) {
	service := &DefaultJWTService{
		cfg: cfg,
	}

	if cfg.Algorithm == "" {
		cfg.Algorithm = "HS256"
	}

	var err error
	switch cfg.Algorithm {
	case "HS256":
		if cfg.Secret == "" {
			return nil, fmt.Errorf("HS256 algorithm requires secret to be set")
		}
		service.signingKey = []byte(cfg.Secret)
		service.verificationKey = []byte(cfg.Secret)
	case "RS256":
		if cfg.PrivateKeyPath == "" || cfg.PublicKeyPath == "" {
			return nil, fmt.Errorf("RS256 algorithm requires both private_key_path and public_key_path to be set")
		}
		service.signingKey, err = loadRSAPrivateKey(cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
		service.verificationKey, err = loadRSAPublicKey(cfg.PublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s (supported: HS256, RS256)", cfg.Algorithm)
	}

	return service, nil
}

// CreateAccessToken creates a new access token for the given user.
func (s *DefaultJWTService) CreateAccessToken(userID string, roles []string) (string, error) {
	expiry := s.cfg.AccessTokenExpiry
	if expiry == 0 {
		expiry = 15 * time.Minute
	}

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

	var method jwt.SigningMethod
	switch s.cfg.Algorithm {
	case "HS256":
		method = jwt.SigningMethodHS256
	case "RS256":
		method = jwt.SigningMethodRS256
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", s.cfg.Algorithm)
	}

	token := jwt.NewWithClaims(method, claims)
	return token.SignedString(s.signingKey)
}

// CreateRefreshToken creates a new refresh token for the given user.
func (s *DefaultJWTService) CreateRefreshToken(userID string) (string, error) {
	expiry := s.cfg.RefreshTokenExpiry
	if expiry == 0 {
		expiry = 168 * time.Hour // 7 days
	}

	// Generate unique token ID for rotation support
	tokenID, err := generateTokenID()
	if err != nil {
		return "", fmt.Errorf("failed to generate token ID: %w", err)
	}

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

	var method jwt.SigningMethod
	switch s.cfg.Algorithm {
	case "HS256":
		method = jwt.SigningMethodHS256
	case "RS256":
		method = jwt.SigningMethodRS256
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", s.cfg.Algorithm)
	}

	token := jwt.NewWithClaims(method, claims)
	return token.SignedString(s.signingKey)
}

// ValidateAccessToken validates and parses an access token.
func (s *DefaultJWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != s.getExpectedAlgorithm() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.verificationKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	return claims, nil
}

// ValidateRefreshToken validates and parses a refresh token.
func (s *DefaultJWTService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != s.getExpectedAlgorithm() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.verificationKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	return claims, nil
}

func (s *DefaultJWTService) getExpectedAlgorithm() string {
	switch s.cfg.Algorithm {
	case "HS256":
		return "HS256"
	case "RS256":
		return "RS256"
	default:
		return "HS256"
	}
}

func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		keyPKCS8, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key: %w (tried PKCS1 and PKCS8)", err)
		}
		rsaKey, ok := keyPKCS8.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
		return rsaKey, nil
	}

	return key, nil
}

func loadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format
		keyPKCS1, err2 := x509.ParsePKCS1PublicKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse public key: %w (tried PKIX and PKCS1)", err)
		}
		return keyPKCS1, nil
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	return rsaKey, nil
}

// generateTokenID generates a unique token ID for refresh tokens.
func generateTokenID() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

