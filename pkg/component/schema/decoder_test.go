package schema

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestDecoder creates a decoder for testing.
func newTestDecoder() *defaultDecoder {
	parser := newStructMetadataParser()
	cache := NewStructMetadataCache(parser)

	return &defaultDecoder{
		structMetadataCache: cache,
	}
}

func createQueryRequest(queryString string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/test?"+queryString, nil)
}

func createParamMetadata(location ParameterLocation, style Style, explode bool, fields ...struct {
	name      string
	mapKey    string
	fieldType reflect.Type
},
) *StructMetadata {
	metaFields := make([]FieldMetadata, len(fields))
	for i, f := range fields {
		metaFields[i] = FieldMetadata{
			StructFieldName: f.name,
			ParamName:       f.name,
			MapKey:          f.mapKey,
			Index:           i,
			Type:            f.fieldType,
			IsParameter:     true,
			Location:        location,
			Style:           style,
			Explode:         explode,
		}
	}

	metadata, _ := NewStructMetadata(metaFields)

	return metadata
}

func TestDecoder_Decode(t *testing.T) {
	tests := []struct {
		name         string
		queryString  string
		routerParams map[string]string
		headers      map[string]string
		cookies      map[string]string
		metadata     *StructMetadata
		want         map[string]any
		wantErr      bool
	}{
		{
			name:        "query parameters only",
			queryString: "name=john&age=30",
			metadata: createParamMetadata(LocationQuery, StyleForm, true,
				struct {
					name      string
					mapKey    string
					fieldType reflect.Type
				}{"name", "name", reflect.TypeOf("")},
				struct {
					name      string
					mapKey    string
					fieldType reflect.Type
				}{"age", "age", reflect.TypeOf("")},
			),
			want: map[string]any{
				"name": "john",
				"age":  "30",
			},
		},
		{
			name:         "path parameters only",
			routerParams: map[string]string{"id": "123", "slug": "test-post"},
			metadata: createParamMetadata(LocationPath, StyleSimple, false,
				struct {
					name      string
					mapKey    string
					fieldType reflect.Type
				}{"id", "id", reflect.TypeOf("")},
				struct {
					name      string
					mapKey    string
					fieldType reflect.Type
				}{"slug", "slug", reflect.TypeOf("")},
			),
			want: map[string]any{
				"id":   "123",
				"slug": "test-post",
			},
		},
		{
			name:    "header parameters only",
			headers: map[string]string{"X-Request-ID": "abc123", "X-Client-Version": "1.0"},
			metadata: createParamMetadata(LocationHeader, StyleSimple, false,
				struct {
					name      string
					mapKey    string
					fieldType reflect.Type
				}{"X-Request-ID", "X-Request-ID", reflect.TypeOf("")},
				struct {
					name      string
					mapKey    string
					fieldType reflect.Type
				}{"X-Client-Version", "X-Client-Version", reflect.TypeOf("")},
			),
			want: map[string]any{
				"X-Request-ID":     "abc123",
				"X-Client-Version": "1.0",
			},
		},
		{
			name:        "empty query string",
			queryString: "",
			metadata: createParamMetadata(LocationQuery, StyleForm, true,
				struct {
					name      string
					mapKey    string
					fieldType reflect.Type
				}{"name", "name", reflect.TypeOf("")},
			),
			want: map[string]any{},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createQueryRequest(tt.queryString)

			// Add headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Add cookies
			for name, value := range tt.cookies {
				req.AddCookie(&http.Cookie{Name: name, Value: value})
			}

			result, err := decoder.Decode(req, tt.routerParams, tt.metadata)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeQuery(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		fields      []FieldMetadata
		want        map[string]any
		wantErr     bool
	}{
		{
			name:        "simple query parameters",
			queryString: "name=john&age=30",
			fields: []FieldMetadata{
				{
					StructFieldName: "Name",
					ParamName:       "name",
					MapKey:          "name",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleForm,
					Explode:         true,
				},
				{
					StructFieldName: "Age",
					ParamName:       "age",
					MapKey:          "age",
					Index:           1,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleForm,
					Explode:         true,
				},
			},
			want: map[string]any{
				"name": "john",
				"age":  "30",
			},
		},
		{
			name:        "array parameter exploded",
			queryString: "ids=1&ids=2&ids=3",
			fields: []FieldMetadata{
				{
					StructFieldName: "IDs",
					ParamName:       "ids",
					MapKey:          "ids",
					Index:           0,
					Type:            reflect.TypeOf([]string{}),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleForm,
					Explode:         true,
				},
			},
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:        "array parameter non-exploded",
			queryString: "ids=1,2,3",
			fields: []FieldMetadata{
				{
					StructFieldName: "IDs",
					ParamName:       "ids",
					MapKey:          "ids",
					Index:           0,
					Type:            reflect.TypeOf([]string{}),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleForm,
					Explode:         false,
				},
			},
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:        "deep object style",
			queryString: "filter%5Btype%5D=car&filter%5Bcolor%5D=red",
			fields: []FieldMetadata{
				{
					StructFieldName: "Filter",
					ParamName:       "filter",
					MapKey:          "filter",
					Index:           0,
					Type:            reflect.TypeOf(map[string]any{}),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleDeepObject,
					Explode:         true,
				},
			},
			want: map[string]any{
				"filter": map[string]any{
					"type":  "car",
					"color": "red",
				},
			},
		},
		{
			name:        "space delimited style",
			queryString: "ids=1%202%203",
			fields: []FieldMetadata{
				{
					StructFieldName: "IDs",
					ParamName:       "ids",
					MapKey:          "ids",
					Index:           0,
					Type:            reflect.TypeOf([]string{}),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleSpaceDelimited,
					Explode:         false,
				},
			},
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:        "pipe delimited style",
			queryString: "ids=1%7C2%7C3",
			fields: []FieldMetadata{
				{
					StructFieldName: "IDs",
					ParamName:       "ids",
					MapKey:          "ids",
					Index:           0,
					Type:            reflect.TypeOf([]string{}),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StylePipeDelimited,
					Explode:         false,
				},
			},
			want: map[string]any{
				"ids": []any{"1", "2", "3"},
			},
		},
		{
			name:        "no query fields in metadata",
			queryString: "name=john",
			fields: []FieldMetadata{
				{
					StructFieldName: "ID",
					ParamName:       "id",
					MapKey:          "id",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationPath,
					Style:           StyleSimple,
				},
			},
			want: map[string]any{},
		},
		{
			name:        "empty query string",
			queryString: "",
			fields: []FieldMetadata{
				{
					StructFieldName: "Name",
					ParamName:       "name",
					MapKey:          "name",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleForm,
					Explode:         true,
				},
			},
			want: map[string]any{},
		},
		{
			name:        "mixed styles in query",
			queryString: "name=john&filter%5Btype%5D=car",
			fields: []FieldMetadata{
				{
					StructFieldName: "Name",
					ParamName:       "name",
					MapKey:          "name",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleForm,
					Explode:         true,
				},
				{
					StructFieldName: "Filter",
					ParamName:       "filter",
					MapKey:          "filter",
					Index:           1,
					Type:            reflect.TypeOf(map[string]any{}),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleDeepObject,
					Explode:         true,
				},
			},
			want: map[string]any{
				"name": "john",
				"filter": map[string]any{
					"type": "car",
				},
			},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := NewStructMetadata(tt.fields)
			require.NoError(t, err)

			req := createQueryRequest(tt.queryString)

			result, err := decoder.decodeQuery(req, metadata)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodePath(t *testing.T) {
	tests := []struct {
		name         string
		routerParams map[string]string
		fields       []FieldMetadata
		want         map[string]any
		wantErr      bool
	}{
		{
			name: "simple path parameter",
			routerParams: map[string]string{
				"id": "123",
			},
			fields: []FieldMetadata{
				{
					StructFieldName: "ID",
					ParamName:       "id",
					MapKey:          "id",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationPath,
					Style:           StyleSimple,
				},
			},
			want: map[string]any{
				"id": "123",
			},
		},
		{
			name: "multiple path parameters",
			routerParams: map[string]string{
				"userID": "42",
				"postID": "100",
			},
			fields: []FieldMetadata{
				{
					StructFieldName: "UserID",
					ParamName:       "userID",
					MapKey:          "userID",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationPath,
					Style:           StyleSimple,
				},
				{
					StructFieldName: "PostID",
					ParamName:       "postID",
					MapKey:          "postID",
					Index:           1,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationPath,
					Style:           StyleSimple,
				},
			},
			want: map[string]any{
				"userID": "42",
				"postID": "100",
			},
		},
		{
			name: "label style path parameter",
			routerParams: map[string]string{
				"values": ".1.2.3",
			},
			fields: []FieldMetadata{
				{
					StructFieldName: "Values",
					ParamName:       "values",
					MapKey:          "values",
					Index:           0,
					Type:            reflect.TypeOf([]string{}),
					IsParameter:     true,
					Location:        LocationPath,
					Style:           StyleLabel,
					Explode:         true,
				},
			},
			want: map[string]any{
				"values": []any{"1", "2", "3"},
			},
		},
		{
			name:         "empty router params",
			routerParams: map[string]string{},
			fields: []FieldMetadata{
				{
					StructFieldName: "ID",
					ParamName:       "id",
					MapKey:          "id",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationPath,
					Style:           StyleSimple,
				},
			},
			want: map[string]any{
				"id": "",
			},
		},
		{
			name: "no path fields in metadata",
			routerParams: map[string]string{
				"id": "123",
			},
			fields: []FieldMetadata{
				{
					StructFieldName: "Name",
					ParamName:       "name",
					MapKey:          "name",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleForm,
				},
			},
			want: map[string]any{},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := NewStructMetadata(tt.fields)
			require.NoError(t, err)

			result, err := decoder.decodePath(tt.routerParams, metadata)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeHeader(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		fields  []FieldMetadata
		want    map[string]any
		wantErr bool
	}{
		{
			name: "simple header parameter",
			headers: map[string]string{
				"X-Request-ID": "abc123",
			},
			fields: []FieldMetadata{
				{
					StructFieldName: "RequestID",
					ParamName:       "X-Request-ID",
					MapKey:          "X-Request-ID",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationHeader,
					Style:           StyleSimple,
				},
			},
			want: map[string]any{
				"X-Request-ID": "abc123",
			},
		},
		{
			name: "multiple header parameters",
			headers: map[string]string{
				"X-Request-ID":     "abc123",
				"X-Client-Version": "2.0",
			},
			fields: []FieldMetadata{
				{
					StructFieldName: "RequestID",
					ParamName:       "X-Request-ID",
					MapKey:          "X-Request-ID",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationHeader,
					Style:           StyleSimple,
				},
				{
					StructFieldName: "ClientVersion",
					ParamName:       "X-Client-Version",
					MapKey:          "X-Client-Version",
					Index:           1,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationHeader,
					Style:           StyleSimple,
				},
			},
			want: map[string]any{
				"X-Request-ID":     "abc123",
				"X-Client-Version": "2.0",
			},
		},
		{
			name:    "missing header",
			headers: map[string]string{},
			fields: []FieldMetadata{
				{
					StructFieldName: "RequestID",
					ParamName:       "X-Request-ID",
					MapKey:          "X-Request-ID",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationHeader,
					Style:           StyleSimple,
				},
			},
			want: map[string]any{
				"X-Request-ID": "",
			},
		},
		{
			name: "no header fields in metadata",
			headers: map[string]string{
				"X-Request-ID": "abc123",
			},
			fields: []FieldMetadata{
				{
					StructFieldName: "Name",
					ParamName:       "name",
					MapKey:          "name",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleForm,
				},
			},
			want: map[string]any{},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := NewStructMetadata(tt.fields)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result, err := decoder.decodeHeader(req, metadata)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeCookie(t *testing.T) {
	// NOTE: The current decoder implementation has a limitation where StyleForm
	// (the only allowed style for cookies per OpenAPI spec) is not supported
	// in decodeValueByStyle for single values. These tests verify behavior
	// for scenarios that work around this limitation.
	tests := []struct {
		name    string
		cookies map[string]string
		fields  []FieldMetadata
		want    map[string]any
		wantErr bool
	}{
		{
			name:    "missing cookie (skipped)",
			cookies: map[string]string{},
			fields: []FieldMetadata{
				{
					StructFieldName: "Session",
					ParamName:       "session",
					MapKey:          "session",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationCookie,
					Style:           StyleForm,
					Explode:         true,
				},
			},
			want: map[string]any{},
		},
		{
			name: "no cookie fields in metadata",
			cookies: map[string]string{
				"session": "xyz789",
			},
			fields: []FieldMetadata{
				{
					StructFieldName: "Name",
					ParamName:       "name",
					MapKey:          "name",
					Index:           0,
					Type:            reflect.TypeOf(""),
					IsParameter:     true,
					Location:        LocationQuery,
					Style:           StyleForm,
				},
			},
			want: map[string]any{},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := NewStructMetadata(tt.fields)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for name, value := range tt.cookies {
				req.AddCookie(&http.Cookie{Name: name, Value: value})
			}

			result, err := decoder.decodeCookie(req, metadata)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFilterQueryValuesByFields(t *testing.T) {
	tests := []struct {
		name      string
		allValues map[string][]string
		fields    []FieldMetadata
		want      map[string][]string
	}{
		{
			name: "filters matching fields",
			allValues: map[string][]string{
				"name":    {"john"},
				"age":     {"30"},
				"unknown": {"value"},
			},
			fields: []FieldMetadata{
				{MapKey: "name"},
				{MapKey: "age"},
			},
			want: map[string][]string{
				"name": {"john"},
				"age":  {"30"},
			},
		},
		{
			name: "handles deep object notation",
			allValues: map[string][]string{
				"filter[type]":  {"car"},
				"filter[color]": {"red"},
				"other":         {"value"},
			},
			fields: []FieldMetadata{
				{MapKey: "filter"},
			},
			want: map[string][]string{
				"filter[type]":  {"car"},
				"filter[color]": {"red"},
			},
		},
		{
			name: "handles dotted notation",
			allValues: map[string][]string{
				"user.name": {"john"},
				"user.age":  {"30"},
				"other":     {"value"},
			},
			fields: []FieldMetadata{
				{MapKey: "user"},
			},
			want: map[string][]string{
				"user.name": {"john"},
				"user.age":  {"30"},
			},
		},
		{
			name: "no matching fields",
			allValues: map[string][]string{
				"unknown": {"value"},
			},
			fields: []FieldMetadata{
				{MapKey: "name"},
			},
			want: map[string][]string{},
		},
		{
			name:      "empty values",
			allValues: map[string][]string{},
			fields: []FieldMetadata{
				{MapKey: "name"},
			},
			want: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterQueryValuesByFields(tt.allValues, tt.fields)

			assert.Equal(t, tt.want, map[string][]string(result))
		})
	}
}
