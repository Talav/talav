package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoder_DecodeFormStyle(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    map[string]any
		wantErr bool
	}{
		{
			name: "single key-value",
			data: "name=john",
			want: map[string]any{
				"name": "john",
			},
		},
		{
			name: "multiple key-values",
			data: "name=john&age=30&city=NYC",
			want: map[string]any{
				"name": "john",
				"age":  "30",
				"city": "NYC",
			},
		},
		{
			name: "comma separated values",
			data: "ids=1,2,3",
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name: "repeated keys (explode)",
			data: "ids=1&ids=2&ids=3",
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name: "dotted notation (nested)",
			data: "filter.type=car&filter.color=red",
			want: map[string]any{
				"filter": map[string]any{
					"type":  "car",
					"color": "red",
				},
			},
		},
		{
			name: "deeply nested dotted notation",
			data: "user.profile.name=john&user.profile.age=30",
			want: map[string]any{
				"user": map[string]any{
					"profile": map[string]any{
						"name": "john",
						"age":  "30",
					},
				},
			},
		},
		{
			name: "empty string",
			data: "",
			want: map[string]any{},
		},
		{
			name: "empty value",
			data: "name=",
			want: map[string]any{},
		},
		{
			name: "URL encoded values",
			data: "name=John%20Doe&email=john%40example.com",
			want: map[string]any{
				"name":  "John Doe",
				"email": "john@example.com",
			},
		},
		{
			name: "special characters",
			data: "query=hello+world&tag=%23golang",
			want: map[string]any{
				"query": "hello world",
				"tag":   "#golang",
			},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.decodeFormStyle(tt.data)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeSimpleStyle(t *testing.T) {
	tests := []struct {
		name string
		data string
		want any
	}{
		{
			name: "simple value",
			data: "john",
			want: "john",
		},
		{
			name: "empty string",
			data: "",
			want: "",
		},
		{
			name: "value with special chars",
			data: "hello-world_123",
			want: "hello-world_123",
		},
		{
			name: "comma separated (returned as-is)",
			data: "1,2,3",
			want: "1,2,3",
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.decodeSimpleStyle(tt.data)

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeLabelStyle(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		explode bool
		want    any
	}{
		{
			name:    "single value non-exploded",
			data:    ".red",
			explode: false,
			want:    "red",
		},
		{
			name:    "array non-exploded (comma-separated)",
			data:    ".1,2,3",
			explode: false,
			want:    []any{"1", "2", "3"},
		},
		{
			name:    "object non-exploded (comma-separated key-value pairs)",
			data:    ".x,1024,y,768",
			explode: false,
			want: map[string]any{
				"x": "1024",
				"y": "768",
			},
		},
		{
			name:    "array exploded (period-separated)",
			data:    ".1.2.3",
			explode: true,
			want:    []any{"1", "2", "3"},
		},
		{
			name:    "object exploded (period-separated key-value pairs)",
			data:    ".x.1024.y.768",
			explode: true,
			want: map[string]any{
				"x": "1024",
				"y": "768",
			},
		},
		{
			name:    "empty string non-exploded",
			data:    ".",
			explode: false,
			want:    nil,
		},
		{
			name:    "single value exploded",
			data:    ".red",
			explode: true,
			want:    "red",
		},
		{
			name:    "without leading dot (handled)",
			data:    "red",
			explode: false,
			want:    "red",
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.decodeLabelStyle(tt.data, tt.explode)

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeMatrixStyle(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		explode bool
		want    map[string]any
		wantErr bool
	}{
		{
			name:    "single key-value non-exploded",
			data:    ";color=blue",
			explode: false,
			want: map[string]any{
				"color": "blue",
			},
		},
		{
			name:    "array non-exploded (comma-separated)",
			data:    ";ids=1,2,3",
			explode: false,
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:    "array exploded (repeated keys)",
			data:    ";ids=1;ids=2;ids=3",
			explode: true,
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:    "multiple keys non-exploded",
			data:    ";color=blue;size=large",
			explode: false,
			want: map[string]any{
				"color": "blue",
				"size":  "large",
			},
		},
		{
			name:    "multiple keys exploded",
			data:    ";color=blue;size=large",
			explode: true,
			want: map[string]any{
				"color": []any{"blue"},
				"size":  []any{"large"},
			},
		},
		{
			name:    "empty string",
			data:    ";",
			explode: false,
			want:    map[string]any{},
		},
		{
			name:    "without leading semicolon",
			data:    "color=blue",
			explode: false,
			want: map[string]any{
				"color": "blue",
			},
		},
		{
			name:    "invalid format - no equals sign",
			data:    ";color",
			explode: false,
			want:    nil,
			wantErr: true,
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.decodeMatrixStyle(tt.data, tt.explode)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeDelimitedStyles(t *testing.T) {
	decoder := newTestDecoder()

	// Define decoder functions to test
	decoders := []struct {
		name      string
		decodeFn  func(string) (map[string]any, error)
		separator string
		tests     []struct {
			name    string
			data    string
			want    map[string]any
			wantErr bool
		}
	}{
		{
			name:      "space_delimited",
			decodeFn:  decoder.decodeSpaceDelimited,
			separator: " ",
			tests: []struct {
				name    string
				data    string
				want    map[string]any
				wantErr bool
			}{
				{
					name: "single key with separated values",
					data: "ids=1%202%203",
					want: map[string]any{"ids": []any{"1", "2", "3"}},
				},
				{
					name: "single value",
					data: "ids=1",
					want: map[string]any{"ids": "1"},
				},
				{
					name: "multiple keys",
					data: "ids=1%202&names=john%20jane",
					want: map[string]any{"ids": []any{"1", "2"}, "names": []any{"john", "jane"}},
				},
				{
					name: "empty string",
					data: "",
					want: map[string]any{},
				},
			},
		},
		{
			name:      "pipe_delimited",
			decodeFn:  decoder.decodePipeDelimited,
			separator: "|",
			tests: []struct {
				name    string
				data    string
				want    map[string]any
				wantErr bool
			}{
				{
					name: "single key with separated values",
					data: "ids=1%7C2%7C3",
					want: map[string]any{"ids": []any{"1", "2", "3"}},
				},
				{
					name: "single value",
					data: "ids=1",
					want: map[string]any{"ids": "1"},
				},
				{
					name: "multiple keys",
					data: "ids=1%7C2&names=john%7Cjane",
					want: map[string]any{"ids": []any{"1", "2"}, "names": []any{"john", "jane"}},
				},
				{
					name: "empty string",
					data: "",
					want: map[string]any{},
				},
			},
		},
	}

	for _, dec := range decoders {
		t.Run(dec.name, func(t *testing.T) {
			for _, tt := range dec.tests {
				t.Run(tt.name, func(t *testing.T) {
					result, err := dec.decodeFn(tt.data)

					if tt.wantErr {
						require.Error(t, err)

						return
					}

					require.NoError(t, err)
					assert.Equal(t, tt.want, result)
				})
			}
		})
	}
}

func TestDecoder_DecodeDeepObject(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    map[string]any
		wantErr bool
	}{
		{
			name: "simple deep object",
			data: "filter%5Btype%5D=car",
			want: map[string]any{
				"filter": map[string]any{
					"type": "car",
				},
			},
		},
		{
			name: "multiple nested properties",
			data: "filter%5Btype%5D=car&filter%5Bcolor%5D=red",
			want: map[string]any{
				"filter": map[string]any{
					"type":  "car",
					"color": "red",
				},
			},
		},
		{
			name: "deeply nested",
			data: "user%5Bprofile%5D%5Baddress%5D%5Bcity%5D=NYC",
			want: map[string]any{
				"user": map[string]any{
					"profile": map[string]any{
						"address": map[string]any{
							"city": "NYC",
						},
					},
				},
			},
		},
		{
			name: "multiple values for same key",
			data: "ids%5B%5D=1&ids%5B%5D=2&ids%5B%5D=3",
			want: map[string]any{
				"ids": map[string]any{
					"": []any{"1", "2", "3"},
				},
			},
		},
		{
			name: "mixed bracket and regular keys",
			data: "filter%5Btype%5D=car&sort=name",
			want: map[string]any{
				"filter": map[string]any{
					"type": "car",
				},
				"sort": "name",
			},
		},
		{
			name: "empty string",
			data: "",
			want: map[string]any{},
		},
		{
			name: "regular key-value (no brackets)",
			data: "name=john",
			want: map[string]any{
				"name": "john",
			},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.decodeDeepObject(tt.data)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_ProcessFormValue(t *testing.T) {
	tests := []struct {
		name     string
		valSlice []string
		want     any
	}{
		{
			name:     "empty slice",
			valSlice: []string{},
			want:     nil,
		},
		{
			name:     "single value",
			valSlice: []string{"john"},
			want:     "john",
		},
		{
			name:     "single empty value",
			valSlice: []string{""},
			want:     nil,
		},
		{
			name:     "multiple values",
			valSlice: []string{"john", "jane"},
			want:     []any{"john", "jane"},
		},
		{
			name:     "comma separated in single value",
			valSlice: []string{"1,2,3"},
			want:     []any{"1", "2", "3"},
		},
		{
			name:     "value with comma at end",
			valSlice: []string{"a,b,"},
			want:     []any{"a", "b"},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decoder.processFormValue(tt.valSlice)

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeDelimited(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		sep     string
		want    map[string]any
		wantErr bool
	}{
		{
			name:  "pipe delimiter",
			query: "ids=1%7C2%7C3",
			sep:   "|",
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:  "space delimiter",
			query: "ids=1%202%203",
			sep:   " ",
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:  "custom delimiter",
			query: "ids=1%3B2%3B3",
			sep:   ";",
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:  "single value (no delimiter)",
			query: "ids=1",
			sep:   "|",
			want: map[string]any{
				"ids": "1",
			},
		},
		{
			name:  "empty query",
			query: "",
			sep:   "|",
			want:  map[string]any{},
		},
		{
			name:  "empty value",
			query: "ids=",
			sep:   "|",
			want:  map[string]any{},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.decodeDelimited(tt.query, tt.sep)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeValueByStyle(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		style   Style
		explode bool
		want    any
		wantErr bool
	}{
		{
			name:    "simple style",
			value:   "john",
			style:   StyleSimple,
			explode: false,
			want:    "john",
		},
		{
			name:    "label style non-exploded",
			value:   ".red",
			style:   StyleLabel,
			explode: false,
			want:    "red",
		},
		{
			name:    "label style exploded array",
			value:   ".1.2.3",
			style:   StyleLabel,
			explode: true,
			want:    []any{"1", "2", "3"},
		},
		{
			name:    "invalid style for single value - form",
			value:   "test",
			style:   StyleForm,
			explode: false,
			wantErr: true,
		},
		{
			name:    "invalid style for single value - matrix",
			value:   "test",
			style:   StyleMatrix,
			explode: false,
			wantErr: true,
		},
		{
			name:    "invalid style for single value - deepObject",
			value:   "test",
			style:   StyleDeepObject,
			explode: false,
			wantErr: true,
		},
		{
			name:    "unknown style",
			value:   "test",
			style:   "unknown",
			explode: false,
			wantErr: true,
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.decodeValueByStyle(tt.value, tt.style, tt.explode)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeByStyle(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		style   Style
		explode bool
		want    map[string]any
		wantErr bool
	}{
		{
			name:    "form style",
			value:   "name=john&age=30",
			style:   StyleForm,
			explode: true,
			want: map[string]any{
				"name": "john",
				"age":  "30",
			},
		},
		{
			name:    "matrix style",
			value:   ";color=blue;size=large",
			style:   StyleMatrix,
			explode: false,
			want: map[string]any{
				"color": "blue",
				"size":  "large",
			},
		},
		{
			name:    "space delimited",
			value:   "ids=1%202%203",
			style:   StyleSpaceDelimited,
			explode: false,
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:    "pipe delimited",
			value:   "ids=1%7C2%7C3",
			style:   StylePipeDelimited,
			explode: false,
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:    "deep object",
			value:   "filter%5Btype%5D=car",
			style:   StyleDeepObject,
			explode: true,
			want: map[string]any{
				"filter": map[string]any{
					"type": "car",
				},
			},
		},
		{
			name:    "invalid style for map - simple",
			value:   "test",
			style:   StyleSimple,
			explode: false,
			wantErr: true,
		},
		{
			name:    "invalid style for map - label",
			value:   ".test",
			style:   StyleLabel,
			explode: false,
			wantErr: true,
		},
		{
			name:    "unknown style",
			value:   "test",
			style:   "unknown",
			explode: false,
			wantErr: true,
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decoder.decodeByStyle(tt.value, tt.style, tt.explode)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}
