/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"testing"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_sumOwnedPodGroupResources(t *testing.T) {
	type args struct {
		workloadMetadata *metadata.WorkloadMetadata
		ownedPodGroups   []kaiv2alpha2.PodGroup
	}
	type expected struct {
		workloadMetadata *metadata.WorkloadMetadata
	}
	tests := []struct {
		name     string
		args     args
		expected expected
	}{
		{
			"One pod group sum",
			args{
				&metadata.WorkloadMetadata{},
				[]kaiv2alpha2.PodGroup{
					{
						Status: kaiv2alpha2.PodGroupStatus{
							Running: 1,
							Pending: 1,
							ResourcesStatus: kaiv2alpha2.PodGroupResourcesStatus{
								Requested: v1.ResourceList{
									clustercommonconstants.NvidiaGpuResourceName: resource.MustParse("0.5"),
								},
								Allocated: v1.ResourceList{
									clustercommonconstants.NvidiaGpuResourceName: resource.MustParse("0.25"),
								},
							},
						},
					},
				},
			},
			expected{
				&metadata.WorkloadMetadata{
					RunningPods:          1,
					PendingPods:          1,
					RequestedGPUs:        0.5,
					CurrentRequestedGPUs: 0.5,
					AllocatedGpus:        0.25,
				},
			},
		},
		{
			"two pod groups sum",
			args{
				&metadata.WorkloadMetadata{},
				[]kaiv2alpha2.PodGroup{
					{
						Status: kaiv2alpha2.PodGroupStatus{
							Running: 1,
							Pending: 1,
							ResourcesStatus: kaiv2alpha2.PodGroupResourcesStatus{
								Requested: v1.ResourceList{
									clustercommonconstants.NvidiaGpuResourceName: resource.MustParse("0.5"),
								},
								Allocated: v1.ResourceList{
									clustercommonconstants.NvidiaGpuResourceName: resource.MustParse("0.25"),
								},
							},
						},
					},
					{
						Status: kaiv2alpha2.PodGroupStatus{
							Running: 2,
							Pending: 0,
							ResourcesStatus: kaiv2alpha2.PodGroupResourcesStatus{
								Requested: v1.ResourceList{
									clustercommonconstants.NvidiaGpuResourceName: resource.MustParse("1"),
								},
								Allocated: v1.ResourceList{
									clustercommonconstants.NvidiaGpuResourceName: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			expected{
				&metadata.WorkloadMetadata{
					RunningPods:          3,
					PendingPods:          1,
					RequestedGPUs:        1.5,
					CurrentRequestedGPUs: 1.5,
					AllocatedGpus:        1.25,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SumOwnedPodGroupResources(tt.args.workloadMetadata, tt.args.ownedPodGroups)
			assert.Equal(t, *tt.args.workloadMetadata, *tt.expected.workloadMetadata)
		})
	}
}
