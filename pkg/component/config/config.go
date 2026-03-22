package config

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/v2"
)

// Config allows to access the application configuration using koanf.
type Config struct {
	k *koanf.Koanf
}

// UnmarshalKey unmarshals the configuration at the given key path into the provided struct.
// Uses "config" struct tag by default (instead of "koanf").
//
// Decode hooks applied automatically:
//   - strings → time.Duration  (e.g. "15s", "100ms")
//   - comma-separated strings → []string slices
func (c *Config) UnmarshalKey(key string, dest any) error {
	return c.k.UnmarshalWithConf(key, dest, koanf.UnmarshalConf{
		Tag: "config",
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
			),
			WeaklyTypedInput: true,
			Result:           dest,
		},
	})
}

type AppConfig struct {
	Env     string `config:"env"`
	Name    string `config:"name"`
	Version string `config:"version"`
}
