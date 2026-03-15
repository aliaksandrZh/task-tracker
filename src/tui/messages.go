package tui

// Screen identifies which TUI screen is active.
type Screen int

const (
	ScreenSummary Screen = iota
)

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

// Notifications holds data for the notification zone rendered by the active screen.
type Notifications struct {
	TimerLine string // pre-formatted timer status line (empty if no timer)
	FlashLine string // pre-formatted flash message (empty if none)
	UpdateLine string // pre-formatted update notice (empty if none)
}

// NotificationReceiver is implemented by screens that render the notification zone.
type NotificationReceiver interface {
	SetNotifications(n Notifications)
}
