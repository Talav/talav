package framework

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/fx"
)

// Run provides opinionated default bootstrap with automatic environment detection.
func Run(opts ...Option) error {
	// Auto-detect environment from APP_ENV
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	// Prepend automatic options
	allOpts := []Option{
		WithEnvironment(env),
	}
	allOpts = append(allOpts, opts...)

	// Create and run application
	app := NewApplication(allOpts...)
	return app.Execute()
}

// RunDefault is the simplest bootstrap - just provide modules.
func RunDefault(modules ...fx.Option) error {
	return Run(
		WithName(filepath.Base(os.Args[0])),
		WithVersion("dev"),
		WithModules(modules...),
	)
}

// RunWithContext allows custom context for graceful shutdown.
func RunWithContext(ctx context.Context, opts ...Option) error {
	app := NewApplication(opts...)

	// Set context on root command
	app.rootCmd.SetContext(ctx)

	return app.Execute()
}
