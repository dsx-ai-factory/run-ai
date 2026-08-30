package clusterservice

import (
	"sync"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	runaiconfig "github.com/run-ai/runai/runai-common-packages/config"
)

var (
	once          sync.Once
	deFactoClient clusterservice.ClientInterface
)

// Get the real or mock client
func Get() clusterservice.ClientInterface {
	once.Do(initRealClient)
	return deFactoClient
}

// SetMock is used by unit test to apply mock client instead of the real client
func SetMock(mockClient clusterservice.ClientInterface) {
	deFactoClient = mockClient
}

// initialize the "real" client, invoked once if SetMock was not invoked by tests
func initRealClient() {
	if deFactoClient == nil {
		serviceConfig := runaiconfig.Get[*config.ServicesConfig]()
		deFactoClient = clusterservice.GetRealClient(serviceConfig.ClusterServiceUrl)
	}
}
