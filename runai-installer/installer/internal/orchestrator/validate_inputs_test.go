/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"os"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	k8smocks "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient/mocks"
	pkgvalidation "github.com/run-ai/runai/runai-installer/installer/pkg/validation"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func createTestRunAIInstallData(mode config.InstallerMode) *model.RunAIInstallData {
	cfg := &config.InstallConfig{
		Mode: mode,
	}
	return &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Prerequisites: []string{},
			Inputs:        make(map[configinputs.InstallInputType]string),
		},
		Config: cfg,
	}
}

func newFakeK8s(t *testing.T) *k8smocks.MockK8sClients {
	cl := fake.NewClientBuilder().Build()
	m := k8smocks.NewMockK8sClients(t)
	m.On("GetClient").Return(cl).Maybe()
	m.On("GetConfig").Return(nil).Maybe()
	m.On("GetContextName").Return("test-context").Maybe()
	return m
}

func TestValidateInputs_ModeTUI(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeTUI)
	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	assert.NoError(t, err, "ValidateInputs should return no error in TUI mode")
}

func TestValidateInputs_ModeLib_MissingMandatoryInput(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeCLI)

	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	_ = err
}

func TestValidateInputs_ModeLib_ValidInput(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputClusterName] = "my-test-cluster"

	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid value for input Cluster Name",
			"Valid input should not produce validation error")
	}
}

func TestValidateInputs_ModeLib_InvalidInput(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputClusterName] = "My_Invalid_Cluster!"

	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	assert.Error(t, err, "Invalid input should produce an error")
}

func TestValidateInputs_ModeLib_JfrogCredentials_Valid(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputJfrogCreds] = "dXNlcm5hbWU6cGFzc3dvcmQ="

	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid value for input JFROG credentials",
			"Valid Jfrog credentials should not produce validation error")
	}
}

func TestValidateInputs_ModeLib_JfrogCredentials_Invalid(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputJfrogCreds] = "invalid@credentials#"

	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	assert.Error(t, err, "Invalid Jfrog credentials should produce an error")
}

func TestValidateInputs_ModeLib_IngressPort_Valid(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputIngressControllerHTTPSPort] = "30443"

	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid value for input Ingress Controller HTTPS port",
			"Valid ingress port should not produce validation error")
	}
}

func TestValidateInputs_ModeLib_IngressPort_Invalid(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputIngressControllerHTTPSPort] = "8080"

	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	assert.Error(t, err, "Invalid ingress port should produce an error")
}

func TestValidateInputs_EmptyInputsMap(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	configItemsMap := make(map[configinputs.InstallInputType]*configinputs.ConfigurationItem)

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	if err != nil {
		assert.Contains(t, err.Error(), "configuration item not found",
			"Empty inputs map should produce 'configuration item not found' error")
	}
}

func TestValidateInputs_OptionalEmptyInput(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputIngressControllerHTTPSPort] = ""

	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid value for input Ingress Controller HTTPS port",
			"Empty optional input should not produce validation error")
	}
}

func TestValidateInputs_ModeLib_WorkloadRouting_ValidPath(t *testing.T) {
	mockK8s := newFakeK8s(t)
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputWorkloadRouting] = pkgvalidation.WorkloadRoutingPath
	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid value for input Workload Routing",
			"'path' should be a valid workload routing value")
	}
}

func TestValidateInputs_ModeLib_WorkloadRouting_ValidHost(t *testing.T) {
	mockK8s := newFakeK8s(t)
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputWorkloadRouting] = pkgvalidation.WorkloadRoutingHost
	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid value for input Workload Routing",
			"'host' should be a valid workload routing value")
	}
}

func TestValidateInputs_ModeLib_WorkloadRouting_Invalid(t *testing.T) {
	mockK8s := newFakeK8s(t)
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputWorkloadRouting] = "subdomain"

	certFile, err := os.CreateTemp("", "test-cert-*.pem")
	assert.NoError(t, err)
	t.Cleanup(func() { os.Remove(certFile.Name()) })
	certFile.WriteString("test cert")
	certFile.Close()

	keyFile, err := os.CreateTemp("", "test-key-*.pem")
	assert.NoError(t, err)
	t.Cleanup(func() { os.Remove(keyFile.Name()) })
	keyFile.WriteString("test key")
	keyFile.Close()

	runaiData.Inputs[configinputs.InstallInputFQDN] = "google.com"
	runaiData.Inputs[configinputs.InstallInputClusterName] = "test-cluster"
	runaiData.Inputs[configinputs.InstallInputSSLPath] = certFile.Name()
	runaiData.Inputs[configinputs.InstallInputSSLKeyPath] = keyFile.Name()
	runaiData.Inputs[configinputs.InstallInputJfrogCreds] = "dGVzdDp0ZXN0"
	configItemsMap := configinputs.GetConfigInputs()

	err = ValidateInputs(runaiData, mockK8s, configItemsMap)
	assert.Error(t, err, "Invalid workload routing value should produce an error")
	assert.Contains(t, err.Error(), "invalid value for input Workload Routing")
}

func TestValidateInputs_ModeLib_HostRouting_MissingCertPath(t *testing.T) {
	mockK8s := newFakeK8s(t)
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputWorkloadRouting] = pkgvalidation.WorkloadRoutingHost
	configItemsMap := configinputs.GetConfigInputs()

	err := ValidateInputs(runaiData, mockK8s, configItemsMap)
	assert.Error(t, err, "Missing host routing cert path should produce an error when routing=host")
}

func TestValidateInputs_ModeLib_HostRouting_CertPathProvided(t *testing.T) {
	mockK8s := newFakeK8s(t)
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputWorkloadRouting] = pkgvalidation.WorkloadRoutingHost

	certFile, err := os.CreateTemp("", "test-host-cert-*.pem")
	assert.NoError(t, err)
	t.Cleanup(func() { os.Remove(certFile.Name()) })
	certFile.WriteString("test cert")
	certFile.Close()

	keyFile, err := os.CreateTemp("", "test-host-key-*.pem")
	assert.NoError(t, err)
	t.Cleanup(func() { os.Remove(keyFile.Name()) })
	keyFile.WriteString("test key")
	keyFile.Close()

	runaiData.Inputs[configinputs.InstallInputHostRoutingCertPath] = certFile.Name()
	runaiData.Inputs[configinputs.InstallInputHostRoutingKeyPath] = keyFile.Name()
	configItemsMap := configinputs.GetConfigInputs()

	err = ValidateInputs(runaiData, mockK8s, configItemsMap)
	if err != nil {
		assert.NotContains(t, err.Error(), "missing mandatory input: Host routing SSL cert path",
			"Provided host routing cert path should not produce missing input error")
		assert.NotContains(t, err.Error(), "missing mandatory input: Host routing SSL key path",
			"Provided host routing key path should not produce missing input error")
	}
}
