package resizer

// Test Image Naming Convention:
// All test images are saved to testdata/output/ with the following naming pattern:
//   crop-{test-name-snake-case-lowercase}-input-{width}x{height}.{ext}
//   crop-{test-name-snake-case-lowercase}-output-{width}x{height}.{ext}
//
// All characters in filenames must be lowercase.
//
// Examples:
//   - crop-exact-dimensions-input-200x100.jpg
//   - crop-exact-dimensions-output-150x75.jpg
//   - crop-wider-input-to-narrower-output-input-400x200.jpg
//   - crop-wider-input-to-narrower-output-output-200x200.jpg
//
// This convention makes it clear which test generates which images for visual inspection.

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupCropResizer creates a new CropResizer with ImagingCodec for testing.
func setupCropResizer(t *testing.T) *CropResizer {
	t.Helper()
	codec := NewImagingCodec()

	return NewCropResizer(codec)
}

func TestNewCropResizer(t *testing.T) {
	resizer := setupCropResizer(t)

	require.NotNil(t, resizer)
	assert.NotNil(t, resizer.codec)
}

func TestCropResizer_Resize_BasicCases(t *testing.T) {
	resizer := setupCropResizer(t)

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
			name:           "wider-input-to-narrower-output",
			inputWidth:     400,
			inputHeight:    200,
			opts:           ResizeOptions{Width: 200, Height: 200, Format: "jpg"},
			expectedWidth:  200,
			expectedHeight: 200,
		},
		{
			name:           "taller-input-to-wider-output",
			inputWidth:     200,
			inputHeight:    400,
			opts:           ResizeOptions{Width: 400, Height: 200, Format: "jpg"},
			expectedWidth:  400,
			expectedHeight: 200,
		},
		{
			name:           "square-input-to-rectangular",
			inputWidth:     200,
			inputHeight:    200,
			opts:           ResizeOptions{Width: 300, Height: 150, Format: "jpg"},
			expectedWidth:  300,
			expectedHeight: 150,
		},
		{
			name:           "rectangular-to-square",
			inputWidth:     400,
			inputHeight:    200,
			opts:           ResizeOptions{Width: 200, Height: 200, Format: "jpg"},
			expectedWidth:  200,
			expectedHeight: 200,
		},
		{
			name:           "downscale",
			inputWidth:     800,
			inputHeight:    600,
			opts:           ResizeOptions{Width: 200, Height: 150, Format: "jpg"},
			expectedWidth:  200,
			expectedHeight: 150,
		},
		{
			name:           "upscale",
			inputWidth:     200,
			inputHeight:    150,
			opts:           ResizeOptions{Width: 800, Height: 600, Format: "jpg"},
			expectedWidth:  800,
			expectedHeight: 600,
		},
		{
			name:           "same-aspect-ratio",
			inputWidth:     400,
			inputHeight:    200,
			opts:           ResizeOptions{Width: 200, Height: 100, Format: "jpg"},
			expectedWidth:  200,
			expectedHeight: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testResize(t, resizer, tt.name, "crop",
				tt.inputWidth, tt.inputHeight,
				tt.opts,
				tt.expectedWidth, tt.expectedHeight,
			)
		})
	}
}

func TestCropResizer_Resize_DifferentAspectRatios(t *testing.T) {
	resizer := setupCropResizer(t)

	testCases := []struct {
		name         string
		inputWidth   int
		inputHeight  int
		outputWidth  int
		outputHeight int
		inputAspect  string
		outputAspect string
	}{
		{
			name:         "wide_to_portrait",
			inputWidth:   600,
			inputHeight:  200,
			outputWidth:  200,
			outputHeight: 400,
			inputAspect:  "3:1",
			outputAspect: "1:2",
		},
		{
			name:         "portrait_to_wide",
			inputWidth:   200,
			inputHeight:  600,
			outputWidth:  600,
			outputHeight: 200,
			inputAspect:  "1:3",
			outputAspect: "3:1",
		},
		{
			name:         "wide_to_square",
			inputWidth:   600,
			inputHeight:  200,
			outputWidth:  200,
			outputHeight: 200,
			inputAspect:  "3:1",
			outputAspect: "1:1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test image with checkerboard pattern
			img := createTestImage(tc.inputWidth, tc.inputHeight)
			inputBuf := encodeImageToBuffer(t, img)

			opts := ResizeOptions{
				Width:  tc.outputWidth,
				Height: tc.outputHeight,
				Format: "jpg",
			}

			var outputBuf bytes.Buffer
			result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

			require.NoError(t, err)
			assert.Equal(t, tc.outputWidth, result.Width)
			assert.Equal(t, tc.outputHeight, result.Height)
			assert.Greater(t, result.FileSize, int64(0))

			// Decode output to verify dimensions
			outputImg, err := imaging.Decode(&outputBuf)
			require.NoError(t, err)
			bounds := outputImg.Bounds()
			assert.Equal(t, tc.outputWidth, bounds.Dx())
			assert.Equal(t, tc.outputHeight, bounds.Dy())

			// Save input and output for visual inspection
			saveImageToOutputDir(t, img, fmt.Sprintf("crop-different-aspect-ratios-%s-input-%dx%d.jpg", tc.name, tc.inputWidth, tc.inputHeight))
			saveImageToOutputDir(t, outputImg, fmt.Sprintf("crop-different-aspect-ratios-%s-output-%dx%d.jpg", tc.name, tc.outputWidth, tc.outputHeight))
		})
	}
}

func TestCropResizer_Resize_DifferentFormats(t *testing.T) {
	resizer := setupCropResizer(t)
	testResizeDifferentFormats(t, resizer, "crop")
}

func TestCropResizer_Resize_WithEncodingOptions(t *testing.T) {
	resizer := setupCropResizer(t)
	testResizeWithEncodingOptions(t, resizer, "crop")
}

func TestCropResizer_Resize_DecodeError(t *testing.T) {
	resizer := setupCropResizer(t)

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

func TestCropResizer_Resize_EncodeError(t *testing.T) {
	resizer := setupCropResizer(t)

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

func TestCropResizer_Resize_InvalidImageData(t *testing.T) {
	resizer := setupCropResizer(t)

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

func TestCropResizer_Resize_WithRealImageFile(t *testing.T) {
	resizer := setupCropResizer(t)
	testResizeWithRealImageFile(t, resizer, "crop", 800, 600)
}

func TestCropResizer_Resize_ValidImageRoundTrip(t *testing.T) {
	resizer := setupCropResizer(t)
	testResizeValidImageRoundTrip(t, resizer, "crop",
		200, 100,
		ResizeOptions{Width: 150, Height: 75, Format: "jpg"},
		150, 75,
	)
}
