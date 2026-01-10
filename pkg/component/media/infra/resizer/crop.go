package resizer

import (
	"context"
	"fmt"
	"io"

	"github.com/disintegration/imaging"
)

// CropResizer resizes images by cropping to maintain aspect ratio.
type CropResizer struct {
	codec ImageCodec
}

// NewCropResizer creates a new CropResizer.
func NewCropResizer(codec ImageCodec) *CropResizer {
	return &CropResizer{codec: codec}
}

// Resize resizes the media by cropping to the specified dimensions.
func (r *CropResizer) Resize(ctx context.Context, reader io.Reader, opts ResizeOptions, writer io.Writer) (ResizeResult, error) {
	// Decode image
	img, err := r.codec.Decode(reader)
	if err != nil {
		return ResizeResult{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// Crop to center maintaining aspect ratio, then resize
	resizedImg := imaging.Fill(img, opts.Width, opts.Height, imaging.Center, imaging.Lanczos)

	// Encode and measure result
	return encodeAndMeasure(r.codec, resizedImg, opts.Format, opts.Options, writer)
}
