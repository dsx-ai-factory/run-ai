/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNGCChartRepo(t *testing.T) {
	assert.Equal(t, "https://helm.ngc.nvidia.com/nvidia/runai",
		NGCChartRepo(config.EnvProd))

	assert.Equal(t, "https://helm.ngc.nvidia.com/nvstaging/runai",
		NGCChartRepo(config.EnvStaging))
}

func TestChartAuth_IsZero(t *testing.T) {
	assert.True(t, ChartAuth{}.IsZero())
	assert.False(t, ChartAuth{Username: "u", Password: "p"}.IsZero())
}
