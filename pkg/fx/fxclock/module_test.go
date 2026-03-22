package fxclock

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/talav/pkg/component/clock"
	"github.com/talav/talav/pkg/component/config"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestModule_FxClockModule_ProvidesClock(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_fxclockmodule_providesclock")
	t.Setenv("APP_ENV", "dev")

	var clk clock.Clock

	fxtest.New(
		t,
		fx.NopLogger,
		fxconfig.FxConfigModule,
		fxconfig.AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		FxClockModule,
		fx.Populate(&clk),
	).RequireStart().RequireStop()

	require.NotNil(t, clk)
}

func TestNewClock_SystemMode(t *testing.T) {
	clk, err := newClock(ClockConfig{Mode: ModeSystem})
	require.NoError(t, err)
	assert.IsType(t, clock.SystemClock{}, clk)
}

func TestNewClock_FixedMode(t *testing.T) {
	clk, err := newClock(ClockConfig{Mode: ModeFixed, FixedTime: "2024-02-29T09:30:00Z"})
	require.NoError(t, err)

	want := time.Date(2024, time.February, 29, 9, 30, 0, 0, time.UTC)
	assert.Equal(t, want, clk.Now())
}

func TestNewClock_OffsetMode(t *testing.T) {
	clk, err := newClock(ClockConfig{Mode: ModeOffset, Offset: -6 * time.Hour})
	require.NoError(t, err)
	assert.IsType(t, clock.OffsetClock{}, clk)

	before := time.Now().Add(-6 * time.Hour)
	got := clk.Now()
	after := time.Now().Add(-6 * time.Hour)

	assert.False(t, got.Before(before))
	assert.False(t, got.After(after.Add(time.Second)))
}

func TestNewClock_EmptyMode_DefaultsToSystem(t *testing.T) {
	clk, err := newClock(ClockConfig{Mode: ""})
	require.NoError(t, err)
	assert.IsType(t, clock.SystemClock{}, clk)
}

func TestNewClock_UnknownMode_ReturnsError(t *testing.T) {
	clk, err := newClock(ClockConfig{Mode: "bogus"})
	assert.Error(t, err)
	assert.Nil(t, clk)
	assert.Contains(t, err.Error(), "bogus")
}

func TestNewClock_InvalidFixedTime_ReturnsError(t *testing.T) {
	clk, err := newClock(ClockConfig{Mode: ModeFixed, FixedTime: "not-a-timestamp"})
	assert.Error(t, err)
	assert.Nil(t, clk)
	assert.Contains(t, err.Error(), "fixed_time")
}

func TestDefaultClockConfig_IsSystem(t *testing.T) {
	cfg := DefaultClockConfig()
	assert.Equal(t, ModeSystem, cfg.Mode)
}
