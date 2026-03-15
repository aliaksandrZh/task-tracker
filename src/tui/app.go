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
	homeModel     ScreenModel
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
	tmr := timer.New(store.DataDir())
	app := App{
		repo: repo,
		tmr:  tmr,
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
		if a.homeModel != nil {
			newModel, cmd := a.homeModel.Update(msg)
			a.homeModel = newModel
			return a, cmd
		}

	case tea.KeyMsg:
		// Handle update shortcut at app level before forwarding
		if msg.String() == "u" && a.updateCount > 0 {
			cmd := "git pull && go build -o tt ."
			copyToClipboard(cmd)
			a.flash = "Copied to clipboard: " + cmd
			return a, a.clearFlashAfter(flashUpdateDuration)
		}
		if a.homeModel != nil {
			newModel, cmd := a.homeModel.Update(msg)
			a.homeModel = newModel
			return a, cmd
		}

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
		if a.homeModel != nil {
			newModel, cmd := a.homeModel.Update(msg)
			a.homeModel = newModel
			return a, cmd
		}
	}

	return a, nil
}

func (a App) View() string {
	// Pass notification data to the screen
	if nr, ok := a.homeModel.(NotificationReceiver); ok {
		nr.SetNotifications(a.buildNotifications())
	}

	if a.homeModel != nil {
		return a.homeModel.View()
	}
	return ""
}

func (a App) buildNotifications() Notifications {
	var n Notifications

	if a.timerInfo != nil {
		elapsed := timer.FormatElapsed(time.Now().UnixMilli() - a.timerInfo.StartedAt)
		n.TimerLine = TimerStyle.Render(
			fmt.Sprintf("⏱ %s %s: %s — %s", a.timerInfo.Type, a.timerInfo.Number, a.timerInfo.Name, elapsed))
	}

	if a.flash != "" {
		n.FlashLine = FlashStyle.Render(a.flash)
	}

	if a.updateCount > 0 {
		plural := ""
		if a.updateCount > 1 {
			plural = "s"
		}
		n.UpdateLine = UpdateStyle.Render(
			fmt.Sprintf("Update available (%d commit%s behind). u=copy update command", a.updateCount, plural))
	}

	return n
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
