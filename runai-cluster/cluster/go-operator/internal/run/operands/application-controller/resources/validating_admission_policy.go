/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	applicationIdentityPolicyName  = "runai-application-identity"
	applicationIdentityBindingName = applicationIdentityPolicyName + "-binding"
	runaiCreatedByLabelKey         = "run.ai/created-by"
	runaiOperatorName              = "runai-operator"
)

func ValidatingAdmissionPolicyDefinition(runaiConfig *runv1.RunaiConfig) admissionpolicy.Definition {
	applicationControllerUsername := fmt.Sprintf("system:serviceaccount:%s:%s",
		runaiConfig.GetNamespace(), operands.RunaiApplicationControllerName)

	return admissionpolicy.Definition{
		Policy: &admissionv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: applicationIdentityPolicyName,
				Labels: map[string]string{
					runaiCreatedByLabelKey: runaiOperatorName,
				},
			},
			Spec: admissionv1.ValidatingAdmissionPolicySpec{
				FailurePolicy: ptr.To(admissionv1.Fail),
				MatchConstraints: &admissionv1.MatchResources{
					ResourceRules: []admissionv1.NamedRuleWithOperations{
						{
							RuleWithOperations: admissionv1.RuleWithOperations{
								Operations: []admissionv1.OperationType{
									admissionv1.Create,
									admissionv1.Update,
									admissionv1.Delete,
								},
								Rule: admissionv1.Rule{
									APIGroups:   []string{"run.ai"},
									APIVersions: []string{"*"},
									Resources:   []string{"applications"},
									Scope:       ptr.To(admissionv1.NamespacedScope),
								},
							},
						},
					},
				},
				Validations: []admissionv1.Validation{
					{
						Expression: fmt.Sprintf("request.userInfo.username == '%s'", applicationControllerUsername),
						Message:    "only application-controller may manage Applications",
					},
				},
			},
		},
		Binding: &admissionv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: applicationIdentityBindingName,
				Labels: map[string]string{
					runaiCreatedByLabelKey: runaiOperatorName,
				},
			},
			Spec: admissionv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName:        applicationIdentityPolicyName,
				ValidationActions: []admissionv1.ValidationAction{admissionv1.Audit},
			},
		},
	}
}

func ValidatingAdmissionPolicyStatus(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	return admissionpolicy.StatusOld(ctx, operands.RunaiApplicationControllerName, k8sReader,
		ValidatingAdmissionPolicyDefinition(runaiConfig))
}
