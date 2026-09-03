/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// BuildSystemDeploymentSpecFg builds the system-scoped "deploymentSpec" field group,
// covering the deployment-level fields above the pod template (everything in
// apps/Deployment/v1 .spec other than .spec.selector and .spec.template).
//
// Field paths and item-selector queries are rooted at the workload .spec.
func BuildSystemDeploymentSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(deploymentSpecFgFields())
	selectors := BuildItemSelectorMap(deploymentSpecItemSelectorDefs())
	return fields, selectors
}

func deploymentSpecItemSelectorDefs() []SystemItemSel {
	// No item selectors needed: every deployment-level field in this group is either a
	// scalar or a singleton object, so there are no arrays/maps to iterate or filter.
	return nil
}

func deploymentSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	deploymentSpec := SystemDeploymentSpecFormatterName

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// Plain scalars
		////////////////////////////////////////////////////////
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
		{
			Name:      "revisionHistoryLimit",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "progressDeadlineSeconds",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "paused",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},

		////////////////////////////////////////////////////////
		// strategy sub-fields
		//
		// The parent `strategy` object is not modeled as a settable field —
		// it's an intermediate wrapper similar to `selector`. Policies set
		// the meaningful children (type, maxSurge, maxUnavailable) directly.
		//
		// strategy.rollingUpdate.maxSurge / maxUnavailable are IntOrString in the
		// Kubernetes API. We expose them as STRING so policies can express both forms
		// uniformly: an integer count ("5") or a percentage ("25%"). The K8s API server
		// accepts either form when the value is JSON-encoded as a string.
		////////////////////////////////////////////////////////
		{
			Name:      "strategy-Type",
			Formatter: deploymentSpec,
			SpecRef:   convutils.Ref("strategy.type"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "strategy-MaxSurge",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("strategy.rollingUpdate.maxSurge"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "strategy-MaxUnavailable",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("strategy.rollingUpdate.maxUnavailable"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}
}
