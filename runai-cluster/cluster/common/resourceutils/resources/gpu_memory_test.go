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
		{"per-container limit takes precedence", map[string]string{requestKey: "500M", limitKey: "525M", constants.AnnotationGpuMemory: "500"}, "1G", "500M", "525M"},
		{"legacy request", map[string]string{constants.AnnotationGpuMemory: "500"}, "", "500M", "500M"},
		{"legacy env limit", map[string]string{constants.AnnotationGpuMemory: "500"}, "600M", "500M", "600M"},
		{"legacy request with larger per-container limit", map[string]string{constants.AnnotationGpuMemory: "500", limitKey: "600M"}, "", "500M", "600M"},
		{"legacy request with equal per-container limit", map[string]string{constants.AnnotationGpuMemory: "500", limitKey: "500M"}, "", "500M", "500M"},
		{"legacy request with binary per-container limit", map[string]string{constants.AnnotationGpuMemory: "524", limitKey: "600Mi"}, "", "524M", "600Mi"},
		{"KAI-added request preserves legacy accounting", map[string]string{constants.AnnotationGpuMemory: "500", requestKey: "500Mi", limitKey: "600M"}, "", "500M", "600M"},
		{"per-container limit overrides legacy env limit", map[string]string{constants.AnnotationGpuMemory: "500", limitKey: "600M"}, "1G", "500M", "600M"},
		{"per-container request with legacy env limit", map[string]string{requestKey: "500M"}, "600M", "500M", "600M"},
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

func TestReadGPUMemoryAnnotationsInvalid(t *testing.T) {
	for _, tc := range []struct{ name, key, value string }{
		{"invalid quantity", "nvidia.com/container.worker.gpu-memory.request", "invalid"},
		{"zero request", "nvidia.com/container.worker.gpu-memory.request", "0"},
		{"negative limit", "nvidia.com/container.worker.gpu-memory.limit", "-1M"},
		{"missing container", "nvidia.com/container..gpu-memory.request", "500M"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ReadGPUMemoryAnnotations(map[string]string{tc.key: tc.value})
			require.Error(t, err)
		})
	}
}

func TestReadGPUMemoryAnnotationsDefaultsPerContainer(t *testing.T) {
	annotations := map[string]string{
		"nvidia.com/container.worker.gpu-memory.request": "500M",
		"nvidia.com/container.worker.gpu-memory.limit":   "600M",
		"nvidia.com/container.other.gpu-memory.limit":    "500Mi",
		"nvidia.com/container.worker.gpu-compute.mode":   "sm-sharing",
	}
	got, err := ReadGPUMemoryAnnotations(annotations)
	require.NoError(t, err)
	require.True(t, got.HasRequest)
	require.True(t, HasPerContainerGPUMemoryAnnotations(annotations))
	require.Zero(t, got.Request.Cmp(resource.MustParse("1024288k")))
	require.Zero(t, got.Limit.Cmp(resource.MustParse("1124288k")))
}

func TestReadGPUMemoryAnnotationsIgnoresUnrelated(t *testing.T) {
	annotations := map[string]string{
		"nvidia.com/container.worker.gpu-compute.mode":    "sm-sharing",
		"nvidia.com/container.worker.gpu-devices":         "GPU-example",
		"example.com/container.worker.gpu-memory.request": "500M",
	}
	got, err := ReadGPUMemoryAnnotations(annotations)
	require.NoError(t, err)
	require.False(t, got.HasRequest)
	require.False(t, HasPerContainerGPUMemoryAnnotations(annotations))
}

func TestPerContainerMemoryDoesNotReplaceLegacyFraction(t *testing.T) {
	for _, withRequest := range []bool{false, true} {
		t.Run(map[bool]string{false: "limit only", true: "request and limit"}[withRequest], func(t *testing.T) {
			annotations := map[string]string{
				constants.AnnotationGpuFraction:                "0.25",
				"nvidia.com/container.worker.gpu-memory.limit": "600M",
			}
			if withRequest {
				annotations["nvidia.com/container.worker.gpu-memory.request"] = "500M"
			}
			requests, limits := PodRunaiRequestsAndLimits(nil, nil, annotations)
			gpu := requests[constants.NvidiaGpuResourceName]
			memory := requests[constants.GpuMemoryResourceName]
			limit := limits[constants.GpuMemoryResourceName]
			require.Zero(t, gpu.Cmp(resource.MustParse("0.25")))
			require.True(t, memory.IsZero())
			require.Zero(t, limit.Cmp(resource.MustParse("600M")))
		})
	}
}

func TestInvalidMemoryRequestDoesNotDefaultFromLimit(t *testing.T) {
	for _, requestKey := range []string{constants.AnnotationGpuMemory, "nvidia.com/container.worker.gpu-memory.request"} {
		t.Run(requestKey, func(t *testing.T) {
			_, err := ReadGPUMemoryAnnotations(map[string]string{
				requestKey: "invalid",
				"nvidia.com/container.worker.gpu-memory.limit": "600M",
			})
			require.Error(t, err)
		})
	}
}
