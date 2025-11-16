package blob

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

func TestFilesystemRegistry_Get_Exists(t *testing.T) {
	ctx := context.Background()

	// Create test buckets
	bucket1, err := blob.OpenBucket(ctx, "mem://")
	require.NoError(t, err, "should create bucket1")

	bucket2, err := blob.OpenBucket(ctx, "mem://")
	require.NoError(t, err, "should create bucket2")

	buckets := map[string]*blob.Bucket{
		"default": bucket1,
		"backup":  bucket2,
	}

	registry := NewFilesystemRegistry(buckets)

	// Test getting existing filesystem
	retrieved, err := registry.Get("default")
	require.NoError(t, err, "should retrieve existing filesystem")
	assert.Same(t, bucket1, retrieved, "should return the same bucket instance")

	retrieved, err = registry.Get("backup")
	require.NoError(t, err, "should retrieve existing filesystem")
	assert.Same(t, bucket2, retrieved, "should return the same bucket instance")
}

func TestFilesystemRegistry_Get_NotExists(t *testing.T) {
	ctx := context.Background()

	// Create test bucket
	bucket, err := blob.OpenBucket(ctx, "mem://")
	require.NoError(t, err, "should create bucket")

	buckets := map[string]*blob.Bucket{
		"default": bucket,
	}

	registry := NewFilesystemRegistry(buckets)

	// Test getting non-existing filesystem
	retrieved, err := registry.Get("nonexistent")
	require.Error(t, err, "should return error for non-existing filesystem")
	assert.Nil(t, retrieved, "should return nil bucket")
	assert.Contains(t, err.Error(), "filesystem \"nonexistent\" not found", "error message should contain filesystem name")
}
