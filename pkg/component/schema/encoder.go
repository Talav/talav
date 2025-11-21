package schema

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// DefaultEncoder handles encoding of maps to parameter strings.
type DefaultEncoder struct{}

// NewEncoder creates a new encoder.
func NewDefaultEncoder() Encoder {
	return &DefaultEncoder{}
}

// Encode encodes a nested map into a parameter value based on the given options.
// The location in opts determines how the value is encoded.
//
//nolint:cyclop // Complex switch logic required for OpenAPI parameter encoding
func (e *DefaultEncoder) Encode(values map[string]any, opts Options) (string, error) {
	location := opts.Location()
	style := opts.Style()
	explode := opts.getExplode()

	switch location {
	case LocationQuery:
		switch style {
		case StyleForm:
			return e.encodeFormStyle(values, explode, true)
		case StyleSpaceDelimited:
			return e.encodeSpaceDelimited(values)
		case StylePipeDelimited:
			return e.encodePipeDelimited(values)
		case StyleDeepObject:
			return e.encodeDeepObject(values)
		case StyleMatrix, StyleLabel, StyleSimple:
			return "", fmt.Errorf("%w %q for query parameters", ErrUnsupportedStyle, style)
		default:
			return "", fmt.Errorf("%w %q for query parameters", ErrUnsupportedStyle, style)
		}

	case LocationCookie:
		switch style {
		case StyleForm:
			return e.encodeFormStyle(values, explode, true)
		case StyleMatrix, StyleLabel, StyleSimple, StyleSpaceDelimited, StylePipeDelimited, StyleDeepObject:
			return "", fmt.Errorf("%w %q for cookie parameters", ErrUnsupportedStyle, style)
		default:
			return "", fmt.Errorf("%w %q for cookie parameters", ErrUnsupportedStyle, style)
		}

	case LocationHeader:
		switch style {
		case StyleSimple:
			return e.encodeSimpleStyle(values, "")
		case StyleForm, StyleSpaceDelimited, StylePipeDelimited, StyleDeepObject, StyleMatrix, StyleLabel:
			return "", fmt.Errorf("%w %q for header parameters", ErrUnsupportedStyle, style)
		default:
			return "", fmt.Errorf("%w %q for header parameters", ErrUnsupportedStyle, style)
		}

	case LocationPath:
		switch style {
		case StyleSimple:
			return e.encodeSimpleStyle(values, "")
		case StyleLabel:
			return e.encodeLabelStyle(values, explode)
		case StyleMatrix:
			return e.encodeMatrixStyle(values, explode)
		case StyleForm, StyleSpaceDelimited, StylePipeDelimited, StyleDeepObject:
			return "", fmt.Errorf("%w %q for path parameters", ErrUnsupportedStyle, style)
		default:
			return "", fmt.Errorf("%w %q for path parameters", ErrUnsupportedStyle, style)
		}

	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedLocation, location)
	}
}

// encodeDelimitedStyle encodes delimited-style data with configurable separators.
func (e *DefaultEncoder) encodeDelimitedStyle(values map[string]any, arraySep, pairSep string, explode bool, addQueryPrefix bool) (string, error) {
	parts := make([]string, 0, len(values))

	for key, value := range values {
		encoded, err := e.encodeValue(key, value, explode, arraySep, pairSep)
		if err != nil {
			return "", err
		}

		if encoded != "" {
			parts = append(parts, encoded)
		}
	}

	result := strings.Join(parts, pairSep)
	if addQueryPrefix && result != "" {
		result = "?" + result
	}

	return result, nil
}

// encodeFormStyle encodes form-style data (query string or form data).
func (e *DefaultEncoder) encodeFormStyle(values map[string]any, explode, queryString bool) (string, error) {
	return e.encodeDelimitedStyle(values, ",", "&", explode, queryString)
}

// encodeSpaceDelimited encodes space-delimited query parameters.
func (e *DefaultEncoder) encodeSpaceDelimited(values map[string]any) (string, error) {
	return e.encodeDelimitedStyle(values, " ", "&", false, true)
}

// encodePipeDelimited encodes pipe-delimited query parameters.
func (e *DefaultEncoder) encodePipeDelimited(values map[string]any) (string, error) {
	return e.encodeDelimitedStyle(values, "|", "&", false, true)
}

// encodeDeepObject encodes deepObject style query parameters.
func (e *DefaultEncoder) encodeDeepObject(values map[string]any) (string, error) {
	parts := make([]string, 0, len(values))

	for key, value := range values {
		encoded := e.encodeDeepObjectValue(key, value)

		if encoded != "" {
			parts = append(parts, encoded)
		}
	}

	result := strings.Join(parts, "&")
	if result != "" {
		result = "?" + result
	}

	return result, nil
}

// encodeDeepObjectValue encodes a value in deepObject format (filter[type]=car).
func (e *DefaultEncoder) encodeDeepObjectValue(key string, value any) string {
	switch v := value.(type) {
	case map[string]any:
		parts := make([]string, 0, len(v))
		for k, val := range v {
			objKey := fmt.Sprintf("%s[%s]", key, k)
			valStr := e.valueToString(val)
			parts = append(parts, url.QueryEscape(objKey)+"="+url.QueryEscape(valStr))
		}

		return strings.Join(parts, "&")
	default:
		valStr := e.valueToString(v)

		return url.QueryEscape(key) + "=" + url.QueryEscape(valStr)
	}
}

// encodeSimpleStyle encodes simple style (no prefix/suffix, comma-separated).
func (e *DefaultEncoder) encodeSimpleStyle(values map[string]any, key string) (string, error) {
	if key != "" {
		// Single key
		if value, ok := values[key]; ok {
			return e.encodeArrayOrValue(value, ","), nil
		}

		return "", nil
	}

	// Empty key or single value
	if len(values) == 1 {
		for _, value := range values {
			return e.encodeArrayOrValue(value, ","), nil
		}
	}

	// Multiple keys - comma-separated
	parts := make([]string, 0, len(values))
	for k, v := range values {
		valStr := e.encodeArrayOrValue(v, ",")
		parts = append(parts, k+"="+valStr)
	}

	return strings.Join(parts, ","), nil
}

// encodeLabelStyle encodes label style (period-prefixed).
func (e *DefaultEncoder) encodeLabelStyle(values map[string]any, explode bool) (string, error) {
	if len(values) == 0 {
		return "", nil
	}

	// Label style typically has empty key
	if value, ok := values[""]; ok {
		if explode {
			// Each value is a separate label: .1.2.3
			return "." + e.encodeArrayOrValue(value, "."), nil
		}

		// Comma-separated: .1,2,3
		return "." + e.encodeArrayOrValue(value, ","), nil
	}

	// Multiple keys - not typical for label style, but handle it
	parts := make([]string, 0, len(values))
	for k, v := range values {
		valStr := e.encodeArrayOrValue(v, ",")
		parts = append(parts, "."+k+"="+valStr)
	}

	return strings.Join(parts, "."), nil
}

// encodeMatrixStyle encodes matrix style (semicolon-prefixed).
func (e *DefaultEncoder) encodeMatrixStyle(values map[string]any, explode bool) (string, error) {
	if len(values) == 0 {
		return "", nil
	}

	// Estimate capacity: at least len(values), potentially more if arrays are exploded
	estimatedCap := len(values)
	for _, value := range values {
		if arr := e.valueToArray(value); len(arr) > 1 {
			estimatedCap += len(arr) - 1
		}
	}
	parts := make([]string, 0, estimatedCap)

	for key, value := range values {
		if explode {
			// Each value is separate: ;ids=1;ids=2;ids=3
			arr := e.valueToArray(value)
			for _, v := range arr {
				valStr := e.valueToString(v)
				parts = append(parts, ";"+url.QueryEscape(key)+"="+url.QueryEscape(valStr))
			}
		} else {
			// Comma-separated: ;ids=1,2,3
			valStr := e.encodeArrayOrValue(value, ",")
			parts = append(parts, ";"+url.QueryEscape(key)+"="+url.QueryEscape(valStr))
		}
	}

	return strings.Join(parts, ""), nil
}

// encodeValue encodes a key-value pair based on explode setting.
//
//nolint:cyclop,unparam // Complex conditional logic required; error return for interface compatibility
func (e *DefaultEncoder) encodeValue(key string, value any, explode bool, arraySep, pairSep string) (string, error) {
	// Handle nested maps (dotted notation)
	if m, ok := value.(map[string]any); ok {
		if explode {
			// Exploded: filter.type=car&filter.color=red
			parts := make([]string, 0, len(m))
			for k, v := range m {
				nestedKey := key + "." + k
				valStr := e.valueToString(v)
				parts = append(parts, url.QueryEscape(nestedKey)+"="+url.QueryEscape(valStr))
			}

			return strings.Join(parts, pairSep), nil
		}

		// Not exploded: filter=type,car,color,red
		parts := make([]string, 0, len(m)*2)
		for k, v := range m {
			parts = append(parts, k)
			parts = append(parts, e.valueToString(v))
		}
		valStr := strings.Join(parts, ",")

		return url.QueryEscape(key) + "=" + url.QueryEscape(valStr), nil
	}

	// Handle arrays
	arr := e.valueToArray(value)
	if len(arr) > 1 || (len(arr) == 1 && e.isArray(value)) {
		if explode {
			// Exploded: ids=1&ids=2&ids=3
			parts := make([]string, 0, len(arr))
			for _, v := range arr {
				valStr := e.valueToString(v)
				parts = append(parts, url.QueryEscape(key)+"="+url.QueryEscape(valStr))
			}

			return strings.Join(parts, pairSep), nil
		}

		// Not exploded: ids=1,2,3
		valStr := e.encodeArrayOrValue(value, arraySep)

		return url.QueryEscape(key) + "=" + url.QueryEscape(valStr), nil
	}

	// Single value
	if len(arr) == 1 {
		valStr := e.valueToString(arr[0])

		return url.QueryEscape(key) + "=" + url.QueryEscape(valStr), nil
	}

	return "", nil
}

// encodeArrayOrValue encodes an array or single value with the given separator.
func (e *DefaultEncoder) encodeArrayOrValue(value any, sep string) string {
	arr := e.valueToArray(value)
	if len(arr) > 1 || (len(arr) == 1 && e.isArray(value)) {
		parts := make([]string, len(arr))
		for i, v := range arr {
			parts[i] = e.valueToString(v)
		}

		return strings.Join(parts, sep)
	}

	if len(arr) == 1 {
		return e.valueToString(arr[0])
	}

	return ""
}

// valueToArray converts a value to an array of any.
func (e *DefaultEncoder) valueToArray(value any) []any {
	switch v := value.(type) {
	case []any:
		return v
	case []string:
		arr := make([]any, len(v))
		for i, s := range v {
			arr[i] = s
		}

		return arr
	case []int:
		arr := make([]any, len(v))
		for i, n := range v {
			arr[i] = n
		}

		return arr
	default:
		return []any{v}
	}
}

// isArray checks if a value is an array/slice type.
func (e *DefaultEncoder) isArray(value any) bool {
	switch value.(type) {
	case []any, []string, []int, []float64, []bool:
		return true
	default:
		return false
	}
}

// valueToString converts any value to a string.
//
//nolint:cyclop // Large switch statement required for all numeric types
func (e *DefaultEncoder) valueToString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}
