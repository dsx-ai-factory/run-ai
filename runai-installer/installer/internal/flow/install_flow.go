/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flow

import (
	"sync"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
)

type InstallFlow struct {
	mu                   sync.Mutex
	runaiData            *model.RunAIInstallData
	steps                []step.InstallStep
	flowName             string
	currentStepIndex     int
	status               model.InstallPhase
	currentStepError     error
	nonFatalTimeoutSteps map[step.InstallStep]struct{}
}

func NewInstallFlow(flowName string, steps []step.InstallStep, runaiData *model.RunAIInstallData) *InstallFlow {
	installFlow := &InstallFlow{
		runaiData:        runaiData,
		steps:            steps,
		flowName:         flowName,
		currentStepIndex: 0,
		status:           model.NotInstalled,
	}
	return installFlow
}

// MarkStepTimeoutNonFatal marks a step so that a Timeout status from that step
// does not fail the flow. Call this from the flow builder for steps where a wait
// timeout is expected and subsequent steps handle cleanup (e.g. helm uninstall
// followed by ForceDeleteNamespace).
func (i *InstallFlow) MarkStepTimeoutNonFatal(s step.InstallStep) {
	if i.nonFatalTimeoutSteps == nil {
		i.nonFatalTimeoutSteps = make(map[step.InstallStep]struct{})
	}
	i.nonFatalTimeoutSteps[s] = struct{}{}
}

// IsCurrentStepTimeoutNonFatal reports whether the current step's timeout is
// non-fatal (i.e. a Timeout status should not fail the flow).
func (i *InstallFlow) IsCurrentStepTimeoutNonFatal() bool {
	currentStep := i.GetCurrentStep()
	if currentStep == nil {
		return false
	}
	_, ok := i.nonFatalTimeoutSteps[currentStep]
	return ok
}

func (i *InstallFlow) GetName() string {
	return i.flowName
}

func (i *InstallFlow) GetCurrentStep() step.InstallStep {
	if i.currentStepIndex < len(i.steps) {
		return i.steps[i.currentStepIndex]
	}
	return nil
}

func (i *InstallFlow) ProceedToNextStep() step.InstallStep {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Only advance if the current step is actually in a terminal state.
	// This prevents multiple nextStepMsg from advancing the flow multiple times.
	currentStep := i.GetCurrentStep()
	if currentStep != nil {
		status := currentStep.GetCachedStatus()
		isTerminal := status == model.Installed || status == model.Failed || status == model.Timeout
		if !isTerminal {
			// Current step not finished yet, don't advance
			return currentStep
		}
	}

	i.currentStepIndex++
	i.calculateFlowStatus()
	if i.currentStepIndex < len(i.steps) {
		return i.steps[i.currentStepIndex]
	}
	return nil
}

func (i *InstallFlow) IsDone() bool {
	return i.status == model.Installed
}

func (i *InstallFlow) IsTerminalStatus() bool {
	return i.status == model.Failed || i.status == model.Installed
}

func (i *InstallFlow) GetStatus() model.InstallPhase {
	return i.status
}

func (i *InstallFlow) GetSteps() []step.InstallStep {
	return i.steps
}

func (i *InstallFlow) calculateFlowStatus() {
	currentStep := i.GetCurrentStep()

	// If no steps at all, nothing to calculate
	if len(i.steps) == 0 {
		return
	}

	// Check PREVIOUS steps - if any failed or timed out, the flow has failed
	// This prevents the flow from showing as successful when later steps succeed
	// after an earlier step has failed/timed out
	// Only check steps that actually exist (bounds check)
	maxIdx := i.currentStepIndex
	if maxIdx > len(i.steps) {
		maxIdx = len(i.steps)
	}
	for idx := 0; idx < maxIdx; idx++ {
		s := i.steps[idx]
		stepStatus := s.GetCachedStatus()
		if stepStatus == model.Failed {
			i.status = model.Failed
			return
		}
		_, nonFatal := i.nonFatalTimeoutSteps[s]
		timeoutFailsTheFlow := !nonFatal
		if stepStatus == model.Timeout && timeoutFailsTheFlow {
			i.status = model.Failed
			return
		}
	}

	// If we're beyond all steps (currentStep is nil) and no previous failures,
	// then all steps completed successfully
	if currentStep == nil {
		i.status = model.Installed
		return
	}

	// If no previous failures, determine status based on current step
	status := model.Installing
	currentStepStatus := currentStep.GetStatusAsync(i.runaiData)
	if currentStepStatus == model.Failed || currentStepStatus == model.Timeout {
		status = model.Failed
	}
	if currentStepStatus == model.Installed && i.currentStepIndex == len(i.steps)-1 {
		status = model.Installed
	}
	i.status = status
}

// GetCurrentStepError returns the error from the current step execution
func (i *InstallFlow) GetCurrentStepError() error {
	return i.currentStepError
}

// GetStepErrors returns all errors from steps that have failed or timed out
func (i *InstallFlow) GetStepErrors() []error {
	var errors []error
	for _, s := range i.steps {
		// Use cached status to avoid I/O during rendering
		status := s.GetCachedStatus()
		if (status == model.Failed || status == model.Timeout) && s.GetError() != nil {
			errors = append(errors, s.GetError())
		}
	}
	return errors
}
