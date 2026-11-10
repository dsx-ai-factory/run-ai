/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateRequestsGpuFromAnnotations(t *testing.T) {
	for _, tc := range []struct {
		name, key, value string
		gpu              bool
	}{
		{"request", "nvidia.com/container.worker.gpu-memory.request", "500M", true},
		{"limit only", "nvidia.com/container.worker.gpu-memory.limit", "500Mi", true},
		{"missing container limit", "nvidia.com/container.gpu-memory.limit", "500Mi", false},
		{"missing container request", "nvidia.com/container.gpu-memory.request", "500M", false},
		{"empty container limit", "nvidia.com/container..gpu-memory.limit", "500Mi", false},
		{"empty container request", "nvidia.com/container..gpu-memory.request", "500M", false},
		{"zero request", "nvidia.com/container.worker.gpu-memory.request", "0", false},
		{"negative limit", "nvidia.com/container.worker.gpu-memory.limit", "-1M", false},
		{"invalid quantity", "nvidia.com/container.worker.gpu-memory.request", "invalid", false},
		{"compute mode only", "nvidia.com/container.worker.gpu-compute.mode", "sm-sharing", false},
		{"unrelated annotation", "example.com/container.worker.gpu-memory.request", "500M", false},
		{"legacy memory", "gpu-memory", "500", true},
		{"legacy fraction", "gpu-fraction", "0.25", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			template := map[string]any{"spec": map[string]any{"template": map[string]any{
				"metadata": map[string]any{"annotations": map[string]any{tc.key: tc.value}},
			}}}
			assert.Equal(t, tc.gpu, templateRequestsGpu(template))
		})
	}
}
