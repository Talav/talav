package fxhttpserver

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/talav/talav/pkg/component/httpserver"
	"github.com/talav/talav/pkg/component/httpserver/cmd"
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/component/zorya/adapters"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxcore"
	"go.uber.org/fx"
)

// ModuleName is the module name.
const ModuleName = "fx-httpserver"

// FxHTTPServerModule is the [Fx] HTTP server module.
// It provides HTTP server components via DI without automatic lifecycle management.
// Commands control the server lifecycle.
var FxHTTPServerModule = fx.Module(
	ModuleName,
	fxconfig.AsConfigWithDefaults("httpserver", httpserver.DefaultConfig(), httpserver.Config{}),
	fx.Provide(
		newFxZoryaAPI,
		func(cfg httpserver.Config, api zorya.API, logger *slog.Logger) (*httpserver.Server, error) {
			return httpserver.NewServer(cfg.Server, api, logger)
		},
	),
	// Register serve-http command
	fxcore.AsRootCommand(cmd.NewServeHTTPCmd),
)

// MiddlewareParams allows injection of registered middlewares.
type MiddlewareParams struct {
	fx.In
	Middlewares []middlewareEntry `group:"httpserver-middlewares"`
}

// newFxZoryaAPI creates a new Zorya API instance with router and infrastructure middleware configured.
// The router is created here, middleware is added, then Zorya adapter wraps it.
func newFxZoryaAPI(cfg httpserver.Config, logger *slog.Logger, params MiddlewareParams) (zorya.API, error) {
	// Create router
	router := chi.NewRouter()

	// Build all middlewares (built-in + user)
	allMiddlewares := buildAllMiddlewares(params.Middlewares, cfg, logger)

	// Sort all middlewares by priority
	sortedMiddlewares := sortMiddlewares(allMiddlewares)

	// Apply all middlewares in priority order
	for _, mw := range sortedMiddlewares {
		if mw.name != "" {
			logger.Debug("registering middleware", "name", mw.name, "priority", mw.priority)
		}
		router.Use(mw.middleware)
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
func buildAllMiddlewares(userMiddlewares []middlewareEntry, cfg httpserver.Config, logger *slog.Logger) []middlewareEntry {
	allMiddlewares := make([]middlewareEntry, 0, len(userMiddlewares)+2)

	// Add built-in RequestID middleware (always)
	// Use order 0 to ensure it's first among same-priority middlewares
	allMiddlewares = append(allMiddlewares, middlewareEntry{
		middleware: middleware.RequestID,
		priority:   PriorityRequestID,
		name:       "request-id",
		order:      0,
	})

	// Add built-in CORS middleware (if enabled)
	// Use order 0 to ensure it's first among same-priority middlewares
	if cfg.CORS.Enabled {
		allMiddlewares = append(allMiddlewares, middlewareEntry{
			middleware: httpserver.NewCORSMiddleware(cfg.CORS),
			priority:   PriorityCORS,
			name:       "cors",
			order:      0,
		})
	}

	// Add built-in HTTPLog middleware (if enabled)
	// Use order 0 to ensure it's first among same-priority middlewares
	if cfg.Logging.Enabled {
		opts := httpserver.NewHTTPLogOptions(cfg.Logging)
		allMiddlewares = append(allMiddlewares, middlewareEntry{
			middleware: httplog.RequestLogger(logger, opts),
			priority:   PriorityHTTPLog,
			name:       "http-log",
			order:      0,
		})
	}

	// Add user middlewares
	// Order is preserved via the order field assigned at registration time
	allMiddlewares = append(allMiddlewares, userMiddlewares...)

	return allMiddlewares
}
