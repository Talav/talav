package fxclock

import (
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/talav/talav/pkg/component/clock"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

// ModuleName is the module name.
const ModuleName = "clock"

// FxClockModule provides clockwork.Clock and clock.Clock backed by one instance.
// The *clockwork.FakeClock controller is nil in system mode.
var FxClockModule = fx.Module(
	ModuleName,
	fxconfig.AsConfigWithDefaults("clock", DefaultClockConfig(), ClockConfig{}),
	fx.Provide(
		newClock,
		func(clk clockwork.Clock) clock.Clock { return clk },
	),
)

func newClock(cfg ClockConfig) (clockwork.Clock, *clockwork.FakeClock, error) {
	switch cfg.Mode {
	case ModeSystem, "":
		return clockwork.NewRealClock(), nil, nil

	case ModeFake:
		if cfg.InitialTime == "" {
			fake := clockwork.NewFakeClock()

			return fake, fake, nil
		}
		initialTime, err := time.Parse(time.RFC3339, cfg.InitialTime)
		if err != nil {
			return nil, nil, fmt.Errorf("clock: invalid initial_time %q: %w", cfg.InitialTime, err)
		}
		fake := clockwork.NewFakeClockAt(initialTime)

		return fake, fake, nil

	default:
		return nil, nil, fmt.Errorf("clock: unknown mode %q", cfg.Mode)
	}
}
