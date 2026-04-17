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

// TestUnmarshalKey_KoanfPathNumericSegments documents Koanf path resolution for “numeric” segments:
//
//   - YAML sequences are stored under one key (e.g. "items"); there are no separate "items.0" keys
//     in the flat map, so k.Get("items.0") is nil and UnmarshalKey("items.0", ...) does not fill a struct.
//   - A mapping with a quoted key "0" is flattened as map_string_zero.0.*; those paths unmarshal normally.
//
// UnmarshalMergeKeys uses the same UnmarshalKey machinery, so per-index paths into YAML arrays are
// not available unless you model the chain as maps or unmarshal whole slices and split in Go.
func TestUnmarshalKey_KoanfPathNumericSegments(t *testing.T) {
	cfg := loadYAMLConfig(t, "scenario_numeric_paths", "dev")

	type item struct {
		ID    string `config:"id"`
		Label string `config:"label"`
	}

	t.Run("yaml_sequence_not_addressable_by_items_dot_index", func(t *testing.T) {
		var got item
		require.NoError(t, cfg.UnmarshalKey("items.0", &got))
		assert.Empty(t, got.ID, "Koanf does not expose slice elements as items.0 for unmarshaling")

		var id string
		require.NoError(t, cfg.UnmarshalKey("items.0.id", &id))
		assert.Empty(t, id)
	})

	t.Run("yaml_sequence_unmarshal_whole_slice", func(t *testing.T) {
		var items []item
		require.NoError(t, cfg.UnmarshalKey("items", &items))
		require.Len(t, items, 2)
		assert.Equal(t, "zero", items[0].ID)
		assert.Equal(t, "one", items[1].ID)
	})

	t.Run("map_with_string_key_zero_path_works", func(t *testing.T) {
		var got item
		require.NoError(t, cfg.UnmarshalKey("map_string_zero.0", &got))
		assert.Equal(t, "map-key-string-zero", got.ID)
		assert.Equal(t, "from-map", got.Label)
	})
}

// TestUnmarshalMergeKeys_SubtreeWithSliceReplacedByLaterKey: merge keys that each carry a slice
// still follow replace-on-later-key semantics for the whole slice, not element-wise merge.
func TestUnmarshalMergeKeys_SubtreeWithSliceReplacedByLaterKey(t *testing.T) {
	cfg := loadYAMLConfig(t, "scenario_numeric_paths", "dev")

	type item struct {
		ID    string `config:"id"`
		Label string `config:"label"`
	}

	var got struct {
		Items []item `config:"items"`
	}

	require.NoError(t, cfg.UnmarshalMergeKeys([]string{"merge_base", "merge_overlay"}, &got))
	require.Len(t, got.Items, 1)
	assert.Equal(t, "overlay-replaces", got.Items[0].ID)
	assert.Equal(t, "overlay", got.Items[0].Label)
}

// TestUnmarshalMergeKeys_OverlayKeyWithNumericPathSegment: second merge key path may include a
// segment like ".0." when that segment is a map key (e.g. merge_mpk_overlay."0".second_key), not a
// YAML array index. The merge key must be a prefix whose subtree maps into the struct (here
// merge_mpk_overlay.0); a leaf-only key such as merge_mpk_overlay.0.second_key unmarshals a scalar
// and mapstructure rejects it for a struct destination.
func TestUnmarshalMergeKeys_OverlayKeyWithNumericPathSegment(t *testing.T) {
	cfg := loadYAMLConfig(t, "scenario_numeric_paths", "dev")

	var got struct {
		Title     string `config:"title"`
		Keep      string `config:"keep"`
		SecondKey string `config:"second_key"`
	}

	require.NoError(t, cfg.UnmarshalMergeKeys([]string{"merge_mpk_base", "merge_mpk_overlay.0"}, &got))
	assert.Equal(t, "base-title", got.Title)
	assert.Equal(t, "from-base", got.Keep)
	assert.Equal(t, "from-overlay-zero", got.SecondKey)
}

// TestUnmarshalMergeKeys_LeafScalarSecondMergeKey: a merge key that points at a scalar leaf
// (merge_mpk_overlay.0.second_key) cannot decode into the same struct destination as earlier keys;
// the second step errors. The first step still ran, so base fields are populated; overlay field is not.
func TestUnmarshalMergeKeys_LeafScalarSecondMergeKey(t *testing.T) {
	cfg := loadYAMLConfig(t, "scenario_numeric_paths", "dev")

	var got struct {
		Title     string `config:"title"`
		Keep      string `config:"keep"`
		SecondKey string `config:"second_key"`
	}

	err := cfg.UnmarshalMergeKeys([]string{"merge_mpk_base", "merge_mpk_overlay.0.second_key"}, &got)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `config key "merge_mpk_overlay.0.second_key"`)
	assert.Contains(t, err.Error(), "map or struct")

	assert.Equal(t, "base-title", got.Title, "first merge key applied")
	assert.Equal(t, "from-base", got.Keep, "first merge key applied")
	assert.Empty(t, got.SecondKey, "second merge step did not apply")
}

func TestUnmarshalMergeKeys_K1AndK20K3ExactPaths(t *testing.T) {
	cfg := loadYAMLConfig(t, "scenario_merge_keys_k1_k2_0_k3", "dev")

	var got struct {
		Title string `config:"title"`
		Keep  string `config:"keep"`
	}

	require.NoError(t, cfg.UnmarshalMergeKeys([]string{"k1", "k2.0.k3"}, &got))
	assert.Equal(t, "from-k2-0-k3", got.Title)
	assert.Equal(t, "from-k1-keep", got.Keep)
}
