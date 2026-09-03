/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/httputils"
	workloadmanager "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

type WorkloadV2 struct {
	apiClient *workloadmanager.APIClient
}
type SubmitResponse struct {
	Name string
	Id   string
}

func NewWorkloadV2(urlStr, token string) (*WorkloadV2, error) {
	conf, err := config.CreateWorkloadManagerServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &WorkloadV2{apiClient: workloadmanager.NewAPIClient(conf)}, nil
}

// Submit submits a workload using a workload blob using the workload v2 API
// returns SubmitResponse struct which contain the workload name and id.
func (wmc *WorkloadV2) Submit(
	ctx context.Context,
	name string,
	useGivenNameAsPrefix bool,
	projectId string,
	priority string,
	category string,
	workloadBlob any,
	clusterId string,
	preemptibility workloadmanager.NullablePreemptibility,
	mnnvl workloadmanager.NullableMnnvlMode,
) (SubmitResponse, error) {
	metadata := workloadmanager.NewWorkloadV2MetadataCreateParams()
	var submitResponse SubmitResponse
	metadata.SetName(name)
	metadata.SetUseGivenNameAsPrefix(useGivenNameAsPrefix)

	if projectId != "" {
		metadata.SetProjectId(projectId)
	} else {
		metadata.SetClusterId(clusterId)
	}

	if priority != "" {
		metadata.SetPriority(priority)
	}
	if category != "" {
		metadata.SetCategory(category)
	}
	if preemptibility.IsSet() {
		metadata.SetPreemptibility(*preemptibility.Get())
	}
	if mnnvl.IsSet() {
		configuration := workloadmanager.NewWorkloadV2Configuration()
		configuration.SetMnnvl(*mnnvl.Get())
		metadata.SetConfiguration(*configuration)
	}

	createRequest := workloadmanager.NewWorkloadV2CreateRequest(*metadata)
	manifestMap := workloadBlob.(map[string]any)
	createRequest.SetManifest(manifestMap)

	request := wmc.apiClient.WorkloadsV2API.CreateWorkloadV2(ctx)
	request = request.WorkloadV2CreateRequest(*createRequest)

	workloadV2, response, err := request.Execute()
	result, errResponse := httputils.HandleResponse(workloadV2, response, err)
	if errResponse != nil {
		return submitResponse, errResponse
	}

	responseMetadata := result.GetMetadata()
	submitResponse.Name = responseMetadata.Name
	submitResponse.Id = responseMetadata.Id

	return submitResponse, nil
}
func (wmc *WorkloadV2) Get(ctx context.Context, id string) (*workloadmanager.WorkloadV2GetResponse, error) {
	request := wmc.apiClient.WorkloadsV2API.GetWorkloadV2ById(ctx, id)

	workload, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return httputils.HandleResponse(workload, response, err)
}
