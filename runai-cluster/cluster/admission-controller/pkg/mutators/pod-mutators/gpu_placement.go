/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Knative can disallow nodeSelector on its Revision, but the resulting Pod
// supports it. Pin the final Pod independently of the required-affinity terms
// before removing the scheduling barrier left by the composer.
func (pm *PodMutator) enforceSelectedGPUPlacement(pod *corev1.Pod) error {
	resource, selected := pod.Annotations[constants.AnnotationGPUPlacementResource]
	if !selected {
		return nil
	}
	if !strings.Contains(resource, "/") || strings.Contains(resource, corev1.ResourceDefaultNamespacePrefix) ||
		strings.HasPrefix(resource, corev1.DefaultResourceRequestsPrefix) || len(validation.IsQualifiedName(corev1.DefaultResourceRequestsPrefix+resource)) > 0 {
		return fmt.Errorf("invalid GPU placement resource %q", resource)
	}
	if pod.Spec.NodeSelector == nil {
		pod.Spec.NodeSelector = make(map[string]string)
	}
	pin := func(key, value string) error {
		if existing, found := pod.Spec.NodeSelector[key]; found && existing != value {
			return fmt.Errorf("GPU placement %s=%s conflicts with nodeSelector value %q", key, value, existing)
		}
		pod.Spec.NodeSelector[key] = value
		return nil
	}
	if err := pin(constants.AcceleratorResourceNodeLabel(resource), "true"); err != nil {
		return err
	}
	if model, supplied := pod.Annotations[constants.AnnotationGPUTypeLabel]; supplied {
		key, value, ok := strings.Cut(model, "=")
		if !ok || len(validation.IsQualifiedName(key)) != 0 || value == "" || len(validation.IsValidLabelValue(value)) != 0 {
			return fmt.Errorf("invalid GPU placement label %q", model)
		}
		if err := pin(key, value); err != nil {
			return err
		}
	}
	if pod.Spec.Affinity == nil || pod.Spec.Affinity.NodeAffinity == nil || pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return fmt.Errorf("GPU placement requires node affinity")
	}
	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	if len(terms) == 0 {
		return fmt.Errorf("GPU placement has no guarded scheduling terms")
	}
	for _, term := range terms {
		if !hasGPUPlacementGuard(term.MatchFields, constants.GPUPlacementAdmissionNode) &&
			!hasGPUPlacementGuard(term.MatchFields, constants.GPUPlacementAdmissionNode+"-empty") {
			return fmt.Errorf("GPU placement admission barrier is missing from a required scheduling term")
		}
	}
	// Intersect explicit pool selectors with whatever the normal admission
	// steps established. An annotation alone must never suppress those steps.
	if pools := strings.Fields(pod.Annotations[constants.AnnotationNodepools]); len(pools) != 0 {
		if err := pm.setNodeAffinityForDefaultNodePools(pod, pools); err != nil {
			return err
		}
	}
	// Pool intersection may have expanded OR branches. Release only the
	// prepended pair and leave originally empty terms permanently ineligible.
	terms = pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	for i := range terms {
		if hasGPUPlacementGuard(terms[i].MatchFields, constants.GPUPlacementAdmissionNode) {
			terms[i].MatchFields = terms[i].MatchFields[2:]
		}
	}
	// Mark this Pod as admitted, so reinvocation never consumes reserved-looking
	// fields that a caller had independently put after the composer's pair.
	delete(pod.Annotations, constants.AnnotationGPUPlacementResource)
	return nil
}

func hasGPUPlacementGuard(fields []corev1.NodeSelectorRequirement, name string) bool {
	if len(fields) < 2 {
		return false
	}
	in, notIn := fields[0], fields[1]
	return in.Key == "metadata.name" && notIn.Key == in.Key &&
		in.Operator == corev1.NodeSelectorOpIn && notIn.Operator == corev1.NodeSelectorOpNotIn &&
		slices.Equal(in.Values, []string{name}) && slices.Equal(notIn.Values, in.Values)
}
