package config

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadYAMLConfig(t *testing.T, relDir, appEnv string) *Config {
	t.Helper()
	t.Setenv("APP_ENV", appEnv)
	factory := NewDefaultConfigFactory()
	cfg, err := factory.Create(ConfigSource{
		Path:     filepath.Join("testdata", relDir),
		Patterns: []string{"config.yaml"},
		Parser:   yaml.Parser(),
	})
	require.NoError(t, err)

	return cfg
}

func TestUnmarshalMergeKeys(t *testing.T) {
	cases := []struct {
		name   string
		relDir string
		run    func(t *testing.T, cfg *Config)
	}{
		{
			name:   "two_keys_overlay",
			relDir: "scenario_merge_keys",
			run: func(t *testing.T, cfg *Config) {
				t.Helper()
				var s struct {
					Host string `config:"host"`
					Port int    `config:"port"`
				}
				require.NoError(t, cfg.UnmarshalMergeKeys([]string{"merge_base", "merge_overlay"}, &s))
				assert.Equal(t, "localhost", s.Host)
				assert.Equal(t, 9090, s.Port)
			},
		},
		{
			name:   "second_key_unmarshal_error_wraps_key",
			relDir: "scenario_merge_keys_bad_overlay",
			run: func(t *testing.T, cfg *Config) {
				t.Helper()
				var s struct {
					Host string `config:"host"`
					Port int    `config:"port"`
				}
				err := cfg.UnmarshalMergeKeys([]string{"merge_base", "merge_bad"}, &s)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `config key "merge_bad"`)
			},
		},
		{
			name:   "slice_replaced_by_later_key",
			relDir: "scenario_merge_keys",
			run: func(t *testing.T, cfg *Config) {
				t.Helper()
				var tc struct {
					Tags []string `config:"tags"`
				}
				require.NoError(t, cfg.UnmarshalMergeKeys([]string{"merge_slice_base", "merge_slice_overlay"}, &tc))
				assert.Equal(t, []string{"c"}, tc.Tags)
			},
		},
		{
			name:   "empty_keys_no_op",
			relDir: "scenario_merge_keys",
			run: func(t *testing.T, cfg *Config) {
				t.Helper()
				var s struct {
					Host string `config:"host"`
					Port int    `config:"port"`
				}
				require.NoError(t, cfg.UnmarshalMergeKeys(nil, &s))
				require.NoError(t, cfg.UnmarshalMergeKeys([]string{}, &s))
				assert.Equal(t, "", s.Host)
				assert.Equal(t, 0, s.Port)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := loadYAMLConfig(t, tc.relDir, "dev")
			tc.run(t, cfg)
		})
	}
}

func TestUnmarshalKey_TimeDurationHook(t *testing.T) {
	factory := NewDefaultConfigFactory()
	testdataDir := filepath.Join("testdata", "scenario_duration")
	t.Setenv("APP_ENV", "dev")

	cfg, err := factory.Create(ConfigSource{
		Path:     testdataDir,
		Patterns: []string{"config.yaml"},
		Parser:   yaml.Parser(),
	})
	require.NoError(t, err)

	type ServerConfig struct {
		ReadTimeout    time.Duration `config:"read_timeout"`
		WriteTimeout   time.Duration `config:"write_timeout"`
		IdleTimeout    time.Duration `config:"idle_timeout"`
		ConnectTimeout time.Duration `config:"connect_timeout"`
	}

	var srv ServerConfig
	require.NoError(t, cfg.UnmarshalKey("server", &srv))

	assert.Equal(t, 15*time.Second, srv.ReadTimeout)
	assert.Equal(t, 30*time.Second, srv.WriteTimeout)
	assert.Equal(t, time.Minute, srv.IdleTimeout)
	assert.Equal(t, 500*time.Millisecond, srv.ConnectTimeout)
}
