package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/automaat/local-k8s-manager/internal/backend"
)

// keyMsgHelp creates a KeyMsg from a string for testing
func keyMsgHelp(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "?":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	case "q":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestRenderHelpView(t *testing.T) {
	m := Model{
		width:  80,
		height: 24,
	}

	result := m.renderHelpView()

	if result == "" {
		t.Error("expected non-empty help view")
	}
}

func TestHandleHelpViewKeys(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		previousView viewState
		shouldQuit   bool
	}{
		{
			name:         "question mark closes help",
			key:          "?",
			previousView: listView,
			shouldQuit:   false,
		},
		{
			name:         "esc closes help",
			key:          "esc",
			previousView: detailView,
			shouldQuit:   false,
		},
		{
			name:         "q quits application",
			key:          "q",
			previousView: listView,
			shouldQuit:   true,
		},
		{
			name:         "ctrl+c quits application",
			key:          "ctrl+c",
			previousView: listView,
			shouldQuit:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				view:         helpView,
				previousView: tt.previousView,
			}

			newModel, cmd := m.handleHelpViewKeys(keyMsgHelp(tt.key))
			m = newModel.(Model)

			if tt.shouldQuit {
				if cmd == nil {
					t.Error("expected quit command")
				}
			} else {
				if m.view != tt.previousView {
					t.Errorf("expected view %v, got %v", tt.previousView, m.view)
				}
			}
		})
	}
}

func TestRenderHelpViewWithSmallWidth(t *testing.T) {
	m := Model{
		width:  40,
		height: 24,
	}

	result := m.renderHelpView()

	if result == "" {
		t.Error("expected non-empty help view even with small width")
	}
}

func TestRenderHelpViewFromDifferentPreviousViews(t *testing.T) {
	providers := []backend.Provider{
		backend.NewK3dProvider(),
	}

	views := []viewState{listView, detailView, createView}

	for _, prevView := range views {
		m := Model{
			view:         helpView,
			previousView: prevView,
			width:        80,
			height:       24,
			providers:    providers,
		}

		result := m.renderHelpView()
		if result == "" {
			t.Errorf("expected non-empty help view for previousView %v", prevView)
		}
	}
}
