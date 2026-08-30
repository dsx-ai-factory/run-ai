package policies

import (
	"context"

	policyService "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
)

type PolicyClientOptions struct {
	NonCreationMode bool // policy service should not attempt to create policy for tenants not having any given policies
}

type ClientInterface interface {
	GetWorkspacePolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.WorkspacePolicyV2, *policyService.Error)
	GetTrainingPolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.TrainingPolicyV2, *policyService.Error)
	GetDistributedPolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.DistributedPolicyV2, *policyService.Error)
	GetInferencePolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.InferencePolicyV2, *policyService.Error)
	GetDistributedInferencePolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.DistributedInferencePolicyV2, *policyService.Error)
	GetNimServicePolicyOfProject(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.NimServicePolicy, *policyService.Error)

	GetGvkPolicyOfProject(ctx context.Context, projectId string, workloadType policyService.GVK) (*policyService.GetGvkPolicyResp, *policyService.Error)

	NotifyAssetModified(ctx context.Context, assetId string, scopeType policyService.ScopeType, scopeId string) *policyService.Error
	NotifyAssetDeleted(ctx context.Context, assetId string, scopeType policyService.ScopeType, scopeId string) *policyService.Error
}
