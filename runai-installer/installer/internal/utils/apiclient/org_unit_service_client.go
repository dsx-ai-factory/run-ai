/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apiclient

import (
	"context"
	"fmt"

	orgunit "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	runaiclient "github.com/dsx-ai-factory/run-ai/runai-installer/clients/pkg/clients"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
)

type OrgUnitServiceClient struct {
	client *runaiclient.RunAIClient
}

func NewOrgUnitServiceClientFromRunAI(client *runaiclient.RunAIClient) *OrgUnitServiceClient {
	return &OrgUnitServiceClient{
		client: client,
	}
}

func (c *OrgUnitServiceClient) GetProjectByName(ctx context.Context, clusterUUID, projectName string) (*orgunit.Project, int, error) {
	logging.Infof("[ProjectsService] Fetching projects for cluster %s", clusterUUID)
	projects, resp, err := c.client.ProjectsAPI().GetProjects(ctx).Limit(1).Execute()
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		logging.Errorf("[ProjectsService] Failed to get projects: %v", err)
		return nil, statusCode, fmt.Errorf("failed to get projects: %w", err)
	}
	if len(projects.Projects) == 0 {
		return nil, resp.StatusCode, fmt.Errorf("project %s not found", projectName)
	}

	logging.Infof("[ProjectsService] Successfully fetched %d projects", len(projects.Projects))
	return &projects.Projects[0], resp.StatusCode, nil
}

func (c *OrgUnitServiceClient) CreateProject(ctx context.Context, clusterUUID string, req orgunit.ProjectCreationRequest) (*orgunit.Project, int, error) {
	logging.Infof("[ProjectsService] Creating project '%s' in cluster %s", req.Name, clusterUUID)

	project, resp, err := c.client.ProjectsAPI().CreateProject(ctx).ProjectCreationRequest(req).Execute()
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		logging.Errorf("[ProjectsService] Failed to create project: %v", err)
		return nil, statusCode, fmt.Errorf("failed to create project: %w", err)
	}

	logging.Infof("[ProjectsService] Successfully created project '%s' (ID: %s)", project.GetName(), project.GetId())
	return project, resp.StatusCode, nil
}

func (c *OrgUnitServiceClient) GetDepartments(ctx context.Context, clusterUUID string) ([]orgunit.DepartmentForList, int, error) {
	logging.Infof("[ProjectsService] Fetching departments for cluster %s", clusterUUID)

	filterBy := []string{fmt.Sprintf("clusterId==%s", clusterUUID)}
	response, resp, err := c.client.DepartmentsAPI().GetDepartments(ctx).FilterBy(filterBy).Execute()
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		logging.Errorf("[ProjectsService] Failed to get departments: %v", err)
		return nil, statusCode, fmt.Errorf("failed to get departments: %w", err)
	}

	logging.Infof("[ProjectsService] Successfully fetched %d departments", len(response.Departments))
	return response.Departments, resp.StatusCode, nil
}

func (c *OrgUnitServiceClient) GetDefaultDepartmentID(ctx context.Context, clusterUUID string) (string, error) {
	logging.Infof("[ProjectsService] Getting default department ID for cluster: %s", clusterUUID)

	departments, _, err := c.GetDepartments(ctx, clusterUUID)
	if err != nil {
		return "", err
	}

	if len(departments) == 0 {
		return "", fmt.Errorf("no departments found for cluster %s", clusterUUID)
	}

	departmentID := departments[0].GetId()
	logging.Infof("[ProjectsService] Found default department with ID: %s", departmentID)
	return departmentID, nil
}
