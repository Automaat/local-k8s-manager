package tui

import (
	"testing"
	"time"

	"github.com/automaat/local-k8s-manager/internal/backend"
	"github.com/automaat/local-k8s-manager/internal/models"
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

func TestRenderListView(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	tests := []struct {
		name  string
		model Model
	}{
		{
			name: "no clusters loading",
			model: Model{
				loading:   true,
				clusters:  []models.Cluster{},
				width:     80,
				providers: providers,
			},
		},
		{
			name: "no clusters not loading",
			model: Model{
				loading:   false,
				clusters:  []models.Cluster{},
				width:     80,
				providers: providers,
			},
		},
		{
			name: "with clusters",
			model: Model{
				loading: false,
				clusters: []models.Cluster{
					{Name: "test1", Provider: "k3d", Status: models.StatusRunning, Nodes: 2, CreatedAt: time.Now()},
					{Name: "test2", Provider: "kind", Status: models.StatusStopped, Nodes: 1, CreatedAt: time.Now()},
				},
				width:     80,
				providers: providers,
				cursor:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.renderListView()
			if result == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestRenderClusterTable(t *testing.T) {
	m := Model{
		clusters: []models.Cluster{
			{Name: "test-cluster", Provider: "k3d", Status: models.StatusRunning, Nodes: 3, CreatedAt: time.Now()},
			{Name: "very-long-cluster-name-that-exceeds-width", Provider: "kind-provider-long", Status: models.StatusStopped, Nodes: 1, CreatedAt: time.Now()},
		},
		cursor: 0,
		width:  80,
	}

	// Available width for table (width - baseStyle padding - border - box padding)
	tableWidth := m.width - 10
	result := m.renderClusterTable(tableWidth)
	if result == "" {
		t.Error("expected non-empty table")
	}
}

func TestHandleListViewKeys(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	clusters := []models.Cluster{
		{Name: "test1", Provider: "k3d", Status: models.StatusRunning},
		{Name: "test2", Provider: "kind", Status: models.StatusStopped},
		{Name: "test3", Provider: "k3d", Status: models.StatusRunning},
	}

	tests := []struct {
		name           string
		key            string
		initialCursor  int
		expectedCursor int
		expectedView   viewState
		shouldQuit     bool
		hasCmd         bool
	}{
		{
			name:           "j moves cursor down",
			key:            "j",
			initialCursor:  0,
			expectedCursor: 1,
			expectedView:   listView,
			shouldQuit:     false,
		},
		{
			name:           "down moves cursor down",
			key:            "down",
			initialCursor:  0,
			expectedCursor: 1,
			expectedView:   listView,
			shouldQuit:     false,
		},
		{
			name:           "k moves cursor up",
			key:            "k",
			initialCursor:  1,
			expectedCursor: 0,
			expectedView:   listView,
			shouldQuit:     false,
		},
		{
			name:           "up moves cursor up",
			key:            "up",
			initialCursor:  1,
			expectedCursor: 0,
			expectedView:   listView,
			shouldQuit:     false,
		},
		{
			name:           "j at bottom stays at bottom",
			key:            "j",
			initialCursor:  2,
			expectedCursor: 2,
			expectedView:   listView,
			shouldQuit:     false,
		},
		{
			name:           "k at top stays at top",
			key:            "k",
			initialCursor:  0,
			expectedCursor: 0,
			expectedView:   listView,
			shouldQuit:     false,
		},
		{
			name:           "c switches to create view",
			key:            "c",
			initialCursor:  0,
			expectedCursor: 0,
			expectedView:   createView,
			shouldQuit:     false,
		},
		{
			name:           "enter switches to detail view",
			key:            "enter",
			initialCursor:  0,
			expectedCursor: 0,
			expectedView:   detailView,
			shouldQuit:     false,
		},
		{
			name:           "? switches to help view",
			key:            "?",
			initialCursor:  0,
			expectedCursor: 0,
			expectedView:   helpView,
			shouldQuit:     false,
		},
		{
			name:           "q quits",
			key:            "q",
			initialCursor:  0,
			expectedCursor: 0,
			expectedView:   listView,
			shouldQuit:     true,
		},
		{
			name:           "ctrl+c quits",
			key:            "ctrl+c",
			initialCursor:  0,
			expectedCursor: 0,
			expectedView:   listView,
			shouldQuit:     true,
		},
		{
			name:           "d deletes cluster",
			key:            "d",
			initialCursor:  0,
			expectedCursor: 0,
			expectedView:   listView,
			shouldQuit:     false,
			hasCmd:         true,
		},
		{
			name:           "s starts cluster",
			key:            "s",
			initialCursor:  0,
			expectedCursor: 0,
			expectedView:   listView,
			shouldQuit:     false,
			hasCmd:         true,
		},
		{
			name:           "x stops cluster",
			key:            "x",
			initialCursor:  0,
			expectedCursor: 0,
			expectedView:   listView,
			shouldQuit:     false,
			hasCmd:         true,
		},
		{
			name:           "r refreshes clusters",
			key:            "r",
			initialCursor:  0,
			expectedCursor: 0,
			expectedView:   listView,
			shouldQuit:     false,
			hasCmd:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				view:      listView,
				providers: providers,
				clusters:  clusters,
				cursor:    tt.initialCursor,
			}

			newModel, cmd := m.handleListViewKeys(keyMsg(tt.key))
			m = newModel.(Model)

			if m.cursor != tt.expectedCursor {
				t.Errorf("expected cursor %d, got %d", tt.expectedCursor, m.cursor)
			}

			if m.view != tt.expectedView {
				t.Errorf("expected view %v, got %v", tt.expectedView, m.view)
			}

			if tt.shouldQuit && cmd == nil {
				t.Error("expected quit command")
			}

			if tt.hasCmd && cmd == nil {
				t.Error("expected command to be returned")
			}
		})
	}
}

func TestHandleListViewKeysNoClusters(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := Model{
		view:      listView,
		providers: providers,
		clusters:  []models.Cluster{},
	}

	// Test that delete/start/stop/enter do nothing when no clusters
	keys := []string{"d", "s", "x", "enter"}
	for _, key := range keys {
		newModel, _ := m.handleListViewKeys(keyMsg(key))
		m = newModel.(Model)

		if m.view != listView {
			t.Errorf("expected to stay in listView for key %s, got %v", key, m.view)
		}
	}
}
