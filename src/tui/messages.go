package tui

// Screen identifies which TUI screen is active.
type Screen int

const (
	ScreenMenu Screen = iota
	ScreenAdd
	ScreenPaste
	ScreenSummary
	ScreenTimerStart
)

// NavigateMsg requests a screen transition.
type NavigateMsg struct {
	Screen Screen
}

// DoneMsg signals a sub-screen is finished and wants to return to menu.
type DoneMsg struct{}

// FlashMsg displays a temporary message.
type FlashMsg struct {
	Text string
}

// StopTimerMsg requests the app to stop the running timer.
type StopTimerMsg struct{}

// TimerTickMsg triggers periodic timer refresh.
type TimerTickMsg struct{}

// UpdateAvailableMsg reports how many commits behind.
type UpdateAvailableMsg struct {
	Count int
}
