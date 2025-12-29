package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/automaat/local-k8s-manager/internal/backend"
)

// createStep represents the current step in the multi-step wizard
type createStep int

const (
	stepProvider createStep = iota + 1
	stepName
	stepWorkers
	stepReview
)

// createFormModel holds the state of the create cluster form
type createFormModel struct {
	providers        []backend.Provider
	selectedProvider int
	name             string
	workers          string
	currentStep      createStep
}

// newCreateFormModel creates a new create form model
func newCreateFormModel(providers []backend.Provider) *createFormModel {
	return &createFormModel{
		providers:        providers,
		selectedProvider: 0,
		name:             "",
		workers:          "1",
		currentStep:      stepProvider,
	}
}

// renderCreateView renders the create cluster view
func (m Model) renderCreateView() string {
	if m.createForm == nil {
		return "Form not initialized"
	}

	switch m.createForm.currentStep {
	case stepProvider:
		return m.renderStepProvider()
	case stepName:
		return m.renderStepName()
	case stepWorkers:
		return m.renderStepWorkers()
	case stepReview:
		return m.renderStepReview()
	default:
		return "Unknown step"
	}
}

// renderStepProvider renders step 1: provider selection
func (m Model) renderStepProvider() string {
	var b strings.Builder
	form := m.createForm

	// Title with step indicator
	title := titleStyle.Render(fmt.Sprintf("Create New Cluster (Step 1/4)"))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Instruction
	b.WriteString(headerStyle.Bold(true).Foreground(colorBlue).Render("Select Provider:"))
	b.WriteString("\n\n")

	// Provider list
	for i, p := range form.providers {
		prefix := "  "
		if i == form.selectedProvider {
			prefix = "▶ "
			providerLine := selectedStyle.Render(fmt.Sprintf("%s%s", prefix, p.Name()))
			b.WriteString(providerLine)
		} else {
			b.WriteString(fmt.Sprintf("%s%s", prefix, p.Name()))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Error message
	if m.err != nil {
		b.WriteString(renderError(m.err, m.width))
		b.WriteString("\n")
	}

	// Help
	help := helpStyle.Render("↑/↓: select • Enter: next • Esc: cancel")
	b.WriteString(help)

	return baseStyle.Render(b.String())
}

// renderStepName renders step 2: cluster name input
func (m Model) renderStepName() string {
	var b strings.Builder
	form := m.createForm

	// Title with step indicator
	title := titleStyle.Render(fmt.Sprintf("Create New Cluster (Step 2/4)"))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Instruction
	b.WriteString(headerStyle.Bold(true).Foreground(colorBlue).Render("Cluster Name:"))
	b.WriteString("\n\n")

	// Name input
	nameValue := form.name
	if nameValue == "" {
		nameValue = "(enter cluster name)"
	}
	b.WriteString(selectedStyle.Render(fmt.Sprintf("  %s", nameValue)))
	b.WriteString("\n\n")

	// Error message
	if m.err != nil {
		b.WriteString(renderError(m.err, m.width))
		b.WriteString("\n")
	}

	// Help
	help := helpStyle.Render("Type name • Enter: next • Esc: back")
	b.WriteString(help)

	return baseStyle.Render(b.String())
}

// renderStepWorkers renders step 3: worker nodes input
func (m Model) renderStepWorkers() string {
	var b strings.Builder
	form := m.createForm

	// Title with step indicator
	title := titleStyle.Render(fmt.Sprintf("Create New Cluster (Step 3/4)"))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Instruction
	b.WriteString(headerStyle.Bold(true).Foreground(colorBlue).Render("Worker Nodes:"))
	b.WriteString("\n\n")

	// Workers input
	b.WriteString(selectedStyle.Render(fmt.Sprintf("  %s", form.workers)))
	b.WriteString("\n\n")

	// Error message
	if m.err != nil {
		b.WriteString(renderError(m.err, m.width))
		b.WriteString("\n")
	}

	// Help
	help := helpStyle.Render("Type number • Enter: next • Esc: back")
	b.WriteString(help)

	return baseStyle.Render(b.String())
}

// renderStepReview renders step 4: review and confirm
func (m Model) renderStepReview() string {
	var b strings.Builder
	form := m.createForm

	// Title with step indicator
	title := titleStyle.Render(fmt.Sprintf("Create New Cluster (Step 4/4)"))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Review header
	b.WriteString(headerStyle.Bold(true).Foreground(colorBlue).Render("Review Configuration:"))
	b.WriteString("\n\n")

	// Provider
	b.WriteString(headerStyle.Render("Provider:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n\n", form.providers[form.selectedProvider].Name()))

	// Cluster Name
	b.WriteString(headerStyle.Render("Cluster Name:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n\n", form.name))

	// Worker Nodes
	b.WriteString(headerStyle.Render("Worker Nodes:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n\n", form.workers))

	// Error message
	if m.err != nil {
		b.WriteString(renderError(m.err, m.width))
		b.WriteString("\n")
	}

	// Help
	help := helpStyle.Render("Enter: create cluster • Esc: back")
	b.WriteString(help)

	return baseStyle.Render(b.String())
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
		// Go back to previous step or cancel
		if form.currentStep == stepProvider {
			// Cancel entire form
			m.view = listView
			m.createForm = nil
			m.err = nil
		} else {
			// Go back to previous step
			form.currentStep--
			m.err = nil
		}

	case "?":
		m.previousView = createView
		m.view = helpView

	case "enter":
		return m.handleEnterKey()

	case "up":
		if form.currentStep == stepProvider && form.selectedProvider > 0 {
			form.selectedProvider--
		}

	case "down":
		if form.currentStep == stepProvider && form.selectedProvider < len(form.providers)-1 {
			form.selectedProvider++
		}

	case "backspace":
		if form.currentStep == stepName && len(form.name) > 0 {
			form.name = form.name[:len(form.name)-1]
		} else if form.currentStep == stepWorkers && len(form.workers) > 0 {
			form.workers = form.workers[:len(form.workers)-1]
		}

	default:
		// Handle text input
		if len(msg.String()) == 1 {
			if form.currentStep == stepName {
				form.name += msg.String()
			} else if form.currentStep == stepWorkers && msg.String() >= "0" && msg.String() <= "9" {
				form.workers += msg.String()
			}
		}
	}

	return m, nil
}

// handleEnterKey handles the enter key press based on current step
func (m Model) handleEnterKey() (tea.Model, tea.Cmd) {
	form := m.createForm

	switch form.currentStep {
	case stepProvider:
		// Move to next step
		form.currentStep = stepName
		m.err = nil

	case stepName:
		// Validate name
		if form.name == "" {
			m.err = fmt.Errorf("cluster name is required")
			return m, nil
		}
		// Move to next step
		form.currentStep = stepWorkers
		m.err = nil

	case stepWorkers:
		// Validate workers
		workers, err := strconv.Atoi(form.workers)
		if err != nil || workers < 0 {
			m.err = fmt.Errorf("workers must be a non-negative number")
			return m, nil
		}
		// Move to review step
		form.currentStep = stepReview
		m.err = nil

	case stepReview:
		// Create cluster
		workers, _ := strconv.Atoi(form.workers)
		provider := form.providers[form.selectedProvider]
		m.loading = true
		m.err = nil

		// Don't clear form or switch views yet - wait for success/failure
		return m, createClusterCmd(m.providers, provider.Name(), form.name, workers)
	}

	return m, nil
}
