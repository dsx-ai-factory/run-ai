/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gvk_compliance

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/itchyny/gojq"
)

const (
	KeyValueBuiltin    = "KeyValue"
	KeyValueMapBuiltin = "KeyValueMap"
)

type RuntimeFormatter struct {
	GetExpr *gojq.Code
	SetExpr *gojq.Code

	// GetPathExpr is the compiled path() form of the get expression. It yields
	// the JQ path to each value GetExpr produces (in lockstep) and is used by
	// validation enforcements to report workload-relative locations in error
	// messages. Nil when the get expression isn't path-eligible (e.g. uses
	// the alternative operator //).
	GetPathExpr *gojq.Code

	// this is a workaround of the KeyValueMap builtin formatter that needs to copy the
	// target key to the source key before setting the value
	CopyTargetKeys bool

	// SpecPath is the dot-split spec ref path used by the KeyValueMap builtin
	// when the field has a specRef. The compiled set expression nests its
	// result under that path (e.g., specRef "labels" produces
	// `{ "labels": { ($ARGS.named.key // "value"): $ARGS.named.value } }`),
	// so CopyTargetKeys must walk both srcValue and the target the same way
	// before fanning the placeholder across existing keys. Empty when no
	// navigation is needed: KeyValue, KeyValueMap without specRef, or
	// non-builtin formatters.
	SpecPath []string

	// ArrayKey identifies an element in an array by the value of a specific field
	// (e.g. the "name" field of a container). It is used by the set enforcement to merge
	// array elements instead of replacing the whole array.
	ArrayKey *RuntimeArrayKey
}

// RuntimeArrayKey identifies an element in an array for set-enforcement merging.
type RuntimeArrayKey struct {
	//	Arg is the name of the argument in $ARGS.named that carries the key value
	//  (e.g. "name" for containers keyed by their name).
	Arg string

	//	SpecField is the dotted path within each array element that holds the key value.
	//	          When empty, falls back to Arg.
	SpecField string
}

// NewRuntimeFormatter builds a RuntimeFormatter for the given field.
// If rule.variation.Formatter.Builtin is set, the get/set expressions are composed from a spec path
// (rule.variation.SpecRef, defaulting to fieldName). Otherwise, formatter.GetExpr/SetExpr are used as-is.
// Mapping/formatter issues (missing formatter, unknown builtin, missing expr, invalid JQ) are logged
// via LogBrokenRule and nil is returned; they are not propagated as errors.
func NewRuntimeFormatter(fieldName string, rule *ResolvedRule) *RuntimeFormatter {
	variation := rule.variation
	if variation == nil || variation.Formatter == nil {
		LogBrokenRule(rule.rule, "field %s has no formatter", fieldName)
		return nil
	}

	formatter := variation.Formatter

	var getExprStr, setExprStr *string
	copyTargetKeys := false
	var kvMapSpecPath []string

	if convutils.IsStringHasValue(formatter.Builtin.Get()) {
		specPath := fieldName
		hasSpecRef := convutils.IsStringHasValue(variation.SpecRef)
		if hasSpecRef {
			specPath = *variation.SpecRef
		}

		builtin := *formatter.Builtin.Get()
		var get, set string
		switch builtin {
		case KeyValueBuiltin:
			get, set = keyValueExpressions(specPath)
		case KeyValueMapBuiltin:
			get, set = keyValueMapExpressions(specPath, hasSpecRef)
			copyTargetKeys = true
			// Only when hasSpecRef does the set expression nest its result
			// under specPath; without it the placeholder lives at srcValue
			// root and no navigation is needed.
			if hasSpecRef {
				kvMapSpecPath = strings.Split(specPath, ".")
			}
		default:
			LogBrokenRule(rule.rule, "field %s has unknown builtin formatter %q", fieldName, builtin)
			return nil
		}
		getExprStr = &get
		setExprStr = &set
	} else {
		getExprStr = formatter.GetExpr.Get()
		setExprStr = formatter.SetExpr.Get()
		if !convutils.IsStringHasValue(getExprStr) && !convutils.IsStringHasValue(setExprStr) {
			LogBrokenRule(rule.rule, "field %s formatter has neither get nor set expression", fieldName)
			return nil
		}
	}

	result := &RuntimeFormatter{
		CopyTargetKeys: copyTargetKeys,
		SpecPath:       kvMapSpecPath,
	}

	if arrayKey := formatter.ArrayKey.Get(); arrayKey != nil {
		specField := ""
		if convutils.IsStringHasValue(arrayKey.SpecRef.Get()) {
			specField = *arrayKey.SpecRef.Get()
		}
		result.ArrayKey = &RuntimeArrayKey{
			Arg:       arrayKey.ArgName,
			SpecField: specField,
		}
	}

	if convutils.IsStringHasValue(getExprStr) {
		getCode, apiErr := compileJQNoArgs(*getExprStr)
		if apiErr != nil {
			LogBrokenRule(rule.rule, "field %s has invalid get expression: %s", fieldName, apiErr.Message)
		} else {
			result.GetExpr = getCode
			result.GetPathExpr = compilePathExpr(*getExprStr, compileJQNoArgs)
		}
	}

	if convutils.IsStringHasValue(setExprStr) {
		setCode, apiErr := compileJQWithArgs(*setExprStr)
		if apiErr != nil {
			LogBrokenRule(rule.rule, "field %s has invalid set expression: %s", fieldName, apiErr.Message)
		} else {
			result.SetExpr = setCode
		}
	}

	return result
}

// keyValueExpressions composes get/set expressions for the KeyValue builtin formatter.
// Keys are quoted / bracketed so the expressions compile for spec path segments
// that are not bare JQ identifiers (e.g. containing hyphens).
// For specPath "spec.replicas":
//
//	get: `.["spec"]["replicas"]`
//	set: `{ "spec": { "replicas": $ARGS.named.value } }`
func keyValueExpressions(specPath string) (string, string) {
	parts := strings.Split(specPath, ".")

	getParts := make([]string, 0, len(parts))
	for _, part := range parts {
		getParts = append(getParts, fmt.Sprintf(`["%s"]`, part))
	}
	getExpr := "." + strings.Join(getParts, "")

	var startSet, endSet strings.Builder

	for i := 0; i < len(parts)-1; i++ {
		startSet.WriteString(fmt.Sprintf(`{ "%s": `, parts[i]))
		endSet.WriteString(" }")
	}

	setExpr := fmt.Sprintf(`%s{ "%s": $ARGS.named.value }%s`,
		startSet.String(), parts[len(parts)-1], endSet.String())

	return getExpr, setExpr
}

// keyValueMapExpressions returns the get/set expressions for the KeyValueMap
// builtin formatter. KeyValueMap intentionally has no get expression — validation
// of map values requires a separate field with an itemSelector that iterates map
// entries. When hasSpecRef is true, specPath navigates to the target map (e.g.
// "selector.matchLabels") and the set expression wraps the map access. When
// hasSpecRef is false, the set expression operates on the input directly (assuming
// the spec/item selector already narrowed to the target map).
// The set expression uses ($ARGS.named.key // "value") so that if a "key" arg
// is provided it sets that specific key, otherwise it emits a "value"
// placeholder that SetEnforcement.Exec's CopyTargetKeys fans across every existing
// key of the target map.
func keyValueMapExpressions(specPath string, hasSpecRef bool) (string, string) {
	if !hasSpecRef {
		// Original flat expressions — spec/item selector already emits the target map.
		setExpr := `{ ($ARGS.named.key // "value"): $ARGS.named.value }`
		return "", setExpr
	}

	// specRef provided — wrap set expression with specPath navigation.
	parts := strings.Split(specPath, ".")

	var startSet, endSet strings.Builder
	for i := 0; i < len(parts); i++ {
		startSet.WriteString(fmt.Sprintf(`{ "%s": `, parts[i]))
		endSet.WriteString(" }")
	}
	setExpr := startSet.String() + `{ ($ARGS.named.key // "value"): $ARGS.named.value }` + endSet.String()

	return "", setExpr
}
