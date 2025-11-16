package blob

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultBlobFactory_Create_Success(t *testing.T) {
	ctx := context.Background()
	factory := NewDefaultBlobFactory()

	cfg := FilesystemConfig{
		URL: "mem://",
	}

	bucket, err := factory.Create(ctx, cfg)
	require.NoError(t, err, "should create bucket successfully")
	assert.NotNil(t, bucket, "bucket should not be nil")
}

func TestDefaultBlobFactory_Create_InvalidURL(t *testing.T) {
	ctx := context.Background()
	factory := NewDefaultBlobFactory()

	cfg := FilesystemConfig{
		URL: "invalid://scheme",
	}

	bucket, err := factory.Create(ctx, cfg)
	require.Error(t, err, "should return error for invalid URL")
	assert.Nil(t, bucket, "bucket should be nil on error")
	assert.Contains(t, err.Error(), "failed to open blob bucket", "error message should indicate bucket opening failure")
}

