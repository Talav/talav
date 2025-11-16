package logger

import (
	"io"
	"log/slog"
	"os"
)

// LoggerFactory is the interface for [slog.Logger] factories.
type LoggerFactory interface {
	Create(cfg LoggerConfig) (*slog.Logger, error)
}

// DefaultLoggerFactory is the default [LoggerFactory] implementation.
type DefaultLoggerFactory struct{}

// NewDefaultLoggerFactory returns a [DefaultLoggerFactory], implementing [LoggerFactory].
func NewDefaultLoggerFactory() LoggerFactory {
	return &DefaultLoggerFactory{}
}

// Create returns a new [slog.Logger] from the given [LoggerConfig].
//
// Example:
//
//	var factory = NewDefaultLoggerFactory()
//	var logger, _ = factory.Create(LoggerConfig{
//		Level:  "info",
//		Format: "json",
//		Output: "stdout",
//	})
func (f *DefaultLoggerFactory) Create(cfg LoggerConfig) (*slog.Logger, error) {
	// Parse log level
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Configure output
	var output io.Writer
	if cfg.Output == "stdout" || cfg.Output == "" {
		output = os.Stdout
	} else {
		//bearer:disable go_gosec_file_permissions_file_perm
		file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			// Fall back to stdout if file creation fails
			output = os.Stdout
		} else {
			// Use multi-writer to write to both file and stdout
			output = io.MultiWriter(file, os.Stdout)
		}
	}

	// Create handler based on format
	var handler slog.Handler
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(output, opts)
	} else {
		// Default to JSON format
		handler = slog.NewJSONHandler(output, opts)
	}

	return slog.New(handler), nil
}
