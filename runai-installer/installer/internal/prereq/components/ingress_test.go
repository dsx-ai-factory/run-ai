/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/ids"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	configinput "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient/mocks"
)

func TestNewNgnixIngress(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	ingress := NewNgnixIngress(mockK8sClient)

	if ingress == nil {
		t.Fatal("expected ngnixIngress instance, got nil")
	}

	if ingress.PreReqImpl == nil {
		t.Fatal("expected PreReqImpl to be initialized")
	}
}

func TestNgnixIngress_GetName(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	ingress := NewNgnixIngress(mockK8sClient)

	expectedName := ids.PrerequisiteIdentifierIngress
	if ingress.GetName() != expectedName {
		t.Errorf("expected name '%s', got '%s'", expectedName, ingress.GetName())
	}
}

func TestNgnixIngress_GetVersion(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	ingress := NewNgnixIngress(mockK8sClient)

	expectedVersion := "4.13.0"
	if ingress.GetVersion() != expectedVersion {
		t.Errorf("expected version '%s', got '%s'", expectedVersion, ingress.GetVersion())
	}
}

func TestNgnixIngress_IsMandatory(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	ingress := NewNgnixIngress(mockK8sClient)

	if !ingress.IsMandatory() {
		t.Error("expected ngnixIngress to be mandatory")
	}
}

func TestNgnixIngress_IsDefaultSelected(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	ingress := NewNgnixIngress(mockK8sClient)

	if !ingress.IsDefaultSelected() {
		t.Error("expected ngnixIngress to be default selected")
	}
}

func TestNgnixIngress_GetMandatoryInputs(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	ingress := NewNgnixIngress(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	inputs := ingress.GetMandatoryInputs(runaiData)

	if inputs == nil {
		t.Error("expected non-nil slice of mandatory inputs")
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 mandatory inputs, got %d", len(inputs))
	}
}

func TestNgnixIngress_GetOptionalInputs(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	ingress := NewNgnixIngress(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	inputs := ingress.GetOptionalInputs(runaiData)

	if inputs == nil {
		t.Error("expected non-nil slice of optional inputs")
	}

	expectedCount := 1
	if len(inputs) != expectedCount {
		t.Errorf("expected %d optional input, got %d", expectedCount, len(inputs))
	}

	if len(inputs) > 0 && inputs[0] != configinput.InstallInputIngressControllerHTTPSPort {
		t.Errorf("expected optional input to be InstallInputIngressControllerHTTPSPort, got %v", inputs[0])
	}
}

func TestNgnixIngress_GetInstallSteps(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	ingress := NewNgnixIngress(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	steps := ingress.GetInstallSteps(mockK8sClient, nil, nil, runaiData)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}

func TestNgnixIngress_GetInstallSteps_WithNilK8sClient(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	ingress := NewNgnixIngress(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	// Should not panic even with nil k8s client
	steps := ingress.GetInstallSteps(nil, nil, nil, runaiData)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}

func TestNgnixIngress_GetInstallSteps_WithNilRunaiData(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	ingress := NewNgnixIngress(mockK8sClient)

	// Should not panic even with nil runai data
	steps := ingress.GetInstallSteps(mockK8sClient, nil, nil, nil)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}
