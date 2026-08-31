/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy

import fields "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"

func IntFieldMinMax(min *int, max *int) *fields.IntField {
	return &fields.IntField{
		Rules: &fields.IntRules{
			Min: min,
			Max: max,
		},
	}
}
