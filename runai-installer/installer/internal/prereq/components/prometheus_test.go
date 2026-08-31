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

func TestPrometheus_GetName(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	prometheus := NewPrometheus(mockK8sClient)

	expectedName := ids.PrerequisiteIdentifierPrometheus
	if prometheus.GetName() != expectedName {
		t.Errorf("expected name '%s', got '%s'", expectedName, prometheus.GetName())
	}
}

func TestPrometheus_GetVersion(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	prometheus := NewPrometheus(mockK8sClient)

	expectedVersion := "76.2.0"
	if prometheus.GetVersion() != expectedVersion {
		t.Errorf("expected version '%s', got '%s'", expectedVersion, prometheus.GetVersion())
	}
}

func TestPrometheus_IsMandatory(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	prometheus := NewPrometheus(mockK8sClient)

	if !prometheus.IsMandatory() {
		t.Error("expected prometheus to be mandatory")
	}
}

func TestPrometheus_IsDefaultSelected(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	prometheus := NewPrometheus(mockK8sClient)

	if !prometheus.IsDefaultSelected() {
		t.Error("expected prometheus to be default selected")
	}
}

func TestPrometheus_GetMandatoryInputs(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	prometheus := NewPrometheus(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	inputs := prometheus.GetMandatoryInputs(runaiData)

	if inputs == nil {
		t.Error("expected non-nil slice of mandatory inputs")
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 mandatory inputs, got %d", len(inputs))
	}
}

func TestPrometheus_GetOptionalInputs(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	prometheus := NewPrometheus(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	inputs := prometheus.GetOptionalInputs(runaiData)

	if inputs == nil {
		t.Error("expected non-nil slice of optional inputs")
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 optional inputs, got %d", len(inputs))
	}
}

func TestPrometheus_GetInstallSteps(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	prometheus := NewPrometheus(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	steps := prometheus.GetInstallSteps(mockK8sClient, nil, nil, runaiData)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}

func TestPrometheus_GetInstallSteps_WithNilK8sClient(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	prometheus := NewPrometheus(mockK8sClient)
	runaiData := &model.RunAIInstallData{}

	// Should not panic even with nil k8s client
	steps := prometheus.GetInstallSteps(nil, nil, nil, runaiData)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}

func TestPrometheus_GetInstallSteps_WithNilRunaiData(t *testing.T) {
	mockK8sClient := &mocks.MockK8sClients{}
	prometheus := NewPrometheus(mockK8sClient)

	// Should not panic even with nil runai data
	steps := prometheus.GetInstallSteps(mockK8sClient, nil, nil, nil)

	if steps == nil {
		t.Fatal("expected non-nil slice of install steps")
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 install step, got %d", len(steps))
	}
}
