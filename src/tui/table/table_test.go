package table

import (
	"strings"
	"testing"

	"github.com/aliaksandrZh/worklog/src/internal/model"
)

func TestHeaderLabel_Uppercase(t *testing.T) {
	tests := []struct {
		col  string
		want string
	}{
		{"date", "DATE"},
		{"type", "TYPE"},
		{"number", "NUMBER"},
		{"name", "NAME"},
		{"timeSpent", "TIME"},
		{"comments", "COMMENTS"},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		got := HeaderLabel(tt.col)
		if got != tt.want {
			t.Errorf("HeaderLabel(%q) = %q, want %q", tt.col, got, tt.want)
		}
	}
}

func TestRender_SortArrowOnActiveColumn(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Date: "3/13/2026", Type: "Bug", Number: "123", Name: "Test", TimeSpent: "1h"}, Index: 0},
	}
	cfg := Config{
		Width:            100,
		SortBy:           "name",
		SortDir:          "desc",
		SelectedRow:      -1,
		ConfirmDeleteRow: -1,
	}
	out := Render(tasks, cfg)
	lines := strings.Split(out, "\n")
	header := lines[0]

	// Active sort column should have ▼
	if !strings.Contains(header, "NAME ▼") {
		t.Errorf("header should contain 'NAME ▼', got: %s", header)
	}
	// Other sortable columns should NOT have arrows
	if strings.Contains(header, "DATE ▼") || strings.Contains(header, "DATE ▲") || strings.Contains(header, "DATE ▽") {
		t.Errorf("non-active column DATE should not have an arrow, got: %s", header)
	}
}

func TestRender_SortArrowAsc(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Date: "3/13/2026", Type: "Bug", Number: "123", Name: "Test", TimeSpent: "1h"}, Index: 0},
	}
	cfg := Config{
		Width:            100,
		SortBy:           "date",
		SortDir:          "asc",
		SelectedRow:      -1,
		ConfirmDeleteRow: -1,
	}
	out := Render(tasks, cfg)
	lines := strings.Split(out, "\n")
	header := lines[0]

	if !strings.Contains(header, "DATE ▲") {
		t.Errorf("header should contain 'DATE ▲' for asc sort, got: %s", header)
	}
}

func TestRender_NoSortArrowWhenNoSort(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Date: "3/13/2026", Type: "Bug", Number: "123", Name: "Test", TimeSpent: "1h"}, Index: 0},
	}
	cfg := Config{
		Width:            100,
		SortBy:           "",
		SortDir:          "asc",
		SelectedRow:      -1,
		ConfirmDeleteRow: -1,
	}
	out := Render(tasks, cfg)
	lines := strings.Split(out, "\n")
	header := lines[0]

	for _, arrow := range []string{"▲", "▼"} {
		if strings.Contains(header, arrow) {
			t.Errorf("header should have no arrows when sort is off, got: %s", header)
		}
	}
}

func TestRender_ConfirmDeleteRow(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Date: "3/13/2026", Type: "Bug", Number: "123", Name: "Test", TimeSpent: "1h", Comments: "some comment"}, Index: 0},
		{Task: model.Task{Date: "3/13/2026", Type: "Task", Number: "456", Name: "Other", TimeSpent: "2h", Comments: "other"}, Index: 1},
	}
	cfg := Config{
		Width:            100,
		SelectedRow:      1,
		ConfirmDeleteRow: 1,
	}
	out := Render(tasks, cfg)

	// Delete row should show confirmation prompt
	if !strings.Contains(out, "Delete? (y/n)") {
		t.Errorf("delete row should contain 'Delete? (y/n)', got: %s", out)
	}

	// First row should NOT have delete prompt
	lines := strings.Split(out, "\n")
	// line 0 = header, line 1 = first task, line 2 = second task (delete row)
	if strings.Contains(lines[1], "Delete? (y/n)") {
		t.Errorf("non-delete row should not contain delete prompt")
	}
}

func TestRender_Empty(t *testing.T) {
	out := Render(nil, Config{Width: 80})
	if !strings.Contains(out, "No tasks to display") {
		t.Errorf("empty render should show 'No tasks to display', got: %s", out)
	}
}
