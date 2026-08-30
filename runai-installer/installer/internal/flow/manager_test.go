package flow

import (
	"context"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/step/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTestFlowManager() *FlowManager {
	return NewFlowManager(nil)
}

func createTestFlow(name string, steps []step.InstallStep, runaiData *model.RunAIInstallData) *InstallFlow {
	return NewInstallFlow(name, steps, runaiData)
}

func createMockStep(t *testing.T, name string, status model.InstallPhase) *mocks.MockInstallStep {
	mockStep := mocks.NewMockInstallStep(t)
	mockStep.On("GetName").Return(name).Maybe()
	mockStep.On("GetStatusAsync", mock.Anything).Return(status).Maybe()
	mockStep.On("GetCachedStatus").Return(status).Maybe()
	mockStep.On("GetError").Return(nil).Maybe()
	mockStep.On("GetElapsedTime").Return(time.Duration(0)).Maybe()
	mockStep.On("GetTimeout").Return(5 * time.Minute).Maybe()
	return mockStep
}

func TestNewFlowManager(t *testing.T) {
	fm := NewFlowManager([]*InstallFlow{})

	assert.NotNil(t, fm)
	assert.NotNil(t, fm.flows)
	assert.Equal(t, 0, fm.GetCurrentFlowIndex())
	assert.Equal(t, 0, len(fm.flows))
}

func TestFlowManager_SetFlows(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.NotInstalled)
	mockStep2 := createMockStep(t, "Step2", model.NotInstalled)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	flows := []*InstallFlow{flow1, flow2}
	fm.SetFlows(flows)

	assert.Equal(t, 2, len(fm.flows))
	assert.Equal(t, flow1, fm.flows[0])
	assert.Equal(t, flow2, fm.flows[1])
}

func TestFlowManager_GetCurrentFlow(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installed)
	mockStep2 := createMockStep(t, "Step2", model.Installed)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	fm.SetFlows([]*InstallFlow{flow1, flow2})

	t.Run("returns first flow initially", func(t *testing.T) {
		currentFlow := fm.GetCurrentFlow()
		assert.Equal(t, flow1, currentFlow)
	})

	t.Run("returns nil when beyond all flows", func(t *testing.T) {
		// Manually advance to beyond all flows (both flows succeed)
		fm.ProceedToNextStep()
		fm.ProceedToNextStep()
		currentFlow := fm.GetCurrentFlow()
		assert.Nil(t, currentFlow)
	})

	t.Run("stops at failed flow and does not advance", func(t *testing.T) {
		fm2 := createTestFlowManager()
		mockStep3 := createMockStep(t, "Step3", model.Installed)
		mockStep4 := createMockStep(t, "Step4", model.Failed)

		flow3 := createTestFlow("Flow3", []step.InstallStep{mockStep3}, runaiData)
		flow4 := createTestFlow("Flow4", []step.InstallStep{mockStep4}, runaiData)

		fm2.SetFlows([]*InstallFlow{flow3, flow4})

		// Advance past first flow (succeeds)
		fm2.ProceedToNextStep()
		assert.Equal(t, flow4, fm2.GetCurrentFlow())

		// Try to advance past second flow (fails) - should NOT advance
		fm2.ProceedToNextStep()
		assert.Equal(t, flow4, fm2.GetCurrentFlow(), "Should still be at failed flow, not advance")
		assert.True(t, fm2.IsDone(), "Should be done when flow fails")
		assert.True(t, fm2.HasFailed(), "Should report as failed")
	})
}

func TestFlowManager_GetCurrentFlowStep(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installed)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	fm.SetFlows([]*InstallFlow{flow1})

	t.Run("returns first step initially", func(t *testing.T) {
		step := fm.GetCurrentFlowStep()
		assert.Equal(t, mockStep1, step)
	})

	t.Run("returns nil when beyond all flows", func(t *testing.T) {
		// Manually advance to beyond all flows
		fm.ProceedToNextStep()
		step := fm.GetCurrentFlowStep()
		assert.Nil(t, step)
	})
}

func TestFlowManager_ProceedToNextStep(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installed)
	mockStep2 := createMockStep(t, "Step2", model.NotInstalled)
	mockStep3 := createMockStep(t, "Step3", model.NotInstalled)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2, mockStep3}, runaiData)

	fm.SetFlows([]*InstallFlow{flow1, flow2})

	t.Run("proceeds to next step in same flow", func(t *testing.T) {
		assert.Equal(t, 0, fm.GetCurrentFlowIndex())
		currentFlow := fm.GetCurrentFlow()
		assert.Equal(t, flow1, currentFlow)

		// Proceed to next step (which will move to next flow since flow1 has only 1 step)
		fm.ProceedToNextStep()

		assert.Equal(t, 1, fm.GetCurrentFlowIndex())
		currentFlow = fm.GetCurrentFlow()
		assert.Equal(t, flow2, currentFlow)
	})

	t.Run("does nothing when beyond all flows", func(t *testing.T) {
		// Advance to beyond all flows
		fm.ProceedToNextStep() // Will move to end
		currentIdx := fm.GetCurrentFlowIndex()
		fm.ProceedToNextStep()
		assert.Equal(t, currentIdx, fm.GetCurrentFlowIndex())
	})
}

func TestFlowManager_IsDone(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := createMockStep(t, "Step1", model.NotInstalled)
	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep}, runaiData)

	fm.SetFlows([]*InstallFlow{flow1})

	t.Run("not done when flows remain", func(t *testing.T) {
		assert.False(t, fm.IsDone())
	})

	t.Run("done when beyond all flows", func(t *testing.T) {
		fm2 := createTestFlowManager()
		mockStep2 := createMockStep(t, "Step1", model.Installed)
		flow2 := createTestFlow("Flow1", []step.InstallStep{mockStep2}, runaiData)
		fm2.SetFlows([]*InstallFlow{flow2})

		fm2.ProceedToNextStep()
		assert.True(t, fm2.IsDone())
	})
}

func TestFlowManager_MultipleFlows_CompleteWorkflow(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	// Create multiple flows with multiple steps
	mockStep1 := createMockStep(t, "Flow1-Step1", model.Installed)
	mockStep2 := createMockStep(t, "Flow1-Step2", model.Installed)
	mockStep3 := createMockStep(t, "Flow2-Step1", model.Installed)

	flow1 := createTestFlow("PrerequisitesFlow", []step.InstallStep{mockStep1, mockStep2}, runaiData)
	flow2 := createTestFlow("RunAIFlow", []step.InstallStep{mockStep3}, runaiData)

	fm.SetFlows([]*InstallFlow{flow1, flow2})

	// Verify initial state
	assert.Equal(t, 0, fm.GetCurrentFlowIndex())
	assert.False(t, fm.IsDone())
	assert.Equal(t, flow1, fm.GetCurrentFlow())
	assert.Equal(t, mockStep1, fm.GetCurrentFlowStep())

	// Proceed through flow1 step1
	fm.ProceedToNextStep()
	assert.Equal(t, 0, fm.GetCurrentFlowIndex())
	assert.Equal(t, mockStep2, fm.GetCurrentFlowStep())

	// Proceed through flow1 step2 (should move to flow2)
	fm.ProceedToNextStep()
	assert.Equal(t, 1, fm.GetCurrentFlowIndex())
	assert.Equal(t, flow2, fm.GetCurrentFlow())
	assert.Equal(t, mockStep3, fm.GetCurrentFlowStep())

	// Complete flow2
	fm.ProceedToNextStep()
	assert.Equal(t, 2, fm.GetCurrentFlowIndex())
	assert.True(t, fm.IsDone())
	assert.Nil(t, fm.GetCurrentFlow())
	assert.Nil(t, fm.GetCurrentFlowStep())
}

func TestFlowManager_ProceedToNextStep_MovesToNextFlow(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installed)
	mockStep2 := createMockStep(t, "Step2", model.NotInstalled)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	fm.SetFlows([]*InstallFlow{flow1, flow2})

	// Initially at flow 0
	assert.Equal(t, 0, fm.GetCurrentFlowIndex())

	// Complete the single step in flow1, should move to flow2
	fm.ProceedToNextStep()
	assert.Equal(t, 1, fm.GetCurrentFlowIndex())
	assert.Equal(t, flow2, fm.GetCurrentFlow())
}

func TestFlowManager_EmptyFlows(t *testing.T) {
	fm := createTestFlowManager()

	assert.NotNil(t, fm)
	assert.True(t, fm.IsDone())
	assert.Nil(t, fm.GetCurrentFlow())
	assert.Nil(t, fm.GetCurrentFlowStep())

	// Should not panic
	fm.ProceedToNextStep()
	assert.Equal(t, 0, fm.GetCurrentFlowIndex())
}

func TestFlowManager_SingleFlow_SingleStep(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := createMockStep(t, "OnlyStep", model.Installed)
	flow := createTestFlow("OnlyFlow", []step.InstallStep{mockStep}, runaiData)

	fm.SetFlows([]*InstallFlow{flow})

	assert.False(t, fm.IsDone())
	assert.Equal(t, flow, fm.GetCurrentFlow())
	assert.Equal(t, mockStep, fm.GetCurrentFlowStep())

	fm.ProceedToNextStep()
	assert.True(t, fm.IsDone())
}

func TestFlowManager_HasFailed_NoFailures(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installed)
	mockStep2 := createMockStep(t, "Step2", model.Installed)
	mockStep3 := createMockStep(t, "Step3", model.NotInstalled)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)
	flow3 := createTestFlow("Flow3", []step.InstallStep{mockStep3}, runaiData)

	fm.SetFlows([]*InstallFlow{flow1, flow2, flow3})

	assert.False(t, fm.HasFailed(), "Should not have failed when no flows are failed or timed out")
}

func TestFlowManager_HasFailed_WithFailedFlow(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installed)
	mockStep2 := createMockStep(t, "Step2", model.Failed)
	mockStep3 := createMockStep(t, "Step3", model.NotInstalled)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)
	flow3 := createTestFlow("Flow3", []step.InstallStep{mockStep3}, runaiData)

	// Advance flow2 to trigger status calculation (Failed step should cause flow to fail)
	flow2.ProceedToNextStep()

	fm.SetFlows([]*InstallFlow{flow1, flow2, flow3})

	assert.True(t, fm.HasFailed(), "Should have failed when at least one flow has Failed status")
}

func TestFlowManager_HasFailed_WithTimeoutFlow(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installed)
	mockStep2 := createMockStep(t, "Step2", model.Timeout)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	// Advance flow2 to trigger status calculation (Timeout step should cause flow to fail)
	flow2.ProceedToNextStep()

	fm.SetFlows([]*InstallFlow{flow1, flow2})

	assert.True(t, fm.HasFailed(), "Should have failed when at least one flow has Timeout status")
}

func TestFlowManager_HasFailed_EmptyFlows(t *testing.T) {
	fm := createTestFlowManager()
	fm.SetFlows([]*InstallFlow{})

	assert.False(t, fm.HasFailed(), "Should not have failed when there are no flows")
}

func TestFlowManager_HasFailed_MultipleFailures(t *testing.T) {
	fm := createTestFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Failed)
	mockStep2 := createMockStep(t, "Step2", model.Timeout)
	mockStep3 := createMockStep(t, "Step3", model.Installed)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)
	flow3 := createTestFlow("Flow3", []step.InstallStep{mockStep3}, runaiData)

	// Advance flows to trigger status calculation
	flow1.ProceedToNextStep()
	flow2.ProceedToNextStep()

	fm.SetFlows([]*InstallFlow{flow1, flow2, flow3})

	assert.True(t, fm.HasFailed(), "Should have failed when multiple flows have Failed or Timeout status")
}

func TestFlowManager_VerifyOnly_SkipsPostInstall(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := mocks.NewMockInstallStep(t)
	mockStep.On("GetName").Return("TestStep").Maybe()
	mockStep.On("GetCachedStatus").Return(model.NotInstalled).Once()
	mockStep.On("GetStatusSync", mock.Anything).Return(model.Installed).Once()
	mockStep.On("GetStatusAsync", mock.Anything).Return(model.Installed).Maybe()
	mockStep.On("GetElapsedTime").Return(time.Duration(0)).Maybe()
	mockStep.On("GetError").Return(nil).Maybe()
	mockStep.On("GetTimeout").Return(5 * time.Minute).Maybe()

	flow := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)

	fm := NewFlowManagerWithExternalEvents([]*InstallFlow{flow}, nil, true)

	events := make(chan interface{}, 10)
	fm.ProvisionStep(context.Background(), runaiData, events)

	mockStep.AssertNotCalled(t, "PostInstall", mock.Anything)
}

func TestFlowManager_NotVerifyOnly_CallsPostInstall(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := mocks.NewMockInstallStep(t)
	mockStep.On("GetName").Return("TestStep").Maybe()
	mockStep.On("GetCachedStatus").Return(model.NotInstalled).Once()
	mockStep.On("GetStatusSync", mock.Anything).Return(model.Installed).Once()
	mockStep.On("GetStatusAsync", mock.Anything).Return(model.Installed).Maybe()
	mockStep.On("GetElapsedTime").Return(time.Duration(0)).Maybe()
	mockStep.On("GetError").Return(nil).Maybe()
	mockStep.On("GetTimeout").Return(5 * time.Minute).Maybe()
	mockStep.On("PostInstall", mock.Anything).Return().Once()

	flow := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)

	fm := NewFlowManagerWithExternalEvents([]*InstallFlow{flow}, nil, false)

	events := make(chan interface{}, 10)
	fm.ProvisionStep(context.Background(), runaiData, events)

	mockStep.AssertCalled(t, "PostInstall", mock.Anything)
}

func TestFlowManager_ContextCancellation_SetsTimeoutStatus(t *testing.T) {
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := mocks.NewMockInstallStep(t)
	mockStep.On("GetName").Return("TestStep").Maybe()
	mockStep.On("GetCachedStatus").Return(model.Installing).Maybe()
	mockStep.On("GetStatusSync", mock.Anything).Return(model.Installing).Maybe()
	mockStep.On("GetStatusAsync", mock.Anything).Return(model.Installing).Maybe()
	mockStep.On("GetElapsedTime").Return(time.Duration(0)).Maybe()
	mockStep.On("GetError").Return(nil).Maybe()
	mockStep.On("GetTimeout").Return(5 * time.Minute).Maybe()
	mockStep.On("CheckTimeout").Return(nil).Maybe()
	mockStep.On("SetStatus", model.Timeout, mock.Anything).Return(model.Timeout).Once()

	flow := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)

	externalEvents := make(chan *FlowStatusReportingMsg, 10)
	fm := NewFlowManagerWithExternalEvents([]*InstallFlow{flow}, externalEvents, false)

	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan interface{}, 10)

	fm.ExecuteFlow(ctx, runaiData, events)

	cancel()

	fm.Wait()

	mockStep.AssertCalled(t, "SetStatus", model.Timeout, mock.Anything)

	var receivedTimeout bool
	timeout := time.After(200 * time.Millisecond)
drainLoop:
	for {
		select {
		case msg := <-externalEvents:
			if msg.FlowStatus == model.Timeout {
				receivedTimeout = true
				assert.Equal(t, "TestFlow", msg.FlowName)
				break drainLoop
			}
		case <-timeout:
			break drainLoop
		}
	}
	assert.True(t, receivedTimeout, "Expected timeout event to be published to external events")
}
