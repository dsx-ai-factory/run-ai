/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

// RunAI label and annotation constants used across the platform
const (
	WorkloadLabel = "run.ai/workload"

	ProjectIdLabel = "run.ai/project"

	CategoryLabel = "run.ai/category"
	PriorityLabel = "priorityClassName"

	ControlPlaneManagedLabel = "run.ai/control-plane-managed"

	WorkloadIdAnnotation = "workloadId"

	TrueAsString = "true"
)

// Default values and placeholders
const (
	DefaultCategory      = "WAIT_FOR_DEFAULT_CATEGORY"
	RunaiNamespacePrefix = "runai-"
)

// GVK format strings
const (
	GVKFormatCoreAPI      = "%s/%s"      // version/kind for core APIs
	GVKFormatGroupAPI     = "%s/%s/%s"   // group/version/kind for group APIs
	ResourceKeyFormat     = "%s/%s (%s)" // namespace/name (kind)
	ResourceKeyFormatNoNS = "%s (%s)"    // name (kind) for cluster-scoped resources
)

// API endpoints
const (
	InternalWorkloadsAPIPath = "/internal/v2/workloads"
	NimServicesAPIPath       = "/api/v2/workloads/nim-services"
)

const (
	WorkloadsCategory = "workloads"
)
