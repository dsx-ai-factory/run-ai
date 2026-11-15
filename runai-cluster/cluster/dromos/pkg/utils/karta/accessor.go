/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package karta

import (
	"context"
	"fmt"
	"strings"

	kartaresource "github.com/run-ai/karta/pkg/resource"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const nvidiaGPUResourceName = "nvidia.com/gpu"

func GetAllComponents(factory *kartaresource.ComponentFactory) ([]*kartaresource.Component, error) {
	children, err := factory.GetChildComponents()
	if err != nil {
		return nil, fmt.Errorf("failed to get child components: %w", err)
	}

	root, err := factory.GetRootComponent()
	if err != nil {
		return nil, fmt.Errorf("failed to get root component: %w", err)
	}

	return append(children, root), nil
}

func UpdateComponent(
	ctx context.Context,
	component Component,
	instances map[string]kartaresource.ExtractedInstance,
) error {
	if len(instances) == 0 {
		return nil
	}

	sample := getFirstInstance(instances)

	if sample.FragmentedPodSpec != nil {
		return component.UpdateFragmentedPodSpec(ctx, extractFragmentedSpecs(instances))
	} else if sample.PodSpec != nil {
		return component.UpdatePodSpec(ctx, extractPodSpecs(instances))
	} else if sample.PodTemplateSpec != nil {
		return component.UpdatePodTemplateSpec(ctx, extractPodTemplates(instances))
	}
	return nil
}

// CanStampPodLabels reports whether the instance can carry a pod label. A bare
// PodSpec cannot - Karta drops its metadata on write-back.
func CanStampPodLabels(instance kartaresource.ExtractedInstance) bool {
	return instance.PodTemplateSpec != nil || instance.FragmentedPodSpec != nil
}

func HasGPU(instance kartaresource.ExtractedInstance) bool {
	var containers []corev1.Container

	switch {
	case instance.PodTemplateSpec != nil:
		containers = instance.PodTemplateSpec.Spec.Containers
	case instance.PodSpec != nil:
		containers = instance.PodSpec.Containers
	case instance.FragmentedPodSpec != nil:
		frag := instance.FragmentedPodSpec
		if HasGPUInFragmentedResources(frag.Resources) ||
			(frag.Container != nil && containerHasNvidiaGPUResources(frag.Container.Resources)) {
			return true
		}
		containers = frag.Containers
	}

	for _, container := range containers {
		if containerHasNvidiaGPUResources(container.Resources) {
			return true
		}
	}

	return false
}

// HasGPUInFragmentedResources also recognizes the GPU shorthand used by Dynamo
// service resources. The shorthand is not a Kubernetes container resource name.
func HasGPUInFragmentedResources(resources *corev1.ResourceRequirements) bool {
	if resources == nil {
		return false
	}
	if quantity, ok := resources.Requests["gpu"]; ok && quantity.Sign() > 0 {
		return true
	}
	if quantity, ok := resources.Limits["gpu"]; ok && quantity.Sign() > 0 {
		return true
	}
	return containerHasNvidiaGPUResources(*resources)
}

func containerHasNvidiaGPUResources(resources corev1.ResourceRequirements) bool {
	for resourceName := range resources.Requests {
		if strings.Contains(string(resourceName), nvidiaGPUResourceName) {
			return true
		}
	}
	for resourceName := range resources.Limits {
		if strings.Contains(string(resourceName), nvidiaGPUResourceName) {
			return true
		}
	}
	return false
}

// CalcInstanceReplicas preserves replicas when set, including zero. Otherwise it
// uses minReplicas, or 1 when neither count is set. Missing scale is treated as unset.
func CalcInstanceReplicas(instance kartaresource.ExtractedInstance) int32 {
	scale := ptr.Deref(instance.Scale, kartaresource.Scale{})
	return ptr.Deref(scale.Replicas, ptr.Deref(scale.MinReplicas, 1))
}

func getFirstInstance(instances map[string]kartaresource.ExtractedInstance) kartaresource.ExtractedInstance {
	for _, instance := range instances {
		return instance
	}
	return kartaresource.ExtractedInstance{}
}

func extractFragmentedSpecs(instances map[string]kartaresource.ExtractedInstance) map[string]kartaresource.FragmentedPodSpec {
	return lo.MapValues(instances, func(instance kartaresource.ExtractedInstance, _ string) kartaresource.FragmentedPodSpec {
		return *instance.FragmentedPodSpec
	})
}

func extractPodSpecs(instances map[string]kartaresource.ExtractedInstance) map[string]corev1.PodSpec {
	return lo.MapValues(instances, func(instance kartaresource.ExtractedInstance, _ string) corev1.PodSpec {
		return *instance.PodSpec
	})
}

func extractPodTemplates(instances map[string]kartaresource.ExtractedInstance) map[string]corev1.PodTemplateSpec {
	return lo.MapValues(instances, func(instance kartaresource.ExtractedInstance, _ string) corev1.PodTemplateSpec {
		return *instance.PodTemplateSpec
	})
}
