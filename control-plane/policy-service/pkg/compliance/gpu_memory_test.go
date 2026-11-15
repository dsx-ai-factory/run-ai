/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package compliance

import (
	"context"
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
)

func TestEffectiveMemoryDefaultsPreserveLegacyValidation(t *testing.T) {
	for _, value := range []float64{.25, 4, 4.5} {
		for _, kind := range []string{"Workspace", "Training", "Inference", "Distributed", "DistributedInference"} {
			t.Run(fmt.Sprintf("%s/%v", kind, value), func(t *testing.T) {
				defaults := map[string]any{"compute": map[string]any{"gpuPortionRequest": value, "gpuRequestType": "portion"}}
				switch kind {
				case "Distributed":
					defaults = map[string]any{"worker": defaults}
				case "DistributedInference":
					defaults = map[string]any{"worker": defaults, "leader": defaults}
				}
				canonical, err := gpucompat.NormalizeLegacyGPUFields(defaults)
				if err != nil {
					t.Fatal(err)
				}
				legacyErr := ValidateEffectivePolicy(context.Background(), kind, nil, defaults)
				canonicalErr := ValidateEffectivePolicy(context.Background(), kind, nil, canonical)
				if (legacyErr != nil) != (canonicalErr != nil) {
					t.Fatalf("normalization changed validation: legacy=%v canonical=%v", legacyErr, canonicalErr)
				}
				if value == 4.5 && canonicalErr == nil {
					t.Fatal("fractional whole GPU default accepted")
				}
				if value != 4.5 && canonicalErr != nil {
					t.Fatal(canonicalErr)
				}
			})
		}
	}
}
