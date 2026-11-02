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
	Request    resource.Quantity
	Limit      resource.Quantity
	HasRequest bool
}

func HasPerContainerGPUMemoryAnnotations(annotations map[string]string) bool {
	for key := range annotations {
		if isPerContainerGPUMemoryAnnotation(key) {
			return true
		}
	}
	return false
}

// ReadGPUMemoryAnnotations preserves legacy accounting even when KAI has added
// per-container annotations to an existing pod. New workloads carry only the
// per-container request, whose quantity retains its original units.
func ReadGPUMemoryAnnotations(annotations map[string]string) (GPUMemoryResources, error) {
	containers := map[string]GPUMemoryResources{}
	for key, value := range annotations {
		if !isPerContainerGPUMemoryAnnotation(key) {
			continue
		}
		suffix := constants.AnnotationGpuMemoryRequestSuffix
		isLimit := strings.HasSuffix(key, constants.AnnotationGpuMemoryLimitSuffix)
		if isLimit {
			suffix = constants.AnnotationGpuMemoryLimitSuffix
		}
		name := strings.TrimSuffix(strings.TrimPrefix(key, constants.AnnotationGpuMemoryContainerPrefix), suffix)
		if name == "" {
			return GPUMemoryResources{}, fmt.Errorf("GPU memory annotation %s has no container name", key)
		}
		quantity, err := resource.ParseQuantity(value)
		if err != nil {
			return GPUMemoryResources{}, fmt.Errorf("parse GPU memory annotation %s: %w", key, err)
		}
		if quantity.Sign() <= 0 {
			return GPUMemoryResources{}, fmt.Errorf("GPU memory annotation %s must be positive", key)
		}
		container := containers[name]
		if isLimit {
			container.Limit = quantity
		} else {
			container.Request = quantity
			container.HasRequest = true
		}
		containers[name] = container
	}

	var total GPUMemoryResources
	legacyMemory := annotations[constants.AnnotationGpuMemory]
	hasLegacyRequest := legacyMemory != "" || annotations[constants.AnnotationGpuFraction] != ""
	if legacyMemory != "" {
		quantity, err := resource.ParseQuantity(legacyMemory)
		if err != nil {
			return GPUMemoryResources{}, fmt.Errorf("parse GPU memory annotation %s: %w", constants.AnnotationGpuMemory, err)
		}
		quantity.Set(int64(quantity.AsApproximateFloat64() * 1000000))
		total.Request = quantity
		total.HasRequest = true
	}
	for _, container := range containers {
		if hasLegacyRequest {
			// A per-container limit must not replace an explicit legacy request.
			total.Limit.Add(container.Limit)
			continue
		}
		if !container.HasRequest {
			container.Request = container.Limit.DeepCopy()
		}
		if container.Limit.IsZero() {
			container.Limit = container.Request.DeepCopy()
		}
		total.Request.Add(container.Request)
		total.Limit.Add(container.Limit)
		total.HasRequest = true
	}
	if hasLegacyRequest && total.Limit.IsZero() {
		total.Limit = total.Request.DeepCopy()
	}
	return total, nil
}

func hasPerContainerGPUMemoryLimit(annotations map[string]string) bool {
	for key := range annotations {
		if isPerContainerGPUMemoryAnnotation(key) && strings.HasSuffix(key, constants.AnnotationGpuMemoryLimitSuffix) {
			return true
		}
	}
	return false
}

func isPerContainerGPUMemoryAnnotation(key string) bool {
	return strings.HasPrefix(key, constants.AnnotationGpuMemoryContainerPrefix) &&
		(strings.HasSuffix(key, constants.AnnotationGpuMemoryRequestSuffix) || strings.HasSuffix(key, constants.AnnotationGpuMemoryLimitSuffix))
}
