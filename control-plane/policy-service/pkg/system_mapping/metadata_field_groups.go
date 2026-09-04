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

// BuildSystemMetadataFg builds the system-scoped "metadata" field group for
// metadata (annotations and labels). The spec selector that binds this field
// group points at a metadata-bearing object (e.g. .metadata for workload-level,
// .spec.template.metadata for pod-template-level). The SpecRefs below
// ("labels", "annotations") then narrow the KeyValueMap target to the right
// nested map under that anchor.
//
// Read companions:
//
//	The KeyValueMap builtin formatter on the parent fields (labels,
//	annotations) only knows how to WRITE entries into the map at the
//	SpecRef-narrowed path. To READ the values for validation, policies use
//	the per-key companions:
//	  - labels-value / annotations-value: single value, by key (uses
//	    labelsByKey / annotationsByKey item selectors).
//	The companions use the runai-metadata custom formatter (GetExpr "."),
//	since the item selector already resolves to the final scalar.
func BuildSystemMetadataFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(metadataFgFields())
	selectors := BuildItemSelectorMap(metadataItemSelectorDefs())
	return fields, selectors
}

func metadataItemSelectorDefs() []SystemItemSel {
	return []SystemItemSel{
		{
			Name:     "labelsByKey",
			Desc:     "label value with a given key",
			Query:    `.labels[$ARGS.named.key]`,
			ArgsInfo: map[string]api.MappingFieldType{"key": api.MAPPINGFIELDTYPE_STRING},
		},
		{
			Name:     "annotationsByKey",
			Desc:     "annotation value with a given key",
			Query:    `.annotations[$ARGS.named.key]`,
			ArgsInfo: map[string]api.MappingFieldType{"key": api.MAPPINGFIELDTYPE_STRING},
		},
	}
}

func metadataFgFields() []SystemFgField {
	keyValueMap := dbmodel.BuiltinFormatterKeyValueMap
	metadataFmt := SystemMetadataFormatterName

	selectorsLabelsByKey := []string{"labelsByKey"}
	selectorsAnnotationsByKey := []string{"annotationsByKey"}

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// labels (KeyValueMap, set-only) + read companions
		////////////////////////////////////////////////////////
		{
			Name:      "labels",
			Formatter: keyValueMap,
			SpecRef:   convutils.Ref("labels"),
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "labels-value",
			Formatter:     metadataFmt,
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsLabelsByKey,
		},

		////////////////////////////////////////////////////////
		// annotations (KeyValueMap, set-only) + read companions
		////////////////////////////////////////////////////////
		{
			Name:      "annotations",
			Formatter: keyValueMap,
			SpecRef:   convutils.Ref("annotations"),
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "annotations-value",
			Formatter:     metadataFmt,
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsAnnotationsByKey,
		},
	}
}
