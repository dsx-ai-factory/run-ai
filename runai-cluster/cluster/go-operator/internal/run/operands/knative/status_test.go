package knative

import (
	"fmt"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type contextMock struct{}

func (c *contextMock) Deadline() (deadline time.Time, ok bool) { return time.Now(), true }
func (c *contextMock) Done() <-chan struct{}                   { return make(<-chan struct{}) }
func (c *contextMock) Err() error                              { return nil }
func (c *contextMock) Value(key any) any                       { return nil }

func newFakeClient(objects ...client.Object) client.Client {
	return fake.NewClientBuilder().WithObjects(objects...).Build()
}

func initKnativeResources() (*appsv1.Deployment, *appsv1.Deployment, *appsv1.Deployment, *corev1.ConfigMap, *corev1.ConfigMap) {
	deploymentHpa := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "autoscaler-hpa", Namespace: consts.KnativeNamespace},
	}
	deploymentKourier := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "3scale-kourier-gateway", Namespace: consts.KnativeNamespace},
	}
	deploymentKnative := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "controller", Namespace: consts.KnativeNamespace},
	}
	cmKnative := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: consts.KnativeFeaturesConfigMapName, Namespace: consts.KnativeNamespace},
		Data: map[string]string{
			"kubernetes.podspec-schedulername": "enabled",
			"kubernetes.podspec-affinity":      "enabled",
		},
	}
	cmKnativeAutoscaler := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: consts.KnativeAutoScalerConfigMapName, Namespace: consts.KnativeNamespace},
		Data:       map[string]string{"enable-scale-to-zero": "true"},
	}
	return deploymentHpa, deploymentKourier, deploymentKnative, cmKnative, cmKnativeAutoscaler
}

func TestKnativeFieldWhenNotAvailable(t *testing.T) {
	status := GetKnativeStatus(&contextMock{}, newFakeClient())
	assert.Equal(t, false, *status.Components["hpa"].Available)
	assert.NotEmpty(t, status.Components["hpa"].Reason)
	assert.Equal(t, false, *status.Components["kourier"].Available)
	assert.NotEmpty(t, status.Components["kourier"].Reason)
	assert.Equal(t, false, *status.Components["knative"].Available)
	assert.NotEmpty(t, status.Components["knative"].Reason)
}

func TestKnativeFieldWhenMinRequirementsAvailable(t *testing.T) {
	deploymentHpa, deploymentKourier, deploymentKnative, cmKnative, cmKnativeAutoscaler := initKnativeResources()
	status := GetKnativeStatus(&contextMock{}, newFakeClient(deploymentHpa, deploymentKourier, deploymentKnative, cmKnative, cmKnativeAutoscaler))
	assert.Equal(t, true, *status.Components["hpa"].Available)
	assert.Empty(t, status.Components["hpa"].Reason)
	assert.Equal(t, true, *status.Components["kourier"].Available)
	assert.Empty(t, status.Components["kourier"].Reason)
	assert.Equal(t, true, *status.Components["knative"].Available)
	assert.Empty(t, status.Components["knative"].Reason)
}

func TestKnativeFieldFeatureSpecificChecks(t *testing.T) {
	_, _, deploymentKnative, cmKnative, cmKnativeAutoscaler := initKnativeResources()
	cmKnative.Data["kubernetes.podspec-init-containers"] = "enabled"
	cmKnativeAutoscaler.Data["enable-scale-to-zero"] = "false"

	status := GetKnativeStatus(&contextMock{}, newFakeClient(deploymentKnative, cmKnative, cmKnativeAutoscaler))
	assert.Equal(t, true, *status.Components["knative"].Available)
	assert.Empty(t, status.Components["knative"].Reason)
	assert.Equal(t, false, *status.Components["knative"].Components["enable-scale-to-zero"].Available)
	assert.NotEmpty(t, status.Components["knative"].Components["enable-scale-to-zero"].Reason)
	assert.Equal(t, true, *status.Components["knative"].Components["kubernetes.podspec-init-containers"].Available)
	assert.Empty(t, status.Components["knative"].Components["kubernetes.podspec-init-containers"].Reason)
}

func TestKnativeFieldRequiredFeaturesNotAvailable(t *testing.T) {
	_, _, deploymentKnative, cmKnative, cmKnativeAutoscaler := initKnativeResources()
	cmKnative.Data["kubernetes.podspec-schedulername"] = "disabled"

	status := GetKnativeStatus(&contextMock{}, newFakeClient(deploymentKnative, cmKnative, cmKnativeAutoscaler))
	assert.Equal(t, false, *status.Components["knative"].Available)
	assert.NotEmpty(t, status.Components["knative"].Reason)
}

func TestValidateKnativeFeatureEnabled(t *testing.T) {
	tests := []struct {
		name           string
		cm             *corev1.ConfigMap
		feature        string
		cmName         string
		expectedStatus runv1.PublicStatus
	}{
		{
			name: "required feature enabled",
			cm: &corev1.ConfigMap{
				Data: map[string]string{consts.KnativeFeaturePodSpecSchedulerName: consts.KnativeFeatureValueEnabled},
			},
			feature: consts.KnativeFeaturePodSpecSchedulerName,
			cmName:  consts.KnativeFeaturesConfigMapName,
			expectedStatus: runv1.PublicStatus{
				Available: ptr.To(true),
				Components: map[string]runv1.PublicStatus{
					consts.KnativeFeaturePodSpecSchedulerName: {Available: ptr.To(true)},
				},
			},
		},
		{
			name: "required feature disabled",
			cm: &corev1.ConfigMap{
				Data: map[string]string{consts.KnativeFeaturePodSpecSchedulerName: consts.KnativeFeatureValueDisabled},
			},
			feature: consts.KnativeFeaturePodSpecSchedulerName,
			cmName:  consts.KnativeFeaturesConfigMapName,
			expectedStatus: runv1.PublicStatus{
				Available: ptr.To(false),
				Reason: fmt.Sprintf("Feature %s is not enabled in knative %s config-map from %s namespace, actual value <%s> expected value <%s>",
					consts.KnativeFeaturePodSpecSchedulerName, consts.KnativeFeaturesConfigMapName, consts.KnativeNamespace,
					consts.KnativeFeatureValueDisabled, consts.KnativeFeatureValueEnabled),
				Components: map[string]runv1.PublicStatus{
					consts.KnativeFeaturePodSpecSchedulerName: {
						Available: ptr.To(false),
						Reason: fmt.Sprintf("Feature %s in knative %s config-map from %s namespace, has actual value <%s> expected value <%s>",
							consts.KnativeFeaturePodSpecSchedulerName, consts.KnativeFeaturesConfigMapName, consts.KnativeNamespace,
							consts.KnativeFeatureValueDisabled, consts.KnativeFeatureValueEnabled),
					},
				},
			},
		},
		{
			name: "optional feature explicitly disabled",
			cm: &corev1.ConfigMap{
				Data: map[string]string{consts.KnativeFeaturePodSpecTolerations: consts.KnativeFeatureValueDisabled},
			},
			feature: consts.KnativeFeaturePodSpecTolerations,
			cmName:  consts.KnativeFeaturesConfigMapName,
			expectedStatus: runv1.PublicStatus{
				Available: ptr.To(true),
				Components: map[string]runv1.PublicStatus{
					consts.KnativeFeaturePodSpecTolerations: {
						Available: ptr.To(false),
						Reason: fmt.Sprintf("Feature %s in knative %s config-map from %s namespace, has actual value <%s> expected value <%s>",
							consts.KnativeFeaturePodSpecTolerations, consts.KnativeFeaturesConfigMapName, consts.KnativeNamespace,
							consts.KnativeFeatureValueDisabled, consts.KnativeFeatureValueEnabled),
					},
				},
			},
		},
		{
			name:    "scale to zero with default value",
			cm:      &corev1.ConfigMap{Data: map[string]string{}},
			feature: consts.KnativeFeatureEnableScaleToZero,
			cmName:  consts.KnativeAutoScalerConfigMapName,
			expectedStatus: runv1.PublicStatus{
				Available: ptr.To(true),
				Components: map[string]runv1.PublicStatus{
					consts.KnativeFeatureEnableScaleToZero: {Available: ptr.To(true)},
				},
			},
		},
		{
			name:    "scale to zero explicitly disabled",
			cm:      &corev1.ConfigMap{Data: map[string]string{consts.KnativeFeatureEnableScaleToZero: "false"}},
			feature: consts.KnativeFeatureEnableScaleToZero,
			cmName:  consts.KnativeAutoScalerConfigMapName,
			expectedStatus: runv1.PublicStatus{
				Available: ptr.To(true),
				Components: map[string]runv1.PublicStatus{
					consts.KnativeFeatureEnableScaleToZero: {
						Available: ptr.To(false),
						Reason: fmt.Sprintf("Feature %s in knative %s config-map from %s namespace, has actual value <false> expected value <%s>",
							consts.KnativeFeatureEnableScaleToZero, consts.KnativeAutoScalerConfigMapName, consts.KnativeNamespace,
							consts.KnativeFeatureValueTrue),
					},
				},
			},
		},
		{
			name:    "scale to zero explicitly enabled",
			cm:      &corev1.ConfigMap{Data: map[string]string{consts.KnativeFeatureEnableScaleToZero: "true"}},
			feature: consts.KnativeFeatureEnableScaleToZero,
			cmName:  consts.KnativeAutoScalerConfigMapName,
			expectedStatus: runv1.PublicStatus{
				Available: ptr.To(true),
				Components: map[string]runv1.PublicStatus{
					consts.KnativeFeatureEnableScaleToZero: {Available: ptr.To(true)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			knativeStatus := &runv1.PublicStatus{}
			var config consts.KnativeFeatureConfig
			if tt.cmName == consts.KnativeFeaturesConfigMapName {
				config = consts.KnativeFeaturesConfigMap[tt.feature]
			} else {
				config = consts.KnativeAutoscalerConfigMap[tt.feature]
			}
			validateKnativeFeatureEnabled(tt.cm, knativeStatus, tt.feature, config, tt.cmName)

			if tt.expectedStatus.Available != nil && !*tt.expectedStatus.Available {
				assert.Equal(t, tt.expectedStatus.Available, knativeStatus.Available)
			}
			if tt.expectedStatus.Reason != "" {
				assert.Equal(t, tt.expectedStatus.Reason, knativeStatus.Reason)
			}
			assert.Equal(t, len(tt.expectedStatus.Components), len(knativeStatus.Components))
			for feature, expectedComponent := range tt.expectedStatus.Components {
				actualComponent, exists := knativeStatus.Components[feature]
				assert.True(t, exists, "component %s not found", feature)
				assert.Equal(t, expectedComponent.Available, actualComponent.Available)
				assert.Equal(t, expectedComponent.Reason, actualComponent.Reason)
			}
		})
	}
}

const (
	emptyDomainReason    = "Domain configuration is empty - external access may not work properly"
	domainNotFoundReason = "Knative 'config-domain' config-map not found in %s namespace"

	exampleContentWithPlaceholder = "example content"
)

func TestKnativeExternalConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		testType string
		setup    func() (*corev1.ConfigMap, bool, runv1.PublicStatus)
	}{
		{
			name:     "domain validation: nil configmap data",
			testType: "validation",
			setup: func() (*corev1.ConfigMap, bool, runv1.PublicStatus) {
				return &corev1.ConfigMap{Data: nil},
					false, runv1.PublicStatus{Available: runv1.AvailableStatus(false), Reason: emptyDomainReason}
			},
		},
		{
			name:     "domain validation: empty configmap data",
			testType: "validation",
			setup: func() (*corev1.ConfigMap, bool, runv1.PublicStatus) {
				return &corev1.ConfigMap{Data: map[string]string{}},
					false, runv1.PublicStatus{Available: runv1.AvailableStatus(false), Reason: emptyDomainReason}
			},
		},
		{
			name:     "domain validation: placeholder domain in _example with valid domain",
			testType: "validation",
			setup: func() (*corev1.ConfigMap, bool, runv1.PublicStatus) {
				return &corev1.ConfigMap{Data: map[string]string{"_example": exampleContentWithPlaceholder, "mydomain.com": ""}},
					false, runv1.PublicStatus{Available: runv1.AvailableStatus(true), Reason: ""}
			},
		},
		{
			name:     "domain validation: only _example data should be treated as empty",
			testType: "validation",
			setup: func() (*corev1.ConfigMap, bool, runv1.PublicStatus) {
				return &corev1.ConfigMap{Data: map[string]string{"_example": exampleContentWithPlaceholder}},
					false, runv1.PublicStatus{Available: runv1.AvailableStatus(false), Reason: emptyDomainReason}
			},
		},
		{
			name:     "external config: kourier available with valid domain",
			testType: "external-config",
			setup: func() (*corev1.ConfigMap, bool, runv1.PublicStatus) {
				return &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: consts.KnativeDomainConfigMapName, Namespace: consts.KnativeNamespace},
					Data:       map[string]string{"mydomain.com": ""},
				}, true, runv1.PublicStatus{Available: runv1.AvailableStatus(true), Reason: ""}
			},
		},
		{
			name:     "external config: kourier available but domain missing",
			testType: "external-config",
			setup: func() (*corev1.ConfigMap, bool, runv1.PublicStatus) {
				return nil, true, runv1.PublicStatus{Available: runv1.AvailableStatus(false), Reason: fmt.Sprintf(domainNotFoundReason, consts.KnativeNamespace)}
			},
		},
		{
			name:     "external config: kourier unavailable",
			testType: "external-config",
			setup: func() (*corev1.ConfigMap, bool, runv1.PublicStatus) {
				return &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: consts.KnativeDomainConfigMapName, Namespace: consts.KnativeNamespace},
					Data:       map[string]string{"mydomain.com": ""},
				}, false, runv1.PublicStatus{}
			},
		},
		{
			name:     "integration: full knative status with valid domain",
			testType: "integration",
			setup: func() (*corev1.ConfigMap, bool, runv1.PublicStatus) {
				return &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: consts.KnativeDomainConfigMapName, Namespace: consts.KnativeNamespace},
					Data:       map[string]string{"mydomain.com": ""},
				}, true, runv1.PublicStatus{Available: runv1.AvailableStatus(true), Reason: ""}
			},
		},
		{
			name:     "integration: full knative status with missing domain",
			testType: "integration",
			setup: func() (*corev1.ConfigMap, bool, runv1.PublicStatus) {
				return nil, false, runv1.PublicStatus{Available: runv1.AvailableStatus(false), Reason: fmt.Sprintf(domainNotFoundReason, consts.KnativeNamespace)}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &contextMock{}
			configMap, kourierAvailable, expectedStatus := tt.setup()

			switch tt.testType {
			case "integration":
				objects := []client.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "3scale-kourier-gateway", Namespace: consts.KnativeNamespace},
						Status:     appsv1.DeploymentStatus{AvailableReplicas: 1},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "controller", Namespace: consts.KnativeNamespace},
						Status:     appsv1.DeploymentStatus{AvailableReplicas: 1},
					},
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{Name: consts.KnativeFeaturesConfigMapName, Namespace: consts.KnativeNamespace},
						Data: map[string]string{
							consts.KnativeFeaturePodSpecSchedulerName: consts.KnativeFeatureValueEnabled,
							consts.KnativeFeaturePodSpecAffinity:      consts.KnativeFeatureValueEnabled,
						},
					},
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{Name: consts.KnativeAutoScalerConfigMapName, Namespace: consts.KnativeNamespace},
						Data:       map[string]string{consts.KnativeFeatureEnableScaleToZero: consts.KnativeFeatureValueTrue},
					},
				}
				if configMap != nil {
					objects = append(objects, configMap)
				}
				fakeClient := fake.NewClientBuilder().WithObjects(objects...).Build()
				result := GetKnativeStatus(ctx, fakeClient)

				assert.NotNil(t, result.Components)
				kourierStatus, exists := result.Components["kourier"]
				assert.True(t, exists)
				assert.NotNil(t, kourierStatus.Components)
				actualComponent, exists := kourierStatus.Components["external-config-domain"]
				assert.True(t, exists)
				assert.Equal(t, expectedStatus.Available, actualComponent.Available)
				assert.Equal(t, expectedStatus.Reason, actualComponent.Reason)

			case "external-config":
				var objects []client.Object
				if configMap != nil {
					objects = append(objects, configMap)
				}
				fakeClient := fake.NewClientBuilder().WithObjects(objects...).Build()
				kourierStatus := &runv1.PublicStatus{Available: runv1.AvailableStatus(kourierAvailable)}

				setKnativeExternalConfigStatus(ctx, fakeClient, kourierStatus)

				if kourierAvailable && (configMap != nil || expectedStatus.Reason != "") {
					assert.NotNil(t, kourierStatus.Components)
					actualComponent, exists := kourierStatus.Components["external-config-domain"]
					assert.True(t, exists)
					assert.Equal(t, expectedStatus.Available, actualComponent.Available)
					assert.Equal(t, expectedStatus.Reason, actualComponent.Reason)
				} else {
					assert.Nil(t, kourierStatus.Components)
				}

			case "validation":
				result := validateKnativeDomainConfig(configMap)
				assert.Equal(t, expectedStatus.Available, result.Available)
				assert.Equal(t, expectedStatus.Reason, result.Reason)
			}
		})
	}
}
