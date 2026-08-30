// Package terminal provides line-mode styled output and interactive prompts
// for the installer CLI. It wraps pterm so callers get spinners, progress
// bars, colored banners, and prompts without managing pterm globals.
//
// Terminal is not safe for concurrent use. Output methods write sequentially
// to the configured writer.
package terminal

import (
	"io"
	"os"

	"github.com/pterm/pterm"
	"golang.org/x/term"
)

type Terminal struct {
	out io.Writer
	err io.Writer
	yes bool
	tty bool
	// ttyForced records that WithTTY set tty explicitly, so New does not
	// overwrite it with auto-detection. confirmFn, when set, replaces the
	// interactive prompt backend. Both exist so tests can drive the interactive
	// paths without a real TTY/stdin.
	ttyForced bool
	confirmFn func(question string, defaultYes bool) (bool, error)

	info    pterm.PrefixPrinter
	success pterm.PrefixPrinter
	warning pterm.PrefixPrinter
	skip    pterm.PrefixPrinter
	errp    pterm.PrefixPrinter
	plain   pterm.BasicTextPrinter
	heading pterm.SectionPrinter
}

type Option func(*Terminal)

func WithWriter(w io.Writer) Option { return func(t *Terminal) { t.out = w } }

func WithErrWriter(w io.Writer) Option { return func(t *Terminal) { t.err = w } }

func WithYes(yes bool) Option { return func(t *Terminal) { t.yes = yes } }

// WithTTY forces the interactive (TTY) state instead of auto-detecting it from
// the output writer. Intended for tests that need to exercise interactive
// paths against a buffer.
func WithTTY(tty bool) Option {
	return func(t *Terminal) { t.tty = tty; t.ttyForced = true }
}

// WithConfirmFunc overrides the backend used by Confirm. Intended for tests to
// inject a deterministic answer in place of the real interactive prompt.
func WithConfirmFunc(fn func(question string, defaultYes bool) (bool, error)) Option {
	return func(t *Terminal) { t.confirmFn = fn }
}

func New(opts ...Option) *Terminal {
	t := &Terminal{
		out: os.Stdout,
		err: os.Stderr,
	}
	for _, opt := range opts {
		opt(t)
	}
	if !t.ttyForced {
		t.tty = isTTY(t.out)
	}

	t.info = *pterm.Info.WithWriter(t.out)
	t.success = *pterm.Success.WithWriter(t.out)
	t.warning = *pterm.Warning.WithWriter(t.out)
	skip := pterm.Warning
	skip.Prefix.Text = "SKIPPED"
	t.skip = *skip.WithWriter(t.out)
	t.errp = *pterm.Error.WithWriter(t.err)
	t.plain = pterm.BasicTextPrinter{Writer: t.out}
	t.heading = *pterm.DefaultSection.WithWriter(t.out)

	return t
}

func (t *Terminal) Yes() bool { return t.yes }

func (t *Terminal) IsTTY() bool { return t.tty }

// Width returns the terminal width in columns, defaulting to 80.
func (t *Terminal) Width() int {
	if w := pterm.GetTerminalWidth(); w > 0 {
		return w
	}
	return 80
}

func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func (t *Terminal) Info(format string, args ...any) {
	t.info.Printfln(format, args...)
}

func (t *Terminal) Success(format string, args ...any) {
	t.success.Printfln(format, args...)
}

func (t *Terminal) Warning(format string, args ...any) {
	t.warning.Printfln(format, args...)
}

func (t *Terminal) Skipped(format string, args ...any) {
	t.skip.Printfln(format, args...)
}

func (t *Terminal) Error(format string, args ...any) {
	t.errp.Printfln(format, args...)
}

func (t *Terminal) Print(format string, args ...any) {
	t.plain.Printfln(format, args...)
}

func (t *Terminal) Heading(title string) {
	t.heading.Println(title)
}
