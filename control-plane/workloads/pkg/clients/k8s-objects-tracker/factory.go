package k8s_objects_tracker

import (
	"context"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"sync"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"

	k8sObjectsTracker "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client"
	k8sObjectsApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
)

// ClientInterface lists the assets services that workloads service needs
type ClientInterface interface {
	GetStorageClassesConfigInt(ctx context.Context, clusterId string, storageClassName []string) (map[string]k8sObjectsApi.StorageClassConfiguration, *k8sObjectsApi.Error)
	// ListSecretsByScope lists all secrets of a given scope, including ancestors, all origins
	ListSecretsByScope(ctx context.Context, scopeType string, scopeId string) (*k8sObjectsApi.SecretListResponse, *k8sObjectsApi.Error)
}

var once sync.Once
var deFactoClient ClientInterface

// Get the real or mock client
func Get() ClientInterface {
	once.Do(initRealClient)
	return deFactoClient
}

// SetMock is used by unit test to apply mock client instead of the real client
func SetMock(mockClient ClientInterface) {
	deFactoClient = mockClient
}

// initialize the "real" client, invoked once if SetMock was not invoked by tests
func initRealClient() {
	if deFactoClient == nil {
		deFactoClient = k8sObjectsTracker.NewK8SObjectTrackerClient(runai_config.Get[*config.ServicesConfig]().ObjectsTrackerUrl)
	}
}
