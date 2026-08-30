package models

import (
	"github.com/google/uuid"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"
)

type DepartmentNodePoolResources struct {
	dao_utils.DeleteBaseModel
	TenantId     string
	ClusterId    uuid.UUID
	OrgUnitId    string
	OrgUnitName  string
	NodePoolId   string
	NodePoolName string
	Rank         string
	ResourcesPerNodePool
}

func (DepartmentNodePoolResources) FilterByFieldsConfigurationForSum() validators.FieldsValidation {
	return map[string]validators.FieldValidator{
		"orgUnitName": validators.NewAlwaysValidFieldValidation("org_unit_name"),
		"orgUnitId":   validators.NewAlwaysValidFieldValidation("org_unit_id"),
		"rank":        validators.NewAlwaysValidFieldValidation("rank"),
		"clusterId":   validators.NewUUIDFieldValidation("cluster_id"),
	}
}

func (DepartmentNodePoolResources) FilterByFieldsConfiguration() validators.FieldsValidation {
	return map[string]validators.FieldValidator{
		"orgUnitName":  validators.NewAlwaysValidFieldValidation("org_unit_name"),
		"orgUnitId":    validators.NewAlwaysValidFieldValidation("org_unit_id"),
		"nodePoolName": validators.NewAlwaysValidFieldValidation(ResourcesNodePoolNameColumn),
		"nodePoolId":   validators.NewAlwaysValidFieldValidation(ResourcesNodePoolIdColumn),
		"rank":         validators.NewAlwaysValidFieldValidation("rank"),
		"clusterId":    validators.NewUUIDFieldValidation("cluster_id"),
	}
}

func (DepartmentNodePoolResources) SortByFieldsConfiguration() validators.FieldsValidation {
	filterByValidation := DepartmentNodePoolResources{}.FilterByFieldsConfiguration()
	result := make(validators.FieldsValidation)
	for k, v := range filterByValidation {
		result[k] = v
	}

	result["gpuDeserved"] = validators.NewAlwaysValidFieldValidation("gpu_deserved")
	result["gpuOverQuotaWeight"] = validators.NewAlwaysValidFieldValidation("gpu_over_quota_weight")
	return result
}

func (p *DepartmentNodePoolResources) ToResources() Resources {
	return Resources{
		NodePoolId:            p.NodePoolId,
		NodePoolName:          p.NodePoolName,
		Rank:                  &p.Rank,
		GpuDeserved:           p.GpuDeserved,
		GpuLimit:              p.GpuLimit,
		GpuOverQuotaWeight:    p.GpuOverQuotaWeight,
		CpuDeserved:           p.CpuDeserved,
		CpuLimit:              p.CpuLimit,
		CpuOverQuotaWeight:    p.CpuOverQuotaWeight,
		MemoryDeserved:        p.MemoryDeserved,
		MemoryLimit:           p.MemoryLimit,
		MemoryOverQuotaWeight: p.MemoryOverQuotaWeight,
		MemoryUnits:           p.MemoryUnits,
	}
}
