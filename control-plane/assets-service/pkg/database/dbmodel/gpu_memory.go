/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
)

func NormalizeLegacyComputeAsset(source api.ComputeAssetSpec) (api.ComputeAssetSpec, error) {
	var err error
	source.GpuMemoryPortionRequest, err = gpucompat.ResolveLegacyGPUMemory("Request", source.GpuPortionRequest, source.GpuMemoryPortionRequest)
	if err != nil {
		return api.ComputeAssetSpec{}, err
	}
	source.GpuMemoryPortionLimit, err = gpucompat.ResolveLegacyGPUMemory("Limit", source.GpuPortionLimit, source.GpuMemoryPortionLimit)
	if err != nil {
		return api.ComputeAssetSpec{}, err
	}
	source.GpuPortionRequest, source.GpuPortionLimit = nil, nil
	return source, nil
}

// ComputeAssetForDB uses the existing columns; aliases are API-only fields.
func ComputeAssetForDB(source api.ComputeAssetSpec) (api.ComputeAssetSpec, error) {
	source, err := NormalizeLegacyComputeAsset(source)
	if err != nil {
		return api.ComputeAssetSpec{}, err
	}
	source.GpuPortionRequest, source.GpuPortionLimit = source.GpuMemoryPortionRequest, source.GpuMemoryPortionLimit
	source.GpuMemoryPortionRequest, source.GpuMemoryPortionLimit = nil, nil
	return source, nil
}

func LegacyCompatibleComputeAsset(source api.ComputeAssetSpec) api.ComputeAssetSpec {
	if source.GpuMemoryPortionRequest == nil {
		source.GpuMemoryPortionRequest = source.GpuPortionRequest
	}
	if source.GpuMemoryPortionLimit == nil {
		source.GpuMemoryPortionLimit = source.GpuPortionLimit
	}
	source.GpuPortionRequest = copyGPUValue(source.GpuMemoryPortionRequest)
	source.GpuPortionLimit = copyGPUValue(source.GpuMemoryPortionLimit)
	source.GpuMemoryPortionRequest = copyGPUValue(source.GpuMemoryPortionRequest)
	if source.GpuMemoryPortionRequest != nil && *source.GpuMemoryPortionRequest > 1 {
		source.GpuMemoryPortionRequest = nil
	}
	source.GpuMemoryPortionLimit = copyGPUValue(source.GpuMemoryPortionLimit)
	return source
}

func copyGPUValue(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
