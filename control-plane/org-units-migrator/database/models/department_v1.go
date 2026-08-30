package models

import (
	"time"
)

type DepartmentV1 struct {
	ID                    int       `gorm:"column:id;primaryKey"`
	Name                  string    `gorm:"column:name"`
	TenantID              uint      `gorm:"column:tenant_id"`
	ClusterUUID           string    `gorm:"column:cluster_uuid"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	CPUDeserved           *float32  `gorm:"column:node_pool_resources.cpu_deserved"`
	CPUMaxAllowed         *float32  `gorm:"column:node_pool_resources.cpu_max_allowed"`
	CPUOverQuotaWeight    *float32  `gorm:"column:node_pool_resources.cpu_over_quota_weight"`
	MemoryDeserved        *float32  `gorm:"column:node_pool_resources.memory_deserved"`
	MemoryMaxAllowed      *float32  `gorm:"column:node_pool_resources.memory_max_allowed"`
	MemoryOverQuotaWeight *float32  `gorm:"column:node_pool_resources.memory_over_quota_weight"`
	GPUDeserved           *float32  `gorm:"column:node_pool_resources.gpu_deserved"`
	GPUMaxAllowed         *float32  `gorm:"column:node_pool_resources.gpu_max_allowed"`
	GPUOverQuotaWeight    *float32  `gorm:"column:node_pool_resources.gpu_over_quota_weight"`
	NodePoolID            uint      `gorm:"column:node_pool_resources.node_pool_id"`
	NodePoolName          string    `gorm:"column:node_pool_resources.node_pool_name"`
}

func (DepartmentV1) TableName() string {
	return "departments_view"
}

type AssignedResource struct {
	CPUDeserved           *float32 `gorm:"column:cpu_deserved"`
	CPUMaxAllowed         *float32 `gorm:"column:cpu_max_allowed"`
	CPUOverQuotaWeight    *float32 `gorm:"column:cpu_over_quota_weight"`
	MemoryDeserved        *float32 `gorm:"column:memory_deserved"`
	MemoryMaxAllowed      *float32 `gorm:"column:memory_max_allowed"`
	MemoryOverQuotaWeight *float32 `gorm:"column:memory_over_quota_weight"`
	GPUDeserved           *float32 `gorm:"column:gpu_deserved"`
	GPUMaxAllowed         *float32 `gorm:"column:gpu_max_allowed"`
	GPUOverQuotaWeight    *float32 `gorm:"column:gpu_over_quota_weight"`
	NodePoolID            uint     `gorm:"column:node_pool_id"`
	NodePoolName          string   `gorm:"column:node_pool_name"`
}

// DepartmentV1Grouped In departments view there is a row per department-nodepool, so we want to groupd it to a single model per department
type DepartmentV1Grouped struct {
	ID                int                `gorm:"column:id;primaryKey"`
	Name              string             `gorm:"column:name"`
	TenantID          uint               `gorm:"column:tenant_id"`
	ClusterUUID       string             `gorm:"column:cluster_uuid"`
	CreatedAt         time.Time          `gorm:"column:created_at"`
	NodepoolResources []AssignedResource `gorm:"-"`
}

// DepartmentV1ForTest gorm auto migrate does not work well with columns that has a dot in their name.
// We duplicated the struct for test purposes only, it allow us to create and insert data to the table, but fetching the data is using the regular struct.
type DepartmentV1ForTest struct {
	ID                    int       `gorm:"column:id;uniqueIndex:dep_np"`
	Name                  string    `gorm:"column:name"`
	TenantID              uint      `gorm:"column:tenant_id"`
	ClusterUUID           string    `gorm:"column:cluster_uuid"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	CPUDeserved           *float32  `gorm:"column:\"node_pool_resources.cpu_deserved\""`
	CPUMaxAllowed         *float32  `gorm:"column:\"node_pool_resources.cpu_max_allowed\""`
	CPUOverQuotaWeight    *float32  `gorm:"column:\"node_pool_resources.cpu_over_quota_weight\""`
	MemoryDeserved        *float32  `gorm:"column:\"node_pool_resources.memory_deserved\""`
	MemoryMaxAllowed      *float32  `gorm:"column:\"node_pool_resources.memory_max_allowed\""`
	MemoryOverQuotaWeight *float32  `gorm:"column:\"node_pool_resources.memory_over_quota_weight\""`
	GPUDeserved           *float32  `gorm:"column:\"node_pool_resources.gpu_deserved\""`
	GPUMaxAllowed         *float32  `gorm:"column:\"node_pool_resources.gpu_max_allowed\""`
	GPUOverQuotaWeight    *float32  `gorm:"column:\"node_pool_resources.gpu_over_quota_weight\""`
	NodePoolID            uint      `gorm:"column:\"node_pool_resources.node_pool_id\";uniqueIndex:dep_np"`
	NodePoolName          string    `gorm:"column:\"node_pool_resources.node_pool_name\""`
}

func (DepartmentV1ForTest) TableName() string {
	return "departments_view"
}
