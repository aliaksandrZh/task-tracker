# Worklog

[![Tests](https://github.com/aliaksandrZh/worklog/actions/workflows/test.yml/badge.svg)](https://github.com/aliaksandrZh/worklog/actions/workflows/test.yml)

A terminal UI for tracking daily work tasks (bugs, tasks) with time spent. Replaces plain-text tracking with a proper interface supporting paste, summaries, editing, and a built-in timer.

## Quick Start

### Install

**Prerequisites:** [Node.js](https://nodejs.org/) v18+

**macOS / Linux:**

```bash
git clone https://github.com/aliaksandrZh/worklog.git
cd worklog
chmod +x install.sh
./install.sh
```

**Windows:**

```cmd
git clone https://github.com/aliaksandrZh/worklog.git
cd worklog
install.bat
```

Both scripts install dependencies, link the `tt` command globally, and verify it's on your PATH.

### Manual Install

```bash
npm install
npm install -g .
```

## Usage

Run `tt` with no arguments to launch the interactive TUI, or use subcommands for quick actions:

```bash
tt                              # interactive TUI menu
tt add Bug 12345: Fix login 1h  # add a task instantly
tt paste                        # read clipboard, parse, and save
tt today                        # print today's tasks
tt week                         # print current week's tasks
```

### Timer

Track time on the current task — works from both CLI and TUI:

```bash
tt start Bug 123: Fix login     # start timer
tt status                       # show what's being timed
tt stop                         # stop timer, save task with elapsed time
```

In the TUI, the running timer is displayed in the header with elapsed time, and you can start/stop timers via the menu (`t` shortcut).

### TUI Menu

The interactive TUI provides keyboard shortcuts for quick navigation:

- **(a) Add Task** — sequential form with back-navigation (Backspace on empty field)
- **(p) Paste Tasks** — paste a line, parser extracts fields and prompts for anything missing
- **(s) View Summary** — daily/weekly summaries with date navigation (←/→ arrows)
- **(e) Edit/Delete** — select a task to edit or remove
- **(t) Start/Stop Timer** — paste a task line to start timing
- **(q) Exit**

Optional fields are marked with `?` in forms. The app checks for updates on startup.

## Data Storage

Tasks are stored in `tasks.csv` (auto-created on first run) in the current directory.

| Column    | Description                    |
|-----------|--------------------------------|
| date      | Date of the task (M/D/YYYY)    |
| type      | Bug, Task, etc.                |
| number    | Task/ticket number             |
| name      | Short description              |
| timeSpent | Duration (e.g. 1h, 30m)       |
| comments  | Optional notes                 |

## Paste Format

The parser is lenient — it extracts what it can and prompts for the rest. Supported formats:

```
Bug 12345: Fix login page redirect 1h 30m
Task 67890: Update API docs 45m
Pull Request 19082: Bug 31601: Fix date filter 1.5
```

Recognized patterns:
- **Type** — `Bug`, `Task` at start of line (color-coded: Bug in red, Task in yellow)
- **Number** — `123`, `#123`, `123:`
- **Time** — `1h`, `30m`, `1h 30m`, or bare number like `1.5` (treated as hours)
- **Name** — whatever remains after extracting other fields
- **Pull Request prefix** — `Pull Request XXXXX:` is stripped and saved to comments

## Development

```bash
npm start       # run the app (same as tt)
npm test        # run all tests
```

Tests use Node.js built-in `node:test` and require no extra dependencies. CI runs on Node 18, 20, and 22 via GitHub Actions.

## Tech Stack

- **Node.js + [Ink 5](https://github.com/vadimdemedes/ink)** — React for the terminal
- **CSV** via [PapaParse](https://www.papaparse.com/) — simple, portable storage
- **tsx** — run JSX/ESM with zero config

## License

MIT
