package resizer

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

// testResize is a generic helper for testing resize operations.
// It handles the common pattern of creating test images, resizing, verifying results, and saving output.
func testResize(
	t *testing.T,
	resizer Resizer,
	testName string,
	prefix string, // "crop", "simple", or "square"
	inputWidth, inputHeight int,
	opts ResizeOptions,
	expectedWidth, expectedHeight int,
) {
	t.Helper()

	// Create test image
	img := createTestImage(inputWidth, inputHeight)
	inputBuf := encodeImageToBuffer(t, img)

	// Resize
	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	// Assert results
	require.NoError(t, err)
	assert.Equal(t, expectedWidth, result.Width)
	assert.Equal(t, expectedHeight, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, expectedWidth, bounds.Dx())
	assert.Equal(t, expectedHeight, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, fmt.Sprintf("%s-%s-input-%dx%d.jpg", prefix, testName, inputWidth, inputHeight))
	saveImageToOutputDir(t, outputImg, fmt.Sprintf("%s-%s-output-%dx%d.jpg", prefix, testName, expectedWidth, expectedHeight))
}

// testResizeWithRealImageFile tests resizing with a real image file from testdata.
func testResizeWithRealImageFile(
	t *testing.T,
	resizer Resizer,
	prefix string, // "crop", "simple", or "square"
	targetWidth, targetHeight int,
) {
	t.Helper()

	// Load real image file from testdata
	testImagePath := filepath.Join("..", "..", "testdata", "test-1.jpg")
	//nolint:gosec // Test file path is safe - it's a relative path within testdata
	file, err := os.Open(testImagePath)
	require.NoError(t, err, "Failed to open test image file")
	defer func() { _ = file.Close() }()

	// Decode original image to get dimensions
	originalImg, err := imaging.Decode(file)
	require.NoError(t, err)
	originalBounds := originalImg.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

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
	saveImageToOutputDir(t, originalImg, fmt.Sprintf("%s-with-real-image-file-input-%dx%d.jpg", prefix, originalWidth, originalHeight))

	outputPath := filepath.Join("..", "..", "testdata", "output")
	err = os.MkdirAll(outputPath, 0o750)
	require.NoError(t, err)
	outputPath = filepath.Join(outputPath, fmt.Sprintf("%s-with-real-image-file-output-%dx%d.jpg", prefix, targetWidth, targetHeight))
	err = os.WriteFile(outputPath, outputBytes, 0o600)
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

// testResizeValidImageRoundTrip tests resizing and roundtrip verification.
func testResizeValidImageRoundTrip(
	t *testing.T,
	resizer Resizer,
	prefix string, // "crop", "simple", or "square"
	inputWidth, inputHeight int,
	opts ResizeOptions,
	expectedWidth, expectedHeight int,
) {
	t.Helper()

	// Create test image
	img := createTestImage(inputWidth, inputHeight)
	inputBuf := encodeImageToBuffer(t, img)

	// Resize
	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)
	require.NoError(t, err)
	assert.Equal(t, expectedWidth, result.Width)
	assert.Equal(t, expectedHeight, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Save bytes before decoding (decoding may consume the buffer)
	outputBytes := outputBuf.Bytes()

	// Decode output to verify dimensions
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, expectedWidth, bounds.Dx())
	assert.Equal(t, expectedHeight, bounds.Dy())

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, fmt.Sprintf("%s-valid-image-roundtrip-input-%dx%d.jpg", prefix, inputWidth, inputHeight))
	saveImageToOutputDir(t, outputImg, fmt.Sprintf("%s-valid-image-roundtrip-output-%dx%d.jpg", prefix, expectedWidth, expectedHeight))

	// Save to disk for roundtrip verification
	outputDir := filepath.Join("..", "..", "testdata", "output")
	err = os.MkdirAll(outputDir, 0o750)
	require.NoError(t, err)
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s-valid-image-roundtrip-saved-%dx%d.jpg", prefix, expectedWidth, expectedHeight))
	err = os.WriteFile(outputPath, outputBytes, 0o600)
	require.NoError(t, err)

	// Read back from disk
	//nolint:gosec // Test file path is safe - it's a relative path within testdata/output
	file, err := os.Open(outputPath)
	require.NoError(t, err, "Failed to open saved image file")
	defer func() { _ = file.Close() }()

	// Decode the file to verify it's a valid image
	decodedImg, err := imaging.Decode(file)
	require.NoError(t, err, "Failed to decode saved image file")

	// Verify dimensions
	decodedBounds := decodedImg.Bounds()
	assert.Equal(t, expectedWidth, decodedBounds.Dx(), "Decoded image width should be %d", expectedWidth)
	assert.Equal(t, expectedHeight, decodedBounds.Dy(), "Decoded image height should be %d", expectedHeight)

	// Verify file size matches what we wrote
	fileInfo, err := file.Stat()
	require.NoError(t, err)
	assert.Equal(t, result.FileSize, fileInfo.Size(), "File size on disk should match ResizeResult.FileSize")
}

// testResizeDifferentFormats tests resizing with different image formats.
func testResizeDifferentFormats(
	t *testing.T,
	resizer Resizer,
	prefix string, // "crop", "simple", or "square"
) {
	t.Helper()

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
			saveImageToOutputDir(t, img, fmt.Sprintf("%s-different-formats-%s-input-200x100.jpg", prefix, formatName))
			saveImageToOutputDir(t, outputImg, fmt.Sprintf("%s-different-formats-%s-output-150x75.%s", prefix, formatName, ext))
		})
	}
}

// testResizeWithEncodingOptions tests resizing with different encoding options.
func testResizeWithEncodingOptions(
	t *testing.T,
	resizer Resizer,
	prefix string, // "crop", "simple", or "square"
) {
	t.Helper()

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
	saveImageToOutputDir(t, img, fmt.Sprintf("%s-with-encoding-options-input-200x100.jpg", prefix))
	saveImageToOutputDir(t, outputImgHigh, fmt.Sprintf("%s-with-encoding-options-output-high-quality-150x75.jpg", prefix))
	saveImageToOutputDir(t, outputImgLow, fmt.Sprintf("%s-with-encoding-options-output-low-quality-150x75.jpg", prefix))
}

// testSquareResize tests resizing with square resizer and verifies output is square.
//
//nolint:unparam // expectedSize parameter allows for flexibility in test cases
func testSquareResize(
	t *testing.T,
	resizer Resizer,
	testName string,
	inputWidth, inputHeight int,
	opts ResizeOptions,
	expectedSize int,
) {
	t.Helper()

	// Create test image
	img := createTestImage(inputWidth, inputHeight)
	inputBuf := encodeImageToBuffer(t, img)

	// Resize
	var outputBuf bytes.Buffer
	result, err := resizer.Resize(context.Background(), inputBuf, opts, &outputBuf)

	require.NoError(t, err)
	assert.Equal(t, expectedSize, result.Width)
	assert.Equal(t, expectedSize, result.Height)
	assert.Greater(t, result.FileSize, int64(0))

	// Decode output to verify it's square
	outputImg, err := imaging.Decode(&outputBuf)
	require.NoError(t, err)
	bounds := outputImg.Bounds()
	assert.Equal(t, expectedSize, bounds.Dx())
	assert.Equal(t, expectedSize, bounds.Dy())
	assert.Equal(t, bounds.Dx(), bounds.Dy(), "Output image should be square")

	// Save input and output for visual inspection
	saveImageToOutputDir(t, img, fmt.Sprintf("square-%s-input-%dx%d.jpg", testName, inputWidth, inputHeight))
	saveImageToOutputDir(t, outputImg, fmt.Sprintf("square-%s-output-%dx%d.jpg", testName, expectedSize, expectedSize))
}
