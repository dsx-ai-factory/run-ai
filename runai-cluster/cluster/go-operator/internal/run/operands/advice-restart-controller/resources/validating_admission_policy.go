package resources

import (
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	vapName                 = "runai-heimdall-workflow-id-immutable"
	vapBindingName          = "runai-heimdall-workflow-id-immutable-binding"
	heimdallWorkflowIDLabel = "heimdall.nvidia.com/workflow-id"
)

var labels = map[string]string{
	"app": "runai",
}

func failurePolicyFail() *admissionv1.FailurePolicyType {
	fail := admissionv1.Fail
	return &fail
}

// runnerTemplateLabels is a CEL returning labels from spec.template or {}
func runnerTemplateLabels(root string) string {
	return `has(` + root + `.spec) && has(` + root + `.spec.template) && ` +
		`has(` + root + `.spec.template.metadata) && has(` + root + `.spec.template.metadata.labels) ? ` +
		root + `.spec.template.metadata.labels : {}`
}

// ValidatingAdmissionPolicyDefinition guards the heimdall workflow-id label on
// WorkloadRunners, preventing it from being removed or modified once set
func ValidatingAdmissionPolicyDefinition() admissionpolicy.Definition {
	return admissionpolicy.Definition{
		Policy: &admissionv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:   vapName,
				Labels: labels,
			},
			Spec: admissionv1.ValidatingAdmissionPolicySpec{
				FailurePolicy: failurePolicyFail(),
				MatchConstraints: &admissionv1.MatchResources{
					NamespaceSelector: utils.NamespaceSelectorRunaiQueue(),
					ResourceRules: []admissionv1.NamedRuleWithOperations{
						{
							RuleWithOperations: admissionv1.RuleWithOperations{
								Operations: []admissionv1.OperationType{admissionv1.Update},
								Rule: admissionv1.Rule{
									APIGroups:   []string{"run.ai"},
									APIVersions: []string{"v1alpha1"},
									Resources:   []string{"workloadrunners"},
								},
							},
						},
					},
				},
				Variables: []admissionv1.Variable{
					{Name: "oldLabels", Expression: runnerTemplateLabels("oldObject")},
					{Name: "newLabels", Expression: runnerTemplateLabels("object")},
				},
				Validations: []admissionv1.Validation{
					{
						Expression: `!('` + heimdallWorkflowIDLabel + `' in variables.oldLabels) || ` +
							`'` + heimdallWorkflowIDLabel + `' in variables.newLabels`,
						Message: "The label " + heimdallWorkflowIDLabel + " cannot be removed once set",
					},
					{
						Expression: `!('` + heimdallWorkflowIDLabel + `' in variables.oldLabels) || ` +
							`!('` + heimdallWorkflowIDLabel + `' in variables.newLabels) || ` +
							`variables.newLabels['` + heimdallWorkflowIDLabel + `'] == ` +
							`variables.oldLabels['` + heimdallWorkflowIDLabel + `']`,
						Message: "The label " + heimdallWorkflowIDLabel + " cannot be modified once set",
					},
				},
			},
		},
		Binding: &admissionv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:   vapBindingName,
				Labels: labels,
			},
			Spec: admissionv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName:        vapName,
				ValidationActions: []admissionv1.ValidationAction{admissionv1.Deny},
			},
		},
	}
}
