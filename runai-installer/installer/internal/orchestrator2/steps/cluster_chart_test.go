/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/stretchr/testify/require"
)

func TestClusterChartPostValidateSkipsSuppliedCredentials(t *testing.T) {
	s := NewClusterChart("install-cluster-chart", actions.InstallClusterChart{
		ClusterUID:  "cluster-uid",
		ReleaseName: "runai-cluster",
		Namespace:   "runai",
	}, orchestrator2.Deps{}).(*clusterChartStep)

	require.NoError(t, s.PostValidate(context.Background(), nil))
}
