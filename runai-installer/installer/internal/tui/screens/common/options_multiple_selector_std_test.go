package common

import (
	"strings"
	"testing"
)

// mandatoryOption wraps SimpleOption to return true for IsMandatory when requested
type mandatoryOption struct {
	SimpleOption
	mandatory bool
}

func (m *mandatoryOption) IsMandatory() bool { return m.mandatory }

func TestOptionsMultipleSelector_ToggleAndMandatory(t *testing.T) {
	opts := []Options{
		NewSimpleOption("One", 0, false),
		NewSimpleOption("Two", 1, false),
	}
	sel := NewOptionsMultipleSelector(opts)

	sel.selectItem(0)
	if !sel.isSelected(0) {
		t.Fatalf("expected index 0 selected after toggle on")
	}
	sel.selectItem(0)
	if sel.isSelected(0) {
		t.Fatalf("expected index 0 deselected after toggle off")
	}
}

func TestOptionsMultipleSelector_MandatoryNotUnselected(t *testing.T) {
	opts := []Options{
		&mandatoryOption{SimpleOption: *NewSimpleOption("Must", 0, true), mandatory: true},
		NewSimpleOption("Other", 1, false),
	}
	sel := NewOptionsMultipleSelector(opts)

	// Attempt to unselect mandatory (index 0)
	sel.selectItem(0)
	if !sel.isSelected(0) {
		t.Fatalf("mandatory option should remain selected when toggled")
	}

	sel.selectItem(1)
	if !sel.isSelected(1) {
		t.Fatalf("expected non-mandatory option to be selected")
	}
}

func TestOptionsMultipleSelector_Render(t *testing.T) {
	opts := []Options{
		NewSimpleOption("First", 0, true),
		NewSimpleOption("Second", 1, false),
	}
	sel := NewOptionsMultipleSelector(opts)
	out := sel.render(1)
	if !strings.Contains(out, "First") || !strings.Contains(out, "Second") {
		t.Fatalf("render should include option names; got: %q", out)
	}
	if !strings.Contains(out, "[X]") || !strings.Contains(out, "[ ]") {
		t.Fatalf("render should reflect selection markers; got: %q", out)
	}
}
