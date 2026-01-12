package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/talav/talav/pkg/component/media/app/provider"
	"github.com/talav/talav/pkg/component/media/infra/repo"
	"gorm.io/gorm"
)

// DeleteMediaCommand represents the command to delete a media entry.
type DeleteMediaCommand struct {
	ID string
}

// DeleteMediaResult represents the result of deleting a media entry.
type DeleteMediaResult struct {
	// Empty result
}

// DeleteMediaHandler handles media deletion commands.
type DeleteMediaHandler struct {
	mediaRepo        repo.MediaRepository
	providerRegistry provider.Registry
	logger           *slog.Logger
}

// NewDeleteMediaHandler creates a new DeleteMediaHandler.
func NewDeleteMediaHandler(mediaRepo repo.MediaRepository, providerRegistry provider.Registry, logger *slog.Logger) *DeleteMediaHandler {
	return &DeleteMediaHandler{
		mediaRepo:        mediaRepo,
		providerRegistry: providerRegistry,
		logger:           logger,
	}
}

// Handle processes the DeleteMediaCommand and deletes the media entry and file.
func (h *DeleteMediaHandler) Handle(ctx context.Context, cmd *DeleteMediaCommand) (*DeleteMediaResult, error) {
	// Find existing media to get provider name and file information
	media, err := h.mediaRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		// Preserve GORM's ErrRecordNotFound so HTTP layer can return 404
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("failed to find media: %w", err)
	}

	// Get provider from registry
	provider, err := h.providerRegistry.GetProvider(media.Provider)
	if err != nil {
		h.logger.Error("Failed to get provider for deletion", "error", err, "provider", media.Provider, "mediaID", cmd.ID)

		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Delete from repository first
	err = h.mediaRepo.Delete(ctx, cmd.ID)
	if err != nil {
		h.logger.Error("Failed to delete media from repository", "error", err, "mediaID", cmd.ID)

		return nil, fmt.Errorf("failed to delete media: %w", err)
	}

	// Delete original file from storage using provider (best effort - log error but don't fail)
	// Repository deletion succeeded, so we continue even if file deletion fails
	if err := provider.Delete(ctx, media); err != nil {
		h.logger.Warn("Failed to delete file from storage", "error", err, "storagePath", media.URL, "provider", media.Provider, "mediaID", cmd.ID)
	}

	return &DeleteMediaResult{}, nil
}
