package provider

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"

	"github.com/talav/talav/pkg/component/media/app/thumbnail"
	"github.com/talav/talav/pkg/component/media/domain"
	"github.com/talav/talav/pkg/component/media/infra/cdn"
	"github.com/talav/talav/pkg/component/media/infra/resizer"
	"gocloud.dev/blob"
)

// ImageProvider implements a provider that handles image files with thumbnail generation.
type ImageProvider struct {
	*FileProvider
	thumbnail thumbnail.Thumbnailer
}

// NewImageProvider creates a new ImageProvider with the given extensions, bucket, CDN, and thumbnail service.
func NewImageProvider(extensions []string, bucket *blob.Bucket, cdnInstance cdn.CDN, thumb thumbnail.Thumbnailer, logger *slog.Logger) *ImageProvider {
	return &ImageProvider{
		FileProvider: NewFileProvider(extensions, bucket, cdnInstance, logger),
		thumbnail:    thumb,
	}
}

// generateThumbnailPath generates a storage path for a thumbnail
// Uses same folder structure as generateMediaPath: {preset}/{2 chars from id}/{2 chars from id}/{id}/{slug}-{formatName}.{ext}.
func (p *ImageProvider) generateThumbnailPath(media *domain.Media, formatName, extension string) string {
	thumbnailFileName := fmt.Sprintf("%s_%s.%s", media.GetSlug(), formatName, extension)

	return p.generateBasePath(media) + "/" + thumbnailFileName
}

// GenerateThumbnails generates thumbnails for the given media based on formats
// Pattern: {preset}/{id}/{slug}-{formatName}.{ext}
// Returns a map of format name to thumbnail metadata with storage paths.
func (p *ImageProvider) GenerateThumbnails(ctx context.Context, mediaEntity *domain.Media, data []byte, formats map[string]FormatConfig) (map[string]domain.ThumbnailMetadata, error) {
	thumbnailInputs := make(map[string]thumbnail.ThumbnailInput)
	for formatName, formatConfig := range formats {
		// Use format from config, fallback to original format
		thumbFormat := formatConfig.Format
		if thumbFormat == "" {
			thumbFormat = mediaEntity.GetExtension()
		}

		options := formatConfig.Options
		if options == nil {
			options = make(map[string]any)
		}

		thumbnailInputs[formatName] = thumbnail.ThumbnailInput{
			Resizer: formatConfig.Resizer,
			ResizeOptions: resizer.ResizeOptions{
				Width:   formatConfig.Width,
				Height:  formatConfig.Height,
				Format:  thumbFormat,
				Options: options,
			},
		}
	}

	// Create writer factory that uses bucket and handles path template substitution
	writerFactory := func(ctx context.Context, formatName string, extension string) (io.WriteCloser, error) {
		storagePath := p.generateThumbnailPath(mediaEntity, formatName, extension)
		writer, err := p.bucket.NewWriter(ctx, storagePath, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create thumbnail writer: %w", err)
		}

		return writer, nil
	}

	// Generate thumbnails (thumbnail service handles iteration)
	results, err := p.thumbnail.GenerateThumbnails(ctx, data, thumbnailInputs, writerFactory)
	if err != nil {
		p.logger.Error("Failed to generate thumbnails", "error", err, "mediaID", mediaEntity.ID)

		return nil, err
	}

	// Add storage paths to metadata
	for formatName, metadata := range results {
		metadata.URL = p.generateThumbnailPath(mediaEntity, formatName, metadata.Extension)
		results[formatName] = metadata
	}

	return results, nil
}

// GetDimensions extracts width and height from image data.
func (p *ImageProvider) GetDimensions(data []byte) (width, height int, err error) {
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image config: %w", err)
	}

	return img.Width, img.Height, nil
}
