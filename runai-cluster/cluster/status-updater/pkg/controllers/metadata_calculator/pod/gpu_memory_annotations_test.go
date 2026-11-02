/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNativeGpuMemoryAllocation(t *testing.T) {
	for _, tc := range []struct {
		name, key, value string
		memoryMb         int64
		gpus             float64
	}{
		{"decimal", "nvidia.com/container.worker.gpu-memory.request", "500M", 500, 500e6 / (16384 * 1048576)},
		{"binary", "nvidia.com/container.worker.gpu-memory.request", "500Mi", 524, 500.0 / 16384},
		{"limit only", "nvidia.com/container.worker.gpu-memory.limit", "500M", 500, 500e6 / (16384 * 1048576)},
		{"sub-megabyte", "nvidia.com/container.worker.gpu-memory.request", "500k", 0, 500e3 / (16384 * 1048576)},
		{"legacy", "gpu-memory", "500", 500, 500e6 / (16384 * 1048576)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{tc.key: tc.value, receivedResourceAnnotationName: receivedTypeFraction}},
				Spec:       v1.PodSpec{NodeName: "gpu-node"},
				Status:     v1.PodStatus{Phase: v1.PodRunning},
			}
			node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "gpu-node", Labels: map[string]string{nodeGpuMemoryLabelName: "16384"}}}
			client := fake.NewClientBuilder().WithObjects(node).Build()
			metrics := &metadata.PodMetadata{}
			require.NoError(t, CalcPodGpuResources(context.Background(), pod, metrics, client))
			require.Equal(t, tc.memoryMb, metrics.GpuMemory)
			require.InDelta(t, tc.gpus, metrics.GPUs, 1e-12)
		})
	}
}

func TestNativeGpuMemoryTakesPrecedenceOverLegacyFraction(t *testing.T) {
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
		"nvidia.com/container.worker.gpu-memory.request": "500M",
		gpuFractionAnnotation:                            "0.75",
	}}}
	require.Zero(t, GetGpuRequest(pod))
	require.EqualValues(t, 500, GetGpuMemoryRequest(pod))
}
