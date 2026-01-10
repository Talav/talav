package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/talav/talav/pkg/component/user"
	"golang.org/x/crypto/bcrypt"
)

const PasswordResetTokenIDPrefix = "pwt"

// PasswordResetToken represents a password reset token in the domain layer.
type PasswordResetToken struct {
	ID          string `json:"id" gorm:"primaryKey"`
	UserID      string `json:"user_id" gorm:"index;not null"`
	TokenHash   string `json:"token_hash" gorm:"not null"`         // bcrypt hash for verification
	TokenLookup string `json:"token_lookup" gorm:"index;not null"` // SHA256 hash for fast lookup
	ExpiresAt   int64  `json:"expires_at" gorm:"index;not null"`
	Used        bool   `json:"used" gorm:"index;default:false;not null"`
	CreatedAt   int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewPasswordResetToken creates a new PasswordResetToken entity.
// token is the raw token string (will be hashed before storage).
// expiresAt is the Unix timestamp when the token expires.
func NewPasswordResetToken(userID string, rawToken string, expiresAt int64) (*PasswordResetToken, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if rawToken == "" {
		return nil, fmt.Errorf("token is required")
	}
	if expiresAt <= time.Now().Unix() {
		return nil, fmt.Errorf("expires_at must be in the future")
	}

	tokenHash, err := hashToken(rawToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}

	tokenLookup := HashTokenForLookup(rawToken)

	token := &PasswordResetToken{
		ID:          user.GenerateID(PasswordResetTokenIDPrefix),
		UserID:      userID,
		TokenHash:   tokenHash,
		TokenLookup: tokenLookup,
		ExpiresAt:   expiresAt,
		Used:        false,
	}

	return token, nil
}

// GenerateRawToken generates a cryptographically secure random token.
// Returns the raw token string (should be sent to user via email).
func GenerateRawToken() (string, error) {
	// Generate 32 bytes of random data (256 bits)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Encode as base64 URL-safe string for use in URLs
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// hashToken hashes a token using bcrypt (same approach as password hashing).
func hashToken(rawToken string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawToken), 10)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}

	return string(hashedBytes), nil
}

// HashTokenForLookup creates a fast SHA256 hash for database lookup.
// This allows efficient searching while bcrypt provides security.
func HashTokenForLookup(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))

	return hex.EncodeToString(hash[:])
}

// CompareToken compares a raw token with the stored token hash.
func (t *PasswordResetToken) CompareToken(rawToken string) error {
	return bcrypt.CompareHashAndPassword([]byte(t.TokenHash), []byte(rawToken))
}

// IsExpired checks if the token has expired.
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().Unix() > t.ExpiresAt
}

// IsValid checks if the token is valid (not used and not expired).
func (t *PasswordResetToken) IsValid() bool {
	return !t.Used && !t.IsExpired()
}

// MarkAsUsed marks the token as used.
func (t *PasswordResetToken) MarkAsUsed() {
	t.Used = true
}
