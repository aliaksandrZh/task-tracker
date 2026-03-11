package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aliaksandrZh/worklog/src/internal/model"
	"github.com/aliaksandrZh/worklog/src/internal/store"
	"github.com/aliaksandrZh/worklog/src/internal/timer"
	"github.com/aliaksandrZh/worklog/src/internal/update"
)

const (
	flashDuration       = 5 * time.Second
	flashUpdateDuration = 15 * time.Second
)

// ScreenModel is the interface each TUI screen must implement.
type ScreenModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (ScreenModel, tea.Cmd)
	View() string
}

// Reloadable is implemented by screens that can refresh their data.
type Reloadable interface {
	Reload()
}

// ScreenFactory creates screen models. Set by the cmd package to avoid import cycles.
var ScreenFactory func(screen Screen, repo store.TaskRepository, tmr *timer.Timer) ScreenModel

// App is the root Bubble Tea model.
type App struct {
	screen        Screen
	activeModel   ScreenModel // current sub-screen (add, paste, timerstart) or nil
	homeModel     ScreenModel // the summary screen (always alive)
	flash         string
	timerInfo     *timer.TimerStatus
	updateCount   int
	repo          store.TaskRepository
	tmr           *timer.Timer
	width, height int
}

// NewApp creates the root app model.
func NewApp() App {
	repo := store.New()
	tmr := timer.New(".")
	app := App{
		screen:      ScreenSummary,
		repo:        repo,
		tmr:         tmr,
	}
	if ScreenFactory != nil {
		app.homeModel = ScreenFactory(ScreenSummary, repo, tmr)
	}
	return app
}

func (a App) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.SetWindowTitle("Worklog"),
		a.refreshTimer(),
		a.checkUpdates(),
		a.timerTick(),
	}
	if a.homeModel != nil {
		cmds = append(cmds, a.homeModel.Init())
	}
	return tea.Batch(cmds...)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Forward to active sub-screen or home
		target := a.currentModel()
		if target != nil {
			newModel, cmd := target.Update(msg)
			a.setCurrentModel(newModel)
			return a, cmd
		}

	case tea.KeyMsg:
		// Handle update shortcut at app level before forwarding
		if msg.String() == "u" && a.updateCount > 0 && a.activeModel == nil {
			cmd := "git pull && go build -o tt ."
			copyToClipboard(cmd)
			a.flash = "Copied to clipboard: " + cmd
			return a, a.clearFlashAfter(flashUpdateDuration)
		}
		target := a.currentModel()
		if target != nil {
			newModel, cmd := target.Update(msg)
			a.setCurrentModel(newModel)
			return a, cmd
		}

	case NavigateMsg:
		return a.navigate(msg.Screen)

	case DoneMsg:
		// Return to summary (home)
		return a.returnHome()

	case StopTimerMsg:
		return a.stopTimer()

	case FlashMsg:
		a.flash = msg.Text
		return a, a.clearFlashAfter(flashDuration)

	case clearFlashMsg:
		a.flash = ""

	case TimerTickMsg:
		a.timerInfo = a.tmr.GetStatus()
		return a, a.timerTick()

	case UpdateAvailableMsg:
		a.updateCount = msg.Count

	default:
		target := a.currentModel()
		if target != nil {
			newModel, cmd := target.Update(msg)
			a.setCurrentModel(newModel)
			return a, cmd
		}
	}

	return a, nil
}

// currentModel returns the active sub-screen, or the home model if on summary.
func (a *App) currentModel() ScreenModel {
	if a.activeModel != nil {
		return a.activeModel
	}
	return a.homeModel
}

// setCurrentModel updates the active sub-screen or home model.
func (a *App) setCurrentModel(m ScreenModel) {
	if a.activeModel != nil {
		a.activeModel = m
	} else {
		a.homeModel = m
	}
}

func (a App) View() string {
	var b strings.Builder

	// Active screen or home
	target := a.currentModel()
	if target != nil {
		b.WriteString(target.View())
	}

	// Footer: fixed 3-line area (timer, flash, update) — always rendered to prevent shifting
	b.WriteString("\n")

	if a.timerInfo != nil {
		elapsed := timer.FormatElapsed(time.Now().UnixMilli() - a.timerInfo.StartedAt)
		b.WriteString(TimerStyle.Render(
			fmt.Sprintf("⏱ %s %s: %s — %s", a.timerInfo.Type, a.timerInfo.Number, a.timerInfo.Name, elapsed)) + "\n")
	} else {
		b.WriteString("\n")
	}

	if a.flash != "" {
		b.WriteString(FlashStyle.Render(a.flash) + "\n")
	} else {
		b.WriteString("\n")
	}

	if a.updateCount > 0 {
		plural := ""
		if a.updateCount > 1 {
			plural = "s"
		}
		b.WriteString(UpdateStyle.Render(
			fmt.Sprintf("Update available (%d commit%s behind). u=copy update command", a.updateCount, plural)) + "\n")
	} else {
		b.WriteString("\n")
	}

	return b.String()
}

func (a App) navigate(screen Screen) (tea.Model, tea.Cmd) {
	a.timerInfo = a.tmr.GetStatus()

	if screen == ScreenSummary {
		return a.returnHome()
	}

	a.screen = screen
	if ScreenFactory != nil {
		a.activeModel = ScreenFactory(screen, a.repo, a.tmr)
		if a.activeModel != nil {
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

func (a App) returnHome() (tea.Model, tea.Cmd) {
	a.screen = ScreenSummary
	a.activeModel = nil
	a.timerInfo = a.tmr.GetStatus()

	// Reload summary data to pick up any changes from sub-screens
	if r, ok := a.homeModel.(Reloadable); ok {
		r.Reload()
	}
	// Forward current window size
	if a.homeModel != nil {
		newModel, cmd := a.homeModel.Update(tea.WindowSizeMsg{
			Width: a.width, Height: a.height,
		})
		a.homeModel = newModel
		return a, cmd
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

	// Reload summary to show the new task
	if r, ok := a.homeModel.(Reloadable); ok {
		r.Reload()
	}
	return a, a.clearFlashAfter(flashDuration)
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

func copyToClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "windows":
		cmd = exec.Command("clip")
	default:
		return
	}
	cmd.Stdin = strings.NewReader(text)
	_ = cmd.Run()
}

type clearFlashMsg struct{}

func (a App) clearFlashAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearFlashMsg{}
	})
}
