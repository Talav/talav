package fxconfig

import (
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
	ConfigSources []config.ConfigSource `group:"config-sources"`
}

// NewFxConfig returns a [config.Config].
func NewFxConfig(p FxConfigParam) (*config.Config, error) {
	if len(p.ConfigSources) > 0 {
		return p.Factory.Create(p.ConfigSources...)
	}

	return p.Factory.Create()
}
