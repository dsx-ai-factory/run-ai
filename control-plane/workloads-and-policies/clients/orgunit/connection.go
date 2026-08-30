package orgunit

import (
	org_unit_service_client "github.com/run-ai/runai/runai-clients/org-unit-service-client"
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
		apiClient: org_unit_service_client.NewOrgUnitClient(orgUnitServiceUrl),
	}
}

func SetMock(clientFactory FactoryInterface) {
	factory = clientFactory
}

var orgUnitServiceUrl string

func InitConnection(orgUnitServiceUrlStr string) {
	orgUnitServiceUrl = orgUnitServiceUrlStr
}
