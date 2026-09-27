package framework

import (
	"bytes"
	"errors"
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
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
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
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		WithEnvironment("test"),
		WithModules(testModule),
	)

	require.NotNil(t, app, "application should be created")
}

func TestApplication_RootCommand(t *testing.T) {
	app := NewApplication(
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		WithName("test-app"),
		WithVersion("1.0.0"),
	)

	require.NotNil(t, app.rootCmd, "root command should not be nil")
	assert.Equal(t, "test-app", app.rootCmd.Use)
	assert.Equal(t, "1.0.0", app.rootCmd.Version)
}

func TestApplication_RequiresLogger(t *testing.T) {
	var app Application
	require.ErrorContains(t, app.initFX(t.Context()), "*slog.Logger")
}

func TestWithLogger_Custom(t *testing.T) {
	var injected *slog.Logger
	custom := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := NewApplication(
		WithName("test-app"),
		WithVersion("1.0.0"),
		WithEnvironment("test"),
		WithLogger(custom),
		WithModules(fx.Populate(&injected)),
	)
	assert.Same(t, custom, injected)
	require.NoError(t, app.Shutdown(t.Context()))
}

func TestWithRootCommandHook(t *testing.T) {
	app := NewApplication(
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
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
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
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

func TestApplication_FXLogging_UsesLoggerLevel(t *testing.T) {
	for _, level := range []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		t.Run(level.String(), func(t *testing.T) {
			var output bytes.Buffer
			log := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: level}))
			require.Panics(t, func() {
				NewApplication(
					WithModules(
						fx.Provide(func() *slog.Logger { return log }),
						fx.Invoke(func(logger *slog.Logger) error {
							logger.Info("application info")
							logger.Error("application error")
							return errors.New("test failure")
						}),
					),
				)
			})

			if level <= slog.LevelInfo {
				assert.Contains(t, output.String(), `"level":"INFO","msg":"provided"`)
				assert.Contains(t, output.String(), `"msg":"application info"`)
			} else {
				assert.NotContains(t, output.String(), `"msg":"provided"`)
				assert.NotContains(t, output.String(), `"msg":"application info"`)
			}
			assert.Contains(t, output.String(), `"level":"ERROR","msg":"invoke failed"`)
			assert.Contains(t, output.String(), `"level":"ERROR","msg":"application error"`)
			assert.Contains(t, output.String(), "test failure")
		})
	}
}
