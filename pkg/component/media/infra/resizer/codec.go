package resizer

import (
	"fmt"
	"image"
	"image/png"
	"io"

	"github.com/disintegration/imaging"
)

// ImageCodec defines the interface for image encoding and decoding operations
// This abstraction allows using any image processing library.
type ImageCodec interface {
	// Decode decodes an image from a reader
	Decode(reader io.Reader) (image.Image, error)
	// Encode encodes an image to writer based on file extension
	// encodeOpts is a map of encoding options (e.g., "jpeg_quality": int, "png_compression_level": int)
	Encode(writer io.Writer, img image.Image, ext string, encodeOpts map[string]any) error
}

// ImagingCodec is the default implementation of ImageCodec using the imaging library.
type ImagingCodec struct{}

// NewImagingCodec creates a new ImagingCodec.
func NewImagingCodec() *ImagingCodec {
	return &ImagingCodec{}
}

// Decode decodes an image from a reader using imaging library.
func (c *ImagingCodec) Decode(reader io.Reader) (image.Image, error) {
	return imaging.Decode(reader)
}

// Encode encodes an image to writer based on file extension using imaging library.
func (c *ImagingCodec) Encode(writer io.Writer, img image.Image, ext string, encodeOpts map[string]any) error {
	format, err := imaging.FormatFromExtension(ext)
	if err != nil {
		return fmt.Errorf("unsupported image format: %w", err)
	}

	// Convert options map to imaging.EncodeOption slice
	opts := buildEncodeOptions(encodeOpts)

	if err := imaging.Encode(writer, img, format, opts...); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}

// buildEncodeOptions converts a map of encoding options to imaging.EncodeOption slice
// Supported options:
//   - "jpeg_quality": int (1-100, default 95)
//   - "png_compression_level": int (0-9, maps to png.CompressionLevel)
//   - "gif_num_colors": int (1-256, default 256)
func buildEncodeOptions(opts map[string]any) []imaging.EncodeOption {
	var encodeOpts []imaging.EncodeOption
	if opts == nil {
		return encodeOpts
	}

	// JPEG quality (1-100)
	if quality, ok := opts["jpeg_quality"].(int); ok && quality >= 1 && quality <= 100 {
		encodeOpts = append(encodeOpts, imaging.JPEGQuality(quality))
	}

	// PNG compression level (0-9, maps to png.CompressionLevel)
	if level, ok := opts["png_compression_level"].(int); ok {
		var compressionLevel png.CompressionLevel
		switch {
		case level <= 0:
			compressionLevel = png.NoCompression
		case level == 1:
			compressionLevel = png.BestSpeed
		case level >= 9:
			compressionLevel = png.BestCompression
		default:
			// Map 2-8 to BestCompression for simplicity
			compressionLevel = png.BestCompression
		}
		encodeOpts = append(encodeOpts, imaging.PNGCompressionLevel(compressionLevel))
	}

	// GIF number of colors (1-256)
	if numColors, ok := opts["gif_num_colors"].(int); ok && numColors >= 1 && numColors <= 256 {
		encodeOpts = append(encodeOpts, imaging.GIFNumColors(numColors))
	}

	return encodeOpts
}

// encodeAndMeasure encodes an image and returns metadata about the result
// This is a helper function used by all resizers to avoid code duplication.
func encodeAndMeasure(codec ImageCodec, img image.Image, format string, encodeOpts map[string]any, writer io.Writer) (ResizeResult, error) {
	// Get actual dimensions from image
	bounds := img.Bounds()
	actualWidth := bounds.Dx()
	actualHeight := bounds.Dy()

	// Use counting writer to track file size
	countingWriter := &counterWriter{writer: writer}

	// Encode directly to writer with encoding options
	if err := codec.Encode(countingWriter, img, format, encodeOpts); err != nil {
		return ResizeResult{}, fmt.Errorf("failed to encode image: %w", err)
	}

	return ResizeResult{
		Width:    actualWidth,
		Height:   actualHeight,
		FileSize: countingWriter.bytesWritten,
	}, nil
}
