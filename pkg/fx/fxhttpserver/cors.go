package fxhttpserver

import (
	"net/http"

	"github.com/rs/cors"
	"github.com/talav/talav/pkg/component/httpserver"
)

// buildCORSMiddleware creates a CORS middleware from the provided configuration.
func buildCORSMiddleware(cfg httpserver.CORSConfig) func(http.Handler) http.Handler {
	corsOptions := cors.Options{
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		ExposedHeaders:   cfg.ExposedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	}

	// Set default methods if not specified
	if len(corsOptions.AllowedMethods) == 0 {
		corsOptions.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"}
	}

	// Set default headers if not specified
	if len(corsOptions.AllowedHeaders) == 0 {
		corsOptions.AllowedHeaders = []string{"*"}
	}

	// Handle wildcard origins
	if len(cfg.AllowedOrigins) > 0 {
		allowsAll := false
		for _, origin := range cfg.AllowedOrigins {
			if origin == "*" {
				allowsAll = true
				break
			}
		}

		if allowsAll {
			corsOptions.AllowOriginFunc = func(origin string) bool {
				return true
			}
		} else {
			corsOptions.AllowedOrigins = cfg.AllowedOrigins
		}
	}

	return cors.New(corsOptions).Handler
}

