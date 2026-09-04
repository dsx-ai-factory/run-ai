/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"gomodules.xyz/jsonpatch/v2"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"

	kube_client2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/kube-client"
	podMutators "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/mutators/pod-mutators"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodHandler struct {
	mutator *podMutators.PodMutator
}

func NewPodHandler(client kube_client2.Client, handlerParams *types.MutatingParams, draParser *dra.DRAParser) *PodHandler {
	return &PodHandler{
		mutator: podMutators.NewPodMutator(client, handlerParams, draParser),
	}
}

func (ph *PodHandler) HandleMutatePod(review v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {
	return mutatingPodHandlerFunc(review, ph.mutator.MutatePod)
}

func mutatingPodHandlerFunc(review v1beta1.AdmissionReview,
	mutatingFunc podMutators.PodMutatingFunc) (*v1beta1.AdmissionResponse, error) {
	var pod *v1.Pod
	request := review.Request
	if err := json.Unmarshal(request.Object.Raw, &pod); err != nil {
		return nil, fmt.Errorf("unable unmarshal pod json object %w", err)
	}

	mutatedPod := pod.DeepCopy()
	err := mutatingFunc(mutatedPod, request)
	if err != nil {
		return nil, err
	}

	patches, err := getPatchesFromPodDiff(pod, mutatedPod)
	if err != nil {
		return nil, err
	}

	response := &v1beta1.AdmissionResponse{}
	response.UID = request.UID
	response.Allowed = true
	pT := v1beta1.PatchTypeJSONPatch
	response.PatchType = &pT
	response.Patch = patches
	response.Result = &metav1.Status{
		Status: "Success",
	}
	return response, nil
}

func getPatchesFromPodDiff(orig *v1.Pod, mutated *v1.Pod) ([]byte, error) {
	origJson, err := json.Marshal(orig)
	if err != nil {
		return nil, err
	}

	mutatedJson, err := json.Marshal(mutated)
	if err != nil {
		return nil, err
	}

	patches, err := jsonpatch.CreatePatch(origJson, mutatedJson)
	if err != nil {
		return nil, err
	}

	log.Printf("Applying the following patches:")
	for idx, patch := range patches {
		log.Printf("%s -> %s", strconv.Itoa(idx), patch)
	}

	marshaledPatch, err := json.Marshal(patches)
	if err != nil {
		log.Printf("could not mashal into JSON, the JSON patch: %s", err)
		return nil, err
	}
	return marshaledPatch, err
}
