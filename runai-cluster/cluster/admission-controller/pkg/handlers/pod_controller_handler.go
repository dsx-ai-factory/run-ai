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

	kube_client2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/kube-client"
	cMut "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/mutators/pod-controller-mutators"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodControllerHandler struct {
	mutator *cMut.PodControllerMutator
}

func NewPodControllerHandler(client kube_client2.Client, params *types.MutatingParams) *PodControllerHandler {
	return &PodControllerHandler{
		mutator: cMut.NewPodControllerMutator(client, params),
	}
}

func (pch *PodControllerHandler) HandleMutateAll(review v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {
	return pch.mutatingPodControllerHandlerFunc(review, pch.mutator.ControllerMutateAll)
}

func (pch *PodControllerHandler) mutatingPodControllerHandlerFunc(review v1beta1.AdmissionReview,
	mutatingFunc cMut.PodControllerMutatingFunc) (*v1beta1.AdmissionResponse, error) {
	var podController *types.PodControllerCommon
	request := review.Request
	if err := json.Unmarshal(request.Object.Raw, &podController); err != nil {
		return nil, fmt.Errorf("unable unmarshal pod controller json object %w", err)
	}

	mutatedPodController := podController.DeepCopy()
	err := mutatingFunc(mutatedPodController, request)
	if err != nil {
		return nil, err
	}

	patches, err := getPatchesFromDiff(podController, mutatedPodController)
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

func getPatchesFromDiff(orig *types.PodControllerCommon,
	mutated *types.PodControllerCommon) ([]byte, error) {
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
	return marshaledPatch, nil
}
