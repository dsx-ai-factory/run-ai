/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package screens

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	viewscommon "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/screens/common"
	screencomponent "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/screens/screen_component"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/screens/screen_component/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// testK8sClients implements K8sClients for testing
type testK8sClients struct {
	client client.Client
	config *rest.Config
}

func (t *testK8sClients) GetClient() client.Client {
	return t.client
}

func (t *testK8sClients) GetConfig() *rest.Config {
	return t.config
}

func (t *testK8sClients) GetContextName() string {
	return "test-context"
}

// createFakeK8sClients creates a test K8sClients implementation with fake clients
func createFakeK8sClients(objects ...client.Object) k8s.K8sClients {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	builder := fake.NewClientBuilder().WithScheme(scheme)
	if len(objects) > 0 {
		builder = builder.WithObjects(objects...)
	}

	return &testK8sClients{
		client: builder.Build(),
		config: &rest.Config{},
	}
}

// createTestRunAIData creates a RunAIInstallData instance for testing
func createTestRunAIData() *model.RunAIInstallData {
	opts := &config.InstallConfig{DevMode: false, StepTimeout: 5 * time.Second}
	runaiData, _ := model.NewRunAIInstallData(opts)
	runaiData.Inputs = make(map[inputs.InstallInputType]string)
	return runaiData
}

// createMockScreenComponents creates mock screen components for testing
func createMockScreenComponents(t *testing.T) []screencomponent.ScreenFrame {
	mockInnerView := mocks.NewMockScreenComponent(t)
	mockMessages := mocks.NewMockScreenComponent(t)
	mockNavButtons := mocks.NewMockScreenComponent(t)

	// Set up flexible expectations that can be called multiple times
	setupMockExpectations(mockInnerView, 0, "inner view")
	setupMockExpectations(mockMessages, 1, "messages")
	setupMockExpectations(mockNavButtons, 2, "nav buttons")

	return []screencomponent.ScreenFrame{mockInnerView, mockMessages, mockNavButtons}
}

// setupMockExpectations sets up flexible mock expectations that can be called multiple times
func setupMockExpectations(mockComponent *mocks.MockScreenComponent, updateOrder int, renderText string) {
	mockComponent.On("GetUpdateOrder").Return(updateOrder).Maybe()
	mockComponent.On("Render").Return(renderText).Maybe()
	mockComponent.On("Update", mock.Anything).Return(tea.Cmd(nil)).Maybe()
	mockComponent.On("OnFocus").Maybe()
	mockComponent.On("OnBlur").Maybe()
	mockComponent.On("SetSize", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Maybe()
	mockComponent.On("Event").Return(func() tea.Cmd { return nil }).Maybe()
}

func TestNewMainScreen(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()
	mockComponents := createMockScreenComponents(t)

	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, &mockScreen{})
	mainScreen := NewAppModel(runaiData, mainScreenData, mockComponents, k8sClient)

	assert.NotNil(t, mainScreen)
	assert.NotNil(t, mainScreen.multiFrameScreenData)
	assert.Equal(t, len(mockComponents), len(mainScreen.screenFrames))
	assert.Equal(t, viewscommon.ScreenFrameTypeMain, mainScreen.multiFrameScreenData.FocusedScreenFrame)
}

func TestMainScreen_Init(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()
	mockComponents := createMockScreenComponents(t)

	// Mock the current screen's Init method
	mockScreen := &mockScreen{}
	mockScreen.On("Init", runaiData).Return(tea.Cmd(nil))
	mockScreen.On("GetItemCount").Return(1) // Has selectable items, so main frame will be focused
	mockScreen.On("OnFocus").Maybe()

	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, mockScreen)
	mainScreen := NewAppModel(runaiData, mainScreenData, mockComponents, k8sClient)

	cmd := mainScreen.Init()

	// Init now always returns a command (includes tea.WindowSize() for proper initial layout)
	assert.NotNil(t, cmd)
	mockScreen.AssertExpectations(t)
}

func TestMainScreen_Update_WindowSizeMsg(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()
	mockComponents := createMockScreenComponents(t)

	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, &mockScreen{})
	mainScreen := NewAppModel(runaiData, mainScreenData, mockComponents, k8sClient)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	model, cmd := mainScreen.Update(msg)

	assert.Equal(t, mainScreen, model)
	assert.NotNil(t, cmd) // Returns tea.ClearScreen to prevent rendering artifacts
}

func TestMainScreen_Update_EscapeKey(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()
	mockComponents := createMockScreenComponents(t)

	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, &mockScreen{})
	mainScreen := NewAppModel(runaiData, mainScreenData, mockComponents, k8sClient)

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	model, cmd := mainScreen.Update(msg)

	assert.Equal(t, mainScreen, model)
	assert.NotNil(t, cmd)
	// Test that the command returns a quit message
	quitMsg := cmd()
	assert.IsType(t, tea.QuitMsg{}, quitMsg)
}

func TestMainScreen_Update_CtrlCKey(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()
	mockComponents := createMockScreenComponents(t)

	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, &mockScreen{})
	mainScreen := NewAppModel(runaiData, mainScreenData, mockComponents, k8sClient)

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	model, cmd := mainScreen.Update(msg)

	assert.Equal(t, mainScreen, model)
	assert.NotNil(t, cmd)
	// Test that the command returns a quit message
	quitMsg := cmd()
	assert.IsType(t, tea.QuitMsg{}, quitMsg)
}

func TestMainScreen_Update_OtherKeyMsg(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()
	mockComponents := createMockScreenComponents(t)

	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, &mockScreen{})
	mainScreen := NewAppModel(runaiData, mainScreenData, mockComponents, k8sClient)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	model, cmd := mainScreen.Update(msg)

	assert.Equal(t, mainScreen, model)
	assert.Nil(t, cmd)
}

func TestMainScreen_Update_NonKeyMsg(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()
	mockComponents := createMockScreenComponents(t)

	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, &mockScreen{})
	mainScreen := NewAppModel(runaiData, mainScreenData, mockComponents, k8sClient)

	msg := "some other message"
	model, cmd := mainScreen.Update(msg)

	assert.Equal(t, mainScreen, model)
	assert.Nil(t, cmd)
}

func TestMainScreen_View(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()
	mockComponents := createMockScreenComponents(t)

	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, &mockScreen{})
	mainScreen := NewAppModel(runaiData, mainScreenData, mockComponents, k8sClient)

	view := mainScreen.View()

	assert.Contains(t, view, "NVIDIA Run:ai Installer")
	assert.Contains(t, view, "inner view")
	assert.Contains(t, view, "messages")
	assert.Contains(t, view, "nav buttons")
}

func TestMainScreen_UpdateComponent_WithCommand(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()

	// Create fresh mocks for this specific test
	mockComp1 := mocks.NewMockScreenComponent(t)
	mockComp2 := mocks.NewMockScreenComponent(t)
	mockComp3 := mocks.NewMockScreenComponent(t)

	// Set up the first component to return a command, others to return nil
	testCmd := func() tea.Msg { return "test message" }
	mockComp1.On("GetUpdateOrder").Return(0)
	mockComp1.On("Update", mock.Anything).Return(tea.Cmd(testCmd))

	mockComp2.On("GetUpdateOrder").Return(1)
	mockComp3.On("GetUpdateOrder").Return(2)

	components := []screencomponent.ScreenFrame{mockComp1, mockComp2, mockComp3}
	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, &mockScreen{})
	mainScreen := NewAppModel(runaiData, mainScreenData, components, k8sClient)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd := mainScreen.updateFrames(msg)

	assert.NotNil(t, cmd)
	assert.Equal(t, "test message", cmd())
}

func TestMainScreen_UpdateComponent_ComponentOrdering(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()

	// Create components with different update orders (not in array order)
	mockComp1 := mocks.NewMockScreenComponent(t) // updateOrder: 2 - should be called third
	mockComp2 := mocks.NewMockScreenComponent(t) // updateOrder: 0 - should be called first
	mockComp3 := mocks.NewMockScreenComponent(t) // updateOrder: 1 - should be called second

	mockComp1.On("GetUpdateOrder").Return(2)
	mockComp2.On("GetUpdateOrder").Return(0)
	mockComp3.On("GetUpdateOrder").Return(1)

	// Track call order
	callOrder := 0
	mockComp2.On("Update", mock.Anything).Return(tea.Cmd(nil)).Run(func(args mock.Arguments) {
		assert.Equal(t, 0, callOrder, "Component with update order 0 should be called first")
		callOrder++
	})
	mockComp3.On("Update", mock.Anything).Return(tea.Cmd(nil)).Run(func(args mock.Arguments) {
		assert.Equal(t, 1, callOrder, "Component with update order 1 should be called second")
		callOrder++
	})
	mockComp1.On("Update", mock.Anything).Return(tea.Cmd(nil)).Run(func(args mock.Arguments) {
		assert.Equal(t, 2, callOrder, "Component with update order 2 should be called third")
		callOrder++
	})

	components := []screencomponent.ScreenFrame{mockComp1, mockComp2, mockComp3}
	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, &mockScreen{})
	mainScreen := NewAppModel(runaiData, mainScreenData, components, k8sClient)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd := mainScreen.updateFrames(msg)

	assert.Nil(t, cmd)
	assert.Equal(t, 3, callOrder, "All components should have been called")
}

func TestMainScreen_UpdateComponent_NoCommand(t *testing.T) {
	runaiData := createTestRunAIData()
	k8sClient := createFakeK8sClients()
	mockComponents := createMockScreenComponents(t)

	mainScreenData := screencomponent.NewMultiFrameScreenData(runaiData, k8sClient, &mockScreen{})
	mainScreen := NewAppModel(runaiData, mainScreenData, mockComponents, k8sClient)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd := mainScreen.updateFrames(msg)

	assert.Nil(t, cmd)
}

// mockScreen is a simple mock implementation of the Screen interface for testing
type mockScreen struct {
	mock.Mock
}

func (m *mockScreen) Init(runaiData *model.RunAIInstallData) tea.Cmd {
	args := m.Called(runaiData)
	return args.Get(0).(tea.Cmd)
}

func (m *mockScreen) Update(msg tea.Msg) tea.Cmd {
	args := m.Called(msg)
	return args.Get(0).(tea.Cmd)
}

func (m *mockScreen) Render() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockScreen) NextScreen() viewscommon.Screen {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(viewscommon.Screen)
}

func (m *mockScreen) GetHeader() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockScreen) CanProceed() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockScreen) UpdateSelection(runaiData *model.RunAIInstallData) {
	m.Called(runaiData)
}

func (m *mockScreen) GetMessages() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *mockScreen) OnFocus() {
	m.Called()
}

func (m *mockScreen) OnBlur() {
	m.Called()
}

func (m *mockScreen) FocusFirst() {
	m.Called()
}

func (m *mockScreen) FocusLast() {
	m.Called()
}

func (m *mockScreen) GetItemCount() int {
	args := m.Called()
	return args.Int(0)
}
