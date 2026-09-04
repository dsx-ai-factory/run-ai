/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/runaimetrics"
	metricsClient "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetWorkloadsCountMetric(t *testing.T) {
	clusterID1 := uuid.New()
	clusterID2 := uuid.New()
	nodePool1 := "pool1"

	tests := []struct {
		name            string
		clusterIds      []uuid.UUID
		groupBy         MetricGroupBy
		expectedMetric  metricsClient.Metric
		expectedFilters any
	}{
		{
			name:           fmt.Sprintf("WorkloadCount_%s", GroupByCategory),
			groupBy:        GroupByCategory,
			clusterIds:     []uuid.UUID{clusterID1, clusterID2},
			expectedMetric: runaimetrics.RunaiWorkloadsCountMetric{},
			expectedFilters: runaimetrics.RunaiWorkloadsCountFilters{
				ClusterId: []uuid.UUID{clusterID1, clusterID2},
				Nodepool:  []string{nodePool1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := GetWorkloadsCountMetric(tt.clusterIds, tt.groupBy.ToClientGroupBy(), nodePool1)

			// Verify metric type
			assert.IsType(t, tt.expectedMetric, metric)

			// Verify filters based on metric type
			switch m := metric.(type) {
			case runaimetrics.RunaiWorkloadsCountMetric:
				assert.Equal(t, tt.expectedFilters.(runaimetrics.RunaiWorkloadsCountFilters), m.Filters)
			default:
				t.Errorf("Unexpected metric type: %T", metric)
			}
		})
	}
}
