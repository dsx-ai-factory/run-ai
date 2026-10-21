/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/lib/pq"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_integration_crud"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"k8s.io/utils/ptr"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	telemetry_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/telemetry"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/topology"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	orgUnitClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/common"
	config "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

const (
	ControlPlaneCreatingMessage = "The workload is being created in the control plane"
	WorkloadsResourceType       = "workloads"
	ProjectsResourceType        = "projects"
	ClusterId                   = "ClusterId"
	DepartmentId                = "DepartmentId"
	ProjectId                   = "ProjectId"
	Type                        = "Type"
	CurrentNodepools            = "CurrentNodepools"
	RequestedNodepools          = "RequestedNodepools"
	Phase                       = "Phase"
	Category                    = "Category"
	Source                      = "Source"
	Bucket                      = "Bucket"
)

var supportedWorkloadsMetricsVersion = withVLLMMetricsVersion(map[string]int{
	string(api.WORKLOADMETRICTYPE_NVLINK_BANDWIDTH_TOTAL): checker.MinVersionNvlinkMetrics,
})

type WorkloadsService struct {
	ModelFieldToDbFieldMap map[string][]string
	WorkloadsDao           *dao.WorkloadsDao
}

func (s *WorkloadsService) InternalPatchWorkloadPhase(ctx context.Context, workloadId string, fields api.WorkloadPatchFields) (api.ImplResponse, error) {
	return s.PatchWorkloadPhase(ctx, workloadId, fields)
}

func NewWorkloadsService() api.WorkloadsAPIServicer {
	return &WorkloadsService{
		ModelFieldToDbFieldMap: createFieldsMap(),
		WorkloadsDao:           dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]()),
	}
}

func createFieldsMap() map[string][]string {
	return map[string][]string{
		ClusterId:          {"cluster_id"},
		DepartmentId:       {"department_id", "department_name"},
		ProjectId:          {"project_id", "project_name"},
		Type:               {"type"},
		CurrentNodepools:   {"current_nodepools"},
		RequestedNodepools: {"unnest(requested_nodepools)"},
		Phase:              {"phase"},
		Category:           {"category"},
		Source:             {"source"},
	}
}

func isPendingOnly(phases []api.Phase) bool {
	return len(phases) == 1 && phases[0] == api.PHASE_PENDING
}

func containsGroupBy(groupBy []string, key string) bool {
	for _, g := range groupBy {
		if g == key {
			return true
		}
	}
	return false
}

// validateNodepoolGroupByPhase validates that nodepool groupBy options match the phase filter:
// - RequestedNodepools requires phases=[Pending] (pending workloads have no current nodepool)
// - CurrentNodepools is not allowed with phases=[Pending] (current_nodepools is empty for pending)
func validateNodepoolGroupByPhase(groupBy []string, phases []api.Phase) error {
	pending := isPendingOnly(phases)
	if containsGroupBy(groupBy, RequestedNodepools) && !pending {
		return errors.New("RequestedNodepools groupBy is only supported when filtering by phase=Pending")
	}
	if containsGroupBy(groupBy, CurrentNodepools) && pending {
		return errors.New("CurrentNodepools groupBy is not supported when filtering by phase=Pending, use RequestedNodepools instead")
	}
	return nil
}

func validateNodepoolFilterGroups(clusterFilter string, nodepoolFilter string, groups []string, clusterGroupName string, nodepoolGroupName string) error {
	if clusterFilter == "" {
		if nodepoolFilter != "" {
			return fmt.Errorf("filter of nodepools %s without cluster filter is not allowed", nodepoolFilter)
		}
		firstGroup := ""
		if len(groups) > 0 {
			firstGroup = groups[0]
		}
		secondGroup := ""
		if len(groups) > 1 {
			secondGroup = groups[1]
		}
		if (firstGroup == nodepoolGroupName && secondGroup != clusterGroupName) ||
			(secondGroup == nodepoolGroupName && firstGroup != clusterGroupName) {
			return fmt.Errorf("group by nodepools %s without cluster filter or cluster groups is not allowed", nodepoolGroupName)
		}

	}
	return nil
}

func (s *WorkloadsService) GetWorkloadsTelemetry(ctx context.Context, telemetryType api.WorkloadTelemetryType,
	clusterId string, nodepoolName string, departmentId string, projectId string, groupBy []string, phases []api.Phase,
	filterBy []string,
) (api.ImplResponse, error) {
	permittedScopes, scopesErr := authorization_client.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}
	filters, filterErr := queryapi.NewParser(models.WorkloadView{}.TelemetryFilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}
	return s.getWorkloadsTelemetry(ctx, telemetryType, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, phases, filters)
}

func (s *WorkloadsService) GetInternalWorkloadsTelemetry(ctx context.Context, telemetryType api.WorkloadTelemetryType,
	clusterId string, nodepoolName string, departmentId string, projectId string, groupBy []string, phases []api.Phase,
	filterBy []string,
) (api.ImplResponse, error) {
	filters, filterErr := queryapi.NewParser(models.WorkloadView{}.TelemetryFilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}
	return s.getWorkloadsTelemetry(ctx, telemetryType, clusterId, nodepoolName, departmentId, projectId, nil, groupBy, phases, filters)
}

func (s *WorkloadsService) getWorkloadsTelemetry(ctx context.Context, telemetryType api.WorkloadTelemetryType,
	clusterId string, nodepoolName string, departmentId string, projectId string,
	permittedScopes *authorization_api.PermittedScopes, groupBy []string, phases []api.Phase,
	filterBy []queryapi.Filter,
) (api.ImplResponse, error) {
	err := validateNodepoolFilterGroups(clusterId, nodepoolName, groupBy, ClusterId, CurrentNodepools)
	if err != nil {
		return ErrorToApiError(err, http.StatusBadRequest), nil
	}
	if err := validateNodepoolGroupByPhase(groupBy, phases); err != nil {
		return ErrorToApiError(err, http.StatusBadRequest), nil
	}
	modelGroupBy, err := telemetry_utils.GetModelGroupBy(groupBy, s.ModelFieldToDbFieldMap)
	if err != nil {
		return ErrorToApiError(err, http.StatusBadRequest), nil
	}

	for _, phase := range phases {
		if !phase.IsValid() {
			apiErr := api.NewError(http.StatusBadRequest, fmt.Sprintf("Invalid phase: %s", string(phase)), nil)
			return NewErrorResponse(apiErr), nil
		}
	}

	// Validate groupBy limit for pending time distribution (bucket is always the first group)
	if telemetryType == api.WORKLOADTELEMETRYTYPE_PENDING_TIME_DISTRIBUTION && len(groupBy) > 1 {
		apiErr := api.NewError(http.StatusBadRequest,
			"Pending time distribution telemetry supports at most 1 groupBy parameter.", nil)
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	var telemetryGroups *[]telemetry_utils.TelemetryGroupValue
	var dbErr *api.Error

	timestamp := time.Now()
	switch telemetryType {
	case api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT:
		telemetryGroups, dbErr = s.WorkloadsDao.DbGetWorkloadCount(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, modelGroupBy, phases, filterBy)

	case api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION:
		telemetryGroups, dbErr = s.WorkloadsDao.DbGetGpuAllocation(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, modelGroupBy, phases, filterBy)

	case api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS:
		telemetryGroups, dbErr = s.WorkloadsDao.DbGetIdleAllocatedGpus(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, modelGroupBy, phases, filterBy)

	case api.WORKLOADTELEMETRYTYPE_PENDING_TIME_DISTRIBUTION:
		telemetryGroups, dbErr = s.WorkloadsDao.DbGetPendingTimeDistribution(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, modelGroupBy, filterBy)

	// case AllocationRatioTelemetryType:

	default:
		supportedTypes := []string{
			string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT),
			string(api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION),
			string(api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS),
			string(api.WORKLOADTELEMETRYTYPE_PENDING_TIME_DISTRIBUTION),
		}

		// Create a string with enums separated by commas
		apiErr = api.NewError(http.StatusBadRequest,
			fmt.Sprintf("telemetry type %s is invalid. supported types are: %s", string(telemetryType), strings.Join(supportedTypes, ", ")), nil)
		return NewErrorResponse(apiErr), nil
	}

	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	// For PENDING_TIME_DISTRIBUTION, "bucket" is always the first group
	effectiveGroupBy := groupBy
	if telemetryType == api.WORKLOADTELEMETRYTYPE_PENDING_TIME_DISTRIBUTION {
		effectiveGroupBy = append([]string{Bucket}, groupBy...)
	}

	telemetryGroupsApi := TelemetryGroupsToApi(string(telemetryType), timestamp, *telemetryGroups, effectiveGroupBy)
	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		data := GetTelemetryCsv(telemetryGroupsApi)
		response := runaihttp.CsvFile{Data: data}
		loc, err := time.LoadLocation("UTC")
		if err == nil {
			response.FileName = fmt.Sprintf("%s-%s.csv", string(telemetryType), telemetryGroupsApi.Timestamp.In(loc).Format("2006-01-02_15:04:05"))
		} else {
			response.FileName = fmt.Sprintf("%s.csv", string(telemetryType))
		}
		return api.Response(http.StatusOK, response), nil
	}
	return api.Response(http.StatusOK, telemetryGroupsApi), nil
}

// DeleteWorkload - Delete the workload as it does not exist in the cluster anymore.
func (s *WorkloadsService) DeleteWorkload(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting workload <%s> delete request in tenant <%d>", workloadId, tenantId)

	id, parseErr := uuid.Parse(workloadId)
	if parseErr != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("workload Id must be a valid uuid. <%s>", workloadId), parseErr)), nil
	}

	dbErr := s.WorkloadsDao.DbDeleteWorkload(ctx, tenantId, id)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete workload <%s>. Err: %+v", workloadId, dbErr)
		return NewErrorResponse(dbErr), nil
	}

	if snapshotErr := dao_crud.DbDeleteWorkloadV2Snapshot(ctx, tenantId, workloadId, nil); snapshotErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete V2 workload snapshot for workload %s: %v", workloadId, snapshotErr.Message)
	}

	dbErr = services_integration_crud.UpdateCrudOnDeletedWorkload(ctx, id)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete workload <%s> from workloads_crud. Err: %+v", workloadId, dbErr)
		//	Not returning error since this is not breaking error. the workload is already deleted from the workloads_view
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully deleted workload <%s> in tenant <%d>", workloadId, tenantId)
	return NoContentResponse(), nil
}

// GetWorkload - Get the workload data.
func (s *WorkloadsService) GetWorkload(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	_, parseErr := uuid.Parse(workloadId)
	if parseErr != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("workload Id must be a valid uuid. <%s>", workloadId), parseErr)), nil
	}
	workload, apiErr := dao.GetWorkloadViewFromDb(ctx, workloadId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	apiWorkload := WorkloadModelToWorkloadDetailedApi(ctx, *workload)

	return api.Response(http.StatusOK, apiWorkload), nil
}

func (s *WorkloadsService) CountWorkloads(ctx context.Context, deleted bool, filterBy []string, search string, precise bool) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	countFromTable := models.Workload{}.AreFilterableFromTable(filterBy)

	filters, filterErr := queryapi.NewParser(models.WorkloadView{}.FilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	count, dbErr := s.WorkloadsDao.DbCountWorkloads(ctx, tenantId, permittedScopes, deleted, filters, search, precise, countFromTable)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	response := api.CountWorkloads200Response{
		Count: count,
	}

	return api.Response(http.StatusOK, response), nil
}

// InternalEnrichOrgUnitRanking - Enrich org unit ranking items with workload counts (Internal API)
// Running workloads: filters by currentNodepools (where workload is running)
// Pending workloads: filters by requestedNodepools (only single-nodepool requests)
func (s *WorkloadsService) InternalEnrichOrgUnitRanking(ctx context.Context, internalEnrichOrgUnitRankingRequest api.InternalEnrichOrgUnitRankingRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	if len(internalEnrichOrgUnitRankingRequest.Ids) == 0 {
		return api.Response(http.StatusOK, api.InternalEnrichOrgUnitRankingResponse{
			Items: []api.OrgUnitRankingItem{},
		}), nil
	}

	var nodePoolPtr *string
	if internalEnrichOrgUnitRankingRequest.NodePoolName != nil && *internalEnrichOrgUnitRankingRequest.NodePoolName != "" {
		nodePoolPtr = internalEnrichOrgUnitRankingRequest.NodePoolName
	}

	countsByPhase, err := s.WorkloadsDao.DbCountWorkloadsForRanking(
		ctx,
		tenantId,
		internalEnrichOrgUnitRankingRequest.ClusterId,
		internalEnrichOrgUnitRankingRequest.Ids,
		internalEnrichOrgUnitRankingRequest.OrgUnitType,
		nodePoolPtr,
	)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"Failed to count workloads for ranking %s in cluster %s: %v",
			internalEnrichOrgUnitRankingRequest.OrgUnitType,
			internalEnrichOrgUnitRankingRequest.ClusterId,
			err,
		)
		return NewErrorResponse(err), nil
	}

	// Build enriched items
	enrichedItems := make([]api.OrgUnitRankingItem, 0, len(internalEnrichOrgUnitRankingRequest.Ids))
	for _, id := range internalEnrichOrgUnitRankingRequest.Ids {
		var runningCount, pendingCount int64
		if counts, exists := countsByPhase[id]; exists {
			runningCount = counts.RunningCount
			pendingCount = counts.PendingCount
		}

		enrichedItem := api.OrgUnitRankingItem{
			Id:                    id,
			RunningWorkloadsCount: ptr.To(int32(runningCount)),
			PendingWorkloadsCount: ptr.To(int32(pendingCount)),
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf(
			"Enriched %s %s: running=%d, pending=%d",
			internalEnrichOrgUnitRankingRequest.OrgUnitType, id, runningCount, pendingCount,
		)

		enrichedItems = append(enrichedItems, enrichedItem)
	}

	response := api.InternalEnrichOrgUnitRankingResponse{
		Items: enrichedItems,
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Enriched %d org unit ranking items with workload counts for tenant %d", len(enrichedItems), tenantId)

	return api.Response(http.StatusOK, response), nil
}

// GetWorkloads - List workloads. [Experimental]
func (s *WorkloadsService) GetWorkloads(ctx context.Context, deleted bool, offset int32, limit int32, sortOrder string, sortBy string,
	filterBy []string, search string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	limit = setDefaultLimit(limit)

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	filters, filterErr := queryapi.NewParser(models.WorkloadView{}.FilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	sortColumn, sortError := queryapi.NewParser(models.WorkloadView{}.SortFieldsToColumn()).GetMapping(sortBy)
	if sortError != nil {
		return ErrorToApiError(sortError, http.StatusBadRequest), nil
	}

	workloads, dbErr := s.WorkloadsDao.DbGetWorkloads(ctx, tenantId, permittedScopes, deleted, offset, limit, sortOrder, sortColumn, filters, search)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	clusterConfigs := map[string]*checker.ClusterConfiguration{}
	apiWorkloads := make([]api.Workload, 0, len(*workloads))
	for _, workload := range *workloads {
		clusterId := workload.ClusterID.String()
		clusterConfig, found := clusterConfigs[clusterId]
		if !found {
			clusterConfig = getClusterConfiguration(ctx, clusterId)
			clusterConfigs[clusterId] = clusterConfig
		}
		apiWorkloads = append(apiWorkloads, WorkloadModelToApi(ctx, workload, clusterConfig))
	}

	recordsCount := int32(len(apiWorkloads))

	response := api.GetWorkloads200Response{
		Workloads: apiWorkloads,
	}
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

// GetWorkloadsSlim - List workloads, slim. Serves GET /api/v1/workloads/slim from the workloads table, sorted by id
// ascending.
func (s *WorkloadsService) GetWorkloadsSlim(ctx context.Context, offset int32, limit int32, filterBy []string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	limit = setDefaultLimit(limit)

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	filters, filterErr := queryapi.NewParser(models.WorkloadSlim{}.FilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	workloads, dbErr := s.WorkloadsDao.DbGetWorkloadsSlim(ctx, tenantId, permittedScopes, offset, limit, filters)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	response := api.GetWorkloadsSlimResponse{
		Workloads: make([]api.WorkloadSlim, 0, len(*workloads)),
	}
	for _, workload := range *workloads {
		response.Workloads = append(response.Workloads, api.WorkloadSlim{
			Id:                   workload.ID.String(),
			Name:                 workload.Name,
			Type:                 workload.Type,
			ProjectName:          workload.ProjectName,
			ClusterId:            workload.ClusterID.String(),
			DistributedFramework: workload.DistributedFramework,
		})
	}

	recordsCount := int32(len(response.Workloads))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

// SyncWorkloads - Sync workloads.
func (s *WorkloadsService) SyncWorkloads(ctx context.Context, filterBy []string, workloadsSyncRequest api.WorkloadsSyncRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting workloads sync in tenant <%d> and project <%s>", tenantId, workloadsSyncRequest.ProjectId)
	from, to, apiErr := parseFromToUuids(workloadsSyncRequest.From, workloadsSyncRequest.To)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	deletionFilters, filterErr := queryapi.NewParser(models.WorkloadView{}.FilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	workloadModels := []models.Workload{}
	workloadChildrenModels := []models.WorkloadChild{}
	workloadElementsByID := make(map[uuid.UUID][]models.WorkloadElement)
	existingWorkloadIds := workloadsSyncRequest.IdsToIgnore
	for _, workload := range workloadsSyncRequest.Workloads {
		workloadModel, workloadChildren, workloadElements, err := SyncRequestToWorkloadsModel(ctx, tenantId, workloadsSyncRequest.ProjectId, workload)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Workload <%s> is invalid, err: %v", workload.Name, err)
			return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Workload <%s> is invalid", workload.Name), err)), nil
		}

		workloadModels = append(workloadModels, workloadModel)
		existingWorkloadIds = append(existingWorkloadIds, workloadModel.ID.String())
		workloadChildrenModels = append(workloadChildrenModels, workloadChildren...)
		workloadElementsByID[workloadModel.ID] = workloadElements
	}

	syncedWorkloads, failedWorkloads, deletedWorkloads, dbErr := s.WorkloadsDao.DbSyncWorkloads(ctx, tenantId,
		workloadsSyncRequest.ProjectId, workloadsSyncRequest.ClusterId, workloadModels, to, from,
		workloadsSyncRequest.IdsToIgnore, deletionFilters)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync workloads in project <%s>. Err: %+v", workloadsSyncRequest.ProjectId, dbErr)
		return NewErrorResponse(dbErr), nil
	}

	if snapshotErr := dao_crud.DbDeleteWorkloadV2Snapshots(ctx, tenantId, deletedWorkloads, nil); snapshotErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete V2 workload snapshots for deleted synced workloads: %v", snapshotErr.Message)
	}

	err := services_integration_crud.UpdateCrudOnSyncWorkloads(ctx, tenantId, workloadsSyncRequest.ProjectId, workloadsSyncRequest.ClusterId, to, from, existingWorkloadIds, deletionFilters)
	if err != nil {
		//  Not returning error since this is not breaking error. the workloads are already synced in the workloads_view
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync crud_workloads in project <%s>. Err: %v", workloadsSyncRequest.ProjectId, err)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(
		"Successfully synced workloads with ids: %v in project <%s>. Workloads excluded from deletion: %v. From: <%s>. To: <%s>",
		syncedWorkloads, workloadsSyncRequest.ProjectId, workloadsSyncRequest.IdsToIgnore, from, to)

	if len(syncedWorkloads) > 0 {
		dbErr = s.WorkloadsDao.DbUpsertWorkloadsChildren(ctx, workloadChildrenModels, failedWorkloads)
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync workloads children in project <%s>. Err: %+v", workloadsSyncRequest.ProjectId, dbErr)
			return NewErrorResponse(dbErr), nil
		}

		dao.DbUpsertWorkloadsElements(ctx, tenantId, workloadsSyncRequest.ClusterId, workloadElementsByID, failedWorkloads)
	}

	if len(failedWorkloads) != 0 {
		return api.Response(http.StatusMultiStatus, api.InternalSyncWorkloads207Response{
			SuccessfullySyncedWorkloads: syncedWorkloads,
			FailedWorkloads:             failedWorkloads,
		}), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully synced workloads children for workloads with ids: %v", syncedWorkloads)
	return NoContentResponse(), nil
}

// UpsertWorkload - Upsert the workload info from the cluster.
func (s *WorkloadsService) UpsertWorkload(ctx context.Context, workloadId string, workloadSyncFields api.WorkloadSyncFields) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf(
		"Starting workload id: <%s> name: <%s> upserting request in project id: <%s> name: <%s>",
		workloadId, workloadSyncFields.Name, workloadSyncFields.ProjectId, workloadSyncFields.ProjectName)

	if workloadId != workloadSyncFields.Id {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Path Id <%s> and body Id <%s> are different", workloadId, workloadSyncFields.Id), nil)), nil
	}

	workloadModel, workloadChildrenModel, workloadElementModels, err := SyncRequestToWorkloadsModel(ctx, tenantId, workloadSyncFields.ProjectId, workloadSyncFields)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Workload id: <%s> name: <%s> is invalid, err: %v", workloadId, workloadSyncFields.Name, err)
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Workload <%s> is invalid", workloadSyncFields.Name), err)), nil
	}

	dbErr := s.WorkloadsDao.DbUpsertWorkloadFull(ctx, tenantId, workloadSyncFields.ClusterId, workloadModel, workloadChildrenModel, workloadElementModels)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync workload <%s> in project <%s>. Err: %+v", workloadId, workloadSyncFields.ProjectId, dbErr)
		return NewErrorResponse(dbErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(
		"Successfully upserted workload id: <%s> name: <%s> in project id: <%s> name: <%s>",
		workloadId, workloadSyncFields.Name, workloadSyncFields.ProjectId, workloadSyncFields.ProjectName)
	return NoContentResponse(), nil
}

func (s *WorkloadsService) PatchWorkloadPhase(ctx context.Context, workloadId string, patchFields api.WorkloadPatchFields) (api.ImplResponse, error) {
	err := PatchWorkloadPhaseHelper(ctx, workloadId, patchFields, time.Now())
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return NoContentResponse(), nil
}

func PatchWorkloadPhaseHelper(ctx context.Context, workloadId string, patchFields api.WorkloadPatchFields, updateOnlyIfLastUpdateSmallerThan time.Time) *api.Error {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return apiErr
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting workload id: <%s> patch request to phase %s", workloadId, patchFields.Phase)

	if patchFields.Phase == api.PHASE_DELETING {
		err := validateDeletingPhase(ctx, tenantId, workloadId, &patchFields)
		if err != nil {
			return err
		}
	}

	dbErr := dao.DbUpdateWorkloadPhase(ctx, tenantId, workloadId, patchFields, updateOnlyIfLastUpdateSmallerThan)
	if dbErr != nil {
		return dbErr
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully updated workload id: <%s> to phase: %v", workloadId, patchFields.Phase)
	return nil
}

func (s *WorkloadsService) InternalCreateWorkload(ctx context.Context, workloadCreationFields api.WorkloadCreationFields) (api.ImplResponse, error) {
	workloadModel, apiErr := s.InternalCreateWorkloadHelper(ctx, workloadCreationFields)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	workload, dbErr := s.WorkloadsDao.DbGetWorkload(ctx, workloadModel.TenantID, workloadModel.ID)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully created workload <%s>", workloadCreationFields.Id)
	return api.Response(http.StatusCreated, WorkloadModelToApi(ctx, *workload, getClusterConfiguration(ctx, workload.ClusterID.String()))), nil
}

func (s *WorkloadsService) InternalCreateWorkloadHelper(ctx context.Context, workloadCreationFields api.WorkloadCreationFields) (models.Workload, *api.Error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return models.Workload{}, apiErr
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting workload <%s> internal creation request in tenant <%d>", workloadCreationFields.Id, tenantId)

	workloadSyncFields := api.WorkloadSyncFields{
		Type:                 workloadCreationFields.Type,
		Group:                workloadCreationFields.Group,
		Version:              workloadCreationFields.Version,
		Name:                 workloadCreationFields.Name,
		Id:                   workloadCreationFields.Id,
		SubmittedBy:          workloadCreationFields.SubmittedBy,
		PriorityClassName:    "",
		ClusterId:            workloadCreationFields.ClusterId,
		ProjectName:          workloadCreationFields.ProjectName,
		ProjectId:            workloadCreationFields.ProjectId,
		DepartmentId:         workloadCreationFields.DepartmentId,
		DepartmentName:       workloadCreationFields.DepartmentName,
		Namespace:            "",
		CreatedAt:            workloadCreationFields.CreatedAt,
		RequestedResources:   workloadCreationFields.RequestedResources,
		ActionsSupport:       api.ActionsSupport{Delete: true, Suspend: true},
		Phase:                api.PHASE_CREATING,
		PhaseMessage:         ControlPlaneCreatingMessage,
		Source:               api.SOURCE_CONTROL_PLANE,
		Conditions:           nil,
		K8sPhase:             "",
		RequestedPods:        workloadCreationFields.RequestedPods,
		RequestedNodePools:   workloadCreationFields.RequestedNodePools,
		Images:               workloadCreationFields.Images,
		DistributedFramework: workloadCreationFields.DistributedFramework,
		ChildrenIds:          nil,
		Urls:                 make([]string, 0),
		AdditionalFields:     "",
		FullySaasWorkload:    workloadCreationFields.FullySaasWorkload,
		Category:             workloadCreationFields.Category,
	}

	if workloadSyncFields.ProjectId == "" || workloadSyncFields.DepartmentId == "" {
		err := s.setProjectData(ctx, &workloadSyncFields)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to enrich project data. Err %v", err)
		}
	}

	workloadModel, _, _, err := SyncRequestToWorkloadsModel(ctx, tenantId, workloadCreationFields.ProjectId, workloadSyncFields)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Workload <%s> is invalid, err: %v", workloadSyncFields.Name, err)
		return models.Workload{}, api.NewError(http.StatusBadRequest, fmt.Sprintf("Workload <%s> is invalid", workloadSyncFields.Name), err)
	}

	internalCreatedWorkloadModel := CreationRequestToInternalCreatedWorkloadModel(workloadCreationFields)
	if internalCreatedWorkloadModel.SourceApi != nil {
		if *internalCreatedWorkloadModel.SourceApi == api.SOURCEAPI_WORKLOADS_V2 {
			workloadModel.SupportDelete = true
			suspendLookupKind := suspendSupportLookupKind(workloadCreationFields.Type, workloadCreationFields.Group, workloadCreationFields.Version, workloadCreationFields.DistributedFramework)
			workloadModel.SupportSuspend = s.isSuspendSupported(ctx, tenantId, suspendLookupKind, workloadCreationFields.Group, workloadCreationFields.Version, workloadCreationFields.ClusterId)
		}
	}

	dbErr := s.WorkloadsDao.DbCreateInternalWorkload(ctx, workloadModel, internalCreatedWorkloadModel)

	return workloadModel, dbErr
}

func suspendSupportLookupKind(typeName, group, version, distributedFramework string) string {
	workloadType := common.NewWorkloadType(typeName)
	if (workloadType == common.WorkloadTypeTraining || workloadType == common.WorkloadTypeWorkspace) &&
		group == "batch" && version == "v1" {
		return "Job"
	}
	if workloadType == common.WorkloadTypeDistributed && group == "kubeflow.org" {
		if kind := distributedFrameworkToKind(distributedFramework); kind != "" {
			// Preserve the submitted version and let the following exact Karta lookup
			// decide whether this concrete GVK supports suspend/resume.
			return kind
		}
	}

	return typeName
}

func distributedFrameworkToKind(distributedFramework string) string {
	switch api.DistributedFramework(distributedFramework) {
	case api.DISTRIBUTEDFRAMEWORK_MPI:
		return constants.KindMpiJob
	case api.DISTRIBUTEDFRAMEWORK_PY_TORCH:
		return constants.KindPyTorchJob
	case api.DISTRIBUTEDFRAMEWORK_TF:
		return constants.KindTFJob
	case api.DISTRIBUTEDFRAMEWORK_XG_BOOST:
		return constants.KindXGBoostJob
	case api.DISTRIBUTEDFRAMEWORK_JAX:
		return constants.KindJAXJob
	default:
		return ""
	}
}

// isSuspendSupported returns true when both the workload type has a suspendDefinition in its Karta
// AND the cluster version supports suspend/resume (≥ 2.26).
// Karta is checked first; if absent we skip the cluster call entirely.
// The cluster capability check is cheap because the checker caches cluster info per cluster ID.
func (s *WorkloadsService) isSuspendSupported(ctx context.Context, tenantId uint, typeName, group, version, clusterId string) bool {
	karta, err := dao.DbGetKartaByGVK(ctx, tenantId, typeName, group, version)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to look up karta for workload type %q, group %q, version %q, defaulting SupportSuspend=false: %v", typeName, group, version, err)
		return false
	}
	if karta == nil || karta.Karta.Spec.StructureDefinition.RootComponent.SuspendDefinition == nil {
		return false
	}

	clusterSupportsSuspend, capErr := checker.NewClusterCapabilitiesChecker(clusterId).ValidateClusterCapability(ctx, checker.MinVersionOfAnyWorkloadSuspendResume)
	if capErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to check cluster capability for suspend/resume on cluster %q, defaulting SupportSuspend=false: %v", clusterId, capErr)
		return false
	}
	return clusterSupportsSuspend
}

func (s *WorkloadsService) GetWorkloadAncestors(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	ancestors, err := authorization_utils.GetAncestorsCache(ctx, WorkloadsResourceType, workloadId)
	if err == nil {
		return api.Response(http.StatusOK, ancestors), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Could not find workload <%s> ancestors in cache. err: %v, continuing", workloadId, err)

	id, err := uuid.Parse(workloadId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("workload Id must be a valid uuid. <%s>", workloadId), err)), nil
	}

	workload, dbErr := s.WorkloadsDao.DbGetWorkload(ctx, tenantId, id)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	ancestors = &authorization_utils.Ancestors{
		TenantId:     fmt.Sprintf("%d", workload.TenantID),
		ClusterUuid:  ptr.To(workload.ClusterID.String()),
		DepartmentId: ptr.To(workload.DepartmentID),
		ProjectId:    ptr.To(workload.ProjectID),
	}

	authorization_utils.SetAncestorsCache(ctx, WorkloadsResourceType, workloadId, *ancestors)

	return api.Response(http.StatusOK, *ancestors), nil
}

func (s *WorkloadsService) GetWorkloadMetrics(ctx context.Context, workloadId string, metricTypes []api.WorkloadMetricType, start time.Time, end time.Time, numberOfSamples int32) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("getting workload id %s metrics of types %+v", workloadId, metricTypes)
	apiErr := validateRequestedMetrics(len(metricTypes))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	clusterId, dbErr := s.WorkloadsDao.DbGetWorkloadClusterId(ctx, tenantId, workloadId)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	if err := validateRequestedNIMWorkloadMetrics(ctx, clusterId, metricTypes, s.WorkloadsDao, tenantId, workloadId); err != nil {
		return NewErrorResponse(err), nil
	}

	if err := validateRequestedVLLMWorkloadMetrics(ctx, metricTypes, s.WorkloadsDao, tenantId, workloadId); err != nil {
		return NewErrorResponse(err), nil
	}

	if err := validateMetricsClusterVersions(ctx, clusterId.String(), metricTypes, supportedWorkloadsMetricsVersion); err != nil {
		return NewErrorResponse(err), nil
	}

	jobIdFilter, err := s.getJobIdFilter(ctx, workloadId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	workloadServiceResponse := api.MetricsCompositeResponse{
		Measurements: []api.MeasurementResponse{},
	}
	responseCode := http.StatusOK

	for _, metricType := range metricTypes {
		result := queryWorkloadMetricType(ctx, workloadId, clusterId, jobIdFilter, start, end, metricType, numberOfSamples)
		buildMetricsComposite(ctx, string(metricType), &workloadServiceResponse, &responseCode, result, false)
	}

	if workloadServiceResponse.Histogram != nil {
		workloadServiceResponse.Histogram = convertHistogramSeriesOrdered(*workloadServiceResponse.Histogram)
	}

	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		response := buildCsvFile(metricTypes, workloadServiceResponse, "workload-metrics.csv")
		return api.Response(responseCode, response), nil
	}

	return api.Response(responseCode, workloadServiceResponse), nil
}

func (s *WorkloadsService) setProjectData(ctx context.Context, workloadData *api.WorkloadSyncFields) error {
	project, apiErr := orgUnitClient.Get().GetProjectById(ctx, workloadData.ProjectId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get the project %s in cluster %s. err: %v", workloadData.ProjectId, workloadData.ClusterId, apiErr)
		return fmt.Errorf("failed to get project %s: %s", workloadData.ProjectId, apiErr.Message)
	}

	workloadData.ProjectName = project.Name
	workloadData.DepartmentId = project.ParentId

	if project.Parent.IsSet() && project.Parent.Get() != nil {
		workloadData.DepartmentName = project.Parent.Get().Name
	}

	return nil
}

func (s *WorkloadsService) InternalSyncWorkloads(ctx context.Context, filterBy []string, workloadSyncRequest api.InternalWorkloadsSyncRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	deletionFilters, filterErr := queryapi.NewParser(models.WorkloadView{}.FilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting internal workloads sync in cluster <%s>", workloadSyncRequest.ClusterId)

	workloadModels := []models.Workload{}
	workloadChildrenModels := []models.WorkloadChild{}
	workloadElementsByID := make(map[uuid.UUID][]models.WorkloadElement)
	for _, workload := range workloadSyncRequest.Workloads {
		workloadModel, workloadChildren, workloadElements, err := SyncRequestToWorkloadsModel(ctx, tenantId, workload.ProjectId, workload)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Workload <%s> is invalid, err: %v", workload.Name, err)
			return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Workload <%s> is invalid", workload.Name), err)), nil
		}

		workloadModels = append(workloadModels, workloadModel)
		workloadChildrenModels = append(workloadChildrenModels, workloadChildren...)
		workloadElementsByID[workloadModel.ID] = workloadElements
	}

	syncedWorkloads, failedWorkloads, _, dbErr := s.WorkloadsDao.DbSyncWorkloads(ctx, tenantId, "", workloadSyncRequest.ClusterId, workloadModels, nil, nil, []string{}, deletionFilters)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully synced internal workloads with ids: %v in cluster <%s>.", syncedWorkloads, workloadSyncRequest.ClusterId)

	if len(syncedWorkloads) > 0 {
		dbErr = s.WorkloadsDao.DbUpsertWorkloadsChildren(ctx, workloadChildrenModels, failedWorkloads)
		if dbErr != nil {
			return NewErrorResponse(dbErr), nil
		}

		dao.DbUpsertWorkloadsElements(ctx, tenantId, workloadSyncRequest.ClusterId, workloadElementsByID, failedWorkloads)
	}

	if len(failedWorkloads) != 0 {
		return api.Response(http.StatusMultiStatus, api.InternalSyncWorkloads207Response{
			SuccessfullySyncedWorkloads: syncedWorkloads,
			FailedWorkloads:             failedWorkloads,
		}), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully synced internal workloads children for workloads with ids: %v in cluster <%s>.", syncedWorkloads, workloadSyncRequest.ClusterId)
	return NoContentResponse(), nil
}

func (s *WorkloadsService) GetWorkloadElements(ctx context.Context, workloadId string, offset int32, limit int32) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	wid, err := uuid.Parse(workloadId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Invalid workload id %q", workloadId), nil)), nil
	}

	clusterId, apiErr := s.WorkloadsDao.DbGetWorkloadClusterId(ctx, tenantId, workloadId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiErr = validateWorkloadElementsSupported(ctx, clusterId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	elements, apiErr := dao.DbGetWorkloadElements(ctx, tenantId, wid, offset, limit)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiElements := modelsToWorkloadElements(elements)

	response := api.WorkloadElementsResponse{
		Elements: apiElements,
	}
	if int32(len(apiElements)) == limit {
		response.Next = offset + int32(len(apiElements))
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *WorkloadsService) GetWorkloadElement(ctx context.Context, workloadId string, elementId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	wid, err := uuid.Parse(workloadId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Invalid workload id %q", workloadId), nil)), nil
	}

	if _, err = uuid.Parse(elementId); err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Invalid element id %q", elementId), nil)), nil
	}

	workload, apiErr := s.WorkloadsDao.DbGetWorkload(ctx, tenantId, wid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiErr = validateWorkloadElementsSupported(ctx, workload.ClusterID)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	element, apiErr := dao.DbGetWorkloadElement(ctx, tenantId, wid, elementId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	pods, apiErr := dao.DbGetElementPods(ctx, tenantId, wid, elementId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	response := modelToWorkloadElementResponse(*element, pods)

	if topologyName := resolveTopologyNameForElement(ctx, *workload, pods); topologyName != "" {
		actual := getActualTopologyPlacementForElement(
			ctx, tenantId, wid, elementId, workload.ClusterID.String(), topologyName,
		)
		if response.Topology != nil {
			response.Topology.ActualPlacement = actual
		} else if actual != nil {
			response.Topology = &api.WorkloadElementTopology{
				Name:            topologyName,
				ActualPlacement: actual,
			}
		}
		if response.Topology != nil && response.Topology.Name == "" {
			response.Topology.Name = topologyName
		}
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *WorkloadsService) CountWorkloadElements(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	wid, err := uuid.Parse(workloadId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Invalid workload id %q", workloadId), nil)), nil
	}

	workload, apiErr := s.WorkloadsDao.DbGetWorkload(ctx, tenantId, wid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiErr = validateWorkloadElementsSupported(ctx, workload.ClusterID)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	storedCount, apiErr := dao.DbCountWorkloadElements(ctx, tenantId, wid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	totalElementsCount := int32(storedCount)
	if workload.StructureElementsTotalCount != nil {
		totalElementsCount = int32(*workload.StructureElementsTotalCount)
	}

	response := api.CountWorkloadElements200Response{
		StoredElementsCount: int32(storedCount),
		TotalElementsCount:  totalElementsCount,
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *WorkloadsService) GetWorkloadElementMetrics(ctx context.Context, workloadId string, elementId string,
	metricTypes []api.ElementMetricType, start time.Time, end time.Time, numberOfSamples int32,
) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("getting element %s metrics of types %+v for workload %s", elementId, metricTypes, workloadId)
	apiErr := validateRequestedMetrics(len(metricTypes))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	wid, err := uuid.Parse(workloadId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Invalid workload id %q", workloadId), nil)), nil
	}

	if _, err = uuid.Parse(elementId); err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Invalid element id %q", elementId), nil)), nil
	}

	clusterId, apiErr := s.WorkloadsDao.DbGetWorkloadClusterId(ctx, tenantId, workloadId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiErr = validateWorkloadElementsSupported(ctx, clusterId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	_, apiErr = dao.DbGetWorkloadElement(ctx, tenantId, wid, elementId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	elementMetricsResponse := api.MetricsCompositeResponse{
		Measurements: []api.MeasurementResponse{},
	}
	responseCode := http.StatusOK

	for _, metricType := range metricTypes {
		result := queryElementMetricType(ctx, elementId, clusterId, start, end, metricType, numberOfSamples)
		buildMetricsComposite(ctx, string(metricType), &elementMetricsResponse, &responseCode, result, false)
	}

	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		response := buildCsvFile(metricTypes, elementMetricsResponse, "element-metrics.csv")
		return api.Response(responseCode, response), nil
	}

	return api.Response(responseCode, elementMetricsResponse), nil
}

func (s *WorkloadsService) GetWorkloadEndpoints(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	_, parseErr := uuid.Parse(workloadId)
	if parseErr != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("workload Id must be a valid uuid. <%s>", workloadId), parseErr)), nil
	}
	workload, apiErr := dao.GetWorkloadViewFromDb(ctx, workloadId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	var endpoints []api.Endpoint
	if common.WorkloadType(workload.Type).IsNative() {
		apiWorkload := WorkloadModelToApi(ctx, *workload, getClusterConfiguration(ctx, workload.ClusterID.String()))
		endpoints = apiWorkload.Endpoints
	} else {
		endpoints = ModelEndpointToApi(workload.Endpoints)
	}

	return api.Response(http.StatusOK, api.GetWorkloadEndpoints200Response{
		Endpoints: endpoints,
	}), nil
}

// getJobIdFilter -
// in cluster versions >=2.16, all metrics have "workloadId" label which is always the workloadId.
// but for backwards compatibility of this API, the IDs might be different -
// in clusters 2.9, all metrics have "jobId" label which is always equal to the workloadId. so no problem as well.
// but for clusters 2.10-2.15, there's no workloadId label,
// and the "jobId" label is the ChildId of type RunaiJob of the workload.
// we can't easily know the cluster version from here, so we'll just fetch the RunaiJob child and that's it...
// TODO: delete after all clusters are 2.16+ ....
func (s *WorkloadsService) getJobIdFilter(ctx context.Context, workloadId string) ([]string, *api.Error) {
	id, err := uuid.Parse(workloadId)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("workload Id must be a valid uuid. <%s>", workloadId), err)
	}

	possibleChildTypes := []string{dao.RunaiJobType, dao.RevisionType}
	children, dbErr := s.WorkloadsDao.DbGetChildren(ctx, id, &possibleChildTypes)
	if dbErr != nil {
		return nil, dbErr
	}

	runaijobId := ""
	revisionId := ""
	for _, child := range *children {
		if child.ChildType == dao.RunaiJobType {
			runaijobId = child.ChildID.String()
			continue
		}
		if child.ChildType == dao.RevisionType {
			revisionId = child.ChildID.String()
		}
	}

	if runaijobId != "" {
		return []string{runaijobId}, nil
	}
	if revisionId != "" {
		return []string{revisionId}, nil
	}

	// if no "runaijob" or "revision" children - use the workloadId as the jobId
	// (this happens, for example, for "deployments" in old clusters - there's no runaijob,
	// and the "workloadId" is the "job_uuid" in the metrics)
	return []string{workloadId}, nil
}

func validateDeletingPhase(ctx context.Context, tenantId uint, workloadId string, patchFields *api.WorkloadPatchFields) *api.Error {
	workloadUuid, err := uuid.Parse(workloadId)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("workload Id must be a valid uuid. <%s>", workloadId), err)
	}
	workload, apiErr := dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]()).DbGetWorkload(ctx, tenantId, workloadUuid)
	if apiErr != nil {
		return apiErr
	}
	if workload.Phase == api.PHASE_CREATING && workload.PhaseMessage == ControlPlaneCreatingMessage {
		patchFields.SoftDelete = true
	}

	return nil
}

// createConnectionsFromUrls converts a list of URLs into Connection structs
// Each URL is parsed to extract subdomain, path, and port information
// The EW kind (environment workload kind) is used as the connection type/name
func createConnectionsFromUrls(urls []string, ewKind string) ([]models.Connection, error) {
	if len(urls) == 0 {
		return nil, nil
	}
	connections := make([]models.Connection, 0, len(urls))

	for index, rawUrl := range urls {
		if rawUrl == "" {
			continue
		}

		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL %s: %w", rawUrl, err)
		}

		// Extract subdomain from host (e.g., "sub-domain" from "sub-domain.example.com")
		var subdomain *string
		if parsedUrl.Host != "" {
			hostParts := strings.Split(parsedUrl.Host, ".")
			if len(hostParts) > 0 {
				subdomain = &hostParts[0]
			}
		}

		// Extract path
		var path *string
		if parsedUrl.Path != "" && parsedUrl.Path != "/" {
			path = &parsedUrl.Path
		}

		// Extract port or default to 8080
		containerPort := int32(8080)
		if parsedUrl.Port() != "" {
			if port, err := strconv.Atoi(parsedUrl.Port()); err == nil {
				containerPort = int32(port)
			}
		}

		// Create connection struct
		conn := models.Connection{
			Name:                toConnectionName(ewKind, index+1, len(urls)),
			Type:                ewKind,
			Url:                 rawUrl,
			ConnectionType:      "ExternalUrl",
			ContainerPort:       containerPort,
			CorrelationName:     "url",
			ExposedUrlSubDomain: subdomain,
			ExposedUrlPath:      path,
		}

		connections = append(connections, conn)
	}

	return connections, nil
}

func toConnectionName(kind string, currentIndex, totalUrlsCount int) string {
	if totalUrlsCount == 1 {
		return kind
	}
	return fmt.Sprintf("%s-%d", kind, currentIndex)
}

// backwards

func (s *WorkloadsService) SyncJobsBackwards(ctx context.Context, clusterId string, syncJobsBackwardsRequest api.SyncJobsBackwardsRequest) (api.ImplResponse, error) {
	return api.Response(http.StatusCreated, nil), nil
}

func (s *WorkloadsService) SyncPodsBackwards(ctx context.Context, clusterId string, syncJobsBackwardsRequest api.SyncJobsBackwardsRequest) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, api.SyncPodsBackwards200Response{Ids: []string{}}), nil
}

func (s *WorkloadsService) SyncPodgroupsBackwards(ctx context.Context, clusterId string, projectName string, syncJobsBackwardsRequest api.SyncJobsBackwardsRequest) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, api.SyncPodsBackwards200Response{Ids: []string{}}), nil
}

func (s *WorkloadsService) PostSyncPodsBackwards(ctx context.Context, clusterId string, syncJobsBackwardsRequest api.SyncJobsBackwardsRequest) (api.ImplResponse, error) {
	return api.Response(http.StatusCreated, "Pods Synced in DB"), nil
}

func (s *WorkloadsService) PostSyncPodgroupsBackwards(ctx context.Context, clusterId string, syncJobsBackwardsRequest api.SyncJobsBackwardsRequest) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, "The pod Group was Synced successfully"), nil
}

// resolveTopologyNameForWorkload returns the network-topology name to use when computing
// actual placement: the requested constraint name if set; otherwise the topology shared by
// the workload's CurrentNodepools (via their NetworkTopologyName).
// Returns "" when no topology applies — no running pods to place, no current nodepools, or
// no shared topology among them.
func resolveTopologyNameForWorkload(ctx context.Context, workload models.WorkloadView) string {
	if workload.Topology != nil && workload.Topology.Name != "" {
		return workload.Topology.Name
	}
	if workload.NumberOfRunningPods == 0 || len(workload.CurrentNodepools) == 0 {
		return ""
	}

	actualNodePoolsResponse, getErr := clusterservice.Get().GetNodePoolsInternalByCluster(ctx, workload.ClusterID.String())
	if getErr != nil {
		logging.Errorf("failed to get node pools for cluster %s: %v", workload.ClusterID, getErr)
	}

	actualNodePools := []clusterApi.Nodepool{}
	if actualNodePoolsResponse != nil {
		actualNodePools = actualNodePoolsResponse.Nodepools
	}

	found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools, workload.CurrentNodepools)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to resolve topology for workload %s nodepools %v: %v", workload.ID, workload.CurrentNodepools, err)
		return ""
	}
	if !found {
		return ""
	}
	return topologyName
}

// resolveTopologyNameForElement is the per-element variant of resolveTopologyNameForWorkload:
// it uses the constraint name when set, otherwise discovers the topology shared by the
// element's running pods (rather than the entire workload's nodepools). This way an element
// running on a topology-aware nodepool still reports actualPlacement even when other elements
// of the same workload sit on non-topology nodepools.
func resolveTopologyNameForElement(ctx context.Context, workload models.WorkloadView, elementPods []models.Pod) string {
	if workload.Topology != nil && workload.Topology.Name != "" {
		return workload.Topology.Name
	}
	nodepools := distinctRunningNodepools(elementPods)
	if len(nodepools) == 0 {
		return ""
	}

	actualNodePoolsResponse, getErr := clusterservice.Get().GetNodePoolsInternalByCluster(ctx, workload.ClusterID.String())
	if getErr != nil {
		logging.Errorf("failed to get node pools for cluster %s: %v", workload.ClusterID, getErr)
	}

	actualNodePools := []clusterApi.Nodepool{}
	if actualNodePoolsResponse != nil {
		actualNodePools = actualNodePoolsResponse.Nodepools
	}

	found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools, nodepools)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to resolve topology for element pods of workload %s: %v", workload.ID, err)
		return ""
	}
	if !found {
		return ""
	}
	return topologyName
}

// distinctRunningNodepools collects the unique non-empty CurrentNodepool values from running
// pods. Returns an empty slice when no running pod has a current nodepool set.
func distinctRunningNodepools(pods []models.Pod) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, p := range pods {
		if p.Phase != string(api.PHASE_RUNNING) {
			continue
		}
		if p.CurrentNodepool == nil || *p.CurrentNodepool == "" {
			continue
		}
		if _, ok := seen[*p.CurrentNodepool]; ok {
			continue
		}
		seen[*p.CurrentNodepool] = struct{}{}
		result = append(result, *p.CurrentNodepool)
	}
	return result
}

func getActualTopologyPlacementForWorkload(ctx context.Context, tenantId uint, workloadId uuid.UUID, clusterId string, topologyName string) *api.WorkloadTopologyActualPlacement {
	return resolveActualTopologyPlacement(ctx, clusterId, topologyName, workloadId, func(ctx context.Context) (pq.StringArray, *api.Error) {
		return dao.DbGetWorkloadTopologyLca(ctx, tenantId, workloadId)
	})
}

func getActualTopologyPlacementForElement(ctx context.Context, tenantId uint, workloadId uuid.UUID, elementId string, clusterId string, topologyName string) *api.WorkloadTopologyActualPlacement {
	return resolveActualTopologyPlacement(ctx, clusterId, topologyName, workloadId, func(ctx context.Context) (pq.StringArray, *api.Error) {
		return dao.DbGetElementTopologyLca(ctx, tenantId, workloadId, elementId)
	})
}

// resolveActualTopologyPlacement gates on cluster topology support, runs the supplied LCA query,
// and maps the LCA path to a network-topology level/value pair. Returns nil when topology is
// unsupported, the LCA cannot be determined, or the topology definition cannot be resolved.
func resolveActualTopologyPlacement(
	ctx context.Context,
	clusterId string,
	topologyName string,
	workloadId uuid.UUID,
	lcaFn func(context.Context) (pq.StringArray, *api.Error),
) *api.WorkloadTopologyActualPlacement {
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	topologySupported, _ := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionTopologyAddress)
	if !topologySupported {
		return nil
	}

	lcaPath, lcaErr := lcaFn(ctx)
	if lcaErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get topology LCA for workload %s: %v", workloadId, lcaErr)
		return nil
	}
	if len(lcaPath) == 0 {
		return nil
	}

	networkTopology, topologyErr := topology.GetByName(ctx, clusterservice.Get(), topologyName)
	if topologyErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get network topology %s for workload %s: %v", topologyName, workloadId, topologyErr)
		return nil
	}
	if len(lcaPath) > len(networkTopology.Levels) {
		return nil
	}

	return &api.WorkloadTopologyActualPlacement{
		Level: networkTopology.Levels[len(lcaPath)-1],
		Value: lcaPath[len(lcaPath)-1],
	}
}
