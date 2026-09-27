package fxconfig

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/talav/talav/pkg/component/config"
	"go.uber.org/fx"
)

// ModuleName is the module name.
const ModuleName = "config"

// FxConfigModule is the [Fx] config module.
var FxConfigModule = fx.Module(
	ModuleName,
	fx.Provide(
		config.NewDefaultConfigFactory,
		NewFxConfig,
	),
)

// FxConfigParam allows injection of the required dependencies in [NewFxConfig].
type FxConfigParam struct {
	fx.In
	Factory       config.ConfigFactory
	ConfigSources []config.ConfigSource         `group:"config-sources"`
	DecodeHooks   []mapstructure.DecodeHookFunc `group:"config-decode-hooks"`
}

// NewFxConfig returns a [config.Config].
func NewFxConfig(p FxConfigParam) (*config.Config, error) {
	cfg, err := p.Factory.Create(p.ConfigSources...)
	if err != nil {
		return nil, err
	}

	return cfg.WithDecodeHooks(p.DecodeHooks...), nil
}
