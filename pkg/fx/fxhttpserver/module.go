package fxhttpserver

import (
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/talav/talav/pkg/component/httpserver"
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/component/zorya/adapters"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

// ModuleName is the module name.
const ModuleName = "httpserver"

// FxHTTPServerModule is the [Fx] HTTP server module.
var FxHTTPServerModule = fx.Module(
	ModuleName,
	fxconfig.AsConfigWithDefaults("httpserver", httpserver.DefaultConfig(), httpserver.Config{}),
	fx.Provide(
		NewFxZoryaAPI,
		NewFxServer,
	),
	fx.Invoke(RegisterLifecycle),
)

// MiddlewareParams allows injection of registered middlewares.
type MiddlewareParams struct {
	fx.In
	Middlewares []MiddlewareEntry `group:"httpserver-middlewares"`
}

// NewFxZoryaAPI creates a new Zorya API instance with router and infrastructure middleware configured.
// The router is created here, middleware is added, then Zorya adapter wraps it.
func NewFxZoryaAPI(cfg httpserver.Config, logger *slog.Logger, params MiddlewareParams) (zorya.API, error) {
	// Create router
	router := chi.NewRouter()

	// Build all middlewares (built-in + user)
	allMiddlewares := buildAllMiddlewares(params.Middlewares, cfg, logger)

	// Sort all middlewares by priority
	sortedMiddlewares := sortMiddlewares(allMiddlewares)

	// Apply all middlewares in priority order
	for _, mw := range sortedMiddlewares {
		if mw.Name != "" {
			logger.Debug("registering middleware", "name", mw.Name, "priority", mw.Priority)
		}
		router.Use(mw.Middleware)
	}

	// Create Zorya adapter with the configured router
	adapter := adapters.NewChi(router)

	// Create Zorya API with the adapter
	api := zorya.NewAPI(
		adapter,
		zorya.WithOpenAPI(cfg.ToZoryaOpenAPI()),
		zorya.WithConfig(cfg.ToZoryaConfig()),
	)

	return api, nil
}

// buildAllMiddlewares combines built-in middlewares with user-registered middlewares.
func buildAllMiddlewares(userMiddlewares []MiddlewareEntry, cfg httpserver.Config, logger *slog.Logger) []MiddlewareEntry {
	allMiddlewares := make([]MiddlewareEntry, 0, len(userMiddlewares)+2)

	// Add built-in RequestID middleware (always)
	// Use order 0 to ensure it's first among same-priority middlewares
	allMiddlewares = append(allMiddlewares, MiddlewareEntry{
		Middleware: middleware.RequestID,
		Priority:   PriorityRequestID,
		Name:       "request-id",
		order:      0,
	})

	// Add built-in HTTPLog middleware (if enabled)
	// Use order 0 to ensure it's first among same-priority middlewares
	if cfg.Logging.Enabled {
		opts := httpserver.BuildHTTPLogOptions(cfg.Logging)
		allMiddlewares = append(allMiddlewares, MiddlewareEntry{
			Middleware: httplog.RequestLogger(logger, opts),
			Priority:   PriorityHTTPLog,
			Name:       "http-log",
			order:      0,
		})
	}

	// Add user middlewares
	// Order is preserved via the order field assigned at registration time
	allMiddlewares = append(allMiddlewares, userMiddlewares...)

	return allMiddlewares
}

// NewFxServer creates a new HTTP server instance that wraps the Zorya API for lifecycle management.
func NewFxServer(cfg httpserver.Config, api zorya.API, logger *slog.Logger) (*httpserver.Server, error) {
	return httpserver.NewServer(cfg.Server, api, logger)
}

// RegisterLifecycle registers the HTTP server lifecycle hooks.
func RegisterLifecycle(lc fx.Lifecycle, server *httpserver.Server, logger *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start server in a goroutine
			go func() {
				if err := server.Start(ctx); err != nil {
					logger.Error("HTTP server error", "error", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Server will stop when context is cancelled
			return nil
		},
	})
}
