package paste

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aliaksandrZh/worklog/src/internal/model"
	"github.com/aliaksandrZh/worklog/src/internal/parser"
	"github.com/aliaksandrZh/worklog/src/internal/store"
	appTui "github.com/aliaksandrZh/worklog/src/tui"
	"github.com/aliaksandrZh/worklog/src/tui/table"
)

var fieldLabels = map[string]string{
	"date":      "Date (M/D/YYYY)",
	"type":      "Type (Bug/Task)",
	"number":    "Number",
	"name":      "Name",
	"timeSpent": "Time Spent (e.g. 1h, 30m)",
}

type pastePhase int

const (
	phaseInput pastePhase = iota
	phaseFill
	phasePreview
)

// Model is the paste tasks screen.
type Model struct {
	phase      pastePhase
	input      textinput.Model
	parsed     []model.ParsedTask
	fillTaskIdx  int
	fillFieldIdx int
	repo       store.TaskRepository
	width      int
}

// New creates a new paste model.
func New(repo store.TaskRepository) *Model {
	ti := textinput.New()
	ti.Placeholder = "Bug 123: Fix login 1h"
	ti.Focus()
	ti.CharLimit = 500
	return &Model{
		phase: phaseInput,
		input: ti,
		repo:  repo,
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (appTui.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch m.phase {
		case phaseInput:
			return m.updateInput(msg)
		case phaseFill:
			return m.updateFill(msg)
		case phasePreview:
			return m.updatePreview(msg)
		}
	}

	if m.phase == phaseInput || m.phase == phaseFill {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) updateInput(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		return m, done()
	case tea.KeyEnter:
		text := m.input.Value()
		if text == "" {
			return m, nil
		}
		tasks := parser.ParsePastedText(text)
		if len(tasks) == 0 {
			return m, tea.Batch(flash("No tasks parsed. Check format."), done())
		}
		m.parsed = tasks
		m.startFillOrPreview()
		m.input.SetValue("")
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) updateFill(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		return m, done()
	case tea.KeyEnter:
		return m.submitFill()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) updatePreview(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		tasks := make([]model.Task, len(m.parsed))
		for i, p := range m.parsed {
			tasks[i] = p.Task
		}
		_ = m.repo.AddTasks(tasks)
		return m, tea.Batch(
			flash(fmt.Sprintf("%d task(s) saved!", len(tasks))),
			done(),
		)
	case tea.KeyEscape:
		return m, done()
	}
	return m, nil
}

func (m *Model) startFillOrPreview() {
	idx := m.findNextFill(0, 0)
	if idx == nil {
		m.phase = phasePreview
	} else {
		m.fillTaskIdx = idx[0]
		m.fillFieldIdx = idx[1]
		m.input.SetValue("")
		m.input.Focus()
		m.phase = phaseFill
	}
}

// findNextFill finds the next required missing field (skip timeSpent and date).
func (m *Model) findNextFill(fromTask, fromField int) []int {
	for t := fromTask; t < len(m.parsed); t++ {
		required := requiredMissing(m.parsed[t].Missing)
		startF := 0
		if t == fromTask {
			startF = fromField
		}
		if startF < len(required) {
			return []int{t, startF}
		}
	}
	return nil
}

func requiredMissing(missing []string) []string {
	var result []string
	for _, f := range missing {
		if f != "timeSpent" && f != "date" {
			result = append(result, f)
		}
	}
	return result
}

func (m *Model) submitFill() (appTui.ScreenModel, tea.Cmd) {
	task := &m.parsed[m.fillTaskIdx]
	required := requiredMissing(task.Missing)
	fieldName := required[m.fillFieldIdx]
	val := m.input.Value()

	// Set the field value
	switch fieldName {
	case "type":
		task.Type = val
	case "number":
		task.Number = val
	case "name":
		task.Name = val
	}

	// Remove from missing
	var newMissing []string
	for _, f := range task.Missing {
		if f != fieldName {
			newMissing = append(newMissing, f)
		}
	}
	task.Missing = newMissing
	m.input.SetValue("")

	// Find next
	next := m.findNextFill(m.fillTaskIdx, m.fillFieldIdx)
	if next == nil {
		m.phase = phasePreview
	} else {
		m.fillTaskIdx = next[0]
		m.fillFieldIdx = next[1]
	}
	return m, nil
}

func (m *Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}

	switch m.phase {
	case phaseInput:
		return m.viewInput()
	case phaseFill:
		return m.viewFill()
	default:
		return m.viewPreview(w)
	}
}

func (m *Model) viewInput() string {
	var b strings.Builder
	b.WriteString(appTui.TitleStyle.Render("Paste Tasks") + "\n")
	b.WriteString(appTui.HintStyle.Render("Type or paste your task. Enter=submit, Escape=cancel") + "\n")
	b.WriteString(appTui.HintStyle.Render("Missing fields will be prompted. Examples:") + "\n")
	b.WriteString(appTui.HintStyle.Render("  Bug 123: Fix login 1h") + "\n")
	b.WriteString(appTui.HintStyle.Render("  Task 456: Add feature 2h 30m") + "\n")
	b.WriteString(appTui.HintStyle.Render("  123: Fix crash") + "\n\n")
	// Set input width to terminal width so long text scrolls instead of overflowing
	termW := m.width
	if termW <= 0 {
		termW = 80
	}
	m.input.Width = termW - 2
	b.WriteString(m.input.View() + "\n")
	return b.String()
}

func (m *Model) viewFill() string {
	var b strings.Builder
	b.WriteString(appTui.TitleStyle.Render("Fill Missing Fields") + "\n")
	b.WriteString(appTui.HintStyle.Render("Press Escape to cancel") + "\n\n")

	task := m.parsed[m.fillTaskIdx]
	taskLabel := task.Name
	if taskLabel == "" && task.Number != "" {
		taskLabel = "#" + task.Number
	}
	if taskLabel == "" {
		taskLabel = "(unknown)"
	}
	b.WriteString(fmt.Sprintf("Task %d/%d: %s\n", m.fillTaskIdx+1, len(m.parsed), taskLabel))

	// Show known fields
	if task.Date != "" {
		b.WriteString(fmt.Sprintf("  date: %s\n", task.Date))
	}
	if task.Type != "" {
		b.WriteString(fmt.Sprintf("  type: %s\n", task.Type))
	}
	if task.Number != "" {
		b.WriteString(fmt.Sprintf("  number: %s\n", task.Number))
	}
	if task.Name != "" {
		b.WriteString(fmt.Sprintf("  name: %s\n", task.Name))
	}

	required := requiredMissing(task.Missing)
	fieldName := required[m.fillFieldIdx]
	label := fieldLabels[fieldName]
	if label == "" {
		label = fieldName
	}

	// Limit textinput width so long values don't overflow
	termW := m.width
	if termW <= 0 {
		termW = 80
	}
	labelWidth := len(label) + 2 // label + ": "
	inputWidth := termW - labelWidth - 2
	if inputWidth < 20 {
		inputWidth = 20
	}
	m.input.Width = inputWidth

	b.WriteString("\n")
	b.WriteString(appTui.PromptStyle.Render(label+": ") + m.input.View() + "\n")
	return b.String()
}

func (m *Model) viewPreview(w int) string {
	var b strings.Builder
	b.WriteString(appTui.TitleStyle.Render(fmt.Sprintf("Preview (%d tasks parsed)", len(m.parsed))) + "\n")

	// Convert to IndexedTask for table
	indexed := make([]model.IndexedTask, len(m.parsed))
	for i, p := range m.parsed {
		indexed[i] = model.IndexedTask{Task: p.Task, Index: i}
	}
	cfg := table.Config{
		Width:       w,
		SelectedRow: -1,
	}
	b.WriteString(table.Render(indexed, cfg) + "\n\n")
	b.WriteString(appTui.FlashStyle.Render("Press Enter to save, Escape to cancel") + "\n")
	return b.String()
}

func done() tea.Cmd {
	return func() tea.Msg { return appTui.DoneMsg{} }
}

func flash(text string) tea.Cmd {
	return func() tea.Msg { return appTui.FlashMsg{Text: text} }
}
