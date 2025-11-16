package fxblob

import (
	"context"
	"fmt"

	"github.com/talav/talav/pkg/component/blob"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
	cloudblob "gocloud.dev/blob"
)

const ModuleName = "blob"

// FxBlobModule is the [Fx] blob module.
var FxBlobModule = fx.Module(
	ModuleName,
	fxconfig.AsConfig("blob", blob.BlobConfig{}),
	fx.Provide(NewFilesystemRegistry),
)

// FxFilesystemRegistryParam allows injection of registered filesystems.
type FxFilesystemRegistryParam struct {
	fx.In
	Filesystems []FilesystemEntry `group:"talav-blob-filesystems"`
}

// NewFilesystemRegistry creates a registry of filesystems from config and registered filesystems.
// The registry can be injected and used with Get for lookups.
func NewFilesystemRegistry(ctx context.Context, cfg blob.BlobConfig, p FxFilesystemRegistryParam) (*blob.FilesystemRegistry, error) {
	buckets := make(map[string]*cloudblob.Bucket)

	// Add filesystems from config
	for name, fsConfig := range cfg.Filesystems {
		bucket, err := cloudblob.OpenBucket(ctx, fsConfig.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to open filesystem %q: %w", name, err)
		}
		buckets[name] = bucket
	}

	// Add registered filesystems (override config if name conflicts)
	for _, entry := range p.Filesystems {
		buckets[entry.Name] = entry.Bucket
	}

	return blob.NewFilesystemRegistry(buckets), nil
}
