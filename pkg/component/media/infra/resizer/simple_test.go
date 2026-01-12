package resizer

// Test Image Naming Convention:
// All test images are saved to testdata/output/ with the following naming pattern:
//   simple-{test-name-snake-case-lowercase}-input-{width}x{height}.{ext}
//   simple-{test-name-snake-case-lowercase}-output-{width}x{height}.{ext}
//
// All characters in filenames must be lowercase.
//
// Examples:
//   - simple-exact-dimensions-input-200x100.jpg
//   - simple-exact-dimensions-output-150x75.jpg
//   - simple-different-formats-jpeg-input-200x100.jpg
//   - simple-different-formats-jpeg-output-150x75.jpg
//
// This convention makes it clear which test generates which images for visual inspection.

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSimpleResizer creates a new SimpleResizer with ImagingCodec for testing.
func setupSimpleResizer(t *testing.T) *SimpleResizer {
	t.Helper()
	codec := NewImagingCodec()

	return NewSimpleResizer(codec)
}

func TestNewSimpleResizer(t *testing.T) {
	resizer := setupSimpleResizer(t)

	require.NotNil(t, resizer)
	assert.NotNil(t, resizer.codec)
}

func TestSimpleResizer_Resize_BasicCases(t *testing.T) {
	resizer := setupSimpleResizer(t)

	tests := []struct {
		name           string
		inputWidth     int
		inputHeight    int
		opts           ResizeOptions
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "exact-dimensions",
			inputWidth:     200,
			inputHeight:    100,
			opts:           ResizeOptions{Width: 150, Height: 75, Format: "jpg"},
			expectedWidth:  150,
			expectedHeight: 75,
		},
		{
			name:           "downscale",
			inputWidth:     400,
			inputHeight:    300,
			opts:           ResizeOptions{Width: 200, Height: 150, Format: "jpg"},
			expectedWidth:  200,
			expectedHeight: 150,
		},
		{
			name:           "upscale",
			inputWidth:     100,
			inputHeight:    75,
			opts:           ResizeOptions{Width: 300, Height: 225, Format: "jpg"},
			expectedWidth:  300,
			expectedHeight: 225,
		},
		{
			name:           "square-input",
			inputWidth:     200,
			inputHeight:    200,
			opts:           ResizeOptions{Width: 300, Height: 150, Format: "jpg"},
			expectedWidth:  300,
			expectedHeight: 150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testResize(t, resizer, tt.name, "simple",
				tt.inputWidth, tt.inputHeight,
				tt.opts,
				tt.expectedWidth, tt.expectedHeight,
			)
		})
	}
}

func TestSimpleResizer_Resize_PreserveAspectRatio_WidthZero(t *testing.T) {
	resizer := setupSimpleResizer(t)
	// When width=0, preserve aspect ratio based on height
	// Expected width: 150 * (200/100) = 300
	testResize(t, resizer, "preserve-aspect-ratio-width-zero", "simple",
		200, 100,
		ResizeOptions{Width: 0, Height: 150, Format: "jpg"},
		300, 150,
	)
}

func TestSimpleResizer_Resize_PreserveAspectRatio_HeightZero(t *testing.T) {
	resizer := setupSimpleResizer(t)
	// When height=0, preserve aspect ratio based on width
	// Expected height: 400 * (100/200) = 200
	testResize(t, resizer, "preserve-aspect-ratio-height-zero", "simple",
		200, 100,
		ResizeOptions{Width: 400, Height: 0, Format: "jpg"},
		400, 200,
	)
}

func TestSimpleResizer_Resize_DifferentFormats(t *testing.T) {
	resizer := setupSimpleResizer(t)
	testResizeDifferentFormats(t, resizer, "simple")
}

func TestSimpleResizer_Resize_WithEncodingOptions(t *testing.T) {
	resizer := setupSimpleResizer(t)
	testResizeWithEncodingOptions(t, resizer, "simple")
}

func TestSimpleResizer_Resize_DecodeError(t *testing.T) {
	resizer := setupSimpleResizer(t)

	// Invalid image data
	invalidData := bytes.NewReader([]byte("not an image"))
	opts := ResizeOptions{
		Width:  150,
		Height: 75,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), invalidData, opts, &outputBuf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode image")
	assert.Equal(t, ResizeResult{}, result)
}

func TestSimpleResizer_Resize_EncodeError(t *testing.T) {
	resizer := setupSimpleResizer(t)

	// Create valid test image with checkerboard pattern
	img := createTestImage(200, 100)
	inputBuf := encodeImageToBuffer(t, img)

	// Invalid format
	opts := ResizeOptions{
		Width:  150,
		Height: 75,
		Format: "invalid",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to encode image")
	assert.Equal(t, ResizeResult{}, result)
}

func TestSimpleResizer_Resize_InvalidImageData(t *testing.T) {
	resizer := setupSimpleResizer(t)

	// Random bytes that are not a valid image
	randomData := bytes.NewReader([]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0})
	opts := ResizeOptions{
		Width:  150,
		Height: 75,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), randomData, opts, &outputBuf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode image")
	assert.Equal(t, ResizeResult{}, result)
}

func TestSimpleResizer_Resize_WithRealImageFile(t *testing.T) {
	resizer := setupSimpleResizer(t)
	testResizeWithRealImageFile(t, resizer, "simple", 800, 600)
}

func TestSimpleResizer_Resize_ValidImageRoundTrip(t *testing.T) {
	resizer := setupSimpleResizer(t)
	testResizeValidImageRoundTrip(t, resizer, "simple",
		200, 100,
		ResizeOptions{Width: 150, Height: 75, Format: "jpg"},
		150, 75,
	)
}
