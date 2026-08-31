/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

func BuildSystemJobSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(jobSpecFgFields())
	selectors := BuildItemSelectorMap(jobSpecItemSelectorDefs())
	return fields, selectors
}

func jobSpecItemSelectorDefs() []SystemItemSel {
	return []SystemItemSel{
		{
			Name:  "allPodFailurePolicyRules",
			Desc:  "all rules of the pod failure policy",
			Query: `(.podFailurePolicy.rules // [])[]`,
		},
		{
			Name:     "podFailurePolicyRuleByAction",
			Desc:     "the pod failure policy rule that triggers a given action (FailJob, Ignore, Count, or FailIndex)",
			Query:    `(.podFailurePolicy.rules // [])[] | select(.action == $ARGS.named.action)`,
			ArgsInfo: map[string]api.MappingFieldType{"action": api.MAPPINGFIELDTYPE_STRING},
			ArgsEnum: map[string][]string{"action": {"FailJob", "Ignore", "Count", "FailIndex"}},
		},
		{
			Name:  "allSuccessPolicyRules",
			Desc:  "all rules of the success policy",
			Query: `(.successPolicy.rules // [])[]`,
		},
	}
}

func jobSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	jobSpec := SystemJobSpecFormatterName

	selectorsPodFailurePolicyRules := []string{"allPodFailurePolicyRules", "podFailurePolicyRuleByAction"}
	selectorsSuccessPolicyRules := []string{"allSuccessPolicyRules"}

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// Plain scalars
		////////////////////////////////////////////////////////
		{
			Name:      "parallelism",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "completions",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "manualSelector",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "managedBy",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "suspend",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "activeDeadlineSeconds",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "ttlSecondsAfterFinished",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "backoffLimit",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "backoffLimitPerIndex",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "maxFailedIndexes",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},

		////////////////////////////////////////////////////////
		// Enum scalars (custom formatter for enum enforcement)
		////////////////////////////////////////////////////////
		{
			Name:      "completionMode",
			Formatter: jobSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "podReplacementPolicy",
			Formatter: jobSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},

		////////////////////////////////////////////////////////
		// podFailurePolicy sub-fields
		//
		// The parent `podFailurePolicy` object is not modeled as a settable
		// field — it's an intermediate wrapper similar to `selector`.
		// Policies add rules to .podFailurePolicy.rules via the
		// `podFailurePolicyRules` custom formatter.
		////////////////////////////////////////////////////////
		{
			Name:      "podFailurePolicyRules",
			Formatter: jobSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "podFailurePolicyRules-onExitCodes",
			Formatter:     jobSpec,
			SpecRef:       convutils.Ref("onExitCodes"),
			Type:          api.MAPPINGFIELDTYPE_OBJECT,
			ItemSelectors: selectorsPodFailurePolicyRules,
		},
		{
			Name:          "podFailurePolicyRules-onPodConditions",
			Formatter:     jobSpec,
			SpecRef:       convutils.Ref("onPodConditions"),
			Type:          api.MAPPINGFIELDTYPE_OBJECT,
			ItemSelectors: selectorsPodFailurePolicyRules,
		},

		////////////////////////////////////////////////////////
		// successPolicy sub-fields
		//
		// The parent `successPolicy` object is not modeled as a settable
		// field — it's an intermediate wrapper similar to `selector`.
		// Policies add rules to .successPolicy.rules via the
		// `successPolicyRules` custom formatter.
		////////////////////////////////////////////////////////
		{
			Name:      "successPolicyRules",
			Formatter: jobSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "successPolicyRules-SucceededCount",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("succeededCount"),
			Type:          api.MAPPINGFIELDTYPE_INTEGER,
			ItemSelectors: selectorsSuccessPolicyRules,
		},
		{
			Name:          "successPolicyRules-SucceededIndexes",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("succeededIndexes"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsSuccessPolicyRules,
		},
	}
}
