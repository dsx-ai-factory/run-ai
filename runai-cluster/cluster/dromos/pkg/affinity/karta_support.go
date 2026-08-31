/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package affinity

import (
	"errors"
	"fmt"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"
)

var (
	ErrNoSpecDefinition   = errors.New("karta has no component that carries a pod spec")
	ErrNoNodeAffinityPath = errors.New("karta component with a fragmented pod spec requires nodeAffinityPath")
)

// SupportsNodeAffinity checks if node affinity can be written to every
// pod-bearing component of karta.
func SupportsNodeAffinity(karta *kartav1alpha1.Karta) error {
	structure := karta.Spec.StructureDefinition
	components := append([]kartav1alpha1.ComponentDefinition{structure.RootComponent}, structure.ChildComponents...)

	hasPodBearingComponent := false
	for _, component := range components {
		switch supportsNodeAffinity, err := componentSupportsNodeAffinity(&component); {
		case err != nil:
			return fmt.Errorf("component %s: %w", component.Name, err)
		case supportsNodeAffinity:
			hasPodBearingComponent = true
		}
	}

	if !hasPodBearingComponent {
		return ErrNoSpecDefinition
	}
	return nil
}

func componentSupportsNodeAffinity(def *kartav1alpha1.ComponentDefinition) (bool, error) {
	if def.SpecDefinition == nil {
		return false, nil
	}
	if def.SpecDefinition.PodTemplateSpecPath != nil || def.SpecDefinition.PodSpecPath != nil {
		return true, nil
	}

	fragmented := def.SpecDefinition.FragmentedPodSpecDefinition
	if fragmented == nil {
		return false, nil
	}
	if ptr.Deref(fragmented.NodeAffinityPath, "") == "" {
		return false, ErrNoNodeAffinityPath
	}
	return true, nil
}
