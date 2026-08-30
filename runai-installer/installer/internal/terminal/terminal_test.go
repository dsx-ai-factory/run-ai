package terminal

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Predictable output for assertions: no ANSI styling, no fancy unicode.
	pterm.DisableColor()
	pterm.DisableStyling()
}

func TestTerminal_Info(t *testing.T) {
	var buf bytes.Buffer
	term := New(WithWriter(&buf))

	term.Info("loading %s", "config")

	out := buf.String()
	assert.Contains(t, out, "INFO")
	assert.Contains(t, out, "loading config")
}

func TestTerminal_Success(t *testing.T) {
	var buf bytes.Buffer
	term := New(WithWriter(&buf))

	term.Success("installed %s", "gpu-operator")

	out := buf.String()
	assert.Contains(t, out, "SUCCESS")
	assert.Contains(t, out, "installed gpu-operator")
}

func TestTerminal_Warning(t *testing.T) {
	var buf bytes.Buffer
	term := New(WithWriter(&buf))

	term.Warning("custom values overridden")

	out := buf.String()
	assert.Contains(t, out, "WARNING")
	assert.Contains(t, out, "custom values overridden")
}

func TestTerminal_Error_GoesToErrWriter(t *testing.T) {
	var out, errBuf bytes.Buffer
	term := New(WithWriter(&out), WithErrWriter(&errBuf))

	term.Error("install failed")

	assert.Contains(t, errBuf.String(), "ERROR")
	assert.Contains(t, errBuf.String(), "install failed")
	assert.Empty(t, out.String(), "errors must not pollute stdout")
}

func TestTerminal_Print_Plain(t *testing.T) {
	var buf bytes.Buffer
	term := New(WithWriter(&buf))

	term.Print("hello %s", "world")

	out := strings.TrimSpace(buf.String())
	assert.Equal(t, "hello world", out)
}

func TestTerminal_Heading(t *testing.T) {
	var buf bytes.Buffer
	term := New(WithWriter(&buf))

	term.Heading("Install")

	assert.Contains(t, buf.String(), "Install")
}

func TestTerminal_Yes(t *testing.T) {
	t.Run("default false", func(t *testing.T) {
		assert.False(t, New().Yes())
	})
	t.Run("set true", func(t *testing.T) {
		assert.True(t, New(WithYes(true)).Yes())
	})
}

func TestTerminal_Spinner_SmokeTest(t *testing.T) {
	var buf bytes.Buffer
	term := New(WithWriter(&buf))

	s, err := term.StartSpinner("loading")
	require.NoError(t, err)
	s.Update("still loading")
	s.Success("done")

	assert.Contains(t, buf.String(), "done")
}

func TestTerminal_Progress_SmokeTest(t *testing.T) {
	var buf bytes.Buffer
	term := New(WithWriter(&buf))

	p, err := term.StartProgress(3, "Steps")
	require.NoError(t, err)
	p.Increment()
	p.Add(2)
	require.NoError(t, p.Stop())

	assert.Contains(t, buf.String(), "Steps")
}
