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
		estimated                bool
	}{
		{"fractional BF16", "bf16", "vllm", 15, false},
		{"whole GPU NVFP4", "nvfp4", "vllm", 24, false},
		{"estimated NVFP4", "nvfp4", "sglang", 82, true},
		{"TensorRT NVFP4", "nvfp4", "tensorrt_llm", 24, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-A10G", MemoryBytes: &capacity}, []ngcmetadata.Profile{{
				ID: "published-profile", Precision: tc.precision, InferenceBackend: tc.backend,
				MinimumVRAMPerDeviceGB: &tc.minimum, TensorParallelSize: &one, PipelineParallelSize: &one,
			}})
			if result.Recommendation == nil || result.ProfileId != "published-profile" || result.Recommendation.Estimated != tc.estimated {
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
			if tc.estimated && !strings.Contains(warning, "not a validated tensor-parallel configuration") {
				t.Fatalf("naive-count caveat lost: %q", warning)
			}
		})
	}
}
