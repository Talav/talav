package negotiation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAcceptValue(t *testing.T) {
	tests := []struct {
		name           string
		value          string
		expectedType   string
		expectedParams map[string]string
		expectedQ      float64
		expectErr      bool
	}{
		{
			name:         "empty value",
			value:        "",
			expectedType: "",
			expectedQ:    1.0,
		},
		{
			name:         "simple type",
			value:        "text/html",
			expectedType: "text/html",
			expectedQ:    1.0,
		},
		{
			name:         "with quality",
			value:        "text/html;q=0.8",
			expectedType: "text/html",
			expectedQ:    0.8,
		},
		{
			name:         "with parameters",
			value:        "text/html; charset=UTF-8; level=2",
			expectedType: "text/html",
			expectedParams: map[string]string{
				"charset": "UTF-8",
				"level":   "2",
			},
			expectedQ: 1.0,
		},
		{
			name:         "with quality and parameters",
			value:        "text/html; charset=UTF-8; q=0.7; level=2",
			expectedType: "text/html",
			expectedParams: map[string]string{
				"charset": "UTF-8",
				"level":   "2",
			},
			expectedQ: 0.7,
		},
		{
			name:         "quoted parameter",
			value:        "text/html; foo=\"bar\"",
			expectedType: "text/html",
			expectedParams: map[string]string{
				"foo": "bar",
			},
			expectedQ: 1.0,
		},
		{
			name:      "empty type",
			value:     ";q=0.8",
			expectErr: true,
		},
		{
			name:         "with spaces",
			value:        "text/html ; q = 0.8 ; charset = UTF-8",
			expectedType: "text/html",
			expectedParams: map[string]string{
				"charset": "UTF-8",
			},
			expectedQ: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, params, q, err := parseAcceptValue(tt.value)

			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, typ)
			assert.Equal(t, tt.expectedQ, q)

			if tt.expectedParams != nil {
				for k, v := range tt.expectedParams {
					assert.Equal(t, v, params[k])
				}
			}
		})
	}
}

func TestParseQuality(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expected  float64
		expectErr bool
	}{
		{"valid", "0.8", 0.8, false},
		{"valid 1.0", "1.0", 1.0, false},
		{"valid 0.0", "0.0", 0.0, false},
		{"clamped above 1", "1.5", 1.0, false},
		{"clamped below 0", "-0.5", 0.0, false},
		{"invalid", "abc", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := parseQuality(tt.value)
			if tt.expectErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, q)
		})
	}
}

func TestBuildNormalizedValue(t *testing.T) {
	tests := []struct {
		name     string
		typ      string
		params   map[string]string
		expected string
	}{
		{
			name:     "no parameters",
			typ:      "text/html",
			params:   nil,
			expected: "text/html",
		},
		{
			name:     "sorted parameters",
			typ:      "text/html",
			params:   map[string]string{"z": "y", "a": "b", "c": "d"},
			expected: "text/html; a=b; c=d; z=y",
		},
		{
			name:     "single parameter",
			typ:      "text/html",
			params:   map[string]string{"charset": "UTF-8"},
			expected: "text/html; charset=UTF-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildNormalizedValue(tt.typ, tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		expected  []string
		expectErr bool
	}{
		{
			name:     "simple",
			header:   "text/html",
			expected: []string{"text/html"},
		},
		{
			name:     "multiple",
			header:   "text/html, application/json",
			expected: []string{"text/html", "application/json"},
		},
		{
			name:     "with quality",
			header:   "text/html;q=0.8, application/json;q=0.9",
			expected: []string{"text/html;q=0.8", "application/json;q=0.9"},
		},
		{
			name:     "with quoted strings",
			header:   "text/html; foo=\"bar\", application/json",
			expected: []string{"text/html; foo=\"bar\"", "application/json"},
		},
		{
			name:     "with spaces",
			header:   "text/html , application/json , text/xml",
			expected: []string{"text/html", "application/json", "text/xml"},
		},
		{
			name:     "comma inside quotes",
			header:   "text/html; foo=\"bar,baz\", application/json",
			expected: []string{"text/html; foo=\"bar,baz\"", "application/json"},
		},
		{
			name:     "escaped quotes",
			header:   "text/html; profile=\"\\\"http://example.com/profile\\\"\", application/json",
			expected: []string{"text/html; profile=\"\\\"http://example.com/profile\\\"\"", "application/json"},
		},
		{
			name:      "empty",
			header:    "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseHeader(tt.header)

			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
