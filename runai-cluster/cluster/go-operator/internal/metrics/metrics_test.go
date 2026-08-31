/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperandStatusMetric_SetAndGet(t *testing.T) {
	// Set a value
	OperandStatus.WithLabelValues("test-operand", OperandStatusConditionReady).Set(1)

	// Gather the metric
	metrics, err := OperandStatus.GetMetricWithLabelValues("test-operand", OperandStatusConditionReady)
	assert.NoError(t, err)

	// The prometheus.Gauge does not expose a direct Get, but we can check that no panic occurs and the metric is registered
	assert.NotNil(t, metrics)
}
