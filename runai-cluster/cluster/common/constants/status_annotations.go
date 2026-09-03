/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

const (
	PodStatusAnnotation      = "runai-calculated-status"
	PodGroupStatusAnnotation = PodStatusAnnotation

	PodAllocatedGpusAnnotation      = "runai-allocated-gpus"
	PodAllocatedGpuMemoryAnnotation = "runai-allocated-gpu-memory"
	PodAllocatedMigGpusAnnotation   = "runai-allocated-mig-gpus"
)
