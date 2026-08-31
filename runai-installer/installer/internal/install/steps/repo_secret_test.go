/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newTestRunAIInstallData(ngcCreds string, devMode bool) *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:     devMode,
		StepTimeout: 10 * time.Minute,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
	}
	d, _ := model.NewRunAIInstallData(opts)
	d.Inputs[configinputs.InstallInputJfrogCreds] = ngcCreds
	return d
}

func TestGetName(t *testing.T) {
	mockClients := step.SetupMockK8sClients()
	inst := NewInstallRepoAccessSecret(mockClients)
	if inst.GetName() != "Install repo access secret" {
		t.Errorf("unexpected name: %s", inst.GetName())
	}
}

func TestGetStatus_SecretExists(t *testing.T) {
	mockClients := step.SetupMockK8sClients(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RunaiRegCredsSecret,
			Namespace: step.CPNamespace,
		},
	})
	runaiData := newTestRunAIInstallData("", false)
	inst := NewInstallRepoAccessSecret(mockClients)
	if inst.GetStatusAsync(runaiData) != model.Installed {
		t.Errorf("expected Installed, got %v", inst.GetStatusAsync(runaiData))
	}
}

func TestGetStatus_SecretNotFound(t *testing.T) {
	mockClients := step.SetupMockK8sClients()
	runaiData := newTestRunAIInstallData("", false)
	inst := NewInstallRepoAccessSecret(mockClients)
	if inst.GetStatusAsync(runaiData) != model.NotInstalled {
		t.Errorf("expected NotInstalled, got %v", inst.GetStatusAsync(runaiData))
	}
}

func TestExecute_SecretAlreadyExists(t *testing.T) {
	mockClients := step.SetupMockK8sClients(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RunaiRegCredsSecret,
			Namespace: step.CPNamespace,
		},
	})
	runaiData := newTestRunAIInstallData("", false)
	inst := NewInstallRepoAccessSecret(mockClients)
	err := inst.Execute(context.Background(), runaiData)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if inst.GetStatusAsync(runaiData) != model.Installed {
		t.Errorf("expected Installed, got %v", inst.GetStatusAsync(runaiData))
	}
}

func TestExecute_MissingNGCCreds(t *testing.T) {
	mockClients := step.SetupMockK8sClients()
	runaiData := newTestRunAIInstallData("", false)
	inst := NewInstallRepoAccessSecret(mockClients)
	err := inst.Execute(context.Background(), runaiData)
	if err == nil || err.Error() != "NGC credentials are required but not provided" {
		t.Errorf("expected missing creds error, got %v", err)
	}
	// After a failed execution, GetStatus() returns the cached Failed status
	// because Failed is a terminal status
	if inst.GetStatusAsync(runaiData) != model.Failed {
		t.Errorf("expected Failed, got %v", inst.GetStatusAsync(runaiData))
	}
}

func TestExecute_CreateSecretWithNGCCreds(t *testing.T) {
	mockClients := step.SetupMockK8sClients()
	rawCreds := "dummy-creds"
	ngcCredsEncoded := base64.StdEncoding.EncodeToString([]byte(rawCreds))
	runaiData := newTestRunAIInstallData(ngcCredsEncoded, false)
	inst := NewInstallRepoAccessSecret(mockClients)
	err := inst.Execute(context.Background(), runaiData)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if inst.GetStatusAsync(runaiData) != model.Installed {
		t.Errorf("expected Installed, got %v", inst.GetStatusAsync(runaiData))
	}
	secret := &corev1.Secret{}
	err = mockClients.GetClient().Get(context.Background(), client.ObjectKey{
		Namespace: step.CPNamespace,
		Name:      RunaiRegCredsSecret,
	}, secret)
	if err != nil {
		t.Errorf("expected secret to be created, got %v", err)
	}
	if string(secret.Data[".dockerconfigjson"]) != rawCreds {
		t.Errorf("unexpected secret data: %s", secret.Data[".dockerconfigjson"])
	}
}

func TestExecute_DevModeKubectl(t *testing.T) {
	mockClients := step.SetupMockK8sClients()
	err := os.Setenv("HOME", "/tmp")
	if err != nil {
		t.Fatalf("failed to set HOME env var: %v", err)
	}
	runaiData := newTestRunAIInstallData("", true)
	inst := NewInstallRepoAccessSecret(mockClients)
	origCmd := execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo")
	}
	defer func() { execCommand = origCmd }()
	err = inst.Execute(context.Background(), runaiData)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// After successful execution in dev mode, GetStatus() returns the cached Installed status
	// because Installed is a terminal status
	if inst.GetStatusAsync(runaiData) != model.Installed {
		t.Errorf("expected Installed, got %v", inst.GetStatusAsync(runaiData))
	}
}
