/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_groups

import (
	"context"

	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
)

// +k8s:deepcopy-gen=true
type DistributedInferenceSchedulerFields struct {
	// Specifies the list of node pools to use for scheduling the job, ordered by preference.
	NodePools *flds.StringField `json:"nodePools,omitempty"`
}

func DistributedInferenceSchedulerRunaiRules(ctx context.Context, nodePoolLister NodePoolNameLister) DistributedInferenceSchedulerFields {
	var result DistributedInferenceSchedulerFields

	// the rules will be calculated from NodePool field;
	// but we need it here just so the existence of this field is known when asked.
	result.NodePools = getNodePoolRules(ctx, nodePoolLister)

	return result
}
