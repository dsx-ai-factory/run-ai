/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

func GetTenantProjectsV1(tenantId int32) ([]models.ProjectV1Grouped, error) {
	projects := []models.ProjectV1{}

	result := sourceClient.
		Where("tenant_id = ?", tenantId).
		Order("project_id, \"project_to_node_pool.default_node_pool_list_index\" ASC").
		Find(&projects)
	if result.Error != nil {
		logging.Errorf("Failed to query tenant projects for tenant %d. err: %v", tenantId, result.Error)
		return nil, result.Error
	}

	projectsGrouped := GroupProjects(projects)
	logging.Infof("Found %d projects for tenant %d.", len(projectsGrouped), tenantId)

	return projectsGrouped, nil
}

// GroupProjects We are fetching project using the view that has a row per project-nodepool, so we would like to group to a single object per project
func GroupProjects(projects []models.ProjectV1) []models.ProjectV1Grouped {
	projectsGrouped := []models.ProjectV1Grouped{}
	curProjectsGrouped := models.ProjectV1Grouped{
		Id:                -1,
		NodepoolResources: []models.AssignedResource{},
	}

	for _, project := range projects {
		if project.Id != curProjectsGrouped.Id {
			if curProjectsGrouped.Id != -1 {
				projectsGrouped = append(projectsGrouped, curProjectsGrouped)
			}

			curProjectsGrouped = models.ProjectV1Grouped{
				Id:                                  project.Id,
				Name:                                project.Name,
				TenantId:                            project.TenantId,
				ClusterUuid:                         project.ClusterUuid,
				CreatedAt:                           project.CreatedAt,
				DepartmentId:                        project.DepartmentId,
				DepartmentName:                      project.DepartmentName,
				Status:                              project.Status,
				Phase:                               project.Phase,
				Namespace:                           project.Namespace,
				CreatedBy:                           project.CreatedBy,
				UpdatedBy:                           project.UpdatedBy,
				InteractiveJobTimeLimitSec:          project.InteractiveJobTimeLimitSec,
				InteractiveJobMaxIdleSec:            project.InteractiveJobMaxIdleSec,
				InteractivePreemptibleJobMaxIdleSec: project.InteractivePreemptibleJobMaxIdleSec,
				TrainingJobMaxIdleSec:               project.TrainingJobMaxIdleSec,
				TrainingJobTimeLimitSec:             project.TrainingJobTimeLimitSec,
				NodeAffinityTrainingIds:             parseNodeTypesArray(project.NodeAffinityTrainingIds),
				NodeAffinityWorkspaceIds:            parseNodeTypesArray(project.NodeAffinityWorkspaceIds),
				NodepoolResources:                   []models.AssignedResource{},
				QuotaStatuses:                       &[]api.QuotaStatusV1Inner{},
				DefaultNodePools:                    &[]string{},
			}
		}

		curProjectsGrouped.NodepoolResources = append(curProjectsGrouped.NodepoolResources, models.AssignedResource{
			CPUDeserved:           project.CPUDeserved,
			CPUMaxAllowed:         project.CPUMaxAllowed,
			CPUOverQuotaWeight:    project.CPUOverQuotaWeight,
			MemoryDeserved:        project.MemoryDeserved,
			MemoryMaxAllowed:      project.MemoryMaxAllowed,
			MemoryOverQuotaWeight: project.MemoryOverQuotaWeight,
			GPUDeserved:           project.GPUDeserved,
			GPUMaxAllowed:         project.GPUMaxAllowed,
			GPUOverQuotaWeight:    project.GPUOverQuotaWeight,
			NodePoolID:            project.NodePoolID,
			NodePoolName:          project.NodePoolName,
		})

		*curProjectsGrouped.QuotaStatuses = append(*curProjectsGrouped.QuotaStatuses, api.QuotaStatusV1Inner{
			NodePoolName: project.NodePoolName,
			Allocated: api.QuotaStatusResourceListV1{
				Gpu:    project.GPUAllocated,
				Cpu:    project.CPUAllocated,
				Memory: project.MemoryAllocated,
			},
			AllocatedNonPreemptible: api.QuotaStatusResourceListV1{
				Gpu:    project.GPUAllocatedNonPreemptible,
				Cpu:    project.CPUAllocatedNonPreemptible,
				Memory: project.MemoryAllocatedNonPreemptible,
			},
			Requested: api.QuotaStatusResourceListV1{
				Gpu:    project.GPURequested,
				Cpu:    project.CPURequested,
				Memory: project.MemoryRequested,
			},
		})

		*curProjectsGrouped.DefaultNodePools = append(*curProjectsGrouped.DefaultNodePools, project.NodePoolName)
	}

	if curProjectsGrouped.Id != -1 {
		projectsGrouped = append(projectsGrouped, curProjectsGrouped)
	}

	return projectsGrouped
}

func parseNodeTypesArray(ids *string) []string {
	parsed := []string{}
	if ids == nil || *ids == "{}" || strings.ToLower(*ids) == "{null}" {
		return parsed
	}

	// Arrays in Postgres are not in json struct, need to replace the parenthesis type.
	runes := []rune(*ids)
	runes[0] = '['
	runes[len(runes)-1] = ']'
	jsonIds := string(runes)

	var intsArray []int32
	err := json.Unmarshal([]byte(jsonIds), &intsArray)
	if err != nil {
		logging.Errorf("Failed to parse ids: %v", err)
		return parsed
	}

	for _, id := range intsArray {
		parsed = append(parsed, fmt.Sprintf("%d", id))
	}

	return parsed
}
