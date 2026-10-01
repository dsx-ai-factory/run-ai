/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	imageparser "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/common/image-parser"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/installation/util"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	internalv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/jarcoal/httpmock"
	ocpconf "github.com/openshift/api/config/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
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
	ClusterVersion = "test-version"
	Namespace      = "test"
	BuildVersion   = "test_version"
	Registry       = "my_registry"

	DockerImageRegExp = `(?m)^[\n\r\s]+.*:\s*([^\/|\n|:]*)\/(.*?):([^\s|"]*)$`
)

func createEmptyKubeClientMock(apiContext context.Context) client.Client {
	v1alpha1.AddToScheme(scheme.Scheme)
	monitoringv1.AddToScheme(scheme.Scheme)
	clientSet := fakeclientset.NewSimpleClientset()
	namespace := k8sv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: consts.GPUOperatorDeploymentName,
		},
	}
	clientSet.CoreV1().Namespaces().Create(apiContext, &namespace, metav1.CreateOptions{})

	service := k8sv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dcgm-exporter",
			Namespace: consts.GPUOperatorDeploymentName,
			Labels: map[string]string{
				"app": "nvidia-dcgm-exporter",
			},
		},
		Spec: k8sv1.ServiceSpec{
			ClusterIP: "100.0.0.127",
			Ports: []k8sv1.ServicePort{
				{Port: 12345},
			},
		},
	}
	k8sClient := fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(&service, &namespace).Build()
	return k8sClient
}

func TestWatchedResourcesIncludesAdmissionPolicies(t *testing.T) {
	countAdmissionPolicyResources := func(vapSupported bool) (policies, bindings int) {
		for _, resource := range watchedResources(vapSupported) {
			switch resource.(type) {
			case *admissionv1.ValidatingAdmissionPolicy:
				policies++
			case *admissionv1.ValidatingAdmissionPolicyBinding:
				bindings++
			}
		}
		return policies, bindings
	}

	policies, bindings := countAdmissionPolicyResources(true)
	assert.Equal(t, 1, policies)
	assert.Equal(t, 1, bindings)

	policies, bindings = countAdmissionPolicyResources(false)
	assert.Equal(t, 0, policies)
	assert.Equal(t, 0, bindings)
}

func getRunaiConfig() v1.RunaiConfig {
	runaiConfig := v1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runaiv1.DefaultName,
			Namespace: runaiv1.DefaultNamespace,
		},
		Status: v1.RunaiConfigStatus{
			RunaiCustomTypeStatus: v1.RunaiCustomTypeStatus{
				Conditions: []metav1.Condition{
					{
						Type: "runai_test",
					},
				},
			},
		},
		Spec: v1.RunaiConfigSpec{
			Internal: &internalv1.Internal{
				Global: &internalv1.Global{
					Version: ClusterVersion,
				},
			},
		},
	}
	runaiConfig.Namespace = Namespace

	runaiConfig.SetDefaultsWhereNeeded()
	return runaiConfig
}
func TestImages(t *testing.T) {
	os.Setenv("IMG_DEFAULT_VERSION", BuildVersion)
	os.Setenv("IMG_DEFAULT_REGISTRY", Registry)
	t.Setenv("IMG_THIRD_PARTY_GIT_SYNC_REGISTRY", "registry.k8s.io")
	t.Setenv("IMG_THIRD_PARTY_GIT_SYNC_REPOSITORY", "git-sync/git-sync")
	t.Setenv("IMG_THIRD_PARTY_GIT_SYNC_TAG", "v4.7.0")
	t.Setenv("IMG_THIRD_PARTY_OAUTH2_PROXY_REGISTRY", "quay.io")
	t.Setenv("IMG_THIRD_PARTY_OAUTH2_PROXY_REPOSITORY", "oauth2-proxy/oauth2-proxy")
	t.Setenv("IMG_THIRD_PARTY_OAUTH2_PROXY_TAG", "v7.15.4")
	t.Setenv("IMG_THIRD_PARTY_VALKEY_REGISTRY", "docker.io")
	t.Setenv("IMG_THIRD_PARTY_VALKEY_REPOSITORY", "valkey/valkey")
	t.Setenv("IMG_THIRD_PARTY_VALKEY_TAG", "9.0.4")
	t.Setenv("IMG_THIRD_PARTY_WAIT_UBI_MINIMAL_REGISTRY", "registry.access.redhat.com")
	t.Setenv("IMG_THIRD_PARTY_WAIT_UBI_MINIMAL_REPOSITORY", "ubi9/ubi-minimal")
	t.Setenv("IMG_THIRD_PARTY_WAIT_UBI_MINIMAL_TAG", "9.8")

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
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := getRunaiConfig()
	runaiConfig.SetDefaultsWhereNeeded()
	mo := NewManagedOperands(apiContext, kbClient, kbClient, &runaiConfig)
	operandsSlice := append(mo.GetRemoteClusterEnabled(), mo.GetLocalClusterEnabled()...)
	resources, _ := ResourcesForOperands(apiContext, operandsSlice)
	images := []*common.Image{}
	for _, resource := range resources.CreateOrUpdate {
		testedResource, _ := resource.Object(apiContext)
		res, _ := getImagesFromK8sObject(testedResource)
		if len(res) > 0 {
			images = append(images, res...)
		}
	}
	for _, resource := range resources.Delete {
		testedResource, _ := resource.Object(apiContext)
		res, _ := getImagesFromK8sObject(testedResource)
		if len(res) > 0 {
			images = append(images, res...)
		}
	}
	thirdPartyImage := identifyThirdPartImages(images)
	newThirdPartyImagesFoundInOperator := checkNoNewThirdPartyImageInOperator(thirdPartyImage)
	assert.False(t, newThirdPartyImagesFoundInOperator)
}

func checkNoNewThirdPartyImageInOperator(opImageMap map[string]*common.Image) bool {
	imageMap := map[string]bool{}
	for _, name := range []string{"GIT_SYNC", "OAUTH2_PROXY", "VALKEY", "WAIT_UBI_MINIMAL"} {
		image := common.ThirdPartyImage(name)
		imageMap[*image.Name+":"+*image.Tag] = true
	}
	for _, thirdPartyImageInOp := range opImageMap {
		imageMapKey := *thirdPartyImageInOp.Name + ":" + *thirdPartyImageInOp.Tag
		if !imageMap[imageMapKey] {
			return true
		}
	}
	return false
}

func createDynamicClient() *fakediscovery.FakeDiscovery {
	fakeClient := fakeclientset.NewSimpleClientset()
	dClient := &fakediscovery.FakeDiscovery{Fake: &fakeClient.Fake}
	return dClient
}

func getImagesFromK8sObject(st interface{}) ([]*common.Image, error) {
	installationConfigurationYAML, err := yaml.Marshal(st)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal installation configuration: %w", err)
	}
	images, _ := imageparser.ParseMultilineContent(string(installationConfigurationYAML))
	return images, nil
}

func identifyThirdPartImages(images []*common.Image) map[string]*common.Image {
	thirdPartyImage := map[string]*common.Image{}
	for _, image := range images {
		if *image.Tag != BuildVersion {
			key := *image.Registry + ":" + *image.Name + ":" + *image.Tag
			if _, ok := thirdPartyImage[key]; ok {
				continue
			}
			thirdPartyImage[key] = image
		}
	}
	return thirdPartyImage
}

func TestSetOperandStatusMetrics(t *testing.T) {
	operandStatusMap := map[string]v1.TimedOperandStatus{
		"test-operand": {
			OperandStatus: v1.OperandStatus{Ready: true},
		},
	}
	setOperandStatusMetrics(operandStatusMap)

	gauge, err := metrics.OperandStatus.GetMetricWithLabelValues("test-operand", metrics.OperandStatusConditionReady)
	assert.NoError(t, err)
	assert.NotNil(t, gauge)
	// We cannot directly get the value from prometheus.Gauge, but this ensures the metric is set and registered.
}

// --- Mock types for reconcileCluster tests ---

type mockResource struct {
	objectErr   error
	statusErr   error
	enabled     bool
	object      client.Object
	operandName string
}

var _ operands.Resource = &mockResource{}

func (m *mockResource) OperandName() string {
	if m.operandName == "" {
		panic("mockResource: operandName must be set")
	}
	return m.operandName
}
func (m *mockResource) Object(ctx context.Context) (client.Object, error) {
	if m.objectErr != nil {
		return nil, m.objectErr
	}
	if m.object == nil {
		panic("mockResource: object must be set")
	}
	return m.object, nil
}
func (m *mockResource) Status(ctx context.Context) error {
	return m.statusErr
}
func (m *mockResource) Enabled() bool { return m.enabled }

func NewMockResource(objectErr, statusErr error, enabled bool, object client.Object, operandName string) *mockResource {
	return &mockResource{
		objectErr:   objectErr,
		statusErr:   statusErr,
		enabled:     enabled,
		object:      object,
		operandName: operandName,
	}
}

type mockOperand struct {
	resources []operands.Resource
	status    v1.OperandStatus
	statusErr error
	enabled   bool
	name      string
}

var _ operands.Operand = &mockOperand{}

func (m *mockOperand) Name() string {
	if m.name == "" {
		panic("mockOperand: name must be set")
	}
	return m.name
}
func (m *mockOperand) Resources() []operands.Resource {
	return m.resources
}
func (m *mockOperand) Status(ctx context.Context) (v1.OperandStatus, error) {
	return m.status, m.statusErr
}
func (m *mockOperand) Enabled() bool { return m.enabled }

func NewMockOperand(resources []operands.Resource, status v1.OperandStatus, statusErr error, enabled bool, name string) *mockOperand {
	return &mockOperand{
		resources: resources,
		status:    status,
		statusErr: statusErr,
		enabled:   enabled,
		name:      name,
	}
}

// --- End mock types ---

func setupTestEnv(t *testing.T) (context.Context, kc.HybridClient, *runtime.Scheme, v1.RunaiConfig) {
	t.Helper()
	ctx := t.Context()
	s := scheme_init.Scheme()
	fc := fakeClient.NewClientBuilder().WithScheme(s).WithObjects(&runaiv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: runaiv1.ClusterName,
		},
	}).Build()
	k8sClient := kc.New(fc, fc, nil)
	runaiConfig := getRunaiConfig()
	runaiConfig.SetDefaultsWhereNeeded()
	return ctx, k8sClient, s, runaiConfig
}

func TestReconcileCluster_NormalReconciliation_Success(t *testing.T) {
	ctx, k8sClient, s, runaiConfig := setupTestEnv(t)
	mockRes := NewMockResource(
		nil,
		nil,
		true,
		&k8sv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: Namespace,
			},
		},
		"mock",
	)
	testOperands := []operands.Operand{
		NewMockOperand(
			[]operands.Resource{mockRes},
			v1.OperandStatus{Ready: true},
			nil,
			true,
			"mock",
		),
	}
	statusMap, finalized, err := reconcileCluster(
		ctx,
		k8sClient,
		s,
		&runaiConfig,
		testOperands,
		true,
		false,
	)
	assert.NoError(t, err)
	assert.False(t, finalized)
	assert.Contains(t, statusMap, "mock")
	assert.True(t, statusMap["mock"].Ready)
}

func TestReconcileCluster_Finalization_DeletesConfigMap(t *testing.T) {
	ctx, k8sClient, s, runaiConfig := setupTestEnv(t)
	mockRes := NewMockResource(
		nil,
		nil,
		true,
		&k8sv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: Namespace,
			},
		},
		"mock",
	)
	testOperands := []operands.Operand{
		NewMockOperand(
			[]operands.Resource{mockRes},
			v1.OperandStatus{Ready: true},
			nil,
			true,
			"mock",
		),
	}

	// Create the ConfigMap in the fake client before reconciliation
	cmObj := mockRes.object.(*k8sv1.ConfigMap)
	err := k8sClient.Create(ctx, cmObj)
	assert.NoError(t, err, "failed to create ConfigMap before reconciliation")

	_, finalized, err := reconcileCluster(
		ctx,
		k8sClient,
		s,
		&runaiConfig,
		testOperands,
		true,
		true,
	)
	assert.NoError(t, err)
	assert.True(t, finalized)

	// Now check that the ConfigMap was deleted
	cm := &k8sv1.ConfigMap{}
	err = k8sClient.Get(ctx, client.ObjectKey{Name: cmObj.Name, Namespace: cmObj.Namespace}, cm)
	assert.Error(t, err)
	assert.True(t, client.IgnoreNotFound(err) == nil, "expected NotFound error after finalization")
}

func TestReconcileCluster_ObjectGathering_Error(t *testing.T) {
	ctx, k8sClient, s, runaiConfig := setupTestEnv(t)
	mockRes2 := NewMockResource(
		fmt.Errorf("object error"),
		nil,
		true,
		&k8sv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: Namespace,
			},
		},
		"mock",
	)
	mockOperandErr := NewMockOperand(
		[]operands.Resource{mockRes2},
		v1.OperandStatus{Ready: true},
		nil,
		true,
		"mock",
	)
	operandsObjErr := []operands.Operand{mockOperandErr}
	_, _, err := reconcileCluster(
		ctx,
		k8sClient,
		s,
		&runaiConfig,
		operandsObjErr,
		true,
		false,
	)
	assert.Error(t, err)
}

func TestReconcileCluster_OwnerAnnotations_NoError(t *testing.T) {
	ctx, k8sClient, s, runaiConfig := setupTestEnv(t)
	mockRes := NewMockResource(
		nil,
		nil,
		true,
		&k8sv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: Namespace,
			},
		},
		"mock",
	)
	testOperands := []operands.Operand{
		NewMockOperand(
			[]operands.Resource{mockRes},
			v1.OperandStatus{Ready: true},
			nil,
			true,
			"mock",
		),
	}
	_, _, err := reconcileCluster(
		ctx,
		k8sClient,
		s,
		&runaiConfig,
		testOperands,
		true,
		false,
	)
	// This will only fail if the real SetOwnerAnnotations returns an error
	assert.NoError(t, err)
}

func TestReconcileCluster_OwnerReferences_NoError(t *testing.T) {
	ctx, k8sClient, s, runaiConfig := setupTestEnv(t)
	mockRes := NewMockResource(
		nil,
		nil,
		true,
		&k8sv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: Namespace,
			},
		},
		"mock",
	)
	testOperands := []operands.Operand{
		NewMockOperand(
			[]operands.Resource{mockRes},
			v1.OperandStatus{Ready: true},
			nil,
			true,
			"mock",
		),
	}
	_, _, err := reconcileCluster(
		ctx,
		k8sClient,
		s,
		&runaiConfig,
		testOperands,
		true,
		false,
	)
	// This will only fail if the real SetOwnerReferencesForApplicableResources returns an error
	assert.NoError(t, err)
	cm := &k8sv1.ConfigMap{}
	err = k8sClient.Get(ctx, client.ObjectKeyFromObject(mockRes.object), cm)
	assert.NoError(t, err)
	assert.Len(t, cm.GetOwnerReferences(), 2)

	foundRunaiConfig := false
	foundCluster := false
	for _, owner := range cm.GetOwnerReferences() {
		if owner.Name == runaiConfig.Name {
			foundRunaiConfig = true
		}
		if owner.Name == runaiv1.ClusterName {
			foundCluster = true
		}
	}
	assert.True(t, foundRunaiConfig)
	assert.True(t, foundCluster)
}

func TestReconcileCluster_CreateOrUpdateResources_NoError(t *testing.T) {
	ctx, k8sClient, s, runaiConfig := setupTestEnv(t)
	mockRes := NewMockResource(
		nil,
		nil,
		true,
		&k8sv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: Namespace,
			},
		},
		"mock",
	)
	testOperands := []operands.Operand{
		NewMockOperand(
			[]operands.Resource{mockRes},
			v1.OperandStatus{Ready: true},
			nil,
			true,
			"mock",
		),
	}
	_, _, err := reconcileCluster(
		ctx,
		k8sClient,
		s,
		&runaiConfig,
		testOperands,
		true,
		false,
	)
	// This will only fail if the real CreateOrUpdateResources returns an error
	assert.NoError(t, err)
}

func TestReconcileCluster_DeleteResources_NoError(t *testing.T) {
	ctx, k8sClient, s, runaiConfig := setupTestEnv(t)
	mockRes := NewMockResource(
		nil,
		nil,
		true,
		&k8sv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: Namespace,
			},
		},
		"mock",
	)
	testOperands := []operands.Operand{
		NewMockOperand(
			[]operands.Resource{mockRes},
			v1.OperandStatus{Ready: true},
			nil,
			true,
			"mock",
		),
	}
	_, _, err := reconcileCluster(
		ctx,
		k8sClient,
		s,
		&runaiConfig,
		testOperands,
		true,
		false,
	)
	// This will only fail if the real DeleteResources returns an error
	assert.NoError(t, err)
}

func TestReconcileCluster_OperandStatusCalculation_NoError(t *testing.T) {
	ctx, k8sClient, s, runaiConfig := setupTestEnv(t)
	mockRes := NewMockResource(
		nil,
		nil,
		true,
		&k8sv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: Namespace,
			},
		},
		"mock",
	)
	testOperands := []operands.Operand{
		NewMockOperand(
			[]operands.Resource{mockRes},
			v1.OperandStatus{Ready: true},
			nil,
			true,
			"mock",
		),
	}
	_, _, err := reconcileCluster(
		ctx,
		k8sClient,
		s,
		&runaiConfig,
		testOperands,
		true,
		false,
	)
	// This will only fail if the real calculateOperandStatusMap returns an error
	assert.NoError(t, err)
}

// --- Tests for handleDependencyChecks ---

func TestHandleDependencyChecks_SkipChecks_NMCEnabled(t *testing.T) {
	ctx, k8sClient, _, _ := setupTestEnv(t)

	// Create runaiConfig with NMC enabled (which triggers skip)
	clusterRunaiconfig := v1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:       runaiv1.DefaultName,
			Namespace:  runaiv1.DefaultNamespace,
			Generation: 1,
		},
		Spec: v1.RunaiConfigSpec{
			Internal: &internalv1.Internal{
				Global: &internalv1.Global{
					NMC: internalv1.NMC{
						RemoteAPIServerURL: "https://remote-cluster.example.com",
					},
				},
			},
		},
	}

	runaiConfigWithDefaults := clusterRunaiconfig.DeepCopy()
	runaiConfigWithDefaults.SetDefaultsWhereNeeded()

	err := handleDependencyChecks(ctx, k8sClient, &clusterRunaiconfig, runaiConfigWithDefaults)

	assert.NoError(t, err)
	condition := findCondition(clusterRunaiconfig.Status.Conditions, runaiv1.ConditionDependenciesFulfilled)
	assert.NotNil(t, condition)
	if assert.NotNil(t, condition) {
		assert.Equal(t, metav1.ConditionTrue, condition.Status)
		assert.Equal(t, runaiv1.DependenciesFulfilled, v1.ConditionReason(condition.Reason))
		assert.Equal(t, "Dependency checks skipped", condition.Message)
		assert.Equal(t, int64(1), condition.ObservedGeneration)
	}
}

func TestHandleDependencyChecks_SkipChecks_DevelopmentFlag(t *testing.T) {
	ctx, k8sClient, _, _ := setupTestEnv(t)

	// Create runaiConfig with SkipDependencyChecks enabled
	clusterRunaiconfig := v1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:       runaiv1.DefaultName,
			Namespace:  runaiv1.DefaultNamespace,
			Generation: 2,
		},
		Spec: v1.RunaiConfigSpec{
			Internal: &internalv1.Internal{
				Development: &internalv1.Development{
					SkipDependencyChecks: true,
				},
			},
		},
	}

	runaiConfigWithDefaults := clusterRunaiconfig.DeepCopy()
	runaiConfigWithDefaults.SetDefaultsWhereNeeded()

	err := handleDependencyChecks(ctx, k8sClient, &clusterRunaiconfig, runaiConfigWithDefaults)

	assert.NoError(t, err)
	condition := findCondition(clusterRunaiconfig.Status.Conditions, runaiv1.ConditionDependenciesFulfilled)
	if assert.NotNil(t, condition) {
		assert.Equal(t, metav1.ConditionTrue, condition.Status)
		assert.Equal(t, runaiv1.DependenciesFulfilled, v1.ConditionReason(condition.Reason))
		assert.Equal(t, "Dependency checks skipped", condition.Message)
		assert.Equal(t, int64(2), condition.ObservedGeneration)
	}
}

func TestHandleDependencyChecks_DependenciesNotFulfilled(t *testing.T) {
	ctx, k8sClient, _, _ := setupTestEnv(t)

	// Create runaiConfig without skip flag and without creating dependencies
	clusterRunaiconfig := v1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:       runaiv1.DefaultName,
			Namespace:  runaiv1.DefaultNamespace,
			Generation: 3,
		},
		Spec: v1.RunaiConfigSpec{
			Internal: &internalv1.Internal{
				Global: &internalv1.Global{
					Version: ClusterVersion,
				},
				Development: &internalv1.Development{
					Headless: true, // Set headless to skip connectivity check
				},
			},
		},
	}

	runaiConfigWithDefaults := clusterRunaiconfig.DeepCopy()
	runaiConfigWithDefaults.SetDefaultsWhereNeeded()

	// Don't create dependencies, so check should fail
	err := handleDependencyChecks(ctx, k8sClient, &clusterRunaiconfig, runaiConfigWithDefaults)

	assert.Error(t, err)
	condition := findCondition(clusterRunaiconfig.Status.Conditions, runaiv1.ConditionDependenciesFulfilled)
	if assert.NotNil(t, condition) {
		assert.Equal(t, metav1.ConditionFalse, condition.Status)
		assert.Equal(t, runaiv1.DependenciesNotFulfilled, v1.ConditionReason(condition.Reason))
		assert.NotEmpty(t, condition.Message)
		assert.Equal(t, int64(3), condition.ObservedGeneration)
	}
}

func TestHandleDependencyChecks_ClusterRunaiconfigNilValues(t *testing.T) {
	ctx, k8sClient, _, _ := setupTestEnv(t)

	// Create runaiConfig without calling SetDefaultsWhereNeeded - simulating fresh Get from cluster
	clusterRunaiconfig := v1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:       runaiv1.DefaultName,
			Namespace:  runaiv1.DefaultNamespace,
			Generation: 4,
		},
		// Spec fields are nil/empty, mimicking fresh Get from cluster
	}

	// Create runaiConfigWithDefaults with defaults set
	runaiConfigWithDefaults := clusterRunaiconfig.DeepCopy()
	runaiConfigWithDefaults.SetDefaultsWhereNeeded()

	// Enable skip in the defaults version (NMC enabled)
	if runaiConfigWithDefaults.Spec.Internal == nil {
		runaiConfigWithDefaults.Spec.Internal = &internalv1.Internal{}
	}
	if runaiConfigWithDefaults.Spec.Internal.Global == nil {
		runaiConfigWithDefaults.Spec.Internal.Global = &internalv1.Global{}
	}
	runaiConfigWithDefaults.Spec.Internal.Global.NMC = internalv1.NMC{
		RemoteAPIServerURL: "https://remote-cluster.example.com",
	}

	err := handleDependencyChecks(ctx, k8sClient, &clusterRunaiconfig, runaiConfigWithDefaults)

	assert.NoError(t, err)
	// Verify condition is set on clusterRunaiconfig (the one without defaults)
	condition := findCondition(clusterRunaiconfig.Status.Conditions, runaiv1.ConditionDependenciesFulfilled)
	if assert.NotNil(t, condition) {
		assert.Equal(t, metav1.ConditionTrue, condition.Status)
		assert.Equal(t, int64(4), condition.ObservedGeneration)
	}
}

func TestHandleDependencyChecks_GenerationTracked(t *testing.T) {
	ctx, k8sClient, _, _ := setupTestEnv(t)

	clusterRunaiconfig := v1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:       runaiv1.DefaultName,
			Namespace:  runaiv1.DefaultNamespace,
			Generation: 42,
		},
		Spec: v1.RunaiConfigSpec{
			Internal: &internalv1.Internal{
				Development: &internalv1.Development{
					SkipDependencyChecks: true,
				},
			},
		},
	}

	runaiConfigWithDefaults := clusterRunaiconfig.DeepCopy()
	runaiConfigWithDefaults.SetDefaultsWhereNeeded()

	err := handleDependencyChecks(ctx, k8sClient, &clusterRunaiconfig, runaiConfigWithDefaults)

	assert.NoError(t, err)
	condition := findCondition(clusterRunaiconfig.Status.Conditions, runaiv1.ConditionDependenciesFulfilled)
	if assert.NotNil(t, condition) {
		assert.Equal(t, int64(42), condition.ObservedGeneration)
	}
}

func TestHandleDependencyChecks_ConditionUpdatesOnClusterRunaiconfig(t *testing.T) {
	ctx, k8sClient, _, _ := setupTestEnv(t)

	// Test that conditions are set on clusterRunaiconfig, not runaiConfigWithDefaults
	clusterRunaiconfig := v1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:       runaiv1.DefaultName,
			Namespace:  runaiv1.DefaultNamespace,
			Generation: 5,
		},
		Spec: v1.RunaiConfigSpec{
			Internal: &internalv1.Internal{
				Development: &internalv1.Development{
					SkipDependencyChecks: true,
				},
			},
		},
	}

	runaiConfigWithDefaults := clusterRunaiconfig.DeepCopy()
	runaiConfigWithDefaults.SetDefaultsWhereNeeded()

	// Verify no conditions initially
	assert.Empty(t, clusterRunaiconfig.Status.Conditions)
	assert.Empty(t, runaiConfigWithDefaults.Status.Conditions)

	err := handleDependencyChecks(ctx, k8sClient, &clusterRunaiconfig, runaiConfigWithDefaults)

	assert.NoError(t, err)
	// Condition should be set on clusterRunaiconfig
	assert.NotEmpty(t, clusterRunaiconfig.Status.Conditions)
	condition := findCondition(clusterRunaiconfig.Status.Conditions, runaiv1.ConditionDependenciesFulfilled)
	assert.NotNil(t, condition)

	// runaiConfigWithDefaults should not have been modified
	assert.Empty(t, runaiConfigWithDefaults.Status.Conditions)
}

func TestHandleDependencyChecks_NilInternalSpec(t *testing.T) {
	ctx, k8sClient, _, _ := setupTestEnv(t)

	// Create runaiConfig with nil Internal but other spec fields
	clusterRunaiconfig := v1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:       runaiv1.DefaultName,
			Namespace:  runaiv1.DefaultNamespace,
			Generation: 6,
		},
		Spec: v1.RunaiConfigSpec{
			// Internal is nil
		},
	}

	runaiConfigWithDefaults := clusterRunaiconfig.DeepCopy()
	runaiConfigWithDefaults.SetDefaultsWhereNeeded()
	// Set skip in defaults to avoid actual dependency checks
	if runaiConfigWithDefaults.Spec.Internal.Development == nil {
		runaiConfigWithDefaults.Spec.Internal.Development = &internalv1.Development{}
	}
	runaiConfigWithDefaults.Spec.Internal.Development.SkipDependencyChecks = true

	err := handleDependencyChecks(ctx, k8sClient, &clusterRunaiconfig, runaiConfigWithDefaults)

	assert.NoError(t, err)
	condition := findCondition(clusterRunaiconfig.Status.Conditions, runaiv1.ConditionDependenciesFulfilled)
	if assert.NotNil(t, condition) {
		assert.Equal(t, metav1.ConditionTrue, condition.Status)
	}
}

// Helper function to find a condition by type
func findCondition(conditions []metav1.Condition, conditionType v1.ConditionType) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == string(conditionType) {
			return &conditions[i]
		}
	}
	return nil
}

// --- Tests for SetupWithManager and Uncached client usage ---

// TestSetupWithManager_UncachedRequired_OpenShiftDetection tests that .Uncached()
// is necessary during SetupWithManager to properly detect OpenShift when the
// manager's cache hasn't started yet. This simulates the production scenario where
// SetupWithManager is called before mgr.Start().
func TestSetupWithManager_UncachedRequired_OpenShiftDetection(t *testing.T) {
	ctx := context.Background()
	s := scheme_init.Scheme()

	// Register OpenShift API types
	err := ocpconf.AddToScheme(s)
	assert.NoError(t, err, "failed to register OpenShift scheme")

	// Create OpenShift ClusterVersion resource
	clusterVersion := &ocpconf.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: "version",
		},
		Spec: ocpconf.ClusterVersionSpec{
			ClusterID: "test-cluster-id",
		},
	}

	// Simulate production scenario:
	// - Cached client (empty, simulating unstarted informer cache)
	// - Uncached API reader (has real data from API server)
	emptyCachedClient := fakeClient.NewClientBuilder().WithScheme(s).Build()
	realAPIReader := fakeClient.NewClientBuilder().
		WithScheme(s).
		WithObjects(clusterVersion).
		Build()

	// Create hybrid client: cached is empty, uncached has data
	// This mimics the state during SetupWithManager in production
	hybridClient := kc.New(emptyCachedClient, realAPIReader, nil)

	// Test 1: Using .Uncached() correctly detects OpenShift
	t.Run("UncachedClient_DetectsOpenShift", func(t *testing.T) {
		isOcp, err := util.IsOpenshift(ctx, hybridClient.Uncached())
		assert.NoError(t, err, "IsOpenshift should not error with uncached client")
		assert.True(t, isOcp, "Should detect OpenShift when using uncached client")
	})

	// Test 2: Using cached client fails to detect OpenShift (demonstrates the bug)
	t.Run("CachedClient_FailsToDetectOpenShift", func(t *testing.T) {
		isOcp, err := util.IsOpenshift(ctx, hybridClient)
		assert.NoError(t, err, "IsOpenshift should not error")
		assert.False(t, isOcp, "Should NOT detect OpenShift when using cached client with empty cache")
	})
}

// TestSetupWithManager_UncachedRequired_NonOpenShift tests that .Uncached()
// correctly identifies non-OpenShift clusters even when cache is empty
func TestSetupWithManager_UncachedRequired_NonOpenShift(t *testing.T) {
	ctx := context.Background()
	s := scheme_init.Scheme()

	// Register OpenShift API types (but don't create any resources)
	err := ocpconf.AddToScheme(s)
	assert.NoError(t, err, "failed to register OpenShift scheme")

	// Create clients without any ClusterVersion resources (non-OpenShift cluster)
	emptyCachedClient := fakeClient.NewClientBuilder().WithScheme(s).Build()
	emptyAPIReader := fakeClient.NewClientBuilder().WithScheme(s).Build()

	hybridClient := kc.New(emptyCachedClient, emptyAPIReader, nil)

	// Both cached and uncached should correctly identify as non-OpenShift
	t.Run("UncachedClient_CorrectlyIdentifiesNonOpenShift", func(t *testing.T) {
		isOcp, err := util.IsOpenshift(ctx, hybridClient.Uncached())
		assert.NoError(t, err, "IsOpenshift should not error")
		assert.False(t, isOcp, "Should correctly identify non-OpenShift cluster")
	})

	t.Run("CachedClient_AlsoIdentifiesNonOpenShift", func(t *testing.T) {
		isOcp, err := util.IsOpenshift(ctx, hybridClient)
		assert.NoError(t, err, "IsOpenshift should not error")
		assert.False(t, isOcp, "Cached client also identifies non-OpenShift (both empty)")
	})
}

// TestSetupWithManager_UncachedRequired_NoMatchError tests that .Uncached()
// correctly handles clusters where OpenShift CRDs don't exist at all
func TestSetupWithManager_UncachedRequired_NoMatchError(t *testing.T) {
	ctx := context.Background()
	s := scheme_init.Scheme()
	// Note: NOT registering OpenShift types to simulate CRD not existing

	emptyCachedClient := fakeClient.NewClientBuilder().WithScheme(s).Build()
	emptyAPIReader := fakeClient.NewClientBuilder().WithScheme(s).Build()

	hybridClient := kc.New(emptyCachedClient, emptyAPIReader, nil)

	// Should handle NoMatchError gracefully and return false
	t.Run("UncachedClient_HandlesNoMatchError", func(t *testing.T) {
		isOcp, err := util.IsOpenshift(ctx, hybridClient.Uncached())
		// May error due to unregistered type, but should not panic
		if err == nil {
			assert.False(t, isOcp, "Should not detect OpenShift when CRD doesn't exist")
		}
		// If there's an error, that's also acceptable behavior for unregistered types
	})
}

// TestSetupWithManager_UncachedWithOpenShiftCRD_DetectsOpenShift is an explicit test
// that verifies when .Uncached() is used with OpenShift CRDs present in the API server,
// it correctly detects that OpenShift exists. This is the expected production behavior
// during SetupWithManager.
func TestSetupWithManager_UncachedWithOpenShiftCRD_DetectsOpenShift(t *testing.T) {
	ctx := context.Background()
	s := scheme_init.Scheme()

	// Register OpenShift API types (simulating OpenShift cluster)
	err := ocpconf.AddToScheme(s)
	assert.NoError(t, err, "failed to register OpenShift scheme")

	// Create OpenShift ClusterVersion resource to simulate real OpenShift cluster
	clusterVersion := &ocpconf.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: "version",
		},
		Spec: ocpconf.ClusterVersionSpec{
			ClusterID: "production-cluster-123",
		},
		Status: ocpconf.ClusterVersionStatus{
			Desired: ocpconf.Release{
				Version: "4.12.0",
			},
		},
	}

	// Create API reader with OpenShift resources (simulates real API server)
	apiReaderWithOpenShift := fakeClient.NewClientBuilder().
		WithScheme(s).
		WithObjects(clusterVersion).
		Build()

	// Create empty cached client (simulates unstarted manager cache during SetupWithManager)
	emptyCachedClient := fakeClient.NewClientBuilder().WithScheme(s).Build()

	// Create hybrid client with empty cache but populated API reader
	hybridClient := kc.New(emptyCachedClient, apiReaderWithOpenShift, nil)

	// THE KEY TEST: With .Uncached() and OpenShift CRDs present, should detect OpenShift
	isOcp, err := util.IsOpenshift(ctx, hybridClient.Uncached())
	assert.NoError(t, err, "IsOpenshift should not error when using uncached client")
	assert.True(t, isOcp, "MUST detect OpenShift when using .Uncached() with OpenShift CRDs present in API server")

	// Verify we can also list the ClusterVersions directly
	cvList := &ocpconf.ClusterVersionList{}
	err = hybridClient.Uncached().List(ctx, cvList)
	assert.NoError(t, err, "Should be able to list ClusterVersions with uncached client")
	assert.Len(t, cvList.Items, 1, "Should find one ClusterVersion resource via uncached client")
	assert.Equal(t, "version", cvList.Items[0].Name, "Should retrieve correct ClusterVersion resource")

	// Show the contrast: cached client has empty cache
	cvListCached := &ocpconf.ClusterVersionList{}
	err = hybridClient.List(ctx, cvListCached)
	assert.NoError(t, err, "List should not error even with empty cache")
	assert.Len(t, cvListCached.Items, 0, "Cached client returns empty list because cache not started")
}
