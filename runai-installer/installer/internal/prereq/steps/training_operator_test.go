package steps

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	stepsmocks "github.com/run-ai/runai/runai-installer/installer/internal/prereq/steps/mocks"
	k8smocks "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient/mocks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupMockK8sClientsForTraining(t *testing.T, objects ...client.Object) *k8smocks.MockK8sClients {
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

func newTestRunAIInstallDataForTraining() *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:     false,
		StepTimeout: 10 * time.Minute,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
	}
	data, _ := model.NewRunAIInstallData(opts)
	return data
}

func TestNewTrainingOperatorInstallStep(t *testing.T) {
	mockClients := setupMockK8sClientsForTraining(t)

	step := NewTrainingOperatorInstallStep(mockClients, "v1.8.1")

	if step.GetName() != TrainingOperatorStepName {
		t.Errorf("expected name '%s', got %s", TrainingOperatorStepName, step.GetName())
	}

	trainingOp := step.(*TrainingOperatorInstallStep)
	if trainingOp.version != "v1.8.1" {
		t.Errorf("expected version 'v1.8.1', got %s", trainingOp.version)
	}
}

func TestTrainingOperatorInstallStep_GetStatus_NotInstalled(t *testing.T) {
	mockClients := setupMockK8sClientsForTraining(t)
	runaiData := newTestRunAIInstallDataForTraining()

	trainingOp := NewTrainingOperatorInstallStep(mockClients, "v1.8.1")

	status := trainingOp.GetStatusAsync(runaiData)
	if status != model.NotInstalled {
		t.Errorf("expected NotInstalled status when training operator not found, got %v", status)
	}
}

func TestTrainingOperatorInstallStep_GetStatus_InstalledWithDeployment(t *testing.T) {
	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeflow",
		},
	}

	// Create a deployment that matches training operator pattern
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "training-operator",
			Namespace: "kubeflow",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
		},
	}

	mockClients := setupMockK8sClientsForTraining(t, namespace, deployment)
	runaiData := newTestRunAIInstallDataForTraining()

	trainingOp := NewTrainingOperatorInstallStep(mockClients, "v1.8.1").(*TrainingOperatorInstallStep)

	// Mock CRD checker to return false for all CRDs
	// Even with deployment, without CRDs it should be NotInstalled
	trainingOp.crdChecker = &mockCRDChecker{
		isCRDInstalledFunc: func(cfg *rest.Config, crdName string) bool {
			return false
		},
	}

	status := trainingOp.GetStatusAsync(runaiData)
	if status != model.NotInstalled {
		t.Errorf("expected NotInstalled status without CRDs, got %v", status)
	}
}

func TestTrainingOperatorInstallStep_GetStatus_InstalledWithCRD(t *testing.T) {
	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeflow",
		},
	}

	// Create a deployment that matches training operator pattern
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "training-operator",
			Namespace: "kubeflow",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
		},
	}

	mockClients := setupMockK8sClientsForTraining(t, namespace, deployment)
	runaiData := newTestRunAIInstallDataForTraining()

	trainingOp := NewTrainingOperatorInstallStep(mockClients, "v1.8.1").(*TrainingOperatorInstallStep)

	// Mock CRD checker to return true for at least one CRD
	trainingOp.crdChecker = &mockCRDChecker{
		isCRDInstalledFunc: func(cfg *rest.Config, crdName string) bool {
			if crdName == "tfjobs.kubeflow.org" {
				return true
			}
			return false
		},
	}

	// GetStatus now runs probe in background, poll until status changes
	var status model.InstallPhase
	for i := 0; i < 50; i++ {
		status = trainingOp.GetStatusAsync(runaiData)
		if status == model.Installed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if status != model.Installed {
		t.Errorf("expected Installed status with deployment and CRD, got %v", status)
	}
}

func TestTrainingOperatorInstallStep_GetStatus_TerminalStatus(t *testing.T) {
	mockClients := setupMockK8sClientsForTraining(t)
	runaiData := newTestRunAIInstallDataForTraining()

	trainingOp := NewTrainingOperatorInstallStep(mockClients, "v1.8.1").(*TrainingOperatorInstallStep)

	// Set a terminal status
	trainingOp.SetStatus(model.Installed, nil)

	status := trainingOp.GetStatusAsync(runaiData)
	if status != model.Installed {
		t.Errorf("expected Installed status to be returned as terminal, got %v", status)
	}
}

func TestTrainingOperatorInstallStep_Execute(t *testing.T) {
	mockClients := setupMockK8sClientsForTraining(t)
	runaiData := newTestRunAIInstallDataForTraining()

	trainingOp := NewTrainingOperatorInstallStep(mockClients, "v1.8.1").(*TrainingOperatorInstallStep)

	// Create a mock executor using mockery-generated mock
	mockExecutor := stepsmocks.NewMockTrainingOperatorExecutor(t)
	mockExecutor.On("ApplyTrainingOperatorKubectl", "v1.8.1").Return(nil)
	mockExecutor.On("PatchDeployment", mockClients).Return(nil)
	mockExecutor.On("DeleteCRD", mockClients).Return(nil)
	mockExecutor.On("ApplyMPIOperator").Return(nil)

	// Inject the mock executor
	trainingOp.executor = mockExecutor

	ctx := context.Background()
	err := trainingOp.Execute(ctx, runaiData)

	// Should succeed with mocked operations
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify that status was set to Installing (since we don't set it to Installed in Execute)
	if trainingOp.GetCachedStatus() != model.Installing {
		t.Errorf("expected Installing status, got %v", trainingOp.GetCachedStatus())
	}
}

func TestTrainingOperatorInstallStep_Execute_Error(t *testing.T) {
	mockClients := setupMockK8sClientsForTraining(t)
	runaiData := newTestRunAIInstallDataForTraining()

	trainingOp := NewTrainingOperatorInstallStep(mockClients, "v1.8.1").(*TrainingOperatorInstallStep)

	// Create a mock executor that fails on the first operation
	expectedErr := fmt.Errorf("kubectl apply failed")
	mockExecutor := stepsmocks.NewMockTrainingOperatorExecutor(t)
	mockExecutor.On("ApplyTrainingOperatorKubectl", "v1.8.1").Return(expectedErr)

	// Inject the mock executor
	trainingOp.executor = mockExecutor

	ctx := context.Background()
	err := trainingOp.Execute(ctx, runaiData)

	// Should fail with the expected error
	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}

	// Verify that status was set to Failed
	if trainingOp.GetCachedStatus() != model.Failed {
		t.Errorf("expected Failed status, got %v", trainingOp.GetCachedStatus())
	}
}

func TestTrainingOperatorInstallStep_Execute_ErrorOnPatchDeployment(t *testing.T) {
	mockClients := setupMockK8sClientsForTraining(t)
	runaiData := newTestRunAIInstallDataForTraining()

	trainingOp := NewTrainingOperatorInstallStep(mockClients, "v1.8.1").(*TrainingOperatorInstallStep)

	// Create a mock executor that fails on PatchDeployment
	expectedErr := fmt.Errorf("patch deployment failed")
	mockExecutor := stepsmocks.NewMockTrainingOperatorExecutor(t)
	mockExecutor.On("ApplyTrainingOperatorKubectl", "v1.8.1").Return(nil)
	mockExecutor.On("PatchDeployment", mockClients).Return(expectedErr)

	// Inject the mock executor
	trainingOp.executor = mockExecutor

	ctx := context.Background()
	err := trainingOp.Execute(ctx, runaiData)

	// Should fail with the expected error
	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}

	// Verify that status was set to Failed
	if trainingOp.GetCachedStatus() != model.Failed {
		t.Errorf("expected Failed status, got %v", trainingOp.GetCachedStatus())
	}

	// Verify that both methods were called (mockery will assert this)
	mockExecutor.AssertCalled(t, "ApplyTrainingOperatorKubectl", "v1.8.1")
	mockExecutor.AssertCalled(t, "PatchDeployment", mockClients)
	mockExecutor.AssertNotCalled(t, "DeleteCRD", mockClients)
	mockExecutor.AssertNotCalled(t, "ApplyMPIOperator")
}

func TestTrainingOperatorInstallStep_getInstallationPhaseWithClientAndCRDChecker_NoDeployment(t *testing.T) {
	mockClients := setupMockK8sClientsForTraining(t)
	runaiData := newTestRunAIInstallDataForTraining()
	trainingOp := NewTrainingOperatorInstallStep(mockClients, "v1.8.1").(*TrainingOperatorInstallStep)

	mockCRDCheckerFunc := func(cfg *rest.Config, crdName string) bool {
		return true
	}

	status := trainingOp.getInstallationPhaseWithClientAndCRDChecker(
		mockClients.GetClient(),
		mockClients.GetConfig(),
		mockCRDCheckerFunc,
		runaiData,
	)

	// Without deployment, should be NotInstalled
	if status != model.NotInstalled {
		t.Errorf("expected NotInstalled without deployment, got %v", status)
	}
}

func TestTrainingOperatorInstallStep_getInstallationPhaseWithClientAndCRDChecker_WithDeployment(t *testing.T) {
	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeflow",
		},
	}

	// Create a deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "training-operator",
			Namespace: "kubeflow",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
		},
	}

	mockClients := setupMockK8sClientsForTraining(t, namespace, deployment)
	runaiData := newTestRunAIInstallDataForTraining()
	trainingOp := NewTrainingOperatorInstallStep(mockClients, "v1.8.1").(*TrainingOperatorInstallStep)

	mockCRDCheckerFunc := func(cfg *rest.Config, crdName string) bool {
		return crdName == "pytorchjobs.kubeflow.org"
	}

	status := trainingOp.getInstallationPhaseWithClientAndCRDChecker(
		mockClients.GetClient(),
		mockClients.GetConfig(),
		mockCRDCheckerFunc,
		runaiData,
	)

	// With deployment and at least one CRD, should be Installed
	if status != model.Installed {
		t.Errorf("expected Installed with deployment and CRD, got %v", status)
	}
}

func TestTrainingOperatorInstallStep_getInstallationPhaseWithClientAndCRDChecker_MultipleCRDs(t *testing.T) {
	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeflow",
		},
	}

	// Create a deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "training-operator",
			Namespace: "kubeflow",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
		},
	}

	mockClients := setupMockK8sClientsForTraining(t, namespace, deployment)
	runaiData := newTestRunAIInstallDataForTraining()
	trainingOp := NewTrainingOperatorInstallStep(mockClients, "v1.8.1").(*TrainingOperatorInstallStep)

	// Test with different CRDs
	testCases := []struct {
		name        string
		crdName     string
		shouldMatch bool
	}{
		{"TFJob CRD", "tfjobs.kubeflow.org", true},
		{"PyTorchJob CRD", "pytorchjobs.kubeflow.org", true},
		{"MPIJob CRD", "mpijobs.kubeflow.org", true},
		{"XGBoostJob CRD", "xgboostjobs.kubeflow.org", true},
		{"MXJob CRD", "mxjobs.kubeflow.org", true},
		{"PaddleJob CRD", "paddlejobs.kubeflow.org", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCRDCheckerFunc := func(cfg *rest.Config, crdName string) bool {
				return crdName == tc.crdName
			}

			status := trainingOp.getInstallationPhaseWithClientAndCRDChecker(
				mockClients.GetClient(),
				mockClients.GetConfig(),
				mockCRDCheckerFunc,
				runaiData,
			)

			if tc.shouldMatch && status != model.Installed {
				t.Errorf("expected Installed with %s, got %v", tc.crdName, status)
			}
		})
	}
}
