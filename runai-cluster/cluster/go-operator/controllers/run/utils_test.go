/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/jarcoal/httpmock"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	versioninfo "k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	clienttesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCheckKubeVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectedErr string
	}{
		{name: "minimum", version: "v1.30.0"},
		{name: "newer", version: "v1.31.2"},
		{name: "vendor suffix", version: "v1.30.0-eks-123456"},
		{name: "too old", version: "v1.29.9", expectedErr: "kubernetes version v1.29.9 is not supported; minimum required version is 1.30.0"},
		{name: "invalid", version: "not-a-version", expectedErr: "failed to check server version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkKubeVersion(fakeDiscoveryWithVersion(tt.version))
			if tt.expectedErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func TestCheckKubeVersionsIdentifiesCluster(t *testing.T) {
	t.Run("local cluster", func(t *testing.T) {
		err := checkKubeVersions(fakeDiscoveryWithVersion("v1.29.9"), fakeDiscoveryWithVersion("v1.30.0"))
		assert.ErrorContains(t, err, "local cluster: kubernetes version v1.29.9 is not supported")
	})

	t.Run("remote cluster", func(t *testing.T) {
		err := checkKubeVersions(fakeDiscoveryWithVersion("v1.30.0"), fakeDiscoveryWithVersion("v1.29.9"))
		assert.ErrorContains(t, err, "remote cluster: kubernetes version v1.29.9 is not supported")
	})

	t.Run("both clusters", func(t *testing.T) {
		err := checkKubeVersions(fakeDiscoveryWithVersion("v1.30.0"), fakeDiscoveryWithVersion("v1.31.0"))
		assert.NoError(t, err)
	})
}

func fakeDiscoveryWithVersion(version string) *fakediscovery.FakeDiscovery {
	return &fakediscovery.FakeDiscovery{
		Fake: &clienttesting.Fake{},
		FakedServerVersion: &versioninfo.Info{
			GitVersion: version,
		},
	}
}

func TestIsVAPSupported(t *testing.T) {
	tests := []struct {
		name      string
		resources []*metav1.APIResourceList
		expected  bool
	}{
		{
			name: "policy kind served",
			resources: []*metav1.APIResourceList{{
				GroupVersion: "admissionregistration.k8s.io/v1",
				APIResources: []metav1.APIResource{{Kind: "ValidatingAdmissionPolicy"}},
			}},
			expected: true,
		},
		{
			// pre-1.30: group version exists (webhook configurations) but VAP is not GA
			name: "group version without the policy kind",
			resources: []*metav1.APIResourceList{{
				GroupVersion: "admissionregistration.k8s.io/v1",
				APIResources: []metav1.APIResource{{Kind: "ValidatingWebhookConfiguration"}},
			}},
			expected: false,
		},
		{name: "group version missing", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dClient := &fakediscovery.FakeDiscovery{Fake: &clienttesting.Fake{Resources: tt.resources}}
			supported, err := isVAPSupported(dClient)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, supported)
		})
	}
}

func TestBoolToFloat64(t *testing.T) {
	assert.Equal(t, 1.0, boolToFloat64(true), "Expected 1.0 for true input")
	assert.Equal(t, 0.0, boolToFloat64(false), "Expected 0.0 for false input")
}

func TestResourcesForOperands_VerifyRunaiConfigDoesNotChange(t *testing.T) {
	os.Setenv("IMG_DEFAULT_VERSION", BuildVersion)
	os.Setenv("IMG_DEFAULT_REGISTRY", Registry)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "http://100.0.0.127:12345/metrics", func(request *http.Request) (*http.Response, error) {
		resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
			"DCGM_FI_DEV_FB_USED": 1, "DCGM_FI_DEV_FB_FREE": 1, "DCGM_FI_DEV_GPU_UTIL": 1,
		})
		return resp, err
	})
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)

	// Add mock Prometheus Operator deployment
	promOperator := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consts.PrometheusOperatorDeploymentName,
			Namespace: consts.MonitoringNamespace,
		},
	}
	fc.Create(apiContext, promOperator)

	kbClient := kc.New(fc, fc, nil)
	runaiConfig := getRunaiConfig()
	runaiConfig.SetDefaultsWhereNeeded()
	runaiConfig.Spec.Prometheus.Spec.RemoteWrite = []monitoringv1.RemoteWriteSpec{
		{
			URL: "http://prometheus-operator.monitoring.svc:9090/api/v1/write",
			BasicAuth: &monitoringv1.BasicAuth{
				Username: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: consts.PrometheusOperatorDeploymentName,
					},
				},
			},
		},
	}
	runaiConfigClone := runaiConfig.DeepCopy()
	mo := NewManagedOperands(apiContext, kbClient, kbClient, &runaiConfig)
	operandsSlice := append(mo.GetLocalClusterOperands(), mo.GetRemoteClusterOperands()...)
	ResourcesForOperands(apiContext, operandsSlice)
	assert.Equal(t, runaiConfig, *runaiConfigClone)
}

func TestManagedOperands_WorkloadsAuthSurvivesReconciliation(t *testing.T) {
	runaiConfig := utils.GetDefaultRunaiConfig()

	workloadsAuth := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.WorkloadAuthSecretName,
			Namespace: runaiConfig.GetNamespace(),
			UID:       "existing-uid",
			Labels: map[string]string{
				"existing-label": "existing-value",
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "run.ai/v1",
				Kind:       "RunaiConfig",
				Name:       "runai",
				UID:        "runai-config-uid",
			}},
		},
		Data: map[string][]byte{
			"OAUTH2_PROXY_CLIENT_ID":               []byte("client-id"),
			constants.WorkloadAuthClientSecretKey:  []byte("client-secret"),
			constants.WorkloadAuthCookieSecretKey:  []byte("cookie-secret"),
			constants.WorkloadAuthRedisPasswordKey: []byte("redis-password"),
			"existing-key":                         []byte("existing-data"),
		},
	}
	redisSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster-redis", Namespace: runaiConfig.GetNamespace()},
		Data: map[string][]byte{
			constants.WorkloadAuthRedisPasswordKey: []byte("redis-password"),
		},
	}
	k8sClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(workloadsAuth, redisSecret).Build()
	hybridClient := kc.New(k8sClient, k8sClient, nil)
	managed := NewManagedOperands(context.Background(), hybridClient, hybridClient, runaiConfig)

	anyworkloadController := findOperandByName(managed.GetRemoteClusterOperands(), "anyworkload-controller")
	assert.NotNil(t, anyworkloadController)

	createdOrUpdated := findResourceByName(anyworkloadController.Resources(), constants.WorkloadAuthSecretName)
	if !assert.NotNil(t, createdOrUpdated) {
		return
	}
	assert.Equal(t, "anyworkload-controller", createdOrUpdated.OperandName())

	object, err := createdOrUpdated.Object(context.Background())
	if !assert.NoError(t, err) {
		return
	}
	secret := object.(*v1.Secret)
	assert.Equal(t, workloadsAuth.UID, secret.UID)
	assert.Equal(t, []byte("client-secret"), secret.Data[constants.WorkloadAuthClientSecretKey])
	assert.Equal(t, []byte("cookie-secret"), secret.Data[constants.WorkloadAuthCookieSecretKey])
	assert.Equal(t, []byte("existing-data"), secret.Data["existing-key"])
	assert.Equal(t, "existing-value", secret.Labels["existing-label"])
	assert.Equal(t, workloadsAuth.OwnerReferences, secret.OwnerReferences)
}

func findResourceByName(resources []operands.Resource, name string) operands.Resource {
	for _, resource := range resources {
		object, err := resource.Object(context.Background())
		if err == nil && object.GetName() == name {
			return resource
		}
	}
	return nil
}
