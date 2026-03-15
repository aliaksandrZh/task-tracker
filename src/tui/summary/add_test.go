package summary

import (
	"testing"
)

func TestRequiredAddMissing_FiltersOptional(t *testing.T) {
	missing := []string{"type", "timeSpent", "number", "date", "name"}
	got := requiredAddMissing(missing)
	expected := []string{"type", "number", "name"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d fields, got %d: %v", len(expected), len(got), got)
	}
	for i, v := range expected {
		if got[i] != v {
			t.Errorf("index %d: expected %q, got %q", i, v, got[i])
		}
	}
}

func TestRequiredAddMissing_Empty(t *testing.T) {
	got := requiredAddMissing(nil)
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestRequiredAddMissing_AllOptional(t *testing.T) {
	got := requiredAddMissing([]string{"timeSpent", "date"})
	if len(got) != 0 {
		t.Errorf("expected empty result for all-optional, got %v", got)
	}
}
