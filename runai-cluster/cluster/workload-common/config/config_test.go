/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestOpenshift(t *testing.T) {
	assert.Equal(t, false, IsOpenshift(), "not set")

	t.Setenv(EnvVarOpenshift, "TRUE")
	assert.Equal(t, true, IsOpenshift(), "set to TRUE")

	t.Setenv(EnvVarOpenshift, "true")
	assert.Equal(t, true, IsOpenshift(), "set to true")

	t.Setenv(EnvVarOpenshift, "FaLsE")
	assert.Equal(t, false, IsOpenshift(), "set to FaLsE")
}

func TestIngressEnabled(t *testing.T) {
	assert.Equal(t, true, IngressEnabled(), "defaults to true when not set")

	t.Setenv(EnvVarIngressEnabled, "true")
	assert.Equal(t, true, IngressEnabled(), "true when set to true")

	t.Setenv(EnvVarIngressEnabled, "TRUE")
	assert.Equal(t, true, IngressEnabled(), "true when set to TRUE")

	t.Setenv(EnvVarIngressEnabled, "false")
	assert.Equal(t, false, IngressEnabled(), "false when set to false")

	t.Setenv(EnvVarIngressEnabled, "")
	assert.Equal(t, true, IngressEnabled(), "defaults to true when set to empty string")
}

func TestGatewayEnabled(t *testing.T) {
	assert.Equal(t, false, GatewayEnabled(), "false when not set")

	t.Setenv(EnvVarGatewayAPIEnabled, "true")
	assert.Equal(t, true, GatewayEnabled(), "true when set to true")

	t.Setenv(EnvVarGatewayAPIEnabled, "TRUE")
	assert.Equal(t, true, GatewayEnabled(), "true when set to TRUE")

	t.Setenv(EnvVarGatewayAPIEnabled, "false")
	assert.Equal(t, false, GatewayEnabled(), "false when set to false")

	t.Setenv(EnvVarGatewayAPIEnabled, "")
	assert.Equal(t, false, GatewayEnabled(), "false when set to empty string")
}

func TestGatewayName(t *testing.T) {
	assert.Equal(t, "", GatewayName(), "empty when not set")

	t.Setenv(EnvVarGatewayName, "runai-gateway")
	assert.Equal(t, "runai-gateway", GatewayName(), "returns value when set")
}

func TestGatewayNamespace(t *testing.T) {
	assert.Equal(t, "", GatewayNamespace(), "empty when not set")

	t.Setenv(EnvVarGatewayNamespace, "runai")
	assert.Equal(t, "runai", GatewayNamespace(), "returns value when set")
}

func TestParseWorkloadInjectedContainerResources(t *testing.T) {
	const envVar = "TEST_WORKLOAD_INJECTED_CONTAINER_RESOURCES"

	t.Run("distinguishes unset", func(t *testing.T) {
		require.NoError(t, os.Unsetenv(envVar))
		resources, configured, err := ParseWorkloadInjectedContainerResources(envVar)
		require.NoError(t, err)
		assert.False(t, configured)
		assert.Equal(t, corev1.ResourceRequirements{}, resources)
	})

	t.Run("treats an empty value as unset", func(t *testing.T) {
		t.Setenv(envVar, "")
		resources, configured, err := ParseWorkloadInjectedContainerResources(envVar)
		require.NoError(t, err)
		assert.False(t, configured)
		assert.Equal(t, corev1.ResourceRequirements{}, resources)
	})

	t.Run("preserves explicit empty", func(t *testing.T) {
		t.Setenv(envVar, "{}")
		resources, configured, err := ParseWorkloadInjectedContainerResources(envVar)
		require.NoError(t, err)
		assert.True(t, configured)
		assert.Equal(t, corev1.ResourceRequirements{}, resources)
	})

	t.Run("parses requests and limits", func(t *testing.T) {
		t.Setenv(envVar, `{"requests":{"cpu":"50m"},"limits":{"memory":"100Mi"}}`)
		resources, configured, err := ParseWorkloadInjectedContainerResources(envVar)
		require.NoError(t, err)
		assert.True(t, configured)
		assert.Equal(t, resource.MustParse("50m"), resources.Requests[corev1.ResourceCPU])
		assert.Equal(t, resource.MustParse("100Mi"), resources.Limits[corev1.ResourceMemory])
	})

	t.Run("rejects malformed or invalid requirements", func(t *testing.T) {
		for _, value := range []string{
			`{`,
			`{"requests":{"cpu":"not-a-quantity"}}`,
			`{"claims":[]}`,
			`{"limits":{"cpu":"-1m"}}`,
			`{"requests":{"memory":"256Mi"},"limits":{"memory":"128Mi"}}`,
			`{} {}`,
		} {
			t.Setenv(envVar, value)
			_, configured, err := ParseWorkloadInjectedContainerResources(envVar)
			assert.True(t, configured)
			assert.Error(t, err)
		}
	})
}
