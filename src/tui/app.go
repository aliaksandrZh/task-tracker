package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aliaksandrZh/worklog/src/internal/model"
	"github.com/aliaksandrZh/worklog/src/internal/store"
	"github.com/aliaksandrZh/worklog/src/internal/timer"
	"github.com/aliaksandrZh/worklog/src/internal/update"
)

// ScreenModel is the interface each TUI screen must implement.
type ScreenModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (ScreenModel, tea.Cmd)
	View() string
}

// ScreenFactory creates screen models. Set by the cmd package to avoid import cycles.
var ScreenFactory func(screen Screen, repo store.TaskRepository, tmr *timer.Timer) ScreenModel

// App is the root Bubble Tea model.
type App struct {
	screen        Screen
	activeModel   ScreenModel
	flash         string
	timerInfo     *timer.TimerStatus
	updateCount   int
	repo          store.TaskRepository
	tmr           *timer.Timer
	width, height int
	menuCursor    int
}

// NewApp creates the root app model.
func NewApp() App {
	repo := store.New()
	tmr := timer.New(".")
	return App{
		screen: ScreenMenu,
		repo:   repo,
		tmr:    tmr,
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("Worklog"),
		a.refreshTimer(),
		a.checkUpdates(),
		a.timerTick(),
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Forward to active screen
		if a.activeModel != nil {
			newModel, cmd := a.activeModel.Update(msg)
			a.activeModel = newModel
			return a, cmd
		}

	case tea.KeyMsg:
		if a.screen == ScreenMenu {
			return a.handleMenuKey(msg)
		}
		if a.activeModel != nil {
			newModel, cmd := a.activeModel.Update(msg)
			a.activeModel = newModel
			return a, cmd
		}

	case NavigateMsg:
		return a.navigate(msg.Screen)

	case DoneMsg:
		return a.navigate(ScreenMenu)

	case FlashMsg:
		a.flash = msg.Text
		return a, a.clearFlashAfter(3 * time.Second)

	case clearFlashMsg:
		a.flash = ""

	case TimerTickMsg:
		a.timerInfo = a.tmr.GetStatus()
		return a, a.timerTick()

	case UpdateAvailableMsg:
		a.updateCount = msg.Count

	default:
		// Forward all other messages (blink, etc.) to active screen
		if a.activeModel != nil {
			newModel, cmd := a.activeModel.Update(msg)
			a.activeModel = newModel
			return a, cmd
		}
	}

	return a, nil
}

func (a App) View() string {
	var b strings.Builder

	// Header
	bar := strings.Repeat("━", 40)
	b.WriteString(HeaderStyle.Render(bar) + "\n")
	b.WriteString(HeaderStyle.Render("  Worklog") + "\n")
	b.WriteString(HeaderStyle.Render(bar) + "\n")

	// Update notification
	if a.updateCount > 0 {
		plural := ""
		if a.updateCount > 1 {
			plural = "s"
		}
		b.WriteString(UpdateStyle.Render(
			fmt.Sprintf("Update available (%d commit%s behind). Run: git pull", a.updateCount, plural)) + "\n")
	}

	// Timer display
	if a.timerInfo != nil {
		elapsed := timer.FormatElapsed(time.Now().UnixMilli() - a.timerInfo.StartedAt)
		b.WriteString(TimerStyle.Render(
			fmt.Sprintf("⏱ %s %s: %s — %s", a.timerInfo.Type, a.timerInfo.Number, a.timerInfo.Name, elapsed)) + "\n")
	}

	// Flash message
	if a.flash != "" {
		b.WriteString(FlashStyle.Render(a.flash) + "\n")
	}

	b.WriteString("\n")

	// Active screen
	if a.activeModel != nil {
		b.WriteString(a.activeModel.View())
	} else {
		b.WriteString(a.menuView())
	}

	return b.String()
}

func (a App) menuItems() []struct{ key, label string } {
	items := []struct{ key, label string }{
		{"a", "Add Task"},
		{"p", "Paste Tasks"},
		{"s", "View Summary"},
	}
	if a.timerInfo != nil {
		items = append(items, struct{ key, label string }{"t", "Stop Timer"})
	} else {
		items = append(items, struct{ key, label string }{"t", "Start Timer"})
	}
	items = append(items, struct{ key, label string }{"q", "Exit"})
	return items
}

// handleMenuKey processes keyboard shortcuts and arrow navigation on the menu screen.
func (a App) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := a.menuItems()

	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "a":
		return a.navigate(ScreenAdd)
	case "p":
		return a.navigate(ScreenPaste)
	case "s":
		return a.navigate(ScreenSummary)
	case "t":
		if a.timerInfo != nil {
			return a.stopTimer()
		}
		return a.navigate(ScreenTimerStart)
	case "up", "k":
		a.menuCursor--
		if a.menuCursor < 0 {
			a.menuCursor = len(items) - 1
		}
	case "down", "j":
		a.menuCursor++
		if a.menuCursor >= len(items) {
			a.menuCursor = 0
		}
	case "enter":
		if a.menuCursor >= 0 && a.menuCursor < len(items) {
			key := items[a.menuCursor].key
			switch key {
			case "a":
				return a.navigate(ScreenAdd)
			case "p":
				return a.navigate(ScreenPaste)
			case "s":
				return a.navigate(ScreenSummary)
			case "t":
				if a.timerInfo != nil {
					return a.stopTimer()
				}
				return a.navigate(ScreenTimerStart)
			case "q":
				return a, tea.Quit
			}
		}
	}
	return a, nil
}

func (a App) navigate(screen Screen) (tea.Model, tea.Cmd) {
	a.screen = screen
	a.timerInfo = a.tmr.GetStatus()

	if screen == ScreenMenu {
		a.activeModel = nil
		return a, nil
	}

	if ScreenFactory != nil {
		a.activeModel = ScreenFactory(screen, a.repo, a.tmr)
		if a.activeModel != nil {
			// Send initial WindowSizeMsg so screens know the width
			cmd := a.activeModel.Init()
			newModel, sizeCmd := a.activeModel.Update(tea.WindowSizeMsg{
				Width: a.width, Height: a.height,
			})
			a.activeModel = newModel
			return a, tea.Batch(cmd, sizeCmd)
		}
	}
	return a, nil
}

func (a App) stopTimer() (tea.Model, tea.Cmd) {
	result, err := a.tmr.Stop()
	if err != nil {
		a.flash = err.Error()
		return a, nil
	}

	task := model.Task{
		Date:      fmt.Sprintf("%d/%d/%d", time.Now().Month(), time.Now().Day(), time.Now().Year()),
		Type:      result.Type,
		Number:    result.Number,
		Name:      result.Name,
		TimeSpent: result.TimeSpent,
	}
	_ = a.repo.AddTask(task)
	a.timerInfo = nil
	a.flash = fmt.Sprintf("Timer stopped: %s %s: %s (%s)", task.Type, task.Number, task.Name, task.TimeSpent)
	return a, a.clearFlashAfter(3 * time.Second)
}

func (a App) menuView() string {
	var b strings.Builder
	b.WriteString(HintStyle.Render("Arrow keys + Enter or press shortcut key") + "\n\n")

	items := a.menuItems()
	for i, item := range items {
		cursor := "  "
		if i == a.menuCursor {
			cursor = "> "
		}
		line := fmt.Sprintf("%s(%s) %s", cursor, item.key, item.label)
		if i == a.menuCursor {
			b.WriteString(SelectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}

func (a App) refreshTimer() tea.Cmd {
	return func() tea.Msg {
		return TimerTickMsg{}
	}
}

func (a App) timerTick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return TimerTickMsg{}
	})
}

func (a App) checkUpdates() tea.Cmd {
	return func() tea.Msg {
		count := update.CheckForUpdates()
		return UpdateAvailableMsg{Count: count}
	}
}

type clearFlashMsg struct{}

func (a App) clearFlashAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearFlashMsg{}
	})
}
