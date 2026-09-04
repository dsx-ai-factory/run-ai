/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	usedNodesDefault = "-"
)

func TestCalcPodsUsedNodesMetadata(t *testing.T) {
	type args struct {
		workloadMetadata *metadata.WorkloadMetadata
		relatedPods      []v1.Pod
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
			name: "Pod group with no pods",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					UsedNodes: usedNodesDefault,
				},
			},
		},
		{
			name: "Pod group with single running pod",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "t1",
							Namespace: "testNamespace",
						},
						Spec: v1.PodSpec{
							NodeName: "NodeOne",
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
				},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					UsedNodes: "NodeOne",
				},
			},
		},
		{
			name: "Pod group with single pending pod",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "t2",
							Namespace: "testNamespace",
						},
						Spec: v1.PodSpec{},
						Status: v1.PodStatus{
							Phase: v1.PodPending,
						},
					},
				},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					UsedNodes: usedNodesDefault,
				},
			},
		},
		{
			name: "Pod group with two running pods",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "t3",
							Namespace: "testNamespace",
						},
						Spec: v1.PodSpec{
							NodeName: "NodeOne",
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "t4",
							Namespace: "testNamespace",
						},
						Spec: v1.PodSpec{
							NodeName: "NodeTwo",
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
				},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					UsedNodes: "NodeOne, NodeTwo",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CalcPodsUsedNodesMetadata(tt.args.workloadMetadata, tt.args.relatedPods)
			assert.Equal(t, tt.args.workloadMetadata.UsedNodes, tt.expected.workloadMetadata.UsedNodes)
		})
	}
}

func TestSumPodGpuMemoryResources(t *testing.T) {
	type args struct {
		workloadMetadata *metadata.WorkloadMetadata
		relatedPods      []v1.Pod
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
			name: "No pods",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					RequestedGPUsMemory:        0,
					CurrentRequestedGPUsMemory: 0,
					AllocatedGPUsMemory:        0,
					UsedNodes:                  usedNodesDefault,
				},
			},
		},
		{
			name: "Single pod with no GPU memory request",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod1",
							Namespace: "testNamespace",
						},
						Spec: v1.PodSpec{
							NodeName: "NodeOne",
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
				},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					RequestedGPUsMemory:        0,
					CurrentRequestedGPUsMemory: 0,
					AllocatedGPUsMemory:        0,
					UsedNodes:                  usedNodesDefault,
				},
			},
		},
		{
			name: "Single pod with GPU memory request (not bound)",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod2",
							Namespace: "testNamespace",
							Annotations: map[string]string{
								"gpu-memory": "4096",
							},
						},
						Spec: v1.PodSpec{},
						Status: v1.PodStatus{
							Phase: v1.PodPending,
						},
					},
				},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					RequestedGPUsMemory:        4096,
					CurrentRequestedGPUsMemory: 4096,
					AllocatedGPUsMemory:        0,
					UsedNodes:                  usedNodesDefault,
				},
			},
		},
		{
			name: "Single pod with GPU memory request (bound/running)",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod3",
							Namespace: "testNamespace",
							Annotations: map[string]string{
								"gpu-memory": "8192",
							},
						},
						Spec: v1.PodSpec{
							NodeName: "NodeOne",
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
				},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					RequestedGPUsMemory:        8192,
					CurrentRequestedGPUsMemory: 8192,
					AllocatedGPUsMemory:        8192,
					UsedNodes:                  usedNodesDefault,
				},
			},
		},
		{
			name: "Multiple pods with mixed states",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod4",
							Namespace: "testNamespace",
							Annotations: map[string]string{
								"gpu-memory": "2048",
							},
						},
						Spec: v1.PodSpec{
							NodeName: "NodeOne",
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod5",
							Namespace: "testNamespace",
							Annotations: map[string]string{
								"gpu-memory": "4096",
							},
						},
						Spec: v1.PodSpec{},
						Status: v1.PodStatus{
							Phase: v1.PodPending,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod6",
							Namespace: "testNamespace",
							Annotations: map[string]string{
								"gpu-memory": "8192",
							},
						},
						Spec: v1.PodSpec{
							NodeName: "NodeTwo",
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
				},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					RequestedGPUsMemory:        14336, // 2048 + 4096 + 8192
					CurrentRequestedGPUsMemory: 14336,
					AllocatedGPUsMemory:        10240, // 2048 + 8192 (only running pods)
					UsedNodes:                  usedNodesDefault,
				},
			},
		},
		{
			name: "Pod with invalid GPU memory annotation",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod7",
							Namespace: "testNamespace",
							Annotations: map[string]string{
								"gpu-memory": "invalid",
							},
						},
						Spec: v1.PodSpec{
							NodeName: "NodeOne",
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
				},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					RequestedGPUsMemory:        0,
					CurrentRequestedGPUsMemory: 0,
					AllocatedGPUsMemory:        0,
					UsedNodes:                  usedNodesDefault,
				},
			},
		},
		{
			name: "Pods with different phases (Succeeded, Failed)",
			args: args{
				metadata.CreateWorkloadMetadata(),
				[]v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod8",
							Namespace: "testNamespace",
							Annotations: map[string]string{
								"gpu-memory": "1024",
							},
						},
						Spec: v1.PodSpec{
							NodeName: "NodeOne",
						},
						Status: v1.PodStatus{
							Phase: v1.PodSucceeded,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod9",
							Namespace: "testNamespace",
							Annotations: map[string]string{
								"gpu-memory": "2048",
							},
						},
						Spec: v1.PodSpec{
							NodeName: "NodeTwo",
						},
						Status: v1.PodStatus{
							Phase: v1.PodFailed,
						},
					},
				},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					RequestedGPUsMemory:        3072, // 1024 + 2048
					CurrentRequestedGPUsMemory: 3072,
					AllocatedGPUsMemory:        0, // Succeeded/Failed pods are not considered bound
					UsedNodes:                  usedNodesDefault,
				},
			},
		},
		{
			name: "Accumulate on pre-existing metadata",
			args: args{
				&metadata.WorkloadMetadata{
					RequestedGPUsMemory:        1000,
					CurrentRequestedGPUsMemory: 1000,
					AllocatedGPUsMemory:        500,
					UsedNodes:                  usedNodesDefault,
				},
				[]v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod10",
							Namespace: "testNamespace",
							Annotations: map[string]string{
								"gpu-memory": "2048",
							},
						},
						Spec: v1.PodSpec{
							NodeName: "NodeOne",
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
				},
			},
			expected: expected{
				&metadata.WorkloadMetadata{
					RequestedGPUsMemory:        3048, // 1000 + 2048
					CurrentRequestedGPUsMemory: 3048,
					AllocatedGPUsMemory:        2548, // 500 + 2048
					UsedNodes:                  usedNodesDefault,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SumPodGpuMemoryResources(tt.args.workloadMetadata, tt.args.relatedPods)
			assert.Equal(t, tt.expected.workloadMetadata.RequestedGPUsMemory, tt.args.workloadMetadata.RequestedGPUsMemory)
			assert.Equal(t, tt.expected.workloadMetadata.CurrentRequestedGPUsMemory, tt.args.workloadMetadata.CurrentRequestedGPUsMemory)
			assert.Equal(t, tt.expected.workloadMetadata.AllocatedGPUsMemory, tt.args.workloadMetadata.AllocatedGPUsMemory)
		})
	}
}
