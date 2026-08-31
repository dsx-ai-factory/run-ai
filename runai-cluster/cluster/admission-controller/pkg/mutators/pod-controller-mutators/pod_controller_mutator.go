/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_controller_mutators

import (
	"log"

	"k8s.io/api/admission/v1beta1"

	kube_client2 "github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
)

type PodControllerMutatingFunc func(*types.PodControllerCommon, *v1beta1.AdmissionRequest) error

type PodControllerMutator struct {
	client kube_client2.Client
	params *types.MutatingParams
}

func NewPodControllerMutator(client kube_client2.Client, params *types.MutatingParams) *PodControllerMutator {
	return &PodControllerMutator{
		client: client,
		params: params,
	}
}

func (pcm *PodControllerMutator) ControllerMutateAll(podController *types.PodControllerCommon,
	ar *v1beta1.AdmissionRequest) error {
	if !pcm.shouldMutateController(podController) {
		log.Printf("Job isn't a runai job, scheduler: %v - do nothing",
			podController.Spec.Template.Spec.SchedulerName)
		return nil
	}

	singleActionMutators := pcm.listSingleActionMutators()
	for _, mutator := range singleActionMutators {
		err := mutator(podController, ar)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pcm *PodControllerMutator) shouldMutateController(podController *types.PodControllerCommon) bool {
	return podController.Spec.Template.Spec.SchedulerName == types.RunaiScheduler
}

func (pcm *PodControllerMutator) listSingleActionMutators() []PodControllerMutatingFunc {
	mutators := []PodControllerMutatingFunc{
		pcm.mutateJobResources,
		pcm.mutateNodeAffinityForNodeType,
	}

	return mutators
}
