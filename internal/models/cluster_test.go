package models

import (
	"testing"
	"time"
)

func TestCluster(t *testing.T) {
	now := time.Now()

	cluster := Cluster{
		Name:      "test-cluster",
		Provider:  "k3d",
		Status:    StatusRunning,
		Nodes:     3,
		CreatedAt: now,
	}

	if cluster.Name != "test-cluster" {
		t.Errorf("expected Name to be 'test-cluster', got '%s'", cluster.Name)
	}

	if cluster.Provider != "k3d" {
		t.Errorf("expected Provider to be 'k3d', got '%s'", cluster.Provider)
	}

	if cluster.Status != StatusRunning {
		t.Errorf("expected Status to be 'running', got '%s'", cluster.Status)
	}

	if cluster.Nodes != 3 {
		t.Errorf("expected Nodes to be 3, got %d", cluster.Nodes)
	}

	if !cluster.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt to be %v, got %v", now, cluster.CreatedAt)
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{"running status", StatusRunning, "running"},
		{"stopped status", StatusStopped, "stopped"},
		{"unknown status", StatusUnknown, "unknown"},
		{"error status", StatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected status '%s', got '%s'", tt.expected, string(tt.status))
			}
		})
	}
}
