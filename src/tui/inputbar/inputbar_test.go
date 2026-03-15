package inputbar

import (
	"strings"
	"testing"
)

func TestNew_inactive(t *testing.T) {
	m := New()
	if m.Active() {
		t.Fatal("new input bar should be inactive")
	}
	if m.Value() != "" {
		t.Fatalf("expected empty value, got %q", m.Value())
	}
	if m.View() != "" {
		t.Fatalf("inactive input bar should render empty, got %q", m.View())
	}
	if m.Height() != 0 {
		t.Fatalf("inactive height should be 0, got %d", m.Height())
	}
}

func TestActivate(t *testing.T) {
	m := New()
	m.Activate(Config{
		Placeholder: "type a task...",
		Hints:       "Enter=submit  Escape=cancel",
	})

	if !m.Active() {
		t.Fatal("should be active after Activate")
	}
	if m.Value() != "" {
		t.Fatalf("expected empty value after activate, got %q", m.Value())
	}

	view := m.View()
	if !strings.Contains(view, "Enter=submit") {
		t.Fatalf("view should contain hints, got %q", view)
	}
	if m.Height() != 4 {
		t.Fatalf("expected height 4 (border + input + border + hints), got %d", m.Height())
	}
}

func TestActivateWithValue(t *testing.T) {
	m := New()
	m.ActivateWithValue(Config{
		Placeholder: "edit value",
		Hints:       "Enter=save",
	}, "hello world")

	if !m.Active() {
		t.Fatal("should be active")
	}
	if m.Value() != "hello world" {
		t.Fatalf("expected 'hello world', got %q", m.Value())
	}
}

func TestDeactivate(t *testing.T) {
	m := New()
	m.Activate(Config{Placeholder: "test", Hints: "test"})
	m.Deactivate()

	if m.Active() {
		t.Fatal("should be inactive after Deactivate")
	}
	if m.Value() != "" {
		t.Fatalf("value should be cleared after deactivate, got %q", m.Value())
	}
	if m.View() != "" {
		t.Fatal("view should be empty after deactivate")
	}
}

func TestSetValue(t *testing.T) {
	m := New()
	m.Activate(Config{Placeholder: "test"})
	m.SetValue("new value")

	if m.Value() != "new value" {
		t.Fatalf("expected 'new value', got %q", m.Value())
	}
}

func TestHeight_noHints(t *testing.T) {
	m := New()
	m.Activate(Config{Placeholder: "test"})

	if m.Height() != 3 {
		t.Fatalf("expected height 3 (border + input + border), got %d", m.Height())
	}
}

func TestSetWidth(t *testing.T) {
	m := New()
	m.SetWidth(80)
	// Just verify it doesn't panic; width is used internally by textinput
	m.Activate(Config{Placeholder: "test"})
	_ = m.View()
}
