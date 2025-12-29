package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/automaat/local-k8s-manager/internal/backend"
)

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

	if form.currentStep != stepProvider {
		t.Errorf("expected currentStep to be stepProvider, got %v", form.currentStep)
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

func TestRenderStepProvider(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := Model{
		createForm: newCreateFormModel(providers),
		width:      80,
	}

	result := m.renderStepProvider()
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestRenderStepName(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := Model{
		createForm: &createFormModel{
			providers:        providers,
			selectedProvider: 0,
			name:             "",
			workers:          "1",
			currentStep:      stepName,
		},
		width: 80,
	}

	result := m.renderStepName()
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestRenderStepWorkers(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := Model{
		createForm: &createFormModel{
			providers:        providers,
			selectedProvider: 0,
			name:             "test",
			workers:          "1",
			currentStep:      stepWorkers,
		},
		width: 80,
	}

	result := m.renderStepWorkers()
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestRenderStepReview(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	m := Model{
		createForm: &createFormModel{
			providers:        providers,
			selectedProvider: 0,
			name:             "test",
			workers:          "2",
			currentStep:      stepReview,
		},
		width: 80,
	}

	result := m.renderStepReview()
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestHandleCreateViewKeys(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
		backend.NewKindProvider(),
	}

	tests := []struct {
		name         string
		key          string
		initialForm  *createFormModel
		expectedView viewState
		checkFormNil bool
		validateForm func(*testing.T, *createFormModel)
	}{
		{
			name:         "esc key on step 1 returns to list view",
			key:          "esc",
			initialForm:  newCreateFormModel(providers),
			expectedView: listView,
			checkFormNil: true,
		},
		{
			name: "esc key on step 2 goes back to step 1",
			key:  "esc",
			initialForm: &createFormModel{
				providers:        providers,
				selectedProvider: 0,
				name:             "",
				workers:          "1",
				currentStep:      stepName,
			},
			expectedView: createView,
			validateForm: func(t *testing.T, form *createFormModel) {
				if form.currentStep != stepProvider {
					t.Errorf("expected stepProvider, got %v", form.currentStep)
				}
			},
		},
		{
			name:         "down key selects next provider on step 1",
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
			name: "up key selects previous provider on step 1",
			key:  "up",
			initialForm: &createFormModel{
				providers:        providers,
				selectedProvider: 1,
				name:             "",
				workers:          "1",
				currentStep:      stepProvider,
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
				currentStep:      stepName,
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
				name:             "test",
				workers:          "10",
				currentStep:      stepWorkers,
			},
			expectedView: createView,
			validateForm: func(t *testing.T, form *createFormModel) {
				if form.workers != "1" {
					t.Errorf("expected workers '1', got %s", form.workers)
				}
			},
		},
		{
			name:         "question mark shows help",
			key:          "?",
			initialForm:  newCreateFormModel(providers),
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

	// Test name field text input on step 2
	m := Model{
		providers: providers,
		createForm: &createFormModel{
			providers:        providers,
			selectedProvider: 0,
			name:             "",
			workers:          "1",
			currentStep:      stepName,
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

	// Test workers field numeric input on step 3
	m.createForm.currentStep = stepWorkers
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

func TestHandleEnterKeyStepNavigation(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	tests := []struct {
		name         string
		currentStep  createStep
		clusterName  string
		workers      string
		expectedStep createStep
		expectError  bool
		expectSubmit bool
	}{
		{
			name:         "step 1 to step 2",
			currentStep:  stepProvider,
			clusterName:  "",
			workers:      "1",
			expectedStep: stepName,
			expectError:  false,
		},
		{
			name:         "step 2 to step 3 with valid name",
			currentStep:  stepName,
			clusterName:  "test-cluster",
			workers:      "1",
			expectedStep: stepWorkers,
			expectError:  false,
		},
		{
			name:         "step 2 stays on error with empty name",
			currentStep:  stepName,
			clusterName:  "",
			workers:      "1",
			expectedStep: stepName,
			expectError:  true,
		},
		{
			name:         "step 3 to step 4 with valid workers",
			currentStep:  stepWorkers,
			clusterName:  "test",
			workers:      "2",
			expectedStep: stepReview,
			expectError:  false,
		},
		{
			name:         "step 3 stays on error with invalid workers",
			currentStep:  stepWorkers,
			clusterName:  "test",
			workers:      "abc",
			expectedStep: stepWorkers,
			expectError:  true,
		},
		{
			name:         "step 4 submits form",
			currentStep:  stepReview,
			clusterName:  "test",
			workers:      "2",
			expectSubmit: true,
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
					currentStep:      tt.currentStep,
				},
				view: createView,
			}

			newModel, _ := m.handleCreateViewKeys(tea.KeyMsg{Type: tea.KeyEnter})
			m = newModel.(Model)

			if tt.expectError {
				if m.err == nil {
					t.Error("expected error, got nil")
				}
				if m.createForm == nil {
					t.Fatal("expected form to still exist on error")
				}
				if m.createForm.currentStep != tt.expectedStep {
					t.Errorf("expected to stay on step %v, got %v", tt.expectedStep, m.createForm.currentStep)
				}
			} else if tt.expectSubmit {
				// After pressing enter on step 4, we stay in create view and set loading
				// Form is not cleared until operationCompleteMsg is received
				if m.err != nil {
					t.Errorf("expected no error when starting submit, got %v", m.err)
				}
				if !m.loading {
					t.Error("expected loading to be true when submitting")
				}
				if m.createForm == nil {
					t.Error("expected form to still exist while operation is in progress")
				}
			} else {
				if m.err != nil {
					t.Errorf("expected no error, got %v", m.err)
				}
				if m.createForm == nil {
					t.Fatal("expected form to still exist")
				}
				if m.createForm.currentStep != tt.expectedStep {
					t.Errorf("expected step %v, got %v", tt.expectedStep, m.createForm.currentStep)
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
