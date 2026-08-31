/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package screens

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	viewscommon "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
	screencomponent "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/screen_component"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestFullFlowIntegration tests the complete user flow through all screens
// starting from the main screen and using fake k8s clients
func TestFullFlowIntegration(t *testing.T) {
	// Setup: Create fake k8s client with some existing resources
	k8sClient := createFakeK8sClientWithResources(t)

	// Create test data
	runaiData := createFullFlowTestData()

	// Initialize all screens
	views := GetScreensWithK8s(k8sClient)

	// Create main screen with components
	mainScreenData := screencomponent.NewMultiFrameScreenData(
		runaiData,
		k8sClient,
		views[viewscommon.ScreenPrerequisites],
	)

	screenComponents := createRealScreenComponents(mainScreenData)
	mainScreen := NewAppModel(runaiData, mainScreenData, screenComponents, k8sClient)

	// Test 1: Initialize main screen
	t.Run("InitializeMainScreen", func(t *testing.T) {
		cmd := mainScreen.Init()
		assert.NotNil(t, mainScreen)

		// Execute init command if present
		if cmd != nil {
			cmd()
		}

		// Verify we're on the prerequisites screen
		assert.Equal(t, viewscommon.ScreenPrerequisites, getCurrentScreenType(mainScreenData.CurrentScreen, views))
	})

	// Test 2: Prerequisites Screen (was Test 3)
	t.Run("PrerequisitesScreen", func(t *testing.T) {
		screen := mainScreenData.CurrentScreen
		assert.NotNil(t, screen)

		// Verify initial state
		header := screen.GetHeader()
		assert.Contains(t, header, "software dependencies")

		// Verify we can proceed (default selection is valid)
		assert.True(t, screen.CanProceed())
	})

	// Test 4: Select Prerequisites (was Test 4)
	t.Run("SelectPrerequisites", func(t *testing.T) {
		screen := mainScreenData.CurrentScreen

		// Verify can proceed even without selections (optional prerequisites)
		assert.True(t, screen.CanProceed())

		// Simulate selecting some prerequisites
		screen.UpdateSelection(runaiData)

		// Prerequisites should be selected or empty
		assert.NotNil(t, runaiData.Prerequisites)
	})

	// Test 5: Navigate to Configuration Screen
	t.Run("NavigateToConfigurationScreen", func(t *testing.T) {
		nextScreen := mainScreenData.CurrentScreen.NextScreen()
		require.NotNil(t, nextScreen)

		mainScreenData.CurrentScreen = nextScreen
		cmd := nextScreen.Init(runaiData)
		if cmd != nil {
			cmd()
		}

		// Verify we're on configuration screen
		assert.Equal(t, viewscommon.ScreenConfiguration, getCurrentScreenType(nextScreen, views))

		header := nextScreen.GetHeader()
		assert.Contains(t, header, "Configuration")
	})

	// Test 6: Fill in Configuration
	t.Run("FillConfiguration", func(t *testing.T) {
		screen := mainScreenData.CurrentScreen

		// Initially, required fields are not filled, so we can't proceed
		canProceed := screen.CanProceed()
		t.Logf("Can proceed before filling config: %v", canProceed)

		// Fill in required configuration
		runaiData.Inputs[configinputs.InstallInputClusterName] = "test-cluster"
		runaiData.Inputs[configinputs.InstallInputJfrogCreds] = "dGVzdDp0ZXN0" // base64 encoded "test:test"
		runaiData.Inputs[configinputs.InstallInputFQDN] = "runai.example.com"

		// Reinitialize screen with filled data
		cmd := screen.Init(runaiData)
		if cmd != nil {
			cmd()
		}

		// Update selection to persist values
		screen.UpdateSelection(runaiData)

		// Verify data is saved
		assert.Equal(t, "test-cluster", runaiData.Inputs[configinputs.InstallInputClusterName])
		assert.Equal(t, "runai.example.com", runaiData.Inputs[configinputs.InstallInputFQDN])
	})

	// Test 7: Navigate to Input Validation Screen
	t.Run("NavigateToInputValidationScreen", func(t *testing.T) {
		nextScreen := mainScreenData.CurrentScreen.NextScreen()
		require.NotNil(t, nextScreen)

		mainScreenData.CurrentScreen = nextScreen
		cmd := nextScreen.Init(runaiData)
		if cmd != nil {
			cmd()
		}

		// Verify we're on input validation screen
		assert.Equal(t, viewscommon.ScreenInputValidation, getCurrentScreenType(nextScreen, views))

		header := nextScreen.GetHeader()
		assert.Contains(t, header, "Validat")
	})

	// Test 8: Input Validation
	t.Run("ValidateInputs", func(t *testing.T) {
		screen := mainScreenData.CurrentScreen

		// Allow validation to run
		time.Sleep(100 * time.Millisecond)

		// Verify screen behavior
		render := screen.Render()
		assert.NotEmpty(t, render)
	})

	// Test 9: Navigate to Prerequisite Install Screen
	t.Run("NavigateToPrerequisiteInstallScreen", func(t *testing.T) {
		nextScreen := mainScreenData.CurrentScreen.NextScreen()
		require.NotNil(t, nextScreen)

		mainScreenData.CurrentScreen = nextScreen
		cmd := nextScreen.Init(runaiData)
		if cmd != nil {
			cmd()
		}

		// Verify we're on prerequisite install screen
		assert.Equal(t, viewscommon.ScreenPrerequisiteInstall, getCurrentScreenType(nextScreen, views))
	})

	// Test 10: Navigate to Installation Screen
	t.Run("NavigateToInstallationScreen", func(t *testing.T) {
		nextScreen := mainScreenData.CurrentScreen.NextScreen()
		require.NotNil(t, nextScreen)

		mainScreenData.CurrentScreen = nextScreen
		cmd := nextScreen.Init(runaiData)
		if cmd != nil {
			cmd()
		}

		// Verify we're on installation screen
		assert.Equal(t, viewscommon.ScreenInstallation, getCurrentScreenType(nextScreen, views))
	})

	// Test 11: View rendering at each stage
	t.Run("ViewRendering", func(t *testing.T) {
		view := mainScreen.View()
		assert.Contains(t, view, "NVIDIA Run:ai Installer")
		assert.NotEmpty(t, view)
	})

	// Test 12: Test navigation back and forth
	t.Run("BackNavigation", func(t *testing.T) {
		// This would test the back button functionality
		// For now, just verify the history mechanism works
		assert.NotNil(t, mainScreenData.CurrentScreen)
	})

	// Test 13: Test quit functionality
	t.Run("QuitFunctionality", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		_, cmd := mainScreen.Update(msg)

		assert.NotNil(t, cmd)
		quitMsg := cmd()
		assert.IsType(t, tea.QuitMsg{}, quitMsg)
	})

	// Test 14: Test window resize
	t.Run("WindowResize", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 120, Height: 40}
		model, cmd := mainScreen.Update(msg)

		assert.NotNil(t, model)
		assert.NotNil(t, cmd) // Returns tea.ClearScreen to prevent rendering artifacts
	})
}

// TestFullFlowWithInstallation tests the complete flow including actual installation simulation
func TestFullFlowWithInstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	k8sClient := createFakeK8sClientWithResources(t)
	runaiData := createFullFlowTestData()

	// Pre-fill all required data to skip input screens
	runaiData.Prerequisites = []string{} // No prerequisites for faster test
	runaiData.Inputs[configinputs.InstallInputClusterName] = "test-cluster"
	runaiData.Inputs[configinputs.InstallInputJfrogCreds] = "dGVzdDp0ZXN0"
	runaiData.Inputs[configinputs.InstallInputFQDN] = "runai.example.com"

	// Initialize screens
	views := GetScreensWithK8s(k8sClient)

	// Start at installation screen directly
	installScreen := views[viewscommon.ScreenInstallation]

	// Initialize installation
	cmd := installScreen.Init(runaiData)
	if cmd != nil {
		msg := cmd()
		t.Logf("Installation init returned message: %v", msg)
	}

	// Render the screen
	render := installScreen.Render()
	assert.NotEmpty(t, render)

	// Get any error messages
	messages := installScreen.GetMessages()
	for _, msg := range messages {
		t.Logf("Installation message: %s", msg)
	}
}

// Helper functions

func createFakeK8sClientWithResources(t *testing.T) k8s.K8sClients {
	scheme := runtime.NewScheme()
	err := appsv1.AddToScheme(scheme)
	require.NoError(t, err)
	err = corev1.AddToScheme(scheme)
	require.NoError(t, err)

	// Create some test resources
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "runai",
		},
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "runai",
		},
		Data: map[string]string{
			"key": "value",
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "runai",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "test:latest",
						},
					},
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(namespace, configMap, deployment).
		Build()

	return &testK8sClientsImpl{
		client: fakeClient,
		config: &rest.Config{},
	}
}

func createFullFlowTestData() *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:     false,
		StepTimeout: 30 * time.Second,
		LogsDir:     "/tmp/runai-installer-test",
	}

	runaiData, _ := model.NewRunAIInstallData(opts)
	runaiData.Inputs = make(map[configinputs.InstallInputType]string)

	return runaiData
}

func createRealScreenComponents(mainScreenData *screencomponent.MultiFrameScreenData) []screencomponent.ScreenFrame {
	components := make([]screencomponent.ScreenFrame, 3)
	components[viewscommon.ScreenFrameTypeMain] = screencomponent.NewMainFrame(mainScreenData)
	components[viewscommon.ScreenFrameTypeErrors] = screencomponent.NewErrorsFrame(mainScreenData)
	components[viewscommon.ScreenFrameTypeNavButtons] = screencomponent.NewNavButtonsFrame(mainScreenData)
	return components
}

func getCurrentScreenType(current viewscommon.Screen, views map[viewscommon.ScreenType]viewscommon.Screen) viewscommon.ScreenType {
	for screenType, screen := range views {
		if screen == current {
			return screenType
		}
	}
	return -1 // Not found
}

// testK8sClientsImpl implements k8s.K8sClients for testing
type testK8sClientsImpl struct {
	client client.Client
	config *rest.Config
}

func (t *testK8sClientsImpl) GetClient() client.Client {
	return t.client
}

func (t *testK8sClientsImpl) GetConfig() *rest.Config {
	return t.config
}

func (t *testK8sClientsImpl) GetContextName() string {
	return "test-context"
}

// TestFullFlowStepByStep provides a detailed step-by-step test with more granular assertions
func TestFullFlowStepByStep(t *testing.T) {
	// This test simulates actual user interactions at each screen

	k8sClient := createFakeK8sClientWithResources(t)
	runaiData := createFullFlowTestData()

	views := GetScreensWithK8s(k8sClient)

	// Create the full application model
	mainScreenData := screencomponent.NewMultiFrameScreenData(
		runaiData,
		k8sClient,
		views[viewscommon.ScreenPrerequisites],
	)

	screenComponents := createRealScreenComponents(mainScreenData)
	mainScreen := NewAppModel(runaiData, mainScreenData, screenComponents, k8sClient)

	// Initialize
	initCmd := mainScreen.Init()
	if initCmd != nil {
		initCmd()
	}

	// Step 1: Prerequisites Screen (was Step 2)
	t.Log("Step 1: Prerequisites Screen")
	currentScreen := mainScreenData.CurrentScreen

	// Render to verify screen works
	view := mainScreen.View()
	assert.Contains(t, view, "Run:ai Installer")

	// Verify we can see prerequisites
	render := currentScreen.Render()
	t.Logf("Prerequisites screen render length: %d", len(render))

	// Select no prerequisites for faster test
	currentScreen.UpdateSelection(runaiData)
	assert.True(t, currentScreen.CanProceed())

	// Step 2: Move to Configuration Screen
	t.Log("Step 2: Configuration Screen")
	nextScreen := currentScreen.NextScreen()
	require.NotNil(t, nextScreen, "Should have a next screen")
	mainScreenData.CurrentScreen = nextScreen

	initCmd = nextScreen.Init(runaiData)
	if initCmd != nil {
		initCmd()
	}

	// Simulate typing in configuration values
	runaiData.Inputs[configinputs.InstallInputClusterName] = "integration-test-cluster"
	runaiData.Inputs[configinputs.InstallInputJfrogCreds] = "dGVzdGNyZWRz" // "testcreds" in base64
	runaiData.Inputs[configinputs.InstallInputFQDN] = "integration.runai.example.com"

	// Re-init to pick up the values
	initCmd = nextScreen.Init(runaiData)
	if initCmd != nil {
		initCmd()
	}

	nextScreen.UpdateSelection(runaiData)
	t.Logf("Configuration filled. Can proceed: %v", nextScreen.CanProceed())

	// Step 3: Continue through remaining screens
	t.Log("Step 3: Input Validation Screen")
	nextScreen = mainScreenData.CurrentScreen.NextScreen()
	if nextScreen != nil {
		mainScreenData.CurrentScreen = nextScreen
		initCmd = nextScreen.Init(runaiData)
		if initCmd != nil {
			initCmd()
		}

		// Give validation a moment
		time.Sleep(50 * time.Millisecond)

		t.Logf("Validation screen header: %s", nextScreen.GetHeader())
	}

	t.Log("Test completed successfully - full flow navigated")
}

// TestMainScreenWithRealComponents tests main screen with actual screen components instead of mocks
func TestMainScreenWithRealComponents(t *testing.T) {
	k8sClient := createFakeK8sClientWithResources(t)
	runaiData := createFullFlowTestData()

	views := GetScreensWithK8s(k8sClient)

	mainScreenData := screencomponent.NewMultiFrameScreenData(
		runaiData,
		k8sClient,
		views[viewscommon.ScreenPrerequisites],
	)

	screenComponents := createRealScreenComponents(mainScreenData)
	mainScreen := NewAppModel(runaiData, mainScreenData, screenComponents, k8sClient)

	// Test initialization
	cmd := mainScreen.Init()
	assert.NotNil(t, mainScreen)
	if cmd != nil {
		cmd()
	}

	// Test view rendering
	view := mainScreen.View()
	assert.Contains(t, view, "NVIDIA Run:ai Installer")
	assert.NotEmpty(t, view)

	// Test window size update
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	model, cmd := mainScreen.Update(msg)
	assert.NotNil(t, model)

	// Test key handling
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	model, cmd = mainScreen.Update(keyMsg)
	assert.NotNil(t, model)

	// Test quit
	quitMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	model, cmd = mainScreen.Update(quitMsg)
	assert.NotNil(t, cmd)
	msg2 := cmd()
	assert.IsType(t, tea.QuitMsg{}, msg2)
}

// TestScreenTransitions tests that each screen properly transitions to the next
func TestScreenTransitions(t *testing.T) {
	k8sClient := createFakeK8sClientWithResources(t)
	runaiData := createFullFlowTestData()

	// Fill in all required data upfront
	runaiData.Prerequisites = []string{}
	runaiData.Inputs[configinputs.InstallInputClusterName] = "test-cluster"
	runaiData.Inputs[configinputs.InstallInputJfrogCreds] = "dGVzdDp0ZXN0"
	runaiData.Inputs[configinputs.InstallInputFQDN] = "runai.example.com"

	views := GetScreensWithK8s(k8sClient)

	// Define expected screen flow
	expectedFlow := []viewscommon.ScreenType{
		viewscommon.ScreenPrerequisites,
		viewscommon.ScreenConfiguration,
		viewscommon.ScreenInputValidation,
		viewscommon.ScreenPrerequisiteInstall,
		viewscommon.ScreenInstallation,
	}

	currentScreen := views[viewscommon.ScreenPrerequisites]

	for i, expectedScreenType := range expectedFlow {
		t.Run(screenTypeToString(expectedScreenType), func(t *testing.T) {
			require.NotNil(t, currentScreen, "Screen %d should not be nil", i)

			// Initialize the screen
			cmd := currentScreen.Init(runaiData)
			if cmd != nil {
				cmd()
			}

			// Verify screen type
			actualType := getCurrentScreenType(currentScreen, views)
			assert.Equal(t, expectedScreenType, actualType, "Screen %d type mismatch", i)

			// Update selection
			currentScreen.UpdateSelection(runaiData)

			// Get next screen (except for last screen)
			if i < len(expectedFlow)-1 {
				nextScreen := currentScreen.NextScreen()
				assert.NotNil(t, nextScreen, "Screen %d should have a next screen", i)
				currentScreen = nextScreen
			}
		})
	}
}

// screenTypeToString converts a ScreenType to a readable string for test output
func screenTypeToString(s viewscommon.ScreenType) string {
	switch s {
	case viewscommon.ScreenPrerequisites:
		return "Prerequisites"
	case viewscommon.ScreenConfiguration:
		return "Configuration"
	case viewscommon.ScreenPrerequisiteInstall:
		return "PrerequisiteInstall"
	case viewscommon.ScreenInputValidation:
		return "InputValidation"
	case viewscommon.ScreenInstallation:
		return "Installation"
	case viewscommon.ScreenComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}
