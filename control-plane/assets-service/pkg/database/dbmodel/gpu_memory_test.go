/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
)

func TestGPUMemoryAssetRoundTrip(t *testing.T) {
	request, limit := .25, .5
	mode := api.GPUCOMPUTEMODE_SM_SHARING
	input := api.ComputeAssetSpec{GpuMemoryPortionRequest: &request, GpuMemoryPortionLimit: &limit, GpuComputeMode: &mode}
	stored, err := ComputeAssetForDB(input)
	if err != nil {
		t.Fatal(err)
	}
	if stored.GpuMemoryPortionRequest != nil || stored.GpuMemoryPortionLimit != nil {
		t.Fatal("persisted new alias")
	}
	response := LegacyCompatibleComputeAsset(stored)
	if response.GpuComputeMode == nil || *response.GpuComputeMode != mode {
		t.Fatal("response lost compute mode")
	}
	if *response.GpuPortionRequest != request || *response.GpuMemoryPortionRequest != request || *response.GpuMemoryPortionLimit != limit {
		t.Fatal("response lost aliases")
	}
	*response.GpuPortionRequest = .9
	if *stored.GpuPortionRequest != .25 || *response.GpuMemoryPortionRequest != .25 {
		t.Fatal("shared alias pointers")
	}
}

func TestLegacyWholeGPUAssetResponse(t *testing.T) {
	value := 4.0
	stored := api.ComputeAssetSpec{GpuPortionRequest: &value}
	response := LegacyCompatibleComputeAsset(stored)
	if response.GpuMemoryPortionRequest != nil || response.GpuPortionRequest == nil || *response.GpuPortionRequest != value || response.GpuDevicesRequest != nil {
		t.Fatal("legacy record changed on read")
	}
}
