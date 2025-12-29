package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// renderHelpView renders the help modal
func (m Model) renderHelpView() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Help")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Calculate box dimensions
	contentWidth := min(m.width-8, 80)

	// Help sections
	sections := []struct {
		title string
		items []string
	}{
		{
			title: "List View",
			items: []string{
				"j/k or ↑/↓  Navigate clusters",
				"Enter       View cluster details",
				"c           Create new cluster",
				"d           Delete selected cluster",
				"s           Start selected cluster",
				"x           Stop selected cluster",
				"r           Refresh cluster list",
				"?           Toggle this help",
				"q           Quit application",
			},
		},
		{
			title: "Detail View",
			items: []string{
				"Esc         Back to list",
				"d           Delete cluster",
				"s           Start cluster",
				"x           Stop cluster",
				"?           Toggle this help",
			},
		},
		{
			title: "Create View",
			items: []string{
				"Tab         Next field",
				"Shift+Tab   Previous field",
				"↑/↓         Navigate provider options",
				"Enter       Submit form",
				"Esc         Cancel",
			},
		},
	}

	// Render sections
	for i, section := range sections {
		if i > 0 {
			b.WriteString("\n")
		}

		sectionTitle := headerStyle.Render(section.title)
		b.WriteString(sectionTitle)
		b.WriteString("\n")

		for _, item := range section.items {
			b.WriteString(infoStyle.Render(item))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Esc or ? to close this help"))

	// Create box
	content := baseStyle.Width(contentWidth).Render(b.String())

	// Center the box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Padding(1, 2)

	box := boxStyle.Render(content)

	// Center on screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// handleHelpViewKeys handles keyboard input in help view
func (m Model) handleHelpViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "?", "esc":
		m.view = m.previousView
	}

	return m, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
