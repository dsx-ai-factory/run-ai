/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

const (
	KataNvidiaGpuNodeLabel   = "kata.feature.node.kubernetes.io/nvidia-gpu"
	AmdSevSnpNodeLabel       = "amd.feature.node.kubernetes.io/snp"
	IntelTdxNodeLabel        = "intel.feature.node.kubernetes.io/tdx"
	NvidiaCCReadyStateLabel  = "nvidia.com/cc.ready.state"
	KataRuntimeNodeLabel     = "katacontainers.io/kata-runtime"
	ConfidentialLabelEnabled = "true"
)
