/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package departments

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/overtime_metrics_writer"

	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	workloads_client "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"

	"github.com/google/uuid"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	cluster_openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/lib/pq"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	ResourceType    = "departments"
	AuditEntityType = "department"
)

type DepartmentApiService struct {
	authorizationClient auth_client.IAuthorizationClient
	clusterClient       cluster_service_client.IClusterExternalClient
	workloadsClient     workloads_client.IWorkloadsClient
}

func NewDepartmentApiService(
	clusterClient cluster_service_client.IClusterExternalClient,
	workloadsClient workloads_client.IWorkloadsClient,
) api.DepartmentsAPIServicer {
	return &DepartmentApiService{
		authorizationClient: auth_client.NewAuthorizationClient(),
		clusterClient:       clusterClient,
		workloadsClient:     workloadsClient,
	}
}

func (s *DepartmentApiService) CountDepartments(ctx context.Context, filterBy []string) (api.ImplResponse, error) {
	tenantId, tenantErr := services.ParseTenantId(ctx)

	if tenantErr != nil {
		return services.HandleError(tenantErr, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	filters, filterErr := queryapi.NewParser(models.Department{}.SortAndFilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy))
		return errResponse, nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	count, dbErr := dao.CountDepartments(ctx, *tenantId, permittedScopes, filters, models.VerboseVerbosity)

	if dbErr != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to count departments")
		return errResponse, nil
	}

	return api.Response(http.StatusOK, api.CountResponse{Count: count}), nil
}

func (s *DepartmentApiService) CreateDepartment(ctx context.Context, request api.DepartmentCreationRequest) (api.ImplResponse, error) {
	// Acquire read lock on cluster - allows multiple Create/Update department calls concurrently,
	// but blocks when NodePoolUpdated has the write lock
	services.GetClusterLockManager().RLock(request.ClusterId)
	defer services.GetClusterLockManager().RUnlock(request.ClusterId)

	clusterUuid, err := uuid.Parse(request.ClusterId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("failed to parse cluster id %s", request.ClusterId))
		return errResponse, nil
	}
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	clusterPermitted := isClusterPermitted(ctx, tenantId, request.ClusterId)
	if !clusterPermitted {
		return services.HandleApiError(*api.NewError(http.StatusNotFound, "cluster not found", err)), nil
	}

	supported, err := services.IsResourceRankSupported(ctx, request.Resources, clusterUuid, s.clusterClient)
	if !supported || err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("rank value for department is not supported for clusters with version less than %s", services.ResourcePriorityMinVersion))
		return errResponse, nil
	}
	nodePoolsMetadata, err := services.GetNodePoolsMetadataMap(ctx, s.clusterClient, request.ClusterId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
		return errResponse, nil
	}

	err = validateCreateRequest(ctx, request, nodePoolsMetadata, request.ClusterId, *tenantId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("create department request validation failed: %s", err.Error()))
		return errResponse, nil
	}

	department, err := MapCreateRequestToModel(ctx, request, nodePoolsMetadata)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to map request to model"), nil
	}
	department.TenantId = *tenantId
	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	} else {
		department.CreatedBy = *subjectId
	}
	response, err := dao.CreateDepartment(ctx, &department)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to create department")
		return errResponse, nil
	}

	// Enrich node type names on the in-memory model so the response includes them
	allNodeTypes, err := dao.GetNodeTypes(ctx, *tenantId, clusterUuid, nil)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get node types for enrichment: %v", err)
	} else if allNodeTypes != nil {
		services.EnrichNodeTypeNames(&response.OrgUnitData, services.BuildNodeTypeNameMap(*allNodeTypes))
	}

	audit.AuditFromCtx(ctx).WithEntityName(response.Name).WithEntityType(AuditEntityType).WithEntityId(strconv.FormatInt(response.ID, 10))

	return api.Response(http.StatusCreated, MapModelToResponse(*response, models.SlimVerbosity, nodePoolsMetadata)), nil
}

func isClusterPermitted(ctx context.Context, tenantId *string, clusterId string) bool {
	permittedScopes, _ := auth_client.ParsePermittedScopes(ctx)
	if permittedScopes == nil {
		return true
	}
	tenantPermitted := dao.IsTenantPermitted(*tenantId, permittedScopes)
	if tenantPermitted {
		return true
	}
	for _, cluster := range permittedScopes.Clusters {
		if cluster == clusterId {
			return true
		}
	}
	return false
}

func (s *DepartmentApiService) DeleteDepartment(ctx context.Context, departmentId string) (api.ImplResponse, error) {
	departmentUid, err := strconv.ParseInt(departmentId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "delete department failed")
		return errResponse, nil
	}

	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	err = validateDeleteRequest(ctx, departmentUid, *tenantId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "delete department request validation failed")
		return errResponse, nil
	}

	departmentName, err := dao.DeleteDepartment(ctx, departmentUid, *tenantId, subjectId, services.DefaultDepartmentName)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to delete department")
		return errResponse, nil
	}

	request := s.authorizationClient.DeleteScopeAccessRules(ctx).ScopeType(string(openapi.SCOPETYPE_DEPARTMENT)).ScopeId(departmentId)
	execute, err := s.authorizationClient.DeleteScopeAccessRulesExecute(request)
	if err != nil || execute.StatusCode != http.StatusNoContent {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete department admins for department %s", departmentId)
	}

	audit.AuditFromCtx(ctx).WithEntityName(departmentName).WithEntityType(AuditEntityType).WithEntityId(departmentId)

	cacheError := cache.Del[authorization_utils.Descendants](ctx, authorization_utils.GetDescendantsKey(departmentId, string(api.SCOPETYPE_DEPARTMENT)))
	if cacheError != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete cache for department %s: %v", departmentId, cacheError)
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (s *DepartmentApiService) GetDepartment(ctx context.Context, departmentId string) (api.ImplResponse, error) {
	verbosityEnum := models.VerboseVerbosity
	departmentUid, err := strconv.ParseInt(departmentId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "get department failed")
		return errResponse, nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	department, err := dao.GetDepartmentById(ctx, departmentUid, *tenantId, verbosityEnum, nil)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department")
		return errResponse, nil
	}

	nodePools, err := services.GetNodePoolsMetadataMap(ctx, s.clusterClient, department.ClusterId.String())
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
		return errResponse, nil
	}

	return api.Response(http.StatusOK, MapModelToResponse(*department, verbosityEnum, nodePools)), nil
}

func (s *DepartmentApiService) GetDepartmentsDescendants(ctx context.Context, orgUnitsIds string) (api.ImplResponse, error) {
	departmentsIds, err := services.ParseOrgUnitIds(orgUnitsIds)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("Failed to parse org units ids %s", orgUnitsIds)), nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil || tenantId == nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	projects, err := dao.GetProjectsIdsByDepartmentIds(ctx, *tenantId, departmentsIds)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department descendants")
		return errResponse, nil
	}

	projectIds := services.ArrayToString(toIds(projects))
	descendants := &authorization_utils.Descendants{
		ProjectIds: projectIds,
	}

	addToDescendantsCache(ctx, projects)
	return api.Response(http.StatusOK, *descendants), nil
}

func addToDescendantsCache(ctx context.Context, projects []models.Project) {
	mappedByDepartment := make(map[int64][]int64)
	for _, project := range projects {
		if project.ParentID == nil {
			continue
		}
		mappedByDepartment[*project.ParentID] = append(mappedByDepartment[*project.ParentID], project.ID)
	}

	for departmentId, projectIds := range mappedByDepartment {
		authorization_utils.SetDescendantsCache(ctx, string(openapi.RESOURCETYPE_DEPARTMENT),
			strconv.Itoa(int(departmentId)),
			authorization_utils.Descendants{
				ProjectIds: services.ArrayToString(projectIds),
			})
	}
}

func toIds(arr []models.Project) []int64 {
	if len(arr) == 0 {
		return []int64{}
	}
	result := make([]int64, len(arr))
	for i, v := range arr {
		result[i] = v.ID
	}
	return result
}

func (s *DepartmentApiService) GetDepartmentAncestors(ctx context.Context, departmentId string) (api.ImplResponse, error) {
	ancestors, err := authorization_utils.GetAncestorsCache(ctx, ResourceType, departmentId)
	if err == nil {
		return api.Response(http.StatusOK, ancestors), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Could not find department <%s> ancestors in cache. err: %v, continuing", departmentId, err)

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	departmentUid, err := strconv.ParseInt(departmentId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("get department <%s> ancestors failed", departmentId))
		return errResponse, nil
	}

	department, err := dao.GetDepartmentById(ctx, departmentUid, *tenantId, models.MetaDataVerbosity, nil)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department ancestors")
		return errResponse, nil
	}

	ancestors = &authorization_utils.Ancestors{
		TenantId:    *tenantId,
		ClusterUuid: convutils.Ref(department.ClusterId.String()),
	}

	authorization_utils.SetAncestorsCache(ctx, ResourceType, departmentId, *ancestors)

	return api.Response(http.StatusOK, *ancestors), nil
}

func (s *DepartmentApiService) GetDepartments(
	ctx context.Context,
	filterBy []string,
	sortBy api.DepartmentFilterSortFields,
	verbosity api.Verbosity,
	sortOrder string,
	offset int32,
	limit int32,
) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	filters, filterErr := queryapi.NewParser(models.Department{}.SortAndFilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy))
		return errResponse, nil
	}

	sortColumn, sortError := queryapi.NewParser(models.Department{}.SortAndFilterFieldsToColumn()).GetMapping(string(sortBy))
	if sortError != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("sort %s invalid", sortBy))
		return errResponse, nil
	}
	modelVerbosity, err := ValidateAndGetVerbosity(ctx, string(verbosity))
	if err != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("verbosity %s invalid", verbosity))
		return errResponse, nil
	}
	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	departments, err := dao.GetDepartments(ctx, *tenantId, filters, sortColumn, sortOrder, offset, limit, permittedScopes, false, modelVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department")
		return errResponse, nil
	}

	var nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal
	if len(*departments) > 0 {
		clusterIds := services.FetchClusterIdsOfDepartments(*departments)
		nodePoolsMetadata, err = services.GetNodePoolsMultiCluster(ctx, s.clusterClient, clusterIds)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
			return errResponse, nil
		}
	}

	apiDepartments := make([]api.DepartmentForList, len(*departments))
	for index, department := range *departments {
		tempDepartment := MapModelToResponse(department, modelVerbosity, nodePoolsMetadata)
		var status *api.DepartmentForListAllOfStatus
		if !api.IsZeroValue(tempDepartment.Status.QuotaStatus) {
			status = &api.DepartmentForListAllOfStatus{
				QuotaStatus: tempDepartment.Status.QuotaStatus,
			}
		}

		apiDepartments[index] = api.DepartmentForList{
			Description:    tempDepartment.Description,
			Resources:      tempDepartment.Resources,
			Name:           tempDepartment.Name,
			ClusterId:      tempDepartment.ClusterId,
			Id:             tempDepartment.Id,
			CreatedAt:      tempDepartment.CreatedAt,
			UpdatedAt:      tempDepartment.UpdatedAt,
			CreatedBy:      tempDepartment.CreatedBy,
			UpdatedBy:      tempDepartment.UpdatedBy,
			Children:       tempDepartment.Children,
			Status:         status,
			TotalResources: tempDepartment.TotalResources,
			OvertimeData:   tempDepartment.OvertimeData,
		}
	}

	response := api.GetDepartments200Response{
		Departments: apiDepartments,
	}

	recordsCount := int32(len(apiDepartments))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *DepartmentApiService) UpdateDepartment(ctx context.Context, departmentId string, departmentUpdateRequest api.DepartmentUpdateRequest) (api.ImplResponse, error) {
	departmentUid, err := strconv.ParseInt(departmentId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "update department resources failed")
		return errResponse, nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	// Acquire read lock on cluster - allows multiple Create/Update department calls concurrently,
	// but blocks when NodePoolUpdated has the write lock
	clusterId, err := dao.GetClusterIdByDepartment(ctx, departmentUid, *tenantId)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department cluster")
		return errResponse, nil
	}
	services.GetClusterLockManager().RLock(clusterId)
	defer services.GetClusterLockManager().RUnlock(clusterId)

	department, err := dao.GetDepartmentById(ctx, departmentUid, *tenantId, models.VerboseVerbosity, nil)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to update department resources")
		return errResponse, nil
	}

	nodePools, err := services.GetNodePoolsMetadataMap(ctx, s.clusterClient, department.ClusterId.String())
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
		return errResponse, nil
	}

	apiErr := s.validateUpdateRequest(ctx, departmentUpdateRequest, nodePools)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	updated, err := MapUpdateRequestToModel(ctx, departmentUpdateRequest, department, nodePools)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to map request to model"), nil
	}

	projects := []models.Project{}
	if shouldUpdateOverrideFields(department.OrgUnitData, updated.OrgUnitData) {
		projectsByParent, err := dao.GetProjectsByParent(ctx, department.ID, department.TenantId, models.SlimVerbosity)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusServiceUnavailable, "failed to update department")
			return errResponse, nil
		}
		err = validateProjects(updated, projectsByParent)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusBadRequest, "failed to update department, projects fields cannot be override")
			return errResponse, nil
		}
		if projectsByParent != nil {
			for _, project := range *projectsByParent {
				project.OrgUnitData.EffectiveData = services.MapEffectiveFields(department.OrgUnitDataID, &updated.OrgUnitData, project.OrgUnitData, false)
				projects = append(projects, project)
			}
		}
	}

	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	} else {
		updated.UpdatedBy = *subjectId
	}

	apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx, *tenantId, updated.OrgUnitData.Resources, departmentUid)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	response, err := dao.UpdateDepartmentWithChildren(ctx, &updated, &projects)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to update department")
		return errResponse, nil
	}
	audit.AuditFromCtx(ctx).WithEntityName(response.Name).WithEntityType(AuditEntityType).WithEntityId(strconv.FormatInt(response.ID, 10))

	// for response to contain verbose data
	response.ProjectChildren = department.ProjectChildren
	response.QuotaStatus = department.QuotaStatus
	response.OrgUnitData.Children = department.OrgUnitData.Children

	// Enrich node type names on the in-memory model so the response includes them
	allNodeTypes, err := dao.GetNodeTypes(ctx, *tenantId, department.ClusterId, nil)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get node types for enrichment: %v", err)
	} else if allNodeTypes != nil {
		services.EnrichNodeTypeNames(&response.OrgUnitData, services.BuildNodeTypeNameMap(*allNodeTypes))
	}

	return api.Response(http.StatusOK, MapModelToResponse(*response, models.VerboseVerbosity, nodePools)), nil
}

func (s *DepartmentApiService) UpdateDepartmentResources(ctx context.Context, departmentId string, resources []api.ResourcesNullable) (api.ImplResponse, error) {
	departmentUid, err := strconv.ParseInt(departmentId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "update department resources failed")
		return errResponse, nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	// Acquire read lock on cluster - allows multiple Create/Update department calls concurrently,
	// but blocks when NodePoolUpdated has the write lock
	clusterId, err := dao.GetClusterIdByDepartment(ctx, departmentUid, *tenantId)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department cluster")
		return errResponse, nil
	}
	services.GetClusterLockManager().RLock(clusterId)
	defer services.GetClusterLockManager().RUnlock(clusterId)

	department, err := dao.GetDepartmentById(ctx, departmentUid, *tenantId, models.SlimVerbosity, nil)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to update department resources")
		return errResponse, nil
	}
	supported, err := services.IsResourceRankSupported(ctx, resources, department.ClusterId, s.clusterClient)
	if !supported || err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("rank value for department is not supported for clusters with version less than %s", services.ResourcePriorityMinVersion))
		return errResponse, nil
	}
	nodePools, err := services.GetNodePoolsMetadataMap(ctx, s.clusterClient, department.ClusterId.String())
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
		return errResponse, nil
	}

	err = services.ValidateResourcesUpdateRequest(ctx, resources, nodePools)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "update department resources validation failed")
		return errResponse, nil
	}

	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	} else {
		department.UpdatedBy = *subjectId
	}

	originalResourcesNodePoolToId := make(map[string]uuid.UUID)
	for _, originalResources := range department.OrgUnitData.Resources {
		originalResourcesNodePoolToId[originalResources.NodePoolId] = originalResources.ID
	}
	newResources, err := services.ApiResourcesToModel(ctx, resources, true, nodePools, true, originalResourcesNodePoolToId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusServiceUnavailable, "failed to map resources to model")
		return errResponse, nil
	}
	services.UpdateOrgUnitDataResources(newResources, &department.OrgUnitData)

	apiErr := services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx, *tenantId, department.OrgUnitData.Resources, departmentUid)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	_, err = dao.UpdateDepartment(ctx, department)
	if err != nil {
		message := services.EnrichUpdateMessage("Failed to update department", "Department", err)
		errResponse := services.HandleDbError(ctx, err, message)
		return errResponse, nil
	}

	audit.AuditFromCtx(ctx).WithEntityName(department.Name).WithEntityType(AuditEntityType).WithEntityId(strconv.FormatInt(department.ID, 10))

	return api.Response(http.StatusOK, departmentApiResourcesFromModel(*department, nodePools)), nil
}

func (s *DepartmentApiService) GetDepartmentsTelemetry(
	ctx context.Context,
	telemetryType api.OrgUnitTelemetryType,
	clusterId string,
	nodepoolName string,
	departmentId string,
	groupBy []string,
) (api.ImplResponse, error) {
	panic("implement me")
}

func (s *DepartmentApiService) GetDepartmentMetrics(
	ctx context.Context,
	departmentId string,
	metricTypes []api.OrgUnitMetricType,
	start time.Time,
	end time.Time,
	numberOfSamples int32,
	nodePoolName string,
	groupBy string,
) (api.ImplResponse, error) {
	apiErr := services.ValidateRequestedMetrics(len(metricTypes))
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	departmentUid, err := strconv.ParseInt(departmentId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "get department metrics failed")
		return errResponse, nil
	}
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	department, err := dao.GetDepartmentById(ctx, departmentUid, *tenantId, models.MetaDataVerbosity, nil)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "get department metrics failed")
		return errResponse, nil
	}

	err = services.ValidateMinClusterVersionForMetrics(ctx, department.ClusterId, metricTypes, s.clusterClient)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, err.Error()), nil
	}

	return services.GetOrgUnitMetrics(ctx, department.ClusterId.String(), nil, &department.Name, &nodePoolName, metricTypes, start, end, numberOfSamples, groupBy)
}

func (s *DepartmentApiService) patchDepartmentResources(
	ctx context.Context,
	departmentId string,
	resources []api.ResourcesNullable,
	tenantId string,
	subjectId *string,
	nodePoolsPerCluster *sync.Map,
	permittedScopes *openapi.PermittedScopes,
) (*[]api.Resources, string, *api.Error) {
	departmentUid, err := strconv.ParseInt(departmentId, 10, 64)
	if err != nil {
		return nil, "", api.NewError(http.StatusBadRequest, "failed to parse department id", err)
	}

	department, err := dao.GetDepartmentById(ctx, departmentUid, tenantId, models.SlimVerbosity, permittedScopes)
	if err != nil {
		return nil, "", api.NewError(http.StatusInternalServerError, "failed to retrieve department", err)
	}

	// Resolve nodepools for this department's cluster using cache
	nodePools, apiErr := services.GetNodePoolsForClusterFromServiceOrCache(ctx, s.clusterClient, department.ClusterId.String(), nodePoolsPerCluster)
	if apiErr != nil {
		return nil, "", apiErr
	}

	err = services.ValidateResourcesPatchRequest(ctx, resources, nodePools)
	if err != nil {
		return nil, "", api.NewError(http.StatusBadRequest, fmt.Sprintf("patch department resources validation failed: %s", err.Error()), err)
	}

	if subjectId != nil {
		department.UpdatedBy = *subjectId
	}

	originalResourcesNodePoolToId := make(map[string]uuid.UUID)
	for _, originalResources := range department.OrgUnitData.Resources {
		originalResourcesNodePoolToId[originalResources.NodePoolId] = originalResources.ID
	}
	newResources, err := services.ApiResourcesToModel(ctx, resources, true, nodePools, false, originalResourcesNodePoolToId)
	if err != nil {
		return nil, "", api.NewError(http.StatusInternalServerError, "failed to map resources to model", err)
	}
	err = services.PatchOrgUnitDataResources(newResources, &department.OrgUnitData)
	if err != nil {
		return nil, "", api.NewError(http.StatusBadRequest, "failed to map resources to model", err)
	}
	apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx, tenantId, department.OrgUnitData.Resources, departmentUid)
	if apiErr != nil {
		return nil, "", apiErr
	}

	_, err = dao.UpdateDepartment(ctx, department)
	if err != nil {
		return nil, "", api.NewError(http.StatusInternalServerError, "failed to update department", err)
	}

	updated := departmentApiResourcesFromModel(*department, nodePools)
	return &updated, department.Name, nil
}

func (s *DepartmentApiService) patchDepartmentNodePoolResources(
	ctx context.Context,
	departmentId string,
	patchRequest api.NodePoolResourcesPatchRequest,
	tenantId string,
	subjectId *string,
	nodePoolsPerCluster *sync.Map,
	permittedScopes *openapi.PermittedScopes,
) (*api.DepartmentNodePoolResources, string, *api.Error) {
	departmentUid, err := strconv.ParseInt(departmentId, 10, 64)
	if err != nil {
		return nil, "", api.NewError(http.StatusBadRequest, "failed to parse department id", err)
	}

	departmentNodePoolResources, err := dao.GetDepartmentNodePoolResourcesPerDepartmentNodePool(ctx, tenantId, departmentId, patchRequest.NodePoolId, permittedScopes)
	if err != nil {
		return nil, "", api.NewError(http.StatusInternalServerError, "failed to get department node pool resources", err)
	}

	// Resolve nodepools for this department's cluster using cache
	nodePools, apiErr := services.GetNodePoolsForClusterFromServiceOrCache(ctx, s.clusterClient, departmentNodePoolResources.ClusterId.String(), nodePoolsPerCluster)
	if apiErr != nil {
		return nil, "", apiErr
	}

	convertedResources := services.ConvertNodePoolResourcesPatchRequest(patchRequest, departmentNodePoolResources.GpuDeserved)
	err = services.ValidateNodePoolResourcesPatchRequest(ctx, convertedResources, nodePools)
	if err != nil {
		return nil, "", api.NewError(http.StatusBadRequest, fmt.Sprintf("patch department resources validation failed: %s", err.Error()), err)
	}

	orgUnitResources := []models.Resources{departmentNodePoolResources.ToResources()}
	patchedResources, err := services.MapNodePoolPatchResourcesRequestToModel(ctx, []api.ResourcesNullable{convertedResources}, orgUnitResources, nodePools, []api.NodePoolResourcesPatchRequest{patchRequest})
	if err != nil {
		return nil, "", api.NewError(http.StatusInternalServerError, "failed to map resources to model", err)
	}
	apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx, tenantId, patchedResources, departmentUid)
	if apiErr != nil {
		return nil, "", apiErr
	}

	// map and patch in db
	patchUpdates, err := services.ApiNodePoolResourcesPatchRequestToModelResourcesUpdates(patchRequest, nodePools, true)
	if err != nil {
		return nil, "", api.NewError(http.StatusInternalServerError, "failed to map patch request to model updates", err)
	}
	err = dao.PatchDepartmentNodePoolResources(ctx, tenantId, departmentNodePoolResources.ClusterId.String(), departmentId, patchRequest.NodePoolId, departmentNodePoolResources.ID.String(), subjectId, patchUpdates)
	if err != nil {
		return nil, "", api.NewError(http.StatusInternalServerError, "failed to patch department node pool resources", err)
	}

	// get for response
	nodePoolResources, err := dao.GetDepartmentNodePoolResourcesPerDepartmentNodePool(ctx, tenantId, departmentId, patchRequest.NodePoolId, nil)
	if err != nil {
		return nil, "", api.NewError(http.StatusInternalServerError, "failed to get department node pool resources", err)
	}

	result := services.MapModelDepartmentNodePoolResourcesToApi(*nodePoolResources, nodePools)
	return result, departmentNodePoolResources.OrgUnitName, nil
}

func (s *DepartmentApiService) PatchDepartmentResources(ctx context.Context, departmentId string, resources []api.ResourcesNullable) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	}
	nodePoolsCache := sync.Map{}

	updatedResources, departmentName, apiErr := s.patchDepartmentResources(ctx, departmentId, resources, *tenantId, subjectId, &nodePoolsCache, nil)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	audit.AuditFromCtx(ctx).WithEntityName(departmentName).WithEntityType(AuditEntityType).WithEntityId(departmentId)
	return api.Response(http.StatusOK, *updatedResources), nil
}

func (s *DepartmentApiService) BatchPatchDepartmentNodePoolResources(ctx context.Context, request api.BatchPatchNodePoolResourcesRequest) (api.ImplResponse, error) {
	apiErr := services.ValidateBatchPatchNodePoolResourcesRequest(request)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	succeeded, failed, apiErr := s.bulkPatchDepartmentResources(ctx, request, *tenantId)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	response := api.BatchPatchDepartmentResourcesResponse{
		Id:              uuid.New().String(),
		Succeeded:       succeeded,
		Failed:          failed,
		TotalRequested:  int32(len(request.PatchRequests)),
		TotalSuccessful: int32(len(succeeded)),
		TotalFailed:     int32(len(failed)),
	}

	// Per convention: 500 when all operations fail, 200 otherwise
	if len(succeeded) == 0 {
		return api.Response(http.StatusInternalServerError, response), nil
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *DepartmentApiService) bulkPatchDepartmentResources(
	ctx context.Context,
	request api.BatchPatchNodePoolResourcesRequest,
	tenantId string,
) ([]api.DepartmentNodePoolResources, []api.FailedNodePoolResourcesPatchRequest, *api.Error) {
	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id in bulkPatchDepartmentResources: %v", err)
	}
	nodePoolsCache := sync.Map{}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	totalRequested := len(request.PatchRequests)
	numberOfWorkers := min(runtime.NumCPU(), totalRequested)
	workChan := make(chan api.NodePoolResourcesPatchRequest, totalRequested)
	resultChan := make(chan struct{}, totalRequested)
	succeeded := services.SyncSlice[api.DepartmentNodePoolResources]{}
	failed := services.SyncSlice[api.FailedNodePoolResourcesPatchRequest]{}

	// worker function
	var wg sync.WaitGroup
	departmentsAuditIdToName := map[string]string{}
	worker := func() {
		defer wg.Done()
		for patchRequest := range workChan {
			resourcesResponse, departmentName, apiErr := s.patchDepartmentNodePoolResources(ctx, patchRequest.OrgUnitId, patchRequest, tenantId, subjectId, &nodePoolsCache, permittedScopes)
			if apiErr != nil {
				failed.Append(api.FailedNodePoolResourcesPatchRequest{
					Payload: patchRequest,
					Code:    apiErr.Code,
					Message: apiErr.Message,
				})
			} else {
				succeeded.Append(*resourcesResponse)
				departmentsAuditIdToName[patchRequest.OrgUnitId] = departmentName
			}
			resultChan <- struct{}{}
		}
	}
	for departmentId, departmentName := range departmentsAuditIdToName {
		auditLogError := audit.AuditFromAudit(audit.AuditFromCtx(ctx)).
			WithAction("Patch").
			WithHttpMethod(http.MethodPatch).
			WithEntityName(departmentName).
			WithEntityType(AuditEntityType).
			WithEntityId(departmentId).
			WithStatusCode(http.StatusOK).
			Write()
		if auditLogError != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to write audit log for patch resources of department %s (id %s): %v",
				departmentName, departmentId, auditLogError)
		}
	}
	audit.AuditFromCtx(ctx).Drop()

	// start workers
	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go worker()
	}
	// send work to workers
	go func() {
		for _, patchRequest := range request.PatchRequests {
			workChan <- patchRequest
		}
		close(workChan)
	}()

	// wait for all work to complete
	for range request.PatchRequests {
		<-resultChan
	}
	// wait for all workers to finish
	wg.Wait()

	return (&succeeded).GetAll(), (&failed).GetAll(), nil
}

func (s *DepartmentApiService) GetDepartmentsRanking(ctx context.Context, metricType api.RankingMetricType, orderType api.RankingOrderType, clusterId string, nodePoolName string) (api.ImplResponse, error) {
	const NumberOfDepartments = 15
	var results []api.DepartmentGetRanking
	var err error

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	if err = services.ValidateGetOrgUnitRankingParams(ctx, metricType, orderType, clusterId, s.clusterClient); err != nil {
		return api.Response(http.StatusBadRequest, err.Error()), nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	// gpuAllocationRatio with nodepool filter: query resources table directly for per-nodepool ratio
	if metricType == api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO && nodePoolName != "" {
		rankingResults, innerErr := dao.GetAllocationRatioRankingByNodepool(ctx, *tenantId, clusterId, nodePoolName, permittedScopes, orderType, NumberOfDepartments, true, "")
		if innerErr != nil {
			return services.HandleDbError(ctx, innerErr, "Failed to get department ranking"), nil
		}
		for _, r := range rankingResults {
			results = append(results, api.DepartmentGetRanking{
				Name: r.Name, Id: r.Id, GpuQuota: r.GpuQuota,
				AllocationValue: r.AllocationValue, UtilizationValue: r.UtilizationValue,
				MemoryUtilizationValue: r.MemoryUtilizationValue,
			})
		}
		err = nil
	} else if nodePoolName == "" || metricType == api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO {
		columnNames, innerErr := dao.GetColumnNamesForRanking(metricType, true)
		if innerErr != nil {
			return api.Response(http.StatusBadRequest, innerErr.Error()), nil
		}
		orderBy, innerErr := dao.GetOrderBy(metricType, true)
		if innerErr != nil {
			return api.Response(http.StatusBadRequest, innerErr.Error()), nil
		}

		results, err = dao.GetDepartmentsRanking(ctx, *tenantId, clusterId, permittedScopes, columnNames, orderBy, orderType, NumberOfDepartments)
		if err != nil {
			return services.HandleDbError(ctx, err, "Failed to get department ranking"), nil
		}
	} else {
		clusterFilter := []string{clusterId}
		var departmentsNames []string

		if permittedScopes != nil {
			departmentsNames, err = dao.GetDepartmentNamesByIds(ctx, permittedScopes.Departments)
			if err != nil {
				return services.HandleDbError(ctx, err, "Failed to get department ranking"), nil
			}
		}

		results, err = overtime_metrics_writer.QueryDepartmentsOvertimeRankingMetrics(ctx, *tenantId, metricType, orderType, clusterFilter, nodePoolName, departmentsNames)
	}

	if err != nil {
		return api.Response(http.StatusInternalServerError, err.Error()), nil
	}

	if len(results) == 0 {
		results = []api.DepartmentGetRanking{}
	}

	enrichedResults, enrichErr := s.enrichDepartmentsRankingWithWorkloadCounts(ctx, results, clusterId, nodePoolName)
	if enrichErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to enrich departments ranking with workload counts: %v", enrichErr)
	} else {
		results = enrichedResults
	}

	response := api.GetDepartmentsRanking200Response{
		Departments: results,
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *DepartmentApiService) GetSyncDepartments(ctx context.Context, clusterId string, updatedAfter time.Time, offset int32, limit int32) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("failed to parse cluster id %s", clusterId))
		return errResponse, nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	departments, err := dao.GetSyncDepartments(ctx, *tenantId, clusterUuid, updatedAfter, offset, limit, permittedScopes)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "get departments sync failed")
		return errResponse, nil
	}

	var nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal
	if len(*departments) > 0 {
		nodePoolsMetadata, err = services.GetNodePoolsMetadataMap(ctx, s.clusterClient, clusterId)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
			return errResponse, nil
		}
	}

	apiDepartments := make([]api.DepartmentSync, len(*departments))

	for index, department := range *departments {
		apiDepartments[index] = MapModelToSyncResponse(department, nodePoolsMetadata)
	}

	response := api.GetSyncDepartments200Response{
		Departments: apiDepartments,
	}

	recordsCount := int32(len(apiDepartments))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *DepartmentApiService) UpdateDepartmentStatus(ctx context.Context, departmentId string, request api.DepartmentStatusUpdateRequest) (api.ImplResponse, error) {
	clusterUuid, err := uuid.Parse(request.ClusterId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("failed to parse cluster id %s", request.ClusterId))
		return errResponse, nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	err = dao.UpdateDepartmentStatus(ctx, clusterUuid, *tenantId, departmentId, request.Status.LastUpdatedTime)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "update department status failed")
		return errResponse, nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (s *DepartmentApiService) DepartmentsResetResources(ctx context.Context, request api.DepartmentsResourcesResetRequest) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	err = services.ValidateResetResourcesRequest(ctx, *tenantId, request.ClusterId, "", request.NodePoolIds, s.clusterClient)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, err.Error())
		return errResponse, nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	apiErr := validateResetResourcesPermitted(ctx, *tenantId, request.ClusterId, permittedScopes)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	dbErr := dao.ResetResources(ctx, *tenantId, request.ClusterId, "", request.NodePoolIds, request.Action, false)
	if dbErr != nil {
		errResponse := services.HandleDbError(ctx, dbErr, "failed to reset departments resources")
		return errResponse, nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (s *DepartmentApiService) GetDepartmentNodePoolResources(ctx context.Context, filterBy []string,
	sortBy api.DepartmentNodePoolResourcesFilterSortFields, sortOrder string, offset int32, limit int32, search string,
) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	filters, filterErr := queryapi.NewParser(models.DepartmentNodePoolResources{}.FilterByFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy))
		return errResponse, nil
	}
	sortColumn, sortError := queryapi.NewParser(models.DepartmentNodePoolResources{}.SortByFieldsToColumn()).GetMapping(string(sortBy))
	if sortError != nil {
		errResponse := services.HandleError(sortError, http.StatusBadRequest, fmt.Sprintf("sort %s invalid", sortBy))
		return errResponse, nil
	}

	shouldValidateNodePoolAdjustableWeights, apiErr := services.ValidateSortByAllowedForGetNodePoolResources(string(sortBy), filters)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	nodePoolResources, err := dao.GetDepartmentNodePoolResources(ctx, *tenantId, filters, sortColumn, sortOrder, offset, limit, permittedScopes, search)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department nodePool resources")
		return errResponse, nil
	}
	var nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal
	if len(*nodePoolResources) > 0 {
		clusterIds := services.FetchClusterIdsOfDepartmentNodePoolResources(*nodePoolResources)
		nodePoolsMetadata, err = services.GetNodePoolsMultiCluster(ctx, s.clusterClient, clusterIds)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
			return errResponse, nil
		}

		if shouldValidateNodePoolAdjustableWeights {
			// should have been filtered by a specific node pool, so just check the first one
			capabilities, _ := services.GetSchedulingCapabilitiesPerNodePoolId(nodePoolsMetadata, (*nodePoolResources)[0].NodePoolId)
			if !capabilities.AdjustableWeights {
				apiErr = api.NewError(http.StatusBadRequest, "sorting by gpuOverQuotaWeight is only allowed when filtering by a specific node pool that has AdjustableWeights capability", nil)
				return services.HandleApiError(*apiErr), nil
			}
		}
	}

	apiResources := []api.DepartmentNodePoolResources{}
	for _, npResources := range *nodePoolResources {
		apiResources = append(apiResources, *services.MapModelDepartmentNodePoolResourcesToApi(npResources, nodePoolsMetadata))
	}

	response := api.GetDepartmentNodePoolResources200Response{
		NodePoolResources: apiResources,
	}

	recordsCount := int32(len(apiResources))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *DepartmentApiService) CountDepartmentNodePoolResources(ctx context.Context, filterBy []string, search string) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	filters, filterErr := queryapi.NewParser(models.DepartmentNodePoolResources{}.FilterByFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy))
		return errResponse, nil
	}
	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	nodePoolResourcesCount, err := dao.CountDepartmentNodePoolResources(ctx, *tenantId, filters, search, permittedScopes)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department nodePool resources")
		return errResponse, nil
	}

	return api.Response(http.StatusOK, api.CountResponse{Count: *nodePoolResourcesCount}), nil
}

func (s *DepartmentApiService) SumDepartmentNodePoolResources(ctx context.Context, nodePoolId string, filterBy []string) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	filters, filterErr := queryapi.NewParser(models.DepartmentNodePoolResources{}.FilterByFieldsToColumnForSum()).Filters(filterBy)
	if filterErr != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy))
		return errResponse, nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	nodePoolMetadata, err := services.GetNodePoolMetadata(ctx, s.clusterClient, nodePoolId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pool metadata")
		return errResponse, nil
	}
	capabilities := ptr.Deref(nodePoolMetadata.SchedulingCapabilities, cluster_openapi.SchedulingCapabilities{})

	sums, err := dao.SumDepartmentNodePoolResources(ctx, *tenantId, nodePoolId, filters, capabilities.AdjustableWeights, permittedScopes)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department nodePool resources")
		return errResponse, nil
	}

	return api.Response(http.StatusOK, api.DepartmentSumResponse{DeservedSum: sums.DeservedSum, WeightsSum: sums.WeightsSum}), nil
}

func (s *DepartmentApiService) validateUpdateRequest(ctx context.Context, request api.DepartmentUpdateRequest,
	nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal,
) *api.Error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating department update request")
	err := services.ValidateResources(ctx, request.Resources, nodePoolsMetadata)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("update department request validation failed: %s", err.Error()), err)
	}
	err = services.ValidateNodePools(ctx, request.Resources)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("update department request validation failed: %s", err.Error()), err)
	}

	return nil
}

func validateDeleteRequest(ctx context.Context, departmentId int64, tenantId string) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating department deletion request")
	err := validateNoChildren(ctx, departmentId, tenantId)
	if err != nil {
		return err
	}
	err = validateNotDefault(ctx, departmentId, tenantId)
	if err != nil {
		return err
	}
	return nil
}

func validateCreateRequest(ctx context.Context, request api.DepartmentCreationRequest, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal, clusterId string, tenantId string) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating department creation request")
	err := services.ValidateOrgName(request.Name, "department")
	if err != nil {
		return err
	}
	err = services.ValidateResources(ctx, request.Resources, nodePoolsMetadata)
	if err != nil {
		return err
	}
	err = services.ValidateNodePools(ctx, request.Resources)
	return err
}

func ValidateAndGetVerbosity(ctx context.Context, verbosity string) (models.Verbosity, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating get project request")
	return services.GetVerbosity(verbosity)
}

func validateNoChildren(ctx context.Context, departmentId int64, tenantId string) error {
	departmentChildren, err := dao.GetDepartmentsByParent(ctx, departmentId, tenantId)
	if err != nil {
		return err
	}

	projectChildren, err := dao.GetProjectsByParent(ctx, departmentId, tenantId, models.MetaDataVerbosity)
	if err != nil {
		return err
	}
	if len(*projectChildren) > 0 || len(*departmentChildren) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("department %v cannot be deleted. the department has children", departmentId)
		return fmt.Errorf("department %v cannot be deleted. the department has children", departmentId)
	}
	return nil
}

func validateNotDefault(ctx context.Context, departmentId int64, tenantId string) error {
	department, err := dao.GetDepartmentById(ctx, departmentId, tenantId, models.MetaDataVerbosity, nil)
	if err != nil {
		return err
	}

	if department.Name == services.DefaultDepartmentName {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("department %v cannot be deleted. The department is a default department", departmentId)
		return fmt.Errorf("department %v cannot be deleted. The department is a default department", departmentId)
	}

	return nil
}

func validateProjects(department models.Department, projects *[]models.Project) error {
	for _, project := range *projects {
		failedFields := services.ValidateInheritanceFields(project, department)
		if len(failedFields) > 0 {
			return fmt.Errorf("failed to update department, project %v fields %v conflicts with department fields", project.Name, failedFields)
		}
	}
	return nil
}

func compareStringArrays(arr1, arr2 *pq.StringArray) bool {
	if arr1 == nil {
		return arr2 == nil
	}
	if arr2 == nil {
		return false
	}
	return reflect.DeepEqual(*arr1, *arr2)
}

func compareNodeTypeArrays(arr1, arr2 *[]models.NodeType) bool {
	if arr1 == nil {
		return arr2 == nil
	}
	if arr2 == nil {
		return false
	}
	return reflect.DeepEqual(*arr1, *arr2)
}

func shouldUpdateOverrideFields(old models.OrgUnitData, new models.OrgUnitData) bool {
	return old.InteractiveJobTimeLimit != new.InteractiveJobTimeLimit ||
		old.InteractiveJobPreemptIdleDuration != new.InteractiveJobPreemptIdleDuration ||
		old.TrainingJobMaxIdleDuration != new.TrainingJobMaxIdleDuration ||
		old.InteractiveJobMaxIdleDuration != new.InteractiveJobMaxIdleDuration ||
		old.TrainingJobTimeLimit != new.TrainingJobTimeLimit ||
		!compareNodeTypeArrays(old.TrainingNodeTypes, new.TrainingNodeTypes) ||
		!compareNodeTypeArrays(old.WorkspaceNodeTypes, new.WorkspaceNodeTypes) ||
		!compareStringArrays(old.DefaultNodePoolList, new.DefaultNodePoolList)
}

func (s *DepartmentApiService) enrichDepartmentsRankingWithWorkloadCounts(
	ctx context.Context,
	departments []api.DepartmentGetRanking,
	clusterId string,
	nodePoolName string,
) ([]api.DepartmentGetRanking, error) {
	wrappers := make([]*services.DepartmentRankingWrapper, len(departments))
	for i := range departments {
		wrappers[i] = &services.DepartmentRankingWrapper{DepartmentGetRanking: &departments[i]}
	}

	_, err := services.EnrichOrgUnitsWithWorkloadCounts(
		ctx,
		wrappers,
		"department",
		clusterId,
		nodePoolName,
		s.workloadsClient,
	)

	return departments, err
}
