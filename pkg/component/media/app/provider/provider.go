package provider

import (
	"context"

	"github.com/talav/talav/pkg/component/media/domain"
)

// Provider defines the interface for file providers.
type Provider interface {
	// CanProcess checks if the provider can process the given media
	CanProcess(media *domain.Media) bool
	// Create creates/stores the file for the given media and returns the storage URL
	Create(ctx context.Context, media *domain.Media, data []byte) (string, error)
	// Delete deletes the file for the given media
	Delete(ctx context.Context, media *domain.Media) error
	// GenerateThumbnails generates thumbnails for the given media based on formats
	// Resizers are handled internally by the thumbnail service
	// data should contain the full file content in memory
	// Returns a map of format name to thumbnail metadata
	// Providers that don't support thumbnails (e.g., FileProvider) return an empty map
	GenerateThumbnails(ctx context.Context, media *domain.Media, data []byte, formats map[string]FormatConfig) (map[string]domain.ThumbnailMetadata, error)
	// GetDimensions extracts width and height from the media file data
	// Returns 0, 0 for providers that don't support dimension extraction (e.g., FileProvider)
	GetDimensions(data []byte) (width, height int, err error)

	GetPublicURL(media *domain.Media) string
}
