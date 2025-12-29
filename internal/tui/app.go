package tui

import (
	"time"

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

	// View-specific state
	selectedCluster *models.Cluster
	createForm      *createFormModel
}

// Message types
type clustersLoadedMsg struct {
	clusters []models.Cluster
	err      error
}

type tickMsg time.Time

type operationCompleteMsg struct {
	err error
}

// NewModel creates a new TUI model
func NewModel(providers []backend.Provider) Model {
	return Model{
		view:      listView,
		providers: providers,
		clusters:  []models.Cluster{},
		loading:   true,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadClustersCmd(m.providers),
		tickCmd(),
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
		return m, tea.Batch(
			loadClustersCmd(m.providers),
			tickCmd(),
		)

	case operationCompleteMsg:
		m.err = msg.err
		m.loading = false
		return m, loadClustersCmd(m.providers)
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
			return operationCompleteMsg{err: nil}
		}

		err := provider.Delete(clusterName)
		logger.Log("cluster.delete", map[string]interface{}{
			"provider": providerName,
			"name":     clusterName,
			"error":    err,
		})

		return operationCompleteMsg{err: err}
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
			return operationCompleteMsg{err: nil}
		}

		err := provider.Start(clusterName)
		logger.Log("cluster.start", map[string]interface{}{
			"provider": providerName,
			"name":     clusterName,
			"error":    err,
		})

		return operationCompleteMsg{err: err}
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
			return operationCompleteMsg{err: nil}
		}

		err := provider.Stop(clusterName)
		logger.Log("cluster.stop", map[string]interface{}{
			"provider": providerName,
			"name":     clusterName,
			"error":    err,
		})

		return operationCompleteMsg{err: err}
	}
}

// createClusterCmd creates a new cluster
func createClusterCmd(providers []backend.Provider, providerName, name string, workers int) tea.Cmd {
	return func() tea.Msg {
		var provider backend.Provider
		for _, p := range providers {
			if p.Name() == providerName {
				provider = p
				break
			}
		}

		if provider == nil {
			return operationCompleteMsg{err: nil}
		}

		err := provider.Create(name, backend.CreateOptions{Workers: workers})
		logger.Log("cluster.create", map[string]interface{}{
			"provider": providerName,
			"name":     name,
			"workers":  workers,
			"error":    err,
		})

		return operationCompleteMsg{err: err}
	}
}

// renderError renders an error message
func renderError(err error, width int) string {
	if err == nil {
		return ""
	}
	msg := errorMessageStyle.Width(width).Render("Error: " + err.Error())
	return lipgloss.Place(width, 1, lipgloss.Left, lipgloss.Top, msg)
}
