/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"

	org_unit_service_swagger "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/structs"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	org_unit_service_client "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
	tenants_api "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
	datavolumes_client "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client"
	workloads_client "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type OrgHierarchyManager struct {
	WorkloadsClient      workloads_client.IWorkloadsClient
	ClustersClient       *cluster_service_client.APIClient
	OrgUnitServiceClient org_unit_service_client.IOrgUnitClient
	DatavolumesClient    datavolumes_client.IDatavolumesClient
	TenantsClient        tenants_api.ITenantsManager
}

var once sync.Once

var mgr *OrgHierarchyManager

func newOrgHierarchyManager() *OrgHierarchyManager {
	logging.Debugf("initializing OrgHierarchyManager")
	if err := cache.RegisterSharedStore[authorization_utils.Ancestors](authorization_utils.Field, options.MaxRecordSizeKB(1), options.MaxStoreSizeMB(200)); err != nil {
		logging.Errorf("Failed to register shared store for ancestors: %v", err)
	}
	return &OrgHierarchyManager{
		WorkloadsClient:      workloads_client.NewWorkloadsClient(config.Get[*authorization_config.Server]().WorkloadsUrl),
		OrgUnitServiceClient: org_unit_service_client.NewOrgUnitClient(config.Get[*authorization_config.Server]().OrgUnitServiceUrl),
		ClustersClient: cluster_service_client.NewAPIClient(&cluster_service_client.Configuration{
			Servers: cluster_service_client.ServerConfigurations{
				{
					URL: config.Get[*authorization_config.Server]().ClusterServiceUrl,
				},
			},
		}),
		DatavolumesClient: datavolumes_client.NewDatavolumesClient(config.Get[*authorization_config.Server]().DatavolumesServiceUrl),
		TenantsClient:     tenants_api.NewTenantsManagerClient(config.Get[*authorization_config.Server]().TenantServiceUrl),
	}
}

func GetOrgHierarchyManager() *OrgHierarchyManager {
	once.Do(func() {
		mgr = newOrgHierarchyManager()
	})

	return mgr
}

func (mgr *OrgHierarchyManager) GetAncestors(ctx context.Context, resourceId string, resourceType api.ResourceType) (ancestors *authorization_utils.Ancestors, cacheHit bool, err error) {
	baseResourceType := authorization_utils.ConvertSubResource(string(resourceType))
	ancestors, err = authorization_utils.GetAncestorsCache(ctx, baseResourceType, resourceId)
	if err == nil && ancestors != nil {
		return ancestors, true, nil
	}

	ancestors, err = mgr.getAncestorsSync(ctx, resourceId, api.ResourceType(baseResourceType))
	if err != nil {
		if strings.HasPrefix(err.Error(), "404") {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Got an error trying to fetch the ancestor list of %s %s from server. err: %v", string(resourceType), resourceId, err)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error trying to fetch the ancestor list of %s %s from server. err: %v", string(resourceType), resourceId, err)
		}
		return nil, false, err
	}

	return ancestors, false, nil
}

func (mgr *OrgHierarchyManager) getAncestorsSync(ctx context.Context, resourceId string, resourceType api.ResourceType) (*authorization_utils.Ancestors, error) {
	switch resourceType {
	case api.RESOURCETYPE_TENANT:
		return &authorization_utils.Ancestors{TenantId: resourceId}, nil
	case api.RESOURCETYPE_WORKLOADS, api.RESOURCETYPE_TRAININGS, api.RESOURCETYPE_WORKSPACES:
		return mgr.getWorkloadsAncestorsSync(ctx, resourceId)
	case api.RESOURCETYPE_INFERENCES:
		return mgr.getInferenceAncestorsSync(ctx, resourceId)
	case api.RESOURCETYPE_CLUSTER, api.RESOURCETYPE_CLUSTERS_MINIMAL:
		return mgr.getClustersAncestorsSync(ctx, resourceId)
	case api.RESOURCETYPE_CLUSTER_CONFIG:
		return mgr.getClusterConfigAncestorsSync(ctx, resourceId)
	case api.RESOURCETYPE_PROJECT, api.RESOURCETYPE_DEPARTMENT:
		return mgr.getOrgUnitAncestorsSync(ctx, resourceId, resourceType)
	case api.RESOURCETYPE_DATAVOLUMES:
		return mgr.getDatavolumeAncestorsSync(ctx, resourceId)
	case api.RESOURCETYPE_NODEPOOLS, api.RESOURCETYPE_NODEPOOLS_MINIMAL:
		return mgr.getNodepoolsAncestorsSync(ctx, resourceId)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func (mgr *OrgHierarchyManager) getOrgUnitAncestorsSync(ctx context.Context, resourceId string, resourceType api.ResourceType) (*authorization_utils.Ancestors, error) {
	ancestors, err := mgr.OrgUnitServiceClient.GetAncestorsOfScope(ctx, org_unit_service_swagger.ScopeType(resourceType), resourceId)
	if err != nil {
		return nil, errors.New(err.Message)
	}

	return &authorization_utils.Ancestors{
		TenantId:     ancestors.TenantId,
		ClusterUuid:  ancestors.ClusterUuid.Get(),
		DepartmentId: ancestors.DepartmentId.Get(),
		ProjectId:    ancestors.ProjectId.Get(),
	}, nil
}

func (mgr *OrgHierarchyManager) getWorkloadsAncestorsSync(ctx context.Context, resourceId string) (*authorization_utils.Ancestors, error) {
	ancestors, err := mgr.WorkloadsClient.GetAncestors(ctx, resourceId)
	if err != nil {
		return nil, err
	}

	return &authorization_utils.Ancestors{
		TenantId:     ancestors.TenantId,
		ClusterUuid:  ancestors.ClusterUuid.Get(),
		DepartmentId: ancestors.DepartmentId.Get(),
		ProjectId:    ancestors.ProjectId.Get(),
	}, nil
}

func (mgr *OrgHierarchyManager) getInferenceAncestorsSync(ctx context.Context, resourceId string) (*authorization_utils.Ancestors, error) {
	ancestors, err := mgr.WorkloadsClient.GetInferenceAncestors(ctx, resourceId)
	if err != nil {
		return nil, err
	}

	return &authorization_utils.Ancestors{
		TenantId:     ancestors.TenantId,
		ClusterUuid:  ancestors.ClusterUuid.Get(),
		DepartmentId: ancestors.DepartmentId.Get(),
		ProjectId:    ancestors.ProjectId.Get(),
	}, nil
}

func (mgr *OrgHierarchyManager) getDatavolumeAncestorsSync(ctx context.Context, resourceId string) (*authorization_utils.Ancestors, error) {
	ancestors, err := mgr.DatavolumesClient.GetDatavolumeAncestorsSync(ctx, resourceId)
	if err != nil {
		return nil, err
	}

	return &authorization_utils.Ancestors{
		TenantId:     ancestors.TenantId,
		ClusterUuid:  ancestors.ClusterUuid.Get(),
		DepartmentId: ancestors.DepartmentId.Get(),
		ProjectId:    ancestors.ProjectId.Get(),
	}, nil
}

func (mgr *OrgHierarchyManager) getClustersAncestorsSync(ctx context.Context, resourceId string) (*authorization_utils.Ancestors, error) {
	request := mgr.ClustersClient.ClustersAPI.GetClusterAncestors(ctx, resourceId)

	ancestors, _, err := request.Execute()
	if err != nil {
		return nil, err
	}

	return &authorization_utils.Ancestors{
		TenantId: ancestors.TenantId,
	}, nil
}

func (mgr *OrgHierarchyManager) getClusterConfigAncestorsSync(ctx context.Context, resourceId string) (*authorization_utils.Ancestors, error) {
	ancestors, err := mgr.getClustersAncestorsSync(ctx, resourceId)
	if err != nil {
		return nil, err
	}
	ancestors.ClusterUuid = &resourceId
	return ancestors, nil
}

func (mgr *OrgHierarchyManager) getNodepoolsAncestorsSync(ctx context.Context, resourceId string) (*authorization_utils.Ancestors, error) {
	request := mgr.ClustersClient.NodePoolsAPI.GetNodepoolAncestors(ctx, resourceId)

	ancestors, _, err := request.Execute()
	if err != nil {
		return nil, err
	}

	return &authorization_utils.Ancestors{
		TenantId:     ancestors.TenantId,
		ClusterUuid:  ancestors.ClusterUuid.Get(),
		DepartmentId: ancestors.DepartmentId.Get(),
		ProjectId:    ancestors.ProjectId.Get(),
	}, nil
}

func (mgr *OrgHierarchyManager) GetScopeName(ctx context.Context, scopeType api.ScopeType, scopeId string) (string, *api.Error) {
	errMsg := fmt.Sprintf("Failed to get the scope name of resource %s-%s", string(scopeType), scopeId)
	var name string
	var response *http.Response
	var err error
	switch scopeType {
	case api.SCOPETYPE_SYSTEM:
		return "system-scope", nil
	case api.SCOPETYPE_TENANT:
		name, response, err = mgr.getTenantName(ctx, scopeId)
	case api.SCOPETYPE_CLUSTER:
		name, response, err = mgr.getClusterName(ctx, scopeId)
	default:
		name, response, err = mgr.getOrgUnitScopeName(ctx, scopeType, scopeId)
	}

	if err != nil || (response != nil && response.StatusCode >= http.StatusBadRequest) {
		if response != nil && response.StatusCode >= http.StatusBadRequest {
			if response.StatusCode == http.StatusNotFound {
				return "", api.NewError(http.StatusNotFound, fmt.Sprintf("Cluster %s not found.", scopeId), nil)
			}
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error response from client on get scope name of resource %s-%s. status: %s", string(scopeType), scopeId, response.Status)

			return "", api.NewError(http.StatusInternalServerError, errMsg, nil)
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error from client on get scope name of resource %s-%s. error: %+v", string(scopeType), scopeId, err)
		return "", api.NewError(http.StatusInternalServerError, errMsg, nil)
	}

	return name, nil
}

func (mgr *OrgHierarchyManager) getClusterName(ctx context.Context, scopeId string) (string, *http.Response, error) {
	request := mgr.ClustersClient.ClustersAPI.GetClusterName(ctx, scopeId)
	responseObject, response, clientError := request.Execute()
	var name string
	if responseObject != nil {
		name = responseObject.Name
	}

	return name, response, clientError
}

func (mgr *OrgHierarchyManager) getTenantName(ctx context.Context, scopeId string) (string, *http.Response, error) {
	tenantId, err := strconv.ParseInt(scopeId, 10, 32)
	if err != nil {
		return "", nil, fmt.Errorf("invalid tenant id %s: %w", scopeId, err)
	}

	tenant, response, err := mgr.TenantsClient.GetTenantByIdInternal(ctx, int32(tenantId))
	if err != nil {
		return "", response, err
	}

	return tenant.Name, response, nil
}

func (mgr *OrgHierarchyManager) getOrgUnitScopeName(ctx context.Context, scopeType api.ScopeType, scopeId string) (string, *http.Response, error) {
	responseObject, err := mgr.OrgUnitServiceClient.GetOrgUnits(ctx, (*org_unit_service_swagger.ScopeType)(&scopeType), &scopeId)
	if err != nil {
		return "", &http.Response{
			StatusCode: int(err.Code),
		}, errors.New(err.Message)
	}

	var name string
	if len(responseObject) == 1 {
		name = responseObject[0].Name
	}

	return name, nil, nil
}

func (mgr *OrgHierarchyManager) GetOrgTree(ctx context.Context) (structs.OrgUnitMap, map[structs.OrgUnitKey]string, error) {
	orgTree, orgTreeNameMap, err := mgr.OrgUnitServiceClient.GetOrgTreeMap(ctx)
	if err != nil {
		return nil, nil, errors.New(err.Message)
	}

	return orgTree, orgTreeNameMap, nil
}

func (mgr *OrgHierarchyManager) GetDepartmentsDescendants(ctx context.Context, departmentIds []string) ([]string, error) {
	departmentIdsToFetch := []string{}
	projectIds := []string{}
	for _, departmentId := range departmentIds {
		descendants, err := authorization_utils.GetDescendantsCache(ctx, string(api.SCOPETYPE_DEPARTMENT), departmentId)
		if err == nil && descendants != nil {
			if descendants.ProjectIds != nil {
				projectIds = append(projectIds, descendants.ProjectIds...)
			}
		} else {
			departmentIdsToFetch = append(departmentIdsToFetch, departmentId)
		}
	}
	if len(departmentIdsToFetch) == 0 {
		return projectIds, nil
	}

	departmentsDescendants, getErr := mgr.OrgUnitServiceClient.GetDepartmentsDescendants(ctx, strings.Join(departmentIdsToFetch, ","))
	if getErr != nil {
		return nil, errors.New(getErr.Message)
	}
	if departmentsDescendants != nil {
		projectIds = append(projectIds, (*departmentsDescendants).ProjectIds...)
	}
	return projectIds, nil
}
