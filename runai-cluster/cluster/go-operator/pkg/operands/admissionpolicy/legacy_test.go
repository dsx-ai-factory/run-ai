/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package admissionpolicy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
)

func TestNewResourcesOld(t *testing.T) {
	definition := validDefinition()

	resources, err := NewResourcesOld(context.Background(), operandName, fakeReader(t), definition)

	require.NoError(t, err)
	require.Len(t, resources, 2)
	assert.Equal(t, operandName, resources[0].OperandName)
	assert.Equal(t, operandName, resources[1].OperandName)
	assert.True(t, resources[0].Enabled)
	assert.True(t, resources[1].Enabled)
	assert.IsType(t, &admissionv1.ValidatingAdmissionPolicy{}, resources[0].Object)
	assert.IsType(t, &admissionv1.ValidatingAdmissionPolicyBinding{}, resources[1].Object)
}

func TestNewResourcesOldPreservesEnabledCallback(t *testing.T) {
	definition := validDefinition()
	definition.Enabled = func() bool { return false }

	resources, err := NewResourcesOld(context.Background(), operandName, fakeReader(t), definition)

	require.NoError(t, err)
	require.Len(t, resources, 2)
	assert.False(t, resources[0].Enabled)
	assert.False(t, resources[1].Enabled)
}

func TestStatusOldReportsMissingPolicy(t *testing.T) {
	definition := validDefinition()

	status, err := StatusOld(context.Background(), operandName, fakeReader(t, definition.Binding), definition)

	require.Error(t, err)
	assert.False(t, status.Ready)
	assert.Contains(t, status.Reasons, "validating admission policy [runai-test-policy] is missing")
}
