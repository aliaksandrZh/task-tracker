package timer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TimerData represents a running timer persisted to .timer.json.
type TimerData struct {
	Type      string `json:"type"`
	Number    string `json:"number"`
	Name      string `json:"name"`
	StartedAt int64  `json:"startedAt"` // Unix milliseconds
}

// TimerStatus is a running timer with calculated elapsed time.
type TimerStatus struct {
	TimerData
	Elapsed   int64  `json:"elapsed"`   // milliseconds
	TimeSpent string `json:"timeSpent"` // formatted
}

// Timer manages timer persistence.
type Timer struct {
	Dir string
}

// New creates a Timer that stores .timer.json in the given directory.
func New(dir string) *Timer {
	return &Timer{Dir: dir}
}

func (t *Timer) path() string {
	return filepath.Join(t.Dir, ".timer.json")
}

// FormatElapsed converts milliseconds to a human-readable time string (e.g. "1h 30m").
func FormatElapsed(ms int64) string {
	totalMin := int(float64(ms) / 60000)
	if totalMin < 1 {
		return "0m"
	}
	h := totalMin / 60
	m := totalMin % 60
	if h > 0 && m > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if h > 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dm", m)
}

// FormatElapsedDecimal converts milliseconds to decimal hours (e.g. "1.5h").
func FormatElapsedDecimal(ms int64) string {
	hours := float64(ms) / 3600000
	if hours < 0.1 {
		return "0h"
	}
	s := fmt.Sprintf("%.1f", hours)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s + "h"
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}

// Start begins a new timer. Errors if one is already running.
func (t *Timer) Start(typ, number, name string) (*TimerData, error) {
	if _, err := os.Stat(t.path()); err == nil {
		return nil, fmt.Errorf("Timer already running. Stop it first with: tt stop")
	}
	data := &TimerData{
		Type:      typ,
		Number:    number,
		Name:      name,
		StartedAt: nowMs(),
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(t.path(), b, 0o644); err != nil {
		return nil, err
	}
	return data, nil
}

// Stop stops the running timer and returns the result. Errors if no timer running.
func (t *Timer) Stop() (*TimerStatus, error) {
	b, err := os.ReadFile(t.path())
	if err != nil {
		return nil, fmt.Errorf("No timer running. Start one with: tt start <type> <number>: <name>")
	}
	var data TimerData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	elapsed := nowMs() - data.StartedAt
	if err := os.Remove(t.path()); err != nil {
		return nil, err
	}
	return &TimerStatus{
		TimerData: data,
		Elapsed:   elapsed,
		TimeSpent: FormatElapsedDecimal(elapsed),
	}, nil
}

// GetStatus returns the current timer status, or nil if no timer running.
func (t *Timer) GetStatus() *TimerStatus {
	b, err := os.ReadFile(t.path())
	if err != nil {
		return nil
	}
	var data TimerData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil
	}
	elapsed := nowMs() - data.StartedAt
	return &TimerStatus{
		TimerData: data,
		Elapsed:   elapsed,
		TimeSpent: FormatElapsed(elapsed),
	}
}
