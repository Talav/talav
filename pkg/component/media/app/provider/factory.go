package provider

import (
	"fmt"
	"log/slog"

	"github.com/talav/talav/pkg/component/media/app/thumbnail"
	"github.com/talav/talav/pkg/component/media/infra/cdn"
	"gocloud.dev/blob"
)

// Factory is the interface for Provider factories.
type Factory interface {
	Create(cfg ProviderConfig, bucket *blob.Bucket, cdnInstance cdn.CDN, thumb thumbnail.Thumbnailer, logger *slog.Logger) (Provider, error)
}

// DefaultFactory is the default [Factory] implementation.
type DefaultFactory struct{}

// NewDefaultFactory returns a [DefaultFactory], implementing [Factory].
func NewDefaultFactory() Factory {
	return &DefaultFactory{}
}

// Create returns a new [Provider] based on the configuration.
func (f *DefaultFactory) Create(cfg ProviderConfig, bucket *blob.Bucket, cdnInstance cdn.CDN, thumb thumbnail.Thumbnailer, logger *slog.Logger) (Provider, error) {
	switch cfg.Type {
	case "file":
		return f.createFileProvider(cfg, bucket, cdnInstance, logger)
	case "image":
		return f.createImageProvider(cfg, bucket, cdnInstance, thumb, logger)
	default:
		if cfg.Type == "" {
			return nil, fmt.Errorf("provider.type is required")
		}

		return nil, fmt.Errorf("unknown provider type: %s", cfg.Type)
	}
}

// createFileProvider creates a file provider with validation.
func (f *DefaultFactory) createFileProvider(cfg ProviderConfig, bucket *blob.Bucket, cdnInstance cdn.CDN, logger *slog.Logger) (Provider, error) {
	if len(cfg.Extensions) == 0 {
		return nil, fmt.Errorf("provider.extensions is required for file provider")
	}
	if bucket == nil {
		return nil, fmt.Errorf("bucket is required for file provider")
	}
	if cdnInstance == nil {
		return nil, fmt.Errorf("cdn is required for file provider")
	}

	return NewFileProvider(cfg.Extensions, bucket, cdnInstance, logger), nil
}

// createImageProvider creates an image provider with validation.
func (f *DefaultFactory) createImageProvider(cfg ProviderConfig, bucket *blob.Bucket, cdnInstance cdn.CDN, thumb thumbnail.Thumbnailer, logger *slog.Logger) (Provider, error) {
	if len(cfg.Extensions) == 0 {
		return nil, fmt.Errorf("provider.extensions is required for image provider")
	}
	if bucket == nil {
		return nil, fmt.Errorf("bucket is required for image provider")
	}
	if cdnInstance == nil {
		return nil, fmt.Errorf("cdn is required for image provider")
	}
	if thumb == nil {
		return nil, fmt.Errorf("thumbnail is required for image provider")
	}

	return NewImageProvider(cfg.Extensions, bucket, cdnInstance, thumb, logger), nil
}
