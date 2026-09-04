/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"golang.org/x/exp/slices"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	cluster_openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	metrics_api "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"k8s.io/utils/ptr"
)

const (
	CreatedPhase    = "Creating"
	UpdatedPhase    = "Updating"
	DeletedNodePool = "Deleting"
)

func getProjectsDeservedGpus(projects []models.Project) string {
	if len(projects) == 0 {
		return ""
	}
	deservedGpus := float64(0)
	for _, project := range projects {
		for _, resource := range project.OrgUnitData.Resources {
			if resource.GpuDeserved != nil {
				deservedGpus = deservedGpus + *resource.GpuDeserved
			}
		}
	}
	return strconv.FormatFloat(deservedGpus, 'f', -1, 32)
}

func MapProjectModelToV1Response(ctx context.Context, project models.Project, accessRules *[]openapi.AccessRule, memoryUnitMb bool, oqFlag bool, cpuResourcesFlag bool) (api.ProjectV1, error) {
	apiProject := api.ProjectV1{}

	projectId := int32(project.ID)
	if int64(projectId) != project.ID {
		return api.ProjectV1{}, fmt.Errorf("failed to map project. project id %d is not int32", project.ID)
	}

	tenantId64, err := strconv.Atoi(project.TenantId)
	if err != nil {
		return api.ProjectV1{}, err
	}

	tenantId := int32(tenantId64)
	if int(tenantId) != tenantId64 {
		return api.ProjectV1{}, fmt.Errorf("failed to map project. id %d is not int32", tenantId64)
	}

	departmentId := *project.ParentID
	departmentIdInt32 := int32(departmentId)
	if int64(departmentIdInt32) != departmentId {
		return api.ProjectV1{}, fmt.Errorf("failed to map project. department id %d is not int32", departmentId)
	}

	orgUnitData := project.OrgUnitData

	nodePoolAssignedResources, assignedResources, err := mapProjectResourcesToResponse(ctx, orgUnitData.Resources, departmentIdInt32, memoryUnitMb, oqFlag, cpuResourcesFlag)
	if err != nil {
		return api.ProjectV1{}, err
	}
	totalGpus := assignedResources.Gpu.Deserved
	maxAllowedGpus := assignedResources.Gpu.MaxAllowed
	oqWeightGpus := assignedResources.Gpu.OverQuotaWeight

	apiProject = api.ProjectV1{
		NodePoolsResources: nodePoolAssignedResources,
		DeservedGpus:       convutils.Deref(totalGpus, 0),
		MaxAllowedGpus:     convutils.Deref(maxAllowedGpus, -1),
		GpuOverQuotaWeight: convutils.Deref(oqWeightGpus, 0),
		Resources:          *assignedResources,
		Name:               project.Name,
		Namespace:          project.RequestedNamespace,
		Id:                 projectId,
		DepartmentId:       departmentIdInt32,
		TenantId:           tenantId,
		ClusterUuid:        project.ClusterId.String(),
		DepartmentName:     convutils.Deref(project.ParentName, ""),
		CreatedAt:          project.CreatedAt,
	}
	if orgUnitData.DefaultNodePoolListEffective != nil {
		apiProject.DefaultNodePools = *orgUnitData.DefaultNodePoolListEffective
	}

	mapProjectStatusModelToApiV1(ctx, project, &apiProject)
	mapNodeAffinityToApi(orgUnitData, &apiProject)
	mapSchedulingRules(orgUnitData, &apiProject)
	if accessRules != nil {
		apiProject.Permissions = mapAccessRules(*accessRules)
	} else {
		apiProject.Permissions = api.ProjectV1ResponseCommonFieldsPermissions{
			Users:  &[]string{},
			Groups: &[]string{},
		}
	}
	return apiProject, nil
}

func MapProjectModelToV1Response2(ctx context.Context,
	project models.Project,
	accessRules *[]openapi.AccessRule,
	memoryUnitMb bool,
	nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal,
) (api.ProjectV1, error) {
	apiProject := api.ProjectV1{}

	projectId := int32(project.ID)
	if int64(projectId) != project.ID {
		return api.ProjectV1{}, fmt.Errorf("failed to map project. project id %d is not int32", project.ID)
	}

	tenantId64, err := strconv.Atoi(project.TenantId)
	if err != nil {
		return api.ProjectV1{}, err
	}

	tenantId := int32(tenantId64)
	if int(tenantId) != tenantId64 {
		return api.ProjectV1{}, fmt.Errorf("failed to map project. id %d is not int32", tenantId64)
	}

	departmentId := *project.ParentID
	departmentIdInt32 := int32(departmentId)
	if int64(departmentIdInt32) != departmentId {
		return api.ProjectV1{}, fmt.Errorf("failed to map project. department id %d is not int32", departmentId)
	}

	orgUnitData := project.OrgUnitData

	nodePoolAssignedResources, err := mapProjectResourcesToResponse2(orgUnitData.Resources, departmentIdInt32, memoryUnitMb, nodePoolsMetadata)
	if err != nil {
		return api.ProjectV1{}, err
	}

	apiProject = api.ProjectV1{
		NodePoolsResources: nodePoolAssignedResources,
		Name:               project.Name,
		Namespace:          project.RequestedNamespace,
		Id:                 projectId,
		DepartmentId:       departmentIdInt32,
		TenantId:           tenantId,
		ClusterUuid:        project.ClusterId.String(),
		DepartmentName:     convutils.Deref(project.ParentName, ""),
		CreatedAt:          project.CreatedAt,
	}
	if orgUnitData.DefaultNodePoolListEffective != nil {
		apiProject.DefaultNodePools = *orgUnitData.DefaultNodePoolListEffective
	}

	mapProjectStatusModelToApiV1(ctx, project, &apiProject)
	mapNodeAffinityToApi(orgUnitData, &apiProject)
	mapSchedulingRules(orgUnitData, &apiProject)
	if accessRules != nil {
		apiProject.Permissions = mapAccessRules(*accessRules)
	} else {
		apiProject.Permissions = api.ProjectV1ResponseCommonFieldsPermissions{
			Users:  &[]string{},
			Groups: &[]string{},
		}
	}
	return apiProject, nil
}

func MapProjectsModelToV1Response(
	ctx context.Context,
	projects []*models.Project,
	accessRules map[string]*[]openapi.AccessRule,
	memoryUnitMb bool,
	nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal,
	excludePermissions bool,
) ([]api.ProjectV1, error) {
	apiProjects := []api.ProjectV1{}
	for _, project := range projects {
		var rules *[]openapi.AccessRule
		if !excludePermissions {
			rules = accessRules[strconv.FormatInt(project.ID, 10)]
		}
		apiProject, err := MapProjectModelToV1Response2(ctx, *project, rules, memoryUnitMb, nodePoolsMetadata)
		if err != nil {
			return apiProjects, err
		}
		apiProjects = append(apiProjects, apiProject)
	}
	return apiProjects, nil
}

func mapNodeAffinityToApi(orgUnitData models.OrgUnitData, apiProject *api.ProjectV1) {
	apiProject.NodeAffinity = &api.ProjectV1NodeAffinityResponse{}
	mapTrainNodeAffinity(&orgUnitData, apiProject)
	mapInteractiveNodeAffinity(&orgUnitData, apiProject)
}

func mapInteractiveNodeAffinity(orgUnitData *models.OrgUnitData, apiProject *api.ProjectV1) {
	if orgUnitData.WorkspaceNodeTypesEffective == nil || len(*orgUnitData.WorkspaceNodeTypesEffective) == 0 {
		apiProject.NodeAffinity.Interactive = api.ProjectV1NodeAffinityResponseInteractive{
			AffinityType:  AffinityTypeNoLimit,
			SelectedTypes: []api.ProjectV1NodeAffinityResponseTrainSelectedTypesInner{},
		}
		return
	}
	selected := []api.ProjectV1NodeAffinityResponseTrainSelectedTypesInner{}
	for _, node := range *orgUnitData.WorkspaceNodeTypesEffective {
		selected = append(selected, api.ProjectV1NodeAffinityResponseTrainSelectedTypesInner{
			Id:   float32(node.ID),
			Name: node.Name,
		})
	}
	apiProject.NodeAffinity.Interactive = api.ProjectV1NodeAffinityResponseInteractive{
		AffinityType:  AffinityTypeOnlySelected,
		SelectedTypes: selected,
	}
}

func mapTrainNodeAffinity(orgUnitData *models.OrgUnitData, apiProject *api.ProjectV1) {
	if orgUnitData.TrainingNodeTypesEffective == nil || len(*orgUnitData.TrainingNodeTypesEffective) == 0 {
		apiProject.NodeAffinity.Train = api.ProjectV1NodeAffinityResponseTrain{
			AffinityType:  AffinityTypeNoLimit,
			SelectedTypes: []api.ProjectV1NodeAffinityResponseTrainSelectedTypesInner{},
		}
	} else {
		selected := []api.ProjectV1NodeAffinityResponseTrainSelectedTypesInner{}
		for _, node := range *orgUnitData.TrainingNodeTypesEffective {
			selected = append(selected, api.ProjectV1NodeAffinityResponseTrainSelectedTypesInner{
				Id:   float32(node.ID),
				Name: node.Name,
			})
		}
		apiProject.NodeAffinity.Train = api.ProjectV1NodeAffinityResponseTrain{
			AffinityType:  AffinityTypeOnlySelected,
			SelectedTypes: selected,
		}
	}
}

func mapAccessRules(rules []openapi.AccessRule) api.ProjectV1ResponseCommonFieldsPermissions {
	users := []string{}
	groups := []string{}
	for _, rule := range rules {
		if rule.SubjectType == openapi.SUBJECTTYPEEXTENDED_GROUP {
			groups = append(groups, rule.SubjectId)
		} else {
			users = append(users, rule.SubjectId)
		}
	}
	slices.Sort(users)
	slices.Sort(groups)

	return api.ProjectV1ResponseCommonFieldsPermissions{
		Users:        &users,
		Groups:       &groups,
		Applications: []string{},
	}
}

func mapProjectStatusModelToApiV1(ctx context.Context, project models.Project, apiProject *api.ProjectV1) {
	status := project.ProjectStatus
	if status.Phase != nil {
		apiProject.Phase = *status.Phase
		apiProject.Status.Phase = *status.Phase
	}
	if status.PhaseMessage != nil {
		apiProject.Status.Message = *status.PhaseMessage
	}
	if status.StatusNamespace != nil {
		apiProject.Status.Namespace = *status.StatusNamespace
	}
	if status.Conditions != nil {
		apiProject.Status.Conditions = conditionsModelToApiV1(status.Conditions)
	}
	nodePoolQuotaStatuses := []api.QuotaStatusV1Inner{}
	for _, resources := range project.OrgUnitData.Resources {
		if !hasStatus(resources) {
			continue
		}
		allocatedMemory := float32(services.MbToMib(ptr.Deref(resources.AllocatedMemory, 0)))
		allocatedNonPreemptibleMemory := float32(services.MbToMib(ptr.Deref(resources.AllocatedNonPreemptibleMemory, 0)))
		requestedMemory := float32(services.MbToMib(ptr.Deref(resources.RequestedMemory, 0)))
		nodePoolQuotaStatuses = append(nodePoolQuotaStatuses, api.QuotaStatusV1Inner{
			NodePoolName: resources.NodePoolName,
			Allocated: api.QuotaStatusResourceListV1{
				Gpu:    ptr.Deref(ptrToFloat32(resources.AllocatedGpu, nil), 0),
				Cpu:    ptr.Deref(ptrToFloat32(resources.AllocatedCpu, nil), 0),
				Memory: allocatedMemory,
			},
			AllocatedNonPreemptible: api.QuotaStatusResourceListV1{
				Gpu:    ptr.Deref(ptrToFloat32(resources.AllocatedNonPreemptibleGpu, nil), 0),
				Cpu:    ptr.Deref(ptrToFloat32(resources.AllocatedNonPreemptibleCpu, nil), 0),
				Memory: allocatedNonPreemptibleMemory,
			},
			Requested: api.QuotaStatusResourceListV1{
				Gpu:    ptr.Deref(ptrToFloat32(resources.RequestedGpu, nil), 0),
				Cpu:    ptr.Deref(ptrToFloat32(resources.RequestedCpu, nil), 0),
				Memory: requestedMemory,
			},
		})
	}
	apiProject.Status.QuotaStatuses = nodePoolQuotaStatuses
}

func hasStatus(resources models.Resources) bool {
	return resources.AllocatedGpu != nil || resources.AllocatedCpu != nil || resources.AllocatedMemory != nil ||
		resources.AllocatedNonPreemptibleGpu != nil || resources.AllocatedNonPreemptibleCpu != nil || resources.AllocatedNonPreemptibleMemory != nil ||
		resources.RequestedGpu != nil || resources.RequestedCpu != nil || resources.RequestedMemory != nil
}

func MapProjectStatusApiV1ToModel(status api.ProjectV1StatusRequest, project *models.Project) (models.ProjectStatus, map[string]models.ResourcesStatus, error) {
	result := models.ProjectStatus{}
	resourcesStatuses := make(map[string]models.ResourcesStatus)
	if status.Status != "" {
		rawStatus := RawStatus{}
		err := json.Unmarshal([]byte(status.Status), &rawStatus)
		if err != nil {
			return result, resourcesStatuses, err
		}

		result.Phase = ptr.To(rawStatus.Phase)
		result.PhaseMessage = ptr.To(rawStatus.Message)
		result.StatusNamespace = ptr.To(rawStatus.Namespace)
		result.Conditions = rawStatus.Conditions
	}

	for _, nodePoolQuotaStatus := range status.QuotaStatuses {
		resourcesStatuses[nodePoolQuotaStatus.NodePoolName] = models.ResourcesStatus{
			AllocatedGpu:                  services.MapAllocationStatus(float64(nodePoolQuotaStatus.Allocated.Gpu)),
			AllocatedCpu:                  services.MapAllocationStatus(float64(nodePoolQuotaStatus.Allocated.Cpu)),
			AllocatedMemory:               services.MapAllocationStatus(services.MibToMb(float64(nodePoolQuotaStatus.Allocated.Memory))),
			AllocatedNonPreemptibleGpu:    services.MapAllocationStatus(float64(nodePoolQuotaStatus.AllocatedNonPreemptible.Gpu)),
			AllocatedNonPreemptibleCpu:    services.MapAllocationStatus(float64(nodePoolQuotaStatus.AllocatedNonPreemptible.Cpu)),
			AllocatedNonPreemptibleMemory: services.MapAllocationStatus(services.MibToMb(float64(nodePoolQuotaStatus.AllocatedNonPreemptible.Memory))),
			RequestedGpu:                  services.MapAllocationStatus(float64(nodePoolQuotaStatus.Requested.Gpu)),
			RequestedCpu:                  services.MapAllocationStatus(float64(nodePoolQuotaStatus.Requested.Cpu)),
			RequestedMemory:               services.MapAllocationStatus(services.MibToMb(float64(nodePoolQuotaStatus.Requested.Memory))),
		}
	}

	for _, nodePoolResources := range project.OrgUnitData.Resources {
		_, found := resourcesStatuses[nodePoolResources.NodePoolName]
		if !found {
			resourcesStatuses[nodePoolResources.NodePoolName] = models.ResourcesStatus{}
		}
	}
	if len(resourcesStatuses) > 0 {
		quotaStatus := api.QuotaStatus{}
		for _, npQs := range status.QuotaStatuses {
			quotaStatus.Allocated.Gpu += float64(npQs.Allocated.Gpu)
			quotaStatus.Allocated.Cpu += float64(npQs.Allocated.Cpu)
			quotaStatus.Allocated.Memory += quotaStatusMemoryV1ToV2(float64(npQs.Allocated.Memory))
			quotaStatus.AllocatedNonPreemptible.Gpu += float64(npQs.AllocatedNonPreemptible.Gpu)
			quotaStatus.AllocatedNonPreemptible.Cpu += float64(npQs.AllocatedNonPreemptible.Cpu)
			quotaStatus.AllocatedNonPreemptible.Memory += quotaStatusMemoryV1ToV2(float64(npQs.AllocatedNonPreemptible.Memory))
			quotaStatus.Requested.Gpu += float64(npQs.Requested.Gpu)
			quotaStatus.Requested.Cpu += float64(npQs.Requested.Cpu)
			quotaStatus.Requested.Memory += quotaStatusMemoryV1ToV2(float64(npQs.Requested.Memory))
		}

		result.QuotaStatus.GpuAllocated = ptr.To(quotaStatus.Allocated.Gpu)
		result.QuotaStatus.CpuAllocated = ptr.To(quotaStatus.Allocated.Cpu)
		result.QuotaStatus.MemoryAllocated = ptr.To(quotaStatus.Allocated.Memory)
		result.QuotaStatus.GpuAllocatedNonPreemptible = ptr.To(quotaStatus.AllocatedNonPreemptible.Gpu)
		result.QuotaStatus.CpuAllocatedNonPreemptible = ptr.To(quotaStatus.AllocatedNonPreemptible.Cpu)
		result.QuotaStatus.MemoryAllocatedNonPreemptible = ptr.To(quotaStatus.AllocatedNonPreemptible.Memory)
		result.QuotaStatus.GpuRequested = ptr.To(quotaStatus.Requested.Gpu)
		result.QuotaStatus.CpuRequested = ptr.To(quotaStatus.Requested.Cpu)
		result.QuotaStatus.MemoryRequested = ptr.To(quotaStatus.Requested.Memory)
	}

	return result, resourcesStatuses, nil
}

func conditionsModelToApiV1(conditions models.ConditionArray) []map[string]string {
	marshaled, err := json.Marshal(conditions)
	if err != nil {
		return []map[string]string{}
	}

	result := []map[string]string{}
	err = json.Unmarshal(marshaled, &result)
	if err != nil {
		return []map[string]string{}
	}
	return result
}

func quotaStatusMemoryV1ToV2(memory float64) float64 {
	return services.MibToMb(memory)
}

func quotaStatusMemoryV2ToV1(memory float64) float64 {
	return services.MbToMib(memory)
}

func mapSchedulingRules(orgUnitData models.OrgUnitData, apiProject *api.ProjectV1) {
	if orgUnitData.InteractiveJobTimeLimitEffective != nil {
		apiProject.InteractiveJobTimeLimitSecs = float32(*orgUnitData.InteractiveJobTimeLimitEffective)
	}
	if orgUnitData.InteractiveJobPreemptIdleDurationEffective != nil {
		apiProject.InteractivePreemptibleJobMaxIdleDurationSecs = float32(*orgUnitData.InteractiveJobPreemptIdleDurationEffective)
	}
	if orgUnitData.InteractiveJobMaxIdleDurationEffective != nil {
		apiProject.InteractiveJobMaxIdleDurationSecs = float32(*orgUnitData.InteractiveJobMaxIdleDurationEffective)
	}
	if orgUnitData.TrainingJobMaxIdleDurationEffective != nil {
		apiProject.TrainingJobMaxIdleDurationSecs = float32(*orgUnitData.TrainingJobMaxIdleDurationEffective)
	}
	if orgUnitData.TrainingJobTimeLimitEffective != nil {
		apiProject.TrainingJobTimeLimitSecs = float32(*orgUnitData.TrainingJobTimeLimitEffective)
	}
}

func mapProjectResourcesToResponse(
	ctx context.Context, modelResources []models.Resources, id int32, memoryUnitMb bool, oqFlag bool, cpuResourcesFlag bool,
) ([]api.NodePoolAssignedResourcesV1Response, *api.AssignedResourcesV1Response, error) {
	gpuDeserved := float32(-1)
	gpuLimit := float32(-1)
	cpuDeservedNull := false
	cpuDeserved := float32(-1)
	cpuLimit := float32(-1)
	memoryDeserved := float32(-1)
	memoryDeservedNull := false
	memoryLimit := float32(-1)
	gpuOq := ptr.To(float32(2))
	cpuOq := ptr.To(float32(2))
	memoryOq := ptr.To(float32(2))

	apiNodePoolResources := []api.NodePoolAssignedResourcesV1Response{}
	for _, nodePoolModelResource := range modelResources {
		nodePoolId, err := getNodePoolId(nodePoolModelResource.NodePoolId)
		if err != nil {
			return nil, nil, err
		}
		singleApiNodePoolResources := api.NodePoolAssignedResourcesV1Response{
			Id: float32(id),
			NodePool: api.NodePoolAssignedResourcesV1ResponseNodePool{
				Id:   nodePoolId,
				Name: nodePoolModelResource.NodePoolName,
			},
			Gpu: api.ResourceV1Response{
				Deserved:        ptrToFloat32(nodePoolModelResource.GpuDeserved, nil),
				MaxAllowed:      ptrToFloat32(nodePoolModelResource.GpuLimit, ptr.To(UnlimitedValue)),
				OverQuotaWeight: ptrToFloat32(nodePoolModelResource.GpuOverQuotaWeight, nil),
			},
		}
		if cpuResourcesFlag {
			singleApiNodePoolResources.Cpu = api.ResourceV1Response{
				Deserved:        ptrToFloat32(nodePoolModelResource.CpuDeserved, nil),
				MaxAllowed:      ptrToFloat32(nodePoolModelResource.CpuLimit, ptr.To(UnlimitedValue)),
				OverQuotaWeight: ptrToFloat32(nodePoolModelResource.CpuOverQuotaWeight, nil),
			}
			conversionFn := services.GetConversionFn(services.ModelUnitsToApi(nodePoolModelResource.MemoryUnits), memoryUnitMb)
			singleApiNodePoolResources.Memory = api.ResourceV1Response{
				Deserved:        ptrToFloat32(services.ApplyMemoryConversion(nodePoolModelResource.MemoryDeserved, conversionFn), nil),
				MaxAllowed:      ptrToFloat32(services.ApplyMemoryConversion(nodePoolModelResource.MemoryLimit, conversionFn), ptr.To(UnlimitedValue)),
				OverQuotaWeight: ptrToFloat32(nodePoolModelResource.MemoryOverQuotaWeight, nil),
			}
		}
		if !oqFlag {
			singleApiNodePoolResources.Gpu.OverQuotaWeight = singleApiNodePoolResources.Gpu.Deserved
			if cpuResourcesFlag {
				singleApiNodePoolResources.Cpu.OverQuotaWeight = singleApiNodePoolResources.Cpu.Deserved
				singleApiNodePoolResources.Memory.OverQuotaWeight = singleApiNodePoolResources.Memory.Deserved
			}
		}
		deserved := ptrToFloat32(nodePoolModelResource.GpuDeserved, nil)
		limit := ptrToFloat32(nodePoolModelResource.GpuLimit, nil)
		gpuDeserved, gpuLimit = updateJoinedDeservedLimit(gpuDeserved, gpuLimit, deserved, limit)
		if cpuResourcesFlag {
			if nodePoolModelResource.CpuDeserved == nil {
				cpuDeservedNull = true
			} else {
				deserved = ptrToFloat32(nodePoolModelResource.CpuDeserved, nil)
				limit = ptrToFloat32(nodePoolModelResource.CpuLimit, nil)
				cpuDeserved, cpuLimit = updateJoinedDeservedLimit(cpuDeserved, cpuLimit, deserved, limit)
			}
			if nodePoolModelResource.MemoryDeserved == nil {
				memoryDeservedNull = true
			} else {
				deserved = ptrToFloat32(nodePoolModelResource.MemoryDeserved, nil)
				limit = ptrToFloat32(nodePoolModelResource.MemoryLimit, nil)
				memoryDeserved, memoryLimit = updateJoinedDeservedLimit(memoryDeserved, memoryLimit, deserved, limit)
			}
		}
		apiNodePoolResources = append(apiNodePoolResources, singleApiNodePoolResources)
	}
	if len(modelResources) > 0 {
		gpuOq = ptrToFloat32(modelResources[0].GpuOverQuotaWeight, nil)
		cpuOq = ptrToFloat32(modelResources[0].CpuOverQuotaWeight, nil)
		memoryOq = ptrToFloat32(modelResources[0].MemoryOverQuotaWeight, nil)
	}
	var legacyCpuDeserved *float32 = nil
	var legacyCpuLimit *float32 = nil
	var legacyMemoryDeserved *float32 = nil
	var legacyMemoryLimit *float32 = nil
	if cpuResourcesFlag {
		if !cpuDeservedNull {
			legacyCpuDeserved = ptr.To(cpuDeserved)
			legacyCpuLimit = ptr.To(cpuLimit)
		}
		if !memoryDeservedNull {
			legacyMemoryDeserved = ptr.To(memoryDeserved)
			legacyMemoryLimit = ptr.To(memoryLimit)
		}
	}
	apiResources := mapProjectApiResourcesToResponse(ctx, modelResources, gpuDeserved, gpuLimit, gpuOq, legacyCpuDeserved, legacyCpuLimit, cpuOq, legacyMemoryDeserved, legacyMemoryLimit, memoryOq, oqFlag)
	if !oqFlag {
		updateApiOverQuotaWeight(apiResources, apiNodePoolResources)
	}
	return apiNodePoolResources, apiResources, nil
}

func mapProjectResourcesToResponse2(modelResources []models.Resources,
	id int32,
	memoryUnitMb bool,
	nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal,
) ([]api.NodePoolAssignedResourcesV1Response, error) {
	apiNodePoolResources := []api.NodePoolAssignedResourcesV1Response{}
	for _, nodePoolModelResource := range modelResources {
		nodePoolId, err := getNodePoolId(nodePoolModelResource.NodePoolId)
		if err != nil {
			return nil, err
		}
		singleApiNodePoolResources := api.NodePoolAssignedResourcesV1Response{
			Id: float32(id),
			NodePool: api.NodePoolAssignedResourcesV1ResponseNodePool{
				Id:   nodePoolId,
				Name: nodePoolModelResource.NodePoolName,
			},
			Gpu: api.ResourceV1Response{
				Deserved:        ptrToFloat32(nodePoolModelResource.GpuDeserved, nil),
				MaxAllowed:      ptrToFloat32(nodePoolModelResource.GpuLimit, ptr.To(UnlimitedValue)),
				OverQuotaWeight: ptrToFloat32(nodePoolModelResource.GpuOverQuotaWeight, nil),
			},
		}

		npSchedulingCapabilities, err := services.GetSchedulingCapabilitiesPerNodePoolId(nodePoolsMetadata, nodePoolModelResource.NodePoolId)
		if err != nil {
			return nil, err
		}
		oqFlag := npSchedulingCapabilities.AdjustableWeights
		cpuResourcesFlag := npSchedulingCapabilities.CpuQuota

		if cpuResourcesFlag {
			singleApiNodePoolResources.Cpu = api.ResourceV1Response{
				Deserved:        ptrToFloat32(nodePoolModelResource.CpuDeserved, nil),
				MaxAllowed:      ptrToFloat32(nodePoolModelResource.CpuLimit, ptr.To(UnlimitedValue)),
				OverQuotaWeight: ptrToFloat32(nodePoolModelResource.CpuOverQuotaWeight, nil),
			}
			conversionFn := services.GetConversionFn(services.ModelUnitsToApi(nodePoolModelResource.MemoryUnits), memoryUnitMb)
			singleApiNodePoolResources.Memory = api.ResourceV1Response{
				Deserved:        ptrToFloat32(services.ApplyMemoryConversion(nodePoolModelResource.MemoryDeserved, conversionFn), nil),
				MaxAllowed:      ptrToFloat32(services.ApplyMemoryConversion(nodePoolModelResource.MemoryLimit, conversionFn), ptr.To(UnlimitedValue)),
				OverQuotaWeight: ptrToFloat32(nodePoolModelResource.MemoryOverQuotaWeight, nil),
			}
		}
		if !oqFlag {
			singleApiNodePoolResources.Gpu.OverQuotaWeight = singleApiNodePoolResources.Gpu.Deserved
			if cpuResourcesFlag {
				singleApiNodePoolResources.Cpu.OverQuotaWeight = singleApiNodePoolResources.Cpu.Deserved
				singleApiNodePoolResources.Memory.OverQuotaWeight = singleApiNodePoolResources.Memory.Deserved
			}
		}
		apiNodePoolResources = append(apiNodePoolResources, singleApiNodePoolResources)
	}
	return apiNodePoolResources, nil
}

func updateJoinedDeservedLimit(joinedDeserved float32, joinedLimit float32, deserved *float32, limit *float32) (float32, float32) {
	if deserved != nil {
		if joinedDeserved == -1 || *deserved == -1 {
			joinedDeserved = *deserved
		} else {
			joinedDeserved = joinedDeserved + *deserved
		}
	}
	if limit != nil {
		if joinedLimit == -1 || *limit == -1 {
			joinedLimit = *limit
		} else {
			joinedLimit = joinedLimit + *limit
		}
	}
	return joinedDeserved, joinedLimit
}

func updateApiOverQuotaWeight(apiResources *api.AssignedResourcesV1Response, apiNodePoolResources []api.NodePoolAssignedResourcesV1Response) {
	apiResources.Gpu.OverQuotaWeight = apiResources.Gpu.Deserved
	apiResources.Cpu.OverQuotaWeight = apiResources.Cpu.Deserved
	apiResources.Memory.OverQuotaWeight = apiResources.Memory.Deserved
	for _, singleResources := range apiNodePoolResources {
		singleResources.Gpu.OverQuotaWeight = singleResources.Gpu.Deserved
		if singleResources.Cpu.Deserved != nil {
			*singleResources.Cpu.OverQuotaWeight = *singleResources.Cpu.Deserved
		}
		if singleResources.Memory.Deserved != nil {
			*singleResources.Memory.OverQuotaWeight = *singleResources.Memory.Deserved
		}
	}
}

func mapProjectApiResourcesToResponse(ctx context.Context, modelResources []models.Resources, gpuDeserved float32, gpuLimit float32, gpuOq *float32,
	cpuDeserved *float32, cpuLimit *float32, cpuOq *float32, memoryDeserved *float32, memoryLimit *float32, memoryOq *float32, oqFlag bool,
) *api.AssignedResourcesV1Response {
	apiResources := api.AssignedResourcesV1Response{
		Gpu: api.ResourceV1Response{
			Deserved:        ptr.To(gpuDeserved),
			MaxAllowed:      ptr.To(gpuLimit),
			OverQuotaWeight: gpuOq,
		},
	}
	apiResources.Cpu.Deserved = cpuDeserved
	apiResources.Cpu.MaxAllowed = cpuLimit
	apiResources.Cpu.OverQuotaWeight = cpuOq
	apiResources.Memory.Deserved = memoryDeserved
	apiResources.Memory.MaxAllowed = memoryLimit
	apiResources.Memory.OverQuotaWeight = memoryOq
	if !oqFlag {
		apiResources.Gpu.OverQuotaWeight = ptr.To(gpuDeserved)
		apiResources.Cpu.OverQuotaWeight = cpuDeserved
		apiResources.Memory.OverQuotaWeight = memoryDeserved
	}
	unlimited := float64(-1)
	if len(modelResources) > 0 {
		if modelResources[0].GpuLimit != nil && *modelResources[0].GpuLimit == unlimited {
			apiResources.Gpu.MaxAllowed = ptrToFloat32(modelResources[0].GpuLimit, nil)
		}
		if modelResources[0].CpuLimit != nil && *modelResources[0].CpuLimit == unlimited {
			apiResources.Cpu.MaxAllowed = ptrToFloat32(modelResources[0].CpuLimit, nil)
		}
		if modelResources[0].MemoryLimit != nil && *modelResources[0].MemoryLimit == unlimited {
			apiResources.Memory.MaxAllowed = ptrToFloat32(modelResources[0].MemoryLimit, nil)
		}
	}
	return &apiResources
}

func MapProjectsMetricsToResponse(
	ctx context.Context,
	projects []*models.Project,
	currentMetrics map[string]map[string][]CurrentMetricMeasurement,
	timeRangeMetrics map[string][]metrics_api.Result,
	clusterId string,
	clusterName string,
	fetchUtilization bool,
) (api.ProjectsMetricsV1Response, error) {
	response := api.ProjectsMetricsV1Response{}

	timeRangeResources, err := MapTimeRangeMetricsToApi(timeRangeMetrics, fetchUtilization)
	if err != nil {
		return response, err
	}

	response.TimeRange = api.TimeRangeV1{Resources: timeRangeResources}

	data := []api.ProjectsMetricsV1ResponseDataInner{}
	for _, project := range projects {
		projectMetric, found := currentMetrics[project.Name]
		if !found {
			continue
		}
		currentProjectResources, err := MapCurrentMetricsToApi(ctx, project.OrgUnitData, projectMetric)
		if err != nil {
			return response, err
		}
		data = append(data, api.ProjectsMetricsV1ResponseDataInner{
			Metadata: api.ProjectMetricsV1Metadata{
				ProjectId:      int32(project.ID),
				ProjectName:    project.Name,
				DepartmentId:   int32(*project.ParentID),
				DepartmentName: convutils.Deref(project.ParentName, ""),
				ClusterId:      clusterId,
				ClusterName:    clusterName,
			},
			Current: api.ProjectMetricsV1Current{Resources: currentProjectResources},
		})

	}
	response.Data = data
	return response, nil
}

func MapQuotaMetricsAndProjectsToResponse(ctx context.Context, projects []*models.Project, metricsResults []metrics_api.Result, tenantId string, clusterId string) ([]api.ProjectQuotasV1Inner, error) {
	response := []api.ProjectQuotasV1Inner{}
	projectToMetricMap := make(map[string]metrics_api.Result)
	for i := range metricsResults {
		projectName := metricsResults[i].GetMetricLabels()["project"]
		projectToMetricMap[projectName] = metricsResults[i]
	}
	for _, project := range projects {
		allocatedGpu := float64(0)
		projectMetric, found := projectToMetricMap[project.Name]
		if found && len(projectMetric.Measurements) > 0 {
			float64Value, err := strconv.ParseFloat(projectMetric.Measurements[0].Value, 32)
			if err != nil {
				return response, err
			}
			allocatedGpu = allocatedGpu + float64Value
		}
		deservedGpu := float64(0)
		for _, resources := range project.OrgUnitData.Resources {
			deservedGpu = deservedGpu + *resources.GpuDeserved
		}
		numberTenant, _ := strconv.Atoi(tenantId)
		deservedGpu32 := float32(deservedGpu)
		response = append(response, api.ProjectQuotasV1Inner{
			Id:            int32(project.ID),
			TenantId:      int32(numberTenant),
			ClusterUuid:   clusterId,
			Name:          project.Name,
			DeservedGpus:  deservedGpu32,
			AllocatedGpus: float32(allocatedGpu),
		})
	}
	return response, nil
}

func MapProjectMetricsToResponse(
	ctx context.Context,
	clusterId string,
	clusterName string,
	project *models.Project,
	currentMetrics map[string][]CurrentMetricMeasurement,
	timeRangeMetrics map[string][]metrics_api.Result,
) (api.ProjectMetricsV1Response, error) {
	response := api.ProjectMetricsV1Response{}

	currentResources, err := MapCurrentMetricsToApi(ctx, project.OrgUnitData, currentMetrics)
	if err != nil {
		return response, err
	}

	timeRangeResources, err := MapTimeRangeMetricsToApi(timeRangeMetrics, false)
	if err != nil {
		return response, err
	}
	response.Metadata = api.ProjectMetricsV1Metadata{
		ProjectId:      int32(project.ID),
		ProjectName:    project.Name,
		DepartmentId:   int32(*project.ParentID),
		DepartmentName: convutils.Deref(project.ParentName, ""),
		ClusterId:      clusterId,
		ClusterName:    clusterName,
	}
	response.Current = api.ProjectMetricsV1Current{Resources: currentResources}
	response.TimeRange = api.TimeRangeV1{Resources: timeRangeResources}
	return response, nil
}
