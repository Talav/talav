package config

import (
	"strconv"
	"strings"
	"testing"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrictDecode(t *testing.T) {
	type instrumentID string
	type price int64
	type settings struct {
		Root   struct{}     `config:"root"`
		ID     instrumentID `config:"id"`
		Credit price        `config:"credit"`
		Inner  struct {
			Threshold int `config:"threshold"`
		} `config:"inner"`
		Roots map[string]string `config:"roots"`
		List  []string          `config:"list"`
		Opt   *instrumentID     `config:"opt"`
	}
	for _, tc := range []struct {
		name   string
		key    string
		yaml   string
		delete bool
		err    string
	}{
		{name: "valid zero and omitted pointer"},
		{name: "root struct", key: "root", yaml: "~", err: "has null fields: root"},
		{name: "named string", key: "id", yaml: "", err: "has null fields: id"},
		{name: "named integer", key: "credit", yaml: "~", err: "has null fields: credit"},
		{name: "nested struct", key: "inner", yaml: "{threshold: ~}", err: "'inner' has null fields: threshold"},
		{name: "map value", key: "roots", yaml: "{SPX: ~}", err: "'roots' has null fields: SPX"},
		{name: "sorted null fields", key: "roots", yaml: "{SPX: ~, NDX: ~}", err: "'roots' has null fields: NDX, SPX"},
		{name: "slice element", key: "list", yaml: "[~]", err: `'list' element 0 is null`},
		{name: "optional pointer", key: "opt", yaml: "~", err: "has null fields: opt"},
		{name: "missing id", key: "id", delete: true, err: "has unset fields: id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := yaml.Parser().Unmarshal([]byte("root: {}\nid: es\ncredit: '0'\ninner: {threshold: 0}\nroots: {}\nlist: []"))
			require.NoError(t, err)
			if tc.key != "" {
				if tc.delete {
					delete(raw, tc.key)
				} else {
					value, err := yaml.Parser().Unmarshal([]byte("value: " + tc.yaml))
					require.NoError(t, err)
					raw[tc.key] = value["value"]
				}
			}
			cfg := (&Config{k: koanf.New(".")}).WithDecodeHooks(
				StringParserHook(func(raw string) (instrumentID, error) { return instrumentID(strings.ToUpper(raw)), nil }),
				StringParserHook(func(raw string) (price, error) {
					value, err := strconv.ParseInt(raw, 10, 64)

					return price(value), err
				}),
			)
			require.NoError(t, cfg.k.Set("settings", raw))
			var got settings
			err = cfg.UnmarshalKey("settings", &got, StrictDecode)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, instrumentID("ES"), got.ID)
			assert.Zero(t, got.Credit)
			assert.Nil(t, got.Opt)
			assert.Empty(t, got.Roots)
			assert.Empty(t, got.List)
		})
	}
}

func TestStrictDecode_WithoutExistingHooks(t *testing.T) {
	var got struct {
		Count    int
		Optional *int
	}
	dc := &mapstructure.DecoderConfig{Result: &got}
	StrictDecode(dc)
	decoder, err := mapstructure.NewDecoder(dc)
	require.NoError(t, err)
	require.NoError(t, decoder.Decode(map[string]any{"Count": 0}))
	assert.Zero(t, got)
	require.ErrorContains(t, decoder.Decode(map[string]any{"Count": nil}), "has null fields: Count")
}

func TestStrictDecode_Merge(t *testing.T) {
	cfg := &Config{k: koanf.New(".")}
	require.NoError(t, cfg.k.Set("base", map[string]any{"id": "ES"}))
	require.NoError(t, cfg.k.Set("overlay", map[string]any{"id": nil}))
	var got struct {
		ID string `config:"id"`
	}
	err := cfg.UnmarshalMergeKeys([]string{"base", "overlay"}, &got, StrictDecode)
	require.ErrorContains(t, err, `config key "overlay"`)
	require.ErrorContains(t, err, "has null fields: id")
	assert.Equal(t, "ES", got.ID)
}

func TestStrictDecode_SelectedNullKeepsNativeBehavior(t *testing.T) {
	cfg := &Config{k: koanf.New(".")}
	require.NoError(t, cfg.k.Set("null", nil))
	for _, key := range []string{"null", "missing"} {
		var got struct {
			ID string `config:"id"`
		}
		require.NoError(t, cfg.UnmarshalKey(key, &got, StrictDecode))
		assert.Zero(t, got)
	}
}
