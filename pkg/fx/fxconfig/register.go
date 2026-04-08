package fxconfig

import (
	"fmt"

	"github.com/talav/talav/pkg/component/config"
	"go.uber.org/fx"
)

// AsConfigSource registers an additional config source.
func AsConfigSource(source config.ConfigSource) fx.Option {
	return fx.Supply(
		fx.Annotate(
			source,
			fx.ResultTags(`group:"config-sources"`),
		),
	)
}

// AsConfig registers a config provider that extracts a config of type T from the main config at the given key.
// This is a generic function that each module can use to declare its config.
//
// After a successful unmarshal, if *T implements [config.Validatable], [config.Validatable.Validate] is called.
// Validation errors are wrapped with the config key and support [errors.Unwrap].
//
// Use struct type T (for example logger.LoggerConfig{}), not *T, with AsConfig.
//
// Example usage:
//
//	fxconfig.AsConfig("logger", logger.LoggerConfig{})
//
// This will provide logger.LoggerConfig as a dependency that can be injected into other constructors.
func AsConfig[T any](key string, _ T) fx.Option {
	return fx.Provide(
		func(mainConfig *config.Config) (T, error) {
			var cfg T
			if err := mainConfig.UnmarshalKey(key, &cfg); err != nil {
				return cfg, err
			}

			if err := config.Validate(&cfg); err != nil {
				return cfg, fmt.Errorf("config key %q: %w", key, err)
			}

			return cfg, nil
		},
	)
}

// AsConfigWithDefaults registers a config provider that starts with defaults, then unmarshals user config on top.
// Only fields present in user config override defaults. Uses standard Go unmarshaling behavior.
//
// After a successful unmarshal, if *T implements [config.Validatable], [config.Validatable.Validate] is called.
// Validation errors are wrapped with the config key and support [errors.Unwrap].
//
// Use struct type T (for example httpserver.Config{}), not *T, with AsConfigWithDefaults.
//
// Example usage:
//
//	fxconfig.AsConfigWithDefaults("httpserver", httpserver.DefaultConfig(), httpserver.Config{})
//
// This will provide httpserver.Config with defaults applied for missing values.
func AsConfigWithDefaults[T any](key string, defaults T, _ T) fx.Option {
	return fx.Provide(
		func(mainConfig *config.Config) (T, error) {
			cfg := defaults
			if err := mainConfig.UnmarshalKey(key, &cfg); err != nil {
				return cfg, err
			}

			if err := config.Validate(&cfg); err != nil {
				return cfg, fmt.Errorf("config key %q: %w", key, err)
			}

			return cfg, nil
		},
	)
}

// AsConfigMergeKeys registers a provider that unmarshals mergeKeys in order into a new T (zero value),
// then runs [config.Validate] when *T implements [config.Validatable].
//
// validateErrorKey is used only when validation fails (wrapped as config key %q with %w).
// It is not a koanf path and is not used by Fx for dependency resolution; types are resolved by T.
// Unmarshal failures retain wrapping from [config.Config.UnmarshalMergeKeys] (the failing path key).
//
// Use struct type T, not *T. The first mergeKeys entry is usually the YAML baseline subtree; later keys overlay.
func AsConfigMergeKeys[T any](validateErrorKey string, mergeKeys []string, _ T) fx.Option {
	return fx.Provide(
		func(mainConfig *config.Config) (T, error) {
			var cfg T
			if err := mainConfig.UnmarshalMergeKeys(mergeKeys, &cfg); err != nil {
				return cfg, err
			}

			if err := config.Validate(&cfg); err != nil {
				return cfg, fmt.Errorf("config key %q: %w", validateErrorKey, err)
			}

			return cfg, nil
		},
	)
}
