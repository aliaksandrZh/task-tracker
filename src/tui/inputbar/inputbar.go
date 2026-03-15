package inputbar

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	appTui "github.com/aliaksandrZh/worklog/src/tui"
)

// Config holds the display settings for an active input bar.
type Config struct {
	Placeholder string // ghost text shown when input is empty
	Hints       string // keybinding hints (e.g., "Enter=submit  Escape=cancel")
}

// Model is a reusable bottom input bar with placeholder and keybinding hints.
type Model struct {
	input  textinput.Model
	config Config
	active bool
	width  int
}

// New creates an inactive input bar.
func New() Model {
	ti := textinput.New()
	ti.Prompt = "▷ "
	ti.CharLimit = 500
	return Model{input: ti}
}

// Activate shows the input bar with the given config and focuses it.
func (m *Model) Activate(cfg Config) {
	m.config = cfg
	m.active = true
	m.input.Placeholder = cfg.Placeholder
	m.input.SetValue("")
	m.input.Focus()
}

// ActivateWithValue shows the input bar pre-filled with a value.
func (m *Model) ActivateWithValue(cfg Config, value string) {
	m.config = cfg
	m.active = true
	m.input.Placeholder = cfg.Placeholder
	m.input.SetValue(value)
	m.input.SetCursor(len(value))
	m.input.Focus()
}

// Deactivate hides the input bar and clears its value.
func (m *Model) Deactivate() {
	m.active = false
	m.input.SetValue("")
	m.input.Blur()
}

// Active returns whether the input bar is currently shown.
func (m Model) Active() bool {
	return m.active
}

// Value returns the current input text.
func (m Model) Value() string {
	return m.input.Value()
}

// SetValue sets the input text.
func (m *Model) SetValue(s string) {
	m.input.SetValue(s)
	m.input.SetCursor(len(s))
}

// SetWidth updates the rendering width.
func (m *Model) SetWidth(w int) {
	m.width = w
	m.input.Width = w - 4 // account for prompt and padding
}

// Update forwards messages to the underlying textinput.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the input bar: hints line + input line.
// Returns empty string when inactive.
func (m Model) View() string {
	if !m.active {
		return ""
	}

	var out string
	if m.config.Hints != "" {
		out += appTui.HintStyle.Render(m.config.Hints) + "\n"
	}
	out += m.input.View()
	return out
}

// Height returns the number of lines the input bar occupies.
func (m Model) Height() int {
	if !m.active {
		return 0
	}
	if m.config.Hints != "" {
		return 2 // hints + input
	}
	return 1 // input only
}
