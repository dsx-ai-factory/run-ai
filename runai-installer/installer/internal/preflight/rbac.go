package preflight

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/steps"
)

func init() { builders = append(builders, rbacBuilder) }

func rbacBuilder(_ *spec.Spec, _ orchestrator2.Deps) orchestrator2.Step {
	return steps.NewActionStepTitled("preflight-rbac", "Cluster permissions (RBAC)", actions.CheckRBAC{})
}
