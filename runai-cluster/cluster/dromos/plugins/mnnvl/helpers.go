/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"context"
	"fmt"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	kartautils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/karta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func extractMnnvlConfig(runner *dromosv1alpha1.WorkloadRunner) *dromosv1alpha1.MnnvlConfig {
	if runner.Spec.Instructions == nil {
		return nil
	}
	if runner.Spec.Instructions.GpuOptimizations == nil {
		return nil
	}
	return runner.Spec.Instructions.GpuOptimizations.Mnnvl
}

func mnnvlInstructionOf(runner *dromosv1alpha1.WorkloadRunner) any {
	config := extractMnnvlConfig(runner)
	if config == nil || config.Mode == dromosv1alpha1.MnnvlModeNone {
		return nil
	}
	return config
}

func mnnvlInstructionPath() *field.Path {
	return field.NewPath("spec", "instructions", "gpuOptimizations", "mnnvl")
}

func (p *Plugin) calculateNumNodes(
	ctx context.Context,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
) (int, error) {
	factory := kartaresource.NewComponentFactoryFromObject(karta, resource)
	components, err := kartautils.GetAllComponents(factory)
	if err != nil {
		return 0, err
	}

	totalNodes := 0
	for _, component := range components {
		instances, err := component.GetExtractedInstances(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to get extracted instances for component %s: %w", component.Name(), err)
		}

		for _, instance := range instances {
			if !kartautils.HasGPU(instance) {
				continue
			}
			replicas := kartautils.CalcInstanceReplicas(instance)
			totalNodes += int(replicas)
		}
	}

	return totalNodes, nil
}

func (p *Plugin) applyComputeDomain(
	ctx context.Context,
	k8sClient client.Client,
	updatedCd *nvapi.ComputeDomain,
) error {
	log.FromContext(ctx).Info("applying compute domain", "compute domain", updatedCd.Name)
	updatedCd.APIVersion = nvapi.SchemeGroupVersion.String()
	updatedCd.Kind = nvapi.ComputeDomainKind
	return k8sClient.Patch(ctx, updatedCd, client.Apply, client.ForceOwnership, client.FieldOwner(PluginName))
}
