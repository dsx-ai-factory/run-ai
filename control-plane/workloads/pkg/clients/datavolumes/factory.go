package datavolumes

import (
	"context"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"sync"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	datavolumes_service_client "github.com/run-ai/runai/runai-clients/datavolumes-service-client"
	openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
)

// ClientInterface lists the assets services that workloads service needs
type ClientInterface interface {
	GetDataVolumesNotPaginated(ctx context.Context, requestType openapi.DatavolumeRequestType, usableInProjectId string,
		filter []string, sortBy string, sortOrder string) ([]openapi.DatavolumeNoSharedScopes, error)
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
		deFactoClient = datavolumes_service_client.NewDatavolumesClient(runai_config.Get[*config.ServicesConfig]().DataVolumeUrl)
	}
}
