package cache

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	datavolumes_config "github.com/run-ai/runai/control-plane/datavolumes-service/pkg/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	org_unit_service_client "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"k8s.io/utils/ptr"
)

const (
	ScopeNameField = "datavolumes-scopeName"
)

type OrgHierarchyManager struct {
	OrgUnitServiceClient *org_unit_service_client.APIClient
	ClustersClient       *cluster_service_client.APIClient
}

var once sync.Once

var mgr *OrgHierarchyManager

func newOrgHierarchyManager() *OrgHierarchyManager {
	logging.Debugf("initializing OrgHierarchyManager")
	orgHierarchyManagerConfig, err := runai_config.GetR[*datavolumes_config.OrgHierarchyManagerConfig]()
	if err != nil {
		logging.Fatalf("failed to get org hierarchy manager config: %v", err)
	}

	return &OrgHierarchyManager{
		OrgUnitServiceClient: org_unit_service_client.NewAPIClient(&org_unit_service_client.Configuration{
			Servers: org_unit_service_client.ServerConfigurations{
				{
					URL: orgHierarchyManagerConfig.OrgUnitServiceUrl,
				},
			},
		}),
		ClustersClient: cluster_service_client.NewAPIClient(&cluster_service_client.Configuration{
			Servers: cluster_service_client.ServerConfigurations{
				{
					URL: orgHierarchyManagerConfig.ClusterServiceUrl,
				},
			},
		}),
	}
}

func GetOrgHierarchyManager() *OrgHierarchyManager {
	once.Do(func() {
		mgr = newOrgHierarchyManager()
	})

	return mgr
}

// TODO: create an API for multi?
func (mgr *OrgHierarchyManager) GetAncestorsMulti(ctx context.Context, scopeIds []string, scopeType api.ScopeType) ([]*authorization_utils.Ancestors, error) {
	result := make([]*authorization_utils.Ancestors, len(scopeIds))
	for i, scopeId := range scopeIds {
		ancestors, _, err := mgr.GetAncestors(ctx, scopeId, scopeType)
		if err != nil {
			return nil, err
		}
		result[i] = ancestors
	}
	return result, nil
}

// GetAncestors - will also populate self scope as ancestor;
// this will be used later when validating RBAC in permitted scopes.
func (mgr *OrgHierarchyManager) GetAncestors(ctx context.Context, scopeId string, scopeType api.ScopeType) (ancestors *authorization_utils.Ancestors, cacheHit bool, err error) {
	ancestors, cacheHit, err = mgr.getAncestors(ctx, scopeId, scopeType)
	if err != nil || ancestors == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get ancestors for %v %s. err: %v", scopeType, scopeId, err)
		return ancestors, cacheHit, err
	}

	switch scopeType {
	case api.SCOPETYPE_PROJECT:
		ancestors.ProjectId = ptr.To(scopeId)
	case api.SCOPETYPE_DEPARTMENT:
		ancestors.DepartmentId = ptr.To(scopeId)
	case api.SCOPETYPE_CLUSTER:
		ancestors.ClusterUuid = ptr.To(scopeId)
	}

	return ancestors, cacheHit, err
}

func (mgr *OrgHierarchyManager) getAncestors(ctx context.Context, scopeId string, scopeType api.ScopeType) (ancestors *authorization_utils.Ancestors, cacheHit bool, err error) {
	ancestors, err = authorization_utils.GetAncestorsCache(ctx, string(scopeType), scopeId)
	if err == nil && ancestors != nil {
		return ancestors, true, nil
	}

	if ancestors == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache miss. %v %s ancestors are not in cache, calculating it synchronously.", scopeType, scopeId)
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Got an error trying to fetch the ancestor list of %v %s from cache, calculating it synchronously. err: %v", scopeType, scopeId, err)
	}

	ancestors, err = mgr.getAncestorsFromService(ctx, scopeId, scopeType)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error trying to fetch the ancestor list of %v %s from server. err: %v", scopeType, scopeId, err)
		return nil, false, err
	}

	return ancestors, false, nil
}

func (mgr *OrgHierarchyManager) getAncestorsFromService(ctx context.Context, scopeId string, scopeType api.ScopeType) (*authorization_utils.Ancestors, error) {
	switch scopeType {
	case api.SCOPETYPE_PROJECT, api.SCOPETYPE_DEPARTMENT:
		return mgr.getOrgUnitAncestorsFromService(ctx, scopeId, scopeType)
	case api.SCOPETYPE_CLUSTER:
		return mgr.getClusterAncestorsFromService(ctx, scopeId)
	default:
		return nil, fmt.Errorf("unsupported scope type %s", scopeType)
	}
}

func (mgr *OrgHierarchyManager) getOrgUnitAncestorsFromService(ctx context.Context, scopeId string, scopeType api.ScopeType) (*authorization_utils.Ancestors, error) {
	var err error
	var response *http.Response
	var ancestors *org_unit_service_client.Ancestors
	switch scopeType {
	case api.SCOPETYPE_PROJECT:
		request := mgr.OrgUnitServiceClient.ProjectsAPI.GetProjectAncestors(ctx, scopeId)
		ancestors, response, err = mgr.OrgUnitServiceClient.ProjectsAPI.GetProjectAncestorsExecute(request)
	case api.SCOPETYPE_DEPARTMENT:
		request := mgr.OrgUnitServiceClient.DepartmentsAPI.GetDepartmentAncestors(ctx, scopeId)
		ancestors, response, err = mgr.OrgUnitServiceClient.DepartmentsAPI.GetDepartmentAncestorsExecute(request)
	}

	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get %v <%s> ancestors, response status code: %d",
			scopeType, scopeId, response.StatusCode)
	}

	return &authorization_utils.Ancestors{
		TenantId:     ancestors.TenantId,
		ClusterUuid:  ancestors.ClusterUuid.Get(),
		DepartmentId: ancestors.DepartmentId.Get(),
		ProjectId:    ancestors.ProjectId.Get(),
	}, nil
}

func (mgr *OrgHierarchyManager) getClusterAncestorsFromService(ctx context.Context, clusterId string) (*authorization_utils.Ancestors, error) {
	request := mgr.ClustersClient.ClustersAPI.GetClusterAncestors(ctx, clusterId)
	ancestors, response, err := mgr.ClustersClient.ClustersAPI.GetClusterAncestorsExecute(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get cluster <%s> ancestors, response status code: %d",
			clusterId, response.StatusCode)
	}

	return &authorization_utils.Ancestors{
		TenantId:    ancestors.TenantId,
		ClusterUuid: ancestors.ClusterUuid.Get(),
	}, nil
}

func (mgr *OrgHierarchyManager) GetScopesNames(ctx context.Context, scopes []api.SharedScopeBase, clusterId string) (map[api.SharedScopeBase]string, *api.Error) {
	// TODO: use a new API for multi-scope-get-names?
	// TODO: make sure the "GetScopesNames" is logged
	// TODO: make sure all errors here are logged!! GetAncestros as well.

	result := make(map[api.SharedScopeBase]string)

	for _, scope := range scopes {
		scopeName, _, apiErr := mgr.GetScopeName(ctx, scope.ScopeId, scope.ScopeType, clusterId)
		if apiErr != nil {
			return nil, apiErr
		}
		result[scope] = scopeName
	}

	return result, nil
}

func (mgr *OrgHierarchyManager) GetScopeName(ctx context.Context, scopeId string, scopeType api.ScopeType, clusterId string) (string, bool, *api.Error) {

	scopeName, err := cache.Get[string](ctx, GetScopeNameKey(scopeId, scopeType, clusterId))
	if err == nil && scopeName != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache hit. Scope %s/%s name is in cache.", string(scopeType), scopeId)
		return *scopeName, true, nil
	}

	if scopeName == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache miss. Scope %s/%s name are not in cache, calculating it synchronously.", string(scopeType), scopeId)
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Got an error trying to fetch the name of scope %s/%s from cache, calculating it synchronously. err: %v",
			string(scopeType), scopeId, err)
	}

	var apiErr *api.Error
	var name string
	name, apiErr = mgr.getScopeNameFromServiceAndSaveToCache(ctx, scopeId, scopeType, clusterId)
	return name, false, apiErr
}

func (mgr *OrgHierarchyManager) getScopeNameFromServiceAndSaveToCache(ctx context.Context, scopeId string, scopeType api.ScopeType, clusterId string) (string, *api.Error) {
	scopeName, apiErr := mgr.getScopeNameFromService(ctx, scopeId, scopeType, clusterId)
	if apiErr != nil {
		return "", apiErr
	}

	if err := cache.Set[string](ctx, GetScopeNameKey(scopeId, scopeType, clusterId), scopeName); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Failed to save scope name to cache. err: %v", err)
	}
	return scopeName, nil
}

func (mgr *OrgHierarchyManager) getScopeNameFromService(ctx context.Context, scopeId string, scopeType api.ScopeType, clusterId string) (string, *api.Error) {
	errMsg := fmt.Sprintf("Failed to get the scope name of resource %s/%s", string(scopeType), scopeId)
	var name string
	var response *http.Response
	var err error
	switch scopeType {
	case api.SCOPETYPE_CLUSTER:
		if scopeId != clusterId {
			return "", api.NewError(http.StatusBadRequest, "Requested cluster scope is not in original datavolume's cluster", nil)
		}
		name, response, err = mgr.getClusterName(ctx, scopeId)
	default:
		name, response, err = mgr.getOrgUnitNameInCluster(ctx, scopeId, scopeType, clusterId)
	}

	if err != nil || (response != nil && response.StatusCode >= http.StatusBadRequest) {
		if response != nil && response.StatusCode >= http.StatusBadRequest {
			if response.StatusCode == http.StatusNotFound {
				return "", api.NewError(http.StatusNotFound, fmt.Sprintf("Scope %s/%s not found.", string(scopeType), scopeId), nil)
			}
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error response from client on get scope name of resource %s/%s. status: %s",
				string(scopeType), scopeId, response.Status)

			return "", api.NewError(http.StatusInternalServerError, errMsg, nil)
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error from client on get scope name of resource %s/%s. error: %+v",
			string(scopeType), scopeId, err)
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

func (mgr *OrgHierarchyManager) getOrgUnitNameInCluster(ctx context.Context, scopeId string, scopeType api.ScopeType, clusterId string) (string, *http.Response, error) {
	namesSlice, resp, err := mgr.OrgUnitServiceClient.OrgUnitAPI.GetOrgUnitsNames(ctx).OrgUnitType(string(scopeType)).OrgUnitIds(scopeId).Execute()
	if len(namesSlice) > 0 {
		return namesSlice[0].Name, resp, err
	}

	return "", resp, err
}

func GetScopeNameKey(scopeId string, scopeType api.ScopeType, clusterId string) string {
	return fmt.Sprintf("%s-%s-%s-%s", scopeType, scopeId, clusterId, ScopeNameField)
}
