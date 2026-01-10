package resizer

import (
	"context"
	"fmt"
	"io"

	"github.com/disintegration/imaging"
)

// SquareResizer resizes images to square dimensions by cropping.
type SquareResizer struct {
	codec ImageCodec
}

// NewSquareResizer creates a new SquareResizer.
func NewSquareResizer(codec ImageCodec) *SquareResizer {
	return &SquareResizer{codec: codec}
}

// Resize resizes the media to square dimensions (uses the larger of width/height).
func (r *SquareResizer) Resize(ctx context.Context, reader io.Reader, opts ResizeOptions, writer io.Writer) (ResizeResult, error) {
	// Decode image
	img, err := r.codec.Decode(reader)
	if err != nil {
		return ResizeResult{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// Use the larger dimension for square
	size := max(opts.Height, opts.Width)
	// Crop to square, then resize
	resizedImg := imaging.Fill(img, size, size, imaging.Center, imaging.Lanczos)

	// Encode and measure result
	return encodeAndMeasure(r.codec, resizedImg, opts.Format, opts.Options, writer)
}
