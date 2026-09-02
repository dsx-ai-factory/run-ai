/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/flow"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	installMocks "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step/mocks"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("InstallationScreen checkStatusAndAdvance", func() {
	It("sends a NextStepMsg with stepPhase Installed when current step is Installed", func() {
		// Arrange screen and runai data
		screen := NewInstallationScreen("Test Installation", nil)
		opts := &config.InstallConfig{
			DevMode:     false,
			StepTimeout: 5 * time.Minute,
			GPUType:     config.GPUTypeReal,
			LogsDir:     "/tmp",
		}
		runaiData, _ := model.NewRunAIInstallData(opts)
		screen.Init(runaiData)

		// Mock steps: first is Installed
		step1 := installMocks.NewMockInstallStep(GinkgoT())
		step1.On("GetName").Return("Step1").Maybe()
		step1.On("GetCachedStatus").Return(model.Installed).Maybe()
		step1.On("GetStatusAsync", mock.Anything).Return(model.Installed).Maybe()
		step1.On("PostInstall", runaiData).Return().Once()
		step1.On("GetError").Return(nil).Maybe()

		step2 := installMocks.NewMockInstallStep(GinkgoT())
		step2.On("GetName").Return("Step2").Maybe()
		step2.On("GetCachedStatus").Return(model.NotInstalled).Maybe()
		step2.On("GetStatusSync", mock.Anything).Return(model.NotInstalled).Maybe()
		step2.On("GetStatusAsync", mock.Anything).Return(model.NotInstalled).Maybe()
		step2.On("GetError").Return(nil).Maybe()

		step3 := installMocks.NewMockInstallStep(GinkgoT())
		step3.On("GetName").Return("Step3").Maybe()
		step3.On("GetCachedStatus").Return(model.NotInstalled).Maybe()
		step3.On("GetStatusSync", mock.Anything).Return(model.NotInstalled).Maybe()
		step3.On("GetStatusAsync", mock.Anything).Return(model.NotInstalled).Maybe()
		step3.On("GetError").Return(nil).Maybe()

		fl := flow.NewInstallFlow("MultiStepFlow",
			[]step.InstallStep{step1, step2, step3},
			runaiData,
		)
		screen.SetFlows([]*flow.InstallFlow{fl})

		// Act
		screen.checkStatusAndAdvance()

		// Assert - expect NextStepMsg on events channel
		var received tea.Msg
		select {
		case received = <-screen.events:
			// got a message
		case <-time.After(1 * time.Second):
			Fail("expected a message on events channel, but timed out")
		}

		nextMsg, ok := received.(flow.NextStepMsg)
		Expect(ok).To(BeTrue(), "expected NextStepMsg")
		Expect(nextMsg.StepPhase).To(Equal(model.Installed))

		// Verify PostInstall was called on the installed step
		step1.AssertCalled(GinkgoT(), "PostInstall", runaiData)
	})

	It("emits CurrentStepStatusCheckMsg only after Execute is signaled to finish, and then reports Installed on next call", func() {
		// Arrange: screen and runai data
		screen := NewInstallationScreen("Transition Flow", nil)
		opts := &config.InstallConfig{
			StepTimeout: 5 * time.Second,
			LogsDir:     "/tmp",
		}
		runaiData, _ := model.NewRunAIInstallData(opts)
		screen.Init(runaiData)

		// Mock a step starting as NotInstalled, then Installing, then Installed
		s := installMocks.NewMockInstallStep(GinkgoT())
		s.On("GetName").Return("Step1").Maybe()

		// Sequence of statuses we want to simulate across calls
		statuses := []model.InstallPhase{model.NotInstalled, model.Installing, model.Installed}
		var mu sync.Mutex
		current := 0
		s.On("GetCachedStatus").Return(func() model.InstallPhase {
			mu.Lock()
			defer mu.Unlock()
			return statuses[current]
		}).Maybe()
		s.On("GetStatusAsync", mock.Anything).Return(func(_ *model.RunAIInstallData) model.InstallPhase {
			// Return current status value and do not advance here (advance on Execute)
			mu.Lock()
			defer mu.Unlock()
			return statuses[current]
		}).Maybe()

		// Mock GetStatusSync for NotInstalled status
		s.On("GetStatusSync", mock.Anything).Return(func(_ *model.RunAIInstallData) model.InstallPhase {
			mu.Lock()
			defer mu.Unlock()
			return statuses[current]
		}).Maybe()

		// Mock TryStartExecute/EndExecute/SetStatus for executeCurrentStep
		s.On("TryStartExecute").Return(true).Maybe()
		s.On("EndExecute").Return().Maybe()
		s.On("SetStatus", model.Installing, nil).Return(model.Installing).Maybe()
		s.On("GetTimeout").Return(5 * time.Minute).Maybe()
		// Central timeout check is called for Installing status
		s.On("CheckTimeout").Return(nil).Maybe()

		// Execute should block until a signal, then advance Installing -> Installed
		proceedCh := make(chan struct{})
		s.On("Execute", mock.Anything, runaiData).Return(func(ctx context.Context, _ *model.RunAIInstallData) error {
			// Move to Installing and block until signaled or context cancelled
			mu.Lock()
			current = 1
			mu.Unlock()
			select {
			case <-proceedCh:
				mu.Lock()
				current = 2
				mu.Unlock()
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}).Once()

		s.On("PostInstall", runaiData).Return().Once()
		s.On("GetError").Return(nil).Maybe()

		fl := flow.NewInstallFlow("TransitionFlow",
			[]step.InstallStep{s},
			runaiData,
		)
		screen.SetFlows([]*flow.InstallFlow{fl})

		// Act: first call - NotInstalled starts execution in background
		screen.checkStatusAndAdvance()

		// Assert: no event should be emitted before we signal Execute to continue
		select {
		case <-screen.events:
			Fail("did not expect an event before proceed signal is sent")
		case <-time.After(150 * time.Millisecond):
			// ok, nothing received yet
		}

		// Second call while Execute is running and status is Installing
		screen.checkStatusAndAdvance()

		// Now signal Execute to finish, and expect CurrentStepStatusCheckMsg
		close(proceedCh)

		var received tea.Msg
		select {
		case received = <-screen.events:
			// got a message
		case <-time.After(2 * time.Second):
			Fail("expected a message on events channel after signaling proceed, but timed out")
		}
		_, isCheck := received.(flow.CurrentStepStatusCheckMsg)
		Expect(isCheck).To(BeTrue(), "expected CurrentStepStatusCheckMsg after Execute completes")

		// Third call after execution finished should now report Installed and emit NextStepMsg
		screen.checkStatusAndAdvance()

		var received2 tea.Msg
		select {
		case received2 = <-screen.events:
			// got a message
		case <-time.After(1 * time.Second):
			Fail("expected NextStepMsg on events channel, but timed out")
		}

		nextMsg, ok := received2.(flow.NextStepMsg)
		Expect(ok).To(BeTrue(), "expected NextStepMsg")
		Expect(nextMsg.StepPhase).To(Equal(model.Installed))

		// Verify PostInstall invoked
		s.AssertCalled(GinkgoT(), "PostInstall", runaiData)
	})

	Context("different step phases", func() {
		var screen *InstallationScreen
		var runaiData *model.RunAIInstallData

		BeforeEach(func() {
			screen = NewInstallationScreen("Phase Tests", nil)
			opts := &config.InstallConfig{
				DevMode:     false,
				StepTimeout: 5 * time.Minute,
				GPUType:     config.GPUTypeReal,
				LogsDir:     "/tmp",
			}
			runaiData, _ = model.NewRunAIInstallData(opts)
			screen.Init(runaiData)
		})

		It("sends NextStepMsg with Installed phase when step status is Installed", func() {
			// Arrange
			s := installMocks.NewMockInstallStep(GinkgoT())
			s.On("GetName").Return("InstalledStep").Maybe()
			s.On("GetCachedStatus").Return(model.Installed).Maybe()
			s.On("GetStatusAsync", mock.Anything).Return(model.Installed).Maybe()
			s.On("PostInstall", runaiData).Return().Once()
			s.On("GetError").Return(nil).Maybe()

			fl := flow.NewInstallFlow("TestFlow",
				[]step.InstallStep{s},
				runaiData,
			)
			screen.SetFlows([]*flow.InstallFlow{fl})

			// Act
			screen.checkStatusAndAdvance()

			// Assert - expect NextStepMsg on events channel
			var received tea.Msg
			select {
			case received = <-screen.events:
				// got a message
			case <-time.After(1 * time.Second):
				Fail("expected a message on events channel, but timed out")
			}

			nextMsg, ok := received.(flow.NextStepMsg)
			Expect(ok).To(BeTrue(), "expected NextStepMsg")
			Expect(nextMsg.StepPhase).To(Equal(model.Installed))
			s.AssertCalled(GinkgoT(), "PostInstall", runaiData)
		})

		It("sends NextStepMsg with Failed phase when step status is Failed", func() {
			// Arrange
			s := installMocks.NewMockInstallStep(GinkgoT())
			s.On("GetName").Return("FailedStep").Maybe()
			s.On("GetCachedStatus").Return(model.Failed).Maybe()
			s.On("GetStatusAsync", mock.Anything).Return(model.Failed).Maybe()
			s.On("GetError").Return(nil).Maybe()

			fl := flow.NewInstallFlow("TestFlow",
				[]step.InstallStep{s},
				runaiData,
			)
			screen.SetFlows([]*flow.InstallFlow{fl})

			// Act
			screen.checkStatusAndAdvance()

			// Assert - expect NextStepMsg on events channel
			var received tea.Msg
			select {
			case received = <-screen.events:
				// got a message
			case <-time.After(1 * time.Second):
				Fail("expected a message on events channel, but timed out")
			}

			nextMsg, ok := received.(flow.NextStepMsg)
			Expect(ok).To(BeTrue(), "expected NextStepMsg")
			Expect(nextMsg.StepPhase).To(Equal(model.Failed))
		})

		It("sends NextStepMsg with Timeout phase when step status is Timeout", func() {
			// Arrange
			s := installMocks.NewMockInstallStep(GinkgoT())
			s.On("GetName").Return("TimeoutStep").Maybe()
			s.On("GetCachedStatus").Return(model.Timeout).Maybe()
			s.On("GetStatusAsync", mock.Anything).Return(model.Timeout).Maybe()
			s.On("GetError").Return(nil).Maybe()

			fl := flow.NewInstallFlow("TestFlow",
				[]step.InstallStep{s},
				runaiData,
			)
			screen.SetFlows([]*flow.InstallFlow{fl})

			// Act
			screen.checkStatusAndAdvance()

			// Assert - expect NextStepMsg on events channel
			var received tea.Msg
			select {
			case received = <-screen.events:
				// got a message
			case <-time.After(1 * time.Second):
				Fail("expected a message on events channel, but timed out")
			}

			nextMsg, ok := received.(flow.NextStepMsg)
			Expect(ok).To(BeTrue(), "expected NextStepMsg")
			Expect(nextMsg.StepPhase).To(Equal(model.Timeout))
		})

		It("does not send messages when step status is NotInstalled and execution already started", func() {
			// Arrange
			s := installMocks.NewMockInstallStep(GinkgoT())
			s.On("GetName").Return("NotInstalledStep").Maybe()
			s.On("GetCachedStatus").Return(model.NotInstalled).Maybe()
			s.On("GetStatusSync", mock.Anything).Return(model.NotInstalled).Maybe()
			s.On("GetStatusAsync", mock.Anything).Return(model.NotInstalled).Maybe()
			s.On("Execute", mock.Anything, runaiData).Return(nil).Maybe()
			s.On("GetError").Return(nil).Maybe()
			// Mock TryStartExecute to prevent actual execution (returns false = already executing)
			s.On("TryStartExecute").Return(false).Maybe()

			fl := flow.NewInstallFlow("TestFlow",
				[]step.InstallStep{s},
				runaiData,
			)
			screen.SetFlows([]*flow.InstallFlow{fl})

			// Act
			screen.checkStatusAndAdvance()

			// Assert - no message should be sent when TryStartExecute returns false
			select {
			case <-screen.events:
				Fail("did not expect an event when step is NotInstalled and already executing")
			case <-time.After(150 * time.Millisecond):
				// ok, nothing received
			}
		})

		It("sends NextStepMsg with Timeout phase when step reports Timeout", func() {
			// Arrange - step handles its own timeout via GetStatusAsync
			s := installMocks.NewMockInstallStep(GinkgoT())
			s.On("GetName").Return("TimeoutStep").Maybe()
			// Step returns Timeout from GetStatusAsync (step handles its own timeout)
			s.On("GetCachedStatus").Return(model.Timeout).Maybe()
			s.On("GetStatusAsync", mock.Anything).Return(model.Timeout).Maybe()
			s.On("GetError").Return(nil).Maybe()

			fl := flow.NewInstallFlow("TestFlow",
				[]step.InstallStep{s},
				runaiData,
			)
			screen.SetFlows([]*flow.InstallFlow{fl})

			// Act
			screen.checkStatusAndAdvance()

			// Assert - expect NextStepMsg on events channel
			var received tea.Msg
			select {
			case received = <-screen.events:
				// got a message
			case <-time.After(1 * time.Second):
				Fail("expected a message on events channel, but timed out")
			}

			nextMsg, ok := received.(flow.NextStepMsg)
			Expect(ok).To(BeTrue(), "expected NextStepMsg when step reports Timeout")
			Expect(nextMsg.StepPhase).To(Equal(model.Timeout))
		})
	})

})
