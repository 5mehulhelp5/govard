package tests

import (
	"testing"

	"govard/internal/engine"
)

func TestCompareNumericDotVersions(t *testing.T) {
	tests := []struct {
		name       string
		left       string
		right      string
		expected   int
		comparable bool
	}{
		{
			name:       "equal simple versions",
			left:       "10.6",
			right:      "10.6.0",
			expected:   0,
			comparable: true,
		},
		{
			name:       "left greater",
			left:       "11.4",
			right:      "10.6",
			expected:   1,
			comparable: true,
		},
		{
			name:       "right greater",
			left:       "2.4.7",
			right:      "2.4.8",
			expected:   -1,
			comparable: true,
		},
		{
			name:       "suffix after numeric segment remains comparable",
			left:       "10.6.20-MariaDB",
			right:      "10.6.19",
			expected:   1,
			comparable: true,
		},
		{
			name:       "invalid leading non numeric",
			left:       "v10.6",
			right:      "10.6",
			expected:   0,
			comparable: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, comparable := engine.CompareNumericDotVersions(tt.left, tt.right)
			if comparable != tt.comparable {
				t.Fatalf("expected comparable=%t, got %t", tt.comparable, comparable)
			}
			if comparable && got != tt.expected {
				t.Fatalf("expected compare result %d, got %d", tt.expected, got)
			}
		})
	}
}
