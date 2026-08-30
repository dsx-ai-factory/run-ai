package assets

import (
	"context"
	"sync"

	"github.com/run-ai/runai/control-plane/catalog-service/pkg/config"
	assetsService "github.com/run-ai/runai/runai-clients/assets-service-client"
	assetsApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
)

// ClientInterface lists the assets services that catalog service needs
type ClientInterface interface {
	GetNgcApiKeyByIdInternal(ctx context.Context, assetId string) (*assetsApi.NgcApiKeyInternal, *assetsApi.Error)
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
		baseUrl := runai_config.Get[*config.Services]().AssetsServiceUrl
		deFactoClient = assetsService.NewAssetsClient(baseUrl)
	}
}
