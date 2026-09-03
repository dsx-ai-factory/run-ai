/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/components"
	k8smocks "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient/mocks"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func newFakeK8sWithSecret(t *testing.T, savedInputs map[configinputs.InstallInputType]string) *k8smocks.MockK8sClients {
	var objects []client.Object
	if savedInputs != nil {
		inputsJSON, err := json.Marshal(savedInputs)
		assert.NoError(t, err, "Failed to marshal saved inputs")

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runai-installer-session",
				Namespace: "default",
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"inputs.json": inputsJSON,
			},
		}
		objects = append(objects, secret)
	}

	cl := fake.NewClientBuilder().WithObjects(objects...).Build()
	m := k8smocks.NewMockK8sClients(t)
	m.On("GetClient").Return(cl).Maybe()
	m.On("GetConfig").Return(nil).Maybe()
	m.On("GetContextName").Return("test-context").Maybe()
	return m
}

func TestValidateAndSetConfiguredPreReq_EmptyPrerequisites(t *testing.T) {
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	k8sClient := newFakeK8s(t)

	ingress := components.NewNgnixIngress(k8sClient)
	lws := components.NewLWS(k8sClient)

	preReqMap := map[string]components.Prerequisite{
		ingress.GetName(): ingress,
		lws.GetName():     lws,
	}

	validateAndSetConfiguredPreReq(runaiData, preReqMap)

	assert.Contains(t, runaiData.Prerequisites, ingress.GetName(),
		"Mandatory prerequisites should be added automatically")

	assert.NotContains(t, runaiData.Prerequisites, lws.GetName(),
		"Optional prerequisites should not be added automatically")
}

func TestValidateAndSetConfiguredPreReq_UnknownPrerequisite(t *testing.T) {
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Config.PreReq = []string{"unknown-prereq"}
	k8sClient := newFakeK8s(t)

	lws := components.NewLWS(k8sClient)

	preReqMap := map[string]components.Prerequisite{
		lws.GetName(): lws,
	}

	validateAndSetConfiguredPreReq(runaiData, preReqMap)
	assert.Contains(t, runaiData.Prerequisites, "unknown-prereq")
}

func TestValidateAndSetConfiguredPreReq_AllPrerequisites(t *testing.T) {
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Config.PreReq = []string{"all"}
	k8sClient := newFakeK8s(t)

	ingress := components.NewNgnixIngress(k8sClient)
	prometheus := components.NewPrometheus(k8sClient)
	lws := components.NewLWS(k8sClient)

	preReqMap := map[string]components.Prerequisite{
		ingress.GetName():    ingress,
		prometheus.GetName(): prometheus,
		lws.GetName():        lws,
	}

	validateAndSetConfiguredPreReq(runaiData, preReqMap)

	assert.Contains(t, runaiData.Prerequisites, ingress.GetName(),
		"All mandatory prerequisites should be included")
	assert.Contains(t, runaiData.Prerequisites, prometheus.GetName(),
		"All mandatory prerequisites should be included")

	assert.NotContains(t, runaiData.Prerequisites, lws.GetName(),
		"Optional prerequisites should not be included with 'all'")
}

func TestValidateAndSetConfiguredPreReq_ValidPrerequisites(t *testing.T) {
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	k8sClient := newFakeK8s(t)

	lws := components.NewLWS(k8sClient)
	knative := components.NewKnative(k8sClient)
	ingress := components.NewNgnixIngress(k8sClient)

	runaiData.Config.PreReq = []string{lws.GetName(), knative.GetName()}

	preReqMap := map[string]components.Prerequisite{
		lws.GetName():     lws,
		knative.GetName(): knative,
		ingress.GetName(): ingress,
	}

	validateAndSetConfiguredPreReq(runaiData, preReqMap)

	assert.Contains(t, runaiData.Prerequisites, lws.GetName())
	assert.Contains(t, runaiData.Prerequisites, knative.GetName())

	assert.Contains(t, runaiData.Prerequisites, ingress.GetName())
}

func TestFillAllPreReq(t *testing.T) {
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	k8sClient := newFakeK8s(t)

	ingress := components.NewNgnixIngress(k8sClient)
	prometheus := components.NewPrometheus(k8sClient)
	lws := components.NewLWS(k8sClient)

	preReqMap := map[string]components.Prerequisite{
		"key1": ingress,
		"key2": prometheus,
		"key3": lws,
	}

	fillAllPreReq(runaiData, preReqMap)

	assert.Contains(t, runaiData.Prerequisites, ingress.GetName())
	assert.Contains(t, runaiData.Prerequisites, prometheus.GetName())

	assert.NotContains(t, runaiData.Prerequisites, lws.GetName())
}

func TestAlignInstallData(t *testing.T) {
	mockK8s := newFakeK8s(t)

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Prerequisites = []string{}

	AlignInstallData(runaiData, mockK8s)

	assert.NotNil(t, runaiData.Prerequisites, "Prerequisites should be initialized")
}

func TestMergeInputData_NoSavedInputs_NewInputsPreserved(t *testing.T) {
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputClusterName] = "my-cluster"
	runaiData.Inputs[configinputs.InstallInputFQDN] = "example.com"

	mockK8s := newFakeK8s(t)

	mergeInputData(runaiData, mockK8s)

	assert.Equal(t, "my-cluster", runaiData.Inputs[configinputs.InstallInputClusterName])
	assert.Equal(t, "example.com", runaiData.Inputs[configinputs.InstallInputFQDN])
	assert.Len(t, runaiData.Inputs, 2)
}

func TestMergeInputData_SavedInputsExist_NoNewInputs(t *testing.T) {
	savedInputs := map[configinputs.InstallInputType]string{
		configinputs.InstallInputClusterName: "saved-cluster",
		configinputs.InstallInputFQDN:        "saved.example.com",
	}

	runaiData := createTestRunAIInstallData(config.ModeCLI)

	mockK8s := newFakeK8sWithSecret(t, savedInputs)

	mergeInputData(runaiData, mockK8s)

	assert.Equal(t, "saved-cluster", runaiData.Inputs[configinputs.InstallInputClusterName])
	assert.Equal(t, "saved.example.com", runaiData.Inputs[configinputs.InstallInputFQDN])
	assert.Len(t, runaiData.Inputs, 2)
}

func TestMergeInputData_NewInputsOverrideSavedInputs(t *testing.T) {
	savedInputs := map[configinputs.InstallInputType]string{
		configinputs.InstallInputClusterName: "saved-cluster",
		configinputs.InstallInputFQDN:        "saved.example.com",
	}

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputClusterName] = "new-cluster"

	mockK8s := newFakeK8sWithSecret(t, savedInputs)

	mergeInputData(runaiData, mockK8s)

	assert.Equal(t, "new-cluster", runaiData.Inputs[configinputs.InstallInputClusterName])
	assert.Equal(t, "saved.example.com", runaiData.Inputs[configinputs.InstallInputFQDN])
	assert.Len(t, runaiData.Inputs, 2)
}

func TestMergeInputData_MergesSavedAndNewInputs(t *testing.T) {
	savedInputs := map[configinputs.InstallInputType]string{
		configinputs.InstallInputClusterName: "saved-cluster",
	}

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputFQDN] = "new.example.com"

	mockK8s := newFakeK8sWithSecret(t, savedInputs)

	mergeInputData(runaiData, mockK8s)

	assert.Equal(t, "saved-cluster", runaiData.Inputs[configinputs.InstallInputClusterName])
	assert.Equal(t, "new.example.com", runaiData.Inputs[configinputs.InstallInputFQDN])
	assert.Len(t, runaiData.Inputs, 2)
}

func TestMergeInputData_EmptyInputsAndNoSecret(t *testing.T) {
	runaiData := createTestRunAIInstallData(config.ModeCLI)

	mockK8s := newFakeK8s(t)

	mergeInputData(runaiData, mockK8s)

	assert.Empty(t, runaiData.Inputs)
}

func TestMergeInputData_EmptySecretData(t *testing.T) {
	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputClusterName] = "my-cluster"

	emptyInputs := map[configinputs.InstallInputType]string{}
	mockK8s := newFakeK8sWithSecret(t, emptyInputs)

	mergeInputData(runaiData, mockK8s)

	assert.Equal(t, "my-cluster", runaiData.Inputs[configinputs.InstallInputClusterName])
	assert.Len(t, runaiData.Inputs, 1)
}

func TestMergeInputData_MultipleInputsOverride(t *testing.T) {
	savedInputs := map[configinputs.InstallInputType]string{
		configinputs.InstallInputClusterName:                "saved-cluster",
		configinputs.InstallInputFQDN:                       "saved.example.com",
		configinputs.InstallInputIngressControllerHTTPSPort: "30443",
	}

	runaiData := createTestRunAIInstallData(config.ModeCLI)
	runaiData.Inputs[configinputs.InstallInputClusterName] = "new-cluster"
	runaiData.Inputs[configinputs.InstallInputFQDN] = "new.example.com"

	mockK8s := newFakeK8sWithSecret(t, savedInputs)

	mergeInputData(runaiData, mockK8s)

	assert.Equal(t, "new-cluster", runaiData.Inputs[configinputs.InstallInputClusterName])
	assert.Equal(t, "new.example.com", runaiData.Inputs[configinputs.InstallInputFQDN])
	assert.Equal(t, "30443", runaiData.Inputs[configinputs.InstallInputIngressControllerHTTPSPort])
	assert.Len(t, runaiData.Inputs, 3)
}
