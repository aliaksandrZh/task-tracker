package inputbar

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	appTui "github.com/aliaksandrZh/worklog/src/tui"
)

var borderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("6")).
	Padding(0, 1)

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
	m.input.Width = w - 8 // account for prompt, border, and padding
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

// View renders the input bar: bordered input, then hints below.
// Returns empty string when inactive.
func (m Model) View() string {
	if !m.active {
		return ""
	}

	boxWidth := m.width - 2 // leave margin for border
	if boxWidth < 20 {
		boxWidth = 20
	}
	out := borderStyle.Width(boxWidth).Render(m.input.View())
	if m.config.Hints != "" {
		out += "\n" + appTui.HintStyle.Render(m.config.Hints)
	}
	return out
}

// Height returns the number of lines the input bar occupies.
func (m Model) Height() int {
	if !m.active {
		return 0
	}
	// border adds 2 lines (top + bottom), plus optional hints
	h := 3 // top border + input + bottom border
	if m.config.Hints != "" {
		h++ // hints line
	}
	return h
}
