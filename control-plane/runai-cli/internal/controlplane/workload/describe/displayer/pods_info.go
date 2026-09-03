/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

type PodsInfo struct {
	Pods        []Pod `json:"pods"`
	RunningPods int32 `json:"runningPods"`
}
