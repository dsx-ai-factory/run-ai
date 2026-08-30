package terminal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfirm_YesShortCircuits(t *testing.T) {
	term := New(WithYes(true))

	got, err := term.Confirm("Continue?", false)
	require.NoError(t, err)
	assert.True(t, got, "--yes must auto-confirm regardless of defaultYes")
}

func TestSelect_YesReturnsDefault(t *testing.T) {
	term := New(WithYes(true))

	got, err := term.Select("Pick:", []string{"abort", "skip", "retry"}, "skip")
	require.NoError(t, err)
	assert.Equal(t, "skip", got)
}

func TestSelect_YesErrorsWithoutDefault(t *testing.T) {
	term := New(WithYes(true))

	_, err := term.Select("Pick:", []string{"a", "b"}, "")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoDefault)
}

func TestInput_YesReturnsDefault(t *testing.T) {
	term := New(WithYes(true))

	got, err := term.Input("Name:", "default-name")
	require.NoError(t, err)
	assert.Equal(t, "default-name", got)
}

func TestInput_YesErrorsWithoutDefault(t *testing.T) {
	term := New(WithYes(true))

	_, err := term.Input("Name:", "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoDefault))
}
