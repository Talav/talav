package fxblob

import (
	"go.uber.org/fx"
	cloudblob "gocloud.dev/blob"
)

// FilesystemEntry represents a filesystem with its name for the group.
type FilesystemEntry struct {
	Name   string
	Bucket *cloudblob.Bucket
}

// FilesystemResult is used to register a filesystem in the filesystems group.
type FilesystemResult struct {
	fx.Out
	Filesystem FilesystemEntry `group:"talav-blob-filesystems"`
}

// AsFilesystem registers a filesystem that will be added to the blob.FilesystemRegistry.
// The constructor should return *cloudblob.Bucket or (*cloudblob.Bucket, error).
// Additional annotations (like fx.ParamTags) can be passed as variadic arguments.
//
// Example:
//
//	AsFilesystem("custom", func(ctx context.Context) (*cloudblob.Bucket, error) {
//		return cloudblob.OpenBucket(ctx, "file:///custom/path")
//	})
func AsFilesystem(name string, constructor any, annotations ...fx.Annotation) fx.Option {
	wrapper := func(bucket *cloudblob.Bucket) FilesystemResult {
		return FilesystemResult{
			Filesystem: FilesystemEntry{Name: name, Bucket: bucket},
		}
	}

	return fx.Options(
		fx.Provide(constructor),
		fx.Provide(
			fx.Annotate(
				wrapper,
				annotations...,
			),
		),
	)
}
