/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package appearance

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The aggregate endpoint and the per-setting endpoint describe the same banner,
// so they have to return the same value: a field that only one of them fills is
// reported as unset through that endpoint.
func TestTransformToAppearanceSettingsResponseMatchesTheSettingTransform(t *testing.T) {
	fontSize := int32(18)
	banner := &BannerConfig{
		Enabled:         true,
		Text:            "Maintenance window",
		BackgroundColor: "#000000",
		FontColor:       "#ffffff",
		FontSize:        &fontSize,
	}

	response := TransformToAppearanceSettingsResponse([]setting_pkg.Settinger{banner})

	settings, ok := response.(*api.AppearanceSettingsGetAllResponse)
	require.True(t, ok, "got %T, want the appearance settings response", response)
	assert.Equal(t, banner.TransformToResponse(), settings.BannerConfig)
	require.NotNil(t, settings.BannerConfig.Value.FontSize)
	assert.Equal(t, fontSize, *settings.BannerConfig.Value.FontSize)
}

func TestTransformToAppearanceSettingsResponseIgnoresOtherSettings(t *testing.T) {
	response := TransformToAppearanceSettingsResponse(nil)

	settings, ok := response.(*api.AppearanceSettingsGetAllResponse)
	require.True(t, ok)
	assert.Nil(t, settings.BannerConfig.Value.Text)
}
