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
	bcryptCost int
	saltLength int
}

// NewPasswordHasher creates a new PasswordHasher with the given configuration.
func NewPasswordHasher(cfg SecurityConfig) PasswordHasher {
	bcryptCost := cfg.BcryptCost
	if bcryptCost == 0 {
		bcryptCost = 10 // default bcrypt cost
	}

	saltLength := cfg.SaltLength
	if saltLength == 0 {
		saltLength = 32 // default salt length in bytes
	}

	return &DefaultPasswordHasher{
		bcryptCost: bcryptCost,
		saltLength: saltLength,
	}
}

// GenerateSalt generates a cryptographically secure random salt.
func (h *DefaultPasswordHasher) GenerateSalt() (string, error) {
	// Generate random bytes
	saltBytes := make([]byte, h.saltLength)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Encode as base64 string for storage
	return base64.StdEncoding.EncodeToString(saltBytes), nil
}

// HashPassword hashes a plain password with the provided salt using bcrypt.
func (h *DefaultPasswordHasher) HashPassword(plainPassword, salt string) (string, error) {
	// Combine password and salt
	passwordWithSalt := plainPassword + salt

	// Hash with bcrypt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordWithSalt), h.bcryptCost)
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

