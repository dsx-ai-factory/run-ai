package steps

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

// ensureClusterStep registers the cluster with the control plane. Detect
// reports it as installed when the control plane already knows the cluster by
// name.
type ensureClusterStep struct {
	*orchestrator2.BaseStep
	deps   orchestrator2.Deps
	action actions.EnsureCluster
}

func NewEnsureCluster(name string, action actions.EnsureCluster, deps orchestrator2.Deps) orchestrator2.Step {
	return &ensureClusterStep{BaseStep: orchestrator2.NewBaseStep(name), deps: deps, action: action}
}

func (s *ensureClusterStep) Detect(ctx context.Context, _ *spec.Spec) (orchestrator2.DetectResult, error) {
	apiClient, err := s.deps.RunAIClientProvider(actions.ControlPlaneURL(s.action.Domain), s.action.CACertPath)
	if err != nil {
		return orchestrator2.DetectResult{}, err
	}
	uuid, err := actions.FindClusterUUID(ctx, apiClient, s.action.ClusterName)
	if err != nil {
		return orchestrator2.DetectResult{}, err
	}
	if uuid != "" {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectInstalled}, nil
	}
	return orchestrator2.DetectResult{Phase: orchestrator2.DetectNotInstalled}, nil
}

func (s *ensureClusterStep) BuildActions(*spec.Spec) ([]orchestrator2.PerformAction, error) {
	return []orchestrator2.PerformAction{s.action}, nil
}
