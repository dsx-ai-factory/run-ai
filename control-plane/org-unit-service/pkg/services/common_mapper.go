/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

	telemetry_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/telemetry"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/lib/pq"
	"k8s.io/utils/ptr"
)

const (
	defaultNodePoolName        = "default"
	ResourcePriorityMinVersion = "2.20"
	// MinOverQuotaWeight is the floor applied to the API response value of
	// OverQuotaWeight when adjustable weights are disabled and deserved == 0,
	// so the cluster scheduler can still fairshare/reclaim for the queue.
	MinOverQuotaWeight = 0.01
)

type QuotaStatus struct {
	QuotaStatuses []api.QuotaStatusV1Inner `json:"quotaStatuses"`
}

func MapSchedulingRules(orgUnitData models.OrgUnitData) api.SchedulingRules {
	schedulingRules := api.SchedulingRules{}
	if orgUnitData.InteractiveJobTimeLimit != nil {
		schedulingRules.InteractiveJobTimeLimitSeconds = orgUnitData.InteractiveJobTimeLimit
	}
	if orgUnitData.InteractiveJobMaxIdleDuration != nil {
		schedulingRules.InteractiveJobMaxIdleDurationSeconds = orgUnitData.InteractiveJobMaxIdleDuration
	}
	if orgUnitData.InteractiveJobPreemptIdleDuration != nil {
		schedulingRules.InteractiveJobPreemptIdleDurationSeconds = orgUnitData.InteractiveJobPreemptIdleDuration
	}
	if orgUnitData.TrainingJobMaxIdleDuration != nil {
		schedulingRules.TrainingJobMaxIdleDurationSeconds = orgUnitData.TrainingJobMaxIdleDuration
	}
	if orgUnitData.TrainingJobTimeLimit != nil {
		schedulingRules.TrainingJobTimeLimitSeconds = orgUnitData.TrainingJobTimeLimit
	}
	return schedulingRules
}

func MapSchedulingRulesEffective(orgUnitData models.OrgUnitData) api.SchedulingRules {
	schedulingRules := api.SchedulingRules{}
	if orgUnitData.InteractiveJobTimeLimitEffective != nil {
		schedulingRules.InteractiveJobTimeLimitSeconds = orgUnitData.InteractiveJobTimeLimitEffective
	}
	if orgUnitData.InteractiveJobMaxIdleDurationEffective != nil {
		schedulingRules.InteractiveJobMaxIdleDurationSeconds = orgUnitData.InteractiveJobMaxIdleDurationEffective
	}
	if orgUnitData.InteractiveJobPreemptIdleDurationEffective != nil {
		schedulingRules.InteractiveJobPreemptIdleDurationSeconds = orgUnitData.InteractiveJobPreemptIdleDurationEffective
	}
	if orgUnitData.TrainingJobMaxIdleDurationEffective != nil {
		schedulingRules.TrainingJobMaxIdleDurationSeconds = orgUnitData.TrainingJobMaxIdleDurationEffective
	}
	if orgUnitData.TrainingJobTimeLimitEffective != nil {
		schedulingRules.TrainingJobTimeLimitSeconds = orgUnitData.TrainingJobTimeLimitEffective
	}
	return schedulingRules
}

func MapRequestToOrgUnitData(
	ctx context.Context,
	schedulingRules *api.SchedulingRules,
	nodeTypes api.NodeTypesPerWorkload,
	defaultNodePools *[]string,
	resources []api.ResourcesNullable,
	isDepartment bool,
	nodePoolsMetadata map[string]openapi.NodepoolMinimal,
	parent *models.OrgUnitData,
	originalOrgUnitData *models.OrgUnitData,
) (models.OrgUnitData, error) {
	data := models.OrgUnitData{}
	var err error
	if originalOrgUnitData != nil {
		originalResourcesNodePoolToId := make(map[string]uuid.UUID)
		originalResourcesNodePoolIdToStatus := make(map[string]models.ResourcesStatus)
		for _, originalResources := range originalOrgUnitData.Resources {
			originalResourcesNodePoolToId[originalResources.NodePoolId] = originalResources.ID
			originalResourcesNodePoolIdToStatus[originalResources.NodePoolId] = originalResources.ResourcesStatus
		}
		data.ID = originalOrgUnitData.ID
		data.Resources, err = ApiResourcesToModel(ctx, resources, isDepartment, nodePoolsMetadata, true, originalResourcesNodePoolToId)
		if err != nil {
			return data, err
		}
		// Status columns are written by the cluster sync; carry them over so the save doesn't null them.
		for i := range data.Resources {
			data.Resources[i].ResourcesStatus = originalResourcesNodePoolIdToStatus[data.Resources[i].NodePoolId]
		}
	} else {
		data.Resources, err = ApiNewResourcesToModel(ctx, resources, isDepartment, nodePoolsMetadata)
		if err != nil {
			return data, err
		}
	}
	if schedulingRules != nil {
		data.InteractiveJobTimeLimit = schedulingRules.InteractiveJobTimeLimitSeconds
		data.InteractiveJobMaxIdleDuration = schedulingRules.InteractiveJobMaxIdleDurationSeconds
		data.InteractiveJobPreemptIdleDuration = schedulingRules.InteractiveJobPreemptIdleDurationSeconds
		data.TrainingJobMaxIdleDuration = schedulingRules.TrainingJobMaxIdleDurationSeconds
		data.TrainingJobTimeLimit = schedulingRules.TrainingJobTimeLimitSeconds
	}
	if defaultNodePools != nil && len(*defaultNodePools) > 0 {
		dbArray := pq.StringArray(*defaultNodePools)
		data.DefaultNodePoolList = &dbArray
	}
	if len(nodeTypes.Workspace) > 0 {
		nodeTypesList := []models.NodeType{}
		for _, nodeType := range nodeTypes.Workspace {
			nodeTypeId, err := strconv.ParseInt(nodeType, 10, 64)
			if err != nil {
				return data, err
			}
			nt := models.NodeType{IntBaseModel: models.IntBaseModel{ID: nodeTypeId}}
			if nodeTypes.Names != nil {
				nt.Name = (*nodeTypes.Names)[nodeType]
			}
			nodeTypesList = append(nodeTypesList, nt)
		}
		data.WorkspaceNodeTypes = ptr.To(nodeTypesList)
	}
	if len(nodeTypes.Training) > 0 {
		nodeTypesList := []models.NodeType{}
		for _, nodeType := range nodeTypes.Training {
			nodeTypeId, err := strconv.ParseInt(nodeType, 10, 64)
			if err != nil {
				return data, err
			}
			nt := models.NodeType{IntBaseModel: models.IntBaseModel{ID: nodeTypeId}}
			if nodeTypes.Names != nil {
				nt.Name = (*nodeTypes.Names)[nodeType]
			}
			nodeTypesList = append(nodeTypesList, nt)
		}
		data.TrainingNodeTypes = ptr.To(nodeTypesList)
	}
	parentId := uuid.Nil
	if parent != nil {
		parentId = parent.ID
	}
	data.EffectiveData = MapEffectiveFields(parentId, parent, data, isDepartment)
	return data, nil
}

func MapEffectiveFields(parentId uuid.UUID, parent *models.OrgUnitData, data models.OrgUnitData, isDepartment bool) models.EffectiveData {
	var effectiveData models.EffectiveData
	if parent == nil {
		effectiveData = models.EffectiveData{
			InteractiveJobTimeLimitEffective:           data.InteractiveJobTimeLimit,
			InteractiveJobMaxIdleDurationEffective:     data.TrainingJobMaxIdleDuration,
			InteractiveJobPreemptIdleDurationEffective: data.InteractiveJobPreemptIdleDuration,
			TrainingJobTimeLimitEffective:              data.TrainingJobTimeLimit,
			TrainingJobMaxIdleDurationEffective:        data.TrainingJobMaxIdleDuration,
			DefaultNodePoolListEffective:               data.DefaultNodePoolList,
			TrainingNodeTypesEffective:                 data.TrainingNodeTypes,
			WorkspaceNodeTypesEffective:                data.WorkspaceNodeTypesEffective,
		}
	} else {
		effectiveData.InteractiveJobTimeLimitEffective, effectiveData.InteractiveJobTimeLimitEffectiveOrigin = mapEffectiveTimeLimits(data.InteractiveJobTimeLimit, parent.InteractiveJobTimeLimit, parentId)
		effectiveData.InteractiveJobMaxIdleDurationEffective, effectiveData.InteractiveJobMaxIdleDurationEffectiveOrigin =
			mapEffectiveTimeLimits(data.InteractiveJobMaxIdleDuration, parent.InteractiveJobMaxIdleDuration, parentId)
		effectiveData.InteractiveJobPreemptIdleDurationEffective, effectiveData.InteractiveJobPreemptIdleDurationEffectiveOrigin =
			mapEffectiveTimeLimits(data.InteractiveJobPreemptIdleDuration, parent.InteractiveJobPreemptIdleDuration, parentId)
		effectiveData.TrainingJobTimeLimitEffective, effectiveData.TrainingJobTimeLimitEffectiveOrigin = mapEffectiveTimeLimits(data.TrainingJobTimeLimit, parent.TrainingJobTimeLimit, parentId)
		effectiveData.TrainingJobMaxIdleDurationEffective, effectiveData.TrainingJobMaxIdleDurationEffectiveOrigin = mapEffectiveTimeLimits(data.TrainingJobMaxIdleDuration, parent.TrainingJobMaxIdleDuration, parentId)
		effectiveData.TrainingNodeTypesEffective, effectiveData.TrainingNodeTypesEffectiveOrigin = mapEffectiveNodeTypes(data.TrainingNodeTypes, parent.TrainingNodeTypes, parentId)
		effectiveData.WorkspaceNodeTypesEffective, effectiveData.WorkspaceNodeTypesEffectiveOrigin = mapEffectiveNodeTypes(data.WorkspaceNodeTypes, parent.WorkspaceNodeTypes, parentId)
		if data.DefaultNodePoolList != nil {
			effectiveData.DefaultNodePoolListEffective = data.DefaultNodePoolList
		} else if parent.DefaultNodePoolList != nil {
			effectiveData.DefaultNodePoolListEffective = parent.DefaultNodePoolList
			effectiveData.DefaultNodePoolListEffectiveOrigin = &parentId
		}
	}
	if (effectiveData.DefaultNodePoolListEffective == nil || len(*effectiveData.DefaultNodePoolListEffective) == 0) && !isDepartment {
		dbArray := pq.StringArray{defaultNodePoolName}
		effectiveData.DefaultNodePoolListEffective = &dbArray
	}
	return effectiveData
}

func mapEffectiveTimeLimits(childField *int32, parentField *int32, parentId uuid.UUID) (*int32, *uuid.UUID) {
	if childField != nil {
		return childField, nil
	} else if parentField != nil {
		return parentField, &parentId
	}
	return nil, nil
}

func mapEffectiveNodeTypes(childField *[]models.NodeType, parentField *[]models.NodeType, parentId uuid.UUID) (*[]models.NodeType, *uuid.UUID) {
	if childField != nil {
		return childField, nil
	} else if parentField != nil {
		return parentField, &parentId
	}
	return nil, nil
}

func getDefaultNodePool(nodePoolsMetadata map[string]openapi.NodepoolMinimal) string {
	for _, nodePool := range nodePoolsMetadata {
		if nodePool.Name == defaultNodePoolName {
			return nodePool.Id
		}
	}
	return ""
}

func ParseRank(rank *string) (*string, error) {
	if rank != nil {
		if isKnownRank(*rank) {
			return rank, nil
		}
		return nil, fmt.Errorf("unknown rank %s", *rank)
	}

	return ptr.To(DefaultResourceRank), nil
}

func isKnownRank(rank string) bool {
	_, exists := RankValuesMap[rank]
	return exists
}

func MapRankToNumber(rank *string) int32 {
	if rank == nil {
		return DefaultResourceRankValue
	}

	rankValue, exists := RankValuesMap[*rank]
	if !exists {
		return DefaultResourceRankValue
	}
	return rankValue
}

func ApiNewResourcesToModel(ctx context.Context, apiResources []api.ResourcesNullable, isDepartment bool, nodePoolsMetadata map[string]openapi.NodepoolMinimal) ([]models.Resources, error) {
	return ApiResourcesToModel(ctx, apiResources, isDepartment, nodePoolsMetadata, true, make(map[string]uuid.UUID))
}

func isResourceRankNilOrNormal(apiResources []api.ResourcesNullable) (nilOrNormal bool) {
	nilOrNormal = true
	for _, resource := range apiResources {
		if resource.Rank != nil && *resource.Rank != DefaultResourceRank {
			nilOrNormal = false
			break
		}
	}
	return nilOrNormal
}

func IsResourceRankSupported(ctx context.Context, apiResources []api.ResourcesNullable, clusterId uuid.UUID, clustersClient cluster_service_client.IClusterExternalClient) (bool, error) {
	if isResourceRankNilOrNormal(apiResources) {
		return true, nil
	}
	return IsVersionSupportedInCluster(ctx, clusterId, clustersClient, ResourcePriorityMinVersion)
}

func ApiResourcesToModel(
	ctx context.Context,
	apiResources []api.ResourcesNullable,
	isDepartment bool,
	nodePoolsMetadata map[string]openapi.NodepoolMinimal,
	addEmptyResources bool,
	originalResourcesNodePoolToId map[string]uuid.UUID,
) ([]models.Resources, error) {
	modelResources := make([]models.Resources, len(apiResources))
	highestGpuLimit := float64(-1)
	highestCpuLimit := float64(-1)
	highestMemoryLimit := float64(-1)
	nodePoolsInResources := make(map[string]*api.ResourcesNodePool)
	if len(apiResources) == 1 && apiResources[0].NodePool == nil {
		nodePoolId := getDefaultNodePool(nodePoolsMetadata)
		if nodePoolId == "" {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to retrieve default node pool id")
			return modelResources, errors.New("failed to retrieve default node pool id")
		}
		apiResources[0].NodePool = &api.ResourcesNodePool{
			Id:   nodePoolId,
			Name: defaultNodePoolName,
		}
	}
	for i, resources := range apiResources {
		if resources.NodePool.Name == "" {
			existing, ok := nodePoolsMetadata[resources.NodePool.Id]
			if ok && existing.Name != "" {
				resources.NodePool.Name = existing.Name
			}
		}

		npSchedulingCapabilities, _, err := getSchedulingCapabilitiesPerNodePool(nodePoolsMetadata, resources.NodePool)
		if err != nil {
			return modelResources, err
		}
		adjustableWeightsFlag := npSchedulingCapabilities.AdjustableWeights

		nodePoolId := resources.NodePool.Id
		nodePoolsInResources[nodePoolId] = resources.NodePool
		rank, err := ParseRank(resources.Rank)
		if err != nil {
			return nil, err
		}
		gpuDeserved := ptr.To(ptr.Deref(resources.Gpu.Deserved, 0))
		modelResources[i] = models.Resources{
			NodePoolId:         nodePoolId,
			NodePoolName:       resources.NodePool.Name,
			GpuDeserved:        gpuDeserved,
			GpuLimit:           resources.Gpu.Limit,
			GpuOverQuotaWeight: calcOverQuotaWeight(adjustableWeightsFlag, isDepartment, gpuDeserved, resources.Gpu.OverQuotaWeight),
			Rank:               rank,
		}
		resourcesId, found := originalResourcesNodePoolToId[nodePoolId]
		if found {
			modelResources[i].ID = resourcesId
		}
		if resources.Gpu.Limit != nil && *resources.Gpu.Limit > highestGpuLimit {
			highestGpuLimit = *resources.Gpu.Limit
		}
		if resources.Cpu != nil {
			modelResources[i].CpuDeserved = resources.Cpu.Deserved
			modelResources[i].CpuLimit = resources.Cpu.Limit
			modelResources[i].CpuOverQuotaWeight = calcOverQuotaWeight(adjustableWeightsFlag, isDepartment, resources.Cpu.Deserved, resources.Cpu.OverQuotaWeight)
			if resources.Cpu.Limit != nil && *resources.Cpu.Limit > highestCpuLimit {
				highestCpuLimit = *resources.Cpu.Limit
			}
		}
		if resources.Memory != nil {
			modelResources[i].MemoryDeserved = resources.Memory.Deserved
			modelResources[i].MemoryLimit = resources.Memory.Limit
			modelResources[i].MemoryOverQuotaWeight = calcOverQuotaWeight(adjustableWeightsFlag, isDepartment, resources.Memory.Deserved, resources.Memory.OverQuotaWeight)
			modelResources[i].MemoryUnits = apiUnitsToModel(resources.Memory.Units)
			if resources.Memory.Limit != nil && *resources.Memory.Limit > highestMemoryLimit {
				highestMemoryLimit = *resources.Memory.Limit
			}
		}
	}
	if addEmptyResources {
		emptyResources := getAllEmptyResources(nodePoolsMetadata, nodePoolsInResources, highestGpuLimit, highestCpuLimit, highestMemoryLimit, isDepartment)
		modelResources = append(modelResources, emptyResources...)
	}
	return modelResources, nil
}

func apiUnitsToModel(units api.Units) models.Units {
	switch units {
	case api.UNITS_MB:
		return models.MB
	case api.UNITS_GB:
		return models.GB
	case api.UNITS_MIB:
		return models.Mib
	}
	return models.Mib
}

func nonAdjustableOverQuotaWeight(deserved *float64) *float64 {
	if deserved == nil {
		return nil
	}
	if *deserved == 0 {
		return ptr.To(MinOverQuotaWeight)
	}
	return ptr.To(*deserved)
}

// IsValidNonAdjustableOverQuotaWeight reports whether weight is acceptable for a
// queue whose adjustable weights are off. The base rule is weight == deserved; we
// also accept weight == MinOverQuotaWeight when deserved is 0, so a GET → PUT
// round-trip of a queue with deserved=0 does not trip validation.
func IsValidNonAdjustableOverQuotaWeight(weight float64, deserved *float64) bool {
	if deserved == nil {
		return true
	}
	if weight == *deserved {
		return true
	}
	return *deserved == 0 && weight == MinOverQuotaWeight
}

func calcOverQuotaWeight(adjustableWeightsFlag bool, isDepartment bool, deserved *float64, overQuotaWeight *float64) *float64 {
	if !adjustableWeightsFlag {
		if isDepartment {
			return deserved
		} else {
			return ptr.To(models.DefaultOverQuotaWeight)
		}
	}
	return overQuotaWeight
}

func getAllEmptyResources(allClusterNodePools map[string]openapi.NodepoolMinimal,
	nodePoolsInResources map[string]*api.ResourcesNodePool,
	highestGpuLimit float64,
	highestCpuLimit float64,
	highestMemoryLimit float64,
	isDepartment bool,
) []models.Resources {
	emptyResources := []models.Resources{}
	for _, nodePool := range allClusterNodePools {
		if nodePool.Phase != nil && (*nodePool.Phase) != openapi.NODEPOOLPHASE_DELETING {
			_, found := nodePoolsInResources[nodePool.Id]
			if !found {
				maxAllowedGpu := ptr.To(highestGpuLimit)
				if highestGpuLimit == -1 {
					maxAllowedGpu = nil
				}
				maxAllowedCpu := ptr.To(highestCpuLimit)
				if highestCpuLimit == -1 {
					maxAllowedCpu = nil
				}
				maxAllowedMemory := ptr.To(highestMemoryLimit)
				if highestMemoryLimit == -1 {
					maxAllowedMemory = nil
				}
				emptyResource := createEmptyResources(nodePool, isDepartment, maxAllowedGpu, maxAllowedCpu, maxAllowedMemory)
				emptyResources = append(emptyResources, emptyResource)
			}
		}
	}
	return emptyResources
}

func createEmptyResources(nodePoolMetadata openapi.NodepoolMinimal, isDepartment bool, gpuLimit *float64, cpuLimit *float64, memoryLimit *float64) models.Resources {
	gpuDeserved := float64(0)
	unlimited := float64(-1)
	emptyResources := models.Resources{
		BaseModel: dbmodels.BaseModel{
			CreatedAt: time.Now(),
		},
		NodePoolId:            nodePoolMetadata.Id,
		NodePoolName:          nodePoolMetadata.Name,
		GpuDeserved:           ptr.To(gpuDeserved),
		GpuLimit:              gpuLimit,
		GpuOverQuotaWeight:    ptr.To(models.DefaultOverQuotaWeight),
		CpuDeserved:           ptr.To(unlimited),
		CpuLimit:              cpuLimit,
		CpuOverQuotaWeight:    ptr.To(models.DefaultOverQuotaWeight),
		MemoryDeserved:        ptr.To(unlimited),
		MemoryLimit:           memoryLimit,
		MemoryOverQuotaWeight: ptr.To(models.DefaultOverQuotaWeight),
		MemoryUnits:           models.Mib,
		OrgUnitDataID:         uuid.UUID{},
	}
	if isDepartment {
		emptyResources.GpuOverQuotaWeight = ptr.To(gpuDeserved)
		emptyResources.CpuOverQuotaWeight = ptr.To(unlimited)
		emptyResources.MemoryOverQuotaWeight = ptr.To(unlimited)
	}
	return emptyResources
}

func ApiResourcesFromModel(modelResources []models.Resources, nodePoolsMetadata map[string]openapi.NodepoolMinimal) []api.Resources {
	apiResources := make([]api.Resources, len(modelResources))

	for i, resources := range modelResources {
		npSchedulingCapabilities, _ := GetSchedulingCapabilitiesPerNodePoolId(nodePoolsMetadata, resources.NodePoolId)
		// error shouldn't happen, as node pool existence is already validated
		adjustableWeightsFlag := npSchedulingCapabilities.AdjustableWeights
		cpuQuotaFlag := npSchedulingCapabilities.CpuQuota
		apiResources[i] = ApiResourceFromModel(resources, adjustableWeightsFlag, cpuQuotaFlag, false)
	}
	return apiResources
}

func ApiResourcesFromModelSync(modelResources []models.Resources, nodePools map[string]openapi.NodepoolMinimal) []api.ResourcesSync {
	apiResources := make([]api.ResourcesSync, len(modelResources))

	for i, resources := range modelResources {
		npSchedulingCapabilities, _ := GetSchedulingCapabilitiesPerNodePoolId(nodePools, resources.NodePoolId)
		// error shouldn't happen, as node pool existence is already validated
		adjustableWeightsFlag := npSchedulingCapabilities.AdjustableWeights
		cpuQuotaFlag := npSchedulingCapabilities.CpuQuota

		apiResourceFromModel := ApiResourceFromModel(resources, adjustableWeightsFlag, cpuQuotaFlag, true)
		rankValue := MapRankToNumber(resources.Rank)
		apiResources[i] = api.ResourcesSync{
			NodePool: apiResourceFromModel.NodePool,
			Gpu: api.ResourcesSyncGpu{
				Deserved:        apiResourceFromModel.Gpu.Deserved,
				Limit:           apiResourceFromModel.Gpu.Limit,
				OverQuotaWeight: apiResourceFromModel.Gpu.OverQuotaWeight,
			},
			Cpu:      apiResourceFromModel.Cpu,
			Rank:     rankValue,
			Priority: rankValue, // cluster agents that predate rank still read priority
		}
		if apiResourceFromModel.Memory != nil {
			apiResources[i].Memory = &api.ResourcesSyncMemory{
				Deserved:        apiResourceFromModel.Memory.Deserved,
				Limit:           apiResourceFromModel.Memory.Limit,
				OverQuotaWeight: apiResourceFromModel.Memory.OverQuotaWeight,
			}
		}
	}
	return apiResources
}

func ApiResourceFromModel(resources models.Resources, adjustableWeightsFlag, cpuQuotaFlag bool, returnClusterUnits bool) api.Resources {
	apiResourcesNodePool := api.ResourcesNodePool{
		Id:   resources.NodePoolId,
		Name: resources.NodePoolName,
	}
	result := api.Resources{
		NodePool: &apiResourcesNodePool,
		Gpu: api.ResourcesGpu{
			Limit:           PtrOrNil(resources.GpuLimit),
			Deserved:        UnlimitedOrNil(resources.GpuDeserved),
			OverQuotaWeight: PtrOrNil(resources.GpuOverQuotaWeight),
		},
		Rank: RankOrDefault(resources.Rank),
	}
	if !adjustableWeightsFlag {
		result.Gpu.OverQuotaWeight = nonAdjustableOverQuotaWeight(resources.GpuDeserved)
	}
	if cpuQuotaFlag && (resources.CpuDeserved != nil || resources.CpuLimit != nil || resources.CpuOverQuotaWeight != nil) {
		result.Cpu = &api.ResourcesCpu{
			Limit:           PtrOrUnlimitedPtr(resources.CpuLimit),
			Deserved:        PtrOrUnlimitedPtr(resources.CpuDeserved),
			OverQuotaWeight: PtrOrNil(resources.CpuOverQuotaWeight),
		}
		if !adjustableWeightsFlag {
			result.Cpu.OverQuotaWeight = nonAdjustableOverQuotaWeight(resources.CpuDeserved)
		}
	}
	if cpuQuotaFlag && (resources.MemoryDeserved != nil || resources.MemoryLimit != nil || resources.MemoryOverQuotaWeight != nil) {
		limit := getMemoryResources(resources.MemoryLimit, resources.MemoryUnits, returnClusterUnits)
		deserved := getMemoryResources(resources.MemoryDeserved, resources.MemoryUnits, returnClusterUnits)

		result.Memory = &api.ResourcesMemory{
			Limit:           limit,
			Deserved:        deserved,
			OverQuotaWeight: PtrOrNil(resources.MemoryOverQuotaWeight),
			Units:           ModelUnitsToApi(resources.MemoryUnits),
		}
		if !adjustableWeightsFlag {
			result.Memory.OverQuotaWeight = nonAdjustableOverQuotaWeight(deserved)
		}
	}

	return result
}

func getMemoryResources(val *float64, units models.Units, convertToCluster bool) *float64 {
	if val == nil {
		return ptr.To(-1.0)
	}
	if !convertToCluster || *val == -1 {
		return ptr.To(*val)
	}
	switch units {
	case models.Mib:
		return ptr.To(MibToMb(*val))
	case models.GB:
		return ptr.To(GbToMb(*val))
	case models.MB:
		return ptr.To(*val)
	default:
		return ptr.To(MibToMb(*val))
	}
}

// EnrichNodeTypeNames populates the Name field on NodeType objects using the provided name map (ID -> Name).
func EnrichNodeTypeNames(orgUnitData *models.OrgUnitData, nameMap map[int64]string) {
	enrichSlice := func(nodeTypes *[]models.NodeType) {
		if nodeTypes == nil {
			return
		}
		for i := range *nodeTypes {
			if (*nodeTypes)[i].Name == "" {
				if name, ok := nameMap[(*nodeTypes)[i].ID]; ok {
					(*nodeTypes)[i].Name = name
				}
			}
		}
	}
	enrichSlice(orgUnitData.TrainingNodeTypes)
	enrichSlice(orgUnitData.WorkspaceNodeTypes)
	enrichSlice(orgUnitData.TrainingNodeTypesEffective)
	enrichSlice(orgUnitData.WorkspaceNodeTypesEffective)
}

// BuildNodeTypeNameMap builds a map of ID -> Name from NodeType slices.
func BuildNodeTypeNameMap(nodeTypes []models.NodeType) map[int64]string {
	nameMap := make(map[int64]string, len(nodeTypes))
	for _, nt := range nodeTypes {
		if nt.Name != "" {
			nameMap[nt.ID] = nt.Name
		}
	}
	return nameMap
}

func ApiNodeTypesFromModel(trainingNodeTypes *[]models.NodeType, workspaceNodeTypes *[]models.NodeType) api.NodeTypesPerWorkload {
	names := make(map[string]string)
	nodeTypesPerWorkload := api.NodeTypesPerWorkload{}
	if trainingNodeTypes != nil {
		nodeTypesPerWorkload.Training = []string{}
		for _, trainNodeType := range *trainingNodeTypes {
			id := strconv.FormatInt(trainNodeType.ID, 10)
			nodeTypesPerWorkload.Training = append(nodeTypesPerWorkload.Training, id)
			if trainNodeType.Name != "" {
				names[id] = trainNodeType.Name
			}
		}
	}
	if workspaceNodeTypes != nil {
		nodeTypesPerWorkload.Workspace = []string{}
		for _, workspaceNodeType := range *workspaceNodeTypes {
			id := strconv.FormatInt(workspaceNodeType.ID, 10)
			nodeTypesPerWorkload.Workspace = append(nodeTypesPerWorkload.Workspace, id)
			if workspaceNodeType.Name != "" {
				names[id] = workspaceNodeType.Name
			}
		}
	}
	if len(names) > 0 {
		nodeTypesPerWorkload.Names = ptr.To(names)
	}
	return nodeTypesPerWorkload
}

func MapQuotaStatusModelToApi(quotaStatus models.QuotaStatus) api.QuotaStatus {
	return api.QuotaStatus{
		Allocated: api.QuotaStatusResource{
			Gpu:    DereferenceFloat64(quotaStatus.GpuAllocated),
			Cpu:    DereferenceFloat64(quotaStatus.CpuAllocated),
			Memory: DereferenceFloat64(quotaStatus.MemoryAllocated),
		},
		AllocatedNonPreemptible: api.QuotaStatusResource{
			Gpu:    DereferenceFloat64(quotaStatus.GpuAllocatedNonPreemptible),
			Cpu:    DereferenceFloat64(quotaStatus.CpuAllocatedNonPreemptible),
			Memory: DereferenceFloat64(quotaStatus.MemoryAllocatedNonPreemptible),
		},
		Requested: api.QuotaStatusResource{
			Gpu:    DereferenceFloat64(quotaStatus.GpuRequested),
			Cpu:    DereferenceFloat64(quotaStatus.CpuRequested),
			Memory: DereferenceFloat64(quotaStatus.MemoryRequested),
		},
	}
}

func MapOvertimeDataModelToApi(overtimeData models.OvertimeData, orgUnitName, projOrDepStr string) api.OvertimeData {
	return api.OvertimeData{
		Range24hData: parseOvertimeRangeData(overtimeData.AvgGpuAllocation24h, overtimeData.AvgGpuUtilization24h, overtimeData.AvgGpuMemoryUtilization24h, overtimeData.Avg24hRangeUpdatedAt,
			"24h", orgUnitName, projOrDepStr),
		Range7dData: parseOvertimeRangeData(overtimeData.AvgGpuAllocation7d, overtimeData.AvgGpuUtilization7d, overtimeData.AvgGpuMemoryUtilization7d, overtimeData.Avg7dRangeUpdatedAt,
			"7d", orgUnitName, projOrDepStr),
		Range30dData: parseOvertimeRangeData(overtimeData.AvgGpuAllocation30d, overtimeData.AvgGpuUtilization30d, overtimeData.AvgGpuMemoryUtilization30d, overtimeData.Avg30dRangeUpdatedAt,
			"30d", orgUnitName, projOrDepStr),
	}
}

func parseOvertimeRangeData(averageGpuAllocation, averageGpuUtilization, averageGpuMemoryUtilization *float64,
	updatedAt *time.Time, overtimeDataRange, orgUnitName, projOrDepStr string,
) *api.OvertimeRangeData {
	if averageGpuAllocation == nil || averageGpuUtilization == nil || updatedAt == nil ||
		isNaN(averageGpuAllocation) || isNaN(averageGpuUtilization) {
		return nil
	}

	if isNaN(averageGpuMemoryUtilization) {
		logging.Warnf("For %s '%s', averageGpuMemoryUtilization for range %s is NaN, ignoring it",
			projOrDepStr, orgUnitName, overtimeDataRange)
		averageGpuMemoryUtilization = nil
	}

	return &api.OvertimeRangeData{
		AverageGpuAllocation:        *averageGpuAllocation,
		AverageGpuUtilization:       *averageGpuUtilization,
		AverageGpuMemoryUtilization: averageGpuMemoryUtilization,
		UpdatedAt:                   *updatedAt,
	}
}

func isNaN(value *float64) bool {
	// there's an edge case where golang's 'float64' can be NaN (Not a Number),
	// when it is somehow fetched from db for overtime data.
	// if this is the case, ignore it.
	return value != nil && *value != *value
}

func PtrOrNil[T any](value *T) *T {
	if value == nil {
		return nil
	}
	return ptr.To(*value)
}

func UnlimitedOrNil(value *float64) float64 {
	if value == nil {
		return -1
	}
	return *value
}

func PtrOrUnlimitedPtr(value *float64) *float64 {
	if value == nil {
		return ptr.To(-1.0)
	}
	return value
}

func RankOrDefault(rank *string) *string {
	if rank == nil {
		return ptr.To(DefaultResourceRank)
	}
	return rank
}

func TelemetryGroupsToApi(telemetryType string, timestamp time.Time, groups []telemetry_utils.TelemetryGroupValue, groupBy []string) *api.TelemetryResponse {
	telemetryGroups := &[]api.TelemetryResponseValuesInner{}
	for _, group := range groups {
		telemetryGroupApi := telemetryGroupToApi(group, groupBy)
		*telemetryGroups = append(*telemetryGroups, *telemetryGroupApi)
	}
	telemetryResponse := api.TelemetryResponse{
		Type:      telemetryType,
		Timestamp: timestamp,
		Values:    *telemetryGroups,
	}
	return &telemetryResponse
}

func telemetryGroupToApi(group telemetry_utils.TelemetryGroupValue, groupBy []string) *api.TelemetryResponseValuesInner {
	var apiGroups []api.TelemetryResponseValuesInnerGroupsInner
	for i, fieldName := range groupBy {
		if i == 0 {
			apiGroups = keyValueGroupToApi(fieldName, apiGroups, group.Group1MainValue, group.Group1SecondaryValue)
		}
		if i == 1 {
			apiGroups = keyValueGroupToApi(fieldName, apiGroups, group.Group2MainValue, group.Group2SecondaryValue)
		}
	}
	telemetryGroup := api.TelemetryResponseValuesInner{
		Value:  group.Data,
		Groups: apiGroups,
	}
	return &telemetryGroup
}

func keyValueGroupToApi(fieldName string, apiGroups []api.TelemetryResponseValuesInnerGroupsInner, groupValue string, groupName string) []api.TelemetryResponseValuesInnerGroupsInner {
	return append(apiGroups, api.TelemetryResponseValuesInnerGroupsInner{
		Key:   fieldName,
		Value: groupValue,
		Name:  groupName,
	})
}

func MapAllocationStatus(status float64) *float64 {
	if status == 0 {
		return nil
	}
	return ptr.To(status)
}

func ConvertNodePoolResourcesPatchRequest(npr api.NodePoolResourcesPatchRequest, storedGpuDeserved *float64) api.ResourcesNullable {
	result := api.ResourcesNullable{
		NodePool: &api.ResourcesNodePool{
			Id: npr.NodePoolId,
		},
		Rank: npr.Rank,
		Gpu:  api.ResourcesNullableGpu{Deserved: storedGpuDeserved},
	}

	if npr.Resources != nil {
		if npr.Resources.Gpu != nil {
			result.Gpu.Limit = npr.Resources.Gpu.Limit
			result.Gpu.OverQuotaWeight = npr.Resources.Gpu.OverQuotaWeight
			if npr.Resources.Gpu.Deserved != nil {
				result.Gpu.Deserved = npr.Resources.Gpu.Deserved
			}
		}
		if npr.Resources.Cpu != nil {
			result.Cpu = npr.Resources.Cpu
		}
		if npr.Resources.Memory != nil {
			result.Memory = npr.Resources.Memory
		}
	}

	return result
}

func ApiNodePoolResourcesPatchRequestToModelResourcesUpdates(
	patchRequest api.NodePoolResourcesPatchRequest,
	nodePoolsMetadata map[string]openapi.NodepoolMinimal,
	isDepartment bool,
) (map[string]interface{}, error) {
	npSchedulingCapabilities, err := GetSchedulingCapabilitiesPerNodePoolId(nodePoolsMetadata, patchRequest.NodePoolId)
	if err != nil {
		return nil, err
	}

	adjustableWeightsFlag := npSchedulingCapabilities.AdjustableWeights

	// only set the fields that are provided in the patch request
	updates := map[string]interface{}{}
	if patchRequest.Resources.Gpu != nil {
		if patchRequest.Resources.Gpu.Deserved != nil {
			updates["gpu_deserved"] = patchRequest.Resources.Gpu.Deserved
			if !adjustableWeightsFlag {
				updates["gpu_over_quota_weight"] = calcOverQuotaWeight(adjustableWeightsFlag, isDepartment, patchRequest.Resources.Gpu.Deserved, patchRequest.Resources.Gpu.Deserved)
			}
		}
		if adjustableWeightsFlag && patchRequest.Resources.Gpu.OverQuotaWeight != nil {
			updates["gpu_over_quota_weight"] = patchRequest.Resources.Gpu.OverQuotaWeight
		}
		if patchRequest.Resources.Gpu.Limit != nil {
			updates["gpu_limit"] = patchRequest.Resources.Gpu.Limit
		}
	}
	if patchRequest.Resources.Cpu != nil {
		if patchRequest.Resources.Cpu.Deserved != nil {
			updates["cpu_deserved"] = patchRequest.Resources.Cpu.Deserved
			if !adjustableWeightsFlag {
				updates["cpu_over_quota_weight"] = calcOverQuotaWeight(adjustableWeightsFlag, isDepartment, patchRequest.Resources.Cpu.Deserved, patchRequest.Resources.Cpu.Deserved)
			}
		}
		if adjustableWeightsFlag && patchRequest.Resources.Cpu.OverQuotaWeight != nil {
			updates["cpu_over_quota_weight"] = patchRequest.Resources.Cpu.OverQuotaWeight
		}
		if patchRequest.Resources.Cpu.Limit != nil {
			updates["cpu_limit"] = patchRequest.Resources.Cpu.Limit
		}
	}
	if patchRequest.Resources.Memory != nil {
		if patchRequest.Resources.Memory.Deserved != nil {
			updates["memory_deserved"] = patchRequest.Resources.Memory.Deserved
			if !adjustableWeightsFlag {
				updates["memory_over_quota_weight"] = calcOverQuotaWeight(adjustableWeightsFlag, isDepartment, patchRequest.Resources.Memory.Deserved, patchRequest.Resources.Memory.Deserved)
			}
		}
		if adjustableWeightsFlag && patchRequest.Resources.Memory.OverQuotaWeight != nil {
			updates["memory_over_quota_weight"] = patchRequest.Resources.Memory.OverQuotaWeight
		}
		if patchRequest.Resources.Memory.Limit != nil {
			updates["memory_limit"] = patchRequest.Resources.Memory.Limit
		}
		if patchRequest.Resources.Memory.Units != "" {
			updates["memory_units"] = apiUnitsToModel(patchRequest.Resources.Memory.Units)
		}
	}
	if patchRequest.Rank != nil {
		updates["priority"] = *patchRequest.Rank
	}

	return updates, nil
}

// MapNodePoolPatchResourcesRequestToModel is similar to mapPatchResourcesRequestToModel, but is used
// for node-pool level batch patch APIs where omitted CPU/Memory/Rank fields should NOT clear the
// existing values. It applies a partial patch over the existing OrgUnitData resources.
func MapNodePoolPatchResourcesRequestToModel(
	ctx context.Context,
	requestedResources []api.ResourcesNullable,
	orgUnitResources []models.Resources,
	nodePoolsMetadata map[string]openapi.NodepoolMinimal,
	patchRequests []api.NodePoolResourcesPatchRequest,
) ([]models.Resources, error) {
	originalResourcesNodePoolToId := make(map[string]uuid.UUID)
	for _, originalNPResources := range orgUnitResources {
		originalResourcesNodePoolToId[originalNPResources.NodePoolId] = originalNPResources.ID
	}

	// Map the incoming API resources to model resources (per-nodepool)
	newResources, err := ApiResourcesToModel(ctx, requestedResources, false, nodePoolsMetadata, false, originalResourcesNodePoolToId)
	if err != nil {
		return nil, err
	}

	err = PatchOrgUnitDataResourcesPartial(newResources, orgUnitResources, patchRequests)
	if err != nil {
		return nil, err
	}

	return orgUnitResources, nil
}
