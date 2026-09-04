/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// BuildSystemMetadataFormatterFields builds the runai-metadata system formatter,
// which holds GET-only companion fields for the metadata field group's
// KeyValueMap entries (labels / annotations).
//
// Why these exist:
//
//	The KeyValueMap builtin formatter intentionally has no get expression — it
//	only knows how to write entries into a map at a SpecRef-narrowed path.
//	To validate or read the map's values, policies use the per-key companion
//	fields below: each one has a custom GetExpr and is paired with an item
//	selector that does the navigation work, so the formatter just returns the
//	already-resolved string value.
//
// The companions use GetExpr "." because the item selector reaches the final
// scalar (a single string keyed by the supplied label/annotation name), so
// no further extraction is needed by the formatter.
func BuildSystemMetadataFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	identityGet := "."

	addGetterField(m, "labels-value", identityGet)
	addGetterField(m, "annotations-value", identityGet)

	return m
}

// addGetterField registers a custom-formatter entry whose only purpose is to
// return whatever the spec/item selector emits. SetExpr is intentionally nil:
// these companions are GET-only — set enforcements on the parent KeyValueMap
// field (e.g. labels) are how the map gets written.
func addGetterField(dst map[string]api.FormatterField, name, getExpr string) {
	dst[name] = api.FormatterField{
		GetExpr: convutils.Ref(getExpr),
	}
}
