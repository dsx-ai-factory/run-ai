/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

type DynamoGraphDeploymentState string

const (
	ComponentTypePlanner  = "planner"
	ComponentTypeFrontend = "frontend"
	ComponentTypeWorker   = "worker"
	ComponentTypePrefill  = "prefill"
	ComponentTypeDecode   = "decode"
	ComponentTypeDefault  = "default"
)

type DynamoGraphDeploymentReason string

const (
	SomeResourcesNotReady         DynamoGraphDeploymentReason = "some_resources_are_not_ready"
	AllResourcesAreReady          DynamoGraphDeploymentReason = "all_resources_are_ready"
	FailedToHandleTheFinalizer    DynamoGraphDeploymentReason = "failed_to_handle_the_finalizer"
	FailedToReconcileTheResources DynamoGraphDeploymentReason = "failed_to_reconcile_the_resources"
	ValidationFailed              DynamoGraphDeploymentReason = "ValidationFailed"
	Undefined                     DynamoGraphDeploymentReason = "undefined"
)
