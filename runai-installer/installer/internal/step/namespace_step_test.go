package step

import (
	"context"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newTestRunAIInstallDataForNamespace() *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:     false,
		StepTimeout: 10 * time.Minute,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
	}
	data, _ := model.NewRunAIInstallData(opts)
	return data
}

func TestNewCreateNamespace(t *testing.T) {
	mockClients := SetupMockK8sClients()
	runaiData := newTestRunAIInstallDataForNamespace()

	step := NewCreateNamespace("test-namespace", mockClients)

	expectedName := "Create namespace test-namespace"
	if step.GetName() != expectedName {
		t.Errorf("expected name '%s', got %s", expectedName, step.GetName())
	}

	if step.namespace != "test-namespace" {
		t.Errorf("expected namespace 'test-namespace', got %s", step.namespace)
	}

	if step.GetStatusAsync(runaiData) != model.NotInstalled {
		t.Errorf("expected status NotInstalled, got %v", step.GetStatusAsync(runaiData))
	}
}

func TestCreateNamespace_GetStatus_NamespaceExists(t *testing.T) {
	// Create a namespace that already exists
	existingNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-namespace",
		},
	}

	mockClients := SetupMockK8sClients(existingNamespace)
	runaiData := newTestRunAIInstallDataForNamespace()
	step := NewCreateNamespace("existing-namespace", mockClients)

	status := step.GetStatusAsync(runaiData)
	if status != model.Installed {
		t.Errorf("expected Installed status for existing namespace, got %v", status)
	}
}

func TestCreateNamespace_GetStatus_NamespaceNotFound(t *testing.T) {
	mockClients := SetupMockK8sClients()
	runaiData := newTestRunAIInstallDataForNamespace()
	step := NewCreateNamespace("non-existing-namespace", mockClients)

	status := step.GetStatusAsync(runaiData)
	if status != model.NotInstalled {
		t.Errorf("expected NotInstalled status for non-existing namespace, got %v", status)
	}
}

func TestCreateNamespace_Execute_NamespaceAlreadyExists(t *testing.T) {
	// Create a namespace that already exists
	existingNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-namespace",
		},
	}

	mockClients := SetupMockK8sClients(existingNamespace)
	runaiData := newTestRunAIInstallDataForNamespace()
	step := NewCreateNamespace("existing-namespace", mockClients)

	err := step.Execute(context.Background(), runaiData)
	if err != nil {
		t.Errorf("expected no error when namespace already exists, got %v", err)
	}

	if step.GetStatusAsync(runaiData) != model.Installed {
		t.Errorf("expected Installed status after execution, got %v", step.GetStatusAsync(runaiData))
	}
}

func TestCreateNamespace_Execute_CreateNewNamespace(t *testing.T) {
	mockClients := SetupMockK8sClients()
	runaiData := newTestRunAIInstallDataForNamespace()
	step := NewCreateNamespace("new-namespace", mockClients)

	err := step.Execute(context.Background(), runaiData)
	if err != nil {
		t.Errorf("expected no error when creating new namespace, got %v", err)
	}

	if step.GetStatusAsync(runaiData) != model.Installed {
		t.Errorf("expected Installed status after execution, got %v", step.GetStatusAsync(runaiData))
	}

	// Verify the namespace was created
	namespace := &corev1.Namespace{}
	err = mockClients.GetClient().Get(context.Background(), client.ObjectKey{
		Name: "new-namespace",
	}, namespace)
	if err != nil {
		t.Errorf("expected namespace to be created, got error: %v", err)
	}

	if namespace.Name != "new-namespace" {
		t.Errorf("expected namespace name 'new-namespace', got %s", namespace.Name)
	}
}

func TestCreateNamespace_Execute_ContextCancellation(t *testing.T) {
	mockClients := SetupMockK8sClients()
	runaiData := newTestRunAIInstallDataForNamespace()
	step := NewCreateNamespace("test-namespace", mockClients)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := step.Execute(ctx, runaiData)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}

	if step.GetStatusAsync(runaiData) != model.Failed {
		t.Errorf("expected Failed status after context cancellation, got %v", step.GetStatusAsync(runaiData))
	}
}
