/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"encoding/base64"
	"strings"
	"testing"

	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestCleanAndBase64Encode(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Test with single email",
			input:    "email@run.ai",
			expected: base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{"email@run.ai"}, "\n"))),
		},
		{
			name:     "Test with multiple emails",
			input:    "email@run.ai,email2@run.ai,email3@run.ai",
			expected: base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{"email@run.ai", "email2@run.ai", "email3@run.ai"}, "\n"))),
		},
		{
			name:     "Test with extra commas",
			input:    "email@run.ai,,email2@run.ai,email3@run.ai,",
			expected: base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{"email@run.ai", "email2@run.ai", "email3@run.ai"}, "\n"))),
		},
		{
			name:     "Test with extra spaces",
			input:    "email@run.ai ,email2@run.ai , email3@run.ai",
			expected: base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{"email@run.ai", "email2@run.ai", "email3@run.ai"}, "\n"))),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CleanAndBase64Encode(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, but got %s", tc.expected, result)
			}
		})
	}
}

func TestUsernameLoaderSecurityContext(t *testing.T) {
	t.Run("sets an explicit non-root UID on vanilla Kubernetes", func(t *testing.T) {
		t.Setenv(workloadconfig.EnvVarOpenshift, "false")

		securityContext := UsernameLoaderSecurityContext()

		if securityContext.RunAsNonRoot == nil || !*securityContext.RunAsNonRoot {
			t.Fatal("expected runAsNonRoot to be true")
		}
		if securityContext.RunAsUser == nil || *securityContext.RunAsUser != UsernameLoaderRunAsUser {
			t.Fatalf("expected runAsUser to be %d", UsernameLoaderRunAsUser)
		}
	})

	t.Run("lets the SCC assign an allowed UID on OpenShift", func(t *testing.T) {
		t.Setenv(workloadconfig.EnvVarOpenshift, "true")

		securityContext := UsernameLoaderSecurityContext()

		if securityContext.RunAsNonRoot == nil || !*securityContext.RunAsNonRoot {
			t.Fatal("expected runAsNonRoot to be true")
		}
		if securityContext.RunAsUser != nil {
			t.Fatalf("expected runAsUser to be unset, got %d", *securityContext.RunAsUser)
		}
	})
}

func TestUsernameLoaderResources(t *testing.T) {
	t.Run("keeps legacy defaults when unset", func(t *testing.T) {
		container, err := BuildUsernameLoaderInitContainerWithError("user@example.com", "test")
		if err != nil {
			t.Fatal(err)
		}
		expected := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
		}
		if !container.Resources.Requests[corev1.ResourceCPU].Equal(expected.Requests[corev1.ResourceCPU]) ||
			!container.Resources.Requests[corev1.ResourceMemory].Equal(expected.Requests[corev1.ResourceMemory]) ||
			!container.Resources.Limits[corev1.ResourceCPU].Equal(expected.Limits[corev1.ResourceCPU]) ||
			!container.Resources.Limits[corev1.ResourceMemory].Equal(expected.Limits[corev1.ResourceMemory]) {
			t.Fatalf("expected legacy defaults, got %#v", container.Resources)
		}
	})

	t.Run("applies configured resources from both entry points", func(t *testing.T) {
		t.Setenv(workloadconfig.EnvVarUsernameLoaderResources, `{"requests":{"cpu":"25m"}}`)
		first, err := BuildUsernameLoaderInitContainerWithError("user@example.com", "test")
		if err != nil {
			t.Fatal(err)
		}
		users := "user@example.com"
		second, err := GetUsernameLoaderInitContainerWithError(nil, &users)
		if err != nil {
			t.Fatal(err)
		}
		for _, container := range []*corev1.Container{first, second} {
			if quantity := container.Resources.Requests[corev1.ResourceCPU]; !quantity.Equal(resource.MustParse("25m")) {
				t.Fatalf("expected configured CPU request, got %s", quantity.String())
			}
		}
	})

	t.Run("explicit empty overrides legacy defaults", func(t *testing.T) {
		t.Setenv(workloadconfig.EnvVarUsernameLoaderResources, `{}`)
		container, err := BuildUsernameLoaderInitContainerWithError("user@example.com", "test")
		if err != nil {
			t.Fatal(err)
		}
		if len(container.Resources.Requests) != 0 || len(container.Resources.Limits) != 0 {
			t.Fatalf("expected empty resources, got %#v", container.Resources)
		}
	})

	t.Run("rejects malformed resources", func(t *testing.T) {
		t.Setenv(workloadconfig.EnvVarUsernameLoaderResources, `{"requests":{"cpu":"invalid"}}`)
		container, err := BuildUsernameLoaderInitContainerWithError("user@example.com", "test")
		if err == nil || container != nil {
			t.Fatalf("expected malformed resources to fail, got container=%#v err=%v", container, err)
		}
	})
}
