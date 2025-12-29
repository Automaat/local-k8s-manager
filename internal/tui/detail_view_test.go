package tui

import (
	"testing"
	"time"
)

func TestFormatDetailTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"zero time", time.Time{}, "unknown"},
		{"valid time", time.Date(2025, 12, 29, 10, 30, 0, 0, time.UTC), "2025-12-29 10:30:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDetailTime(tt.time)
			if result != tt.expected {
				t.Errorf("formatDetailTime() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestRenderDetailHelp(t *testing.T) {
	m := Model{}
	help := m.renderDetailHelp()
	if help == "" {
		t.Error("renderDetailHelp returned empty string")
	}
}
