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
	// Title
	title := titleStyle.Render("Local Kubernetes Manager")

	// Render help footer
	help := m.renderListHelp()
	helpHeight := lipgloss.Height(help)

	// Render error message if present
	var errorMsg string
	errorHeight := 0
	if m.err != nil {
		errorMsg = renderError(m.err, m.width)
		errorHeight = lipgloss.Height(errorMsg) + 1 // +1 for spacing
	}

	// Calculate available height for bordered box
	titleHeight := lipgloss.Height(title) + 2 // +2 for spacing
	basePadding := 2                          // baseStyle padding top+bottom
	availableBoxHeight := m.height - titleHeight - helpHeight - errorHeight - basePadding
	if availableBoxHeight < 3 {
		availableBoxHeight = 3
	}

	// Render table content
	var tableContent strings.Builder
	if len(m.clusters) == 0 {
		if m.loading {
			msg := m.spinner.View() + " Loading clusters..."
			tableContent.WriteString(msg)
		} else {
			msg := "No clusters found. Press 'c' to create one."
			tableContent.WriteString(msg)
		}
	} else {
		tableContent.WriteString(m.renderClusterTable())
		if m.loading {
			tableContent.WriteString("\n\n")
			msg := m.spinner.View() + " Refreshing..."
			tableContent.WriteString(msg)
		}
	}

	// Calculate content width and set box to fill remaining space
	// Subtract baseStyle padding (4) + border (2)
	contentWidth := m.width - 6

	// Wrap table in bordered box with calculated dimensions
	// Use Inline(false) and Width to make box fill the space
	borderedTable := listBoxStyle.
		Width(contentWidth).
		Height(availableBoxHeight).
		Inline(false).
		Render(tableContent.String())

	// Build final output
	var final strings.Builder
	final.WriteString(title)
	final.WriteString("\n\n")
	final.WriteString(borderedTable)
	if m.err != nil {
		final.WriteString("\n")
		final.WriteString(errorMsg)
	}
	final.WriteString("\n")
	final.WriteString(help)

	return baseStyle.Render(final.String())
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

	case "?":
		m.previousView = listView
		m.view = helpView
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
