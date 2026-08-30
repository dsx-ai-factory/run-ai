// Package specrun wires the terminal, event manager, failure handler, and
// orchestrator runner around a resolved spec. It is command-agnostic: any CLI
// command builds a *spec.Spec however it likes and hands it here for execution.
// It is cobra-free so it can be driven from commands or component tests.
package specrun

import (
	"context"
	"errors"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/events"
	"github.com/run-ai/runai/runai-installer/installer/internal/failure"
	"github.com/run-ai/runai/runai-installer/installer/internal/inputcheck"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/terminal"
)

// ErrPreflightFailed is returned when preflight checks fail (or the user
// declines to continue past failures). The CLI maps it to exit code 2 to
// distinguish a preflight failure from an install failure.
var ErrPreflightFailed = errors.New("preflight checks failed")

// ExecOptions carries cross-cutting knobs that aren't part of the spec.
type ExecOptions struct {
	// PlanBuilder constructs the preflight and install plans from the spec. Each
	// command supplies its own (e.g. plan.Resolve), keeping specrun
	// command-agnostic.
	PlanBuilder orchestrator2.PlanBuilder
	// ConfirmPrompt, when non-empty, asks the user to confirm before the install
	// pipeline. Skipped for non-interactive and dry-run executions.
	ConfirmPrompt string
	// SkipPreflight bypasses the preflight pipeline (escape hatch).
	SkipPreflight bool
	// PreflightOnly runs preflight and stops before the install pipeline.
	PreflightOnly bool
	StepTimeout   time.Duration
	Timeout       time.Duration
	// FailureHandler overrides the default terminal-driven handler. When nil,
	// specrun builds one from the spec via failure.NewFromSpec.
	FailureHandler orchestrator2.FailureHandler
	// OnFailureSet reports whether --on-failure was explicitly set; when it is,
	// the input/preflight gates follow the policy instead of prompting.
	OnFailureSet bool
	// DepsBuilder, when set, constructs the orchestrator deps after offline
	// input validation passes, overriding the deps argument. The CLI uses it so
	// that cluster-touching setup (e.g. connecting to the API server) never
	// runs before a misconfigured spec is reported. When nil, the deps argument
	// is used as-is (component tests inject fakes directly).
	DepsBuilder func(context.Context) (orchestrator2.Deps, error)
}

func Execute(
	ctx context.Context,
	s *spec.Spec,
	deps orchestrator2.Deps,
	term *terminal.Terminal,
	opts ExecOptions,
) error {
	if s.Output.Format == spec.OutputFormatNone {
		if term.IsTTY() {
			s.Output.Format = spec.OutputFormatText
		} else {
			s.Output.Format = spec.OutputFormatJSON
		}
	}

	start := time.Now()

	mgr, err := events.New(events.ConfigFromSpec(s, term))
	if err != nil {
		return err
	}
	defer mgr.Close()

	// Offline input validation runs before any work. In interactive mode the
	// user is prompted to continue past the findings; otherwise the run aborts.
	if report := inputcheck.Validate(s); len(report.Findings) > 0 {
		if emitErr := mgr.HandleEvent(inputValidationEvent(report)); emitErr != nil {
			return emitErr
		}
		proceed, err := proceedPastFailure(s, term, opts, "Input validation reported problems. Continue anyway?")
		if err != nil {
			return err
		}
		if !proceed {
			return report.Err()
		}
		term.Warning("proceeding despite %d input validation problem(s)", len(report.Findings))
	}

	if err := s.Validate(); err != nil {
		return err
	}

	if err := prepareSelfSignedTLS(s, mgr); err != nil {
		return err
	}

	// Cluster-touching deps are built only after offline validation passes, so
	// an unreachable cluster never masks a misconfigured spec.
	if opts.DepsBuilder != nil {
		deps, err = opts.DepsBuilder(ctx)
		if err != nil {
			return err
		}
	}

	preflightPlan, installPlan, err := opts.PlanBuilder(s, deps)
	if err != nil {
		return err
	}

	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	failureHandler := opts.FailureHandler
	if failureHandler == nil {
		failureHandler = failure.NewFromSpec(term, s)
	}
	deps.Progress = mgr.Progress
	runner := &orchestrator2.Runner{
		Deps:           deps,
		EventSink:      mgr,
		FailureHandler: failureHandler,
		StepTimeout:    opts.StepTimeout,
	}

	// Preflight stage: read-only, runs in every mode (including dry-run). Its
	// per-check results are aggregated into a single preflight_completed event.
	var summary orchestrator2.PreflightSummary
	if preflightPlan != nil && !opts.SkipPreflight {
		if err := runner.RunPlan(ctx, preflightPlan, s); err != nil {
			return err
		}
		summary = preflightSummary(mgr.GetEvents())
		if err := mgr.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindPreflightCompleted, Preflight: &summary}); err != nil {
			return err
		}
	}

	proceed, err := decideProceed(ctx, s, term, opts, summary)
	if err != nil {
		return err
	}
	if !proceed {
		if summary.Failed > 0 {
			return ErrPreflightFailed
		}
		return nil
	}

	// Plan-time warnings (e.g. a compatibility-matrix fallback) surface just
	// before the install work they describe.
	for _, e := range installPlan.PreEvents {
		if err := mgr.HandleEvent(e); err != nil {
			return err
		}
	}

	// The install plan runs pipeline-by-pipeline so each pipeline's outcome can
	// be aggregated into a single *_completed event the moment its pipeline
	// finishes — mirroring how preflight emits its summary. The summary is
	// emitted even when a pipeline fails, so the user gets a per-component
	// verdict (including the failure) before the run aborts. Once the install
	// phase has begun, a single run_completed event is always emitted at the end
	// (success or failure) as the terminal summary of the whole run.
	var installErr error
	for _, p := range installPlan.Pipelines {
		runErr := runner.RunPlan(ctx, &orchestrator2.Plan{Pipelines: []*orchestrator2.Pipeline{p}}, s)
		if err := emitPipelineSummary(mgr, p, s); err != nil {
			return err
		}
		if runErr != nil {
			installErr = runErr
			break
		}
	}

	rs := runSummary(mgr.GetEvents(), s, time.Since(start), installErr)
	if err := mgr.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindRunCompleted, Run: &rs}); err != nil {
		return err
	}
	return installErr
}

// emitPipelineSummary emits the per-pipeline completion event (prereq/features/
// runai) once a pipeline finishes. Preflight has its own summary emitted earlier.
func emitPipelineSummary(mgr *events.Manager, p *orchestrator2.Pipeline, s *spec.Spec) error {
	if len(p.Flows) == 0 {
		return nil
	}
	switch p.Name {
	case "prereq":
		ps := prereqSummary(mgr.GetEvents())
		return mgr.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindPrereqCompleted, Prereq: &ps})
	case "features":
		fs := featuresSummary(mgr.GetEvents())
		return mgr.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindFeaturesCompleted, Features: &fs})
	case "runai":
		rs := runaiSummary(mgr.GetEvents(), s)
		return mgr.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindRunaiCompleted, Runai: &rs})
	}
	return nil
}

// decideProceed reports whether to run the install pipeline after preflight:
// PreflightOnly stops, a failed preflight is gated by proceedPastFailure, and
// otherwise the normal confirmation applies.
func decideProceed(ctx context.Context, s *spec.Spec, term *terminal.Terminal, opts ExecOptions, summary orchestrator2.PreflightSummary) (bool, error) {
	if opts.PreflightOnly {
		return false, nil
	}
	if summary.Failed > 0 {
		return proceedPastFailure(s, term, opts, "Preflight reported failures. Continue anyway?")
	}
	return confirm(ctx, s, term, opts.ConfirmPrompt)
}

// proceedPastFailure gates a pre-install failure (input validation or preflight):
// interactive with no explicit policy prompts (default No); otherwise the
// failure policy decides (skip proceeds, abort stops).
func proceedPastFailure(s *spec.Spec, term *terminal.Terminal, opts ExecOptions, prompt string) (bool, error) {
	interactive := term.IsTTY() && !s.NonInteractive
	if interactive && !opts.OnFailureSet {
		return term.Confirm(prompt, false)
	}
	return s.OnStepFailure == spec.FailurePolicySkip, nil
}

// inputValidationEvent maps an inputcheck report onto the dependency-free
// event payload the sinks render.
func inputValidationEvent(report inputcheck.Report) orchestrator2.Event {
	findings := make([]orchestrator2.ValidationFinding, len(report.Findings))
	for i, f := range report.Findings {
		findings[i] = orchestrator2.ValidationFinding{Check: f.Check, Field: f.Field, Message: f.Message}
	}
	return orchestrator2.Event{Kind: orchestrator2.EventKindInputValidationFailed, Findings: findings}
}

// confirm prompts for confirmation and reports whether to proceed. It is
// skipped (proceeds) when no prompt is configured, for non-interactive runs,
// and for dry-run. A non-TTY interactive run is a hard error rather than an
// implicit auto-confirm.
func confirm(ctx context.Context, s *spec.Spec, term *terminal.Terminal, prompt string) (bool, error) {
	if prompt == "" || s.NonInteractive || s.DryRun {
		return true, nil
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if !term.IsTTY() {
		return false, errors.New("interactive confirmation required but no TTY; re-run with --yes")
	}
	ok, err := term.Confirm(prompt, true)
	if err != nil {
		return false, err
	}
	if !ok {
		term.Warning("aborted by user")
	}
	return ok, nil
}
