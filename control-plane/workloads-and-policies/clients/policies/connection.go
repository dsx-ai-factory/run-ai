package policies

import (
	policyService "github.com/run-ai/runai/runai-clients/policy-service-client"
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
		apiClient: policyService.NewPoliciesClient(policyServiceUrl),
	}
}

func SetMock(clientFactory FactoryInterface) {
	factory = clientFactory
}

var policyServiceUrl string

func InitConnection(policyServiceUrlStr string) {
	policyServiceUrl = policyServiceUrlStr
}
