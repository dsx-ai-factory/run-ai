/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	nodes "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type BackendDistributedClient struct {
	workloadsClient *openapi.APIClient
	nodesService    *nodes.BackendNodeService
}

func NewDistributedClient(urlStr, token string) (*BackendDistributedClient, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}

	nodesService, err := nodes.CreateBackendNodeService(urlStr, token)
	if err != nil {
		return nil, err
	}

	return &BackendDistributedClient{workloadsClient: openapi.NewAPIClient(conf), nodesService: nodesService}, nil
}

func (bwc *BackendDistributedClient) Delete(ctx context.Context, uuid string) error {
	_, response, err := bwc.workloadsClient.DistributedAPI.DeleteDistributed(ctx, uuid).Execute()
	return handleDeleteResponse(response, err)
}

func (bwc *BackendDistributedClient) Suspend(ctx context.Context, uuid string) error {
	return handleSuspendResponse(bwc.workloadsClient.DistributedAPI.SuspendDistributed(ctx, uuid).Execute())
}

func (bwc *BackendDistributedClient) Resume(ctx context.Context, uuid string) error {
	return handleSuspendResponse(bwc.workloadsClient.DistributedAPI.ResumeDistributed(ctx, uuid).Execute())
}

func (bwc *BackendDistributedClient) Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	request := bwc.workloadsClient.DistributedAPI.CreateDistributed(ctx)
	distributed := *openapi.NewDistributedCreationRequest(name, projectId, clusterId)
	wsSpec, err := spec.GenerateDistributedSpec(ctx, clusterId, bwc.nodesService)
	if err != nil {
		return "", err
	}
	distributed.SetSpec(wsSpec.GetSpec())
	distributed.SetUseGivenNameAsPrefix(spec.HasPrefixName())
	if wsSpec.HasMasterSpec() {
		distributed.SetMasterSpec(wsSpec.GetMasterSpec())
	}
	if wsSpec.HasMasterSpecSameAsWorker() {
		distributed.SetMasterSpecSameAsWorker(wsSpec.GetMasterSpecSameAsWorker())
	}
	if spec.GetTemplateId() != "" {
		distributed.SetTemplateId(spec.GetTemplateId())
	}
	if assets := spec.GetAssetSubmissionIds(); assets != nil {
		distributed.SetAssets(*assets)
	}
	if masterAssets := spec.GetMasterAssetSubmissionIds(); masterAssets != nil {
		distributed.SetMasterAssets(*masterAssets)
	}
	if spec.IsDryRun() {
		var byteSpec []byte
		byteSpec, err = distributed.MarshalJSON()
		return string(byteSpec), err
	}
	request = request.DistributedCreationRequest(distributed)

	var distributedName string
	distributedName, err = handleSubmitResponse(request.Execute())
	return distributedName, err
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

func handleSuspendResponse(_ *openapi.HttpResponse, response *http.Response, err error) error {
	if err != nil {
		return clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		return response.Body.Close()
	}
	return nil
}

func handleSubmitResponse(distributed *openapi.DistributedWorkload, response *http.Response, err error) (string, error) {
	if err != nil {
		return "", clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	return distributed.Name, nil
}
