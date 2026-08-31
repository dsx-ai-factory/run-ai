/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flow

type InstallFlowState int

const (
	InstallFlowStateNotStarted InstallFlowState = iota
	InstallFlowStateInProgress
	InstallFlowStateCompleted
	InstallFlowStateFailed
)
