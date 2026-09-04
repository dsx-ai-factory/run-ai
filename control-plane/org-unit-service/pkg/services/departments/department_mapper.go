/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package departments

import (
	"context"
	"strconv"

	"k8s.io/utils/ptr"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/google/uuid"
)

func MapCreateRequestToModel(ctx context.Context, request api.DepartmentCreationRequest, nodePoolsMetadata map[string]openapi.NodepoolMinimal) (models.Department, error) {
	clusterId, err := uuid.Parse(request.ClusterId)
	if err != nil {
		return models.Department{}, err
	}
	data, err := services.MapRequestToOrgUnitData(ctx, request.SchedulingRules, request.NodeTypes, request.DefaultNodePools, request.Resources, true, nodePoolsMetadata, nil, nil)
	if err != nil {
		return models.Department{}, err
	}

	return models.Department{
		Name:        request.Name,
		Description: request.Description,
		ClusterId:   clusterId,
		OrgUnitData: data,
	}, nil
}

func MapUpdateRequestToModel(ctx context.Context, request api.DepartmentUpdateRequest, department *models.Department, nodePoolsMetadata map[string]openapi.NodepoolMinimal) (models.Department, error) {
	data, err := services.MapRequestToOrgUnitData(ctx, request.SchedulingRules, request.NodeTypes, request.DefaultNodePools, request.Resources, true, nodePoolsMetadata, nil, &department.OrgUnitData)
	if err != nil {
		return models.Department{}, err
	}

	update := models.Department{
		TenantId: department.TenantId,
		IntBaseModel: models.IntBaseModel{
			ID:        department.ID,
			CreatedBy: department.CreatedBy,
			CreatedAt: department.CreatedAt,
		},
		Name:          department.Name,
		ClusterId:     department.ClusterId,
		ParentID:      department.ParentID,
		OrgUnitDataID: department.OrgUnitDataID,
		OrgUnitData:   data,
	}
	update.Description = request.Description
	return update, nil
}

func MapModelToResponse(department models.Department, verbosity models.Verbosity, nodePoolsMetadata map[string]openapi.NodepoolMinimal) api.Department {
	response := mapBasicDepartment(department, nodePoolsMetadata)

	if verbosity == models.VerboseVerbosity {
		response.Status = api.DepartmentAllOfStatus{
			QuotaStatus: api.QuotaStatus{
				Allocated: api.QuotaStatusResource{
					Gpu:    ptr.Deref(department.QuotaStatus.GpuAllocated, 0),
					Cpu:    ptr.Deref(department.QuotaStatus.CpuAllocated, 0),
					Memory: ptr.Deref(department.QuotaStatus.MemoryAllocated, 0),
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    ptr.Deref(department.QuotaStatus.GpuAllocatedNonPreemptible, 0),
					Cpu:    ptr.Deref(department.QuotaStatus.CpuAllocatedNonPreemptible, 0),
					Memory: ptr.Deref(department.QuotaStatus.MemoryAllocatedNonPreemptible, 0),
				},
				Requested: api.QuotaStatusResource{
					Gpu:    ptr.Deref(department.QuotaStatus.GpuRequested, 0),
					Cpu:    ptr.Deref(department.QuotaStatus.CpuRequested, 0),
					Memory: ptr.Deref(department.QuotaStatus.MemoryRequested, 0),
				},
			},
		}

		response.Children = mapChildrenToResponse(department.OrgUnitData.Children)
		response.ProjectsAggregatedResources = mapProjectAggregatedNodePoolResources(department.ProjectChildren)
		response.Status.NodePoolQuotaStatuses = mapQuotaStatusesToResponse(department.ProjectChildren)
	}
	return response
}

func MapModelToSyncResponse(department models.Department, nodePoolsMetadata map[string]openapi.NodepoolMinimal) api.DepartmentSync {
	resources := departmentApiResourcesFromModelSync(department, nodePoolsMetadata)
	response := api.DepartmentSync{
		Id:        strconv.FormatInt(department.ID, 10),
		Name:      department.Name,
		ClusterId: department.ClusterId.String(),
		CreatedAt: department.CreatedAt,
		UpdatedAt: department.UpdatedAt,
		Resources: resources,
	}
	if department.DeletedAt.Valid {
		response.DeletedAt = &department.DeletedAt.Time
	}
	return response
}

func mapBasicDepartment(department models.Department, nodePools map[string]openapi.NodepoolMinimal) api.Department {
	resources := departmentApiResourcesFromModel(department, nodePools)
	orgUnitData := department.OrgUnitData
	response := api.Department{
		Id:          strconv.FormatInt(department.ID, 10),
		CreatedAt:   department.CreatedAt,
		UpdatedAt:   department.UpdatedAt,
		CreatedBy:   department.CreatedBy,
		UpdatedBy:   department.UpdatedBy,
		Name:        department.Name,
		Description: department.Description,
		ClusterId:   department.ClusterId.String(),
		Resources:   resources,
		TotalResources: api.TotalResources{
			GpuQuota:    orgUnitData.TotalGpuQuota,
			CpuQuota:    orgUnitData.TotalCpuQuota,
			MemoryQuota: orgUnitData.TotalMemoryQuota,
		},
		OvertimeData: services.MapOvertimeDataModelToApi(orgUnitData.OvertimeData, department.Name, "department"),
	}
	if orgUnitData.DefaultNodePoolList != nil {
		apiNodePools := []string(*orgUnitData.DefaultNodePoolList)
		response.DefaultNodePools = &apiNodePools
	}
	response.NodeTypes = services.ApiNodeTypesFromModel(orgUnitData.TrainingNodeTypes, orgUnitData.WorkspaceNodeTypes)
	schedulingRules := services.MapSchedulingRules(orgUnitData)
	if (schedulingRules != api.SchedulingRules{}) {
		response.SchedulingRules = &schedulingRules
	}

	return response
}

func departmentApiResourcesFromModel(department models.Department, nodePoolsMetadata map[string]openapi.NodepoolMinimal) []api.Resources {
	return services.ApiResourcesFromModel(department.OrgUnitData.Resources, nodePoolsMetadata)
}

func departmentApiResourcesFromModelSync(department models.Department, nodePoolsMetadata map[string]openapi.NodepoolMinimal) []api.ResourcesSync {
	return services.ApiResourcesFromModelSync(department.OrgUnitData.Resources, nodePoolsMetadata)
}

func mapChildrenToResponse(children []models.Child) *[]api.DepartmentProjectChild {
	apiChildren := &[]api.DepartmentProjectChild{}
	for _, child := range children {
		apiChild := api.DepartmentProjectChild{
			Id:   strconv.FormatInt(child.Id, 10),
			Name: child.Name,
		}
		*apiChildren = append(*apiChildren, apiChild)
	}

	return apiChildren
}

func mapQuotaStatusesToResponse(projectChildren []models.Project) []api.QuotaStatusNodePool {
	if len(projectChildren) == 0 {
		return []api.QuotaStatusNodePool{}
	}

	npQuotaStatuses := map[string]api.QuotaStatusNodePool{}
	for _, project := range projectChildren {
		for _, quotaStatus := range project.OrgUnitData.Resources {
			npQuotaStatus, ok := npQuotaStatuses[quotaStatus.NodePoolName]
			if !ok {
				npQuotaStatus = api.QuotaStatusNodePool{
					NodePoolName: quotaStatus.NodePoolName,
					NodePoolId:   quotaStatus.NodePoolId,
					Allocated: api.QuotaStatusResource{
						Gpu:    ptr.Deref(quotaStatus.AllocatedGpu, 0),
						Cpu:    ptr.Deref(quotaStatus.AllocatedCpu, 0),
						Memory: ptr.Deref(quotaStatus.AllocatedMemory, 0),
					},
					AllocatedNonPreemptible: api.QuotaStatusResource{
						Gpu:    ptr.Deref(quotaStatus.AllocatedNonPreemptibleGpu, 0),
						Cpu:    ptr.Deref(quotaStatus.AllocatedNonPreemptibleCpu, 0),
						Memory: ptr.Deref(quotaStatus.AllocatedNonPreemptibleMemory, 0),
					},
					Requested: api.QuotaStatusResource{
						Gpu:    ptr.Deref(quotaStatus.RequestedGpu, 0),
						Cpu:    ptr.Deref(quotaStatus.RequestedCpu, 0),
						Memory: ptr.Deref(quotaStatus.RequestedMemory, 0),
					},
				}
			} else {
				npQuotaStatus.Allocated = api.QuotaStatusResource{
					Gpu:    npQuotaStatus.Allocated.Gpu + ptr.Deref(quotaStatus.AllocatedGpu, 0),
					Cpu:    npQuotaStatus.Allocated.Cpu + ptr.Deref(quotaStatus.AllocatedCpu, 0),
					Memory: npQuotaStatus.Allocated.Memory + ptr.Deref(quotaStatus.AllocatedMemory, 0),
				}
				npQuotaStatus.AllocatedNonPreemptible = api.QuotaStatusResource{
					Gpu:    npQuotaStatus.AllocatedNonPreemptible.Gpu + ptr.Deref(quotaStatus.AllocatedNonPreemptibleGpu, 0),
					Cpu:    npQuotaStatus.AllocatedNonPreemptible.Cpu + ptr.Deref(quotaStatus.AllocatedNonPreemptibleCpu, 0),
					Memory: npQuotaStatus.AllocatedNonPreemptible.Memory + ptr.Deref(quotaStatus.AllocatedNonPreemptibleMemory, 0),
				}
				npQuotaStatus.Requested = api.QuotaStatusResource{
					Gpu:    npQuotaStatus.Requested.Gpu + ptr.Deref(quotaStatus.RequestedGpu, 0),
					Cpu:    npQuotaStatus.Requested.Cpu + ptr.Deref(quotaStatus.RequestedCpu, 0),
					Memory: npQuotaStatus.Requested.Memory + ptr.Deref(quotaStatus.RequestedMemory, 0),
				}
			}

			npQuotaStatuses[quotaStatus.NodePoolName] = npQuotaStatus
		}
	}

	result := make([]api.QuotaStatusNodePool, 0, len(npQuotaStatuses))
	for _, npQuotaStatus := range npQuotaStatuses {
		result = append(result, npQuotaStatus)
	}

	return result
}

func mapProjectAggregatedNodePoolResources(projectChildren []models.Project) *[]api.AggregatedResources {
	if len(projectChildren) == 0 {
		return nil
	}

	npAggregatedResources := map[string]api.AggregatedResources{}
	for _, project := range projectChildren {
		for _, resources := range project.OrgUnitData.Resources {
			npResources, ok := npAggregatedResources[resources.NodePoolId]
			if !ok {
				npResources = api.AggregatedResources{
					NodePool: api.AggregatedResourcesNodePool{
						Id:   resources.NodePoolId,
						Name: resources.NodePoolName,
					},
					Gpu: api.DeservedResource{
						Deserved: deservedWithExplicitUnlimited(resources.GpuDeserved),
					},
					Cpu: api.DeservedResource{
						Deserved: deservedWithExplicitUnlimited(resources.CpuDeserved),
					},
					Memory: api.AggregatedResourcesMemory{
						Deserved: deservedWithExplicitUnlimited(resources.MemoryDeserved),
						Units:    services.ModelUnitsToApi(resources.MemoryUnits),
					},
				}
			} else {
				sumModelToApiResources(&npResources, resources)
			}

			npAggregatedResources[resources.NodePoolId] = npResources
		}
	}

	result := make([]api.AggregatedResources, 0, len(npAggregatedResources))
	for _, npAggregatedResource := range npAggregatedResources {
		result = append(result, npAggregatedResource)
	}
	return &result
}

func sumModelToApiResources(result *api.AggregatedResources, modelResources models.Resources) {
	if modelResources.GpuDeserved != nil {
		result.Gpu.Deserved = sumResourceWithUnlimited(result.Gpu.Deserved, modelResources.GpuDeserved)
	}

	result.Cpu.Deserved = sumResourceWithUnlimited(result.Cpu.Deserved, modelResources.CpuDeserved)
	if modelResources.MemoryDeserved != nil {
		conversionFunc := services.GetMemoryConversionFn(services.ModelUnitsToApi(modelResources.MemoryUnits), result.Memory.Units)
		result.Memory.Deserved = sumResourceWithUnlimited(result.Memory.Deserved, ptr.To(conversionFunc(*modelResources.MemoryDeserved)))
	}
}

// deservedWithExplicitUnlimited - returns the resource value,
// ensuring that unlimited is explicitly represented as ptr to -1.
func deservedWithExplicitUnlimited(value *float64) *float64 {
	if value == nil {
		return ptr.To(-1.0)
	}
	return value
}

// sumResourceWithUnlimited sums two resource values, propagating unlimited (-1).
// - If either value is nil or unlimited (-1), returns -1 (unlimited)
// - Otherwise returns the sum
func sumResourceWithUnlimited(left, right *float64) *float64 {
	// If either is unlimited, the result is unlimited
	if isUnlimited(left) || isUnlimited(right) {
		return ptr.To(-1.0)
	}
	return ptr.To(*left + *right)
}

// isUnlimited checks if a resource value represents unlimited quota
func isUnlimited(value *float64) bool {
	return value == nil || *value < 0
}
