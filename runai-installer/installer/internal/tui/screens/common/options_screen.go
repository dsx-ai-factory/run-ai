/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	tea "github.com/charmbracelet/bubbletea"
)

type OptionsModel struct {
	cursor                 int
	lastCursor             int
	options                []Options
	allowMultipleSelection bool
	selector               OptionsSelector
	devMode                bool
}

func NewOptionsModel(options []Options,
	allowMultipleSelection bool,
	devMode bool,
) *OptionsModel {
	optionsModel := &OptionsModel{
		options:                options,
		allowMultipleSelection: allowMultipleSelection,
		devMode:                devMode,
	}
	if allowMultipleSelection {
		optionsModel.selector = NewOptionsMultipleSelector(options)
	} else {
		optionsModel.selector = NewOptionsSingleSelect(options)
	}
	// Initialize cursor to first selectable (non-disabled, non-mandatory) option
	// If no selectable options exist, cursor stays at -1
	if optionsModel.GetItemCount() > 0 {
		optionsModel.cursor = optionsModel.findNextAvailableCursurPosition(0, false)
		optionsModel.lastCursor = optionsModel.cursor
	} else {
		optionsModel.cursor = -1
		optionsModel.lastCursor = -1
	}
	return optionsModel
}

func (o OptionsModel) Init() tea.Cmd {
	return nil
}

func (o *OptionsModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return o.handleKeyPress(msg)
	}
	return nil
}

func (o *OptionsModel) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	newCursor := o.handleSelection(msg)
	if newCursor == -2 {
		return func() tea.Msg { return FocusFrameMsg{FrameType: ScreenFrameTypeNavButtons} }
	}
	o.cursor = newCursor
	if newCursor != -1 {
		return ScreenSelectionChange()
	}
	return nil
}

func (o *OptionsModel) Render() string {
	return o.selector.render(o.cursor)
}

func (o *OptionsModel) SelectItem(nextIndex int) {
	o.selector.selectItem(nextIndex)
}

// GetItemCount returns the number of selectable items (not disabled and not mandatory)
func (o *OptionsModel) GetItemCount() int {
	count := 0
	for _, option := range o.options {
		if !option.IsDisabled() && !option.IsMandatory() {
			count++
		}
	}
	return count
}

func (o *OptionsModel) GetCursor() int {
	return o.cursor
}

func (o *OptionsModel) GetOptions() []Options {
	return o.options
}

// CanProceed returns true if the user can proceed to the next screen
func (o *OptionsModel) CanProceed() bool {
	if o.devMode {
		return true
	}
	for _, option := range o.options {
		mandatory := option.IsMandatory()
		selected := o.selector.isSelected(option.GetOrder())
		if mandatory && !selected {
			return false
		}
	}
	return true
}

func (o *OptionsModel) handleSelection(msg tea.KeyMsg) int {
	cursor := o.handleListKeys(msg, o.GetCursor())
	if cursor == -1 || cursor == -2 {
		o.lastCursor = o.cursor
	}
	if cursor >= 0 {
		switch msg.Type {
		case tea.KeySpace:
			o.SelectItem(cursor)
		}
	}
	return cursor
}

// GetAllowMultipleSelection returns whether multiple selection is allowed
func (o *OptionsModel) GetAllowMultipleSelection() bool {
	return o.allowMultipleSelection
}

// GetDevMode returns whether dev mode is enabled
func (o *OptionsModel) GetDevMode() bool {
	return o.devMode
}

func (o *OptionsModel) handleListKeys(msg tea.KeyMsg, cursor int) int {
	newCusorLocation := cursor
	totalOptions := len(o.options)
	switch msg.Type {
	case tea.KeyUp:
		start := cursor - 1
		if start < 0 {
			return -2
		}
		newPos := o.findNextAvailableCursurPosition(start, true)
		if newPos >= cursor {
			return -2
		}
		newCusorLocation = newPos
	case tea.KeyDown:
		start := cursor + 1
		if start >= totalOptions {
			return -2
		}
		newPos := o.findNextAvailableCursurPosition(start, false)
		if newPos <= cursor {
			return -2
		}
		newCusorLocation = newPos
	}
	return newCusorLocation
}

func (o *OptionsModel) findNextAvailableCursurPosition(cursor int, up bool) int {
	optionsCount := len(o.options)
	for i := 0; i < optionsCount; i++ {
		var idx int
		if !up {
			idx = (cursor + i) % optionsCount
		} else {
			idx = optionsCount - (i - cursor)
			idx = idx % optionsCount
		}
		option := o.options[idx]
		// Skip disabled and mandatory options when moving the cursor
		if !(option.IsDisabled() || option.IsMandatory()) {
			return idx
		}
	}
	return cursor
}

func (o *OptionsModel) OnFocus() {
	// If no selectable items, keep cursor at -1
	if o.GetItemCount() == 0 {
		o.cursor = -1
		return
	}
	if o.cursor == -1 {
		o.cursor = o.lastCursor
	}
	if o.cursor >= len(o.options) || o.cursor < 0 {
		o.cursor = o.findNextAvailableCursurPosition(0, false)
	}
}

func (o *OptionsModel) FocusFirst() {
	if o.GetItemCount() == 0 {
		o.cursor = -1
		return
	}
	o.cursor = o.findNextAvailableCursurPosition(0, false)
}

func (o *OptionsModel) FocusLast() {
	if o.GetItemCount() == 0 {
		o.cursor = -1
		return
	}
	o.cursor = o.findNextAvailableCursurPosition(len(o.options)-1, true)
}

func (o *OptionsModel) OnBlur() {
	if o.cursor != -1 {
		o.lastCursor = o.cursor
		o.cursor = -1
	}
}
