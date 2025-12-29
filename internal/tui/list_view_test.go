package tui

import (
	"testing"
	"time"
)

func TestFormatAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		contains string
	}{
		{"zero time", time.Time{}, "unknown"},
		{"seconds", now.Add(-30 * time.Second), "s"},
		{"minutes", now.Add(-5 * time.Minute), "m"},
		{"hours", now.Add(-3 * time.Hour), "h"},
		{"days", now.Add(-2 * 24 * time.Hour), "d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAge(tt.time)
			if result == "" {
				t.Error("formatAge returned empty string")
			}
			// For zero time, check exact match
			if tt.time.IsZero() && result != "unknown" {
				t.Errorf("expected 'unknown', got '%s'", result)
			}
		})
	}
}

func TestRenderListHelp(t *testing.T) {
	m := Model{}
	help := m.renderListHelp()
	if help == "" {
		t.Error("renderListHelp returned empty string")
	}
}
