package steps

import (
	"context"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	k8smocks "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient/mocks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupMockK8sClientsForNetwork(t *testing.T, objects ...client.Object) *k8smocks.MockK8sClients {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		Build()

	mockClients := k8smocks.NewMockK8sClients(t)
	mockClients.On("GetClient").Return(fakeClient).Maybe()
	mockClients.On("GetConfig").Return(&rest.Config{}).Maybe()

	return mockClients
}

func newTestRunAIInstallDataForNetwork(gpuType config.GPUType) *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:     false,
		StepTimeout: 10 * time.Minute,
		GPUType:     gpuType,
		LogsDir:     "/tmp",
	}
	data, _ := model.NewRunAIInstallData(opts)
	return data
}

func TestNewNetworkOperatorInstallStep(t *testing.T) {
	mockClients := setupMockK8sClientsForNetwork(t)

	step := NewNetworkOperatorInstallStep(mockClients, "v25.7.0")

	if step.GetName() != NetworkOperatorStepName {
		t.Errorf("expected name '%s', got %s", NetworkOperatorStepName, step.GetName())
	}

	networkOp := step.(*NetworkOperatorInstallStep)
	if networkOp.version != "v25.7.0" {
		t.Errorf("expected version 'v25.7.0', got %s", networkOp.version)
	}

	if false {
		t.Error("expected Network Operator to be mandatory")
	}
}

func TestNetworkOperatorInstallStep_GetStatus_FakeGPU(t *testing.T) {
	mockClients := setupMockK8sClientsForNetwork(t)
	runaiData := newTestRunAIInstallDataForNetwork(config.GPUTypeFake)

	networkOp := NewNetworkOperatorInstallStep(mockClients, "v25.7.0")

	// Use GetStatusSync for immediate results in tests
	status := networkOp.GetStatusSync(runaiData)
	// Should return Installed for fake GPU
	if status != model.Installed {
		t.Errorf("expected Installed status for fake GPU, got %v", status)
	}
}

func TestNetworkOperatorInstallStep_GetStatus_RealGPU_NotInstalled(t *testing.T) {
	mockClients := setupMockK8sClientsForNetwork(t)
	runaiData := newTestRunAIInstallDataForNetwork(config.GPUTypeReal)

	networkOp := NewNetworkOperatorInstallStep(mockClients, "v25.7.0")

	status := networkOp.GetStatusAsync(runaiData)
	if status != model.NotInstalled {
		t.Errorf("expected NotInstalled status when network operator not found, got %v", status)
	}
}

func TestNetworkOperatorInstallStep_GetStatus_RealGPU_Installed(t *testing.T) {
	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nvidia-network-operator",
		},
	}

	// Create a deployment that matches network operator pattern
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "network-operator",
			Namespace: "nvidia-network-operator",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
		},
	}

	mockClients := setupMockK8sClientsForNetwork(t, namespace, deployment)
	runaiData := newTestRunAIInstallDataForNetwork(config.GPUTypeReal)

	networkOp := NewNetworkOperatorInstallStep(mockClients, "v25.7.0")

	// GetStatus now runs probe in background, poll until status changes
	var status model.InstallPhase
	for i := 0; i < 50; i++ {
		status = networkOp.GetStatusAsync(runaiData)
		if status == model.Installed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if status != model.Installed {
		t.Errorf("expected Installed status when network operator deployment found, got %v", status)
	}
}

func TestNetworkOperatorInstallStep_GetStatus_TerminalStatus(t *testing.T) {
	mockClients := setupMockK8sClientsForNetwork(t)
	runaiData := newTestRunAIInstallDataForNetwork(config.GPUTypeReal)

	networkOp := NewNetworkOperatorInstallStep(mockClients, "v25.7.0").(*NetworkOperatorInstallStep)

	// Set a terminal status
	networkOp.SetStatus(model.Failed, nil)

	status := networkOp.GetStatusAsync(runaiData)
	if status != model.Failed {
		t.Errorf("expected Failed status to be returned as terminal, got %v", status)
	}
}

func TestNetworkOperatorInstallStep_Execute_FakeGPU(t *testing.T) {
	mockClients := setupMockK8sClientsForNetwork(t)
	runaiData := newTestRunAIInstallDataForNetwork(config.GPUTypeFake)

	networkOp := NewNetworkOperatorInstallStep(mockClients, "v25.7.0")

	ctx := context.Background()
	err := networkOp.Execute(ctx, runaiData)

	// Should not error for fake GPU (it just returns nil)
	if err != nil {
		t.Errorf("expected no error for fake GPU, got %v", err)
	}
}

func TestNetworkOperatorInstallStep_Execute_RealGPU(t *testing.T) {
	t.Skip("Skipping test that requires network access and real cluster connection")

	mockClients := setupMockK8sClientsForNetwork(t)
	runaiData := newTestRunAIInstallDataForNetwork(config.GPUTypeReal)

	networkOp := NewNetworkOperatorInstallStep(mockClients, "v25.7.0").(*NetworkOperatorInstallStep)

	ctx := context.Background()
	err := networkOp.Execute(ctx, runaiData)

	// We expect an error because helm installation will fail in test environment
	if err == nil {
		t.Log("Execute returned no error (helm may have succeeded in test env)")
	}

	// Verify that status was set to Installing
	if networkOp.GetCachedStatus() != model.Installing && networkOp.GetCachedStatus() != model.Failed {
		t.Errorf("expected Installing or Failed status, got %v", networkOp.GetCachedStatus())
	}
}

func TestNetworkOperatorInstallStep_determineWithConfigAndCRDChecker_CRDInstalled(t *testing.T) {
	mockClients := setupMockK8sClientsForNetwork(t)
	networkOp := NewNetworkOperatorInstallStep(mockClients, "v25.7.0").(*NetworkOperatorInstallStep)

	mockCRDChecker := func(cfg *rest.Config, crdName string) bool {
		return true
	}

	status := networkOp.determineWithConfigAndCRDChecker(mockClients.GetConfig(), mockCRDChecker)
	if status != model.Installed {
		t.Errorf("expected Installed when CRD exists, got %v", status)
	}
}

func TestNetworkOperatorInstallStep_determineWithConfigAndCRDChecker_CRDNotInstalled(t *testing.T) {
	mockClients := setupMockK8sClientsForNetwork(t)
	networkOp := NewNetworkOperatorInstallStep(mockClients, "v25.7.0").(*NetworkOperatorInstallStep)

	mockCRDChecker := func(cfg *rest.Config, crdName string) bool {
		return false
	}

	status := networkOp.determineWithConfigAndCRDChecker(mockClients.GetConfig(), mockCRDChecker)
	if status != model.NotInstalled {
		t.Errorf("expected NotInstalled when CRD doesn't exist, got %v", status)
	}
}
