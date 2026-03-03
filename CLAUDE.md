# Task Tracker TUI

A terminal UI app for tracking daily work tasks (bugs, tasks) with time spent. Replaces plain-text tracking with a proper interface supporting paste, summaries, and editing.

## Tech Stack

- **Node.js + Ink 5.x** (React for CLI) with JSX
- **CSV file** (`tasks.csv`) for storage via `papaparse`
- **tsx** to run JSX/ESM with zero config
- All ESM (`"type": "module"` in package.json)

## Commands

- `npm start` — run the app
- `npm test` — run all tests (uses built-in `node:test`, no extra deps)

## Project Structure

```
task-tracker/
├── src/
│   ├── index.js              # Entry point, renders <App>
│   ├── app.jsx               # Root component, screen router
│   ├── store.js              # CSV CRUD (load, save, add, update, delete)
│   ├── parser.js             # Lenient paste-format parser
│   ├── utils.js              # Shared utilities (parseTime, parseDate, groupByDate, etc.)
│   └── components/
│       ├── MainMenu.jsx      # Main menu (Add, Paste, Summary, Edit/Delete, Exit)
│       ├── AddTask.jsx       # Sequential single-task form
│       ├── PasteTasks.jsx    # Multi-line paste mode with missing field prompts
│       ├── ViewSummary.jsx   # Daily/weekly summaries with navigation
│       ├── EditDelete.jsx    # Task list → select to edit or delete
│       ├── EditForm.jsx      # Pre-populated edit form
│       └── TaskTable.jsx     # Reusable formatted task table
├── tests/
│   ├── patterns.test.js      # Unit tests for each pattern group
│   ├── parser.test.js        # Integration tests for parsePastedText
│   ├── store.test.js         # CSV CRUD tests
│   └── utils.test.js         # Utility function tests (parseTime, parseDate, groupByDate, etc.)
├── package.json
└── .gitignore
```

## CSV Format

File: `tasks.csv` (auto-created on first run, gitignored)

Columns: `date,type,number,name,timeSpent,comments`

## Architecture

### Screen Router (`app.jsx`)

State-based routing: `screen` state controls which component renders. Components receive `onDone` (go back to menu) and `onMessage` (flash message) callbacks.

### Parser (`parser.js`)

Lenient, field-by-field extraction pipeline. Each field has its own pattern array:

- `DATE_PATTERNS` — standalone date lines (`M/D/YYYY`, `YYYY-MM-DD`)
- `TYPE_PATTERNS` — task type at start of line (`Bug`, `Task`)
- `NUMBER_PATTERNS` — task number (`123`, `#123`, `123:`)
- `TIME_PATTERNS` — time at end of line (`1h`, `30m`, `1h 30m`)

Pipeline: type → number → time → name (whatever remains). Unknown fields are reported in a `missing` array so the UI can prompt the user.

To add new patterns, append to the relevant array in `parser.js` and add tests in `tests/patterns.test.js`.

### Paste Mode (`PasteTasks.jsx`)

Three phases: `input` → `fill` → `preview`. After parsing, if any required fields are missing (type, number, name — not timeSpent which is optional), the user is prompted to fill them before saving.

### Store (`store.js`)

Synchronous CSV read/write using `papaparse`. Values are trimmed on load. All functions: `loadTasks`, `saveTasks`, `addTask`, `addTasks`, `updateTask`, `deleteTask`.

## Testing

Tests use Node.js built-in `node:test` + `node:assert/strict`, run via `tsx`.

- `patterns.test.js` — tests each pattern array in isolation (DATE, TYPE, NUMBER, TIME)
- `parser.test.js` — tests `parsePastedText` end-to-end (full format, partial format, mixed blocks, date handling)
- `store.test.js` — tests CSV CRUD in a temp directory
- `utils.test.js` — tests parseTime, parseDate, getWeekBounds, formatDateShort, groupByDate

## Rules

- All new code must be covered with tests. Run `npm test` before committing/pushing.

## Key Design Decisions

- Parser is lenient by design: extracts what it can, prompts for the rest
- `timeSpent` is always optional (not prompted during fill phase)
- Date defaults to today if no date line is provided
- Each pattern group is a separate exported array for easy expansion
