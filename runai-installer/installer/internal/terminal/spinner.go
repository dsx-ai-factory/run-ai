package terminal

import (
	"time"

	"github.com/pterm/pterm"
)

// SpinnerDelay is how often an animated spinner advances a frame.
const SpinnerDelay = 200 * time.Millisecond

// SpinnerFrame returns the styled spinner glyph for frame i, matching the
// spinner shown for ordinary steps.
func SpinnerFrame(i int) string {
	seq := pterm.DefaultSpinner.Sequence
	if len(seq) == 0 {
		return ""
	}
	glyph := seq[((i%len(seq))+len(seq))%len(seq)]
	if pterm.DefaultSpinner.Style != nil {
		return pterm.DefaultSpinner.Style.Sprint(glyph)
	}
	return glyph
}

// Spinner wraps a pterm spinner with non-TTY fallback. On a TTY the inner
// pterm spinner renders animated frames; otherwise the start text is dropped
// and Success/Warning/Fail print plain lines through the Terminal.
type Spinner struct {
	t     *Terminal
	inner *pterm.SpinnerPrinter
}

func (t *Terminal) StartSpinner(message string) (*Spinner, error) {
	if !t.tty {
		return &Spinner{t: t}, nil
	}
	inner, err := pterm.DefaultSpinner.WithWriter(t.out).Start(message)
	if err != nil {
		return nil, err
	}
	return &Spinner{t: t, inner: inner}, nil
}

func (s *Spinner) Update(message string) {
	if s.inner != nil {
		s.inner.UpdateText(message)
	}
}

func (s *Spinner) Success(message string) {
	if s.inner != nil {
		s.inner.Success(message)
		return
	}
	s.t.Success("%s", message)
}

func (s *Spinner) Fail(message string) {
	if s.inner != nil {
		s.inner.Fail(message)
		return
	}
	s.t.Error("%s", message)
}

func (s *Spinner) Warning(message string) {
	if s.inner != nil {
		s.inner.Warning(message)
		return
	}
	s.t.Warning("%s", message)
}

func (s *Spinner) Skip(message string) {
	if s.inner != nil {
		s.inner.InfoPrinter = &s.t.skip
		s.inner.Info(message)
		return
	}
	s.t.Skipped("%s", message)
}

func (s *Spinner) Stop() error {
	if s.inner != nil {
		return s.inner.Stop()
	}
	return nil
}

// StopSilently stops the spinner and clears its line (used when handing
// rendering over to a live area).
func (s *Spinner) StopSilently() {
	if s.inner != nil {
		s.inner.RemoveWhenDone = true
		_ = s.inner.Stop()
	}
}

// LiveArea is a multi-line region that redraws in place on each Update.
type LiveArea struct {
	inner *pterm.AreaPrinter
}

// StartArea begins a live multi-line area. On a non-TTY it is a no-op.
func (t *Terminal) StartArea() (*LiveArea, error) {
	if !t.tty {
		return &LiveArea{}, nil
	}
	area := pterm.DefaultArea.WithRemoveWhenDone(true)
	area.SetWriter(t.out)
	inner, err := area.Start()
	if err != nil {
		return nil, err
	}
	return &LiveArea{inner: inner}, nil
}

func (a *LiveArea) Update(content string) {
	if a.inner != nil {
		a.inner.Update(content)
	}
}

// Stop clears the area and releases it.
func (a *LiveArea) Stop() {
	if a.inner != nil {
		_ = a.inner.Stop()
	}
}
