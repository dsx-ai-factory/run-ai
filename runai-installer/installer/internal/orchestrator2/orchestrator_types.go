package orchestrator2

import (
	"context"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/display"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

type Pipeline struct {
	Name string
	// Title is the human-readable label shown in the text UI. When empty the
	// name is humanized. It never affects JSON output or control flow.
	Title  string
	DryRun bool
	Flows  []*Flow
}

// DisplayTitle returns the pipeline's human label (custom Title, or humanized
// Name).
func (p *Pipeline) DisplayTitle() string { return display.Or(p.Title, p.Name) }

type Flow struct {
	Name string
	// Title is the human-readable label shown in the text UI. When empty the
	// name is humanized. It never affects JSON output or control flow.
	Title string
	Steps []Step
}

// DisplayTitle returns the flow's human label (custom Title, or humanized Name).
func (f *Flow) DisplayTitle() string { return display.Or(f.Title, f.Name) }

type Step interface {
	Name() string
	// Title is the human-readable label shown in the text UI. When no custom
	// title is set it defaults to the humanized name. It never affects JSON
	// output or control flow.
	Title() string
	PreValidate(s *spec.Spec) error
	Detect(ctx context.Context, s *spec.Spec) (DetectResult, error)
	BuildActions(s *spec.Spec) ([]PerformAction, error)
	PostValidate(ctx context.Context, s *spec.Spec) error
	PostPerform(s *spec.Spec)
	Output() string
	// ExtraTimeout is added to the runner's per-step timeout for this step.
	// Zero (the default) leaves the configured timeout unchanged.
	ExtraTimeout() time.Duration
}

type stepOutcome struct {
	Skipped    bool
	SkipReason SkipReason
	Output     string
	Results    []ActionResult
}

type BaseStep struct {
	name         string
	title        string
	extraTimeout time.Duration
}

func NewBaseStep(name string) *BaseStep {
	return &BaseStep{name: name}
}

// NewBaseStepTitled builds a step with a custom human-readable title.
func NewBaseStepTitled(name, title string) *BaseStep {
	return &BaseStep{name: name, title: title}
}

func (b *BaseStep) Name() string { return b.name }

// Title returns the custom title when set, otherwise the humanized name.
func (b *BaseStep) Title() string { return display.Or(b.title, b.name) }

// SetTitle assigns a custom human-readable title, used by the steps.Titled
// helper so the plan can label a step without changing its constructor.
func (b *BaseStep) SetTitle(title string) { b.title = title }

func (b *BaseStep) PreValidate(s *spec.Spec) error { return nil }

func (b *BaseStep) Detect(ctx context.Context, s *spec.Spec) (DetectResult, error) {
	return DetectResult{Phase: DetectNotInstalled}, nil
}

func (b *BaseStep) BuildActions(s *spec.Spec) ([]PerformAction, error) { return nil, nil }

func (b *BaseStep) PostValidate(ctx context.Context, s *spec.Spec) error { return nil }

func (b *BaseStep) PostPerform(s *spec.Spec) {}

func (b *BaseStep) Output() string { return "" }

func (b *BaseStep) ExtraTimeout() time.Duration { return b.extraTimeout }

// SetExtraTimeout sets the step's extra timeout, used by steps.WithExtraTimeout.
func (b *BaseStep) SetExtraTimeout(d time.Duration) { b.extraTimeout = d }
