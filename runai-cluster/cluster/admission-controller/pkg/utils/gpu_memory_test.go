/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNativeGpuMemoryDetection(t *testing.T) {
	for _, tc := range []struct {
		name, key, value string
		gpus             float64
	}{
		{"request", "nvidia.com/container.worker.gpu-memory.request", "500M", 1},
		{"limit only", "nvidia.com/container.worker.gpu-memory.limit", "500Mi", 1},
		{"compute mode", "nvidia.com/container.worker.gpu-compute.mode", "sm-sharing", 0},
		{"unrelated annotation", "example.com/container.worker.gpu-memory.request", "500M", 0},
		{"invalid memory", "nvidia.com/container.worker.gpu-memory.request", "invalid", 0},
		{"zero memory", "nvidia.com/container.worker.gpu-memory.request", "0", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			annotations := map[string]string{tc.key: tc.value}
			assert.Equal(t, tc.gpus, GetRequestedGpus(nil, &metav1.ObjectMeta{Annotations: annotations}))
			assert.Equal(t, tc.gpus > 0, HasNativeGpuMemoryAnnotation(annotations))
		})
	}
}
