package authorization

import authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"

type FactoryInterface interface {
	Client() ClientInterface
}

var factory FactoryInterface = RealFactory{}

type RealFactory struct{}

func (f RealFactory) Client() ClientInterface {
	return &Client{
		apiClient: authClient.NewAuthorizationClient(),
	}
}

func SetMock(testFactory FactoryInterface) {
	factory = testFactory
}

func Get() FactoryInterface { return factory }
