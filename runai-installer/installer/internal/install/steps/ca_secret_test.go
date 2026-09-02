/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	k8smocks "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient/mocks"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newFakeClient(t *testing.T, objs ...runtime.Object) *k8smocks.MockK8sClients {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 to scheme: %v", err)
	}
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
	mockClients := k8smocks.NewMockK8sClients(t)
	mockClients.On("GetClient").Return(cl).Maybe()
	return mockClients
}

func TestInstallCAChainSecret_GetStatus_NotInstalled(t *testing.T) {
	secretName := "runai-ca-cert"
	namespace := "runai-backend"
	k8s := newFakeClient(t /* no objects */)

	step := NewInstallCAChainSecret(secretName, namespace, "runai-backend", nil, k8s)
	status := step.GetStatusAsync(&model.RunAIInstallData{})
	if status != model.NotInstalled {
		t.Fatalf("expected NotInstalled, got %v", status)
	}
}

func TestInstallCAChainSecret_GetStatus_Installed(t *testing.T) {
	secretName := "runai-ca-cert"
	namespace := "runai-backend"
	existing := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}
	k8s := newFakeClient(t, existing)

	step := NewInstallCAChainSecret(secretName, namespace, "runai-backend", nil, k8s)
	status := step.GetStatusAsync(&model.RunAIInstallData{})
	if status != model.Installed {
		t.Fatalf("expected Installed, got %v", status)
	}
}

func TestInstallCAChainSecret_Execute_SecretAlreadyExists(t *testing.T) {
	secretName := "runai-ca-cert"
	namespace := "runai-backend"
	existing := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}
	k8s := newFakeClient(t, existing)
	step := NewInstallCAChainSecret(secretName, namespace, "runai-backend", nil, k8s)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := step.Execute(ctx, &model.RunAIInstallData{})
	if err != nil {
		t.Fatalf("expected nil error when secret exists, got %v", err)
	}
}

func TestInstallCAChainSecret_Execute_CreateSecretSuccess(t *testing.T) {
	secretName := "runai-ca-cert"
	namespace := "runai-backend"
	k8s := newFakeClient(t /* no objects */)

	labels := map[string]string{"app": "runai", "scope": "cluster"}
	step := NewInstallCAChainSecret(secretName, namespace, "runai-backend", labels, k8s)

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cacert.pem")
	certContent := []byte("-----BEGIN CERTIFICATE-----\nABC\n-----END CERTIFICATE-----\n")
	if err := os.WriteFile(certPath, certContent, 0o600); err != nil {
		t.Fatalf("failed to write temp cert: %v", err)
	}

	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputCAChainPath: certPath,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := step.Execute(ctx, runaiData); err != nil {
		t.Fatalf("expected success creating secret, got %v", err)
	}

	// Verify secret was created with correct content and metadata
	got := &corev1.Secret{}
	if err := k8s.GetClient().Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: namespace,
	}, got); err != nil {
		t.Fatalf("secret not found after creation: %v", err)
	}
	if got.Type != corev1.SecretTypeOpaque {
		t.Fatalf("unexpected secret type: %v", got.Type)
	}
	if string(got.Data["runai-ca.pem"]) != string(certContent) {
		t.Fatalf("unexpected cert data in secret")
	}
	for k, v := range labels {
		if got.Labels[k] != v {
			t.Fatalf("missing/incorrect label %s=%s", k, v)
		}
	}
}

func TestInstallCAChainSecret_Execute_ReadFileError(t *testing.T) {
	secretName := "runai-ca-cert"
	namespace := "runai-backend"
	k8s := newFakeClient(t /* no objects */)
	step := NewInstallCAChainSecret(secretName, namespace, "runai-backend", nil, k8s)

	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputCAChainPath: "/path/does/not/exist",
			},
		},
	}
	ctx := context.Background()
	if err := step.Execute(ctx, runaiData); err == nil {
		t.Fatalf("expected error when reading missing file")
	}
}

func TestInstallCAChainSecret_Execute_ContextCancelled(t *testing.T) {
	secretName := "runai-ca-cert"
	namespace := "runai-backend"
	k8s := newFakeClient(t /* no objects */)
	step := NewInstallCAChainSecret(secretName, namespace, "runai-backend", nil, k8s)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := step.Execute(ctx, &model.RunAIInstallData{})
	if err == nil {
		t.Fatalf("expected error due to context cancellation")
	}
}
