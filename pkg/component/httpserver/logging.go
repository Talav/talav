package httpserver

import (
	"log/slog"
	"net/http"
	"slices"

	"github.com/go-chi/httplog/v3"
)

// NewHTTPLogOptions converts LoggingConfig to httplog.Options.
func NewHTTPLogOptions(cfg LoggingConfig) *httplog.Options {
	opts := &httplog.Options{
		Level:         ParseLogLevel(cfg.Level),
		Schema:        ParseSchema(cfg.Schema),
		RecoverPanics: cfg.RecoverPanics,
	}

	// Configure body logging if enabled
	if cfg.LogRequestBody {
		opts.LogRequestBody = func(*http.Request) bool {
			return true
		}
	}

	if cfg.LogResponseBody {
		opts.LogResponseBody = func(*http.Request) bool {
			return true
		}
	}

	// Configure skip paths if provided
	if len(cfg.SkipPaths) > 0 {
		opts.Skip = func(req *http.Request, respStatus int) bool {
			return slices.Contains(cfg.SkipPaths, req.URL.Path)
		}
	}

	return opts
}

// ParseLogLevel converts string level to slog.Level.
func ParseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ParseSchema converts string schema to *httplog.Schema.
func ParseSchema(schema string) *httplog.Schema {
	switch schema {
	case "ecs":
		return httplog.SchemaECS
	case "otel":
		return httplog.SchemaOTEL
	case "gcp":
		return httplog.SchemaGCP
	default:
		// Default to standard schema (nil means no schema transformation)
		return nil
	}
}
