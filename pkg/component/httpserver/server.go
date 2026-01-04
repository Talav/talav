package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/talav/talav/pkg/component/zorya"
)

// Server represents an HTTP server that wraps Zorya API for lifecycle management.
type Server struct {
	api     zorya.API
	config  ServerConfig
	logger  *slog.Logger
	httpSrv *http.Server
}

// NewServer creates a new HTTP server that wraps the provided Zorya API.
// The router and middleware should be configured in the Zorya API before passing it here.
func NewServer(cfg ServerConfig, api zorya.API, logger *slog.Logger) (*Server, error) {
	return &Server{
		api:    api,
		config: cfg,
		logger: logger,
	}, nil
}

// Start starts the HTTP server and blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Parse timeout durations
	readTimeoutDuration, _ := time.ParseDuration(s.config.ReadTimeout)
	writeTimeoutDuration, _ := time.ParseDuration(s.config.WriteTimeout)
	idleTimeoutDuration, _ := time.ParseDuration(s.config.IdleTimeout)
	readHeaderTimeoutDuration, _ := time.ParseDuration(s.config.ReadHeaderTimeout)
	shutdownTimeoutDuration, _ := time.ParseDuration(s.config.ShutdownTimeout)

	s.httpSrv = &http.Server{
		Addr:              addr,
		Handler:           s.api.Adapter(),
		ReadTimeout:       readTimeoutDuration,
		WriteTimeout:      writeTimeoutDuration,
		IdleTimeout:       idleTimeoutDuration,
		ReadHeaderTimeout: readHeaderTimeoutDuration,
		MaxHeaderBytes:    s.config.MaxHeaderBytes,
	}

	s.logger.Info("starting HTTP server", "addr", addr)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeoutDuration)
		defer cancel()

		if err := s.httpSrv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}

		return nil
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}
