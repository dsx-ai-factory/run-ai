package database

import (
	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func GetTenantDepartmentsV1(tenantId int32) ([]models.DepartmentV1Grouped, error) {
	departments := []models.DepartmentV1{}

	result := sourceClient.
		Where("tenant_id = ?", tenantId).
		Order("id").
		Find(&departments)
	if result.Error != nil {
		logging.Errorf("Failed to query tenant departments for tenant %d. err: %v", tenantId, result.Error)
		return nil, result.Error
	}

	departmentsGrouped := GroupDepartments(departments)
	logging.Infof("Found %d departments for tenant %d.", len(departmentsGrouped), tenantId)

	return departmentsGrouped, nil
}

// GroupDepartments We are fetching departments using the view that has a row per departments-nodepool, so we would like to group to a single object per departments
func GroupDepartments(departments []models.DepartmentV1) []models.DepartmentV1Grouped {
	departmentsGrouped := []models.DepartmentV1Grouped{}
	curDepartmentsGrouped := models.DepartmentV1Grouped{
		ID:                -1,
		NodepoolResources: []models.AssignedResource{},
	}

	for _, department := range departments {
		if department.ID != curDepartmentsGrouped.ID {
			if curDepartmentsGrouped.ID != -1 {
				departmentsGrouped = append(departmentsGrouped, curDepartmentsGrouped)
			}

			curDepartmentsGrouped.Name = department.Name
			curDepartmentsGrouped.ID = department.ID
			curDepartmentsGrouped.ClusterUUID = department.ClusterUUID
			curDepartmentsGrouped.TenantID = department.TenantID
			curDepartmentsGrouped.CreatedAt = department.CreatedAt
			curDepartmentsGrouped.NodepoolResources = []models.AssignedResource{}
		}

		curDepartmentsGrouped.NodepoolResources = append(curDepartmentsGrouped.NodepoolResources, models.AssignedResource{
			CPUDeserved:           department.CPUDeserved,
			CPUMaxAllowed:         department.CPUMaxAllowed,
			CPUOverQuotaWeight:    department.CPUOverQuotaWeight,
			MemoryDeserved:        department.MemoryDeserved,
			MemoryMaxAllowed:      department.MemoryMaxAllowed,
			MemoryOverQuotaWeight: department.MemoryOverQuotaWeight,
			GPUDeserved:           department.GPUDeserved,
			GPUMaxAllowed:         department.GPUMaxAllowed,
			GPUOverQuotaWeight:    department.GPUOverQuotaWeight,
			NodePoolID:            department.NodePoolID,
			NodePoolName:          department.NodePoolName,
		})
	}

	if curDepartmentsGrouped.ID != -1 {
		departmentsGrouped = append(departmentsGrouped, curDepartmentsGrouped)
	}

	return departmentsGrouped
}
