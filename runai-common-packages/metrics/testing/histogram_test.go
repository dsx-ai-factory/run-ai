/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
)

func TestHistogramObserve(t *testing.T) {
	manager := SetMock(t)
	manager.Histogram("histogram_test").Get().Observe(3.14)
	manager.Histogram("histogram_test").Register("histogram_test", nil, []float64{0.1, 1.0, 10.0}).Observe(3.14)

	metrics.Histogram("histogram_test").Get().Observe(3.14)
	metrics.Histogram("histogram_test").Register("histogram_test", nil, []float64{0.1, 1.0, 10.0}).Observe(3.14)
}
