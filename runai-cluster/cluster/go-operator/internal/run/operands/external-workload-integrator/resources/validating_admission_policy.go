/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"fmt"

	runutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// PodOwnershipPolicyDefinition guards existing Pod ownership labels. Creation and
// initially absent labels remain allowed for native and legacy workload flows.
// EWI may complete or correct partially populated ownership metadata.
func PodOwnershipPolicyDefinition(runaiConfig *runv1.RunaiConfig) admissionpolicy.Definition {
	const policyName = "runai-pod-ownership"
	ewiUsername := fmt.Sprintf("system:serviceaccount:%s:%s", runaiConfig.GetNamespace(),
		utils.GetNMCServiceAccountNameOrDefault(runaiConfig, operands.RunaiExternalWorkloadIntegratorName))
	labels := map[string]string{"run.ai/created-by": "runai-operator"}
	var validations []admissionv1.Validation
	// Protect both selectors while grouping consumers still use top-owner-uid.
	//nolint:staticcheck // Guard the deprecated label while grouping consumers still select it.
	for _, key := range []string{constants.TopOwnerWorkloadIdLabel, constants.TopOwnerUidLabel} {
		validations = append(validations, admissionv1.Validation{
			Expression: fmt.Sprintf(`request.userInfo.username == variables.ewiUsername ||
    !has(oldObject.metadata.labels) || !('%[1]s' in oldObject.metadata.labels) ||
    oldObject.metadata.labels['%[1]s'] == '' ||
    (has(object.metadata.labels) && '%[1]s' in object.metadata.labels &&
     object.metadata.labels['%[1]s'] == oldObject.metadata.labels['%[1]s'])`, key),
			Message: "only external-workload-integrator may change or remove existing Pod ownership label " + key,
			Reason:  ptr.To(metav1.StatusReasonForbidden),
		})
	}
	return admissionpolicy.Definition{
		Policy: &admissionv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: policyName, Labels: labels},
			Spec: admissionv1.ValidatingAdmissionPolicySpec{
				FailurePolicy: ptr.To(admissionv1.Fail),
				MatchConstraints: &admissionv1.MatchResources{
					// Match EWI's managed-namespace boundary.
					NamespaceSelector: runutils.NamespaceSelectorRunaiQueue(),
					ResourceRules: []admissionv1.NamedRuleWithOperations{{
						RuleWithOperations: admissionv1.RuleWithOperations{
							Operations: []admissionv1.OperationType{admissionv1.Update},
							Rule: admissionv1.Rule{APIGroups: []string{""}, APIVersions: []string{"v1"},
								Resources: []string{"pods", "pods/*"}, Scope: ptr.To(admissionv1.NamespacedScope)},
						},
					}}},
				Variables:   []admissionv1.Variable{{Name: "ewiUsername", Expression: fmt.Sprintf("'%s'", ewiUsername)}},
				Validations: validations,
			},
		},
		Binding: &admissionv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{Name: policyName + "-binding", Labels: labels},
			Spec: admissionv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName: policyName,
				// Observe existing cluster-side writers before enforcing this new restriction.
				ValidationActions: []admissionv1.ValidationAction{admissionv1.Audit},
			},
		},
	}
}
