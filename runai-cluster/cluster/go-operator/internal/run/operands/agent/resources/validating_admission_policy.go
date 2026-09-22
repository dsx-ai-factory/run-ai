/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// AssetIdentityAdmissionPolicyDefinitions replaces the Secret/PVC webhook with
// the same authorization rules: any ServiceAccount, or a configured username
// substring. Kubernetes RBAC still determines access to the requested resource.
func AssetIdentityAdmissionPolicyDefinitions(runaiConfig *runv1.RunaiConfig) []admissionpolicy.Definition {
	// Match the agent's strings.Join -> splitAndTrim environment-variable path,
	// including comma-separated values inside individual configuration entries.
	var patterns []string
	for _, pattern := range strings.Split(strings.Join(runaiConfig.Spec.Internal.Global.NMC.AllowedServiceAccounts, ","), ",") {
		if pattern = strings.TrimSpace(pattern); pattern != "" {
			patterns = append(patterns, strconv.Quote(pattern))
		}
	}
	var definitions []admissionpolicy.Definition
	for _, asset := range []struct{ name, resource, kind string }{
		{"runai-secret-identity", "secrets", "Secret"},
		{"runai-pvc-identity", "persistentvolumeclaims", "PersistentVolumeClaim"},
	} {
		labels := map[string]string{"run.ai/created-by": "runai-operator"}
		definitions = append(definitions, admissionpolicy.Definition{
			Policy: &admissionv1.ValidatingAdmissionPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: asset.name, Labels: labels},
				Spec: admissionv1.ValidatingAdmissionPolicySpec{
					// Preserve the replaced webhook's handling of evaluation errors.
					// A false validation is still rejected by the Deny binding.
					FailurePolicy: ptr.To(admissionv1.Ignore),
					MatchConstraints: &admissionv1.MatchResources{
						ResourceRules: []admissionv1.NamedRuleWithOperations{{
							RuleWithOperations: admissionv1.RuleWithOperations{
								Operations: []admissionv1.OperationType{admissionv1.Create, admissionv1.Update, admissionv1.Delete},
								Rule: admissionv1.Rule{
									APIGroups: []string{"*"}, APIVersions: []string{"*"},
									Resources: []string{asset.resource}, Scope: ptr.To(admissionv1.NamespacedScope),
								},
							},
						}},
					},
					Variables: []admissionv1.Variable{{Name: "allowedPatterns", Expression: "[" + strings.Join(patterns, ", ") + "]"}},
					Validations: []admissionv1.Validation{{
						Reason: ptr.To(metav1.StatusReasonForbidden),
						Expression: `request.userInfo.username.startsWith("system:serviceaccount:") || ` +
							`variables.allowedPatterns.exists(pattern, request.userInfo.username.contains(pattern))`,
						Message: "username is not permitted to do any action on " + asset.kind + " with the run.ai/asset label",
						MessageExpression: `"username " + request.userInfo.username + ` +
							strconv.Quote(" is not permitted to do any action on "+asset.kind+" with the run.ai/asset label"),
					}},
				},
			},
			Binding: &admissionv1.ValidatingAdmissionPolicyBinding{
				ObjectMeta: metav1.ObjectMeta{Name: asset.name + "-binding", Labels: labels},
				Spec: admissionv1.ValidatingAdmissionPolicyBindingSpec{
					PolicyName: asset.name, ValidationActions: []admissionv1.ValidationAction{admissionv1.Deny},
					MatchResources: &admissionv1.MatchResources{
						ObjectSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{
							Key: constants.RunaiAssetLabel, Operator: metav1.LabelSelectorOpExists,
						}}},
					},
				},
			},
		})
	}
	return definitions
}
