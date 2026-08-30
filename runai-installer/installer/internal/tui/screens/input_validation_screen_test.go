package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/inputs/validation"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	prereqids "github.com/run-ai/runai/runai-installer/installer/internal/prereq/ids"
	"github.com/run-ai/runai/runai-installer/installer/internal/tui/inputs"
)

// Helper function to create test RunAI data with minimal setup
func newInputValidationTestData(devMode bool) *model.RunAIInstallData {
	opts := &config.InstallConfig{DevMode: devMode}
	r, _ := model.NewRunAIInstallData(opts)
	r.Prerequisites = []string{prereqids.PrerequisiteIdentifierIngress}
	r.Inputs = make(map[configinputs.InstallInputType]string)

	// Set test values for inputs
	r.Inputs[configinputs.InstallInputFQDN] = "test.example.com"
	r.Inputs[configinputs.InstallInputClusterName] = "test-cluster"
	r.Inputs[configinputs.InstallInputIngressControllerHTTPSPort] = "30443"

	return r
}

func TestInputValidationScreen_NewInputValidationScreen(t *testing.T) {
	screen := NewInputValidationScreen(nil, nil)

	if screen == nil {
		t.Fatal("expected screen to be created")
	}

	if screen.currentValidation != 0 {
		t.Errorf("expected currentValidation to be 0, got %d", screen.currentValidation)
	}
}

func TestInputValidationScreen_Init(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)

	cmd := screen.Init(runaiData)

	if screen.runaiData != runaiData {
		t.Error("expected runaiData to be set")
	}

	if len(screen.validationResults) == 0 {
		t.Error("expected validation results to be initialized")
	}

	if len(screen.validations) == 0 {
		t.Error("expected validations to be initialized")
	}

	if cmd == nil {
		t.Error("expected init command to be returned")
	}

	// Test that initial validation results are set to pending
	for _, result := range screen.validationResults {
		if result.Status {
			t.Error("expected initial validation status to be false (pending)")
		}
		if result.Message != "Waiting..." {
			t.Errorf("expected initial message to be 'Waiting...', got %s", result.Message)
		}
	}
}

func TestInputValidationScreen_Init_EmptyInputs(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	runaiData.Inputs = make(map[configinputs.InstallInputType]string) // Empty inputs
	screen := NewInputValidationScreen(nil, nil)

	cmd := screen.Init(runaiData)

	if len(screen.validationResults) != 0 {
		t.Errorf("expected 0 validation results for empty inputs, got %d", len(screen.validationResults))
	}

	if len(screen.validations) != 0 {
		t.Errorf("expected 0 validations for empty inputs, got %d", len(screen.validations))
	}

	if cmd == nil {
		t.Error("expected init command to be returned even with empty inputs")
	}
}

func TestInputValidationScreen_Update_ValidationStartMsg(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Simulate starting validation at index 0
	cmd := screen.Update(validationStartMsg{index: 0})

	if screen.currentValidation != 0 {
		t.Errorf("expected currentValidation to be 0, got %d", screen.currentValidation)
	}

	if cmd == nil {
		t.Error("expected command to be returned for validation start")
	}
}

func TestInputValidationScreen_Update_ValidationProgressMsg(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Create a mock validation result
	mockResult := validation.ValidationResult{
		Name:    "Test Input",
		Status:  true,
		Message: "Valid",
	}

	// Simulate progress message for first validation
	cmd := screen.Update(validationProgressMsg{index: 0, result: mockResult})

	// Check if result was stored
	if screen.validationResults[0].Status != true {
		t.Error("expected validation result to be updated")
	}

	if screen.validationResults[0].Message != "Valid" {
		t.Errorf("expected message to be 'Valid', got %s", screen.validationResults[0].Message)
	}

	// Should return command for next validation if there are more
	if len(screen.validationResults) > 1 && cmd == nil {
		t.Error("expected command for next validation")
	}
}

func TestInputValidationScreen_Update_LastValidationProgress(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Create a mock validation result
	mockResult := validation.ValidationResult{
		Name:    "Last Input",
		Status:  true,
		Message: "Valid",
	}

	// Simulate progress message for last validation
	lastIndex := len(screen.validationResults) - 1
	cmd := screen.Update(validationProgressMsg{index: lastIndex, result: mockResult})

	// Should not return command when all validations complete
	if cmd != nil {
		t.Error("expected no command when all validations are complete")
	}

	if screen.currentValidation != len(screen.validationResults) {
		t.Errorf("expected currentValidation to be %d, got %d", len(screen.validationResults), screen.currentValidation)
	}
}

func TestInputValidationScreen_CanProceed_DevMode(t *testing.T) {
	runaiData := newInputValidationTestData(true) // Dev mode enabled
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	if !screen.CanProceed() {
		t.Error("expected CanProceed to be true in dev mode")
	}
}

func TestInputValidationScreen_CanProceed_NonDevMode_AllValid(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Set all validation results to successful
	for i := range screen.validationResults {
		screen.validationResults[i].Status = true
	}

	if !screen.CanProceed() {
		t.Error("expected CanProceed to be true when all validations pass")
	}
}

func TestInputValidationScreen_CanProceed_NonDevMode_HasFailures(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Set some validation results to failed
	if len(screen.validationResults) > 0 {
		screen.validationResults[0].Status = false
	}

	if screen.CanProceed() {
		t.Error("expected CanProceed to be false when validations fail")
	}
}

func TestInputValidationScreen_Render_InProgress(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Set current validation to simulate in-progress state
	screen.currentValidation = 1

	output := screen.Render()

	if output == "" {
		t.Fatal("expected non-empty render output")
	}

	if !strings.Contains(output, "Validating user inputs...") {
		t.Error("expected render to show validation in progress message")
	}

	// Should show pending status for current validation
	if !strings.Contains(output, "⏳") {
		t.Error("expected render to show sand clock for current validation")
	}
}

func TestInputValidationScreen_Render_Complete(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Set all validations to complete and successful
	for i := range screen.validationResults {
		screen.validationResults[i].Status = true
	}
	screen.currentValidation = len(screen.validationResults) // All complete

	output := screen.Render()

	if output == "" {
		t.Fatal("expected non-empty render output")
	}

	if !strings.Contains(output, "Validation complete.") {
		t.Error("expected render to show validation complete message")
	}

	// Should show success checkmarks
	if !strings.Contains(output, "✓") {
		t.Error("expected render to show success checkmarks")
	}
}

func TestInputValidationScreen_Render_WithFailures(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Set some validations to failed
	if len(screen.validationResults) > 0 {
		screen.validationResults[0].Status = false
		screen.validationResults[0].Message = "Validation failed"
	}
	screen.currentValidation = len(screen.validationResults) // All complete

	output := screen.Render()

	if output == "" {
		t.Fatal("expected non-empty render output")
	}

	// Should show error marks for failed validations
	if !strings.Contains(output, "✗") {
		t.Error("expected render to show error marks for failed validations")
	}
}

func TestInputValidationScreen_NextScreen(t *testing.T) {
	mockNextScreen := &mockScreen{}
	screen := NewInputValidationScreen(mockNextScreen, nil)

	nextScreen := screen.NextScreen()

	if nextScreen != mockNextScreen {
		t.Error("expected next screen to be the mock screen passed to constructor")
	}
}

func TestInputValidationScreen_GetRunAIInstallData(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	retrievedData := screen.GetRunAIInstallData()

	if retrievedData != runaiData {
		t.Error("expected to get the same runai data that was set")
	}
}

func TestInputValidationScreen_GetHeader(t *testing.T) {
	screen := NewInputValidationScreen(nil, nil)

	header := screen.GetHeader()
	expected := "Input Validation"

	if header != expected {
		t.Errorf("expected header to be '%s', got '%s'", expected, header)
	}
}

func TestInputValidationScreen_UpdateSelection(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// UpdateSelection should be a no-op for this screen
	screen.UpdateSelection(runaiData)

	// Should not crash or cause issues - this is mainly a compatibility method
}

func TestInputValidationScreen_validateNextItem_ValidIndex(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	if len(screen.validations) == 0 {
		t.Skip("no validations to test")
	}

	cmd := screen.validateNextItem(0)

	if cmd == nil {
		t.Error("expected command to be returned")
	}

	// Execute the command to get the result
	msg := cmd()

	if progressMsg, ok := msg.(validationProgressMsg); ok {
		if progressMsg.index != 0 {
			t.Errorf("expected index 0, got %d", progressMsg.index)
		}
		if progressMsg.result.Name == "" {
			t.Error("expected result to have a name")
		}
	} else {
		t.Errorf("expected validationProgressMsg, got %T", msg)
	}
}

func TestInputValidationScreen_validateNextItem_InvalidIndex(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Test with out-of-bounds index
	invalidIndex := len(screen.validations) + 10
	cmd := screen.validateNextItem(invalidIndex)

	if cmd == nil {
		t.Error("expected command to be returned even for invalid index")
	}

	// Execute the command to get the result
	msg := cmd()

	if progressMsg, ok := msg.(validationProgressMsg); ok {
		if progressMsg.result.Status {
			t.Error("expected validation to fail for invalid index")
		}
		if !strings.Contains(progressMsg.result.Message, "Unknown validation index") {
			t.Error("expected error message for invalid index")
		}
	} else {
		t.Errorf("expected validationProgressMsg, got %T", msg)
	}
}

func TestInputValidationScreen_validateNextItem_EmptyValidations(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	runaiData.Inputs = make(map[configinputs.InstallInputType]string) // Empty inputs
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	cmd := screen.validateNextItem(0)

	if cmd == nil {
		t.Error("expected command to be returned")
	}

	// Execute the command to get the result
	msg := cmd()

	if progressMsg, ok := msg.(validationProgressMsg); ok {
		if progressMsg.result.Status {
			t.Error("expected validation to fail for empty validations")
		}
		if !strings.Contains(progressMsg.result.Message, "No validations") {
			t.Error("expected 'No validations' message")
		}
	} else {
		t.Errorf("expected validationProgressMsg, got %T", msg)
	}
}

func TestInputValidationScreen_startProgressiveValidation(t *testing.T) {
	screen := NewInputValidationScreen(nil, nil)

	cmd := screen.startProgressiveValidation()

	if cmd == nil {
		t.Error("expected command to be returned")
	}

	// Execute the command to get the result
	msg := cmd()

	if startMsg, ok := msg.(validationStartMsg); ok {
		if startMsg.index != 0 {
			t.Errorf("expected start index to be 0, got %d", startMsg.index)
		}
	} else {
		t.Errorf("expected validationStartMsg, got %T", msg)
	}
}

func TestInputValidationScreen_Update_UnknownMessage(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Test with unknown message type
	cmd := screen.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should return nil for unknown messages
	if cmd != nil {
		t.Error("expected nil command for unknown message types")
	}
}

func TestInputValidationScreen_Render_StatusIcons(t *testing.T) {
	runaiData := newInputValidationTestData(false)
	screen := NewInputValidationScreen(nil, nil)
	screen.Init(runaiData)

	// Test different states
	if len(screen.validationResults) >= 3 {
		// Set different states for testing
		screen.validationResults[0].Status = true  // Success
		screen.validationResults[1].Status = false // Failed
		screen.currentValidation = 2               // Currently validating third item
	}

	output := screen.Render()

	// Check that different status icons are present
	if screen.validationResults[0].Status && !strings.Contains(output, "✓") {
		t.Error("expected success checkmark for successful validation")
	}

	// Check for failure mark
	if !screen.validationResults[1].Status && screen.currentValidation > 1 && !strings.Contains(output, "✗") {
		t.Error("expected error mark for failed validation")
	}
}

func TestInputValidationScreen_ValidationItem_Structure(t *testing.T) {
	// Test the ValidationItem struct - create a mock input
	configItem := configinputs.NewConfigurationItem("Test Label", "Test Placeholder", configinputs.InstallInputFQDN,
		func(value string) validation.ValidationResult {
			return validation.ValidationResult{Name: "Test", Status: true, Message: "Valid"}
		})
	input := inputs.NewTextInput(configItem, false)

	item := ValidationItem{
		Name:    "Test Validation",
		Enabled: true,
		Input:   input,
	}

	if item.Name != "Test Validation" {
		t.Errorf("expected name 'Test Validation', got %s", item.Name)
	}

	if !item.Enabled {
		t.Error("expected enabled to be true")
	}

	if item.Input != input {
		t.Error("expected input to be set correctly")
	}
}
