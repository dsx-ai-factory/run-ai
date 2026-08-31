/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package appearance

import (
	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/setting"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type BannerConfig struct {
	Enabled         bool   `json:"enabled"`
	Text            string `json:"text"`
	BackgroundColor string `json:"backgroundColor"`
	FontColor       string `json:"fontColor"`
	FontSize        *int32 `json:"fontSize,omitempty"`
}

func NewBannerConfig() setting_pkg.Settinger {
	return &BannerConfig{}
}

func (b *BannerConfig) Category() string {
	return string(utils.UIAppearanceCategory)
}

func (b *BannerConfig) Description() string {
	return "Configure a persistent banner displayed at the top of the Run:AI UI for all authenticated users."
}

func (b *BannerConfig) DisplayName() string {
	return "Banner configuration"
}

func (b *BannerConfig) Key() string {
	return string(api.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG)
}

func (b *BannerConfig) SetDefault() {
	b.Enabled = false
	b.Text = ""
	b.BackgroundColor = ""
	b.FontColor = ""
}

func (b *BannerConfig) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

func (b *BannerConfig) Validate(_ interface{}) bool {
	return true
}

func (b *BannerConfig) TransformToResponse() interface{} {
	return api.GetBannerConfig{
		Value: api.GetBannerConfigValue{
			Enabled:         &b.Enabled,
			Text:            &b.Text,
			BackgroundColor: &b.BackgroundColor,
			FontColor:       &b.FontColor,
			FontSize:        b.FontSize,
		},
		DisplayName: b.DisplayName(),
		Description: b.Description(),
	}
}

func GetBannerConfig(tenantID string) (setting_pkg.Settinger, error) {
	return setting.Get[*BannerConfig](tenantID)
}

func UpdateBannerConfig(tenantID string, updateRequest any) error {
	req, ok := updateRequest.(api.UpdateBannerConfig)
	if !ok {
		return utils.ErrInvalidInput
	}

	existing, err := setting.Get[*BannerConfig](tenantID)
	if err != nil {
		return err
	}
	settingValue := *existing

	if req.Enabled != nil {
		settingValue.Enabled = *req.Enabled
	}
	if req.Text != nil {
		settingValue.Text = *req.Text
	}
	if req.BackgroundColor != nil {
		settingValue.BackgroundColor = *req.BackgroundColor
	}
	if req.FontColor != nil {
		settingValue.FontColor = *req.FontColor
	}
	if req.FontSize != nil {
		settingValue.FontSize = req.FontSize
	}

	if settingValue.Enabled && settingValue.Text == "" {
		return utils.ErrBannerTextRequiredWhenEnabled
	}

	return setting.Update[*BannerConfig](tenantID, settingValue).UpdatedByUser().Execute()
}
