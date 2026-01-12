package query

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/talav/talav/pkg/component/media/app/provider"
	"github.com/talav/talav/pkg/component/media/domain"
	"github.com/talav/talav/pkg/component/media/infra/repo"
)

// GetMediaQuery represents a query to get a single media entry.
type GetMediaQuery struct {
	ID string
}

// GetMediaResult represents the result of getting a media entry.
type GetMediaResult struct {
	Media     *domain.Media
	PublicURL string
}

// GetMediaQueryHandler handles get media queries.
type GetMediaQueryHandler struct {
	mediaRepo        repo.MediaRepository
	providerRegistry provider.Registry
	logger           *slog.Logger
}

// NewGetMediaQueryHandler creates a new GetMediaQueryHandler.
func NewGetMediaQueryHandler(mediaRepo repo.MediaRepository, providerRegistry provider.Registry, logger *slog.Logger) *GetMediaQueryHandler {
	return &GetMediaQueryHandler{
		mediaRepo:        mediaRepo,
		providerRegistry: providerRegistry,
		logger:           logger,
	}
}

// Handle processes the GetMediaQuery and returns the media entry.
func (h *GetMediaQueryHandler) Handle(ctx context.Context, query *GetMediaQuery) (*GetMediaResult, error) {
	media, err := h.mediaRepo.FindByID(ctx, query.ID)
	if err != nil {
		return nil, err
	}
	// Get provider from registry
	provider, err := h.providerRegistry.GetProvider(media.Provider)
	if err != nil {
		h.logger.Error("Failed to get provider for deletion", "error", err, "provider", media.Provider, "mediaID", query.ID)

		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	return &GetMediaResult{Media: media, PublicURL: provider.GetPublicURL(media)}, nil
}
