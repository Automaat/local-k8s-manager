package backend

import (
	"testing"

	"github.com/automaat/local-k8s-manager/internal/models"
)

func TestMinikubeProvider_Name(t *testing.T) {
	p := NewMinikubeProvider()
	if p.Name() != "minikube" {
		t.Errorf("expected name 'minikube', got '%s'", p.Name())
	}
}

func TestMinikubeProvider_IsInstalled(t *testing.T) {
	p := NewMinikubeProvider()
	result := p.IsInstalled()
	_ = result
}

func TestNewMinikubeProvider(t *testing.T) {
	p := NewMinikubeProvider()
	if p == nil {
		t.Error("NewMinikubeProvider() returned nil")
	}
}

func TestParseMinikubeStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected models.Status
	}{
		{"running status", "Running", models.StatusRunning},
		{"stopped status", "Stopped", models.StatusStopped},
		{"paused status", "Paused", models.StatusUnknown},
		{"unknown status", "SomeOtherStatus", models.StatusUnknown},
		{"empty status", "", models.StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMinikubeStatus(tt.input)
			if result != tt.expected {
				t.Errorf("parseMinikubeStatus(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}
