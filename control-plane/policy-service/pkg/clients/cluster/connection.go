package cluster

import (
	clusterService "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/config"
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
		apiClient: clusterService.NewClusterClient(runai_config.Get[*config.ServicesConfig]().ClusterServiceUrl),
	}
}

func SetMock(mockFactory FactoryInterface) {
	factory = mockFactory
}
