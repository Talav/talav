package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	labelOneTwoThree       = ".1.2.3"
	labelRed               = ".red"
	csvOneTwoThree         = "1,2,3"
	matrixColorSize        = ";color=blue;size=large"
	goFieldAPIKey          = "APIKey"
	goFieldAge             = "Age"
	goFieldBody            = "Body"
	goFieldEmail           = "Email"
	goFieldFile            = "File"
	goFieldFilter          = "Filter"
	goFieldForm            = "Form"
	goFieldItems           = "Items"
	valueJane              = "Jane"
	valueJohn              = "John"
	valueNYC               = "NYC"
	goFieldName            = "Name"
	goFieldTags            = "Tags"
	goFieldUser            = "User"
	headerClientVersion    = "X-Client-Version"
	headerRequestID        = "X-Request-ID"
	valueABC123            = "abc123"
	keyAge                 = "age"
	mimeJSON               = "application/json"
	mimeOctetStream        = "application/octet-stream"
	valueBlue              = "blue"
	valueCar               = "car"
	keyCity                = "city"
	keyColor               = "color"
	keyEmail               = "email"
	caseEmptyString        = "empty string"
	bodyFile               = "file"
	keyFilter              = "filter"
	keyIDs                 = "ids"
	queryIDsSingle         = "ids=1"
	queryIDsSpaceDelimited = "ids=1%202%203"
	queryIDsPipeDelimited  = "ids=1%7C2%7C3"
	msgInvalidBodyType     = "invalid body type"
	msgInvalidStyle        = "invalid style"
	keyItems               = "items"
	valueJaneLower         = "jane"
	valueJohnLower         = "john"
	valueEmailJohn         = "john@example.com"
	tagJSON                = "json"
	valueLarge             = "large"
	bodyMultipart          = "multipart"
	keyName                = "name"
	queryName              = "name=john"
	queryNameAge           = "name=john&age=30"
	valueRed               = "red"
	caseSingleValue        = "single value"
	keySize                = "size"
	bodyStructured         = "structured"
	keyTag                 = "tag"
	keyTags                = "tags"
	valueTest              = "test"
	keyType                = "type"
	valueUnknown           = "unknown"
	keyUser                = "user"
	keyUserName            = "user_name"
	valuePlain             = "value"
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
			data: queryName,
			want: map[string]any{
				keyName: valueJohnLower,
			},
		},
		{
			name: "multiple key-values",
			data: "name=john&age=30&city=NYC",
			want: map[string]any{
				keyName: valueJohnLower,
				keyAge:  "30",
				keyCity: valueNYC,
			},
		},
		{
			name: "comma separated values",
			data: "ids=1,2,3",
			want: map[string]any{
				keyIDs: []any{"1", "2", "3"},
			},
		},
		{
			name: "repeated keys (explode)",
			data: "ids=1&ids=2&ids=3",
			want: map[string]any{
				keyIDs: []any{"1", "2", "3"},
			},
		},
		{
			name: "dotted notation (nested)",
			data: "filter.type=car&filter.color=red",
			want: map[string]any{
				keyFilter: map[string]any{
					keyType:  valueCar,
					keyColor: valueRed,
				},
			},
		},
		{
			name: "deeply nested dotted notation",
			data: "user.profile.name=john&user.profile.age=30",
			want: map[string]any{
				keyUser: map[string]any{
					"profile": map[string]any{
						keyName: valueJohnLower,
						keyAge:  "30",
					},
				},
			},
		},
		{
			name: caseEmptyString,
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
				keyName:  "John Doe",
				keyEmail: valueEmailJohn,
			},
		},
		{
			name: "special characters",
			data: "query=hello+world&tag=%23golang",
			want: map[string]any{
				"query": "hello world",
				keyTag:  "#golang",
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
			data: valueJohnLower,
			want: valueJohnLower,
		},
		{
			name: caseEmptyString,
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
			data: csvOneTwoThree,
			want: csvOneTwoThree,
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
			data:    labelRed,
			explode: false,
			want:    valueRed,
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
			data:    labelOneTwoThree,
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
			data:    labelRed,
			explode: true,
			want:    valueRed,
		},
		{
			name:    "without leading dot (handled)",
			data:    valueRed,
			explode: false,
			want:    valueRed,
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
				keyColor: valueBlue,
			},
		},
		{
			name:    "array non-exploded (comma-separated)",
			data:    ";ids=1,2,3",
			explode: false,
			want: map[string]any{
				keyIDs: []any{"1", "2", "3"},
			},
		},
		{
			name:    "array exploded (repeated keys)",
			data:    ";ids=1;ids=2;ids=3",
			explode: true,
			want: map[string]any{
				keyIDs: []any{"1", "2", "3"},
			},
		},
		{
			name:    "multiple keys non-exploded",
			data:    matrixColorSize,
			explode: false,
			want: map[string]any{
				keyColor: valueBlue,
				keySize:  valueLarge,
			},
		},
		{
			name:    "multiple keys exploded",
			data:    matrixColorSize,
			explode: true,
			want: map[string]any{
				keyColor: []any{valueBlue},
				keySize:  []any{valueLarge},
			},
		},
		{
			name:    caseEmptyString,
			data:    ";",
			explode: false,
			want:    map[string]any{},
		},
		{
			name:    "without leading semicolon",
			data:    "color=blue",
			explode: false,
			want: map[string]any{
				keyColor: valueBlue,
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
			decodeFn:  func(q string) (map[string]any, error) { return decoder.decodeSpaceDelimited(q) },
			separator: " ",
			tests: []struct {
				name    string
				data    string
				want    map[string]any
				wantErr bool
			}{
				{
					name: "single key with separated values",
					data: queryIDsSpaceDelimited,
					want: map[string]any{keyIDs: []any{"1", "2", "3"}},
				},
				{
					name: caseSingleValue,
					data: queryIDsSingle,
					want: map[string]any{keyIDs: "1"},
				},
				{
					name: "multiple keys",
					data: "ids=1%202&names=john%20jane",
					want: map[string]any{keyIDs: []any{"1", "2"}, "names": []any{valueJohnLower, valueJaneLower}},
				},
				{
					name: caseEmptyString,
					data: "",
					want: map[string]any{},
				},
			},
		},
		{
			name:      "pipe_delimited",
			decodeFn:  func(q string) (map[string]any, error) { return decoder.decodePipeDelimited(q) },
			separator: "|",
			tests: []struct {
				name    string
				data    string
				want    map[string]any
				wantErr bool
			}{
				{
					name: "single key with separated values",
					data: queryIDsPipeDelimited,
					want: map[string]any{keyIDs: []any{"1", "2", "3"}},
				},
				{
					name: caseSingleValue,
					data: queryIDsSingle,
					want: map[string]any{keyIDs: "1"},
				},
				{
					name: "multiple keys",
					data: "ids=1%7C2&names=john%7Cjane",
					want: map[string]any{keyIDs: []any{"1", "2"}, "names": []any{valueJohnLower, valueJaneLower}},
				},
				{
					name: caseEmptyString,
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
				keyFilter: map[string]any{
					keyType: valueCar,
				},
			},
		},
		{
			name: "multiple nested properties",
			data: "filter%5Btype%5D=car&filter%5Bcolor%5D=red",
			want: map[string]any{
				keyFilter: map[string]any{
					keyType:  valueCar,
					keyColor: valueRed,
				},
			},
		},
		{
			name: "deeply nested",
			data: "user%5Bprofile%5D%5Baddress%5D%5Bcity%5D=NYC",
			want: map[string]any{
				keyUser: map[string]any{
					"profile": map[string]any{
						"address": map[string]any{
							keyCity: valueNYC,
						},
					},
				},
			},
		},
		{
			name: "multiple values for same key",
			data: "ids%5B%5D=1&ids%5B%5D=2&ids%5B%5D=3",
			want: map[string]any{
				keyIDs: map[string]any{
					"": []any{"1", "2", "3"},
				},
			},
		},
		{
			name: "mixed bracket and regular keys",
			data: "filter%5Btype%5D=car&sort=name",
			want: map[string]any{
				keyFilter: map[string]any{
					keyType: valueCar,
				},
				"sort": keyName,
			},
		},
		{
			name: caseEmptyString,
			data: "",
			want: map[string]any{},
		},
		{
			name: "regular key-value (no brackets)",
			data: queryName,
			want: map[string]any{
				keyName: valueJohnLower,
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
			name:     caseSingleValue,
			valSlice: []string{valueJohnLower},
			want:     valueJohnLower,
		},
		{
			name:     "single empty value",
			valSlice: []string{""},
			want:     nil,
		},
		{
			name:     "multiple values",
			valSlice: []string{valueJohnLower, valueJaneLower},
			want:     []any{valueJohnLower, valueJaneLower},
		},
		{
			name:     "comma separated in single value",
			valSlice: []string{csvOneTwoThree},
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
			query: queryIDsPipeDelimited,
			sep:   "|",
			want: map[string]any{
				keyIDs: []any{"1", "2", "3"},
			},
		},
		{
			name:  "space delimiter",
			query: queryIDsSpaceDelimited,
			sep:   " ",
			want: map[string]any{
				keyIDs: []any{"1", "2", "3"},
			},
		},
		{
			name:  "custom delimiter",
			query: "ids=1%3B2%3B3",
			sep:   ";",
			want: map[string]any{
				keyIDs: []any{"1", "2", "3"},
			},
		},
		{
			name:  "single value (no delimiter)",
			query: queryIDsSingle,
			sep:   "|",
			want: map[string]any{
				keyIDs: "1",
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
			value:   valueJohnLower,
			style:   StyleSimple,
			explode: false,
			want:    valueJohnLower,
		},
		{
			name:    "label style non-exploded",
			value:   labelRed,
			style:   StyleLabel,
			explode: false,
			want:    valueRed,
		},
		{
			name:    "label style exploded array",
			value:   labelOneTwoThree,
			style:   StyleLabel,
			explode: true,
			want:    []any{"1", "2", "3"},
		},
		{
			name:    "invalid style for single value - form",
			value:   valueTest,
			style:   StyleForm,
			explode: false,
			wantErr: true,
		},
		{
			name:    "invalid style for single value - matrix",
			value:   valueTest,
			style:   StyleMatrix,
			explode: false,
			wantErr: true,
		},
		{
			name:    "invalid style for single value - deepObject",
			value:   valueTest,
			style:   StyleDeepObject,
			explode: false,
			wantErr: true,
		},
		{
			name:    "unknown style",
			value:   valueTest,
			style:   valueUnknown,
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
			value:   queryNameAge,
			style:   StyleForm,
			explode: true,
			want: map[string]any{
				keyName: valueJohnLower,
				keyAge:  "30",
			},
		},
		{
			name:    "matrix style",
			value:   matrixColorSize,
			style:   StyleMatrix,
			explode: false,
			want: map[string]any{
				keyColor: valueBlue,
				keySize:  valueLarge,
			},
		},
		{
			name:    "space delimited",
			value:   queryIDsSpaceDelimited,
			style:   StyleSpaceDelimited,
			explode: false,
			want: map[string]any{
				keyIDs: []any{"1", "2", "3"},
			},
		},
		{
			name:    "pipe delimited",
			value:   queryIDsPipeDelimited,
			style:   StylePipeDelimited,
			explode: false,
			want: map[string]any{
				keyIDs: []any{"1", "2", "3"},
			},
		},
		{
			name:    "deep object",
			value:   "filter%5Btype%5D=car",
			style:   StyleDeepObject,
			explode: true,
			want: map[string]any{
				keyFilter: map[string]any{
					keyType: valueCar,
				},
			},
		},
		{
			name:    "invalid style for map - simple",
			value:   valueTest,
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
			value:   valueTest,
			style:   valueUnknown,
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

func TestDecoder_StyleDecoders_UsesTagNames(t *testing.T) {
	decoder := newTestDecoder()

	tests := []struct {
		name     string
		data     string
		style    Style
		explode  bool
		decoder  func(string) (map[string]any, error)
		expected map[string]any
	}{
		{
			name:  "form style - uses tag names (not field names)",
			data:  "user_name=John&age=30&email_address=john@example.com",
			style: StyleForm,
			decoder: func(d string) (map[string]any, error) {
				return decoder.decodeFormStyle(d)
			},
			expected: map[string]any{
				keyUserName:     valueJohn,      // Tag name, not "UserName"
				keyAge:          "30",           // Tag name, not "UserAge"
				"email_address": valueEmailJohn, // Tag name, not "Email"
			},
		},
		{
			name:    "matrix style - uses tag names (not field names)",
			data:    ";user_name=John;age=30",
			style:   StyleMatrix,
			explode: true,
			decoder: func(d string) (map[string]any, error) {
				return decoder.decodeMatrixStyle(d, true)
			},
			expected: map[string]any{
				keyUserName: []any{valueJohn}, // Tag name, not "UserName" (explode=true creates arrays)
				keyAge:      []any{"30"},      // Tag name, not "UserAge"
			},
		},
		{
			name:  "space delimited - uses tag names (not field names)",
			data:  "user_name=1 2 3&age=30 40",
			style: StyleSpaceDelimited,
			decoder: func(d string) (map[string]any, error) {
				return decoder.decodeSpaceDelimited(d)
			},
			expected: map[string]any{
				keyUserName: []any{"1", "2", "3"}, // Tag name, not "UserName"
				keyAge:      []any{"30", "40"},    // Tag name, not "UserAge"
			},
		},
		{
			name:  "pipe delimited - uses tag names (not field names)",
			data:  "user_name=1|2|3&age=30|40",
			style: StylePipeDelimited,
			decoder: func(d string) (map[string]any, error) {
				return decoder.decodePipeDelimited(d)
			},
			expected: map[string]any{
				keyUserName: []any{"1", "2", "3"}, // Tag name, not "UserName"
				keyAge:      []any{"30", "40"},    // Tag name, not "UserAge"
			},
		},
		{
			name:  "deep object - uses tag names in bracket notation",
			data:  "user_name[first]=John&user_name[last]=Doe&age[value]=30",
			style: StyleDeepObject,
			decoder: func(d string) (map[string]any, error) {
				return decoder.decodeDeepObject(d)
			},
			expected: map[string]any{
				keyUserName: map[string]any{ // Tag name, not "UserName"
					"first": valueJohn,
					"last":  "Doe",
				},
				keyAge: map[string]any{ // Tag name, not "UserAge"
					valuePlain: "30",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.decoder(tt.data)
			require.NoError(t, err)

			// Verify that keys are tag names (ParamName), not field names (MapKey)
			for expectedKey, expectedValue := range tt.expected {
				actualValue, exists := result[expectedKey]
				assert.True(t, exists, "Expected key %q (tag name) should exist in result", expectedKey)
				assert.Equal(t, expectedValue, actualValue, "Value for key %q should match", expectedKey)
			}

			// Verify that field names (MapKey) are NOT in the result
			fieldNames := []string{"UserName", "UserAge", goFieldEmail}
			for _, fieldName := range fieldNames {
				_, exists := result[fieldName]
				assert.False(t, exists, "Field name %q should NOT be in result (should use tag name instead)", fieldName)
			}
		})
	}
}

// TestDecoder_StyleDecoders_UnknownFields tests that unknown fields
// are included (filtering happens at higher level, not in style decoders).
func TestDecoder_StyleDecoders_UnknownFields(t *testing.T) {
	decoder := newTestDecoder()

	tests := []struct {
		name     string
		data     string
		decoder  func(string) (map[string]any, error)
		expected map[string]any
	}{
		{
			name: "form style - all fields included (no filtering at style level)",
			data: "user_name=John&unknown_field=value&another_unknown=test",
			decoder: func(d string) (map[string]any, error) {
				return decoder.decodeFormStyle(d)
			},
			expected: map[string]any{
				keyUserName:       valueJohn,  // Tag name, not "UserName"
				"unknown_field":   valuePlain, // Unknown fields are not filtered at style decoder level
				"another_unknown": valueTest,  // Unknown fields are not filtered at style decoder level
			},
		},
		{
			name: "form style - nested unknown objects included",
			data: "user_name=John&nested.unknown=value",
			decoder: func(d string) (map[string]any, error) {
				return decoder.decodeFormStyle(d)
			},
			expected: map[string]any{
				keyUserName: valueJohn, // Tag name, not "UserName"
				"nested": map[string]any{
					valueUnknown: valuePlain, // Nested unknown fields are included
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.decoder(tt.data)
			require.NoError(t, err)

			for expectedKey, expectedValue := range tt.expected {
				actualValue, exists := result[expectedKey]
				assert.True(t, exists, "Expected key %q should exist in result", expectedKey)
				assert.Equal(t, expectedValue, actualValue, "Value for key %q should match", expectedKey)
			}
		})
	}
}

// TestDecoder_StyleDecoders_NestedStructures tests that nested structures
// use tag names at all levels (not field names).
func TestDecoder_StyleDecoders_NestedStructures(t *testing.T) {
	decoder := newTestDecoder()

	tests := []struct {
		name     string
		data     string
		decoder  func(string) (map[string]any, error)
		expected map[string]any
	}{
		{
			name: "form style - nested dotted notation uses tag names",
			data: "filter.type1=car&filter.color=red&user.name=John&user.age=30",
			decoder: func(d string) (map[string]any, error) {
				return decoder.decodeFormStyle(d)
			},
			expected: map[string]any{
				keyFilter: map[string]any{
					"type1":  valueCar, // Tag name, not "Type"
					keyColor: valueRed, // Tag name, not "Color"
				},
				keyUser: map[string]any{
					keyName: valueJohn, // Tag name, not "Name"
					keyAge:  "30",      // Tag name, not "Age"
				},
			},
		},
		{
			name: "deep object - nested bracket notation uses tag names",
			data: "filter[type1]=car&filter[color]=red&user[name]=John&user[age]=30",
			decoder: func(d string) (map[string]any, error) {
				return decoder.decodeDeepObject(d)
			},
			expected: map[string]any{
				keyFilter: map[string]any{
					"type1":  valueCar, // Tag name, not "Type"
					keyColor: valueRed, // Tag name, not "Color"
				},
				keyUser: map[string]any{
					keyName: valueJohn, // Tag name, not "Name"
					keyAge:  "30",      // Tag name, not "Age"
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.decoder(tt.data)
			require.NoError(t, err)

			// Verify nested structure with tag names (not field names)
			assertNestedMap(t, result, tt.expected, "")

			// Verify that field names are NOT present at any level
			assertNoFieldNames(t, result, []string{goFieldFilter, "Type", "Color", goFieldUser, goFieldName, goFieldAge})
		})
	}
}

// assertNestedMap recursively asserts that nested maps match expected structure.
func assertNestedMap(t *testing.T, actual, expected map[string]any, path string) {
	t.Helper()
	for key, expectedValue := range expected {
		fullPath := key
		if path != "" {
			fullPath = path + "." + key
		}

		actualValue, exists := actual[key]
		assert.True(t, exists, "Expected key %q should exist at path %q", key, path)

		if expectedMap, ok := expectedValue.(map[string]any); ok {
			actualMap, ok := actualValue.(map[string]any)
			assert.True(t, ok, "Value at %q should be a map, got %T", fullPath, actualValue)
			assertNestedMap(t, actualMap, expectedMap, fullPath)
		} else {
			assert.Equal(t, expectedValue, actualValue, "Value at %q should match", fullPath)
		}
	}
}

// assertNoFieldNames recursively checks that field names are not present in the map.
func assertNoFieldNames(t *testing.T, m map[string]any, fieldNames []string) {
	t.Helper()
	for key, value := range m {
		for _, fieldName := range fieldNames {
			assert.NotEqual(t, fieldName, key, "Field name %q should not be present as key", fieldName)
		}

		// Recursively check nested maps
		if nestedMap, ok := value.(map[string]any); ok {
			assertNoFieldNames(t, nestedMap, fieldNames)
		}
	}
}
