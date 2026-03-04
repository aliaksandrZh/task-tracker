# UX Improvements TODO

## 1. Add keyboard shortcuts to main menu
Add single-key shortcuts (a/p/s/e/q) to the main menu as additional navigation alongside existing arrow keys. Users should be able to press a letter to jump to a menu item instantly.

**Plan:**
- Add `useInput` hook from `ink` to listen for key presses
- Map keys: `a` → Add Task, `p` → Paste Tasks, `s`/`v` → View Summary, `e` → Edit/Delete, `q` → Exit
- Show shortcut keys in the labels, e.g. `(a) Add Task`
- Keep `SelectInput` and arrow key navigation as-is — shortcuts are additive
- Update help text to mention shortcut keys

**Files:** `src/components/MainMenu.jsx`

---

## 2. Add back navigation in sequential forms
Allow users to press Backspace on an empty field to go back to the previous field and fix it, instead of having to cancel and start over. Applies to AddTask and EditForm components.

**Plan:**
- In `useInput`, detect Backspace/Delete when `input` is empty → go back one step
- Restore the previous field's value into the text input so user can edit it
- Show hint text: `"Backspace on empty to go back"`
- First field (date) — ignore Backspace (nowhere to go)

**Files:** `src/components/AddTask.jsx`, `src/components/EditForm.jsx`

---

## 3. Add timer support to TUI
Integrate the timer (start/stop/status) into the TUI. Currently timer is CLI-only.

**Plan:**
- **Timer status in header** (`app.jsx`): poll `getTimerStatus()` on interval (every 30s), display running timer info below title bar, e.g. `"⏱ Bug 123: Fix login — 45m"`
- **New menu items** in `MainMenu.jsx`:
  - "Start Timer" → navigates to new `TimerStart` screen (form: type, number, name)
  - "Stop Timer" → stops timer, saves task, shows confirmation
- **New component: `TimerStart.jsx`** — 3-field form (type, number, name), calls `startTimer()`
- **Stop logic** — in `app.jsx` handler: call `stopTimer()`, then `addTask()` with result, show message
- **Conditional menu** — show "Start Timer" when no timer running, "Stop Timer" when one is

**Files:** `src/app.jsx`, `src/components/MainMenu.jsx`, new `src/components/TimerStart.jsx`

---

## 4. Mark optional fields with ? in form prompts
In AddTask and EditForm, mark optional fields (timeSpent, comments) with a "?" indicator so users know they can skip them.

**Plan:**
- In `AddTask.jsx` FIELDS array, update labels: `"Time Spent? (e.g. 1h, 30m)"`, `"Comments?"`
- In `EditForm.jsx`, add field label mapping with `?` markers for timeSpent and comments
- Update hint text to mention "? fields are optional"

**Files:** `src/components/AddTask.jsx`, `src/components/EditForm.jsx`

---

## 5. Show paste format hints in paste input screen
Display example input formats in the PasteTasks input phase so users know what patterns are recognized.

**Plan:**
- In the `input` phase, add dimmed example lines below instructions:
  ```
  Examples:  Bug 123: Fix login 1h
             Task 456: Add feature 2h 30m
             3/4/2026           ← date line (applies to tasks below)
  ```
- Keep it compact — 3-4 example lines max

**Files:** `src/components/PasteTasks.jsx`

---

## 6. Add date navigation to Summary view
Allow ←/→ arrow keys in ViewSummary to browse past/future days (daily mode) or weeks (weekly mode).

**Plan:**
- Add `offset` state (default 0). `←` decrements, `→` increments (capped at 0 = current)
- **Daily mode**: show one date at a time, navigate through `dailyGroups` array by index
- **Weekly mode**: shift week by 7 days. Add `filterWeekByOffset(tasks, offset)` to `utils.js`
- Show navigation hints: `"← prev | → next | d = daily | w = weekly | Escape = back"`
- Add test for `filterWeekByOffset`

**Files:** `src/components/ViewSummary.jsx`, `src/utils.js`, `tests/utils.test.js`
