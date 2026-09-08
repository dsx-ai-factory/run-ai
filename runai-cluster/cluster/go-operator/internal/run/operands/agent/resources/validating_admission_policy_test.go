/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAssetIdentityAdmissionPolicyDefinitions(t *testing.T) {
	runaiConfig := newAssetIdentityRunaiConfig("custom-runai")
	definitions := AssetIdentityAdmissionPolicyDefinitions(runaiConfig)
	require.Len(t, definitions, 2)

	tests := []struct {
		name                   string
		definitionIndex        int
		policyName             string
		resource               string
		allowedServiceAccounts string
		deniedMessage          string
	}{
		{
			name:            "Secret",
			definitionIndex: 0,
			policyName:      "runai-secret-identity",
			resource:        "secrets",
			allowedServiceAccounts: `[` +
				`"system:serviceaccount:custom-runai:runai-agent", ` +
				`"system:serviceaccount:custom-runai:shared-objects-controller", ` +
				`"system:serviceaccount:custom-runai:anyworkload-controller"` +
				`]`,
			deniedMessage: "only Run:ai-managed service accounts may modify asset Secrets",
		},
		{
			name:            "PersistentVolumeClaim",
			definitionIndex: 1,
			policyName:      "runai-pvc-identity",
			resource:        "persistentvolumeclaims",
			allowedServiceAccounts: `[` +
				`"system:serviceaccount:custom-runai:runai-agent", ` +
				`"system:serviceaccount:custom-runai:shared-objects-controller"` +
				`]`,
			deniedMessage: "only Run:ai-managed service accounts may modify asset PersistentVolumeClaims",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definition := definitions[tt.definitionIndex]
			policy := definition.Policy
			binding := definition.Binding
			require.NotNil(t, policy)
			require.NotNil(t, binding)

			assert.Equal(t, tt.policyName, policy.Name)
			assert.Equal(t, tt.policyName+"-binding", binding.Name)
			assert.Equal(t, tt.policyName, binding.Spec.PolicyName)
			assert.Empty(t, policy.Namespace)
			assert.Empty(t, binding.Namespace)
			assert.Equal(t, map[string]string{
				"run.ai/created-by": "runai-operator",
			}, policy.Labels)
			assert.Equal(t, policy.Labels, binding.Labels)

			require.NotNil(t, policy.Spec.FailurePolicy)
			assert.Equal(t, admissionv1.Ignore, *policy.Spec.FailurePolicy)
			require.NotNil(t, policy.Spec.MatchConstraints)
			require.Len(t, policy.Spec.MatchConstraints.ResourceRules, 1)
			rule := policy.Spec.MatchConstraints.ResourceRules[0]
			assert.Equal(t, []admissionv1.OperationType{
				admissionv1.Create,
				admissionv1.Update,
				admissionv1.Delete,
			}, rule.Operations)
			assert.Equal(t, []string{"*"}, rule.APIGroups)
			assert.Equal(t, []string{"*"}, rule.APIVersions)
			assert.Equal(t, []string{tt.resource}, rule.Resources)
			require.NotNil(t, rule.Scope)
			assert.Equal(t, admissionv1.NamespacedScope, *rule.Scope)

			require.Len(t, policy.Spec.Variables, 1)
			assert.Equal(t, allowedServiceAccountsVariable, policy.Spec.Variables[0].Name)
			assert.Equal(t, tt.allowedServiceAccounts, policy.Spec.Variables[0].Expression)
			require.Len(t, policy.Spec.Validations, 1)
			assert.Equal(t,
				"request.userInfo.username in variables.allowedServiceAccounts",
				policy.Spec.Validations[0].Expression)
			assert.Equal(t, tt.deniedMessage, policy.Spec.Validations[0].Message)

			assert.Equal(t, []admissionv1.ValidationAction{admissionv1.Audit}, binding.Spec.ValidationActions)
			require.NotNil(t, binding.Spec.MatchResources)
			require.NotNil(t, binding.Spec.MatchResources.ObjectSelector)
			require.Len(t, binding.Spec.MatchResources.ObjectSelector.MatchExpressions, 1)
			selector := binding.Spec.MatchResources.ObjectSelector.MatchExpressions[0]
			assert.Equal(t, constants.RunaiAssetLabel, selector.Key)
			assert.Equal(t, metav1.LabelSelectorOpExists, selector.Operator)
			assert.Empty(t, selector.Values)
		})
	}
}

func TestAssetIdentityAdmissionPolicyDefinitionsUseEffectiveNMCServiceAccount(t *testing.T) {
	runaiConfig := newAssetIdentityRunaiConfig("nmc-runai")
	runaiConfig.Spec.Internal.Global.NMC.ServiceAccountName = "nmc-service-account"

	definitions := AssetIdentityAdmissionPolicyDefinitions(runaiConfig)

	assert.Equal(t,
		`["system:serviceaccount:nmc-runai:nmc-service-account", `+
			`"system:serviceaccount:nmc-runai:anyworkload-controller"]`,
		definitions[0].Policy.Spec.Variables[0].Expression)
	assert.Equal(t,
		`["system:serviceaccount:nmc-runai:nmc-service-account"]`,
		definitions[1].Policy.Spec.Variables[0].Expression)
}

func newAssetIdentityRunaiConfig(namespace string) *runv1.RunaiConfig {
	runaiConfig := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: namespace}}
	runaiConfig.SetDefaultsWhereNeeded()
	return runaiConfig
}
