/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"fmt"
	"time"

	service_config "github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/google/uuid"
	"k8s.io/utils/ptr"
)

const (
	TestTenantsCount               = 8
	TestDepartmentsCountPerTenant  = 50
	TestProjectsCountPerDepartment = 3
	TestNodepoolsCountPerTenant    = 3
	TestNodeTypesCountPerTenant    = 100
)

var MapTenantIdToClusterId = map[int32]string{}

func SetupTables() {
	conf := config.Get[*service_config.ConfigData]()
	if _, err := targetClient.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", conf.SchemaSource)); err != nil {
		logging.Errorf("Failed to create schema. err: %v", err)
		return
	}

	if _, err := targetClient.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", models.TenantsSchema)); err != nil {
		logging.Errorf("Failed to create schema. err: %v", err)
		return
	}

	if _, err := targetClient.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", models.ClustersSchema)); err != nil {
		logging.Errorf("Failed to create schema. err: %v", err)
		return
	}

	err := globalClient.Actual().Migrator().CreateTable(models.Tenant{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}

	err = sourceClient.Actual().Migrator().CreateTable(models.MigrationControl{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}

	err = sourceClient.Actual().Migrator().CreateTable(models.TenantSetting{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}

	err = globalClient.Actual().Migrator().CreateTable(models.Cluster{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}

	err = sourceClient.Actual().Migrator().CreateTable(models.DepartmentV1ForTest{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}
	if _, err = sourceClient.Exec("alter table departments_view drop constraint if exists departments_view_pkey"); err != nil {
		logging.Errorf("Failed to drop pkey. err: %v", err)
		return
	}

	err = sourceClient.Actual().Migrator().CreateTable(models.NodeType{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}

	err = sourceClient.Actual().Migrator().CreateTable(models.ProjectV1ForTest{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}
	if _, err = sourceClient.Exec("alter table projects_new_view drop constraint if exists projects_new_view_pkey"); err != nil {
		logging.Errorf("Failed to drop pkey. err: %v", err)
		return
	}

	departmentId := 1
	projectId := 1

	for i := 0; i < TestTenantsCount; i++ {
		tenantId := int32(i + 1)
		clusterId := uuid.NewString()
		MapTenantIdToClusterId[tenantId] = clusterId

		globalClient.Create(&models.Tenant{
			Id:   tenantId,
			Name: fmt.Sprintf("test-%d", i),
		})

		if (i/2)%2 == 0 {
			sourceClient.Create(&models.TenantSetting{
				TenantId: tenantId,
				Key:      ProjectManualOverQuota,
				Value:    "TRUE",
			})

			globalClient.Create(&models.Cluster{
				Id:       clusterId,
				TenantId: tenantId,
				Version:  "2.9.3",
			})
		} else {
			sourceClient.Create(&models.TenantSetting{
				TenantId: tenantId,
				Key:      ProjectManualOverQuota,
				Value:    "false",
			})

			globalClient.Create(&models.Cluster{
				Id:       clusterId,
				TenantId: tenantId,
				Version:  "2.11.7",
			})
		}

		if (i/4)%2 == 0 {
			sourceClient.Create(&models.TenantSetting{
				TenantId: tenantId,
				Key:      CPUResourcesQuota,
				Value:    "true",
			})
		}

		nodeTypes := []*models.NodeType{}
		for j := 0; j < TestNodeTypesCountPerTenant; j++ {
			nodeTypeId := j + int(TestNodeTypesCountPerTenant*tenantId)
			nodeTypes = append(nodeTypes, &models.NodeType{
				ID:          nodeTypeId,
				Name:        fmt.Sprintf("node-type-%d", nodeTypeId),
				TenantID:    uint(tenantId),
				ClusterUUID: uuid.MustParse(clusterId),
			})
		}
		sourceClient.Create(&nodeTypes)

		departments := []*models.DepartmentV1ForTest{}
		projects := []*models.ProjectV1ForTest{}
		for ; departmentId <= int(TestDepartmentsCountPerTenant*tenantId); departmentId++ {
			for ; projectId <= TestProjectsCountPerDepartment*departmentId; projectId++ {
				for nodepoolId := 1; nodepoolId <= TestNodepoolsCountPerTenant; nodepoolId++ {
					projects = append(projects, &models.ProjectV1ForTest{
						Id:                                  projectId,
						Name:                                fmt.Sprintf("proj-%d", projectId),
						TenantId:                            uint(tenantId),
						ClusterUuid:                         uuid.MustParse(clusterId),
						CreatedAt:                           time.Now(),
						DepartmentId:                        int64(departmentId),
						Phase:                               ptr.To[string]("Ready"),
						Namespace:                           fmt.Sprintf("ruani-proj-%d", projectId),
						CreatedBy:                           "test@example.com",
						UpdatedBy:                           "test-2@example.com",
						InteractiveJobTimeLimitSec:          1,
						InteractivePreemptibleJobMaxIdleSec: 3,
						TrainingJobMaxIdleSec:               4,
						TrainingJobTimeLimitSec:             5,
						NodeAffinityTrainingIds:             ptr.To[string](fmt.Sprintf("{%d,%d,%d}", tenantId*100+1, tenantId*100+2, tenantId*100+3)),
						NodeAffinityWorkspaceIds:            ptr.To[string]("{NULL}"),
						CPUDeserved:                         ptr.To[float32](1.5),
						CPUMaxAllowed:                       ptr.To[float32](1.5),
						CPUOverQuotaWeight:                  ptr.To[float32](1.5),
						MemoryDeserved:                      ptr.To[float32](1.5),
						MemoryMaxAllowed:                    ptr.To[float32](1.5),
						MemoryOverQuotaWeight:               ptr.To[float32](1.5),
						GPUDeserved:                         ptr.To[float32](1.5),
						GPUMaxAllowed:                       ptr.To[float32](1.5),
						GPUOverQuotaWeight:                  ptr.To[float32](1.5),
						NodePoolID:                          uint(nodepoolId),
						NodePoolName:                        fmt.Sprintf("np-%d", nodepoolId),
						CPUAllocated:                        0.5,
						CPUAllocatedNonPreemptible:          0.5,
						CPURequested:                        1,
						GPUAllocated:                        0.5,
						GPUAllocatedNonPreemptible:          0.5,
						GPURequested:                        1,
						MemoryAllocated:                     0.5,
						MemoryAllocatedNonPreemptible:       0.5,
						MemoryRequested:                     1,
						DefaultNodePoolIndex:                TestNodepoolsCountPerTenant - int32(nodepoolId),
					})
				}
			}

			for nodepoolId := 1; nodepoolId <= TestNodepoolsCountPerTenant; nodepoolId++ {
				departments = append(departments, &models.DepartmentV1ForTest{
					ID:                    departmentId,
					Name:                  fmt.Sprintf("dep-%d", departmentId),
					TenantID:              uint(tenantId),
					ClusterUUID:           clusterId,
					CreatedAt:             time.Now(),
					CPUDeserved:           ptr.To[float32](0.5),
					CPUMaxAllowed:         ptr.To[float32](0.5),
					CPUOverQuotaWeight:    ptr.To[float32](0.5),
					MemoryDeserved:        ptr.To[float32](0.5),
					MemoryMaxAllowed:      ptr.To[float32](0.5),
					MemoryOverQuotaWeight: ptr.To[float32](0.5),
					GPUDeserved:           ptr.To[float32](0.5),
					GPUMaxAllowed:         ptr.To[float32](0.5),
					GPUOverQuotaWeight:    ptr.To[float32](0.5),
					NodePoolID:            uint(nodepoolId),
					NodePoolName:          fmt.Sprintf("np-%d", nodepoolId),
				})
			}
		}
		sourceClient.Create(&departments)
		sourceClient.Create(&projects)
	}
}

func CleanTables() {
}
