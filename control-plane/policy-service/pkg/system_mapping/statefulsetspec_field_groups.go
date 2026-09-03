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

// BuildSystemStatefulSetSpecFg builds the system-scoped "statefulSetSpec" field
// group, covering the stateful-set-level fields above the pod template
// (everything in apps/StatefulSet/v1 .spec other than .spec.selector and
// .spec.template).
//
// Field paths and item-selector queries are rooted at the workload .spec.
//
// updateStrategy.rollingUpdate.maxUnavailable is IntOrString in the K8s API.
// We expose it as STRING so policies can express both forms uniformly: an
// integer count ("5") or a percentage ("25%"). The K8s API server accepts
// either form when JSON-encoded as a string. (Same convention as the
// Deployment field group's strategy-MaxSurge / strategy-MaxUnavailable.)
//
// volumeClaimTemplates is an array of full PersistentVolumeClaim objects.
// We expose:
//   - a `volumeClaimTemplates` rule-builder field (custom formatter) that
//     creates or upserts one PVC entry, merged into the array by metadata.name
//   - per-attribute leaf fields (`-AccessModes`, `-StorageClassName`,
//     `-VolumeMode`, `-StorageRequest`) that combine with the
//     `allVolumeClaimTemplates` / `volumeClaimTemplateByName` item selectors
//     to express constraints over existing templates.
//
// The full PVC schema is intentionally not modeled — these fields cover the
// commonly-policed attributes (capacity, access modes, storage class, volume
// mode); other fields (selector, dataSource, volumeName, storage limits) can
// be added incrementally on the same pattern.
//
// persistentVolumeClaimRetentionPolicy.whenDeleted / .whenScaled enums are
// not declared in the K8s OpenAPI v3 schema (no `enum` field at the type
// level). We still enforce them via runai-statefulSetSpec argsEnum so
// policies cannot accidentally write nonsense values that K8s admission
// would later reject.
func BuildSystemStatefulSetSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(statefulSetSpecFgFields())
	selectors := BuildItemSelectorMap(statefulSetSpecItemSelectorDefs())
	return fields, selectors
}

func statefulSetSpecItemSelectorDefs() []SystemItemSel {
	return []SystemItemSel{
		{
			Name:  "allVolumeClaimTemplates",
			Desc:  "all volume claim templates of the stateful set",
			Query: `(.volumeClaimTemplates // [])[]`,
		},
		{
			Name:     "volumeClaimTemplateByName",
			Desc:     "the volume claim template with a given name",
			Query:    `(.volumeClaimTemplates // [])[] | select(.metadata.name == $ARGS.named.templateName)`,
			ArgsInfo: map[string]api.MappingFieldType{"templateName": api.MAPPINGFIELDTYPE_STRING},
		},
	}
}

func statefulSetSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	statefulSetSpec := SystemStatefulSetSpecFormatterName

	selectorsVolumeClaimTemplates := []string{"allVolumeClaimTemplates", "volumeClaimTemplateByName"}

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
			Name:      "serviceName",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},

		////////////////////////////////////////////////////////
		// Enum scalar (custom formatter for enum enforcement)
		////////////////////////////////////////////////////////
		{
			Name:      "podManagementPolicy",
			Formatter: statefulSetSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},

		////////////////////////////////////////////////////////
		// updateStrategy sub-fields
		//
		// The parent `updateStrategy` object is not modeled as a settable
		// field — it's an intermediate wrapper similar to `selector`.
		// Policies set the meaningful children (type, partition,
		// maxUnavailable) directly.
		////////////////////////////////////////////////////////
		{
			Name:      "updateStrategy-Type",
			Formatter: statefulSetSpec,
			SpecRef:   convutils.Ref("updateStrategy.type"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "updateStrategy-Partition",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("updateStrategy.rollingUpdate.partition"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "updateStrategy-MaxUnavailable",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("updateStrategy.rollingUpdate.maxUnavailable"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},

		////////////////////////////////////////////////////////
		// volumeClaimTemplates
		////////////////////////////////////////////////////////
		{
			Name:      "volumeClaimTemplates",
			Formatter: statefulSetSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "volumeClaimTemplates-AccessModes",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("spec.accessModes"),
			Type:          api.MAPPINGFIELDTYPE_STRINGS,
			ItemSelectors: selectorsVolumeClaimTemplates,
		},
		{
			Name:          "volumeClaimTemplates-StorageClassName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("spec.storageClassName"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumeClaimTemplates,
		},
		{
			Name:          "volumeClaimTemplates-VolumeMode",
			Formatter:     statefulSetSpec,
			SpecRef:       convutils.Ref("spec.volumeMode"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumeClaimTemplates,
		},
		{
			Name:          "volumeClaimTemplates-StorageRequest",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("spec.resources.requests.storage"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsVolumeClaimTemplates,
		},

		////////////////////////////////////////////////////////
		// persistentVolumeClaimRetentionPolicy sub-fields
		//
		// The parent `persistentVolumeClaimRetentionPolicy` object is not
		// modeled as a settable field — it's an intermediate wrapper similar
		// to `selector`. Policies set the meaningful children (whenDeleted,
		// whenScaled) directly.
		////////////////////////////////////////////////////////
		{
			Name:      "persistentVolumeClaimRetentionPolicy-WhenDeleted",
			Formatter: statefulSetSpec,
			SpecRef:   convutils.Ref("persistentVolumeClaimRetentionPolicy.whenDeleted"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "persistentVolumeClaimRetentionPolicy-WhenScaled",
			Formatter: statefulSetSpec,
			SpecRef:   convutils.Ref("persistentVolumeClaimRetentionPolicy.whenScaled"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},

		////////////////////////////////////////////////////////
		// ordinals
		////////////////////////////////////////////////////////
		{
			Name:      "ordinals-Start",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("ordinals.start"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
	}
}
