package specrun

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/terminal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDecideProceed_PreflightFailed_Interactive covers the branch where preflight
// reported failures and an interactive TTY user is prompted to continue anyway.
func TestDecideProceed_PreflightFailed_Interactive(t *testing.T) {
	failed := orchestrator2.PreflightSummary{Outcome: "failed", Failed: 1}
	opts := ExecOptions{ConfirmPrompt: "Proceed?"}
	s := &spec.Spec{} // interactive: NonInteractive=false, DryRun=false

	t.Run("user confirms -> proceed", func(t *testing.T) {
		var asked string
		term := terminal.New(
			terminal.WithTTY(true),
			terminal.WithConfirmFunc(func(q string, _ bool) (bool, error) { asked = q; return true, nil }),
		)
		proceed, err := decideProceed(context.Background(), s, term, opts, failed)
		require.NoError(t, err)
		assert.True(t, proceed, "confirming the prompt proceeds past a failed preflight")
		assert.Equal(t, "Preflight reported failures. Continue anyway?", asked)
	})

	t.Run("user declines -> stop", func(t *testing.T) {
		term := terminal.New(
			terminal.WithTTY(true),
			terminal.WithConfirmFunc(func(string, bool) (bool, error) { return false, nil }),
		)
		proceed, err := decideProceed(context.Background(), s, term, opts, failed)
		require.NoError(t, err)
		assert.False(t, proceed, "declining the prompt stops the install")
	})

	t.Run("non-TTY never prompts and stops", func(t *testing.T) {
		prompted := false
		term := terminal.New(
			terminal.WithTTY(false),
			terminal.WithConfirmFunc(func(string, bool) (bool, error) { prompted = true; return true, nil }),
		)
		proceed, err := decideProceed(context.Background(), s, term, opts, failed)
		require.NoError(t, err)
		assert.False(t, proceed, "a non-interactive run stops on a failed preflight")
		assert.False(t, prompted, "no prompt is shown without a TTY")
	})

	t.Run("dry-run interactive still prompts", func(t *testing.T) {
		var asked string
		dryRun := &spec.Spec{DryRun: true}
		term := terminal.New(
			terminal.WithTTY(true),
			terminal.WithConfirmFunc(func(q string, _ bool) (bool, error) { asked = q; return true, nil }),
		)
		proceed, err := decideProceed(context.Background(), dryRun, term, opts, failed)
		require.NoError(t, err)
		assert.True(t, proceed, "a dry run must not hard-stop on a failed preflight; the user is prompted")
		assert.Equal(t, "Preflight reported failures. Continue anyway?", asked)
	})
}

// TestDecideProceed_PreflightFailed_ExplicitPolicy covers the branch where the
// user set --on-failure: the policy decides without prompting, even on a TTY.
func TestDecideProceed_PreflightFailed_ExplicitPolicy(t *testing.T) {
	failed := orchestrator2.PreflightSummary{Outcome: "failed", Failed: 1}
	s := &spec.Spec{}

	newTerm := func(prompted *bool) *terminal.Terminal {
		return terminal.New(
			terminal.WithTTY(true),
			terminal.WithConfirmFunc(func(string, bool) (bool, error) { *prompted = true; return true, nil }),
		)
	}

	t.Run("skip proceeds without prompting", func(t *testing.T) {
		prompted := false
		s.OnStepFailure = spec.FailurePolicySkip
		opts := ExecOptions{ConfirmPrompt: "Proceed?", OnFailureSet: true}
		proceed, err := decideProceed(context.Background(), s, newTerm(&prompted), opts, failed)
		require.NoError(t, err)
		assert.True(t, proceed, "explicit skip proceeds past a failed preflight")
		assert.False(t, prompted, "an explicit policy must not prompt")
	})

	t.Run("abort stops without prompting", func(t *testing.T) {
		prompted := false
		s.OnStepFailure = spec.FailurePolicyAbort
		opts := ExecOptions{ConfirmPrompt: "Proceed?", OnFailureSet: true}
		proceed, err := decideProceed(context.Background(), s, newTerm(&prompted), opts, failed)
		require.NoError(t, err)
		assert.False(t, proceed, "explicit abort stops on a failed preflight")
		assert.False(t, prompted, "an explicit policy must not prompt")
	})
}
