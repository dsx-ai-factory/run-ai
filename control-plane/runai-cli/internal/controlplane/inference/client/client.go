/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/httputils"
	nodes "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

type BackendInferenceClient struct {
	apiClient    *openapi.APIClient
	nodesService *nodes.BackendNodeService
}

func NewInferenceClient(urlStr, token string) (*BackendInferenceClient, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}

	nodesService, err := nodes.CreateBackendNodeService(urlStr, token)
	if err != nil {
		return nil, err
	}

	return &BackendInferenceClient{apiClient: openapi.NewAPIClient(conf), nodesService: nodesService}, nil
}

func (bwc *BackendInferenceClient) Delete(ctx context.Context, uuid string) error {
	// Delete standard inference workload
	_, response, err := bwc.apiClient.InferencesAPI.DeleteInference(ctx, uuid).Execute()
	return handleDeleteResponse(response, err)
}

func (bwc *BackendInferenceClient) Update(ctx context.Context, uuid string, clusterId string, spec *spec.GeneralSpec) (string, error) {
	request := bwc.apiClient.InferencesAPI.UpdateInferenceSpec(ctx, uuid)
	inference := *openapi.NewInferenceUpdateRequest()
	infwSpec, err := spec.GenerateInferenceUpdateSpec(ctx, clusterId, bwc.nodesService)
	if err != nil || infwSpec == nil {
		return "", err
	}

	// set the inference spec
	inference.SetSpec(*infwSpec)

	if spec.IsDryRun() {
		var byteSpec []byte
		byteSpec, err = inference.MarshalJSON()
		return string(byteSpec), err
	}

	request = request.InferenceUpdateRequest(inference)

	workload, response, err := request.Execute() //nolint:bodyclose // handle in handler
	_, err = httputils.HandleResponse(workload, response, err)
	if err != nil {
		return "", err
	}

	return workload.Name, err
}

func (bwc *BackendInferenceClient) Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	request := bwc.apiClient.InferencesAPI.CreateInference(ctx)
	inference := *openapi.NewInferenceCreationRequest(name, projectId, clusterId)

	infwSpec, err := spec.GenerateInferenceSpec(ctx, clusterId, bwc.nodesService)
	if err != nil {
		return "", err
	}

	// set the inference spec
	inference.SetSpec(*infwSpec)
	inference.SetUseGivenNameAsPrefix(spec.HasPrefixName())
	if spec.GetTemplateId() != "" {
		inference.SetTemplateId(spec.GetTemplateId())
	}
	if assets := spec.GetAssetSubmissionIds(); assets != nil {
		inference.SetAssets(*assets)
	}
	if spec.IsDryRun() {
		var byteSpec []byte
		byteSpec, err = inference.MarshalJSON()
		return string(byteSpec), err
	}
	request = request.InferenceCreationRequest(inference)

	var inferenceName string
	inferenceName, err = handleBackendResponse(request.Execute())
	return inferenceName, err
}

func handleDeleteResponse(response *http.Response, err error) error {
	if err != nil {
		return clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		return response.Body.Close()
	}
	return nil
}

func handleBackendResponse(inference *openapi.Inference, response *http.Response, err error) (string, error) {
	if err != nil {
		return "", clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	return inference.Name, nil
}
