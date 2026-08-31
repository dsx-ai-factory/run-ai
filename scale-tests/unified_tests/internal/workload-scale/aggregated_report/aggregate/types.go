/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregate

type StagePerformanceMetric struct {
	PendingRate float64 `json:"pending_rate"`
	LatencyMS   Latency `json:"latency_ms"`
	SampleSize  int     `json:"sample_size"`
}

type Latency struct {
	Mean   float64 `json:"mean"`
	Max    float64 `json:"max"`
	Min    float64 `json:"min"`
	P95    float64 `json:"p95"`
	P50    float64 `json:"p50"`
	P75    float64 `json:"p75"`
	StdDev float64 `json:"std_dev"`
}

type SuccessFailureAnalysisMetric struct {
	TotalWorkloads    int                         `json:"total_workloads"`
	SuccessCount      int                         `json:"success_count"`
	FailureCount      int                         `json:"failure_count"`
	SuccessRate       float64                     `json:"success_rate"`
	FailureBreakdown  map[string]int              `json:"failure_breakdown"`
	TopFailureReasons []FailureReasonCount        `json:"top_failure_reasons"`
	StageFailures     map[string]StageFailureInfo `json:"stage_failures"`
}

type FailureReasonCount struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type StageFailureInfo struct {
	SuccessCount int `json:"success_count"`
	FailureCount int `json:"failure_count"`
}
