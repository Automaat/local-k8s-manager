package tui

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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

func TestInit(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := NewModel(providers)
	cmd := m.Init()

	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestUpdate(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	tests := []struct {
		name      string
		msg       tea.Msg
		checkFunc func(*testing.T, Model)
	}{
		{
			name: "window size message",
			msg:  tea.WindowSizeMsg{Width: 100, Height: 50},
			checkFunc: func(t *testing.T, m Model) {
				if m.width != 100 {
					t.Errorf("expected width 100, got %d", m.width)
				}
				if m.height != 50 {
					t.Errorf("expected height 50, got %d", m.height)
				}
			},
		},
		{
			name: "clusters loaded message",
			msg: clustersLoadedMsg{
				clusters: []models.Cluster{
					{Name: "test", Provider: "k3d"},
				},
				err: nil,
			},
			checkFunc: func(t *testing.T, m Model) {
				if len(m.clusters) != 1 {
					t.Errorf("expected 1 cluster, got %d", len(m.clusters))
				}
				if m.loading {
					t.Error("expected loading to be false")
				}
			},
		},
		{
			name: "clusters loaded with error",
			msg: clustersLoadedMsg{
				clusters: nil,
				err:      errors.New("test error"),
			},
			checkFunc: func(t *testing.T, m Model) {
				if m.err == nil {
					t.Error("expected error to be set")
				}
				if m.loading {
					t.Error("expected loading to be false")
				}
			},
		},
		{
			name: "operation complete message",
			msg: operationCompleteMsg{
				err: nil,
			},
			checkFunc: func(t *testing.T, m Model) {
				if m.loading {
					t.Error("expected loading to be false")
				}
			},
		},
		{
			name: "operation complete with error",
			msg: operationCompleteMsg{
				err: errors.New("operation failed"),
			},
			checkFunc: func(t *testing.T, m Model) {
				if m.err == nil {
					t.Error("expected error to be set")
				}
			},
		},
		{
			name: "tick message in list view",
			msg:  tickMsg(time.Now()),
			checkFunc: func(t *testing.T, m Model) {
				if m.view != listView {
					t.Error("view should remain listView")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(providers)
			m.loading = false
			newModel, _ := m.Update(tt.msg)
			m = newModel.(Model)

			tt.checkFunc(t, m)
		})
	}
}

func TestUpdateSelectedClusterRefresh(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	cluster := &models.Cluster{
		Name:     "test",
		Provider: "k3d",
		Status:   models.StatusRunning,
	}

	m := NewModel(providers)
	m.view = detailView
	m.selectedCluster = cluster

	// Simulate clusters loaded with updated cluster
	msg := clustersLoadedMsg{
		clusters: []models.Cluster{
			{Name: "test", Provider: "k3d", Status: models.StatusStopped},
		},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.selectedCluster.Status != models.StatusStopped {
		t.Error("expected selectedCluster to be updated")
	}
}

func TestView(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	tests := []struct {
		name  string
		view  viewState
		width int
	}{
		{
			name:  "list view",
			view:  listView,
			width: 80,
		},
		{
			name:  "detail view",
			view:  detailView,
			width: 80,
		},
		{
			name:  "create view",
			view:  createView,
			width: 80,
		},
		{
			name:  "help view",
			view:  helpView,
			width: 80,
		},
		{
			name:  "zero width",
			view:  listView,
			width: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(providers)
			m.view = tt.view
			m.width = tt.width
			m.height = 24

			if tt.view == detailView {
				m.selectedCluster = &models.Cluster{
					Name:     "test",
					Provider: "k3d",
					Status:   models.StatusRunning,
				}
			}

			if tt.view == createView {
				m.createForm = newCreateFormModel(providers)
			}

			result := m.View()

			if tt.width == 0 && result != "" {
				t.Error("expected empty string for zero width")
			}

			if tt.width > 0 && result == "" && tt.view != listView {
				t.Error("expected non-empty view")
			}
		})
	}
}

func TestHandleKeyPress(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	views := []viewState{listView, detailView, createView, helpView}

	for _, view := range views {
		t.Run(string(rune(view)), func(t *testing.T) {
			m := NewModel(providers)
			m.view = view

			if view == detailView {
				m.selectedCluster = &models.Cluster{
					Name:     "test",
					Provider: "k3d",
				}
			}

			if view == createView {
				m.createForm = newCreateFormModel(providers)
			}

			msg := tea.KeyMsg{Type: tea.KeyEsc}
			newModel, _ := m.handleKeyPress(msg)

			if newModel == nil {
				t.Error("expected non-nil model")
			}
		})
	}
}

func TestTickCmd(t *testing.T) {
	cmd := tickCmd()
	if cmd == nil {
		t.Error("expected non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(tickMsg); !ok {
		t.Error("expected tickMsg")
	}
}

func TestLoadClustersCmd(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	cmd := loadClustersCmd(providers)
	if cmd == nil {
		t.Error("expected non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(clustersLoadedMsg); !ok {
		t.Error("expected clustersLoadedMsg")
	}
}

func TestDeleteClusterCmd(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	cmd := deleteClusterCmd(providers, "test", "k3d")
	if cmd == nil {
		t.Error("expected non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(operationCompleteMsg); !ok {
		t.Error("expected operationCompleteMsg")
	}
}

func TestDeleteClusterCmdProviderNotFound(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	cmd := deleteClusterCmd(providers, "test", "nonexistent")
	msg := cmd()

	opMsg, ok := msg.(operationCompleteMsg)
	if !ok {
		t.Fatal("expected operationCompleteMsg")
	}

	if opMsg.err != nil {
		t.Error("expected no error when provider not found")
	}
}

func TestStartClusterCmd(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	cmd := startClusterCmd(providers, "test", "k3d")
	if cmd == nil {
		t.Error("expected non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(operationCompleteMsg); !ok {
		t.Error("expected operationCompleteMsg")
	}
}

func TestStopClusterCmd(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	cmd := stopClusterCmd(providers, "test", "k3d")
	if cmd == nil {
		t.Error("expected non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(operationCompleteMsg); !ok {
		t.Error("expected operationCompleteMsg")
	}
}

func TestCreateClusterCmd(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	cmd := createClusterCmd(providers, "k3d", "test", 2)
	if cmd == nil {
		t.Error("expected non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(operationCompleteMsg); !ok {
		t.Error("expected operationCompleteMsg")
	}
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
