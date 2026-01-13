package framework

import (
	"testing"

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
