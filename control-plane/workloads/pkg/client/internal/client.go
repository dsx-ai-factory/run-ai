/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

type WorkloadsExternalClient struct {
	WorkloadsClient
}

type WorkloadsClient struct {
	ApiClient *openapi.APIClient
}

func (wc *WorkloadsClient) GetInferenceAncestors(ctx context.Context, resourceId string) (*openapi.Ancestors, error) {
	request := wc.ApiClient.InferencesAPI.GetInferenceAncestors(ctx, resourceId)

	ancestors, _, err := wc.ApiClient.InferencesAPI.GetInferenceAncestorsExecute(request)
	if err != nil {
		return nil, err
	}

	return ancestors, nil
}

func (wc *WorkloadsClient) GetAncestors(ctx context.Context, resourceId string) (*openapi.Ancestors, error) {
	request := wc.ApiClient.WorkloadsAPI.GetWorkloadAncestors(ctx, resourceId)

	ancestors, _, err := wc.ApiClient.WorkloadsAPI.GetWorkloadAncestorsExecute(request)
	if err != nil {
		return nil, err
	}

	return ancestors, nil
}

// ListWorkloadTypes retrieves all workload types from the workloads service
func (wc *WorkloadsClient) ListWorkloadTypes(ctx context.Context, filters []string) (*openapi.ListWorkloadTypes200Response, *openapi.Error) {
	request := wc.ApiClient.WorkloadPropertiesAPI.
		ListWorkloadTypes(ctx)

	if len(filters) > 0 {
		request = request.FilterBy(filters)
	}

	response, httpResponse, err := request.Execute()
	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusOK {
			return response, nil
		}

		data, _ := io.ReadAll(httpResponse.Body)
		formattedErr := fmt.Sprintf("failed to list workload types: [%d] %s", httpResponse.StatusCode, data)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return nil, openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	if err == nil {
		err = fmt.Errorf("could not get response from list workload types API")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", err.Error())
	return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
}

func (wc *WorkloadsClient) GetAvailablePrioritiesAndCategories(ctx context.Context, clusterId *string,
	sortOrder *string, sortBy *string, filters *[]string,
) (*openapi.GetAvailablePrioritiesAndCategories200Response, *openapi.Error) {
	request := wc.ApiClient.WorkloadPropertiesAPI.
		GetAvailablePrioritiesAndCategories(ctx)

	if clusterId != nil {
		request = request.ClusterId(*clusterId)
	}

	if sortOrder != nil {
		request = request.SortOrder(*sortOrder)
	}

	if sortBy != nil {
		request = request.SortBy(*sortBy)
	}

	if filters != nil && len(*filters) > 0 {
		request = request.FilterBy(*filters)
	}

	response, httpResponse, err := request.Execute()
	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusOK {
			return response, nil
		}

		data, _ := io.ReadAll(httpResponse.Body)
		formattedErr := fmt.Sprintf("failed to list available priorities and categories: [%d] %s", httpResponse.StatusCode, data)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return nil, openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	if err == nil {
		err = fmt.Errorf("could not get response from get available priorities and categories")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", err.Error())
	return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
}

func (wc *WorkloadsClient) GetWorkloads(ctx context.Context,
	limit, offset int32, filter []string, includeDeleted bool, sortBy string, sortOrder string) (*openapi.GetWorkloads200Response, error) {
	request := wc.ApiClient.WorkloadsAPI.GetWorkloads(ctx)
	request = request.Limit(limit)
	request = request.Offset(offset)
	request = request.Deleted(includeDeleted)

	if len(filter) != 0 {
		request = request.FilterBy(filter)
	}

	if len(sortBy) != 0 {
		request = request.SortBy(sortBy)
	}

	if len(sortOrder) != 0 {
		request = request.SortOrder(sortOrder)
	}

	workloadList, _, err := request.Execute()
	if err != nil {
		return nil, err
	}

	return workloadList, nil
}

func (wc *WorkloadsClient) GetWorkload(ctx context.Context, workloadId string) (*openapi.WorkloadDetailed, *openapi.Error) {
	request := wc.ApiClient.WorkloadsAPI.GetWorkload(ctx, workloadId)

	response, httpResponse, err := request.Execute()
	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusOK {
			return response, nil
		}

		// we need the error as Json, not as []byte
		data, _ := io.ReadAll(httpResponse.Body)

		formattedErr := fmt.Sprintf("failed to get workload %s: [%d] %s", workloadId, httpResponse.StatusCode, data)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return nil, openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to request workloads service to get workload %s: %v", workloadId, err)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return nil, openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}

func (wc *WorkloadsClient) CreateWorkload(ctx context.Context, workload openapi.WorkloadCreationFields) (*openapi.Workload, *openapi.Error) {
	request := wc.ApiClient.WorkloadsAPI.InternalCreateWorkload(ctx)

	response, httpResponse, err := request.WorkloadCreationFields(workload).Execute()
	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusCreated {
			return response, nil
		}

		// we need the error as Json, not as []byte
		data, _ := io.ReadAll(httpResponse.Body)

		formattedErr := fmt.Sprintf("failed to notify workloads service about the creation of workload %s: [%d] %s", workload.Id, httpResponse.StatusCode, data)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return nil, openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to notify workloads service about the creation of workload %s: %v", workload.Id, err.Error())

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return nil, openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}

// CreateWorkloadActions stores accepted actions forwarded by workloads-manager.
func (wc *WorkloadsClient) CreateWorkloadActions(ctx context.Context, actions openapi.CreateWorkloadActionsRequest) *openapi.Error {
	request := wc.ApiClient.WorkloadActionsAPI.InternalCreateWorkloadActions(ctx).CreateWorkloadActionsRequest(actions)
	httpResponse, err := request.Execute()
	if httpResponse != nil {
		defer httpResponse.Body.Close()
		if httpResponse.StatusCode == http.StatusNoContent {
			return nil
		}
		data, _ := io.ReadAll(httpResponse.Body)
		message := fmt.Sprintf("failed to request workloads service to create workload actions: [%d] %s", httpResponse.StatusCode, data)
		return openapi.NewError(int32(httpResponse.StatusCode), message)
	}
	if err == nil {
		err = fmt.Errorf("no HTTP response received")
	}
	return openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request workloads service to create workload actions: %v", err))
}

func (wc *WorkloadsClient) PatchWorkload(ctx context.Context, workloadId string, fields openapi.WorkloadPatchFields) *openapi.Error {
	request := wc.ApiClient.WorkloadsAPI.InternalPatchWorkloadPhase(ctx, workloadId)

	httpResponse, err := request.WorkloadPatchFields(fields).Execute()
	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusCreated {
			return nil
		}

		// we need the error as Json, not as []byte
		data, _ := io.ReadAll(httpResponse.Body)

		formattedErr := fmt.Sprintf("failed to patch workload %s: [%d] %s", workloadId, httpResponse.StatusCode, data)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to request workloads service to patch workload %s: %v", workloadId, err.Error())

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}

func (wc *WorkloadsClient) ProcessSyncReport(ctx context.Context, clusterId string, report openapi.ClusterSyncReport) *openapi.Error {
	httpResponse, err := wc.ApiClient.SyncAPI.ReportSyncIssues(ctx).ClusterId(clusterId).ClusterSyncReport(report).Execute()
	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusNoContent {
			return nil
		}

		// we need the error as Json, not as []byte
		data, _ := io.ReadAll(httpResponse.Body)

		formattedErr := fmt.Sprintf("failed to process sync report of %s: [%d] %s", clusterId, httpResponse.StatusCode, data)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to request workloads service to process sync report of %s: %v", clusterId, err)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}

func (wc *WorkloadsClient) CountWorkloads(ctx context.Context, filter []string, search string, includeDeleted bool, precise bool) (int64, *openapi.Error) {
	response, httpResponse, err := wc.ApiClient.WorkloadsAPI.CountWorkloads(ctx).
		FilterBy(filter).
		Search(search).
		Precise(precise).
		Deleted(includeDeleted).Execute()
	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusOK && response != nil {
			return response.Count, nil
		}

		// we need the error as Json, not as []byte
		data, _ := io.ReadAll(httpResponse.Body)

		formattedErr := fmt.Sprintf("failed to count workloads: [%d] %s", httpResponse.StatusCode, data)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return 0, openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to request workloads service to count workloads: %v", err)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return 0, openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}

func (wc *WorkloadsClient) InternalEnrichOrgUnitRanking(ctx context.Context, request openapi.InternalEnrichOrgUnitRankingRequest) (*openapi.InternalEnrichOrgUnitRankingResponse, error) {
	apiRequest := wc.ApiClient.WorkloadsAPI.InternalEnrichOrgUnitRanking(ctx).InternalEnrichOrgUnitRankingRequest(request)

	response, httpResponse, err := apiRequest.Execute()
	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusOK {
			return response, nil
		}

		data, _ := io.ReadAll(httpResponse.Body)
		formattedErr := fmt.Errorf("failed to enrich org unit ranking: [%d] %s", httpResponse.StatusCode, data)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr.Error())
		return nil, formattedErr
	}

	if err == nil {
		err = fmt.Errorf("could not get response from enrich org unit ranking API")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", err.Error())
	return nil, err
}

func (wc *WorkloadsClient) InternalUpdateNodePods(ctx context.Context, clusterId string, nodeName string, networkTopologyLabelValues []string) *openapi.Error {
	request := wc.ApiClient.PodsAPI.InternalUpdateNodePods(ctx, clusterId, nodeName).
		UpdateNodePodsRequest(openapi.UpdateNodePodsRequest{
			NetworkTopologyLabelValues: networkTopologyLabelValues,
		})

	httpResponse, err := request.Execute()
	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusNoContent || httpResponse.StatusCode == http.StatusOK {
			return nil
		}

		data, _ := io.ReadAll(httpResponse.Body)
		formattedErr := fmt.Sprintf("failed to update node pods for cluster %s, node %s: [%d] %s",
			clusterId, nodeName, httpResponse.StatusCode, data)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to request workloads service to update node pods for cluster %s, node %s: %v",
		clusterId, nodeName, err)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}
