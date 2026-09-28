/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"testing"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/topology"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/stretchr/testify/require"
)

func TestDefaultTopologyAnnotations(t *testing.T) {
	for _, tc := range []struct {
		name             string
		defaultNodePools []string
		annotations      map[string]interface{}
		selectedPool     string
		wantAnnotations  map[string]interface{}
		wantError        string
	}{
		{
			name:             "accepts defaults with and without topology",
			defaultNodePools: []string{"pool-a", "pool-b"},
		},
		{
			name:             "accepts mixed defaults in reverse order",
			defaultNodePools: []string{"pool-b", "pool-a"},
		},
		{
			name:             "accepts defaults with different topologies",
			defaultNodePools: []string{"pool-a", "pool-c"},
		},
		{
			name:             "preserves explicit node selection with mixed defaults",
			defaultNodePools: []string{"pool-a", "pool-b"},
			selectedPool:     "pool-b",
		},
		{
			name:             "preserves explicitly requested topology with mixed defaults",
			defaultNodePools: []string{"pool-a", "pool-b"},
			annotations: map[string]interface{}{
				topology.KAITopologyAnnotationKey:         "requested-topology",
				topology.KAITopologyAnnotationRequiredKey: "zone",
			},
			wantAnnotations: map[string]interface{}{
				topology.KAITopologyAnnotationKey:         "requested-topology",
				topology.KAITopologyAnnotationRequiredKey: "zone",
			},
		},
		{
			name:             "retains a common topology default",
			defaultNodePools: []string{"pool-a", "pool-d"},
			wantAnnotations: map[string]interface{}{
				topology.KAITopologyAnnotationKey:          "topology-a",
				topology.KAITopologyAnnotationPreferredKey: "rack",
			},
		},
		{
			name:             "still rejects missing node pools",
			defaultNodePools: []string{"pool-a", "missing"},
			wantError:        "the following node pools do not exist: missing",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const clusterID = "topology-cluster"
			client := clusterservice.NewSimpleMock().
				WithNodePoolsWithTopology(clusterID, "topology-a", "pool-a", "pool-d").
				WithNodePools(clusterID, "pool-b").
				WithNodePoolsWithTopology(clusterID, "topology-c", "pool-c").
				WithTopology("topology-a", &clusterApi.NetworkTopology{Levels: []string{"zone", "rack"}})
			clusters.SetMock(client)
			requestedAnnotations := make(map[string]interface{})
			for name, value := range tc.annotations {
				requestedAnnotations[name] = value
			}
			podSpec := map[string]interface{}{
				"restartPolicy": "Never",
				"containers": []interface{}{
					map[string]interface{}{"name": "training", "image": "busybox"},
				},
			}
			if tc.selectedPool != "" {
				podSpec["nodeSelector"] = map[string]interface{}{"example.com/pool": tc.selectedPool}
			}
			manifest := map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name":        "training",
					"annotations": requestedAnnotations,
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": podSpec,
					},
				},
			}
			project := &orgUnitApi.Project{
				Id:               "project-id",
				Name:             "project",
				ClusterId:        clusterID,
				DefaultNodePools: tc.defaultNodePools,
			}

			effective, err := NewK8sResourceUtil().UpdateResourceMetadata(context.Background(), manifest, "workload-id", project)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			annotations := effective["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
			wantAnnotations := map[string]interface{}{WorkloadIdAnnotation: "workload-id"}
			for name, value := range tc.wantAnnotations {
				wantAnnotations[name] = value
			}
			require.Equal(t, wantAnnotations, annotations)
			require.Equal(t, manifest["spec"], effective["spec"])
		})
	}
}
