/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/mutators/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/patch"
)

const (
	autoscalerEvictAnnotationKey         = "cluster-autoscaler.kubernetes.io/safe-to-evict"
	autoscalerEvictAnnotationValue       = "false"
	karpenterDoNotDisruptAnnotationKey   = karpenterv1.DoNotDisruptAnnotationKey
	karpenterDoNotDisruptAnnotationValue = "true"
)

func (pm *PodMutator) setAutoScalerAnnotations(pod *v1.Pod, _ *v1beta1.AdmissionRequest) error {
	priorityClassName := getPriorityClassName(pod)
	if !utils.IsKnownPriorityClassName(priorityClassName) {
		return nil
	}
	preemptible := utils.IsPreemptible(priorityClassName)

	if !preemptible {
		patch.AddAnnotationToObjectMeta(&pod.ObjectMeta,
			autoscalerEvictAnnotationKey, autoscalerEvictAnnotationValue)
		patch.AddAnnotationToObjectMeta(&pod.ObjectMeta,
			karpenterDoNotDisruptAnnotationKey, karpenterDoNotDisruptAnnotationValue)
	}
	return nil
}
