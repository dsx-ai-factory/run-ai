/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/utils/ptr"

	logging "github.com/run-ai/runai/runai-common-packages/logging"

	clusteropenapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

	openapi "github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	orgUnitsServiceDao "github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	orgUnitServiceModels "github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

func CreateDepartments(tenant models.Tenant, departments []models.DepartmentV1Grouped, settings map[string]bool) error {
	departmentsToCreate := []*orgUnitServiceModels.Department{}
	nodePools := createStubNodepools(departments, settings)
	for _, departmentV1 := range departments {
		logging.Debugf("Migrating department %s of tenant %s.", departmentV1.Name, tenant.Name)
		departmentV2, err := MapDepartmentV1ToV2(departmentV1, nodePools)
		if err != nil {
			logging.Errorf("Failed to convert department %s of tenant %s to V2 struct. Err: %v", departmentV1.Name, tenant.Name, err)
			return err
		}
		departmentsToCreate = append(departmentsToCreate, departmentV2)
	}

	err := orgUnitsServiceDao.CreateDepartments(context.Background(), departmentsToCreate)
	if err != nil {
		logging.Errorf("Failed to create departments of tenant %s. Err: %v", tenant.Name, err)
		return err
	}

	return nil
}

func MapDepartmentV1ToV2(departmentV1 models.DepartmentV1Grouped, nodePools map[string]clusteropenapi.NodepoolMinimal) (*orgUnitServiceModels.Department, error) {
	departmentV2Request := openapi.DepartmentCreationRequest{
		Resources: convutils.ConvertSliceTypes(departmentV1.NodepoolResources, getDepartmentNodepoolResourcesToResourcesV2()),
		Name:      departmentV1.Name,
		ClusterId: departmentV1.ClusterUUID,
	}
	departmentV2, err := departments.MapCreateRequestToModel(context.Background(), departmentV2Request, nodePools)
	if err != nil {
		return nil, err
	}

	departmentV2.ID = int64(departmentV1.ID)
	departmentV2.TenantId = fmt.Sprintf("%d", departmentV1.TenantID)
	departmentV2.CreatedAt = departmentV1.CreatedAt
	departmentV2.CreatedBy = ""
	departmentV2.UpdatedBy = ""

	return &departmentV2, nil
}

func createStubNodepools(departments []models.DepartmentV1Grouped, settings map[string]bool) map[string]clusteropenapi.NodepoolMinimal {
	result := make(map[string]clusteropenapi.NodepoolMinimal)
	for _, department := range departments {
		for _, assignedResources := range department.NodepoolResources {
			nodePoolId := strconv.Itoa(int(assignedResources.NodePoolID))
			if _, exists := result[nodePoolId]; !exists {
				result[nodePoolId] = clusteropenapi.NodepoolMinimal{
					Id: strconv.Itoa(int(assignedResources.NodePoolID)),
					SchedulingCapabilities: &clusteropenapi.SchedulingCapabilities{
						CpuQuota:          settings[database.CPUResourcesQuota],
						AdjustableWeights: settings[database.ProjectManualOverQuota],
					},
				}
			}
		}
	}
	return result
}

func getDepartmentNodepoolResourcesToResourcesV2() func(resource models.AssignedResource) openapi.ResourcesNullable {
	return func(assignedResources models.AssignedResource) openapi.ResourcesNullable {
		return openapi.ResourcesNullable{
			NodePool: &openapi.ResourcesNodePool{
				Id:   fmt.Sprintf("%d", assignedResources.NodePoolID),
				Name: assignedResources.NodePoolName,
			},
			Gpu: openapi.ResourcesNullableGpu{
				Limit:           convertDepartmentMaxAllowedToLimit(assignedResources.GPUMaxAllowed),
				Deserved:        convertDepartmentDeservedToV2NotNull(assignedResources.GPUDeserved),
				OverQuotaWeight: convertDepartmentDeservedToV2(assignedResources.GPUDeserved),
			},
			Cpu: &openapi.ResourcesCpu{
				Limit:           convertDepartmentMaxAllowedToLimit(assignedResources.CPUMaxAllowed),
				Deserved:        convertDepartmentDeservedToV2(assignedResources.CPUDeserved),
				OverQuotaWeight: convertDepartmentDeservedToV2(assignedResources.CPUDeserved),
			},
			Memory: &openapi.ResourcesMemory{
				Limit:           convertDepartmentMaxAllowedToLimit(assignedResources.MemoryMaxAllowed),
				Deserved:        convertDepartmentDeservedToV2(assignedResources.MemoryDeserved),
				OverQuotaWeight: convertDepartmentDeservedToV2(assignedResources.MemoryDeserved),
				Units:           "Mib",
			},
		}
	}
}

func convertDepartmentMaxAllowedToLimit(maxAllowed *float32) *float64 {
	if maxAllowed == nil {
		return nil
	}

	return ptr.To(float64(*maxAllowed))
}

func convertDepartmentDeservedToV2(deserved *float32) *float64 {
	if deserved == nil {
		return nil
	}

	return ptr.To(float64(*deserved))
}

func convertDepartmentDeservedToV2NotNull(deserved *float32) float64 {
	deservedNullable := convertDepartmentDeservedToV2(deserved)
	if deservedNullable == nil {
		return -1
	}
	return *deservedNullable
}
