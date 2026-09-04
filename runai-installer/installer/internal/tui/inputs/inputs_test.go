/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputs

import (
	"testing"

	configinput "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/stretchr/testify/assert"
)

func TestTextInput_SetAndGetValue(t *testing.T) {
	ci := &configinput.ConfigurationItem{Label: "Test", Placeholder: "ph", InputIdentifier: configinput.InstallInputFQDN}
	input := NewTextInput(ci, true)
	input.SetValue("abc")
	assert.Equal(t, "abc", input.GetValue())
}

func TestTextInput_IsRequired(t *testing.T) {
	ci := &configinput.ConfigurationItem{Label: "Test", Placeholder: "ph", InputIdentifier: configinput.InstallInputFQDN}
	input := NewTextInput(ci, true)
	assert.True(t, input.IsRequired())
	input2 := NewTextInput(ci, false)
	assert.False(t, input2.IsRequired())
}

func TestTextInput_CursorMovement(t *testing.T) {
	ci := &configinput.ConfigurationItem{Label: "Test", Placeholder: "ph", InputIdentifier: configinput.InstallInputFQDN}
	input := NewTextInput(ci, true)
	input.SetValue("abc")
	input.MoveCursorToBeginning()
	assert.Equal(t, 0, input.GetTextCursor())
	input.MoveCursorToEnd()
	assert.Equal(t, 3, input.GetTextCursor())
	input.MoveCursorLeft()
	assert.Equal(t, 2, input.GetTextCursor())
	input.MoveCursorRight()
	assert.Equal(t, 3, input.GetTextCursor())
}

func TestTextInput_HandleInput(t *testing.T) {
	ci := &configinput.ConfigurationItem{Label: "Test", Placeholder: "ph", InputIdentifier: configinput.InstallInputFQDN}
	input := NewTextInput(ci, true)
	input.HandleInput("a")
	input.HandleInput("b")
	input.HandleInput("c")
	assert.Equal(t, "abc", input.GetValue())
	assert.Equal(t, 3, input.GetTextCursor())
}

func TestTextInput_DeleteInput(t *testing.T) {
	ci := &configinput.ConfigurationItem{Label: "Test", Placeholder: "ph", InputIdentifier: configinput.InstallInputFQDN}
	input := NewTextInput(ci, true)
	input.SetValue("abc")
	input.MoveCursorToEnd()
	input.DeleteInput()
	assert.Equal(t, "ab", input.GetValue())
	input.DeleteInput()
	input.DeleteInput()
	assert.Equal(t, "", input.GetValue())
}

func TestTextInput_SelectAndSetSelected(t *testing.T) {
	ci := &configinput.ConfigurationItem{Label: "Test", Placeholder: "ph", InputIdentifier: configinput.InstallInputFQDN}
	input := NewTextInput(ci, true)
	assert.False(t, input.selected)
	input.SelectInput("")
	assert.True(t, input.selected)
	input.SetSelected(false)
	assert.False(t, input.selected)
}

func TestTextInput_GetInputIdentifier(t *testing.T) {
	ci := &configinput.ConfigurationItem{Label: "Test", Placeholder: "ph", InputIdentifier: configinput.InstallInputFQDN}
	input := NewTextInput(ci, true)
	assert.Equal(t, configinput.InstallInputFQDN, input.GetInputIdentifier())
}
