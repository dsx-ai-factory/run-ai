package assets

import (
	"context"
	"sync"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	assetsService "github.com/run-ai/runai/runai-clients/assets-service-client"
	assetsApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

// ClientInterface lists the assets services that workloads service needs
type ClientInterface interface {
	RequestAssetFlatFormat(ctx context.Context, request assetsApi.AssetFlatRequest) (*assetsApi.AssetFlatResponse, *assetsApi.Error)

	ListTemplatesOfTenant(ctx context.Context) (*assetsApi.WorkloadTemplateListResponse, *assetsApi.Error)

	ListCredentialsAssets(ctx context.Context, privateOnly bool) (*assetsApi.CredentialsListResponse, *assetsApi.Error)
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
		deFactoClient = assetsService.NewAssetsClient(runai_config.Get[*config.ServicesConfig]().AssetsServiceUrl)
	}
}
