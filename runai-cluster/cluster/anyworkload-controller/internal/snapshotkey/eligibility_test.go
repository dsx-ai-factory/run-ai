/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotkey

import (
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/karta"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// TargetContainer duplicates karta's unexported per-container GPU predicate,
// so the two can drift. They must not: the plugin decides eligibility with
// karta.HasGPU, and a pod that is eligible but unkeyable opts into a feature
// that then never runs, while the reverse keys a pod the plugin never offered.
//
// karta.HasGPU is "any container has a GPU"; TargetContainer is the stricter
// "exactly one does". The invariant is therefore one-directional:
//
//	TargetContainer succeeds  =>  karta.HasGPU is true
//	karta.HasGPU is true      =>  TargetContainer succeeds, or reports the
//	                              several-container case
func TestAgreesWithKartaHasGPU(t *testing.T) {
	gpu := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")},
	}
	gpuRequest := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")},
	}
	// karta matches on a substring, so a vendor-prefixed name counts too.
	gpuShared := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{"nvidia.com/gpu-shared": resource.MustParse("1")},
	}
	cpu := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{"cpu": resource.MustParse("1")},
	}

	cases := map[string][]corev1.Container{
		"no containers":        {},
		"one cpu container":    {{Name: "a", Resources: cpu}},
		"one gpu limit":        {{Name: "a", Resources: gpu}},
		"one gpu request":      {{Name: "a", Resources: gpuRequest}},
		"vendor-prefixed gpu":  {{Name: "a", Resources: gpuShared}},
		"gpu beside cpu":       {{Name: "a", Resources: cpu}, {Name: "b", Resources: gpu}},
		"two gpu containers":   {{Name: "a", Resources: gpu}, {Name: "b", Resources: gpu}},
		"three gpu containers": {{Name: "a", Resources: gpu}, {Name: "b", Resources: gpu}, {Name: "c", Resources: gpuRequest}},
	}

	for name, containers := range cases {
		t.Run(name, func(t *testing.T) {
			pod := gpuPod()
			pod.Spec.Containers = containers

			instance := kartaresource.ExtractedInstance{
				PodTemplateSpec: &corev1.PodTemplateSpec{Spec: pod.Spec},
			}
			eligible := karta.HasGPU(instance)

			target, err := TargetContainer(pod)

			if err == nil && !eligible {
				t.Fatalf("TargetContainer returned %q but karta.HasGPU says the pod has no GPU", target)
			}
			if eligible && err != nil && !errors.Is(err, ErrNotKeyable) {
				t.Fatalf("karta.HasGPU is true but TargetContainer failed with an unexpected error: %v", err)
			}
			if !eligible && err == nil {
				t.Fatal("karta.HasGPU is false but the pod is keyable")
			}
		})
	}
}
