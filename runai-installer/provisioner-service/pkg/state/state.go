package state

import (
	"sync"
	"time"

	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/types"
)

type store struct {
	mu           sync.RWMutex
	environments map[types.EnvironmentID]types.EnvironmentStatus
}

var s = &store{
	environments: map[types.EnvironmentID]types.EnvironmentStatus{},
}

func CreateEnvironmentStatus(id types.EnvironmentID, mode types.ProvisioningMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.environments[id] = types.EnvironmentStatus{
		ID:        id,
		Phase:     types.EnvironmentPhasePending,
		Message:   "Queued",
		UpdatedAt: time.Now().UTC(),
		Mode:      mode,
	}
}

func SetInProgress(id types.EnvironmentID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	env, ok := s.environments[id]
	if !ok {
		return
	}
	env.Phase = types.EnvironmentPhaseInProgress
	env.Message = "Provisioning started"
	env.UpdatedAt = time.Now().UTC()
	s.environments[id] = env
}

func UpdateMessage(id types.EnvironmentID, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if env, ok := s.environments[id]; ok {
		env.Message = msg
		env.UpdatedAt = time.Now().UTC()
		s.environments[id] = env
	}
}

func SetCreatingRunaiAssets(id types.EnvironmentID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	env, ok := s.environments[id]
	if !ok {
		return
	}
	env.Phase = types.EnvironmentPhaseCreatingRunaiAssets
	env.Message = "Creating Run:AI assets (user, project, and access rule)"
	env.UpdatedAt = time.Now().UTC()
	s.environments[id] = env
}

func SetUserCredentials(id types.EnvironmentID, email, password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	env, ok := s.environments[id]
	if !ok {
		return
	}
	env.UserCredentials = &types.RunaiUserCredentials{
		Email:    email,
		Password: password,
	}
	env.UpdatedAt = time.Now().UTC()
	s.environments[id] = env
}

func SetSucceeded(id types.EnvironmentID, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	env, ok := s.environments[id]
	if !ok {
		return
	}
	env.Phase = types.EnvironmentPhaseSucceeded
	env.Message = msg
	env.UpdatedAt = time.Now().UTC()
	s.environments[id] = env
}

func SetFailed(id types.EnvironmentID, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	env, ok := s.environments[id]
	if !ok {
		return
	}
	env.Phase = types.EnvironmentPhaseFailed
	env.Message = msg
	env.UpdatedAt = time.Now().UTC()
	s.environments[id] = env
}

func Get(id types.EnvironmentID) (types.EnvironmentStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	env, ok := s.environments[id]
	return env, ok
}

func GetInProgressEnvironment() (types.EnvironmentStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, env := range s.environments {
		if env.Phase == types.EnvironmentPhaseInProgress || env.Phase == types.EnvironmentPhasePending {
			return env, true
		}
	}
	return types.EnvironmentStatus{}, false
}

func Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.environments = map[types.EnvironmentID]types.EnvironmentStatus{}
}
