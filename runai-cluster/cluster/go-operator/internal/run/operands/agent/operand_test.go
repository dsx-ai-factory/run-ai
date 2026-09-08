/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package agent

import (
	"context"
	"testing"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
)

func TestOperandOwnsIndependentAssetIdentityPolicies(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{}
	runaiConfig.SetDefaultsWhereNeeded()

	operandResources := Operand(nil, runaiConfig).Resources()
	require.GreaterOrEqual(t, len(operandResources), 4)

	policyNames := map[string]bool{}
	bindingNames := map[string]bool{}
	for _, resource := range operandResources[len(operandResources)-4:] {
		object, err := resource.Object(context.Background())
		require.NoError(t, err)
		switch typedObject := object.(type) {
		case *admissionv1.ValidatingAdmissionPolicy:
			policyNames[typedObject.Name] = true
		case *admissionv1.ValidatingAdmissionPolicyBinding:
			bindingNames[typedObject.Name] = true
		}
	}

	assert.Equal(t, map[string]bool{
		"runai-secret-identity": true,
		"runai-pvc-identity":    true,
	}, policyNames)
	assert.Equal(t, map[string]bool{
		"runai-secret-identity-binding": true,
		"runai-pvc-identity-binding":    true,
	}, bindingNames)
}
