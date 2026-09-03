package testing

import "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

const (
	UnitTestDefaultClusterVersion = "9.9.9"

	OldClustersPreK8sObjectsTracker = "2.15.0"
)

func GetDefaultClustersCacheConfig() *checker.ClustersCacheConfig {
	return &checker.ClustersCacheConfig{ExpirationSec: 600, CleanupSec: 10}
}
