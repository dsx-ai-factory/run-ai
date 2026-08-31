/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
)

// copyIfTrue copies constants.LabelNIMWorkload with value "true" from the getter into 'to'.
// The getter returns (value, exists).
func copyIfTrue(to map[string]string, get func() (string, bool)) {
	if to == nil || get == nil {
		return
	}
	if nimVal, exists := get(); exists && nimVal == "true" {
		to[constants.LabelNIMWorkload] = "true"
	}
}

func CopyNimLabel(fromDict *flds.DictionaryField, to map[string]string) {
	if fromDict != nil {
		copyIfTrue(to, func() (string, bool) {
			item, ok := fromDict.Deref()[constants.LabelNIMWorkload]
			if !ok || item.Value == nil {
				return "", false
			}
			return *item.Value, true
		})
		return
	}
}
