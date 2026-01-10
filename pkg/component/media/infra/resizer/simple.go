package resizer

import (
	"context"
	"fmt"
	"io"

	"github.com/disintegration/imaging"
)

// SimpleResizer resizes images using Lanczos filter.
type SimpleResizer struct {
	codec ImageCodec
}

// NewSimpleResizer creates a new SimpleResizer.
func NewSimpleResizer(codec ImageCodec) *SimpleResizer {
	return &SimpleResizer{codec: codec}
}

// Resize resizes the media from reader to writer with specified dimensions.
func (r *SimpleResizer) Resize(ctx context.Context, reader io.Reader, opts ResizeOptions, writer io.Writer) (ResizeResult, error) {
	// Decode image
	img, err := r.codec.Decode(reader)
	if err != nil {
		return ResizeResult{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize using imaging library
	resizedImg := imaging.Resize(img, opts.Width, opts.Height, imaging.Lanczos)

	// Encode and measure result
	return encodeAndMeasure(r.codec, resizedImg, opts.Format, opts.Options, writer)
}
