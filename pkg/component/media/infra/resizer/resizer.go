package resizer

import (
	"context"
	"io"
)

// ResizeResult contains metadata about the resized image.
type ResizeResult struct {
	Width    int   // actual output width
	Height   int   // actual output height
	FileSize int64 // size of encoded output in bytes
}

// ResizeOptions contains options for resizing operations.
type ResizeOptions struct {
	Width   int            // target width
	Height  int            // target height
	Format  string         // target format extension without leading dot (e.g., "jpg", "png")
	Options map[string]any // resizer-specific configuration
}

// Resizer defines the interface for media resizing operations
// Resizers work with media directly, abstracting away image processing internals.
type Resizer interface {
	// Resize resizes media from reader to writer with specified dimensions
	// The resizer handles decoding, resizing, and encoding internally
	// Returns metadata about the resized image including actual dimensions and file size
	Resize(ctx context.Context, reader io.Reader, opts ResizeOptions, writer io.Writer) (ResizeResult, error)
}
