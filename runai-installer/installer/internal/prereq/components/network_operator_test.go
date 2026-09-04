/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/ids"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient/mocks"
)

func TestNewNetworkOperator(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	networkOp := NewNetworkOperator(mockK8sClient)

	if networkOp == nil {
		t.Fatal("expected networkOperator instance, got nil")
	}

	if networkOp.PreReqImpl == nil {
		t.Fatal("expected PreReqImpl to be initialized")
	}
}

func TestNetworkOperator_GetName(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	networkOp := NewNetworkOperator(mockK8sClient)

	expectedName := ids.PrerequisiteIdentifierNetworkOperator
	if networkOp.GetName() != expectedName {
		t.Errorf("expected name '%s', got '%s'", expectedName, networkOp.GetName())
	}
}

func TestNetworkOperator_GetVersion(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	networkOp := NewNetworkOperator(mockK8sClient)

	expectedVersion := "v25.10.0"
	if networkOp.GetVersion() != expectedVersion {
		t.Errorf("expected version '%s', got '%s'", expectedVersion, networkOp.GetVersion())
	}
}

func TestNetworkOperator_IsMandatory(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	networkOp := NewNetworkOperator(mockK8sClient)

	if !networkOp.IsMandatory() {
		t.Error("expected networkOperator to be mandatory")
	}
}

func TestNetworkOperator_IsDefaultSelected(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	networkOp := NewNetworkOperator(mockK8sClient)

	if !networkOp.IsDefaultSelected() {
		t.Error("expected networkOperator to be default selected")
	}
}

func TestNetworkOperator_GetMandatoryInputs(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	networkOp := NewNetworkOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	inputs := networkOp.GetMandatoryInputs(runaiData)

	if inputs == nil {
		t.Error("expected non-nil slice of mandatory inputs")
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 mandatory inputs, got %d", len(inputs))
	}
}

func TestNetworkOperator_GetOptionalInputs(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	networkOp := NewNetworkOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	inputs := networkOp.GetOptionalInputs(runaiData)

	if inputs == nil {
		t.Error("expected non-nil slice of optional inputs")
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 optional inputs, got %d", len(inputs))
	}
}

func TestNetworkOperator_GetInstallSteps(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	networkOp := NewNetworkOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	steps := networkOp.GetInstallSteps(mockK8sClient, nil, nil, runaiData)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}

func TestNetworkOperator_GetInstallSteps_WithNilK8sClient(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	networkOp := NewNetworkOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	// Should not panic even with nil k8s client
	steps := networkOp.GetInstallSteps(nil, nil, nil, runaiData)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}

func TestNetworkOperator_GetInstallSteps_WithNilRunaiData(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	networkOp := NewNetworkOperator(mockK8sClient)

	// Should not panic even with nil runai data
	steps := networkOp.GetInstallSteps(mockK8sClient, nil, nil, nil)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}
