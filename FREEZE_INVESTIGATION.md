# App Freeze Investigation (Windows)

## Issue
App becomes unresponsive after 1-2 hours of idle. Keys (a, e, q) stop working. Only Ctrl+C works. Reported on Windows.

## Root Cause (most likely)
Windows console mode reset after system sleep/idle.

Bubble Tea sets `ENABLE_VIRTUAL_TERMINAL_INPUT` on the console handle once at startup (`tty_windows.go`). After Windows sleep/idle/background, the console can reset these flags, reverting to "cooked mode" where individual keystrokes aren't sent to the app. Only Ctrl+C works because it's handled at the OS level, not through VT input.

On Windows, `suspendSupported = false` in Bubble Tea, so there's no SIGCONT-style recovery path (unlike Unix).

## Not caused by recent code changes
- The `timerTick` (30s interval) was already running before the summary-as-home-screen change
- The `textinput` blink only activates during edit mode, not while idle
- No new goroutines, blocking calls, or loops were introduced
- Commit `8652751` (make summary the home screen) is structural, not input-related

## Relevant files
- Bubble Tea TTY init: `~/go/pkg/mod/github.com/charmbracelet/bubbletea@v1.3.10/tty_windows.go`
- App init/ticks: `src/tui/app.go` (lines 66-77, 273-277)
- Already on latest Bubble Tea (v1.3.10)

## Possible fixes
1. **File upstream issue** on github.com/charmbracelet/bubbletea for Windows console mode restoration
2. **Workaround**: Add a periodic (60s) tick that re-applies Windows console mode flags via `golang.org/x/sys/windows` — platform-specific file, only runs on Windows
3. **Alternative**: Ask the reporter which terminal they use (Windows Terminal, PowerShell, cmd.exe) — some handle idle/sleep better than others
