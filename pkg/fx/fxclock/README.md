# fxclock

FX module for [pkg/component/clock](../../component/clock). Provides a `clock.Clock` into the dependency graph, selecting the implementation based on configuration.

## Quick start

```go
fx.New(
    fxconfig.FxConfigModule,
    fxclock.FxClockModule,
    fx.Invoke(func(clk clock.Clock) {
        fmt.Println(clk.Now())
    }),
)
```

## Configuration

```yaml
clock:
  mode: system   # system (default) | fixed | offset
```

### Modes

**`system`** (default) — delegates to `time.Now()`.

**`fixed`** — returns a frozen instant. Useful when you need a specific date to be active at runtime (e.g., testing leap-year handling in a deployed environment).

```yaml
clock:
  mode: fixed
  fixed_time: "2024-02-29T09:30:00Z"   # RFC3339
```

**`offset`** — shifts wall time by a fixed duration. Useful for replaying a historical time window while still advancing at real speed.

```yaml
clock:
  mode: offset
  offset: "-6h"   # any time.Duration string
```

## Injected type

`clock.Clock` — the interface from `pkg/component/clock`.

In unit tests, skip this module and inject `clock.FixedClock{T: ...}` directly:

```go
fxtest.New(t,
    fx.Supply(clock.Clock(clock.FixedClock{T: myTime})),
    fx.Invoke(func(clk clock.Clock) { ... }),
).RequireStart().RequireStop()
```
