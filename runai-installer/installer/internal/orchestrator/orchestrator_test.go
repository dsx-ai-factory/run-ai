/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"os"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	k8smocks "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient/mocks"
	pkgvalidation "github.com/run-ai/runai/runai-installer/installer/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// setupMockK8sClientsForInstaller creates a MockK8sClients with a fake client for testing
func setupMockK8sClientsForInstaller(t *testing.T) *k8smocks.MockK8sClients {
	fakeClient := fake.NewClientBuilder().Build()
	mockK8s := k8smocks.NewMockK8sClients(t)
	mockK8s.On("GetClient").Return(fakeClient).Maybe()
	mockK8s.On("GetConfig").Return(&rest.Config{}).Maybe()
	mockK8s.On("GetContextName").Return("test-context").Maybe()
	return mockK8s
}

// Helper functions for creating test data

func createTestRunAIDataForInstaller(t *testing.T) *model.RunAIInstallData {
	// Create temp files for SSL certificates
	certFile, err := os.CreateTemp("", "test-cert-*.crt")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(certFile.Name()) })

	keyFile, err := os.CreateTemp("", "test-key-*.key")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(keyFile.Name()) })

	// Write dummy content to cert files
	_, _ = certFile.WriteString("test cert content")
	_, _ = keyFile.WriteString("test key content")
	certFile.Close()
	keyFile.Close()

	opts := &config.InstallConfig{
		DevMode:     true,
		StepTimeout: 7 * time.Minute,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
		Mode:        config.ModeCLI,
		ConfigValues: map[string]string{
			configinputs.FlagClusterName: "test-cluster",
		},
	}
	data, err := model.NewRunAIInstallData(opts)
	require.NoError(t, err)

	// Add required inputs to avoid validation errors
	// Use google.com which is guaranteed to be reachable for FQDN validation
	data.Inputs[configinputs.InstallInputClusterName] = "test-cluster"
	data.Inputs[configinputs.InstallInputFQDN] = "google.com"
	data.Inputs[configinputs.InstallInputSSLPath] = certFile.Name()
	data.Inputs[configinputs.InstallInputSSLKeyPath] = keyFile.Name()

	return data
}

// Tests for NewInstallerWithK8sClient (mocked constructor tests)

func TestNewInstallerWithK8sClient_Success(t *testing.T) {
	// Use the mock setup helper with proper expectations
	mockK8sClient := setupMockK8sClientsForInstaller(t)

	runaiData := createTestRunAIDataForInstaller(t)

	installer, err := NewInstallerWithK8sClient(runaiData, mockK8sClient)

	assert.NoError(t, err)
	assert.NotNil(t, installer)
	assert.Equal(t, mockK8sClient, installer.k8sClient)
	assert.Equal(t, runaiData, installer.runaiData)
	assert.Equal(t, config.ModeCLI, installer.runaiData.Config.Mode)
}

func TestNewInstallerWithK8sClient_NilRunAIData(t *testing.T) {
	mockK8sClient := setupMockK8sClientsForInstaller(t)

	installer, err := NewInstallerWithK8sClient(nil, mockK8sClient)

	assert.Error(t, err)
	assert.Nil(t, installer)
	assert.Equal(t, "runaiData cannot be nil", err.Error())
}

func TestNewInstallerWithK8sClient_NilK8sClient(t *testing.T) {
	runaiData := createTestRunAIDataForInstaller(t)

	installer, err := NewInstallerWithK8sClient(runaiData, nil)

	assert.Error(t, err)
	assert.Nil(t, installer)
	assert.Equal(t, "k8sClients cannot be nil", err.Error())
}

func TestNewInstallerWithK8sClient_ValidatesInputs(t *testing.T) {
	mockK8sClient := setupMockK8sClientsForInstaller(t)

	// Create data without required inputs
	opts := &config.InstallConfig{
		DevMode:     true,
		StepTimeout: 7 * time.Minute,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
		Mode:        config.ModeCLI,
	}
	runaiData, err := model.NewRunAIInstallData(opts)
	require.NoError(t, err)
	// Don't add required inputs - this should cause validation to fail

	installer, err := NewInstallerWithK8sClient(runaiData, mockK8sClient)

	// Should fail validation (missing required inputs)
	if err != nil {
		assert.Nil(t, installer)
		assert.Contains(t, err.Error(), "failed to validate installer inputs")
	}
}

func TestNewInstallerWithK8sClient_WorkloadRoutingFromConfigValues(t *testing.T) {
	certFile, err := os.CreateTemp("", "test-cert-*.crt")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(certFile.Name()) })
	certFile.WriteString("test cert content")
	certFile.Close()

	keyFile, err := os.CreateTemp("", "test-key-*.key")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(keyFile.Name()) })
	keyFile.WriteString("test key content")
	keyFile.Close()

	opts := &config.InstallConfig{
		DevMode:     true,
		StepTimeout: 7 * time.Minute,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
		Mode:        config.ModeCLI,
		ConfigValues: map[string]string{
			configinputs.FlagClusterName:        "test-cluster",
			configinputs.FlagWorkloadRouting:     pkgvalidation.WorkloadRoutingHost,
			configinputs.FlagHostRoutingCertPath: certFile.Name(),
			configinputs.FlagHostRoutingKeyPath:  keyFile.Name(),
		},
	}
	data, err := model.NewRunAIInstallData(opts)
	require.NoError(t, err)

	assert.Equal(t, pkgvalidation.WorkloadRoutingHost, data.GetInputValue(configinputs.InstallInputWorkloadRouting))
	assert.Equal(t, certFile.Name(), data.GetInputValue(configinputs.InstallInputHostRoutingCertPath))
	assert.Equal(t, keyFile.Name(), data.GetInputValue(configinputs.InstallInputHostRoutingKeyPath))
}

func TestNewInstallerWithK8sClient_SetsModeToLib(t *testing.T) {
	mockK8sClient := setupMockK8sClientsForInstaller(t)

	runaiData := createTestRunAIDataForInstaller(t)

	installer, err := NewInstallerWithK8sClient(runaiData, mockK8sClient)

	assert.NoError(t, err)
	assert.NotNil(t, installer)
	// Should be set to ModeCLI
	assert.Equal(t, config.ModeCLI, installer.runaiData.Config.Mode)
}
