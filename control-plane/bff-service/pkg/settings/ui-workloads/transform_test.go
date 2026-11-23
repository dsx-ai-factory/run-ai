/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_test

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	ui_workloads "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/settings/ui-workloads"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The aggregate endpoint and the per-setting endpoint describe the same
// setting, so they have to return the same value: a field that only one of them
// fills is reported as unset through that endpoint.
func TestTransformToWorkloadsSettingsResponseMatchesTheSettingTransform(t *testing.T) {
	setting := &ui_workloads.AllowGPT2Model{Enabled: true}

	response := ui_workloads.TransformToWorkloadsSettingsResponse([]setting_pkg.Settinger{setting})

	settings, ok := response.(*api.WorkloadsSettingsGetAllResponse)
	require.True(t, ok, "got %T, want the workloads settings response", response)
	assert.Equal(t, setting.TransformToResponse(), settings.AllowGPT2Model)
	assert.Equal(t, "Alpha", settings.AllowGPT2Model.Stage)
}
