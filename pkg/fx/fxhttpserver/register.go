package fxhttpserver

import (
	"net/http"

	"go.uber.org/fx"
)

// AsMiddleware registers a middleware with the specified priority.
// Lower priority values execute first. Middlewares with the same priority
// execute in registration order.
//
// Recommended priority ranges:
//   - < 100: Before RequestID (rare, usually not needed)
//   - 100-199: Between RequestID and HTTPLog
//   - 200-299: Between HTTPLog and Zorya (recommended default)
//   - >= 300: Right before Zorya adapter
//
// Example:
//
//	fxhttpserver.AsMiddleware(corsMiddleware, 250, "cors")
func AsMiddleware(middleware func(http.Handler) http.Handler, priority int, name string) fx.Option {
	entry := MiddlewareEntry{
		Middleware: middleware,
		Priority:   priority,
		Name:       name,
	}

	return fx.Supply(
		fx.Annotate(
			entry,
			fx.ResultTags(`group:"httpserver-middlewares"`),
		),
	)
}
