package fxclock

import (
	"fmt"
	"time"

	"github.com/talav/talav/pkg/component/clock"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

// ModuleName is the module name.
const ModuleName = "clock"

// FxClockModule is the [fx] clock module.
//
// It provides a [clock.Clock] implementation selected by the "clock.mode" config key:
//   - "system"  (default) — delegates to time.Now; correct for production.
//   - "fixed"             — frozen at clock.fixed_time (RFC3339); useful for
//     deterministic testing or simulating a specific instant (e.g. leap day).
//   - "offset"            — shifts wall time by clock.offset; preserves elapsed
//     time but anchors the start to a different window (e.g. market replay).
var FxClockModule = fx.Module(
	ModuleName,
	fxconfig.AsConfigWithDefaults("clock", DefaultClockConfig(), ClockConfig{}),
	fx.Provide(newClock),
)

// newClock constructs a [clock.Clock] from the resolved [ClockConfig].
func newClock(cfg ClockConfig) (clock.Clock, error) {
	switch cfg.Mode {
	case ModeSystem, "":
		return clock.SystemClock{}, nil

	case ModeFixed:
		t, err := time.Parse(time.RFC3339, cfg.FixedTime)
		if err != nil {
			return nil, fmt.Errorf("clock: invalid fixed_time %q: %w", cfg.FixedTime, err)
		}

		return clock.FixedClock{T: t}, nil

	case ModeOffset:
		return clock.OffsetClock{Base: clock.SystemClock{}, Offset: cfg.Offset}, nil

	default:
		return nil, fmt.Errorf("clock: unknown mode %q", cfg.Mode)
	}
}
