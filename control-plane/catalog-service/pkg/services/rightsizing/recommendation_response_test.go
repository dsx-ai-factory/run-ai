// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	"github.com/stretchr/testify/require"
)

func TestRecommendationResponseContainsOnlyAllocationFields(t *testing.T) {
	capacity := int64(80_000_000_000)
	for _, tc := range []struct {
		name, precision, want string
		minimum               float64
		tp                    int
	}{
		{"shared GPU", "bf16", `{"allocationType":"gpu_memory","gpuCount":1,"gpuMemoryPerDeviceBytes":30000000000}`, 30, 1},
		{"whole GPU", "fp8", `{"allocationType":"gpu_count","gpuCount":1,"gpuMemoryPerDeviceBytes":80000000000}`, 72, 1},
		{"parallel profile", "fp8", `{"allocationType":"gpu_count","gpuCount":2,"gpuMemoryPerDeviceBytes":80000000000}`, 57, 2},
		{"NVFP4 profile", "nvfp4", `{"allocationType":"gpu_memory","gpuCount":1,"gpuMemoryPerDeviceBytes":24000000000}`, 24, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			one := 1
			profile := ngcmetadata.Profile{
				ID: "published-profile", Precision: tc.precision, InferenceBackend: "vllm",
				MinimumVRAMPerDeviceGB: &tc.minimum, TensorParallelSize: &tc.tp, PipelineParallelSize: &one,
			}
			result := recommendForHardware(gpuinventory.Hardware{
				Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8,
			}, []ngcmetadata.Profile{profile}, 90)
			require.Equal(t, "recommended", string(result.Status))
			body, err := json.Marshal(result.Recommendation)
			require.NoError(t, err)
			require.JSONEq(t, tc.want, string(body))
		})
	}
}
