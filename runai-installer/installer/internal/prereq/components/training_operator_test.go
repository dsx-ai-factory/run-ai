/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/ids"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient/mocks"
)

func TestNewTrainingOperator(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	trainingOp := NewTrainingOperator(mockK8sClient)

	if trainingOp == nil {
		t.Fatal("expected trainingOperator instance, got nil")
	}

	if trainingOp.PreReqImpl == nil {
		t.Fatal("expected PreReqImpl to be initialized")
	}
}

func TestTrainingOperator_GetName(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	trainingOp := NewTrainingOperator(mockK8sClient)

	expectedName := ids.PrerequisiteIdentifierTrainingOperator
	if trainingOp.GetName() != expectedName {
		t.Errorf("expected name '%s', got '%s'", expectedName, trainingOp.GetName())
	}
}

func TestTrainingOperator_GetVersion(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	trainingOp := NewTrainingOperator(mockK8sClient)

	expectedVersion := "v1.8.1"
	if trainingOp.GetVersion() != expectedVersion {
		t.Errorf("expected version '%s', got '%s'", expectedVersion, trainingOp.GetVersion())
	}
}

func TestTrainingOperator_IsMandatory(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	trainingOp := NewTrainingOperator(mockK8sClient)

	if trainingOp.IsMandatory() {
		t.Error("expected trainingOperator to not be mandatory")
	}
}

func TestTrainingOperator_IsDefaultSelected(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	trainingOp := NewTrainingOperator(mockK8sClient)

	if !trainingOp.IsDefaultSelected() {
		t.Error("expected trainingOperator to be selected by default")
	}
}

func TestTrainingOperator_GetMandatoryInputs(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	trainingOp := NewTrainingOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	inputs := trainingOp.GetMandatoryInputs(runaiData)

	if inputs == nil {
		t.Error("expected non-nil slice of mandatory inputs")
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 mandatory inputs, got %d", len(inputs))
	}
}

func TestTrainingOperator_GetOptionalInputs(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	trainingOp := NewTrainingOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	inputs := trainingOp.GetOptionalInputs(runaiData)

	if inputs == nil {
		t.Error("expected non-nil slice of optional inputs")
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 optional inputs, got %d", len(inputs))
	}
}

func TestTrainingOperator_GetInstallSteps(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	trainingOp := NewTrainingOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	steps := trainingOp.GetInstallSteps(mockK8sClient, nil, nil, runaiData)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}

func TestTrainingOperator_GetInstallSteps_WithNilK8sClient(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	trainingOp := NewTrainingOperator(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	// Should not panic even with nil k8s client
	steps := trainingOp.GetInstallSteps(nil, nil, nil, runaiData)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}

func TestTrainingOperator_GetInstallSteps_WithNilRunaiData(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	trainingOp := NewTrainingOperator(mockK8sClient)

	// Should not panic even with nil runai data
	steps := trainingOp.GetInstallSteps(mockK8sClient, nil, nil, nil)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}
