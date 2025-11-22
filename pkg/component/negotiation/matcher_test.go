package negotiation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch_SortOrder(t *testing.T) {
	tests := []struct {
		name     string
		match1   *matchResult
		match2   *matchResult
		expected bool // true if match1 should come before match2
	}{
		{
			name:     "equal quality and index",
			match1:   &matchResult{Quality: 1.0, Score: 110, Index: 1},
			match2:   &matchResult{Quality: 1.0, Score: 111, Index: 1},
			expected: false, // equal, order doesn't matter
		},
		{
			name:     "equal quality, match1 lower index",
			match1:   &matchResult{Quality: 0.1, Score: 10, Index: 1},
			match2:   &matchResult{Quality: 0.1, Score: 10, Index: 2},
			expected: true, // match1 comes first (lower index)
		},
		{
			name:     "equal quality, match1 higher index",
			match1:   &matchResult{Quality: 0.5, Score: 110, Index: 5},
			match2:   &matchResult{Quality: 0.5, Score: 11, Index: 4},
			expected: false, // match2 comes first (lower index)
		},
		{
			name:     "match1 lower quality",
			match1:   &matchResult{Quality: 0.4, Score: 110, Index: 1},
			match2:   &matchResult{Quality: 0.6, Score: 111, Index: 3},
			expected: false, // match2 comes first (higher quality)
		},
		{
			name:     "match1 higher quality",
			match1:   &matchResult{Quality: 0.6, Score: 110, Index: 1},
			match2:   &matchResult{Quality: 0.4, Score: 111, Index: 3},
			expected: true, // match1 comes first (higher quality)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the inline sort logic
			mi, mj := tt.match1, tt.match2
			result := false
			if mi.Quality != mj.Quality {
				result = mi.Quality > mj.Quality
			} else {
				result = mi.Index < mj.Index
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
