/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type installationStartMsg struct{}

type statusPollMsg struct{}

type InstallationScreen struct {
	uiFlowManager  *uiFlowManager
	currentSpinner spinner.Model
	runaiData      *model.RunAIInstallData
	nextScreen     Screen
	screenHeader   string
	events         chan interface{}
}

func NewInstallationScreen(screenHeader string, nextScreen Screen) *InstallationScreen {
	s := spinner.New()
	// Braille dot spinner, clockwise rotation
	s.Spinner = spinner.Spinner{
		Frames: []string{"⠏", "⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇"},
		FPS:    time.Millisecond * 100,
	}
	return &InstallationScreen{
		nextScreen:     nextScreen,
		screenHeader:   screenHeader,
		uiFlowManager:  NewUIFlowManager(),
		events:         make(chan interface{}, 1),
		currentSpinner: s,
	}
}

func (s *InstallationScreen) Init(runaiData *model.RunAIInstallData) tea.Cmd {
	s.runaiData = runaiData
	return tea.Batch(listenForExternalEvents(s.events), s.startInstallation())
}

func (s *InstallationScreen) SetFlows(flows []*flow.InstallFlow) {
	s.uiFlowManager.SetFlows(flows)
}

func (s *InstallationScreen) startInstallation() tea.Cmd {
	return func() tea.Msg {
		return installationStartMsg{}
	}
}

func listenForExternalEvents(ch <-chan interface{}) tea.Cmd {
	return func() tea.Msg {
		msg := <-ch // blocks until a message arrives
		return msg.(tea.Msg)
	}
}

func (s *InstallationScreen) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		s.currentSpinner, cmd = s.currentSpinner.Update(msg)
		return cmd
	case installationStartMsg:
		// Start the single tick chain and process first step
		cmds = append(cmds, s.currentSpinner.Tick, scheduleStatusPoll())
	case flow.NextStepMsg:
		s.uiFlowManager.flowManager.ProceedToNextStep()
		if s.uiFlowManager.flowManager.IsDone() {
			s.currentSpinner = spinner.New()
			return nil
		}
		cmds = append(cmds, listenForExternalEvents(s.events))
	case flow.CurrentStepStatusCheckMsg:
		// Check status without starting tick (tick chain already running)
		s.checkStatusAndAdvance()
		cmds = append(cmds, listenForExternalEvents(s.events))
	case statusPollMsg:
		// Tick chain: always schedule next tick, then check status
		// This is the ONLY place scheduleStatusPoll() is returned, ensuring single tick chain
		// GetStatusAsync() is non-blocking - it returns cached status immediately
		// while status evaluation runs in the background (guarded against goroutine pileup)
		s.checkStatusAndAdvance()
		cmds = append(cmds, s.currentSpinner.Tick, scheduleStatusPoll(), listenForExternalEvents(s.events))
	}
	return tea.Batch(cmds...)
}

func (s *InstallationScreen) checkStatusAndAdvance() {
	s.uiFlowManager.flowManager.ProvisionStep(context.Background(), s.runaiData, s.events)
}

func (s *InstallationScreen) Render() string {
	return s.uiFlowManager.Render(s.currentSpinner, s.runaiData, s.runaiData.Config.StepTimeout)
}

func (s *InstallationScreen) NextScreen() Screen {
	return s.nextScreen
}

func (s *InstallationScreen) GetHeader() string {
	return s.screenHeader
}

func (s *InstallationScreen) CanProceed() bool {
	return true
}

func (s *InstallationScreen) UpdateSelection(runaiData *model.RunAIInstallData) {
}

func (f *InstallationScreen) GetMessages() []string {
	return f.uiFlowManager.getMessages()
}

func (s *InstallationScreen) OnFocus() {
}

func (s *InstallationScreen) OnBlur() {
}

func (s *InstallationScreen) FocusFirst() {
}

func (s *InstallationScreen) FocusLast() {
}

func (s *InstallationScreen) GetItemCount() int {
	return 0
}

func scheduleStatusPoll() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return statusPollMsg{}
	})
}
