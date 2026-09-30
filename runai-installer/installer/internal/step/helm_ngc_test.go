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
		name    string
		env     config.RunaiEnvType
		nonProd string
		want    string
	}{
		{name: "prod ignores the non-prod repo", env: config.EnvProd, nonProd: "https://helm.example.com/org/team", want: "https://helm.ngc.nvidia.com/nvidia/runai"},
		{name: "prod without the non-prod repo", env: config.EnvProd, want: "https://helm.ngc.nvidia.com/nvidia/runai"},
		{name: "staging uses the non-prod repo", env: config.EnvStaging, nonProd: "https://helm.example.com/org/team/", want: "https://helm.example.com/org/team"},
		{name: "test uses the non-prod repo", env: config.EnvTest, nonProd: "https://helm.example.com/org/team", want: "https://helm.example.com/org/team"},
		{name: "non-prod without the non-prod repo", env: config.EnvLab, want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(NGCNonProdChartRepoEnv, tc.nonProd)
			assert.Equal(t, tc.want, NGCChartRepo(tc.env))
		})
	}
}

func TestChartAuth_IsZero(t *testing.T) {
	assert.True(t, ChartAuth{}.IsZero())
	assert.False(t, ChartAuth{Username: "u", Password: "p"}.IsZero())
}
