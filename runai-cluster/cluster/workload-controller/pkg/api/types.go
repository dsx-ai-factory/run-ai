/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

type SubresourceComponent string

const (
	SpecSubResource       SubresourceComponent = "spec"
	WorkerSpecSubResource SubresourceComponent = "workerSpec"
	MasterSpecSubResource SubresourceComponent = "masterSpec"
	LeaderSpecSubResource SubresourceComponent = "leaderSpec"
)
