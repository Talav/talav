package schema

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_convertBool(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      bool
		wantError bool
	}{
		{"valid true", "true", true, false},
		{"valid false", "false", false, false},
		{"case insensitive true uppercase", "TRUE", true, false},
		{"case insensitive true title case", "True", true, false},
		{"case insensitive false uppercase", "FALSE", false, false},
		{"case insensitive false title case", "False", false, false},
		{"integer one", "1", true, false},
		{"integer zero", "0", false, false},
		{"integer negative", "-1", true, false},
		{"integer positive", "42", true, false},
		{"float one", "1.0", true, false},
		{"float zero", "0.0", false, false},
		{"float negative", "-0.5", true, false},
		{"invalid string", "invalid", false, true},
		{"empty string", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertBool(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Bool())
		})
	}
}

func TestConverter_convertString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"valid string", "hello", "hello"},
		{"empty string", "", ""},
		{"unicode", "世界", "世界"},
		{"with spaces", "hello world", "hello world"},
		{"special chars", "!@#$%^&*()", "!@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertString(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.String())
		})
	}
}

func TestConverter_convertInt(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      int
		wantError bool
	}{
		{"valid integer", "42", 42, false},
		{"zero", "0", 0, false},
		{"negative", "-42", -42, false},
		{"float truncated", "12.7", 12, false},
		{"float truncated negative", "-12.7", -12, false},
		{"float zero", "0.0", 0, false},
		{"invalid string", "invalid", 0, true},
		{"empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, int(result.Int()))
		})
	}
}

func TestConverter_convertInt8(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      int8
		wantError bool
	}{
		{"valid in range", "127", 127, false},
		{"valid negative", "-128", -128, false},
		{"zero", "0", 0, false},
		{"overflow", "128", 0, true},
		{"underflow", "-129", 0, true},
		{"invalid string", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt8(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, int8(result.Int()))
		})
	}
}

func TestConverter_convertInt16(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      int16
		wantError bool
	}{
		{"valid in range", "32767", 32767, false},
		{"valid negative", "-32768", -32768, false},
		{"zero", "0", 0, false},
		{"overflow", "32768", 0, true},
		{"underflow", "-32769", 0, true},
		{"invalid string", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt16(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, int16(result.Int()))
		})
	}
}

func TestConverter_convertInt32(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      int32
		wantError bool
	}{
		{"valid in range", "2147483647", 2147483647, false},
		{"valid negative", "-2147483648", -2147483648, false},
		{"zero", "0", 0, false},
		{"overflow", "2147483648", 0, true},
		{"underflow", "-2147483649", 0, true},
		{"invalid string", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt32(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, int32(result.Int()))
		})
	}
}

func TestConverter_convertInt64(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      int64
		wantError bool
	}{
		{"valid in range", "9223372036854775807", 9223372036854775807, false},
		{"valid negative", "-9223372036854775808", -9223372036854775808, false},
		{"zero", "0", 0, false},
		{"overflow", "9223372036854775808", 0, true},
		{"underflow", "-9223372036854775809", 0, true},
		{"invalid string", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt64(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Int())
		})
	}
}

func TestConverter_convertUint(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      uint
		wantError bool
	}{
		{"valid in range", "42", 42, false},
		{"zero", "0", 0, false},
		{"negative", "-1", 0, true},
		{"invalid string", "invalid", 0, true},
		{"empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, uint(result.Uint()))
		})
	}
}

func TestConverter_convertUint8(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      uint8
		wantError bool
	}{
		{"valid in range", "255", 255, false},
		{"zero", "0", 0, false},
		{"overflow", "256", 0, true},
		{"negative", "-1", 0, true},
		{"invalid string", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint8(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, uint8(result.Uint()))
		})
	}
}

func TestConverter_convertUint16(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      uint16
		wantError bool
	}{
		{"valid in range", "65535", 65535, false},
		{"zero", "0", 0, false},
		{"overflow", "65536", 0, true},
		{"negative", "-1", 0, true},
		{"invalid string", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint16(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, uint16(result.Uint()))
		})
	}
}

func TestConverter_convertUint32(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      uint32
		wantError bool
	}{
		{"valid in range", "4294967295", 4294967295, false},
		{"zero", "0", 0, false},
		{"overflow", "4294967296", 0, true},
		{"negative", "-1", 0, true},
		{"invalid string", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint32(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, uint32(result.Uint()))
		})
	}
}

func TestConverter_convertUint64(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      uint64
		wantError bool
	}{
		{"valid in range", "18446744073709551615", 18446744073709551615, false},
		{"zero", "0", 0, false},
		{"overflow", "18446744073709551616", 0, true},
		{"negative", "-1", 0, true},
		{"invalid string", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint64(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Uint())
		})
	}
}

func TestConverter_convertFloat32(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      float32
		wantError bool
	}{
		{"valid positive", "3.14", 3.14, false},
		{"valid negative", "-3.14", -3.14, false},
		{"zero", "0.0", 0.0, false},
		{"integer", "42", 42.0, false},
		{"scientific notation", "1e10", 1e10, false},
		{"invalid string", "invalid", 0, true},
		{"empty string", "", 0, true},
		{"infinity", "+Inf", float32(math.Inf(1)), false},
		{"negative infinity", "-Inf", float32(math.Inf(-1)), false},
		{"infinity no sign", "Inf", float32(math.Inf(1)), false},
		{"NaN", "NaN", float32(math.NaN()), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertFloat32(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			got := float32(result.Float())
			switch tt.name {
			case "NaN":
				assert.True(t, math.IsNaN(float64(got)))
			case "infinity", "negative infinity", "infinity no sign":
				assert.Equal(t, math.IsInf(float64(tt.want), 0), math.IsInf(float64(got), 0))
				if math.IsInf(float64(tt.want), 1) {
					assert.True(t, math.IsInf(float64(got), 1))
				} else if math.IsInf(float64(tt.want), -1) {
					assert.True(t, math.IsInf(float64(got), -1))
				}
			default:
				assert.InDelta(t, tt.want, got, 0.0001)
			}
		})
	}
}

func TestConverter_convertFloat64(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      float64
		wantError bool
	}{
		{"valid positive", "3.141592653589793", 3.141592653589793, false},
		{"valid negative", "-3.141592653589793", -3.141592653589793, false},
		{"zero", "0.0", 0.0, false},
		{"integer", "42", 42.0, false},
		{"scientific notation", "1e100", 1e100, false},
		{"invalid string", "invalid", 0, true},
		{"empty string", "", 0, true},
		{"infinity", "+Inf", math.Inf(1), false},
		{"negative infinity", "-Inf", math.Inf(-1), false},
		{"infinity no sign", "Inf", math.Inf(1), false},
		{"NaN", "NaN", math.NaN(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertFloat64(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			got := result.Float()
			switch tt.name {
			case "NaN":
				assert.True(t, math.IsNaN(got))
			case "infinity", "negative infinity", "infinity no sign":
				assert.Equal(t, math.IsInf(tt.want, 0), math.IsInf(got, 0))
				if math.IsInf(tt.want, 1) {
					assert.True(t, math.IsInf(got, 1))
				} else if math.IsInf(tt.want, -1) {
					assert.True(t, math.IsInf(got, -1))
				}
			default:
				assert.InDelta(t, tt.want, got, 0.0001)
			}
		})
	}
}
