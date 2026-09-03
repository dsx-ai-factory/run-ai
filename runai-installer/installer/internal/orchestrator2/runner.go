/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

type Runner struct {
	Deps           Deps
	EventSink      EventSink
	FailureHandler FailureHandler
	StepTimeout    time.Duration
	OverallTimeout time.Duration
}

func (r *Runner) RunPlan(ctx context.Context, p *Plan, s *spec.Spec) error {
	if r.EventSink == nil {
		return errors.New("runner: EventSink is required")
	}
	if r.FailureHandler == nil {
		return errors.New("runner: FailureHandler is required")
	}
	if r.OverallTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.OverallTimeout)
		defer cancel()
	}
	for _, pipeline := range p.Pipelines {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := r.runPipeline(ctx, pipeline, s); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) emit(e Event) error {
	return r.EventSink.HandleEvent(e)
}

func (r *Runner) runPipeline(ctx context.Context, p *Pipeline, s *spec.Spec) error {
	if err := r.emit(Event{Kind: EventKindPipelineStarted, Pipeline: p.Name, PipelineTitle: p.DisplayTitle()}); err != nil {
		return err
	}
	runErr := r.runPipelineBody(ctx, p, s)
	return r.finishPipeline(p, runErr)
}

func (r *Runner) runPipelineBody(ctx context.Context, p *Pipeline, s *spec.Spec) error {
	for _, f := range p.Flows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := r.runFlow(ctx, p, f, s); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) finishPipeline(p *Pipeline, runErr error) error {
	kind := EventKindPipelineSucceeded
	if runErr != nil {
		kind = EventKindPipelineFailed
	}
	if emitErr := r.emit(Event{Kind: kind, Pipeline: p.Name, PipelineTitle: p.DisplayTitle(), Error: runErr}); emitErr != nil {
		return emitErr
	}
	return runErr
}

func (r *Runner) runFlow(ctx context.Context, p *Pipeline, f *Flow, s *spec.Spec) error {
	if err := r.emit(Event{Kind: EventKindFlowStarted, Pipeline: p.Name, Flow: f.Name, FlowTitle: f.DisplayTitle()}); err != nil {
		return err
	}
	failedStep, runErr := r.runFlowBody(ctx, p, f, s)
	return r.finishFlow(p, f, failedStep, runErr)
}

func (r *Runner) runFlowBody(ctx context.Context, p *Pipeline, f *Flow, s *spec.Spec) (Step, error) {
	for _, step := range f.Steps {
		if err := ctx.Err(); err != nil {
			return step, err
		}
		if err := r.runStep(ctx, p, f, step, s); err != nil {
			return step, err
		}
	}
	return nil, nil
}

func (r *Runner) finishFlow(p *Pipeline, f *Flow, failedStep Step, runErr error) error {
	if runErr != nil {
		e := Event{Kind: EventKindFlowFailed, Pipeline: p.Name, Flow: f.Name, FlowTitle: f.DisplayTitle(), Error: runErr}
		if failedStep != nil {
			e.Step = failedStep.Name()
			e.StepTitle = failedStep.Title()
		}
		if emitErr := r.emit(e); emitErr != nil {
			return emitErr
		}
		return runErr
	}
	return r.emit(Event{Kind: EventKindFlowSucceeded, Pipeline: p.Name, Flow: f.Name, FlowTitle: f.DisplayTitle()})
}

func (r *Runner) runStep(ctx context.Context, p *Pipeline, f *Flow, step Step, s *spec.Spec) error {
	name := step.Name()
	title := step.Title()
	if err := r.emit(Event{Kind: EventKindStepStarted, Pipeline: p.Name, Flow: f.Name, Step: name, StepTitle: title}); err != nil {
		return err
	}

	for {
		outcome, phase, err := r.executeStepAttempt(ctx, p, step, s)
		if err == nil {
			if outcome.Skipped {
				return r.emit(Event{
					Kind:       EventKindStepSkipped,
					Pipeline:   p.Name,
					Flow:       f.Name,
					Step:       name,
					StepTitle:  title,
					SkipReason: outcome.SkipReason,
					Output:     outcome.Output,
				})
			}
			return r.emit(Event{Kind: EventKindStepSucceeded, Pipeline: p.Name, Flow: f.Name, Step: name, StepTitle: title, Output: step.Output(), Results: outcome.Results})
		}

		// The prompt uses the parent ctx, not the step-timeout ctx, so time
		// spent waiting on the operator does not count against the next attempt.
		decision := r.FailureHandler.OnFailure(ctx, FailureInfo{
			Pipeline: p.Name,
			Flow:     f.Name,
			Step:     name,
			Phase:    phase,
			Err:      err,
		})
		switch decision {
		case FailureRetry:
			// No cap: the handler bounds retries (interactive is human-driven;
			// the non-interactive handler never returns retry).
			continue
		case FailureSkip:
			return r.emit(Event{Kind: EventKindStepSkipped, Pipeline: p.Name, Flow: f.Name, Step: name, StepTitle: title, SkipReason: SkipReasonUserSelected, Error: err, Output: phase})
		default:
			return r.emitStepFailed(p, f, step, phase, err)
		}
	}
}

// executeStepAttempt runs one attempt of a step under a fresh per-attempt step
// timeout. Creating the timeout here (rather than once around the retry loop)
// means each retry gets the full StepTimeout instead of inheriting the elapsed
// deadline of the previous attempt.
func (r *Runner) executeStepAttempt(ctx context.Context, p *Pipeline, step Step, s *spec.Spec) (stepOutcome, string, error) {
	if r.StepTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.StepTimeout+step.ExtraTimeout())
		defer cancel()
	}
	return r.executeStepPhases(ctx, p, step, s)
}

func (r *Runner) emitStepFailed(p *Pipeline, f *Flow, step Step, phase string, stepErr error) error {
	if emitErr := r.emit(Event{Kind: EventKindStepFailed, Pipeline: p.Name, Flow: f.Name, Step: step.Name(), StepTitle: step.Title(), Error: stepErr, Output: phase}); emitErr != nil {
		return emitErr
	}
	return stepErr
}

func (r *Runner) executeStepPhases(ctx context.Context, p *Pipeline, step Step, s *spec.Spec) (stepOutcome, string, error) {
	if err := step.PreValidate(s); err != nil {
		return stepOutcome{}, "prevalidate", err
	}

	// Detect runs in every mode (it is read-only) so dry-run can report which
	// steps would be skipped as already-installed versus actually run. A detect
	// failure is fatal on a real run, but in dry-run it is non-fatal: a step
	// whose detection depends on an earlier step that has not been performed yet
	// is simply reported as "would run".
	detected, detectErr := step.Detect(ctx, s)
	if detectErr != nil && !p.DryRun {
		return stepOutcome{}, "detect", detectErr
	}
	if detectErr == nil && detected.Phase == DetectInstalled {
		output := "already installed"
		if detected.Version != "" {
			output = fmt.Sprintf("already installed (version %s)", detected.Version)
		}
		return stepOutcome{Skipped: true, SkipReason: SkipReasonAlreadyInstalled, Output: output}, "", nil
	}

	actions, err := step.BuildActions(s)
	if err != nil {
		return stepOutcome{}, "build_actions", err
	}

	if p.DryRun {
		return stepOutcome{Skipped: true, SkipReason: SkipReasonDryRun, Output: describeActions(actions)}, "", nil
	}

	// A returned Go error means the action could not run (a step failure). A nil
	// error with a result is a verdict: the action ran successfully and is
	// reporting an outcome — including a StatusFail verdict (e.g. a preflight
	// check that ran fine but found the cluster unsuitable). Such a step is
	// therefore reported as step_succeeded carrying a failed result, not
	// step_failed; callers aggregate the verdicts (see preflightSummary).
	var results []ActionResult
	for _, a := range actions {
		res, err := a.Execute(ctx, r.Deps)
		if err != nil {
			return stepOutcome{}, "execute_" + a.Kind(), err
		}
		if !res.IsZero() {
			results = append(results, res)
		}
	}
	if err := step.PostValidate(ctx, s); err != nil {
		return stepOutcome{}, "postvalidate", err
	}
	step.PostPerform(s)
	return stepOutcome{Results: results}, "", nil
}
