package provider

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/talav/talav/pkg/component/media/domain"
	"github.com/talav/talav/pkg/component/media/infra/cdn"
	"gocloud.dev/blob"
)

// FileProvider implements a provider that handles specific file extensions
// FileProvider does not generate thumbnails - use ImageProvider for thumbnail support.
type FileProvider struct {
	extensions map[string]bool
	bucket     *blob.Bucket
	cdn        cdn.CDN
	logger     *slog.Logger
}

// NewFileProvider creates a new FileProvider with the given extensions, bucket, and CDN.
func NewFileProvider(extensions []string, bucket *blob.Bucket, cdnInstance cdn.CDN, logger *slog.Logger) *FileProvider {
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		// Normalize to lowercase and remove leading dot
		normalized := strings.TrimPrefix(strings.ToLower(ext), ".")
		if normalized != "" {
			extMap[normalized] = true
		}
	}

	return &FileProvider{
		extensions: extMap,
		bucket:     bucket,
		cdn:        cdnInstance,
		logger:     logger,
	}
}

// CanProcess checks if the provider can process the given media.
func (p *FileProvider) CanProcess(media *domain.Media) bool {
	return p.extensions[media.GetExtension()]
}

// Create creates/stores the file for the given media and returns the storage URL.
func (p *FileProvider) Create(ctx context.Context, media *domain.Media, data []byte) (string, error) {
	// Generate storage path
	storagePath := p.generateMediaPath(media)

	// Write to blob storage
	writer, err := p.bucket.NewWriter(ctx, storagePath, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create writer: %w", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		_ = writer.Close()

		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	return storagePath, nil
}

// Delete deletes the file for the given media and all its thumbnails.
func (p *FileProvider) Delete(ctx context.Context, media *domain.Media) error {
	// Delete original file
	if media.URL != "" {
		if err := p.bucket.Delete(ctx, media.URL); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	// Delete all thumbnails
	if media.Thumbnails != nil {
		for formatName, thumbMeta := range media.Thumbnails {
			if thumbMeta.URL != "" {
				if err := p.bucket.Delete(ctx, thumbMeta.URL); err != nil {
					p.logger.Warn("Failed to delete thumbnail", "error", err, "format", formatName, "url", thumbMeta.URL, "mediaID", media.ID)
					// Continue deleting other thumbnails even if one fails
				}
			}
		}
	}

	return nil
}

// generateMediaPath generates a storage path for the original media file
// Uses last 4 chars of xid (counter portion) for better distribution.
func (p *FileProvider) generateMediaPath(media *domain.Media) string {
	return p.generateBasePath(media) + "/" + media.FileName
}

// generateBasePath generates the base folder path: {preset}/{2 chars from id}/{2 chars from id}/{id}
// Uses last 4 chars of xid (counter portion) for better distribution.
func (p *FileProvider) generateBasePath(media *domain.Media) string {
	id := media.ID[4:]

	return fmt.Sprintf("%s/%s/%s/%s", media.Preset, id[16:18], id[18:20], media.ID)
}

// GenerateThumbnails returns an empty map as FileProvider does not support thumbnail generation
// Use ImageProvider for thumbnail support.
func (p *FileProvider) GenerateThumbnails(ctx context.Context, mediaEntity *domain.Media, data []byte, formats map[string]FormatConfig) (map[string]domain.ThumbnailMetadata, error) {
	return map[string]domain.ThumbnailMetadata{}, nil
}

// GetDimensions returns 0, 0 as FileProvider does not support dimension extraction.
func (p *FileProvider) GetDimensions(data []byte) (width, height int, err error) {
	return 0, 0, nil
}

func (p *FileProvider) GetPublicURL(media *domain.Media) string {
	return p.cdn.GetPath(media.URL)
}
