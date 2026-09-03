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

func TestNewGpuOperator(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	gpuOp := NewGpuOperator(mockK8sClient)

	if gpuOp == nil {
		t.Fatal("expected gpuOperator instance, got nil")
	}

	if gpuOp.PreReqImpl == nil {
		t.Fatal("expected PreReqImpl to be initialized")
	}
}

func TestGpuOperator_GetName(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	gpuOp := NewGpuOperator(mockK8sClient)

	expectedName := ids.PrerequisiteIdentifierGpuOperator
	if gpuOp.GetName() != expectedName {
		t.Errorf("expected name '%s', got '%s'", expectedName, gpuOp.GetName())
	}
}

func TestGpuOperator_GetVersion(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	gpuOp := NewGpuOperator(mockK8sClient)

	expectedVersion := "v25.3.0"
	if gpuOp.GetVersion() != expectedVersion {
		t.Errorf("expected version '%s', got '%s'", expectedVersion, gpuOp.GetVersion())
	}
}

func TestGpuOperator_IsMandatory(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	gpuOp := NewGpuOperator(mockK8sClient)

	if !gpuOp.IsMandatory() {
		t.Error("expected gpuOperator to be mandatory")
	}
}

func TestGpuOperator_IsDefaultSelected(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	gpuOp := NewGpuOperator(mockK8sClient)

	if !gpuOp.IsDefaultSelected() {
		t.Error("expected gpuOperator to be default selected")
	}
}

func TestGpuOperator_GetMandatoryInputs(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	gpuOp := NewGpuOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	inputs := gpuOp.GetMandatoryInputs(runaiData)

	if inputs == nil {
		t.Error("expected non-nil slice of mandatory inputs")
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 mandatory inputs, got %d", len(inputs))
	}
}

func TestGpuOperator_GetOptionalInputs(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	gpuOp := NewGpuOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	inputs := gpuOp.GetOptionalInputs(runaiData)

	if inputs == nil {
		t.Error("expected non-nil slice of optional inputs")
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 optional inputs, got %d", len(inputs))
	}
}

func TestGpuOperator_GetInstallSteps(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	gpuOp := NewGpuOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	steps := gpuOp.GetInstallSteps(mockK8sClient, nil, nil, runaiData)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}

func TestGpuOperator_GetInstallSteps_WithNilK8sClient(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	gpuOp := NewGpuOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	// Should not panic even with nil k8s client
	steps := gpuOp.GetInstallSteps(nil, nil, nil, runaiData)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}

func TestGpuOperator_GetInstallSteps_WithNilRunaiData(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	gpuOp := NewGpuOperator(mockK8sClient)

	// Should not panic even with nil runai data
	steps := gpuOp.GetInstallSteps(mockK8sClient, nil, nil, nil)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}
