package fxlogger

import (
	"log/slog"

	"github.com/talav/talav/pkg/component/logger"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

// ModuleName is the module name.
const ModuleName = "logger"

// FxLoggerModule is the [Fx] logger module.
var FxLoggerModule = fx.Module(
	ModuleName,
	fxconfig.AsConfig("logger", logger.LoggerConfig{}),
	fx.Provide(
		logger.NewDefaultLoggerFactory,
		func(cfg logger.LoggerConfig, factory logger.LoggerFactory) (*slog.Logger, error) {
			return factory.Create(cfg)
		},
	),
)
