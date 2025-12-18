package zorya

import "time"

// DefaultMaxBodyBytes is the default maximum request body size (1MB).
const DefaultMaxBodyBytes int64 = 1024 * 1024

// DefaultBodyReadTimeout is the default timeout for reading request bodies (5 seconds).
const DefaultBodyReadTimeout = 5 * time.Second

// BaseRoute is the base struct for all routes in Fuego.
// It contains the OpenAPI operation and other metadata.
type BaseRoute struct {
	// HTTP method (GET, POST, PUT, PATCH, DELETE)
	Method string

	// URL path. Will be prefixed by the base path of the server and the group path if any
	Path string

	Middlewares Middlewares

	// BodyReadTimeout sets a deadline for reading the request body.
	// If > 0, sets read deadline to now + timeout.
	// If == 0, uses DefaultBodyReadTimeout (5 seconds).
	// If < 0, disables any deadline (no timeout).
	BodyReadTimeout time.Duration

	// MaxBodyBytes limits the size of the request body in bytes.
	// If > 0, enforces the specified limit.
	// If == 0, uses DefaultMaxBodyBytes (1MB).
	// If < 0, disables the limit (no size restriction).
	MaxBodyBytes int64
}
