# Task Tracker

[![Tests](https://github.com/aliaksandrZh/task-tracker/actions/workflows/test.yml/badge.svg)](https://github.com/aliaksandrZh/task-tracker/actions/workflows/test.yml)

A terminal UI for tracking daily work tasks (bugs, tasks) with time spent. Replaces plain-text tracking with a proper interface supporting paste, summaries, editing, and a built-in timer.

## Quick Start

### Install

**Prerequisites:** [Node.js](https://nodejs.org/) v18+

**macOS / Linux:**

```bash
git clone https://github.com/aliaksandrZh/task-tracker.git
cd task-tracker
chmod +x install.sh
./install.sh
```

**Windows:**

```cmd
git clone https://github.com/aliaksandrZh/task-tracker.git
cd task-tracker
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

Track time on the current task:

```bash
tt start Bug 123: Fix login     # start timer
tt status                       # show what's being timed
tt stop                         # stop timer, save task with elapsed time
```

### TUI Menu

The interactive TUI provides:

- **Add Task** — sequential form for a single task
- **Paste Tasks** — paste multiple lines, parser extracts fields automatically and prompts for anything missing
- **View Summary** — daily/weekly summaries with navigation
- **Edit/Delete** — select a task to edit or remove

## Data Storage

Tasks are stored in `tasks.csv` (auto-created on first run) in the current directory.

| Column    | Description                    |
|-----------|--------------------------------|
| date      | Date of the task (YYYY-MM-DD)  |
| type      | Bug, Task, etc.                |
| number    | Task/ticket number             |
| name      | Short description              |
| timeSpent | Duration (e.g. 1h, 30m)       |
| comments  | Optional notes                 |

## Paste Format

The parser is lenient — it extracts what it can and prompts for the rest. Supported formats:

```
3/4/2026
Bug 12345: Fix login page redirect 1h 30m
Task 67890: Update API docs 45m
```

Recognized patterns:
- **Date** — `M/D/YYYY`, `YYYY-MM-DD` (defaults to today if omitted)
- **Type** — `Bug`, `Task` at start of line
- **Number** — `123`, `#123`, `123:`
- **Time** — `1h`, `30m`, `1h 30m` at end of line
- **Name** — whatever remains after extracting other fields

## Development

```bash
npm start       # run the app (same as tt)
npm test        # run all tests
```

Tests use Node.js built-in `node:test` and require no extra dependencies.

## Tech Stack

- **Node.js + [Ink 5](https://github.com/vadimdemedes/ink)** — React for the terminal
- **CSV** via [PapaParse](https://www.papaparse.com/) — simple, portable storage
- **tsx** — run JSX/ESM with zero config

## License

MIT
