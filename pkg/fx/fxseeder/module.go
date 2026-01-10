package fxseeder

import (
	"log/slog"
	"os"

	"github.com/talav/talav/pkg/component/seeder"
	"github.com/talav/talav/pkg/component/seeder/cmd"
	"github.com/talav/talav/pkg/fx/fxcore"
	"go.uber.org/fx"
)

const ModuleName = "seeder"

// FxSeederModule is the [Fx] seeder module.
var FxSeederModule = fx.Module(
	ModuleName,
	fx.Provide(NewFxSeederRegistry),
	// Register seed command as top-level command
	fxcore.AsRootCommand(cmd.NewSeedCmd),
)

// FxSeederRegistryParam allows injection of the required dependencies in [NewFxSeederRegistry].
type FxSeederRegistryParam struct {
	fx.In
	Seeders []seeder.Seeder `group:"seeders"`
	Logger  *slog.Logger
}

// NewFxSeederRegistry returns a [seeder.SeederRegistry] with all registered seeders filtered by environment.
// Only seeders that should run in the current environment are included.
func NewFxSeederRegistry(p FxSeederRegistryParam) *seeder.SeederRegistry {
	currentEnv := getEnvironment()
	p.Logger.Info("creating seeder registry", "total_seeders", len(p.Seeders), "environment", currentEnv)

	return seeder.NewSeederRegistry(p.Seeders, currentEnv)
}

// getEnvironment returns the current environment from APP_ENV variable, defaults to "dev".
func getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	return env
}
