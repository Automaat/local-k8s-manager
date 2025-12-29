package tui

import tea "github.com/charmbracelet/bubbletea"

// keyMsg creates a KeyMsg from a string for testing
func keyMsg(s string) tea.KeyMsg {
	keyMap := map[string]tea.KeyMsg{
		"esc":       {Type: tea.KeyEsc},
		"enter":     {Type: tea.KeyEnter},
		"tab":       {Type: tea.KeyTab},
		"shift+tab": {Type: tea.KeyShiftTab},
		"up":        {Type: tea.KeyUp},
		"down":      {Type: tea.KeyDown},
		"backspace": {Type: tea.KeyBackspace},
		"ctrl+c":    {Type: tea.KeyCtrlC},
		"?":         {Type: tea.KeyRunes, Runes: []rune{'?'}},
		"q":         {Type: tea.KeyRunes, Runes: []rune{'q'}},
		"j":         {Type: tea.KeyRunes, Runes: []rune{'j'}},
		"k":         {Type: tea.KeyRunes, Runes: []rune{'k'}},
		"c":         {Type: tea.KeyRunes, Runes: []rune{'c'}},
		"d":         {Type: tea.KeyRunes, Runes: []rune{'d'}},
		"s":         {Type: tea.KeyRunes, Runes: []rune{'s'}},
		"x":         {Type: tea.KeyRunes, Runes: []rune{'x'}},
		"r":         {Type: tea.KeyRunes, Runes: []rune{'r'}},
	}

	if msg, ok := keyMap[s]; ok {
		return msg
	}

	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}
