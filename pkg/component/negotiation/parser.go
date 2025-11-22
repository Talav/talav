package negotiation

import (
	"strconv"
	"strings"
)

// parseAcceptValue parses an accept header value into type, parameters, and quality.
// Returns the normalized type (lowercase), parameters map (excluding 'q'), and quality value.
func parseAcceptValue(value string) (typ string, params map[string]string, quality float64, err error) {
	if value == "" {
		return "", nil, 1.0, nil
	}

	parts := strings.Split(value, ";")
	typ = strings.TrimSpace(parts[0])
	if typ == "" {
		return "", nil, 0, &InvalidHeaderError{Header: value}
	}

	params = make(map[string]string)
	quality = 1.0

	for i := 1; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			continue
		}

		if !strings.Contains(part, "=") {
			continue
		}

		key, val, _ := strings.Cut(part, "=")
		key = strings.ToLower(strings.TrimSpace(key))
		val = strings.Trim(strings.TrimSpace(val), `"`)

		if key == "q" {
			quality, err = parseQuality(val)
			if err != nil {
				return "", nil, 0, err
			}
		} else {
			params[key] = val
		}
	}

	typ = strings.ToLower(strings.TrimSpace(typ))

	return typ, params, quality, nil
}

// parseQuality parses and validates a quality value string.
// Returns a value between 0.0 and 1.0.
func parseQuality(s string) (float64, error) {
	q, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if q < 0 {
		q = 0
	} else if q > 1 {
		q = 1
	}

	return q, nil
}

// parseHeader parses an Accept* header string into individual accept parts.
// Handles quoted strings, escaped quotes, and commas correctly using a state machine.
func parseHeader(header string) ([]string, error) {
	var parts []string
	start := 0
	inQuotes := false
	escaped := false

	for i := 0; i < len(header); i++ {
		c := header[i]
		var shouldContinue bool
		escaped, inQuotes, shouldContinue = processChar(c, escaped, inQuotes)
		if shouldContinue {
			continue
		}

		if c == ',' && !inQuotes {
			if part := extractPart(header[start:i]); part != "" {
				parts = append(parts, part)
			}
			start = i + 1
		}
	}

	parts = appendFinalPart(parts, header, start)

	if len(parts) == 0 {
		return nil, &InvalidHeaderError{Header: header}
	}

	return parts, nil
}

// processChar processes a single character in the state machine.
// Returns the new escaped state, new inQuotes state, and whether to continue the loop.
func processChar(c byte, escaped, inQuotes bool) (newEscaped, newInQuotes, shouldContinue bool) {
	if escaped {
		return false, inQuotes, true
	}

	if c == '\\' {
		return true, inQuotes, true
	}

	if c == '"' {
		return false, !inQuotes, true
	}

	return false, inQuotes, false
}

// extractPart extracts and trims a part from the header string.
func extractPart(s string) string {
	return strings.TrimSpace(s)
}

// appendFinalPart appends the final part if it exists and is non-empty.
func appendFinalPart(parts []string, header string, start int) []string {
	if start >= len(header) {
		return parts
	}

	if part := extractPart(header[start:]); part != "" {
		parts = append(parts, part)
	}

	return parts
}
