package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliaksandrZh/worklog/src/internal/model"
)

func tempStore(t *testing.T) *CSVStore {
	t.Helper()
	dir := t.TempDir()
	return NewWithPath(filepath.Join(dir, "tasks.csv"))
}

func TestLoadCreatesFile(t *testing.T) {
	s := tempStore(t)
	tasks, err := s.LoadTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		t.Error("file should exist after load")
	}
}

func TestAddAndLoad(t *testing.T) {
	s := tempStore(t)
	err := s.AddTask(model.Task{
		Date: "3/5/2026", Type: "Bug", Number: "123",
		Name: "Fix login", TimeSpent: "1h", Comments: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	tasks, err := s.LoadTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Name != "Fix login" {
		t.Errorf("got name %q", tasks[0].Name)
	}
}

func TestAddTasks(t *testing.T) {
	s := tempStore(t)
	err := s.AddTasks([]model.Task{
		{Date: "3/5/2026", Type: "Bug", Number: "1", Name: "A"},
		{Date: "3/5/2026", Type: "Task", Number: "2", Name: "B"},
	})
	if err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.LoadTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestUpdateTask(t *testing.T) {
	s := tempStore(t)
	_ = s.AddTask(model.Task{Date: "3/5/2026", Type: "Bug", Number: "123", Name: "Old"})
	err := s.UpdateTask(0, map[string]string{"name": "New"})
	if err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.LoadTasks()
	if tasks[0].Name != "New" {
		t.Errorf("expected New, got %q", tasks[0].Name)
	}
	if tasks[0].Type != "Bug" {
		t.Errorf("update should preserve other fields, got type %q", tasks[0].Type)
	}
}

func TestDeleteTask(t *testing.T) {
	s := tempStore(t)
	_ = s.AddTasks([]model.Task{
		{Date: "3/5/2026", Name: "A"},
		{Date: "3/5/2026", Name: "B"},
	})
	err := s.DeleteTask(0)
	if err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.LoadTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Name != "B" {
		t.Errorf("expected B, got %q", tasks[0].Name)
	}
}

func TestDeleteOutOfRange(t *testing.T) {
	s := tempStore(t)
	_ = s.AddTask(model.Task{Date: "3/5/2026", Name: "A"})
	err := s.DeleteTask(5)
	if err == nil {
		t.Error("expected error for out of range")
	}
}

func TestTrimWhitespace(t *testing.T) {
	s := tempStore(t)
	// Write CSV with whitespace manually
	err := os.WriteFile(s.Path, []byte("date,type,number,name,timeSpent,comments\n  3/5/2026 , Bug , 123 , Fix login , 1h , test \n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	tasks, err := s.LoadTasks()
	if err != nil {
		t.Fatal(err)
	}
	if tasks[0].Date != "3/5/2026" {
		t.Errorf("date not trimmed: %q", tasks[0].Date)
	}
	if tasks[0].Type != "Bug" {
		t.Errorf("type not trimmed: %q", tasks[0].Type)
	}
}

func TestInvalidDateDefaultsToToday(t *testing.T) {
	s := tempStore(t)
	err := s.AddTask(model.Task{Date: "invalid", Name: "Test"})
	if err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.LoadTasks()
	if tasks[0].Date == "invalid" {
		t.Error("invalid date should be replaced with today")
	}
}

func TestEmptyDateDefaultsToToday(t *testing.T) {
	s := tempStore(t)
	err := s.AddTask(model.Task{Date: "", Name: "Test"})
	if err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.LoadTasks()
	if tasks[0].Date == "" {
		t.Error("empty date should be replaced with today")
	}
}

func TestValidDatePreserved(t *testing.T) {
	s := tempStore(t)
	err := s.AddTask(model.Task{Date: "1/15/2026", Name: "Test"})
	if err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.LoadTasks()
	if tasks[0].Date != "1/15/2026" {
		t.Errorf("valid date should be preserved, got %q", tasks[0].Date)
	}
}

func TestSanitizeNewlines(t *testing.T) {
	s := tempStore(t)
	err := s.AddTask(model.Task{
		Date:     "3/5/2026",
		Type:     "Bug",
		Number:   "123",
		Name:     "Fix\nlogin\r\npage",
		Comments: "## Header\n\n- bullet 1\n- bullet 2\n\nSome `code` block",
	})
	if err != nil {
		t.Fatal(err)
	}
	tasks, err := s.LoadTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if strings.Contains(tasks[0].Name, "\n") {
		t.Errorf("name should not contain newlines: %q", tasks[0].Name)
	}
	if strings.Contains(tasks[0].Comments, "\n") {
		t.Errorf("comments should not contain newlines: %q", tasks[0].Comments)
	}
	if tasks[0].Name != "Fix login page" {
		t.Errorf("expected 'Fix login page', got %q", tasks[0].Name)
	}
}

func TestSanitizeOnUpdate(t *testing.T) {
	s := tempStore(t)
	_ = s.AddTask(model.Task{Date: "3/5/2026", Name: "Test", Comments: "old"})
	err := s.UpdateTask(0, map[string]string{"comments": "line1\nline2\nline3"})
	if err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.LoadTasks()
	if strings.Contains(tasks[0].Comments, "\n") {
		t.Errorf("updated comments should not contain newlines: %q", tasks[0].Comments)
	}
	if tasks[0].Comments != "line1 line2 line3" {
		t.Errorf("expected 'line1 line2 line3', got %q", tasks[0].Comments)
	}
}

func TestDataDirDefault(t *testing.T) {
	os.Unsetenv("WORKLOG_DIR")
	if dir := DataDir(); dir != "." {
		t.Errorf("expected '.', got %q", dir)
	}
}

func TestDataDirFromEnv(t *testing.T) {
	t.Setenv("WORKLOG_DIR", "/tmp/myworklog")
	if dir := DataDir(); dir != "/tmp/myworklog" {
		t.Errorf("expected '/tmp/myworklog', got %q", dir)
	}
}

func TestNewUsesDataDir(t *testing.T) {
	t.Setenv("WORKLOG_DIR", "/tmp/myworklog")
	s := New()
	if s.Path != "/tmp/myworklog/tasks.csv" {
		t.Errorf("expected '/tmp/myworklog/tasks.csv', got %q", s.Path)
	}
}

func TestIsValidDate(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"3/5/2026", true},
		{"12/31/2025", true},
		{"1/1/2000", true},
		{"", false},
		{"invalid", false},
		{"13/1/2026", false},
		{"1/32/2026", false},
		{"1/1/1999", false},
	}
	for _, tt := range tests {
		got := IsValidDate(tt.input)
		if got != tt.want {
			t.Errorf("IsValidDate(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
