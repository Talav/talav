package zorya

import "time"

// DefaultMaxBodyBytes is the default maximum request body size (1MB).
const DefaultMaxBodyBytes int64 = 1024 * 1024

// DefaultBodyReadTimeout is the default timeout for reading request bodies (5 seconds).
const DefaultBodyReadTimeout = 5 * time.Second

// BaseRoute is the base struct for all routes in Fuego.
// It contains the OpenAPI operation and other metadata.
type BaseRoute struct {
	// OpenAPI operation
	Operation *Operation

	// HTTP method (GET, POST, PUT, PATCH, DELETE)
	Method string

	// URL path. Will be prefixed by the base path of the server and the group path if any
	Path string

	// DefaultStatus is the default HTTP status code for this operation. It will
	// be set to 200 or 204 if not specified, depending on whether the handler
	// returns a response body.
	DefaultStatus int

	// Middlewares is a list of middleware functions to run before the handler.
	// This is useful for adding custom logic to operations, such as logging,
	// authentication, or rate limiting.
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

	// Errors is a list of HTTP status codes that the handler may return. If
	// not specified, then a default error response is added to the OpenAPI.
	// This is a convenience for handlers that return a fixed set of errors
	// where you do not wish to provide each one as an OpenAPI response object.
	// Each error specified here is expanded into a response object with the
	// schema generated from the type returned by `huma.NewError()`
	// or `huma.NewErrorWithContext`.
	Errors []int
}
