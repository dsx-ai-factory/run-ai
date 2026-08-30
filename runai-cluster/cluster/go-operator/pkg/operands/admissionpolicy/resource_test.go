package admissionpolicy

import (
	"context"
	"testing"

	operandsutils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	operandName = "test-service"
	policyName  = "runai-test-policy"
	bindingName = "runai-test-policy-binding"
)

func TestNewResources(t *testing.T) {
	definition := validDefinition()
	resources := NewResources(operandName, fakeReader(t), definition)

	require.Len(t, resources, 2)
	assert.Equal(t, operandName, resources[0].OperandName())
	assert.Equal(t, operandName, resources[1].OperandName())
	assert.True(t, resources[0].Enabled())
	assert.True(t, resources[1].Enabled())

	policyObject, err := resources[0].Object(context.Background())
	require.NoError(t, err)
	bindingObject, err := resources[1].Object(context.Background())
	require.NoError(t, err)

	policy, ok := policyObject.(*admissionv1.ValidatingAdmissionPolicy)
	require.True(t, ok)
	binding, ok := bindingObject.(*admissionv1.ValidatingAdmissionPolicyBinding)
	require.True(t, ok)
	assert.Equal(t, policyName, policy.Name)
	assert.Equal(t, bindingName, binding.Name)
	assert.Equal(t, policyName, binding.Spec.PolicyName)
}

func TestNewResourcesSnapshotsAndCopiesDefinition(t *testing.T) {
	definition := validDefinition()
	resources := NewResources(operandName, fakeReader(t), definition)
	definition.Policy.Spec.Validations[0].Expression = "false"

	firstObject, err := resources[0].Object(context.Background())
	require.NoError(t, err)
	first := firstObject.(*admissionv1.ValidatingAdmissionPolicy)
	assert.Equal(t, "true", first.Spec.Validations[0].Expression)

	first.Spec.Validations[0].Expression = "false"
	secondObject, err := resources[0].Object(context.Background())
	require.NoError(t, err)
	second := secondObject.(*admissionv1.ValidatingAdmissionPolicy)
	assert.Equal(t, "true", second.Spec.Validations[0].Expression)
}

func TestNewResourcesUsesEnabledCallback(t *testing.T) {
	enabled := false
	definition := validDefinition()
	definition.Enabled = func() bool { return enabled }
	resources := NewResources(operandName, fakeReader(t), definition)

	assert.False(t, resources[0].Enabled())
	assert.False(t, resources[1].Enabled())
	enabled = true
	assert.True(t, resources[0].Enabled())
	assert.True(t, resources[1].Enabled())
}

func TestResourceStatus(t *testing.T) {
	definition := validDefinition()

	t.Run("resources exist", func(t *testing.T) {
		reader := fakeReader(t, definition.Policy, definition.Binding)
		resources := NewResources(operandName, reader, definition)

		for _, resource := range resources {
			assert.NoError(t, resource.Status(context.Background()))
		}
	})

	t.Run("policy is missing", func(t *testing.T) {
		resources := NewResources(operandName, fakeReader(t, definition.Binding), definition)
		err := resources[0].Status(context.Background())
		require.EqualError(t, err, "validating admission policy [runai-test-policy] is missing")
	})

	t.Run("binding is missing", func(t *testing.T) {
		resources := NewResources(operandName, fakeReader(t, definition.Policy), definition)
		err := resources[1].Status(context.Background())
		require.EqualError(t, err, "validating admission policy binding [runai-test-policy-binding] is missing")
	})
}

func TestMissingResourceMakesAggregatedStatusNotReady(t *testing.T) {
	definition := validDefinition()
	resources := NewResources(operandName, fakeReader(t, definition.Binding), definition)

	status, err := operandsutils.Status(context.Background(), resources)

	require.Error(t, err)
	assert.False(t, status.Ready)
	assert.Contains(t, status.Reasons, "validating admission policy [runai-test-policy] is missing")
}

func TestDefinitionValidation(t *testing.T) {
	tests := []struct {
		name       string
		definition Definition
		operand    string
		expected   string
	}{
		{name: "missing policy", definition: Definition{Binding: validDefinition().Binding}, operand: operandName, expected: "validating admission policy is required"},
		{name: "missing binding", definition: Definition{Policy: validDefinition().Policy}, operand: operandName, expected: "validating admission policy binding is required"},
		{name: "missing policy name", definition: definitionWith(func(d *Definition) { d.Policy.Name = "" }), operand: operandName, expected: "validating admission policy name is required"},
		{name: "missing binding name", definition: definitionWith(func(d *Definition) { d.Binding.Name = "" }), operand: operandName, expected: "validating admission policy binding name is required"},
		{name: "namespaced policy", definition: definitionWith(func(d *Definition) { d.Policy.Namespace = "runai" }), operand: operandName, expected: "validating admission policy [runai-test-policy] must be cluster-scoped"},
		{name: "namespaced binding", definition: definitionWith(func(d *Definition) { d.Binding.Namespace = "runai" }), operand: operandName, expected: "validating admission policy binding [runai-test-policy-binding] must be cluster-scoped"},
		{name: "wrong policy reference", definition: definitionWith(func(d *Definition) { d.Binding.Spec.PolicyName = "other" }), operand: operandName, expected: "validating admission policy binding [runai-test-policy-binding] references policy [other], expected [runai-test-policy]"},
		{name: "missing operand name", definition: validDefinition(), expected: "operand name is required for admission policy resources"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := NewResources(tt.operand, fakeReader(t), tt.definition)
			_, err := resources[0].Object(context.Background())
			require.EqualError(t, err, tt.expected)
		})
	}
}

func validDefinition() Definition {
	return Definition{
		Policy: &admissionv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: policyName},
			Spec: admissionv1.ValidatingAdmissionPolicySpec{
				Validations: []admissionv1.Validation{{Expression: "true"}},
			},
		},
		Binding: &admissionv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{Name: bindingName},
			Spec: admissionv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName: policyName,
			},
		},
	}
}

func definitionWith(change func(*Definition)) Definition {
	definition := validDefinition()
	change(&definition)
	return definition
}

func fakeReader(t *testing.T, objects ...client.Object) client.Reader {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, admissionv1.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}
