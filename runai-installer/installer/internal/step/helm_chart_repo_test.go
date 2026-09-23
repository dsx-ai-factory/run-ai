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

func TestChartRepos_FromEnvBase(t *testing.T) {
	t.Setenv(ChartRepoBaseEnv, " https://charts.example.com/api/helm/ ")

	for _, tc := range []struct {
		name    string
		env     config.RunaiEnvType
		wantCP  string
		wantCLS string
	}{
		{name: "prod", env: config.EnvProd, wantCP: "https://charts.example.com/api/helm/cp-charts-prod", wantCLS: "https://charts.example.com/api/helm/run-ai-charts"},
		{name: "lab", env: config.EnvLab, wantCP: "https://charts.example.com/api/helm/cp-charts-lab", wantCLS: "https://charts.example.com/api/helm/op-charts-lab"},
		{name: "staging", env: config.EnvStaging, wantCP: "https://charts.example.com/api/helm/cp-charts-staging", wantCLS: "https://charts.example.com/api/helm/op-charts-staging"},
		{name: "test", env: config.EnvTest, wantCP: "https://charts.example.com/api/helm/cp-charts-test", wantCLS: "https://charts.example.com/api/helm/run-ai-charts-test"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantCP, CPChartRepo(tc.env))
			assert.Equal(t, tc.wantCLS, ClusterChartRepo(tc.env))
		})
	}
}

func TestChartRepos_EmptyWithoutBase(t *testing.T) {
	t.Setenv(ChartRepoBaseEnv, "")

	assert.Empty(t, ChartRepoBase())
	assert.Empty(t, CPChartRepo(config.EnvProd))
	assert.Empty(t, ClusterChartRepo(config.EnvProd))
	assert.ErrorContains(t, MissingChartRepoError(CPChartName), ChartRepoBaseEnv)
}
