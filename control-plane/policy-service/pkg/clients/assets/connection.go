package assets

import (
	"github.com/run-ai/runai/control-plane/policy-service/pkg/config"
	assetsService "github.com/run-ai/runai/runai-clients/assets-service-client"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
)

type FactoryInterface interface {
	Client() ClientInterface
}

var factory FactoryInterface = RealFactory{}

func Get() FactoryInterface {
	return factory
}

type RealFactory struct{}

func (c RealFactory) Client() ClientInterface {
	return &Client{
		apiClient: assetsService.NewAssetsClient(runai_config.Get[*config.ServicesConfig]().AssetsServiceUrl),
	}
}

func SetMock(mockFactory FactoryInterface) {
	factory = mockFactory
}
