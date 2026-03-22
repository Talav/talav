package fxclock

import "time"

// Mode identifies which Clock implementation the FX module should provide.
type Mode string

const (
	// ModeSystem provides a SystemClock backed by time.Now. This is the default.
	ModeSystem Mode = "system"

	// ModeFixed provides a FixedClock frozen at the instant given by FixedTime.
	// Use this to test leap-day logic, market-open edge cases, or any scenario
	// that requires a deterministic, non-advancing time in a running process.
	ModeFixed Mode = "fixed"

	// ModeOffset provides an OffsetClock that shifts wall time by Offset.
	// The clock advances at real speed but reports times in a different window.
	// Use this for replaying historical data or anchoring a service to a
	// specific time-of-day without freezing it.
	ModeOffset Mode = "offset"
)

// ClockConfig holds configuration for the FxClockModule.
type ClockConfig struct {
	// Mode selects the clock implementation. Defaults to "system".
	Mode Mode `config:"mode"`

	// FixedTime is the frozen instant used when Mode is "fixed".
	// Must be a valid RFC3339 timestamp, e.g. "2024-02-29T09:30:00Z".
	FixedTime string `config:"fixed_time"`

	// Offset is the duration applied to wall time when Mode is "offset".
	// Accepts any value parseable as a Go duration, e.g. "-6h" or "30m".
	Offset time.Duration `config:"offset"`
}

// DefaultClockConfig returns a ClockConfig that selects the system clock.
func DefaultClockConfig() ClockConfig {
	return ClockConfig{
		Mode: ModeSystem,
	}
}
