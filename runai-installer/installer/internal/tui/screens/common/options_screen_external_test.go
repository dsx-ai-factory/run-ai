/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
	"github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common/mocks"
)

func TestOptionsModel_NewOptionsModel(t *testing.T) {
	mockOption1 := mocks.NewMockOptions(t)
	mockOption2 := mocks.NewMockOptions(t)
	options := []common.Options{mockOption1, mockOption2}

	// Common mocks for all tests
	mockOption1.On("IsDisabled").Return(false).Maybe()
	mockOption2.On("IsDisabled").Return(false).Maybe()
	mockOption1.On("IsMandatory").Return(false).Maybe()
	mockOption2.On("IsMandatory").Return(false).Maybe()
	mockOption1.On("IsSelected").Return(false).Maybe()
	mockOption2.On("IsSelected").Return(false).Maybe()

	t.Run("creates single-select model", func(t *testing.T) {
		model := common.NewOptionsModel(options, false, false)
		assert.NotNil(t, model)
		assert.Equal(t, 0, model.GetCursor())
		assert.Equal(t, 2, model.GetItemCount())
		assert.False(t, model.GetAllowMultipleSelection())
		assert.False(t, model.GetDevMode())
	})

	t.Run("creates multi-select model", func(t *testing.T) {
		model := common.NewOptionsModel(options, true, false)
		assert.NotNil(t, model)
		assert.True(t, model.GetAllowMultipleSelection())
	})

	t.Run("creates dev mode model", func(t *testing.T) {
		model := common.NewOptionsModel(options, false, true)
		assert.NotNil(t, model)
		assert.True(t, model.GetDevMode())
	})
}

func TestOptionsModel_CursorMovement(t *testing.T) {
	mockOption1 := mocks.NewMockOptions(t)
	mockOption2 := mocks.NewMockOptions(t)
	mockOption3 := mocks.NewMockOptions(t)

	mockOption1.On("IsSelected").Return(false).Maybe()
	mockOption2.On("IsSelected").Return(false).Maybe()
	mockOption3.On("IsSelected").Return(false).Maybe()

	mockOption1.On("IsDisabled").Return(false).Maybe()
	mockOption2.On("IsDisabled").Return(false).Maybe()
	mockOption3.On("IsDisabled").Return(false).Maybe()

	mockOption1.On("IsMandatory").Return(false).Maybe()
	mockOption2.On("IsMandatory").Return(false).Maybe()
	mockOption3.On("IsMandatory").Return(false).Maybe()

	options := []common.Options{mockOption1, mockOption2, mockOption3}

	t.Run("moves cursor down", func(t *testing.T) {
		model := common.NewOptionsModel(options, false, false)
		assert.Equal(t, 0, model.GetCursor())

		_ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		assert.Equal(t, 1, model.GetCursor())

		_ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		assert.Equal(t, 2, model.GetCursor())

		// Attempt to wrap around (should fail and return focus msg)
		cmd := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		assert.Equal(t, 2, model.GetCursor())
		assert.NotNil(t, cmd)
		msg := cmd()
		assert.IsType(t, common.FocusFrameMsg{}, msg)
	})

	t.Run("moves cursor up", func(t *testing.T) {
		model := common.NewOptionsModel(options, false, false)
		assert.Equal(t, 0, model.GetCursor())

		// Attempt to wrap around (should fail and return focus msg)
		cmd := model.Update(tea.KeyMsg{Type: tea.KeyUp})
		assert.Equal(t, 0, model.GetCursor())
		assert.NotNil(t, cmd)
		msg := cmd()
		assert.IsType(t, common.FocusFrameMsg{}, msg)
	})
}
