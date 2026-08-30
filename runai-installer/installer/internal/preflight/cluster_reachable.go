package preflight

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/steps"
)

func init() { builders = append(builders, clusterReachableBuilder) }

func clusterReachableBuilder(_ *spec.Spec, _ orchestrator2.Deps) orchestrator2.Step {
	return steps.NewActionStepTitled("preflight-cluster-reachable", "Cluster reachability", actions.CheckClusterReachable{})
}
