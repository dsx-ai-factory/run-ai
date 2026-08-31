/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

const (
	DefaultWidthName            = 10
	DefaultWidthStatus          = 10
	DefaultWidthGpuType         = 10
	DefaultWidthGpuAllocation   = 20
	DefaultWidthGpuCount        = 10
	DefaultWidthMNNVLDetection  = 15
	DefaultWidthFreeGpu         = 10
	DefaultWidthNodepool        = 10
	DefaultWidthNetworkTopology = 20
)

const (
	noDataFormat = "-"
)

type NodeColumn string

const (
	ColumnKeyName          NodeColumn = "name"
	ColumnKeyStatus        NodeColumn = "status"
	ColumnKeyGpuType       NodeColumn = "gputype"
	ColumnKeyGpuAllocation NodeColumn = "gpuallocation"
	ColumnKeyFreeGpu       NodeColumn = "freegpu"
	ColumnKeyNodepool      NodeColumn = "nodepool"
)
