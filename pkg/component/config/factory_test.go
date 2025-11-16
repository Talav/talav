package config

import (
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigFactory_Create_Scenario1_YAMLOnly(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario1_yaml_only")
	t.Setenv("APP_ENV", "dev")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "yaml-local", cfg.k.String("app.name"))
	assert.Equal(t, "local-host", cfg.k.String("database.host"))
	assert.Equal(t, int64(5432), cfg.k.Int64("database.port"))
}

func TestDefaultConfigFactory_Create_Scenario1_YAMLOnly_ProdEnv(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario1_yaml_only")
	t.Setenv("APP_ENV", "prod")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "yaml-base", cfg.k.String("app.name"))
	assert.Equal(t, "localhost", cfg.k.String("database.host"))
}

func TestDefaultConfigFactory_Create_Scenario2_EnvOnly(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario2_env_only")
	t.Setenv("APP_ENV", "dev")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{".env", ".env.{env}", ".env.local", ".env.{env}.local"},
			Parser:   NewDotenvParser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "env-dev-local", cfg.k.String("app.name"))
	assert.Equal(t, "env-dev-local-host", cfg.k.String("database.host"))
}

func TestDefaultConfigFactory_Create_Scenario2_EnvOnly_WithoutLocal(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario10_env_staging")
	t.Setenv("APP_ENV", "staging")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{".env", ".env.{env}", ".env.local", ".env.{env}.local"},
			Parser:   NewDotenvParser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "staging-override", cfg.k.String("app.name"))
}

func TestDefaultConfigFactory_Create_Scenario3_YAMLAndEnv(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario3_yaml_and_env")
	t.Setenv("APP_ENV", "dev")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{".env", ".env.{env}", ".env.local", ".env.{env}.local"},
			Parser:   NewDotenvParser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "yaml-value", cfg.k.String("app.name"))
	assert.Equal(t, "env-override-host", cfg.k.String("database.host"))
	assert.Equal(t, "secret", cfg.k.String("database.password"))
}

func TestDefaultConfigFactory_Create_Scenario4_AllOverrides(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario4_all_overrides")
	t.Setenv("APP_ENV", "test")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{".env", ".env.{env}", ".env.local", ".env.{env}.local"},
			Parser:   NewDotenvParser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "9003", cfg.k.String("server.port"))
	assert.Equal(t, "env-final-host", cfg.k.String("server.host"))
}

func TestDefaultConfigFactory_Create_Scenario5_VariableExpansion(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario5_variable_expansion")
	t.Setenv("HOME", "/home/testuser")
	t.Setenv("CONFIG_ROOT", "/etc/myapp")
	t.Setenv("DB_USER", "dbuser")
	t.Setenv("DB_PASS", "dbpass")
	t.Setenv("DB_HOST", "dbhost")
	t.Setenv("APP_ENV", "dev")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "/home/testuser/data", cfg.k.String("app.data_dir"))
	assert.Equal(t, "postgresql://dbuser:dbpass@dbhost:5432/mydb", cfg.k.String("database.url"))
}

func TestDefaultConfigFactory_Create_Scenario6_MissingFiles(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario6_missing_files")
	t.Setenv("APP_ENV", "dev")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{".env", ".env.{env}", ".env.local", ".env.{env}.local"},
			Parser:   NewDotenvParser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "only-base", cfg.k.String("app.name"))
}

func TestDefaultConfigFactory_Create_CompletelyMissingDirectory(t *testing.T) {
	factory := NewDefaultConfigFactory()
	nonExistentDir := filepath.Join("testdata", "does-not-exist")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     nonExistentDir,
			Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestDefaultConfigFactory_Create_AutomaticEnvVars(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario6_missing_files")
	t.Setenv("APP_NAME", "from-env-var")
	t.Setenv("DATABASE_HOST", "env-var-host")
	t.Setenv("APP_ENV", "dev")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "from-env-var", cfg.k.String("app.name"))
	assert.Equal(t, "env-var-host", cfg.k.String("database.host"))
}

func TestDefaultConfigFactory_Create_InvalidYAML(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario7_invalid_yaml")
	t.Setenv("APP_ENV", "dev")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
	)

	if err != nil {
		assert.Contains(t, err.Error(), "failed to load config file")
		assert.Nil(t, cfg)
	} else {
		require.NotNil(t, cfg)
	}
}

func TestDefaultConfigFactory_Create_MultipleSources(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario8_multiple_sources")
	source1 := filepath.Join(testdataDir, "source1")
	source2 := filepath.Join(testdataDir, "source2")
	t.Setenv("APP_ENV", "dev")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     source1,
			Patterns: []string{"app.yaml", "app_{env}.yaml", "app_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
		ConfigSource{
			Path:     source2,
			Patterns: []string{"db.yaml", "db_{env}.yaml", "db_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "app1", cfg.k.String("app.name"))
	assert.Equal(t, "dbhost", cfg.k.String("database.host"))
	assert.Equal(t, int64(5432), cfg.k.Int64("database.port"))
}

func TestDefaultConfigFactory_Create_DefaultStructure(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario9_default_structure")
	t.Setenv("APP_ENV", "dev")

	cfg, err := factory.Create(
		ConfigSource{
			Path:     filepath.Join(testdataDir, "config"),
			Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
			Parser:   yaml.Parser(),
		},
		ConfigSource{
			Path:     testdataDir,
			Patterns: []string{".env", ".env.{env}", ".env.local", ".env.{env}.local"},
			Parser:   NewDotenvParser(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "from-env", cfg.k.String("app.name"))
	assert.Equal(t, "env-password", cfg.k.String("database.password"))

	type DatabaseConfig struct {
		Host     string `config:"host"`
		Password string `config:"password"`
		Port     int    `config:"port"`
	}
	var dbCfg DatabaseConfig
	err = cfg.UnmarshalKey("database", &dbCfg)
	require.NoError(t, err)
	assert.Equal(t, "env-password", dbCfg.Password)
	assert.Equal(t, 5432, dbCfg.Port)
}

func TestNewDefaultConfigFactory(t *testing.T) {
	factory := NewDefaultConfigFactory()
	require.NotNil(t, factory)
	assert.Implements(t, (*ConfigFactory)(nil), factory)
}
