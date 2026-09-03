/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
)

func TestSummaryObserve(t *testing.T) {
	manager := SetMock(t)
	manager.Summary("summary_test").Get().Observe(2.71)
	manager.Summary("summary_test").Register("summary_test", nil, map[float64]float64{0.5: 0.05, 0.9: 0.01}).Observe(2.71)

	metrics.Summary("summary_test").Get().Observe(2.71)
	metrics.Summary("summary_test").Register("summary_test", nil, map[float64]float64{0.5: 0.05, 0.9: 0.01}).Observe(2.71)
}
