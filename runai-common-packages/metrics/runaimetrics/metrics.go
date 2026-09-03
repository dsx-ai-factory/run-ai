/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

// MetricCounter is an interface for a counter metric.
type MetricCounter interface {
	Inc()
	Add(value float64)
}

// MetricGauge is an interface for a gauge metric.
type MetricGauge interface {
	Set(value float64)
	Inc()
	Dec()
	Add(value float64)
	Sub(value float64)
}

// MetricHistogram is an interface for a histogram metric.
type MetricHistogram interface {
	Observe(value float64)
}

// MetricSummary is an interface for a summary metric.
type MetricSummary interface {
	Observe(value float64)
}
