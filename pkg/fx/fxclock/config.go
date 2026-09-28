package fxclock

// Mode identifies which Clock implementation the FX module should provide.
type Mode string

const (
	// ModeSystem uses real time and is the default.
	ModeSystem Mode = "system"

	// ModeFake uses a manually advanced clock.
	ModeFake Mode = "fake"
)

// ClockConfig holds configuration for the FxClockModule.
type ClockConfig struct {
	// Mode selects the clock implementation. Defaults to "system".
	Mode Mode `config:"mode"`

	// InitialTime is the optional RFC3339 start time for fake mode; defaults to now.
	InitialTime string `config:"initial_time"`
}

// DefaultClockConfig returns a ClockConfig that selects the system clock.
func DefaultClockConfig() ClockConfig {
	return ClockConfig{
		Mode: ModeSystem,
	}
}
