package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher_RoundTrip(t *testing.T) {
	// Setup
	hasher := NewPasswordHasher(HasherConfig{
		BcryptCost: 10,
		SaltLength: 32,
	})

	password := "test-password-123"

	// Action
	salt, err := hasher.GenerateSalt()
	require.NoError(t, err, "GenerateSalt should not return an error")
	require.NotEmpty(t, salt, "Generated salt should not be empty")

	hashedPassword, err := hasher.HashPassword(password, salt)
	require.NoError(t, err, "HashPassword should not return an error")
	require.NotEmpty(t, hashedPassword, "Hashed password should not be empty")

	// Assertion
	err = hasher.ComparePassword(hashedPassword, password, salt)
	assert.NoError(t, err, "ComparePassword should succeed with correct password and salt")
}

func TestPasswordHasher_ComparePassword_WrongPassword(t *testing.T) {
	// Setup
	hasher := NewPasswordHasher(HasherConfig{
		BcryptCost: 10,
		SaltLength: 32,
	})

	correctPassword := "correct-password-123"
	wrongPassword := "wrong-password-456"

	// Action
	salt, err := hasher.GenerateSalt()
	require.NoError(t, err, "GenerateSalt should not return an error")

	hashedPassword, err := hasher.HashPassword(correctPassword, salt)
	require.NoError(t, err, "HashPassword should not return an error")

	// Assertion
	err = hasher.ComparePassword(hashedPassword, wrongPassword, salt)
	assert.Error(t, err, "ComparePassword should return an error with wrong password")
}
