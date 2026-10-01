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

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	anyworkloadcontroller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/anyworkload-controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runaiv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBuildEnvVarsUsesAnyworkloadOpenShiftRouteConfiguration(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.SetDefaultsWhereNeeded()
	runaiConfig.Spec.AnyworkloadController.OpenshiftRouteConfiguration =
		&anyworkloadcontroller.OpenshiftRouteConfiguration{
			TerminationPolicy: anyworkloadcontroller.OpenshiftRouteTerminationPolicyPassthrough,
		}
	resource := NewDeploymentResource("anyworkload-controller", nil, runaiConfig)

	envVars, err := resource.buildEnvVars(runaiConfig.Spec.AnyworkloadController, runaiConfig.Spec.Internal)
	if err != nil {
		t.Fatalf("buildEnvVars() error = %v", err)
	}

	value, found := envValue(envVars, "OPENSHIFT_ROUTE_TERMINATION_POLICY")
	if !found {
		t.Fatal("OPENSHIFT_ROUTE_TERMINATION_POLICY env var not found")
	}
	if value != "passthrough" {
		t.Fatalf("OPENSHIFT_ROUTE_TERMINATION_POLICY = %q, want %q", value, "passthrough")
	}
}

func TestBuildEnvVarsIncludesNetworkingCapabilityFlags(t *testing.T) {
	tests := []struct {
		name               string
		kubernetesFlavor   runaiv1internal.KubernetesFlavor
		gatewayAPIEnabled  bool
		expectedOpenShift  string
		expectedGatewayAPI string
	}{
		{
			name:               "vanilla with gateway API disabled",
			kubernetesFlavor:   runaiv1internal.KubernetesFlavorVanilla,
			expectedOpenShift:  "false",
			expectedGatewayAPI: "false",
		},
		{
			name:               "vanilla with gateway API enabled",
			kubernetesFlavor:   runaiv1internal.KubernetesFlavorVanilla,
			gatewayAPIEnabled:  true,
			expectedOpenShift:  "false",
			expectedGatewayAPI: "true",
		},
		{
			name:               "OpenShift with gateway API disabled",
			kubernetesFlavor:   runaiv1internal.KubernetesFlavorOpenShift,
			expectedOpenShift:  "true",
			expectedGatewayAPI: "false",
		},
		{
			name:               "OpenShift with gateway API enabled",
			kubernetesFlavor:   runaiv1internal.KubernetesFlavorOpenShift,
			gatewayAPIEnabled:  true,
			expectedOpenShift:  "true",
			expectedGatewayAPI: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runaiConfig := &runaiv1.RunaiConfig{}
			runaiConfig.SetDefaultsWhereNeeded()
			runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = tt.kubernetesFlavor
			runaiConfig.Spec.Global.GatewayAPI.Enabled = &tt.gatewayAPIEnabled
			runaiConfig.Spec.Global.GatewayAPI.Name = "workload-gateway"
			runaiConfig.Spec.Global.GatewayAPI.Namespace = "gateway-system"
			resource := NewDeploymentResource("anyworkload-controller", nil, runaiConfig)

			envVars, err := resource.buildEnvVars(runaiConfig.Spec.AnyworkloadController, runaiConfig.Spec.Internal)
			if err != nil {
				t.Fatalf("buildEnvVars() error = %v", err)
			}

			assertEnvValue(t, envVars, "IS_OPENSHIFT", tt.expectedOpenShift)
			assertEnvValue(t, envVars, "GATEWAY_API_ENABLED", tt.expectedGatewayAPI)
			if tt.gatewayAPIEnabled {
				assertEnvValue(t, envVars, "GATEWAY_NAME", "workload-gateway")
				assertEnvValue(t, envVars, "GATEWAY_NAMESPACE", "gateway-system")
			} else {
				for _, name := range []string{"GATEWAY_NAME", "GATEWAY_NAMESPACE"} {
					if _, found := envValue(envVars, name); found {
						t.Fatalf("env var %s unexpectedly configured while Gateway API is disabled", name)
					}
				}
			}
		})
	}
}

func TestBuildEnvVarsRejectsIncompleteGatewayConfiguration(t *testing.T) {
	for _, tt := range []struct {
		name        string
		gateway     string
		namespace   string
		wantMessage string
	}{
		{
			name: "missing name", namespace: "gateway-system",
			wantMessage: `gatewayAPI is enabled but name="" or namespace="gateway-system" is empty`,
		},
		{
			name: "missing namespace", gateway: "workload-gateway",
			wantMessage: `gatewayAPI is enabled but name="workload-gateway" or namespace="" is empty`,
		},
		{
			name:        "missing both",
			wantMessage: `gatewayAPI is enabled but name="" or namespace="" is empty`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runaiConfig := &runaiv1.RunaiConfig{}
			runaiConfig.SetDefaultsWhereNeeded()
			*runaiConfig.Spec.Global.GatewayAPI.Enabled = true
			runaiConfig.Spec.Global.GatewayAPI.Name = tt.gateway
			runaiConfig.Spec.Global.GatewayAPI.Namespace = tt.namespace
			resource := NewDeploymentResource("anyworkload-controller", nil, runaiConfig)

			_, err := resource.buildEnvVars(runaiConfig.Spec.AnyworkloadController, runaiConfig.Spec.Internal)
			if err == nil {
				t.Fatal("buildEnvVars() accepted an incomplete enabled Gateway configuration")
			}
			if err.Error() != tt.wantMessage {
				t.Fatalf("buildEnvVars() error = %q, want %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestBuildEnvVarsDoesNotInjectIngressAnnotationDefaults(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.SetDefaultsWhereNeeded()
	runaiConfig.Spec.Global.Ingress.Annotations = map[string]string{"example.com/default": "value"}
	resource := NewDeploymentResource("anyworkload-controller", nil, runaiConfig)

	envVars, err := resource.buildEnvVars(runaiConfig.Spec.AnyworkloadController, runaiConfig.Spec.Internal)
	if err != nil {
		t.Fatalf("buildEnvVars() error = %v", err)
	}
	assertEnvValue(t, envVars, "INGRESS_CLASS", runaiConfig.Spec.Global.Ingress.Class)
	if _, found := envValue(envVars, "INGRESS_ANNOTATIONS"); found {
		t.Fatal("INGRESS_ANNOTATIONS must not be enabled by routing validation")
	}
}

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

func assertEnvValue(t *testing.T, envVars []corev1.EnvVar, name, expected string) {
	t.Helper()
	actual, found := envValue(envVars, name)
	if !found {
		t.Fatalf("env var %s not found", name)
	}
	if actual != expected {
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

func TestRestartAllowedServiceAccounts(t *testing.T) {
	for _, tc := range []struct {
		name                string
		agentServiceAccount string
		extra               []string
		want                string
	}{
		{name: "built-in controllers", want: "system:serviceaccount:custom-runai:anyworkload-controller,system:serviceaccount:custom-runai:runai-agent,system:serviceaccount:custom-runai:runtime-rules"},
		{
			name:  "built-in controllers and configured identity",
			extra: []string{"system:serviceaccount:management:restart"},
			want: "system:serviceaccount:custom-runai:anyworkload-controller," +
				"system:serviceaccount:custom-runai:runai-agent," +
				"system:serviceaccount:custom-runai:runtime-rules," +
				"system:serviceaccount:management:restart",
		},
		{name: "NMC agent", agentServiceAccount: "nmc-agent", want: "system:serviceaccount:custom-runai:anyworkload-controller,system:serviceaccount:custom-runai:nmc-agent,system:serviceaccount:custom-runai:runtime-rules"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := &runaiv1.RunaiConfig{}
			config.SetDefaultsWhereNeeded()
			config.Namespace = "custom-runai"
			config.Spec.Internal.Global.NMC.ServiceAccountName = tc.agentServiceAccount
			config.Spec.Internal.Global.NMC.AllowedServiceAccounts = tc.extra
			resource := NewDeploymentResource("anyworkload-controller", nil, config)
			variables, err := resource.buildEnvVars(config.Spec.AnyworkloadController, config.Spec.Internal)
			if err != nil {
				t.Fatal(err)
			}
			value, found := envValue(variables, "RESTART_ALLOWED_SERVICE_ACCOUNTS")
			if !found || value != tc.want {
				t.Fatalf("restart identities = %q, want %q", value, tc.want)
			}
		})
	}
}
