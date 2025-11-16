package fxconfig

import (
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

			return cfg, nil
		},
	)
}
