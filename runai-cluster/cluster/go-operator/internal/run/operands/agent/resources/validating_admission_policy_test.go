/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAssetIdentityAdmissionPolicyDefinitions(t *testing.T) {
	config := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "custom-runai"}}
	config.SetDefaultsWhereNeeded()
	definitions := AssetIdentityAdmissionPolicyDefinitions(config)
	require.Len(t, definitions, 2)
	for i, resource := range []string{"secrets", "persistentvolumeclaims"} {
		policy, binding := definitions[i].Policy, definitions[i].Binding
		assert.Equal(t, []string{"runai-secret-identity", "runai-pvc-identity"}[i], policy.Name)
		assert.Equal(t, policy.Name+"-binding", binding.Name)
		assert.Equal(t, policy.Name, binding.Spec.PolicyName)
		assert.Empty(t, policy.Namespace)
		assert.Empty(t, binding.Namespace)
		assert.Equal(t, map[string]string{"run.ai/created-by": "runai-operator"}, policy.Labels)
		assert.Equal(t, policy.Labels, binding.Labels)
		require.NotNil(t, policy.Spec.FailurePolicy)
		assert.Equal(t, admissionv1.Ignore, *policy.Spec.FailurePolicy)
		require.Len(t, policy.Spec.MatchConstraints.ResourceRules, 1)
		rule := policy.Spec.MatchConstraints.ResourceRules[0]
		assert.Equal(t, []admissionv1.OperationType{admissionv1.Create, admissionv1.Update, admissionv1.Delete}, rule.Operations)
		assert.Equal(t, []string{"*"}, rule.APIGroups)
		assert.Equal(t, []string{"*"}, rule.APIVersions)
		assert.Equal(t, []string{resource}, rule.Resources)
		assert.Equal(t, admissionv1.NamespacedScope, *rule.Scope)
		assert.Equal(t, []admissionv1.ValidationAction{admissionv1.Deny}, binding.Spec.ValidationActions)
		assert.Equal(t, &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key: constants.RunaiAssetLabel, Operator: metav1.LabelSelectorOpExists,
		}}}, binding.Spec.MatchResources.ObjectSelector)
		require.Len(t, policy.Spec.Variables, 1)
		assert.Equal(t, "[]", policy.Spec.Variables[0].Expression)
	}
}

func TestAssetIdentityAuthorization(t *testing.T) {
	environment, err := cel.NewEnv(cel.Variable("request", cel.DynType), cel.Variable("variables", cel.DynType))
	require.NoError(t, err)
	for _, patterns := range []struct {
		name                     string
		configured, allowedUsers []string
	}{
		{name: "no patterns"},
		{name: "legacy configuration parsing", configured: []string{" nmc-ops, trusted-user ", "", "  "}, allowedUsers: []string{"nmc-ops-user", "prefix-trusted-user-suffix"}},
		{name: "escaped pattern", configured: []string{`quote"slash\`}, allowedUsers: []string{`quote"slash\user`}},
	} {
		t.Run(patterns.name, func(t *testing.T) {
			config := &runv1.RunaiConfig{}
			config.SetDefaultsWhereNeeded()
			config.Spec.Internal.Global.NMC.AllowedServiceAccounts = patterns.configured
			for _, definition := range AssetIdentityAdmissionPolicyDefinitions(config) {
				policy := definition.Policy
				patternAST, issues := environment.Compile(policy.Spec.Variables[0].Expression)
				require.NoError(t, issues.Err())
				patternProgram, err := environment.Program(patternAST)
				require.NoError(t, err)
				patternValue, _, err := patternProgram.Eval(map[string]any{})
				require.NoError(t, err)
				ast, issues := environment.Compile(policy.Spec.Validations[0].Expression)
				require.NoError(t, issues.Err())
				program, err := environment.Program(ast)
				require.NoError(t, err)
				for _, username := range []string{
					"system:serviceaccount:runai:runai-agent",
					"system:serviceaccount:tenant-a:unrelated",
					"system:serviceaccount:kube-system:persistent-volume-binder",
					"system:serviceaccount:custom:custom-nmc",
					"system:kube-controller-manager", "system:kube-scheduler",
					"alice", "prefix-system:serviceaccount:default:fake",
					"system:serviceaccount", "nmc-ops-user", "prefix-trusted-user-suffix",
					`quote"slash\user`, "",
				} {
					result, _, err := program.Eval(map[string]any{
						"request":   map[string]any{"userInfo": map[string]any{"username": username}},
						"variables": map[string]any{"allowedPatterns": patternValue},
					})
					require.NoError(t, err)
					// Asset policies retain their existing authorization independently of the stricter CR webhooks.
					allowed := false
					for _, identity := range append([]string{
						"system:serviceaccount:runai:runai-agent",
						"system:serviceaccount:tenant-a:unrelated",
						"system:serviceaccount:kube-system:persistent-volume-binder",
						"system:serviceaccount:custom:custom-nmc",
					}, patterns.allowedUsers...) {
						if username == identity {
							allowed = true
						}
					}
					assert.Equal(t, allowed,
						result.Value(), "policy=%s username=%q", policy.Name, username)
				}
			}
		})
	}
}
