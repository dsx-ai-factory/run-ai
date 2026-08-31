/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import "maps"

type NodePoolControllerParams struct {
	// SchedulingShardArgs are cluster-wide KAI scheduler args merged into every
	// SchedulingShard. Loaded from the --scheduling-shard-args flag; empty means the
	// scheduling shard will use its own KAI-scheduler defaults.
	SchedulingShardArgs map[string]string
}

func (in *NodePoolControllerParams) DeepCopy() *NodePoolControllerParams {
	if in == nil {
		return nil
	}
	out := new(NodePoolControllerParams)
	out.SchedulingShardArgs = maps.Clone(in.SchedulingShardArgs)
	return out
}
