/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
)

// BuildSystemReplicaSetSpecFg builds the system-scoped "replicaSetSpec" field group,
// covering the replica-set-level fields above the pod template (everything in
// apps/ReplicaSet/v1 .spec other than .spec.selector and .spec.template).
//
// ReplicaSet's spec is intentionally minimal: it carries only the rollout knobs
// the controller needs to keep N pods alive. There are no enum-bearing fields,
// hence no custom formatter — every field uses KeyValue.
//
// Field paths and item-selector queries are rooted at the workload .spec.
func BuildSystemReplicaSetSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(replicaSetSpecFgFields())
	selectors := BuildItemSelectorMap(replicaSetSpecItemSelectorDefs())
	return fields, selectors
}

func replicaSetSpecItemSelectorDefs() []SystemItemSel {
	// No item selectors needed: every replica-set-level field is a scalar, so
	// there are no arrays/maps to iterate or filter.
	return nil
}

func replicaSetSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue

	return []SystemFgField{
		{
			Name:      "replicas",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "minReadySeconds",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
	}
}
