/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodGpuMemoryRequestsAndLimits(t *testing.T) {
	const requestKey = "nvidia.com/container.worker.gpu-memory.request"
	const limitKey = "nvidia.com/container.worker.gpu-memory.limit"
	for _, tc := range []struct {
		name        string
		annotations map[string]string
		envLimit    string
		request     string
		limit       string
	}{
		{"equal decimal", map[string]string{requestKey: "500M", limitKey: "500M"}, "", "500M", "500M"},
		{"equal binary", map[string]string{requestKey: "500Mi", limitKey: "500Mi"}, "", "500Mi", "500Mi"},
		{"larger limit", map[string]string{requestKey: "500M", limitKey: "525M"}, "", "500M", "525M"},
		{"request only", map[string]string{requestKey: "500Mi"}, "", "500Mi", "500Mi"},
		{"limit only defaults request", map[string]string{limitKey: "500Mi"}, "", "500Mi", "500Mi"},
		{"native takes precedence", map[string]string{requestKey: "500M", limitKey: "525M", constants.AnnotationGpuMemory: "500"}, "1G", "500M", "525M"},
		{"legacy request", map[string]string{constants.AnnotationGpuMemory: "500"}, "", "500M", "500M"},
		{"legacy env limit", map[string]string{constants.AnnotationGpuMemory: "500"}, "600M", "500M", "600M"},
		{"compute mode alone is not a request", map[string]string{"nvidia.com/container.worker.gpu-compute.mode": "sm-sharing"}, "", "0", "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			container := v1.Container{Name: "worker"}
			if tc.envLimit != "" {
				container.Env = []v1.EnvVar{{Name: constants.GpuMemoryLimitEnvVar, Value: tc.envLimit}}
			}
			requests, limits := PodRunaiRequestsAndLimits([]v1.Container{container}, nil, tc.annotations)
			request, limit := requests[constants.GpuMemoryResourceName], limits[constants.GpuMemoryResourceName]
			require.Zero(t, request.Cmp(resource.MustParse(tc.request)), "request = %s", request.String())
			require.Zero(t, limit.Cmp(resource.MustParse(tc.limit)), "limit = %s", limit.String())
		})
	}
}

func TestNativeGpuMemoryFractionDetection(t *testing.T) {
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
		"nvidia.com/container.worker.gpu-memory.request": "500M",
	}}}
	require.True(t, RequestsGPUFraction(pod))
	devices, err := GetNumGPUFractionDevices(pod)
	require.NoError(t, err)
	require.EqualValues(t, 1, devices)
	memory, err := GetGPUMemory(pod)
	require.NoError(t, err)
	require.EqualValues(t, 500, memory)
}
