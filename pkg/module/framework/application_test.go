package framework

import (
	"io"
	"log/slog"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestNewApplication(t *testing.T) {
	app := NewApplication(
		WithName("test-app"),
		WithVersion("1.0.0"),
		WithEnvironment("test"),
	)

	assert.Equal(t, "test-app", app.name)
	assert.Equal(t, "1.0.0", app.version)
	assert.Equal(t, "test", app.environment)
	require.NotNil(t, app.rootCmd)
}

func TestApplication_WithModules(t *testing.T) {
	// Use a type that won't conflict with framework's string (environment)
	type testValue int
	testModule := fx.Module("test",
		fx.Provide(func() testValue { return 42 }),
	)

	// NewApplication now panics on FX init failure, so if this succeeds,
	// the module was registered and initialized correctly
	app := NewApplication(
		WithEnvironment("test"),
		WithModules(testModule),
	)

	require.NotNil(t, app, "application should be created")
}

func TestApplication_RootCommand(t *testing.T) {
	app := NewApplication(
		WithName("test-app"),
		WithVersion("1.0.0"),
	)

	require.NotNil(t, app.rootCmd, "root command should not be nil")
	assert.Equal(t, "test-app", app.rootCmd.Use)
	assert.Equal(t, "1.0.0", app.rootCmd.Version)
}

func TestWithLogger_Default(t *testing.T) {
	app := NewApplication(
		WithName("test-app"),
		WithVersion("1.0.0"),
		WithEnvironment("test"),
	)
	assert.Equal(t, slog.Default(), app.logger)
}

func TestWithLogger_Custom(t *testing.T) {
	custom := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := NewApplication(
		WithName("test-app"),
		WithVersion("1.0.0"),
		WithEnvironment("test"),
		WithLogger(custom),
	)
	assert.Same(t, custom, app.logger)
}

func TestWithRootCommandHook(t *testing.T) {
	app := NewApplication(
		WithName("test-app"),
		WithVersion("1.0.0"),
		WithEnvironment("test"),
		WithRootCommandHook(func(cmd *cobra.Command) {
			if cmd.Annotations == nil {
				cmd.Annotations = make(map[string]string)
			}
			cmd.Annotations["framework_test"] = "1"
		}),
	)
	require.NotNil(t, app.rootCmd)
	assert.Equal(t, "1", app.rootCmd.Annotations["framework_test"])
}

func TestWithRootCommandHook_MultipleLastWins(t *testing.T) {
	var order []int
	app := NewApplication(
		WithName("test-app"),
		WithVersion("1.0.0"),
		WithEnvironment("test"),
		WithRootCommandHook(func(cmd *cobra.Command) {
			order = append(order, 1)
			if cmd.Annotations == nil {
				cmd.Annotations = make(map[string]string)
			}
			cmd.Annotations["k"] = "first"
		}),
		WithRootCommandHook(func(cmd *cobra.Command) {
			order = append(order, 2)
			if cmd.Annotations == nil {
				cmd.Annotations = make(map[string]string)
			}
			cmd.Annotations["k"] = "second"
		}),
	)
	require.NotNil(t, app.rootCmd)
	assert.Equal(t, []int{1, 2}, order)
	assert.Equal(t, "second", app.rootCmd.Annotations["k"])
}
