/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"net/http"
	"time"

	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	ext_api "github.com/run-ai/runai/control-plane/workloads/pkg/api"
	clients "github.com/run-ai/runai/control-plane/workloads/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_utils"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var supportedPodMetricsVersion = withVLLMMetricsVersion(map[string]int{
	string(api.PODMETRICTYPE_NVLINK_BANDWIDTH_TOTAL):         checker.MinVersionNvlinkMetrics,
	string(api.PODMETRICTYPE_NVLINK_BANDWIDTH_TOTAL_PER_GPU): checker.MinVersionNvlinkMetrics,
})

type PodsService struct {
	tenantSettingsClient clients.TenantSettingsClient
	WorkloadsDao         *dao.WorkloadsDao
}

func NewPodsService(client clients.TenantSettingsClient) api.PodsAPIServicer {
	return &PodsService{
		tenantSettingsClient: client,
		WorkloadsDao:         dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]()),
	}
}

func (s *PodsService) InternalSyncPods(ctx context.Context, podsSyncRequest api.InternalPodsSyncRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting pods sync in cluster <%s>", podsSyncRequest.ClusterId)

	podModels := []models.Pod{}
	containersModels := []models.Container{}
	podElementsModels := []models.PodElement{}
	children := []models.WorkloadChild{}
	revisionChildren := []models.RevisionChild{}
	for _, pod := range podsSyncRequest.Pods {
		podModel, containerModels, podElements, child, revisionChild, err := CreationRequestToPodModel(ctx, tenantId, pod.ProjectId, pod)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Pod <%s> is invalid in sync request of cluster <%s>. Err: %+v", pod.Name, podsSyncRequest.ClusterId, err)
			return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Pod <%s> is invalid in sync request of cluster <%s>", pod.Name, podsSyncRequest.ClusterId), err)), nil
		}

		podModels = append(podModels, podModel)
		containersModels = append(containersModels, containerModels...)
		podElementsModels = append(podElementsModels, podElements...)
		children = append(children, child)
		if revisionChild != nil {
			revisionChildren = append(revisionChildren, *revisionChild)
		}
	}

	syncedPods, failedPods, dbErr := dao.DbSyncPods(ctx, tenantId, "", podsSyncRequest.ClusterId, podModels, nil, nil, []string{}, "")
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync pods of cluster <%s>. Err: %+v", podsSyncRequest.ClusterId, dbErr)
		return NewErrorResponse(dbErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully synced pods with ids: %v in cluster <%s>.", syncedPods, podsSyncRequest.ClusterId)

	if len(syncedPods) > 0 {
		dbErr = dao.DbUpsertContainers(ctx, containersModels, failedPods)
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync pods containers of cluster <%s>. Err: %+v", podsSyncRequest.ClusterId, dbErr)
			return NewErrorResponse(dbErr), nil
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully synced pods containers for pods with ids: %v in cluster <%s>", syncedPods, podsSyncRequest.ClusterId)

		if elemErr := dao.DbUpsertPodElements(ctx, podElementsModels, failedPods); elemErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to sync pod elements of cluster <%s>, will retry on next sync. Err: %+v",
				podsSyncRequest.ClusterId, elemErr)
		}

		dbErr = s.WorkloadsDao.DbUpsertWorkloadsPodsChildren(ctx, children, failedPods)
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync workload children of cluster <%s>. Err: %+v", podsSyncRequest.ClusterId, dbErr)
			return NewErrorResponse(dbErr), nil
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully synced pods as workload children for pods with ids: %v", syncedPods)

		if len(revisionChildren) > 0 {
			dbErr = dao.DbUpsertRevisionsPodsChildren(ctx, revisionChildren, failedPods)
			if dbErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync revision children of cluster <%s>. Err: %+v", podsSyncRequest.ClusterId, dbErr)
				return NewErrorResponse(dbErr), nil
			}
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully synced pods as revision children for pods with ids: %v", syncedPods)
		}
	}

	if len(failedPods) != 0 {
		return api.Response(http.StatusMultiStatus, api.InternalSyncPods207Response{
			SuccessfullySyncedPods: syncedPods,
			FailedPods:             failedPods,
		}), nil
	}

	return NoContentResponse(), nil
}

func (s *PodsService) GetWorkloadPod(ctx context.Context, workloadId string, podId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	workloadUuid, err := uuid.Parse(workloadId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("workload Id must be a valid uuid. <%s>", workloadId), err)), nil
	}
	podUuid, err := uuid.Parse(podId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("pod Id must be a valid uuid. <%s>", podId), err)), nil
	}

	pod, dbErr := dao.DbGetPod(ctx, tenantId, workloadUuid, podUuid)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}
	return api.Response(http.StatusOK, PodModelToApi(*pod, pod.ProjectID)), nil
}

func (s *PodsService) DeletePod(ctx context.Context, podId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting pod <%s> delete request", podId)

	podUuid, err := uuid.Parse(podId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("pod Id must be a valid uuid. <%s>", podId), err)), nil
	}

	dbErr := dao.DbDeletePod(ctx, tenantId, podUuid)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete pod <%s>. Err: %+v", podId, dbErr)
		return NewErrorResponse(dbErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully deleted pod <%s>", podId)
	return NoContentResponse(), nil
}

func (s *PodsService) InternalUpdateNodePods(ctx context.Context, clusterId string, nodeName string, updateNodePodsRequest api.UpdateNodePodsRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting update pods request for node <%s> in cluster <%s>", nodeName, clusterId)

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("clusterId must be a valid uuid. <%s>", clusterId), err)), nil
	}

	if nodeName == "" {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "nodeName cannot be empty", nil)), nil
	}

	updatedCount, dbErr := dao.DbUpdatePodsByNodeName(ctx, tenantId, clusterUuid, nodeName, updateNodePodsRequest)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update pods for node <%s> in cluster <%s>. Err: %+v", nodeName, clusterId, dbErr)
		return NewErrorResponse(dbErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully updated %d pods for node <%s> in cluster <%s>", updatedCount, nodeName, clusterId)
	return NoContentResponse(), nil
}

func (s *PodsService) ListPods(ctx context.Context, deleted bool, offset int32, limit int32, sortOrder string, sortBy string,
	filterBy []string, verbosity api.PodVerbosity, completed string, search string,
) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	limit = setDefaultLimit(limit)

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	verbosityType, apiErr := ext_api.ParsePodVerbosity(verbosity)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	filters, filterErr := dao_utils.ValidateAndParameterizeApiFilters(filterBy, models.Pod{}.SortAndFilterFieldsToColumn())
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	sortColumn, sortError := dao_utils.ValidateSortByField(sortBy, models.Pod{}.SortAndFilterFieldsToColumn())
	if sortError != nil {
		return ErrorToApiError(sortError, http.StatusBadRequest), nil
	}

	pods, dbErr := dao.DbListPods(ctx, tenantId, deleted, offset, limit, sortOrder, sortColumn, filters, verbosityType, completed, search, permittedScopes)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	apiPods := []api.Pod{}
	switch p := pods.(type) {
	case *[]models.Pod:
		for _, pod := range *p {
			apiPods = append(apiPods, PodModelToApi(pod, pod.ProjectID))
		}
	case *[]models.PodFull:
		for _, pod := range *p {
			fullApiPod := PodModelToApi(pod.Pod, pod.ProjectID)
			if pod.WorkloadName != nil {
				fullApiPod.WorkloadName = *pod.WorkloadName
			}
			if pod.ProjectName != nil {
				fullApiPod.ProjectName = *pod.ProjectName
			}
			apiPods = append(apiPods, fullApiPod)
		}
	}

	recordsCount := int32(len(apiPods))

	response := api.ListPods200Response{
		Pods: apiPods,
	}
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *PodsService) GetWorkloadPods(ctx context.Context, workloadId string, deleted bool, offset int32, limit int32, search string, sortOrder string, sortBy string, filterBy []string) (api.ImplResponse, error) {
	if filterBy == nil {
		filterBy = []string{}
	}

	filterBy = append(filterBy, fmt.Sprintf("workloadId==%s", workloadId))
	return s.ListPods(ctx, deleted, offset, limit, sortOrder, sortBy, filterBy, "", "", search)
}

func (s *PodsService) GetRevisionPods(ctx context.Context, revisionId string, deleted bool) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	// this call is needed in order to get the project id
	revision, dbErr := dao.DbGetRevision(ctx, tenantId, revisionId)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	pods, dbErr := dao.DbGetRevisionPods(ctx, tenantId, revisionId, deleted, 0, 0, "")
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	projectId := revision.ProjectID
	apiPods := []api.Pod{}
	for _, pod := range *pods {
		apiPods = append(apiPods, PodModelToApi(pod, projectId))
	}

	response := api.GetRevisionPods200Response{
		Pods: apiPods,
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *PodsService) CountPods(ctx context.Context, deleted bool, filterBy []string, search string, completed string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	filters, filterErr := dao_utils.ValidateAndParameterizeApiFilters(filterBy, models.Pod{}.SortAndFilterFieldsToColumn())
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	count, dbErr := dao.DbCountPods(ctx, tenantId, deleted, filters, search, completed, permittedScopes)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	response := api.CountWorkloads200Response{
		Count: count,
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *PodsService) SyncPods(ctx context.Context, workloadPodsSyncRequest []api.WorkloadPodsSyncRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	totalFailedPods := []string{}
	totalSyncedPods := []string{}

	for i := range workloadPodsSyncRequest {
		podsToSyncRequest := workloadPodsSyncRequest[i]
		syncedPods, failedPods, errResponse := s.syncGroupOfPods(ctx, podsToSyncRequest.Pods, podsToSyncRequest.ClusterId,
			podsToSyncRequest.ProjectId, podsToSyncRequest.IdsToIgnore, tenantId, nil, nil, podsToSyncRequest.WorkloadId)
		if errResponse != nil {
			return *errResponse, nil
		}
		totalFailedPods = append(totalFailedPods, failedPods...)
		totalSyncedPods = append(totalSyncedPods, syncedPods...)
	}
	if len(totalFailedPods) != 0 {
		return api.Response(http.StatusMultiStatus, api.InternalSyncPods207Response{
			SuccessfullySyncedPods: totalSyncedPods,
			FailedPods:             totalFailedPods,
		}), nil
	}

	return NoContentResponse(), nil
}

func (s *PodsService) SyncWorkloadsPods(ctx context.Context, podsSyncRequest api.PodsSyncRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting pods sync in tenant <%d> and project <%s>", tenantId, podsSyncRequest.ProjectId)
	from, to, apiErr := parseFromToUuids(podsSyncRequest.From, podsSyncRequest.To)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	syncedPods, failedPods, errResponse := s.syncGroupOfPods(ctx, podsSyncRequest.Pods, podsSyncRequest.ClusterId, podsSyncRequest.ProjectId, podsSyncRequest.IdsToIgnore, tenantId, to, from, "")
	if errResponse != nil {
		return *errResponse, nil
	}

	if len(failedPods) != 0 {
		return api.Response(http.StatusMultiStatus, api.InternalSyncPods207Response{
			SuccessfullySyncedPods: syncedPods,
			FailedPods:             failedPods,
		}), nil
	}

	return NoContentResponse(), nil
}

func (s *PodsService) syncGroupOfPods(
	ctx context.Context,
	pods []api.PodSyncFields,
	clusterId string,
	projectId string,
	idsToIgnore []string,
	tenantId uint,
	to *uuid.UUID,
	from *uuid.UUID,
	workloadId string,
) ([]string, []string, *api.ImplResponse) {
	podModels := []models.Pod{}
	containersModels := []models.Container{}
	podElementsModels := []models.PodElement{}
	children := []models.WorkloadChild{}
	revisionChildren := []models.RevisionChild{}
	for _, pod := range pods {
		podModel, containerModels, podElements, child, revisionChild, err := CreationRequestToPodModel(ctx, tenantId, projectId, pod)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Pod <%s> is invalid within project with id <%s>. Err: %+v", pod.Name, pod.ProjectId, err)
			return nil, nil, ptr.To(NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Pod <%s> is invalid in sync request of project <%s>", pod.Name, projectId), err)))
		}

		podModels = append(podModels, podModel)
		containersModels = append(containersModels, containerModels...)
		podElementsModels = append(podElementsModels, podElements...)
		children = append(children, child)
		if revisionChild != nil {
			revisionChildren = append(revisionChildren, *revisionChild)
		}
	}

	syncedPods, failedPods, dbErr := dao.DbSyncPods(ctx, tenantId, projectId, clusterId, podModels, to, from, idsToIgnore, workloadId)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync pods of project with id <%s>. Err: %+v", projectId, dbErr)
		return nil, nil, ptr.To(NewErrorResponse(dbErr))
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(
		"Successfully synced pods with ids: %v in project <%s>. Pods excluded from deletion: %v. From: <%s>. To: <%s>",
		syncedPods, projectId, idsToIgnore, from, to)

	if len(syncedPods) > 0 {
		dbErr = dao.DbUpsertContainers(ctx, containersModels, failedPods)
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync pods containers of project with id <%s>. Err: %+v", projectId, dbErr)
			return nil, nil, ptr.To(NewErrorResponse(dbErr))
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully synced pods containers for pods with ids: %v", syncedPods)

		if elemErr := dao.DbUpsertPodElements(ctx, podElementsModels, failedPods); elemErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to sync pod elements of project with id <%s>, will retry on next sync. Err: %+v",
				projectId, elemErr)
		}

		dbErr = s.WorkloadsDao.DbUpsertWorkloadsPodsChildren(ctx, children, failedPods)
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync workloads pods children of project with id <%s>. Err: %+v", projectId, dbErr)
			return nil, nil, ptr.To(NewErrorResponse(dbErr))
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully synced pods as workload children for pods with ids: %v", syncedPods)

		if len(revisionChildren) > 0 {
			dbErr = dao.DbUpsertRevisionsPodsChildren(ctx, revisionChildren, failedPods)
			if dbErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync revision pods children of project with id <%s>. Err: %+v", projectId, dbErr)
				return nil, nil, ptr.To(NewErrorResponse(dbErr))
			}
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully synced pods as revision children for pods with ids: %v", syncedPods)
		}
	}
	return syncedPods, failedPods, nil
}

func (s *PodsService) UpsertPod(ctx context.Context, podId string, podSyncFields api.PodSyncFields) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting pod id: <%s> name: <%s> (of workload id: <%s>) upserting request", podId, podSyncFields.Name, podSyncFields.WorkloadId)

	if podId != podSyncFields.Id {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Path Id <%s> and body Id <%s> are different", podId, podSyncFields.Id)
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Path Id <%s> and body Id <%s> are different", podId, podSyncFields.Id), nil)), nil
	}

	podModel, containersModel, podElements, child, revisionChild, err := CreationRequestToPodModel(ctx, tenantId, podSyncFields.ProjectId, podSyncFields)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Pod <%s> is invalid within project with id <%s>. Err: %+v", podSyncFields.Name, podSyncFields.ProjectId, err)
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Pod <%s> is invalid within project with id <%s>", podSyncFields.Name, podSyncFields.ProjectId), err)), nil
	}

	dbErr := dao.DbUpsertPod(ctx, podModel)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	dbErr = dao.DbUpsertContainers(ctx, containersModel, []string{})
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to upsert pod <%s> containers within project with id <%s>. Err: %+v", podSyncFields.Name, podSyncFields.ProjectId, dbErr)
		return NewErrorResponse(dbErr), nil
	}

	if elemErr := dao.DbUpsertPodElements(ctx, podElements, []string{}); elemErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to upsert pod <%s> elements within project with id <%s>, will retry on next sync. Err: %+v",
			podSyncFields.Name, podSyncFields.ProjectId, elemErr)
	}

	dbErr = s.WorkloadsDao.DbUpsertWorkloadsPodsChildren(ctx, []models.WorkloadChild{child}, []string{})
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to upsert pod <%s> as workload child within project with id <%s>. Err: %+v", podSyncFields.Name, podSyncFields.ProjectId, dbErr)
		return NewErrorResponse(dbErr), nil
	}

	if revisionChild != nil {
		dbErr = dao.DbUpsertRevisionsPodsChildren(ctx, []models.RevisionChild{*revisionChild}, []string{})
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to upsert pod <%s> as revision child within project with id <%s>. Err: %+v", podSyncFields.Name, podSyncFields.ProjectId, dbErr)
			return NewErrorResponse(dbErr), nil
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully upserted pod id: <%s> name: <%s> (of workload <%s>)", podId, podSyncFields.Name, podSyncFields.WorkloadId)
	return NoContentResponse(), nil
}

func (s *PodsService) GetWorkloadPodMetrics(ctx context.Context, workloadId string, podId string, metricTypes []api.PodMetricType, start time.Time, end time.Time, numberOfSamples int32) (api.ImplResponse, error) {
	apiErr := validateRequestedMetrics(len(metricTypes))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiErr = validateRequestedAdvancedMetrics(ctx, s.tenantSettingsClient, metricTypes)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	clusterId, dbErr := dao.DbGetPodClusterId(ctx, tenantId, workloadId, podId)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	if err := validateMetricsClusterVersions(ctx, clusterId.String(), metricTypes, supportedPodMetricsVersion); err != nil {
		return NewErrorResponse(err), nil
	}

	if err := validateRequestedNIMWorkloadMetrics(ctx, clusterId, metricTypes, s.WorkloadsDao, tenantId, workloadId); err != nil {
		return NewErrorResponse(err), nil
	}

	if err := validateRequestedVLLMWorkloadMetrics(ctx, metricTypes, s.WorkloadsDao, tenantId, workloadId); err != nil {
		return NewErrorResponse(err), nil
	}

	workloadServiceResponse := api.MetricsCompositeResponse{
		Measurements: []api.MeasurementResponse{},
	}
	responseCode := http.StatusOK

	for _, metricType := range metricTypes {
		result := queryPodMetricType(ctx, podId, clusterId, start, end, metricType, numberOfSamples)
		buildMetricsComposite(ctx, string(metricType), &workloadServiceResponse, &responseCode, result, true)
	}

	if workloadServiceResponse.Histogram != nil {
		workloadServiceResponse.Histogram = convertHistogramSeriesOrdered(*workloadServiceResponse.Histogram)
	}

	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		response := buildCsvFile(metricTypes, workloadServiceResponse, "pods-metrics.csv")
		return api.Response(responseCode, response), nil
	}

	return api.Response(responseCode, workloadServiceResponse), nil
}
