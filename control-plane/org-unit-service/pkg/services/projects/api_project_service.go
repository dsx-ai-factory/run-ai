/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package projects

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"

	errors2 "github.com/run-ai/runai/control-plane/org-unit-service/pkg/errors"

	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/overtime_metrics_writer"

	"github.com/google/uuid"
	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	auth_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	cluster_openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	telemetry "github.com/run-ai/runai/runai-application-packages/telemetry"
	workloads_client "github.com/run-ai/runai/runai-clients/workloads-client"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

const (
	CreatingPhase   = string(api.PROJECTPHASE_CREATING)
	UpdatingPhase   = string(api.PROJECTPHASE_UPDATING)
	ResourceType    = "projects"
	AuditEntityType = "project"
	ClusterId       = "ClusterId"
	ParentId        = "ParentId"
	NodepoolId      = "NodepoolId"
	ProjectId       = "ProjectId"

	GpuDeservedColumn                    = "gpu_deserved"
	CpuDeservedColumn                    = "cpu_deserved"
	MemoryDeservedColumn                 = "memory_deserved"
	GpuAllocationColumn                  = "allocated_gpu"
	CpuAllocationColumn                  = "allocated_cpu"
	MemoryAllocationColumn               = "allocated_memory"
	GpuAllocationNonPreemptibleColumn    = "allocated_non_preemptible_gpu"
	CpuAllocationNonPreemptibleColumn    = "allocated_non_preemptible_cpu"
	MemoryAllocationNonPreemptibleColumn = "allocated_non_preemptible_memory"
)

type ProjectApiService struct {
	authorizationClient    auth_client.IAuthorizationClient
	projectDeletionService services.ProjectDeletionService
	clusterClient          cluster_service_client.IClusterExternalClient
	workloadsClient        workloads_client.IWorkloadsClient
	modelFieldToDbFieldMap map[string][]string
}

func NewProjectApiService(
	projectDeletionService services.ProjectDeletionService,
	clusterClient cluster_service_client.IClusterExternalClient,
	workloadsClient workloads_client.IWorkloadsClient,
) api.ProjectsAPIServicer {
	return &ProjectApiService{
		authorizationClient:    auth_client.NewAuthorizationClient(),
		projectDeletionService: projectDeletionService,
		modelFieldToDbFieldMap: createFieldsMap(),
		clusterClient:          clusterClient,
		workloadsClient:        workloadsClient,
	}
}

func createFieldsMap() map[string][]string {
	return map[string][]string{
		ClusterId:  {"cluster_id"},
		ParentId:   {"parent_id", "parent_name"},
		NodepoolId: {models.ResourcesNodePoolIdColumn, models.ResourcesNodePoolNameColumn},
		ProjectId:  {"project.id", "project.name"},
	}
}

func (s *ProjectApiService) CountProjects(ctx context.Context, filterBy []string, search string) (api.ImplResponse, error) {
	tenantId, tenantErr := services.ParseTenantId(ctx)
	if tenantErr != nil {
		return services.HandleError(tenantErr, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	filters, filterErr := models.Project{}.FieldsConfiguration().ValidateAndParameterizeFilters(filterBy)
	if filterErr != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy))
		return errResponse, nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	count, dbErr := dao.CountProjects(ctx, *tenantId, permittedScopes, filters, search, models.SlimVerbosity)
	if dbErr != nil {
		errResponse := services.HandleDbError(ctx, dbErr, "Failed to count projects")
		return errResponse, nil
	}

	return api.Response(http.StatusOK, api.CountResponse{Count: count}), nil
}

func (s *ProjectApiService) CreateProject(ctx context.Context, request api.ProjectCreationRequest) (api.ImplResponse, error) {
	// Acquire read lock on cluster - allows multiple Create/Update resources calls concurrently,
	// but blocks when NodePoolUpdated has the write lock
	services.GetClusterLockManager().RLock(request.ClusterId)
	defer services.GetClusterLockManager().RUnlock(request.ClusterId)

	clusterUuid, err := uuid.Parse(request.ClusterId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("failed to parse cluster id %s", request.ClusterId))
		return errResponse, nil
	}
	supported, err := services.IsResourceRankSupported(ctx, request.Resources, clusterUuid, s.clusterClient)
	if !supported || err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "rank value for a project is not supported for clusters with version less than 2.20")
		return errResponse, nil
	}
	nodePools, err := services.GetNodePoolsMetadataMap(ctx, s.clusterClient, request.ClusterId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
		return errResponse, nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	var parentDepartment models.Department
	if request.ParentId == nil {
		defaultDepartment, err := dao.GetDefaultDepartment(ctx, *tenantId, clusterUuid, permittedScopes)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusServiceUnavailable, "get default departments failed")
			return errResponse, nil
		}
		parentDepartment = *defaultDepartment
		defaultDepartmentId := strconv.FormatInt(defaultDepartment.ID, 10)
		request.ParentId = &defaultDepartmentId
	} else {
		intId, err := strconv.Atoi(*request.ParentId)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusBadRequest, "Parent id is not of type int")
			return errResponse, nil
		}
		department, err := dao.GetDepartmentById(ctx, int64(intId), *tenantId, models.SlimVerbosity, permittedScopes)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusServiceUnavailable, "get parent department failed")
			return errResponse, nil
		}
		parentDepartment = *department
	}

	apiErr := s.validateCreateRequest(ctx, request, *tenantId, nodePools)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	project, err := MapCreateRequestToModel(ctx, request, parentDepartment, nodePools)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "failed to map request to model")
		return errResponse, nil
	}
	project.TenantId = *tenantId

	err = validateParent(project, parentDepartment)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "failed to create project, override parent's values failed")
		return errResponse, nil
	}

	apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx, *tenantId, project.OrgUnitData.Resources, *request.ParentId)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	} else {
		project.CreatedBy = *subjectId
	}

	response, err := dao.CreateProject(ctx, &project)
	if err != nil {
		details := ""
		humanizer := errors2.TryCreateDuplicateKeyError(err)
		if humanizer != nil {
			details = humanizer.Humanize()
		}
		errResponse := services.HandleDbErrorWithDetails(ctx, err, "Failed to create project", details)
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
	cacheError := cache.Del[authorization_utils.Descendants](ctx, authorization_utils.GetDescendantsKey(strconv.FormatInt(parentDepartment.ID, 10), string(api.SCOPETYPE_DEPARTMENT)))
	if cacheError != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete cache for department %d: %v", parentDepartment.ID, cacheError)
	}
	return api.Response(http.StatusCreated, MapModelToResponse(ctx, *response, nodePools)), nil
}

func (s *ProjectApiService) DeleteProject(ctx context.Context, projectId string, deletionType api.DeletionTypes, force bool) (api.ImplResponse, error) {
	projectUid, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "delete project failed")
		return errResponse, nil
	}

	err = validateDeleteRequest(ctx)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "delete project request validation failed")
		return errResponse, nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	}

	var deletionTypePtr *api.DeletionTypes
	if deletionType != "" {
		deletionTypePtr = &deletionType
	}
	projectName, departmentId, isSyncDeletion, err := s.projectDeletionService.DeleteProject(ctx, projectUid, *tenantId, subjectId, deletionTypePtr, &force)
	if err != nil {
		if errors.As(err, ptr.To(&services.ClusterDoesNotSupportProjectDeletionErr{})) ||
			errors.As(err, ptr.To(&services.ProjectIsNotEmptyErr{})) ||
			errors.Is(err, &services.ProjectIsAlreadyDeletingErr{}) {
			return api.ImplResponse{
				Code: http.StatusBadRequest,
				Body: err.Error(),
			}, nil
		}

		errResponse := services.HandleDbError(ctx, err, "Failed to delete project")
		return errResponse, nil
	}

	audit.AuditFromCtx(ctx).WithEntityName(projectName).WithEntityType(AuditEntityType).WithEntityId(projectId)

	cacheError := cache.Del[authorization_utils.Descendants](ctx, authorization_utils.GetDescendantsKey(strconv.FormatInt(departmentId, 10), string(api.SCOPETYPE_DEPARTMENT)))
	if cacheError != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete cache for department %d: %v", departmentId, cacheError)
	}
	if isSyncDeletion {
		return api.Response(http.StatusNoContent, nil), nil
	}
	return api.Response(http.StatusAccepted, nil), nil
}

func (s *ProjectApiService) GetProject(ctx context.Context, projectId string) (api.ImplResponse, error) {
	// add when multi-hierarchy will be in and verbosity will be a wuery param
	//verbosityEnum, err := ValidateAndGetVerbosity(verbosity)
	//if err != nil {
	//	errResponse := services.HandleError(err, http.StatusBadRequest, "get project request validation failed")
	//	return errResponse, nil
	//}
	projectUid, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "get project failed")
		return errResponse, nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	project, err := dao.GetProjectById(ctx, projectUid, *tenantId, models.SlimVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get project")
		return errResponse, nil
	}

	nodePools, err := services.GetNodePoolsMetadataMap(ctx, s.clusterClient, project.ClusterId.String())
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
		return errResponse, nil
	}
	return api.Response(http.StatusOK, MapModelToResponse(ctx, *project, nodePools)), nil
}

func (s *ProjectApiService) GetProjectAncestors(ctx context.Context, projectId string) (api.ImplResponse, error) {
	ancestors, err := authorization_utils.GetAncestorsCache(ctx, ResourceType, projectId)
	if err == nil {
		return api.Response(http.StatusOK, ancestors), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Could not find project <%s> ancestors in cache. err: %v, continuing", projectId, err)

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	projectUid, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("get project <%s> ancestors failed", projectId))
		return errResponse, nil
	}

	project, err := dao.GetProjectById(ctx, projectUid, *tenantId, models.MetaDataVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get project ancestors")
		return errResponse, nil
	}

	ancestors = &authorization_utils.Ancestors{
		TenantId:     *tenantId,
		ClusterUuid:  convutils.Ref(project.ClusterId.String()),
		DepartmentId: convutils.Ref(strconv.FormatInt(*project.ParentID, 10)),
	}

	authorization_utils.SetAncestorsCache(ctx, ResourceType, projectId, *ancestors)

	return api.Response(http.StatusOK, *ancestors), nil
}

func (s *ProjectApiService) GetProjects(ctx context.Context, filterBy []string, sortBy api.ProjectFilterSortFields, sortOrder string, offset int32, limit int32, search string) (api.ImplResponse, error) {
	// add when multi-hierarchy will be in and verbosity will be a query param
	//verbosityEnum, err := ValidateAndGetVerbosity(verbosity)
	//if err != nil {
	//	errResponse := services.HandleError(err, http.StatusBadRequest, "get project request validation failed")
	//	return errResponse, nil
	//}
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	filters, filterErr := models.Project{}.FieldsConfiguration().ValidateAndParameterizeFilters(filterBy)
	if filterErr != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy))
		return errResponse, nil
	}
	sortColumn, sortError := models.Project{}.FieldsConfiguration().ValidateSortByField(string(sortBy))
	if sortError != nil {
		errResponse := services.HandleError(sortError, http.StatusBadRequest, fmt.Sprintf("sort %s invalid", sortBy))
		return errResponse, nil
	}
	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}
	projects, err := dao.GetProjects(ctx, *tenantId, filters, sortColumn, sortOrder, offset, limit, search, permittedScopes, false, models.SlimVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get project")
		return errResponse, nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("%d projects read from dao: %#v", len(*projects), *projects)

	var nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal
	if len(*projects) > 0 {
		clusterIds := services.FetchClusterIdsOfProjects(*projects)
		nodePoolsMetadata, err = services.GetNodePoolsMultiCluster(ctx, s.clusterClient, clusterIds)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
			return errResponse, nil
		}
	}

	apiProjects := make([]api.Project, len(*projects))
	for index, project := range *projects {
		apiProject := MapModelToResponse(ctx, project, nodePoolsMetadata)
		apiProjects[index] = apiProject
	}

	response := api.GetProjects200Response{
		Projects: apiProjects,
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("response for get projects is %#v", response)

	recordsCount := int32(len(apiProjects))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *ProjectApiService) GetProjectsMetadata(ctx context.Context, filterBy []string, sortBy string, sortOrder string, offset int32, limit int32, search string) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	filters, filterErr := models.Project{}.FieldsConfiguration().ValidateAndParameterizeFilters(filterBy)
	if filterErr != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy))
		return errResponse, nil
	}
	sortColumn, sortError := models.Project{}.FieldsConfiguration().ValidateSortByField(sortBy)
	if sortError != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("sort %s invalid", sortBy))
		return errResponse, nil
	}
	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	projects, err := dao.GetProjects(ctx, *tenantId, filters, sortColumn, sortOrder, offset, limit, search, permittedScopes, false, models.SlimVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get project")
		return errResponse, nil
	}

	apiProjects := make([]api.ProjectMetadata, len(*projects))
	for index, project := range *projects {
		apiProjects[index] = MapModelToGetMetadataResponse(project)
	}

	response := api.GetProjectsMetadata200Response{
		Projects: apiProjects,
	}

	recordsCount := int32(len(apiProjects))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *ProjectApiService) UpdateProject(ctx context.Context, projectId string, updateProjectRequest api.ProjectUpdateRequest) (api.ImplResponse, error) {
	projectUid, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "update project resources failed")
		return errResponse, nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	// Acquire read lock on cluster - allows multiple Create/Update projects calls concurrently,
	// but blocks when NodePoolUpdated has the write lock
	clusterId, err := dao.GetClusterIdByProject(ctx, projectUid, *tenantId)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get project cluster")
		return errResponse, nil
	}
	services.GetClusterLockManager().RLock(clusterId)
	defer services.GetClusterLockManager().RUnlock(clusterId)

	project, err := dao.GetProjectById(ctx, projectUid, *tenantId, models.SlimVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to update project resources")
		return errResponse, nil
	}

	nodePools, err := services.GetNodePoolsMetadataMap(ctx, s.clusterClient, project.ClusterId.String())
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
		return errResponse, nil
	}

	apiErr := s.validateUpdateRequest(ctx, updateProjectRequest, nodePools)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	clearTrainNodeTypes := project.OrgUnitData.TrainingNodeTypes != nil && len(*project.OrgUnitData.TrainingNodeTypes) > 0
	clearWorkspaceNodeTypes := project.OrgUnitData.WorkspaceNodeTypes != nil && len(*project.OrgUnitData.WorkspaceNodeTypes) > 0

	department, err := dao.GetDepartmentById(ctx, *project.ParentID, project.TenantId, models.SlimVerbosity, nil)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusServiceUnavailable, "get parent department failed")
		return errResponse, nil
	}

	updated, err := MapUpdateRequestToModel(ctx, updateProjectRequest, project, nodePools, department)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "failed to map request to model")
		return errResponse, nil
	}

	err = validateParent(updated, *department)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "failed to update project, override parent's values failed")
		return errResponse, nil
	}

	apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx, *tenantId, updated.OrgUnitData.Resources, *project)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	} else {
		project.UpdatedBy = *subjectId
	}
	response, err := dao.UpdateProject(ctx, &updated, clearTrainNodeTypes, clearWorkspaceNodeTypes)
	if err != nil {
		message := services.EnrichUpdateMessage("Failed to update project", "Project", err)
		errResponse := services.HandleDbError(ctx, err, message)
		return errResponse, nil
	}

	// Enrich node type names on the in-memory model so the response includes them
	allNodeTypes, err := dao.GetNodeTypes(ctx, *tenantId, project.ClusterId, nil)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get node types for enrichment: %v", err)
	} else if allNodeTypes != nil {
		services.EnrichNodeTypeNames(&response.OrgUnitData, services.BuildNodeTypeNameMap(*allNodeTypes))
	}

	audit.AuditFromCtx(ctx).WithEntityName(response.Name).WithEntityType(AuditEntityType).WithEntityId(strconv.FormatInt(response.ID, 10))

	return api.Response(http.StatusOK, MapModelToResponse(ctx, *response, nodePools)), nil
}

func (s *ProjectApiService) UpdateProjectResources(ctx context.Context, projectId string, resources []api.ResourcesNullable) (api.ImplResponse, error) {
	projectUid, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "update project resources failed")
		return errResponse, nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	// Acquire read lock on cluster - allows multiple Create/Update projects calls concurrently,
	// but blocks when NodePoolUpdated has the write lock
	clusterId, err := dao.GetClusterIdByProject(ctx, projectUid, *tenantId)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get project cluster")
		return errResponse, nil
	}
	services.GetClusterLockManager().RLock(clusterId)
	defer services.GetClusterLockManager().RUnlock(clusterId)

	project, err := dao.GetProjectById(ctx, projectUid, *tenantId, models.SlimVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to update project resources")
		return errResponse, nil
	}
	supported, err := services.IsResourceRankSupported(ctx, resources, project.ClusterId, s.clusterClient)
	if !supported || err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("rank value for project is not supported for clusters with version less than %s", services.ResourcePriorityMinVersion))
		return errResponse, nil
	}
	nodePools, err := services.GetNodePoolsMetadataMap(ctx, s.clusterClient, project.ClusterId.String())
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
		return errResponse, nil
	}

	err = services.ValidateResourcesUpdateRequest(ctx, resources, nodePools)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "update project resources validation failed")
		return errResponse, nil
	}

	originalResources := services.CopyResourcesDeservedAndStatus(project.OrgUnitData.Resources)

	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	} else {
		project.UpdatedBy = *subjectId
	}

	err = mapUpdateResourcesRequestToModel(ctx, resources, project, nodePools)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusServiceUnavailable, "failed to map resources to model")
		return errResponse, nil
	}

	apiErr := services.ValidateProjectUpdateNotCausingOverSubscriptionWrapper(ctx, *tenantId,
		project.OrgUnitData.Resources, originalResources, project.ParentID, project.ClusterId)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	_, err = dao.UpdateProject(ctx, project, false, false)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to update project")
		return errResponse, nil
	}

	audit.AuditFromCtx(ctx).WithEntityName(project.Name).WithEntityType(AuditEntityType).WithEntityId(strconv.FormatInt(project.ID, 10))

	return api.Response(http.StatusOK, services.ApiResourcesFromModel(project.OrgUnitData.Resources, nodePools)), nil
}

func (s *ProjectApiService) UpdateProjectStatus(ctx context.Context, projectStatusUpdateRequest api.ProjectStatusUpdateRequest) (api.ImplResponse, error) {
	err := validateStatusRequest(ctx, projectStatusUpdateRequest)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "update project status validation failed")
		return errResponse, nil
	}

	clusterUuid, err := uuid.Parse(projectStatusUpdateRequest.ClusterId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("failed to parse cluster id %s", projectStatusUpdateRequest.ClusterId))
		return errResponse, nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	projectUid, err := strconv.ParseInt(projectStatusUpdateRequest.Id, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "failed to update status")
		return errResponse, nil
	}

	if projectStatusUpdateRequest.Status.Phase == api.PROJECTPHASE_DELETED {
		err = s.projectDeletionService.DeleteProjectSyncFromCluster(ctx, projectUid, *tenantId)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusBadRequest, "failed to delete project")
			return errResponse, nil
		}

		return api.Response(http.StatusOK, nil), nil
	}

	project, err := dao.GetProjectById(ctx, projectUid, *tenantId, models.SlimVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to update project")
		return errResponse, nil
	}

	if project.Phase != nil && *project.Phase == string(api.PROJECTPHASE_DELETING) && projectStatusUpdateRequest.Status.Phase != api.PROJECTPHASE_DELETING {
		return api.Response(http.StatusBadRequest, fmt.Sprintf("Can not update phase of a deleting project to phase %s", projectStatusUpdateRequest.Status.Phase)), nil
	}

	projectStatus, resourcesStatuses := MapApiStatusToModel(projectStatusUpdateRequest.Status, project)
	err = dao.UpdateProjectStatus(ctx, *tenantId, clusterUuid, project, projectStatus, resourcesStatuses)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to update project")
		return errResponse, nil
	}

	return api.Response(http.StatusOK, nil), nil
}

func (s *ProjectApiService) GetProjectsTelemetry(ctx context.Context, telemetryType api.OrgUnitTelemetryType,
	clusterId string, nodepoolName string, departmentId string, projectId string, groupBy []string,
) (api.ImplResponse, error) {
	modelGroupBy, err := telemetry.GetModelGroupBy(groupBy, s.modelFieldToDbFieldMap)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("groupBy %v invalid", groupBy))
		return errResponse, nil
	}
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "tenantId invalid")
		return errResponse, nil
	}
	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}
	timestamp := time.Now()
	var telemetryGroups *[]telemetry.TelemetryGroupValue
	var daoErr error
	switch telemetryType {
	case api.ORGUNITTELEMETRYTYPE_GPU_QUOTA:
		telemetryGroups, daoErr = dao.GetProjectResourcesTelemetry(ctx, GpuDeservedColumn, *tenantId, permittedScopes, clusterId, nodepoolName, departmentId, projectId, modelGroupBy)
	case api.ORGUNITTELEMETRYTYPE_CPU_QUOTA:
		telemetryGroups, daoErr = dao.GetProjectResourcesTelemetry(ctx, CpuDeservedColumn, *tenantId, permittedScopes, clusterId, nodepoolName, departmentId, projectId, modelGroupBy)
	case api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA:
		telemetryGroups, daoErr = dao.GetProjectResourcesTelemetry(ctx, MemoryDeservedColumn, *tenantId, permittedScopes, clusterId, nodepoolName, departmentId, projectId, modelGroupBy)
	case api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION:
		telemetryGroups, daoErr = dao.GetProjectResourcesTelemetry(ctx, GpuAllocationColumn, *tenantId, permittedScopes, clusterId, nodepoolName, departmentId, projectId, modelGroupBy)
	case api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION:
		telemetryGroups, daoErr = dao.GetProjectResourcesTelemetry(ctx, CpuAllocationColumn, *tenantId, permittedScopes, clusterId, nodepoolName, departmentId, projectId, modelGroupBy)
	case api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION:
		telemetryGroups, daoErr = dao.GetProjectResourcesTelemetry(ctx, MemoryAllocationColumn, *tenantId, permittedScopes, clusterId, nodepoolName, departmentId, projectId, modelGroupBy)
	case api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE:
		telemetryGroups, daoErr = dao.GetProjectResourcesTelemetry(ctx, GpuAllocationNonPreemptibleColumn, *tenantId, permittedScopes, clusterId, nodepoolName, departmentId, projectId, modelGroupBy)
	case api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION_NON_PREEMPTIBLE:
		telemetryGroups, daoErr = dao.GetProjectResourcesTelemetry(ctx, CpuAllocationNonPreemptibleColumn, *tenantId, permittedScopes, clusterId, nodepoolName, departmentId, projectId, modelGroupBy)
	case api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION_NON_PREEMPTIBLE:
		telemetryGroups, daoErr = dao.GetProjectResourcesTelemetry(ctx, MemoryAllocationNonPreemptibleColumn, *tenantId, permittedScopes, clusterId, nodepoolName, departmentId, projectId, modelGroupBy)
	default:
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("telemetry type %v invalid", telemetryType))
		return errResponse, nil
	}
	if daoErr != nil {
		errResponse := services.HandleDbError(ctx, daoErr, fmt.Sprintf("Failed to get project telemetry %v", telemetryType))
		return errResponse, nil
	}
	telemetryGroupsApi := services.TelemetryGroupsToApi(string(telemetryType), timestamp, *telemetryGroups, groupBy)
	return api.Response(http.StatusOK, telemetryGroupsApi), nil
}

func (s *ProjectApiService) GetProjectMetrics(
	ctx context.Context,
	projectId string,
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

	projectUid, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "get project metrics failed")
		return errResponse, nil
	}
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	project, err := dao.GetProjectById(ctx, projectUid, *tenantId, models.MetaDataVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "get project metrics failed")
		return errResponse, nil
	}

	err = services.ValidateMinClusterVersionForMetrics(ctx, project.ClusterId, metricTypes, s.clusterClient)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, err.Error()), nil
	}

	return services.GetOrgUnitMetrics(ctx, project.ClusterId.String(), &project.Name, project.ParentName, &nodePoolName, metricTypes, start, end, numberOfSamples, groupBy)
}

func (s *ProjectApiService) GetProjectsRanking(ctx context.Context, metricType api.RankingMetricType,
	orderType api.RankingOrderType, clusterId string, nodePoolName string, departmentId string,
) (api.ImplResponse, error) {
	const NumberOfProjects = 15
	var results []api.ProjectGetRanking
	var err error

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	if err := services.ValidateClusterBelongsToTenant(ctx, clusterId, *tenantId, s.clusterClient); err != nil {
		if errors.Is(err, services.ErrClusterTenantMismatch) {
			return services.HandleError(err, http.StatusUnauthorized, "cluster does not belong to tenant"), nil
		}
		return services.HandleError(err, http.StatusBadGateway, "failed to validate cluster"), nil
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
		results, err = dao.GetAllocationRatioRankingByNodepool(ctx, *tenantId, clusterId, nodePoolName, permittedScopes, orderType, NumberOfProjects, false, departmentId)
		if err != nil {
			return services.HandleDbError(ctx, err, "Failed to get project ranking"), nil
		}
	} else if nodePoolName == "" || metricType == api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO {
		columnNames, innerErr := dao.GetColumnNamesForRanking(metricType, false)
		if innerErr != nil {
			return api.Response(http.StatusBadRequest, innerErr.Error()), nil
		}
		orderBy, innerErr := dao.GetOrderBy(metricType, false)
		if innerErr != nil {
			return api.Response(http.StatusBadRequest, innerErr.Error()), nil
		}

		results, err = dao.GetProjectsRanking(ctx, *tenantId, clusterId, permittedScopes, columnNames, orderBy, orderType, NumberOfProjects, departmentId)
		if err != nil {
			return services.HandleDbError(ctx, err, "Failed to get project ranking"), nil
		}
	} else {
		clusterFilter := []string{clusterId}
		var departmentsNames []string

		if permittedScopes != nil {
			departmentsNames, err = dao.GetDepartmentNamesByIds(ctx, permittedScopes.Departments)
			if err != nil {
				return services.HandleDbError(ctx, err, "Failed to get project ranking"), nil
			}
		}

		results, err = overtime_metrics_writer.QueryProjectsOvertimeRankingMetrics(ctx, *tenantId, metricType, orderType, clusterFilter, nodePoolName, departmentsNames)
	}

	if err != nil {
		return api.Response(http.StatusInternalServerError, err.Error()), nil
	}

	if len(results) == 0 {
		results = []api.ProjectGetRanking{}
	}

	enrichedResults, enrichErr := s.enrichProjectsRankingWithWorkloadCounts(ctx, results, clusterId, nodePoolName)
	if enrichErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to enrich projects ranking with workload counts: %v", enrichErr)
	} else {
		results = enrichedResults
	}

	response := api.GetProjectsRanking200Response{
		Projects: results,
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *ProjectApiService) patchProjectResources(
	ctx context.Context,
	projectId string,
	resources []api.ResourcesNullable,
	tenantId string,
	subjectId *string,
	nodePoolsPerCluster *sync.Map,
	permittedScopes *openapi.PermittedScopes,
) (*[]api.Resources, string, *api.Error) {
	projectUid, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		return nil, "", api.NewError(http.StatusBadRequest, "failed to parse project id", err)
	}
	project, err := dao.GetProjectByIdWithPermittedScopes(ctx, projectUid, tenantId, permittedScopes)
	if err != nil {
		return nil, "", api.NewError(http.StatusInternalServerError, "failed to retrieve project", err)
	}

	nodePools, apiErr := services.GetNodePoolsForClusterFromServiceOrCache(ctx, s.clusterClient, project.ClusterId.String(), nodePoolsPerCluster)
	if apiErr != nil {
		return nil, "", apiErr
	}

	err = services.ValidateResourcesPatchRequest(ctx, resources, nodePools)
	if err != nil {
		return nil, "", api.NewError(http.StatusBadRequest, fmt.Sprintf("patch project resources validation failed: %s", err.Error()), err)
	}

	originalResources := services.CopyResourcesDeservedAndStatus(project.OrgUnitData.Resources)
	if subjectId != nil {
		project.UpdatedBy = *subjectId
	}

	err = mapPatchResourcesRequestToModel(ctx, resources, project, nodePools)
	if err != nil {
		return nil, "", api.NewError(http.StatusServiceUnavailable, "failed to map resources to model", err)
	}

	apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionWrapper(ctx, tenantId,
		project.OrgUnitData.Resources, originalResources, project.ParentID, project.ClusterId)
	if apiErr != nil {
		return nil, "", apiErr
	}

	_, err = dao.UpdateProject(ctx, project, false, false)
	if err != nil {
		return nil, "", api.NewError(http.StatusInternalServerError, "Failed to update project", err)
	}

	updated := services.ApiResourcesFromModel(project.OrgUnitData.Resources, nodePools)
	return &updated, project.Name, nil
}

func (s *ProjectApiService) PatchProjectResources(ctx context.Context, projectId string, resources []api.ResourcesNullable) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	}
	nodePoolsCache := sync.Map{}

	updated, projectName, apiErr := s.patchProjectResources(ctx, projectId, resources, *tenantId, subjectId, &nodePoolsCache, nil)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	audit.AuditFromCtx(ctx).WithEntityName(projectName).WithEntityType(AuditEntityType).WithEntityId(projectId)
	return api.Response(http.StatusOK, *updated), nil
}

func (s *ProjectApiService) BatchPatchProjectNodePoolResources(ctx context.Context, request api.BatchPatchNodePoolResourcesRequest) (api.ImplResponse, error) {
	apiErr := services.ValidateBatchPatchNodePoolResourcesRequest(request)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	succeeded, failed, apiErr := s.bulkPatchProjectResources(ctx, request, *tenantId)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	response := api.BatchPatchProjectResourcesResponse{
		Id:              uuid.New().String(),
		Succeeded:       succeeded,
		Failed:          failed,
		TotalRequested:  int32(len(request.PatchRequests)),
		TotalSuccessful: int32(len(succeeded)),
		TotalFailed:     int32(len(failed)),
	}
	if len(succeeded) == 0 {
		return api.Response(http.StatusInternalServerError, response), nil
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *ProjectApiService) bulkPatchProjectResources(
	ctx context.Context,
	request api.BatchPatchNodePoolResourcesRequest,
	tenantId string,
) ([]api.ProjectNodePoolResources, []api.FailedNodePoolResourcesPatchRequest, *api.Error) {
	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	// Step 1: Validate IDs, batch fetch projects and index by ID
	patchRequestsWithValidIdPerProject, failed, dbProjectById, apiErr := validateAndFetchProjectsFromDb(ctx, tenantId, permittedScopes, request)
	if apiErr != nil {
		return nil, nil, apiErr
	}

	// Step 2: Group by department (and cluster)
	departmentIdToPatchGroup, groupFailed := buildDepartmentPatchGroups(dbProjectById, patchRequestsWithValidIdPerProject)
	failed = append(failed, groupFailed...)

	// Step 3: Process per-department using the department-aware patcher
	subjectId, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id in bulkPatchProjectResources: %v", err)
	}
	nodePoolsCache := sync.Map{}
	stagedGroupsPerDepartment := make([][]dbProjectWithPatchRequests, 0, len(departmentIdToPatchGroup))
	for departmentId, group := range departmentIdToPatchGroup {
		stagedGroup, validationFailed, validateErr := prepareAndValidatePatchResourcesForDepartment(ctx, tenantId, departmentId, group, &nodePoolsCache, s.clusterClient)
		if validateErr != nil {
			// If department-level validation failed, mark all group requests as failed with the validation error
			for _, pp := range group.patches {
				appendFailedPatchRequests(pp.patchRequests, validateErr, &failed)
			}
			continue
		}
		stagedGroupsPerDepartment = append(stagedGroupsPerDepartment, stagedGroup)
		failed = append(failed, validationFailed...)
	}

	// Step 4: Execute project updates, for each department in parallel
	succeededCommit, failedCommit := commitProjectPatchesPerDepartmentInParallel(ctx, tenantId, stagedGroupsPerDepartment, subjectId, &nodePoolsCache, s.clusterClient)
	return succeededCommit, append(failed, failedCommit...), nil
}

func (s *ProjectApiService) GetSyncProjects(ctx context.Context, clusterId string, updatedAfter time.Time, offset int32, limit int32, excludePermissions bool) (api.ImplResponse, error) {
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

	projects, err := dao.GetSyncProjects(ctx, *tenantId, clusterUuid, updatedAfter, offset, limit, permittedScopes)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get project")
		return errResponse, nil
	}
	var accessRulesPerId map[string]*[]openapi.AccessRule
	if !excludePermissions {
		accessRulesPerId, err = services.GetProjectsPermissions(ctx, s.authorizationClient, clusterId, projects)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusServiceUnavailable, "Failed to get project permissions")
			return errResponse, nil
		}
	}

	var nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal
	if len(projects) > 0 {
		nodePoolsMetadata, err = services.GetNodePoolsMetadataMap(ctx, s.clusterClient, clusterId)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
			return errResponse, nil
		}
	}

	apiProjects := make([]api.ProjectSync, len(projects))
	for index, project := range projects {
		apiProjects[index] = MapModelToSyncResponse(*project, nodePoolsMetadata, accessRulesPerId[strconv.FormatInt(project.ID, 10)])
	}

	response := api.GetSyncProjects200Response{
		Projects: apiProjects,
	}

	recordsCount := int32(len(apiProjects))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *ProjectApiService) SyncProjects(ctx context.Context, syncRequest api.SyncProjectsRequest) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	clusterUuid, err := uuid.Parse(syncRequest.ClusterId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, fmt.Sprintf("failed to parse cluster id %s", syncRequest.ClusterId))
		return errResponse, nil
	}
	err = s.projectDeletionService.DeleteProjectsSyncFromCluster(ctx, *tenantId, clusterUuid, syncRequest.Projects)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, fmt.Sprintf("failed to delete missing projects of cluster id %s", syncRequest.ClusterId))
		return errResponse, nil
	}

	return api.Response(http.StatusOK, nil), nil
}

func (s *ProjectApiService) ProjectsResetResources(ctx context.Context, request api.ProjectsResourcesResetRequest) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	err = services.ValidateResetResourcesRequest(ctx, *tenantId, request.ClusterId, request.ParentId, request.NodePoolIds, s.clusterClient)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, err.Error())
		return errResponse, nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	apiErr := validateResetResourcesPermitted(ctx, *tenantId, request.ClusterId, request.ParentId, permittedScopes)
	if apiErr != nil {
		return services.HandleApiError(*apiErr), nil
	}

	dbErr := dao.ResetResources(ctx, *tenantId, request.ClusterId, request.ParentId, request.NodePoolIds, request.Action, true)
	if dbErr != nil {
		errResponse := services.HandleDbError(ctx, dbErr, "failed to reset projects resources")
		return errResponse, nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (s *ProjectApiService) GetProjectNodePoolResources(ctx context.Context, filterBy []string,
	sortBy api.ProjectNodePoolResourcesFilterSortFields, sortOrder string, offset int32, limit int32, search string,
) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	filters, filterErr := models.ProjectNodePoolResources{}.FilterByFieldsConfiguration().ValidateAndParameterizeFilters(filterBy)
	if filterErr != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy))
		return errResponse, nil
	}
	sortColumn, sortError := models.ProjectNodePoolResources{}.SortByFieldsConfiguration().ValidateSortByField(string(sortBy))
	if sortError != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("sort %s invalid", sortBy))
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

	nodePoolResources, err := dao.GetProjectNodePoolResources(ctx, *tenantId, filters, sortColumn, sortOrder, offset, limit, permittedScopes, search)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get project nodePool resources")
		return errResponse, nil
	}
	var nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal
	if len(*nodePoolResources) > 0 {
		clusterIds := services.FetchClusterIdsOfProjectNodePoolResources(*nodePoolResources)
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
	apiResources := []api.ProjectNodePoolResources{}
	for _, npResources := range *nodePoolResources {
		apiResources = append(apiResources, *services.MapModelProjectNodePoolResourcesToApi(npResources, nodePoolsMetadata))
	}
	response := api.GetProjectNodePoolResources200Response{
		NodePoolResources: apiResources,
	}
	recordsCount := int32(len(apiResources))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *ProjectApiService) CountProjectNodePoolResources(ctx context.Context, filterBy []string, search string) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	filters, filterErr := models.ProjectNodePoolResources{}.FilterByFieldsConfiguration().ValidateAndParameterizeFilters(filterBy)
	if filterErr != nil {
		errResponse := services.HandleError(filterErr, http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy))
		return errResponse, nil
	}
	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	nodePoolResourcesCount, err := dao.CountProjectNodePoolResources(ctx, *tenantId, filters, search, permittedScopes)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get project nodePool resources")
		return errResponse, nil
	}

	return api.Response(http.StatusOK, api.CountResponse{Count: *nodePoolResourcesCount}), nil
}

func (s *ProjectApiService) SumProjectNodePoolResources(ctx context.Context, nodePoolId string, filterBy []string) (api.ImplResponse, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	filters, filterErr := models.ProjectNodePoolResources{}.FilterByFieldsConfigurationForSum().ValidateAndParameterizeFilters(filterBy)
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

	gpuWeightSum, err := dao.SumProjectNodePoolResourcesGpuWeights(ctx, *tenantId, nodePoolId, filters, capabilities.AdjustableWeights, permittedScopes)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get project nodePool resources")
		return errResponse, nil
	}

	return api.Response(http.StatusOK, api.SumResponse{WeightsSum: *gpuWeightSum}), nil
}

func (s *ProjectApiService) validateCreateRequest(ctx context.Context, request api.ProjectCreationRequest,
	tenantId string, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal,
) *api.Error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating project creation request")
	err := services.ValidateOrgName(request.Name, "project")
	if err != nil {
		return api.NewError(http.StatusBadRequest, "failed to create project", err)
	}
	err = services.ValidateResources(ctx, request.Resources, nodePoolsMetadata)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("failed to create project: %s", err.Error()), err)
	}
	err = services.ValidateNodePools(ctx, request.Resources)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("failed to create project: %s", err.Error()), err)
	}
	parent, err := services.ValidateParent(ctx, request.ParentId, tenantId, request.ClusterId)
	if err != nil {
		return api.NewError(http.StatusBadRequest, "failed to create project", err)
	}
	err = services.ValidateJobsConfigurations(parent, request.SchedulingRules)
	if err != nil {
		return api.NewError(http.StatusBadRequest, "failed to create project", err)
	}

	return nil
}

func validateDeleteRequest(ctx context.Context) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating project deletion request")
	return nil
}

func (s *ProjectApiService) validateUpdateRequest(ctx context.Context, request api.ProjectUpdateRequest,
	nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal,
) *api.Error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating project update request")
	err := services.ValidateResources(ctx, request.Resources, nodePoolsMetadata)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("update project request validation failed: %s", err.Error()), err)
	}
	err = services.ValidateNodePools(ctx, request.Resources)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("update project request validation failed: %s", err.Error()), err)
	}

	return nil
}

func validateStatusRequest(ctx context.Context, request api.ProjectStatusUpdateRequest) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating project update status request")
	if request.ClusterId == "" {
		return errors.New("clusterId should be defined")
	}
	if request.Id == "" {
		return errors.New("id should be defined")
	}
	if request.Status.Phase == "" {
		return errors.New("status.phase should be defined")
	}
	return nil
}

func validateParent(project models.Project, parentDepartment models.Department) error {
	failedFields := services.ValidateInheritanceFields(project, parentDepartment)
	if len(failedFields) > 0 {
		return fmt.Errorf("failed to create project, fields %v cannot override parent's values", failedFields)
	}
	return nil
}

func (s *ProjectApiService) enrichProjectsRankingWithWorkloadCounts(
	ctx context.Context,
	projects []api.ProjectGetRanking,
	clusterId string,
	nodePoolName string,
) ([]api.ProjectGetRanking, error) {
	wrappers := make([]*services.ProjectRankingWrapper, len(projects))
	for i := range projects {
		wrappers[i] = &services.ProjectRankingWrapper{ProjectGetRanking: &projects[i]}
	}

	_, err := services.EnrichOrgUnitsWithWorkloadCounts(
		ctx,
		wrappers,
		"project",
		clusterId,
		nodePoolName,
		s.workloadsClient,
	)

	return projects, err
}
