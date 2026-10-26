/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"context"
	"log"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	nodev1 "k8s.io/api/node/v1"
	"k8s.io/utils/ptr"
)

const nvidiaGpuResourceName = "nvidia.com/gpu"

func (pm *PodMutator) mutateConfidentialRuntimeClass(pod *v1.Pod, _ *v1beta1.AdmissionRequest) error {
	if !pm.params.ConfidentialComputing {
		return nil
	}

	if pod.Spec.RuntimeClassName != nil && *pod.Spec.RuntimeClassName != "" {
		return nil
	}

	if !requestsNvidiaGpu(pod) {
		return nil
	}

	runtimeClasses := &nodev1.RuntimeClassList{}
	if err := pm.client.GetCachedReader().List(context.Background(), runtimeClasses); err != nil {
		log.Printf("Warning: could not list runtime classes: %v", err)
		return nil
	}

	runtimeClass := confidentialRuntimeClass(runtimeClasses.Items)
	if runtimeClass == nil {
		log.Printf("Warning: confidential computing is on but no confidential runtime class was found")
		return nil
	}

	pod.Spec.RuntimeClassName = ptr.To(runtimeClass.Name)
	applyRuntimeClassScheduling(pod, runtimeClass)
	return nil
}

func requestsNvidiaGpu(pod *v1.Pod) bool {
	containers := append(pod.Spec.InitContainers, pod.Spec.Containers...) //nolint:gocritic
	for i := range containers {
		quantity, found := containers[i].Resources.Limits[nvidiaGpuResourceName]
		if found && !quantity.IsZero() {
			return true
		}
		quantity, found = containers[i].Resources.Requests[nvidiaGpuResourceName]
		if found && !quantity.IsZero() {
			return true
		}
	}
	return false
}

func confidentialRuntimeClass(runtimeClasses []nodev1.RuntimeClass) *nodev1.RuntimeClass {
	for i := range runtimeClasses {
		if isConfidentialRuntimeClass(&runtimeClasses[i]) {
			return &runtimeClasses[i]
		}
	}
	return nil
}

// The built-in RuntimeClass admission plugin runs before mutating webhooks, so a pod that
// receives its runtime class here never has the class's scheduling and overhead merged.
func applyRuntimeClassScheduling(pod *v1.Pod, runtimeClass *nodev1.RuntimeClass) {
	if runtimeClass.Scheduling != nil {
		if pod.Spec.NodeSelector == nil {
			pod.Spec.NodeSelector = map[string]string{}
		}
		for key, value := range runtimeClass.Scheduling.NodeSelector {
			if _, found := pod.Spec.NodeSelector[key]; !found {
				pod.Spec.NodeSelector[key] = value
			}
		}

		pod.Spec.Tolerations = append(pod.Spec.Tolerations, runtimeClass.Scheduling.Tolerations...)
	}

	if runtimeClass.Overhead != nil && pod.Spec.Overhead == nil {
		pod.Spec.Overhead = runtimeClass.Overhead.PodFixed
	}
}
