/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy"
	operandsutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	secretIdentityPolicyName = "runai-secret-identity"
	pvcIdentityPolicyName    = "runai-pvc-identity"
	identityBindingSuffix    = "-binding"

	allowedServiceAccountsVariable = "allowedServiceAccounts"
	runaiCreatedByLabelKey         = "run.ai/created-by"
	runaiOperatorName              = "runai-operator"
)

// AssetIdentityAdmissionPolicyDefinitions returns independent policies for the
// two resources protected by the legacy /validate-secrets-pvc webhook.
func AssetIdentityAdmissionPolicyDefinitions(runaiConfig *runv1.RunaiConfig) []admissionpolicy.Definition {
	namespace := runaiConfig.GetNamespace()
	nmcServiceAccount := func(defaultName string) string {
		return operandsutils.GetNMCServiceAccountNameOrDefault(runaiConfig, defaultName)
	}

	// The agent writes source asset Secrets and PVCs. shared-objects-controller
	// replicates both while retaining run.ai labels. Both deployments use the
	// NMC ServiceAccount override when configured.
	sharedAssetWriters := serviceAccountUsernames(namespace,
		nmcServiceAccount(operands.RunaiAgentName),
		nmcServiceAccount(operands.RunaiSharedObjectsControllerName),
	)

	// anyworkload-controller also materializes personal-credential Secrets and
	// retains the source asset label. Its deployment uses its fixed ServiceAccount.
	secretWriters := append([]string{}, sharedAssetWriters...)
	secretWriters = appendUnique(secretWriters,
		serviceAccountUsername(namespace, operands.RunaiAnyworkloadControllerName))

	return []admissionpolicy.Definition{
		assetIdentityAdmissionPolicyDefinition(
			secretIdentityPolicyName,
			"secrets",
			"only Run:ai-managed service accounts may modify asset Secrets",
			secretWriters,
		),
		assetIdentityAdmissionPolicyDefinition(
			pvcIdentityPolicyName,
			"persistentvolumeclaims",
			"only Run:ai-managed service accounts may modify asset PersistentVolumeClaims",
			sharedAssetWriters,
		),
	}
}

func assetIdentityAdmissionPolicyDefinition(
	policyName string,
	resource string,
	deniedMessage string,
	allowedServiceAccounts []string,
) admissionpolicy.Definition {
	labels := map[string]string{runaiCreatedByLabelKey: runaiOperatorName}
	return admissionpolicy.Definition{
		Policy: &admissionv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:   policyName,
				Labels: labels,
			},
			Spec: admissionv1.ValidatingAdmissionPolicySpec{
				FailurePolicy: ptr.To(admissionv1.Ignore),
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
									APIGroups:   []string{"*"},
									APIVersions: []string{"*"},
									Resources:   []string{resource},
									Scope:       ptr.To(admissionv1.NamespacedScope),
								},
							},
						},
					},
				},
				Variables: []admissionv1.Variable{
					{
						Name:       allowedServiceAccountsVariable,
						Expression: celStringList(allowedServiceAccounts),
					},
				},
				Validations: []admissionv1.Validation{
					{
						Expression: fmt.Sprintf(
							"request.userInfo.username in variables.%s",
							allowedServiceAccountsVariable,
						),
						Message: deniedMessage,
					},
				},
			},
		},
		Binding: &admissionv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:   policyName + identityBindingSuffix,
				Labels: labels,
			},
			Spec: admissionv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName:        policyName,
				ValidationActions: []admissionv1.ValidationAction{admissionv1.Audit},
				MatchResources: &admissionv1.MatchResources{
					ObjectSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      constants.RunaiAssetLabel,
								Operator: metav1.LabelSelectorOpExists,
							},
						},
					},
				},
			},
		},
	}
}

func serviceAccountUsernames(namespace string, serviceAccountNames ...string) []string {
	usernames := make([]string, 0, len(serviceAccountNames))
	for _, serviceAccountName := range serviceAccountNames {
		usernames = appendUnique(usernames, serviceAccountUsername(namespace, serviceAccountName))
	}
	return usernames
}

func serviceAccountUsername(namespace, name string) string {
	return fmt.Sprintf("system:serviceaccount:%s:%s", namespace, name)
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func celStringList(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, strconv.Quote(value))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
