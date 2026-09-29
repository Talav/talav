package config

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/v2"
)

// Config allows to access the application configuration using koanf.
type Config struct {
	k           *koanf.Koanf
	decodeHooks []mapstructure.DecodeHookFunc
}

// WithDecodeHooks returns a copy with additional hooks, run before the built-in hooks.
// It shares the loaded values; hooks must pass unrelated inputs through.
func (c *Config) WithDecodeHooks(hooks ...mapstructure.DecodeHookFunc) *Config {
	return &Config{k: c.k, decodeHooks: slices.Concat(c.decodeHooks, hooks)}
}

// UnmarshalKey unmarshals the configuration at the given key path into the provided struct.
// Uses "config" struct tag by default (instead of "koanf").
// Options run in order after defaults and registered hooks; Result always points to dest.
//
// Decode hooks applied automatically:
//   - strings → time.Duration  (e.g. "15s", "100ms")
//   - comma-separated strings → []string slices
func (c *Config) UnmarshalKey(key string, dest any, options ...func(*mapstructure.DecoderConfig)) error {
	return c.unmarshalKey(key, dest, options...)
}

// UnmarshalMergeKeys unmarshals each key path in order into dest using the same rules as [Config.UnmarshalKey].
// Later keys overlay earlier ones for fields present in each subtree; fields absent from a subtree are left unchanged.
// Slices and maps are replaced when the source subtree defines them, not merged element-wise.
//
// An empty keys slice is a no-op. On failure, the error wraps the failing key with [fmt.Errorf] using %w.
// Options configure a fresh decoder for each key; ErrorUnset checks each input separately.
func (c *Config) UnmarshalMergeKeys(keys []string, dest any, options ...func(*mapstructure.DecoderConfig)) error {
	for _, k := range keys {
		if err := c.unmarshalKey(k, dest, options...); err != nil {
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

func (c *Config) unmarshalKey(key string, dest any, options ...func(*mapstructure.DecoderConfig)) error {
	decoderConfig := &mapstructure.DecoderConfig{
		TagName: "config",
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.ComposeDecodeHookFunc(c.decodeHooks...),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
		WeaklyTypedInput: true,
		// Explicit false so overlay via [Config.UnmarshalMergeKeys] keeps fields missing from later keys.
		ZeroFields: false,
	}
	for _, option := range options {
		option(decoderConfig)
	}

	return c.k.UnmarshalWithConf(key, dest, koanf.UnmarshalConf{
		Tag:           decoderConfig.TagName,
		DecoderConfig: decoderConfig,
	})
}

type AppConfig struct {
	Env     string `config:"env"`
	Name    string `config:"name"`
	Version string `config:"version"`
}

// StringParserHook parses strings into T. Non-string inputs must already be T or *T;
// other destination types pass through unchanged. Strings are parsed even when T is string.
func StringParserHook[T any](parse func(string) (T, error)) mapstructure.DecodeHookFuncType {
	target := reflect.TypeFor[T]()

	return func(_ reflect.Type, to reflect.Type, data any) (any, error) {
		if to != target {
			return data, nil
		}

		switch value := data.(type) {
		case string:
			return parse(value)
		case *T, T:
			return data, nil
		default:
			return nil, fmt.Errorf("expected a string for %s, got %T", target, data)
		}
	}
}

// StrictDecode requires non-pointer fields and rejects null map values and slice elements.
// Existing decode hooks are preserved; optional pointer fields must be omitted rather than null.
func StrictDecode(dc *mapstructure.DecoderConfig) {
	dc.ErrorUnset = true
	dc.AllowUnsetPointer = true
	if dc.DecodeHook == nil {
		dc.DecodeHook = rejectNull

		return
	}
	dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(rejectNull, dc.DecodeHook)
}

func rejectNull(_ reflect.Type, to reflect.Type, data any) (any, error) {
	//nolint:exhaustive // Null entries are checked at container boundaries.
	switch to.Kind() {
	case reflect.Struct, reflect.Map:
		if values, ok := data.(map[string]any); ok {
			var fields []string
			for key, value := range values {
				if value == nil {
					fields = append(fields, key)
				}
			}
			if len(fields) > 0 {
				slices.Sort(fields)

				return nil, fmt.Errorf("has null fields: %s", strings.Join(fields, ", "))
			}
		}
	case reflect.Slice:
		if values, ok := data.([]any); ok {
			for i, value := range values {
				if value == nil {
					return nil, fmt.Errorf("element %d is null", i)
				}
			}
		}
	}

	return data, nil
}
