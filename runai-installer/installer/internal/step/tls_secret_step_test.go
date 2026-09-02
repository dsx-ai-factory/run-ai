/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestRunAIInstallDataForTLS() *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:     false,
		StepTimeout: 10 * time.Minute,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
	}
	data, _ := model.NewRunAIInstallData(opts)
	// Set required inputs for TLS certificate generation
	data.Inputs[configinputs.InstallInputFQDN] = "test.example.com"
	return data
}

func TestNewInstallTlsSecret(t *testing.T) {
	mockClients := SetupMockK8sClients()
	runaiData := newTestRunAIInstallDataForTLS()

	step := NewInstallTlsSecret("test-tls-secret", "test-namespace", configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, mockClients)

	expectedName := "Install test-tls-secret secret"
	if step.GetName() != expectedName {
		t.Errorf("expected name '%s', got %s", expectedName, step.GetName())
	}

	if step.secretName != "test-tls-secret" {
		t.Errorf("expected secretName 'test-tls-secret', got %s", step.secretName)
	}

	if step.namespace != "test-namespace" {
		t.Errorf("expected namespace 'test-namespace', got %s", step.namespace)
	}

	if step.GetStatusAsync(runaiData) != model.NotInstalled {
		t.Errorf("expected status NotInstalled, got %v", step.GetStatusAsync(runaiData))
	}
}

func TestInstallTlsSecret_GetStatus_SecretExists(t *testing.T) {
	// Create a TLS secret that already exists
	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-tls-secret",
			Namespace: "test-namespace",
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte("dummy-cert"),
			"tls.key": []byte("dummy-key"),
		},
	}

	mockClients := SetupMockK8sClients(existingSecret)
	runaiData := newTestRunAIInstallDataForTLS()
	step := NewInstallTlsSecret("existing-tls-secret", "test-namespace", configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, mockClients)

	status := step.GetStatusAsync(runaiData)
	if status != model.Installed {
		t.Errorf("expected Installed status for existing secret, got %v", status)
	}
}

func TestInstallTlsSecret_GetStatus_SecretNotFound(t *testing.T) {
	mockClients := SetupMockK8sClients()
	runaiData := newTestRunAIInstallDataForTLS()
	step := NewInstallTlsSecret("non-existing-secret", "test-namespace", configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, mockClients)

	status := step.GetStatusAsync(runaiData)
	if status != model.NotInstalled {
		t.Errorf("expected NotInstalled status for non-existing secret, got %v", status)
	}
}

func TestInstallTlsSecret_Execute_SecretAlreadyExists(t *testing.T) {
	// Create a TLS secret that already exists
	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-tls-secret",
			Namespace: "test-namespace",
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte("dummy-cert"),
			"tls.key": []byte("dummy-key"),
		},
	}

	mockClients := SetupMockK8sClients(existingSecret)
	runaiData := newTestRunAIInstallDataForTLS()
	step := NewInstallTlsSecret("existing-tls-secret", "test-namespace", configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, mockClients)

	err := step.Execute(context.Background(), runaiData)
	if err != nil {
		t.Errorf("expected no error when secret already exists, got %v", err)
	}

	if step.GetStatusAsync(runaiData) != model.Installed {
		t.Errorf("expected Installed status after execution, got %v", step.GetStatusAsync(runaiData))
	}
}

func TestInstallTlsSecret_Execute_ContextCancellation(t *testing.T) {
	mockClients := SetupMockK8sClients()
	runaiData := newTestRunAIInstallDataForTLS()
	step := NewInstallTlsSecret("test-tls-secret", "test-namespace", configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, mockClients)

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

func TestInstallTlsSecret_GetError(t *testing.T) {
	mockClients := SetupMockK8sClients()
	step := NewInstallTlsSecret("test-tls-secret", "test-namespace", configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, mockClients)

	// Initially no error
	if step.GetError() != nil {
		t.Errorf("expected no error initially, got %v", step.GetError())
	}

	// Set an error and verify it's returned
	testErr := fmt.Errorf("test error")
	step.SetStatus(model.Failed, testErr)

	if step.GetError() == nil || step.GetError().Error() != "test error" {
		t.Errorf("expected 'test error', got %v", step.GetError())
	}
}

// Note: Testing the actual certificate creation would require mocking the cert.GetCertHandler
// function, which is more complex and would require additional setup. The tests above cover
// the main logic paths that can be tested without external dependencies.
