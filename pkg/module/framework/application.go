package framework

import (
	"context"
	"fmt"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/talav/talav/pkg/fx/fxcore"
	"go.uber.org/fx"
)

// Application represents the main framework application.
type Application struct {
	name        string
	version     string
	environment string

	// FX container
	modules  []fx.Option
	fxApp    *fx.App
	commands []*cobra.Command

	// Cobra CLI
	rootCmd *cobra.Command
}

// NewApplication creates a new Application with the given options.
func NewApplication(opts ...Option) *Application {
	a := &Application{
		modules: make([]fx.Option, 0),
	}

	for _, opt := range opts {
		opt(a)
	}

	// Create root command
	a.rootCmd = a.createRootCommand()

	// Initialize FX container immediately to register all module commands
	ctx := context.Background()
	if err := a.initFX(ctx); err != nil {
		panic(fmt.Errorf("FX initialization failed: %w", err))
	}

	// Register commands from modules
	_ = a.injectModuleCommands(a.rootCmd)

	return a
}

// Execute starts the CLI application.
// It sets up signal handling and ensures FX cleanup happens after the command completes.
func (a *Application) Execute() error {
	// Setup signal handling at framework level - all commands inherit this
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Set signal-aware context on root command
	a.rootCmd.SetContext(ctx)

	// Ensure FX cleanup when Execute returns
	defer func() {
		if a.fxApp != nil {
			// Use background context for cleanup (independent of command context)
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()
			_ = a.Shutdown(shutdownCtx)
		}
	}()

	return a.rootCmd.Execute()
}

// createRootCommand creates the root Cobra command.
func (a *Application) createRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     a.name,
		Short:   fmt.Sprintf("%s - powered by Talav Framework", a.name),
		Version: a.version,
	}

	// Add built-in version command
	cmd.AddCommand(a.createVersionCommand())

	return cmd
}

// createVersionCommand creates the version command.
func (a *Application) createVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", a.name, a.version)
			fmt.Printf("Environment: %s\n", a.environment)
			fmt.Printf("Go version: %s\n", runtime.Version())
		},
	}
}

// initFX initializes the FX app.
func (a *Application) initFX(ctx context.Context) error {
	var commandsParam fxcore.FxCommandsParam

	// Build FX options
	opts := []fx.Option{
		fx.NopLogger, // Suppress FX logs
		fx.Supply(a.environment),
	}
	opts = append(opts, a.modules...)

	// Collect commands BEFORE starting (FX way)
	opts = append(opts, fx.Populate(&commandsParam))

	// Create FX app
	a.fxApp = fx.New(opts...)

	// Start FX app (DI only, no lifecycle hooks)
	if err := a.fxApp.Start(ctx); err != nil {
		return err
	}

	// Store collected commands
	a.commands = commandsParam.Commands
	return nil
}

// injectModuleCommands collects commands from FX modules and adds them to root.
func (a *Application) injectModuleCommands(rootCmd *cobra.Command) error {
	// Get commands collected during FX initialization
	for _, cmd := range a.commands {
		// Check if command already exists
		existingCmd, _, _ := rootCmd.Find([]string{cmd.Name()})
		if existingCmd == nil || existingCmd == rootCmd {
			rootCmd.AddCommand(cmd)
		}
	}

	return nil
}

// Shutdown gracefully shuts down the application.
func (a *Application) Shutdown(ctx context.Context) error {
	if a.fxApp != nil {
		if err := a.fxApp.Stop(ctx); err != nil {
			return fmt.Errorf("FX shutdown failed: %w", err)
		}
	}

	return nil
}
