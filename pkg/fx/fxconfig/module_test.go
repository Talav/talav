package fxconfig

import (
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/talav/pkg/component/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestModule_NewFxConfig_WithConfigSources(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_newfxconfig_withconfigsources")
	t.Setenv("APP_ENV", "dev")

	var cfg *config.Config

	fxtest.New(
		t,
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"custom.yaml", "custom_{env}.yaml"},
			Parser:   yaml.Parser(),
		}),
		fx.Populate(&cfg),
	).RequireStart().RequireStop()

	require.NotNil(t, cfg)

	var appCfg struct {
		Name string `config:"name"`
	}
	err := cfg.UnmarshalKey("app", &appCfg)
	require.NoError(t, err)
	assert.Equal(t, "custom-app", appCfg.Name)
}

func TestModule_NewFxConfig_WithEnvironment(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_newfxconfig_withenvironment")
	t.Setenv("APP_ENV", "test")

	var cfg *config.Config

	fxtest.New(
		t,
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml", "config_{env}.yaml"},
			Parser:   yaml.Parser(),
		}),
		fx.Populate(&cfg),
	).RequireStart().RequireStop()

	require.NotNil(t, cfg)

	var appCfg struct {
		Name string `config:"name"`
		Env  string `config:"env"`
	}
	err := cfg.UnmarshalKey("app", &appCfg)
	require.NoError(t, err)
	assert.Equal(t, "test-app", appCfg.Name)
	assert.Equal(t, "test", appCfg.Env)
}

func TestModule_AsConfig_Provider(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_asconfig_provider")
	t.Setenv("APP_ENV", "dev")

	type AppConfig struct {
		Name    string `config:"name"`
		Version string `config:"version"`
	}

	type DatabaseConfig struct {
		Host string `config:"host"`
		Port int    `config:"port"`
	}

	var appCfg AppConfig
	var dbCfg DatabaseConfig

	fxtest.New(
		t,
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		AsConfig("app", AppConfig{}),
		AsConfig("database", DatabaseConfig{}),
		fx.Populate(&appCfg, &dbCfg),
	).RequireStart().RequireStop()

	assert.Equal(t, "test-app", appCfg.Name)
	assert.Equal(t, "1.0.0", appCfg.Version)
	assert.Equal(t, "localhost", dbCfg.Host)
	assert.Equal(t, 5432, dbCfg.Port)
}

func TestModule_NewFxConfig_ErrorHandling(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_newfxconfig_errorhandling")
	t.Setenv("APP_ENV", "dev")

	var cfg *config.Config

	// This test verifies that invalid YAML causes an error during app construction
	// Use fx.New() directly to catch the error instead of fxtest.New() which fails on error
	app := fx.New(
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"invalid.yaml"},
			Parser:   yaml.Parser(),
		}),
		fx.Populate(&cfg),
	)

	// Error should occur during construction due to invalid YAML
	err := app.Err()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "yaml")
}
