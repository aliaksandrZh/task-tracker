package summary

import (
	"testing"

	"github.com/aliaksandrZh/worklog/src/internal/model"
)

func TestFilterTasks_Empty(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Name: "Fix login", Type: "Bug", Number: "123"}, Index: 0},
	}
	got := filterTasks(tasks, "")
	if len(got) != 1 {
		t.Errorf("empty filter should return all tasks, got %d", len(got))
	}
}

func TestFilterTasks_ByName(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Name: "Fix login page", Type: "Bug", Number: "123"}, Index: 0},
		{Task: model.Task{Name: "Add dashboard", Type: "Task", Number: "456"}, Index: 1},
		{Task: model.Task{Name: "Update login flow", Type: "Task", Number: "789"}, Index: 2},
	}
	got := filterTasks(tasks, "login")
	if len(got) != 2 {
		t.Errorf("expected 2 tasks matching 'login', got %d", len(got))
	}
}

func TestFilterTasks_CaseInsensitive(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Name: "Fix Login Page", Type: "Bug"}, Index: 0},
	}
	got := filterTasks(tasks, "LOGIN")
	if len(got) != 1 {
		t.Errorf("filter should be case-insensitive, got %d", len(got))
	}
}

func TestFilterTasks_ByType(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Name: "Fix login", Type: "Bug"}, Index: 0},
		{Task: model.Task{Name: "Add feature", Type: "Task"}, Index: 1},
	}
	got := filterTasks(tasks, "bug")
	if len(got) != 1 {
		t.Errorf("expected 1 task matching type 'bug', got %d", len(got))
	}
	if got[0].Index != 0 {
		t.Errorf("expected task index 0, got %d", got[0].Index)
	}
}

func TestFilterTasks_ByNumber(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Name: "Fix login", Number: "12345"}, Index: 0},
		{Task: model.Task{Name: "Add feature", Number: "67890"}, Index: 1},
	}
	got := filterTasks(tasks, "123")
	if len(got) != 1 {
		t.Errorf("expected 1 task matching number '123', got %d", len(got))
	}
}

func TestFilterTasks_ByComments(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Name: "Fix login", Comments: "Pull Request 21479"}, Index: 0},
		{Task: model.Task{Name: "Add feature", Comments: ""}, Index: 1},
	}
	got := filterTasks(tasks, "pull request")
	if len(got) != 1 {
		t.Errorf("expected 1 task matching comments, got %d", len(got))
	}
}

func TestFilterTasks_NoMatch(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{Name: "Fix login", Type: "Bug", Number: "123"}, Index: 0},
	}
	got := filterTasks(tasks, "nonexistent")
	if len(got) != 0 {
		t.Errorf("expected 0 tasks for non-matching filter, got %d", len(got))
	}
}

func TestSumHours(t *testing.T) {
	tasks := []model.IndexedTask{
		{Task: model.Task{TimeSpent: "1h"}, Index: 0},
		{Task: model.Task{TimeSpent: "2h 30m"}, Index: 1},
		{Task: model.Task{TimeSpent: ""}, Index: 2},
	}
	got := sumHours(tasks)
	if got != 3.5 {
		t.Errorf("expected 3.5h, got %v", got)
	}
}
