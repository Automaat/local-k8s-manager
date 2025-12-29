package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/automaat/local-k8s-manager/internal/backend"
)

// keyMsg creates a KeyMsg from a string for testing
func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "?":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	case "q":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	case "j":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	case "k":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	case "c":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	case "d":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	case "s":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	case "x":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	case "r":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestNewCreateFormModel(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
		backend.NewKindProvider(),
	}

	form := newCreateFormModel(providers)

	if form == nil {
		t.Fatal("expected non-nil form")
	}

	if len(form.providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(form.providers))
	}

	if form.selectedProvider != 0 {
		t.Errorf("expected selectedProvider to be 0, got %d", form.selectedProvider)
	}

	if form.name != "" {
		t.Errorf("expected empty name, got %s", form.name)
	}

	if form.workers != "1" {
		t.Errorf("expected workers to be '1', got %s", form.workers)
	}

	if form.focusedField != providerField {
		t.Errorf("expected focusedField to be providerField, got %v", form.focusedField)
	}
}

func TestRenderCreateView(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := Model{
		providers:  providers,
		createForm: newCreateFormModel(providers),
		width:      80,
	}

	result := m.renderCreateView()

	if result == "" {
		t.Error("expected non-empty result")
	}

	if result == "Form not initialized" {
		t.Error("expected form to be initialized")
	}
}

func TestRenderCreateViewWithoutForm(t *testing.T) {
	m := Model{
		width: 80,
	}

	result := m.renderCreateView()

	if result != "Form not initialized" {
		t.Errorf("expected 'Form not initialized', got %s", result)
	}
}

func TestRenderFormField(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := Model{
		createForm: newCreateFormModel(providers),
	}

	result := m.renderFormField("Test Label", providerField)
	if result == "" {
		t.Error("expected non-empty result")
	}

	// Test with different field
	result = m.renderFormField("Other Label", nameField)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestRenderCreateHelp(t *testing.T) {
	m := Model{}

	result := m.renderCreateHelp()
	if result == "" {
		t.Error("expected non-empty help text")
	}
}

func TestHandleCreateViewKeys(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
		backend.NewKindProvider(),
	}

	tests := []struct {
		name          string
		key           string
		initialForm   *createFormModel
		expectedView  viewState
		checkFormNil  bool
		validateForm  func(*testing.T, *createFormModel)
	}{
		{
			name:         "esc key returns to list view",
			key:          "esc",
			initialForm:  newCreateFormModel(providers),
			expectedView: listView,
			checkFormNil: true,
		},
		{
			name:         "tab key cycles to next field",
			key:          "tab",
			initialForm:  newCreateFormModel(providers),
			expectedView: createView,
			validateForm: func(t *testing.T, form *createFormModel) {
				if form.focusedField != nameField {
					t.Errorf("expected nameField, got %v", form.focusedField)
				}
			},
		},
		{
			name: "shift+tab key cycles to previous field",
			key:  "shift+tab",
			initialForm: &createFormModel{
				providers:        providers,
				selectedProvider: 0,
				name:             "",
				workers:          "1",
				focusedField:     nameField,
			},
			expectedView: createView,
			validateForm: func(t *testing.T, form *createFormModel) {
				if form.focusedField != providerField {
					t.Errorf("expected providerField, got %v", form.focusedField)
				}
			},
		},
		{
			name:         "down key selects next provider",
			key:          "down",
			initialForm:  newCreateFormModel(providers),
			expectedView: createView,
			validateForm: func(t *testing.T, form *createFormModel) {
				if form.selectedProvider != 1 {
					t.Errorf("expected selectedProvider to be 1, got %d", form.selectedProvider)
				}
			},
		},
		{
			name: "up key selects previous provider",
			key:  "up",
			initialForm: &createFormModel{
				providers:        providers,
				selectedProvider: 1,
				name:             "",
				workers:          "1",
				focusedField:     providerField,
			},
			expectedView: createView,
			validateForm: func(t *testing.T, form *createFormModel) {
				if form.selectedProvider != 0 {
					t.Errorf("expected selectedProvider to be 0, got %d", form.selectedProvider)
				}
			},
		},
		{
			name: "backspace in name field",
			key:  "backspace",
			initialForm: &createFormModel{
				providers:        providers,
				selectedProvider: 0,
				name:             "test",
				workers:          "1",
				focusedField:     nameField,
			},
			expectedView: createView,
			validateForm: func(t *testing.T, form *createFormModel) {
				if form.name != "tes" {
					t.Errorf("expected name 'tes', got %s", form.name)
				}
			},
		},
		{
			name: "backspace in workers field",
			key:  "backspace",
			initialForm: &createFormModel{
				providers:        providers,
				selectedProvider: 0,
				name:             "",
				workers:          "10",
				focusedField:     workersField,
			},
			expectedView: createView,
			validateForm: func(t *testing.T, form *createFormModel) {
				if form.workers != "1" {
					t.Errorf("expected workers '1', got %s", form.workers)
				}
			},
		},
		{
			name: "question mark shows help",
			key:  "?",
			initialForm: newCreateFormModel(providers),
			expectedView: helpView,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				providers:  providers,
				createForm: tt.initialForm,
				view:       createView,
			}

			newModel, _ := m.handleCreateViewKeys(keyMsg(tt.key))
			m = newModel.(Model)

			if m.view != tt.expectedView {
				t.Errorf("expected view %v, got %v", tt.expectedView, m.view)
			}

			if tt.checkFormNil && m.createForm != nil {
				t.Error("expected form to be nil")
			}

			if tt.validateForm != nil && m.createForm != nil {
				tt.validateForm(t, m.createForm)
			}
		})
	}
}

func TestHandleCreateViewKeysTextInput(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	// Test name field text input
	m := Model{
		providers: providers,
		createForm: &createFormModel{
			providers:        providers,
			selectedProvider: 0,
			name:             "",
			workers:          "1",
			focusedField:     nameField,
		},
		view: createView,
	}

	// Simulate typing 'a'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.handleCreateViewKeys(msg)
	m = newModel.(Model)

	if m.createForm.name != "a" {
		t.Errorf("expected name 'a', got %s", m.createForm.name)
	}

	// Test workers field numeric input
	m.createForm.focusedField = workersField
	m.createForm.workers = ""

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, _ = m.handleCreateViewKeys(msg)
	m = newModel.(Model)

	if m.createForm.workers != "5" {
		t.Errorf("expected workers '5', got %s", m.createForm.workers)
	}

	// Test workers field rejects non-numeric input
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ = m.handleCreateViewKeys(msg)
	m = newModel.(Model)

	if m.createForm.workers != "5" {
		t.Errorf("expected workers to remain '5', got %s", m.createForm.workers)
	}
}

func TestHandleCreateViewKeysEnterValidation(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	tests := []struct {
		name        string
		clusterName string
		workers     string
		expectError bool
	}{
		{
			name:        "valid cluster",
			clusterName: "test-cluster",
			workers:     "2",
			expectError: false,
		},
		{
			name:        "empty name",
			clusterName: "",
			workers:     "1",
			expectError: true,
		},
		{
			name:        "invalid workers",
			clusterName: "test",
			workers:     "abc",
			expectError: true,
		},
		{
			name:        "negative workers",
			clusterName: "test",
			workers:     "-1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				providers: providers,
				createForm: &createFormModel{
					providers:        providers,
					selectedProvider: 0,
					name:             tt.clusterName,
					workers:          tt.workers,
					focusedField:     providerField,
				},
				view: createView,
			}

			newModel, _ := m.handleCreateViewKeys(tea.KeyMsg{Type: tea.KeyEnter})
			m = newModel.(Model)

			if tt.expectError {
				if m.err == nil {
					t.Error("expected error, got nil")
				}
				if m.view != createView {
					t.Errorf("expected to stay in createView on error, got %v", m.view)
				}
			} else {
				if m.err != nil {
					t.Errorf("expected no error, got %v", m.err)
				}
				if m.view != listView {
					t.Errorf("expected to switch to listView, got %v", m.view)
				}
				if m.createForm != nil {
					t.Error("expected form to be nil after successful submit")
				}
			}
		})
	}
}

func TestHandleCreateViewKeysWithoutForm(t *testing.T) {
	m := Model{
		view: createView,
	}

	newModel, _ := m.handleCreateViewKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.view != listView {
		t.Errorf("expected listView when form is nil, got %v", m.view)
	}
}
