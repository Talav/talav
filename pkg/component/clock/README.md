# clock

Testable time source. Replaces direct `time.Now()` calls with an injectable interface so time-sensitive logic can be tested deterministically and replayed against historical data.

## Types

```go
type Clock interface {
    Now() time.Time
}

type SystemClock struct{}                        // delegates to time.Now(); use in production
type FixedClock struct{ T time.Time }            // frozen instant; use in unit tests
type OffsetClock struct{ Base Clock; Offset time.Duration } // shifts Base by Offset; use for replay
```

## Usage

```go
// Production
clk := clock.SystemClock{}

// Unit test — freeze at a specific instant
clk := clock.FixedClock{T: time.Date(2024, time.February, 29, 9, 30, 0, 0, time.UTC)}

// Replay — run at real speed but anchored to a historical window
// offset = historicalStart - time.Now() at the moment replay begins
clk := clock.OffsetClock{Base: clock.SystemClock{}, Offset: -6 * time.Hour}
```

Inject `Clock` via constructor:

```go
type OrderService struct {
    clock clock.Clock
}

func NewOrderService(clk clock.Clock) *OrderService {
    return &OrderService{clock: clk}
}

func (s *OrderService) IsMarketOpen() bool {
    now := s.clock.Now()
    hour := now.In(nyc).Hour()
    return hour >= 9 && hour < 16
}
```

## Notes

- `OffsetClock.Base` must not be nil.
