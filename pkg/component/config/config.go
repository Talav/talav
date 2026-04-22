package config

import (
	"fmt"

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
	return c.unmarshalKey(key, dest)
}

// UnmarshalMergeKeys unmarshals each key path in order into dest using the same rules as [Config.UnmarshalKey].
// Later keys overlay earlier ones for fields present in each subtree; fields absent from a subtree are left unchanged.
// Slices and maps are replaced when the source subtree defines them, not merged element-wise.
//
// An empty keys slice is a no-op. On failure, the error wraps the failing key with [fmt.Errorf] using %w.
func (c *Config) UnmarshalMergeKeys(keys []string, dest any) error {
	for _, k := range keys {
		if err := c.unmarshalKey(k, dest); err != nil {
			return fmt.Errorf("config key %q: %w", k, err)
		}
	}

	return nil
}

// Koanf returns the underlying [*koanf.Koanf]. Use it for Koanf APIs that do not need this
// package’s struct unmarshaling hooks ([Config.UnmarshalKey] duration and comma-slice decode, etc.):
// Keys, String, Get, All, Marshal, Raw, and similar. Prefer [Config.UnmarshalKey] when decoding
// into structs so those hooks apply.
func (c *Config) Koanf() *koanf.Koanf {
	return c.k
}

func (c *Config) unmarshalKey(key string, dest any) error {
	return c.k.UnmarshalWithConf(key, dest, koanf.UnmarshalConf{
		Tag: "config",
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
			),
			WeaklyTypedInput: true,
			// Explicit false so overlay via [Config.UnmarshalMergeKeys] keeps fields missing from later keys.
			ZeroFields: false,
			Result:     dest,
		},
	})
}

type AppConfig struct {
	Env     string `config:"env"`
	Name    string `config:"name"`
	Version string `config:"version"`
}
