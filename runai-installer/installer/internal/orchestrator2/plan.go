package orchestrator2

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

type Plan struct {
	Pipelines []*Pipeline
	// PreEvents are emitted (in order) before the plan's pipelines run. They
	// carry plan-time information the user should see up front, e.g. a
	// compatibility-matrix fallback warning. They are informational only.
	PreEvents []Event
}

// PlanBuilder turns a spec into two executable plans: the read-only preflight
// plan (run first, in every mode) and the install plan (the pipelines that do
// the work, gated by confirmation and dry-run). Plan construction depends on
// the actions package, so it lives outside the orchestrator2 core (injected
// via this type) to avoid an import cycle. A nil preflight plan means there is
// no preflight to run (e.g. upgrade/uninstall).
type PlanBuilder func(s *spec.Spec, deps Deps) (preflight *Plan, install *Plan, err error)
