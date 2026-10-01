/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNGCChartRepo(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  config.RunaiEnvType
		want string
	}{
		{name: "prod", env: config.EnvProd, want: "https://helm.ngc.nvidia.com/nvidia/runai-oss"},
		{name: "staging", env: config.EnvStaging, want: ""},
		{name: "lab", env: config.EnvLab, want: ""},
		{name: "test", env: config.EnvTest, want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, NGCChartRepo(tc.env))
		})
	}
}

func TestChartAuth_IsZero(t *testing.T) {
	assert.True(t, ChartAuth{}.IsZero())
	assert.False(t, ChartAuth{Username: "u", Password: "p"}.IsZero())
}
