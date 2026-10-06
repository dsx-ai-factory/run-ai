/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpu_validator

import (
	"context"
	"strconv"
	"strings"

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
