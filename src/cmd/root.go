package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aliaksandrZh/worklog/src/internal/store"
	"github.com/aliaksandrZh/worklog/src/internal/timer"
	appTui "github.com/aliaksandrZh/worklog/src/tui"
	"github.com/aliaksandrZh/worklog/src/tui/summary"
)

func init() {
	appTui.ScreenFactory = func(screen appTui.Screen, repo store.TaskRepository, tmr *timer.Timer) appTui.ScreenModel {
		switch screen {
		case appTui.ScreenSummary:
			return summary.New(repo, tmr)
		default:
			return nil
		}
	}
}

// Execute launches the TUI.
func Execute() {
	app := appTui.NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
