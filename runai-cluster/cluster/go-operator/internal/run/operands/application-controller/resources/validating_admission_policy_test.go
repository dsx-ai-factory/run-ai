package resources

import (
	"testing"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidatingAdmissionPolicyDefinition(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "custom-runai"}}

	definition := ValidatingAdmissionPolicyDefinition(runaiConfig)

	require.NotNil(t, definition.Policy)
	require.NotNil(t, definition.Binding)
	assert.Equal(t, "runai-application-identity", definition.Policy.Name)
	assert.Equal(t, "runai-application-identity-binding", definition.Binding.Name)
	assert.Equal(t, "runai-application-identity", definition.Binding.Spec.PolicyName)
	assert.Equal(t, map[string]string{
		"run.ai/created-by": "runai-operator",
	}, definition.Policy.Labels)
	assert.Equal(t, map[string]string{
		"run.ai/created-by": "runai-operator",
	}, definition.Binding.Labels)
	require.NotNil(t, definition.Policy.Spec.FailurePolicy)
	assert.Equal(t, admissionv1.Fail, *definition.Policy.Spec.FailurePolicy)
	assert.Equal(t, []admissionv1.ValidationAction{admissionv1.Audit}, definition.Binding.Spec.ValidationActions)

	require.NotNil(t, definition.Policy.Spec.MatchConstraints)
	require.Len(t, definition.Policy.Spec.MatchConstraints.ResourceRules, 1)
	rule := definition.Policy.Spec.MatchConstraints.ResourceRules[0]
	assert.Equal(t, []admissionv1.OperationType{admissionv1.Create, admissionv1.Update, admissionv1.Delete}, rule.Operations)
	assert.Equal(t, []string{"run.ai"}, rule.APIGroups)
	assert.Equal(t, []string{"*"}, rule.APIVersions)
	assert.Equal(t, []string{"applications"}, rule.Resources)
	require.NotNil(t, rule.Scope)
	assert.Equal(t, admissionv1.NamespacedScope, *rule.Scope)

	require.Len(t, definition.Policy.Spec.Validations, 1)
	assert.Equal(t,
		"request.userInfo.username == 'system:serviceaccount:custom-runai:application-controller'",
		definition.Policy.Spec.Validations[0].Expression)
}
