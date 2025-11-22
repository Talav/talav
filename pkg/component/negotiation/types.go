package negotiation

import (
	"fmt"
	"sort"
	"strings"
)

// Header represents a parsed Accept* header value.
// Fields are exported for direct access (idiomatic Go).
type Header struct {
	// Value is the original header value.
	Value string
	// Type is the accept type (e.g., "text/html", "en", "utf-8").
	Type string
	// Quality is the quality value (q-value), defaulting to 1.0.
	Quality float64
	// Parameters contains all parameters except 'q'.
	Parameters map[string]string
	// BasePart is the base part (e.g., "text" from "text/html", "en" from "en-US").
	// Empty for types that don't use base/sub parts.
	BasePart string
	// SubPart is the sub part (e.g., "html" from "text/html", "US" from "en-US").
	// Empty for types that don't use base/sub parts.
	SubPart string

	// NormalizedValue is the normalized value with sorted parameters.
	NormalizedValue string

	// originalIndex is the original position in the header string (for stable sorting).
	originalIndex int
}

// BuildNormalizedValue builds the normalized value string with sorted parameters.
func buildNormalizedValue(typ string, params map[string]string) string {
	if len(params) == 0 {
		return typ
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	return fmt.Sprintf("%s; %s", typ, strings.Join(parts, "; "))
}

// newHeader creates a new Header from a value.
func newHeader(value, typ, basePart, subPart string, quality float64, parameters map[string]string) *Header {
	return &Header{
		Value:           value,
		NormalizedValue: buildNormalizedValue(typ, parameters),
		Type:            typ,
		Quality:         quality,
		Parameters:      parameters,
		BasePart:        basePart,
		SubPart:         subPart,
	}
}
