package fxcasbin

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

// ModuleName is the module name.
const ModuleName = "casbin"

// FxCasbinModule is the Fx Casbin module.
var FxCasbinModule = fx.Module(
	ModuleName,
	fxconfig.AsConfigWithDefaults("casbin", CasbinConfig{}, CasbinConfig{}),
	fx.Provide(NewEnforcer),
)

// NewEnforcer creates a new Casbin enforcer with the provided adapter.
func NewEnforcer(adapter persist.Adapter, cfg CasbinConfig) (*casbin.Enforcer, error) {
	return casbin.NewEnforcer(cfg.ModelPath, adapter)
}

