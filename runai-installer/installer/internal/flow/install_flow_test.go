/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flow

import (
	"fmt"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/step/mocks"
)

func createTestRunAIData(stepTimeout time.Duration) *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:     false,
		StepTimeout: stepTimeout,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
	}
	data, _ := model.NewRunAIInstallData(opts)
	return data
}

func TestNewInstallFlow(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)
	steps := []step.InstallStep{mockStep1, mockStep2}
	flowName := "Test Flow"

	flow := NewInstallFlow(flowName, steps, runaiData)

	if flow == nil {
		t.Fatal("expected flow to be created, got nil")
	}
	if flow.GetName() != flowName {
		t.Errorf("expected flow name '%s', got '%s'", flowName, flow.GetName())
	}
	if flow.currentStepIndex != 0 {
		t.Errorf("expected currentStepIndex to be 0, got %d", flow.currentStepIndex)
	}
	if flow.GetStatus() != model.NotInstalled {
		t.Errorf("expected status NotInstalled, got %v", flow.GetStatus())
	}
	if len(flow.GetSteps()) != 2 {
		t.Errorf("expected 2 steps, got %d", len(flow.GetSteps()))
	}
}

func TestInstallFlow_GetName(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	flowName := "My Custom Flow"

	flow := NewInstallFlow(flowName, []step.InstallStep{}, runaiData)

	if flow.GetName() != flowName {
		t.Errorf("expected flow name '%s', got '%s'", flowName, flow.GetName())
	}
}

func TestInstallFlow_GetCurrentStep(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)
	steps := []step.InstallStep{mockStep1, mockStep2}

	flow := NewInstallFlow("Test Flow", steps, runaiData)

	currentStep := flow.GetCurrentStep()
	if currentStep != mockStep1 {
		t.Error("expected current step to be mockStep1")
	}
}

func TestInstallFlow_GetCurrentStep_NoSteps(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	flow := NewInstallFlow("Test Flow", []step.InstallStep{}, runaiData)

	currentStep := flow.GetCurrentStep()
	if currentStep != nil {
		t.Error("expected current step to be nil when no steps exist")
	}
}

func TestInstallFlow_GetCurrentStep_BeyondSteps(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep := mocks.NewMockInstallStep(t)
	mockStep.On("GetCachedStatus").Return(model.Installed)
	// GetStatusAsync is called on the step after incrementing index
	// Since we only have 1 step, after incrementing we're beyond all steps
	// so calculateFlowStatus returns early without calling GetStatusAsync
	steps := []step.InstallStep{mockStep}

	flow := NewInstallFlow("Test Flow", steps, runaiData)
	flow.ProceedToNextStep() // Move beyond the last step

	currentStep := flow.GetCurrentStep()
	if currentStep != nil {
		t.Error("expected current step to be nil when beyond all steps")
	}
}

func TestInstallFlow_ProceedToNextStep(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)
	mockStep3 := mocks.NewMockInstallStep(t)

	// GetCachedStatus is called by ProceedToNextStep to check if step is terminal
	mockStep1.On("GetCachedStatus").Return(model.Installed)
	mockStep2.On("GetCachedStatus").Return(model.Installed)
	mockStep3.On("GetCachedStatus").Return(model.Installed)

	// After incrementing index, GetStatusAsync is called on the NEW current step
	mockStep2.On("GetStatusAsync", runaiData).Return(model.NotInstalled)
	mockStep3.On("GetStatusAsync", runaiData).Return(model.NotInstalled)

	steps := []step.InstallStep{mockStep1, mockStep2, mockStep3}
	flow := NewInstallFlow("Test Flow", steps, runaiData)

	// Initially at step 0
	if flow.currentStepIndex != 0 {
		t.Errorf("expected currentStepIndex to be 0, got %d", flow.currentStepIndex)
	}

	// Move to step 1
	nextStep := flow.ProceedToNextStep()
	if nextStep != mockStep2 {
		t.Error("expected next step to be mockStep2")
	}
	if flow.currentStepIndex != 1 {
		t.Errorf("expected currentStepIndex to be 1, got %d", flow.currentStepIndex)
	}

	// Move to step 2
	nextStep = flow.ProceedToNextStep()
	if nextStep != mockStep3 {
		t.Error("expected next step to be mockStep3")
	}
	if flow.currentStepIndex != 2 {
		t.Errorf("expected currentStepIndex to be 2, got %d", flow.currentStepIndex)
	}

	// Move beyond last step
	nextStep = flow.ProceedToNextStep()
	if nextStep != nil {
		t.Error("expected next step to be nil when beyond all steps")
	}
	if flow.currentStepIndex != 3 {
		t.Errorf("expected currentStepIndex to be 3, got %d", flow.currentStepIndex)
	}
}

func TestInstallFlow_IsDone(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep := mocks.NewMockInstallStep(t)
	steps := []step.InstallStep{mockStep}

	flow := NewInstallFlow("Test Flow", steps, runaiData)

	// Not done initially
	if flow.IsDone() {
		t.Error("expected flow to not be done initially")
	}

	// Set to Installed status
	flow.status = model.Installed
	if !flow.IsDone() {
		t.Error("expected flow to be done when status is Installed")
	}

	// Set to Failed status
	flow.status = model.Failed
	if flow.IsDone() {
		t.Error("expected flow to not be done when status is Failed")
	}

	// Set to Installing status
	flow.status = model.Installing
	if flow.IsDone() {
		t.Error("expected flow to not be done when status is Installing")
	}
}

func TestInstallFlow_IsTerminalStatus(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep := mocks.NewMockInstallStep(t)
	steps := []step.InstallStep{mockStep}

	flow := NewInstallFlow("Test Flow", steps, runaiData)

	testCases := []struct {
		status      model.InstallPhase
		isTerminal  bool
		description string
	}{
		{model.NotInstalled, false, "NotInstalled should not be terminal"},
		{model.Installing, false, "Installing should not be terminal"},
		{model.Failed, true, "Failed should be terminal"},
		{model.Installed, true, "Installed should be terminal"},
		{model.Timeout, false, "Timeout should not be terminal for flow (only for steps)"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			flow.status = tc.status
			result := flow.IsTerminalStatus()
			if result != tc.isTerminal {
				t.Errorf("expected %v for status %v, got %v", tc.isTerminal, tc.status, result)
			}
		})
	}
}

func TestInstallFlow_GetStatus(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep := mocks.NewMockInstallStep(t)
	steps := []step.InstallStep{mockStep}

	flow := NewInstallFlow("Test Flow", steps, runaiData)

	// Test initial status
	if flow.GetStatus() != model.NotInstalled {
		t.Errorf("expected initial status to be NotInstalled, got %v", flow.GetStatus())
	}

	// Test status changes
	flow.status = model.Installing
	if flow.GetStatus() != model.Installing {
		t.Errorf("expected status to be Installing, got %v", flow.GetStatus())
	}

	flow.status = model.Installed
	if flow.GetStatus() != model.Installed {
		t.Errorf("expected status to be Installed, got %v", flow.GetStatus())
	}

	flow.status = model.Failed
	if flow.GetStatus() != model.Failed {
		t.Errorf("expected status to be Failed, got %v", flow.GetStatus())
	}
}

func TestInstallFlow_GetSteps(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)
	mockStep3 := mocks.NewMockInstallStep(t)
	steps := []step.InstallStep{mockStep1, mockStep2, mockStep3}

	flow := NewInstallFlow("Test Flow", steps, runaiData)

	retrievedSteps := flow.GetSteps()
	if len(retrievedSteps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(retrievedSteps))
	}
	if retrievedSteps[0] != mockStep1 {
		t.Error("expected first step to be mockStep1")
	}
	if retrievedSteps[1] != mockStep2 {
		t.Error("expected second step to be mockStep2")
	}
	if retrievedSteps[2] != mockStep3 {
		t.Error("expected third step to be mockStep3")
	}
}

func TestInstallFlow_CalculateFlowStatus_Failed(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)

	mockStep1.On("GetCachedStatus").Return(model.Failed)
	// GetStatusAsync is NOT called because calculateFlowStatus finds a previous failed step
	// and returns early

	steps := []step.InstallStep{mockStep1, mockStep2}
	flow := NewInstallFlow("Test Flow", steps, runaiData)

	flow.ProceedToNextStep()

	if flow.GetStatus() != model.Failed {
		t.Errorf("expected flow status to be Failed, got %v", flow.GetStatus())
	}
}

func TestInstallFlow_CalculateFlowStatus_Timeout(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)

	mockStep1.On("GetCachedStatus").Return(model.Timeout)
	// GetStatusAsync is NOT called because calculateFlowStatus finds a previous timed-out step
	// and returns early

	steps := []step.InstallStep{mockStep1, mockStep2}
	flow := NewInstallFlow("Test Flow", steps, runaiData)

	flow.ProceedToNextStep()

	if flow.GetStatus() != model.Failed {
		t.Errorf("expected flow status to be Failed when step times out, got %v", flow.GetStatus())
	}
}

func TestInstallFlow_CalculateFlowStatus_Installed(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)

	mockStep1.On("GetCachedStatus").Return(model.Installed)
	mockStep2.On("GetCachedStatus").Return(model.Installed)

	// After first proceed, we're at step 2 which hasn't been executed yet
	mockStep2.On("GetStatusAsync", runaiData).Return(model.NotInstalled).Once()
	// After second proceed, we check if last step is done and it is
	// (We're beyond all steps so GetStatusAsync isn't called)

	steps := []step.InstallStep{mockStep1, mockStep2}
	flow := NewInstallFlow("Test Flow", steps, runaiData)

	// Advance from step 1 to step 2
	flow.ProceedToNextStep()
	if flow.GetStatus() != model.Installing {
		t.Errorf("expected flow status to be Installing after advancing to step 2, got %v", flow.GetStatus())
	}

	// Advance from step 2 beyond all steps (all steps are Installed)
	flow.ProceedToNextStep()
	if flow.GetStatus() != model.Installed {
		t.Errorf("expected flow status to be Installed after completing last step, got %v", flow.GetStatus())
	}
}

func TestInstallFlow_CalculateFlowStatus_Installing(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)

	mockStep1.On("GetCachedStatus").Return(model.Installed)
	// After advancing, step 2 becomes current and its status is checked
	mockStep2.On("GetStatusAsync", runaiData).Return(model.NotInstalled)

	steps := []step.InstallStep{mockStep1, mockStep2}
	flow := NewInstallFlow("Test Flow", steps, runaiData)

	flow.ProceedToNextStep()

	if flow.GetStatus() != model.Installing {
		t.Errorf("expected flow status to be Installing, got %v", flow.GetStatus())
	}
}

func TestInstallFlow_GetCurrentStepError(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep := mocks.NewMockInstallStep(t)
	steps := []step.InstallStep{mockStep}

	flow := NewInstallFlow("Test Flow", steps, runaiData)

	// Initially no error
	err := flow.GetCurrentStepError()
	if err != nil {
		t.Errorf("expected no error initially, got %v", err)
	}

	// Set an error
	testErr := fmt.Errorf("test error")
	flow.currentStepError = testErr

	err = flow.GetCurrentStepError()
	if err == nil || err.Error() != "test error" {
		t.Errorf("expected 'test error', got %v", err)
	}
}

func TestInstallFlow_GetStepErrors(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)
	mockStep3 := mocks.NewMockInstallStep(t)

	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")

	// GetCachedStatus is used during rendering to avoid I/O
	mockStep1.On("GetCachedStatus").Return(model.Failed)
	mockStep1.On("GetError").Return(err1)

	mockStep2.On("GetCachedStatus").Return(model.Installed)
	// GetError should not be called for non-failed steps

	mockStep3.On("GetCachedStatus").Return(model.Failed)
	mockStep3.On("GetError").Return(err2)

	steps := []step.InstallStep{mockStep1, mockStep2, mockStep3}
	flow := NewInstallFlow("Test Flow", steps, runaiData)

	errors := flow.GetStepErrors()

	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
	}

	if errors[0].Error() != "error 1" {
		t.Errorf("expected 'error 1', got %v", errors[0])
	}

	if errors[1].Error() != "error 2" {
		t.Errorf("expected 'error 2', got %v", errors[1])
	}
}

func TestInstallFlow_GetStepErrors_NoErrors(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)

	// GetCachedStatus is used during rendering to avoid I/O
	mockStep1.On("GetCachedStatus").Return(model.Installed)
	mockStep2.On("GetCachedStatus").Return(model.Installing)

	steps := []step.InstallStep{mockStep1, mockStep2}
	flow := NewInstallFlow("Test Flow", steps, runaiData)

	errors := flow.GetStepErrors()

	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}
}

func TestInstallFlow_GetStepErrors_FailedWithNoError(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep := mocks.NewMockInstallStep(t)

	// GetCachedStatus is used during rendering to avoid I/O
	mockStep.On("GetCachedStatus").Return(model.Failed)
	mockStep.On("GetError").Return(nil)

	steps := []step.InstallStep{mockStep}
	flow := NewInstallFlow("Test Flow", steps, runaiData)

	errors := flow.GetStepErrors()

	// Should not include steps with Failed status but no error
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}
}

func TestInstallFlow_ConcurrentAccess(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)
	mockStep3 := mocks.NewMockInstallStep(t)

	// GetCachedStatus is used during rendering to avoid I/O
	mockStep1.On("GetCachedStatus").Return(model.Installed)
	mockStep2.On("GetCachedStatus").Return(model.Installed)
	mockStep3.On("GetCachedStatus").Return(model.Installed)

	steps := []step.InstallStep{mockStep1, mockStep2, mockStep3}
	flow := NewInstallFlow("Test Flow", steps, runaiData)

	done := make(chan bool, 3)

	// Goroutine 1: Read status repeatedly
	go func() {
		for i := 0; i < 100; i++ {
			_ = flow.GetStatus()
		}
		done <- true
	}()

	// Goroutine 2: Read current step repeatedly
	go func() {
		for i := 0; i < 100; i++ {
			_ = flow.GetCurrentStep()
		}
		done <- true
	}()

	// Goroutine 3: Read step errors repeatedly
	go func() {
		for i := 0; i < 100; i++ {
			_ = flow.GetStepErrors()
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	<-done
	<-done
	<-done

	// If we get here without a race condition, the test passes
}

func TestInstallFlow_CalculateFlowStatus_NilCurrentStep(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	flow := NewInstallFlow("Test Flow", []step.InstallStep{}, runaiData)

	// calculateFlowStatus should handle nil current step gracefully
	flow.calculateFlowStatus()

	// Status should remain NotInstalled
	if flow.GetStatus() != model.NotInstalled {
		t.Errorf("expected status to remain NotInstalled, got %v", flow.GetStatus())
	}
}

func TestInstallFlow_MultipleStepsWorkflow(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep2 := mocks.NewMockInstallStep(t)
	mockStep3 := mocks.NewMockInstallStep(t)

	// GetCachedStatus is called multiple times:
	// 1. By ProceedToNextStep to check if step is terminal
	// 2. By calculateFlowStatus to check previous steps for failures
	mockStep1.On("GetCachedStatus").Return(model.Installed)
	mockStep2.On("GetCachedStatus").Return(model.Installed)
	mockStep3.On("GetCachedStatus").Return(model.Installed)

	// GetStatusAsync is called on the NEW current step after each proceed
	// First proceed: step2 becomes current, it's Installing
	mockStep2.On("GetStatusAsync", runaiData).Return(model.Installing)
	// Second proceed: step3 becomes current, it's Installing
	mockStep3.On("GetStatusAsync", runaiData).Return(model.Installing)

	steps := []step.InstallStep{mockStep1, mockStep2, mockStep3}
	flow := NewInstallFlow("Test Flow", steps, runaiData)

	// Verify initial state
	if flow.GetCurrentStep() != mockStep1 {
		t.Error("expected current step to be mockStep1")
	}
	if flow.GetStatus() != model.NotInstalled {
		t.Errorf("expected initial status to be NotInstalled, got %v", flow.GetStatus())
	}
	if flow.IsDone() {
		t.Error("expected flow to not be done initially")
	}

	// Complete step 1
	flow.ProceedToNextStep()
	if flow.GetCurrentStep() != mockStep2 {
		t.Error("expected current step to be mockStep2")
	}
	if flow.GetStatus() != model.Installing {
		t.Errorf("expected status to be Installing, got %v", flow.GetStatus())
	}
	if flow.IsDone() {
		t.Error("expected flow to not be done after first step")
	}

	// Complete step 2
	flow.ProceedToNextStep()
	if flow.GetCurrentStep() != mockStep3 {
		t.Error("expected current step to be mockStep3")
	}
	if flow.GetStatus() != model.Installing {
		t.Errorf("expected status to be Installing, got %v", flow.GetStatus())
	}
	if flow.IsDone() {
		t.Error("expected flow to not be done after second step")
	}

	// Complete step 3 (last step)
	flow.ProceedToNextStep()
	if flow.GetCurrentStep() != nil {
		t.Error("expected current step to be nil after all steps completed")
	}
	if flow.GetStatus() != model.Installed {
		t.Errorf("expected status to be Installed, got %v", flow.GetStatus())
	}
	if !flow.IsDone() {
		t.Error("expected flow to be done after all steps completed")
	}
	if !flow.IsTerminalStatus() {
		t.Error("expected flow to have terminal status after all steps completed")
	}
}
