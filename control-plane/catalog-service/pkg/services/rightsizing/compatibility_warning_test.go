// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
)

func TestSizingRecommendationsExposeUnverifiedRuntimeCompatibility(t *testing.T) {
	capacity, one := int64(24_146_608_128), 1
	for _, tc := range []struct {
		name, precision, backend string
		minimum                  float64
	}{
		{"fractional BF16", "bf16", "vllm", 15},
		{"whole GPU NVFP4", "nvfp4", "vllm", 24},
		{"SGLang NVFP4", "nvfp4", "sglang", 24},
		{"TensorRT NVFP4", "nvfp4", "tensorrt_llm", 24},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-A10G", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, []ngcmetadata.Profile{{
				ID: "published-profile", Precision: tc.precision, InferenceBackend: tc.backend,
				MinimumVRAMPerDeviceGB: &tc.minimum, TensorParallelSize: &one, PipelineParallelSize: &one,
			}}, 90)
			if result.Recommendation == nil || result.ProfileId != "published-profile" || result.Recommendation.Estimated {
				t.Fatalf("sizing/source information changed: %+v", result)
			}
			warning := result.Recommendation.Warning
			if !strings.Contains(warning, "compatibility") || !strings.Contains(warning, "unverified") {
				t.Fatalf("runtime compatibility presented as verified: %q", warning)
			}
			if tc.precision == "nvfp4" {
				for _, prerequisite := range []string{"Blackwell", "NIM_ALLOW_NVFP4_EMULATION=1", "vLLM", "SGLang", "TensorRT-LLM"} {
					if !strings.Contains(warning, prerequisite) {
						t.Errorf("NVFP4 runtime prerequisite %q is missing: %q", prerequisite, warning)
					}
				}
			}
		})
	}
}
