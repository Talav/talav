package mapstructure

import (
	"io"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_convertBool(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      bool
		wantError bool
	}{
		// Native bool
		{"bool true", true, true, false},
		{"bool false", false, false, false},
		// Native int
		{"int positive", 42, true, false},
		{"int zero", 0, false, false},
		{"int negative", -1, true, false},
		{"int8", int8(1), true, false},
		{"int64", int64(0), false, false},
		// Native uint
		{"uint positive", uint(42), true, false},
		{"uint zero", uint(0), false, false},
		{"uint8", uint8(1), true, false},
		// Native float
		{"float64 positive", 3.14, true, false},
		{"float64 zero", 0.0, false, false},
		{"float32", float32(1.5), true, false},
		// String parsing
		{"string true", "true", true, false},
		{"string false", "false", false, false},
		{"string TRUE", "TRUE", true, false},
		{"string 1", "1", true, false},
		{"string 0", "0", false, false},
		{"string empty", "", false, false},
		{"string invalid", "invalid", false, true},
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
		name      string
		input     any
		want      string
		wantError bool
	}{
		// Native string
		{"string valid", "hello", "hello", false},
		{"string empty", "", "", false},
		{"string unicode", "世界", "世界", false},
		{"string with spaces", "hello world", "hello world", false},
		// Native bool
		{"bool true", true, "1", false},
		{"bool false", false, "0", false},
		// Native int
		{"int positive", 42, "42", false},
		{"int negative", -42, "-42", false},
		{"int zero", 0, "0", false},
		{"int64", int64(1234567890), "1234567890", false},
		// Native uint
		{"uint", uint(42), "42", false},
		{"uint64", uint64(1234567890), "1234567890", false},
		// Native float
		{"float positive", 3.14, "3.14", false},
		{"float negative", -3.14, "-3.14", false},
		{"float integer", 42.0, "42", false},
		// []byte
		{"bytes", []byte("hello"), "hello", false},
		{"bytes empty", []byte{}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertString(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.String())
		})
	}
}

func TestConverter_convertInt(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int
		wantError bool
	}{
		// Native int
		{"int positive", 42, 42, false},
		{"int negative", -42, -42, false},
		{"int zero", 0, 0, false},
		{"int64", int64(100), 100, false},
		// Native uint
		{"uint", uint(42), 42, false},
		{"uint64", uint64(100), 100, false},
		// Native float
		{"float positive", 42.9, 42, false},
		{"float negative", -42.9, -42, false},
		// Native bool
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		// String parsing
		{"string valid", "42", 42, false},
		{"string negative", "-42", -42, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string invalid", "invalid", 0, true},
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

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertInt8(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int8
		wantError bool
	}{
		// Native types
		{"int positive", 127, 127, false},
		{"int negative", -128, -128, false},
		{"uint", uint(100), 100, false},
		{"float", 50.9, 50, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "127", 127, false},
		{"string negative", "-128", -128, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "128", 0, true},
		{"string underflow", "-129", 0, true},
		{"string invalid", "invalid", 0, true},
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

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertInt16(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int16
		wantError bool
	}{
		// Native types
		{"int positive", 32767, 32767, false},
		{"int negative", -32768, -32768, false},
		{"uint", uint(30000), 30000, false},
		{"float", 1000.9, 1000, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "32767", 32767, false},
		{"string negative", "-32768", -32768, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "32768", 0, true},
		{"string underflow", "-32769", 0, true},
		{"string invalid", "invalid", 0, true},
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

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertInt32(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int32
		wantError bool
	}{
		// Native types
		{"int positive", 1000000, 1000000, false},
		{"int negative", -1000000, -1000000, false},
		{"uint", uint(1000000), 1000000, false},
		{"float", 1000000.9, 1000000, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "2147483647", 2147483647, false},
		{"string negative", "-2147483648", -2147483648, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "2147483648", 0, true},
		{"string underflow", "-2147483649", 0, true},
		{"string invalid", "invalid", 0, true},
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
		input     any
		want      int64
		wantError bool
	}{
		// Native types
		{"int positive", 1000000, 1000000, false},
		{"int negative", -1000000, -1000000, false},
		{"int64", int64(9223372036854775807), 9223372036854775807, false},
		{"uint", uint(1000000), 1000000, false},
		{"float", 1000000.9, 1000000, false},
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		// String
		{"string valid", "9223372036854775807", 9223372036854775807, false},
		{"string negative", "-9223372036854775808", -9223372036854775808, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "9223372036854775808", 0, true},
		{"string underflow", "-9223372036854775809", 0, true},
		{"string invalid", "invalid", 0, true},
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
		input     any
		want      uint
		wantError bool
	}{
		// Native int
		{"int positive", 42, 42, false},
		{"int zero", 0, 0, false},
		{"int negative", -1, 0, true},
		// Native uint
		{"uint", uint(42), 42, false},
		{"uint64", uint64(100), 100, false},
		// Native float
		{"float positive", 42.9, 42, false},
		{"float negative", -1.5, 0, true},
		// Native bool
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		// String parsing
		{"string valid", "42", 42, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string negative", "-1", 0, true},
		{"string invalid", "invalid", 0, true},
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

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertUint8(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint8
		wantError bool
	}{
		// Native types
		{"int positive", 255, 255, false},
		{"int negative", -1, 0, true},
		{"uint", uint(200), 200, false},
		{"float", 100.9, 100, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "255", 255, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "256", 0, true},
		{"string negative", "-1", 0, true},
		{"string invalid", "invalid", 0, true},
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

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertUint16(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint16
		wantError bool
	}{
		// Native types
		{"int positive", 65535, 65535, false},
		{"int negative", -1, 0, true},
		{"uint", uint(50000), 50000, false},
		{"float", 1000.9, 1000, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "65535", 65535, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "65536", 0, true},
		{"string negative", "-1", 0, true},
		{"string invalid", "invalid", 0, true},
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

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertUint32(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint32
		wantError bool
	}{
		// Native types
		{"int positive", 1000000, 1000000, false},
		{"int negative", -1, 0, true},
		{"uint", uint(4294967295), 4294967295, false},
		{"float", 1000000.9, 1000000, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "4294967295", 4294967295, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "4294967296", 0, true},
		{"string negative", "-1", 0, true},
		{"string invalid", "invalid", 0, true},
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
		input     any
		want      uint64
		wantError bool
	}{
		// Native types
		{"int positive", 1000000, 1000000, false},
		{"int negative", -1, 0, true},
		{"uint", uint(1000000), 1000000, false},
		{"uint64", uint64(18446744073709551615), 18446744073709551615, false},
		{"float", 1000000.9, 1000000, false},
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		// String
		{"string valid", "18446744073709551615", 18446744073709551615, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "18446744073709551616", 0, true},
		{"string negative", "-1", 0, true},
		{"string invalid", "invalid", 0, true},
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
		input     any
		want      float32
		wantError bool
	}{
		// Native int
		{"int positive", 42, 42.0, false},
		{"int negative", -42, -42.0, false},
		{"int zero", 0, 0.0, false},
		{"int64", int64(100), 100.0, false},
		// Native uint
		{"uint", uint(42), 42.0, false},
		{"uint64", uint64(100), 100.0, false},
		// Native float
		{"float32", float32(3.14), 3.14, false},
		{"float64", 3.14, 3.14, false},
		// Native bool
		{"bool true", true, 1.0, false},
		{"bool false", false, 0.0, false},
		// String parsing
		{"string positive", "3.14", 3.14, false},
		{"string negative", "-3.14", -3.14, false},
		{"string zero", "0.0", 0.0, false},
		{"string integer", "42", 42.0, false},
		{"string scientific", "1e10", 1e10, false},
		{"string empty", "", 0.0, false},
		{"string invalid", "invalid", 0, true},
		{"string infinity", "+Inf", float32(math.Inf(1)), false},
		{"string negative infinity", "-Inf", float32(math.Inf(-1)), false},
		{"string NaN", "NaN", float32(math.NaN()), false},
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
			case "string NaN":
				assert.True(t, math.IsNaN(float64(got)))
			case "string infinity", "string negative infinity":
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
		input     any
		want      float64
		wantError bool
	}{
		// Native int
		{"int positive", 42, 42.0, false},
		{"int negative", -42, -42.0, false},
		{"int zero", 0, 0.0, false},
		{"int64", int64(100), 100.0, false},
		// Native uint
		{"uint", uint(42), 42.0, false},
		{"uint64", uint64(100), 100.0, false},
		// Native float
		{"float32", float32(3.14), 3.14, false},
		{"float64", 3.141592653589793, 3.141592653589793, false},
		// Native bool
		{"bool true", true, 1.0, false},
		{"bool false", false, 0.0, false},
		// String parsing
		{"string positive", "3.141592653589793", 3.141592653589793, false},
		{"string negative", "-3.141592653589793", -3.141592653589793, false},
		{"string zero", "0.0", 0.0, false},
		{"string integer", "42", 42.0, false},
		{"string scientific", "1e100", 1e100, false},
		{"string empty", "", 0.0, false},
		{"string invalid", "invalid", 0, true},
		{"string infinity", "+Inf", math.Inf(1), false},
		{"string negative infinity", "-Inf", math.Inf(-1), false},
		{"string NaN", "NaN", math.NaN(), false},
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
			case "string NaN":
				assert.True(t, math.IsNaN(got))
			case "string infinity", "string negative infinity":
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

func TestConverter_convertBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    []byte
		wantErr bool
	}{
		{name: "nil", input: nil, want: nil},
		{name: "bytes", input: []byte{1, 2, 3}, want: []byte{1, 2, 3}},
		{name: "bytes empty", input: []byte{}, want: []byte{}},
		{name: "string", input: "hello", want: []byte("hello")},
		{name: "string empty", input: "", want: []byte("")},
		{name: "any slice", input: []any{1, 2, 3}, want: []byte{1, 2, 3}},
		{name: "any slice empty", input: []any{}, want: []byte{}},
		{name: "any slice with ints", input: []any{int(65), int(66), int(67)}, want: []byte{65, 66, 67}},
		{name: "reader", input: io.NopCloser(strings.NewReader("test")), want: []byte("test")},
		{name: "reader empty", input: io.NopCloser(strings.NewReader("")), want: []byte{}},
		{name: "invalid int", input: 42, wantErr: true},
		{name: "invalid bool", input: true, wantErr: true},
		{name: "any slice invalid element", input: []any{1, "invalid", 3}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertBytes(tt.input)
			if tt.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:forcetypeassert // Test code
			assert.Equal(t, tt.want, result.Interface().([]byte))
		})
	}
}

func TestConverter_convertReadCloser(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		wantContent string
		wantNil     bool
		wantErr     bool
	}{
		{name: "nil", input: nil, wantNil: true},
		{name: "readcloser", input: io.NopCloser(strings.NewReader("hello")), wantContent: "hello"},
		{name: "reader", input: strings.NewReader("world"), wantContent: "world"},
		{name: "bytes", input: []byte("bytes"), wantContent: "bytes"},
		{name: "bytes empty", input: []byte{}, wantContent: ""},
		{name: "string", input: "string", wantContent: "string"},
		{name: "string empty", input: "", wantContent: ""},
		{name: "invalid int", input: 42, wantErr: true},
		{name: "invalid bool", input: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertReadCloser(tt.input)
			if tt.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)

			if tt.wantNil {
				assert.True(t, result.IsNil())

				return
			}

			//nolint:forcetypeassert // Test code
			rc := result.Interface().(io.ReadCloser)
			defer func() { _ = rc.Close() }()

			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, string(content))
		})
	}
}
