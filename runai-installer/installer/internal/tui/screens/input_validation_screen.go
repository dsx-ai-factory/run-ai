package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/run-ai/runai/runai-installer/installer/internal/install"
	configinput "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	validation "github.com/run-ai/runai/runai-installer/installer/internal/inputs/validation"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	prereqcomponents "github.com/run-ai/runai/runai-installer/installer/internal/prereq/components"
	inputs "github.com/run-ai/runai/runai-installer/installer/internal/tui/inputs"
	viewscommon "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
	styles "github.com/run-ai/runai/runai-installer/installer/internal/tui/styles"
)

// InputValidationScreen represents the input validation screen
type InputValidationScreen struct {
	validationResults []*validation.ValidationResult
	currentValidation int // Index of currently validating item
	runaiData         *model.RunAIInstallData
	validations       []*configinput.ConfigurationItem
	nextScreen        viewscommon.Screen
	prerequisitesMap  map[string]prereqcomponents.Prerequisite
}

// ValidationItem represents a single validation with its metadata
type ValidationItem struct {
	Name    string
	Enabled bool
	Input   inputs.Input
}

// NewInputValidationScreen creates a new input validation screen with configuration and prerequisite data
func NewInputValidationScreen(nextScreen viewscommon.Screen, prerequisitesMap map[string]prereqcomponents.Prerequisite) *InputValidationScreen {
	return &InputValidationScreen{
		currentValidation: 0, // Start with first validation
		nextScreen:        nextScreen,
		prerequisitesMap:  prerequisitesMap,
	}
}

// validationProgressMsg is a message sent when a single validation completes
type validationProgressMsg struct {
	index  int
	result validation.ValidationResult
}

// validationStartMsg is a message sent when starting a new validation
type validationStartMsg struct {
	index int
}

// Init initializes the screen
func (s *InputValidationScreen) Init(runaiData *model.RunAIInstallData) tea.Cmd {
	s.runaiData = runaiData

	// Get relevant input types using shared helper
	relevantInputs := install.GetRelevantInputTypes(runaiData, s.prerequisitesMap)

	// Initialize validation results only for relevant inputs that have values
	validationResults := []*validation.ValidationResult{}
	validations := []*configinput.ConfigurationItem{}
	for inputType := range relevantInputs {
		// Only validate if the input has a value in runaiData.Inputs
		if _, hasValue := s.runaiData.Inputs[inputType]; hasValue {
			configItem := configinput.GetConfigInputs()[inputType]
			if configItem != nil {
				validations = append(validations, configItem)
				validationResults = append(validationResults, &validation.ValidationResult{
					Name:    configItem.Label,
					Status:  false, // Pending status
					Message: "Waiting...",
				})
			}
		}
	}
	s.validationResults = validationResults
	s.validations = validations
	// Start progressive validation
	return s.startProgressiveValidation()
}

func (s *InputValidationScreen) GetMessages() []string {
	var messages []string
	for _, validationResult := range s.validationResults {
		if !validationResult.Status && validationResult.Message != "" {
			styledName := styles.AppStyles.Normal.Render(validationResult.Name + ": ")
			errorMessage := styledName + styles.AppStyles.ErrorRed.Render(validationResult.Message)
			messages = append(messages, errorMessage)
		}
	}
	return messages
}

// startProgressiveValidation starts the validation process progressively
func (s *InputValidationScreen) startProgressiveValidation() tea.Cmd {
	return func() tea.Msg {
		// Start with the first validation
		return validationStartMsg{index: 0}
	}
}

// validateNextItem validates the next item in the list
func (s *InputValidationScreen) validateNextItem(index int) tea.Cmd {
	return func() tea.Msg {
		if len(s.validations) == 0 {
			return validationProgressMsg{
				index: index,
				result: validation.ValidationResult{
					Name:    "No validations",
					Status:  false,
					Message: "No validations",
				},
			}
		}
		// Validate only the specific item at the given index
		if index >= 0 && index < len(s.validations) {
			configItem := s.validations[index]
			configValue := s.runaiData.Inputs[configItem.InputIdentifier]
			result := configItem.ValidationFunc(configValue)

			// Update the name to match the display name
			result.Name = s.validations[index].Label

			return validationProgressMsg{
				index:  index,
				result: result,
			}
		}

		// If index is out of bounds, return an error result
		return validationProgressMsg{
			index: index,
			result: validation.ValidationResult{
				Name:    "Unknown validation index",
				Status:  false,
				Message: "Unknown validation index",
			},
		}
	}
}

func (s *InputValidationScreen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case validationStartMsg:
		s.currentValidation = msg.index
		return s.validateNextItem(msg.index)
	case validationProgressMsg:
		nextIndex := msg.index + 1
		s.currentValidation = nextIndex
		s.validationResults[msg.index] = &msg.result
		if nextIndex < len(s.validationResults) {
			return s.validateNextItem(nextIndex)
		} else {
			return nil
		}
	}
	return nil
}

// View renders the screen content
func (s *InputValidationScreen) Render() string {
	var sb strings.Builder
	if !s.CanProceed() {
		sb.WriteString(styles.AppStyles.Normal.Render("Validating user inputs...") + "\n\n")
	} else {
		sb.WriteString(styles.AppStyles.Normal.Render("Validation complete.") + "\n\n")
	}

	for i, validationResult := range s.validationResults {
		var statusStyle lipgloss.Style
		var statusText string
		var nameStyle lipgloss.Style

		// Determine status and styling based on validation state
		if validationResult.Status {
			// Success
			statusStyle = styles.AppStyles.SuccessGreen
			statusText = "✓"
			nameStyle = styles.AppStyles.Normal
		} else {
			// Failure or pending
			if !s.CanProceed() && i == s.currentValidation {
				// Currently being validated - show sand clock
				statusStyle = styles.AppStyles.Warning
				statusText = "⏳"
				nameStyle = styles.AppStyles.Selected // Highlight the current item
			} else {
				// Failed validation
				statusStyle = styles.AppStyles.ErrorRed
				statusText = "✗"
				nameStyle = styles.AppStyles.Normal
			}
		}

		sb.WriteString(fmt.Sprintf("%s %s\n",
			statusStyle.Render(statusText),
			nameStyle.Render(validationResult.Name),
		))
	}
	return sb.String()
}

// CanProceed returns true if the user can proceed to the next screen
func (s *InputValidationScreen) CanProceed() bool {
	if s.runaiData.Config.DevMode {
		return true
	}
	// Check if all validations are successful
	for _, validationResult := range s.validationResults {
		if !validationResult.Status {
			return false
		}
	}
	return true
}

// NextScreen returns the next screen if the current screen can proceed, otherwise returns nil
func (s *InputValidationScreen) NextScreen() viewscommon.Screen {
	return s.nextScreen
}

// GetRunAIInstallData returns the RunAIInstallData for session persistence
func (s *InputValidationScreen) GetRunAIInstallData() *model.RunAIInstallData {
	return s.runaiData
}

func (s *InputValidationScreen) GetHeader() string {
	return "Input Validation"
}

func (s *InputValidationScreen) UpdateSelection(runaiData *model.RunAIInstallData) {
	s.runaiData = runaiData
}

func (s *InputValidationScreen) OnFocus() {
}

func (s *InputValidationScreen) OnBlur() {
}

func (s *InputValidationScreen) FocusFirst() {
}

func (s *InputValidationScreen) FocusLast() {
}

func (s *InputValidationScreen) GetItemCount() int {
	return 0
}
