package resources

import (
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy"
	operandsutils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	workloadRunnerCreationPolicyName        = "runai-workloadrunner-creator"
	workloadRunnerCreationPolicyBindingName = workloadRunnerCreationPolicyName + "-binding"
	workloadRunnerCreationDeniedMessage     = "WorkloadRunner creation is restricted to the runai-agent service account"
	runaiCreatedByLabelKey                  = "run.ai/created-by"
	runaiOperatorName                       = "runai-operator"
)

func ValidatingAdmissionPolicyDefinition(runaiConfig *runv1.RunaiConfig) admissionpolicy.Definition {
	return admissionpolicy.Definition{
		Policy: &admissionv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: workloadRunnerCreationPolicyName,
				Labels: map[string]string{
					runaiCreatedByLabelKey: runaiOperatorName,
				},
			},
			Spec: admissionv1.ValidatingAdmissionPolicySpec{
				ParamKind: &admissionv1.ParamKind{
					APIVersion: "v1",
					Kind:       "ServiceAccount",
				},
				MatchConstraints: &admissionv1.MatchResources{
					ResourceRules: []admissionv1.NamedRuleWithOperations{
						{
							RuleWithOperations: admissionv1.RuleWithOperations{
								Operations: []admissionv1.OperationType{admissionv1.Create},
								Rule: admissionv1.Rule{
									APIGroups:   []string{"run.ai"},
									APIVersions: []string{"v1alpha1"},
									Resources:   []string{"workloadrunners"},
									Scope:       ptr.To(admissionv1.NamespacedScope),
								},
							},
						},
					},
				},
				Validations: []admissionv1.Validation{
					{
						Expression: "request.userInfo.uid == params.metadata.uid",
						Message:    workloadRunnerCreationDeniedMessage,
					},
				},
				FailurePolicy: ptr.To(admissionv1.Fail),
			},
		},
		Binding: &admissionv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: workloadRunnerCreationPolicyBindingName,
				Labels: map[string]string{
					runaiCreatedByLabelKey: runaiOperatorName,
				},
			},
			Spec: admissionv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName: workloadRunnerCreationPolicyName,
				ParamRef: &admissionv1.ParamRef{
					Name: operandsutils.GetNMCServiceAccountNameOrDefault(
						runaiConfig,
						operands.RunaiAgentName,
					),
					Namespace:               runaiConfig.GetNamespace(),
					ParameterNotFoundAction: ptr.To(admissionv1.DenyAction),
				},
				ValidationActions: []admissionv1.ValidationAction{admissionv1.Deny},
			},
		},
	}
}
