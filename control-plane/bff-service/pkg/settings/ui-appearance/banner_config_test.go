/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package appearance

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestBannerConfig_SetDefault(t *testing.T) {
	setting := &BannerConfig{}
	setting.SetDefault()
	assert.False(t, setting.Enabled)
	assert.Equal(t, "", setting.Text)
	assert.Equal(t, "", setting.BackgroundColor)
	assert.Equal(t, "", setting.FontColor)
}

func TestBannerConfig_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "valid BannerConfig struct",
			input:    BannerConfig{Enabled: true, Text: "test"},
			expected: true,
		},
		{
			name:     "valid BannerConfig pointer",
			input:    &BannerConfig{Enabled: false},
			expected: true,
		},
		{
			name:     "any value is valid",
			input:    "any string",
			expected: true,
		},
		{
			name:     "nil is valid",
			input:    nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &BannerConfig{}
			result := setting.Validate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBannerConfig_Key(t *testing.T) {
	setting := &BannerConfig{}
	assert.Equal(t, string(api.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG), setting.Key())
}

func TestBannerConfig_Category(t *testing.T) {
	setting := &BannerConfig{}
	assert.Equal(t, "UI-Appearance", setting.Category())
}

func TestBannerConfig_TransformToResponse(t *testing.T) {
	enabled := true
	text := "PRODUCTION"
	bgColor := "#D32F2F"
	fontColor := "#FFFFFF"

	setting := &BannerConfig{
		Enabled:         enabled,
		Text:            text,
		BackgroundColor: bgColor,
		FontColor:       fontColor,
	}

	result := setting.TransformToResponse()
	response, ok := result.(api.GetBannerConfig)
	assert.True(t, ok)
	assert.Equal(t, &enabled, response.Value.Enabled)
	assert.Equal(t, &text, response.Value.Text)
	assert.Equal(t, &bgColor, response.Value.BackgroundColor)
	assert.Equal(t, &fontColor, response.Value.FontColor)
	assert.Equal(t, "Banner configuration", response.DisplayName)
	assert.NotEmpty(t, response.Description)
}

func TestUpdateBannerConfig_AllFields(t *testing.T) {
	enabled := true
	text := "STAGING"
	bgColor := "#1565C0"
	fontColor := "#FFFFFF"

	req := api.UpdateBannerConfig{
		Enabled:         &enabled,
		Text:            &text,
		BackgroundColor: &bgColor,
		FontColor:       &fontColor,
	}

	setting := &BannerConfig{}
	if req.Enabled != nil {
		setting.Enabled = *req.Enabled
	}
	if req.Text != nil {
		setting.Text = *req.Text
	}
	if req.BackgroundColor != nil {
		setting.BackgroundColor = *req.BackgroundColor
	}
	if req.FontColor != nil {
		setting.FontColor = *req.FontColor
	}

	assert.True(t, setting.Enabled)
	assert.Equal(t, "STAGING", setting.Text)
	assert.Equal(t, "#1565C0", setting.BackgroundColor)
	assert.Equal(t, "#FFFFFF", setting.FontColor)
}
