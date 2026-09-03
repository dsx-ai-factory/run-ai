/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"

	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

const (
	ResearcherRoleName = "AI practitioner"
)

type GetL1ResearcherRoleID struct {
	*step.RunAIInstallStepImpl
	RoleID          int32
	clientsProvider step.RunAIClientsProvider
}

func NewGetL1ResearcherRoleID(clientsProvider step.RunAIClientsProvider) *GetL1ResearcherRoleID {
	s := &GetL1ResearcherRoleID{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(fmt.Sprintf("Get %s role ID", ResearcherRoleName)),
		clientsProvider:      clientsProvider,
	}
	return s
}

func (s *GetL1ResearcherRoleID) GetStatusAsync(runaiData *model.RunAIInstallData) model.InstallPhase {
	return s.GetCachedStatus()
}

func (s *GetL1ResearcherRoleID) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}

	s.SetStatus(model.Installing, nil)

	url, err := step.GetCPUrl(runaiData)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Get %s ID] Failed to get Control Plane URL: %v", ResearcherRoleName, err)
		s.SetStatus(model.Failed, err)
		return err
	}

	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	clients, err := s.clientsProvider(url, caCertPath)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Get %s ID] Failed to get clients: %v", ResearcherRoleName, err)
		s.SetStatus(model.Failed, err)
		return err
	}

	// Use OrgUnitServiceClient to get role ID
	roleID, err := clients.AuthServiceClient.GetRoleIDByName(ctx, ResearcherRoleName)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Get %s role ID] %v", ResearcherRoleName, err)
		s.SetStatus(model.Failed, err)
		return err
	}

	s.RoleID = roleID
	s.SetStatus(model.Installed, nil)
	return nil
}
