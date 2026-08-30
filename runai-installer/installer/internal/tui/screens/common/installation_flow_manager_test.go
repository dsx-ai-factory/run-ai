package common

import (
	"fmt"
	"testing"
	"time"

	spinner "github.com/charmbracelet/bubbles/spinner"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/step/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTestUIFlowManager() *uiFlowManager {
	return NewUIFlowManager()
}

func createTestFlow(name string, steps []step.InstallStep, runaiData *model.RunAIInstallData) *flow.InstallFlow {
	return flow.NewInstallFlow(name, steps, runaiData)
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

func TestUIFlowManager_GetMessages_NoErrors(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installed)
	mockStep2 := createMockStep(t, "Step2", model.Installed)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	fm.SetFlows([]*flow.InstallFlow{flow1, flow2})

	messages := fm.getMessages()
	assert.Empty(t, messages)
}

func TestUIFlowManager_GetMessages_WithErrors(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	// Create step that will fail
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep1.On("GetName").Return("FailingStep").Maybe()
	mockStep1.On("GetStatusAsync", mock.Anything).Return(model.Failed).Maybe()
	mockStep1.On("GetCachedStatus").Return(model.Failed).Maybe()
	mockStep1.On("GetError").Return(fmt.Errorf("test error 1")).Maybe()

	mockStep2 := mocks.NewMockInstallStep(t)
	mockStep2.On("GetName").Return("Step2").Maybe()
	mockStep2.On("GetStatusAsync", mock.Anything).Return(model.Installed).Maybe()
	mockStep2.On("GetCachedStatus").Return(model.Installed).Maybe()
	mockStep2.On("GetError").Return(nil).Maybe()

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	// Trigger the flow to proceed so it's not in NotInstalled state
	flow1.ProceedToNextStep()

	fm.SetFlows([]*flow.InstallFlow{flow1, flow2})

	messages := fm.getMessages()
	assert.NotEmpty(t, messages)
	assert.Contains(t, messages[0], "Error in Flow1")
	assert.Contains(t, messages[0], "test error 1")
}

func TestUIFlowManager_GetMessages_CurrentFlowFailed(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := mocks.NewMockInstallStep(t)
	mockStep.On("GetName").Return("FailingStep").Maybe()
	mockStep.On("GetStatusAsync", mock.Anything).Return(model.Failed).Maybe()
	mockStep.On("GetCachedStatus").Return(model.Failed).Maybe()
	mockStep.On("GetError").Return(fmt.Errorf("current step error")).Maybe()

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep}, runaiData)
	fm.SetFlows([]*flow.InstallFlow{flow1})

	// Trigger the flow to execute and fail (set status through the flow)
	flow1.ProceedToNextStep()

	messages := fm.getMessages()
	// The message should appear since the step is in failed state
	assert.NotEmpty(t, messages)
}

func TestUIFlowManager_GetMessages_SkipsNotInstalledFlows(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep1.On("GetName").Return("Step1").Maybe()
	mockStep1.On("GetStatusAsync", mock.Anything).Return(model.Failed).Maybe()
	mockStep1.On("GetError").Return(fmt.Errorf("error 1")).Maybe()

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	// Keep flow in NotInstalled status (don't proceed)

	fm.SetFlows([]*flow.InstallFlow{flow1})

	messages := fm.getMessages()
	// Since the flow is in NotInstalled status, it should be skipped
	assert.Empty(t, messages)
}

func TestUIFlowManager_GetMessages_MultipleFlowsWithErrors(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	// Flow 1 with error
	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep1.On("GetName").Return("FailingStep1").Maybe()
	mockStep1.On("GetStatusAsync", mock.Anything).Return(model.Failed).Maybe()
	mockStep1.On("GetCachedStatus").Return(model.Failed).Maybe()
	mockStep1.On("GetError").Return(fmt.Errorf("error in flow 1")).Maybe()

	// Flow 2 with error
	mockStep2 := mocks.NewMockInstallStep(t)
	mockStep2.On("GetName").Return("FailingStep2").Maybe()
	mockStep2.On("GetStatusAsync", mock.Anything).Return(model.Failed).Maybe()
	mockStep2.On("GetCachedStatus").Return(model.Failed).Maybe()
	mockStep2.On("GetError").Return(fmt.Errorf("error in flow 2")).Maybe()

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	// Trigger both flows to execute (so they're not in NotInstalled state)
	flow1.ProceedToNextStep()
	flow2.ProceedToNextStep()

	fm.SetFlows([]*flow.InstallFlow{flow1, flow2})

	messages := fm.getMessages()

	assert.Len(t, messages, 2)
	assert.Contains(t, messages[0], "Error in Flow1")
	assert.Contains(t, messages[0], "error in flow 1")
	assert.Contains(t, messages[1], "Error in Flow2")
	assert.Contains(t, messages[1], "error in flow 2")
}

func TestUIFlowManager_Render_EmptyFlows(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	fm.SetFlows([]*flow.InstallFlow{})

	s := spinner.New()
	content := fm.Render(s, runaiData, 5*time.Minute)

	assert.Contains(t, content, "Status: All components ready")
}

func TestUIFlowManager_Render_WithFlows(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installing)
	mockStep2 := createMockStep(t, "Step2", model.NotInstalled)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	fm.SetFlows([]*flow.InstallFlow{flow1, flow2})

	s := spinner.New()
	content := fm.Render(s, runaiData, 5*time.Minute)

	assert.Contains(t, content, "Status: Installing")
	assert.Contains(t, content, "timeout: 5m0s")
}

func TestUIFlowManager_Render_AllComponentsReady(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installed)
	mockStep2 := createMockStep(t, "Step2", model.Installed)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	fm.SetFlows([]*flow.InstallFlow{flow1, flow2})

	// Mark flows as complete
	flow1.ProceedToNextStep()
	flow2.ProceedToNextStep()
	// Move beyond all flows
	fm.flowManager.ProceedToNextStep()
	fm.flowManager.ProceedToNextStep()

	s := spinner.New()
	content := fm.Render(s, runaiData, 5*time.Minute)

	assert.Contains(t, content, "Status: All components ready")
}

func TestUIFlowManager_Render_SomeComponentsFailed(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := mocks.NewMockInstallStep(t)
	mockStep1.On("GetName").Return("FailingStep").Maybe()
	mockStep1.On("GetStatusAsync", mock.Anything).Return(model.Failed).Maybe()
	mockStep1.On("GetCachedStatus").Return(model.Failed).Maybe()
	mockStep1.On("GetError").Return(fmt.Errorf("test error")).Maybe()

	mockStep2 := createMockStep(t, "Step2", model.Installed)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	fm.SetFlows([]*flow.InstallFlow{flow1, flow2})

	// Mark first flow as failed
	flow1.ProceedToNextStep()
	flow2.ProceedToNextStep()
	// Move beyond all flows
	fm.flowManager.ProceedToNextStep()
	fm.flowManager.ProceedToNextStep()

	s := spinner.New()
	content := fm.Render(s, runaiData, 5*time.Minute)

	assert.Contains(t, content, "Status: Some components failed")
}

func TestUIFlowManager_GetInstallationStatusDesc_NotInstalled(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := createMockStep(t, "TestStep", model.NotInstalled)
	flow := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)

	s := spinner.New()
	statusDesc := fm.getInstallationStatusDesc(flow, runaiData, s)

	assert.Contains(t, statusDesc, "[ ]")
	assert.Contains(t, statusDesc, "TestStep")
}

func TestUIFlowManager_GetInstallationStatusDesc_Installing(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := createMockStep(t, "TestStep", model.Installing)
	flow := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)

	s := spinner.New()
	statusDesc := fm.getInstallationStatusDesc(flow, runaiData, s)

	assert.Contains(t, statusDesc, "TestStep")
	// The spinner will be rendering, so we just check the step name is there
}

func TestUIFlowManager_GetInstallationStatusDesc_Installed(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := createMockStep(t, "TestStep", model.Installed)
	flow := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)
	flow.ProceedToNextStep() // Mark as terminal

	s := spinner.New()
	statusDesc := fm.getInstallationStatusDesc(flow, runaiData, s)

	assert.Contains(t, statusDesc, "[✓]")
	assert.Contains(t, statusDesc, "TestFlow")
}

func TestUIFlowManager_GetInstallationStatusDesc_Failed(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := mocks.NewMockInstallStep(t)
	mockStep.On("GetName").Return("FailingStep").Maybe()
	mockStep.On("GetStatusAsync", mock.Anything).Return(model.Failed).Maybe()
	mockStep.On("GetCachedStatus").Return(model.Failed).Maybe()
	mockStep.On("GetError").Return(fmt.Errorf("test error")).Maybe()

	flow := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)
	flow.ProceedToNextStep() // Mark as terminal/failed

	s := spinner.New()
	statusDesc := fm.getInstallationStatusDesc(flow, runaiData, s)

	assert.Contains(t, statusDesc, "[✗]")
	assert.Contains(t, statusDesc, "TestFlow")
}

func TestUIFlowManager_GetInstallationStatusDesc_Timeout(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := mocks.NewMockInstallStep(t)
	mockStep.On("GetName").Return("TimeoutStep").Maybe()
	mockStep.On("GetStatusAsync", mock.Anything).Return(model.Timeout).Maybe()
	mockStep.On("GetCachedStatus").Return(model.Timeout).Maybe()
	mockStep.On("GetError").Return(fmt.Errorf("timeout error")).Maybe()

	flow := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)

	s := spinner.New()
	statusDesc := fm.getInstallationStatusDesc(flow, runaiData, s)

	assert.Contains(t, statusDesc, "[✗]")
	assert.Contains(t, statusDesc, "TimeoutStep")
}

func TestUIFlowManager_Render_CurrentFlowNotInTerminalStatus(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Step1", model.Installing)
	mockStep2 := createMockStep(t, "Step2", model.NotInstalled)

	flow1 := createTestFlow("Flow1", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("Flow2", []step.InstallStep{mockStep2}, runaiData)

	fm.SetFlows([]*flow.InstallFlow{flow1, flow2})

	s := spinner.New()
	content := fm.Render(s, runaiData, 5*time.Minute)

	// Flow1 should be rendered with spinner (current flow)
	assert.Contains(t, content, "Step1")
	// Flow2 should be rendered with [ ] (not started)
	assert.Contains(t, content, "[ ]")
	assert.Contains(t, content, "Flow2")
}

func TestUIFlowManager_Render_MixedFlowStatuses(t *testing.T) {
	fm := createTestUIFlowManager()
	runaiData := createTestRunAIData(5 * time.Minute)

	// Flow 1: Completed
	mockStep1 := createMockStep(t, "Step1", model.Installed)

	// Flow 2: In progress
	mockStep2 := createMockStep(t, "Step2", model.Installing)

	// Flow 3: Not started
	mockStep3 := createMockStep(t, "Step3", model.NotInstalled)

	flow1 := createTestFlow("CompletedFlow", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("InProgressFlow", []step.InstallStep{mockStep2}, runaiData)
	flow3 := createTestFlow("PendingFlow", []step.InstallStep{mockStep3}, runaiData)

	// Complete flow1
	flow1.ProceedToNextStep()

	fm.SetFlows([]*flow.InstallFlow{flow1, flow2, flow3})
	// Move to flow2
	fm.flowManager.ProceedToNextStep()

	s := spinner.New()
	content := fm.Render(s, runaiData, 5*time.Minute)

	// Verify all flows appear in the render
	assert.Contains(t, content, "CompletedFlow")
	assert.Contains(t, content, "Step2") // Current step should show
	assert.Contains(t, content, "PendingFlow")
	assert.Contains(t, content, "Status: Installing")
}
