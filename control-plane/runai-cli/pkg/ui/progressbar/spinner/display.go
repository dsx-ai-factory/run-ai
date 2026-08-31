/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spinner

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
	title    string
}

func initialModel(title string) model {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.PinkStr))
	return model{spinner: s, title: title}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case ui.CtrlPlusC, ui.KeyQ, ui.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		default:
			return m, nil
		}
	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() (s string) {
	if m.err != nil {
		return m.err.Error()
	}
	str := fmt.Sprintf("\n %s %s\n", m.spinner.View(), m.title)
	if m.quitting {
		return str
	}
	return str
}
