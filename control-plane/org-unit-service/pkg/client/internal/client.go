/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/structs"
	clientutils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	"k8s.io/utils/ptr"
)

const (
	clusterIdField = "clusterId"
	parentIdField  = "parentId"
)

type OrgUnitClient struct {
	ApiClient *openapi.APIClient
}

func (client *OrgUnitClient) GetProjectsSync(ctx context.Context, clusterId string, limit int32, offset int32, UpdatedAfter time.Time, excludePermissions bool) (*openapi.GetSyncProjects200Response, error) {
	response, h, err := client.ApiClient.ProjectsAPI.GetSyncProjects(ctx).
		ClusterId(clusterId).
		Limit(limit).
		Offset(offset).
		UpdatedAfter(UpdatedAfter).
		ExcludePermissions(excludePermissions).Execute()

	if h != nil && h.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Failed to get sync projects. response %v", h))
	}
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (client *OrgUnitClient) UpdateProjectsStatus(ctx context.Context, projectStatusUpdateRequest openapi.ProjectStatusUpdateRequest) (*http.Response, error) {
	return client.ApiClient.ProjectsAPI.UpdateProjectStatus(ctx).ProjectStatusUpdateRequest(projectStatusUpdateRequest).Execute()
}

func (client *OrgUnitClient) SyncClusterProjects(ctx context.Context, clusterId string, syncProjects []openapi.ProjectInSyncRequest) (*http.Response, error) {
	return client.ApiClient.ProjectsAPI.SyncProjects(ctx).SyncProjectsRequest(openapi.SyncProjectsRequest{
		ClusterId: clusterId,
		Projects:  syncProjects,
	}).Execute()
}

func (client *OrgUnitClient) GetDepartmentsSync(ctx context.Context, clusterId string, limit int32, offset int32, UpdatedAfter time.Time) (*openapi.GetSyncDepartments200Response, error) {
	response, h, err := client.ApiClient.DepartmentsAPI.GetSyncDepartments(ctx).
		ClusterId(clusterId).
		Limit(limit).
		Offset(offset).
		UpdatedAfter(UpdatedAfter).Execute()

	if err != nil {
		return nil, err
	}
	if h != nil && h.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Failed to get sync projects. response %v", h))
	}
	return response, nil
}

func (client *OrgUnitClient) UpdateDepartmentsStatus(ctx context.Context, departmentId string, departmentStatusUpdateRequest openapi.DepartmentStatusUpdateRequest) (*http.Response, error) {
	return client.ApiClient.DepartmentsAPI.UpdateDepartmentStatus(ctx, departmentId).DepartmentStatusUpdateRequest(departmentStatusUpdateRequest).Execute()
}

func (client *OrgUnitClient) NodePoolCreated(ctx context.Context, clusterId string, nodePoolId string, nodePoolName string) error {
	_, err := client.ApiClient.NotificationsAPI.NodePoolUpdated(ctx).Default_(false).NodePool(openapi.NodePool{
		Id:        &nodePoolId,
		Name:      &nodePoolName,
		ClusterId: clusterId,
		Operation: "CREATED",
	}).Execute()
	return err
}

// UpdateOrgUnitSettings calls org-unit-service's internal settings endpoint. Setting
// "project.quota.cpu_resources" to "false" triggers org-unit-service to wipe CPU/memory quota
// values across all projects and departments (dao.DisableCpuQuota).
func (client *OrgUnitClient) UpdateOrgUnitSettings(ctx context.Context, name string, value string) error {
	resp, err := client.ApiClient.OrgUnitAPI.UpdateOrgUnitSettings(ctx).
		OrgUnitSingleSetting(openapi.OrgUnitSingleSetting{Name: &name, Value: &value}).Execute()
	if err != nil {
		return fmt.Errorf("UpdateOrgUnitSettings: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateOrgUnitSettings returned status %d", resp.StatusCode)
	}
	return nil
}

func (client *OrgUnitClient) NodePoolSummary(ctx context.Context, nodePoolId string, summaryType openapi.OrgUnitsNodepoolSummaryType) ([]openapi.OrgUnitsNodepoolSummaryInner, error) {
	response, resp, err := client.ApiClient.OrgUnitAPI.OrgsNodePoolSummary(ctx, nodePoolId).
		NodepoolSummaryType(summaryType).Execute()
	if err != nil {
		return nil, fmt.Errorf("NodePoolSummary: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("NodePoolSummary returned status %d", resp.StatusCode)
	}
	if response == nil {
		return nil, errors.New("NodePoolSummary returned an empty response")
	}
	return response.OrgUnitsNodepoolSummary, nil
}

func (client *OrgUnitClient) NodePoolResetResources(ctx context.Context, clusterId string, nodePoolId string, action openapi.NodepoolResourcesResetAction) error {
	resp, err := client.ApiClient.OrgUnitAPI.NodePoolResetResources(ctx).
		OrgUnitsNodepoolResourcesResetRequest(openapi.OrgUnitsNodepoolResourcesResetRequest{
			ClusterId:  clusterId,
			NodePoolId: nodePoolId,
			Action:     action,
		}).Execute()
	if err != nil {
		return fmt.Errorf("NodePoolResetResources: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("NodePoolResetResources returned status %d", resp.StatusCode)
	}
	return nil
}

func (client *OrgUnitClient) NodePoolDeleted(ctx context.Context, clusterId string, nodePoolId string, nodePoolName string) error {
	_, err := client.ApiClient.NotificationsAPI.NodePoolUpdated(ctx).Default_(false).NodePool(openapi.NodePool{
		Id:        &nodePoolId,
		Name:      &nodePoolName,
		ClusterId: clusterId,
		Operation: "DELETED",
	}).Execute()
	return err
}

func (client *OrgUnitClient) DefaultNodePoolCreated(ctx context.Context, clusterId string, nodePoolId string) error {
	_, err := client.ApiClient.NotificationsAPI.NodePoolUpdated(ctx).Default_(true).NodePool(openapi.NodePool{
		Id:        &nodePoolId,
		ClusterId: clusterId,
		Name:      ptr.To("default"),
		Operation: "CREATED",
	}).Execute()
	return err
}

func (client *OrgUnitClient) ClusterDeleted(ctx context.Context, clusterId string) error {
	_, err := client.ApiClient.NotificationsAPI.ClusterDeleted(ctx, clusterId).Execute()
	return err
}

func (client *OrgUnitClient) GetProjectById(ctx context.Context, projectId string) (*openapi.Project, *openapi.Error) {
	project, resp, err := client.ApiClient.ProjectsAPI.GetProject(ctx, projectId).Execute()
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return nil, openapi.NewError(int32(resp.StatusCode), fmt.Sprintf("Failed to get project: %v", resp))
	}
	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	if project == nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, "Failed to get response of get project by ID")
	}

	return project, nil
}

func (client *OrgUnitClient) GetOrgUnits(ctx context.Context, scopeType *openapi.ScopeType, scopeId *string) ([]openapi.OrgUnit, *openapi.Error) {
	request := client.ApiClient.OrgUnitAPI.GetOrgUnitsNames(ctx)

	if scopeType != nil {
		request = request.OrgUnitType(string(*scopeType))
	}

	if scopeId != nil {
		request = request.OrgUnitIds(*scopeId)
	}

	orgUnits, resp, err := client.ApiClient.OrgUnitAPI.GetOrgUnitsNamesExecute(request)
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return nil, openapi.NewError(int32(resp.StatusCode), "Failed to get org units")
	}
	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	if orgUnits == nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, "failed to get response of get org units")
	}

	return orgUnits, nil
}

func (client *OrgUnitClient) GetOrgTreeMap(ctx context.Context) (structs.OrgUnitMap, map[structs.OrgUnitKey]string, *openapi.Error) {
	orgTree, resp, err := client.ApiClient.OrgUnitAPI.GetInternalOrgTree(ctx).Execute()
	if err != nil {
		return nil, nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, nil, openapi.NewError(int32(resp.StatusCode), "Failed to get org units")
	}
	if orgTree == nil {
		return nil, nil, openapi.NewError(http.StatusServiceUnavailable, "failed to get response of get org tree")
	}

	orgUnitMap := make(structs.OrgUnitMap)
	orgUnitNameMap := map[structs.OrgUnitKey]string{}
	for _, orgUnit := range orgTree.OrgUnits {
		orgUnitKey := structs.OrgUnitKey{Scope: string(orgUnit.Type), Id: orgUnit.Id}
		orgUnitNameMap[orgUnitKey] = orgUnit.Name

		if orgUnit.ParentId != nil && orgUnit.ParentType != nil {
			orgUnitMap[orgUnitKey] = structs.OrgUnitKey{Scope: string(*orgUnit.ParentType), Id: *orgUnit.ParentId}
		}
	}

	return orgUnitMap, orgUnitNameMap, nil
}

func (client *OrgUnitClient) GetProjectsMetadataByCluster(ctx context.Context, clusterId string) ([]openapi.ProjectMetadata, *openapi.Error) {
	req := client.ApiClient.ProjectsAPI.GetProjectsMetadata(ctx)
	req = req.FilterBy(clientutils.NewFilterBuilder().AddFilter(clusterIdField, clientutils.EqualOp, clusterId).Get())

	nextOffset := int32(0)

	projects := make([]openapi.ProjectMetadata, 0)

	for true {
		resp, hResp, err := req.Offset(nextOffset).Limit(500).Execute()
		if hResp != nil && hResp.StatusCode >= http.StatusBadRequest {
			return nil, openapi.NewError(int32(hResp.StatusCode), "Failed to get projects")
		}
		if err != nil {
			return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
		}
		if resp == nil {
			return nil, openapi.NewError(http.StatusServiceUnavailable, "Failed to obtain response of get projects")
		}

		if len(resp.Projects) == 0 {
			break
		}

		projects = append(projects, resp.Projects...)

		if resp.GetNext() == 0 {
			break
		}
		if resp.GetNext() <= nextOffset {
			return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("received next offset %d while current offset is %d", resp.GetNext(), nextOffset))
		}
		nextOffset = resp.GetNext()
	}

	return projects, nil
}

func (client *OrgUnitClient) GetProjectsMetadataByDepartment(ctx context.Context, departmentId string) ([]openapi.ProjectMetadata, *openapi.Error) {
	req := client.ApiClient.ProjectsAPI.GetProjectsMetadata(ctx)
	req = req.FilterBy(clientutils.NewFilterBuilder().AddFilter(parentIdField, clientutils.EqualOp, departmentId).Get())

	nextOffset := int32(0)

	projects := make([]openapi.ProjectMetadata, 0)

	for true {
		resp, hResp, err := req.Offset(nextOffset).Limit(500).Execute()
		if hResp != nil && hResp.StatusCode >= http.StatusBadRequest {
			return nil, openapi.NewError(int32(hResp.StatusCode), "Failed to get projects")
		}
		if err != nil {
			return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
		}
		if resp == nil {
			return nil, openapi.NewError(http.StatusServiceUnavailable, "Failed to obtain response of get projects")
		}

		if len(resp.Projects) == 0 {
			break
		}

		projects = append(projects, resp.Projects...)

		if resp.GetNext() == 0 {
			break
		}
		if resp.GetNext() <= nextOffset {
			return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("received next offset %d while current offset is %d", resp.GetNext(), nextOffset))
		}
		nextOffset = resp.GetNext()
	}

	return projects, nil
}

func (client *OrgUnitClient) GetDepartmentsDescendants(ctx context.Context, orgUnitIds string) (*openapi.Descendants, *openapi.Error) {
	response, h, err := client.ApiClient.DepartmentsAPI.GetDepartmentsDescendants(ctx).OrgUnitIds(orgUnitIds).Execute()
	if h != nil && h.StatusCode != http.StatusOK {
		return nil, openapi.NewError(int32(h.StatusCode), fmt.Sprintf("failed to get descendants: %v", response))
	}
	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	if response == nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, "failed to get response of get department descendants")
	}

	return response, nil
}

func (client *OrgUnitClient) GetProjectAncestors(ctx context.Context, projectId string) (*openapi.Ancestors, *openapi.Error) {
	response, h, err := client.ApiClient.ProjectsAPI.GetProjectAncestors(ctx, projectId).Execute()
	if h != nil && h.StatusCode != http.StatusOK {
		return nil, openapi.NewError(int32(h.StatusCode), fmt.Sprintf("failed to get ancestors: %v", response))
	}
	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	if response == nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, "failed to get response of get project ancestors")
	}

	return response, nil
}

func (client *OrgUnitClient) GetDepartmentAncestors(ctx context.Context, projectId string) (*openapi.Ancestors, *openapi.Error) {
	response, h, err := client.ApiClient.DepartmentsAPI.GetDepartmentAncestors(ctx, projectId).Execute()
	if h != nil && h.StatusCode != http.StatusOK {
		return nil, openapi.NewError(int32(h.StatusCode), fmt.Sprintf("failed to get ancestors: %v", response))
	}
	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	if response == nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, "failed to get response of get department ancestors")
	}

	return response, nil
}

func (client *OrgUnitClient) GetAncestorsOfScope(ctx context.Context, scopeType openapi.ScopeType, scopeId string) (*openapi.Ancestors, *openapi.Error) {
	result := openapi.Ancestors{}

	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantId != nil {
		result.TenantId = *tenantId
	}

	switch scopeType {
	case openapi.SCOPETYPE_PROJECT:
		return client.GetProjectAncestors(ctx, scopeId)
	case openapi.SCOPETYPE_DEPARTMENT:
		return client.GetDepartmentAncestors(ctx, scopeId)
	case openapi.SCOPETYPE_CLUSTER:
		result.ClusterUuid = *openapi.NewNullableString(&scopeId)
		return &result, nil
	case openapi.SCOPETYPE_TENANT:
		return &result, nil
	case openapi.SCOPETYPE_SYSTEM:
		return &openapi.Ancestors{}, nil
	default:
		return nil, openapi.NewError(http.StatusBadRequest, fmt.Sprintf("scope %s is not supported", scopeType))
	}
}

func (client *OrgUnitClient) CountClusterDepartments(ctx context.Context, clusterId string) (int64, *openapi.Error) {
	response, h, err := client.ApiClient.DepartmentsAPI.CountDepartments(ctx).
		FilterBy([]string{fmt.Sprintf("clusterId==%s", clusterId)}).Execute()

	if err != nil {
		return 0, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	if h != nil && h.StatusCode != http.StatusOK {
		return 0, openapi.NewError(int32(h.StatusCode), fmt.Sprintf("failed to get departments: %v", response))
	}
	if response == nil {
		return 0, openapi.NewError(http.StatusServiceUnavailable, "failed to get response of get departments")
	}
	return response.Count, nil
}

func (client *OrgUnitClient) GetOrgTree(ctx context.Context) ([]openapi.OrgUnit, *openapi.Error) {
	orgTree, resp, err := client.ApiClient.OrgUnitAPI.GetOrgTree(ctx).Execute()
	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, openapi.NewError(int32(resp.StatusCode), "Failed to get org tree")
	}
	if orgTree == nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, "failed to get response of get org tree")
	}

	return orgTree.OrgUnits, nil
}

func (client *OrgUnitClient) GetInternalDeletedOrgUnits(ctx context.Context, deletedAfter time.Time) ([]openapi.DeletedOrgUnit, *openapi.Error) {
	response, h, err := client.ApiClient.OrgUnitAPI.GetInternalDeletedOrgUnits(ctx).DeletedAfter(deletedAfter).Execute()
	if err != nil {
		return []openapi.DeletedOrgUnit{}, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	if h != nil && h.StatusCode != http.StatusOK {
		return []openapi.DeletedOrgUnit{}, openapi.NewError(int32(h.StatusCode), fmt.Sprintf("failed to get deleted org units: %v", response))
	}
	if response == nil {
		return []openapi.DeletedOrgUnit{}, openapi.NewError(http.StatusServiceUnavailable, "failed to get response of get deleted org units")
	}
	return response.OrgUnits, nil
}
