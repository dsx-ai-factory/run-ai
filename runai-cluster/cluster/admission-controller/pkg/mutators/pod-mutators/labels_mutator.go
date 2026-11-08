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
	"k8s.io/apimachinery/pkg/types"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/mutators/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/patch"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
)

const ccReadyStateNodeLabel = "nvidia.com/cc.ready.state"

func (pm *PodMutator) mutateProjectNameLabel(pod *v1.Pod, request *v1beta1.AdmissionRequest) error {
	_, found := pod.Labels["project"]
	if found {
		return nil
	}

	namespace := getDestinationNamespace(pod, request)
	projectName := utils.GetDesiredProjectName(pm.client, namespace, pod.Spec.Containers)

	if projectName == "" {
		log.Printf("Warning: Could not identify the runai project related to this pod")
		return nil
	}

	patch.AddLabelToObjectMeta(&pod.ObjectMeta, "project", projectName)
	return nil
}

func (pm *PodMutator) mutateConfidentialLabel(pod *v1.Pod, _ *v1beta1.AdmissionRequest) error {
	if pod.Spec.RuntimeClassName == nil || *pod.Spec.RuntimeClassName == "" {
		return nil
	}

	runtimeClass := &nodev1.RuntimeClass{}
	err := pm.client.GetCachedReader().Get(context.Background(),
		types.NamespacedName{Name: *pod.Spec.RuntimeClassName}, runtimeClass)
	if err != nil {
		log.Printf("Warning: could not read runtime class %s: %v", *pod.Spec.RuntimeClassName, err)
		return nil
	}

	if !isConfidentialRuntimeClass(runtimeClass) {
		return nil
	}

	patch.AddLabelToObjectMeta(&pod.ObjectMeta, constants.ConfidentialComputingLabel, "true")
	return nil
}

func isConfidentialRuntimeClass(runtimeClass *nodev1.RuntimeClass) bool {
	if runtimeClass.Scheduling == nil {
		return false
	}
	return runtimeClass.Scheduling.NodeSelector[ccReadyStateNodeLabel] == "true"
}
