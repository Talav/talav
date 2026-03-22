package clock

import "time"

// Clock is a source of the current time.
//
// Callers that accept a Clock instead of calling time.Now directly can be
// tested deterministically and replayed against historical data without
// modifying production logic.
type Clock interface {
	Now() time.Time
}

// SystemClock delegates to time.Now. It is the correct choice for production code.
type SystemClock struct{}

// Now returns the current wall-clock time.
func (SystemClock) Now() time.Time { return time.Now() }

// FixedClock always returns a single frozen instant.
//
// It is useful for unit tests that require fully deterministic time, such as
// verifying behaviour on a leap day or a specific market open.
type FixedClock struct{ T time.Time }

// Now returns the fixed instant regardless of how much wall time has elapsed.
func (c FixedClock) Now() time.Time { return c.T }

// OffsetClock shifts the time returned by Base by a fixed duration.
//
// It preserves real elapsed time — Now advances at normal speed — but anchors
// the starting point at a different instant. Typical uses:
//   - Replaying a historical trading session: set Offset so that Now() returns
//     the historical start time at the moment replay begins.
//   - Simulating a different time zone or a future date without freezing the clock.
//
// Base must not be nil.
type OffsetClock struct {
	Base   Clock
	Offset time.Duration
}

// Now returns Base.Now() shifted by Offset.
func (c OffsetClock) Now() time.Time { return c.Base.Now().Add(c.Offset) }

// Compile-time interface assertions.
var (
	_ Clock = SystemClock{}
	_ Clock = FixedClock{}
	_ Clock = OffsetClock{}
)
