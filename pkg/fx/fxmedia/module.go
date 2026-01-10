package fxmedia

import (
	"fmt"
	"log/slog"

	"github.com/talav/talav/pkg/component/blob"
	"github.com/talav/talav/pkg/component/media"
	"github.com/talav/talav/pkg/component/media/app/command"
	"github.com/talav/talav/pkg/component/media/app/preset"
	"github.com/talav/talav/pkg/component/media/app/provider"
	"github.com/talav/talav/pkg/component/media/app/query"
	"github.com/talav/talav/pkg/component/media/app/thumbnail"
	"github.com/talav/talav/pkg/component/media/cdn"
	"github.com/talav/talav/pkg/component/media/infra/repo"
	"github.com/talav/talav/pkg/component/media/infra/resizer"
	mediastorage "github.com/talav/talav/pkg/component/media/infra/storage"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxorm"
	"go.uber.org/fx"
	cloudblob "gocloud.dev/blob"
)

const ModuleName = "media"

// FxMediaModule registers media-specific handlers and services.
var FxMediaModule = fx.Module(
	ModuleName,
	fxconfig.AsConfig("media", media.MediaConfig{}),
	fx.Provide(
		// Provide CDN factory
		cdn.NewDefaultFactory,
		// Provide CDN map
		NewCDNMap,
		// Provide resizers map
		NewFxResizers,
		// Provide image codec
		NewFxImageCodec,
		// Provide thumbnail (depends on resizers and codec)
		NewFxThumbnail,
		// Provide provider factory
		provider.NewDefaultFactory,
		// Provide providers map (depends on thumbnail)
		NewFxProviders,
		// Provide provider registry
		NewFxProviderRegistry,
		// Provide preset registry
		NewFxPresetRegistry,
		// Provide bucket for media storage (selected by config)
		NewFxMediaStorageBucket,
		// Provide storage service
		NewFxMediaStorage,
		// Provide command handlers
		command.NewCreateMediaHandler,
		command.NewUpdateMediaHandler,
		command.NewDeleteMediaHandler,
		// Provide query handlers
		query.NewGetMediaQueryHandler,
		query.NewListMediaQueryHandler,
	),
	// Register repository to the registry
	fxorm.AsRepository[repo.MediaRepository](repo.NewMediaRepository),
	// Register resizers with keys from config (they depend on codec)
	AsResizer("simple", func(c resizer.ImageCodec) resizer.Resizer {
		return resizer.NewSimpleResizer(c)
	}),
	AsResizer("crop", func(c resizer.ImageCodec) resizer.Resizer {
		return resizer.NewCropResizer(c)
	}),
	AsResizer("square", func(c resizer.ImageCodec) resizer.Resizer {
		return resizer.NewSquareResizer(c)
	}),
)

// NewFxMediaStorageBucket resolves the filesystem bucket by name from the registry.
func NewFxMediaStorageBucket(registry *blob.FilesystemRegistry, cfg media.MediaConfig) (*cloudblob.Bucket, error) {
	// Try to use first provider's filesystem as fallback
	if len(cfg.Providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	// Use first provider's filesystem as default
	for _, providerConfig := range cfg.Providers {
		if providerConfig.Filesystem != "" {
			return registry.Get(providerConfig.Filesystem)
		}
	}

	return nil, fmt.Errorf("no filesystem configured in providers")
}

// NewFxMediaStorage returns a new [mediastorage.MediaStorage].
func NewFxMediaStorage(bucket *cloudblob.Bucket, cfg media.MediaConfig) *mediastorage.MediaStorage {
	return mediastorage.NewMediaStorage(bucket, cfg)
}

// CDNMap maps CDN names to CDN instances.
type CDNMap map[string]cdn.CDN

// ResolveCDN returns the CDN for the given name from the map.
func ResolveCDN(cdnMap CDNMap, name string) (cdn.CDN, error) {
	if name == "" {
		name = "default"
	}

	cdnInstance, exists := cdnMap[name]
	if !exists {
		return nil, fmt.Errorf("cdn %q not found", name)
	}

	return cdnInstance, nil
}

// NewCDNMap creates a map of CDNs from configuration.
func NewCDNMap(cfg media.MediaConfig, factory cdn.Factory, logger *slog.Logger) (CDNMap, error) {
	cdnMap := make(CDNMap)

	for name, cdnSpec := range cfg.CDN {
		cdnInstance, err := factory.Create(cdnSpec, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create CDN %q: %w", name, err)
		}

		cdnMap[name] = cdnInstance
	}

	return cdnMap, nil
}

// NewFxProviders creates a map of providers from configuration.
func NewFxProviders(
	factory provider.Factory,
	cfg media.MediaConfig,
	registry *blob.FilesystemRegistry,
	cdnMap CDNMap,
	thumbnailer thumbnail.Thumbnailer,
	logger *slog.Logger,
) (map[string]provider.Provider, error) {
	providers := make(map[string]provider.Provider)

	for name, providerConfig := range cfg.Providers {
		if providerConfig.Filesystem == "" {
			return nil, fmt.Errorf("provider %q filesystem is required", name)
		}

		bucket, err := registry.Get(providerConfig.Filesystem)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve filesystem %q for provider %q: %w", providerConfig.Filesystem, name, err)
		}

		cdnName := providerConfig.CDN
		if cdnName == "" {
			cdnName = "default"
		}

		cdnInstance, err := ResolveCDN(cdnMap, cdnName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve CDN %q for provider %q: %w", cdnName, name, err)
		}

		p, err := factory.Create(providerConfig, bucket, cdnInstance, thumbnailer, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %q: %w", name, err)
		}

		providers[name] = p
	}

	return providers, nil
}

// NewFxProviderRegistry creates a new provider registry.
func NewFxProviderRegistry(providers map[string]provider.Provider) provider.Registry {
	return provider.NewDefaultRegistry(providers)
}

// NewFxPresetRegistry creates a new preset registry.
func NewFxPresetRegistry(cfg media.MediaConfig) preset.Registry {
	return preset.NewDefaultRegistry(cfg.Presets)
}

// ResizerEntry represents a named resizer entry.
type ResizerEntry struct {
	Name    string
	Resizer resizer.Resizer
}

// FxResizersParam allows injection of registered resizers.
type FxResizersParam struct {
	fx.In
	Resizers []ResizerEntry `group:"resizers"`
}

// NewFxResizers creates a map of resizers from registered resizers.
func NewFxResizers(p FxResizersParam) map[string]resizer.Resizer {
	resizers := make(map[string]resizer.Resizer)
	for _, entry := range p.Resizers {
		resizers[entry.Name] = entry.Resizer
	}

	return resizers
}

// NewFxImageCodec creates a new image codec (default implementation using imaging library).
func NewFxImageCodec() resizer.ImageCodec {
	return resizer.NewImagingCodec()
}

// NewFxThumbnail creates a new thumbnail.
func NewFxThumbnail(resizers map[string]resizer.Resizer) thumbnail.Thumbnailer {
	return thumbnail.NewThumbnail(resizers)
}
