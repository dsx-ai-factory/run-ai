package steps

import (
	"context"
	"fmt"

	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/apiclient"
)

type CreateLocalUser struct {
	*step.RunAIInstallStepImpl
	userCredentials *apiclient.UserCredentials
	clientsProvider step.RunAIClientsProvider
}

func NewCreateLocalUser(clientsProvider step.RunAIClientsProvider) *CreateLocalUser {
	c := &CreateLocalUser{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Create local user"),
		clientsProvider:      clientsProvider,
	}
	c.SetStatusEvaluator(c.EvaluateStatus)
	return c
}

func (s *CreateLocalUser) EvaluateStatus(runaiData *model.RunAIInstallData) {
	if s.ShouldSkipStatusEvaluation() {
		return
	}

	// Check if user with the email already exists
	userEmail := runaiData.GetInputValue(configinputs.InstallInputResearcherUserEmail)
	if userEmail == "" {
		err := fmt.Errorf("user email is required but not provided")
		logging.Errorf("[RunAI Researcher Setup - Create local user] %v", err)
		s.SetStatus(model.Failed, err)
		return
	}

	url, err := step.GetCPUrl(runaiData)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create local user] Failed to get Control Plane URL: %v", err)
		s.SetStatus(model.Failed, err)
		return
	}

	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	clientsInstance, err := s.clientsProvider(url, caCertPath)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create local user] Failed to get clients: %v", err)
		s.SetStatus(model.Failed, err)
		return
	}

	user, err := clientsInstance.AuthServiceClient.GetUserByEmail(context.Background(), userEmail)
	if err != nil {
		return
	}

	if user != nil {
		s.userCredentials = &apiclient.UserCredentials{
			Email:  user.Username,
			UserID: user.Id,
		}
		logging.Infof("[RunAI Assets] User '%s' already exists with ID: %s", userEmail, user.Id)
		s.SetStatus(model.Installed, nil)
	}
}

func (s *CreateLocalUser) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}

	s.SetStatus(model.Installing, nil)

	userEmail := runaiData.GetInputValue(configinputs.InstallInputResearcherUserEmail)
	if userEmail == "" {
		err := fmt.Errorf("user email is required but not provided")
		logging.Errorf("[RunAI Researcher Setup - Create local user] %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	url, err := step.GetCPUrl(runaiData)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create local user] Failed to get Control Plane URL: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	clients, err := s.clientsProvider(url, caCertPath)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create local user] Failed to get clients: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	// Use IDM service to create local user
	userCreds, _, err := clients.AuthServiceClient.CreateLocalUser(ctx, userEmail)
	if err != nil {
		logging.Errorf("[RunAI Researcher Setup - Create local user] Failed to create user: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	s.userCredentials = userCreds
	s.SetStatus(model.Installed, nil)
	return nil
}
