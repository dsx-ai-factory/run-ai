package spinner

import (
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/progressbar"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	SpinnerFull = 100
)

type Spinner struct {
	teaProg *tea.Program
}

func NewSpinner(title string) progressbar.ProgressBar {
	return &Spinner{
		teaProg: tea.NewProgram(initialModel(title)),
	}
}

func (s *Spinner) Stop() {
	s.teaProg.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	time.Sleep(1 * time.Second) // wait for 1 second to let the spinner clean up
	fmt.Println("\r")
}

func (s *Spinner) Run() {
	if _, err := s.teaProg.Run(); err != nil {
		return
	}
}

func (s *Spinner) Set(num int) {
	if num == SpinnerFull {
		s.Stop()
	}
}
