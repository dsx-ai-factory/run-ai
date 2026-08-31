/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/run-ai/runai/runai-clients/policy-service-client/structs"

	openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
)

type PoliciesClient struct {
	ApiClient *openapi.APIClient
}

func (c *PoliciesClient) UpdateDistributedPolicy(ctx context.Context, request openapi.DistributedPolicyChangeRequestV2) (*openapi.DistributedPolicyV2, *http.Response, error) {
	return c.ApiClient.PolicyAPI.UpdateDistributedPolicyV2(ctx).DistributedPolicyChangeRequestV2(request).Execute()
}

func (c *PoliciesClient) UpdateInferencePolicy(ctx context.Context, request openapi.InferencePolicyChangeRequestV2) (*openapi.InferencePolicyV2, *http.Response, error) {
	return c.ApiClient.PolicyAPI.UpdateInferencePolicyV2(ctx).InferencePolicyChangeRequestV2(request).Execute()
}

func (c *PoliciesClient) UpdateTrainingPolicy(ctx context.Context, request openapi.TrainingPolicyChangeRequestV2) (*openapi.TrainingPolicyV2, *http.Response, error) {
	return c.ApiClient.PolicyAPI.UpdateTrainingPolicyV2(ctx).TrainingPolicyChangeRequestV2(request).Execute()
}

func (c *PoliciesClient) UpdateWorkspacePolicy(ctx context.Context, request openapi.WorkspacePolicyChangeRequestV2) (*openapi.WorkspacePolicyV2, *http.Response, error) {
	return c.ApiClient.PolicyAPI.UpdateWorkspacePolicyV2(ctx).WorkspacePolicyChangeRequestV2(request).Execute()
}

func (c *PoliciesClient) UpdateSyncStatusWorkspace(ctx context.Context, syncStatus openapi.SyncStatusUpdate) (*http.Response, error) {
	return c.ApiClient.SynchronizationAPI.UpdateSyncStatusWorkspace(ctx).SyncStatusUpdate(syncStatus).Execute()
}

func (c *PoliciesClient) UpdateSyncStatusTraining(ctx context.Context, syncStatus openapi.SyncStatusUpdate) (*http.Response, error) {
	return c.ApiClient.SynchronizationAPI.UpdateSyncStatusTraining(ctx).SyncStatusUpdate(syncStatus).Execute()
}

func (c *PoliciesClient) UpdateSyncStatusDistributed(ctx context.Context, syncStatus openapi.SyncStatusUpdate) (*http.Response, error) {
	return c.ApiClient.SynchronizationAPI.UpdateSyncStatusDistributed(ctx).SyncStatusUpdate(syncStatus).Execute()
}

func (c *PoliciesClient) UpdateSyncStatusInference(ctx context.Context, syncStatus openapi.SyncStatusUpdate) (*http.Response, error) {
	return c.ApiClient.SynchronizationAPI.UpdateSyncStatusInference(ctx).SyncStatusUpdate(syncStatus).Execute()
}

func (c *PoliciesClient) SyncWorkspacePolicies(ctx context.Context, clusterId string, after int64, syncReport *openapi.ClusterSyncReport) (*openapi.PolicySyncResponse, *http.Response, error) {
	request := c.ApiClient.SynchronizationAPI.SyncRequestWorkspacePolicies(ctx)
	request = request.ClusterId(clusterId)
	request = request.After(after)
	if syncReport != nil {
		request = request.ClusterSyncReport(*syncReport)
	}
	return request.Execute()
}

func (c *PoliciesClient) SyncTrainingPolicies(ctx context.Context, clusterId string, after int64, syncReport *openapi.ClusterSyncReport) (*openapi.PolicySyncResponse, *http.Response, error) {
	request := c.ApiClient.SynchronizationAPI.SyncRequestTrainingPolicies(ctx)
	request = request.ClusterId(clusterId)
	request = request.After(after)
	if syncReport != nil {
		request = request.ClusterSyncReport(*syncReport)
	}
	return request.Execute()
}

func (c *PoliciesClient) SyncDistributedPolicies(ctx context.Context, clusterId string, after int64, syncReport *openapi.ClusterSyncReport) (*openapi.PolicySyncResponse, *http.Response, error) {
	request := c.ApiClient.SynchronizationAPI.SyncRequestDistributedPolicies(ctx)
	request = request.ClusterId(clusterId)
	request = request.After(after)
	if syncReport != nil {
		request = request.ClusterSyncReport(*syncReport)
	}
	return request.Execute()
}

func (c *PoliciesClient) SyncInferencePolicies(ctx context.Context, clusterId string, after int64, syncReport *openapi.ClusterSyncReport) (*openapi.PolicySyncResponse, *http.Response, error) {
	request := c.ApiClient.SynchronizationAPI.SyncRequestInferencePolicies(ctx)
	request = request.ClusterId(clusterId)
	request = request.After(after)
	if syncReport != nil {
		request = request.ClusterSyncReport(*syncReport)
	}
	return request.Execute()
}

func (c *PoliciesClient) GetWorkspacePolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.WorkspacePolicyV2, *http.Response, error) {
	request := c.ApiClient.PolicyAPI.GetWorkspacePolicyV2(ctx).NoCreationMode(noCreationMode).Scope(scope)
	if departmentId != nil {
		request = request.DepartmentId(*departmentId)
	}
	if clusterId != nil {
		request = request.ClusterId(*clusterId)
	}
	if projectId != nil {
		request = request.ProjectId(*projectId)
	}
	return request.Execute()
}

func (c *PoliciesClient) GetNimServicePolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.NimServicePolicy, *http.Response, error) {
	request := c.ApiClient.PolicyAPI.GetNimServicePolicy(ctx).NoCreationMode(noCreationMode).Scope(scope)
	if departmentId != nil {
		request = request.DepartmentId(*departmentId)
	}
	if clusterId != nil {
		request = request.ClusterId(*clusterId)
	}
	if projectId != nil {
		request = request.ProjectId(*projectId)
	}
	return request.Execute()
}

func (c *PoliciesClient) GetTrainingPolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.TrainingPolicyV2, *http.Response, error) {
	request := c.ApiClient.PolicyAPI.GetTrainingPolicyV2(ctx).NoCreationMode(noCreationMode).Scope(scope)
	if departmentId != nil {
		request = request.DepartmentId(*departmentId)
	}
	if clusterId != nil {
		request = request.ClusterId(*clusterId)
	}
	if projectId != nil {
		request = request.ProjectId(*projectId)
	}
	return request.Execute()
}

func (c *PoliciesClient) GetDistributedPolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.DistributedPolicyV2, *http.Response, error) {
	request := c.ApiClient.PolicyAPI.GetDistributedPolicyV2(ctx).NoCreationMode(noCreationMode).Scope(scope)
	if departmentId != nil {
		request = request.DepartmentId(*departmentId)
	}
	if clusterId != nil {
		request = request.ClusterId(*clusterId)
	}
	if projectId != nil {
		request = request.ProjectId(*projectId)
	}
	return request.Execute()
}

func (c *PoliciesClient) GetInferencePolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.InferencePolicyV2, *http.Response, error) {
	request := c.ApiClient.PolicyAPI.GetInferencePolicyV2(ctx).NoCreationMode(noCreationMode).Scope(scope)
	if departmentId != nil {
		request = request.DepartmentId(*departmentId)
	}
	if clusterId != nil {
		request = request.ClusterId(*clusterId)
	}
	if projectId != nil {
		request = request.ProjectId(*projectId)
	}
	return request.Execute()
}

func (c *PoliciesClient) GetDistributedInferencePolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.DistributedInferencePolicyV2, *http.Response, error) {
	request := c.ApiClient.PolicyAPI.GetDistributedInferencePolicyV2(ctx).NoCreationMode(noCreationMode).Scope(scope)
	if departmentId != nil {
		request = request.DepartmentId(*departmentId)
	}
	if clusterId != nil {
		request = request.ClusterId(*clusterId)
	}
	if projectId != nil {
		request = request.ProjectId(*projectId)
	}
	return request.Execute()
}

func (c *PoliciesClient) GetPolicyCrossType(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.PolicyCrossType, *openapi.Error) {
	request := c.ApiClient.PolicyAPI.GetPolicyCrossType(ctx).NoCreationMode(noCreationMode).Scope(scope)
	if departmentId != nil {
		request = request.DepartmentId(*departmentId)
	}
	if clusterId != nil {
		request = request.ClusterId(*clusterId)
	}
	if projectId != nil {
		request = request.ProjectId(*projectId)
	}

	payload, httpResp, err := request.Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		if httpResp.StatusCode != http.StatusOK {
			// Read the body
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read response payload: %v", err.Error()))
			}

			var apiError openapi.Error

			jsonErr := json.Unmarshal(body, &apiError)
			if jsonErr != nil {
				return nil, openapi.NewError(int32(httpResp.StatusCode), string(body))
			} else {
				return nil, &apiError
			}
		}

		if payload != nil {
			return payload, nil
		}
	}

	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("%v", err))
}

func (c *PoliciesClient) GetAssetChangedHook(ctx context.Context, assetId string, notification *openapi.AssetChangeNotification) (*http.Response, error) {
	return c.ApiClient.PolicyAPI.AssetChangedHook(ctx, assetId).AssetChangeNotification(*notification).Execute()
}

func (c *PoliciesClient) GetWorkloadConfigOfTenant(ctx context.Context) (*map[string]*structs.PolicyWorkloadTypeConfig, *openapi.Error) {
	resp, err := c.GetPolicyCrossType(ctx, string(openapi.SCOPETYPE_TENANT), nil, nil, nil, true /*noCreation*/)
	if err != nil {
		return nil, err
	}

	result := map[string]*structs.PolicyWorkloadTypeConfig{}

	if resp.Workspace.IsSet() && resp.Workspace.Get().Effective.IsSet() && resp.Workspace.Get().Effective.Get().Defaults.IsSet() {
		defaults := resp.Workspace.Get().Effective.Get().Defaults.Get()
		config := &structs.PolicyWorkloadTypeConfig{
			PriorityName:   defaults.PriorityClass.Get(),
			CategoryName:   defaults.Category.Get(),
			Preemptibility: defaults.Preemptibility.Get(),
		}
		result["InteractiveWorkload"] = config
	}

	if resp.Training.IsSet() && resp.Training.Get().Effective.IsSet() && resp.Training.Get().Effective.Get().Defaults.IsSet() {
		defaults := resp.Training.Get().Effective.Get().Defaults.Get()
		config := &structs.PolicyWorkloadTypeConfig{
			PriorityName:   defaults.PriorityClass.Get(),
			CategoryName:   defaults.Category.Get(),
			Preemptibility: defaults.Preemptibility.Get(),
		}
		result["TrainingWorkload"] = config
	}

	if resp.Inference.IsSet() && resp.Inference.Get().Effective.IsSet() && resp.Inference.Get().Effective.Get().Defaults.IsSet() {
		defaults := resp.Inference.Get().Effective.Get().Defaults.Get()
		config := &structs.PolicyWorkloadTypeConfig{
			PriorityName:   defaults.PriorityClass.Get(),
			CategoryName:   defaults.Category.Get(),
			Preemptibility: defaults.Preemptibility.Get(),
		}
		result["InferenceWorkload"] = config
	}

	if resp.Distributed.IsSet() && resp.Distributed.Get().Effective.IsSet() && resp.Distributed.Get().Effective.Get().Defaults.IsSet() {
		defaults := resp.Distributed.Get().Effective.Get().Defaults.Get().GetMaster()
		config := &structs.PolicyWorkloadTypeConfig{
			PriorityName:   defaults.PriorityClass.Get(),
			CategoryName:   defaults.Category.Get(),
			Preemptibility: defaults.Preemptibility.Get(),
		}
		result["DistributedWorkload"] = config
	}

	return &result, nil
}

func (c *PoliciesClient) GetGvkPolicyOfProject(ctx context.Context, projectId string, workloadType openapi.GVK) (*openapi.GetGvkPolicyResp, *openapi.Error) {
	request := c.ApiClient.PolicyAPI.GetGvkPolicyByTypeAndScope(ctx).
		WorkloadTypeAndScope(openapi.WorkloadTypeAndScope{
			WorkloadType: workloadType,
			Scope:        openapi.SCOPETYPE_PROJECT,
			ProjectId:    *openapi.NewNullableString(&projectId),
		})

	payload, httpResp, err := request.Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		if httpResp.StatusCode != http.StatusOK {
			// Read the body
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read response payload: %v", err.Error()))
			}

			var apiError openapi.Error

			jsonErr := json.Unmarshal(body, &apiError)
			if jsonErr != nil {
				return nil, openapi.NewError(int32(httpResp.StatusCode), string(body))
			} else {
				return nil, &apiError
			}
		}

		if payload != nil {
			return payload, nil
		}
	}

	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("%v", err))

}
