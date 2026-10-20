/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
)

func TestGPUMemoryAssetSelectorCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		reject     bool
	}{
		{"old", `{"gpuDevicesRequest":1,"gpuRequestType":"portion","gpuPortionRequest":0.25}`, false},
		{"new", `{"gpuDevicesRequest":1,"gpuRequestType":"portion","gpuMemoryPortionRequest":0.25}`, false},
		{"equal", `{"gpuDevicesRequest":1,"gpuRequestType":"portion","gpuPortionRequest":0.25,"gpuMemoryPortionRequest":0.25}`, false},
		{"conflict", `{"gpuDevicesRequest":1,"gpuRequestType":"portion","gpuPortionRequest":0.25,"gpuMemoryPortionRequest":0.5}`, true},
		{"portion missing", `{"gpuDevicesRequest":1,"gpuRequestType":"portion"}`, true},
		{"wrong selector", `{"gpuDevicesRequest":1,"gpuRequestType":"memory","gpuMemoryPortionRequest":0.25}`, true},
		{"selector omitted", `{"gpuDevicesRequest":1,"gpuMemoryPortionRequest":0.25}`, true},
		{"invalid limit", `{"gpuDevicesRequest":1,"gpuRequestType":"portion","gpuMemoryPortionRequest":0.5,"gpuMemoryPortionLimit":0.25}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var spec api.ComputeAssetSpec
			if err := json.Unmarshal([]byte(tc.body), &spec); err != nil {
				t.Fatal(err)
			}
			err := validateSingleComputeGpuRequest(spec, true)
			if (err != nil) != tc.reject {
				t.Fatalf("rejection=%v, error=%v", tc.reject, err)
			}
		})
	}
}
