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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/utils/ptr"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/versions"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	apiClusters "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/settings"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
)

const (
	CpuResourcesFlag      = "project.quota.cpu_resources"
	DefaultDepartmentName = "default"

	// OrgNameRegex - starts and ends with a digit or letter; in between could be a digit, a letter or a hyphen;
	// max length 63; has to include at least one letter; could also be a single letter.
	OrgNameRegex         = `^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$`
	CsvAccept            = "text/csv"
	DepartmentAdminRole  = "Department administrator"
	ResearchManagerRole  = "Research manager"
	ResearcherRole       = "L1 researcher"
	MlEngineerRole       = "ML engineer"
	L2ResearcherRole     = "L2 researcher"
	EditorRole           = "Editor"
	SystemAdminRole      = "System administrator"
	ApplicationAdminRole = "Application administrator"

	NodePoolToOrgUnitConstraint = "pool_to_org_unit_data"

	// MaxBatchPatchNodePoolResourcesRequests limits the number of node pool resources patch operations per request
	MaxBatchPatchNodePoolResourcesRequests = 100
	MaxBatchDeleteReportsRequests          = 100
)

var verbosityMap = map[string]models.Verbosity{
	"meta":    models.MetaDataVerbosity, // Data from department table only
	"brief":   models.SlimVerbosity,     // Department spec
	"verbose": models.VerboseVerbosity,  // Department spec + children + status
}

func HandleError(err error, status int, message string) api.ImplResponse {
	newError := api.NewError(status, message, err)
	return HandleApiError(*newError)
}

func HandleApiError(err api.Error) api.ImplResponse {
	return api.ImplResponse{
		Code: int(err.Code),
		Body: err,
	}
}

func HandleDbError(ctx context.Context, err error, message string) api.ImplResponse {
	return HandleDbErrorWithDetails(ctx, err, message, "")
}

func HandleDbErrorWithDetails(ctx context.Context, err error, message string, details string) api.ImplResponse {
	dbErr := api.NewError(http.StatusInternalServerError, message, err)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("DB failed: %s, %v", message, err)
	if errors.Is(err, orm.ErrDuplicatedKey) {
		dbErr.Code = http.StatusConflict
	} else if errors.Is(err, orm.ErrRecordNotFound) {
		dbErr.Code = http.StatusNotFound
	}
	if details != "" {
		dbErr.Details = details
	} else {
		dbErr.WithDetails(err)
	}
	return api.ImplResponse{
		Code: int(dbErr.Code),
		Body: *dbErr,
	}
}

func UpdateOrgUnitDataResources(newResources []models.Resources, orgUnitData *models.OrgUnitData) {
	resourcesMap := make(map[string]models.Resources)
	for _, resource := range newResources {
		resourcesMap[resource.NodePoolId] = resource
	}

	resources := orgUnitData.Resources
	for i, resource := range resources {
		newResource := resourcesMap[resource.NodePoolId]
		resources[i].GpuLimit = newResource.GpuLimit
		resources[i].GpuDeserved = newResource.GpuDeserved
		resources[i].GpuOverQuotaWeight = newResource.GpuOverQuotaWeight
		resources[i].CpuLimit = newResource.CpuLimit
		resources[i].CpuDeserved = newResource.CpuDeserved
		resources[i].CpuOverQuotaWeight = newResource.CpuOverQuotaWeight
		resources[i].MemoryLimit = newResource.MemoryLimit
		resources[i].MemoryDeserved = newResource.MemoryDeserved
		resources[i].MemoryOverQuotaWeight = newResource.MemoryOverQuotaWeight
		resources[i].Rank = newResource.Rank
	}
}

func PatchOrgUnitDataResources(newResources []models.Resources, orgUnitData *models.OrgUnitData) error {
	resourcesMap := make(map[string]*models.Resources)
	for _, resource := range newResources {
		resourcesMap[resource.NodePoolId] = ptr.To(resource)
	}

	resources := orgUnitData.Resources
	for i, resource := range resources {
		newResource := resourcesMap[resource.NodePoolId]
		if newResource != nil {
			rank, err := ParseRank(newResource.Rank)
			if err != nil {
				return err
			}
			resources[i].GpuLimit = newResource.GpuLimit
			resources[i].GpuDeserved = newResource.GpuDeserved
			resources[i].GpuOverQuotaWeight = newResource.GpuOverQuotaWeight
			resources[i].CpuLimit = newResource.CpuLimit
			resources[i].CpuDeserved = newResource.CpuDeserved
			resources[i].CpuOverQuotaWeight = newResource.CpuOverQuotaWeight
			resources[i].MemoryLimit = newResource.MemoryLimit
			resources[i].MemoryDeserved = newResource.MemoryDeserved
			resources[i].MemoryOverQuotaWeight = newResource.MemoryOverQuotaWeight
			resources[i].MemoryUnits = newResource.MemoryUnits
			resources[i].Rank = rank
		}
	}
	return nil
}

// PatchOrgUnitDataResourcesPartial applies a partial patch of org unit resources:
// for each node pool present in newResources, it inspects the original patchRequests
// to determine which groups (Gpu/Cpu/Memory/Rank) were explicitly touched. Touched groups
// are copied exactly from newResources (nil means clear); untouched groups are left as-is.
func PatchOrgUnitDataResourcesPartial(
	newResources []models.Resources,
	existingResources []models.Resources,
	patchRequests []api.NodePoolResourcesPatchRequest,
) error {
	resourcesMap := make(map[string]*models.Resources)
	for _, resource := range newResources {
		resourcesMap[resource.NodePoolId] = ptr.To(resource)
	}

	patchRequestsPerNodePool := map[string]api.NodePoolResourcesPatchRequest{}
	for _, patchRequest := range patchRequests {
		patchRequestsPerNodePool[patchRequest.NodePoolId] = patchRequest
	}

	for i, resource := range existingResources {
		newResource := resourcesMap[resource.NodePoolId]
		if newResource == nil {
			continue
		}

		patchRequest, ok := patchRequestsPerNodePool[resource.NodePoolId]
		if !ok {
			continue
		}

		// Rank: update only when explicitly touched
		if patchRequest.Rank != nil && newResource.Rank != nil {
			rank, err := ParseRank(newResource.Rank)
			if err != nil {
				return err
			}
			existingResources[i].Rank = rank
		}

		if patchRequest.Resources == nil {
			continue
		}

		// GPU: update when touched (values may be nil to clear)
		if patchRequest.Resources.Gpu != nil {
			existingResources[i].GpuLimit = newResource.GpuLimit
			existingResources[i].GpuDeserved = newResource.GpuDeserved
			existingResources[i].GpuOverQuotaWeight = newResource.GpuOverQuotaWeight
		}

		// CPU: update when touched (values may be nil to clear)
		if patchRequest.Resources.Cpu != nil {
			existingResources[i].CpuLimit = newResource.CpuLimit
			existingResources[i].CpuDeserved = newResource.CpuDeserved
			existingResources[i].CpuOverQuotaWeight = newResource.CpuOverQuotaWeight
		}

		// Memory: update when touched (values may be nil to clear)
		if patchRequest.Resources.Memory != nil {
			existingResources[i].MemoryLimit = newResource.MemoryLimit
			existingResources[i].MemoryDeserved = newResource.MemoryDeserved
			existingResources[i].MemoryOverQuotaWeight = newResource.MemoryOverQuotaWeight
			existingResources[i].MemoryUnits = newResource.MemoryUnits
		}
	}
	return nil
}

func GetVerbosity(verbosity string) (models.Verbosity, error) {
	if verbosity == "" {
		return models.SlimVerbosity, nil
	}
	if val, ok := verbosityMap[verbosity]; ok {
		return val, nil
	}
	return 0, errors.New("verbosity invalid")
}

func ValidateResourcesUpdateRequest(ctx context.Context, resources []api.ResourcesNullable, nodePoolsMetadata map[string]apiClusters.NodepoolMinimal) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating resources request")
	err := ValidateResources(ctx, resources, nodePoolsMetadata)
	if err != nil {
		return err
	}
	return ValidateNodePools(ctx, resources)
}

func ValidateResourcesPatchRequest(ctx context.Context, resources []api.ResourcesNullable, nodePoolsMetadata map[string]apiClusters.NodepoolMinimal) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating resources patch request")
	err := ValidateResources(ctx, resources, nodePoolsMetadata)
	if err != nil {
		return err
	}
	return ValidateNodePools(ctx, resources)
}

func ValidateNodePoolResourcesPatchRequest(ctx context.Context, resources api.ResourcesNullable,
	nodePoolsMetadata map[string]apiClusters.NodepoolMinimal,
) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating node pool resources patch request")
	_, err := validateNodePoolResources(resources, nodePoolsMetadata)
	return err
}

func StoredGpuDeserved(resources []models.Resources, nodePoolId string) *float64 {
	for _, resource := range resources {
		if resource.NodePoolId == nodePoolId {
			return resource.GpuDeserved
		}
	}
	return nil
}

func ValidateNodePools(ctx context.Context, resources []api.ResourcesNullable) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Validating creation resources' node pools")
	if len(resources) <= 1 {
		return nil
	}
	resourcesMap := make(map[string]api.ResourcesNullable)
	for i, resource := range resources {
		if resource.NodePool == nil {
			logging.Errorf("resources are not valid. no node pool for resources in index %d", i)
			return fmt.Errorf("resources are not valid. no node pool for resources in index %d", i)
		}
		if resource.NodePool.Id == "" {
			logging.Errorf("resources are not valid. no nodepool id for resources in index %d", i)
			return fmt.Errorf("resources are not valid. no nodepool id for resources in index %d", i)
		}

		_, found := resourcesMap[resource.NodePool.Id]
		if found {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources are not valid. more than one resources for node pool %s", resource.NodePool.Name)
			return fmt.Errorf("resources are not valid. more than one resources for node pool %s", resource.NodePool.Name)
		}
		resourcesMap[resource.NodePool.Id] = resource
	}
	return nil
}

func validateNodePoolResources(nodePoolResources api.ResourcesNullable, nodePoolsMetadata map[string]apiClusters.NodepoolMinimal) ([][]*float64, error) {
	deservedLimitPairs := [][]*float64{}

	npSchedulingCapabilities, nodepoolId, err := getSchedulingCapabilitiesPerNodePool(nodePoolsMetadata, nodePoolResources.NodePool)
	if err != nil {
		return nil, err
	}
	adjustableWeightsFlag := npSchedulingCapabilities.AdjustableWeights
	cpuQuotaFlag := npSchedulingCapabilities.CpuQuota
	quotaFlag := npSchedulingCapabilities.Quota
	ranksFlag := npSchedulingCapabilities.Ranks

	if quotaFlag && nodePoolResources.Gpu.Deserved == nil {
		return nil, fmt.Errorf("quota is enabled on nodepool id <%s>. GPU deserved is required", nodepoolId)
	}

	gpuDeserved := ptr.Deref(nodePoolResources.Gpu.Deserved, 0)
	if !quotaFlag && (gpuDeserved > 0 || (nodePoolResources.Gpu.Limit != nil && *nodePoolResources.Gpu.Limit != -1)) {
		return nil, fmt.Errorf("quota is not enabled on nodepool id <%s>. GPU deserved should be 0 or omitted, and GPU limit should be -1, null or omitted", nodepoolId)
	}

	if !adjustableWeightsFlag && nodePoolResources.Gpu.OverQuotaWeight != nil && !IsValidNonAdjustableOverQuotaWeight(*nodePoolResources.Gpu.OverQuotaWeight, &gpuDeserved) {
		return nil, errors.New("over quota feature is off. OverQuotaWeight should be deserved or empty")
	}

	rank, err := ParseRank(nodePoolResources.Rank)
	if err != nil {
		return nil, err
	}
	if !ranksFlag && rank != nil && *rank != DefaultResourceRank {
		return nil, fmt.Errorf("ranks is not enabled on nodepool id <%s>. Rank should be empty", nodepoolId)
	}

	deservedLimitPairs = append(deservedLimitPairs, []*float64{ptr.To(gpuDeserved), nodePoolResources.Gpu.Limit})

	if cpuQuotaFlag {
		if nodePoolResources.Cpu != nil {
			if !adjustableWeightsFlag && nodePoolResources.Cpu.OverQuotaWeight != nil && !IsValidNonAdjustableOverQuotaWeight(*nodePoolResources.Cpu.OverQuotaWeight, nodePoolResources.Cpu.Deserved) {
				return nil, errors.New("over quota feature is off. OverQuotaWeight should be deserved or empty")
			}
			deservedLimitPairs = append(deservedLimitPairs, []*float64{nodePoolResources.Cpu.Deserved, nodePoolResources.Cpu.Limit})
		}
		if nodePoolResources.Memory != nil {
			if !adjustableWeightsFlag && nodePoolResources.Memory.OverQuotaWeight != nil && !IsValidNonAdjustableOverQuotaWeight(*nodePoolResources.Memory.OverQuotaWeight, nodePoolResources.Memory.Deserved) {
				return nil, errors.New("over quota feature is off. OverQuotaWeight should be deserved or empty")
			}
			deservedLimitPairs = append(deservedLimitPairs, []*float64{nodePoolResources.Memory.Deserved, nodePoolResources.Memory.Limit})
		}
	} else if (nodePoolResources.Cpu != nil && nodePoolResources.Cpu.Deserved != nil && *nodePoolResources.Cpu.Deserved != -1) ||
		(nodePoolResources.Memory != nil && nodePoolResources.Memory.Deserved != nil && *nodePoolResources.Memory.Deserved != -1) {
		return nil, errors.New("CPU memory feature is off and cpu/memory deserved resources is assigned")
	}

	return deservedLimitPairs, nil
}

func ValidateResources(ctx context.Context, resources []api.ResourcesNullable, nodePoolsMetadata map[string]apiClusters.NodepoolMinimal) error {
	deservedLimitPairs := [][]*float64{}
	for i, nodePoolResources := range resources {
		npDeservedLimitPairs, err := validateNodePoolResources(nodePoolResources, nodePoolsMetadata)
		if err != nil {
			return fmt.Errorf("%s in index %d", err.Error(), i)
		}
		deservedLimitPairs = append(deservedLimitPairs, npDeservedLimitPairs...)
	}
	return ValidateSingleResource(ctx, deservedLimitPairs)
}

func getSchedulingCapabilitiesPerNodePool(nodePoolsMetadata map[string]apiClusters.NodepoolMinimal, resourcesNodePool *api.ResourcesNodePool) (*apiClusters.SchedulingCapabilities, string, error) {
	var nodePoolId string
	if resourcesNodePool == nil {
		// assume default node pool
		nodePoolId = getDefaultNodePool(nodePoolsMetadata)
		if nodePoolId == "" {
			logging.Errorf("resources are not valid. no node pool for resources")
			return nil, "", fmt.Errorf("resources are not valid. no node pool for resources")
		}
	} else {
		nodePoolId = resourcesNodePool.Id
	}

	sc, err := GetSchedulingCapabilitiesPerNodePoolId(nodePoolsMetadata, nodePoolId)
	return sc, nodePoolId, err
}

func GetSchedulingCapabilitiesPerNodePoolId(nodePoolsMetadata map[string]apiClusters.NodepoolMinimal, nodePoolId string) (*apiClusters.SchedulingCapabilities, error) {
	nodePool, ok := nodePoolsMetadata[nodePoolId]
	if !ok {
		return defaultSchedulingCapabilities(), fmt.Errorf("node pool with id %s not found", nodePoolId)
	}

	// currently in master, the "nodePool.schedulingCapabilities" field is a pointer - it is generated as such since it is marked as "x-internal" and is not "required".
	// once we fully enable the whole feature, we will remove the "x-internal" and add "required", then it will not be a pointer and will always have a value.
	// so for now, keep this logic and use default (false, false) if null;
	// although, it shouldn't happen, the fields always return in GET anyway.
	if nodePool.SchedulingCapabilities == nil {
		return defaultSchedulingCapabilities(), nil
	}
	return nodePool.SchedulingCapabilities, nil
}

func defaultSchedulingCapabilities() *apiClusters.SchedulingCapabilities {
	return &apiClusters.SchedulingCapabilities{
		AdjustableWeights: false,
		CpuQuota:          false,
		Quota:             true,
		Ranks:             true,
	}
}

func ValidateParent(ctx context.Context, parentId *string, tenantId string, clusterId string) (*models.Department, error) {
	parentUid, err := strconv.ParseInt(*parentId, 10, 64)
	if err != nil {
		return nil, err
	}
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return nil, err
	}
	parent, err := dao.GetDepartmentByIdAndCluster(ctx, parentUid, tenantId, clusterUuid, models.SlimVerbosity)
	if err != nil {
		return nil, err
	}
	return parent, nil
}

func ValidateJobsConfigurations(parent *models.Department, schedulingRules *api.SchedulingRules) error {
	if schedulingRules == nil {
		return nil
	}
	err := validateDuration(parent.OrgUnitData.InteractiveJobMaxIdleDuration, schedulingRules.InteractiveJobMaxIdleDurationSeconds, "InteractiveJobMaxIdleDuration")
	if err != nil {
		return err
	}
	err = validateDuration(parent.OrgUnitData.InteractiveJobPreemptIdleDuration, schedulingRules.InteractiveJobPreemptIdleDurationSeconds, "InteractiveJobPreemptIdleDuration")
	if err != nil {
		return err
	}
	err = validateDuration(parent.OrgUnitData.InteractiveJobTimeLimit, schedulingRules.InteractiveJobTimeLimitSeconds, "InteractiveJobTimeLimit")
	if err != nil {
		return err
	}
	err = validateDuration(parent.OrgUnitData.TrainingJobMaxIdleDuration, schedulingRules.TrainingJobMaxIdleDurationSeconds, "TrainingJobMaxIdleDuration")
	if err != nil {
		return err
	}
	err = validateDuration(parent.OrgUnitData.TrainingJobTimeLimit, schedulingRules.TrainingJobTimeLimitSeconds, "TrainingJobTimeLimit")
	if err != nil {
		return err
	}
	return nil
}

func validateDuration(parentDuration *int32, childDuration *int32, durationName string) error {
	if parentDuration != nil && childDuration != nil {
		if *parentDuration < *childDuration {
			return fmt.Errorf("%s is %d and cannot be larger than parents duration %d", durationName, *childDuration, *parentDuration)
		}
	}
	return nil
}

func ParseTenantId(ctx context.Context) (*string, error) {
	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)

	if tenantId == nil {
		return nil, errors.New("missing tenant id")
	}
	return tenantId, nil
}

func ParseSubjectId(ctx context.Context) (*string, error) {
	subjectId := runaictx.GetString(ctx, runaictx.CfSubjectID)

	if subjectId == nil {
		return nil, errors.New("missing subject id")
	}
	return subjectId, nil
}

func GetNodePoolsMetadata(ctx context.Context, clusterClient cluster_service_client.IClusterExternalClient, clusterId string) ([]apiClusters.NodepoolMinimal, error) {
	// TODO: once nodepools will have filter, filter out Deleting status
	response, err := clusterClient.GetNodePoolsMinimalByCluster(ctx, clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot validate request. failed to retrieve node pools")
		return nil, err
	}
	return response.Nodepools, nil
}

func GetNodePoolMetadata(ctx context.Context, clusterClient cluster_service_client.IClusterExternalClient, nodepoolId string) (*apiClusters.NodepoolMinimal, error) {
	response, err := clusterClient.GetNodepoolMinimal(ctx, nodepoolId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot validate request. failed to retrieve node pool %s", nodepoolId)
		return nil, err
	}
	return response, nil
}

func GetNodePoolsMetadataMap(ctx context.Context, clusterClient cluster_service_client.IClusterExternalClient, clusterId string) (map[string]apiClusters.NodepoolMinimal, error) {
	response, err := clusterClient.GetNodePoolsMinimalByCluster(ctx, clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot validate request. failed to retrieve node pools")
		return nil, err
	}
	return nodePoolsMetadataToMapById(response.Nodepools), nil
}

func nodePoolsToMapByIdInner(result map[string]apiClusters.NodepoolMinimal, nodePools []apiClusters.NodepoolMinimal) {
	for _, nodePool := range nodePools {
		if nodePool.Id != "" {
			result[nodePool.Id] = nodePool
		}
	}
}

func nodePoolsMetadataToMapById(nodePools []apiClusters.NodepoolMinimal) map[string]apiClusters.NodepoolMinimal {
	result := make(map[string]apiClusters.NodepoolMinimal)
	nodePoolsMetadataToMapByIdInner(result, nodePools)
	return result
}

func nodePoolsMetadataToMapByIdInner(result map[string]apiClusters.NodepoolMinimal, nodePools []apiClusters.NodepoolMinimal) {
	for _, nodePool := range nodePools {
		if nodePool.Id != "" {
			result[nodePool.Id] = nodePool
		}
	}
}

func GetNodePoolsMultiCluster(ctx context.Context, clusterClient cluster_service_client.IClusterExternalClient, clusterIds datastructs.Set[string]) (map[string]apiClusters.NodepoolMinimal, error) {
	result := make(map[string]apiClusters.NodepoolMinimal)
	for clusterId := range clusterIds {
		response, err := clusterClient.GetNodePoolsMinimalByCluster(ctx, clusterId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot validate request. failed to retrieve node pools for cluster %s", clusterId)
			return nil, err
		}
		nodePoolsToMapByIdInner(result, response.Nodepools)
	}
	return result, nil
}

func FetchClusterIdsOfProjects(projects []models.Project) datastructs.Set[string] {
	clusterIds := datastructs.Set[string]{}
	for _, project := range projects {
		clusterIds.Add(project.ClusterId.String())
	}
	return clusterIds
}

func FetchClusterIdsOfProjectNodePoolResources(nodePoolResources []models.ProjectNodePoolResources) datastructs.Set[string] {
	clusterIds := datastructs.Set[string]{}
	for _, resources := range nodePoolResources {
		clusterIds.Add(resources.ClusterId.String())
	}
	return clusterIds
}

func FetchClusterIdsOfDepartmentNodePoolResources(nodePoolResources []models.DepartmentNodePoolResources) datastructs.Set[string] {
	clusterIds := datastructs.Set[string]{}
	for _, resources := range nodePoolResources {
		clusterIds.Add(resources.ClusterId.String())
	}
	return clusterIds
}

func FetchClusterIdsOfDepartments(departments []models.Department) datastructs.Set[string] {
	clusterIds := datastructs.Set[string]{}
	for _, department := range departments {
		clusterIds.Add(department.ClusterId.String())
	}
	return clusterIds
}

func IsVersionSupportedInCluster(ctx context.Context, clusterId uuid.UUID, clusterClient cluster_service_client.IClusterExternalClient, version string) (bool, error) {
	clusterVersion, err := GetClusterVersion(ctx, clusterId, clusterClient)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get cluster version. error: %v", err.Error())
		return false, err
	}
	clusterSupport, err := versions.IsVersionSupported(clusterVersion, version)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cluster version %s not supported. error: %v", clusterVersion, err.Error())
		return false, err
	}
	return clusterSupport, nil
}

// IsLimitQuotaOverSubscriptionEnabled reads the setting from the local settings package, which owns it
// after the migration to org-unit-service.
func IsLimitQuotaOverSubscriptionEnabled(ctx context.Context) (bool, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return false, err
	}
	setting, err := setting_pkg.Get[*settings.LimitQuotaOverSubscription](*tenantId)
	if err != nil {
		return false, err
	}
	return setting.Enabled, nil
}

func ValidateSingleResource(ctx context.Context, deservedLimitPairs [][]*float64) error {
	for _, pair := range deservedLimitPairs {
		deserved := pair[0]
		limit := pair[1]
		if limit != nil && *limit != -1 && (deserved == nil || *deserved > *limit) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources invalid: deserved is more than the limit")
			return errors.New("resources invalid: deserved is more than the limit")
		}
	}
	return nil
}

func CopyResources(resources []models.Resources) []models.Resources {
	newResources := make([]models.Resources, len(resources))
	copy(newResources, resources)
	return newResources
}

func CopyResourcesDeservedAndStatus(resources []models.Resources) []models.Resources {
	newResources := make([]models.Resources, len(resources))
	for i, resource := range resources {
		newResources[i] = models.Resources{
			NodePoolId:      resource.NodePoolId,
			NodePoolName:    resource.NodePoolName,
			GpuDeserved:     CopyPtr(resource.GpuDeserved),
			CpuDeserved:     CopyPtr(resource.CpuDeserved),
			MemoryDeserved:  CopyPtr(resource.MemoryDeserved),
			MemoryUnits:     resource.MemoryUnits,
			ResourcesStatus: resource.ResourcesStatus,
		}
	}
	return newResources
}

func ValidateOrgName(name string, resourceType string) error {
	regex, err := regexp.Compile(OrgNameRegex)
	if err != nil {
		return err
	}

	if !regex.MatchString(name) {
		return fmt.Errorf("%s name should fit regular expression %s", resourceType, OrgNameRegex)
	}

	// in golang, the regexp cannot contain "lookarounds" so we have to check this separately.
	// validate the name contains at least one character a-z
	for _, c := range name {
		if c >= 'a' && c <= 'z' {
			return nil
		}
	}
	return fmt.Errorf("%s name should contain at least one character a-z", resourceType)
}

func ValidateInheritanceFields(project models.Project, department models.Department) []string {
	failedFields := []string{}
	childData := project.OrgUnitData
	parentData := department.OrgUnitData
	if !timeLimitFieldValid(parentData.TrainingJobTimeLimit, childData.TrainingJobTimeLimit) {
		failedFields = append(failedFields, "TrainingJobTimeLimit")
	}
	if !timeLimitFieldValid(parentData.InteractiveJobTimeLimit, childData.InteractiveJobTimeLimit) {
		failedFields = append(failedFields, "InteractiveJobTimeLimit")
	}
	if !timeLimitFieldValid(parentData.TrainingJobMaxIdleDuration, childData.TrainingJobMaxIdleDuration) {
		failedFields = append(failedFields, "TrainingJobMaxIdleDuration")
	}
	if !timeLimitFieldValid(parentData.InteractiveJobMaxIdleDuration, childData.InteractiveJobMaxIdleDuration) {
		failedFields = append(failedFields, "InteractiveJobMaxIdleDuration")
	}
	if !timeLimitFieldValid(parentData.InteractiveJobPreemptIdleDuration, childData.InteractiveJobPreemptIdleDuration) {
		failedFields = append(failedFields, "InteractiveJobPreemptIdleDuration")
	}
	if !nodeTypeFieldValid(parentData.TrainingNodeTypes, childData.TrainingNodeTypes) {
		failedFields = append(failedFields, "TrainingNodeTypes")
	}
	if !nodeTypeFieldValid(parentData.WorkspaceNodeTypes, childData.WorkspaceNodeTypes) {
		failedFields = append(failedFields, "WorkspaceNodeTypes")
	}

	return failedFields
}

func timeLimitFieldValid(parentLimit *int32, childLimit *int32) bool {
	return parentLimit == nil || childLimit == nil || *parentLimit >= *childLimit
}

func nodeTypeFieldValid(parentNodeTypes *[]models.NodeType, childNodeTypes *[]models.NodeType) bool {
	if parentNodeTypes == nil || childNodeTypes == nil {
		return true
	}
	parentNodeTypesMap := make(map[int64]bool)
	for _, nodeType := range *parentNodeTypes {
		parentNodeTypesMap[nodeType.ID] = true
	}
	for _, childNodeType := range *childNodeTypes {
		_, found := parentNodeTypesMap[childNodeType.ID]
		if !found {
			return false
		}
	}
	return true
}

func GetOrgUnitMetrics(ctx context.Context,
	clusterId string, projectName, departmentName, nodePoolName *string,
	metricTypes []api.OrgUnitMetricType, start time.Time, end time.Time, numberOfSamples int32,
	groupBy string,
) (api.ImplResponse, error) {
	if groupBy != "" && groupBy != "Category" {
		return HandleError(fmt.Errorf("invalid groupBy value: %s, allowed values: [Category]", groupBy), http.StatusBadRequest, "invalid groupBy parameter"), nil
	}

	if groupBy != "" {
		for _, mt := range metricTypes {
			if mt != api.ORGUNITMETRICTYPE_WORKLOADS_COUNT {
				return HandleError(fmt.Errorf("groupBy is only supported for WORKLOADS_COUNT metric type"), http.StatusBadRequest, "invalid groupBy parameter"), nil
			}
		}
	}

	measurements := []api.MeasurementResponse{}
	responseCode := http.StatusOK

	for _, metricType := range metricTypes {
		result := QueryProjectMetric(ctx, clusterId, projectName, departmentName, nodePoolName, metricType, start, end, numberOfSamples, groupBy)
		if result == nil {
			responseCode = http.StatusMultiStatus
			measurements = append(measurements, api.MeasurementResponse{
				Type: string(metricType),
			})

			continue
		}

		if groupBy != "" {
			groupedMeasurements := ParseGroupedMetricResultToMeasurements(ctx, *result, groupBy)
			measurements = append(measurements, groupedMeasurements...)
		} else {
			measurement := ParseOrgUnitMetricResultToMeasurement(ctx, string(metricType), *result)
			measurements = append(measurements, measurement)
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Finished querying metrics with responseCode: %d, for clusterId: %s, projectName: %v, departmentName: %v, nodePoolName: %v, start: %v, end: %v, numberOfSamples: %d",
		responseCode, clusterId, projectName, departmentName, nodePoolName, start, end, numberOfSamples)

	responseStruct := api.MetricsResponse{Measurements: measurements}
	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		fileName := "project-metrics.csv"
		if projectName == nil {
			fileName = "department-metrics.csv"
		}

		data := GetMetricsCsv(&responseStruct, metricTypes)
		response := runaihttp.CsvFile{Data: data, FileName: fileName}
		return api.Response(responseCode, response), nil
	}

	return api.Response(responseCode, responseStruct), nil
}

func GetProjectsPermissions(ctx context.Context, authClient auth_client.IAuthorizationClient, clusterId string, projects []*models.Project) (map[string]*[]openapi.AccessRule, error) {
	accessRulesPerId := make(map[string]*[]openapi.AccessRule)
	departmentToProjects := make(map[string]map[string]bool)
	departmentIdsMap := make(map[string]bool)
	projectIds := []string{}
	for _, project := range projects {
		projId := strconv.FormatInt(project.ID, 10)
		depId := strconv.FormatInt(*project.ParentID, 10)
		projectIds = append(projectIds, projId)
		departmentIdsMap[depId] = true

		if _, ok := departmentToProjects[depId]; !ok {
			departmentToProjects[depId] = map[string]bool{}
		}
		departmentToProjects[depId][projId] = true
	}
	departmentIds := []string{}
	for depId := range departmentIdsMap {
		departmentIds = append(departmentIds, depId)
	}

	accessRules, err := GetPermissions(ctx, authClient, strings.Join(projectIds, ","), strings.Join(departmentIds, ","), clusterId)
	if err != nil {
		return accessRulesPerId, err
	}
	for _, accessRule := range *accessRules {
		scopeType := accessRule.ScopeType
		switch scopeType {
		case openapi.SCOPETYPE_PROJECT:
			rulePerScope, ok := accessRulesPerId[accessRule.ScopeId]
			if !ok || rulePerScope == nil {
				rulePerScope = &[]openapi.AccessRule{}
				accessRulesPerId[accessRule.ScopeId] = rulePerScope
			}
			*rulePerScope = append(*rulePerScope, accessRule)
		case openapi.SCOPETYPE_DEPARTMENT:
			ruleProjectIds := departmentToProjects[accessRule.ScopeId]
			for ruleProjectId := range ruleProjectIds {
				rulePerScope, ok := accessRulesPerId[ruleProjectId]
				if !ok || rulePerScope == nil {
					rulePerScope = &[]openapi.AccessRule{}
					accessRulesPerId[ruleProjectId] = rulePerScope
				}
				*rulePerScope = append(*rulePerScope, accessRule)
			}
		default:
			for _, project := range projects {
				projectId := strconv.FormatInt(project.ID, 10)
				rulePerScope, ok := accessRulesPerId[projectId]
				if !ok || rulePerScope == nil {
					rulePerScope = &[]openapi.AccessRule{}
					accessRulesPerId[projectId] = rulePerScope
				}
				*rulePerScope = append(*rulePerScope, accessRule)
			}
		}
	}
	return accessRulesPerId, nil
}

func GetPermissions(ctx context.Context, authClient auth_client.IAuthorizationClient, projectIds string, departmentIds string, clusterId string) (*[]openapi.AccessRule, error) {
	getAccessRules200Response, response, err := authClient.GetScopeAccessRules(ctx).
		ProjectIds(projectIds).
		DepartmentIds(departmentIds).
		ClusterId(clusterId).
		Roles(strings.Join([]string{ResearchManagerRole, ResearcherRole, MlEngineerRole, L2ResearcherRole, EditorRole, DepartmentAdminRole, SystemAdminRole, ApplicationAdminRole}, ",")).Execute()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get project access rules. status is %s", response.Status)
		return nil, err
	}
	accessRules, exists := getAccessRules200Response.GetAccessRulesOk()
	if !exists {
		return &[]openapi.AccessRule{}, nil
	}
	if accessRules == nil {
		return &[]openapi.AccessRule{}, nil
	}
	return &accessRules, nil
}

func validateNodepoolsClusterMatching(ctx context.Context, clusterId string, nodePoolIds []string, clusterClient cluster_service_client.IClusterExternalClient) error {
	response, err := clusterClient.GetNodePoolsMinimalByCluster(ctx, clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot validate request. failed to retrieve node pools metadata for cluster %s: %s",
			clusterId, err.Error())
		return err
	}

	nodePoolIdsSet := datastructs.Set[string]{}
	for _, npMetadata := range response.Nodepools {
		nodePoolIdsSet.Add(npMetadata.Id)
	}

	for _, nodePoolId := range nodePoolIds {
		if !nodePoolIdsSet.Has(nodePoolId) {
			return fmt.Errorf("nodepool %s and cluster %s do not match", nodePoolId, clusterId)
		}
	}

	return nil
}

func ValidateResetResourcesRequest(ctx context.Context, tenantId, clusterId, parentId string, nodePoolIds []string, clusterClient cluster_service_client.IClusterExternalClient) error {
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return fmt.Errorf("failed to parse cluster id %s: %s", clusterId, err.Error())
	}

	err = validateNodepoolsClusterMatching(ctx, clusterId, nodePoolIds, clusterClient)
	if err != nil {
		return err
	}

	if parentId == "" {
		return nil
	}

	// validate parent department exists in cluster
	var departmentId int64
	departmentId, err = strconv.ParseInt(parentId, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse parent id %s: %s", parentId, err.Error())
	}

	_, err = dao.GetDepartmentByIdAndCluster(ctx, departmentId, tenantId, clusterUuid, models.MetaDataVerbosity)
	if err != nil {
		if errors.Is(err, orm.ErrRecordNotFound) {
			return fmt.Errorf("department %s and cluster %s do not match", parentId, clusterId)
		}
		return err
	}

	return nil
}

// GetNodePoolsForClusterFromServiceOrCache tries to load node pools for a cluster from cache; if missing, fetches and caches them.
func GetNodePoolsForClusterFromServiceOrCache(
	ctx context.Context,
	clusterClient cluster_service_client.IClusterExternalClient,
	clusterId string,
	nodePoolsPerClusterCache *sync.Map,
) (map[string]apiClusters.NodepoolMinimal, *api.Error) {
	if nodePoolsPerClusterCache != nil {
		if cached, ok := nodePoolsPerClusterCache.Load(clusterId); ok {
			return cached.(map[string]apiClusters.NodepoolMinimal), nil
		}
	}

	fetched, err := GetNodePoolsMetadataMap(ctx, clusterClient, clusterId)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to retrieve node pools", err)
	}

	if nodePoolsPerClusterCache != nil {
		nodePoolsPerClusterCache.Store(clusterId, fetched)
	}

	return fetched, nil
}

func ValidateBatchPatchNodePoolResourcesRequest(request api.BatchPatchNodePoolResourcesRequest) *api.Error {
	totalRequested := len(request.PatchRequests)
	if totalRequested == 0 {
		return api.NewError(http.StatusBadRequest, "patchRequests length cannot be 0", errors.New("invalid request"))
	}
	if totalRequested > MaxBatchPatchNodePoolResourcesRequests {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("patchRequests length exceeds %d", MaxBatchPatchNodePoolResourcesRequests), errors.New("invalid request"))
	}

	seenOrgUnitNodePoolIds := datastructs.Set[string]{}
	for _, pr := range request.PatchRequests {
		orgUnitNodePoolIds := fmt.Sprintf("%s//%s", pr.OrgUnitId, pr.NodePoolId)
		if seenOrgUnitNodePoolIds.Has(orgUnitNodePoolIds) {
			return api.NewError(http.StatusBadRequest, fmt.Sprintf("duplicate OrgUnitId//NodePoolId in patch requests: %s", orgUnitNodePoolIds), nil)
		}
		seenOrgUnitNodePoolIds.Add(orgUnitNodePoolIds)
	}

	return nil
}

func ValidateBatchDeleteReportsRequest(request api.BatchDeleteReportsRequest) *api.Error {
	if len(request.Ids) == 0 {
		return api.NewError(http.StatusBadRequest, "ids length cannot be 0", errors.New("invalid request"))
	}
	if len(request.Ids) > MaxBatchDeleteReportsRequests {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("ids length exceeds %d", MaxBatchDeleteReportsRequests), errors.New("invalid request"))
	}

	seenReportIds := datastructs.Set[string]{}
	for _, id := range request.Ids {
		if seenReportIds.Has(id) {
			return api.NewError(http.StatusBadRequest, fmt.Sprintf("duplicate report id in request: %s", id), nil)
		}
		seenReportIds.Add(id)
	}

	return nil
}

// ValidateSortByAllowedForGetNodePoolResources
// only allow sorting by gpuOverQuotaWeight if filtering by a specific nodepool,
// and this nodepool has adjustableWeights enabled -
// which will be validated later in the flow if needed.
// in this case, returns true (shouldValidateNodePoolAdjustableWeights), nil.
func ValidateSortByAllowedForGetNodePoolResources(sortBy string, filterBy []queryapi.Filter) (bool, *api.Error) {
	if sortBy != string(api.PROJECTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_OVER_QUOTA_WEIGHT) {
		return false, nil
	}

	for _, filter := range filterBy {
		if filter.Operator == "=" &&
			len(filter.Values) == 1 &&
			(filter.Field == models.ResourcesNodePoolIdColumn ||
				filter.Field == models.ResourcesNodePoolNameColumn) {
			return true, nil
		}
	}

	return false, api.NewError(http.StatusBadRequest, "sorting by gpuOverQuotaWeight is only allowed when filtering by a specific node pool", nil)
}

func EnrichUpdateMessage(message string, orgUnitType string, err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == runaihttp.ErrDuplicateKey && pgErr.ConstraintName == NodePoolToOrgUnitConstraint {
		return fmt.Sprintf("%s: %s is not up to date, refresh and try again", message, orgUnitType)
	}
	return message
}
