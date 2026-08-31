/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_controller_mutators

import (
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/mutators/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/patch"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
)

const (
	nodeTypeAffinityExpressionKey = "run.ai/type"
	interactiveAffinity           = "nodeAffinityInteractive"
	trainAffinity                 = "nodeAffinityTrain"
)

func (pcm *PodControllerMutator) mutateNodeAffinityForNodeType(podController *types.PodControllerCommon,
	request *v1beta1.AdmissionRequest) error {
	namespace := getDestinationNamespace(podController, request)
	projectName := utils.GetDesiredProjectName(pcm.client, namespace, podController.Spec.Template.Spec.Containers)

	err := pcm.setRunaiNodeAffinityForNodeType(podController, projectName)
	return err
}

func (pm *PodControllerMutator) setRunaiNodeAffinityForNodeType(podController *types.PodControllerCommon,
	projectName string) error {
	podSpec := &podController.Spec.Template.Spec
	project, err := utils.GetProjectResource(pm.client, projectName)
	if err != nil {
		return err
	}

	interactive := utils.IsInteractiveByPriority(pm.client.GetCachedReader(),
		podController.Labels, podController.Spec.Template.Spec.PriorityClassName)
	setNodeAffinityForNodeType(podSpec, project, interactive)

	return nil
}

func setNodeAffinityForNodeType(podSpec *v1.PodSpec, project *unstructured.Unstructured, interactive bool) {
	// The node types moved from Project.spec to an annotation with the kai.resources CRD.
	nodeAffinityTypes := utils.NodeAffinityTypesFromAnnotation(project, interactive)

	if len(nodeAffinityTypes) == 0 {
		return
	}

	matchExpressionToAdd := v1.NodeSelectorRequirement{
		Key:      nodeTypeAffinityExpressionKey,
		Operator: v1.NodeSelectorOpIn,
		Values:   nodeAffinityTypes,
	}
	patch.AddNodeAffinityToPodSpec(podSpec, []v1.NodeSelectorRequirement{matchExpressionToAdd})
}
