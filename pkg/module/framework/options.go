package framework

import (
	"log/slog"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Option is a function that configures an Application.
type Option func(*Application)

// WithName sets the application name.
func WithName(name string) Option {
	return func(a *Application) {
		a.name = name
	}
}

// WithVersion sets the application version.
func WithVersion(version string) Option {
	return func(a *Application) {
		a.version = version
	}
}

// WithEnvironment sets the application environment.
func WithEnvironment(env string) Option {
	return func(a *Application) {
		a.environment = env
	}
}

// WithModules adds FX modules to the application.
func WithModules(modules ...fx.Option) Option {
	return func(a *Application) {
		a.modules = append(a.modules, modules...)
	}
}

// WithLogger sets the slog.Logger used for FX's internal event reporting
// (dependency resolution, lifecycle hooks, errors). The logger is captured at
// NewApplication time; configure it before calling Run or NewApplication.
//
// Non-error FX events (provider registration, lifecycle hooks, "started") are
// logged at warn level, suppressing routine DI noise. FX error events (invoke
// failed, missing type, rollback) are logged at error level regardless of the
// warn floor. To see all FX output, pass a logger whose handler enables debug
// level.
//
// Default: slog.Default() at the time NewApplication is called.
func WithLogger(logger *slog.Logger) Option {
	return func(a *Application) {
		a.logger = logger
	}
}

// WithRootCommandHook registers a function that is called on the root
// *cobra.Command after the root and built-in subcommands (e.g. version) are
// set up, but before FX is initialized and module-registered subcommands are
// attached. Use it to set PersistentPreRun, PersistentPostRun, global flags,
// or other root-level Cobra options.
//
// WithRootCommandHook may be specified multiple times; hooks run in
// registration order. If two hooks set the same field on the root command, the
// last registration wins (there is no automatic chaining).
//
// Cobra's default is that only the first PersistentPreRun or PersistentPostRun
// found on the path from the executing subcommand up toward the root is run
// (unless the application sets cobra.EnableTraverseRunHooks to run all
// parents' persistent hooks in order). A root hook that sets only the root’s
// Persistent* runs for a subcommand when no command closer to the leaf defines
// Persistent* (see Cobra’s execute flow). PersistentPostRun is not run if
// earlier execution or validation returns an error before the normal
// post-chain.
func WithRootCommandHook(fn func(*cobra.Command)) Option {
	return func(a *Application) {
		a.rootCommandHooks = append(a.rootCommandHooks, fn)
	}
}
