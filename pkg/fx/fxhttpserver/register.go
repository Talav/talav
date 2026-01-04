package fxhttpserver

import (
	"net/http"
	"sync/atomic"

	"go.uber.org/fx"
)

// middlewareOrderCounter is an atomic counter for tracking middleware registration order.
var middlewareOrderCounter int64

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
	entry := middlewareEntry{
		middleware: middleware,
		priority:   priority,
		name:       name,
		order:      int(atomic.AddInt64(&middlewareOrderCounter, 1)),
	}

	return fx.Supply(
		fx.Annotate(
			entry,
			fx.ResultTags(`group:"httpserver-middlewares"`),
		),
	)
}

// AsMiddlewareConstructor registers a middleware with the specified priority using a constructor.
// The constructor will be called by Fx with dependency injection.
// The constructor must return func(http.Handler) http.Handler.
// Additional annotations (like fx.ParamTags) can be passed as variadic arguments.
//
// Example:
//
//	fxhttpserver.AsMiddlewareConstructor(NewAuthMiddleware, 250, "auth")
//
// Where NewAuthMiddleware is:
//
//	func NewAuthMiddleware(jwtService security.JWTService) func(http.Handler) http.Handler {
//		return func(next http.Handler) http.Handler { ... }
//	}
func AsMiddlewareConstructor(constructor any, priority int, name string, annotations ...fx.Annotation) fx.Option {
	wrapper := func(middleware func(http.Handler) http.Handler) middlewareEntry {
		return middlewareEntry{
			middleware: middleware,
			priority:   priority,
			name:       name,
			order:      int(atomic.AddInt64(&middlewareOrderCounter, 1)),
		}
	}

	annotations = append(annotations, fx.ResultTags(`group:"httpserver-middlewares"`))

	return fx.Options(
		fx.Provide(constructor),
		fx.Provide(
			fx.Annotate(
				wrapper,
				annotations...,
			),
		),
	)
}
