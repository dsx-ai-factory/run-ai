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
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

type BackendNimClient struct {
	apiClient *openapi.APIClient
}

func NewNimClient(urlStr, token string) (*BackendNimClient, error) {
	conf, err := config.CreateWorkloadManagerServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &BackendNimClient{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (bnc *BackendNimClient) Delete(ctx context.Context, uuid string) error {
	_, response, err := bnc.apiClient.WorkloadsV2API.DeleteWorkloadV2(ctx, uuid).Execute()
	return handleDeleteResponse(response, err)
}

func (bnc *BackendNimClient) Update(ctx context.Context, uuid string, clusterId string, generalSpec *spec.GeneralSpec) (string, error) {
	nimSpec, err := generalSpec.GenerateNimServiceSpec(ctx, clusterId)
	if err != nil {
		return "", err
	}

	request := bnc.apiClient.NVIDIANIMAPI.UpdateNimServiceSpec(ctx, uuid)
	updateReq := *openapi.NewNimServiceUpdateRequest(*nimSpec)

	fieldsToRemove := buildFieldsToRemove(generalSpec.GetFlags().Nim.Remove)

	if nimSpec.HasAutoscaling() {
		fieldsToRemove = append(fieldsToRemove, openapi.NIMSERVICEREMOVABLEFIELDS_REPLICAS)
	}

	if len(fieldsToRemove) > 0 {
		updateReq.SetFieldsToRemove(fieldsToRemove)
	}

	if generalSpec.IsDryRun() {
		byteSpec, err := updateReq.MarshalJSON()
		return string(byteSpec), err
	}

	request = request.NimServiceUpdateRequest(updateReq)
	workload, response, err := request.Execute()
	_, err = httputils.HandleResponse(workload, response, err)
	if err != nil {
		return "", err
	}
	return workload.Metadata.Name, err
}

func buildFieldsToRemove(removeFields []string) []openapi.NimServiceRemovableFields {
	if len(removeFields) == 0 {
		return nil
	}

	var fields []openapi.NimServiceRemovableFields
	for _, field := range removeFields {
		switch field {
		case consts.RemovableFieldAutoscaling:
			fields = append(fields, openapi.NIMSERVICEREMOVABLEFIELDS_AUTOSCALING)
		case consts.RemovableFieldReplicas:
			fields = append(fields, openapi.NIMSERVICEREMOVABLEFIELDS_REPLICAS)
		case consts.RemovableFieldEnvironmentVariables:
			fields = append(fields, openapi.NIMSERVICEREMOVABLEFIELDS_ENVIRONMENT_VARIABLES)
		case consts.RemovableFieldLabels:
			fields = append(fields, openapi.NIMSERVICEREMOVABLEFIELDS_LABELS)
		case consts.RemovableFieldAnnotations:
			fields = append(fields, openapi.NIMSERVICEREMOVABLEFIELDS_ANNOTATIONS)
		case consts.RemovableFieldTolerations:
			fields = append(fields, openapi.NIMSERVICEREMOVABLEFIELDS_TOLERATIONS)
		case consts.RemovableFieldProbes:
			fields = append(fields, openapi.NIMSERVICEREMOVABLEFIELDS_PROBES)
		case consts.RemovableFieldCompute:
			fields = append(fields, openapi.NIMSERVICEREMOVABLEFIELDS_COMPUTE)
		case consts.RemovableFieldMultiNode:
			fields = append(fields, openapi.NIMSERVICEREMOVABLEFIELDS_MULTI_NODE)
		case consts.RemovableFieldModelStore:
			fields = append(fields, openapi.NIMSERVICEREMOVABLEFIELDS_MODEL_STORE)
		}
	}
	return fields
}

func (bnc *BackendNimClient) Submit(ctx context.Context, name, projectId, clusterId string, generalSpec *spec.GeneralSpec) (string, error) {
	nimSpec, err := generalSpec.GenerateNimServiceSpec(ctx, clusterId)
	if err != nil {
		return "", err
	}

	request := bnc.apiClient.NVIDIANIMAPI.CreateNimService(ctx)
	metadata := openapi.NewNIMServiceMetadataCreateParams(name, projectId)
	metadata.SetUseGivenNameAsPrefix(generalSpec.HasPrefixName())
	nimServiceCreateReq := *openapi.NewNimServiceCreateRequest(*metadata, *nimSpec)

	if generalSpec.IsDryRun() {
		byteSpec, err := nimServiceCreateReq.MarshalJSON()
		return string(byteSpec), err
	}

	request = request.NimServiceCreateRequest(nimServiceCreateReq)
	nimService, response, err := request.Execute()
	if err != nil {
		return "", clierrors.HandleResponseError(response, err)
	}
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	return nimService.Metadata.Name, nil
}

func (bnc *BackendNimClient) Get(ctx context.Context, uuid string) (*openapi.NimServiceResponse, error) {
	nimService, response, err := bnc.apiClient.NVIDIANIMAPI.GetNimServiceById(ctx, uuid).Execute()
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	return nimService, nil
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
