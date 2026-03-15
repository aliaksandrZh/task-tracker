package summary

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/aliaksandrZh/worklog/src/internal/model"
	"github.com/aliaksandrZh/worklog/src/internal/parser"
	"github.com/aliaksandrZh/worklog/src/internal/prefs"
	"github.com/aliaksandrZh/worklog/src/internal/store"
	"github.com/aliaksandrZh/worklog/src/internal/timeutil"
	"github.com/aliaksandrZh/worklog/src/internal/timer"
	appTui "github.com/aliaksandrZh/worklog/src/tui"
	"github.com/aliaksandrZh/worklog/src/tui/inputbar"
	"github.com/aliaksandrZh/worklog/src/tui/table"
)

var sortColumns = []string{"", "date", "type", "number", "name", "timeSpent"}

type phase int

const (
	phaseView phase = iota
	phaseSelect
	phaseEditing
	phaseConfirmDelete
	phaseFilter
	phaseAdding
	phaseAddFill
)

// Model is the summary screen.
type Model struct {
	mode        string // "daily", "weekly", or "monthly"
	phase       phase
	dailyIdx    int
	weekOffset  int
	monthOffset int

	sortBy  string
	sortDir string

	selectedRow int
	selectedCol int
	editInput   textinput.Model
	editInline  bool // true = edit in cell, false = edit below table

	filterInput textinput.Model
	filterText  string // active filter (applied when non-empty)

	inputBar     inputbar.Model
	addParsed    []model.ParsedTask
	addTaskIdx   int
	addFieldIdx  int

	repo  store.TaskRepository
	tmr   *timer.Timer
	prefs *prefs.Store

	allTasks     []model.Task
	indexedAll   []model.IndexedTask
	dailyGroups  []timeutil.DateGroup
	displayed    []model.IndexedTask
	weeklyGroups  []timeutil.DateGroup // day groups within current week
	monthlyGroups []timeutil.DateGroup // day groups within current month

	notifications appTui.Notifications

	width  int
	height int
	vp     viewport.Model
}

// New creates a new summary model.
func New(repo store.TaskRepository, tmr *timer.Timer) *Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 200

	fi := textinput.New()
	fi.Prompt = ""
	fi.CharLimit = 100
	fi.Placeholder = "type to filter..."

	p := prefs.New(".")
	pref := p.Load()

	m := &Model{
		mode:        "daily",
		phase:       phaseView,
		sortBy:      pref.SortBy,
		sortDir:     pref.SortDir,
		editInput:   ti,
		filterInput: fi,
		inputBar:    inputbar.New(),
		repo:        repo,
		tmr:         tmr,
		prefs:       p,
	}
	if m.sortDir == "" {
		m.sortDir = "asc"
	}
	m.reload()
	return m
}

func (m *Model) reload() {
	tasks, _ := m.repo.LoadTasks()
	m.allTasks = tasks
	m.indexedAll = make([]model.IndexedTask, len(tasks))
	for i, t := range tasks {
		m.indexedAll[i] = model.IndexedTask{Task: t, Index: i}
	}
	m.dailyGroups = timeutil.GroupByDate(m.indexedAll)
	m.ensureTodayGroup()
	m.refreshDisplayed()
}

// ensureTodayGroup makes sure today's date is in dailyGroups and sets dailyIdx to it.
func (m *Model) ensureTodayGroup() {
	idx := timeutil.TodayIndex(m.dailyGroups)
	if idx >= 0 {
		m.dailyIdx = idx
		return
	}
	// Insert synthetic empty group for today at index 0 (most recent)
	today := timeutil.DateGroup{Key: timeutil.TodayStr()}
	m.dailyGroups = append([]timeutil.DateGroup{today}, m.dailyGroups...)
	m.dailyIdx = 0
}

func (m *Model) refreshDisplayed() {
	if m.mode == "monthly" {
		result := timeutil.FilterMonthByOffset(m.indexedAll, m.monthOffset)
		m.monthlyGroups = timeutil.GroupByDate(result.Tasks)
		m.displayed = nil
		for i, g := range m.monthlyGroups {
			filtered := filterTasks(g.Tasks, m.filterText)
			m.monthlyGroups[i].Tasks = filtered
			m.monthlyGroups[i].Total = sumHours(filtered)
			m.displayed = append(m.displayed, timeutil.SortTasks(filtered, m.sortBy, m.sortDir)...)
		}
	} else if m.mode == "weekly" {
		result := timeutil.FilterWeekByOffset(m.indexedAll, m.weekOffset)
		m.weeklyGroups = timeutil.GroupByDate(result.Tasks)
		m.displayed = nil
		for i, g := range m.weeklyGroups {
			filtered := filterTasks(g.Tasks, m.filterText)
			m.weeklyGroups[i].Tasks = filtered
			m.weeklyGroups[i].Total = sumHours(filtered)
			m.displayed = append(m.displayed, timeutil.SortTasks(filtered, m.sortBy, m.sortDir)...)
		}
	} else {
		m.weeklyGroups = nil
		m.monthlyGroups = nil
		var raw []model.IndexedTask
		if m.dailyIdx < len(m.dailyGroups) {
			raw = m.dailyGroups[m.dailyIdx].Tasks
		}
		m.displayed = timeutil.SortTasks(filterTasks(raw, m.filterText), m.sortBy, m.sortDir)
	}
}

// SetNotifications updates the notification zone data.
func (m *Model) SetNotifications(n appTui.Notifications) {
	m.notifications = n
}

// Reload refreshes data from the store (called when returning from sub-screens).
func (m *Model) Reload() {
	m.reload()
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (appTui.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.phase {
		case phaseAdding:
			return m.updateAdding(msg)
		case phaseAddFill:
			return m.updateAddFill(msg)
		case phaseFilter:
			return m.updateFilter(msg)
		case phaseEditing:
			return m.updateEditing(msg)
		case phaseConfirmDelete:
			return m.updateConfirmDelete(msg)
		case phaseSelect:
			return m.updateSelect(msg)
		default:
			return m.updateView(msg)
		}

	case tea.MouseMsg:
		if m.phase == phaseView || m.phase == phaseSelect {
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}
	}

	// Forward non-key messages (blink cursor, etc.) to textinput when editing/filtering/adding
	if m.phase == phaseAdding || m.phase == phaseAddFill {
		var cmd tea.Cmd
		m.inputBar, cmd = m.inputBar.Update(msg)
		return m, cmd
	}
	if m.phase == phaseEditing {
		if m.editInline {
			var cmd tea.Cmd
			m.editInput, cmd = m.editInput.Update(msg)
			return m, cmd
		}
		var cmd tea.Cmd
		m.inputBar, cmd = m.inputBar.Update(msg)
		return m, cmd
	}
	if m.phase == phaseFilter {
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) updateView(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	// Forward scroll keys to viewport
	switch msg.String() {
	case "up", "down", "pgup", "pgdown", "k", "j":
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "a":
		m.phase = phaseAdding
		m.inputBar.SetWidth(m.width)
		m.inputBar.Activate(inputbar.Config{
			Placeholder: "Bug 123: Fix login 1h",
			Hints:       "Enter=submit  Escape=cancel",
		})
		return m, textinput.Blink
	case "t":
		if m.tmr.GetStatus() != nil {
			return m, stopTimer()
		}
		return m, navigate(appTui.ScreenTimerStart)
	case "e":
		if len(m.displayed) > 0 {
			m.phase = phaseSelect
			m.selectedRow = 0
			m.selectedCol = 0
		}
	case "left":
		if m.mode == "daily" && m.dailyIdx < len(m.dailyGroups)-1 {
			m.dailyIdx++
			m.refreshDisplayed()
		} else if m.mode == "weekly" {
			next := m.adjacentWeekOffset(-1)
			if next != m.weekOffset {
				m.weekOffset = next
				m.refreshDisplayed()
			}
		} else if m.mode == "monthly" {
			next := m.adjacentMonthOffset(-1)
			if next != m.monthOffset {
				m.monthOffset = next
				m.refreshDisplayed()
			}
		}
	case "right":
		if m.mode == "daily" && m.dailyIdx > 0 {
			m.dailyIdx--
			m.refreshDisplayed()
		} else if m.mode == "weekly" {
			next := m.adjacentWeekOffset(1)
			if next != m.weekOffset {
				m.weekOffset = next
				m.refreshDisplayed()
			}
		} else if m.mode == "monthly" {
			next := m.adjacentMonthOffset(1)
			if next != m.monthOffset {
				m.monthOffset = next
				m.refreshDisplayed()
			}
		}
	case "d":
		if m.mode != "daily" {
			m.mode = "daily"
			m.refreshDisplayed()
		}
	case "w":
		if m.mode != "weekly" {
			m.mode = "weekly"
			m.refreshDisplayed()
		}
	case "m":
		if m.mode != "monthly" {
			m.mode = "monthly"
			m.monthOffset = 0
			m.refreshDisplayed()
		}
	case "f":
		m.filterInput.SetValue(m.filterText)
		m.filterInput.SetCursor(len(m.filterText))
		m.filterInput.Focus()
		m.phase = phaseFilter
		return m, textinput.Blink
	case "s":
		idx := indexOf(sortColumns, m.sortBy)
		m.sortBy = sortColumns[(idx+1)%len(sortColumns)]
		m.prefs.Save(prefs.Prefs{SortBy: m.sortBy})
		m.refreshDisplayed()
	case "S":
		if m.sortDir == "asc" {
			m.sortDir = "desc"
		} else {
			m.sortDir = "asc"
		}
		m.prefs.Save(prefs.Prefs{SortDir: m.sortDir})
		m.refreshDisplayed()
	}
	return m, nil
}

func (m *Model) updateSelect(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "escape", "e":
		m.phase = phaseView
		m.selectedRow = 0
		m.selectedCol = 0
	case "up", "k":
		if len(m.displayed) > 0 {
			m.selectedRow--
			if m.selectedRow < 0 {
				m.selectedRow = len(m.displayed) - 1
			}
		}
	case "down", "j":
		if len(m.displayed) > 0 {
			m.selectedRow++
			if m.selectedRow >= len(m.displayed) {
				m.selectedRow = 0
			}
		}
	case "left", "h":
		m.selectedCol--
		if m.selectedCol < 0 {
			m.selectedCol = len(table.Columns) - 1
		}
	case "right", "l":
		m.selectedCol++
		if m.selectedCol >= len(table.Columns) {
			m.selectedCol = 0
		}
	case "enter":
		if m.selectedRow >= 0 && m.selectedRow < len(m.displayed) {
			col := table.Columns[m.selectedCol]
			val := getField(m.displayed[m.selectedRow].Task, col)
			colW := table.ColWidth(col, m.width)

			// If value fits in column, edit inline; otherwise use bottom input bar
			if len([]rune(val)) < colW-1 {
				m.editInline = true
				m.editInput.SetValue(val)
				m.editInput.SetCursor(len(val))
				m.editInput.Width = colW
				m.editInput.Focus()
			} else {
				m.editInline = false
				label := table.HeaderLabel(col)
				m.inputBar.SetWidth(m.width)
				m.inputBar.ActivateWithValue(inputbar.Config{
					Placeholder: label,
					Hints:       fmt.Sprintf("Edit %s  Enter=save  Escape=cancel", label),
				}, val)
			}

			m.phase = phaseEditing
			return m, textinput.Blink
		}
	case "x":
		if m.selectedRow >= 0 && m.selectedRow < len(m.displayed) {
			m.phase = phaseConfirmDelete
		}
	case "s":
		idx := indexOf(sortColumns, m.sortBy)
		m.sortBy = sortColumns[(idx+1)%len(sortColumns)]
		m.prefs.Save(prefs.Prefs{SortBy: m.sortBy})
		m.refreshDisplayed()
	case "S":
		if m.sortDir == "asc" {
			m.sortDir = "desc"
		} else {
			m.sortDir = "asc"
		}
		m.prefs.Save(prefs.Prefs{SortDir: m.sortDir})
		m.refreshDisplayed()
	}
	return m, nil
}

func (m *Model) updateEditing(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		if !m.editInline {
			m.inputBar.Deactivate()
		}
		m.phase = phaseSelect
		return m, nil
	case tea.KeyEnter:
		task := m.displayed[m.selectedRow]
		col := table.Columns[m.selectedCol]
		var val string
		if m.editInline {
			val = m.editInput.Value()
		} else {
			val = m.inputBar.Value()
			m.inputBar.Deactivate()
		}
		_ = m.repo.UpdateTask(task.Index, map[string]string{col: val})
		m.reload()
		m.phase = phaseSelect
		return m, flash(fmt.Sprintf("Updated %s.", col))
	}

	// Forward all other keys to the active input
	if m.editInline {
		var cmd tea.Cmd
		m.editInput, cmd = m.editInput.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.inputBar, cmd = m.inputBar.Update(msg)
	return m, cmd
}

func (m *Model) updateFilter(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		if m.filterInput.Value() != "" {
			// First Esc: clear the filter text
			m.filterInput.SetValue("")
			m.filterText = ""
			m.refreshDisplayed()
			return m, nil
		}
		// Second Esc (already empty): close filter mode
		m.phase = phaseView
		return m, nil
	case tea.KeyEnter:
		m.filterText = m.filterInput.Value()
		m.phase = phaseView
		return m, nil
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	// Live filter as user types
	m.filterText = m.filterInput.Value()
	m.refreshDisplayed()
	return m, cmd
}

var addFieldLabels = map[string]string{
	"type":   "Type (Bug/Task)",
	"number": "Number",
	"name":   "Name",
}

func (m *Model) updateAdding(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.inputBar.Deactivate()
		m.phase = phaseView
		return m, nil
	case tea.KeyEnter:
		text := m.inputBar.Value()
		if text == "" {
			return m, nil
		}
		tasks := parser.ParsePastedText(text)
		if len(tasks) == 0 {
			m.inputBar.Deactivate()
			m.phase = phaseView
			return m, flash("No tasks parsed. Check format.")
		}
		m.addParsed = tasks
		m.inputBar.Deactivate()
		return m.startAddFillOrSave()
	}

	var cmd tea.Cmd
	m.inputBar, cmd = m.inputBar.Update(msg)
	return m, cmd
}

func (m *Model) updateAddFill(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.inputBar.Deactivate()
		m.phase = phaseView
		return m, nil
	case tea.KeyEnter:
		val := m.inputBar.Value()
		if val == "" {
			return m, nil
		}
		task := &m.addParsed[m.addTaskIdx]
		required := requiredAddMissing(task.Missing)
		fieldName := required[m.addFieldIdx]

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
		m.inputBar.Deactivate()
		return m.startAddFillOrSave()
	}

	var cmd tea.Cmd
	m.inputBar, cmd = m.inputBar.Update(msg)
	return m, cmd
}

// startAddFillOrSave checks for remaining missing fields; fills next or saves all.
func (m *Model) startAddFillOrSave() (appTui.ScreenModel, tea.Cmd) {
	for t := m.addTaskIdx; t < len(m.addParsed); t++ {
		required := requiredAddMissing(m.addParsed[t].Missing)
		startF := 0
		if t == m.addTaskIdx {
			startF = m.addFieldIdx
		}
		if startF < len(required) {
			m.addTaskIdx = t
			m.addFieldIdx = startF
			fieldName := required[startF]
			label := addFieldLabels[fieldName]
			if label == "" {
				label = fieldName
			}
			m.phase = phaseAddFill
			m.inputBar.SetWidth(m.width)
			m.inputBar.Activate(inputbar.Config{
				Placeholder: label,
				Hints:       fmt.Sprintf("Fill %s  Enter=submit  Escape=cancel", label),
			})
			return m, textinput.Blink
		}
	}

	// All fields filled — save
	tasks := make([]model.Task, len(m.addParsed))
	for i, p := range m.addParsed {
		tasks[i] = p.Task
	}
	_ = m.repo.AddTasks(tasks)
	m.addParsed = nil
	m.addTaskIdx = 0
	m.addFieldIdx = 0
	m.phase = phaseView
	m.reload()
	return m, flash(fmt.Sprintf("%d task(s) saved!", len(tasks)))
}

// requiredAddMissing returns missing fields excluding optional ones (timeSpent, date).
func requiredAddMissing(missing []string) []string {
	var result []string
	for _, f := range missing {
		if f != "timeSpent" && f != "date" {
			result = append(result, f)
		}
	}
	return result
}

func (m *Model) updateConfirmDelete(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		task := m.displayed[m.selectedRow]
		_ = m.repo.DeleteTask(task.Index)
		m.reload()
		if m.selectedRow >= len(m.displayed) {
			m.selectedRow = max(0, len(m.displayed)-1)
		}
		if len(m.displayed) == 0 {
			m.phase = phaseView
		} else {
			m.phase = phaseSelect
		}
		return m, flash("Task deleted.")
	case "n", "N", "esc", "escape":
		m.phase = phaseSelect
	}
	return m, nil
}

func countLines(s string) int {
	return strings.Count(s, "\n")
}

func (m *Model) timerHint() string {
	if m.tmr.GetStatus() != nil {
		return "[t]=stop"
	}
	return "[t]imer"
}

func (m *Model) View() string {
	if len(m.allTasks) == 0 && len(m.dailyGroups) <= 1 {
		return fmt.Sprintf("No tasks yet. [a]dd | %s | [q]uit\n", m.timerHint())
	}

	var header strings.Builder
	var body strings.Builder
	selectedLineY := -1 // track Y position of the selected row in body

	isEdit := m.phase == phaseSelect || m.phase == phaseEditing || m.phase == phaseConfirmDelete
	stateLabel := ""
	if isEdit {
		stateLabel = " (EDIT)"
	} else if m.phase == phaseAdding || m.phase == phaseAddFill {
		stateLabel = " [ADDING]"
	}

	sortHint := "off"
	if m.sortBy != "" {
		sortHint = m.sortBy + " " + m.sortDir
	}
	viewHint := fmt.Sprintf("[a]dd [e]dit %s | [d]aily [w]eekly [m]onthly | ← → nav | [f]ilter [s]ort(%s) [q]uit", m.timerHint(), sortHint)
	editHint := fmt.Sprintf("↑↓ row  ←→ col | Enter=edit [x]=delete | [s]ort(%s) [S]=flip | [e]/Esc=back", sortHint)
	var hintLine string
	if m.phase == phaseFilter {
		hintLine = "Enter=keep filter | Esc=clear, Esc again=close"
	} else if isEdit {
		hintLine = editHint
	} else {
		hintLine = viewHint
	}

	w := m.width
	if w <= 0 {
		w = 80
	}

	if m.mode == "monthly" {
		result := timeutil.FilterMonthByOffset(m.indexedAll, m.monthOffset)
		header.WriteString(appTui.TitleStyle.Render(fmt.Sprintf("Monthly Summary%s", stateLabel)) + "\n")
		header.WriteString(appTui.PromptStyle.Render(
			fmt.Sprintf("%s — %.1fh total (%d tasks)", result.Label, result.Total, len(result.Tasks))) + "\n")

		// Render day-by-day sections
		rowOffset := 0
		for _, g := range m.monthlyGroups {
			body.WriteString("\n")
			body.WriteString(appTui.PromptStyle.Render(
				fmt.Sprintf("%s — %.1fh total (%d tasks)", g.Key, g.Total, len(g.Tasks))))
			body.WriteString(" " + appTui.RemainingLabel(g.Total, 8) + " " + appTui.ProgressBar(g.Total, 8, 10) + "\n")

			sorted := timeutil.SortTasks(g.Tasks, m.sortBy, m.sortDir)
			cfg := table.Config{
				Width:            w,
				SortBy:           m.sortBy,
				SortDir:          m.sortDir,
				SelectedRow:      -1,
				SelectedCol:      m.selectedCol,
				ConfirmDeleteRow: -1,
			}
			if isEdit {
				localRow := m.selectedRow - rowOffset
				if localRow >= 0 && localRow < len(sorted) {
					cfg.SelectedRow = localRow
					if m.phase == phaseEditing && m.editInline {
						cfg.EditingCell = true
						cfg.EditView = m.editInput.View()
					}
					if m.phase == phaseConfirmDelete {
						cfg.ConfirmDeleteRow = localRow
					}
					selectedLineY = countLines(body.String()) + 1 + localRow
				}
			}
			body.WriteString(table.Render(sorted, cfg) + "\n")
			rowOffset += len(g.Tasks)
		}
	} else if m.mode == "weekly" {
		result := timeutil.FilterWeekByOffset(m.indexedAll, m.weekOffset)
		header.WriteString(appTui.TitleStyle.Render(fmt.Sprintf("Weekly Summary%s", stateLabel)) + "\n")
		header.WriteString(appTui.PromptStyle.Render(
			fmt.Sprintf("%s — %.1fh total (%d tasks)", result.Label, result.Total, len(result.Tasks))) + "\n")

		// Render day-by-day sections
		rowOffset := 0
		for _, g := range m.weeklyGroups {
			body.WriteString("\n")
			body.WriteString(appTui.PromptStyle.Render(
				fmt.Sprintf("%s — %.1fh total (%d tasks)", g.Key, g.Total, len(g.Tasks))))
			body.WriteString(" " + appTui.RemainingLabel(g.Total, 8) + " " + appTui.ProgressBar(g.Total, 8, 10) + "\n")

			sorted := timeutil.SortTasks(g.Tasks, m.sortBy, m.sortDir)
			cfg := table.Config{
				Width:            w,
				SortBy:           m.sortBy,
				SortDir:          m.sortDir,
				SelectedRow:      -1,
				SelectedCol:      m.selectedCol,
				ConfirmDeleteRow: -1,
			}
			if isEdit {
				localRow := m.selectedRow - rowOffset
				if localRow >= 0 && localRow < len(sorted) {
					cfg.SelectedRow = localRow
					if m.phase == phaseEditing && m.editInline {
						cfg.EditingCell = true
						cfg.EditView = m.editInput.View()
					}
					if m.phase == phaseConfirmDelete {
						cfg.ConfirmDeleteRow = localRow
					}
					selectedLineY = countLines(body.String()) + 1 + localRow
				}
			}
			body.WriteString(table.Render(sorted, cfg) + "\n")
			rowOffset += len(g.Tasks)
		}
	} else {
		dateLabel := ""
		if m.dailyIdx < len(m.dailyGroups) {
			dateLabel = " — " + m.dailyGroups[m.dailyIdx].Key
		}
		header.WriteString(appTui.TitleStyle.Render(fmt.Sprintf("Daily Summary%s%s", stateLabel, dateLabel)) + "\n")
		if m.dailyIdx < len(m.dailyGroups) {
			g := m.dailyGroups[m.dailyIdx]
			header.WriteString(appTui.PromptStyle.Render(
				fmt.Sprintf("%s — %.1fh total (%d tasks)", g.Key, g.Total, len(g.Tasks))))
			header.WriteString(" " + appTui.RemainingLabel(g.Total, 8) + " " + appTui.ProgressBar(g.Total, 8, 10) + "\n")
		}

		cfg := table.Config{
			Width:            w,
			SortBy:           m.sortBy,
			SortDir:          m.sortDir,
			SelectedRow:      -1,
			SelectedCol:      m.selectedCol,
			ConfirmDeleteRow: -1,
		}
		if isEdit {
			cfg.SelectedRow = m.selectedRow
			if m.phase == phaseEditing && m.editInline {
				cfg.EditingCell = true
				cfg.EditView = m.editInput.View()
			}
			if m.phase == phaseConfirmDelete {
				cfg.ConfirmDeleteRow = m.selectedRow
			}
			selectedLineY = countLines(body.String()) + 1 + m.selectedRow
		}
		body.WriteString(table.Render(m.displayed, cfg) + "\n")
	}


	bodyStr := body.String()

	// Calculate viewport height: terminal height - header lines - footer reserve (3 lines in app.go)
	if m.phase == phaseFilter {
		header.WriteString(appTui.PromptStyle.Render("Filter: ") + m.filterInput.View() + "\n")
	} else if m.filterText != "" {
		header.WriteString(appTui.HintStyle.Render(fmt.Sprintf("filter: %q  [f] to edit", m.filterText)) + "\n")
	}
	headerStr := header.String()
	headerLines := strings.Count(headerStr, "\n") + 1
	footerReserve := 4 // scroll hint + hint line + possible notification + margin
	if m.inputBar.Active() {
		footerReserve += m.inputBar.Height()
	}
	vpHeight := m.height - headerLines - footerReserve
	if vpHeight < 5 {
		vpHeight = 5
	}

	m.vp.Width = w
	m.vp.Height = vpHeight
	m.vp.SetContent(bodyStr)

	// Auto-scroll to keep selected row visible in edit mode
	if isEdit && m.selectedRow >= 0 {
		// Estimate the line position of the selected row in the body content
		if selectedLineY >= 0 && selectedLineY >= m.vp.YOffset+vpHeight {
			m.vp.SetYOffset(selectedLineY - vpHeight + 2)
		} else if selectedLineY >= 0 && selectedLineY < m.vp.YOffset {
			m.vp.SetYOffset(selectedLineY)
		}
	}

	// Scroll indicator — only show when content overflows
	var scrollHint string
	totalLines := strings.Count(bodyStr, "\n")
	if totalLines > vpHeight {
		atTop := m.vp.YOffset == 0
		atBottom := m.vp.YOffset >= totalLines-vpHeight
		pct := 0
		if totalLines-vpHeight > 0 {
			pct = m.vp.YOffset * 100 / (totalLines - vpHeight)
		}
		if atTop {
			scrollHint = appTui.HintStyle.Render("↓ scroll down")
		} else if atBottom {
			scrollHint = appTui.HintStyle.Render("↑ scroll up")
		} else {
			scrollHint = appTui.HintStyle.Render(fmt.Sprintf("↑↓ scroll (%d%%)", pct))
		}
	}

	out := headerStr + m.vp.View()
	if scrollHint != "" {
		out += "\n" + scrollHint
	}

	// Show input bar with its own hints, or the base shortcut line — not both
	if m.inputBar.Active() {
		out += "\n" + m.inputBar.View()
	} else {
		out += "\n" + appTui.HintStyle.Render(hintLine)
	}

	// Notification zone: only render lines that have content
	if m.notifications.TimerLine != "" {
		out += "\n" + m.notifications.TimerLine
	}
	if m.notifications.FlashLine != "" {
		out += "\n" + m.notifications.FlashLine
	}
	if m.notifications.UpdateLine != "" {
		out += "\n" + m.notifications.UpdateLine
	}
	return out
}

// weekOffsetsWithTasks returns sorted (ascending) list of week offsets that contain tasks.
func (m *Model) weekOffsetsWithTasks() []int {
	if len(m.indexedAll) == 0 {
		return nil
	}
	nowMonday, _ := timeutil.GetWeekBounds(time.Now())
	seen := map[int]bool{}
	for _, t := range m.indexedAll {
		d, ok := timeutil.ParseDate(t.Date)
		if !ok {
			continue
		}
		taskMonday, _ := timeutil.GetWeekBounds(d)
		days := int(taskMonday.Sub(nowMonday).Hours() / 24)
		offset := days / 7
		seen[offset] = true
	}
	offsets := make([]int, 0, len(seen))
	for o := range seen {
		offsets = append(offsets, o)
	}
	sort.Ints(offsets)
	return offsets
}

// adjacentWeekOffset finds the next week offset with tasks in the given direction (-1 or +1).
// Returns current offset if no adjacent week with tasks exists.
func (m *Model) adjacentWeekOffset(direction int) int {
	offsets := m.weekOffsetsWithTasks()
	if len(offsets) == 0 {
		return m.weekOffset
	}
	if direction < 0 {
		// Find the largest offset smaller than current
		for i := len(offsets) - 1; i >= 0; i-- {
			if offsets[i] < m.weekOffset {
				return offsets[i]
			}
		}
	} else {
		// Find the smallest offset larger than current
		for _, o := range offsets {
			if o > m.weekOffset {
				return o
			}
		}
	}
	return m.weekOffset
}

// monthOffsetsWithTasks returns sorted (ascending) list of month offsets that contain tasks.
func (m *Model) monthOffsetsWithTasks() []int {
	if len(m.indexedAll) == 0 {
		return nil
	}
	now := time.Now()
	nowYear, nowMonth := now.Year(), now.Month()
	seen := map[int]bool{}
	for _, t := range m.indexedAll {
		d, ok := timeutil.ParseDate(t.Date)
		if !ok {
			continue
		}
		offset := (d.Year()-nowYear)*12 + int(d.Month()-nowMonth)
		seen[offset] = true
	}
	offsets := make([]int, 0, len(seen))
	for o := range seen {
		offsets = append(offsets, o)
	}
	sort.Ints(offsets)
	return offsets
}

// adjacentMonthOffset finds the next month offset with tasks in the given direction (-1 or +1).
func (m *Model) adjacentMonthOffset(direction int) int {
	offsets := m.monthOffsetsWithTasks()
	if len(offsets) == 0 {
		return m.monthOffset
	}
	if direction < 0 {
		for i := len(offsets) - 1; i >= 0; i-- {
			if offsets[i] < m.monthOffset {
				return offsets[i]
			}
		}
	} else {
		for _, o := range offsets {
			if o > m.monthOffset {
				return o
			}
		}
	}
	return m.monthOffset
}

func getField(t model.Task, col string) string {
	switch col {
	case "date":
		return t.Date
	case "type":
		return t.Type
	case "number":
		return t.Number
	case "name":
		return t.Name
	case "timeSpent":
		return t.TimeSpent
	case "comments":
		return t.Comments
	}
	return ""
}

func sumHours(tasks []model.IndexedTask) float64 {
	var total float64
	for _, t := range tasks {
		total += timeutil.ParseTime(t.TimeSpent)
	}
	return total
}

// filterTasks returns tasks where any field contains the filter text (case-insensitive).
func filterTasks(tasks []model.IndexedTask, text string) []model.IndexedTask {
	if text == "" {
		return tasks
	}
	lower := strings.ToLower(text)
	var out []model.IndexedTask
	for _, t := range tasks {
		if strings.Contains(strings.ToLower(t.Date), lower) ||
			strings.Contains(strings.ToLower(t.Type), lower) ||
			strings.Contains(strings.ToLower(t.Number), lower) ||
			strings.Contains(strings.ToLower(t.Name), lower) ||
			strings.Contains(strings.ToLower(t.TimeSpent), lower) ||
			strings.Contains(strings.ToLower(t.Comments), lower) {
			out = append(out, t)
		}
	}
	return out
}

func indexOf(slice []string, val string) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return 0
}

func navigate(screen appTui.Screen) tea.Cmd {
	return func() tea.Msg { return appTui.NavigateMsg{Screen: screen} }
}

func stopTimer() tea.Cmd {
	return func() tea.Msg { return appTui.StopTimerMsg{} }
}

func flash(text string) tea.Cmd {
	return func() tea.Msg { return appTui.FlashMsg{Text: text} }
}
