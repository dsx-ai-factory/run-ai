package orchestrator2

import (
	"context"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopSink struct{}

func (noopSink) HandleEvent(Event) error { return nil }

type abortHandler struct{}

func (abortHandler) OnFailure(context.Context, FailureInfo) FailureDecision { return FailureAbort }

// deadlineCapture records the context deadline it is executed with.
type deadlineCapture struct {
	deadline time.Time
	ok       bool
}

func (a *deadlineCapture) Kind() string         { return "test.deadline" }
func (a *deadlineCapture) Args() map[string]any { return nil }
func (a *deadlineCapture) Describe() string     { return "capture deadline" }
func (a *deadlineCapture) Execute(ctx context.Context, _ Deps) (ActionResult, error) {
	a.deadline, a.ok = ctx.Deadline()
	return ActionResult{}, nil
}

type actionStepStub struct {
	*BaseStep
	acts []PerformAction
}

func (s *actionStepStub) BuildActions(*spec.Spec) ([]PerformAction, error) { return s.acts, nil }

func runCapturingDeadline(t *testing.T, stepTimeout, extra time.Duration) time.Duration {
	t.Helper()
	act := &deadlineCapture{}
	step := &actionStepStub{BaseStep: NewBaseStep("s"), acts: []PerformAction{act}}
	step.SetExtraTimeout(extra)

	r := &Runner{EventSink: noopSink{}, FailureHandler: abortHandler{}, StepTimeout: stepTimeout}
	p := &Plan{Pipelines: []*Pipeline{{Name: "p", Flows: []*Flow{{Name: "f", Steps: []Step{step}}}}}}

	start := time.Now()
	require.NoError(t, r.RunPlan(context.Background(), p, &spec.Spec{}))
	require.True(t, act.ok, "the step must run under a deadline")
	return act.deadline.Sub(start)
}

// TestRunner_StepTimeoutAddsExtraTimeout verifies the runner grants a step its
// configured timeout plus the step's ExtraTimeout, in any pipeline.
func TestRunner_StepTimeoutAddsExtraTimeout(t *testing.T) {
	base := runCapturingDeadline(t, 5*time.Minute, 0)
	assert.InDelta(t, (5 * time.Minute).Seconds(), base.Seconds(), 30, "no extra timeout by default")

	extended := runCapturingDeadline(t, 5*time.Minute, 30*time.Minute)
	assert.InDelta(t, (35 * time.Minute).Seconds(), extended.Seconds(), 30, "extra timeout is added to the configured timeout")
}
