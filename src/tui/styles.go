package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Header bar style
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6")) // cyan

	// Title text
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6"))

	// Flash message
	FlashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")) // green

	// Update available
	UpdateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")) // yellow

	// Timer display
	TimerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")) // magenta

	// Hint text (dimmed)
	HintStyle = lipgloss.NewStyle().
			Faint(true)

	// Error text
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")) // red

	// Selected/highlighted item
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Reverse(true)

	// Color for Bug type
	BugColor = lipgloss.Color("1") // red

	// Color for Task type
	TaskColor = lipgloss.Color("3") // yellow

	// Table header
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("6"))

	// Active editing cell (deep blue background)
	EditingCellStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("17")).
			Foreground(lipgloss.Color("15")) // white on deep blue

	// Confirm delete
	DeleteConfirmStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("1"))

	// Prompt label
	PromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	// Modal overlay box
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6")).
			Padding(0, 2)

	// Modal title
	ModalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6"))

	// Overtime (positive hours beyond workday)
	OvertimeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green

	// Remaining (hours left to reach workday)
	RemainingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
)

// RemainingLabel returns a styled string showing overtime (+) or remaining (-) hours.
func RemainingLabel(total float64, workday float64) string {
	diff := total - workday
	if diff >= 0 {
		return OvertimeStyle.Render(fmt.Sprintf("+%.1fh", diff))
	}
	return RemainingStyle.Render(fmt.Sprintf("-%.1fh", -diff))
}

// ProgressBar renders an ASCII progress bar like [■■■■■■■■□□] 92%.
func ProgressBar(current, target float64, width int) string {
	if target <= 0 {
		return ""
	}
	pct := current / target
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	bar := fmt.Sprintf("[%s%s] %d%%",
		strings.Repeat("■", filled),
		strings.Repeat("□", width-filled),
		int(pct*100))

	var color lipgloss.Color
	switch {
	case pct >= 1.0:
		color = lipgloss.Color("2") // green
	case pct >= 0.5:
		color = lipgloss.Color("3") // yellow
	default:
		color = lipgloss.Color("1") // red
	}
	return lipgloss.NewStyle().Foreground(color).Render(bar)
}

// TypeColor returns the appropriate color for a task type.
func TypeColor(typ string) lipgloss.Color {
	switch {
	case len(typ) > 0 && (typ[0] == 'B' || typ[0] == 'b'):
		return BugColor
	case len(typ) > 0 && (typ[0] == 'T' || typ[0] == 't'):
		return TaskColor
	default:
		return lipgloss.Color("7") // white
	}
}
