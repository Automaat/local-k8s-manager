package backend

import (
	"testing"
)

func TestIsCommandAvailable(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"existing command - ls", "ls", true},
		{"non-existing command", "this-command-does-not-exist-12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCommandAvailable(tt.command)
			if result != tt.expected {
				t.Errorf("IsCommandAvailable(%s) = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestExec(t *testing.T) {
	// Test successful command
	output, err := Exec("echo", "test")
	if err != nil {
		t.Errorf("Exec(echo, test) failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("Exec(echo, test) returned empty output")
	}

	// Test failing command
	_, err = Exec("ls", "/this-path-does-not-exist-12345")
	if err == nil {
		t.Error("Exec(ls, /nonexistent) should have failed")
	}
}
