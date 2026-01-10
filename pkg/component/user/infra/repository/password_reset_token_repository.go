package repository

import (
	"context"

	"github.com/talav/talav/pkg/component/orm"
	"github.com/talav/talav/pkg/component/user/domain"
	"gorm.io/gorm"
)

// PasswordResetTokenRepository interface defines all password reset token repository methods.
type PasswordResetTokenRepository interface {
	// Base repository methods
	Create(ctx context.Context, entity *domain.PasswordResetToken) error
	Update(ctx context.Context, entity *domain.PasswordResetToken) error

	// Token-specific methods
	FindByTokenLookup(ctx context.Context, tokenLookup string) (*domain.PasswordResetToken, error)
	InvalidateByUserID(ctx context.Context, userID string) error
	// DeleteExpired deletes all expired and unused tokens (for cleanup jobs).
	// Only hard-deletes expired tokens that were never used to preserve audit trail.
	DeleteExpired(ctx context.Context) error
}

// passwordResetTokenRepository implements PasswordResetTokenRepository with embedded base repository.
type passwordResetTokenRepository struct {
	*orm.BaseRepository[domain.PasswordResetToken]
}

// NewPasswordResetTokenRepository creates a new password reset token repository instance.
func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{
		BaseRepository: orm.NewBaseRepository[domain.PasswordResetToken](db),
	}
}

// FindByTokenLookup finds a token by its lookup hash (SHA256, for fast searching).
func (r *passwordResetTokenRepository) FindByTokenLookup(ctx context.Context, tokenLookup string) (*domain.PasswordResetToken, error) {
	return r.FindOne(ctx, "token_lookup", tokenLookup)
}

// InvalidateByUserID marks all unused tokens for a user as used (soft delete for audit trail).
// This preserves tokens in the database for security auditing while invalidating them.
func (r *passwordResetTokenRepository) InvalidateByUserID(ctx context.Context, userID string) error {
	return r.GetDB().WithContext(ctx).
		Model(&domain.PasswordResetToken{}).
		Where("user_id = ? AND used = ?", userID, false).
		Update("used", true).Error
}

// DeleteExpired deletes all expired tokens that were never used (for cleanup).
// Used tokens are preserved for audit trail, only expired+unused tokens are hard-deleted.
func (r *passwordResetTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.GetDB().WithContext(ctx).
		Where("expires_at < ? AND used = ?", r.getCurrentTimestamp(), false).
		Delete(&domain.PasswordResetToken{}).Error
}

// getCurrentTimestamp returns current Unix timestamp.
func (r *passwordResetTokenRepository) getCurrentTimestamp() int64 {
	return r.GetDB().NowFunc().Unix()
}
