package clock_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/talav/talav/pkg/component/clock"
)

func TestSystemClock_Now_IsBetweenBeforeAndAfter(t *testing.T) {
	clk := clock.SystemClock{}

	// Bracket Now() so the result must fall within [before, after].
	before := time.Now()
	got := clk.Now()
	after := time.Now()

	assert.False(t, got.Before(before), "Now() must not precede the call")
	assert.False(t, got.After(after), "Now() must not exceed the return time")
}

func TestSystemClock_Now_Advances(t *testing.T) {
	clk := clock.SystemClock{}

	first := clk.Now()
	time.Sleep(time.Millisecond)
	second := clk.Now()

	assert.True(t, second.After(first), "successive calls to Now() should advance")
}

func TestFixedClock_Now_ReturnsFixedInstant(t *testing.T) {
	frozen := time.Date(2024, time.February, 29, 12, 0, 0, 0, time.UTC)
	clk := clock.FixedClock{T: frozen}

	assert.Equal(t, frozen, clk.Now())
}

func TestFixedClock_Now_IsIdempotent(t *testing.T) {
	frozen := time.Date(2024, time.February, 29, 9, 30, 0, 0, time.UTC)
	clk := clock.FixedClock{T: frozen}

	assert.Equal(t, clk.Now(), clk.Now(), "FixedClock must return the same value on every call")
}

func TestOffsetClock_Now_ShiftsPositive(t *testing.T) {
	base := clock.FixedClock{T: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}
	clk := clock.OffsetClock{Base: base, Offset: 2 * time.Hour}

	want := time.Date(2024, time.January, 1, 14, 0, 0, 0, time.UTC)
	assert.Equal(t, want, clk.Now())
}

func TestOffsetClock_Now_ShiftsNegative(t *testing.T) {
	base := clock.FixedClock{T: time.Date(2024, time.June, 15, 10, 0, 0, 0, time.UTC)}
	clk := clock.OffsetClock{Base: base, Offset: -6 * time.Hour}

	want := time.Date(2024, time.June, 15, 4, 0, 0, 0, time.UTC)
	assert.Equal(t, want, clk.Now())
}

func TestOffsetClock_Now_PreservesElapsedTime(t *testing.T) {
	// Base uses a real SystemClock so elapsed time is real, but the reported
	// time is shifted to a historical window.
	const offset = -24 * time.Hour
	clk := clock.OffsetClock{Base: clock.SystemClock{}, Offset: offset}

	first := clk.Now()
	time.Sleep(time.Millisecond)
	second := clk.Now()

	assert.True(t, second.After(first), "OffsetClock should advance like its base clock")
}

func TestOffsetClock_ZeroOffset_MatchesBase(t *testing.T) {
	frozen := time.Date(2025, time.March, 15, 8, 0, 0, 0, time.UTC)
	base := clock.FixedClock{T: frozen}
	clk := clock.OffsetClock{Base: base, Offset: 0}

	assert.Equal(t, frozen, clk.Now())
}
