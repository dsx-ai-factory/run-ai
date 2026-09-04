/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_test

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	ui_workloads "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/settings/ui-workloads"
	"github.com/stretchr/testify/assert"
)

func TestAllowGPT2Model_SetDefault(t *testing.T) {
	setting := &ui_workloads.AllowGPT2Model{}
	setting.SetDefault()
	assert.False(t, setting.Enabled)
}

func TestAllowGPT2Model_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name: "valid AllowGPT2Model struct",
			input: ui_workloads.AllowGPT2Model{
				Enabled: true,
			},
			expected: true,
		},
		{
			name: "valid AllowGPT2Model pointer",
			input: &ui_workloads.AllowGPT2Model{
				Enabled: false,
			},
			expected: true,
		},
		{
			name:     "valid any value",
			input:    "any value",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &ui_workloads.AllowGPT2Model{}
			result := setting.Validate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllowGPT2Model_TransformForMigration(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:  "valid GetAllowGpt2Model from API with true",
			input: true,
			expected: ui_workloads.AllowGPT2Model{
				Enabled: true,
			},
		},
		{
			name:  "valid GetAllowGpt2Model with false",
			input: false,
			expected: ui_workloads.AllowGPT2Model{
				Enabled: false,
			},
		},
		{
			name:     "invalid input type returns as-is",
			input:    "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &ui_workloads.AllowGPT2Model{}
			result := setting.TransformForMigration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllowGPT2Model_TransformToResponse(t *testing.T) {
	tests := []struct {
		name     string
		setting  ui_workloads.AllowGPT2Model
		expected api.GetAllowGpt2Model
	}{
		{
			name: "transform enabled true to response",
			setting: ui_workloads.AllowGPT2Model{
				Enabled: true,
			},
			expected: api.GetAllowGpt2Model{
				Value: api.GetAllowGpt2ModelValue{
					Enabled: boolPtr(true),
				},
				DisplayName: "Allow GPT-2 model",
				Description: "GPT2 model - present OpenAI's GPT2 model in the catalog in favor of demos. This model's compute requirements are relatively low (CPU resources are sufficient for it). The underlying image of this model also consists of a Chatbot UI",
			},
		},
		{
			name: "transform enabled false to response",
			setting: ui_workloads.AllowGPT2Model{
				Enabled: false,
			},
			expected: api.GetAllowGpt2Model{
				Value: api.GetAllowGpt2ModelValue{
					Enabled: boolPtr(false),
				},
				DisplayName: "Allow GPT-2 model",
				Description: "GPT2 model - present OpenAI's GPT2 model in the catalog in favor of demos. This model's compute requirements are relatively low (CPU resources are sufficient for it). The underlying image of this model also consists of a Chatbot UI",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setting.TransformToResponse()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllowGPT2Model_Metadata(t *testing.T) {
	setting := &ui_workloads.AllowGPT2Model{}

	t.Run("Category", func(t *testing.T) {
		assert.Equal(t, "UI-Workloads", setting.Category())
	})

	t.Run("Key", func(t *testing.T) {
		assert.Equal(t, string(api.WORKLOADSSETTINGSKEYENUM_ALLOW_GPT2_MODEL), setting.Key())
	})

	t.Run("DisplayName", func(t *testing.T) {
		assert.Equal(t, "Allow GPT-2 model", setting.DisplayName())
	})

	t.Run("Description", func(t *testing.T) {
		assert.Equal(t, "GPT2 model - present OpenAI's GPT2 model in the catalog in favor of demos. This model's compute requirements are relatively low (CPU resources are sufficient for it). The underlying image of this model also consists of a Chatbot UI", setting.Description())
	})

	t.Run("Stage", func(t *testing.T) {
		assert.Equal(t, "Alpha", string(setting.Stage()))
	})
}
