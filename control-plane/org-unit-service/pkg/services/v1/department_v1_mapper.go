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

	cluster_api "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	metrics_api "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"

	"k8s.io/utils/ptr"
)

type AggregatedResources struct {
	NodePool AggregatedResourcesNodePool `json:"nodePool"`
	Gpu      AggregatedResource          `json:"gpu"`
	Cpu      AggregatedResource          `json:"cpu"`
	Memory   AggregatedResource          `json:"memory"`
}

type AggregatedResourcesNodePool struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AggregatedResource struct {
	Deserved *float32 `json:"deserved"`
}

func MapDepartmentModelToV1Response(
	ctx context.Context, department models.Department, departmentAdmins *[]openapi.AccessRule, memoryUnitMb bool, nodePools map[string]cluster_api.NodepoolMinimal,
) (api.DepartmentForGetV1, error) {
	tenantId64, err := strconv.Atoi(department.TenantId)
	if err != nil {
		return api.DepartmentForGetV1{}, err
	}
	tenantId := int32(tenantId64)
	if int(tenantId) != tenantId64 {
		return api.DepartmentForGetV1{}, fmt.Errorf("failed to map department. id %d is not int32", tenantId64)
	}
	departmentId := int32(department.ID)
	if int64(departmentId) != department.ID {
		return api.DepartmentForGetV1{}, fmt.Errorf("failed to map department. id %d is not int32", department.ID)
	}
	projects, err := mapDepartmentProjects(ctx, department.ProjectChildren)
	if err != nil {
		return api.DepartmentForGetV1{}, err
	}
	orgUnitData := department.OrgUnitData
	nodePoolAssignedResources, err := mapDepartmentResourcesToResponse(ctx, orgUnitData.Resources, departmentId, memoryUnitMb, nodePools)
	if err != nil {
		return api.DepartmentForGetV1{}, err
	}

	projectsDeservedGpus := getProjectsDeservedGpus(department.ProjectChildren)

	projectResources, projectQuotas, err := mapDepartmentProjectToAdditionalData(ctx, department.ProjectChildren)
	if err != nil {
		return api.DepartmentForGetV1{}, err
	}

	response := api.DepartmentForGetV1{
		Id:                                  departmentId,
		Name:                                department.Name,
		TenantId:                            tenantId,
		ClusterUuid:                         department.ClusterId.String(),
		CreatedAt:                           department.CreatedAt,
		ProjectsDeservedGpus:                projectsDeservedGpus,
		Projects:                            projects,
		NodePoolsResources:                  nodePoolAssignedResources,
		QuotaStatuses:                       &projectQuotas,
		ProjectAggregatedNodePoolsResources: &projectResources,
	}
	if departmentAdmins != nil {
		response.DepartmentAdmins = getDepartmentAdmins(*departmentAdmins)
	}
	return response, nil
}

func mapDepartmentProjectToAdditionalData(ctx context.Context, projects []models.Project) ([]map[string]interface{}, []map[string]interface{}, error) {
	nodePoolToResources := make(map[string]AggregatedResources)
	nodePoolToQuota := make(map[string]*api.QuotaStatusV1Inner)
	for _, project := range projects {
		for _, resources := range project.OrgUnitData.Resources {
			mapResources(ctx, resources, nodePoolToResources)
		}
		for _, projQuota := range project.OrgUnitData.Resources {
			mapQuotas(ctx, nodePoolToQuota, projQuota)
		}
	}
	projectsResources := []AggregatedResources{}
	for _, resources := range nodePoolToResources {
		projectsResources = append(projectsResources, resources)
	}
	projectsQuotas := []api.QuotaStatusV1Inner{}
	for _, quotas := range nodePoolToQuota {
		projectsQuotas = append(projectsQuotas, *quotas)
	}

	jsonResources, err := json.Marshal(projectsResources)
	if err != nil {
		return nil, nil, err
	}
	jsonQuotas, err := json.Marshal(projectsQuotas)
	if err != nil {
		return nil, nil, err
	}

	resourcesResult := []map[string]interface{}{}
	err = json.Unmarshal(jsonResources, &resourcesResult)
	if err != nil {
		return nil, nil, err
	}
	quotasResult := []map[string]interface{}{}
	err = json.Unmarshal(jsonQuotas, &quotasResult)
	if err != nil {
		return nil, nil, err
	}

	return resourcesResult, quotasResult, nil
}

func mapQuotas(ctx context.Context, nodePoolToQuota map[string]*api.QuotaStatusV1Inner, resources models.Resources) {
	aggregatedQuotas, found := nodePoolToQuota[resources.NodePoolName]
	if !found {
		aggregatedQuotas = initDefaultQuotaStatus(resources.NodePoolName)
		nodePoolToQuota[resources.NodePoolName] = aggregatedQuotas
	}
	aggregatedQuotas.Allocated.Gpu += ptr.Deref(ptrToFloat32(resources.AllocatedGpu, nil), 0)
	aggregatedQuotas.Allocated.Cpu += ptr.Deref(ptrToFloat32(resources.AllocatedCpu, nil), 0)
	aggregatedQuotas.Allocated.Memory += float32(quotaStatusMemoryV2ToV1(ptr.Deref(resources.AllocatedMemory, 0)))
	aggregatedQuotas.AllocatedNonPreemptible.Gpu += ptr.Deref(ptrToFloat32(resources.AllocatedNonPreemptibleGpu, nil), 0)
	aggregatedQuotas.AllocatedNonPreemptible.Cpu += ptr.Deref(ptrToFloat32(resources.AllocatedNonPreemptibleCpu, nil), 0)
	aggregatedQuotas.AllocatedNonPreemptible.Memory += float32(quotaStatusMemoryV2ToV1(ptr.Deref(resources.AllocatedNonPreemptibleMemory, 0)))
	aggregatedQuotas.Requested.Gpu += ptr.Deref(ptrToFloat32(resources.RequestedGpu, nil), 0)
	aggregatedQuotas.Requested.Cpu += ptr.Deref(ptrToFloat32(resources.RequestedCpu, nil), 0)
	aggregatedQuotas.Requested.Memory += float32(quotaStatusMemoryV2ToV1(ptr.Deref(resources.RequestedMemory, 0)))
}

func initDefaultQuotaStatus(nodePoolName string) *api.QuotaStatusV1Inner {
	return &api.QuotaStatusV1Inner{
		NodePoolName: nodePoolName,
		Allocated: api.QuotaStatusResourceListV1{
			Gpu:    0,
			Cpu:    0,
			Memory: 0,
		},
		AllocatedNonPreemptible: api.QuotaStatusResourceListV1{
			Gpu:    0,
			Cpu:    0,
			Memory: 0,
		},
		Requested: api.QuotaStatusResourceListV1{
			Gpu:    0,
			Cpu:    0,
			Memory: 0,
		},
	}
}

func mapResources(ctx context.Context, resources models.Resources, nodePoolToResources map[string]AggregatedResources) {
	aggregatedResources, found := nodePoolToResources[resources.NodePoolId]
	if !found {
		newResources := AggregatedResources{
			NodePool: AggregatedResourcesNodePool{
				Id:   resources.NodePoolId,
				Name: resources.NodePoolName,
			},
			Gpu:    AggregatedResource{Deserved: ptr.To(float32(0))},
			Cpu:    AggregatedResource{Deserved: ptr.To(float32(0))},
			Memory: AggregatedResource{Deserved: ptr.To(float32(0))},
		}
		aggregatedResources = newResources
	}
	aggregatedResources.Gpu.Deserved = getDeservedSum(aggregatedResources.Gpu, ptrToFloat32(resources.GpuDeserved, nil))
	aggregatedResources.Cpu.Deserved = getDeservedSum(aggregatedResources.Cpu, ptrToFloat32(resources.CpuDeserved, nil))
	aggregatedResources.Memory.Deserved = getDeservedSum(aggregatedResources.Memory, ptrToFloat32(resources.MemoryDeserved, nil))
	nodePoolToResources[resources.NodePoolId] = aggregatedResources
}

func getDeservedSum(currentDeserved AggregatedResource, newDeserved *float32) *float32 {
	if currentDeserved.Deserved == nil || *currentDeserved.Deserved == -1 {
		return currentDeserved.Deserved
	}
	if newDeserved == nil {
		return ptr.To(float32(0))
	}
	if *newDeserved == -1 {
		return newDeserved
	}
	sum := *currentDeserved.Deserved + *newDeserved
	return ptr.To(sum)
}

func MapDepartmentsModelToV1Response(ctx context.Context,
	departments []models.Department,
	departmentAdmins map[string]*[]openapi.AccessRule,
	memoryUnitMb bool,
	nodePools map[string]cluster_api.NodepoolMinimal,
) ([]api.DepartmentForGetV1, error) {
	apiDepartments := []api.DepartmentForGetV1{}
	for _, department := range departments {
		apiDepartment, err := MapDepartmentModelToV1Response(ctx, department, departmentAdmins[strconv.FormatInt(department.ID, 10)], memoryUnitMb, nodePools)
		if err != nil {
			return apiDepartments, err
		}
		apiDepartments = append(apiDepartments, apiDepartment)
	}
	return apiDepartments, nil
}

func getDepartmentAdmins(accessRules []openapi.AccessRule) []string {
	departmentAdmins := []string{}
	for _, accessRule := range accessRules {
		departmentAdmins = append(departmentAdmins, accessRule.SubjectId)
	}
	return departmentAdmins
}

func mapDepartmentProjects(ctx context.Context, children []models.Project) ([]api.ProjectV1, error) {
	departmentProjects := []api.ProjectV1{}
	for _, child := range children {
		deservedGpus := float64(0)
		for _, resources := range child.OrgUnitData.Resources {
			deservedGpus = deservedGpus + *resources.GpuDeserved
		}
		projectId := int32(child.ID)
		if int64(projectId) != child.ID {
			return nil, fmt.Errorf("failed to map project. id %d is not int32", child.ID)
		}
		gpus := float32(deservedGpus)
		departmentProject := api.ProjectV1{
			Id:           projectId,
			Name:         child.Name,
			DeservedGpus: gpus,
		}
		departmentProjects = append(departmentProjects, departmentProject)
	}
	return departmentProjects, nil
}

func mapDepartmentResourcesToResponse(
	ctx context.Context,
	modelResources []models.Resources,
	id int32,
	memoryUnitMb bool,
	nodePools map[string]cluster_api.NodepoolMinimal,
) ([]api.NodePoolAssignedResourcesV1Response, error) {
	gpuDeserved := float32(-1)
	gpuLimit := float32(-1)
	cpuDeserved := float32(-1)
	cpuLimit := float32(-1)
	memoryDeserved := float32(-1)
	memoryLimit := float32(-1)
	apiNodePoolResources := []api.NodePoolAssignedResourcesV1Response{}
	for _, nodePoolModelResource := range modelResources {
		nodePoolId64, err := strconv.ParseFloat(nodePoolModelResource.NodePoolId, 32)
		if err != nil {
			return apiNodePoolResources, err
		}
		nodePoolId := float32(nodePoolId64)
		if float64(nodePoolId) != nodePoolId64 {
			return apiNodePoolResources, fmt.Errorf("failed to map department resources. nodepool id %f is not float32", nodePoolId64)
		}

		npSchedulingCapabilities, err := services.GetSchedulingCapabilitiesPerNodePoolId(nodePools, nodePoolModelResource.NodePoolId)
		if err != nil {
			return apiNodePoolResources, err
		}
		cpuResourcesFlag := npSchedulingCapabilities.CpuQuota

		conversionFn := services.GetConversionFn(services.ModelUnitsToApi(nodePoolModelResource.MemoryUnits), memoryUnitMb)
		singleApiNodePoolResources := MapSingleNodePool(ctx, id, nodePoolId, nodePoolModelResource, conversionFn, cpuResourcesFlag)
		gpuDeserved, gpuLimit = updateJoinedDeservedLimit(gpuDeserved, gpuLimit, singleApiNodePoolResources.Gpu.Deserved, singleApiNodePoolResources.Gpu.MaxAllowed)
		if cpuResourcesFlag {
			if nodePoolModelResource.CpuDeserved == nil {
			} else {
				cpuDeserved, cpuLimit = updateJoinedDeservedLimit(cpuDeserved, cpuLimit, singleApiNodePoolResources.Cpu.Deserved, singleApiNodePoolResources.Cpu.MaxAllowed)
			}
			if nodePoolModelResource.MemoryDeserved == nil {
			} else {
				memoryDeserved, memoryLimit = updateJoinedDeservedLimit(memoryDeserved, memoryLimit, singleApiNodePoolResources.Memory.Deserved, singleApiNodePoolResources.Memory.MaxAllowed)
			}
		}
		apiNodePoolResources = append(apiNodePoolResources, singleApiNodePoolResources)
	}
	return apiNodePoolResources, nil
}

func MapDepartmentMetricsToResponse(
	ctx context.Context,
	clusterId string,
	clusterName string,
	department *models.Department,
	departmentCurrentMetrics map[string]map[string][]CurrentMetricMeasurement,
	projectCurrentMetrics map[string]map[string][]CurrentMetricMeasurement,
	timeRangeMetrics map[string][]metrics_api.Result,
) (api.DepartmentMetricsV1Response, error) {
	response, err := mapCurrentMetricsResponse(ctx, clusterId, clusterName, department, departmentCurrentMetrics, projectCurrentMetrics)
	if err != nil {
		return response, err
	}
	timeRangeResources, err := MapTimeRangeMetricsToApi(timeRangeMetrics, false)
	if err != nil {
		return response, err
	}
	response.TimeRange = api.TimeRangeV1{Resources: timeRangeResources}
	return response, nil
}

func mapCurrentMetricsResponse(
	ctx context.Context,
	clusterId string,
	clusterName string,
	department *models.Department,
	departmentCurrentMetrics map[string]map[string][]CurrentMetricMeasurement,
	projectCurrentMetrics map[string]map[string][]CurrentMetricMeasurement,
) (api.DepartmentMetricsV1Response, error) {
	response := api.DepartmentMetricsV1Response{}

	currentDepartmentResources, err := MapCurrentMetricsToApi(ctx, department.OrgUnitData, departmentCurrentMetrics[department.Name])

	projectCurrentResources := []api.DepartmentMetricsV1CurrentProjectResourcesInner{}
	for _, project := range department.ProjectChildren {
		currentProjResources, err := MapCurrentMetricsToApi(ctx, project.OrgUnitData, projectCurrentMetrics[project.Name])
		if err != nil {
			return response, err
		}
		for _, currentProjNpResources := range currentProjResources {
			projectCurrentResources = append(projectCurrentResources, api.DepartmentMetricsV1CurrentProjectResourcesInner{
				NumberOfPendingWorkloads: currentProjNpResources.NumberOfPendingWorkloads,
				Gpu:                      currentProjNpResources.Gpu,
				Cpu:                      currentProjNpResources.Cpu,
				Memory:                   currentProjNpResources.Memory,
				ProjectName:              project.Name,
				NodepoolName:             currentProjNpResources.NodepoolName,
			})
		}
	}

	if err != nil {
		return response, err
	}
	response.Metadata = api.DepartmentMetricsV1Metadata{
		DepartmentId:   int32(department.ID),
		DepartmentName: department.Name,
		ClusterId:      clusterId,
		ClusterName:    clusterName,
	}
	response.Current = api.DepartmentMetricsV1Current{
		Resources:        currentDepartmentResources,
		ProjectResources: projectCurrentResources,
	}
	return response, nil
}

func MapDepartmentsMetricsToResponse(
	ctx context.Context,
	clusterId string,
	clusterName string,
	departments []models.Department,
	departmentCurrentMetrics map[string]map[string][]CurrentMetricMeasurement,
	projectCurrentMetrics map[string]map[string][]CurrentMetricMeasurement,
	timeRangeMetrics map[string][]metrics_api.Result,
	fetchUtilization bool,
) (api.DepartmentsMetricsV1Response, error) {
	response := api.DepartmentsMetricsV1Response{}
	data := []api.DepartmentsMetricsV1ResponseDataInner{}
	for _, department := range departments {
		metricsResponse, err := mapCurrentMetricsResponse(ctx, clusterId, clusterName, &department, departmentCurrentMetrics, projectCurrentMetrics)
		if err != nil {
			return response, err
		}
		data = append(data, api.DepartmentsMetricsV1ResponseDataInner{
			Metadata: metricsResponse.Metadata,
			Current:  metricsResponse.Current,
		})
	}

	timeRangeResources, err := MapTimeRangeMetricsToApi(timeRangeMetrics, fetchUtilization)
	if err != nil {
		return response, err
	}
	response.Data = data
	response.TimeRange = api.TimeRangeV1{Resources: timeRangeResources}
	return response, nil
}
