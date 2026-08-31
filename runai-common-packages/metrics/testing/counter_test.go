/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/metrics"
)

func TestCounterAdd(t *testing.T) {
	manager := SetMock(t)
	manager.Counter("counter_test").Get().Add(5)
	manager.Counter("counter_test").Register("counter_test", nil).Add(5)

	metrics.Counter("counter_test").Get().Add(5)
	metrics.Counter("counter_test").Register("counter_test", nil).Add(5)
}

func TestCounterInc(t *testing.T) {
	manager := SetMock(t)
	manager.Counter("counter_test").Get().Inc()
	manager.Counter("counter_test").Register("counter_test", nil).Inc()

	metrics.Counter("counter_test").Get().Inc()
	metrics.Counter("counter_test").Register("counter_test", nil).Inc()
}
