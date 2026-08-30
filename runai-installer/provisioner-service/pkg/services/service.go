package services

import (
	"context"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/state"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/types"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/workflow"
)

type ProvisionerService struct {
	provisionFunc func(envID types.EnvironmentID, mode types.ProvisioningMode, domain string, wildcardDomain string, kubernetesDistribution *string, userEmail *string, projectName *string)
}

func NewProvisionerService() *ProvisionerService {
	s := &ProvisionerService{}
	s.provisionFunc = s.provisionEnvironment
	return s
}

func (s *ProvisionerService) CreateEnvironment(mode types.ProvisioningMode, domain string, wildcardDomain string, kubernetesDistribution *string, userEmail *string, projectName *string) (types.EnvironmentID, time.Time, error) {
	m := types.ProvisioningMode(strings.ToLower(strings.TrimSpace(string(mode))))
	logging.Infof("Creating environment with mode: %s", m)
	if existingEnv, createdAt, found := s.checkForInProgressEnvironment(); found {
		logging.Infof("Environment creation already in progress (ID: %s), returning existing environment instead of creating new one", existingEnv)
		return existingEnv, createdAt, nil
	}

	ulidID := ulid.Make()
	id := types.EnvironmentID(ulidID.String())
	createdAt := ulid.Time(ulidID.Time())
	state.CreateEnvironmentStatus(id, m)

	go s.provisionFunc(id, m, domain, wildcardDomain, kubernetesDistribution, userEmail, projectName)

	return id, createdAt, nil
}

func (s *ProvisionerService) provisionEnvironment(envID types.EnvironmentID, mode types.ProvisioningMode, domain string, wildcardDomain string, kubernetesDistribution *string, userEmail *string, projectName *string) {
	state.SetInProgress(envID)
	ctx := context.Background()
	if err := workflow.InstallRunAI(ctx, envID, mode, domain, wildcardDomain, kubernetesDistribution, userEmail, projectName); err != nil {
		state.SetFailed(envID, err.Error())
		return
	}
}

func (s *ProvisionerService) checkForInProgressEnvironment() (types.EnvironmentID, time.Time, bool) {
	if env, found := state.GetInProgressEnvironment(); found {
		logging.Infof("Environment creation already in progress (ID: %s), returning existing environment instead of creating new one", env.ID)
		return env.ID, env.UpdatedAt, true
	}
	return "", time.Time{}, false
}

func (s *ProvisionerService) GetEnvironment(id string) (types.EnvironmentStatus, bool) {
	return state.Get(types.EnvironmentID(id))
}
