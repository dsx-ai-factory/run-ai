/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plugins

import (
	"context"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	EnvVarName                        = "WORKLOAD_RUNNER_NAME"
	resourceMutateAddEnvVarPluginName = "resource-mutate-add-env-var-plugin"
)

// ResourceMutateAddEnvVarPlugin is a plugin that adds an environment variable to pod containers
// using the Karta utils.
// This plugin is used for testing the resource mutation functionality.
type ResourceMutateAddEnvVarPlugin struct{}

func NewResourceMutateAddEnvVarPlugin() *ResourceMutateAddEnvVarPlugin {
	return &ResourceMutateAddEnvVarPlugin{}
}

func (p *ResourceMutateAddEnvVarPlugin) Name() string {
	return resourceMutateAddEnvVarPluginName
}

// MutateResource adds WORKLOAD_RUNNER_NAME environment variable to pod containers.
func (p *ResourceMutateAddEnvVarPlugin) MutateResource(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	state *reconcile.State,
) error {
	factory := kartaresource.NewComponentFactoryFromObject(karta, resource)

	components, err := p.getComponents(factory)
	if err != nil {
		return err
	}

	for _, component := range components {
		instancesMap, err := component.GetExtractedInstances(ctx)
		if err != nil {
			return fmt.Errorf("failed to get extracted instances: %w", err)
		}

		if p.mutateComponentInstancesContainers(instancesMap, runner.Name) {
			if err := p.updateComponentPodSpec(ctx, component, instancesMap); err != nil {
				return fmt.Errorf("failed to update component pod spec: %w", err)
			}
		}
	}

	updatedResource, err := factory.GetResource()
	if err != nil {
		return fmt.Errorf("failed to get updated resource: %w", err)
	}

	// Update the original attachedResource object in-place
	updatedUnstructured := updatedResource.(*unstructured.Unstructured)
	resource.Object = updatedUnstructured.Object

	return nil
}

func (p *ResourceMutateAddEnvVarPlugin) getComponents(factory *kartaresource.ComponentFactory) ([]*kartaresource.Component, error) {
	components, err := factory.GetChildComponents()
	if err != nil {
		return nil, fmt.Errorf("failed to get child components: %w", err)
	}

	rootComponent, err := factory.GetRootComponent()
	if err != nil {
		return nil, fmt.Errorf("failed to get root component: %w", err)
	}
	return append(components, rootComponent), nil
}

// mutateComponentInstancesContainers mutates the containers by adding the WORKLOAD_RUNNER_NAME environment variable.
func (p *ResourceMutateAddEnvVarPlugin) mutateComponentInstancesContainers(instancesMap map[string]kartaresource.ExtractedInstance, runnerName string) bool {
	mutated := false
	for _, instance := range instancesMap {
		var containers []corev1.Container
		if instance.PodTemplateSpec != nil {
			containers = instance.PodTemplateSpec.Spec.Containers
		} else if instance.PodSpec != nil {
			containers = instance.PodSpec.Containers
		} else if instance.FragmentedPodSpec != nil {
			containers = instance.FragmentedPodSpec.Containers
		}

		if containers != nil {
			if modifyContainersEnv(containers, runnerName) {
				mutated = true
			}
		}
	}
	return mutated
}

func (p *ResourceMutateAddEnvVarPlugin) updateComponentPodSpec(ctx context.Context, component *kartaresource.Component, instancesMap map[string]kartaresource.ExtractedInstance) error {
	if len(instancesMap) == 0 {
		return nil
	}

	// Use the first instance to determine the type for the update call
	sample := lo.Values(instancesMap)[0]

	if sample.FragmentedPodSpec != nil {
		return component.UpdateFragmentedPodSpec(ctx, lo.MapValues(instancesMap, func(instance kartaresource.ExtractedInstance, _ string) kartaresource.FragmentedPodSpec {
			return *instance.FragmentedPodSpec
		}))
	} else if sample.PodSpec != nil {
		return component.UpdatePodSpec(ctx, lo.MapValues(instancesMap, func(instance kartaresource.ExtractedInstance, _ string) corev1.PodSpec {
			return *instance.PodSpec
		}))
	} else if sample.PodTemplateSpec != nil {
		return component.UpdatePodTemplateSpec(ctx, lo.MapValues(instancesMap, func(instance kartaresource.ExtractedInstance, _ string) corev1.PodTemplateSpec {
			return *instance.PodTemplateSpec
		}))
	} else {
		return fmt.Errorf("unsupported instance type: %T", sample)
	}
}

func modifyContainersEnv(containers []corev1.Container, runnerName string) (mutated bool) {
	mutated = false
	for i := range containers {
		container := &containers[i]
		if lo.ContainsBy(container.Env, func(env corev1.EnvVar) bool {
			return env.Name == EnvVarName
		}) {
			continue
		}
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  EnvVarName,
			Value: runnerName,
		})
		mutated = true
	}
	return mutated
}

var (
	_ framework.ResourceMutatePlugin = &ResourceMutateAddEnvVarPlugin{}
)
