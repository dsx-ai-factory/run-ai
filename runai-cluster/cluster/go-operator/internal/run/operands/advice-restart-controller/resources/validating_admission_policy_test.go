/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestValidatingAdmissionPolicyDefinition(t *testing.T) {
	definition := ValidatingAdmissionPolicyDefinition()
	scheme := runtime.NewScheme()
	require.NoError(t, admissionv1.AddToScheme(scheme))
	reader := fake.NewClientBuilder().WithScheme(scheme).Build()
	policyResources := admissionpolicy.NewResources("advice-restart-controller", reader, definition)

	require.Len(t, policyResources, 2)
	policyObject, err := policyResources[0].Object(context.Background())
	require.NoError(t, err)
	bindingObject, err := policyResources[1].Object(context.Background())
	require.NoError(t, err)

	policy := policyObject.(*admissionv1.ValidatingAdmissionPolicy)
	binding := bindingObject.(*admissionv1.ValidatingAdmissionPolicyBinding)
	assert.Equal(t, vapName, policy.Name)
	assert.Equal(t, vapBindingName, binding.Name)
	assert.Equal(t, policy.Name, binding.Spec.PolicyName)
	assert.Equal(t, []admissionv1.ValidationAction{admissionv1.Deny}, binding.Spec.ValidationActions)

	rules := policy.Spec.MatchConstraints.ResourceRules
	require.Len(t, rules, 1)
	assert.Equal(t, []admissionv1.OperationType{admissionv1.Update}, rules[0].Operations)
	assert.Equal(t, []string{"run.ai"}, rules[0].APIGroups)
	assert.Equal(t, []string{"v1alpha1"}, rules[0].APIVersions)
	assert.Equal(t, []string{"workloadrunners"}, rules[0].Resources)

	// the label lives inside the embedded template
	require.Len(t, policy.Spec.Variables, 2)
	assert.Contains(t, policy.Spec.Variables[0].Expression, "spec.template.metadata.labels")
	assert.Contains(t, policy.Spec.Variables[1].Expression, "spec.template.metadata.labels")
	require.Len(t, policy.Spec.Validations, 2)
	assert.Contains(t, policy.Spec.Validations[0].Expression, heimdallWorkflowIDLabel)
	assert.Contains(t, policy.Spec.Validations[1].Expression, heimdallWorkflowIDLabel)
}
