/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference

const (
	// Serving field names
	FieldLatestCreatedRevisionName        = "Latest Created Revision Name"
	FieldLatestReadyRevisionName          = "Latest Ready Revision Name"
	FieldMinReplicasName                  = "Minimum Replicas"
	FieldMaxReplicasName                  = "Maximum Replicas"
	FieldInitialReplicasName              = "Initial Replicas"
	FieldInitializationTimeoutSecondsName = "Initialization Timeout Seconds"
	FieldRequestTimeoutSecondsName        = "Request Timeout Seconds"
	FieldActivationReplicasName           = "Activation Replicas"
	FieldMetricName                       = "Metric"
	FieldMetricThresholdPercentageName    = "Metric Threshold Percentage"
	FieldMetricThresholdName              = "Metric Threshold"
	FieldScaleDownDelaySecondsName        = "Scale Down Delay Seconds"
	FieldScaleToZeroRetentionSecondsName  = "Scale To Zero Retention Seconds"
	FieldConcurrencyHardLimitName         = "Hard Limit"

	// Section title
	TitleServing = "Serving"
)
