/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package screens

import (
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/run-ai/runai/runai-installer/installer/internal/install"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	prereqcomponents "github.com/run-ai/runai/runai-installer/installer/internal/prereq/components"
	"github.com/run-ai/runai/runai-installer/installer/internal/session"
	"github.com/run-ai/runai/runai-installer/installer/internal/tui/inputs"
	screenscommon "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

// ConfigurationScreen represents the configuration input screen
type ConfigurationScreen struct {
	displayedInputs  []inputs.Input
	cursor           int
	lastCursor       int
	runaiData        *model.RunAIInstallData
	nextScreen       screenscommon.Screen
	k8sClient        k8s.K8sClients
	prerequisitesMap map[string]prereqcomponents.Prerequisite
}

// NewConfigurationScreen creates a new configuration screen
func NewConfigurationScreen(nextScreen screenscommon.Screen, k8sClient k8s.K8sClients, prerequisitesMap map[string]prereqcomponents.Prerequisite) *ConfigurationScreen {
	return &ConfigurationScreen{
		cursor:           0, // Start with first field focused
		lastCursor:       0,
		nextScreen:       nextScreen,
		k8sClient:        k8sClient,
		prerequisitesMap: prerequisitesMap,
	}
}

// Init initializes the screen
func (s *ConfigurationScreen) Init(runaiData *model.RunAIInstallData) tea.Cmd {
	s.runaiData = runaiData
	configParams := s.initConfigParameters(runaiData)
	s.displayedInputs = configParams
	return nil
}

func (s *ConfigurationScreen) GetMessages() []string {
	return nil
}

// Update handles messages and updates the screen state
func (s *ConfigurationScreen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		newCursor := s.handleInputScreenKeys(msg)
		if newCursor == -2 {
			return func() tea.Msg {
				return screenscommon.FocusFrameMsg{FrameType: screenscommon.ScreenFrameTypeNavButtons}
			}
		}
		s.cursor = newCursor
		if s.cursor != -1 {
			return screenscommon.ScreenSelectionChange()
		}
	}
	return nil
}

// View renders the screen content
func (s *ConfigurationScreen) Render() string {
	var sb strings.Builder
	for index, input := range s.displayedInputs {
		sb.WriteString(input.Render(index == s.cursor))
		sb.WriteString("\n")
	}
	return sb.String()
}

// CanProceed checks if the user can proceed to the next screen
func (s *ConfigurationScreen) CanProceed() bool {
	for _, input := range s.displayedInputs {
		if input.IsRequired() && input.GetValue() == "" {
			return false
		}
	}
	return true
}
func (s *ConfigurationScreen) GetHeader() string {
	return "Configurations"
}

// NextScreen returns the next screen if the current screen can proceed, otherwise returns nil
func (s *ConfigurationScreen) NextScreen() screenscommon.Screen {
	return s.nextScreen
}

func (s *ConfigurationScreen) UpdateSelection(runaiData *model.RunAIInstallData) {
	// Preserve existing inputs (including those not currently displayed)
	if runaiData.Inputs == nil {
		runaiData.Inputs = map[configinputs.InstallInputType]string{}
	}
	// Update only the displayed inputs, preserving hidden ones
	for _, input := range s.displayedInputs {
		if input.IsRequired() {
			runaiData.Inputs[input.GetInputIdentifier()] = input.GetValue()
		} else if input.GetValue() != "" {
			runaiData.Inputs[input.GetInputIdentifier()] = input.GetValue()
		} else {
			// Remove empty optional inputs to keep the map clean
			delete(runaiData.Inputs, input.GetInputIdentifier())
		}
	}
	// Save inputs immediately after submission
	_ = session.SaveInputs(s.k8sClient, runaiData.Inputs)
}

func (s *ConfigurationScreen) GetCursor() int {
	return s.cursor
}

func (s *ConfigurationScreen) GetItemCount() int {
	return len(s.displayedInputs)
}

func (s *ConfigurationScreen) handleInputScreenKeys(msg tea.KeyMsg) int {
	if msg.Type == tea.KeyDown {
		if s.cursor >= s.GetItemCount()-1 {
			return -2
		}
	}
	if msg.Type == tea.KeyUp {
		if s.cursor <= 0 {
			return -2
		}
	}

	cursor := screenscommon.HandleListKeys(msg, s.cursor, s.GetItemCount())
	if cursor == -1 {
		// When tab is pressed, we return -1.
		// The actual saving of lastCursor is handled in OnBlur for navigation.
		// But if internal navigation logic needs it:
		s.lastCursor = s.cursor
		return cursor
	}
	currentInput := s.displayedInputs[cursor]
	switch msg.Type {
	case tea.KeyBackspace:
		currentInput.DeleteInput()
	case tea.KeyDelete:
		currentInput.DeleteForward()
	case tea.KeyTab, tea.KeyEnter:
		// Do nothing
	case tea.KeySpace:
		currentInput.SelectInput(msg.String())
	case tea.KeyLeft:
		// Move cursor left within the text field
		currentInput.MoveCursorLeft()
	case tea.KeyRight:
		// Move cursor right within the text field
		currentInput.MoveCursorRight()
	case tea.KeyHome, tea.KeyCtrlA:
		// Move cursor to beginning of line
		currentInput.MoveCursorToBeginning()
	case tea.KeyEnd, tea.KeyCtrlE:
		// Move cursor to end of line
		currentInput.MoveCursorToEnd()
	case tea.KeyRunes:
		// Handle paste operations and regular text input
		if msg.Paste {
			// This is a paste operation (Cmd+V on Mac, Ctrl+V on other systems)
			clipboardText, err := clipboard.ReadAll()
			if err == nil && clipboardText != "" {
				currentInput.HandleInput(clipboardText)
			}
		} else if len(msg.String()) == 1 {
			// Handle regular text input for the currently focused field
			currentInput.HandleInput(msg.String())
		}
	case tea.KeyCtrlV:
		// Handle paste operation (Ctrl+V on Windows/Linux, also works on Mac)
		clipboardText, err := clipboard.ReadAll()
		if err == nil && clipboardText != "" {
			currentInput.HandleInput(clipboardText)
		}
	default:
		// Handle text input for the currently focused field
		if len(msg.String()) == 1 {
			currentInput.HandleInput(msg.String())
		}
	}
	return cursor
}

func (s *ConfigurationScreen) OnFocus() {
	if s.cursor == -1 {
		s.cursor = s.lastCursor
	}
	if s.cursor >= len(s.displayedInputs) {
		s.cursor = len(s.displayedInputs) - 1
	}
	if s.cursor < 0 && len(s.displayedInputs) > 0 {
		s.cursor = 0
	}
}

func (s *ConfigurationScreen) FocusFirst() {
	if len(s.displayedInputs) > 0 {
		s.cursor = 0
	}
}

func (s *ConfigurationScreen) FocusLast() {
	if len(s.displayedInputs) > 0 {
		s.cursor = len(s.displayedInputs) - 1
	}
}

func (s *ConfigurationScreen) OnBlur() {
	if s.cursor != -1 {
		s.lastCursor = s.cursor
		s.cursor = -1
	}
}

func (s *ConfigurationScreen) initConfigParameters(runaiData *model.RunAIInstallData) []inputs.Input {
	// Get relevant input types using shared helper
	relevantInputTypes := install.GetRelevantInputTypes(runaiData, s.prerequisitesMap)

	// Build input map with mandatory/optional flags
	inputsMap := map[configinputs.InstallInputType]inputs.Input{}
	for inputType, isMandatory := range relevantInputTypes {
		if configItem, ok := configinputs.GetConfigInputs()[inputType]; ok {
			input := inputs.NewTextInput(configItem, isMandatory)
			inputsMap[input.GetInputIdentifier()] = input
		}
	}

	// Set saved values
	for installInput, savedConfigValue := range runaiData.Inputs {
		if input, ok := inputsMap[installInput]; ok {
			input.SetValue(savedConfigValue)
		}
	}

	// Convert to sorted array
	inputArr := []inputs.Input{}
	for _, v := range inputsMap {
		inputArr = append(inputArr, v)
	}
	sort.Slice(inputArr, func(i, j int) bool {
		return GetInputsViewOrder()[inputArr[i].GetInputIdentifier()] < GetInputsViewOrder()[inputArr[j].GetInputIdentifier()]
	})
	return inputArr
}
