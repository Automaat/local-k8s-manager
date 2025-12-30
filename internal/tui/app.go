package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/automaat/local-k8s-manager/internal/backend"
	"github.com/automaat/local-k8s-manager/internal/logger"
	"github.com/automaat/local-k8s-manager/internal/models"
)

// viewState represents the current view in the TUI
type viewState int

const (
	listView viewState = iota
	detailView
	createView
	helpView
)

const refreshInterval = 5 * time.Second

// Model is the main TUI model
type Model struct {
	view      viewState
	providers []backend.Provider
	clusters  []models.Cluster
	cursor    int
	width     int
	height    int
	err       error
	loading   bool
	spinner   spinner.Model

	// View-specific state
	selectedCluster *models.Cluster
	createForm      *createFormModel
	previousView    viewState
}

// operationType represents the type of operation in progress
type operationType int

const (
	opCreate operationType = iota
	opDelete
	opStart
	opStop
)

// Message types
type clustersLoadedMsg struct {
	clusters []models.Cluster
	err      error
}

type tickMsg time.Time

type operationCompleteMsg struct {
	operation operationType
	err       error
	output    string
}

type logLineMsg struct {
	line       string
	outputChan <-chan string
}

type autoCloseLogsMsg struct{}

// NewModel creates a new TUI model
func NewModel(providers []backend.Provider) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorBlue)

	return Model{
		view:         listView,
		providers:    providers,
		clusters:     []models.Cluster{},
		loading:      true,
		spinner:      s,
		previousView: listView,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadClustersCmd(m.providers),
		tickCmd(),
		m.spinner.Tick,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case clustersLoadedMsg:
		m.clusters = msg.clusters
		m.err = msg.err
		m.loading = false
		// Update selected cluster if in detail view
		if m.view == detailView && m.selectedCluster != nil {
			m.selectedCluster = m.findCluster(m.selectedCluster.Name, m.selectedCluster.Provider)
		}
		return m, nil

	case tickMsg:
		// Only auto-refresh clusters when in list view to avoid unnecessary API calls
		if m.view == listView {
			return m, tea.Batch(
				loadClustersCmd(m.providers),
				tickCmd(),
			)
		}
		// In other views, keep the ticker running but skip the refresh
		return m, tickCmd()

	case operationCompleteMsg:
		m.loading = false

		// Handle create operation specially
		if msg.operation == opCreate {
			// We're already on logs screen showing streamed output
			// Just stop loading indicator and set error if any
			m.err = msg.err

			// If successful, auto-close logs after 2 seconds
			if msg.err == nil {
				return m, autoCloseLogsCmd()
			}
			return m, nil
		} else {
			// For other operations, just set error
			m.err = msg.err
		}

		return m, loadClustersCmd(m.providers)

	case autoCloseLogsMsg:
		// Close logs and return to list view
		if m.createForm != nil && m.createForm.currentStep == stepLogs {
			m.view = listView
			m.createForm = nil
			m.err = nil
			return m, loadClustersCmd(m.providers)
		}
		return m, nil

	case logLineMsg:
		// Append log line to the current form logs
		if m.createForm != nil && m.createForm.currentStep == stepLogs {
			// Check if we're at the bottom before adding new line
			logLines := strings.Split(m.createForm.logs, "\n")
			oldLineCount := len(logLines)

			// Calculate if user was at bottom (within 2 lines)
			boxHeight := 20
			if m.height-10 < boxHeight {
				boxHeight = m.height - 10
				if boxHeight < 5 {
					boxHeight = 5
				}
			}
			visibleLines := boxHeight - 2
			if visibleLines < 1 {
				visibleLines = 1
			}
			maxOffset := oldLineCount - visibleLines
			if maxOffset < 0 {
				maxOffset = 0
			}
			wasAtBottom := (m.createForm.scrollOffset >= maxOffset-2)

			// Add new line
			if m.createForm.logs != "" {
				m.createForm.logs += "\n"
			}
			m.createForm.logs += msg.line

			// Auto-scroll to bottom if user was already at/near bottom
			if wasAtBottom {
				newLogLines := strings.Split(m.createForm.logs, "\n")
				newMaxOffset := len(newLogLines) - visibleLines
				if newMaxOffset < 0 {
					newMaxOffset = 0
				}
				m.createForm.scrollOffset = newMaxOffset
			}
		}
		// Continue waiting for more log lines
		return m, waitForLogLine(msg.outputChan)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.view {
	case listView:
		return m.renderListView()
	case detailView:
		return m.renderDetailView()
	case createView:
		return m.renderCreateView()
	case helpView:
		return m.renderHelpView()
	}

	return ""
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.view {
	case listView:
		return m.handleListViewKeys(msg)
	case detailView:
		return m.handleDetailViewKeys(msg)
	case createView:
		return m.handleCreateViewKeys(msg)
	case helpView:
		return m.handleHelpViewKeys(msg)
	}
	return m, nil
}

// findCluster finds a cluster by name and provider
func (m Model) findCluster(name, provider string) *models.Cluster {
	for i := range m.clusters {
		if m.clusters[i].Name == name && m.clusters[i].Provider == provider {
			return &m.clusters[i]
		}
	}
	return nil
}

// loadClustersCmd loads clusters from all providers
func loadClustersCmd(providers []backend.Provider) tea.Cmd {
	return func() tea.Msg {
		var allClusters []models.Cluster

		for _, p := range providers {
			clusters, err := p.List()
			if err != nil {
				logger.LogError("cluster.list", err, map[string]interface{}{
					"provider": p.Name(),
				})
				continue
			}
			allClusters = append(allClusters, clusters...)
		}

		return clustersLoadedMsg{
			clusters: allClusters,
		}
	}
}

// tickCmd creates a ticker for auto-refresh
func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// deleteClusterCmd deletes a cluster
func deleteClusterCmd(providers []backend.Provider, clusterName, providerName string) tea.Cmd {
	return func() tea.Msg {
		var provider backend.Provider
		for _, p := range providers {
			if p.Name() == providerName {
				provider = p
				break
			}
		}

		if provider == nil {
			return operationCompleteMsg{operation: opDelete, err: nil}
		}

		err := provider.Delete(clusterName)
		logger.Log("cluster.delete", map[string]interface{}{
			"provider": providerName,
			"name":     clusterName,
			"error":    err,
		})

		return operationCompleteMsg{operation: opDelete, err: err}
	}
}

// startClusterCmd starts a cluster
func startClusterCmd(providers []backend.Provider, clusterName, providerName string) tea.Cmd {
	return func() tea.Msg {
		var provider backend.Provider
		for _, p := range providers {
			if p.Name() == providerName {
				provider = p
				break
			}
		}

		if provider == nil {
			return operationCompleteMsg{operation: opStart, err: nil}
		}

		err := provider.Start(clusterName)
		logger.Log("cluster.start", map[string]interface{}{
			"provider": providerName,
			"name":     clusterName,
			"error":    err,
		})

		return operationCompleteMsg{operation: opStart, err: err}
	}
}

// stopClusterCmd stops a cluster
func stopClusterCmd(providers []backend.Provider, clusterName, providerName string) tea.Cmd {
	return func() tea.Msg {
		var provider backend.Provider
		for _, p := range providers {
			if p.Name() == providerName {
				provider = p
				break
			}
		}

		if provider == nil {
			return operationCompleteMsg{operation: opStop, err: nil}
		}

		err := provider.Stop(clusterName)
		logger.Log("cluster.stop", map[string]interface{}{
			"provider": providerName,
			"name":     clusterName,
			"error":    err,
		})

		return operationCompleteMsg{operation: opStop, err: err}
	}
}

// createClusterStreamingCmd creates a cluster and streams output via messages
func createClusterStreamingCmd(providers []backend.Provider, providerName, name string, workers int) tea.Cmd {
	// Create output channel for streaming
	outputChan := make(chan string, 100)

	// Return batch of commands: one to read from channel, one to execute
	return tea.Batch(
		// Command to wait for log lines
		waitForLogLine(outputChan),
		// Command to execute cluster creation
		func() tea.Msg {
			var provider backend.Provider
			for _, p := range providers {
				if p.Name() == providerName {
					provider = p
					break
				}
			}

			if provider == nil {
				close(outputChan)
				return operationCompleteMsg{operation: opCreate, err: nil, output: ""}
			}

			// Execute with streaming
			output, err := provider.Create(name, backend.CreateOptions{Workers: workers}, outputChan)
			close(outputChan)

			logger.Log("cluster.create", map[string]interface{}{
				"provider": providerName,
				"name":     name,
				"workers":  workers,
				"error":    err,
			})

			return operationCompleteMsg{operation: opCreate, err: err, output: output}
		},
	)
}

// waitForLogLine waits for a log line from the channel and returns it as a message
func waitForLogLine(outputChan <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-outputChan
		if !ok {
			// Channel closed, no more lines
			return nil
		}
		return logLineMsg{line: line, outputChan: outputChan}
	}
}

// autoCloseLogsCmd creates a command that waits 2 seconds then closes logs
func autoCloseLogsCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return autoCloseLogsMsg{}
	})
}

// renderError renders an error message
func renderError(err error, width int) string {
	if err == nil {
		return ""
	}
	msg := errorMessageStyle.Width(width).Render("Error: " + err.Error())
	return lipgloss.Place(width, 1, lipgloss.Left, lipgloss.Top, msg)
}
