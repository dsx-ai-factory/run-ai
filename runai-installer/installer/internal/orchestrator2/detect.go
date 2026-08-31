/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator2

type DetectPhase string

const (
	DetectNotInstalled DetectPhase = "not_installed"
	DetectInstalled    DetectPhase = "installed"
)

type DetectResult struct {
	Phase   DetectPhase
	Version string
}
