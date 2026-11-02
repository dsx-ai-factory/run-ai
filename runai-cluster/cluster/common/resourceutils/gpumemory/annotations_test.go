/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpumemory

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestReadInvalidMemory(t *testing.T) {
	for _, tc := range []struct{ name, key, value string }{
		{"invalid quantity", "nvidia.com/container.worker.gpu-memory.request", "invalid"},
		{"zero request", "nvidia.com/container.worker.gpu-memory.request", "0"},
		{"negative limit", "nvidia.com/container.worker.gpu-memory.limit", "-1M"},
		{"missing container", "nvidia.com/container..gpu-memory.request", "500M"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, found, err := Read(map[string]string{tc.key: tc.value})
			require.True(t, found)
			require.Error(t, err)
		})
	}
}

func TestReadDefaultsPerContainer(t *testing.T) {
	annotations := map[string]string{
		"nvidia.com/container.worker.gpu-memory.request": "500M",
		"nvidia.com/container.worker.gpu-memory.limit":   "600M",
		"nvidia.com/container.other.gpu-memory.limit":    "500Mi",
		"nvidia.com/container.worker.gpu-compute.mode":   "sm-sharing",
	}
	got, found, err := Read(annotations)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, HasRequest(annotations))
	require.Zero(t, got.Request.Cmp(resource.MustParse("1024288k")))
	require.Zero(t, got.Limit.Cmp(resource.MustParse("1124288k")))
}

func TestReadIgnoresUnrelatedAnnotations(t *testing.T) {
	annotations := map[string]string{
		"gpu-memory": "500",
		"nvidia.com/container.worker.gpu-compute.mode":    "sm-sharing",
		"nvidia.com/container.worker.gpu-devices":         "GPU-example",
		"example.com/container.worker.gpu-memory.request": "500M",
	}
	_, found, err := Read(annotations)
	require.NoError(t, err)
	require.False(t, found)
	require.False(t, HasRequest(annotations))
}
