/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpumode"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

func (fc *FieldsComposer) ComposeGpuComputeMode(ctx context.Context, src *string, sizing *GpuSizingInfo,
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
			available, reason := false, "GPU compute mode availability has not been reported"
			if fc.Options.GpuComputeModeStatus != nil {
				var err error
				available, reason, err = fc.Options.GpuComputeModeStatus(ctx, *src)
				if err != nil {
					return workloadsApi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to check compute.gpuComputeMode availability: %v", err))
				}
			}
			if !available {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("compute.gpuComputeMode %q is unavailable on the selected cluster: %s", *src, reason))
			}
		}
		if len(containerName) > maxContainerNameLenForAnnotationKey(gpumode.AnnotationFormat) {
			return workloadsApi.NewError(http.StatusBadRequest, "container name is too long for the GPU compute mode annotation")
		}
		return setter(src, target, fmt.Sprintf(gpumode.AnnotationFormat, containerName))
	}
}
