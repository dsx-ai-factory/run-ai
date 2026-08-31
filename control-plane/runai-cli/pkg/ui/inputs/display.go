/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputs

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	DefaultCharLimit = 32
)

var (
	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type Input struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	response   *map[string]string
	title      string
}

func initialModel(input []textinput.Model, title string) *Input {
	model := Input{
		inputs: make([]textinput.Model, len(input)),
		title:  title,
	}

	var t textinput.Model
	for i := range model.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = DefaultCharLimit
		if input[i].CharLimit > 0 {
			t.CharLimit = input[i].CharLimit
		}

		if i == 0 {
			t.Focus()
		}

		t.Placeholder = input[i].Placeholder
		t.EchoMode = input[i].EchoMode
		t.EchoCharacter = input[i].EchoCharacter
		if input[i].Prompt != "" {
			t.SetValue(input[i].Prompt)
		}

		model.inputs[i] = t
	}

	return &model
}

func (m *Input) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Input) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:gocognit // This is an interface I can't change it
	switch msg := msg.(type) { //nolint:gocritic // keeping this
	case tea.KeyMsg:
		switch msg.String() {
		case ui.CtrlPlusC, ui.KeyEsc:
			return m, tea.Quit

		// Change cursor mode
		case ui.CtrlPlusR:
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case ui.KeyTab, ui.KeyShiftTab, ui.KeyEnter, ui.KeyUp, ui.KeyDown:
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				m.response = new(map[string]string)
				*m.response = make(map[string]string, len(m.inputs))
				for i := range m.inputs {
					(*m.response)[m.inputs[i].Placeholder] = m.inputs[i].Value()
				}

				return m, tea.Quit
			}

			// Cycle indexes
			if s == ui.KeyUp || s == ui.KeyShiftTab {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *Input) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *Input) View() string {
	var b strings.Builder

	b.WriteString("\n\n" + textStyle.Render(m.title))
	b.WriteRune('\n')
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}

func MultiTexts(input []textinput.Model, title string) (*map[string]string, error) {
	if title == "" {
		title = "Please set values to the fields below:"
	}
	model := initialModel(input, title)
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		return nil, err
	}

	if model.response == nil {
		return nil, fmt.Errorf("no input was set")
	}

	return model.response, nil
}
