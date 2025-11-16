package blob

import (
	"fmt"

	"gocloud.dev/blob"
)

// FilesystemRegistry maps filesystem names to buckets.
type FilesystemRegistry struct {
	buckets map[string]*blob.Bucket
}

// Get returns the bucket for the given filesystem name from the registry.
func (r *FilesystemRegistry) Get(name string) (*blob.Bucket, error) {
	bucket, exists := r.buckets[name]
	if !exists {
		return nil, fmt.Errorf("filesystem %q not found", name)
	}

	return bucket, nil
}

// NewFilesystemRegistry creates a new filesystem registry with the given buckets.
func NewFilesystemRegistry(buckets map[string]*blob.Bucket) *FilesystemRegistry {
	return &FilesystemRegistry{buckets: buckets}
}
