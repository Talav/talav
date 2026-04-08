package fxconfig

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/talav/pkg/component/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// errFxValidatableTest is returned by test [Validatable] implementations when validation must fail.
var errFxValidatableTest = errors.New("fx validatable test")

type widgetConfig struct {
	Reject bool `config:"reject"`
}

func (c *widgetConfig) Validate() error {
	if c.Reject {
		return errFxValidatableTest
	}

	return nil
}

type serverConfigWithValidate struct {
	Host string `config:"host"`
	Port int    `config:"port"`
}

func (c *serverConfigWithValidate) Validate() error {
	if c.Port == 9000 {
		return errFxValidatableTest
	}

	return nil
}

// mergeKeysFlatConfig is used for AsConfigMergeKeys success tests (no Validatable).
type mergeKeysFlatConfig struct {
	Name  string `config:"name"`
	Extra string `config:"extra"`
}

// mergeKeysPortConfig is used when testing merge unmarshal errors.
type mergeKeysPortConfig struct {
	Port int `config:"port"`
}

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

func TestModule_AsConfigWithDefaults_Provider(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_asconfigwithdefaults_provider")
	t.Setenv("APP_ENV", "dev")

	type ServerConfig struct {
		Host string `config:"host"`
		Port int    `config:"port"`
	}

	// Default config with all fields set
	defaultConfig := ServerConfig{
		Host: "localhost",
		Port: 8080,
	}

	var serverCfg ServerConfig

	fxtest.New(
		t,
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		AsConfigWithDefaults("server", defaultConfig, ServerConfig{}),
		fx.Populate(&serverCfg),
	).RequireStart().RequireStop()

	// User config only sets port, host should use default
	assert.Equal(t, "localhost", serverCfg.Host) // default value
	assert.Equal(t, 9000, serverCfg.Port)        // user config overrides default
}

func TestModule_AsConfig_Validatable_OK(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_asconfig_validatable_ok")
	t.Setenv("APP_ENV", "dev")

	var widgetCfg widgetConfig

	fxtest.New(
		t,
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		AsConfig("widget", widgetConfig{}),
		fx.Populate(&widgetCfg),
	).RequireStart().RequireStop()

	assert.False(t, widgetCfg.Reject)
}

func TestModule_AsConfig_Validatable_Error(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_asconfig_validatable_err")
	t.Setenv("APP_ENV", "dev")

	var widgetCfg widgetConfig

	app := fx.New(
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		AsConfig("widget", widgetConfig{}),
		fx.Populate(&widgetCfg),
	)

	err := app.Err()
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), `config key "widget"`))
	require.ErrorIs(t, err, errFxValidatableTest)
}

func TestModule_AsConfigWithDefaults_Validatable_Error(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_asconfigwithdefaults_validatable_err")
	t.Setenv("APP_ENV", "dev")

	defaults := serverConfigWithValidate{
		Host: "localhost",
		Port: 8080,
	}

	var serverCfg serverConfigWithValidate

	app := fx.New(
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		AsConfigWithDefaults("server", defaults, serverConfigWithValidate{}),
		fx.Populate(&serverCfg),
	)

	err := app.Err()
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), `config key "server"`))
	require.ErrorIs(t, err, errFxValidatableTest)
}

func TestModule_AsConfigMergeKeys_OK(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_asconfig_mergekeys_ok")
	t.Setenv("APP_ENV", "dev")

	var got mergeKeysFlatConfig

	fxtest.New(
		t,
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		AsConfigMergeKeys("merged_cfg", []string{"mergekeys_base", "mergekeys_overlay"}, mergeKeysFlatConfig{}),
		fx.Populate(&got),
	).RequireStart().RequireStop()

	assert.Equal(t, "merged", got.Name)
	assert.Equal(t, "from-overlay", got.Extra)
}

func TestModule_AsConfigMergeKeys_Validate_Error(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_asconfig_mergekeys_validate_err")
	t.Setenv("APP_ENV", "dev")

	var got widgetConfig

	app := fx.New(
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		AsConfigMergeKeys("mergekeysValidateLabel", []string{"mergekeys_v_base", "mergekeys_v_overlay"}, widgetConfig{}),
		fx.Populate(&got),
	)

	err := app.Err()
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), `config key "mergekeysValidateLabel"`))
	require.ErrorIs(t, err, errFxValidatableTest)
}

func TestModule_AsConfigMergeKeys_Unmarshal_ErrorWrapsFailingPath(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_asconfig_mergekeys_unmarshal_err")
	t.Setenv("APP_ENV", "dev")

	var got mergeKeysPortConfig

	app := fx.New(
		fx.NopLogger,
		FxConfigModule,
		AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		AsConfigMergeKeys("service", []string{"mergekeys_u_base", "mergekeys_u_bad"}, mergeKeysPortConfig{}),
		fx.Populate(&got),
	)

	err := app.Err()
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), `config key "mergekeys_u_bad"`))
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
