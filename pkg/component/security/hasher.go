package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher provides password hashing and comparison operations.
type PasswordHasher interface {
	// GenerateSalt generates a cryptographically secure random salt.
	GenerateSalt() (string, error)
	// HashPassword hashes a plain password with the provided salt.
	HashPassword(plainPassword, salt string) (string, error)
	// ComparePassword compares a plain password with a hashed password using the provided salt.
	ComparePassword(hashedPassword, plainPassword, salt string) error
}

// DefaultPasswordHasher is the default implementation of PasswordHasher.
type DefaultPasswordHasher struct {
	cfg HasherConfig
}

// NewPasswordHasher creates a new PasswordHasher with the given configuration.
func NewPasswordHasher(cfg HasherConfig) PasswordHasher {
	return &DefaultPasswordHasher{
		cfg: cfg,
	}
}

// GenerateSalt generates a cryptographically secure random salt.
func (h *DefaultPasswordHasher) GenerateSalt() (string, error) {
	// Generate random bytes
	saltBytes := make([]byte, h.cfg.SaltLength)
	_, _ = rand.Read(saltBytes)

	// Encode as base64 string for storage
	return base64.StdEncoding.EncodeToString(saltBytes), nil
}

// HashPassword hashes a plain password with the provided salt using bcrypt.
func (h *DefaultPasswordHasher) HashPassword(plainPassword, salt string) (string, error) {
	// Combine password and salt
	passwordWithSalt := plainPassword + salt

	// Hash with bcrypt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordWithSalt), h.cfg.BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// ComparePassword compares a plain password with a hashed password using the provided salt.
func (h *DefaultPasswordHasher) ComparePassword(hashedPassword, plainPassword, salt string) error {
	// Combine password and salt
	passwordWithSalt := plainPassword + salt

	// Compare with bcrypt
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(passwordWithSalt))
}
