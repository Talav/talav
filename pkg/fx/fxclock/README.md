# fxclock

Provides a configured [clockwork](https://github.com/jonboulle/clockwork) clock through Fx. Existing consumers can keep using the narrow [clock.Clock](../../component/clock) interface.

## Configuration

`system` is the default and uses real time:

```yaml
clock:
  mode: system
```

`fake` starts at the current time unless `initial_time` supplies an RFC3339 timestamp, with optional fractional seconds. It advances only when explicitly controlled:

```yaml
clock:
  mode: fake
  initial_time: "2024-02-29T09:30:00Z"
```

Unknown modes and invalid supplied fake start times fail initialization. An empty mode selects `system`. `APP_ENV=test` does not change the selected mode.

## Injection

Load `fxconfig.FxConfigModule` and `fxclock.FxClockModule`. The module provides:

| Type | Purpose |
|---|---|
| `clockwork.Clock` | Time reads, timers, tickers and waits |
| `clock.Clock` | Existing consumers that only need `Now()` |
| `*clockwork.FakeClock` | Controller for fake mode; nil in system mode |

Both interfaces refer to the same instance. Each application constructs its own clock.

```go
fx.New(
    fxconfig.FxConfigModule,
    fxclock.FxClockModule,
    fx.Invoke(func(clk clockwork.Clock) {
        fmt.Println(clk.Now())
    }),
)
```

## Fake time

With fake mode configured, populate the shared clock and controller:

```go
var clk clockwork.Clock
var fake *clockwork.FakeClock

app := fxtest.New(t,
    fxconfig.FxConfigModule,
    fxclock.FxClockModule,
    fx.Populate(&clk, &fake),
).RequireStart()
t.Cleanup(func() { app.RequireStop() })

require.NotNil(t, fake)
timer := clk.NewTimer(time.Minute)
defer timer.Stop()

fake.Advance(time.Minute)
<-timer.Chan()
```

For timers created by another goroutine, use `fake.BlockUntilContext(ctx, count)` to wait for registration before advancing time. Observe application completion separately: registration does not mean awakened work has finished.

The clock controls only calls made through it. Standard-library timers, `time.Now()`, and `context.WithTimeout` continue using real time.

## Custom time sources

Omit `FxClockModule` and provide the narrow interface directly:

```go
fx.Provide(func() clock.Clock {
    return clock.FixedClock{T: initialTime}
})
```

A custom `clock.Clock` does not need to implement timers.

## Migration

This version replaces the Fx module's `fixed` and `offset` modes:

| Previous configuration | Replacement |
|---|---|
| `system` | Unchanged |
| `fixed` with `fixed_time` | `fake` with `initial_time`; remains frozen until advanced |
| `offset` with `offset` | Supply `clock.OffsetClock` explicitly as a custom `clock.Clock` |

The component's `Clock` interface, `SystemClock`, `FixedClock`, and `OffsetClock` remain available. `ModeFixed`, `ModeOffset`, and the `FixedTime`/`Offset` configuration fields are removed.
