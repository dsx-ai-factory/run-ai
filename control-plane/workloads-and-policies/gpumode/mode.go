/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpumode

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	leaderSection    = "leader"
	masterSection    = "master"
	workerSection    = "worker"
	TimeSlicing      = "time-slicing"
	SMSharing        = "sm-sharing"
	AnnotationFormat = "nvidia.com/container.%s.gpu-compute.mode"
)

// VisitCompute visits only workload compute sections, never arbitrary annotation,
// environment-variable or application configuration maps. Policy defaults and rules
// are excluded: those are inputs to resolution, not workload GPU allocations.
func VisitCompute(spec map[string]interface{}, visit func(map[string]interface{})) {
	if compute, ok := spec["compute"].(map[string]interface{}); ok {
		visit(compute)
	}
	for _, key := range []string{"spec", "masterSpec", workerSection, masterSection, leaderSection, "combinedSpec", "combinedMasterSpec"} {
		if child, ok := spec[key].(map[string]interface{}); ok {
			VisitCompute(child, visit)
		}
	}
}

func HasGPU(compute map[string]interface{}) bool {
	for _, key := range []string{"gpuDevicesRequest", "gpuMemoryPortionRequest", "gpuPortionRequest"} {
		value, err := strconv.ParseFloat(fmt.Sprint(compute[key]), 64)
		if err == nil && value > 0 {
			return true
		}
	}
	memory, ok := compute["gpuMemoryRequest"].(string)
	if !ok {
		return false
	}
	quantity, err := resource.ParseQuantity(memory)
	return err == nil && quantity.Sign() > 0
}
