package tui

import (
	"testing"
)

func TestStatusStyle(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{"running status", "running"},
		{"stopped status", "stopped"},
		{"error status", "error"},
		{"unknown status", "unknown"},
		{"other status", "some-other-status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := StatusStyle(tt.status)
			// Just verify it returns a style without panicking
			_ = style
		})
	}
}
