/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
)

// ScreenType represents different screens in the TUI
type ScreenType int

const (
	// ScreenInstallationType
	// ScreenDeploymentEnvironment
	ScreenPrerequisites ScreenType = iota
	ScreenConfiguration
	ScreenPrerequisiteInstall
	ScreenInputValidation
	ScreenInstallation
	ScreenComplete
)

type ScreenFrameType int

const (
	ScreenFrameTypeMain ScreenFrameType = iota
	ScreenFrameTypeErrors
	ScreenFrameTypeNavButtons
)

type FocusPosition int

const (
	FocusPositionDefault FocusPosition = iota
	FocusPositionFirst
	FocusPositionLast
)

type FocusFrameMsg struct {
	FrameType ScreenFrameType
	Position  FocusPosition
}

// Screen interface that all screens must implement
type Screen interface {
	Update(msg tea.Msg) tea.Cmd

	Init(runaiData *model.RunAIInstallData) tea.Cmd

	NextScreen() Screen

	UpdateSelection(runaiData *model.RunAIInstallData)

	GetHeader() string

	CanProceed() bool

	Render() string
	// GetMessages returns messages to display to the user
	GetMessages() []string

	OnFocus()
	OnBlur()

	FocusFirst()
	FocusLast()

	GetItemCount() int
}

type OptionsSelector interface {
	selectItem(nextIndex int)
	render(cursor int) string
	isSelected(index int) bool
}

type Options interface {
	GetUiLabel() string
	GetOrder() int
	IsMandatory() bool
	IsSelected() bool
	IsDisabled() bool
	SetSelected(selected bool)
}

type SimpleOption struct {
	name     string
	id       int
	selected bool
}

func NewSimpleOption(name string, id int, selected bool) *SimpleOption {
	return &SimpleOption{
		name:     name,
		id:       id,
		selected: selected,
	}
}

func (o *SimpleOption) GetUiLabel() string {
	return o.name
}

func (o *SimpleOption) GetOrder() int {
	return o.id
}

func (o *SimpleOption) IsMandatory() bool {
	return false
}

func (o *SimpleOption) IsSelected() bool {
	return o.selected
}

func (o *SimpleOption) SetSelected(selected bool) {
	o.selected = selected
}

func (o *SimpleOption) IsDisabled() bool {
	return false
}
