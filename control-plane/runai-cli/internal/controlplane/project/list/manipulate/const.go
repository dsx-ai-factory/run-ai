/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

const (
	DefaultWidthName            = 10
	DefaultWidthPhase           = 10
	DefaultWidthDepartment      = 15
	DefaultWidthGPUQuota        = 10
	DefaultWidthAllocatedGPUs   = 15
	DefaultWidthAllocationRatio = 20
)

const (
	nameColumnIndex = 0
	noDataFormat    = "-"
)

type ProjectColumn string

const (
	ColumnKeyName            ProjectColumn = "name"
	ColumnKeyStatus          ProjectColumn = "status"
	ColumnKeyDepartment      ProjectColumn = "department"
	ColumnKeyGPUQuota        ProjectColumn = "gpuquota"
	ColumnKeyAllocatedGPUs   ProjectColumn = "allocatedgpus"
	ColumnKeyAllocationRatio ProjectColumn = "allocationratio"
)
