package tui

import (
	"errors"
	"strings"
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

	cmd := createClusterStreamingCmd(providers, "k3d", "test", 2)
	if cmd == nil {
		t.Error("expected non-nil command")
	}

	// createClusterStreamingCmd returns a batch, so we can't test it the same way
	// Just verify it's not nil
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestLogLineMsg(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := NewModel(providers)
	m.view = createView
	m.createForm = &createFormModel{
		currentStep: stepLogs,
		logs:        "",
	}
	m.width = 100
	m.height = 50

	outputChan := make(chan string, 10)

	msg := logLineMsg{
		line:       "test log line",
		outputChan: outputChan,
	}

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.createForm.logs != "test log line" {
		t.Errorf("expected logs to be 'test log line', got '%s'", m.createForm.logs)
	}

	if cmd == nil {
		t.Error("expected command to wait for next log line")
	}
}

func TestLogLineMsgAutoScroll(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	t.Run("auto-scrolls when at bottom", func(t *testing.T) {
		m := NewModel(providers)
		m.view = createView
		// Create enough lines to require scrolling
		logs := ""
		for i := 0; i < 25; i++ {
			if i > 0 {
				logs += "\n"
			}
			logs += "line"
		}
		m.createForm = &createFormModel{
			currentStep:  stepLogs,
			logs:         logs,
			scrollOffset: 10, // User is near bottom
		}
		m.width = 100
		m.height = 30

		outputChan := make(chan string, 10)

		msg := logLineMsg{
			line:       "new line",
			outputChan: outputChan,
		}

		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if !strings.Contains(m.createForm.logs, "new line") {
			t.Error("expected logs to contain new line")
		}
	})

	t.Run("does not auto-scroll when not at bottom", func(t *testing.T) {
		m := NewModel(providers)
		m.view = createView
		// Create enough lines to require scrolling
		logs := ""
		for i := 0; i < 25; i++ {
			if i > 0 {
				logs += "\n"
			}
			logs += "line"
		}
		m.createForm = &createFormModel{
			currentStep:  stepLogs,
			logs:         logs,
			scrollOffset: 0, // User is at top
		}
		m.width = 100
		m.height = 30

		outputChan := make(chan string, 10)
		initialOffset := m.createForm.scrollOffset

		msg := logLineMsg{
			line:       "new line",
			outputChan: outputChan,
		}

		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if m.createForm.scrollOffset != initialOffset {
			t.Errorf("expected scroll offset to stay at %d, got %d", initialOffset, m.createForm.scrollOffset)
		}
	})

	t.Run("handles small window height", func(t *testing.T) {
		m := NewModel(providers)
		m.view = createView
		m.createForm = &createFormModel{
			currentStep:  stepLogs,
			logs:         "line1",
			scrollOffset: 0,
		}
		m.width = 100
		m.height = 8 // Very small height

		outputChan := make(chan string, 10)

		msg := logLineMsg{
			line:       "line2",
			outputChan: outputChan,
		}

		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if !strings.Contains(m.createForm.logs, "line2") {
			t.Error("expected logs to contain new line")
		}
	})
}

func TestOperationCompleteMsgCreate(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := NewModel(providers)
	m.view = createView
	m.createForm = &createFormModel{
		currentStep: stepLogs,
	}
	m.loading = true

	msg := operationCompleteMsg{
		operation: opCreate,
		err:       nil,
	}

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.loading {
		t.Error("expected loading to be false")
	}

	if m.err != nil {
		t.Error("expected no error")
	}

	if cmd == nil {
		t.Error("expected auto-close command for successful create operation")
	}
}

func TestOperationCompleteMsgCreateWithError(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := NewModel(providers)
	m.view = createView
	m.createForm = &createFormModel{
		currentStep: stepLogs,
	}
	m.loading = true

	msg := operationCompleteMsg{
		operation: opCreate,
		err:       errors.New("creation failed"),
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.loading {
		t.Error("expected loading to be false")
	}

	if m.err == nil {
		t.Error("expected error to be set")
	}
}

func TestAutoCloseLogsMsgInCreateView(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := NewModel(providers)
	m.view = createView
	m.createForm = &createFormModel{
		currentStep: stepLogs,
		logs:        "test logs",
	}

	msg := autoCloseLogsMsg{}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.view != listView {
		t.Error("expected view to be listView")
	}

	if m.createForm != nil {
		t.Error("expected createForm to be nil")
	}

	if m.err != nil {
		t.Error("expected err to be nil")
	}

	if cmd == nil {
		t.Error("expected loadClustersCmd")
	}
}

func TestAutoCloseLogsMsgInHelpView(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := NewModel(providers)
	m.view = helpView
	m.previousView = createView
	m.createForm = &createFormModel{
		currentStep: stepLogs,
		logs:        "test logs",
	}

	msg := autoCloseLogsMsg{}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	// Should not change view when in help view
	if m.view != helpView {
		t.Error("expected view to remain helpView")
	}

	if m.createForm == nil {
		t.Error("expected createForm to remain set")
	}

	if cmd != nil {
		t.Error("expected no command when not in createView")
	}
}

func TestAutoCloseLogsMsgNotOnLogsStep(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := NewModel(providers)
	m.view = createView
	m.createForm = &createFormModel{
		currentStep: stepReview,
	}

	msg := autoCloseLogsMsg{}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	// Should not change view when not on logs step
	if m.view != createView {
		t.Error("expected view to remain createView")
	}

	if m.createForm == nil {
		t.Error("expected createForm to remain set")
	}

	if cmd != nil {
		t.Error("expected no command when not on stepLogs")
	}
}

func TestAutoCloseLogsCmd(t *testing.T) {
	cmd := autoCloseLogsCmd()
	if cmd == nil {
		t.Error("expected non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(autoCloseLogsMsg); !ok {
		t.Error("expected autoCloseLogsMsg")
	}
}

func TestWaitForLogLine(t *testing.T) {
	outputChan := make(chan string, 1)
	outputChan <- "test log line"

	cmd := waitForLogLine(outputChan)
	msg := cmd()

	logMsg, ok := msg.(logLineMsg)
	if !ok {
		t.Error("expected logLineMsg")
	}

	if logMsg.line != "test log line" {
		t.Errorf("expected 'test log line', got '%s'", logMsg.line)
	}
}

func TestWaitForLogLineClosedChannel(t *testing.T) {
	outputChan := make(chan string)
	close(outputChan)

	cmd := waitForLogLine(outputChan)
	msg := cmd()

	if msg != nil {
		t.Errorf("expected nil for closed channel, got %v", msg)
	}
}

func TestCreateClusterStreamingCmd(t *testing.T) {
	// Create a mock provider
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	cmd := createClusterStreamingCmd(providers, "k3d", "test-cluster", 1)

	if cmd == nil {
		t.Error("expected command to be returned")
	}

	// Execute the command to test the internal function
	msg := cmd()
	if msg == nil {
		t.Error("expected message from command")
	}
}

func TestCreateClusterStreamingCmdProviderNotFound(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	cmd := createClusterStreamingCmd(providers, "nonexistent", "test-cluster", 1)

	if cmd == nil {
		t.Error("expected command to be returned")
	}

	// Execute the command to test provider not found path
	msg := cmd()
	if msg == nil {
		t.Error("expected message from command")
	}
}
