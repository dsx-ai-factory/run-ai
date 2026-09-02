/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/cluster"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/utility"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/choises"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/inputs"
)

type SettingSelection string

const (
	SettingsInteractive    SettingSelection = "interactive"
	SettingsServers        SettingSelection = "servers"
	SettingsStatusDuration SettingSelection = "statusDuration"

	SelectedEnabled  string = "Enabled"
	SelectedDisabled string = "Disabled"

	MaxUrlLength      = 1024
	MaxDurationLength = 10

	DocStyleMarginIndex1 = 1
	DocStyleMarginIndex2 = 2
)

var docStyle = lipgloss.NewStyle().Margin(DocStyleMarginIndex1, DocStyleMarginIndex2)
var selected SettingSelection

type item struct {
	title, desc string
	setting     SettingSelection
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case ui.KeyQ, ui.CtrlPlusC:
			selected = ""
			return m, tea.Quit
		case ui.KeyEnter:
			i, ok := m.list.SelectedItem().(item)
			if ok {
				selected = i.setting
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func buildItemList() []list.Item {
	interactive := config.GetInstance().GetConfigValueWithDefault(config.SettingsInteractive, consts.DefaultInteractive)
	items := []list.Item{
		item{title: "Interactive Mode", desc: "Enable/Disable Interactive mode, " + "Enabled: " + interactive, setting: SettingsInteractive},
		item{title: "Server Settings", desc: "Control Plane/Authentication URL", setting: SettingsServers},
		item{title: "Status Timeout Duration", desc: "Set duration timeout for cluster status", setting: SettingsStatusDuration},
	}
	return items
}

func generateFieldMap(cpUrl string, authUrl string, cpInitValue string, authInitValue string) []textinput.Model {
	return []textinput.Model{
		{
			Placeholder: fmt.Sprintf(consts.SelectionInputWithVisibleDefault, consts.SelectionCPUrl, cpUrl),
			CharLimit:   MaxUrlLength,
			Prompt:      cpInitValue,
		},
		{
			Placeholder: fmt.Sprintf(consts.SelectionInputWithVisibleDefault, consts.SelectionAuthUrl, authUrl),
			CharLimit:   MaxUrlLength,
			Prompt:      authInitValue,
		},
	}
}

func generateStatusFieldMap(asdt string, durationInitValue string) []textinput.Model {
	return []textinput.Model{
		{
			Placeholder: fmt.Sprintf(consts.SelectionInputWithVisibleDefault, consts.SelectionStatusDurationTimeout, asdt),
			CharLimit:   MaxDurationLength,
			Prompt:      durationInitValue,
		},
	}
}

func ConfigurationInteractive() error {
	var keepRunning = true
	for keepRunning {
		m := model{list: list.New(buildItemList(), list.NewDefaultDelegate(), 0, 0)}
		m.list.Title = "CLI Settings"

		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}

		switch selected {
		case SettingsInteractive:
			decision, _ := choises.ExecuteTea("Interactive State:", "", []string{SelectedEnabled, SelectedDisabled})
			switch decision {
			case SelectedEnabled:
				_ = config.GetInstance().SetInteractive(true)
			case SelectedDisabled:
				_ = config.GetInstance().SetInteractive(false)
			default:
				logger := config.GetInstance().GetLogger()
				logger.Debug("no selection was made")
			}
		case SettingsServers:
			keepRunningServerSettings := true
			var displayError, cpInitValue, authInitValue string
			for keepRunningServerSettings {
				cpUrl := config.GetInstance().GetConfigValue(config.ControlPlaneUrl)
				authUrl := config.GetInstance().GetConfigValue(config.AuthenticationUrl)
				decisions, _ := inputs.MultiTexts(generateFieldMap(cpUrl, authUrl, cpInitValue, authInitValue), displayError)
				if decisions != nil {
					cpInitValue = (*decisions)[fmt.Sprintf(consts.SelectionInputWithVisibleDefault, consts.SelectionCPUrl, cpUrl)]
					authInitValue = (*decisions)[fmt.Sprintf(consts.SelectionInputWithVisibleDefault, consts.SelectionAuthUrl, authUrl)]
					displayError, keepRunningServerSettings = CheckDecision(decisions)
				} else {
					keepRunningServerSettings = false
				}

			}
		case SettingsStatusDuration:
			keepRunningStatusDuration := true
			var displayError, durationInitValue string
			for keepRunningStatusDuration {
				timeoutDuration := config.GetInstance().GetConfigValueWithDefault(config.ApiStatusTimeoutDuration, cluster.DefaultApiStatusDurationTimeout)
				decisions, _ := inputs.MultiTexts(generateStatusFieldMap(timeoutDuration, durationInitValue), displayError)
				if decisions != nil {
					durationInitValue = (*decisions)[fmt.Sprintf(consts.SelectionInputWithVisibleDefault, consts.SelectionStatusDurationTimeout, timeoutDuration)]
					displayError, keepRunningStatusDuration = CheckDecision(decisions)
				} else {
					keepRunningStatusDuration = false
				}
			}
		default:
			keepRunning = false
		}
	}
	return nil
}

func CheckDecision(decisions *map[string]string) (string, bool) {
	var decisionErrors []error
	for key, value := range *decisions {
		if value == "" {
			continue
		}
		if strings.HasPrefix(key, consts.SelectionCPUrl) {
			if err := utility.ValidateUrl(value); err != nil {
				decisionErrors = append(decisionErrors, fmt.Errorf("'%s': %w", value, err))
				continue
			}
			_ = config.GetInstance().SetCPURL(value)
		}
		if strings.HasPrefix(key, consts.SelectionAuthUrl) {
			if err := utility.ValidateUrl(value); err != nil {
				decisionErrors = append(decisionErrors, fmt.Errorf("'%s': %w", value, err))
				continue
			}
			_ = config.GetInstance().SetAuthURL(value)
		}
		if strings.HasPrefix(key, consts.SelectionStatusDurationTimeout) {
			if duration, err := time.ParseDuration(value); err != nil || duration < 0 {
				if duration < 0 {
					err = errors.New("duration cannot be negative")
				}
				decisionErrors = append(decisionErrors, fmt.Errorf("invalid duration '%s': %w", value, err))
				continue
			}
			_ = config.GetInstance().SetStatusDuration(value)
		}
	}

	if len(decisionErrors) == 0 {
		return "", false
	}

	var errStrings []string
	for _, err := range decisionErrors {
		errStrings = append(errStrings, err.Error())
	}
	return strings.Join(errStrings, "\n"), true
}
