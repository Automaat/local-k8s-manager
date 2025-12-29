package tui

import (
	"testing"

	"github.com/automaat/local-k8s-manager/internal/backend"
	"github.com/automaat/local-k8s-manager/internal/models"
)

func TestNewModel(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := NewModel(providers)

	if m.view != listView {
		t.Errorf("expected initial view to be listView, got %v", m.view)
	}

	if !m.loading {
		t.Error("expected initial loading to be true")
	}

	if len(m.providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(m.providers))
	}
}

func TestFindCluster(t *testing.T) {
	m := Model{
		clusters: []models.Cluster{
			{Name: "test1", Provider: "k3d"},
			{Name: "test2", Provider: "kind"},
		},
	}

	// Test finding existing cluster
	cluster := m.findCluster("test1", "k3d")
	if cluster == nil {
		t.Fatal("expected to find cluster, got nil")
	}
	if cluster.Name != "test1" {
		t.Errorf("expected cluster name 'test1', got '%s'", cluster.Name)
	}

	// Test not finding cluster
	cluster = m.findCluster("nonexistent", "k3d")
	if cluster != nil {
		t.Error("expected nil for nonexistent cluster")
	}
}

func TestRenderError(t *testing.T) {
	// Test with nil error
	result := renderError(nil, 80)
	if result != "" {
		t.Error("expected empty string for nil error")
	}

	// Test with actual error
	err := &testError{msg: "test error"}
	result = renderError(err, 80)
	if result == "" {
		t.Error("expected non-empty string for error")
	}
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
