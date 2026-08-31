/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package model

type InstallPhase int

const (
	NotInstalled InstallPhase = iota
	Installing
	Failed
	Installed
	Timeout
)
