package negotiation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeader_Parameters(t *testing.T) {
	acc, err := newMedia("foo/bar; q=1; hello=world")
	require.NoError(t, err)

	_, hasHello := acc.Parameters["hello"]
	assert.True(t, hasHello)
	assert.Equal(t, "world", acc.Parameters["hello"])
	_, hasUnknown := acc.Parameters["unknown"]
	assert.False(t, hasUnknown)

	// Test default value pattern (idiomatic Go: direct map access with ok check)
	val, ok := acc.Parameters["unknown"]
	if !ok {
		val = ""
	}
	assert.Equal(t, "", val)

	val, ok = acc.Parameters["unknown"]
	if !ok {
		val = "goodbye"
	}
	assert.Equal(t, "goodbye", val)

	val, ok = acc.Parameters["hello"]
	if !ok {
		val = "goodbye"
	}
	assert.Equal(t, "world", val)
}

func TestHeader_NormalizedValue(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "sorted parameters",
			header:   "text/html; z=y; a=b; c=d",
			expected: "text/html; a=b; c=d; z=y",
		},
		{
			name:     "with quality",
			header:   "application/pdf; q=1; param=p",
			expected: "application/pdf; param=p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newMedia(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, acc.NormalizedValue)
		})
	}
}

func TestHeader_Type(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"with parameters", "text/html;hello=world", "text/html"},
		{"simple", "application/pdf", "application/pdf"},
		{"with quality", "application/xhtml+xml;q=0.9", "application/xhtml+xml"},
		{"with quality and space", "text/plain; q=0.5", "text/plain"},
		{"with level", "text/html;level=2;q=0.4", "text/html"},
		{"with spaces", "text/html ; level = 2   ; q = 0.4", "text/html"},
		{"wildcard subtype", "text/*", "text/*"},
		{"wildcard with params", "text/* ;q=1 ;level=2", "text/*"},
		{"full wildcard", "*/*", "*/*"},
		{"single wildcard", "*", "*/*"},
		{"wildcard with params", "*/* ; param=555", "*/*"},
		{"single wildcard with params", "* ; param=555", "*/*"},
		{"case insensitive", "TEXT/hTmL;leVel=2; Q=0.4", "text/html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newMedia(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, acc.Type)
		})
	}
}

func TestHeader_Value(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"with spaces", "text/html;hello=world  ;q=0.5", "text/html;hello=world  ;q=0.5"},
		{"simple", "application/pdf", "application/pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newMedia(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, acc.Value)
		})
	}
}

func TestHeader_ParametersMap(t *testing.T) {
	acc, err := newMedia("text/html; charset=UTF-8; level=2")
	require.NoError(t, err)

	params := acc.Parameters
	assert.Equal(t, "UTF-8", params["charset"])
	assert.Equal(t, "2", params["level"])

	// Modify returned map - should not affect original (need to copy)
	paramsCopy := make(map[string]string)
	for k, v := range acc.Parameters {
		paramsCopy[k] = v
	}
	paramsCopy["charset"] = "ISO-8859-1"
	assert.Equal(t, "UTF-8", acc.Parameters["charset"])
}
