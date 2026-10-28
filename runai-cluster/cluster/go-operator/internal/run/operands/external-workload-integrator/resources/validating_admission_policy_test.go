/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestPodOwnershipPolicyDefinition(t *testing.T) {
	config := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "custom-runai"}}
	config.SetDefaultsWhereNeeded()
	defaultDefinition := PodOwnershipPolicyDefinition(config)
	require.Len(t, defaultDefinition.Policy.Spec.Variables, 1)
	assert.Equal(t, "'system:serviceaccount:custom-runai:external-workload-integrator'", defaultDefinition.Policy.Spec.Variables[0].Expression)
	config.Spec.Internal.Global.NMC.ServiceAccountName = "custom-nmc"
	d := PodOwnershipPolicyDefinition(config)
	require.NotNil(t, d.Policy)
	require.NotNil(t, d.Binding)
	assert.Equal(t, "runai-pod-ownership", d.Policy.Name)
	assert.Equal(t, d.Policy.Name, d.Binding.Spec.PolicyName)
	assert.Equal(t, "runai-operator", d.Policy.Labels["run.ai/created-by"])
	assert.Equal(t, "runai-operator", d.Binding.Labels["run.ai/created-by"])
	assert.Equal(t, admissionv1.Fail, *d.Policy.Spec.FailurePolicy)
	assert.Equal(t, []admissionv1.ValidationAction{admissionv1.Audit}, d.Binding.Spec.ValidationActions)
	require.Len(t, d.Policy.Spec.MatchConstraints.ResourceRules, 1)
	assert.Equal(t, &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
		{Key: "runai/queue", Operator: metav1.LabelSelectorOpExists},
	}}, d.Policy.Spec.MatchConstraints.NamespaceSelector)
	rule := d.Policy.Spec.MatchConstraints.ResourceRules[0]
	assert.Equal(t, []admissionv1.OperationType{admissionv1.Update}, rule.Operations)
	assert.Equal(t, []string{""}, rule.APIGroups)
	assert.Equal(t, []string{"v1"}, rule.APIVersions)
	assert.Equal(t, []string{"pods", "pods/*"}, rule.Resources)
	assert.Equal(t, admissionv1.NamespacedScope, *rule.Scope)
	require.Len(t, d.Policy.Spec.Variables, 1)
	assert.Equal(t, "'system:serviceaccount:custom-runai:custom-nmc'", d.Policy.Spec.Variables[0].Expression)
	require.Len(t, d.Policy.Spec.Validations, 2)
}
