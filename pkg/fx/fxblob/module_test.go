package fxblob

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/talav/pkg/component/blob"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	cloudblob "gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

func TestModule_NewFilesystemRegistry_MultipleFilesystems(t *testing.T) {
	var registry *blob.FilesystemRegistry

	fxtest.New(t,
		fx.Supply(fx.Annotate(context.Background(), fx.As(new(context.Context)))),
		fx.Supply(blob.BlobConfig{
			Filesystems: map[string]blob.FilesystemConfig{
				"default": {URL: "mem://"},
				"backup":  {URL: "mem://"},
			},
		}),
		fx.Provide(NewFilesystemRegistry),
		fx.Populate(&registry),
	).RequireStart()

	require.NotNil(t, registry, "FilesystemRegistry should be provided")

	bucket1, err := registry.Get("default")
	require.NoError(t, err, "Should resolve default filesystem")
	assert.NotNil(t, bucket1, "Default bucket should not be nil")

	bucket2, err := registry.Get("backup")
	require.NoError(t, err, "Should resolve backup filesystem")
	assert.NotNil(t, bucket2, "Backup bucket should not be nil")

	assert.NotSame(t, bucket1, bucket2, "Buckets should be different instances")
}

func TestModule_AsFilesystem_RegistersCustomFilesystem(t *testing.T) {
	var registry *blob.FilesystemRegistry

	fxtest.New(t,
		fx.Supply(fx.Annotate(context.Background(), fx.As(new(context.Context)))),
		fx.Supply(blob.BlobConfig{
			Filesystems: map[string]blob.FilesystemConfig{},
		}),
		AsFilesystem("custom", func(ctx context.Context) (*cloudblob.Bucket, error) {
			return cloudblob.OpenBucket(ctx, "mem://")
		}),
		fx.Provide(NewFilesystemRegistry),
		fx.Populate(&registry),
	).RequireStart()

	require.NotNil(t, registry, "FilesystemRegistry should be provided")

	// Test that the custom filesystem is registered
	bucket, err := registry.Get("custom")
	require.NoError(t, err, "Should resolve custom filesystem")
	assert.NotNil(t, bucket, "Custom bucket should not be nil")
}
