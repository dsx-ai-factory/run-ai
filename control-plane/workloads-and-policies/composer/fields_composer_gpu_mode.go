/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpumode"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
)

func (fc *FieldsComposer) ComposeGpuComputeMode(src *string, sizing *GpuSizingInfo,
	containerName string, target any, setter FnStringSetter,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}
		if *src != gpumode.TimeSlicing && *src != gpumode.SMSharing {
			return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuComputeMode must be time-slicing or sm-sharing")
		}
		hasGPU := sizing != nil && ((sizing.NumDevices != nil && *sizing.NumDevices > 0) ||
			sizing.MemoryRequest != nil || (sizing.PortionRequest != nil && *sizing.PortionRequest > 0))
		if !fc.Options.IgnoreMissingValues {
			if !hasGPU {
				return nil
			}
			if !fc.Options.GpuComputeModeSupported {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("compute.gpuComputeMode requires cluster version %s or later", checker.MinVersionForSupporting(checker.MinVersionGpuComputeMode)))
			}
		}
		if len(containerName) > maxContainerNameLenForAnnotationKey(gpumode.AnnotationFormat) {
			return workloadsApi.NewError(http.StatusBadRequest, "container name is too long for the GPU compute mode annotation")
		}
		return setter(src, target, fmt.Sprintf(gpumode.AnnotationFormat, containerName))
	}
}
