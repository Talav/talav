package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// ErrRefreshTokenInvalid is returned when a refresh token is invalid or expired.
var ErrRefreshTokenInvalid = errors.New("refresh token is invalid or expired")

// RefreshTokenStore provides storage for refresh tokens (for rotation).
type RefreshTokenStore interface {
	// Store stores a refresh token with expiration.
	Store(ctx context.Context, userID string, tokenID string, expiresAt time.Time) error
	// Get retrieves a refresh token by ID.
	Get(ctx context.Context, tokenID string) (string, error)
	// Delete removes a refresh token.
	Delete(ctx context.Context, tokenID string) error
	// DeleteAllForUser removes all refresh tokens for a user.
	DeleteAllForUser(ctx context.Context, userID string) error
}

// RefreshTokenService provides refresh token management with rotation support.
type RefreshTokenService interface {
	// RotateRefreshToken validates an old refresh token and generates a new one.
	// The old token is invalidated.
	RotateRefreshToken(ctx context.Context, oldToken string) (newToken string, userID string, err error)
	// RevokeRefreshToken invalidates a refresh token.
	RevokeRefreshToken(ctx context.Context, token string) error
	// RevokeAllRefreshTokens invalidates all refresh tokens for a user.
	RevokeAllRefreshTokens(ctx context.Context, userID string) error
}

// DefaultRefreshTokenService is the default implementation of RefreshTokenService.
type DefaultRefreshTokenService struct {
	jwtService JWTService
	store      RefreshTokenStore
}

// NewRefreshTokenService creates a new RefreshTokenService.
func NewRefreshTokenService(jwtService JWTService, store RefreshTokenStore) RefreshTokenService {
	return &DefaultRefreshTokenService{
		jwtService: jwtService,
		store:      store,
	}
}

// RotateRefreshToken validates an old refresh token and generates a new one.
func (s *DefaultRefreshTokenService) RotateRefreshToken(ctx context.Context, oldToken string) (string, string, error) {
	// Validate the old token
	claims, err := s.jwtService.ValidateRefreshToken(oldToken)
	if err != nil {
		return "", "", ErrRefreshTokenInvalid
	}

	userID := claims.Subject
	if userID == "" {
		return "", "", ErrRefreshTokenInvalid
	}

	// Extract token ID from claims (use jti if present, otherwise use a hash of the token)
	tokenID := claims.ID
	if tokenID == "" {
		// Fallback: use a hash of the token as ID
		tokenID = hashToken(oldToken)
	}

	// Check if token exists in store (for rotation)
	if s.store != nil {
		storedUserID, err := s.store.Get(ctx, tokenID)
		if err != nil || storedUserID != userID {
			return "", "", ErrRefreshTokenInvalid
		}

		// Delete old token
		if err := s.store.Delete(ctx, tokenID); err != nil {
			return "", "", fmt.Errorf("failed to revoke old token: %w", err)
		}
	}

	// Generate new refresh token
	newToken, err := s.jwtService.CreateRefreshToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to create new refresh token: %w", err)
	}

	// Store new token if store is available
	if s.store != nil {
		newClaims, err := s.jwtService.ValidateRefreshToken(newToken)
		if err != nil {
			return "", "", fmt.Errorf("failed to validate new token: %w", err)
		}

		newTokenID := newClaims.ID
		if newTokenID == "" {
			newTokenID = hashToken(newToken)
		}

		expiresAt := newClaims.ExpiresAt.Time
		if err := s.store.Store(ctx, userID, newTokenID, expiresAt); err != nil {
			return "", "", fmt.Errorf("failed to store new token: %w", err)
		}
	}

	return newToken, userID, nil
}

// RevokeRefreshToken invalidates a refresh token.
func (s *DefaultRefreshTokenService) RevokeRefreshToken(ctx context.Context, token string) error {
	if s.store == nil {
		return nil // No-op if no store
	}

	claims, err := s.jwtService.ValidateRefreshToken(token)
	if err != nil {
		return ErrRefreshTokenInvalid
	}

	tokenID := claims.ID
	if tokenID == "" {
		tokenID = hashToken(token)
	}

	return s.store.Delete(ctx, tokenID)
}

// RevokeAllRefreshTokens invalidates all refresh tokens for a user.
func (s *DefaultRefreshTokenService) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	if s.store == nil {
		return nil // No-op if no store
	}

	return s.store.DeleteAllForUser(ctx, userID)
}

// hashToken creates a simple hash of the token for use as ID.
// In production, consider using a proper hash function.
func hashToken(token string) string {
	// Simple hash - in production use crypto/sha256
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

