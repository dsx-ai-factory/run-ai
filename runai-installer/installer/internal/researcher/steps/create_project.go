/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-installer/clients/runai/orgunit"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
)

const (
	DefaultProjectName     = "test-project"
	DefaultProjectGPUCount = float32(1000) // Fallback if we can't get actual GPU count
)

type CreateProject struct {
	*step.RunAIInstallStepImpl
	clusterUUID         string
	projectID           string
	gpuDiscoverer       GPUDiscoverer
	clientsProvider     step.RunAIClientsProvider
	runaiClientProvider step.RunAIAPIClientProvider
}
type GPUDiscoverer interface {
	GetClusterGPUCount(ctx context.Context) (float32, error)
}

func NewCreateProject(gpuDiscoverer GPUDiscoverer, clientsProvider step.RunAIClientsProvider, runaiClientProvider step.RunAIAPIClientProvider) *CreateProject {
	s := &CreateProject{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Create project"),
		clientsProvider:      clientsProvider,
		runaiClientProvider:  runaiClientProvider,
	}
	s.gpuDiscoverer = gpuDiscoverer
	s.SetStatusEvaluator(s.EvaluateStatus)
	return s
}

func (s *CreateProject) EvaluateStatus(runaiData *model.RunAIInstallData) {
	if s.ShouldSkipStatusEvaluation() {
		return
	}

	projectName := runaiData.GetInputValue(configinputs.InstallInputReasercherProjectName)
	if projectName == "" {
		projectName = DefaultProjectName
	}

	url, err := step.GetCPUrl(runaiData)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create project] Failed to get Control Plane URL: %v", err)
		s.SetStatus(model.Failed, err)
		return
	}

	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	clients, err := s.clientsProvider(url, caCertPath)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create project] Failed to get clients: %v", err)
		s.SetStatus(model.Failed, err)
		return
	}
	clusterUUID, err := step.GetClusterByNameViaProvider(context.Background(), s.runaiClientProvider, runaiData)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create project] Could not find cluster during status evaluation: %v", err)
		s.SetStatus(model.Failed, err)
		return
	}
	s.clusterUUID = clusterUUID

	project, _, err := clients.OrgUnitServiceClient.GetProjectByName(context.Background(), clusterUUID, projectName)
	if err != nil {
		return
	}

	s.projectID = project.GetId()
	logging.Infof("[RunAI Researcher Setup - Create project] Project '%s' already exists with ID: %s", projectName, s.projectID)
	s.SetStatus(model.Installed, nil)
}

func (s *CreateProject) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}

	s.SetStatus(model.Installing, nil)

	projectName := runaiData.GetInputValue(configinputs.InstallInputReasercherProjectName)
	if projectName == "" {
		projectName = DefaultProjectName
	}

	url, err := step.GetCPUrl(runaiData)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create project] Failed to get Control Plane URL: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	clients, err := s.clientsProvider(url, caCertPath)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create project] Failed to get clients: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	// Look up cluster by name
	clusterUUID, err := step.GetClusterByNameViaProvider(ctx, s.runaiClientProvider, runaiData)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create project] Failed to find cluster: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}
	s.clusterUUID = clusterUUID

	logging.Infof("[RunAI Researcher Setup - Create project] Creating project '%s' with all GPUs in cluster %s", projectName, clusterUUID)

	departmentID, err := clients.OrgUnitServiceClient.GetDefaultDepartmentID(ctx, clusterUUID)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create project] %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	gpuCount := DefaultProjectGPUCount
	if s.gpuDiscoverer != nil {
		discoveredGPUs, err := s.gpuDiscoverer.GetClusterGPUCount(ctx)
		if err != nil {
			logging.Warnf("[RunAI Researcher Setup - Create project] Failed to get cluster GPU count, using default (%v): %v", DefaultProjectGPUCount, err)
		} else if discoveredGPUs > 0 {
			gpuCount = discoveredGPUs
		} else {
			logging.Infof("[RunAI Researcher Setup - Create project] No GPUs found in cluster, using default GPU count: %v", DefaultProjectGPUCount)
		}
	}

	logging.Infof("[RunAI Researcher Setup - Create project] Creating project '%s' with %v GPUs", projectName, gpuCount)

	// Create resources with GPU allocation
	gpuCountFloat64 := float64(gpuCount)
	resources := []orgunit.ResourcesNullable{
		{
			Gpu: &orgunit.ResourcesNullableGpu{
				Deserved: &gpuCountFloat64,
			},
		},
	}

	projectRequest := orgunit.NewProjectCreationRequest(projectName, clusterUUID, resources)
	if departmentID != "" {
		projectRequest.SetParentId(departmentID)
	}

	project, statusCode, err := clients.OrgUnitServiceClient.CreateProject(ctx, clusterUUID, *projectRequest)
	if err != nil {
		// Check if project already exists (409 Conflict)
		if statusCode == http.StatusConflict {
			logging.Infof("[RunAI Researcher Setup - Create project] Project '%s' already exists, continuing", projectName)
			// Try to get existing project ID
			existingProject, _, getErr := clients.OrgUnitServiceClient.GetProjectByName(ctx, clusterUUID, projectName)
			if getErr != nil {
				err := fmt.Errorf("[RunAI Researcher Setup - Create project] project exists but failed to get its ID: %w", getErr)
				s.SetStatus(model.Failed, err)
				return err
			}
			s.projectID = existingProject.GetId()
			s.SetStatus(model.Installed, nil)
			return nil
		}
		logging.Errorf("[RunAI Researcher Setup - Create project] %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	s.projectID = project.GetId()
	logging.Infof("[RunAI Researcher Setup - Create project] Successfully created project '%s' with ID: %s", projectName, s.projectID)
	s.SetStatus(model.Installed, nil)
	return nil
}
