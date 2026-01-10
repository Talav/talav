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
	"os"
	"path/filepath"
	"strings"
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

func TestCropResizer_Resize_ExactDimensions(t *testing.T) {
	resizer := setupCropResizer(t)

	// Create test image 200x100 with checkerboard pattern
	img := createTestImage(200, 100)
	inputBuf := encodeImageToBuffer(t, img)

	// Resize to exact dimensions
	opts := ResizeOptions{
		Width:  150,
		Height: 75,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, 150, result.Width)
	assert.Equal(t, 75, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 150, bounds.Dx())
	assert.Equal(t, 75, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "crop-exact-dimensions-input-200x100.jpg")
	saveImageToOutputDir(t, outputImg, "crop-exact-dimensions-output-150x75.jpg")
}

func TestCropResizer_Resize_WiderInputToNarrowerOutput(t *testing.T) {
	resizer := setupCropResizer(t)

	// Create wider test image 400x200 (2:1 aspect ratio) with checkerboard pattern
	img := createTestImage(400, 200)
	inputBuf := encodeImageToBuffer(t, img)

	// Crop to square (1:1 aspect ratio) - should crop sides
	opts := ResizeOptions{
		Width:  200,
		Height: 200,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, 200, result.Width)
	assert.Equal(t, 200, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 200, bounds.Dx())
	assert.Equal(t, 200, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "crop-wider-input-to-narrower-output-input-400x200.jpg")
	saveImageToOutputDir(t, outputImg, "crop-wider-input-to-narrower-output-output-200x200.jpg")
}

func TestCropResizer_Resize_TallerInputToWiderOutput(t *testing.T) {
	resizer := setupCropResizer(t)

	// Create taller test image 200x400 (1:2 aspect ratio) with checkerboard pattern
	img := createTestImage(200, 400)
	inputBuf := encodeImageToBuffer(t, img)

	// Crop to wider output (2:1 aspect ratio) - should crop top/bottom
	opts := ResizeOptions{
		Width:  400,
		Height: 200,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, 400, result.Width)
	assert.Equal(t, 200, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 400, bounds.Dx())
	assert.Equal(t, 200, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "crop-taller-input-to-wider-output-input-200x400.jpg")
	saveImageToOutputDir(t, outputImg, "crop-taller-input-to-wider-output-output-400x200.jpg")
}

func TestCropResizer_Resize_SquareInputToRectangular(t *testing.T) {
	resizer := setupCropResizer(t)

	// Create square test image 200x200 with checkerboard pattern
	img := createTestImage(200, 200)
	inputBuf := encodeImageToBuffer(t, img)

	// Crop to rectangular output (2:1 aspect ratio) - should crop top/bottom
	opts := ResizeOptions{
		Width:  300,
		Height: 150,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, 300, result.Width)
	assert.Equal(t, 150, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 300, bounds.Dx())
	assert.Equal(t, 150, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "crop-square-input-to-rectangular-input-200x200.jpg")
	saveImageToOutputDir(t, outputImg, "crop-square-input-to-rectangular-output-300x150.jpg")
}

func TestCropResizer_Resize_RectangularToSquare(t *testing.T) {
	resizer := setupCropResizer(t)

	// Create rectangular test image 400x200 (2:1 aspect ratio) with checkerboard pattern
	img := createTestImage(400, 200)
	inputBuf := encodeImageToBuffer(t, img)

	// Crop to square output - should crop sides
	opts := ResizeOptions{
		Width:  200,
		Height: 200,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, 200, result.Width)
	assert.Equal(t, 200, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 200, bounds.Dx())
	assert.Equal(t, 200, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "crop-rectangular-to-square-input-400x200.jpg")
	saveImageToOutputDir(t, outputImg, "crop-rectangular-to-square-output-200x200.jpg")
}

func TestCropResizer_Resize_Downscale(t *testing.T) {
	resizer := setupCropResizer(t)

	// Create test image 800x600 with checkerboard pattern
	img := createTestImage(800, 600)
	inputBuf := encodeImageToBuffer(t, img)

	// Downscale with cropping
	opts := ResizeOptions{
		Width:  200,
		Height: 150,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, 200, result.Width)
	assert.Equal(t, 150, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 200, bounds.Dx())
	assert.Equal(t, 150, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "crop-downscale-input-800x600.jpg")
	saveImageToOutputDir(t, outputImg, "crop-downscale-output-200x150.jpg")
}

func TestCropResizer_Resize_Upscale(t *testing.T) {
	resizer := setupCropResizer(t)

	// Create test image 200x150 with checkerboard pattern
	img := createTestImage(200, 150)
	inputBuf := encodeImageToBuffer(t, img)

	// Upscale with cropping
	opts := ResizeOptions{
		Width:  800,
		Height: 600,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, 800, result.Width)
	assert.Equal(t, 600, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 800, bounds.Dx())
	assert.Equal(t, 600, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "crop-upscale-input-200x150.jpg")
	saveImageToOutputDir(t, outputImg, "crop-upscale-output-800x600.jpg")
}

func TestCropResizer_Resize_SameAspectRatio(t *testing.T) {
	resizer := setupCropResizer(t)

	// Create test image 400x200 (2:1 aspect ratio) with checkerboard pattern
	img := createTestImage(400, 200)
	inputBuf := encodeImageToBuffer(t, img)

	// Resize to same aspect ratio (2:1) - should not crop, just resize
	opts := ResizeOptions{
		Width:  200,
		Height: 100,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, 200, result.Width)
	assert.Equal(t, 100, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 200, bounds.Dx())
	assert.Equal(t, 100, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "crop-same-aspect-ratio-input-400x200.jpg")
	saveImageToOutputDir(t, outputImg, "crop-same-aspect-ratio-output-200x100.jpg")
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

	testCases := []struct {
		name   string
		format string
	}{
		{"JPEG", "jpg"},
		{"PNG", "png"},
		{"GIF", "gif"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test image 200x100 with checkerboard pattern
			img := createTestImage(200, 100)
			var inputBuf bytes.Buffer
			err := imaging.Encode(&inputBuf, img, imaging.JPEG)
			require.NoError(t, err)

			opts := ResizeOptions{
				Width:  150,
				Height: 75,
				Format: tc.format,
			}

			var outputBuf bytes.Buffer
			result, err := resizer.Resize(context.Background(), &inputBuf, opts, &outputBuf)

			require.NoError(t, err)
			assert.Equal(t, 150, result.Width)
			assert.Equal(t, 75, result.Height)
			assert.Greater(t, result.FileSize, int64(0))

			// Decode output to verify it's valid
			outputImg, err := imaging.Decode(&outputBuf)
			require.NoError(t, err)
			bounds := outputImg.Bounds()
			assert.Equal(t, 150, bounds.Dx())
			assert.Equal(t, 75, bounds.Dy())

			// Save input and output for visual inspection
			ext := tc.format
			formatName := strings.ToLower(tc.name)
			saveImageToOutputDir(t, img, fmt.Sprintf("crop-different-formats-%s-input-200x100.jpg", formatName))
			saveImageToOutputDir(t, outputImg, fmt.Sprintf("crop-different-formats-%s-output-150x75.%s", formatName, ext))
		})
	}
}

func TestCropResizer_Resize_WithEncodingOptions(t *testing.T) {
	resizer := setupCropResizer(t)

	// Create test image with checkerboard pattern
	img := createTestImage(200, 100)
	inputBuf := encodeImageToBuffer(t, img)

	// Test with high quality
	optsHigh := ResizeOptions{
		Width:  150,
		Height: 75,
		Format: "jpg",
		Options: map[string]any{
			"jpeg_quality": 95,
		},
	}

	var outputBufHigh bytes.Buffer
	resultHigh, err := resizer.Resize(context.Background(), inputBuf, optsHigh, &outputBufHigh)
	require.NoError(t, err)

	// Reset input buffer
	inputBuf = encodeImageToBuffer(t, img)

	// Test with low quality
	optsLow := ResizeOptions{
		Width:  150,
		Height: 75,
		Format: "jpg",
		Options: map[string]any{
			"jpeg_quality": 50,
		},
	}

	var outputBufLow bytes.Buffer
	resultLow, err := resizer.Resize(context.Background(), inputBuf, optsLow, &outputBufLow)
	require.NoError(t, err)

	// Verify both have correct dimensions
	assert.Equal(t, 150, resultHigh.Width)
	assert.Equal(t, 75, resultHigh.Height)
	assert.Equal(t, 150, resultLow.Width)
	assert.Equal(t, 75, resultLow.Height)

	// The important thing is that encoding options are passed through correctly
	assert.Greater(t, resultHigh.FileSize, int64(0))
	assert.Greater(t, resultLow.FileSize, int64(0))
	// Low quality should be smaller or equal (allowing for edge cases with simple images)
	assert.GreaterOrEqual(t, resultHigh.FileSize, resultLow.FileSize, "High quality should produce same or larger file than low quality")

	// Verify both outputs are valid images
	outputImgHigh, err := imaging.Decode(&outputBufHigh)
	require.NoError(t, err)
	outputImgLow, err := imaging.Decode(&outputBufLow)
	require.NoError(t, err)

	boundsHigh := outputImgHigh.Bounds()
	assert.Equal(t, 150, boundsHigh.Dx())
	assert.Equal(t, 75, boundsHigh.Dy())

	boundsLow := outputImgLow.Bounds()
	assert.Equal(t, 150, boundsLow.Dx())
	assert.Equal(t, 75, boundsLow.Dy())

	// Save input and outputs for visual inspection
	saveImageToOutputDir(t, img, "crop-with-encoding-options-input-200x100.jpg")
	saveImageToOutputDir(t, outputImgHigh, "crop-with-encoding-options-output-high-quality-150x75.jpg")
	saveImageToOutputDir(t, outputImgLow, "crop-with-encoding-options-output-low-quality-150x75.jpg")
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

	// Load real image file from testdata
	testImagePath := filepath.Join("..", "..", "testdata", "test-1.jpg")
	file, err := os.Open(testImagePath)
	require.NoError(t, err, "Failed to open test image file")
	defer func() { _ = file.Close() }()

	// Decode original image to get dimensions
	originalImg, err := imaging.Decode(file)
	require.NoError(t, err)
	originalBounds := originalImg.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	// Resize to smaller dimensions with different aspect ratio (will crop)
	targetWidth := 800
	targetHeight := 600

	// Reset file pointer
	_, err = file.Seek(0, 0)
	require.NoError(t, err)

	// Resize options
	opts := ResizeOptions{
		Width:  targetWidth,
		Height: targetHeight,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), file, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, targetWidth, result.Width)
	assert.Equal(t, targetHeight, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Save bytes before decoding (decoding may consume the buffer)
	outputBytes := outputBuf.Bytes()

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, targetWidth, bounds.Dx())
	assert.Equal(t, targetHeight, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, originalImg, fmt.Sprintf("crop-with-real-image-file-input-%dx%d.jpg", originalWidth, originalHeight))

	outputPath := filepath.Join("..", "..", "testdata", "output")
	err = os.MkdirAll(outputPath, 0o755)
	require.NoError(t, err)
	outputPath = filepath.Join(outputPath, fmt.Sprintf("crop-with-real-image-file-output-%dx%d.jpg", targetWidth, targetHeight))
	err = os.WriteFile(outputPath, outputBytes, 0o644)
	require.NoError(t, err)
	t.Logf("Output image saved to: %s", outputPath)

	// Read back from disk and decode to verify it's a valid image
	savedImg, err := imaging.Open(outputPath)
	require.NoError(t, err, "Failed to decode saved image file")

	// Verify dimensions match
	savedBounds := savedImg.Bounds()
	assert.Equal(t, targetWidth, savedBounds.Dx(), "Saved image width should match expected size")
	assert.Equal(t, targetHeight, savedBounds.Dy(), "Saved image height should match expected size")

	// Verify file size matches what we wrote
	fileInfo, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.Equal(t, result.FileSize, fileInfo.Size(), "File size on disk should match ResizeResult.FileSize")
}

func TestCropResizer_Resize_ValidImageRoundTrip(t *testing.T) {
	resizer := setupCropResizer(t)

	// Create test image 200x100 with checkerboard pattern
	img := createTestImage(200, 100)
	inputBuf := encodeImageToBuffer(t, img)

	// Resize options
	opts := ResizeOptions{
		Width:  150,
		Height: 75,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)
	require.NoError(t, err)
	assert.Equal(t, 150, result.Width)
	assert.Equal(t, 75, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Save bytes before decoding (decoding may consume the buffer)
	outputBytes := outputBuf.Bytes()

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 150, bounds.Dx())
	assert.Equal(t, 75, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "crop-valid-image-roundtrip-input-200x100.jpg")
	saveImageToOutputDir(t, outputImg, "crop-valid-image-roundtrip-output-150x75.jpg")

	// Save to disk for roundtrip verification
	outputDir := filepath.Join("..", "..", "testdata", "output")
	err = os.MkdirAll(outputDir, 0o755)
	require.NoError(t, err)
	outputPath := filepath.Join(outputDir, "crop-valid-image-roundtrip-saved-150x75.jpg")
	err = os.WriteFile(outputPath, outputBytes, 0o644)
	require.NoError(t, err)

	// Read back from disk
	file, err := os.Open(outputPath)
	require.NoError(t, err, "Failed to open saved image file")
	defer func() { _ = file.Close() }()

	// Decode the file to verify it's a valid image
	decodedImg, err := imaging.Decode(file)
	require.NoError(t, err, "Failed to decode saved image file")

	// Verify dimensions
	decodedBounds := decodedImg.Bounds()
	assert.Equal(t, 150, decodedBounds.Dx(), "Decoded image width should be 150")
	assert.Equal(t, 75, decodedBounds.Dy(), "Decoded image height should be 75")

	// Verify file size matches what we wrote
	fileInfo, err := file.Stat()
	require.NoError(t, err)
	assert.Equal(t, result.FileSize, fileInfo.Size(), "File size on disk should match ResizeResult.FileSize")
}
