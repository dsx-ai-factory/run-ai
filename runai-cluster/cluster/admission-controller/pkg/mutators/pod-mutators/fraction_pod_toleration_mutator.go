/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/patch"
	podIdentifier "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/pod-identifier"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
)

func (pm *PodMutator) mutateFractionTolerationPod(pod *v1.Pod, _ *v1beta1.AdmissionRequest) error {
	if !podIdentifier.IsFractionalPod(pod.Annotations, pod.Spec.Containers) {
		return nil
	}

	log.Printf("Handling shared pod for GKE")
	patch.AddTolerationToSpec(&pod.Spec, types.GpuResource, "", v1.TolerationOpExists, v1.TaintEffectNoSchedule)
	return nil
}
