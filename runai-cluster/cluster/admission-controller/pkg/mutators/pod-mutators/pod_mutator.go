/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"log"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"

	pod_identifier "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/pod-identifier"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"

	kube_client2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/kube-client"
)

type PodMutatingFunc func(*v1.Pod, *v1beta1.AdmissionRequest) error

type PodMutator struct {
	client    kube_client2.Client
	params    *types.MutatingParams
	draParser *dra.DRAParser
}

func NewPodMutator(client kube_client2.Client, params *types.MutatingParams, draParser *dra.DRAParser) *PodMutator {
	return &PodMutator{
		client:    client,
		params:    params,
		draParser: draParser,
	}
}

func (pm *PodMutator) MutatePod(pod *v1.Pod, request *v1beta1.AdmissionRequest) error {
	namespaceName := getDestinationNamespace(pod, request)
	namespace, err := pm.client.GetNamespace(namespaceName)
	if err != nil {
		return err
	}

	if !pod_identifier.ShouldMutatePod(pod, namespace, pm.params.ScheduleWorkflowPods) {
		log.Printf("Pod should not be mutated: %v/%v - do nothing", pod.Namespace, pod.Name)
		return nil
	}

	singleActionMutators := pm.listSingleActionPodMutators()
	for _, mutator := range singleActionMutators {
		err = mutator(pod, request)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *PodMutator) listSingleActionPodMutators() []PodMutatingFunc {
	mutators := []PodMutatingFunc{
		pm.mutateSchedulerName,
		pm.mutateProjectNameLabel,
		pm.mutatePodResources,
		pm.setEnvVariablesForPod,
		pm.mutateNodeAffinityForNodeType,
		pm.mutateNodeAffinityForDefaultNodePools,
		pm.patchVolumesAndMounts,
		pm.setAutoScalerAnnotations,
	}

	if pod_identifier.IsGKECluster(pm.params) || pod_identifier.IsOKECluster(pm.params) {
		mutators = append(mutators, pm.mutateFractionTolerationPod)
	}

	return mutators
}
