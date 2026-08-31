/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"
	nodes "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type BackendDistributedInferenceClient struct {
	apiClient    *openapi.APIClient
	nodesService *nodes.BackendNodeService
}

func NewDistributedInferenceClient(urlStr, token string) (*BackendDistributedInferenceClient, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}

	nodesService, err := nodes.CreateBackendNodeService(urlStr, token)
	if err != nil {
		return nil, err
	}

	return &BackendDistributedInferenceClient{apiClient: openapi.NewAPIClient(conf), nodesService: nodesService}, nil
}

func (bdc *BackendDistributedInferenceClient) Delete(ctx context.Context, uuid string) error {
	_, response, err := bdc.apiClient.DistributedInferencesAPI.DeleteDistributedInference(ctx, uuid).Execute()
	return handleDeleteResponse(response, err)
}

func (bdc *BackendDistributedInferenceClient) Update(ctx context.Context, uuid string, clusterId string, spec *spec.GeneralSpec) (string, error) {
	request := bdc.apiClient.DistributedInferencesAPI.UpdateDistributedInferenceSpec(ctx, uuid)
	updateReq := *openapi.NewUpdateRequest()
	updateSpec := openapi.NewUpdateSpecSpec()

	updateSpec.SetReplicas(spec.GetFlags().DistributedInference.Replicas)

	updateReq.SetSpec(*updateSpec)

	if spec.IsDryRun() {
		var byteSpec []byte
		byteSpec, err := updateReq.MarshalJSON()
		return string(byteSpec), err
	}

	request = request.UpdateRequest(updateReq)

	workload, response, err := request.Execute() //nolint:bodyclose // handle in handler
	_, err = httputils.HandleResponse(workload, response, err)
	if err != nil {
		return "", err
	}

	return workload.Name, err
}

func (bdc *BackendDistributedInferenceClient) Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	request := bdc.apiClient.DistributedInferencesAPI.CreateDistributedInference(ctx)
	distributedInference := *openapi.NewDistributedInferenceCreationRequest(name, projectId, clusterId)

	distInfSpec, err := spec.GenerateDistributedInferenceSpec(ctx, clusterId, bdc.nodesService)
	if err != nil {
		return "", err
	}

	// set the distributed inference spec
	distributedInference.SetSpec(*distInfSpec)
	distributedInference.SetUseGivenNameAsPrefix(spec.HasPrefixName())
	if spec.GetTemplateId() != "" {
		distributedInference.SetTemplateId(spec.GetTemplateId())
	}
	if spec.IsDryRun() {
		var byteSpec []byte
		byteSpec, err = distributedInference.MarshalJSON()
		return string(byteSpec), err
	}
	request = request.DistributedInferenceCreationRequest(distributedInference)

	inference, response, err := request.Execute()
	if err != nil {
		return "", clierrors.HandleResponseError(response, err)
	}
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	return inference.Name, nil
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
