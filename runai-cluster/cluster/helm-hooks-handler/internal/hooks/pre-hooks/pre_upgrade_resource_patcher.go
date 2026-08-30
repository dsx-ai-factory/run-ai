package pre_hooks

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	dynamicupgrades "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/dynamic-schema-upgrades"
)

type ResourceUpgradeGroup interface {
	Upgrade(ctx context.Context, clients *clients.Clients) error
}

type resourceUpgrades struct {
	upgrade []ResourceUpgradeGroup
}

func NewResourceGroupUpgrader() *resourceUpgrades {
	var upgradeGroups []ResourceUpgradeGroup
	upgradeGroups = append(upgradeGroups, dynamicupgrades.NewRunaiConfigUpgrade())
	return &resourceUpgrades{upgradeGroups}
}

func (r *resourceUpgrades) Upgrade(ctx context.Context, clients *clients.Clients) error {
	for _, upgradeGroup := range r.upgrade {
		err := upgradeGroup.Upgrade(ctx, clients)
		if err != nil {
			return err
		}
	}
	return nil
}
