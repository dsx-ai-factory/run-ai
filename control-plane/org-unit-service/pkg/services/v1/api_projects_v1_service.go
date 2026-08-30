package v1

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"

	"github.com/google/uuid"
	auth_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	MetricsClusterMinVersion = "2.11"
	OnlySelectedAffinityType = "only_selected"
)

type ProjectsV1ApiService struct {
	authorizationClient    auth_client.IAuthorizationClient
	clustersClient         cluster_service_client.IClusterExternalClient
	projectDeletionService services.ProjectDeletionService
}

func NewProjectV1ApiService(
	clustersApiService cluster_service_client.IClusterExternalClient,
	projectDeletionService services.ProjectDeletionService,
) api.ProjectsV1APIServicer {
	return &ProjectsV1ApiService{
		authorizationClient:    auth_client.NewAuthorizationClient(),
		clustersClient:         clustersApiService,
		projectDeletionService: projectDeletionService,
	}
}

func (s *ProjectsV1ApiService) GetProjectsV1(ctx context.Context, clusterId string, excludePermissions bool, memoryUnitMb bool) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get V1 projects")
	tenantId, clusterUuid, err := GetTenantAndClusterUuid(ctx, clusterId)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant and cluster id"), nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	projects, err := dao.GetProjectsByCluster(ctx, *tenantId, clusterUuid, permittedScopes, models.SlimVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get projects")
		return errResponse, nil
	}

	accessRulesPerId := make(map[string]*[]openapi.AccessRule)
	if !excludePermissions {
		accessRulesPerId, err = services.GetProjectsPermissions(ctx, s.authorizationClient, clusterId, projects)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusServiceUnavailable, "Failed to get project permissions")
			return errResponse, nil
		}
	}

	nodePools, err := services.GetNodePoolsMetadataMap(ctx, s.clustersClient, clusterId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
		return errResponse, nil
	}
	response, err := MapProjectsModelToV1Response(ctx, projects, accessRulesPerId, memoryUnitMb, nodePools, excludePermissions)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get projects"), nil
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *ProjectsV1ApiService) UpdateProjectStatusV1(ctx context.Context, clusterId string, projectV1StatusRequest api.ProjectV1StatusRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Updating V1 project status, clusterId %s, project %v", clusterId, projectV1StatusRequest.Name)
	tenantId, clusterUuid, err := GetTenantAndClusterUuid(ctx, clusterId)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant and cluster id"), nil
	}

	project, err := dao.GetProjectByNameAndCluster(ctx, projectV1StatusRequest.Name, clusterUuid, *tenantId, models.SlimVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to update project")
		return errResponse, nil
	}
	projectStatus, resourcesStatuses, err := MapProjectStatusApiV1ToModel(projectV1StatusRequest, project)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "update project status failed")
		return errResponse, nil
	}

	err = dao.UpdateProjectStatus(ctx, *tenantId, clusterUuid, project, projectStatus, resourcesStatuses)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to set project status")
		return errResponse, nil
	}

	return api.Response(http.StatusOK, nil), nil
}

func (s *ProjectsV1ApiService) GetProjectsMetricsV1(ctx context.Context, clusterUuid string, start time.Time, end time.Time, numberOfSamples int32, nodepoolName string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get V1 projects metrics")
	err := ValidateMetricsRequest(ctx, clusterUuid, start, end, &numberOfSamples, s.clustersClient)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to get project's metrics"), nil
	}
	tenantId, cluster, err := GetTenantAndClusterUuid(ctx, clusterUuid)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant and cluster id"), nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	projects, err := dao.GetProjectsByCluster(ctx, *tenantId, cluster, permittedScopes, models.SlimVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get projects")
		return errResponse, nil
	}
	if len(projects) == 0 {
		return api.Response(http.StatusOK, api.ProjectsMetricsV1Response{
			Data: []api.ProjectsMetricsV1ResponseDataInner{},
		}), nil
	}

	projectNames := []string{}
	for _, project := range projects {
		projectNames = append(projectNames, project.Name)
	}
	nodePools := []string{".*"}
	if nodepoolName != "" {
		nodePools = strings.Split(nodepoolName, ",")
	}
	_, currentMetrics, err := GetCurrentMetrics(ctx, clusterUuid, nodePools, []string{".*"}, projectNames, true)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get metrics"), nil
	}

	fetchUtilization := ShouldFetchUtilization(permittedScopes)
	timeRangeMetrics, err := GetTimeRangeMetrics(ctx, clusterUuid, nodePools, []string{".*"}, &projectNames, start, end, numberOfSamples, fetchUtilization)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get metrics"), nil
	}

	clusterName, err := services.GetClusterName(ctx, clusterUuid, s.clustersClient)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get cluster name"), nil
	}

	response, err := MapProjectsMetricsToResponse(ctx, projects, currentMetrics, timeRangeMetrics, clusterUuid, clusterName, fetchUtilization)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get metrics"), nil
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *ProjectsV1ApiService) GetProjectMetricsV1(ctx context.Context, clusterUuid string, id string, start time.Time, end time.Time, numberOfSamples int32, nodepoolName string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get V1 project metrics")
	err := ValidateMetricsRequest(ctx, clusterUuid, start, end, &numberOfSamples, s.clustersClient)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to get project's metrics"), nil
	}

	projectId, err := strconv.Atoi(id)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse project id"), nil
	}
	modelProject, err := getProjectById(ctx, projectId, clusterUuid)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to get project"), nil
	}

	projectNames := []string{modelProject.Name}
	departments := []string{".*"}
	nodePools := []string{".*"}
	if nodepoolName != "" {
		nodePools = strings.Split(nodepoolName, ",")
	}
	_, currentMetrics, err := GetCurrentMetrics(ctx, clusterUuid, nodePools, departments, projectNames, true)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get metrics"), nil
	}

	timeRangeMetrics, err := GetTimeRangeMetrics(ctx, clusterUuid, nodePools, departments, &projectNames, start, end, numberOfSamples, false)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get metrics"), nil
	}

	clusterName, err := services.GetClusterName(ctx, clusterUuid, s.clustersClient)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get cluster name"), nil
	}
	response, err := MapProjectMetricsToResponse(ctx, clusterUuid, clusterName, modelProject, currentMetrics[modelProject.Name], timeRangeMetrics)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get metrics"), nil
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *ProjectsV1ApiService) GetProjectsQuotasV1(ctx context.Context, clusterId string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get V1 projects quota")
	tenantId, cluster, err := GetTenantAndClusterUuid(ctx, clusterId)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant and cluster id"), nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	projects, err := dao.GetProjectsByCluster(ctx, *tenantId, cluster, permittedScopes, models.SlimVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get projects")
		return errResponse, nil
	}
	projectNames := []string{}
	for _, project := range projects {
		projectNames = append(projectNames, project.Name)
	}

	quotaMetrics, err := GetQuotaMetricResults(ctx, clusterId, projectNames)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to get projects quota"), nil
	}

	quotaResponse, err := MapQuotaMetricsAndProjectsToResponse(ctx, projects, quotaMetrics, *tenantId, clusterId)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get projects quota"), nil
	}

	return api.Response(http.StatusOK, quotaResponse), nil
}

func getProjectById(ctx context.Context, projectId int, clusterId string) (*models.Project, error) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return nil, err
	}
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return nil, err
	}
	return dao.GetProjectByIdAndCluster(ctx, int64(projectId), *tenantId, clusterUuid, models.SlimVerbosity)
}
