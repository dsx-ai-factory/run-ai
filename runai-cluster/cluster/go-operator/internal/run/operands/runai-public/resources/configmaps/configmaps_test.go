/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configmaps

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cpclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/jarcoal/httpmock"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	prometheusv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus"
	internalv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/version"

	kcfake "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client/fake"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ContextMock struct{}

func (cm *ContextMock) Deadline() (deadline time.Time, ok bool) {
	return time.Now(), true
}
func (cm *ContextMock) Done() <-chan struct{} {
	return make(<-chan struct{})
}
func (cm *ContextMock) Err() error {
	return nil
}
func (cm *ContextMock) Value(key any) any {
	return nil
}

var (
	clusterVersion = "test-version"
	testNamespace  = runv1.DefaultNamespace
)

func TestRunaiPublic(t *testing.T) {
	kbClient := createKubeClientMock(getTrainingOperatorCm())
	dClient := createDynamicClient()
	kc := kcfake.New(kbClient, kbClient, dClient)
	runaiConfig := getRunaiConfig()
	runai_public := NewRunaiPublic("test", kc, runaiConfig)
	object, err := runai_public.Object(&ContextMock{})
	if err != nil {
		t.Errorf("failed to generate runai-public config %s", err.Error())
	}
	configMap := object.(*corev1.ConfigMap)
	assert.Equal(t, configMap.Name, "runai-public")
	assert.Equal(t, configMap.Namespace, testNamespace)
	assert.Equal(t, configMap.Data["cluster-version"], clusterVersion)

	out := v1.RunaiPublic{}

	yaml.Unmarshal([]byte(configMap.Data["runai-public"]), &out)

	assert.Equal(t, out.Version, clusterVersion)
	assert.NotEmpty(t, out.Dependencies.Required["gpu-stack"])
	assert.NotEmpty(t, out.Dependencies.Required["gpu-stack"].Components)
	assert.NotEmpty(t, out.Dependencies.Required["gpu-stack"].Components["nvidia"])
	assert.NotEmpty(t, out.Dependencies.Optional["mpi"])
	assert.NotEmpty(t, out.Dependencies.Optional["knative"])
	assert.NotEmpty(t, out.Dependencies.Optional["tensorflow"])
	assert.NotEmpty(t, out.Dependencies.Optional["tensorflow"].Components["suspend"])
	assert.NotEmpty(t, out.Dependencies.Optional["pytorch"])
	assert.NotEmpty(t, out.Dependencies.Optional["pytorch"].Components["suspend"])
	assert.NotEmpty(t, out.Dependencies.Optional["xgboost"])
	assert.NotEmpty(t, out.Dependencies.Optional["xgboost"].Components["suspend"])
	assert.NotNil(t, out.Configuration)
	assert.Equal(t, true, *out.Configuration.IngressEnabled)
	assert.Equal(t, false, *out.Configuration.GatewayAPIEnabled)
}

func TestRunaiPublicRejectsInvalidInjectedContainerResources(t *testing.T) {
	runaiConfig := getRunaiConfig()
	runaiConfig.Spec.Global.WorkloadInjectedContainers = &global.WorkloadInjectedContainersConfig{
		OAuth2Proxy: &global.OAuth2ProxyContainerConfig{
			Resources: &global.WorkloadInjectedContainerResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("-1m")},
			},
		},
	}
	kbClient := createKubeClientMock(getTrainingOperatorCm())
	dClient := createDynamicClient()
	kc := kcfake.New(kbClient, kbClient, dClient)

	_, err := NewRunaiPublic("test", kc, runaiConfig).Object(&ContextMock{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "oauth2Proxy resources")
	assert.Contains(t, err.Error(), "must not be negative")
}

func TestRunaiPublic_InitContainerConstructorInputs(t *testing.T) {
	runaiConfig := getRunaiConfig()

	aw := runaiConfig.Spec.AnyworkloadController
	aw.GitSyncImage = &common.Image{
		Registry: ptr.To("reg.example.com"),
		Name:     ptr.To("git-sync"),
		Tag:      ptr.To("v1"),
	}
	aw.S3FileSystemImage = &common.Image{
		Registry: ptr.To("reg.example.com"),
		Name:     ptr.To("goofys"),
		Tag:      ptr.To("v2"),
	}
	aw.S3FileSystemMountCheckEnabled = ptr.To(true)
	aw.RunaiOauth2ProxyImage = &common.Image{
		Registry: ptr.To("reg.example.com"),
		Name:     ptr.To("oauth2-proxy"),
		Tag:      ptr.To("v3"),
	}
	runaiConfig.Spec.Global.WorkloadInjectedContainers = &global.WorkloadInjectedContainersConfig{
		OAuth2Proxy: &global.OAuth2ProxyContainerConfig{
			Resources: &global.WorkloadInjectedContainerResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
		UsernameLoader: &global.UsernameLoaderContainerConfig{
			Resources: &global.WorkloadInjectedContainerResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("25m")},
			},
		},
		GitSync: &global.GitSyncContainerConfig{
			Resources: &global.WorkloadInjectedContainerResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("64Mi")},
			},
		},
		S3FileSystem: &global.S3FileSystemContainerConfig{
			Resources: &global.WorkloadInjectedContainerResourceRequirements{},
		},
	}
	aw.RunaiOauth2ProxyUsernameLoaderImage = &common.Image{
		Registry: ptr.To("reg.example.com"),
		Name:     ptr.To("username-loader"),
		Tag:      ptr.To("v4"),
	}
	aw.RunaiOauth2ProxyRedisFQDN = ptr.To("redis.example.com:6379")
	aw.RunaiOauth2ProxyWorkloadAuthSecretName = ptr.To("workload-auth-secret")
	aw.RunaiOauth2ProxySessionStoreType = ptr.To("cookie")
	aw.ExternalAuthUrlEnabled = ptr.To(false)

	runaiConfig.Spec.Global.CustomCAGit = &common.CustomCA{
		Enabled: ptr.To(true),
		Secret: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "ca-git"},
			Key:                  "ca.crt",
			Optional:             ptr.To(false),
		},
	}
	runaiConfig.Spec.Global.CustomCAS3 = &common.CustomCA{
		Enabled: ptr.To(true),
		Secret: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "ca-s3"},
			Key:                  "ca.crt",
			Optional:             ptr.To(true),
		},
	}
	runaiConfig.Spec.Internal.Global.CustomCA = &common.CustomCA{
		Enabled: ptr.To(false),
	}
	runaiConfig.Spec.Internal.Global.ControlPlane.Config.SetResearcherService(cpclient.ClusterControlPlaneConfigResearcherService{
		OauthProxy: &cpclient.ClusterControlPlaneConfigResearcherServiceOauthProxy{
			OidcIssuerUrl: ptr.To("https://cp.example.com/auth/realms/runai"),
		},
	})

	kbClient := createKubeClientMock(getTrainingOperatorCm())
	dClient := createDynamicClient()
	kc := kcfake.New(kbClient, kbClient, dClient)
	runaiPublic := NewRunaiPublic("test", kc, runaiConfig)

	object, err := runaiPublic.Object(&ContextMock{})
	assert.NoError(t, err)

	configMap := object.(*corev1.ConfigMap)
	out := v1.RunaiPublic{}
	assert.NoError(t, yaml.Unmarshal([]byte(configMap.Data["runai-public"]), &out))

	cfg := out.Configuration
	assert.NotNil(t, cfg)

	assert.Equal(t, "reg.example.com/git-sync:v1", *cfg.GitSyncImage)
	assert.Equal(t, "reg.example.com/goofys:v2", *cfg.S3FileSystemImage)
	assert.Equal(t, "reg.example.com/goofys:v2", *cfg.S3FileSystemMountCheckImage)
	assert.Equal(t, "reg.example.com/oauth2-proxy:v3", *cfg.RunaiOauth2ProxyImage)
	assert.Equal(t, map[string]string{"cpu": "100m", "memory": "128Mi"}, cfg.RunaiOauth2ProxyResources.Requests)
	assert.Equal(t, map[string]string{"cpu": "100m", "memory": "128Mi"}, cfg.RunaiOauth2ProxyResources.Limits)
	assert.Equal(t, map[string]string{"cpu": "25m"}, cfg.UsernameLoaderResources.Requests)
	assert.Nil(t, cfg.UsernameLoaderResources.Limits)
	assert.Equal(t, map[string]string{"memory": "64Mi"}, cfg.GitSyncResources.Limits)
	assert.Nil(t, cfg.GitSyncResources.Requests)
	assert.NotNil(t, cfg.S3FileSystemResources)
	assert.Nil(t, cfg.S3FileSystemResources.Requests)
	assert.Nil(t, cfg.S3FileSystemResources.Limits)
	assert.Equal(t, "reg.example.com/username-loader:v4", *cfg.RunaiOauth2ProxyUsernameLoaderImage)
	assert.Equal(t, "redis.example.com:6379", *cfg.RunaiOauth2ProxyRedisFQDN)
	assert.Equal(t, "workload-auth-secret", *cfg.RunaiOauth2ProxyWorkloadAuthSecretName)
	assert.Equal(t, "cookie", *cfg.RunaiOauth2ProxySessionStoreType)
	assert.Equal(t, "https://cp.example.com/auth/realms/runai", *cfg.RunaiOauth2ProxyOidcIssuerUrl)
	assert.Equal(t, false, *cfg.ExternalAuthUrlEnabled)

	assert.NotNil(t, cfg.CustomCAGit)
	assert.Equal(t, true, *cfg.CustomCAGit.Enabled)
	assert.Equal(t, "ca-git", *cfg.CustomCAGit.SecretName)
	assert.Equal(t, "ca.crt", *cfg.CustomCAGit.SecretKey)
	assert.Equal(t, false, *cfg.CustomCAGit.Optional)

	assert.NotNil(t, cfg.CustomCAS3)
	assert.Equal(t, true, *cfg.CustomCAS3.Enabled)
	assert.Equal(t, "ca-s3", *cfg.CustomCAS3.SecretName)
	assert.Equal(t, "ca.crt", *cfg.CustomCAS3.SecretKey)
	assert.Equal(t, true, *cfg.CustomCAS3.Optional)

	assert.NotNil(t, cfg.CustomCAOauth2Proxy)
	assert.Equal(t, false, *cfg.CustomCAOauth2Proxy.Enabled)
	assert.Nil(t, cfg.CustomCAOauth2Proxy.SecretName)
	assert.Nil(t, cfg.CustomCAOauth2Proxy.SecretKey)
	assert.Nil(t, cfg.CustomCAOauth2Proxy.Optional)
}

func TestRunaiPublic_ExternalRoutingConfig(t *testing.T) {
	runaiConfig := getRunaiConfig()
	runaiConfig.Spec.Global.Ingress.Enabled = ptr.To(false)
	runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
		Enabled:   ptr.To(true),
		Name:      "my-gateway",
		Namespace: "gateway-system",
	}

	kbClient := createKubeClientMock(getTrainingOperatorCm())
	dClient := createDynamicClient()
	kc := kcfake.New(kbClient, kbClient, dClient)
	runaiPublic := NewRunaiPublic("test", kc, runaiConfig)

	object, err := runaiPublic.Object(&ContextMock{})
	assert.NoError(t, err)

	configMap := object.(*corev1.ConfigMap)
	out := v1.RunaiPublic{}
	assert.NoError(t, yaml.Unmarshal([]byte(configMap.Data["runai-public"]), &out))

	cfg := out.Configuration
	assert.NotNil(t, cfg)
	assert.Equal(t, false, *cfg.IngressEnabled)
	assert.Equal(t, true, *cfg.GatewayAPIEnabled)
	assert.Equal(t, "my-gateway", *cfg.GatewayName)
	assert.Equal(t, "gateway-system", *cfg.GatewayNamespace)
}

func TestRunaiPublic_GPUNetworkAcceleration(t *testing.T) {
	kbClient := createKubeClientMock(getTrainingOperatorCm())
	dClient := createDynamicClient()
	kc := kcfake.New(kbClient, kbClient, dClient)

	// Anyworkload-controller owns the common acceleration setting exposed through
	// both legacy public configuration fields.
	defaultConfig := getRunaiConfig()
	defaultPublic := NewRunaiPublic("test", kc, defaultConfig)
	defaultObject, err := defaultPublic.Object(&ContextMock{})
	assert.NoError(t, err)
	defaultOut := v1.RunaiPublic{}
	assert.NoError(t, yaml.Unmarshal([]byte(defaultObject.(*corev1.ConfigMap).Data["runai-public"]), &defaultOut))
	assert.NotNil(t, defaultOut.Configuration)
	assert.NotNil(t, defaultOut.Configuration.DistGPUNetworkAcceleration)
	assert.Equal(t, false, *defaultOut.Configuration.DistGPUNetworkAcceleration)
	assert.NotNil(t, defaultOut.Configuration.DinfGPUNetworkAcceleration)
	assert.Equal(t, false, *defaultOut.Configuration.DinfGPUNetworkAcceleration)

	enabledConfig := getRunaiConfig()
	enabledConfig.Spec.AnyworkloadController.GPUNetworkAccelerationEnabled = ptr.To(true)
	enabledPublic := NewRunaiPublic("test", kc, enabledConfig)
	distObject, err := enabledPublic.Object(&ContextMock{})
	assert.NoError(t, err)
	distOut := v1.RunaiPublic{}
	assert.NoError(t, yaml.Unmarshal([]byte(distObject.(*corev1.ConfigMap).Data["runai-public"]), &distOut))
	assert.NotNil(t, distOut.Configuration)
	assert.Equal(t, true, *distOut.Configuration.DistGPUNetworkAcceleration)
	assert.Equal(t, true, *distOut.Configuration.DinfGPUNetworkAcceleration)
}

func TestFormatImageRef(t *testing.T) {
	assert.Nil(t, formatImageRef(nil))
	assert.Nil(t, formatImageRef(&common.Image{}))
	assert.Nil(t, formatImageRef(&common.Image{Registry: ptr.To("r"), Name: ptr.To("n")}))

	ref := formatImageRef(&common.Image{
		Registry: ptr.To("r"),
		Name:     ptr.To("n"),
		Tag:      ptr.To("t"),
	})
	assert.NotNil(t, ref)
	assert.Equal(t, "r/n:t", *ref)
}

func TestCustomCAToConfig(t *testing.T) {
	assert.Nil(t, customCAToConfig(nil))

	bare := customCAToConfig(&common.CustomCA{Enabled: ptr.To(false)})
	assert.NotNil(t, bare)
	assert.Equal(t, false, *bare.Enabled)
	assert.Nil(t, bare.SecretName)
	assert.Nil(t, bare.SecretKey)
	assert.Nil(t, bare.Optional)

	full := customCAToConfig(&common.CustomCA{
		Enabled: ptr.To(true),
		Secret: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "s"},
			Key:                  "k",
			Optional:             ptr.To(true),
		},
	})
	assert.NotNil(t, full)
	assert.Equal(t, true, *full.Enabled)
	assert.Equal(t, "s", *full.SecretName)
	assert.Equal(t, "k", *full.SecretKey)
	assert.Equal(t, true, *full.Optional)
}

func TestObject_Platform(t *testing.T) {
	// Prepare
	kbClient := createKubeClientMock()
	dClient := createDynamicClient()
	dClient.FakedServerVersion = &version.Info{
		GitVersion: "v1.18.0",
	}
	kc := kcfake.New(kbClient, kbClient, dClient)
	runaiConfig := getRunaiConfig()
	runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = internalv1.KubernetesFlavorOpenShift

	// Act
	runai_public := NewRunaiPublic("test", kc, runaiConfig)
	object, err := runai_public.Object(&ContextMock{})

	// Assert
	if err != nil {
		t.Errorf("failed to generate runai-public config %s", err.Error())
	}
	configMap := object.(*corev1.ConfigMap)
	out := v1.RunaiPublic{}
	yaml.Unmarshal([]byte(configMap.Data["runai-public"]), &out)
	assert.Equal(t, out.Platform.Type, "openshift")
	assert.Equal(t, out.Platform.KubeVersion, "v1.18.0")
}

func TestGpuStackField(t *testing.T) {
	kbClient := createKubeClientMock()
	status := getGpuStackStatus(&ContextMock{}, kbClient)
	assert.Equal(t, false, *status.Components["nvidia"].Available)
	assert.NotEmpty(t, status.Components["nvidia"].Reason)
	assert.NotEmpty(t, status.Components["nvidia"].Components)
}

func TestNvidiaDevicePlugin_Status(t *testing.T) {
	assert.Equal(t, v1.PublicStatus{
		Available: v1.AvailableStatus(true),
		Reason:    "",
		Components: map[string]v1.PublicStatus{
			"nvidia-metrics": {
				Available: v1.AvailableStatus(false),
				Reason:    getMissingMetricsErrorStr(nil, "some_metric"),
			},
		},
	}, getAvailableNvidiaStatus(nil, "some_metric"))

	assert.Equal(t, v1.PublicStatus{
		Available: v1.AvailableStatus(false),
		Reason:    "failed to connect to cluster",
		Components: map[string]v1.PublicStatus{
			"nvidia-metrics": {
				Available: v1.AvailableStatus(false),
				Reason:    "",
			},
		},
	}, getUnavailableNvidiaStatus(fmt.Errorf("failed to connect to cluster")))

	assert.Equal(t, v1.PublicStatus{
		Available: v1.AvailableStatus(true),
		Reason:    "",
		Components: map[string]v1.PublicStatus{
			"nvidia-metrics": {
				Available: v1.AvailableStatus(false),
				Reason:    "failed to connect to dcgm-exporter service in the cluster",
			},
		},
	}, getAvailableNvidiaStatus(fmt.Errorf("failed to connect to dcgm-exporter service in the cluster"), ""))
}

func getRunaiConfig() v1.RunaiConfig {
	runaiConfig := v1.RunaiConfig{
		Status: v1.RunaiConfigStatus{
			RunaiCustomTypeStatus: v1.RunaiCustomTypeStatus{
				Conditions: []metav1.Condition{
					{
						Type: "stam",
					},
				},
			},
		},
		Spec: v1.RunaiConfigSpec{
			Internal: &internalv1.Internal{
				Global: &internalv1.Global{
					Version: clusterVersion,
				},
			},
		},
	}
	runaiConfig.Namespace = testNamespace

	runaiConfig.SetDefaultsWhereNeeded()
	return runaiConfig
}

func createKubeClientMock(objects ...client.Object) client.Client {
	//mockClient := &mocks.Client{}
	//mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	//mockClient.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	//mockClient.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	builder := fake.ClientBuilder{}
	builder.WithObjects(objects...)
	return builder.Build()
}

func TestIsVersionSupportsOauth2(t *testing.T) {
	assert.NotNil(t, isVersionSupportsOauth2("1.44.0"))
	assert.Nil(t, isVersionSupportsOauth2("2.27.0"))
	assert.Nil(t, isVersionSupportsOauth2("2.44.0"))
}

func TestGetPrometheusStatus_versionOlderMinVersionThatSupportOauth2(t *testing.T) {
	httpmock.Activate()
	httpmock.RegisterResponder("GET", "http://prometheus-operated.runai.svc.cluster.local:9090/api/v1/status/buildinfo", func(request *http.Request) (*http.Response, error) {
		resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
			"status": "success", "data": map[string]interface{}{
				"version": "2.44.0",
			},
		})
		return resp, err
	})

	config := v1.RunaiConfig{Spec: v1.RunaiConfigSpec{
		Global: &global.Global{
			PrometheusService: &global.PrometheusService{
				Address: "http://prometheus-operated.runai.svc.cluster.local:9090",
			},
		},
	}}
	config.SetDefaultsWhereNeeded()

	assert.Equal(t, v1.PublicStatus{
		Available: v1.AvailableStatus(true),
		Reason:    "",
		Components: map[string]v1.PublicStatus{
			"Oauth2": {
				Available: v1.AvailableStatus(true),
				Reason:    "",
			},
		},
	}, getPrometheusStatus(&config))
}
func TestGetPrometheusStatus_versionNotSupportOauth2(t *testing.T) {
	httpmock.Activate()
	httpmock.RegisterResponder("GET", "http://prometheus-operated.runai.svc.cluster.local:9090/api/v1/status/buildinfo", func(request *http.Request) (*http.Response, error) {
		resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
			"status": "success", "data": map[string]interface{}{
				"version": "1.48.0",
			},
		})
		return resp, err
	})

	config := v1.RunaiConfig{Spec: v1.RunaiConfigSpec{
		Global: &global.Global{
			PrometheusService: &global.PrometheusService{
				Address: "http://prometheus-operated.runai.svc.cluster.local:9090",
			},
		},
	}}
	config.SetDefaultsWhereNeeded()

	assert.Equal(t, v1.PublicStatus{
		Available: v1.AvailableStatus(true),
		Reason:    "",
		Components: map[string]v1.PublicStatus{
			"Oauth2": {
				Available: v1.AvailableStatus(false),
				Reason:    "Prometheus version: 1.48.0 - not support oauth2, oauth2 valid in Prometheus versions 2.27.0 and newer",
			},
		},
	}, getPrometheusStatus(&config))
}

func TestGetPrometheusStatus_prometheusVersionError(t *testing.T) {
	httpmock.Activate()
	httpmock.RegisterResponder("GET", "http://prometheus-operated.runai.svc.cluster.local:9090/api/v1/status/buildinfo", func(request *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("connection refused")
	})

	config := v1.RunaiConfig{Spec: v1.RunaiConfigSpec{
		Global: &global.Global{
			PrometheusService: &global.PrometheusService{
				Address: "http://prometheus-operated.runai.svc.cluster.local:9090",
			},
		},
	}}
	config.SetDefaultsWhereNeeded()

	expectedStatus := getPrometheusStatus(&config)
	assert.Equal(t, v1.AvailableStatus(false), expectedStatus.Available)
	assert.Contains(t, expectedStatus.Reason, "failed to connect to prometheus")
	assert.Empty(t, expectedStatus.Components)
}

func TestGetPrometheusStatus_versionExactlyAtMinimum(t *testing.T) {
	httpmock.Activate()
	httpmock.RegisterResponder("GET", "http://prometheus-operated.runai.svc.cluster.local:9090/api/v1/status/buildinfo", func(request *http.Request) (*http.Response, error) {
		resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
			"status": "success", "data": map[string]interface{}{
				"version": "2.27.0",
			},
		})
		return resp, err
	})

	config := v1.RunaiConfig{Spec: v1.RunaiConfigSpec{
		Global: &global.Global{
			PrometheusService: &global.PrometheusService{
				Address: "http://prometheus-operated.runai.svc.cluster.local:9090",
			},
		},
	}}
	config.SetDefaultsWhereNeeded()

	assert.Equal(t, v1.PublicStatus{
		Available: v1.AvailableStatus(true),
		Reason:    "",
		Components: map[string]v1.PublicStatus{
			"Oauth2": {
				Available: v1.AvailableStatus(true),
				Reason:    "",
			},
		},
	}, getPrometheusStatus(&config))
}

func TestGetPrometheusStatus_invalidVersionFormat(t *testing.T) {
	httpmock.Activate()
	httpmock.RegisterResponder("GET", "http://prometheus-operated.runai.svc.cluster.local:9090/api/v1/status/buildinfo", func(request *http.Request) (*http.Response, error) {
		resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
			"status": "success", "data": map[string]interface{}{
				"version": "invalid-version",
			},
		})
		return resp, err
	})

	config := v1.RunaiConfig{Spec: v1.RunaiConfigSpec{
		Global: &global.Global{
			PrometheusService: &global.PrometheusService{
				Address: "http://prometheus-operated.runai.svc.cluster.local:9090",
			},
		},
	}}
	config.SetDefaultsWhereNeeded()

	expectedStatus := getPrometheusStatus(&config)
	assert.Equal(t, v1.AvailableStatus(true), expectedStatus.Available)
	assert.Equal(t, "", expectedStatus.Reason)
	assert.NotEmpty(t, expectedStatus.Components)
	assert.Equal(t, v1.AvailableStatus(false), expectedStatus.Components["Oauth2"].Available)
	assert.Contains(t, expectedStatus.Components["Oauth2"].Reason, "invalid-version")
}

func TestGetPrometheusStatus_noPrometheusServiceConfigured(t *testing.T) {
	// No HTTP mock needed since no service is configured

	config := v1.RunaiConfig{Spec: v1.RunaiConfigSpec{
		// No Global.PrometheusService configured
	}}
	config.SetDefaultsWhereNeeded()

	expectedStatus := getPrometheusStatus(&config)
	assert.Equal(t, v1.AvailableStatus(false), expectedStatus.Available)
	assert.NotEmpty(t, expectedStatus.Reason)
	assert.Empty(t, expectedStatus.Components)
}

func TestGetPrometheusStatus_openshiftBuiltinPrometheus(t *testing.T) {
	// No HTTP mock needed since we bypass the version check

	config := v1.RunaiConfig{
		Spec: v1.RunaiConfigSpec{
			Internal: &internalv1.Internal{
				Global: &internalv1.Global{
					Platform: &internalv1.Platform{
						KubernetesFlavor: internalv1.KubernetesFlavorOpenShift,
					},
				},
			},
		},
	}
	// Set the prometheus config separately after initialization
	config.Spec.Prometheus = &prometheusv1.Prometheus{
		Openshift: prometheusv1.Openshift{
			UseBuiltInPrometheus: ptr.To(true),
		},
	}
	config.SetDefaultsWhereNeeded()

	assert.Equal(t, v1.PublicStatus{
		Available: v1.AvailableStatus(true),
		Reason:    "",
		Components: map[string]v1.PublicStatus{
			"Oauth2": {
				Available: v1.AvailableStatus(true),
				Reason:    "Running on OpenShift using built-in Thanos which always supports OAuth2",
			},
		},
	}, getPrometheusStatus(&config))
}

func TestGetTrainingOperatorSuspendStatus(t *testing.T) {
	const imageSupportsSuspend = "kubeflow/training-operator:v1-04f9f13"
	const imageNotSupportsSuspend = "kubeflow/training-operator:v1-855e096"
	ctx := &ContextMock{}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consts.TrainingOperatorDeploymentName,
			Namespace: consts.TrainingOperatorNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: imageSupportsSuspend},
					},
				},
			},
		},
	}

	kbClient := createKubeClientMock(getTrainingOperatorCm(), deployment)
	status := getTrainingOperatorSuspendStatus(ctx, kbClient)
	assert.Equal(t, *status.Available, true)

	deployment.Spec.Template.Spec.Containers[0].Image = imageNotSupportsSuspend
	assert.Nil(t, kbClient.Update(ctx, deployment))

	status = getTrainingOperatorSuspendStatus(ctx, kbClient)
	assert.Equal(t, *status.Available, false)
}

func TestGetKubeflowTrimmedArgsWithEqual(t *testing.T) {
	deployment := createDeploymentWithArgs([]string{"--enable-scheme=jaxjob"})
	args, _ := getKubeFlowEnabledFrameworks(deployment)
	assert.Equal(t, "jaxjob", args[0])
}
func TestGetKubeflowTrimmedArgsWithUnsupportedFramework(t *testing.T) {
	deployment := createDeploymentWithArgs([]string{"--enable-scheme=blabla"})
	args, _ := getKubeFlowEnabledFrameworks(deployment)
	assert.Equal(t, "blabla", args[0])
}
func TestGetKubeflowTrimmedArgsWithSpace(t *testing.T) {
	deployment := createDeploymentWithArgs([]string{"--enable-scheme jaxjob"})
	args, _ := getKubeFlowEnabledFrameworks(deployment)
	assert.Equal(t, "jaxjob", args[0])
}
func TestGetKubeflowTrimmedArgsNoArgs(t *testing.T) {
	deployment := createDeploymentWithArgs([]string{"--enable jaxjob"})
	args, _ := getKubeFlowEnabledFrameworks(deployment)
	var nilSlice []string = nil
	assert.Equal(t, nilSlice, args)
	assert.Equal(t, len(nilSlice), 0)
}
func TestGetKubeflowTrimmedArgsMultiArgs(t *testing.T) {
	deployment := createDeploymentWithArgs([]string{"--enable-scheme jaxjob", "--enable-scheme=mpijob"})
	args, _ := getKubeFlowEnabledFrameworks(deployment)
	assert.Equal(t, "jaxjob", args[0])
	assert.Equal(t, "mpijob", args[1])
}

var _ = Describe("Snir doesn't like ginkgo, it won't run regular tests without this thing", func() {
})

func createDynamicClient() *fakediscovery.FakeDiscovery {
	fakeClient := fakeclientset.NewSimpleClientset()
	dClient := &fakediscovery.FakeDiscovery{Fake: &fakeClient.Fake}
	return dClient
}

func createFakeClientSet() (kubernetes.Interface, error) {
	fakeClient := fakeclientset.NewSimpleClientset()
	return fakeClient, nil
}

func createDeploymentWithArgs(args []string) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consts.TrainingOperatorDeploymentName,
			Namespace: consts.TrainingOperatorNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Args: args},
					},
				},
			},
		},
	}
	return deployment
}

func getTrainingOperatorCm() client.Object {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "training-operator-tags-config",
			Namespace: v1.DefaultNamespace,
		},
		Data: map[string]string{
			"v1-3f15cb8": "1.9.2",
			"v1-5c0e763": "1.9.1",
			"v1-5170a36": "1.9.0",
			"v1-04f9f13": "1.8.1",
			"v1-9e52eb7": "1.8.0",
			"v1-855e096": "1.7.0",
		},
	}
}
