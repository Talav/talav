package blob

import (
	"context"
	"fmt"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
)

// BlobFactory is the interface for [blob.Bucket] factories.
type BlobFactory interface {
	Create(ctx context.Context, cfg FilesystemConfig) (*blob.Bucket, error)
}

// DefaultBlobFactory is the default [BlobFactory] implementation.
type DefaultBlobFactory struct{}

// NewDefaultBlobFactory returns a [DefaultBlobFactory], implementing [BlobFactory].
func NewDefaultBlobFactory() BlobFactory {
	return &DefaultBlobFactory{}
}

// Create returns a new [blob.Bucket] from the given URL.
//
// Example:
//
//	var factory = NewDefaultBlobFactory(logger)
//	var bucket, _ = factory.Create(ctx, "file:///tmp/storage")
//
// Supported URL formats:
//   - file:///path/to/dir - Local filesystem
//   - mem:// - In-memory storage
//   - s3://bucket?region=us-west-1 - AWS S3
//   - gs://bucket - Google Cloud Storage
//   - azblob://container - Azure Blob Storage
func (f *DefaultBlobFactory) Create(ctx context.Context, cfg FilesystemConfig) (*blob.Bucket, error) {
	bucket, err := blob.OpenBucket(ctx, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open blob bucket: %w", err)
	}

	return bucket, nil
}
