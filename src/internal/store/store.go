package store

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aliaksandrZh/worklog/src/internal/model"
	"github.com/aliaksandrZh/worklog/src/internal/timeutil"
)

// TaskReader can load tasks.
type TaskReader interface {
	LoadTasks() ([]model.Task, error)
}

// TaskWriter can modify tasks.
type TaskWriter interface {
	AddTask(task model.Task) error
	AddTasks(tasks []model.Task) error
	UpdateTask(index int, updates map[string]string) error
	DeleteTask(index int) error
}

// TaskRepository combines reading and writing.
type TaskRepository interface {
	TaskReader
	TaskWriter
}

// CSVStore implements TaskRepository using a CSV file.
type CSVStore struct {
	Path string
}

// DataDir returns the directory for data files.
// Uses WORKLOG_DIR env var if set, otherwise the current directory.
func DataDir() string {
	if dir := os.Getenv("WORKLOG_DIR"); dir != "" {
		return dir
	}
	return "."
}

// New creates a CSVStore at the default location (tasks.csv in data dir).
func New() *CSVStore {
	return &CSVStore{Path: filepath.Join(DataDir(), "tasks.csv")}
}

// NewWithPath creates a CSVStore at the given path.
func NewWithPath(path string) *CSVStore {
	return &CSVStore{Path: path}
}

// IsValidDate checks if a date string is valid M/D/YYYY.
func IsValidDate(str string) bool {
	if str == "" {
		return false
	}
	parts := strings.Split(str, "/")
	if len(parts) != 3 {
		return false
	}
	m, err1 := strconv.Atoi(parts[0])
	d, err2 := strconv.Atoi(parts[1])
	y, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return false
	}
	return m >= 1 && m <= 12 && d >= 1 && d <= 31 && y >= 2000
}

// sanitizeField collapses newlines and trims whitespace to keep CSV safe.
func sanitizeField(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

// sanitizeTask cleans all fields of a task for safe CSV storage.
func sanitizeTask(task model.Task) model.Task {
	task.Date = sanitizeField(task.Date)
	task.Type = sanitizeField(task.Type)
	task.Number = sanitizeField(task.Number)
	task.Name = sanitizeField(task.Name)
	task.TimeSpent = sanitizeField(task.TimeSpent)
	task.Comments = sanitizeField(task.Comments)
	return task
}

func ensureDate(task model.Task) model.Task {
	task = sanitizeTask(task)
	if !IsValidDate(task.Date) {
		task.Date = timeutil.TodayStr()
	}
	return task
}

func (s *CSVStore) ensureFile() error {
	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		dir := filepath.Dir(s.Path)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		}
		f, err := os.Create(s.Path)
		if err != nil {
			return err
		}
		w := csv.NewWriter(f)
		_ = w.Write(model.Headers)
		w.Flush()
		f.Close()
	}
	return nil
}

// LoadTasks reads all tasks from the CSV file.
func (s *CSVStore) LoadTasks() ([]model.Task, error) {
	if err := s.ensureFile(); err != nil {
		return nil, err
	}
	f, err := os.Open(s.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	// Skip header
	header := records[0]
	colIndex := make(map[string]int)
	for i, h := range header {
		colIndex[strings.TrimSpace(h)] = i
	}

	var tasks []model.Task
	for _, row := range records[1:] {
		t := model.Task{}
		if idx, ok := colIndex["date"]; ok && idx < len(row) {
			t.Date = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["type"]; ok && idx < len(row) {
			t.Type = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["number"]; ok && idx < len(row) {
			t.Number = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["name"]; ok && idx < len(row) {
			t.Name = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["timeSpent"]; ok && idx < len(row) {
			t.TimeSpent = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["comments"]; ok && idx < len(row) {
			t.Comments = strings.TrimSpace(row[idx])
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// SaveTasks writes all tasks to the CSV file.
func (s *CSVStore) SaveTasks(tasks []model.Task) error {
	f, err := os.Create(s.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	_ = w.Write(model.Headers)
	for _, t := range tasks {
		_ = w.Write([]string{t.Date, t.Type, t.Number, t.Name, t.TimeSpent, t.Comments})
	}
	w.Flush()
	return w.Error()
}

// AddTask appends a single task (with date validation).
func (s *CSVStore) AddTask(task model.Task) error {
	tasks, err := s.LoadTasks()
	if err != nil {
		return err
	}
	tasks = append(tasks, ensureDate(task))
	return s.SaveTasks(tasks)
}

// AddTasks appends multiple tasks (with date validation).
func (s *CSVStore) AddTasks(newTasks []model.Task) error {
	tasks, err := s.LoadTasks()
	if err != nil {
		return err
	}
	for _, t := range newTasks {
		tasks = append(tasks, ensureDate(t))
	}
	return s.SaveTasks(tasks)
}

// UpdateTask updates fields of the task at the given index.
func (s *CSVStore) UpdateTask(index int, updates map[string]string) error {
	tasks, err := s.LoadTasks()
	if err != nil {
		return err
	}
	if index < 0 || index >= len(tasks) {
		return fmt.Errorf("index %d out of range", index)
	}
	t := tasks[index]
	for k, v := range updates {
		v = sanitizeField(v)
		switch k {
		case "date":
			t.Date = v
		case "type":
			t.Type = v
		case "number":
			t.Number = v
		case "name":
			t.Name = v
		case "timeSpent":
			t.TimeSpent = v
		case "comments":
			t.Comments = v
		}
	}
	tasks[index] = t
	return s.SaveTasks(tasks)
}

// DeleteTask removes the task at the given index.
func (s *CSVStore) DeleteTask(index int) error {
	tasks, err := s.LoadTasks()
	if err != nil {
		return err
	}
	if index < 0 || index >= len(tasks) {
		return fmt.Errorf("index %d out of range", index)
	}
	tasks = append(tasks[:index], tasks[index+1:]...)
	return s.SaveTasks(tasks)
}
