/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	orgUnitClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"net/http"
	"slices"
)

const applicationTypeApplication = "application"

type AIApplicationsService struct {
	SyncConf        *config.ApplicationsSyncConfig
	chartInspection *applicationChartInspector
}

func NewAIApplicationsService() api.AIApplicationsAPIServicer {
	return &AIApplicationsService{
		SyncConf:        runai_config.Get[*config.ApplicationsSyncConfig](),
		chartInspection: newApplicationChartInspector(),
	}
}

func (s *AIApplicationsService) ListApplications(ctx context.Context,
	verbosity api.AiApplicationVerbosity,
	offset int32,
	limit int32,
	sortOrder string,
	sortBy api.AiApplicationsSortFilterFields,
	filterBy []string,
	search string,
) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	filters, filterErr := queryapi.NewParser(models.Application{}.FieldsConfiguration()).Filters(filterBy)
	if filterErr != nil {
		return api.Response(http.StatusBadRequest, filterErr), nil
	}
	sortColumn, sortError := queryapi.NewParser(models.Application{}.FieldsConfiguration()).GetMapping(string(sortBy))
	if sortError != nil {
		return api.Response(http.StatusBadRequest, sortError), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	applications, err := dao.GetApplications(ctx, fmt.Sprint(tenantId), permittedScopes, offset, limit, sortOrder, sortColumn, filters, search, verbosity == api.AIAPPLICATIONVERBOSITY_FULL)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to get applications", nil).WithDetails(err)), nil
	}
	apiApplications := []api.AiApplication{}
	if applications != nil {
		for _, application := range *applications {
			apiApplication := applicationViewToApi(application, verbosity)
			sanitizeAIApplicationResponse(ctx, &apiApplication)
			apiApplications = append(apiApplications, apiApplication)
		}
	}

	response := api.ListApplications200Response{
		Applications: apiApplications,
	}
	recordsCount := int32(len(apiApplications))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *AIApplicationsService) GetApplication(ctx context.Context, applicationId string) (api.ImplResponse, error) {
	tenantIdInt, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	tenantId := fmt.Sprint(tenantIdInt)

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}
	applicationUuid, err := uuid.Parse(applicationId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("applicationId should be valid uuid <%s>", applicationId), err)), nil
	}
	application, err := dao.GetApplicationById(ctx, tenantId, applicationUuid, permittedScopes)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewErrorResponse(api.NewError(http.StatusNotFound, fmt.Sprintf("application %s not found", applicationId), nil)), nil
		}
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to get application", nil).WithDetails(err)), nil
	}

	response := applicationViewToApi(*application, api.AIAPPLICATIONVERBOSITY_FULL)
	sanitizeAIApplicationResponse(ctx, &response)
	return api.Response(http.StatusOK, response), nil
}

func (s *AIApplicationsService) CountApplications(ctx context.Context, filterBy []string, search string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	filters, filterErr := queryapi.NewParser(models.Application{}.FieldsConfiguration()).Filters(filterBy)
	if filterErr != nil {
		return api.Response(http.StatusBadRequest, filterErr), nil
	}

	count, dbErr := dao.CountApplications(ctx, fmt.Sprint(tenantId), permittedScopes, filters, search)
	if dbErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to count applications", nil).WithDetails(dbErr)), nil
	}

	response := api.CountResponse{
		Count: count,
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *AIApplicationsService) SyncApplications(ctx context.Context, clusterId string, applicationsForSync api.AiApplicationsForSync) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Syncing applications for: %+v", applicationsForSync)
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return api.Response(http.StatusBadRequest, fmt.Sprintf("invalid cluster id %s", clusterId)), nil
	}

	applicationsForUpsert, applicationsForDelete, err := getApplicationsForSync(ctx, fmt.Sprint(tenantId), clusterId, applicationsForSync)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to map sync applications to model", err)), nil
	}

	synced, failed, err := dao.SyncClusterApplications(ctx, fmt.Sprint(tenantId), clusterUuid, applicationsForUpsert, applicationsForDelete, applicationsForSync.FullSync)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to sync applications", err)), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Synced status for: %v", synced)
	if len(failed) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to sync status for: %v", failed)
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (s *AIApplicationsService) CreateApplication(ctx context.Context, createApplicationRequest api.AiApplicationCreationRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	appSubmissionFlag, err := setting_pkg.Get[*settings.ApplicationSubmission](fmt.Sprint(tenantId))
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to get application submission setting", nil).WithDetails(err)), nil
	}
	if !appSubmissionFlag.Enabled {
		return NewErrorResponse(api.NewError(http.StatusForbidden, "Application submission is not enabled. Please contact the administrator to enable it.", nil)), nil
	}

	if validationErr := validateApplicationSpec(&createApplicationRequest.Spec); validationErr != nil {
		return NewErrorResponse(validationErr), nil
	}

	project, orgUnitErr := orgUnitClient.Get().GetProjectById(ctx, createApplicationRequest.ProjectId)
	if orgUnitErr != nil {
		return NewErrorResponse(api.NewError(int(orgUnitErr.Code), "failed to get project", nil).WithDetails(orgUnitErr)), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get permitted scopes: %s", err.Error())
		permittedScopes = nil
	}

	err = validatePermittedScope(permittedScopes, createApplicationRequest.ProjectId, project.ParentId, createApplicationRequest.ClusterId, fmt.Sprint(tenantId))
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusForbidden, err.Error(), nil)), nil
	}

	if capErr := services_utils.ValidateClusterSupportsCapability(ctx, createApplicationRequest.ClusterId, checker.MinVersionOfApplicationCreation); capErr != nil {
		return NewErrorResponse(api.NewError(int(capErr.GetCode()), capErr.GetMessage(), nil)), nil
	}

	usesPersonalCredentials, credentialsErr := processAIApplicationCredentials(ctx, createApplicationRequest.Name, &createApplicationRequest.Spec)
	if credentialsErr != nil {
		return NewErrorResponse(credentialsErr), nil
	}
	if usesPersonalCredentials {
		if capabilityErr := validateApplicationPersonalCredentialsCapability(ctx, createApplicationRequest.ClusterId); capabilityErr != nil {
			return NewErrorResponse(capabilityErr), nil
		}
	}

	applicationId := uuid.New()
	createdBy := convutils.Deref(runaictx.GetString(ctx, runaictx.CfSubjectID), "")

	departmentName := ""
	if project.Parent.IsSet() && project.Parent.Get() != nil {
		departmentName = project.Parent.Get().Name
	}

	application := models.Application{
		BaseModel: dbmodels.BaseModel{
			ID:        applicationId,
			CreatedBy: createdBy,
			UpdatedBy: createdBy,
		},
		Name:           createApplicationRequest.Name,
		Source:         models.ApplicationSourceRunai,
		Type:           applicationTypeApplication,
		TenantId:       fmt.Sprint(tenantId),
		ClusterId:      createApplicationRequest.ClusterId,
		DepartmentId:   project.ParentId,
		DepartmentName: departmentName,
		ProjectId:      createApplicationRequest.ProjectId,
		ProjectName:    project.Name,
		ActualPhase:    models.ApplicationPhaseCreating,
	}

	dbErr := dao.CreateApplication(ctx, &application, &createApplicationRequest.Spec, convutils.Ref(models.ApplicationPhaseDeployed))
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	created, getErr := dao.GetApplicationById(ctx, fmt.Sprint(tenantId), applicationId, permittedScopes)
	if getErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "application created but failed to retrieve it", nil).WithDetails(getErr)), nil
	}

	response := applicationViewToApi(*created, api.AIAPPLICATIONVERBOSITY_FULL)
	sanitizeAIApplicationResponse(ctx, &response)
	return api.Response(http.StatusOK, response), nil
}

func (s *AIApplicationsService) DeleteApplication(ctx context.Context, aiApplicationId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	applicationUuid, err := uuid.Parse(aiApplicationId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("applicationId should be valid uuid <%s>", aiApplicationId), err)), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	application, err := dao.GetApplicationById(ctx, fmt.Sprint(tenantId), applicationUuid, permittedScopes)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusNotFound, fmt.Sprintf("application %s not found", aiApplicationId), nil).WithDetails(err)), nil
	}

	if application.Source != models.ApplicationSourceRunai {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("application %s was not created via Run:ai and cannot be deleted through this API", aiApplicationId), nil)), nil
	}

	if application.ActualPhase == models.ApplicationPhaseCreating {
		if dbErr := dao.DeleteApplicationImmediate(ctx, fmt.Sprint(tenantId), applicationUuid); dbErr != nil {
			return NewErrorResponse(dbErr), nil
		}
		return api.Response(http.StatusAccepted, nil), nil
	}

	dbErr := dao.UpdateCrudSyncDesiredPhase(ctx, fmt.Sprint(tenantId), applicationUuid, models.ApplicationPhaseDeleted)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	return api.Response(http.StatusAccepted, nil), nil
}

func (s *AIApplicationsService) RequestApplicationsUpdates(ctx context.Context, updatesAfter int64, clusterId string, clusterSyncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantIdStr := fmt.Sprint(tenantId)
	processApplicationsSyncReport(ctx, tenantIdStr, clusterId, &clusterSyncReport)

	updates, dbErr := dao.ListApplicationUpdates(ctx, tenantIdStr, clusterId, updatesAfter, s.SyncConf.BulkSize)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	var pendingDeletionIds []uuid.UUID
	for i := range updates {
		if updates[i].DesiredPhase != nil && *updates[i].DesiredPhase == models.ApplicationPhaseDeleted {
			pendingDeletionIds = append(pendingDeletionIds, updates[i].ID)
			updates[i].ActualPhase = models.ApplicationPhaseDeleting
		}
	}
	updatedBy := convutils.Deref(runaictx.GetString(ctx, runaictx.CfSubjectID), "")
	if markErr := dao.MarkApplicationsAsDeleting(ctx, tenantIdStr, pendingDeletionIds, updatedBy); markErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to mark applications as deleting: %s", markErr.Message)
	}

	syncResponse := api.AiApplicationSyncResponse{
		RequestInfo: api.RequestInfo{
			AfterUpdate: updatesAfter,
		},
	}

	if len(updates) != 0 {
		syncResponse.ResponseInfo.LowestUpdate = updates[0].UpdateCount
		syncResponse.ResponseInfo.HighestUpdate = updates[len(updates)-1].UpdateCount
	}

	chartURLBuilder := newChartDownloadURLBuilder(runai_config.Get[*config.ChartUploadConfig]())
	for _, update := range updates {
		syncItem, err := applicationSyncUpdateToSyncItem(update, chartURLBuilder)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("skipping sync update for application %s: %s", update.ID, err)
			continue
		}

		syncResponse.Applications = append(syncResponse.Applications, api.AiApplicationSyncInfo{
			UpdateCount: update.UpdateCount,
			Application: syncItem,
		})
	}

	maxCount, countErr := dao.GetApplicationsHighestUpdateCount(ctx)
	if countErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get highest application update count: %s", countErr.Message)
	} else {
		syncResponse.SystemInfo.HighestUpdate = maxCount
	}

	if cleanupErr := dao.CleanupApplicationCrudSync(ctx, tenantIdStr, clusterId, updatesAfter); cleanupErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to cleanup application crud sync: %s", cleanupErr.Message)
	}

	return api.Response(http.StatusOK, syncResponse), nil
}

var allowedRepoSchemes = []string{"http://", "https://", "oci://"}

func validateApplicationSpec(spec *api.AiApplicationSpec) *api.Error {
	chartRef := spec.ChartRef
	hasURL := chartRef.ChartURL != ""
	hasChartId := chartRef.ChartId != ""
	hasAnyRepoField := chartRef.Repo != "" || chartRef.ChartName != "" || chartRef.ChartVersion != ""
	hasAllRepoFields := chartRef.Repo != "" && chartRef.ChartName != "" && chartRef.ChartVersion != ""

	sourceCount := 0
	for _, present := range []bool{hasURL, hasChartId, hasAnyRepoField} {
		if present {
			sourceCount++
		}
	}
	if sourceCount == 0 {
		return api.NewError(http.StatusBadRequest, "chartRef must define exactly one chart source: chartURL, chartId, or repo+chartName+chartVersion.", nil)
	}
	if sourceCount > 1 {
		return api.NewError(http.StatusBadRequest, "chartRef must define exactly one chart source: chartURL, chartId, or repo+chartName+chartVersion, not multiple.", nil)
	}

	if hasAnyRepoField && !hasAllRepoFields {
		return api.NewError(http.StatusBadRequest, "chartRef requires repo, chartName, and chartVersion when using a repository reference.", nil)
	}

	if hasChartId {
		if _, err := uuid.Parse(chartRef.ChartId); err != nil {
			return api.NewError(http.StatusBadRequest, "chartRef.chartId must be a valid uuid.", err)
		}
	}

	if chartRef.Repo != "" {
		validScheme := false
		for _, scheme := range allowedRepoSchemes {
			if len(chartRef.Repo) > len(scheme) && chartRef.Repo[:len(scheme)] == scheme {
				validScheme = true
				break
			}
		}
		if !validScheme {
			return api.NewError(http.StatusBadRequest, fmt.Sprintf("chartRef.repo must use one of the following schemes: %v", allowedRepoSchemes), nil)
		}
	}

	if spec.Auth.AssetId != "" {
		if _, err := uuid.Parse(spec.Auth.AssetId); err != nil {
			return api.NewError(http.StatusBadRequest, "spec.auth.assetId must be a valid uuid.", err)
		}
	}

	return nil
}

func processApplicationsSyncReport(ctx context.Context, tenantId string, clusterId string, syncReport *api.ClusterSyncReport) {
	if syncReport == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("missing sync report of cluster %s for applications", clusterId)
		return
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("processing %d application sync reports for cluster %s", len(syncReport.Updates), clusterId)

	for _, update := range syncReport.Updates {
		if update.Message == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("skip report for application %s from cluster %s: no message", update.Id, clusterId)
			continue
		}

		dbErr := dao.UpdateApplicationSyncMessage(ctx, tenantId, clusterId, update.Id, *update.Message)
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update sync message for application %s: %s", update.Id, dbErr.Message)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("update status of application %s from cluster %s: %s", update.Id, clusterId, *update.Message)
		}
	}
}

func applicationSyncUpdateToSyncItem(update models.ApplicationSyncUpdate, chartURLBuilder *chartDownloadURLBuilder) (api.AiApplicationSyncItem, error) {
	if update.Spec == nil {
		return api.AiApplicationSyncItem{}, fmt.Errorf("application %s has no spec", update.ID)
	}

	item := api.AiApplicationSyncItem{
		Id:        update.ID.String(),
		Name:      update.Name,
		Source:    update.Source,
		Type:      update.Type,
		Phase:     update.ActualPhase,
		ProjectId: update.ProjectId,
		Spec:      *update.Spec,
	}

	// An uploaded chart (chartId) is rewritten to a control-plane download URL on the cluster-facing
	// copy only, and the envelope auth flag is set from chartId alone. Because the flag is derived
	// here and never carried on the spec, a user-provided chartURL can never make the controller send
	// its token to an arbitrary URL. The stored spec keeps chartId, so the public API still reports it.
	if item.Spec.ChartRef.ChartId != "" {
		downloadURL, err := chartURLBuilder.downloadURL(item.Spec.ChartRef.ChartId)
		if err != nil {
			return api.AiApplicationSyncItem{}, fmt.Errorf("failed to build chart download url for application %s: %w", update.ID, err)
		}
		item.Spec.ChartRef.ChartURL = downloadURL
		item.Spec.ChartRef.ChartId = ""
		item.RequiresControlPlaneAuth = true
	}

	if update.DesiredPhase != nil {
		item.DesiredPhase = *update.DesiredPhase
	}

	return item, nil
}

func getApplicationsForSync(ctx context.Context, tenantId string, clusterId string, applicationsForSync api.AiApplicationsForSync) ([]models.Application, []string, error) {
	applicationsForUpsert := []models.Application{}
	applicationsForDelete := []string{}

	applications := applicationsForSync.Applications
	for i := range applications {
		if applications[i].Action == api.AIAPPLICATIONSYNCACTION_DELETE {
			applicationsForDelete = append(applicationsForDelete, applications[i].Id)
		} else {
			modelApplication, err := mapSyncApplicationToModel(tenantId, clusterId, applications[i])
			if err != nil {
				return nil, nil, err
			}
			// Older clusters do not send departmentName in the sync payload.
			// The org-unit client caches every department name from
			// a single org-tree fetch, so a sync batch
			// spanning many departments triggers at most one call per TTL window.
			if modelApplication.DepartmentName == "" && modelApplication.DepartmentId != "" {
				departmentName, orgUnitErr := orgUnitClient.Get().GetDepartmentName(ctx, modelApplication.DepartmentId)
				if orgUnitErr != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to resolve department name for department %s: %s", modelApplication.DepartmentId, orgUnitErr.Message)
				} else {
					modelApplication.DepartmentName = departmentName
				}
			}
			applicationsForUpsert = append(applicationsForUpsert, *modelApplication)
		}
	}

	return applicationsForUpsert, applicationsForDelete, nil
}

func mapSyncApplicationToModel(tenantId string, clusterId string, applicationForSync api.AiApplicationForSync) (*models.Application, error) {
	id, err := uuid.Parse(applicationForSync.Id)
	if err != nil {
		return nil, err
	}
	applicationResources := []models.ApplicationResource{}
	for _, resource := range applicationForSync.Resources {
		applicationResources = append(applicationResources, models.ApplicationResource{
			Name:        resource.Name,
			Id:          resource.Id,
			Kind:        resource.Kind,
			ClusterId:   resource.ClusterId,
			ClusterName: resource.ClusterName,
		})
	}
	applicationStatus := models.ApplicationStatus{
		K8sPhase:  applicationForSync.K8sPhase,
		Resources: applicationResources,
	}

	// actualPhase is calculated like this:
	// for a non-completed application -> the status is the status of the action
	// for a completed application -> the status is the helmStatus (deployed, for example) or k8sPhase if no helm status
	actualPhase, messagePtr := determineStatus(applicationForSync)

	return &models.Application{
		BaseModel: dbmodels.BaseModel{
			ID: id,
		},
		Name:              applicationForSync.Name,
		Source:            applicationForSync.Source,
		Type:              applicationForSync.Type,
		TenantId:          tenantId,
		ClusterId:         clusterId,
		DepartmentId:      applicationForSync.DepartmentId,
		DepartmentName:    applicationForSync.DepartmentName,
		ProjectId:         applicationForSync.ProjectId,
		ProjectName:       applicationForSync.ProjectName,
		ApplicationStatus: applicationStatus,
		Message:           messagePtr,
		ActualPhase:       actualPhase,
	}, nil
}

// the actualPhase and message are
func determineStatus(applicationForSync api.AiApplicationForSync) (string, *string) {
	var actualPhase string
	var message string

	if applicationForSync.ActionInfo != nil {
		if applicationForSync.ActionInfo.Status == "completed" {
			actualPhase = applicationForSync.ActionInfo.HelmPhase
			if actualPhase == "" {
				actualPhase = applicationForSync.K8sPhase
			}
			if applicationForSync.ActionInfo.Message != "" {
				message = applicationForSync.ActionInfo.Message
			}
		} else {
			actualPhase = applicationForSync.ActionInfo.Status

			if applicationForSync.ActionInfo.Message != "" {
				message = applicationForSync.ActionInfo.Message
			}

			if applicationForSync.ActionInfo.Error != "" {
				if message != "" {
					message += fmt.Sprintf(" (%s)", applicationForSync.ActionInfo.Error)
				} else {
					message = applicationForSync.ActionInfo.Error
				}
			}
		}
	}

	var messagePtr *string
	if message != "" {
		messagePtr = &message
	}

	return actualPhase, messagePtr
}

func applicationViewToApi(application models.ApplicationView, verbosity api.AiApplicationVerbosity) api.AiApplication {
	resources := []api.AiApplicationResource{}
	for _, resource := range application.Resources {
		resources = append(resources, api.AiApplicationResource{
			Id:   resource.Id,
			Name: resource.Name,
			Kind: resource.Kind,
		})
	}

	var actionStatus *string
	if application.ActualPhase != "" {
		actionStatus = convutils.Ref(application.ActualPhase)
	}

	apiApplication := api.AiApplication{
		Id:             application.ID.String(),
		Name:           application.Name,
		Source:         application.Source,
		Type:           application.Type,
		Phase:          application.K8sPhase,
		ProjectId:      application.ProjectId,
		ProjectName:    application.ProjectName,
		DepartmentId:   application.DepartmentId,
		DepartmentName: application.DepartmentName,
		Resources:      resources,
		CreatedAt:      application.CreatedAt,
		CreatedBy:      application.CreatedBy,
		UpdatedAt:      application.UpdatedAt,
		UpdatedBy:      application.UpdatedBy,
		DesiredPhase:   application.DesiredPhase,
		ClusterId:      application.ClusterId,
		ActionStatus:   actionStatus,
		ActionMessage:  application.Message,
	}
	if len(application.Endpoints) > 0 {
		apiEndpoints := []api.AiApplicationEndpoint{}
		for _, workloadEndpoints := range application.Endpoints {
			for _, endpoint := range workloadEndpoints.Endpoints {
				apiEndpoints = append(apiEndpoints, api.AiApplicationEndpoint{
					WorkloadId: workloadEndpoints.WorkloadId,
					Host:       endpoint.Host,
					Port:       endpoint.Port,
					Path:       endpoint.Path,
					TlsEnabled: endpoint.TlsEnabled,
					Protocol:   endpoint.Protocol,
					Url:        endpoint.Url,
				})
			}
		}
		apiApplication.Endpoints = apiEndpoints
	}

	if len(application.ResourceConnections) > 0 {
		apiResourceConnections := []api.AiApplicationResourceConnection{}
		for _, workloadConnections := range application.ResourceConnections {
			for _, ref := range workloadConnections.ResourceRefs {
				apiResourceConnections = append(apiResourceConnections, api.AiApplicationResourceConnection{
					WorkloadId: workloadConnections.WorkloadId,
					Uid:        ref.UID,
					Name:       ref.Name,
					Kind:       api.ResourceRefKind(ref.Kind),
				})
			}
		}
		apiApplication.ResourceConnections = apiResourceConnections
	}

	if application.Spec != nil {
		apiApplication.Spec = *application.Spec
	}
	if application.UpdateCount != nil {
		apiApplication.UpdateCount = *application.UpdateCount
	}
	if verbosity == api.AIAPPLICATIONVERBOSITY_FULL {
		apiApplication.RequestedResources = getApplicationsRequestedResources(
			application.RequestedGpuLimit,
			application.RequestedGpuRequest,
			application.RequestedGpuMemoryLimit,
			application.RequestedGpuMemoryRequest,
			application.RequestedCpuLimit,
			application.RequestedCpuRequest,
			application.RequestedCpuMemoryLimit,
			application.RequestedCpuMemoryRequest,
		)

		apiApplication.AllocatedResources = &api.AiApplicationAllocatedResources{
			Gpu:       application.AllocatedGpu,
			GpuMemory: sumQuantities(application.AllocatedGpuMemory),
			Cpu:       application.AllocatedCpu,
			CpuMemory: sumQuantities(application.AllocatedCpuMemory),
		}
	}
	return apiApplication
}

func getApplicationsRequestedResources(
	requestedGPULimit *float64,
	requestedGPURequest *float64,
	requestedGPUMemoryLimit *pq.StringArray,
	requestedGPUMemoryRequest *pq.StringArray,
	requestedCPULimit *float64,
	requestedCPURequest *float64,
	requestedCPUMemoryLimit *pq.StringArray,
	requestedCPUMemoryRequest *pq.StringArray,
) *api.AiApplicationRequestResources {
	resources := &api.AiApplicationRequestResources{}

	if requestedGPULimit != nil || requestedGPURequest != nil {
		resources.Gpu = &api.RequestResourceCores{
			Limit:   requestedGPULimit,
			Request: requestedGPURequest,
		}
	} else if requestedGPUMemoryLimit != nil || requestedGPUMemoryRequest != nil {
		resources.GpuMemory = &api.RequestResourceQuantity{
			Limit:   sumQuantities(requestedGPUMemoryLimit),
			Request: sumQuantities(requestedGPUMemoryRequest),
		}
	}

	if requestedCPULimit != nil || requestedCPURequest != nil {
		resources.Cpu = &api.RequestResourceCores{
			Limit:   requestedCPULimit,
			Request: requestedCPURequest,
		}
	}

	if requestedCPUMemoryLimit != nil || requestedCPUMemoryRequest != nil {
		resources.CpuMemory = &api.RequestResourceQuantity{
			Limit:   sumQuantities(requestedCPUMemoryLimit),
			Request: sumQuantities(requestedCPUMemoryRequest),
		}
	}

	return resources
}

func validatePermittedScope(permittedScopes *authorization_api.PermittedScopes, projectId string, departmentId string, clusterId string, tenantId string) error {
	if permittedScopes == nil {
		return nil
	}

	if permittedScopes.System != nil && *permittedScopes.System {
		return nil
	}

	if slices.Contains(permittedScopes.Projects, projectId) {
		return nil
	}

	if slices.Contains(permittedScopes.Departments, departmentId) {
		return nil
	}

	if slices.Contains(permittedScopes.Clusters, clusterId) {
		return nil
	}

	if slices.Contains(permittedScopes.Tenants, tenantId) {
		return nil
	}

	return api.NewError(http.StatusForbidden, "user does not have access to the project, department, cluster, or tenant", nil)
}
