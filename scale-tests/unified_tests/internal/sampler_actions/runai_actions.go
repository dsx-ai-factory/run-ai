package sampler_actions

import (
	"context"
	"fmt"
	generalconfig "github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
)

func SetClusterId(ctx context.Context, clients *Clients, clusterConfig *generalconfig.ClusterConfig) error {
	if clusterConfig.ClusterID == "" {
		clusters, _, err := clients.ClusterServiceClient.ClustersAPI.GetClusters(ctx).Execute()
		if err != nil {
			return fmt.Errorf("failed to get clusters: %w", err)
		}
		if len(clusters) == 0 {
			return fmt.Errorf("no clusters found")
		}

		clusterConfig.ClusterID = clusters[0].Uuid
	}
	return nil
}

func ProjectToNs(project string) string {
	return fmt.Sprintf("runai-%s", project)
}
