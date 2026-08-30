package k8s_objects_tracker

import (
	"context"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"sync"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"

	k8sObjectsTracker "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client"
	k8sObjectsApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
)

// ClientInterface lists the assets services that workloads service needs
type ClientInterface interface {
	GetSecretsOfCluster(ctx context.Context, clusterId string) (*k8sObjectsApi.SecretListResponse, *k8sObjectsApi.Error)
	GetPvcsOfCluster(ctx context.Context, clusterId string) (*k8sObjectsApi.PVCListResponse, *k8sObjectsApi.Error)
	GetConfigMapsOfCluster(ctx context.Context, clusterId string) (*k8sObjectsApi.ConfigMapsListResponse, *k8sObjectsApi.Error)
	GetClusterResourcesStatus(ctx context.Context, resourceType *k8sObjectsApi.ClusterResourceType, scopeType *string,
		scopeId *string, secretType *k8sObjectsApi.SecretType) (*k8sObjectsApi.ClusterObjectsStatus, *k8sObjectsApi.Error)

	GetStorageClassConfigInt(ctx context.Context, clusterId string, storageClassName string) (*k8sObjectsApi.StorageClassConfiguration, *k8sObjectsApi.Error)
	GetStorageClassesConfigInt(ctx context.Context, clusterId string, storageClassName []string) (map[string]k8sObjectsApi.StorageClassConfiguration, *k8sObjectsApi.Error)

	ListPvcsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*k8sObjectsApi.PVCListResponse, *k8sObjectsApi.Error)
	ListConfigMapsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*k8sObjectsApi.ConfigMapsListResponse, *k8sObjectsApi.Error)
	ListSecretsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*k8sObjectsApi.SecretListResponse, *k8sObjectsApi.Error)
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
		deFactoClient = k8sObjectsTracker.NewK8SObjectTrackerClient(runai_config.Get[*config.ServicesConfig]().K8sObjectsTrackerUrl)
	}
}
