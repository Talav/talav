package schema

import (
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Decode Query tests.
func TestDecodeQuery_FormStyle(t *testing.T) {
	decoder := NewDefaultDecoder()
	t.Run("explode false", func(t *testing.T) {
		opts, err := NewOptions(LocationQuery, StyleForm, false)
		require.NoError(t, err)

		result, err := decoder.Decode("ids=1,2,3&name=John", opts)
		require.NoError(t, err)
		assert.Equal(t, []any{"1", "2", "3"}, result["ids"])
		assert.Equal(t, "John", result["name"])
	})

	t.Run("explode true", func(t *testing.T) {
		opts, err := NewOptions(LocationQuery, StyleForm, true)
		require.NoError(t, err)

		result, err := decoder.Decode("ids=1&ids=2&ids=3&name=John", opts)
		require.NoError(t, err)
		assert.Equal(t, []any{"1", "2", "3"}, result["ids"])
		assert.Equal(t, "John", result["name"])
	})

	t.Run("nested structure", func(t *testing.T) {
		opts, err := NewOptions(LocationQuery, StyleForm, true)
		require.NoError(t, err)

		result, err := decoder.Decode("filter.type=car&filter.color=red", opts)
		require.NoError(t, err)
		filter, ok := result["filter"].(map[string]any)
		require.True(t, ok, "filter should be a map")
		assert.Equal(t, "car", filter["type"])
		assert.Equal(t, "red", filter["color"])
	})
}

func TestDecodeQuery_SpaceDelimited(t *testing.T) {
	opts, err := NewOptions(LocationQuery, StyleSpaceDelimited)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("ids=1 2 3", opts)
	require.NoError(t, err)
	assert.Equal(t, []any{"1", "2", "3"}, result["ids"])
}

func TestDecodeQuery_PipeDelimited(t *testing.T) {
	opts, err := NewOptions(LocationQuery, StylePipeDelimited)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("ids=1|2|3", opts)
	require.NoError(t, err)
	assert.Equal(t, []any{"1", "2", "3"}, result["ids"])
}

func TestDecodeQuery_DeepObject(t *testing.T) {
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("filter[type]=car&filter[color]=red", opts)
	require.NoError(t, err)
	filter, ok := result["filter"].(map[string]any)
	require.True(t, ok, "filter should be a map")
	assert.Equal(t, "car", filter["type"])
	assert.Equal(t, "red", filter["color"])
}

func TestDecodeQuery_DeepObject_ArrayIssue(t *testing.T) {
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("user[interests]=coding&user[interests]=reading&user[interests]=music", opts)
	require.NoError(t, err)

	userMap, ok := result["user"].(map[string]any)
	require.True(t, ok, "user should be a map")
	assert.Equal(t, []any{"coding", "reading", "music"}, userMap["interests"], "deep object should create array from multiple values")
}

func TestDecodeQuery_DeepObject_MultipleLevels(t *testing.T) {
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("user[profile][address1][street]=Main+St&user[profile][address][street]=Oak+Ave&user[profile][address2][street]=Elm+St", opts)
	require.NoError(t, err)

	userMap, ok := result["user"].(map[string]any)
	require.True(t, ok, "user should be a map")
	profileMap, ok := userMap["profile"].(map[string]any)
	require.True(t, ok, "profile should be a map")

	address1Map, ok := profileMap["address1"].(map[string]any)
	require.True(t, ok, "address1 should be a map")
	addressMap, ok := profileMap["address"].(map[string]any)
	require.True(t, ok, "address should be a map")
	address2Map, ok := profileMap["address2"].(map[string]any)
	require.True(t, ok, "address2 should be a map")

	assert.Equal(t, "Main St", address1Map["street"])
	assert.Equal(t, "Oak Ave", addressMap["street"])
	assert.Equal(t, "Elm St", address2Map["street"])
}

func TestDecodeQuery_DeepObject_ArrayAtNestedLevel(t *testing.T) {
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("u[l11][l21][l31]=v1&u[l11][l21][l31]=v2&u[l11][l22][l31]=v3&u[l11][l22][l31]=v4", opts)
	require.NoError(t, err)

	uMap, ok := result["u"].(map[string]any)
	require.True(t, ok, "u should be a map")
	l11Map, ok := uMap["l11"].(map[string]any)
	require.True(t, ok, "l11 should be a map")

	// l21 should exist with l31 array
	l21Map, ok := l11Map["l21"].(map[string]any)
	require.True(t, ok, "l21 should be a map")
	l21l31Value := l21Map["l31"]
	assert.Equal(t, []any{"v1", "v2"}, l21l31Value, "l21.l31 should be an array with v1 and v2")

	// l22 should exist with l31 array (sibling of l21)
	l22Map, ok := l11Map["l22"].(map[string]any)
	require.True(t, ok, "l22 should be a map")
	l22l31Value := l22Map["l31"]
	assert.Equal(t, []any{"v3", "v4"}, l22l31Value, "l22.l31 should be an array with v3 and v4")
}

func TestDecodeQuery_DeepObject_TypeConflict(t *testing.T) {
	// Test that setting a key as primitive then as object in same decode returns error
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()

	// Try to set same path as both primitive and object - should fail
	_, err = decoder.Decode("user[name]=John&user[name][first]=John", opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflict")
}

func TestDecodeQuery_DeepObject_EmptyValues(t *testing.T) {
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("user[name]=&user[email]=test@example.com", opts)
	require.NoError(t, err)

	userMap, ok := result["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "", userMap["name"], "empty value should be empty string")
	assert.Equal(t, "test@example.com", userMap["email"])
}

func TestDecodeQuery_DeepObject_OverwriteSingleWithArray(t *testing.T) {
	// Setting a key as single value, then with multiple values should create array
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("user[tags]=coding&user[tags]=reading&user[tags]=music", opts)
	require.NoError(t, err)

	userMap, ok := result["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []any{"coding", "reading", "music"}, userMap["tags"], "multiple values should create array")
}

func TestDecodeQuery_DeepObject_VeryDeepNesting(t *testing.T) {
	// Test 4+ levels of nesting
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("a[b][c][d][e]=value", opts)
	require.NoError(t, err)

	aMap, ok := result["a"].(map[string]any)
	require.True(t, ok)
	bMap, ok := aMap["b"].(map[string]any)
	require.True(t, ok)
	cMap, ok := bMap["c"].(map[string]any)
	require.True(t, ok)
	dMap, ok := cMap["d"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", dMap["e"])
}

func TestDecodeQuery_DeepObject_EmptyBracketKey(t *testing.T) {
	// Edge case: empty bracket key creates map with empty string key
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("user[]=value", opts)
	require.NoError(t, err)

	userMap, ok := result["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", userMap[""], "empty bracket creates empty string key")
}

func TestDecodeQuery_DeepObject_SingleValueNotArray(t *testing.T) {
	// Edge case: single value should remain single, not array
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("user[name]=John", opts)
	require.NoError(t, err)

	userMap, ok := result["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "John", userMap["name"], "single value should be string, not array")
	assert.IsType(t, "", userMap["name"], "single value should be string type")
}

// Decode Path tests.
func TestDecodePath_Simple(t *testing.T) {
	opts, err := NewOptions(LocationPath, StyleSimple)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("1,2,3", opts)
	require.NoError(t, err)
	assert.Equal(t, []any{"1", "2", "3"}, result[""])
}

func TestDecodePath_Label(t *testing.T) {
	t.Run("explode false", func(t *testing.T) {
		opts, err := NewOptions(LocationPath, StyleLabel, false)
		require.NoError(t, err)

		decoder := NewDefaultDecoder()
		result, err := decoder.Decode(".1,2,3", opts)
		require.NoError(t, err)
		assert.Equal(t, []any{"1", "2", "3"}, result[""])
	})

	t.Run("explode true", func(t *testing.T) {
		opts, err := NewOptions(LocationPath, StyleLabel, true)
		require.NoError(t, err)

		decoder := NewDefaultDecoder()
		result, err := decoder.Decode(".1.2.3", opts)
		require.NoError(t, err)
		assert.Equal(t, []any{"1", "2", "3"}, result[""])
	})
}

func TestDecodePath_Matrix(t *testing.T) {
	t.Run("explode false", func(t *testing.T) {
		opts, err := NewOptions(LocationPath, StyleMatrix, false)
		require.NoError(t, err)

		decoder := NewDefaultDecoder()
		result, err := decoder.Decode(";ids=1,2,3", opts)
		require.NoError(t, err)
		assert.Equal(t, []any{"1", "2", "3"}, result["ids"])
	})

	t.Run("explode true", func(t *testing.T) {
		opts, err := NewOptions(LocationPath, StyleMatrix, true)
		require.NoError(t, err)

		decoder := NewDefaultDecoder()
		result, err := decoder.Decode(";ids=1;ids=2;ids=3", opts)
		require.NoError(t, err)
		assert.Equal(t, []any{"1", "2", "3"}, result["ids"])
	})
}

// Decode Header tests.
func TestDecodeHeader_Simple(t *testing.T) {
	opts, err := NewOptions(LocationHeader, StyleSimple)
	require.NoError(t, err)

	decoder := NewDefaultDecoder()
	result, err := decoder.Decode("1,2,3", opts)
	require.NoError(t, err)
	assert.Equal(t, []any{"1", "2", "3"}, result[""])
}

// RoundTrip tests.
func TestRoundTrip(t *testing.T) {
	t.Run("query form explode false", func(t *testing.T) {
		opts, err := NewOptions(LocationQuery, StyleForm, false)
		require.NoError(t, err)

		original := map[string]any{
			"ids":  []any{"1", "2", "3"},
			"name": "John",
		}

		encoder := NewDefaultEncoder()
		encoded, err := encoder.Encode(original, opts)
		require.NoError(t, err)
		// Encoder adds ? prefix for query strings, but RawQuery doesn't include it
		encoded = strings.TrimPrefix(encoded, "?")

		decoder := NewDefaultDecoder()
		decoded, err := decoder.Decode(encoded, opts)
		require.NoError(t, err)

		assert.Equal(t, original["name"], decoded["name"])
		// Arrays should match
		assert.Equal(t, original["ids"], decoded["ids"])
	})

	t.Run("deepObject", func(t *testing.T) {
		opts, err := NewOptions(LocationQuery, StyleDeepObject)
		require.NoError(t, err)

		original := map[string]any{
			"filter": map[string]any{
				"type":  "car",
				"color": "red",
			},
		}

		encoder := NewDefaultEncoder()
		encoded, err := encoder.Encode(original, opts)
		require.NoError(t, err)
		// Encoder adds ? prefix for query strings, but RawQuery doesn't include it
		encoded = strings.TrimPrefix(encoded, "?")

		decoder := NewDefaultDecoder()
		decoded, err := decoder.Decode(encoded, opts)
		require.NoError(t, err)

		filter, ok := decoded["filter"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "car", filter["type"])
		assert.Equal(t, "red", filter["color"])
	})
}

// Integration tests.
func TestSchema_Decode_Query(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api?ids=1,2,3&name=John", nil)
	opts, _ := NewOptions(LocationQuery, StyleForm, false)
	decoder := NewDefaultDecoder()
	result, _ := decoder.Decode(req.URL.RawQuery, opts)
	assert.Equal(t, []any{"1", "2", "3"}, result["ids"])
	assert.Equal(t, "John", result["name"])
}

func TestSchema_Decode_Header(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api", nil)
	req.Header.Set("Accept", "text/html, text/plain, application/json")
	opts, _ := NewOptions(LocationHeader, StyleSimple)
	decoder := NewDefaultDecoder()
	result, _ := decoder.Decode(req.Header.Get("Accept"), opts)
	assert.Equal(t, []any{"text/html", "text/plain", "application/json"}, result[""])
}

func TestSchema_Decode_Cookie(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	req.AddCookie(&http.Cookie{Name: "ids", Value: "1"})
	req.AddCookie(&http.Cookie{Name: "ids", Value: "2"})
	values := make(url.Values)
	for _, cookie := range req.Cookies() {
		values.Add(cookie.Name, cookie.Value)
	}
	opts, _ := NewOptions(LocationCookie, StyleForm, true)
	decoder := NewDefaultDecoder()
	result, _ := decoder.Decode(values.Encode(), opts)
	assert.Equal(t, []any{"1", "2"}, result["ids"])
	assert.Equal(t, "abc123", result["session"])
}

func TestSchema_Decode_Path(t *testing.T) {
	opts, _ := NewOptions(LocationPath, StyleSimple)
	decoder := NewDefaultDecoder()
	result, _ := decoder.Decode("1,2,3", opts)
	assert.Equal(t, []any{"1", "2", "3"}, result[""])
}

func TestSchema_Decode_POSTForm(t *testing.T) {
	body := strings.NewReader("name=John&email=john@example.com&ids=1&ids=2&ids=3")
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/api", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, req.ParseForm())

	opts, _ := NewOptions(LocationQuery, StyleForm, true)
	decoder := NewDefaultDecoder()
	result, _ := decoder.Decode(req.PostForm.Encode(), opts)
	assert.Equal(t, "John", result["name"])
	assert.Equal(t, "john@example.com", result["email"])
	assert.Equal(t, []any{"1", "2", "3"}, result["ids"])
}

func TestSchema_Decode_MultipartForm(t *testing.T) {
	var b strings.Builder
	writer := multipart.NewWriter(&b)
	require.NoError(t, writer.WriteField("name", "John"))
	require.NoError(t, writer.WriteField("email", "john@example.com"))
	require.NoError(t, writer.WriteField("ids", "1"))
	require.NoError(t, writer.WriteField("ids", "2"))
	require.NoError(t, writer.Close())

	req, _ := http.NewRequest(http.MethodPost, "http://example.com/api", strings.NewReader(b.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(32<<20))

	opts, _ := NewOptions(LocationQuery, StyleForm, true)
	decoder := NewDefaultDecoder()
	result, _ := decoder.Decode(req.PostForm.Encode(), opts)
	assert.Equal(t, "John", result["name"])
	assert.Equal(t, "john@example.com", result["email"])
	assert.Equal(t, []any{"1", "2"}, result["ids"])
}
