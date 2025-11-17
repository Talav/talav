package fxcore

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const ModuleName = "core"

// FxCoreModule is the [Fx] core module.
var FxCoreModule = fx.Module(
	ModuleName,
)

// FxCommandsParam allows injection of commands from the cobra-commands group.
type FxCommandsParam struct {
	fx.In
	Commands []*cobra.Command `group:"talav-core-commands"`
}
