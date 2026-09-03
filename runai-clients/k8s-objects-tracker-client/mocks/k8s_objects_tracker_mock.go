package mocks

import (
	"context"
	"net/http"

	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
)

// StorageClassConfigMock supports fixed set of pvcs and possible error
type K8SObjectTrackerMock struct {
	// in reality the key is tenantId + clusterId + name, but we only track the names.
	StorageClassConfig map[string]k8sObjectsTracker.StorageClassConfiguration
	Secrets            []k8sObjectsTracker.SecretListInfo
}

func NewK8SObjectTrackerMock() K8SObjectTrackerMock {
	return K8SObjectTrackerMock{
		StorageClassConfig: make(map[string]k8sObjectsTracker.StorageClassConfiguration),
	}
}

func (s K8SObjectTrackerMock) GetStorageClassConfigInt(ctx context.Context, clusterId string, storageClassName string) (*k8sObjectsTracker.StorageClassConfiguration, *k8sObjectsTracker.Error) {
	result, ok := s.StorageClassConfig[storageClassName]
	if !ok || result.ScopeId != clusterId {
		return nil, k8sObjectsTracker.NewError(http.StatusNotFound, "not found")
	}
	return &result, nil
}

func (s K8SObjectTrackerMock) GetStorageClassesConfigInt(ctx context.Context, clusterId string, storageClassNames []string) (map[string]k8sObjectsTracker.StorageClassConfiguration, *k8sObjectsTracker.Error) {
	configs := make(map[string]k8sObjectsTracker.StorageClassConfiguration)

	for _, name := range storageClassNames {
		result, ok := s.StorageClassConfig[name]
		if ok && result.ScopeId == clusterId {
			configs[name] = result
		}
	}

	return configs, nil
}

func (s K8SObjectTrackerMock) WithStorageClassConfig(name string, config k8sObjectsTracker.StorageClassConfiguration) K8SObjectTrackerMock {
	s.StorageClassConfig[name] = config
	return s
}

func (s K8SObjectTrackerMock) ListSecretsByScope(ctx context.Context, scopeType string, scopeId string) (*k8sObjectsTracker.SecretListResponse, *k8sObjectsTracker.Error) {
	return &k8sObjectsTracker.SecretListResponse{
		Secrets: s.Secrets,
	}, nil
}

func (s K8SObjectTrackerMock) WithSecrets(secrets []k8sObjectsTracker.SecretListInfo) K8SObjectTrackerMock {
	s.Secrets = secrets
	return s
}
