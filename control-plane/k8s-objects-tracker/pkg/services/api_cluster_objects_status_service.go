package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"k8s.io/utils/ptr"

	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/common"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"golang.org/x/exp/maps"
)

// ClusterObjectsStatusApiService is a service that implements the logic for the ClusterObjectsStatusApiService
type ClusterObjectsStatusApiService struct {
	dbAccessor dao.ConfigMapsDBAccessor
}

// NewClusterObjectsStatusApiService creates a default api service
func NewClusterObjectsStatusApiService(dbAccessor dao.ConfigMapsDBAccessor) api.ClusterObjectsStatusAPIServicer {
	return &ClusterObjectsStatusApiService{
		dbAccessor: dbAccessor,
	}
}

// ListClusterObjectsStatus - list of objects in the cluster and their found issues
func (cos *ClusterObjectsStatusApiService) ListClusterObjectsStatus(ctx context.Context, resourceType api.ClusterResourceType, secretType string,
	scopeId string, scopeType string) (api.ImplResponse, error) {
	if scopeId == "" && scopeType != "" {
		return api.Response(http.StatusInternalServerError, nil), errors.New("scopeId must be provided when scopeType is provided")
	}

	if scopeType == "" && scopeId != "" {
		return api.Response(http.StatusInternalServerError, nil), errors.New("scopeType must be provided when scopeId is provided")
	}

	if err := common.ValidateScopeType(scopeType); err != nil {
		return NewErrorResponse(err), nil
	}

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return api.Response(http.StatusUnauthorized, nil), errors.New(err.Message)
	}

	var scopeQuery *models.ScopeQueryParams
	if scopeType != "" && scopeId != "" {
		scopeQuery = &models.ScopeQueryParams{
			ScopeType: scopeType,
			ScopeId:   scopeId,
		}
	}

	clusterObjectIdMap, err := getClusterObjects(ctx, tenantId, string(resourceType), secretType, scopeQuery, &models.ListOptions{
		IncludeAncestors: true, /* backward compatibility with the existing API */
		IncludePrivate:   true,
	})
	if err != nil {
		return api.Response(int(err.Code), nil), errors.New(err.Message)
	}

	clusterObjectsIdSet := make([]string, 0, len(clusterObjectIdMap))
	for id := range clusterObjectIdMap {
		clusterObjectsIdSet = append(clusterObjectsIdSet, id)
	}

	clusterObjectIssuesIdMap, err := getClusterObjectsIssues(ctx, clusterObjectsIdSet)
	if err != nil {
		return api.Response(int(err.Code), nil), errors.New(err.Message)
	}

	patchResourceIssues(clusterObjectIdMap, clusterObjectIssuesIdMap)

	response := api.ClusterObjectsStatus{
		Entries: maps.Values(clusterObjectIdMap),
	}

	return api.Response(http.StatusOK, response), nil
}

// HandleResourceIssues handles the issues of a resource.
// returns the issues changes and an error if any.
func HandleResourceIssues(ctx context.Context,
	tenantId int32,
	clusterId string,
	resourceScopeType string,
	resourceScopeId string,
	resourceId string,
	issues []api.Issue,
	resourceType string) (
	*dao.IssuesChanges,
	*api.Error,
) {
	resourceIssues, err := prepareResourceIssues(ctx, tenantId, clusterId, resourceScopeType, resourceScopeId, resourceId, resourceType, issues)
	if err != nil {
		return nil, err
	}

	issuesChanges, err := dao.UpdateResourceIssues(ctx, tenantId, clusterId, resourceId, resourceIssues)
	if err != nil {
		return nil, err
	}
	return issuesChanges, nil
}

func prepareResourceIssues(ctx context.Context, tenantId int32, clusterId string, resourceScopeType string, resourceScopeId string,
	resourceId string, resourceType string, issues []api.Issue,
) ([]models.ResourceIssue, *api.Error) {
	resourceIssues := make([]models.ResourceIssue, 0, len(issues))
	for _, issue := range issues {
		issueClusterId := clusterId
		filters := models.ResourceIssueFilters{
			TenantId:   tenantId,
			ClusterId:  &issueClusterId,
			ResourceId: &resourceId,
			Type:       (*api.ClusterResourceType)(&resourceType),
		}
		if issue.Namespace == NamespaceRunai {
			// If an issue is in "runai" namespace it is in the scope of the resource it is related to
			filters.ScopeType = ptr.To(resourceScopeType)
			filters.ScopeId = ptr.To(resourceScopeId)
		} else {
			filters.ScopeType = ptr.To(string(authorizationClient.SCOPETYPE_PROJECT))
			scopeId, err := namespaceToProjectId(ctx, clusterId, issue.Namespace)
			if err != nil {
				return nil, err
			}
			filters.ScopeId = &scopeId
		}
		currIssue := issue.Issue
		resourceIssue := models.ResourceIssue{
			ResourceIssueFilters: filters,
			Issue:                currIssue,
		}
		resourceIssues = append(resourceIssues, resourceIssue)
	}
	return resourceIssues, nil
}

func getClusterObjects(ctx context.Context, tenantId int32, resourceType string, secretType string,
	scopeQuery *models.ScopeQueryParams, options *models.ListOptions) (map[string]api.ClusterObjectStatus, *api.Error) {
	clusterObjectIdMap := make(map[string]api.ClusterObjectStatus)

	if resourceType == "" || resourceType == string(api.CLUSTERRESOURCETYPE_PVC) {
		filters := models.PvcFilters{
			TenantId:   tenantId,
			ScopeQuery: scopeQuery,
		}

		pvcs, err := dao.ListPvcs(ctx, &filters, options)
		if err != nil {
			return nil, err
		}

		for _, pvc := range pvcs {
			clusterObject := api.ClusterObjectStatus{
				ResourceName: *pvc.PvcName,
				ResourceType: api.CLUSTERRESOURCETYPE_PVC,
			}

			if pvc.ScopeQuery != nil {
				clusterObject.ScopeType = pvc.ScopeQuery.ScopeType
				clusterObject.ScopeId = pvc.ScopeQuery.ScopeId
			}

			clusterObjectIdMap[pvc.ID.String()] = clusterObject
		}
	}

	if resourceType == "" || resourceType == string(api.CLUSTERRESOURCETYPE_SECRET) {
		filters := models.SecretFilters{
			TenantId:   tenantId,
			ScopeQuery: scopeQuery,
		}

		if secretType != "" {
			filters.SecretType = &secretType
		}

		secrets, err := dao.ListSecrets(ctx, &filters, options)
		if err != nil {
			return nil, err
		}

		for _, secret := range secrets {
			secretInfo := api.ClusterObjectStatus{
				ResourceName: *secret.SecretName,
				ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
				SecretType:   api.SecretType(*secret.SecretType),
			}

			if secret.ScopeQuery != nil {
				secretInfo.ScopeType = secret.ScopeQuery.ScopeType
				secretInfo.ScopeId = secret.ScopeQuery.ScopeId
			}

			// Private credentials are stored as cluster scope in the k8s-objects-tracker,
			// In order to support the private credetials status we need to set the scope type and id to the tenant (since the assets are tenant scoped).
			if convutils.BoolIsTrue(secret.Private) {
				secretInfo.ScopeType = string(authorizationClient.SCOPETYPE_TENANT)
				secretInfo.ScopeId = fmt.Sprintf("%d", tenantId)
			}

			clusterObjectIdMap[secret.ID.String()] = secretInfo
		}
	}

	if resourceType == "" || resourceType == string(api.CLUSTERRESOURCETYPE_CONFIG_MAP) {
		filters := models.ConfigMapFilters{
			TenantId:   tenantId,
			ScopeQuery: scopeQuery,
		}

		cms, err := dao.NewConfigMapsDBAccessor().ListConfigMaps(ctx, &filters, options)
		if err != nil {
			return nil, err
		}

		for _, cm := range cms {
			clusterInfo := api.ClusterObjectStatus{
				ResourceName: *cm.Name,
				ResourceType: api.CLUSTERRESOURCETYPE_CONFIG_MAP,
			}

			if cm.ScopeQuery != nil {
				clusterInfo.ScopeType = cm.ScopeQuery.ScopeType
				clusterInfo.ScopeId = cm.ScopeQuery.ScopeId

			}

			clusterObjectIdMap[cm.ID.String()] = clusterInfo
		}
	}

	return clusterObjectIdMap, nil
}

func getClusterObjectsIssues(ctx context.Context, resourceIds []string) (map[string][]api.ClusterObjectIssue, *api.Error) {
	clusterObjectIssuesIdMap := make(map[string][]api.ClusterObjectIssue)

	issues, err := dao.ListResourceIssuesByIds(ctx, resourceIds)
	if err != nil {
		return nil, err
	}

	for _, issue := range issues {
		clusterObjectIssuesIdMap[*issue.ResourceId] = append(clusterObjectIssuesIdMap[*issue.ResourceId], api.ClusterObjectIssue{
			ScopeType: *issue.ScopeType,
			ScopeId:   *issue.ScopeId,
			Issue:     api.ClusterObjectIssueType(issue.Issue),
		})
	}

	return clusterObjectIssuesIdMap, nil
}

func patchResourceIssues(clusterObjectIdMap map[string]api.ClusterObjectStatus, clusterObjectIssuesIdMap map[string][]api.ClusterObjectIssue) {
	for id, clusterObjectStatus := range clusterObjectIdMap {
		clusterObjectStatus.Issues = clusterObjectIssuesIdMap[id]
		clusterObjectIdMap[id] = clusterObjectStatus
	}
}
