package common_test

import (
	"fmt"
	"testing"
	"time"

	spinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	installMocks "github.com/run-ai/runai/runai-installer/installer/internal/step/mocks"
	"github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// simpleScreen is a simple implementation of Screen interface for testing.
// Note: We cannot use MockScreen from common/mocks because it has a type compatibility
// issue with the self-referential NextScreen() method. The generated mock returns
// common.Screen which creates a type identity mismatch when used across packages.
// This is a known limitation of mockery with self-referential interfaces.
type simpleScreen struct{}

func (s *simpleScreen) Update(msg tea.Msg) tea.Cmd                            { return nil }
func (s *simpleScreen) Init(runaiData *model.RunAIInstallData) tea.Cmd    { return nil }
func (s *simpleScreen) NextScreen() common.Screen                             { return nil }
func (s *simpleScreen) UpdateSelection(runaiData *model.RunAIInstallData) {}
func (s *simpleScreen) GetHeader() string                                     { return "test" }
func (s *simpleScreen) CanProceed() bool                                      { return true }
func (s *simpleScreen) Render() string                                        { return "" }
func (s *simpleScreen) GetMessages() []string                                 { return nil }
func (s *simpleScreen) OnFocus()                                              {}
func (s *simpleScreen) OnBlur()                                               {}
func (s *simpleScreen) FocusFirst()                                           {}
func (s *simpleScreen) FocusLast()                                            {}
func (s *simpleScreen) GetItemCount() int                                     { return 0 }

func createTestInstallationScreen() *common.InstallationScreen {
	return common.NewInstallationScreen("Test Installation", nil)
}

func TestNewInstallationScreen(t *testing.T) {
	screenHeader := "Test Installation Screen"
	mockNextScreen := &simpleScreen{}

	screen := common.NewInstallationScreen(screenHeader, mockNextScreen)

	assert.NotNil(t, screen)
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

func createMockStep(t *testing.T, name string, status model.InstallPhase) *installMocks.MockInstallStep {
	mockStep := installMocks.NewMockInstallStep(t)
	mockStep.On("GetName").Return(name).Maybe()
	mockStep.On("GetStatusAsync", mock.Anything).Return(status).Maybe()
	mockStep.On("GetCachedStatus").Return(status).Maybe()
	mockStep.On("GetError").Return(nil).Maybe()
	mockStep.On("GetElapsedTime").Return(time.Duration(0)).Maybe()
	mockStep.On("GetTimeout").Return(5 * time.Minute).Maybe()
	return mockStep
}

func createTestFlow(name string, steps []step.InstallStep, runaiData *model.RunAIInstallData) *flow.InstallFlow {
	return flow.NewInstallFlow(name, steps, runaiData)
}

func TestInstallationScreen_Init(t *testing.T) {
	screen := createTestInstallationScreen()
	runaiData := createTestRunAIData(5 * time.Minute)

	cmd := screen.Init(runaiData)

	assert.NotNil(t, cmd)
}

func TestInstallationScreen_SetFlows(t *testing.T) {
	screen := createTestInstallationScreen()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep := createMockStep(t, "TestStep", model.NotInstalled)
	fl := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)

	flows := []*flow.InstallFlow{fl}
	screen.SetFlows(flows)

	assert.NotNil(t, screen)
}

func TestInstallationScreen_GetHeader(t *testing.T) {
	screenHeader := "My Custom Header"
	screen := common.NewInstallationScreen(screenHeader, nil)

	result := screen.GetHeader()

	assert.Equal(t, screenHeader, result)
}

func TestInstallationScreen_NextScreen(t *testing.T) {
	mockNextScreen := &simpleScreen{}
	screen := common.NewInstallationScreen("Test", mockNextScreen)

	result := screen.NextScreen()

	assert.Equal(t, mockNextScreen, result)
}

func TestInstallationScreen_CanProceed(t *testing.T) {
	screen := createTestInstallationScreen()

	result := screen.CanProceed()

	assert.True(t, result, "CanProceed should always return true")
}

func TestInstallationScreen_UpdateSelection(t *testing.T) {
	screen := createTestInstallationScreen()
	runaiData := createTestRunAIData(5 * time.Minute)

	// This method is a no-op, but should not panic
	screen.UpdateSelection(runaiData)
	// No assertions needed, just verifying it doesn't panic
}

func TestInstallationScreen_GetMessages(t *testing.T) {
	screen := createTestInstallationScreen()
	runaiData := createTestRunAIData(5 * time.Minute)

	t.Run("returns empty messages when no errors", func(t *testing.T) {
		mockStep := createMockStep(t, "TestStep", model.Installed)
		fl := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)
		screen.SetFlows([]*flow.InstallFlow{fl})

		messages := screen.GetMessages()
		assert.Empty(t, messages)
	})

	t.Run("returns error messages when flow has errors", func(t *testing.T) {
		mockStep := installMocks.NewMockInstallStep(t)
		mockStep.On("GetName").Return("FailingStep").Maybe()
		mockStep.On("GetStatusAsync", mock.Anything).Return(model.Failed).Maybe()
		mockStep.On("GetCachedStatus").Return(model.Failed).Maybe()
		mockStep.On("GetError").Return(fmt.Errorf("test error")).Maybe()

		fl := createTestFlow("FailingFlow", []step.InstallStep{mockStep}, runaiData)
		fl.ProceedToNextStep() // Trigger the flow so it's not in NotInstalled state
		screen.SetFlows([]*flow.InstallFlow{fl})

		messages := screen.GetMessages()
		assert.NotEmpty(t, messages)
		assert.Contains(t, messages[0], "Error in FailingFlow")
	})
}

func TestInstallationScreen_Render(t *testing.T) {
	screen := createTestInstallationScreen()
	runaiData := createTestRunAIData(5 * time.Minute)
	screen.Init(runaiData)

	t.Run("renders empty screen when no flows", func(t *testing.T) {
		screen.SetFlows([]*flow.InstallFlow{})

		output := screen.Render()

		assert.NotEmpty(t, output)
		assert.Contains(t, output, "All components ready")
	})

	t.Run("renders with flows", func(t *testing.T) {
		mockStep := createMockStep(t, "TestStep", model.Installing)
		fl := createTestFlow("TestFlow", []step.InstallStep{mockStep}, runaiData)
		screen.SetFlows([]*flow.InstallFlow{fl})

		output := screen.Render()

		assert.NotEmpty(t, output)
		assert.Contains(t, output, "Installing")
	})
}

func TestInstallationScreen_Update_SpinnerTickMsg(t *testing.T) {
	screen := createTestInstallationScreen()
	runaiData := createTestRunAIData(5 * time.Minute)
	screen.Init(runaiData)

	// Create a spinner tick message
	tickMsg := spinner.TickMsg{Time: time.Now()}

	cmd := screen.Update(tickMsg)

	// Should return a command for updating the spinner
	assert.NotNil(t, cmd)
}

// Note: Tests for private message types (installationStartMsg, nextStepMsg, checkCurrentStepStatusMsg) and
// private methods (processCurrentInstallFlowStep, executeCurrentStepWithTimeout, tick, etc.)
// are not included as they test internal implementation details. The functionality they provide
// is tested indirectly through the public API tests above and through integration tests.

func TestInstallationScreen_Integration_CompleteFlow(t *testing.T) {
	// This is an integration test that simulates a complete installation flow
	screen := createTestInstallationScreen()
	runaiData := createTestRunAIData(5 * time.Minute)

	// Create a flow with multiple steps (all already installed to avoid async execution)
	mockStep1 := installMocks.NewMockInstallStep(t)
	mockStep1.On("GetName").Return("Step1").Maybe()
	mockStep1.On("GetStatusAsync", mock.Anything).Return(model.Installed).Maybe()
	mockStep1.On("GetCachedStatus").Return(model.Installed).Maybe()
	mockStep1.On("PostInstall", runaiData).Return().Maybe()
	mockStep1.On("GetError").Return(nil).Maybe()

	mockStep2 := installMocks.NewMockInstallStep(t)
	mockStep2.On("GetName").Return("Step2").Maybe()
	mockStep2.On("GetStatusAsync", mock.Anything).Return(model.Installed).Maybe()
	mockStep2.On("GetCachedStatus").Return(model.Installed).Maybe()
	mockStep2.On("PostInstall", runaiData).Return().Maybe()
	mockStep2.On("GetError").Return(nil).Maybe()

	fl := createTestFlow("CompleteFlow", []step.InstallStep{mockStep1, mockStep2}, runaiData)
	screen.SetFlows([]*flow.InstallFlow{fl})
	screen.Init(runaiData)

	// Render should show the installation screen
	output := screen.Render()
	assert.NotEmpty(t, output)

	// Verify that both mock steps can be called
	mockStep1.AssertExpectations(t)
	mockStep2.AssertExpectations(t)
}

func TestInstallationScreen_Integration_FlowWithErrors(t *testing.T) {
	screen := createTestInstallationScreen()
	runaiData := createTestRunAIData(5 * time.Minute)

	// Create a flow with a failing step
	mockStep := installMocks.NewMockInstallStep(t)
	mockStep.On("GetName").Return("FailingStep").Maybe()
	mockStep.On("GetStatusAsync", mock.Anything).Return(model.Failed).Maybe()
	mockStep.On("GetCachedStatus").Return(model.Failed).Maybe()
	mockStep.On("GetError").Return(fmt.Errorf("installation failed")).Maybe()

	fl := createTestFlow("FailingFlow", []step.InstallStep{mockStep}, runaiData)
	// Mark as having completed (failed) by proceeding to next step
	fl.ProceedToNextStep()
	screen.SetFlows([]*flow.InstallFlow{fl})
	screen.Init(runaiData)

	// Get messages
	messages := screen.GetMessages()
	assert.NotEmpty(t, messages)
	assert.Contains(t, messages[0], "Error in FailingFlow")
	assert.Contains(t, messages[0], "installation failed")

	// Since the flow is completed but failed, render should show error state or completed state
	output := screen.Render()
	assert.NotEmpty(t, output)
	// The flow is marked as complete, so it should show either "all components" or the flow name
	assert.Contains(t, output, "FailingFlow")
}

func TestInstallationScreen_MultipleFlows(t *testing.T) {
	screen := createTestInstallationScreen()
	runaiData := createTestRunAIData(5 * time.Minute)

	mockStep1 := createMockStep(t, "Flow1Step", model.Installed)
	mockStep2 := createMockStep(t, "Flow2Step", model.NotInstalled)

	flow1 := createTestFlow("FirstFlow", []step.InstallStep{mockStep1}, runaiData)
	flow2 := createTestFlow("SecondFlow", []step.InstallStep{mockStep2}, runaiData)

	screen.SetFlows([]*flow.InstallFlow{flow1, flow2})
	screen.Init(runaiData)

	// Verify render shows the installation screen with both flows
	output := screen.Render()
	assert.NotEmpty(t, output)
}
