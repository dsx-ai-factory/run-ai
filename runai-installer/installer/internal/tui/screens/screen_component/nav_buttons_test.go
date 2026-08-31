/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package screencomponent_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	component "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/screen_component"
)

func TestNavButtons_Definitions(t *testing.T) {
	if component.NextButton.Name != "Next" || component.NextButton.Type != component.NavButtonTypeNext {
		t.Fatalf("NextButton definition mismatch: %+v", component.NextButton)
	}
	if component.BackButton.Name != "Back" || component.BackButton.Type != component.NavButtonTypeBack {
		t.Fatalf("BackButton definition mismatch: %+v", component.BackButton)
	}
	if component.ExitButton.Name != "Exit" || component.ExitButton.Type != component.NavButtonTypeExit {
		t.Fatalf("ExitButton definition mismatch: %+v", component.ExitButton)
	}
}

func TestNavButtons_Events(t *testing.T) {
	// Next button event produces NextScreenMsg
	if cmd := component.NextButton.Event(); cmd == nil {
		t.Fatalf("NextButton.Event returned nil")
	} else {
		msg := cmd()
		if _, ok := msg.(component.NextScreenMsg); !ok {
			t.Fatalf("expected NextScreenMsg, got %T", msg)
		}
	}

	// Back button event produces BackNavigationMsg
	if cmd := component.BackButton.Event(); cmd == nil {
		t.Fatalf("BackButton.Event returned nil")
	} else {
		msg := cmd()
		if _, ok := msg.(component.BackNavigationMsg); !ok {
			t.Fatalf("expected BackNavigationMsg, got %T", msg)
		}
	}

	// Exit button uses tea.Quit
	if cmd := component.ExitButton.Event(); cmd == nil {
		t.Fatalf("ExitButton.Event returned nil")
	} else {
		msg := cmd()
		if _, ok := any(msg).(tea.Msg); !ok {
			t.Fatalf("expected tea.Msg from ExitButton event, got %T", msg)
		}
	}
}

func TestGetNavButtons_Combinations(t *testing.T) {
	first := component.GetNavButtons(true, false)
	if len(first) != 2 || first[0].Type != component.NavButtonTypeNext || first[1].Type != component.NavButtonTypeExit {
		t.Fatalf("unexpected first-screen buttons: %+v", first)
	}

	last := component.GetNavButtons(false, true)
	if len(last) != 2 || last[0].Type != component.NavButtonTypeBack || last[1].Type != component.NavButtonTypeExit {
		t.Fatalf("unexpected last-screen buttons: %+v", last)
	}

	mid := component.GetNavButtons(false, false)
	if len(mid) != 3 || mid[0].Type != component.NavButtonTypeBack || mid[1].Type != component.NavButtonTypeNext || mid[2].Type != component.NavButtonTypeExit {
		t.Fatalf("unexpected mid-screen buttons: %+v", mid)
	}
}
