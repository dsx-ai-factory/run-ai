package services

import (
	"context"
	"net/http"

	org_unit "github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"

	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/auth"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/common"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// ConfigMapsService implement all routes relates to ConfigMap requests and operations
type ConfigMapsService struct {
	dbAccessor dao.ConfigMapsDBAccessor // dbAccessor ep to db operations
}

// NewConfigMapsService returns a new ConfigMap service
func NewConfigMapsService(dbAccessor dao.ConfigMapsDBAccessor) *ConfigMapsService {
	return &ConfigMapsService{
		dbAccessor: dbAccessor,
	}
}

// CreateClusterWideConfigmapV1 handles ConfigMap creation in cluster scope request
func (cms *ConfigMapsService) CreateClusterWideConfigmapV1(ctx context.Context, clusterId string, name string, request api.ConfigMap) (api.ImplResponse, error) {
	res, err := cms.insertConfigMap(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, name, request.Issues)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, *res), nil
}

// CreateDepartmentsWideConfigmapV1 handles ConfigMap creation in department scope request
func (cms *ConfigMapsService) CreateDepartmentsWideConfigmapV1(ctx context.Context, clusterId string, departmentId string, name string, request api.ConfigMap) (api.ImplResponse, error) {
	res, err := cms.insertConfigMap(ctx, clusterId, authorizationClient.SCOPETYPE_DEPARTMENT, departmentId, name, request.Issues)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, *res), nil
}

// CreateProjectsConfigmapV1 handles ConfigMap creation in project scope request
func (cms *ConfigMapsService) CreateProjectsConfigmapV1(ctx context.Context, clusterId string, namespace string, name string, request api.ConfigMap) (api.ImplResponse, error) {
	res, err := cms.insertConfigMap(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, name, request.Issues)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, *res), nil
}

// DeleteClusterWideConfigmapV1 handles ConfigMap deletion in cluster scope request
func (cms *ConfigMapsService) DeleteClusterWideConfigmapV1(ctx context.Context, clusterId string, name string) (api.ImplResponse, error) {
	err := cms.deleteConfigMap(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, name)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return NoContentResponse(), nil
}

// DeleteDepartmentsWideConfigmapV1 handles ConfigMap deletion in department scope request
func (cms *ConfigMapsService) DeleteDepartmentsWideConfigmapV1(ctx context.Context, clusterId string, departmentId string, name string) (api.ImplResponse, error) {
	err := cms.deleteConfigMap(ctx, clusterId, authorizationClient.SCOPETYPE_DEPARTMENT, departmentId, name)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return NoContentResponse(), nil
}

// DeleteProjectsConfigmapV1 handles ConfigMap deletion in project scope request
func (cms *ConfigMapsService) DeleteProjectsConfigmapV1(ctx context.Context, clusterId string, namespace string, name string) (api.ImplResponse, error) {
	err := cms.deleteConfigMap(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, name)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return NoContentResponse(), nil
}

// ListConfigMaps handles list ConfigMaps request
func (cms *ConfigMapsService) ListConfigMaps(ctx context.Context, scopeTypeStr string, scopeId string, optionalName string,
	clusterId string, includeAncestors bool,
) (api.ImplResponse, error) {
	// Filter sanity check
	if scopeTypeStr == "" && clusterId == "" {
		return NewErrorResponse(api.NewValidationError("either scopeType or clusterId must be provided")), nil
	}
	if scopeTypeStr != "" && scopeId == "" {
		return NewErrorResponse(api.NewValidationError("scopeId is mandatory when scopeType is provided")), nil
	}

	if err := common.ValidateScopeType(scopeTypeStr); err != nil {
		return NewErrorResponse(err), nil
	}

	var filters models.ConfigMapFilters
	var ancestorsResolver *org_unit.AncestorsResolver

	if clusterId != "" {
		ancestorsResolver = org_unit.NewAncestorsResolverOfCluster(clusterId)
	} else {
		ancestorsResolver = org_unit.NewAncestorsResolver(orgUnitApi.ScopeType(scopeTypeStr), scopeId)
	}

	filters.ScopeQuery = models.NewScopeQueryParams(ancestorsResolver)

	authErr := auth.CheckReadPermission(ctx, authorizationClient.RESOURCETYPE_CM_VOLUME_ASSETS, ancestorsResolver)
	if authErr != nil {
		return NewErrorResponse(authErr), nil
	}

	if optionalName != "" {
		filters.Name = &optionalName
	}

	result, err := cms.dbAccessor.ListConfigMaps(ctx, &filters, &models.ListOptions{
		IncludeAncestors: includeAncestors,
	})
	if err != nil {
		return NewErrorResponse(err), nil
	}

	payload := api.ConfigMapsListResponse{
		ConfigMaps: make([]api.ConfigMapListInfo, 0, len(result)),
	}
	for _, cm := range result {
		configMapInfo := api.ConfigMapListInfo{
			ClusterId: convutils.Deref(cm.ClusterId, ""),
			Name:      convutils.Deref(cm.Name, ""),
		}

		if cm.ScopeQuery != nil {
			configMapInfo.ScopeType = cm.ScopeQuery.ScopeType
			configMapInfo.ScopeId = cm.ScopeQuery.ScopeId
		}

		payload.ConfigMaps = append(payload.ConfigMaps, configMapInfo)
	}

	return api.Response(http.StatusOK, payload), nil
}

// SyncClusterWideConfigmapsV1 handles ConfigMap sync in cluster scope request
func (cms *ConfigMapsService) SyncClusterWideConfigmapsV1(ctx context.Context, clusterId string, configMaps api.ConfigMaps) (api.ImplResponse, error) {
	res, err := cms.syncConfigMaps(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, configMaps)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, *res), nil
}

// SyncDepartmentsWideConfigmapV1 handles ConfigMap sync in department scope request
func (cms *ConfigMapsService) SyncDepartmentsWideConfigmapV1(ctx context.Context, clusterId string, departmentId string, configMaps api.ConfigMaps) (api.ImplResponse, error) {
	res, err := cms.syncConfigMaps(ctx, clusterId, authorizationClient.SCOPETYPE_DEPARTMENT, departmentId, configMaps)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, *res), nil
}

// SyncProjectsConfigmapsV1 handles ConfigMap sync in project scope request
func (cms *ConfigMapsService) SyncProjectsConfigmapsV1(ctx context.Context, clusterId string, namespace string, configMaps api.ConfigMaps) (api.ImplResponse, error) {
	res, err := cms.syncConfigMaps(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, configMaps)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, *res), nil
}

// insertConfigMap inserts a ConfigMap into the ConfigMaps table
func (cms *ConfigMapsService) insertConfigMap(ctx context.Context, clusterId string, scopeType authorizationClient.ScopeType, scopeId string, name string, issues []api.Issue) (*api.ConfigMap, *api.Error) {
	if scopeType == authorizationClient.SCOPETYPE_PROJECT {
		var scopeError *api.Error
		scopeId, scopeError = namespaceToProjectId(ctx, clusterId, scopeId)
		if scopeError != nil {
			return nil, scopeError
		}
	}
	record := models.ConfigMap{
		ConfigMapFilters: models.ConfigMapFilters{
			ClusterId: &clusterId,
			ScopeQuery: &models.ScopeQueryParams{
				ScopeType: string(scopeType),
				ScopeId:   scopeId,
			},
			Name: &name,
		},
	}

	cm, err := cms.dbAccessor.CreateConfigMap(ctx, &record)
	if err != nil {
		return nil, err
	}

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// We do not handle status changes for config-maps
	_, err = HandleResourceIssues(ctx, tenantId, clusterId, cm.ScopeQuery.ScopeType, cm.ScopeQuery.ScopeId, cm.ID.String(), issues, string(api.CLUSTERRESOURCETYPE_CONFIG_MAP))
	if err != nil {
		return nil, err
	}

	return &api.ConfigMap{
		Name: name,
	}, nil
}

// deleteConfigMap deletes a ConfigMap from the ConfigMaps table
func (cms *ConfigMapsService) deleteConfigMap(ctx context.Context, clusterId string, scopeType authorizationClient.ScopeType, scopeId string, name string) *api.Error {
	if scopeType == authorizationClient.SCOPETYPE_PROJECT {
		var scopeError *api.Error
		scopeId, scopeError = namespaceToProjectId(ctx, clusterId, scopeId)
		if scopeError != nil {
			return scopeError
		}
	}

	filters := models.ConfigMapFilters{
		ClusterId: &clusterId,
		ScopeQuery: &models.ScopeQueryParams{
			ScopeType: string(scopeType),
			ScopeId:   scopeId,
		},
		Name: &name,
	}
	return cms.dbAccessor.DeleteConfigMap(ctx, &filters)
}

// syncConfigMap syncs a batch of ConfigMaps to the ConfigMaps table
func (cms *ConfigMapsService) syncConfigMaps(ctx context.Context, clusterId string, scopeType authorizationClient.ScopeType, scopeId string,
	configMaps api.ConfigMaps,
) (*api.ConfigMaps, *api.Error) {
	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	for _, cm := range configMaps.Configmaps {
		if cm.Name == "" {
			return nil, api.NewValidationError("mandatory ConfigMap name is not provided")
		}
	}

	if scopeType == authorizationClient.SCOPETYPE_PROJECT {
		var scopeError *api.Error
		scopeId, scopeError = namespaceToProjectId(ctx, clusterId, scopeId)
		if scopeError != nil {
			return nil, scopeError
		}
	}

	filters := models.ConfigMapFilters{
		TenantId:  tenantId,
		ClusterId: &clusterId,
		ScopeQuery: &models.ScopeQueryParams{
			ScopeType: string(scopeType),
			ScopeId:   scopeId,
		},
	}

	cmIdMap, err := cms.dbAccessor.ApplyConfigMaps(ctx, &filters, configMaps.Configmaps)
	if err != nil {
		return nil, err
	}

	for id, cm := range cmIdMap {
		// We do not handle status changes for config-maps
		_, err = HandleResourceIssues(ctx, tenantId, clusterId, string(scopeType), scopeId, id, cm.Issues, string(api.CLUSTERRESOURCETYPE_CONFIG_MAP))
		if err != nil {
			return nil, err
		}
	}

	return &configMaps, nil
}
