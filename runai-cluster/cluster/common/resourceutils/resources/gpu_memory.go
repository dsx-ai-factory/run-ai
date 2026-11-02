/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"k8s.io/apimachinery/pkg/api/resource"
)

type GPUMemoryResources struct {
	Request resource.Quantity
	Limit   resource.Quantity
}

func HasNativeGPUMemoryAnnotations(annotations map[string]string) bool {
	for key := range annotations {
		if isNativeGPUMemoryAnnotation(key) {
			return true
		}
	}
	return false
}

// ReadNativeGPUMemoryAnnotations preserves quantity units and defaults a limit-only container's request
// from its limit, matching KAI. Missing limits are reported as requests, as in
// the legacy Run:ai resource reports. Admission remains responsible for checking
// container references and whether the requested combination is supported.
func ReadNativeGPUMemoryAnnotations(annotations map[string]string) (GPUMemoryResources, bool, error) {
	containers := map[string]GPUMemoryResources{}
	for key, value := range annotations {
		if !isNativeGPUMemoryAnnotation(key) {
			continue
		}
		suffix := constants.AnnotationGpuMemoryRequestSuffix
		isLimit := strings.HasSuffix(key, constants.AnnotationGpuMemoryLimitSuffix)
		if isLimit {
			suffix = constants.AnnotationGpuMemoryLimitSuffix
		}
		name := strings.TrimSuffix(strings.TrimPrefix(key, constants.AnnotationGpuMemoryContainerPrefix), suffix)
		if name == "" {
			return GPUMemoryResources{}, true, fmt.Errorf("GPU memory annotation %s has no container name", key)
		}
		quantity, err := resource.ParseQuantity(value)
		if err != nil {
			return GPUMemoryResources{}, true, fmt.Errorf("parse GPU memory annotation %s: %w", key, err)
		}
		if quantity.Sign() <= 0 {
			return GPUMemoryResources{}, true, fmt.Errorf("GPU memory annotation %s must be positive", key)
		}
		container := containers[name]
		if isLimit {
			container.Limit = quantity
		} else {
			container.Request = quantity
		}
		containers[name] = container
	}

	var total GPUMemoryResources
	for _, container := range containers {
		if container.Request.IsZero() {
			container.Request = container.Limit.DeepCopy()
		}
		if container.Limit.IsZero() {
			container.Limit = container.Request.DeepCopy()
		}
		total.Request.Add(container.Request)
		total.Limit.Add(container.Limit)
	}
	return total, len(containers) > 0, nil
}

func isNativeGPUMemoryAnnotation(key string) bool {
	return strings.HasPrefix(key, constants.AnnotationGpuMemoryContainerPrefix) &&
		(strings.HasSuffix(key, constants.AnnotationGpuMemoryRequestSuffix) || strings.HasSuffix(key, constants.AnnotationGpuMemoryLimitSuffix))
}
