package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_integration"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/workloads/pkg/services"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	clients "github.com/run-ai/runai/control-plane/workloads/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
)

const (
	InferencesResourceType = "inferences"
	NgcBaseURL             = "https://api.ngc.nvidia.com/v2/search/catalog/resources/CONTAINER"
)

type NgcCatalogResponse struct {
	ResultTotal     int32                    `json:"resultTotal,omitempty"`
	ResultPageTotal int32                    `json:"resultPageTotal,omitempty"`
	Params          map[string]interface{}   `json:"params,omitempty"`
	Results         []NgcCatalogResultsInner `json:"results,omitempty"`
}

type NgcCatalogResultsInner struct {
	TotalCount int32                                  `json:"totalCount,omitempty"`
	Resources  []NgcCatalogResultsInnerResourcesInner `json:"resources,omitempty"`
}

type NgcCatalogResultsInnerResourcesInner struct {
	OrgName      string `json:"orgName,omitempty"`
	ResourceId   string `json:"resourceId,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Name         string `json:"name,omitempty"`
	ResourceType string `json:"resourceType,omitempty"`
	Description  string `json:"description,omitempty"`
	IsPublic     bool   `json:"isPublic,omitempty"`
}

type InferencesService struct {
	Conf                 *config.ServerConfig
	SyncConf             *config.WorkloadsSyncConfig
	tenantSettingsClient clients.TenantSettingsClient
	WorkloadsDao         *dao.WorkloadsDao
}

func NewInferencesService(client clients.TenantSettingsClient) api.InferencesAPIServicer {
	conf := runai_config.Get[*config.ServerConfig]()
	return &InferencesService{
		Conf:                 conf,
		SyncConf:             runai_config.Get[*config.WorkloadsSyncConfig](),
		tenantSettingsClient: client,
		WorkloadsDao:         dao.NewWorkloadsDao(conf),
	}
}

func (s *InferencesService) GetNimCatalog(ctx context.Context, filters []string, query string, orderBy string, page int32, pageSize int32) (api.ImplResponse, error) {
	ngcQ := buildNgcQuery(filters, query, orderBy, page, pageSize)
	qBytes, err := json.Marshal(ngcQ)
	if err != nil {
		return api.Response(http.StatusInternalServerError, api.Error{Code: 500, Message: "Failed to encode NGC query"}), nil
	}
	qParam := url.QueryEscape(string(qBytes))
	fullURL := NgcBaseURL + "?q=" + qParam

	body, _, err := doNgcRequest(ctx, fullURL)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("NGC API request failed, using fallback catalog. Error: %v", err)
		fallbackResp := getFallbackNimCatalog()
		return api.Response(http.StatusOK, fallbackResp), nil
	}

	ngcResp, err := parseNgcResponse(body)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("NGC API response parsing failed, using fallback catalog. Error: %v", err)
		fallbackResp := getFallbackNimCatalog()
		return api.Response(http.StatusOK, fallbackResp), nil
	}

	return api.Response(http.StatusOK, ngcResp), nil
}

func buildNgcQuery(filters []string, query string, orderBy string, page int32, pageSize int32) ngcQuery {
	var baseFilters []ngcFilter

	// Add predefined filters for NIMs
	baseFilters = append(baseFilters, ngcFilter{Field: "resourceId", Value: "nim/*"})

	// Add user-provided filters
	for _, f := range filters {
		if parts := strings.SplitN(f, ":", 2); len(parts) == 2 {
			baseFilters = append(baseFilters, ngcFilter{Field: parts[0], Value: parts[1]})
		}
	}

	// Parse orderBy as field,direction pair
	orderByArr := []ngcOrderBy{{Field: "dateCreated", Value: "DESC"}}
	if orderBy != "" {
		parts := strings.Split(orderBy, ",")
		field := parts[0]
		direction := "DESC" // default direction
		if len(parts) > 1 && strings.ToUpper(parts[1]) == "ASC" {
			direction = "ASC"
		}
		orderByArr = []ngcOrderBy{{Field: field, Value: direction}}
	}

	return ngcQuery{
		Filters:  baseFilters,
		OrderBy:  orderByArr,
		Page:     int(page),
		PageSize: int(pageSize),
		Query:    query,
	}
}

func doNgcRequest(ctx context.Context, url string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create NGC API request")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "RunAI-Workloads-Service")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("failed to call NGC API: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to read NGC API response")
	}

	// Check if response is empty
	if len(body) == 0 {
		return body, resp.StatusCode, fmt.Errorf("NGC API returned empty response")
	}

	if resp.StatusCode != http.StatusOK {
		return body, resp.StatusCode, fmt.Errorf("NGC API returned status %d: %s", resp.StatusCode, body)
	}

	return body, http.StatusOK, nil
}

func parseNgcResponse(body []byte) (api.Catalog, error) {
	var ngcResp NgcCatalogResponse
	if err := json.Unmarshal(body, &ngcResp); err != nil {
		return api.Catalog{}, fmt.Errorf("failed to parse NGC API response: %v", err)
	}

	var resources []api.CatalogResourcesInner

	// Aggregate all resources from all results
	for _, result := range ngcResp.Results {
		if result.Resources == nil {
			continue // Skip results with no resources
		}
		for _, ngcResource := range result.Resources {
			// Validate and fix resource ID
			fixedResourceId := validateAndFixResourceId(ngcResource.ResourceId, ngcResource.OrgName)

			resource := api.CatalogResourcesInner{
				OrgName:      ngcResource.OrgName,
				ResourceId:   fixedResourceId,
				DisplayName:  ngcResource.DisplayName,
				Name:         ngcResource.Name,
				ResourceType: ngcResource.ResourceType,
				Description:  ngcResource.Description,
				IsPublic:     ngcResource.IsPublic,
			}
			resources = append(resources, resource)
		}
	}

	// Create new catalog structure
	apiResp := api.Catalog{
		IsStaticData:    false, // Always false for real NGC API responses
		ResultTotal:     ngcResp.ResultTotal,
		ResultPageTotal: ngcResp.ResultPageTotal,
		Params:          api.CatalogParams{}, // Use empty params for now since NGC API params structure is different
		Resources:       resources,
	}

	return apiResp, nil
}

// validateAndFixResourceId validates that the resource ID has 2 parts when split by "/".
// If it has 3 parts and the first part equals the orgName, it removes the first part.
func validateAndFixResourceId(resourceId string, orgName string) string {
	if resourceId == "" {
		return resourceId
	}

	parts := strings.Split(resourceId, "/")

	// If resource ID has exactly 2 parts, it's already in the correct format
	if len(parts) == 2 {
		return resourceId
	}

	// If resource ID has 3 parts and the first part equals the orgName, remove the first part
	if len(parts) == 3 && parts[0] == orgName {
		return parts[1] + "/" + parts[2]
	}

	// For any other case, return the original resource ID
	return resourceId
}

func (s *InferencesService) GetInferenceAncestors(ctx context.Context, resourceId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	ancestors, err := authorization_utils.GetAncestorsCache(ctx, InferencesResourceType, resourceId)
	if err == nil {
		return api.Response(http.StatusOK, ancestors), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Could not find inference or revision <%s> ancestors in cache. err: %v, continuing", resourceId, err)

	workload, apiErr := services_integration.GetWorkloadDetails(ctx, resourceId)
	if apiErr != nil && apiErr.Code != http.StatusNotFound {
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	if apiErr == nil {
		ancestors = &authorization_utils.Ancestors{
			TenantId:     fmt.Sprintf("%d", workload.TenantID),
			ClusterUuid:  ptr.To(workload.ClusterID.String()),
			DepartmentId: ptr.To(workload.DepartmentID),
			ProjectId:    ptr.To(workload.ProjectID),
		}
	} else {
		revision, dbErr := dao.DbGetRevision(ctx, tenantId, resourceId)
		if dbErr != nil {
			return api.Response(int(dbErr.Code), dbErr.Message), nil
		}

		ancestors = &authorization_utils.Ancestors{
			TenantId:     fmt.Sprintf("%d", revision.TenantID),
			ClusterUuid:  ptr.To(revision.ClusterID.String()),
			DepartmentId: ptr.To(revision.DepartmentID),
			ProjectId:    ptr.To(revision.ProjectID),
		}
	}

	authorization_utils.SetAncestorsCache(ctx, InferencesResourceType, resourceId, *ancestors)

	return api.Response(http.StatusOK, *ancestors), nil
}

func (s *InferencesService) CreateInference(ctx context.Context, request api.InferenceCreationRequest) (api.ImplResponse, error) {
	return CreateInference(ctx, s.tenantSettingsClient, request)
}

func (s *InferencesService) DeleteInference(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	res, err := ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_DELETED)
	if res.Code >= 200 && res.Code < 300 {
		res.Body = nil // When deleting a single workload, we do not return a response body for 2xx status codes.
	}
	return res, err
}

func (s *InferencesService) GetInference(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return GetWorkloadById(ctx, Inference, workloadId)
}

func (s *InferencesService) UpdateInferenceSpec(ctx context.Context, workloadId string, inferenceUpdateRequest api.InferenceUpdateRequest) (api.ImplResponse, error) {
	return UpdateInferenceSpec(ctx, workloadId, s.tenantSettingsClient, inferenceUpdateRequest)
}

func (s *InferencesService) ListInferencesUpdates(ctx context.Context, afterUpdate int64, clusterId string) (api.ImplResponse, error) {
	return ListUpdates(ctx, Inference, afterUpdate, clusterId, s.SyncConf, nil)
}

func (s *InferencesService) RequestInferencesUpdates(ctx context.Context, afterUpdate int64, clusterId string, syncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return ListUpdates(ctx, Inference, afterUpdate, clusterId, s.SyncConf, &syncReport)
}

func (s *InferencesService) GetInferenceWorkloadMetrics(
	ctx context.Context,
	workloadId string,
	metricTypes []api.InferenceWorkloadMetricType,
	start time.Time,
	end time.Time,
	numberOfSamples int32,
) (api.ImplResponse, error) {
	return services.GetInferenceWorkloadMetrics(ctx, s.WorkloadsDao, workloadId, metricTypes, start, end, numberOfSamples)
}

func (s *InferencesService) GetInferenceWorkloadPodMetrics(
	ctx context.Context,
	workloadId string,
	podId string,
	metricTypes []api.InferencePodMetricType,
	start time.Time,
	end time.Time,
	numberOfSamples int32,
) (api.ImplResponse, error) {
	return services.GetInferenceWorkloadPodMetrics(ctx, workloadId, podId, metricTypes, start, end, numberOfSamples)
}
