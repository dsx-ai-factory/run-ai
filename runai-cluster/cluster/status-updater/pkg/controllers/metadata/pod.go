/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metadata

type PodMetadata struct {
	Status         string
	GPUs           float64
	GpuMemory      int64
	FractionsCount int64
	MigGPUs        float64
}
