/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/apimachinery/pkg/api/resource"
	"strconv"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

// IsFractionalGpu - preferring not to rely on float conversions which are not accurate
// the number is checked as a string. value of 12.0 or 12 indicate non-fractional
func IsFractionalGpu(ctx context.Context, gpu string) bool {
	parts := strings.Split(gpu, ".")
	if len(parts) == 1 {
		return false
	}
	fract, err := strconv.Atoi(parts[1])
	if err != nil {
		log := logger.FromContext(ctx)
		log.Error(err, "invalid value", "gpu", gpu)
		return false
	}
	return fract != 0
}

// IsWholeWithFractionalGpu checks if the given GPU value is a whole number with a fractional part.
// It first checks if the GPU value is less than 1, if so, it returns false.
// If the GPU value is greater than or equal to 1, it converts the GPU value to a string and checks if it has a fractional part.
// It uses the IsFractionalGpu function to check for the fractional part.
// The function returns true if the GPU value is a whole number with a fractional part, otherwise it returns false.
//
// Parameters:
// ctx: The context in which the function is called. It is used for logging purposes in the IsFractionalGpu function.
// gpu: The GPU value to check. It is a float64 value representing the number of GPUs.
//
// Returns:
// The function returns a boolean value. It returns true if the GPU value is a whole number with a fractional part, otherwise it returns false.
func IsWholeWithFractionalGpu(ctx context.Context, gpu float64) bool {
	if gpu < 1 {
		return false
	}
	return IsFractionalGpu(ctx, strconv.FormatFloat(gpu, 'f', -1, 64))
}

func IsFractionalGpuOrMemroy(ctx context.Context, gpu float64, gpuMemory resource.Quantity) bool {
	log := logger.FromContext(ctx)
	log.V(4).Info("partial check", "gpu", gpu, "gpuMemory", gpuMemory)
	if gpu < 1 && gpu > 0 {
		log.V(4).Info("fractional gpu found")
		return true
	}

	if gpuMemory.ScaledValue(resource.Mega) > 0 {
		log.V(4).Info("fractional gpu memory found")
		return true
	}

	log.V(4).Info("not part of gpu")
	return false
}

func AddPartialGpuAnnotations(annotations map[string]string, gpu *field_types.NumberField, memoryQuantity resource.Quantity, gpuDevices int) {
	// when no GPU device is set, memory fraction uses 1 GPU, so it will be clear the workload uses GPUs.
	if gpuDevices == 0 {
		gpuDevices = 1
	}
	annotations[constants.AnnotationGpuFractionDevices] = strconv.Itoa(gpuDevices)

	memoryMb := int(memoryQuantity.ScaledValue(resource.Mega))

	if memoryMb > 0 {
		annotations[constants.AnnotationGpuMemory] = strconv.Itoa(memoryMb)
	}
	if gpu.Deref() > 0 && gpu.Deref() < 1 {
		annotations[constants.AnnotationGpuFraction] = gpu.DerefStr()
	}
}
