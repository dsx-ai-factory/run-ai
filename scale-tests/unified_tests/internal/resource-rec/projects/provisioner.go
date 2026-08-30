package projects

import (
	"context"
	"fmt"
	"time"

	seed_cluster "github.com/run-ai/runai/scale-tests/unified_tests/internal/projects/seed_cluster"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
)

// Provisioner creates and waits for Run:ai projects to become READY.
// Projects are named "<prefix>-1" … "<prefix>-N" and are assigned to
// all nodepools in the cluster with Deserved=0, Limit=-1 (unlimited quota).
// The workload-submitter binary round-robins submissions across these projects.
type Provisioner struct {
	clusterCfg *scale.ClusterConfig
	count      int
	prefix     string
	clients    *sampler_actions.Clients
}

func NewProvisioner(clusterCfg *scale.ClusterConfig, count int, prefix string, clients *sampler_actions.Clients) *Provisioner {
	return &Provisioner{
		clusterCfg: clusterCfg,
		count:      count,
		prefix:     prefix,
		clients:    clients,
	}
}

func (p *Provisioner) Run(ctx context.Context) error {
	fmt.Printf("Creating %d projects with prefix %q in cluster %s...\n",
		p.count, p.prefix, p.clusterCfg.ClusterID)

	projectConfig := &seed_cluster.ProjectConfig{
		ProjectCount:           p.count,
		ControlPlaneQPS:        p.clusterCfg.ControlPlaneQPS,
		ClusterID:              p.clusterCfg.ClusterID,
		ProjectCreationTimeout: 20 * time.Minute,
		MonitorWaitTime:        10 * time.Second,
		BatchSize:              p.clusterCfg.BatchSize,
	}

	sampler := seed_cluster.NewProjectSampler(p.clusterCfg, projectConfig, p.clients, p.prefix)

	if err := sampler.Create(ctx); err != nil {
		return fmt.Errorf("project creation failed: %w", err)
	}

	succeeded := sampler.NumberOfSuccessfulCreatedObjects()
	if succeeded == 0 {
		return fmt.Errorf("no projects were created successfully")
	}
	if succeeded < p.count {
		fmt.Printf("WARNING: only %d/%d projects created successfully\n", succeeded, p.count)
	}

	fmt.Printf("Waiting for %d projects to reach READY...\n", succeeded)
	if err := sampler.Monitor(); err != nil {
		return fmt.Errorf("projects did not reach READY: %w", err)
	}

	fmt.Printf("Projects ready: %d/%d succeeded\n", succeeded, p.count)
	return nil
}
