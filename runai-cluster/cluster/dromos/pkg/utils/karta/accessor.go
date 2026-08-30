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
		containers = instance.FragmentedPodSpec.Containers
	}

	for _, container := range containers {
		if containerHasNvidiaGPUResources(container.Resources) {
			return true
		}
	}

	return false
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

func CalcInstanceReplicas(instance kartaresource.ExtractedInstance) int32 {
	var replicas int32
	if instance.Scale != nil {
		replicas = ptr.Deref(instance.Scale.Replicas, 0)
		if replicas == 0 {
			replicas = ptr.Deref(instance.Scale.MinReplicas, 0)
		}
	}
	if replicas == 0 {
		replicas = 1
	}
	return replicas
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
