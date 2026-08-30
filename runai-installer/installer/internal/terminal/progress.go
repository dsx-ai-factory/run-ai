package terminal

import "github.com/pterm/pterm"

type ProgressBar struct {
	inner *pterm.ProgressbarPrinter
}

func (t *Terminal) StartProgress(total int, title string) (*ProgressBar, error) {
	inner, err := pterm.DefaultProgressbar.
		WithWriter(t.out).
		WithTotal(total).
		WithTitle(title).
		Start()
	if err != nil {
		return nil, err
	}
	return &ProgressBar{inner: inner}, nil
}

func (p *ProgressBar) Increment() { p.inner.Increment() }

func (p *ProgressBar) Add(n int) { p.inner.Add(n) }

func (p *ProgressBar) UpdateTitle(title string) { p.inner.UpdateTitle(title) }

func (p *ProgressBar) Stop() error {
	_, err := p.inner.Stop()
	return err
}
