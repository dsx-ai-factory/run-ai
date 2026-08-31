/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/metrics"
)

func TestGaugeAdd(t *testing.T) {
	manager := SetMock(t)
	manager.Gauge("bla").Get().Add(1)
	manager.Gauge("bla").Register("bla", nil).Add(1)

	metrics.Gauge("bla").Get().Add(1)
	metrics.Gauge("bla").Register("bla", nil).Add(1)
}

func TestGaugeSet(t *testing.T) {
	manager := SetMock(t)
	manager.Gauge("bla").Get().Set(1)
	manager.Gauge("bla").Register("bla", nil).Set(1)

	metrics.Gauge("bla").Get().Set(1)
	metrics.Gauge("bla").Register("bla", nil).Set(1)
}

func TestGaugeInc(t *testing.T) {
	manager := SetMock(t)
	manager.Gauge("bla").Get().Inc()
	manager.Gauge("bla").Register("bla", nil).Inc()

	metrics.Gauge("bla").Get().Inc()
	metrics.Gauge("bla").Register("bla", nil).Inc()
}

func TestGaugeDec(t *testing.T) {
	manager := SetMock(t)
	manager.Gauge("bla").Get().Dec()
	manager.Gauge("bla").Register("bla", nil).Dec()

	metrics.Gauge("bla").Get().Dec()
	metrics.Gauge("bla").Register("bla", nil).Dec()
}

func TestGaugeSub(t *testing.T) {
	manager := SetMock(t)
	manager.Gauge("bla").Get().Sub(1)
	manager.Gauge("bla").Register("bla", nil).Sub(1)

	metrics.Gauge("bla").Get().Sub(1)
	metrics.Gauge("bla").Register("bla", nil).Sub(1)
}
