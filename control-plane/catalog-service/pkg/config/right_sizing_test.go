// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package config_test

import (
	catalogconfig "github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/services/rightsizing"
	sharedconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"testing"
)

func TestSharingCutoffConfigurationChangesAllocation(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		whole       bool
	}{
		{"default 90 percent", "", false}, {"override to 80 percent", "80", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("WHOLE_GPU_CUTOFF_PERCENT", tc.value)
			sharedconfig.FlushAll()
			t.Cleanup(sharedconfig.FlushAll)
			if err := sharedconfig.Register[*catalogconfig.Services](); err != nil {
				t.Fatal(err)
			}
			cutoff := sharedconfig.Get[*catalogconfig.Services]().WholeGPUCutoffPercent
			result, err := rightsizing.SingleGPUAllocation(85, 100, cutoff)
			if err != nil || result.WholeGPU != tc.whole {
				t.Fatalf("cutoff=%d allocation=%+v error=%v", cutoff, result, err)
			}
		})
	}
}

func TestSharingCutoffConfigurationRejectsInvalidPercentages(t *testing.T) {
	for _, value := range []string{"0", "-1", "101"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("WHOLE_GPU_CUTOFF_PERCENT", value)
			sharedconfig.FlushAll()
			t.Cleanup(sharedconfig.FlushAll)
			if err := sharedconfig.Register[*catalogconfig.Services](); err == nil {
				t.Fatalf("accepted invalid percentage %q", value)
			}
		})
	}
}
