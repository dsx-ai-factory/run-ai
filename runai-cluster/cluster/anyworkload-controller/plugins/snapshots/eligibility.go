/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	kartautils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/karta"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// optInAnnotation is the workload's own opt-in (REQ-02-1). Its pods never carry
// it, so MutateResource copies it down as the label the restore webhook selects by.
const optInAnnotation = "run.ai/snapshot"

// errIneligible is reported: the workload asked for snapshots, and cannot have them.
var errIneligible = errors.New("snapshots are off for this workload")

// checkEligibility returns nil for an opted-in workload that can be captured and
// restored, or errIneligible with the first check it fails. Callers check
// optedIn first.
func checkEligibility(ctx context.Context, template *unstructured.Unstructured, karta *kartav1alpha1.Karta) error {
	instances, err := podTemplates(ctx, template, karta)
	if err != nil {
		return err
	}

	if len(instances) != 1 {
		return fmt.Errorf("%w: it has %d pod templates, and a capture copies a single pod",
			errIneligible, len(instances))
	}
	podTemplate := instances[0]

	// The restore webhook selects pods by the label MutateResource stamps.
	if !kartautils.CanStampPodLabels(podTemplate) {
		return fmt.Errorf("%w: its pods cannot carry labels, so the restore webhook never sees them", errIneligible)
	}

	containers := podContainers(podTemplate)
	target, err := snapshotkey.TargetContainer(&corev1.Pod{Spec: corev1.PodSpec{Containers: containers}})
	if err != nil {
		return fmt.Errorf("%w: %w", errIneligible, err)
	}
	for _, container := range containers {
		// A restore pod refuses a target with a seccomp profile of its own, and
		// capture doesn't check it, so without this every restore would fail.
		if container.Name == target && container.SecurityContext != nil &&
			container.SecurityContext.SeccompProfile != nil {
			return fmt.Errorf("%w: container %q sets its own seccomp profile, which a restore pod cannot keep",
				errIneligible, target)
		}
		// Same for a startup probe: the restore gate replaces it.
		if container.Name == target && container.StartupProbe != nil {
			return fmt.Errorf("%w: container %q defines its own startup probe, which a restore pod cannot keep",
				errIneligible, target)
		}
	}
	return nil
}

// podTemplates returns every pod template Karta finds in the workload.
func podTemplates(
	ctx context.Context,
	template *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
) ([]kartaresource.ExtractedInstance, error) {
	components, err := kartautils.GetAllComponents(kartaresource.NewComponentFactoryFromObject(karta, template))
	if err != nil {
		return nil, err
	}

	var instances []kartaresource.ExtractedInstance
	for _, component := range components {
		extracted, err := component.GetExtractedInstances(ctx)
		if err != nil {
			return nil, fmt.Errorf("extract pod templates: %w", err)
		}
		for _, instance := range extracted {
			// Components such as a scaling group have no pod of their own.
			if instance.PodTemplateSpec != nil || instance.PodSpec != nil || instance.FragmentedPodSpec != nil {
				instances = append(instances, instance)
			}
		}
	}
	return instances, nil
}

func optedIn(workload *unstructured.Unstructured) bool {
	return workload.GetAnnotations()[optInAnnotation] == "true"
}

func podContainers(instance kartaresource.ExtractedInstance) []corev1.Container {
	switch {
	case instance.PodTemplateSpec != nil:
		return instance.PodTemplateSpec.Spec.Containers
	case instance.FragmentedPodSpec != nil:
		return instance.FragmentedPodSpec.Containers
	}
	return nil
}
