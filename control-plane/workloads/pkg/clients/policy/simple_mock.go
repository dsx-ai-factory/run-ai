package policy

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/structs"
)

type SimpleMock struct {
	policyWorkloadTypeConfig map[string]*structs.PolicyWorkloadTypeConfig
}

var defaultPolicyWorkloadTypeConfig = map[string]*structs.PolicyWorkloadTypeConfig{
	"InteractiveWorkload": {
		CategoryName:   convutils.Ref("Build"),
		PriorityName:   convutils.Ref("high"),
		Preemptibility: convutils.Ref(openapi.PREEMPTIBILITY_NON_PREEMPTIBLE),
	},
	"DistributedWorkload": {
		CategoryName:   convutils.Ref("Train"),
		PriorityName:   convutils.Ref("low"),
		Preemptibility: convutils.Ref(openapi.PREEMPTIBILITY_PREEMPTIBLE),
	},
	"TrainingWorkload": {
		CategoryName:   convutils.Ref("Train"),
		PriorityName:   convutils.Ref("low"),
		Preemptibility: convutils.Ref(openapi.PREEMPTIBILITY_PREEMPTIBLE),
	},
	"InferenceWorkload": {
		CategoryName:   convutils.Ref("Deploy"),
		PriorityName:   convutils.Ref("very-high"),
		Preemptibility: convutils.Ref(openapi.PREEMPTIBILITY_NON_PREEMPTIBLE),
	},
	"DistributedInferenceWorkload": {
		CategoryName:   convutils.Ref("Deploy"),
		PriorityName:   convutils.Ref("very-high"),
		Preemptibility: convutils.Ref(openapi.PREEMPTIBILITY_NON_PREEMPTIBLE),
	},
}

func NewSimpleMock() *SimpleMock {
	// Create a copy of the default config to avoid test interference
	configCopy := make(map[string]*structs.PolicyWorkloadTypeConfig)
	for k, v := range defaultPolicyWorkloadTypeConfig {
		configCopy[k] = &structs.PolicyWorkloadTypeConfig{
			CategoryName:   v.CategoryName,
			PriorityName:   v.PriorityName,
			Preemptibility: v.Preemptibility,
		}
	}
	return &SimpleMock{
		policyWorkloadTypeConfig: configCopy,
	}
}

func (smc *SimpleMock) SetPoliciesConfig(policyWorkloadTypeConfig map[string]*structs.PolicyWorkloadTypeConfig) {
	smc.policyWorkloadTypeConfig = policyWorkloadTypeConfig
}

func (smc *SimpleMock) GetWorkloadConfigOfTenant(_ context.Context) (*map[string]*structs.PolicyWorkloadTypeConfig, *openapi.Error) {
	return &smc.policyWorkloadTypeConfig, nil
}

func (smc *SimpleMock) UpdateWorkspacePolicy(_ context.Context, request openapi.WorkspacePolicyChangeRequestV2) (*openapi.WorkspacePolicyV2, *http.Response, error) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
	}
	// Update internal state so GetWorkloadConfigOfTenant returns updated values
	if convutils.HasValue(request.Policy) && convutils.HasValue(request.Policy.Get().Defaults) {
		defaults := request.Policy.Get().Defaults.Get()
		smc.updatePolicyConfig("InteractiveWorkload", defaults.Category.Get(), defaults.PriorityClass.Get(), defaults.Preemptibility.Get())
	}
	return &openapi.WorkspacePolicyV2{Policy: request.Policy}, resp, nil
}

func (smc *SimpleMock) UpdateTrainingPolicy(_ context.Context, request openapi.TrainingPolicyChangeRequestV2) (*openapi.TrainingPolicyV2, *http.Response, error) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
	}
	// Update internal state so GetWorkloadConfigOfTenant returns updated values
	if convutils.HasValue(request.Policy) && convutils.HasValue(request.Policy.Get().Defaults) {
		defaults := request.Policy.Get().Defaults.Get()
		smc.updatePolicyConfig("TrainingWorkload", defaults.Category.Get(), defaults.PriorityClass.Get(), defaults.Preemptibility.Get())
	}
	return &openapi.TrainingPolicyV2{Policy: request.Policy}, resp, nil
}

func (smc *SimpleMock) UpdateDistributedPolicy(_ context.Context, request openapi.DistributedPolicyChangeRequestV2) (*openapi.DistributedPolicyV2, *http.Response, error) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
	}
	// Update internal state so GetWorkloadConfigOfTenant returns updated values
	if convutils.HasValue(request.Policy) && convutils.HasValue(request.Policy.Get().Defaults) && convutils.HasValue(request.Policy.Get().Defaults.Get().Master) {
		master := request.Policy.Get().Defaults.Get().Master.Get()
		smc.updatePolicyConfig("DistributedWorkload", master.Category.Get(), master.PriorityClass.Get(), master.Preemptibility.Get())
	}
	return &openapi.DistributedPolicyV2{Policy: request.Policy}, resp, nil
}

func (smc *SimpleMock) UpdateInferencePolicy(_ context.Context, request openapi.InferencePolicyChangeRequestV2) (*openapi.InferencePolicyV2, *http.Response, error) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
	}
	// Update internal state so GetWorkloadConfigOfTenant returns updated values
	if convutils.HasValue(request.Policy) && convutils.HasValue(request.Policy.Get().Defaults) {
		defaults := request.Policy.Get().Defaults.Get()
		smc.updatePolicyConfig("InferenceWorkload", defaults.Category.Get(), defaults.PriorityClass.Get(), defaults.Preemptibility.Get())
	}
	return &openapi.InferencePolicyV2{Policy: request.Policy}, resp, nil
}

func (smc *SimpleMock) updatePolicyConfig(workloadKind string, category *string, priority *string, preemptibility *openapi.Preemptibility) {
	if smc.policyWorkloadTypeConfig[workloadKind] == nil {
		smc.policyWorkloadTypeConfig[workloadKind] = &structs.PolicyWorkloadTypeConfig{}
	}
	if category != nil {
		smc.policyWorkloadTypeConfig[workloadKind].CategoryName = category
	}
	if priority != nil {
		smc.policyWorkloadTypeConfig[workloadKind].PriorityName = priority
	}
	if preemptibility != nil {
		smc.policyWorkloadTypeConfig[workloadKind].Preemptibility = preemptibility
	}
}

func (smc *SimpleMock) GetPolicyCrossType(ctx context.Context, scope string, clusterId *string, departmentId *string, projectId *string, noCreationMode bool) (*openapi.PolicyCrossType, *openapi.Error) {
	return nil, nil
}
