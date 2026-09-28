package fxclock

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/talav/pkg/component/clock"
	"github.com/talav/talav/pkg/component/config"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

const fakeClockConfig = `clock:
  mode: fake
  initial_time: "2024-02-29T09:30:00.123456789-04:00"
`

type injectedClocks struct {
	current clock.Clock
	timed   clockwork.Clock
	fake    *clockwork.FakeClock
}

func clockOptions(t *testing.T, configYAML string) fx.Option {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configYAML), 0o600))

	return fx.Options(
		fx.NopLogger,
		fxconfig.FxConfigModule,
		fxconfig.AsConfigSource(config.ConfigSource{
			Path:     dir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		FxClockModule,
	)
}

func startClocks(t *testing.T, configYAML string) injectedClocks {
	t.Helper()

	var clocks injectedClocks
	app := fxtest.New(t,
		clockOptions(t, configYAML),
		fx.Populate(&clocks.current, &clocks.timed, &clocks.fake),
	).RequireStart()
	t.Cleanup(func() { app.RequireStop() })

	return clocks
}

func TestModule_FxClockModule_System(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	for _, tc := range []struct {
		name   string
		config string
	}{
		{name: "default"},
		{name: "explicit", config: "clock:\n  mode: system\n"},
		{name: "empty", config: "clock:\n  mode: \"\"\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clocks := startClocks(t, tc.config)

			require.Same(t, clocks.current, clocks.timed)
			assert.Nil(t, clocks.fake)

			before := time.Now()
			got := clocks.current.Now()
			after := time.Now()
			assert.False(t, got.Before(before))
			assert.False(t, got.After(after))
		})
	}
}

func TestModule_FxClockModule_FakeDefaultTime(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
	}{
		{name: "omitted", config: "clock:\n  mode: fake\n"},
		{name: "empty", config: "clock:\n  mode: fake\n  initial_time: \"\"\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := time.Now()
			clocks := startClocks(t, tc.config)
			after := time.Now()

			require.NotNil(t, clocks.fake)
			require.Same(t, clocks.fake, clocks.current)
			require.Same(t, clocks.fake, clocks.timed)

			initialTime := clocks.current.Now()
			assert.False(t, initialTime.Before(before))
			assert.False(t, initialTime.After(after))

			clocks.fake.Advance(time.Minute)
			assert.True(t, initialTime.Add(time.Minute).Equal(clocks.current.Now()))
		})
	}
}

func TestModule_FxClockModule_FakeTimer(t *testing.T) {
	clocks := startClocks(t, fakeClockConfig)
	initialTime := time.Date(2024, time.February, 29, 13, 30, 0, 123456789, time.UTC)

	require.NotNil(t, clocks.fake)
	require.Same(t, clocks.fake, clocks.current)
	require.Same(t, clocks.fake, clocks.timed)
	require.True(t, initialTime.Equal(clocks.current.Now()))

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	fired := make(chan time.Time, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		timer := clocks.timed.NewTimer(time.Minute)
		defer timer.Stop()

		select {
		case at := <-timer.Chan():
			fired <- at
		case <-ctx.Done():
		}
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})

	require.NoError(t, clocks.fake.BlockUntilContext(ctx, 1))
	clocks.fake.Advance(time.Minute - time.Nanosecond)
	select {
	case <-fired:
		t.Fatal("timer fired before its deadline")
	default:
	}

	clocks.fake.Advance(time.Nanosecond)
	want := initialTime.Add(time.Minute)
	select {
	case got := <-fired:
		assert.True(t, want.Equal(got))
	case <-ctx.Done():
		t.Fatal("timer did not fire after advancing fake time")
	}
	assert.True(t, want.Equal(clocks.current.Now()))
}

func TestModule_FxClockModule_IsolatedApplications(t *testing.T) {
	first := startClocks(t, fakeClockConfig)
	second := startClocks(t, fakeClockConfig)
	initialTime := second.current.Now()

	first.fake.Advance(time.Hour)

	assert.True(t, initialTime.Add(time.Hour).Equal(first.current.Now()))
	assert.True(t, initialTime.Equal(second.current.Now()))
}

func TestModule_FxClockModule_InvalidConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
		err    string
	}{
		{name: "unknown mode", config: "clock:\n  mode: unknown\n", err: "unknown mode"},
		{name: "fixed mode", config: "clock:\n  mode: fixed\n", err: "unknown mode"},
		{name: "offset mode", config: "clock:\n  mode: offset\n", err: "unknown mode"},
		{
			name:   "invalid initial time",
			config: "clock:\n  mode: fake\n  initial_time: invalid\n",
			err:    "invalid initial_time",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var clk clock.Clock
			app := fx.New(clockOptions(t, tc.config), fx.Populate(&clk))

			require.ErrorContains(t, app.Err(), tc.err)
			assert.Nil(t, clk)
		})
	}
}

func TestClock_CustomProvider(t *testing.T) {
	want := clock.OffsetClock{
		Base:   clock.FixedClock{T: time.Date(2024, time.February, 29, 9, 30, 0, 0, time.UTC)},
		Offset: -time.Hour,
	}
	var got clock.Clock

	fxtest.New(t,
		fx.NopLogger,
		fx.Provide(func() clock.Clock { return want }),
		fx.Populate(&got),
	).RequireStart().RequireStop()

	assert.Equal(t, want.Now(), got.Now())
}
