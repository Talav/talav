package resizer

// Test Image Naming Convention:
// All test images are saved to testdata/output/ with the following naming pattern:
//   square-{test-name-snake-case-lowercase}-input-{width}x{height}.{ext}
//   square-{test-name-snake-case-lowercase}-output-{width}x{height}.{ext}
//
// All characters in filenames must be lowercase.
//
// Examples:
//   - square-width-greater-than-height-input-200x100.jpg
//   - square-width-greater-than-height-output-200x200.jpg
//   - square-different-formats-jpeg-input-200x100.jpg
//   - square-different-formats-jpeg-output-200x200.jpg
//
// This convention makes it clear which test generates which images for visual inspection.

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupResizer creates a new SquareResizer with ImagingCodec for testing.
func setupResizer(t *testing.T) *SquareResizer {
	t.Helper()
	codec := NewImagingCodec()

	return NewSquareResizer(codec)
}

// createTestImage creates a test image with a visible checkerboard pattern
// This makes it easier to visually verify resizing behavior.
func createTestImage(width, height int) *image.NRGBA {
	img := imaging.New(width, height, color.White)
	black := color.RGBA{0, 0, 0, 255}
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}

	// Calculate checker size based on image dimensions (aim for ~5-10 squares)
	checkerSize := max(width, height) / 8
	if checkerSize < 5 {
		checkerSize = 5
	}

	for y := range height {
		for x := range width {
			checkerX := x / checkerSize
			checkerY := y / checkerSize
			if (checkerX+checkerY)%2 == 0 {
				img.Set(x, y, black)
			} else {
				// Alternate between red and blue for more visibility
				if checkerX%2 == 0 {
					img.Set(x, y, red)
				} else {
					img.Set(x, y, blue)
				}
			}
		}
	}

	// Add a white border to make edges visible
	for x := range width {
		img.Set(x, 0, color.White)
		img.Set(x, height-1, color.White)
	}
	for y := range height {
		img.Set(0, y, color.White)
		img.Set(width-1, y, color.White)
	}

	return img
}

// encodeImageToBuffer encodes an image to a bytes.Buffer in JPEG format.
func encodeImageToBuffer(t *testing.T, img *image.NRGBA) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	err := imaging.Encode(&buf, img, imaging.JPEG)
	require.NoError(t, err)

	return &buf
}

// decodeAndVerifySquare decodes an image from buffer and verifies it's square with expected dimensions.
func decodeAndVerifySquare(t *testing.T, buf *bytes.Buffer, expectedSize int) {
	t.Helper()
	img, err := imaging.Decode(buf)
	require.NoError(t, err)
	bounds := img.Bounds()
	assert.Equal(t, expectedSize, bounds.Dx())
	assert.Equal(t, expectedSize, bounds.Dy())
	assert.Equal(t, bounds.Dx(), bounds.Dy(), "Output image should be square")
}

// saveImageToOutputDir saves an image to the testdata/output directory.
func saveImageToOutputDir(t *testing.T, img image.Image, filename string) string {
	t.Helper()
	// Try relative path first (when running from package directory)
	outputDir := filepath.Join("..", "..", "testdata", "output")
	// Check if directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		// If relative path doesn't exist, try from project root
		outputDir = filepath.Join("internal", "module", "media", "testdata", "output")
	}
	// Get absolute path
	absDir, err := filepath.Abs(outputDir)
	require.NoError(t, err)
	// Ensure directory exists - use absolute path for MkdirAll
	err = os.MkdirAll(absDir, 0o755)
	require.NoError(t, err, "Failed to create directory: %s", absDir)
	// Verify directory was created and is accessible
	info, err := os.Stat(absDir)
	require.NoError(t, err, "Directory should exist after MkdirAll: %s", absDir)
	require.True(t, info.IsDir(), "Path should be a directory: %s", absDir)
	outputPath := filepath.Join(absDir, filename)
	// Use absolute path for imaging.Save
	err = imaging.Save(img, outputPath)
	require.NoError(t, err, "Failed to save image to: %s", outputPath)
	t.Logf("Image saved to: %s", outputPath)

	return outputPath
}

func TestNewSquareResizer(t *testing.T) {
	resizer := setupResizer(t)

	require.NotNil(t, resizer)
	assert.NotNil(t, resizer.codec)
}

func TestSquareResizer_Resize_WidthGreaterThanHeight(t *testing.T) {
	resizer := setupResizer(t)

	// Create test image 200x100 with checkerboard pattern
	img := createTestImage(200, 100)
	inputBuf := encodeImageToBuffer(t, img)

	// Resize options
	opts := ResizeOptions{
		Width:  200,
		Height: 100,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, 200, result.Width)
	assert.Equal(t, 200, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify it's square
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 200, bounds.Dx())
	assert.Equal(t, 200, bounds.Dy())
	assert.Equal(t, bounds.Dx(), bounds.Dy(), "Output image should be square")

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "square-width-greater-than-height-input-200x100.jpg")
	saveImageToOutputDir(t, outputImg, "square-width-greater-than-height-output-200x200.jpg")
}

func TestSquareResizer_Resize_HeightGreaterThanWidth(t *testing.T) {
	resizer := setupResizer(t)

	// Create test image 100x200 with checkerboard pattern
	img := createTestImage(100, 200)
	inputBuf := encodeImageToBuffer(t, img)

	// Resize options
	opts := ResizeOptions{
		Width:  100,
		Height: 200,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, 200, result.Width)
	assert.Equal(t, 200, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify it's square
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 200, bounds.Dx())
	assert.Equal(t, 200, bounds.Dy())
	assert.Equal(t, bounds.Dx(), bounds.Dy(), "Output image should be square")

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "square-height-greater-than-width-input-100x200.jpg")
	saveImageToOutputDir(t, outputImg, "square-height-greater-than-width-output-200x200.jpg")
}

func TestSquareResizer_Resize_DifferentFormats(t *testing.T) {
	resizer := setupResizer(t)

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
				Width:  200,
				Height: 100,
				Format: tc.format,
			}

			var outputBuf bytes.Buffer
			result, err := resizer.Resize(context.Background(), &inputBuf, opts, &outputBuf)

			require.NoError(t, err)
			assert.Equal(t, 200, result.Width)
			assert.Equal(t, 200, result.Height)
			assert.Greater(t, result.FileSize, int64(0))

			// Decode output to verify it's valid and square
			outputImg, err := imaging.Decode(&outputBuf)
			require.NoError(t, err)
			bounds := outputImg.Bounds()
			assert.Equal(t, 200, bounds.Dx())
			assert.Equal(t, 200, bounds.Dy())
			assert.Equal(t, bounds.Dx(), bounds.Dy(), "Output image should be square")

			// Save input and output for visual inspection
			ext := tc.format
			formatName := strings.ToLower(tc.name)
			saveImageToOutputDir(t, img, fmt.Sprintf("square-different-formats-%s-input-200x100.jpg", formatName))
			saveImageToOutputDir(t, outputImg, fmt.Sprintf("square-different-formats-%s-output-200x200.%s", formatName, ext))
		})
	}
}

func TestSquareResizer_Resize_WithEncodingOptions(t *testing.T) {
	resizer := setupResizer(t)

	// Create test image with checkerboard pattern
	img := createTestImage(200, 100)
	inputBuf := encodeImageToBuffer(t, img)

	// Test with high quality
	optsHigh := ResizeOptions{
		Width:  200,
		Height: 100,
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
		Width:  200,
		Height: 100,
		Format: "jpg",
		Options: map[string]any{
			"jpeg_quality": 50,
		},
	}

	var outputBufLow bytes.Buffer
	resultLow, err := resizer.Resize(context.Background(), inputBuf, optsLow, &outputBufLow)
	require.NoError(t, err)

	// Verify both are square
	assert.Equal(t, 200, resultHigh.Width)
	assert.Equal(t, 200, resultHigh.Height)
	assert.Equal(t, 200, resultLow.Width)
	assert.Equal(t, 200, resultLow.Height)

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
	assert.Equal(t, 200, boundsHigh.Dx())
	assert.Equal(t, 200, boundsHigh.Dy())
	assert.Equal(t, boundsHigh.Dx(), boundsHigh.Dy(), "High quality output should be square")

	boundsLow := outputImgLow.Bounds()
	assert.Equal(t, 200, boundsLow.Dx())
	assert.Equal(t, 200, boundsLow.Dy())
	assert.Equal(t, boundsLow.Dx(), boundsLow.Dy(), "Low quality output should be square")

	// Save input and outputs for visual inspection
	saveImageToOutputDir(t, img, "square-with-encoding-options-input-200x100.jpg")
	saveImageToOutputDir(t, outputImgHigh, "square-with-encoding-options-output-high-quality-200x200.jpg")
	saveImageToOutputDir(t, outputImgLow, "square-with-encoding-options-output-low-quality-200x200.jpg")
}

func TestSquareResizer_Resize_DecodeError(t *testing.T) {
	resizer := setupResizer(t)

	// Invalid image data
	invalidData := bytes.NewReader([]byte("not an image"))
	opts := ResizeOptions{
		Width:  200,
		Height: 100,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), invalidData, opts, &outputBuf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode image")
	assert.Equal(t, ResizeResult{}, result)
}

func TestSquareResizer_Resize_EncodeError(t *testing.T) {
	resizer := setupResizer(t)

	// Create valid test image with checkerboard pattern
	img := createTestImage(200, 100)
	inputBuf := encodeImageToBuffer(t, img)

	// Invalid format
	opts := ResizeOptions{
		Width:  200,
		Height: 100,
		Format: "invalid",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to encode image")
	assert.Equal(t, ResizeResult{}, result)
}

func TestSquareResizer_Resize_InvalidImageData(t *testing.T) {
	resizer := setupResizer(t)

	// Random bytes that are not a valid image
	randomData := bytes.NewReader([]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0})
	opts := ResizeOptions{
		Width:  200,
		Height: 100,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), randomData, opts, &outputBuf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode image")
	assert.Equal(t, ResizeResult{}, result)
}

func TestSquareResizer_Resize_WithRealImageFile(t *testing.T) {
	resizer := setupResizer(t)

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

	// Calculate expected square size (max of requested width/height from opts)
	// SquareResizer uses max(opts.Width, opts.Height) to determine square size
	expectedSize := max(originalWidth, originalHeight)

	// Reset file pointer
	_, err = file.Seek(0, 0)
	require.NoError(t, err)

	// Resize options
	opts := ResizeOptions{
		Width:  originalWidth,
		Height: originalHeight,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), file, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, expectedSize, result.Width)
	assert.Equal(t, expectedSize, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Save bytes before decoding (decoding may consume the buffer)
	outputBytes := outputBuf.Bytes()

	// Decode output to verify it's square
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, expectedSize, bounds.Dx())
	assert.Equal(t, expectedSize, bounds.Dy())

	// Verify output is actually square
	assert.Equal(t, bounds.Dx(), bounds.Dy(), "Output image should be square")

	// Save input and output for visual inspection
	// Note: originalImg is already decoded from the file
	saveImageToOutputDir(t, originalImg, fmt.Sprintf("square-with-real-image-file-input-%dx%d.jpg", originalWidth, originalHeight))

	outputPath := filepath.Join("..", "..", "testdata", "output")
	err = os.MkdirAll(outputPath, 0o755)
	require.NoError(t, err)
	outputPath = filepath.Join(outputPath, fmt.Sprintf("square-with-real-image-file-output-%dx%d.jpg", expectedSize, expectedSize))
	err = os.WriteFile(outputPath, outputBytes, 0o644)
	require.NoError(t, err)
	t.Logf("Output image saved to: %s", outputPath)

	// Read back from disk and decode to verify it's a valid image
	// Use imaging.Open which handles file paths directly
	savedImg, err := imaging.Open(outputPath)
	require.NoError(t, err, "Failed to decode saved image file")

	// Verify dimensions match
	savedBounds := savedImg.Bounds()
	assert.Equal(t, expectedSize, savedBounds.Dx(), "Saved image width should match expected size")
	assert.Equal(t, expectedSize, savedBounds.Dy(), "Saved image height should match expected size")
	assert.Equal(t, savedBounds.Dx(), savedBounds.Dy(), "Saved image should be square")

	// Verify file size matches what we wrote
	fileInfo, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.Equal(t, result.FileSize, fileInfo.Size(), "File size on disk should match ResizeResult.FileSize")
}

func TestSquareResizer_Resize_SmallerImageThanRequested(t *testing.T) {
	resizer := setupResizer(t)

	// Create small test image 50x50 with checkerboard pattern
	// This will make it obvious when upscaled
	img := createTestImage(50, 50)
	inputBuf := encodeImageToBuffer(t, img)

	// Request larger dimensions (200x200)
	opts := ResizeOptions{
		Width:  200,
		Height: 200,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	// Should upscale to requested size (max of width/height = 200)
	assert.Equal(t, 200, result.Width)
	assert.Equal(t, 200, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify it's square and upscaled
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 200, bounds.Dx(), "Image should be upscaled to requested width")
	assert.Equal(t, 200, bounds.Dy(), "Image should be upscaled to requested height")
	assert.Equal(t, bounds.Dx(), bounds.Dy(), "Output image should be square")

	// Save both input and output for visual inspection
	saveImageToOutputDir(t, img, "square-smaller-image-than-requested-input-50x50.jpg")
	saveImageToOutputDir(t, outputImg, "square-smaller-image-than-requested-output-200x200.jpg")
}

func TestSquareResizer_Resize_SmallerNonSquareImage(t *testing.T) {
	resizer := setupResizer(t)

	// Create small non-square test image 50x30 with checkerboard pattern
	img := createTestImage(50, 30)
	inputBuf := encodeImageToBuffer(t, img)

	// Request larger square dimensions (200x200)
	opts := ResizeOptions{
		Width:  200,
		Height: 200,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	// Should upscale to requested size (max of width/height = 200)
	assert.Equal(t, 200, result.Width)
	assert.Equal(t, 200, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify it's square and upscaled
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 200, bounds.Dx(), "Image should be upscaled to requested width")
	assert.Equal(t, 200, bounds.Dy(), "Image should be upscaled to requested height")
	assert.Equal(t, bounds.Dx(), bounds.Dy(), "Output image should be square")

	// Save both input and output for visual inspection
	saveImageToOutputDir(t, img, "square-smaller-nonsquare-input-50x30.jpg")
	saveImageToOutputDir(t, outputImg, "square-smaller-nonsquare-output-200x200.jpg")
}

func TestSquareResizer_Resize_ValidImageRoundTrip(t *testing.T) {
	resizer := setupResizer(t)

	// Create test image 200x100 with checkerboard pattern
	img := createTestImage(200, 100)
	inputBuf := encodeImageToBuffer(t, img)

	// Resize options
	opts := ResizeOptions{
		Width:  200,
		Height: 100,
		Format: "jpg",
	}

	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)
	require.NoError(t, err)
	assert.Equal(t, 200, result.Width)
	assert.Equal(t, 200, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Save bytes before decoding (decoding may consume the buffer)
	outputBytes := outputBuf.Bytes()

	// Decode output to verify it's square
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, 200, bounds.Dx())
	assert.Equal(t, 200, bounds.Dy())
	assert.Equal(t, bounds.Dx(), bounds.Dy(), "Output image should be square")

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, "square-valid-image-roundtrip-input-200x100.jpg")
	saveImageToOutputDir(t, outputImg, "square-valid-image-roundtrip-output-200x200.jpg")

	// Save to disk for roundtrip verification
	outputDir := filepath.Join("..", "..", "testdata", "output")
	err = os.MkdirAll(outputDir, 0o755)
	require.NoError(t, err)
	outputPath := filepath.Join(outputDir, "square-valid-image-roundtrip-saved-200x200.jpg")
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
	assert.Equal(t, 200, decodedBounds.Dx(), "Decoded image width should be 200")
	assert.Equal(t, 200, decodedBounds.Dy(), "Decoded image height should be 200")
	assert.Equal(t, decodedBounds.Dx(), decodedBounds.Dy(), "Decoded image should be square")

	// Verify file size matches what we wrote
	fileInfo, err := file.Stat()
	require.NoError(t, err)
	assert.Equal(t, result.FileSize, fileInfo.Size(), "File size on disk should match ResizeResult.FileSize")
}
