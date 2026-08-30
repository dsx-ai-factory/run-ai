package topology

import (
	"context"
	"fmt"

	node_pools "github.com/run-ai/runai/scale-tests/unified_tests/internal/node-pools"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
)

// Provisioner sets up the KWOK nodepool topology for a given scale level.
// It creates NP Run:ai NodePools and NP KWOK NodePools (N/NP nodes each),
// then waits for all NodePools to reach READY phase.
type Provisioner struct {
	config  *scale.ClusterConfig
	clients *sampler_actions.Clients
}

func NewProvisioner(config *scale.ClusterConfig, clients *sampler_actions.Clients) *Provisioner {
	return &Provisioner{config: config, clients: clients}
}

func (p *Provisioner) Run(ctx context.Context) error {
	np := p.config.NodepoolCount
	nodesPerPool := p.config.NodesPerPool
	totalNodes := np * nodesPerPool

	fmt.Printf("Setting up KWOK topology: %d nodepools × %d nodes = %d total nodes\n",
		np, nodesPerPool, totalNodes)

	npConfig := node_pools.ToNodePoolConfig(p.config)
	clusterNodePools := node_pools.NewClusterNodePools(npConfig, p.clients)

	if err := clusterNodePools.Create(ctx); err != nil {
		return fmt.Errorf("failed to create nodepools: %w", err)
	}

	if err := clusterNodePools.WaitForNodepoolsReadiness(); err != nil {
		return fmt.Errorf("nodepools did not reach ready state: %w", err)
	}

	fmt.Printf("Topology ready: %d nodepools, %d nodes each\n", np, nodesPerPool)
	return nil
}
