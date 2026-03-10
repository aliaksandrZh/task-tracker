package timeutil

import (
	"testing"
	"time"

	"github.com/aliaksandrZh/worklog/src/internal/model"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"", 0},
		{"1h", 1},
		{"30m", 0.5},
		{"1h 30m", 1.5},
		{"1.5h", 1.5},
		{"2h", 2},
		{"45m", 0.75},
	}
	for _, tt := range tests {
		got := ParseTime(tt.input)
		if got != tt.want {
			t.Errorf("ParseTime(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseDate(t *testing.T) {
	d, ok := ParseDate("3/5/2026")
	if !ok {
		t.Fatal("ParseDate should succeed")
	}
	if d.Month() != 3 || d.Day() != 5 || d.Year() != 2026 {
		t.Errorf("got %v", d)
	}

	_, ok = ParseDate("invalid")
	if ok {
		t.Error("should fail for invalid")
	}
}

func TestGetWeekBounds(t *testing.T) {
	// Wednesday March 5, 2026
	d := time.Date(2026, 3, 5, 12, 0, 0, 0, time.Local)
	monday, sunday := GetWeekBounds(d)
	if monday.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %v", monday.Weekday())
	}
	if sunday.Weekday() != time.Sunday {
		t.Errorf("expected Sunday, got %v", sunday.Weekday())
	}
	if monday.Day() != 2 {
		t.Errorf("expected March 2, got %d", monday.Day())
	}
	if sunday.Day() != 8 {
		t.Errorf("expected March 8, got %d", sunday.Day())
	}
}

func TestFormatDateShort(t *testing.T) {
	d := time.Date(2026, 3, 5, 0, 0, 0, 0, time.Local)
	got := FormatDateShort(d)
	if got != "3/5/2026" {
		t.Errorf("got %q", got)
	}
}

func TestGroupByDate(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Date: "3/5/2026", TimeSpent: "1h"}, Index: 0},
		{Task: model.Task{Date: "3/5/2026", TimeSpent: "2h"}, Index: 1},
		{Task: model.Task{Date: "3/4/2026", TimeSpent: "30m"}, Index: 2},
	}
	groups := GroupByDate(tasks)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	// Most recent first
	if groups[0].Key != "3/5/2026" {
		t.Errorf("first group should be 3/5/2026, got %s", groups[0].Key)
	}
	if groups[0].Total != 3.0 {
		t.Errorf("expected 3.0h total, got %v", groups[0].Total)
	}
	if len(groups[0].Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(groups[0].Tasks))
	}
}

func TestSortTasks(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Date: "3/5/2026", Name: "B"}, Index: 0},
		{Task: model.Task{Date: "3/4/2026", Name: "A"}, Index: 1},
	}

	sorted := SortTasks(tasks, "name", "asc")
	if sorted[0].Name != "A" {
		t.Errorf("expected A first, got %s", sorted[0].Name)
	}

	sorted = SortTasks(tasks, "name", "desc")
	if sorted[0].Name != "B" {
		t.Errorf("expected B first, got %s", sorted[0].Name)
	}

	sorted = SortTasks(tasks, "date", "asc")
	if sorted[0].Date != "3/4/2026" {
		t.Errorf("expected 3/4 first, got %s", sorted[0].Date)
	}

	// No sort
	sorted = SortTasks(tasks, "", "asc")
	if sorted[0].Name != "B" {
		t.Errorf("no sort should preserve order")
	}
}

func TestGroupByDateSortOrder(t *testing.T) {
	// Regression: string sort puts "3/9" before "3/10" because '9' > '1'.
	// Date-aware sort should put 3/10 first (most recent).
	tasks := []model.IndexedTask{
		{Task: model.Task{Date: "3/1/2026", TimeSpent: "1h"}, Index: 0},
		{Task: model.Task{Date: "3/9/2026", TimeSpent: "1h"}, Index: 1},
		{Task: model.Task{Date: "3/10/2026", TimeSpent: "1h"}, Index: 2},
	}
	groups := GroupByDate(tasks)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	want := []string{"3/10/2026", "3/9/2026", "3/1/2026"}
	for i, w := range want {
		if groups[i].Key != w {
			t.Errorf("groups[%d].Key = %s, want %s", i, groups[i].Key, w)
		}
	}
}

func TestTodayIndex(t *testing.T) {
	today := TodayStr()
	groups := []DateGroup{
		{Key: today},
		{Key: "1/1/2020"},
	}
	idx := TodayIndex(groups)
	if idx != 0 {
		t.Errorf("expected 0, got %d", idx)
	}

	// Not found
	groups2 := []DateGroup{{Key: "1/1/2020"}}
	idx2 := TodayIndex(groups2)
	if idx2 != -1 {
		t.Errorf("expected -1, got %d", idx2)
	}
}

func TestSumTimeForDate(t *testing.T) {
	tasks := []model.Task{
		{Date: "3/5/2026", TimeSpent: "1h"},
		{Date: "3/5/2026", TimeSpent: "30m"},
		{Date: "3/4/2026", TimeSpent: "2h"},
	}
	got := SumTimeForDate(tasks, "3/5/2026")
	if got != 1.5 {
		t.Errorf("expected 1.5, got %v", got)
	}
	got2 := SumTimeForDate(tasks, "3/6/2026")
	if got2 != 0 {
		t.Errorf("expected 0, got %v", got2)
	}
}

func TestSortByTimeSpent(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{TimeSpent: "2h"}, Index: 0},
		{Task: model.Task{TimeSpent: "30m"}, Index: 1},
	}
	sorted := SortTasks(tasks, "timeSpent", "asc")
	if sorted[0].TimeSpent != "30m" {
		t.Errorf("expected 30m first, got %s", sorted[0].TimeSpent)
	}
}
