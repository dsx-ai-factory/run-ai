package anyworkload_controller

import (
	"context"
	"testing"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
)

func TestOperandOwnsWorkloadRunnerCreationPolicy(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{}
	runaiConfig.SetDefaultsWhereNeeded()
	operand := Operand(nil, runaiConfig)
	operandResources := operand.Resources()
	if len(operandResources) < 2 {
		t.Fatalf("operand resources count = %d, want at least 2", len(operandResources))
	}

	var hasPolicy, hasBinding bool
	for _, resource := range operandResources[len(operandResources)-2:] {
		object, err := resource.Object(context.Background())
		if err != nil {
			t.Fatalf("failed to build operand resource: %v", err)
		}
		switch object.(type) {
		case *admissionv1.ValidatingAdmissionPolicy:
			hasPolicy = true
		case *admissionv1.ValidatingAdmissionPolicyBinding:
			hasBinding = true
		}
	}

	if !hasPolicy {
		t.Error("anyworkload-controller operand does not own a validating admission policy")
	}
	if !hasBinding {
		t.Error("anyworkload-controller operand does not own a validating admission policy binding")
	}
}
