package schema

import (
	"fmt"
	"net/url"
	"strings"
)

// DefaultDecoder handles decoding of parameter strings to maps.
type DefaultDecoder struct{}

// NewDecoder creates a new decoder.
func NewDefaultDecoder() Decoder {
	return &DefaultDecoder{}
}

// Decode parses a parameter value into a nested map based on the given options.
// The location in opts determines how the value is parsed.
func (d *DefaultDecoder) Decode(value string, opts Options) (map[string]any, error) {
	location := opts.Location()
	style := opts.Style()
	explode := opts.getExplode()

	// Validate style is allowed for location (defensive check, Options should already validate)
	if !IsStyleAllowed(location, style) {
		return nil, fmt.Errorf("%w %q for %s parameters", ErrUnsupportedStyle, style, location)
	}

	// Style-based dispatch
	return d.decodeByStyle(value, style, explode)
}

// decodeByStyle dispatches to the appropriate style-specific decoder.
func (d *DefaultDecoder) decodeByStyle(value string, style Style, explode bool) (map[string]any, error) {
	switch style {
	case StyleForm:
		return d.decodeFormStyle(value)
	case StyleSimple:
		return d.decodeSimpleStyle(value)
	case StyleLabel:
		return d.decodeLabelStyle(value, explode)
	case StyleMatrix:
		return d.decodeMatrixStyle(value, explode)
	case StyleSpaceDelimited:
		return d.decodeSpaceDelimited(value)
	case StylePipeDelimited:
		return d.decodePipeDelimited(value)
	case StyleDeepObject:
		return d.decodeDeepObject(value)
	default:
		// Defensive: should never happen if Options validation works correctly
		return nil, fmt.Errorf("%w: %q", ErrInvalidStyle, style)
	}
}

// decodeFormStyle parses form-style data (query string or form data).
func (d *DefaultDecoder) decodeFormStyle(data string) (map[string]any, error) {
	result := make(map[string]any)

	values, err := url.ParseQuery(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	for key, valSlice := range values {
		value := d.processFormValue(valSlice)
		if value == nil {
			continue
		}

		// Handle nested structures (dotted notation: filter.type, filter.color)
		if err := d.setNestedValue(result, key, value); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// decodeSimpleStyle parses simple style (no prefix/suffix, comma-separated).
func (d *DefaultDecoder) decodeSimpleStyle(data string) (map[string]any, error) {
	result := make(map[string]any)

	if data == "" {
		return result, nil
	}

	result[""] = splitToArray(data, ",")

	return result, nil
}

// decodeLabelStyle parses label style (period-prefixed).
func (d *DefaultDecoder) decodeLabelStyle(path string, explode bool) (map[string]any, error) {
	result := make(map[string]any)

	data := strings.TrimPrefix(path, ".")

	if explode {
		// Each value is a separate label: .1.2.3
		result[""] = splitToArray(data, ".")
	} else {
		// Comma-separated: .1,2,3
		return d.decodeSimpleStyle(data)
	}

	return result, nil
}

// decodeMatrixStyle parses matrix style (semicolon-prefixed).
func (d *DefaultDecoder) decodeMatrixStyle(path string, explode bool) (map[string]any, error) {
	result := make(map[string]any)

	data := strings.TrimPrefix(path, ";")

	// Parse key=value pairs
	for pair := range strings.SplitSeq(data, ";") {
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: matrix pair %q", ErrInvalidFormat, pair)
		}

		key := parts[0]
		val := parts[1]

		if explode {
			// Each value is separate: ;ids=1;ids=2;ids=3
			result[key] = appendToArray(result[key], val)
		} else {
			// Comma-separated: ;ids=1,2,3
			result[key] = splitToArray(val, ",")
		}
	}

	return result, nil
}

// decodeSpaceDelimited parses space-delimited query parameters.
func (d *DefaultDecoder) decodeSpaceDelimited(query string) (map[string]any, error) {
	return d.decodeDelimited(query, " ")
}

// decodePipeDelimited parses pipe-delimited query parameters.
func (d *DefaultDecoder) decodePipeDelimited(query string) (map[string]any, error) {
	return d.decodeDelimited(query, "|")
}

// decodeDeepObject parses deepObject style query parameters.
func (d *DefaultDecoder) decodeDeepObject(query string) (map[string]any, error) {
	result := make(map[string]any)

	values, err := url.ParseQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	for key, valSlice := range values {
		// Deep object format: filter[type]=car
		if strings.Contains(key, "[") && strings.Contains(key, "]") {
			if err := d.setDeepObjectValue(result, key, valSlice); err != nil {
				return nil, err
			}
		} else {
			// Regular key-value
			result[key] = valSlice[0]
			if len(valSlice) > 1 {
				result[key] = stringSliceToAny(valSlice)
			}
		}
	}

	return result, nil
}

// processFormValue processes a form value.
func (d *DefaultDecoder) processFormValue(valSlice []string) any {
	if len(valSlice) == 0 {
		return nil
	}

	// Multiple values: treat as array (aligns with standard HTTP behavior)
	// This handles both explode=true (spec: ?ids=1&ids=2) and
	// explode=false edge case (non-spec: ?ids=1&ids=2 when expecting ?ids=1,2)
	if len(valSlice) > 1 {
		return stringSliceToAny(valSlice)
	}

	// Single value: check if comma-separated (explode=false spec case: ?ids=1,2,3)
	val := valSlice[0]
	if val == "" {
		return nil
	}

	if strings.Contains(val, ",") {
		return splitToArray(val, ",")
	}

	return val
}

// decodeDelimited parses delimited query parameters (space or pipe).
func (d *DefaultDecoder) decodeDelimited(query string, sep string) (map[string]any, error) {
	result := make(map[string]any)

	values, err := url.ParseQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	for key, valSlice := range values {
		if len(valSlice) > 0 {
			val := valSlice[len(valSlice)-1]
			if val != "" {
				result[key] = splitToArray(val, sep)
			}
		}
	}

	return result, nil
}

// setNestedValue sets a value in a nested map using dotted notation.
func (d *DefaultDecoder) setNestedValue(m map[string]any, key string, value any) error {
	parts := strings.Split(key, ".")

	return d.setNestedValueByParts(m, parts, value, key)
}

// setDeepObjectValue sets a value in a deep object structure.
func (d *DefaultDecoder) setDeepObjectValue(result map[string]any, key string, valSlice []string) error {
	// Parse all bracket pairs: user[profile][address][street] -> ["user", "profile", "address", "street"]
	parts, err := splitDeepKey(key)
	if err != nil {
		return err
	}

	if len(parts) < 2 {
		return fmt.Errorf("%w: invalid deep object key format: %q", ErrInvalidFormat, key)
	}

	// Determine final value (array if multiple values, single value otherwise)
	finalValue := any(valSlice[0])
	if len(valSlice) > 1 {
		finalValue = stringSliceToAny(valSlice)
	}

	// Navigate/create nested maps and set value (shared logic with setNestedValue)
	return d.setNestedValueByParts(result, parts, finalValue, key)
}

// setNestedValueByParts navigates to/create nested maps and sets a value at the final key.
// This is the common logic shared by setNestedValue and setDeepObjectValue.
func (d *DefaultDecoder) setNestedValueByParts(m map[string]any, parts []string, finalValue any, originalKey string) error {
	if len(parts) < 1 {
		return fmt.Errorf("%w: invalid key format: %q", ErrInvalidFormat, originalKey)
	}

	current := m
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if current[part] == nil {
			current[part] = make(map[string]any)
		}

		next, ok := current[part].(map[string]any)
		if !ok {
			return fmt.Errorf("conflict: key %q is both object and primitive at path %q", part, originalKey)
		}

		current = next
	}

	finalKey := parts[len(parts)-1]

	// Check for conflict: if final key already exists and is a map, but we're setting a non-map value
	if existing, exists := current[finalKey]; exists {
		if _, isMap := existing.(map[string]any); isMap {
			if _, isFinalValueMap := finalValue.(map[string]any); !isFinalValueMap {
				return fmt.Errorf("conflict: key %q is both object and primitive at path %q", finalKey, originalKey)
			}
		}
	}

	current[finalKey] = finalValue

	return nil
}

// splitDeepKey parses bracket notation into parts (e.g., "user[profile][name]" -> ["user", "profile", "name"]).
// Uses byte-level iteration for O(n) performance instead of O(n²) with repeated strings.Index calls.
func splitDeepKey(key string) ([]string, error) {
	var parts []string
	b := []byte(key)
	i := 0

	for i < len(b) {
		// Find next '['
		start := i
		for start < len(b) && b[start] != '[' {
			start++
		}

		if start == len(b) {
			// No more brackets, add remaining text
			if i < len(b) {
				parts = append(parts, string(b[i:]))
			}

			break
		}

		// Add text before '['
		if start > i {
			parts = append(parts, string(b[i:start]))
		}

		// Find closing ']'
		end := start + 1
		for end < len(b) && b[end] != ']' {
			end++
		}

		if end == len(b) {
			return nil, fmt.Errorf("%w: invalid deep object key format: %q", ErrInvalidFormat, key)
		}

		// Add content between brackets
		parts = append(parts, string(b[start+1:end]))
		i = end + 1
	}

	return parts, nil
}

// stringSliceToAny converts []string to []any.
func stringSliceToAny(strs []string) []any {
	result := make([]any, len(strs))
	for i, s := range strs {
		result[i] = s
	}

	return result
}

// splitAndTrim splits a string by separator, trims spaces from each part, and filters empty strings.
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// splitToArray splits a string and returns as []any, or single value if only one part.
func splitToArray(s, sep string) any {
	parts := splitAndTrim(s, sep)

	if len(parts) > 1 {
		return stringSliceToAny(parts)
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return nil
}

// appendToArray appends a value to an array, creating one if needed.
func appendToArray(existing any, val string) []any {
	if arr, ok := existing.([]any); ok {
		return append(arr, val)
	}

	if existing == nil {
		return []any{val}
	}

	// Convert existing value to array
	return []any{existing, val}
}
