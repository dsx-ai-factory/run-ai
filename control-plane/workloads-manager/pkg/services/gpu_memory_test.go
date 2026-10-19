/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
)

func TestGPUMemoryNimPolicyAndUpdate(t *testing.T) {
	for _, oldDefault := range []bool{false, true} {
		inputKey, defaultKey := "gpuMemoryPortionRequest", "gpuPortionRequest"
		if !oldDefault {
			inputKey, defaultKey = defaultKey, inputKey
		}
		t.Run(defaultKey, func(t *testing.T) {
			for _, locked := range []bool{false, true} {
				var spec api.NimServiceSpec
				if err := json.Unmarshal([]byte(`{"compute":{"`+inputKey+`":0.5}}`), &spec); err != nil {
					t.Fatal(err)
				}
				defaults := map[string]any{"compute": map[string]any{defaultKey: json.Number("0.25")}}
				rules := map[string]any{"compute": map[string]any{defaultKey: map[string]any{"canEdit": !locked}}}
				mutated, err := MutateWorkload(&spec, nil, defaults)
				if err != nil {
					t.Fatal(err)
				}
				_, validationErr := validateMutatedSpecVsPolicy(rules, defaults, mutated)
				if (validationErr != nil) != locked {
					t.Fatalf("lock=%v, error=%v", locked, validationErr)
				}
			}
		})
	}
	var full, update api.NimServiceSpec
	if err := json.Unmarshal([]byte(`{"compute":{"gpuPortionRequest":0.25,"gpuPortionLimit":0.5},"image":"image"}`), &full); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"compute":{"gpuMemoryPortionRequest":0.75}}`), &update); err != nil {
		t.Fatal(err)
	}
	merged, err := (&NimServicesImpl{}).mergeUpdateSpecIntoFullSpec(context.Background(), update, full, nil)
	if err != nil {
		t.Fatal(err)
	}
	if merged.Compute.GpuMemoryPortionRequest == nil || *merged.Compute.GpuMemoryPortionRequest != .75 || merged.Compute.GpuPortionLimit != nil || merged.Compute.GpuMemoryPortionLimit != nil {
		t.Fatal("compute section replacement changed")
	}
	stored := nimSpecForDB(merged)
	response := legacyCompatibleNimSpec(stored)
	values, jsonErr := gojsonq.LoadJson(response)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}
	fields := values["compute"].(map[string]any)
	if fields["gpuPortionRequest"] != fields["gpuMemoryPortionRequest"] || stored.Compute.GpuMemoryPortionRequest != nil {
		t.Fatal("round trip lost DB/response contract")
	}
}

func TestNimLegacyWholeGPURoundTrip(t *testing.T) {
	for _, source := range []string{"request", "policy"} {
		t.Run(source, func(t *testing.T) {
			var input api.NimServiceSpec
			var defaults map[string]any
			if source == "request" {
				if err := json.Unmarshal([]byte(`{"compute":{"gpuPortionRequest":4}}`), &input); err != nil {
					t.Fatal(err)
				}
			} else {
				defaults = map[string]any{"compute": map[string]any{"gpuPortionRequest": json.Number("4")}}
			}
			mutated, err := MutateWorkload(&input, nil, defaults)
			if err != nil {
				t.Fatal(err)
			}
			payload, jsonErr := json.Marshal(mutated)
			if jsonErr != nil {
				t.Fatal(jsonErr)
			}
			var full api.NimServiceSpec
			if jsonErr := json.Unmarshal(payload, &full); jsonErr != nil {
				t.Fatal(jsonErr)
			}
			response := legacyCompatibleNimSpec(nimSpecForDB(full))
			if response.Compute.GpuMemoryPortionRequest != nil || response.Compute.GpuPortionRequest == nil || *response.Compute.GpuPortionRequest != 4 || response.Compute.GpuDevicesRequest != nil {
				t.Fatal("whole GPU response changed")
			}
			if _, err := MutateWorkload(&response, nil, defaults); err != nil {
				t.Fatal("read/resubmit rejected", err)
			}
			updated, updateErr := (&NimServicesImpl{}).mergeUpdateSpecIntoFullSpec(context.Background(), api.NimServiceSpec{}, full, nil)
			if updateErr != nil || updated.Compute.GpuMemoryPortionRequest == nil || *updated.Compute.GpuMemoryPortionRequest != 4 {
				t.Fatalf("unrelated update lost legacy request: %v", updateErr)
			}
		})
	}
	for _, body := range []string{`{"compute":{"gpuMemoryPortionRequest":4}}`, `{"compute":{"gpuPortionRequest":4,"gpuMemoryPortionRequest":4}}`} {
		var input api.NimServiceSpec
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatal(err)
		}
		if _, err := MutateWorkload(&input, nil, nil); err == nil || err.Code != 400 {
			t.Fatalf("invalid new request accepted: %v", err)
		}
		if _, err := (&NimServicesImpl{}).mergeUpdateSpecIntoFullSpec(context.Background(), input, api.NimServiceSpec{}, nil); err == nil {
			t.Fatal("invalid new update accepted")
		}
	}
}
