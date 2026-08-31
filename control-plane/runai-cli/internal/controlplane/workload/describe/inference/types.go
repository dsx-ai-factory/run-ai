/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference

// Fields contains serving/autoscaling information for inference workloads
type Fields struct {
	LatestCreatedRevisionName    *string  `json:"latestCreatedRevisionName"`
	LatestReadyRevisionName      *string  `json:"latestReadyRevisionName"`
	MinReplicas                  *int32   `json:"minReplicas"`
	MaxReplicas                  *int32   `json:"maxReplicas"`
	InitialReplicas              *int32   `json:"initialReplicas"`
	InitializationTimeoutSeconds *int32   `json:"initializationTimeoutSeconds"`
	RequestTimeoutSeconds        *int32   `json:"requestTimeoutSeconds"`
	ActivationReplicas           *int32   `json:"activationReplicas"`
	Metric                       *string  `json:"metric,omitempty"`
	MetricThresholdPercentage    *float32 `json:"metricThresholdPercentage,omitempty"`
	MetricThreshold              *int32   `json:"metricThreshold,omitempty"`
	ScaleDownDelaySeconds        *int32   `json:"scaleDownDelaySeconds,omitempty"`
	ScaleToZeroRetentionSeconds  *int32   `json:"scaleToZeroRetentionSeconds,omitempty"`
	ConcurrencyHardLimit         *int32   `json:"concurrencyHardLimit,omitempty"`
}
