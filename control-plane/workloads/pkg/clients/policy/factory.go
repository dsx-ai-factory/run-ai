package policy

import (
	"context"
	openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"github.com/run-ai/runai/runai-clients/policy-service-client/structs"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"net/http"
	"sync"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	policyService "github.com/run-ai/runai/runai-clients/policy-service-client"
)

// ClientInterface lists the assets services that workloads service needs
type ClientInterface interface {
	GetWorkloadConfigOfTenant(ctx context.Context) (*map[string]*structs.PolicyWorkloadTypeConfig, *openapi.Error)

	UpdateDistributedPolicy(ctx context.Context, request openapi.DistributedPolicyChangeRequestV2) (*openapi.DistributedPolicyV2, *http.Response, error)
	UpdateInferencePolicy(ctx context.Context, request openapi.InferencePolicyChangeRequestV2) (*openapi.InferencePolicyV2, *http.Response, error)
	UpdateTrainingPolicy(ctx context.Context, request openapi.TrainingPolicyChangeRequestV2) (*openapi.TrainingPolicyV2, *http.Response, error)
	UpdateWorkspacePolicy(ctx context.Context, request openapi.WorkspacePolicyChangeRequestV2) (*openapi.WorkspacePolicyV2, *http.Response, error)
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
		deFactoClient = policyService.NewPoliciesClient(runai_config.Get[*config.ServicesConfig]().PolicyServiceUrl)
	}
}
