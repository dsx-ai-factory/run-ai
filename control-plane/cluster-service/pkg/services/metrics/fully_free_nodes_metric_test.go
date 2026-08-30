package metrics

import (
	"testing"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/runaimetrics"
	metricsClient "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	"github.com/stretchr/testify/assert"
)

func TestGetFullyFreeGpuNodesMetric(t *testing.T) {
	clusterID1 := uuid.New()
	clusterID2 := uuid.New()
	nodePool1 := "pool1"
	nodePool2 := "pool2"

	tests := []struct {
		name            string
		clusterIds      []uuid.UUID
		nodePools       []string
		groupBy         MetricGroupBy
		expectedMetric  metricsClient.Metric
		expectedFilters interface{}
	}{
		{
			name:           "GroupByNode",
			clusterIds:     []uuid.UUID{clusterID1, clusterID2},
			nodePools:      []string{nodePool1, nodePool2},
			groupBy:        GroupByNode,
			expectedMetric: runaimetrics.RunaiFullyFreeGpuNodeCountPerNodeMetric{},
			expectedFilters: runaimetrics.RunaiFullyFreeGpuNodeCountPerNodeFilters{
				ClusterId: []uuid.UUID{clusterID1, clusterID2},
				Nodepool:  []string{nodePool1, nodePool2},
			},
		},
		{
			name:           "GroupByNodePool",
			clusterIds:     []uuid.UUID{clusterID1},
			nodePools:      []string{nodePool1},
			groupBy:        GroupByNodePool,
			expectedMetric: runaimetrics.RunaiFullyFreeGpuNodeCountPerNodepoolMetric{},
			expectedFilters: runaimetrics.RunaiFullyFreeGpuNodeCountPerNodepoolFilters{
				ClusterId: []uuid.UUID{clusterID1},
				Nodepool:  []string{nodePool1},
			},
		},
		{
			name:           "GroupByCluster",
			clusterIds:     []uuid.UUID{clusterID1, clusterID2},
			nodePools:      []string{nodePool1, nodePool2},
			groupBy:        GroupByCluster,
			expectedMetric: runaimetrics.RunaiFullyFreeGpuNodeCountPerClusterMetric{},
			expectedFilters: runaimetrics.RunaiFullyFreeGpuNodeCountPerClusterFilters{
				ClusterId: []uuid.UUID{clusterID1, clusterID2},
			},
		},
		{
			name:           "GroupByTenant",
			clusterIds:     []uuid.UUID{clusterID1},
			nodePools:      []string{nodePool1},
			groupBy:        GroupByTenant,
			expectedMetric: runaimetrics.RunaiFullyFreeGpuNodeCountPerTenantMetric{},
			expectedFilters: runaimetrics.RunaiFullyFreeGpuNodeCountPerTenantFilters{
				ClusterId: []uuid.UUID{clusterID1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := GetFullyFreeGpuNodesMetric(tt.clusterIds, tt.nodePools, tt.groupBy)

			// Verify metric type
			assert.IsType(t, tt.expectedMetric, metric)

			// Verify filters based on metric type
			switch m := metric.(type) {
			case runaimetrics.RunaiFullyFreeGpuNodeCountPerNodeMetric:
				assert.Equal(t, tt.expectedFilters.(runaimetrics.RunaiFullyFreeGpuNodeCountPerNodeFilters), m.Filters)
			case runaimetrics.RunaiFullyFreeGpuNodeCountPerNodepoolMetric:
				assert.Equal(t, tt.expectedFilters.(runaimetrics.RunaiFullyFreeGpuNodeCountPerNodepoolFilters), m.Filters)
			case runaimetrics.RunaiFullyFreeGpuNodeCountPerClusterMetric:
				assert.Equal(t, tt.expectedFilters.(runaimetrics.RunaiFullyFreeGpuNodeCountPerClusterFilters), m.Filters)
			case runaimetrics.RunaiFullyFreeGpuNodeCountPerTenantMetric:
				assert.Equal(t, tt.expectedFilters.(runaimetrics.RunaiFullyFreeGpuNodeCountPerTenantFilters), m.Filters)
			default:
				t.Errorf("Unexpected metric type: %T", metric)
			}
		})
	}
}
