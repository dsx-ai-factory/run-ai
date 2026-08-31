/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
)

type NextStepMsg struct {
	StepPhase model.InstallPhase
}

type CurrentStepStatusCheckMsg struct{}

type FlowStatusReportingMsg struct {
	FlowStatus model.InstallPhase
	FlowName   string
}

var DefaultPollInterval = 30 * time.Second

type FlowManager struct {
	flows          []*InstallFlow
	currentFlow    int
	done           chan struct{}
	externalEvents chan *FlowStatusReportingMsg
	verifyOnly     bool
	activeSteps    sync.WaitGroup
	pollInterval   time.Duration
}

func NewFlowManager(flows []*InstallFlow) *FlowManager {
	return NewFlowManagerWithExternalEvents(flows, nil, false)
}

func NewFlowManagerWithExternalEvents(flows []*InstallFlow, externalEvents chan *FlowStatusReportingMsg, verifyOnly bool) *FlowManager {
	manager := &FlowManager{
		flows:          flows,
		currentFlow:    0,
		done:           make(chan struct{}),
		externalEvents: externalEvents,
		verifyOnly:     verifyOnly,
		pollInterval:   DefaultPollInterval,
	}
	if len(flows) == 0 {
		manager.flows = []*InstallFlow{}
	}
	return manager
}

func (f *FlowManager) SetFlows(flows []*InstallFlow) {
	f.flows = flows
}

func (f *FlowManager) GetCurrentFlow() *InstallFlow {
	if f.currentFlow >= len(f.flows) {
		return nil
	}
	return f.flows[f.currentFlow]
}

func (f *FlowManager) GetCurrentFlowIndex() int {
	return f.currentFlow
}

func (f *FlowManager) GetCurrentFlowStep() step.InstallStep {
	if f.currentFlow >= len(f.flows) {
		return nil
	}
	return f.flows[f.currentFlow].GetCurrentStep()
}

func (f *FlowManager) ProceedToNextStep() {
	if f.currentFlow >= len(f.flows) {
		return
	}
	nextStep := f.flows[f.currentFlow].ProceedToNextStep()
	if nextStep == nil {
		flowStatus := f.flows[f.currentFlow].GetStatus()
		f.publishToExternalEvents(&FlowStatusReportingMsg{
			FlowStatus: flowStatus,
			FlowName:   f.flows[f.currentFlow].GetName(),
		})
		// Only advance to next flow if current flow succeeded
		// If flow failed or timed out, stop here
		if flowStatus == model.Failed || flowStatus == model.Timeout {
			logging.Errorf("[FlowManager] Flow '%s' failed, stopping execution", f.flows[f.currentFlow].GetName())
			for _, stepErr := range f.flows[f.currentFlow].GetStepErrors() {
				logging.Errorf("[FlowManager] Step error: %v", stepErr)
			}
			return
		}
		f.currentFlow++
	}
}

func (f *FlowManager) IsDone() bool {
	// Done if we've processed all flows OR if any flow has failed
	return f.currentFlow >= len(f.flows) || f.HasFailed()
}

func (f *FlowManager) Wait() {
	<-f.done
}

func (f *FlowManager) HasFailed() bool {
	for _, fl := range f.flows {
		status := fl.GetStatus()
		if status == model.Failed || status == model.Timeout {
			return true
		}
	}
	return false
}

func (f *FlowManager) publishToExternalEvents(event *FlowStatusReportingMsg) {
	if f.externalEvents != nil {
		f.externalEvents <- event
	}
}

func (f *FlowManager) ExecuteFlow(ctx context.Context, runaiData *model.RunAIInstallData, events chan interface{}) {
	f.initFlow(ctx, runaiData, events)
	go func() {
		defer func() {
			f.activeSteps.Wait()
			close(f.done)
		}()
		ticker := time.NewTicker(f.pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logging.Errorf("Flow manager cancelled via context timeout")
				if currentStep := f.GetCurrentFlowStep(); currentStep != nil {
					currentStep.SetStatus(model.Timeout, fmt.Errorf("installation cancelled: context timeout"))
					f.publishToExternalEvents(&FlowStatusReportingMsg{
						FlowStatus: model.Timeout,
						FlowName:   f.flows[f.currentFlow].GetName(),
					})
				}
				return
			case <-ticker.C:
				f.ProvisionStep(ctx, runaiData, events)
			case event := <-events:
				switch event.(type) {
				case NextStepMsg:
					f.ProceedToNextStep()
					if f.IsDone() {
						return
					}
					f.ProvisionStep(ctx, runaiData, events)
				case CurrentStepStatusCheckMsg:
					f.ProvisionStep(ctx, runaiData, events)
				}
			}
		}
	}()
}

func (f *FlowManager) initFlow(ctx context.Context, runaiData *model.RunAIInstallData, events chan interface{}) {
	currentStepInFlow := f.GetCurrentFlowStep()
	currentStepInFlow.GetStatusSync(runaiData)
	f.ProvisionStep(ctx, runaiData, events)
}

func (f *FlowManager) ProvisionStep(ctx context.Context, runaiData *model.RunAIInstallData, events chan interface{}) {
	currentStepInFlow := f.GetCurrentFlowStep()
	if currentStepInFlow == nil {
		return
	}
	status := currentStepInFlow.GetCachedStatus()
	if status == model.NotInstalled {
		status = currentStepInFlow.GetStatusSync(runaiData)
	} else {
		status = currentStepInFlow.GetStatusAsync(runaiData)
	}
	switch status {
	case model.NotInstalled:
		if f.verifyOnly {
			err := fmt.Errorf("step '%s' is not installed (verify-only mode)", currentStepInFlow.GetName())
			logging.Errorf("[Installation] %v", err)
			currentStepInFlow.SetStatus(model.Failed, err)
			events <- NextStepMsg{StepPhase: model.Failed}
		} else {
			f.activeSteps.Add(1)
			go executeCurrentStep(ctx, currentStepInFlow, runaiData, events, &f.activeSteps)
		}
	case model.Installing:
		// Central timeout check - ensures all steps have timeout protection
		// even if individual step forgets to implement it
		if err := currentStepInFlow.CheckTimeout(); err != nil {
			logging.Errorf("[Installation] %v", err)
			if hint := currentStepInFlow.GetTimeoutHint(); hint != "" {
				logging.Errorf("[Installation] %s", hint)
			}
			currentStepInFlow.SetStatus(model.Timeout, err)
			events <- NextStepMsg{StepPhase: model.Timeout}
		}
		f.publishToExternalEvents(&FlowStatusReportingMsg{
			FlowStatus: model.Installing,
			FlowName:   f.flows[f.currentFlow].GetName(),
		})
	case model.Installed:
		if !f.verifyOnly {
			currentStepInFlow.PostInstall(runaiData)
		}
		events <- NextStepMsg{StepPhase: status}
	case model.Failed, model.Timeout:
		currentFlow := f.flows[f.currentFlow]
		if status == model.Timeout && currentFlow.IsCurrentStepTimeoutNonFatal() {
			logging.Infof("[Installation] Step '%s' timed out (non-fatal), continuing flow", currentStepInFlow.GetName())
		} else {
			logging.Errorf("[Installation] Step '%s' ended with status '%v': %v", currentStepInFlow.GetName(), status, currentStepInFlow.GetError())
			f.publishToExternalEvents(&FlowStatusReportingMsg{
				FlowStatus: status,
				FlowName:   currentFlow.GetName(),
			})
		}
		events <- NextStepMsg{StepPhase: status}
	}
}

func executeCurrentStep(ctx context.Context, s step.InstallStep, runaiData *model.RunAIInstallData, events chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	if s == nil {
		return
	}
	// Guard against concurrent Execute() calls from multiple UI ticks racing
	// This also starts the execution timer automatically
	if !s.TryStartExecute() {
		return
	}
	defer s.EndExecute()
	// Set status to Installing immediately to prevent race conditions where
	// another tick might see NotInstalled and try to start another Execute().
	s.SetStatus(model.Installing, nil)
	// Scope the context to the step's timeout so helm (and any other blocking
	// operation inside Execute) is cancelled when the step deadline fires,
	// rather than waiting for the much longer execution-level context.
	stepCtx, cancel := context.WithTimeout(ctx, s.GetTimeout())
	defer cancel()
	if err := s.Execute(stepCtx, runaiData); err != nil {
		// Safety net: if Execute returned an error but left the status as Installing,
		// force it to Failed so ProvisionStep doesn't loop until the central timeout.
		if s.GetCachedStatus() == model.Installing {
			s.SetStatus(model.Failed, err)
		}
	}
	events <- CurrentStepStatusCheckMsg{}
}
