package summary

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/aliaksandrZh/worklog/src/internal/model"
	"github.com/aliaksandrZh/worklog/src/internal/prefs"
	"github.com/aliaksandrZh/worklog/src/internal/store"
	"github.com/aliaksandrZh/worklog/src/internal/timeutil"
	"github.com/aliaksandrZh/worklog/src/internal/timer"
	appTui "github.com/aliaksandrZh/worklog/src/tui"
	"github.com/aliaksandrZh/worklog/src/tui/table"
)

var sortColumns = []string{"", "date", "type", "number", "name", "timeSpent"}

type phase int

const (
	phaseView phase = iota
	phaseSelect
	phaseEditing
	phaseConfirmDelete
)

// Model is the summary screen.
type Model struct {
	mode       string // "daily" or "weekly"
	phase      phase
	dailyIdx   int
	weekOffset int

	sortBy  string
	sortDir string

	selectedRow int
	selectedCol int
	editInput   textinput.Model
	editInline  bool // true = edit in cell, false = edit below table

	repo  store.TaskRepository
	tmr   *timer.Timer
	prefs *prefs.Store

	allTasks     []model.Task
	indexedAll   []model.IndexedTask
	dailyGroups  []timeutil.DateGroup
	displayed    []model.IndexedTask
	weeklyGroups []timeutil.DateGroup // day groups within current week

	width int
}

// New creates a new summary model.
func New(repo store.TaskRepository, tmr *timer.Timer) *Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 200

	p := prefs.New(".")
	pref := p.Load()

	m := &Model{
		mode:      "daily",
		phase:     phaseView,
		sortBy:    pref.SortBy,
		sortDir:   pref.SortDir,
		editInput: ti,
		repo:      repo,
		tmr:       tmr,
		prefs:     p,
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
	if m.mode == "weekly" {
		result := timeutil.FilterWeekByOffset(m.indexedAll, m.weekOffset)
		m.weeklyGroups = timeutil.GroupByDate(result.Tasks)
		// Build flat displayed list: each day group sorted independently
		m.displayed = nil
		for _, g := range m.weeklyGroups {
			m.displayed = append(m.displayed, timeutil.SortTasks(g.Tasks, m.sortBy, m.sortDir)...)
		}
	} else {
		m.weeklyGroups = nil
		var raw []model.IndexedTask
		if m.dailyIdx < len(m.dailyGroups) {
			raw = m.dailyGroups[m.dailyIdx].Tasks
		}
		m.displayed = timeutil.SortTasks(raw, m.sortBy, m.sortDir)
	}
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
		return m, nil

	case tea.KeyMsg:
		switch m.phase {
		case phaseEditing:
			return m.updateEditing(msg)
		case phaseConfirmDelete:
			return m.updateConfirmDelete(msg)
		case phaseSelect:
			return m.updateSelect(msg)
		default:
			return m.updateView(msg)
		}
	}

	// Forward non-key messages (blink cursor, etc.) to textinput when editing
	if m.phase == phaseEditing {
		var cmd tea.Cmd
		m.editInput, cmd = m.editInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) updateView(msg tea.KeyMsg) (appTui.ScreenModel, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "a":
		return m, navigate(appTui.ScreenAdd)
	case "p":
		return m, navigate(appTui.ScreenPaste)
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

			m.editInput.SetValue(val)
			m.editInput.SetCursor(len(val))

			// If value fits in column, edit inline; otherwise show modal
			if len([]rune(val)) < colW-1 {
				m.editInline = true
				m.editInput.Width = colW
			} else {
				m.editInline = false
				// Modal inner width: terminal - 4 (outer margin) - 6 (border + padding)
				modalInner := m.width - 10
				if modalInner < 30 {
					modalInner = 30
				}
				m.editInput.Width = modalInner
			}

			m.editInput.Focus()
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
		m.phase = phaseSelect
		return m, nil
	case tea.KeyEnter:
		task := m.displayed[m.selectedRow]
		col := table.Columns[m.selectedCol]
		_ = m.repo.UpdateTask(task.Index, map[string]string{col: m.editInput.Value()})
		m.reload()
		m.phase = phaseSelect
		return m, flash(fmt.Sprintf("Updated %s.", col))
	}

	// Forward all other keys to the textinput for typing
	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
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
		return "t=stop timer"
	}
	return "t=timer"
}

func (m *Model) View() string {
	if len(m.allTasks) == 0 && len(m.dailyGroups) <= 1 {
		return fmt.Sprintf("No tasks yet. a=add | p=paste | %s | q=quit\n", m.timerHint())
	}

	var b strings.Builder
	selectedLineY := -1 // track Y position of the selected row

	isEdit := m.phase != phaseView
	editLabel := ""
	if isEdit {
		editLabel = " (EDIT)"
	}

	sortHint := "off"
	if m.sortBy != "" {
		sortHint = m.sortBy + " " + m.sortDir
	}
	viewHint := fmt.Sprintf("a=add | p=paste | %s | ← → nav | e=edit | d=daily | w=weekly | s=sort(%s) | q=quit", m.timerHint(), sortHint)
	editHint := fmt.Sprintf("↑↓=row | ←→=col | Enter=edit | x=delete | s=sort(%s) | S=flip | e/Esc=back", sortHint)

	w := m.width
	if w <= 0 {
		w = 80
	}

	if m.mode == "weekly" {
		result := timeutil.FilterWeekByOffset(m.indexedAll, m.weekOffset)
		b.WriteString(appTui.TitleStyle.Render(fmt.Sprintf("Weekly Summary%s", editLabel)) + "\n")
		if isEdit {
			b.WriteString(appTui.HintStyle.Render(editHint) + "\n")
		} else {
			b.WriteString(appTui.HintStyle.Render(viewHint) + "\n")
		}
		b.WriteString("\n")
		b.WriteString(appTui.PromptStyle.Render(
			fmt.Sprintf("%s — %.1fh total (%d tasks)", result.Label, result.Total, len(result.Tasks))) + "\n")

		// Render day-by-day sections
		rowOffset := 0
		for _, g := range m.weeklyGroups {
			b.WriteString("\n")
			b.WriteString(appTui.PromptStyle.Render(
				fmt.Sprintf("%s — %.1fh total (%d tasks)", g.Key, g.Total, len(g.Tasks))))
			b.WriteString(" " + appTui.RemainingLabel(g.Total, 8) + "\n")

			sorted := timeutil.SortTasks(g.Tasks, m.sortBy, m.sortDir)
			cfg := table.Config{
				Width:       w,
				SortBy:      m.sortBy,
				SortDir:     m.sortDir,
				SelectedRow: -1,
				SelectedCol: m.selectedCol,
			}
			if isEdit {
				localRow := m.selectedRow - rowOffset
				if localRow >= 0 && localRow < len(sorted) {
					cfg.SelectedRow = localRow
					if m.phase == phaseEditing && m.editInline {
						cfg.EditingCell = true
						cfg.EditView = m.editInput.View()
					}
					// line Y = lines so far + table header(1) + localRow
					selectedLineY = countLines(b.String()) + 1 + localRow
				}
			}
			b.WriteString(table.Render(sorted, cfg) + "\n")
			rowOffset += len(g.Tasks)
		}
	} else {
		dateLabel := ""
		if m.dailyIdx < len(m.dailyGroups) {
			dateLabel = " — " + m.dailyGroups[m.dailyIdx].Key
		}
		b.WriteString(appTui.TitleStyle.Render(fmt.Sprintf("Daily Summary%s%s", editLabel, dateLabel)) + "\n")
		if isEdit {
			b.WriteString(appTui.HintStyle.Render(editHint) + "\n")
		} else {
			b.WriteString(appTui.HintStyle.Render(viewHint) + "\n")
		}
		b.WriteString("\n")
		if m.dailyIdx < len(m.dailyGroups) {
			g := m.dailyGroups[m.dailyIdx]
			b.WriteString(appTui.PromptStyle.Render(
				fmt.Sprintf("%s — %.1fh total (%d tasks)", g.Key, g.Total, len(g.Tasks))))
			b.WriteString(" " + appTui.RemainingLabel(g.Total, 8) + "\n")
		}

		cfg := table.Config{
			Width:       w,
			SortBy:      m.sortBy,
			SortDir:     m.sortDir,
			SelectedRow: -1,
			SelectedCol: m.selectedCol,
		}
		if isEdit {
			cfg.SelectedRow = m.selectedRow
			if m.phase == phaseEditing && m.editInline {
				cfg.EditingCell = true
				cfg.EditView = m.editInput.View()
			}
			// line Y = lines so far + table header(1) + selectedRow
			selectedLineY = countLines(b.String()) + 1 + m.selectedRow
		}
		b.WriteString(table.Render(m.displayed, cfg) + "\n")
	}

	if m.phase == phaseConfirmDelete {
		b.WriteString(appTui.DeleteConfirmStyle.Render("Delete this task? (y/n)") + "\n")
	}

	output := b.String()

	// Insert modal right after the selected row
	if m.phase == phaseEditing && !m.editInline && selectedLineY >= 0 {
		col := table.Columns[m.selectedCol]
		label := table.HeaderLabel(col)

		modalW := w - 4
		if modalW < 40 {
			modalW = 40
		}

		var mb strings.Builder
		mb.WriteString(appTui.ModalTitleStyle.Render(fmt.Sprintf("Edit %s", label)) + "\n\n")
		mb.WriteString(m.editInput.View() + "\n\n")
		mb.WriteString(appTui.HintStyle.Render("Enter=save | Escape=cancel"))

		modal := appTui.ModalStyle.Width(modalW).Render(mb.String())

		lines := strings.Split(output, "\n")
		insertAt := selectedLineY + 1
		if insertAt > len(lines) {
			insertAt = len(lines)
		}

		var result []string
		result = append(result, lines[:insertAt]...)
		result = append(result, modal)
		result = append(result, lines[insertAt:]...)
		output = strings.Join(result, "\n")
	}

	return output
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
