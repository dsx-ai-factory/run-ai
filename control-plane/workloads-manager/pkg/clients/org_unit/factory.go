package org_unit

import (
	"context"
	"github.com/run-ai/runai/runai-common-packages/config"
	"sync"

	ourConfig "github.com/run-ai/runai/control-plane/workloads-manager/pkg/config"

	orgUnitService "github.com/run-ai/runai/runai-clients/org-unit-service-client"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

// ClientInterface lists the assets services that workloads service needs
type ClientInterface interface {
	GetProjectById(ctx context.Context, projectId string) (*orgUnitApi.Project, *orgUnitApi.Error)
	GetProjectsMetadataByCluster(ctx context.Context, clusterId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error)
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
		servicesConfig := config.Get[*ourConfig.Services]()
		deFactoClient = orgUnitService.NewOrgUnitClient(servicesConfig.OrgUnitUrl)
	}
}
