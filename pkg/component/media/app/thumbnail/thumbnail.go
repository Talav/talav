package thumbnail

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/talav/talav/pkg/component/media/domain"
	"github.com/talav/talav/pkg/component/media/infra/resizer"
)

// WriterFactory creates a writer for storing a thumbnail
// formatName is the format name that will be used to generate the storage path.
type WriterFactory func(ctx context.Context, formatName string, extension string) (io.WriteCloser, error)

// ThumbnailInput contains all information needed to generate a single thumbnail.
type ThumbnailInput struct {
	Resizer string // resizer name to use
	resizer.ResizeOptions
}

// Thumbnailer interface for thumbnail generation.
type Thumbnailer interface {
	// GenerateThumbnails generates multiple thumbnails for different formats
	GenerateThumbnails(ctx context.Context, data []byte, formats map[string]ThumbnailInput, writerFactory WriterFactory) (map[string]domain.ThumbnailMetadata, error)
}

// Thumbnail is the default implementation of Thumbnailer.
type Thumbnail struct {
	resizers map[string]resizer.Resizer
}

// NewThumbnail creates a new Thumbnail.
func NewThumbnail(resizers map[string]resizer.Resizer) Thumbnailer {
	return &Thumbnail{
		resizers: resizers,
	}
}

// GenerateThumbnails generates thumbnails for multiple formats.
func (s *Thumbnail) GenerateThumbnails(ctx context.Context, data []byte, formats map[string]ThumbnailInput, writerFactory WriterFactory) (map[string]domain.ThumbnailMetadata, error) {
	results := make(map[string]domain.ThumbnailMetadata)

	for formatName, input := range formats {
		// Create writer using factory (provider handles path generation)
		writer, err := writerFactory(ctx, formatName, input.Format)
		if err != nil {
			return nil, fmt.Errorf("failed to create writer for format %q: %w", formatName, err)
		}

		// Generate thumbnail using existing method
		result, err := s.generateThumbnail(ctx, data, input, writer)
		if err != nil {
			_ = writer.Close()

			return nil, fmt.Errorf("failed to generate thumbnail for format %q: %w", formatName, err)
		}

		// Collect metadata (storage path will be set by provider)
		results[formatName] = domain.ThumbnailMetadata{
			Width:     result.Width,
			Height:    result.Height,
			FileSize:  result.FileSize,
			Extension: input.Format,
		}
	}

	return results, nil
}

// GenerateThumbnail generates a single thumbnail based on the input.
func (s *Thumbnail) generateThumbnail(ctx context.Context, data []byte, input ThumbnailInput, writer io.WriteCloser) (resizer.ResizeResult, error) {
	// Get resizer
	r, exists := s.resizers[input.Resizer]
	if !exists {
		return resizer.ResizeResult{}, fmt.Errorf("resizer %q not found", input.Resizer)
	}

	// Create reader from data
	reader := bytes.NewReader(data)

	// Resize using resizer - streams directly to writer, no intermediate buffer
	// Resizer decodes to image.Image (memory), resizes, then encodes to writer
	// Resizer returns actual dimensions and file size
	result, err := r.Resize(ctx, reader, input.ResizeOptions, writer)
	if err != nil {
		_ = writer.Close()

		return resizer.ResizeResult{}, fmt.Errorf("failed to resize image: %w", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return resizer.ResizeResult{}, fmt.Errorf("failed to close writer: %w", err)
	}

	return result, nil
}
