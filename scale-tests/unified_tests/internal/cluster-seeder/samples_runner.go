/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_seeder

import (
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
)

type SamplerRunner interface {
	InitSampledEntities(clients *sampler_actions.Clients) ([]scale.Entity, error)
	UpdateRawResultReportMetaData(report *scale.SamplesReport)
	GetMetricLabels() map[string]string
}
