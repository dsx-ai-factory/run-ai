package choises

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type SelectDisplay struct {
	cursor  int
	choice  string
	choices []string
	subject string
}

func (a *SelectDisplay) Init() tea.Cmd {
	return nil
}

func (a *SelectDisplay) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) { //nolint:gocritic // keeping this
	case tea.KeyMsg:
		switch msg.String() {
		case ui.CtrlPlusC, ui.KeyQ, ui.KeyEsc:
			return a, tea.Quit

		case ui.KeyEnter:
			// Send the choice on the channel and exit.
			a.choice = a.choices[a.cursor]
			return a, tea.Quit

		case ui.KeyDown, ui.KeyJ:
			a.cursor++
			if a.cursor >= len(a.choices) {
				a.cursor = 0
			}

		case ui.KeyUp, ui.KeyK:
			a.cursor--
			if a.cursor < 0 {
				a.cursor = len(a.choices) - 1
			}
		}
	}

	return a, nil
}

func (a *SelectDisplay) View() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("\n%s\n\n", a.subject))

	for i := 0; i < len(a.choices); i++ {
		if a.cursor == i {
			s.WriteString(ui.ForegroundPinkStyle.Render(ui.ButtonSelectedRound))
		} else {
			s.WriteString(ui.ForegroundPinkStyle.Render(ui.ButtonNotSelectedRound))
		}
		s.WriteString(a.choices[i])
		s.WriteString("\n")
	}
	s.WriteString(ui.PressQToQuit)

	return s.String()
}

func ExecuteTea(subject, selected string, choices []string) (string, error) {
	i := 0
	for choice := range choices {
		if choices[choice] == selected {
			i = choice
			break
		}
	}

	if choices[i] != selected {
		i = 0
	}

	p := tea.NewProgram(&SelectDisplay{choices: choices, cursor: i, subject: subject}, tea.WithAltScreen())

	// Run returns the model as a tea.Model.
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	// Assert the final tea.Model to our local model and print the choice.
	if m, ok := m.(*SelectDisplay); ok && m.choice != "" {
		return m.choice, nil
	}

	return "", fmt.Errorf(ui.ErrorNoChoiceWasMade)
}
