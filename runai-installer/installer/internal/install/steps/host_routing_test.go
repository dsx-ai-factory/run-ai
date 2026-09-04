/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	k8smocks "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient/mocks"
	pkgvalidation "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/validation"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupMockK8sClientsForHostRouting(t *testing.T, objects ...client.Object) *k8smocks.MockK8sClients {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		Build()

	mockClients := k8smocks.NewMockK8sClients(t)
	mockClients.On("GetClient").Return(fakeClient).Maybe()
	mockClients.On("GetConfig").Return(&rest.Config{}).Maybe()

	return mockClients
}

func newTestRunAIInstallDataForHostRouting() *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:     false,
		StepTimeout: 10 * time.Minute,
		LogsDir:     "/tmp",
	}
	data, _ := model.NewRunAIInstallData(opts)
	data.Inputs[configinputs.InstallInputFQDN] = "test.example.com"
	data.Inputs[configinputs.InstallInputWorkloadRouting] = pkgvalidation.WorkloadRoutingHost
	return data
}

func TestNewInstallHostBasedRouting(t *testing.T) {
	mockClients := setupMockK8sClientsForHostRouting(t)

	step := NewInstallHostBasedRouting(mockClients)

	if step.GetName() != "Install host-based routing" {
		t.Errorf("expected name 'Install host-based routing', got %s", step.GetName())
	}
}

func TestHostRouting_EvaluateStatus_SecretNotFound(t *testing.T) {
	mockClients := setupMockK8sClientsForHostRouting(t)
	runaiData := newTestRunAIInstallDataForHostRouting()

	step := NewInstallHostBasedRouting(mockClients)
	step.EvaluateStatus(runaiData)

	status := step.GetCachedStatus()
	if status == model.Installed {
		t.Error("expected non-Installed status when secret does not exist")
	}
}

func TestHostRouting_EvaluateStatus_SecretExistsButNoIngress(t *testing.T) {
	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      HostRoutingTlsSecretName,
			Namespace: RunaiClusterNamespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte("dummy-cert"),
			"tls.key": []byte("dummy-key"),
		},
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: RunaiClusterNamespace},
	}

	mockClients := setupMockK8sClientsForHostRouting(t, ns, existingSecret)
	runaiData := newTestRunAIInstallDataForHostRouting()

	step := NewInstallHostBasedRouting(mockClients)
	step.EvaluateStatus(runaiData)

	status := step.GetCachedStatus()
	if status == model.Installed {
		t.Error("expected non-Installed status when ingress does not exist")
	}
}

func TestHostRouting_EvaluateStatus_SkipsWhenTerminal(t *testing.T) {
	mockClients := setupMockK8sClientsForHostRouting(t)
	runaiData := newTestRunAIInstallDataForHostRouting()

	step := NewInstallHostBasedRouting(mockClients)
	step.SetStatus(model.Installed, nil)

	step.EvaluateStatus(runaiData)

	if step.GetCachedStatus() != model.Installed {
		t.Error("expected terminal status to be preserved")
	}
}

func TestHostRouting_Execute_ContextCancellation(t *testing.T) {
	mockClients := setupMockK8sClientsForHostRouting(t)
	runaiData := newTestRunAIInstallDataForHostRouting()

	step := NewInstallHostBasedRouting(mockClients)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := step.Execute(ctx, runaiData)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}

func TestHostRouting_Execute_SecretAlreadyExists_SkipsCreation(t *testing.T) {
	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      HostRoutingTlsSecretName,
			Namespace: RunaiClusterNamespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte("dummy-cert"),
			"tls.key": []byte("dummy-key"),
		},
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: RunaiClusterNamespace},
	}

	mockClients := setupMockK8sClientsForHostRouting(t, ns, existingSecret)
	runaiData := newTestRunAIInstallDataForHostRouting()

	step := NewInstallHostBasedRouting(mockClients)
	err := step.createTlsSecret(context.Background(), runaiData)

	if err != nil {
		t.Errorf("expected no error when secret already exists, got %v", err)
	}
}

func TestHostRouting_GetError_Initially(t *testing.T) {
	mockClients := setupMockK8sClientsForHostRouting(t)
	step := NewInstallHostBasedRouting(mockClients)

	if step.GetError() != nil {
		t.Errorf("expected no error initially, got %v", step.GetError())
	}
}

func TestHostRouting_GetError_AfterFailure(t *testing.T) {
	mockClients := setupMockK8sClientsForHostRouting(t)
	step := NewInstallHostBasedRouting(mockClients)

	testErr := fmt.Errorf("test error")
	step.SetStatus(model.Failed, testErr)

	if step.GetError() == nil || step.GetError().Error() != "test error" {
		t.Errorf("expected 'test error', got %v", step.GetError())
	}
}
