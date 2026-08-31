/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
)

type CreateAccessRule struct {
	*step.RunAIInstallStepImpl
	createUserStep    *CreateLocalUser
	getRoleStep       *GetL1ResearcherRoleID
	createProjectStep *CreateProject
	clientsProvider   step.RunAIClientsProvider
}

func NewCreateAccessRule(
	createUserStep *CreateLocalUser,
	getRoleStep *GetL1ResearcherRoleID,
	createProjectStep *CreateProject,
	clientsProvider step.RunAIClientsProvider,
) *CreateAccessRule {
	s := &CreateAccessRule{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Create access rule"),
		createUserStep:       createUserStep,
		getRoleStep:          getRoleStep,
		createProjectStep:    createProjectStep,
		clientsProvider:      clientsProvider,
	}
	s.SetStatusEvaluator(s.EvaluateStatus)
	return s
}

func (s *CreateAccessRule) EvaluateStatus(runaiData *model.RunAIInstallData) {
	if s.ShouldSkipStatusEvaluation() {
		return
	}

	// Check if access rule already exists
	userEmail := runaiData.GetInputValue(configinputs.InstallInputResearcherUserEmail)
	if userEmail == "" {
		msg := "researcher user email is empty"
		err := errors.New(msg)
		logging.Errorf("[RunAI Researcher Setup - Create access rule] %s", msg)
		s.SetStatus(model.Failed, err)
		return
	}

	url, err := step.GetCPUrl(runaiData)
	if err != nil {
		return
	}

	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	clientsInstance, err := s.clientsProvider(url, caCertPath)
	if err != nil {
		logging.Errorf("[RunAI] Failed to get Control Plane URL: %v", err)
		s.SetStatus(model.Failed, err)
		return
	}

	exists, err := clientsInstance.AuthServiceClient.AccessRuleExists(context.Background(), userEmail, s.getRoleStep.RoleID, s.createProjectStep.clusterUUID, s.createProjectStep.projectID)
	if err != nil {
		logging.Infof("[RunAI Researcher Setup - Create access rule] Could not check access rule during status evaluation: %v", err)
		return
	}

	if exists {
		logging.Infof("[RunAI Researcher Setup - Create access rule] Access rule for user %s on project %s already exists", userEmail, s.createProjectStep.projectID)
		s.SetStatus(model.Installed, nil)
	}
}

func (s *CreateAccessRule) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}

	s.SetStatus(model.Installing, nil)

	userEmail := runaiData.GetInputValue(configinputs.InstallInputResearcherUserEmail)
	if userEmail == "" {
		err := fmt.Errorf("user email is required but not provided")
		logging.Errorf("[RunAI Researcher Setup - Create access rule] %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	url, err := step.GetCPUrl(runaiData)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create access rule] Failed to get Control Plane URL: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	clients, err := s.clientsProvider(url, caCertPath)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create access rule] Failed to get clients: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	statusCode, err := clients.AuthServiceClient.CreateAccessRule(ctx, userEmail, s.getRoleStep.RoleID, s.createProjectStep.clusterUUID, s.createProjectStep.projectID)
	if err != nil {
		if statusCode == http.StatusConflict {
			logging.Infof("[RunAI Researcher Setup - Create access rule] Access rule for user %s on project %s already exists, continuing", userEmail, s.createProjectStep.projectID)
			s.SetStatus(model.Installed, nil)
			return nil
		}
		logging.Errorf("[RunAI Researcher Setup - Create access rule] %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	logging.Infof("[RunAI Researcher Setup - Create access rule] Successfully created access rule for user %s on project %s", userEmail, s.createProjectStep.projectID)
	return nil
}
