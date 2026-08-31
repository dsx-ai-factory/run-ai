/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDeploymentMarksOwnershipWebhookServingVersion(t *testing.T) {
	runaiConfig := utils.GetDefaultRunaiConfig()
	k8sClient := fake.NewClientBuilder().Build()

	object, err := NewDeploymentResource("anyworkload-controller", k8sClient, runaiConfig).Object(context.Background())
	if err != nil {
		t.Fatalf("build deployment: %v", err)
	}
	deployment := object.(*appsv1.Deployment)
	if deployment.Spec.Template.Labels["run.ai/legacy-ownership-webhook-serving"] != "true" {
		t.Fatalf("pod template does not mark the legacy ownership webhook serving version")
	}
	if deployment.Spec.Template.Labels["run.ai/workloadrunner-ownership-webhook-serving"] != "v1" {
		t.Fatalf("pod template does not mark the WorkloadRunner ownership webhook serving version")
	}
}

func TestBuildEnvVarsIncludesWorkloadInjectedContainerResources(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.SetDefaultsWhereNeeded()
	runaiConfig.Spec.Global.WorkloadInjectedContainers = &global.WorkloadInjectedContainersConfig{
		OAuth2Proxy: &global.OAuth2ProxyContainerConfig{
			Resources: &global.WorkloadInjectedContainerResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")},
				Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("128Mi")},
			},
		},
		UsernameLoader: &global.UsernameLoaderContainerConfig{
			Resources: &global.WorkloadInjectedContainerResourceRequirements{},
		},
		GitSync: &global.GitSyncContainerConfig{
			Resources: &global.WorkloadInjectedContainerResourceRequirements{},
		},
		S3FileSystem: &global.S3FileSystemContainerConfig{
			Resources: &global.WorkloadInjectedContainerResourceRequirements{},
		},
	}
	resource := NewDeploymentResource("anyworkload-controller", nil, runaiConfig)

	envVars, err := resource.buildEnvVars(runaiConfig.Spec.AnyworkloadController, runaiConfig.Spec.Internal)
	if err != nil {
		t.Fatalf("buildEnvVars() error = %v", err)
	}

	assertJSONEnvValue(t, envVars, constants.OAuth2ProxyResourcesEnvVar, `{"requests":{"cpu":"100m"},"limits":{"memory":"128Mi"}}`)
	assertJSONEnvValue(t, envVars, constants.UsernameLoaderResourcesEnvVar, `{}`)
	assertJSONEnvValue(t, envVars, constants.GitSyncResourcesEnvVar, `{}`)
	assertJSONEnvValue(t, envVars, constants.S3FileSystemResourcesEnvVar, `{}`)
}

func TestBuildEnvVarsOmitsUnconfiguredWorkloadInjectedContainerResources(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.SetDefaultsWhereNeeded()
	resource := NewDeploymentResource("anyworkload-controller", nil, runaiConfig)

	envVars, err := resource.buildEnvVars(runaiConfig.Spec.AnyworkloadController, runaiConfig.Spec.Internal)
	if err != nil {
		t.Fatalf("buildEnvVars() error = %v", err)
	}

	for _, name := range []string{
		constants.OAuth2ProxyResourcesEnvVar,
		constants.UsernameLoaderResourcesEnvVar,
		constants.GitSyncResourcesEnvVar,
		constants.S3FileSystemResourcesEnvVar,
	} {
		if value, found := envValue(envVars, name); found {
			t.Fatalf("env var %s unexpectedly configured with %q", name, value)
		}
	}
}

func assertJSONEnvValue(t *testing.T, envVars []corev1.EnvVar, name, expected string) {
	t.Helper()
	actual, found := envValue(envVars, name)
	if !found {
		t.Fatalf("env var %s not found", name)
	}
	var actualJSON, expectedJSON any
	if err := json.Unmarshal([]byte(actual), &actualJSON); err != nil {
		t.Fatalf("env var %s contains invalid JSON %q: %v", name, actual, err)
	}
	if err := json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		t.Fatalf("invalid expected JSON %q: %v", expected, err)
	}
	if !reflect.DeepEqual(actualJSON, expectedJSON) {
		t.Fatalf("env var %s = %q, want %q", name, actual, expected)
	}
}

func envValue(envVars []corev1.EnvVar, name string) (string, bool) {
	for _, envVar := range envVars {
		if envVar.Name == name {
			return envVar.Value, true
		}
	}
	return "", false
}
