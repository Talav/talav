package zorya

import (
	"context"
	"net/http"
	"time"
)

// Context is the current request/response context. It provides a generic
// interface to get request information and write responses.
type Context interface {
	// Context returns the underlying request context.
	Context() context.Context
	// Request returns the HTTP request.
	Request() *http.Request
	// RouterParams returns the path parameters from the router.
	RouterParams() map[string]string
	// Header returns the value of the specified request header.
	Header(name string) string
	// SetReadDeadline sets the deadline for reading the request body.
	// If t is zero, disables the deadline.
	SetReadDeadline(t time.Time) error
	// BodyWriter returns the underlying response writer for streaming responses.
	// Use this in Body func(Context) handlers for direct control over the response.
	// For SSE, set appropriate headers via ResponseWriter type assertion before writing.
	BodyWriter() http.ResponseWriter
}

// BodyLimiter is an optional interface that contexts can implement to support
// request body size limiting. Adapters should implement this to enable
// MaxBodyBytes functionality.
type BodyLimiter interface {
	// SetBodyLimit wraps the request body with a size limiter.
	// If limit <= 0, no limit is applied.
	SetBodyLimit(limit int64)
}
