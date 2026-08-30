package resources

import (
	"reflect"
	"testing"

	"github.com/google/cel-go/cel"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	admissionv1 "k8s.io/api/admissionregistration/v1"
)

func TestWorkloadRunnerCreationPolicyUsesRunaiOperatorOwnershipLabel(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{}
	runaiConfig.SetDefaultsWhereNeeded()

	definition := ValidatingAdmissionPolicyDefinition(runaiConfig)
	wantLabels := map[string]string{
		"run.ai/created-by": "runai-operator",
	}

	if !reflect.DeepEqual(definition.Policy.Labels, wantLabels) {
		t.Fatalf("policy labels = %v, want %v", definition.Policy.Labels, wantLabels)
	}
	if !reflect.DeepEqual(definition.Binding.Labels, wantLabels) {
		t.Fatalf("binding labels = %v, want %v", definition.Binding.Labels, wantLabels)
	}
}

func TestWorkloadRunnerCreationPolicy(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{}
	runaiConfig.SetDefaultsWhereNeeded()
	definition := ValidatingAdmissionPolicyDefinition(runaiConfig)
	policy := definition.Policy

	tests := []struct {
		name         string
		operation    admissionv1.OperationType
		requesterUID string
		wantAllowed  bool
	}{
		{
			name:         "create as agent",
			operation:    admissionv1.Create,
			requesterUID: "agent-uid",
			wantAllowed:  true,
		},
		{
			name:         "create as another service account",
			operation:    admissionv1.Create,
			requesterUID: "other-uid",
			wantAllowed:  false,
		},
		{
			name:         "update as another service account",
			operation:    admissionv1.Update,
			requesterUID: "other-uid",
			wantAllowed:  true,
		},
		{
			name:         "delete as another service account",
			operation:    admissionv1.Delete,
			requesterUID: "other-uid",
			wantAllowed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := evaluatePolicy(t, policy, tt.operation, tt.requesterUID, "agent-uid")
			if allowed != tt.wantAllowed {
				t.Fatalf("allowed = %t, want %t", allowed, tt.wantAllowed)
			}
		})
	}
}

func TestWorkloadRunnerCreationPolicyConfiguration(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{
		Spec: runv1.RunaiConfigSpec{
			Internal: &runv1internal.Internal{
				Global: &runv1internal.Global{},
			},
		},
	}
	runaiConfig.Namespace = "runai"

	definition := ValidatingAdmissionPolicyDefinition(runaiConfig)
	policy := definition.Policy

	if policy.Spec.ParamKind == nil || policy.Spec.ParamKind.APIVersion != "v1" || policy.Spec.ParamKind.Kind != "ServiceAccount" {
		t.Fatalf("unexpected parameter kind: %#v", policy.Spec.ParamKind)
	}
	if policy.Spec.FailurePolicy == nil || *policy.Spec.FailurePolicy != admissionv1.Fail {
		t.Fatalf("failure policy = %v, want Fail", policy.Spec.FailurePolicy)
	}
	if len(policy.Spec.Validations) != 1 || policy.Spec.Validations[0].Message != workloadRunnerCreationDeniedMessage {
		t.Fatalf("unexpected validations: %#v", policy.Spec.Validations)
	}
	if len(policy.Spec.MatchConstraints.ResourceRules) != 1 {
		t.Fatalf("resource rules count = %d, want 1", len(policy.Spec.MatchConstraints.ResourceRules))
	}
	rule := policy.Spec.MatchConstraints.ResourceRules[0].RuleWithOperations
	if len(rule.Operations) != 1 || rule.Operations[0] != admissionv1.Create {
		t.Fatalf("operations = %v, want [CREATE]", rule.Operations)
	}
	if len(rule.APIGroups) != 1 || rule.APIGroups[0] != "run.ai" ||
		len(rule.APIVersions) != 1 || rule.APIVersions[0] != "v1alpha1" ||
		len(rule.Resources) != 1 || rule.Resources[0] != "workloadrunners" {
		t.Fatalf("unexpected workload runner rule: %#v", rule.Rule)
	}

	binding := definition.Binding
	if binding.Spec.ParamRef == nil || binding.Spec.ParamRef.Name != "runai-agent" || binding.Spec.ParamRef.Namespace != "runai" {
		t.Fatalf("unexpected parameter reference: %#v", binding.Spec.ParamRef)
	}
	if binding.Spec.ParamRef.ParameterNotFoundAction == nil || *binding.Spec.ParamRef.ParameterNotFoundAction != admissionv1.DenyAction {
		t.Fatalf("parameter-not-found action = %v, want Deny", binding.Spec.ParamRef.ParameterNotFoundAction)
	}
	if len(binding.Spec.ValidationActions) != 1 || binding.Spec.ValidationActions[0] != admissionv1.Deny {
		t.Fatalf("validation actions = %v, want [Deny]", binding.Spec.ValidationActions)
	}
}

func TestWorkloadRunnerCreationPolicyUsesConfiguredAgentServiceAccount(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{
		Spec: runv1.RunaiConfigSpec{
			Internal: &runv1internal.Internal{
				Global: &runv1internal.Global{
					NMC: runv1internal.NMC{
						ServiceAccountName: "nmc-agent",
					},
				},
			},
		},
	}
	runaiConfig.Namespace = "runai"

	binding := ValidatingAdmissionPolicyDefinition(runaiConfig).Binding
	if binding.Spec.ParamRef == nil || binding.Spec.ParamRef.Name != "nmc-agent" {
		t.Fatalf("parameter service account = %#v, want nmc-agent", binding.Spec.ParamRef)
	}
}

func evaluatePolicy(
	t *testing.T,
	policy *admissionv1.ValidatingAdmissionPolicy,
	operation admissionv1.OperationType,
	requesterUID string,
	agentUID string,
) bool {
	t.Helper()

	matched := false
	for _, rule := range policy.Spec.MatchConstraints.ResourceRules {
		for _, matchedOperation := range rule.Operations {
			if matchedOperation == operation {
				matched = true
			}
		}
	}
	if !matched {
		return true
	}

	environment, err := cel.NewEnv(
		cel.Variable("request", cel.DynType),
		cel.Variable("params", cel.DynType),
	)
	if err != nil {
		t.Fatalf("failed to create CEL environment: %v", err)
	}
	ast, issues := environment.Compile(policy.Spec.Validations[0].Expression)
	if issues != nil && issues.Err() != nil {
		t.Fatalf("failed to compile validation: %v", issues.Err())
	}
	program, err := environment.Program(ast)
	if err != nil {
		t.Fatalf("failed to create CEL program: %v", err)
	}
	result, _, err := program.Eval(map[string]any{
		"request": map[string]any{
			"userInfo": map[string]any{
				"uid": requesterUID,
			},
		},
		"params": map[string]any{
			"metadata": map[string]any{
				"uid": agentUID,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to evaluate validation: %v", err)
	}
	allowed, ok := result.Value().(bool)
	if !ok {
		t.Fatalf("validation returned %T, want bool", result.Value())
	}
	return allowed
}
