/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/client-go/kubernetes"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"

	kube_client2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/utils"
)

type RequestHandler struct {
	podControllerHandler *PodControllerHandler
	podHandler           *PodHandler
}

type Handler func(admReview v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error)

func NewRequestHandler(
	ctx context.Context,
	minAllocationEnforcementBytes int,
	sharedDirectoryHostPath string,
	mpsDirectoryHostPath string,
	coreDebugLogsDirectoryHostPath string,
	coreDebugLevel string,
	coreDebugFile string,
	coreAnalytics bool,
	isKubevirtEnabledCluster bool,
	scheduleWorkflowPods bool,
	specificClusterType string,
	overProvisioningRoute bool,
) *RequestHandler {
	handlerParams := &types.MutatingParams{
		MinAllocationEnforcementBytes:  minAllocationEnforcementBytes,
		SharedDirectoryHostPath:        sharedDirectoryHostPath,
		MpsDirectoryHostPath:           mpsDirectoryHostPath,
		CoreDebugLogsDirectoryHostPath: coreDebugLogsDirectoryHostPath,
		CoreDebugLevel:                 coreDebugLevel,
		CoreDebugFile:                  coreDebugFile,
		CoreAnalytics:                  coreAnalytics,
		IsKubevirtEnabledCluster:       isKubevirtEnabledCluster,
		ScheduleWorkflowPods:           scheduleWorkflowPods,
		SpecificClusterType:            specificClusterType,
		OverProvisioningRoute:          overProvisioningRoute,
	}

	client, err := kube_client2.NewClient(getClientConfig(), initScheme())
	if err != nil {
		panic(err.Error())
	}
	err = client.StartAndWaitForSync(ctx)
	if err != nil {
		panic(err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(getClientConfig())
	if err != nil {
		panic(err.Error())
	}
	draParser := dra.NewDRAParser(client.GetCachedReader(), kubeClient)

	return &RequestHandler{
		podControllerHandler: NewPodControllerHandler(client, handlerParams),
		podHandler:           NewPodHandler(client, handlerParams, draParser),
	}
}

func (handler *RequestHandler) HandleMutatePod(w http.ResponseWriter, r *http.Request) {
	handleRequest(w, r, handler.podHandler.HandleMutatePod)
}

func (handler *RequestHandler) HandleMutatePodController(w http.ResponseWriter, r *http.Request) {
	handleRequest(w, r, handler.podControllerHandler.HandleMutateAll)
}

func handleRequest(w http.ResponseWriter, r *http.Request, handler Handler) {
	admReview, err := extractAdmissionReviewFromRequest(w, r)
	if err != nil {
		return
	}

	result, err := runHandlerWithAdmissionReview(w, handler, admReview)
	if err != nil {
		return
	}

	applyAdmissionResultToResponse(w, result)
}

func sanitizeForLog(err error) string {
	replacer := strings.NewReplacer("\n", "", "\r", "")
	return replacer.Replace(err.Error())
}

func extractAdmissionReviewFromRequest(w http.ResponseWriter, r *http.Request) (v1beta1.AdmissionReview, error) {
	// read the body / request
	body, err := io.ReadAll(r.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(r.Body)
	if err != nil {
		log.Println(sanitizeForLog(err))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return v1beta1.AdmissionReview{}, err
	}

	log.Printf("%s\n", utils.IndentJSONString("recv -> ", body))

	// unmarshal request into AdmissionReview struct
	admReview := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return v1beta1.AdmissionReview{}, err
	}
	return admReview, nil
}

func runHandlerWithAdmissionReview(w http.ResponseWriter, handler Handler,
	admReview v1beta1.AdmissionReview) ([]byte, error) {
	// validate the request
	response, err := handler(admReview)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return nil, err
	}
	admReview.Response = response
	return json.Marshal(admReview)
}

func applyAdmissionResultToResponse(w http.ResponseWriter, result []byte) {
	log.Printf("%s\n", utils.IndentJSONString("resp -> ", result))
	// and write it back
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(result)
	if err != nil {
		log.Printf("Error while writing: %v", err)
	}
}
