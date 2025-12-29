package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/automaat/local-k8s-manager/internal/backend"
)

// formField represents which field is currently focused
type formField int

const (
	providerField formField = iota
	nameField
	workersField
)

// createFormModel holds the state of the create cluster form
type createFormModel struct {
	providers        []backend.Provider
	selectedProvider int
	name             string
	workers          string
	focusedField     formField
}

// newCreateFormModel creates a new create form model
func newCreateFormModel(providers []backend.Provider) *createFormModel {
	return &createFormModel{
		providers:        providers,
		selectedProvider: 0,
		name:             "",
		workers:          "1",
		focusedField:     providerField,
	}
}

// renderCreateView renders the create cluster view
func (m Model) renderCreateView() string {
	if m.createForm == nil {
		return "Form not initialized"
	}

	var b strings.Builder

	// Title
	title := titleStyle.Render("Create New Cluster")
	b.WriteString(title)
	b.WriteString("\n\n")

	form := m.createForm

	// Provider selection
	b.WriteString(m.renderFormField("Provider", providerField))
	for i, p := range form.providers {
		prefix := "  "
		if i == form.selectedProvider {
			prefix = "▶ "
		}

		providerLine := fmt.Sprintf("%s%s", prefix, p.Name())

		if form.focusedField == providerField && i == form.selectedProvider {
			providerLine = selectedStyle.Render(providerLine)
		}

		b.WriteString(providerLine)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Name field
	b.WriteString(m.renderFormField("Cluster Name", nameField))
	nameValue := form.name
	if nameValue == "" {
		nameValue = "(enter cluster name)"
	}
	if form.focusedField == nameField {
		nameValue = selectedStyle.Render(nameValue)
	}
	b.WriteString(fmt.Sprintf("  %s\n\n", nameValue))

	// Workers field
	b.WriteString(m.renderFormField("Worker Nodes", workersField))
	workersValue := form.workers
	if form.focusedField == workersField {
		workersValue = selectedStyle.Render(workersValue)
	}
	b.WriteString(fmt.Sprintf("  %s\n\n", workersValue))

	// Error message
	if m.err != nil {
		b.WriteString(renderError(m.err, m.width))
		b.WriteString("\n")
	}

	// Help
	b.WriteString(m.renderCreateHelp())

	return baseStyle.Render(b.String())
}

// renderFormField renders a form field label
func (m Model) renderFormField(label string, field formField) string {
	style := headerStyle
	if m.createForm != nil && m.createForm.focusedField == field {
		style = style.Bold(true).Foreground(colorBlue)
	}
	return style.Render(label+":") + "\n"
}

// renderCreateHelp renders the help text for create view
func (m Model) renderCreateHelp() string {
	help := []string{
		"Tab: next field",
		"Shift+Tab: prev field",
		"↑/↓: select provider",
		"Enter: create",
		"Esc: cancel",
	}

	return helpStyle.Render(strings.Join(help, " • "))
}

// handleCreateViewKeys handles keyboard input in create view
func (m Model) handleCreateViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.createForm == nil {
		m.view = listView
		return m, nil
	}

	form := m.createForm

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.view = listView
		m.createForm = nil
		m.err = nil

	case "tab":
		form.focusedField = (form.focusedField + 1) % 3

	case "shift+tab":
		form.focusedField = (form.focusedField + 2) % 3

	case "up":
		if form.focusedField == providerField && form.selectedProvider > 0 {
			form.selectedProvider--
		}

	case "down":
		if form.focusedField == providerField && form.selectedProvider < len(form.providers)-1 {
			form.selectedProvider++
		}

	case "enter":
		// Validate and create cluster
		if form.name == "" {
			m.err = fmt.Errorf("cluster name is required")
			return m, nil
		}

		workers, err := strconv.Atoi(form.workers)
		if err != nil || workers < 0 {
			m.err = fmt.Errorf("workers must be a valid number")
			return m, nil
		}

		provider := form.providers[form.selectedProvider]
		m.loading = true
		m.view = listView
		m.createForm = nil
		m.err = nil

		return m, createClusterCmd(m.providers, provider.Name(), form.name, workers)

	case "backspace":
		if form.focusedField == nameField && len(form.name) > 0 {
			form.name = form.name[:len(form.name)-1]
		} else if form.focusedField == workersField && len(form.workers) > 0 {
			form.workers = form.workers[:len(form.workers)-1]
		}

	default:
		// Handle text input
		if len(msg.String()) == 1 {
			if form.focusedField == nameField {
				form.name += msg.String()
			} else if form.focusedField == workersField && msg.String() >= "0" && msg.String() <= "9" {
				form.workers += msg.String()
			}
		}
	}

	return m, nil
}
