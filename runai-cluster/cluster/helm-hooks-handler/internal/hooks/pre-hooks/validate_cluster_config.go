package pre_hooks

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-actions"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

// ValidateClusterConfigHook validates that clusterConfig (from hooks params) can be applied to RunaiConfig
// and fails fast if it cannot be unmarshaled/merged.
type ValidateClusterConfigHook struct {
	params *hooks.ClusterHookParams
}

func NewValidateClusterConfigHook(params *hooks.ClusterHookParams) *ValidateClusterConfigHook {
	return &ValidateClusterConfigHook{params: params}
}

func (p *ValidateClusterConfigHook) GetIdentifier() string {
	return "validate-cluster-config"
}

func (p *ValidateClusterConfigHook) Call(ctx context.Context, _ *clients.Clients) error {
	if len(p.params.ClusterConfig) == 0 {
		return nil
	}
	dummy := &runaiv1.RunaiConfig{}
	if err := ra.PatchRunaiConfig(dummy, p.params.ClusterConfig); err != nil {
		return fmt.Errorf("invalid clusterConfig in Helm values: %w", err)
	}
	return nil
}
