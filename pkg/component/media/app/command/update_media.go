package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/talav/talav/pkg/component/media/app/provider"
	"github.com/talav/talav/pkg/component/media/domain"
	"github.com/talav/talav/pkg/component/media/infra/repo"
	"gorm.io/gorm"
)

// UpdateMediaCommand represents the command to update a media entry.
type UpdateMediaCommand struct {
	ID          string
	Description string // Only description can be updated (media is immutable)
}

// UpdateMediaResult represents the result of updating a media entry.
type UpdateMediaResult struct {
	domain.MediaView
}

// UpdateMediaHandler handles media update commands.
type UpdateMediaHandler struct {
	mediaRepo        repo.MediaRepository
	providerRegistry provider.Registry
	logger           *slog.Logger
}

// NewUpdateMediaHandler creates a new UpdateMediaHandler.
func NewUpdateMediaHandler(mediaRepo repo.MediaRepository, providerRegistry provider.Registry, logger *slog.Logger) *UpdateMediaHandler {
	return &UpdateMediaHandler{
		mediaRepo:        mediaRepo,
		providerRegistry: providerRegistry,
		logger:           logger,
	}
}

// Handle processes the UpdateMediaCommand and returns the updated media entry.
func (h *UpdateMediaHandler) Handle(ctx context.Context, cmd *UpdateMediaCommand) (*UpdateMediaResult, error) {
	// Find existing media
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

	// Update description (only mutable field)
	media.UpdateDescription(cmd.Description)

	// Save changes
	err = h.mediaRepo.Update(ctx, media)
	if err != nil {
		h.logger.Error("Failed to update media", "error", err, "mediaID", cmd.ID)

		return nil, fmt.Errorf("failed to update media: %w", err)
	}

	return &UpdateMediaResult{
		MediaView: domain.MediaView{
			Media:     media,
			PublicURL: provider.GetPublicURL(media),
		},
	}, nil
}
