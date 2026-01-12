package thumbnail

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/talav/pkg/component/media/infra/resizer"
)

// mockResizer is a mock implementation of resizer.Resizer.
type mockResizer struct {
	resizeFunc func(ctx context.Context, reader io.Reader, opts resizer.ResizeOptions, writer io.Writer) (resizer.ResizeResult, error)
}

func (m *mockResizer) Resize(ctx context.Context, reader io.Reader, opts resizer.ResizeOptions, writer io.Writer) (resizer.ResizeResult, error) {
	if m.resizeFunc != nil {
		return m.resizeFunc(ctx, reader, opts, writer)
	}

	return resizer.ResizeResult{}, nil
}

// mockWriteCloser is a mock implementation of io.WriteCloser that tracks close state.
type mockWriteCloser struct {
	io.Writer
	closed bool
}

func newMockWriteCloser() *mockWriteCloser {
	return &mockWriteCloser{Writer: io.Discard}
}

func (m *mockWriteCloser) Close() error {
	m.closed = true

	return nil
}

// setupThumbnail creates a Thumbnail instance with mock resizers.
func setupThumbnail(resizers map[string]resizer.Resizer) *Thumbnail {
	return &Thumbnail{
		resizers: resizers,
	}
}

// createMockResizer creates a mock resizer with the given behavior.
func createMockResizer(result resizer.ResizeResult, err error) *mockResizer {
	return &mockResizer{
		resizeFunc: func(ctx context.Context, reader io.Reader, opts resizer.ResizeOptions, writer io.Writer) (resizer.ResizeResult, error) {
			return result, err
		},
	}
}

// createMockWriterFactory creates a mock writer factory.
func createMockWriterFactory(writers map[string]io.WriteCloser, err error) WriterFactory {
	return func(ctx context.Context, formatName string, extension string) (io.WriteCloser, error) {
		if err != nil {
			return nil, err
		}
		if writers != nil {
			if writer, ok := writers[formatName]; ok {
				return writer, nil
			}
		}

		return newMockWriteCloser(), nil
	}
}

func TestNewThumbnail(t *testing.T) {
	t.Run("ValidResizers", func(t *testing.T) {
		resizers := map[string]resizer.Resizer{
			"simple": createMockResizer(resizer.ResizeResult{}, nil),
		}
		thumb := NewThumbnail(resizers)

		require.NotNil(t, thumb)
		assert.Implements(t, (*Thumbnailer)(nil), thumb)
	})

	t.Run("EmptyResizers", func(t *testing.T) {
		resizers := map[string]resizer.Resizer{}
		thumb := NewThumbnail(resizers)

		require.NotNil(t, thumb)
		assert.Implements(t, (*Thumbnailer)(nil), thumb)
	})
}

func TestThumbnail_GenerateThumbnails(t *testing.T) {
	t.Run("SingleFormat_Success", func(t *testing.T) {
		ctx := context.Background()
		data := []byte("test image data")
		expectedResult := resizer.ResizeResult{
			Width:    100,
			Height:   100,
			FileSize: 1024,
		}

		mockResizer := createMockResizer(expectedResult, nil)
		resizers := map[string]resizer.Resizer{
			"simple": mockResizer,
		}
		thumb := setupThumbnail(resizers)

		mockWriter := newMockWriteCloser()
		writers := map[string]io.WriteCloser{
			"thumb": mockWriter,
		}
		writerFactory := createMockWriterFactory(writers, nil)

		formats := map[string]ThumbnailInput{
			"thumb": {
				Resizer: "simple",
				ResizeOptions: resizer.ResizeOptions{
					Width:  100,
					Height: 100,
					Format: ".jpg",
				},
			},
		}

		results, err := thumb.GenerateThumbnails(ctx, data, formats, writerFactory)

		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results, 1)

		metadata, ok := results["thumb"]
		require.True(t, ok)
		assert.Equal(t, expectedResult.Width, metadata.Width)
		assert.Equal(t, expectedResult.Height, metadata.Height)
		assert.Equal(t, expectedResult.FileSize, metadata.FileSize)
		assert.True(t, mockWriter.closed, "Writer should be closed")
	})

	t.Run("MultipleFormats_Success", func(t *testing.T) {
		ctx := context.Background()
		data := []byte("test image data")

		mockResizer := createMockResizer(resizer.ResizeResult{
			Width:    100,
			Height:   100,
			FileSize: 1024,
		}, nil)
		resizers := map[string]resizer.Resizer{
			"simple": mockResizer,
		}
		thumb := setupThumbnail(resizers)

		mockWriter1 := newMockWriteCloser()
		mockWriter2 := newMockWriteCloser()
		writers := map[string]io.WriteCloser{
			"thumb":   mockWriter1,
			"preview": mockWriter2,
		}
		writerFactory := createMockWriterFactory(writers, nil)

		formats := map[string]ThumbnailInput{
			"thumb": {
				Resizer: "simple",
				ResizeOptions: resizer.ResizeOptions{
					Width:  100,
					Height: 100,
					Format: ".jpg",
				},
			},
			"preview": {
				Resizer: "simple",
				ResizeOptions: resizer.ResizeOptions{
					Width:  200,
					Height: 150,
					Format: ".jpg",
				},
			},
		}

		results, err := thumb.GenerateThumbnails(ctx, data, formats, writerFactory)

		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results, 2)
		assert.Contains(t, results, "thumb")
		assert.Contains(t, results, "preview")
		assert.True(t, mockWriter1.closed, "First writer should be closed")
		assert.True(t, mockWriter2.closed, "Second writer should be closed")
	})

	t.Run("WriterFactoryError", func(t *testing.T) {
		ctx := context.Background()
		data := []byte("test image data")

		mockResizer := createMockResizer(resizer.ResizeResult{}, nil)
		resizers := map[string]resizer.Resizer{
			"simple": mockResizer,
		}
		thumb := setupThumbnail(resizers)

		factoryErr := errors.New("writer factory error")
		writerFactory := createMockWriterFactory(nil, factoryErr)

		formats := map[string]ThumbnailInput{
			"thumb": {
				Resizer: "simple",
				ResizeOptions: resizer.ResizeOptions{
					Width:  100,
					Height: 100,
					Format: ".jpg",
				},
			},
		}

		results, err := thumb.GenerateThumbnails(ctx, data, formats, writerFactory)

		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "failed to create writer for format")
		assert.Contains(t, err.Error(), "thumb")
	})

	t.Run("ResizerNotFound", func(t *testing.T) {
		ctx := context.Background()
		data := []byte("test image data")

		resizers := map[string]resizer.Resizer{}
		thumb := setupThumbnail(resizers)

		mockWriter := newMockWriteCloser()
		writers := map[string]io.WriteCloser{
			"thumb": mockWriter,
		}
		writerFactory := createMockWriterFactory(writers, nil)

		formats := map[string]ThumbnailInput{
			"thumb": {
				Resizer: "nonexistent",
				ResizeOptions: resizer.ResizeOptions{
					Width:  100,
					Height: 100,
					Format: ".jpg",
				},
			},
		}

		results, err := thumb.GenerateThumbnails(ctx, data, formats, writerFactory)

		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "failed to generate thumbnail for format")
		assert.Contains(t, err.Error(), "thumb")
		assert.Contains(t, err.Error(), "resizer")
		assert.Contains(t, err.Error(), "not found")
		assert.True(t, mockWriter.closed, "Writer should be closed on error")
	})

	t.Run("ResizeError", func(t *testing.T) {
		ctx := context.Background()
		data := []byte("test image data")

		resizeErr := errors.New("resize error")
		mockResizer := createMockResizer(resizer.ResizeResult{}, resizeErr)
		resizers := map[string]resizer.Resizer{
			"simple": mockResizer,
		}
		thumb := setupThumbnail(resizers)

		mockWriter := newMockWriteCloser()
		writers := map[string]io.WriteCloser{
			"thumb": mockWriter,
		}
		writerFactory := createMockWriterFactory(writers, nil)

		formats := map[string]ThumbnailInput{
			"thumb": {
				Resizer: "simple",
				ResizeOptions: resizer.ResizeOptions{
					Width:  100,
					Height: 100,
					Format: ".jpg",
				},
			},
		}

		results, err := thumb.GenerateThumbnails(ctx, data, formats, writerFactory)

		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "failed to generate thumbnail for format")
		assert.Contains(t, err.Error(), "thumb")
		assert.True(t, mockWriter.closed, "Writer should be closed on error")
	})

	t.Run("EmptyFormatsMap", func(t *testing.T) {
		ctx := context.Background()
		data := []byte("test image data")

		resizers := map[string]resizer.Resizer{
			"simple": createMockResizer(resizer.ResizeResult{}, nil),
		}
		thumb := setupThumbnail(resizers)

		writerFactory := createMockWriterFactory(nil, nil)
		formats := map[string]ThumbnailInput{}

		results, err := thumb.GenerateThumbnails(ctx, data, formats, writerFactory)

		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Empty(t, results)
	})

	t.Run("EmptyData", func(t *testing.T) {
		ctx := context.Background()
		data := []byte{}

		expectedResult := resizer.ResizeResult{
			Width:    100,
			Height:   100,
			FileSize: 0,
		}
		mockResizer := createMockResizer(expectedResult, nil)
		resizers := map[string]resizer.Resizer{
			"simple": mockResizer,
		}
		thumb := setupThumbnail(resizers)

		mockWriter := newMockWriteCloser()
		writers := map[string]io.WriteCloser{
			"thumb": mockWriter,
		}
		writerFactory := createMockWriterFactory(writers, nil)

		formats := map[string]ThumbnailInput{
			"thumb": {
				Resizer: "simple",
				ResizeOptions: resizer.ResizeOptions{
					Width:  100,
					Height: 100,
					Format: ".jpg",
				},
			},
		}

		results, err := thumb.GenerateThumbnails(ctx, data, formats, writerFactory)

		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results, 1)
	})
}
