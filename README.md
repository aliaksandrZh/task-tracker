# Worklog

[![Tests](https://github.com/aliaksandrZh/worklog/actions/workflows/test.yml/badge.svg)](https://github.com/aliaksandrZh/worklog/actions/workflows/test.yml)

A terminal UI for tracking daily work tasks (bugs, tasks) with time spent. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea) for native terminal rendering with zero flicker.

## Quick Start

**Prerequisites:** [Go](https://go.dev/) 1.21+

### macOS / Linux

```bash
git clone https://github.com/aliaksandrZh/worklog.git
cd worklog
go build -o tt .
./tt
```

To make it available globally:

```bash
sudo cp tt /usr/local/bin/
```

### Windows

```powershell
git clone https://github.com/aliaksandrZh/worklog.git
cd worklog
go build -o tt.exe .
.\tt.exe
```

To make it available globally, move `tt.exe` to a directory in your `PATH`, or add the current directory to `PATH`:

```powershell
# Option 1: copy to a directory already in PATH
copy tt.exe C:\Users\%USERNAME%\go\bin\

# Option 2: add current directory to PATH (current session)
$env:PATH += ";$(Get-Location)"
```

## Usage

Run `tt` to launch the interactive TUI.

### TUI Menu

Navigate with arrow keys + Enter, or press shortcut keys:

- **(a) Add Task** — sequential form (Backspace on empty field goes back, Escape cancels)
- **(p) Paste Tasks** — type/paste a task line, parser extracts fields and prompts for anything missing
- **(s) View Summary** — daily/weekly summaries with date navigation
- **(t) Start/Stop Timer** — type a task line to start timing; stop saves with elapsed time
- **(q) Exit**

### View Summary

- **d** / **w** — switch between daily and weekly view
- **←** / **→** — navigate between periods
- **s** — cycle sort column (date, type, number, name, time)
- **S** — flip sort direction (asc/desc)
- **e** — enter edit mode
- **Esc** — go back

### Edit Mode (in Summary)

- **↑↓** — select row
- **←→** — select column
- **Enter** — edit the selected cell inline
- **x** — delete the selected task (with y/n confirmation)
- **Esc** — exit edit mode

### Timer

The running timer is displayed in the header with live elapsed time. Start from the menu (`t`), stop to save the task automatically.

## Custom Data Directory

By default, `tt` stores data (`tasks.csv`, `.timer.json`) in the current directory. To run `tt` from anywhere, set the `WORKLOG_DIR` environment variable:

```bash
# In ~/.zshrc or ~/.bashrc
export WORKLOG_DIR="$HOME/worklog"
alias tt="/usr/local/bin/tt"
```

Or inline with the alias:

```bash
alias tt="WORKLOG_DIR=$HOME/worklog /usr/local/bin/tt"
```

## Data Storage

Tasks are stored in `tasks.csv` (auto-created on first run) in the data directory (`WORKLOG_DIR` or current directory).

| Column    | Description                    |
|-----------|--------------------------------|
| date      | Date of the task (M/D/YYYY)    |
| type      | Bug, Task, etc.                |
| number    | Task/ticket number             |
| name      | Short description              |
| timeSpent | Duration (e.g. 1h, 30m)        |
| comments  | Optional notes                 |

## Paste Format

The parser is lenient — it extracts what it can and prompts for the rest:

```
Bug 12345: Fix login page redirect 1h 30m
Task 67890: Update API docs 45m
Pull Request 19082: Bug 31601: Fix date filter 1.5
```

Recognized patterns:
- **Type** — `Bug`, `Task` at start of line (color-coded: Bug=red, Task=yellow)
- **Number** — `123`, `#123`, `123:`
- **Time** — `1h`, `30m`, `1h 30m`, or bare number like `1.5` (treated as hours)
- **Name** — whatever remains after extracting other fields
- **Pull Request prefix** — `Pull Request XXXXX:` is stripped and saved to comments

## Development

```bash
go build -o tt .    # build
go test ./...       # run all tests
```

## Project Structure

```
main.go                     # Entry point
src/
  cmd/root.go               # Wires screen factory, launches TUI
  internal/
    model/                  # Task, IndexedTask, ParsedTask structs
    store/                  # CSV CRUD (Load, Save, Add, Update, Delete)
    parser/                 # Lenient parser (patterns + pipeline)
    timer/                  # .timer.json persistence
    prefs/                  # .prefs.json persistence
    timeutil/               # ParseTime, ParseDate, GroupByDate, etc.
    format/                 # Pad, column width constants
    update/                 # Git-based update check
  tui/
    app.go                  # Root model (screen router, flash, timer tick)
    styles.go               # Lip Gloss styles
    messages.go             # Screen enum, message types
    table/                  # Reusable table renderer
    addtask/                # Sequential 6-field form
    paste/                  # 3-phase: input → fill → preview
    summary/                # Daily/weekly view + inline edit mode
    timerstart/             # Single input for timer
```

## Tech Stack

- **Go** — single binary, cross-platform
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — TUI framework
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** — terminal styling
- **[Bubbles](https://github.com/charmbracelet/bubbles)** — text input component
- **CSV** — simple, portable storage

## License

MIT
