/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type ResourcesPerNodePool struct {
	GpuDeserved           *float64
	GpuLimit              *float64
	GpuOverQuotaWeight    *float64
	CpuDeserved           *float64
	CpuLimit              *float64
	CpuOverQuotaWeight    *float64
	MemoryDeserved        *float64
	MemoryLimit           *float64
	MemoryOverQuotaWeight *float64
	MemoryUnits           Units
}
