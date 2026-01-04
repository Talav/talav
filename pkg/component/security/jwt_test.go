package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_RoundTrip_HS256(t *testing.T) {
	// Setup
	jwtService, err := NewJWTService(JWTConfig{
		Algorithm:          "HS256",
		Secret:             "test-secret-key-12345",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 168 * time.Hour,
	})
	require.NoError(t, err, "NewJWTService should not return an error")

	userID := "user-123"
	roles := []string{"admin", "user"}

	// Action - Create access token
	accessToken, err := jwtService.CreateAccessToken(userID, roles)
	require.NoError(t, err, "CreateAccessToken should not return an error")
	require.NotEmpty(t, accessToken, "Access token should not be empty")

	// Action - Validate access token
	claims, err := jwtService.ValidateAccessToken(accessToken)
	require.NoError(t, err, "ValidateAccessToken should not return an error")
	require.NotNil(t, claims, "Claims should not be nil")

	// Assertion
	assert.Equal(t, userID, claims.Subject, "User ID should match")
	assert.Equal(t, roles, claims.Roles, "Roles should match")
	assert.NotNil(t, claims.ExpiresAt, "ExpiresAt should be set")
	assert.NotNil(t, claims.IssuedAt, "IssuedAt should be set")

	// Action - Create refresh token
	refreshToken, err := jwtService.CreateRefreshToken(userID)
	require.NoError(t, err, "CreateRefreshToken should not return an error")
	require.NotEmpty(t, refreshToken, "Refresh token should not be empty")

	// Action - Validate refresh token
	refreshClaims, err := jwtService.ValidateRefreshToken(refreshToken)
	require.NoError(t, err, "ValidateRefreshToken should not return an error")
	require.NotNil(t, refreshClaims, "Refresh claims should not be nil")

	// Assertion
	assert.Equal(t, userID, refreshClaims.Subject, "User ID should match")
	assert.NotEmpty(t, refreshClaims.ID, "Token ID should be set")
	assert.NotNil(t, refreshClaims.ExpiresAt, "ExpiresAt should be set")
}

func TestJWTService_RoundTrip_RS256(t *testing.T) {
	// Setup - Generate temporary RSA keys
	privateKeyPath, publicKeyPath := createTempRSAKeys(t)
	defer func() {
		removeFile(privateKeyPath)
		removeFile(publicKeyPath)
	}()

	jwtService, err := NewJWTService(JWTConfig{
		Algorithm:          "RS256",
		PrivateKeyPath:     privateKeyPath,
		PublicKeyPath:      publicKeyPath,
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 168 * time.Hour,
	})
	require.NoError(t, err, "NewJWTService should not return an error")

	userID := "user-456"
	roles := []string{"editor", "viewer"}

	// Action - Create access token
	accessToken, err := jwtService.CreateAccessToken(userID, roles)
	require.NoError(t, err, "CreateAccessToken should not return an error")
	require.NotEmpty(t, accessToken, "Access token should not be empty")

	// Action - Validate access token
	claims, err := jwtService.ValidateAccessToken(accessToken)
	require.NoError(t, err, "ValidateAccessToken should not return an error")
	require.NotNil(t, claims, "Claims should not be nil")

	// Assertion
	assert.Equal(t, userID, claims.Subject, "User ID should match")
	assert.Equal(t, roles, claims.Roles, "Roles should match")
	assert.NotNil(t, claims.ExpiresAt, "ExpiresAt should be set")
	assert.NotNil(t, claims.IssuedAt, "IssuedAt should be set")

	// Action - Create refresh token
	refreshToken, err := jwtService.CreateRefreshToken(userID)
	require.NoError(t, err, "CreateRefreshToken should not return an error")
	require.NotEmpty(t, refreshToken, "Refresh token should not be empty")

	// Action - Validate refresh token
	refreshClaims, err := jwtService.ValidateRefreshToken(refreshToken)
	require.NoError(t, err, "ValidateRefreshToken should not return an error")
	require.NotNil(t, refreshClaims, "Refresh claims should not be nil")

	// Assertion
	assert.Equal(t, userID, refreshClaims.Subject, "User ID should match")
	assert.NotEmpty(t, refreshClaims.ID, "Token ID should be set")
	assert.NotNil(t, refreshClaims.ExpiresAt, "ExpiresAt should be set")
}

func TestJWTService_ValidateAccessToken_Expired(t *testing.T) {
	// Setup
	jwtService, err := NewJWTService(JWTConfig{
		Algorithm:         "HS256",
		Secret:            "test-secret-key-12345",
		AccessTokenExpiry: 100 * time.Millisecond, // Very short expiry
	})
	require.NoError(t, err, "NewJWTService should not return an error")

	userID := "user-789"
	roles := []string{"user"}

	// Action - Create access token
	accessToken, err := jwtService.CreateAccessToken(userID, roles)
	require.NoError(t, err, "CreateAccessToken should not return an error")

	// Wait for token to expire
	time.Sleep(150 * time.Millisecond)

	// Action - Validate expired token
	claims, err := jwtService.ValidateAccessToken(accessToken)

	// Assertion
	assert.Error(t, err, "ValidateAccessToken should return an error for expired token")
	assert.Nil(t, claims, "Claims should be nil when validation fails")
	assert.Contains(t, err.Error(), "expired", "Error should indicate token is expired")
}

func TestJWTService_ValidateAccessToken_InvalidToken(t *testing.T) {
	// Setup
	jwtService, err := NewJWTService(JWTConfig{
		Algorithm:         "HS256",
		Secret:            "test-secret-key-12345",
		AccessTokenExpiry: 15 * time.Minute,
	})
	require.NoError(t, err, "NewJWTService should not return an error")

	tests := []struct {
		name        string
		tokenString string
		description string
	}{
		{
			name:        "empty_token",
			tokenString: "",
			description: "Empty token string",
		},
		{
			name:        "malformed_token",
			tokenString: "not.a.valid.jwt.token",
			description: "Malformed token format",
		},
		{
			name:        "invalid_format",
			tokenString: "header.payload",
			description: "Token missing signature",
		},
		{
			name:        "tampered_signature",
			tokenString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyIsInJvbGVzIjpbInVzZXIiXX0.tampered-signature",
			description: "Token with tampered signature",
		},
		{
			name:        "wrong_algorithm",
			tokenString: createTokenWithWrongAlgorithm(t),
			description: "Token signed with different algorithm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Action - Validate invalid token
			claims, err := jwtService.ValidateAccessToken(tt.tokenString)

			// Assertion
			assert.Error(t, err, "ValidateAccessToken should return an error for %s", tt.description)
			assert.Nil(t, claims, "Claims should be nil when validation fails")
		})
	}
}

// removeFile removes a file, ignoring errors (for test cleanup).
func removeFile(path string) {
	_ = os.Remove(path)
}

// createTempRSAKeys creates temporary RSA key files for testing.
func createTempRSAKeys(t *testing.T) (string, string) {
	t.Helper()

	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate RSA private key")

	// Create temporary files
	privateKeyFile, err := os.CreateTemp("", "test_private_key_*.pem")
	require.NoError(t, err, "Failed to create temporary private key file")
	privateKeyPath := privateKeyFile.Name()
	require.NoError(t, privateKeyFile.Close(), "Failed to close private key file")

	publicKeyFile, err := os.CreateTemp("", "test_public_key_*.pem")
	require.NoError(t, err, "Failed to create temporary public key file")
	publicKeyPath := publicKeyFile.Name()
	require.NoError(t, publicKeyFile.Close(), "Failed to close public key file")

	// Write private key to file
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyData := pem.EncodeToMemory(privateKeyPEM)
	err = os.WriteFile(privateKeyPath, privateKeyData, 0o600)
	require.NoError(t, err, "Failed to write private key file")

	// Write public key to file
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	}
	publicKeyData := pem.EncodeToMemory(publicKeyPEM)
	err = os.WriteFile(publicKeyPath, publicKeyData, 0o600)
	require.NoError(t, err, "Failed to write public key file")

	return privateKeyPath, publicKeyPath
}

// createTokenWithWrongAlgorithm creates a token signed with a different secret to simulate wrong algorithm.
func createTokenWithWrongAlgorithm(t *testing.T) string {
	t.Helper()

	// Create a service with different secret
	wrongService, err := NewJWTService(JWTConfig{
		Algorithm:         "HS256",
		Secret:            "different-secret-key",
		AccessTokenExpiry: 15 * time.Minute,
	})
	require.NoError(t, err, "Failed to create JWT service with different secret")

	// Create token with wrong secret
	token, err := wrongService.CreateAccessToken("user-123", []string{"user"})
	require.NoError(t, err, "Failed to create token with wrong secret")

	return token
}
