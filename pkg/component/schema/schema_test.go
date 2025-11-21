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

func TestValidateStyle(t *testing.T) {
	tests := []struct {
		name        string
		location    ParameterLocation
		style       Style
		expected    Style
		expectError bool
	}{
		// Empty style should return default
		{"query empty style", LocationQuery, "", StyleForm, false},
		{"path empty style", LocationPath, "", StyleSimple, false},
		{"header empty style", LocationHeader, "", StyleSimple, false},
		{"cookie empty style", LocationCookie, "", StyleForm, false},

		// Valid styles
		{"query form", LocationQuery, StyleForm, StyleForm, false},
		{"query spaceDelimited", LocationQuery, StyleSpaceDelimited, StyleSpaceDelimited, false},
		{"path simple", LocationPath, StyleSimple, StyleSimple, false},
		{"path matrix", LocationPath, StyleMatrix, StyleMatrix, false},
		{"header simple", LocationHeader, StyleSimple, StyleSimple, false},
		{"cookie form", LocationCookie, StyleForm, StyleForm, false},

		// Invalid styles
		{"query simple (invalid)", LocationQuery, StyleSimple, "", true},
		{"path form (invalid)", LocationPath, StyleForm, "", true},
		{"header form (invalid)", LocationHeader, StyleForm, "", true},
		{"cookie simple (invalid)", LocationCookie, StyleSimple, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateStyle(tt.location, tt.style)
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateStyle(%q, %q) expected error, got nil", tt.location, tt.style)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateStyle(%q, %q) unexpected error: %v", tt.location, tt.style, err)
				}
				if result != tt.expected {
					t.Errorf("ValidateStyle(%q, %q) = %q, want %q", tt.location, tt.style, result, tt.expected)
				}
			}
		})
	}
}
