/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateHasNativeGpuMemory(t *testing.T) {
	for _, tc := range []struct {
		name, key, value string
		gpu              bool
	}{
		{"request", "nvidia.com/container.worker.gpu-memory.request", "500M", true},
		{"limit only", "nvidia.com/container.worker.gpu-memory.limit", "500Mi", true},
		{"compute mode only", "nvidia.com/container.worker.gpu-compute.mode", "sm-sharing", false},
		{"unrelated annotation", "example.com/container.worker.gpu-memory.request", "500M", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			template := map[string]any{"spec": map[string]any{"template": map[string]any{
				"metadata": map[string]any{"annotations": map[string]any{tc.key: tc.value}},
			}}}
			assert.Equal(t, tc.gpu, templateRequestsGpu(template))
		})
	}
}
