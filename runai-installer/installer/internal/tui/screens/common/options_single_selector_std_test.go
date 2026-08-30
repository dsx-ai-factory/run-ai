package common

import (
	"strings"
	"testing"
)

func TestOptionsSingleSelector_RenderAndSelection(t *testing.T) {
	opts := []Options{
		NewSimpleOption("One", 0, false),
		NewSimpleOption("Two", 1, false),
		NewSimpleOption("Three", 2, false),
	}

	sel := NewOptionsSingleSelect(opts)

	out := sel.render(0)
	if !strings.Contains(out, "One") || !strings.Contains(out, "Two") {
		t.Fatalf("render should include option names; got: %q", out)
	}
	if !strings.Contains(out, "( )") {
		t.Fatalf("expected unselected radio '( )' in initial render; got: %q", out)
	}

	sel.selectItem(1)
	if !sel.isSelected(1) || sel.isSelected(0) {
		t.Fatalf("selection state not updated correctly: sel1=%v sel0=%v", sel.isSelected(1), sel.isSelected(0))
	}

	out = sel.render(1)
	if !strings.Contains(out, "(•)") {
		t.Fatalf("expected selected radio '(•)' after selection; got: %q", out)
	}
}

func TestOptionsSingleSelector_ReselectDeselectsPrevious(t *testing.T) {
	opts := []Options{
		NewSimpleOption("A", 0, false),
		NewSimpleOption("B", 1, false),
		NewSimpleOption("C", 2, false),
	}
	sel := NewOptionsSingleSelect(opts)

	sel.selectItem(0)
	if !sel.isSelected(0) {
		t.Fatalf("expected index 0 to be selected")
	}

	sel.selectItem(2)
	if sel.isSelected(0) || !sel.isSelected(2) {
		t.Fatalf("expected index 0 deselected and index 2 selected")
	}
}
