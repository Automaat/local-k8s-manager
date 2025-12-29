package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorGreen  = lipgloss.Color("#00ff00")
	colorYellow = lipgloss.Color("#ffff00")
	colorRed    = lipgloss.Color("#ff0000")
	colorGray   = lipgloss.Color("#808080")
	colorBlue   = lipgloss.Color("#00aaff")
	colorWhite  = lipgloss.Color("#ffffff")

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Status colors
	runningStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	stoppedStyle = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	unknownStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	// Header style
	headerStyle = lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true).
			Padding(0, 1)

	// Selected row style
	selectedStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(lipgloss.Color("#5f5fff")).
			Bold(true)

	// Help style
	helpStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	// Title style
	titleStyle = lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true).
			PaddingTop(1)

	// Error message style
	errorMessageStyle = lipgloss.NewStyle().
				Foreground(colorRed).
				Bold(true).
				Padding(1, 2)

	// Info message style
	infoStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Padding(0, 2)

	// Border box style for cluster list
	listBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBlue).
			Padding(1, 2)
)

// StatusStyle returns the appropriate style for a cluster status
func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "running":
		return runningStyle
	case "stopped":
		return stoppedStyle
	case "error":
		return errorStyle
	default:
		return unknownStyle
	}
}
