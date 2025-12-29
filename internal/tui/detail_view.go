package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderDetailView renders the cluster detail view
func (m Model) renderDetailView() string {
	if m.selectedCluster == nil {
		return "No cluster selected"
	}

	var b strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf("Cluster Details: %s", m.selectedCluster.Name))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Details
	details := []struct {
		label string
		value string
	}{
		{"Name", m.selectedCluster.Name},
		{"Provider", m.selectedCluster.Provider},
		{"Status", string(m.selectedCluster.Status)},
		{"Nodes", fmt.Sprintf("%d", m.selectedCluster.Nodes)},
		{"Created", formatDetailTime(m.selectedCluster.CreatedAt)},
	}

	for _, d := range details {
		label := headerStyle.Render(fmt.Sprintf("%-12s", d.label+":"))
		value := d.value

		// Apply status styling for status field
		if d.label == "Status" {
			value = StatusStyle(string(m.selectedCluster.Status)).Render(value)
		}

		b.WriteString(fmt.Sprintf("%s %s\n", label, value))
	}

	b.WriteString("\n")

	// Error message
	if m.err != nil {
		b.WriteString(renderError(m.err, m.width))
		b.WriteString("\n")
	}

	// Help
	b.WriteString(m.renderDetailHelp())

	return baseStyle.Render(b.String())
}

// renderDetailHelp renders the help text for detail view
func (m Model) renderDetailHelp() string {
	help := []string{
		"Esc: back",
		"d: delete",
		"s: start",
		"x: stop",
		"q: quit",
	}

	return helpStyle.Render(strings.Join(help, " • "))
}

// handleDetailViewKeys handles keyboard input in detail view
func (m Model) handleDetailViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedCluster == nil {
		m.view = listView
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.view = listView
		m.selectedCluster = nil

	case "d":
		m.loading = true
		cmd := deleteClusterCmd(m.providers, m.selectedCluster.Name, m.selectedCluster.Provider)
		m.view = listView
		m.selectedCluster = nil
		return m, cmd

	case "s":
		m.loading = true
		return m, startClusterCmd(m.providers, m.selectedCluster.Name, m.selectedCluster.Provider)

	case "x":
		m.loading = true
		return m, stopClusterCmd(m.providers, m.selectedCluster.Name, m.selectedCluster.Provider)
	}

	return m, nil
}

// formatDetailTime formats a time for the detail view
func formatDetailTime(t interface {
	IsZero() bool
	Format(string) string
}) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("2006-01-02 15:04:05")
}
