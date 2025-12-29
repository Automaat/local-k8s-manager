package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// renderListView renders the cluster list view
func (m Model) renderListView() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Local Kubernetes Manager")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render table
	if len(m.clusters) == 0 {
		if m.loading {
			msg := infoStyle.Render("Loading clusters...")
			b.WriteString(msg)
		} else {
			msg := infoStyle.Render("No clusters found. Press 'c' to create one.")
			b.WriteString(msg)
		}
	} else {
		b.WriteString(m.renderClusterTable())
	}

	b.WriteString("\n\n")

	// Error message
	if m.err != nil {
		b.WriteString(renderError(m.err, m.width))
		b.WriteString("\n")
	}

	// Help
	b.WriteString(m.renderListHelp())

	return baseStyle.Render(b.String())
}

// renderClusterTable renders the cluster table
func (m Model) renderClusterTable() string {
	// Calculate column widths
	providerWidth := 12
	nameWidth := 20
	statusWidth := 10
	nodesWidth := 7
	ageWidth := 10

	// Header
	header := []string{
		headerStyle.Width(providerWidth).Render("PROVIDER"),
		headerStyle.Width(nameWidth).Render("NAME"),
		headerStyle.Width(statusWidth).Render("STATUS"),
		headerStyle.Width(nodesWidth).Render("NODES"),
		headerStyle.Width(ageWidth).Render("AGE"),
	}

	rows := []string{strings.Join(header, " ")}
	rows = append(rows, strings.Repeat("─", providerWidth+nameWidth+statusWidth+nodesWidth+ageWidth+8))

	// Rows
	for i, cluster := range m.clusters {
		provider := cluster.Provider
		if len(provider) > providerWidth {
			provider = provider[:providerWidth-3] + "..."
		}

		name := cluster.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		status := StatusStyle(string(cluster.Status)).Render(string(cluster.Status))
		nodes := fmt.Sprintf("%d", cluster.Nodes)
		age := formatAge(cluster.CreatedAt)

		row := []string{
			lipgloss.NewStyle().Width(providerWidth).Render(provider),
			lipgloss.NewStyle().Width(nameWidth).Render(name),
			lipgloss.NewStyle().Width(statusWidth).Render(status),
			lipgloss.NewStyle().Width(nodesWidth).Render(nodes),
			lipgloss.NewStyle().Width(ageWidth).Render(age),
		}

		rowStr := strings.Join(row, " ")

		// Highlight selected row
		if i == m.cursor {
			rowWidth := lipgloss.Width(rowStr)
			rowStr = selectedStyle.Width(rowWidth).Render(rowStr)
		}

		rows = append(rows, rowStr)
	}

	return strings.Join(rows, "\n")
}

// renderListHelp renders the help text for list view
func (m Model) renderListHelp() string {
	help := []string{
		"j/k ↑/↓: navigate",
		"Enter: details",
		"c: create",
		"d: delete",
		"s: start",
		"x: stop",
		"r: refresh",
		"?: help",
		"q: quit",
	}

	return helpStyle.Render(strings.Join(help, " • "))
}

// handleListViewKeys handles keyboard input in list view
func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		if m.cursor < len(m.clusters)-1 {
			m.cursor++
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}

	case "c":
		m.view = createView
		m.createForm = newCreateFormModel(m.providers)

	case "d":
		if len(m.clusters) > 0 && m.cursor < len(m.clusters) {
			cluster := m.clusters[m.cursor]
			m.loading = true
			return m, deleteClusterCmd(m.providers, cluster.Name, cluster.Provider)
		}

	case "s":
		if len(m.clusters) > 0 && m.cursor < len(m.clusters) {
			cluster := m.clusters[m.cursor]
			m.loading = true
			return m, startClusterCmd(m.providers, cluster.Name, cluster.Provider)
		}

	case "x":
		if len(m.clusters) > 0 && m.cursor < len(m.clusters) {
			cluster := m.clusters[m.cursor]
			m.loading = true
			return m, stopClusterCmd(m.providers, cluster.Name, cluster.Provider)
		}

	case "r":
		m.loading = true
		return m, loadClustersCmd(m.providers)

	case "enter":
		if len(m.clusters) > 0 && m.cursor < len(m.clusters) {
			m.selectedCluster = &m.clusters[m.cursor]
			m.view = detailView
		}
	}

	return m, nil
}

// formatAge formats a time.Time into a human-readable age string
func formatAge(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	age := time.Since(t)

	if age < time.Minute {
		return fmt.Sprintf("%ds", int(age.Seconds()))
	} else if age < time.Hour {
		return fmt.Sprintf("%dm", int(age.Minutes()))
	} else if age < 24*time.Hour {
		return fmt.Sprintf("%dh", int(age.Hours()))
	}
	return fmt.Sprintf("%dd", int(age.Hours()/24))
}
