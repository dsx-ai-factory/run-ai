package dynamic_schema_upgrades

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions"
	testu "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/test-util"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	anyworkload_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/anyworkload-controller"
	inference_workload_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/inference-workload-controller"
	workload_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-controller"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
)

// newRunaiConfigWithControllers builds an unstructured RunaiConfig where each of the
// three controllers is present (with the given GPUNetworkAccelerationEnabled value)
// only when the corresponding pointer is non-nil. A nil pointer means the controller
// section is absent from the spec entirely.
func newRunaiConfigWithControllers(workload, inference, anyworkload *bool) *unstructured.Unstructured {
	spec := map[string]interface{}{}
	if workload != nil {
		spec[workloadControllerOperandName] = map[string]interface{}{
			gpuNetworkAccelerationEnabledField: *workload,
		}
	}
	if inference != nil {
		spec[inferenceWorkloadControllerOperandName] = map[string]interface{}{
			gpuNetworkAccelerationEnabledField: *inference,
		}
	}
	if anyworkload != nil {
		spec[anyworkloadControllerOperandName] = map[string]interface{}{
			gpuNetworkAccelerationEnabledField: *anyworkload,
		}
	}
	return &unstructured.Unstructured{Object: map[string]interface{}{"spec": spec}}
}

func anyworkloadGpuNetworkAccelerationEnabled(t *testing.T, resource *unstructured.Unstructured) (bool, bool) {
	t.Helper()
	spec, ok := resource.Object["spec"].(map[string]interface{})
	assert.True(t, ok, "spec should be a map[string]interface{}")
	anyworkloadConf, ok := spec[anyworkloadControllerOperandName].(map[string]interface{})
	if !ok {
		return false, false
	}
	enabled, ok := anyworkloadConf[gpuNetworkAccelerationEnabledField].(bool)
	return enabled, ok
}

func TestGpuNetworkAcceleration_ShouldPerformUpgrade(t *testing.T) {
	upgrader := newGpuNetworkAccelerationUpgrader()

	testCases := []struct {
		name        string
		workload    *bool
		inference   *bool
		anyworkload *bool
		expected    bool
	}{
		{
			name:     "workload-controller enabled, anyworkload absent",
			workload: ptr.To(true),
			expected: true,
		},
		{
			name:      "inference-workload-controller enabled, anyworkload absent",
			inference: ptr.To(true),
			expected:  true,
		},
		{
			name:        "both legacy controllers enabled, anyworkload disabled",
			workload:    ptr.To(true),
			inference:   ptr.To(true),
			anyworkload: ptr.To(false),
			expected:    true,
		},
		{
			name:      "neither legacy controller enabled",
			workload:  ptr.To(false),
			inference: ptr.To(false),
			expected:  false,
		},
		{
			name:        "legacy enabled but anyworkload already enabled (idempotent)",
			workload:    ptr.To(true),
			anyworkload: ptr.To(true),
			expected:    false,
		},
		{
			name:     "controllers absent from spec",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource := newRunaiConfigWithControllers(tc.workload, tc.inference, tc.anyworkload)
			assert.Equal(t, tc.expected, upgrader.shouldPerformUpgrade(resource))
		})
	}
}

func TestGpuNetworkAcceleration_UpdateRunaiConfig_AnyworkloadSectionMissing(t *testing.T) {
	upgrader := newGpuNetworkAccelerationUpgrader()
	resource := newRunaiConfigWithControllers(ptr.To(true), nil, nil)

	updated, err := upgrader.updateRunaiConfig(resource)
	assert.NoError(t, err)

	enabled, ok := anyworkloadGpuNetworkAccelerationEnabled(t, updated)
	assert.True(t, ok, "anyworkload-controller section should be created")
	assert.True(t, enabled)
}

func TestGpuNetworkAcceleration_UpdateRunaiConfig_AnyworkloadSectionPresent(t *testing.T) {
	upgrader := newGpuNetworkAccelerationUpgrader()
	resource := newRunaiConfigWithControllers(nil, ptr.To(true), ptr.To(false))

	updated, err := upgrader.updateRunaiConfig(resource)
	assert.NoError(t, err)

	enabled, ok := anyworkloadGpuNetworkAccelerationEnabled(t, updated)
	assert.True(t, ok)
	assert.True(t, enabled)
}

// TestGpuNetworkAcceleration_Upgrade exercises the full Upgrade flow against a fake
// dynamic client, mirroring the nodepool-controller upgrade tests.
func TestGpuNetworkAcceleration_Upgrade(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	runaiConf := &runaiv1.RunaiConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: runaiv1.GroupVersion.String(),
			Kind:       runaiv1.RunaiConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      runaiv1.DefaultName,
		},
		Spec: runaiv1.RunaiConfigSpec{
			WorkloadController: &workload_controller.WorkloadController{
				GPUNetworkAccelerationEnabled: ptr.To(true),
			},
			InferenceWorkloadController: &inference_workload_controller.InferenceWorkloadController{
				GPUNetworkAccelerationEnabled: ptr.To(false),
			},
			AnyworkloadController: &anyworkload_controller.AnyworkloadController{
				GPUNetworkAccelerationEnabled: ptr.To(false),
			},
		},
	}
	testu.AddToDynamicClient(ra.RunaiConfigResource, runaiConf, dynamicClient)

	ctx := context.TODO()
	k8sClients := &clients.Clients{LocalClusterDynamicClient: dynamicClient}

	err := NewRunaiConfigEnableAnyworkloadGPUNetworkAcceleration().Upgrade(ctx, k8sClients)
	assert.NoError(t, err)

	dynamicResource := dynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
	val, err := dynamicResource.Get(ctx, runaiv1.DefaultName, metav1.GetOptions{})
	assert.NoError(t, err)

	enabled, ok := anyworkloadGpuNetworkAccelerationEnabled(t, val)
	assert.True(t, ok)
	assert.True(t, enabled, "anyworkload-controller GPUNetworkAccelerationEnabled should be true after upgrade")
}
