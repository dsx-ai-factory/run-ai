package policy_service_client

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/runai-clients/policy-service-client/structs"

	client_utils "github.com/run-ai/runai/runai-application-packages/client-utils"
	openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"github.com/run-ai/runai/runai-clients/policy-service-client/internal"
)

//go:generate mockery --name IPolicyExternalClient
type IPolicyExternalClient interface {
	UpdateSyncStatusWorkspace(ctx context.Context, syncStatus openapi.SyncStatusUpdate) (*http.Response, error)
	UpdateSyncStatusTraining(ctx context.Context, syncStatus openapi.SyncStatusUpdate) (*http.Response, error)
	UpdateSyncStatusDistributed(ctx context.Context, syncStatus openapi.SyncStatusUpdate) (*http.Response, error)
	UpdateSyncStatusInference(ctx context.Context, syncStatus openapi.SyncStatusUpdate) (*http.Response, error)
	SyncWorkspacePolicies(ctx context.Context, clusterId string, after int64, syncReport *openapi.ClusterSyncReport) (*openapi.PolicySyncResponse, *http.Response, error)
	SyncTrainingPolicies(ctx context.Context, clusterId string, after int64, syncReport *openapi.ClusterSyncReport) (*openapi.PolicySyncResponse, *http.Response, error)
	SyncDistributedPolicies(ctx context.Context, clusterId string, after int64, syncReport *openapi.ClusterSyncReport) (*openapi.PolicySyncResponse, *http.Response, error)
	SyncInferencePolicies(ctx context.Context, clusterId string, after int64, syncReport *openapi.ClusterSyncReport) (*openapi.PolicySyncResponse, *http.Response, error)
	GetWorkspacePolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.WorkspacePolicyV2, *http.Response, error)
	GetTrainingPolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.TrainingPolicyV2, *http.Response, error)
	GetDistributedPolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.DistributedPolicyV2, *http.Response, error)
	GetInferencePolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.InferencePolicyV2, *http.Response, error)
	GetDistributedInferencePolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.DistributedInferencePolicyV2, *http.Response, error)
	GetNimServicePolicy(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.NimServicePolicy, *http.Response, error)
	GetAssetChangedHook(ctx context.Context, assetId string, notification *openapi.AssetChangeNotification) (*http.Response, error)
}

//go:generate mockery --name IPolicyClient
type IPolicyClient interface {
	IPolicyExternalClient

	GetPolicyCrossType(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.PolicyCrossType, *openapi.Error)

	UpdateDistributedPolicy(ctx context.Context, request openapi.DistributedPolicyChangeRequestV2) (*openapi.DistributedPolicyV2, *http.Response, error)
	UpdateInferencePolicy(ctx context.Context, request openapi.InferencePolicyChangeRequestV2) (*openapi.InferencePolicyV2, *http.Response, error)
	UpdateTrainingPolicy(ctx context.Context, request openapi.TrainingPolicyChangeRequestV2) (*openapi.TrainingPolicyV2, *http.Response, error)
	UpdateWorkspacePolicy(ctx context.Context, request openapi.WorkspacePolicyChangeRequestV2) (*openapi.WorkspacePolicyV2, *http.Response, error)
	GetWorkloadConfigOfTenant(ctx context.Context) (*map[string]*structs.PolicyWorkloadTypeConfig, *openapi.Error)
	GetGvkPolicyOfProject(ctx context.Context, projectId string, workloadType openapi.GVK) (*openapi.GetGvkPolicyResp, *openapi.Error)
}

func NewPoliciesClient(baseURL string) IPolicyClient {
	config := openapi.NewConfiguration()
	config.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	return &internal.PoliciesClient{ApiClient: openapi.NewAPIClient(config)}
}

func NewPoliciesExternalClient(baseURL string) IPolicyExternalClient {
	return NewPoliciesClient(client_utils.GetExternalUrl(baseURL))
}

func NewPoliciesClientWithConfig(config *openapi.Configuration) IPolicyClient {
	return &internal.PoliciesClient{ApiClient: openapi.NewAPIClient(config)}
}

func NewPoliciesExternalClientWithConfig(config *openapi.Configuration) IPolicyExternalClient {
	return NewPoliciesClientWithConfig(config)
}
