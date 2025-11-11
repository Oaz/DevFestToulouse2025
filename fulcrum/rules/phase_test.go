package gamerules

import (
	"testing"
)

func TestPhase_Accept(t *testing.T) {
	tests := []struct {
		name     string
		phase    Phase
		assets   Set[string]
		expected bool
	}{
		{
			name: "exact match with single accepted set",
			phase: Phase{
				ID:        1,
				Type:      "selection",
				Available: []string{"a1", "a2"},
				Accepted:  []Set[string]{MakeSet([]string{"a1", "a2"})},
			},
			assets:   MakeSet([]string{"a1", "a2"}),
			expected: true,
		},
		{
			name: "superset of accepted set",
			phase: Phase{
				ID:        1,
				Type:      "selection",
				Available: []string{"a1", "a2", "a3"},
				Accepted:  []Set[string]{MakeSet([]string{"a1", "a2"})},
			},
			assets:   MakeSet([]string{"a1", "a2", "a3"}),
			expected: true,
		},
		{
			name: "subset of accepted set (insufficient)",
			phase: Phase{
				ID:        1,
				Type:      "selection",
				Available: []string{"a1", "a2"},
				Accepted:  []Set[string]{MakeSet([]string{"a1", "a2"})},
			},
			assets:   MakeSet([]string{"a1"}),
			expected: false,
		},
		{
			name: "no match with accepted sets",
			phase: Phase{
				ID:        1,
				Type:      "selection",
				Available: []string{"a1", "a2", "a3"},
				Accepted:  []Set[string]{MakeSet([]string{"a1", "a2"}), MakeSet([]string{"a2", "a3"})},
			},
			assets:   MakeSet([]string{"a1", "a3"}),
			expected: false,
		},
		{
			name: "multiple accepted sets - match first",
			phase: Phase{
				ID:        1,
				Type:      "selection",
				Available: []string{"a1", "a2", "a3"},
				Accepted:  []Set[string]{MakeSet([]string{"a1", "a2"}), MakeSet([]string{"a2", "a3"})},
			},
			assets:   MakeSet([]string{"a1", "a2", "a4"}),
			expected: true,
		},
		{
			name: "multiple accepted sets - match second",
			phase: Phase{
				ID:        1,
				Type:      "selection",
				Available: []string{"a1", "a2", "a3"},
				Accepted:  []Set[string]{MakeSet([]string{"a1", "a2"}), MakeSet([]string{"a2", "a3"})},
			},
			assets:   MakeSet([]string{"a2", "a3", "a4"}),
			expected: true,
		},
		{
			name: "empty assets",
			phase: Phase{
				ID:        1,
				Type:      "selection",
				Available: []string{"a1", "a2"},
				Accepted:  []Set[string]{MakeSet([]string{"a1", "a2"})},
			},
			assets:   MakeSet([]string{}),
			expected: false,
		},
		{
			name: "empty accepted sets",
			phase: Phase{
				ID:        1,
				Type:      "selection",
				Available: []string{"a1", "a2"},
				Accepted:  []Set[string]{},
			},
			assets:   MakeSet([]string{"a1", "a2"}),
			expected: false,
		},
		{
			name: "accepted set with empty subset",
			phase: Phase{
				ID:        1,
				Type:      "selection",
				Available: []string{"a1", "a2"},
				Accepted:  []Set[string]{MakeSet([]string{})},
			},
			assets:   MakeSet([]string{"a1", "a2"}),
			expected: true,
		},
		{
			name: "matching empty assets with empty accepted set",
			phase: Phase{
				ID:        1,
				Type:      "selection",
				Available: []string{"a1", "a2"},
				Accepted:  []Set[string]{MakeSet([]string{})},
			},
			assets:   MakeSet([]string{}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.phase.Accept(tt.assets)
			if result != tt.expected {
				t.Errorf("Phase.Accept() = %v, want %v", result, tt.expected)
			}
		})
	}
}
