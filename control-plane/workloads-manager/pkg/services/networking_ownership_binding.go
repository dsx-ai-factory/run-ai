/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/karta/pkg/jq/execution"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	networkingOwnershipContractV1 = "v1"
	networkingWorkloadIDLabel     = "run.ai/networking-workload-id"
)

func enableNetworkingOwnershipContract(configuration **models.WorkloadConfiguration) {
	if *configuration == nil {
		*configuration = &models.WorkloadConfiguration{}
	}
	(*configuration).NetworkingOwnershipContract = &models.NetworkingOwnershipContract{
		Version: networkingOwnershipContractV1,
	}
}

func usesNetworkingOwnershipContract(configuration *models.WorkloadConfiguration) bool {
	return configuration != nil &&
		configuration.NetworkingOwnershipContract != nil &&
		configuration.NetworkingOwnershipContract.Version == networkingOwnershipContractV1
}

// applyNetworkingOwnershipBinding establishes the part of the networking
// ownership contract that can be determined before the workload reaches the
// cluster. Created Services receive a protected selector and every pod-label
// target described by Karta receives the same control-plane workload ID.
// Referenced Services are deliberately left unchanged and are verified later
// from live cluster state.
func applyNetworkingOwnershipBinding(
	ctx context.Context,
	record *models.WorkloadsQueueRecord,
	workloadType *workloadsApi.WorkloadTypeConfig,
) error {
	if record == nil || !usesNetworkingOwnershipContract(record.Configuration) {
		return nil
	}

	// The ownership label is server-controlled. Remove every caller- or
	// policy-supplied occurrence before deciding whether this workload needs
	// the label. Workloads without a created Service must not be able to claim
	// another workload's identity and join its Service selector.
	stripNetworkingWorkloadIDLabel(map[string]interface{}(record.EffectiveManifest))

	createdServices := bindCreatedServiceSelectors(record.Configuration.Networking, record.ID)
	if createdServices == 0 {
		return nil
	}

	karta, kartaErr := decodeKartaForVersion(workloadType, record.ResourceVersion)
	if kartaErr != nil {
		return fmt.Errorf("get Karta for networking ownership binding: %s", kartaErr.Message)
	}
	if karta == nil {
		return fmt.Errorf("get Karta for networking ownership binding: Karta is nil")
	}

	manifest := &unstructured.Unstructured{Object: map[string]interface{}(record.EffectiveManifest)}
	presenceEvaluator := execution.NewDefaultRunner(manifest)
	factory := kartaresource.NewComponentFactoryFromObject(karta, manifest)
	components, err := networkingKartaComponents(factory)
	if err != nil {
		return err
	}

	stampedTargets := 0
	for _, component := range components {
		absent, err := kartaCompletePodTargetAbsent(ctx, presenceEvaluator, component.Definition())
		if err != nil {
			return fmt.Errorf("evaluate Karta component %q pod target presence: %w", component.Name(), err)
		}
		if absent {
			continue
		}

		instances, err := component.GetExtractedInstances(ctx)
		if err != nil {
			return fmt.Errorf("extract Karta component %q: %w", component.Name(), err)
		}

		componentChanged := false
		for _, instance := range instances {
			target, supported, changed := stampNetworkingWorkloadID(instance, record.ID, component.Definition())
			if target && !supported {
				return fmt.Errorf("Karta component %q cannot carry the pod label required by networking ownership binding", component.Name())
			}
			if target {
				stampedTargets++
			}
			componentChanged = componentChanged || changed
		}

		if componentChanged {
			if err := updateNetworkingKartaComponent(ctx, component, instances); err != nil {
				return fmt.Errorf("update Karta component %q: %w", component.Name(), err)
			}
		}
	}

	if stampedTargets == 0 {
		return fmt.Errorf("Karta exposes no writable pod-label target required by networking ownership binding")
	}

	updated, err := factory.GetResource()
	if err != nil {
		return fmt.Errorf("get workload manifest after networking ownership binding: %w", err)
	}
	updatedManifest, ok := updated.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("get workload manifest after networking ownership binding: unexpected resource type %T", updated)
	}
	record.EffectiveManifest = updatedManifest.Object
	return nil
}

// kartaCompletePodTargetAbsent reports whether every complete pod target
// exposed by a component is absent. Karta's typed extraction converts missing
// and null paths into zero-valued pod structs, so this check must run against
// the raw manifest before extraction to avoid materializing optional targets.
// Fragmented definitions are deliberately excluded because an absent leaf,
// such as a labels map, can belong to a present target that must be stamped.
func kartaCompletePodTargetAbsent(
	ctx context.Context,
	evaluator execution.Evaluator,
	definition kartav1alpha1.ComponentDefinition,
) (bool, error) {
	if definition.SpecDefinition == nil {
		return false, nil
	}

	var path *string
	switch {
	case definition.SpecDefinition.PodTemplateSpecPath != nil:
		path = definition.SpecDefinition.PodTemplateSpecPath
	case definition.SpecDefinition.PodSpecPath != nil:
		path = definition.SpecDefinition.PodSpecPath
	default:
		return false, nil
	}

	targets, err := evaluator.Evaluate(ctx, *path)
	if err != nil {
		return false, fmt.Errorf("evaluate path %q: %w", *path, err)
	}

	present := false
	absent := false
	for _, target := range targets {
		if target == nil {
			absent = true
		} else {
			present = true
		}
	}
	if present && absent {
		return false, fmt.Errorf("path %q returned both present and absent pod targets", *path)
	}

	return !present, nil
}

func bindCreatedServiceSelectors(networking *models.NetworkingConfiguration, workloadID string) int {
	if networking == nil {
		return 0
	}

	createdServices := 0
	for endpointIndex := range *networking {
		services := (*networking)[endpointIndex].Services
		if services == nil {
			continue
		}
		for serviceIndex := range *services {
			spec := (*services)[serviceIndex].Spec
			if spec == nil {
				continue
			}
			createdServices++
			if spec.Selector == nil {
				spec.Selector = &map[string]string{}
			}
			(*spec.Selector)[networkingWorkloadIDLabel] = workloadID
		}
	}
	return createdServices
}

func stripNetworkingWorkloadIDLabel(value interface{}) {
	switch typed := value.(type) {
	case map[string]interface{}:
		delete(typed, networkingWorkloadIDLabel)
		for _, child := range typed {
			stripNetworkingWorkloadIDLabel(child)
		}
	case map[string]string:
		delete(typed, networkingWorkloadIDLabel)
	case []interface{}:
		for _, child := range typed {
			stripNetworkingWorkloadIDLabel(child)
		}
	}
}

func networkingKartaComponents(factory *kartaresource.ComponentFactory) ([]*kartaresource.Component, error) {
	children, err := factory.GetChildComponents()
	if err != nil {
		return nil, fmt.Errorf("get Karta child components: %w", err)
	}
	root, err := factory.GetRootComponent()
	if err != nil {
		return nil, fmt.Errorf("get Karta root component: %w", err)
	}
	return append(children, root), nil
}

func stampNetworkingWorkloadID(
	instance kartaresource.ExtractedInstance,
	workloadID string,
	definition kartav1alpha1.ComponentDefinition,
) (target, supported, changed bool) {
	var labels *map[string]string
	switch {
	case instance.PodTemplateSpec != nil:
		labels = &instance.PodTemplateSpec.Labels
	case instance.FragmentedPodSpec != nil:
		if definition.SpecDefinition == nil {
			return true, false, false
		}
		fragmentedDefinition := definition.SpecDefinition.FragmentedPodSpecDefinition
		if fragmentedDefinition == nil || fragmentedDefinition.LabelsPath == nil || *fragmentedDefinition.LabelsPath == "" {
			return true, false, false
		}
		labels = &instance.FragmentedPodSpec.Labels
	case instance.PodSpec != nil:
		return true, false, false
	default:
		return false, true, false
	}

	if *labels == nil {
		*labels = map[string]string{}
	}
	previous, found := (*labels)[networkingWorkloadIDLabel]
	(*labels)[networkingWorkloadIDLabel] = workloadID
	return true, true, !found || previous != workloadID
}

func updateNetworkingKartaComponent(
	ctx context.Context,
	component *kartaresource.Component,
	instances map[string]kartaresource.ExtractedInstance,
) error {
	for _, instance := range instances {
		switch {
		case instance.FragmentedPodSpec != nil:
			fragmentedSpecs := make(map[string]kartaresource.FragmentedPodSpec, len(instances))
			for id, item := range instances {
				if item.FragmentedPodSpec == nil {
					return fmt.Errorf("Karta component contains mixed pod definition types")
				}
				fragmentedSpecs[id] = *item.FragmentedPodSpec
			}
			return component.UpdateFragmentedPodSpec(ctx, fragmentedSpecs)
		case instance.PodTemplateSpec != nil:
			podTemplates := make(map[string]corev1.PodTemplateSpec, len(instances))
			for id, item := range instances {
				if item.PodTemplateSpec == nil {
					return fmt.Errorf("Karta component contains mixed pod definition types")
				}
				podTemplates[id] = *item.PodTemplateSpec
			}
			return component.UpdatePodTemplateSpec(ctx, podTemplates)
		case instance.PodSpec != nil:
			return fmt.Errorf("bare PodSpec has no writable pod metadata")
		default:
			return nil
		}
	}
	return nil
}
