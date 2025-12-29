package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/automaat/local-k8s-manager/internal/backend"
	"github.com/automaat/local-k8s-manager/internal/models"
)

// keyMsgDetail creates a KeyMsg from a string for testing
func keyMsgDetail(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "?":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	case "q":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	case "d":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	case "s":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	case "x":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

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

func TestRenderDetailView(t *testing.T) {
	cluster := &models.Cluster{
		Name:      "test-cluster",
		Provider:  "k3d",
		Status:    models.StatusRunning,
		Nodes:     3,
		CreatedAt: time.Date(2025, 12, 29, 10, 30, 0, 0, time.UTC),
	}

	tests := []struct {
		name            string
		selectedCluster *models.Cluster
		expected        string
	}{
		{
			name:            "with cluster",
			selectedCluster: cluster,
			expected:        "",
		},
		{
			name:            "no cluster",
			selectedCluster: nil,
			expected:        "No cluster selected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				selectedCluster: tt.selectedCluster,
				width:           80,
			}

			result := m.renderDetailView()
			if result == "" {
				t.Error("expected non-empty result")
			}

			if tt.selectedCluster == nil && result != "No cluster selected" {
				t.Errorf("expected 'No cluster selected', got %s", result)
			}
		})
	}
}

func TestHandleDetailViewKeys(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	cluster := &models.Cluster{
		Name:     "test-cluster",
		Provider: "k3d",
		Status:   models.StatusRunning,
	}

	tests := []struct {
		name         string
		key          string
		expectedView viewState
		shouldQuit   bool
		hasCmd       bool
	}{
		{
			name:         "esc returns to list",
			key:          "esc",
			expectedView: listView,
			shouldQuit:   false,
		},
		{
			name:         "d deletes and returns to list",
			key:          "d",
			expectedView: listView,
			shouldQuit:   false,
			hasCmd:       true,
		},
		{
			name:         "s starts cluster",
			key:          "s",
			expectedView: detailView,
			shouldQuit:   false,
			hasCmd:       true,
		},
		{
			name:         "x stops cluster",
			key:          "x",
			expectedView: detailView,
			shouldQuit:   false,
			hasCmd:       true,
		},
		{
			name:         "? shows help",
			key:          "?",
			expectedView: helpView,
			shouldQuit:   false,
		},
		{
			name:         "q quits",
			key:          "q",
			expectedView: detailView,
			shouldQuit:   true,
		},
		{
			name:         "ctrl+c quits",
			key:          "ctrl+c",
			expectedView: detailView,
			shouldQuit:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				view:            detailView,
				providers:       providers,
				selectedCluster: cluster,
			}

			newModel, cmd := m.handleDetailViewKeys(keyMsgDetail(tt.key))
			m = newModel.(Model)

			if m.view != tt.expectedView {
				t.Errorf("expected view %v, got %v", tt.expectedView, m.view)
			}

			if tt.shouldQuit && cmd == nil {
				t.Error("expected quit command")
			}

			if tt.hasCmd && cmd == nil {
				t.Error("expected command to be returned")
			}

			// Check if delete cleared selected cluster
			if tt.key == "d" && m.selectedCluster != nil {
				t.Error("expected selectedCluster to be nil after delete")
			}

			// Check if esc cleared selected cluster
			if tt.key == "esc" && m.selectedCluster != nil {
				t.Error("expected selectedCluster to be nil after esc")
			}
		})
	}
}

func TestHandleDetailViewKeysWithoutCluster(t *testing.T) {
	m := Model{
		view:            detailView,
		selectedCluster: nil,
	}

	newModel, _ := m.handleDetailViewKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.view != listView {
		t.Errorf("expected listView when cluster is nil, got %v", m.view)
	}
}
