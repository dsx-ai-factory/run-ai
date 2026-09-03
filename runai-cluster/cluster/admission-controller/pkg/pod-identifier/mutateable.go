/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_identifier

import (
	"log"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	v1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
)

const (
	gkeClusterName    = "gke"
	okeClusterName    = "oke"
	azureMlLabelKey   = "ml.azure.com/compute"
	notebookLabelName = "notebook-name"
)

func IsRunaiPod(pod *v1.Pod) bool {
	return pod.Spec.SchedulerName == types.RunaiScheduler
}

func IsOwnerWorkflow(pod *v1.Pod) bool {
	// kubeflow pipelines also have a workflow owner - so this case is a generalization of it
	for _, ownerReference := range pod.OwnerReferences {
		if ownerReference.Kind == "Workflow" {
			log.Printf("Pod %s has a workflow owner - could be a kubeflow pipeline", pod.Name)
			return true
		}
	}
	return false
}

func IsKubeflowNotebook(pod *v1.Pod) bool {
	for labelName := range pod.Labels {
		if labelName == notebookLabelName {
			log.Printf("Pod %s is a kubeflow notebook", pod.Name)
			return true
		}
	}
	return false
}

func IsAzureMLJob(pod *v1.Pod) bool {
	_, found := pod.Labels[azureMlLabelKey]
	return found
}

func IsGKECluster(params *types.MutatingParams) bool {
	return strings.ToLower(params.SpecificClusterType) == gkeClusterName
}

func IsOKECluster(params *types.MutatingParams) bool {
	return strings.ToLower(params.SpecificClusterType) == okeClusterName
}

func isSchedulerEnforcedForNamespace(namespace *v1.Namespace) bool {
	enforce, found := namespace.Annotations[constants.EnforceSchedulerAnnotationName]
	if !found {
		return false
	}

	val, err := strconv.ParseBool(enforce)
	if err != nil {
		log.Printf("Could not parse %s annotation value for namespace %s. Value: %t",
			constants.EnforceSchedulerAnnotationName, namespace.Name, val)
		return false
	}
	return val
}

func ShouldMutatePod(pod *v1.Pod, namesapce *v1.Namespace, scheduleWorkflowPods bool) bool {
	return scheduleWorkflowPods && IsOwnerWorkflow(pod) ||
		IsKubeflowNotebook(pod) ||
		IsAzureMLJob(pod) ||
		IsRunaiPod(pod) ||
		isSchedulerEnforcedForNamespace(namesapce)
}
