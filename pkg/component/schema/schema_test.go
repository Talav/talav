package schema

import (
	"testing"
)

func TestDefaultStyle(t *testing.T) {
	tests := []struct {
		location ParameterLocation
		expected Style
	}{
		{LocationQuery, StyleForm},
		{LocationPath, StyleSimple},
		{LocationHeader, StyleSimple},
		{LocationCookie, StyleForm},
	}

	for _, tt := range tests {
		t.Run(string(tt.location), func(t *testing.T) {
			result := DefaultStyle(tt.location)
			if result != tt.expected {
				t.Errorf("DefaultStyle(%q) = %q, want %q", tt.location, result, tt.expected)
			}
		})
	}
}

func TestAllowedStyles(t *testing.T) {
	tests := []struct {
		location ParameterLocation
		expected []Style
	}{
		{
			LocationQuery,
			[]Style{StyleForm, StyleSpaceDelimited, StylePipeDelimited, StyleDeepObject},
		},
		{
			LocationPath,
			[]Style{StyleSimple, StyleLabel, StyleMatrix},
		},
		{
			LocationHeader,
			[]Style{StyleSimple},
		},
		{
			LocationCookie,
			[]Style{StyleForm},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.location), func(t *testing.T) {
			result := AllowedStyles(tt.location)
			if len(result) != len(tt.expected) {
				t.Fatalf("AllowedStyles(%q) returned %d styles, want %d", tt.location, len(result), len(tt.expected))
			}

			expectedMap := make(map[Style]bool)
			for _, s := range tt.expected {
				expectedMap[s] = true
			}

			for _, s := range result {
				if !expectedMap[s] {
					t.Errorf("AllowedStyles(%q) returned unexpected style %q", tt.location, s)
				}
			}
		})
	}
}

func TestIsStyleAllowed(t *testing.T) {
	tests := []struct {
		location ParameterLocation
		style    Style
		allowed  bool
	}{
		// Query parameters
		{LocationQuery, StyleForm, true},
		{LocationQuery, StyleSpaceDelimited, true},
		{LocationQuery, StylePipeDelimited, true},
		{LocationQuery, StyleDeepObject, true},
		{LocationQuery, StyleSimple, false},
		{LocationQuery, StyleMatrix, false},
		{LocationQuery, StyleLabel, false},

		// Path parameters
		{LocationPath, StyleSimple, true},
		{LocationPath, StyleLabel, true},
		{LocationPath, StyleMatrix, true},
		{LocationPath, StyleForm, false},
		{LocationPath, StyleSpaceDelimited, false},

		// Header parameters
		{LocationHeader, StyleSimple, true},
		{LocationHeader, StyleForm, false},
		{LocationHeader, StyleMatrix, false},

		// Cookie parameters
		{LocationCookie, StyleForm, true},
		{LocationCookie, StyleSimple, false},
		{LocationCookie, StyleMatrix, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.location)+"_"+string(tt.style), func(t *testing.T) {
			result := IsStyleAllowed(tt.location, tt.style)
			if result != tt.allowed {
				t.Errorf("IsStyleAllowed(%q, %q) = %v, want %v", tt.location, tt.style, result, tt.allowed)
			}
		})
	}
}
