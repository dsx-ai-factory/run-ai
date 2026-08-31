/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/itchyny/gojq"
	kartajq "github.com/run-ai/karta/pkg/jq"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
)

// validateParsedJQSafety rejects JQ expressions that are dangerous or may produce
// excessive output. Delegates to karta's exported ValidateParsedJQ.
func validateParsedJQSafety(q *gojq.Query) error {
	return kartajq.ValidateParsedJQ(q)
}

// CollectArgsNamedReferences walks a parsed gojq query tree and records every static
// $ARGS.named.<identifier> reference into out. Keys are the <identifier> segment only
// (e.g. containerPort for $ARGS.named.containerPort).
//
// out must be non-nil. The same map can be reused across multiple expressions to merge names.
//
// Dynamic keys (e.g. $ARGS.named[$var] or $ARGS.named["x"] with string interpolation) are skipped.
func CollectArgsNamedReferences(q *gojq.Query, out datastructs.Set[string]) {
	if q == nil || out == nil {
		return
	}

	for _, fd := range q.FuncDefs {
		CollectArgsNamedReferences(fd.Body, out)
	}

	if q.Term != nil {
		walkTermForArgsNamed(q.Term, out)
	}

	CollectArgsNamedReferences(q.Left, out)
	CollectArgsNamedReferences(q.Right, out)
}

func walkTermForArgsNamed(t *gojq.Term, out datastructs.Set[string]) {
	if t == nil {
		return
	}
	if t.Type == gojq.TermTypeFunc && t.Func != nil && t.Func.Name == "$ARGS" {
		collectArgsNamedFromArgsSuffixes(t.SuffixList, out)
	}
	if t.Func != nil {
		for _, arg := range t.Func.Args {
			CollectArgsNamedReferences(arg, out)
		}
	}
	if t.Unary != nil && t.Unary.Term != nil {
		walkTermForArgsNamed(t.Unary.Term, out)
	}
	if t.Query != nil {
		CollectArgsNamedReferences(t.Query, out)
	}
	if t.Object != nil {
		for _, kv := range t.Object.KeyVals {
			if kv.KeyQuery != nil {
				CollectArgsNamedReferences(kv.KeyQuery, out)
			}
			if kv.Val != nil {
				CollectArgsNamedReferences(kv.Val, out)
			}
		}
	}
	if t.Array != nil && t.Array.Query != nil {
		CollectArgsNamedReferences(t.Array.Query, out)
	}
	if t.If != nil {
		CollectArgsNamedReferences(t.If.Cond, out)
		CollectArgsNamedReferences(t.If.Then, out)
		for _, elif := range t.If.Elif {
			CollectArgsNamedReferences(elif.Cond, out)
			CollectArgsNamedReferences(elif.Then, out)
		}
		CollectArgsNamedReferences(t.If.Else, out)
	}
	if t.Try != nil {
		CollectArgsNamedReferences(t.Try.Body, out)
		CollectArgsNamedReferences(t.Try.Catch, out)
	}
	if t.Reduce != nil {
		CollectArgsNamedReferences(t.Reduce.Query, out)
		CollectArgsNamedReferences(t.Reduce.Start, out)
		CollectArgsNamedReferences(t.Reduce.Update, out)
	}
	if t.Foreach != nil {
		CollectArgsNamedReferences(t.Foreach.Query, out)
		CollectArgsNamedReferences(t.Foreach.Start, out)
		CollectArgsNamedReferences(t.Foreach.Update, out)
		CollectArgsNamedReferences(t.Foreach.Extract, out)
	}
	if t.Label != nil {
		CollectArgsNamedReferences(t.Label.Body, out)
	}
	if t.Str != nil {
		for _, eq := range t.Str.Queries {
			CollectArgsNamedReferences(eq, out)
		}
	}
	for _, suf := range t.SuffixList {
		if suf.Index == nil {
			continue
		}
		if suf.Index.Start != nil {
			CollectArgsNamedReferences(suf.Index.Start, out)
		}
		if suf.Index.End != nil {
			CollectArgsNamedReferences(suf.Index.End, out)
		}
	}
}

func collectArgsNamedFromArgsSuffixes(suffixes []*gojq.Suffix, out datastructs.Set[string]) {
	if len(suffixes) < 2 {
		return
	}
	if !indexIsBareIdent(suffixes[0].Index, "named") {
		return
	}
	if name := staticSingleKeyFromIndex(suffixes[1].Index); name != "" {
		out.Add(name)
	}
}

func indexIsBareIdent(idx *gojq.Index, want string) bool {
	if idx == nil {
		return false
	}
	return idx.Name == want && idx.Str == nil && idx.Start == nil && idx.End == nil && !idx.IsSlice
}

// staticSingleKeyFromIndex returns a key for .ident or ["literal"] without interpolation.
func staticSingleKeyFromIndex(idx *gojq.Index) string {
	if idx == nil {
		return ""
	}
	if idx.Name != "" && idx.Str == nil && idx.Start == nil && idx.End == nil && !idx.IsSlice {
		return idx.Name
	}
	if idx.Str != nil && len(idx.Str.Queries) == 0 {
		return idx.Str.Str
	}
	return ""
}

// EnrichArgsInfoFromExpressions parses the given JQ expressions, collects all
// $ARGS.named.<name> references, and adds any missing arguments to argsInfo
// with type STRING. This ensures all referenced arguments are documented in
// argsInfo for readability, even if they don't need explicit type information.
// Returns the (possibly modified) argsInfo map.
func EnrichArgsInfoFromExpressions(argsInfo map[string]api.MappingFieldType, expressions ...string) map[string]api.MappingFieldType {
	if argsInfo == nil {
		argsInfo = make(map[string]api.MappingFieldType)
	}

	referencedArgs := datastructs.NewSet[string]()
	for _, expr := range expressions {
		if expr == "" {
			continue
		}
		if q, err := gojq.Parse(expr); err == nil {
			CollectArgsNamedReferences(q, referencedArgs)
		}
	}

	for _, argName := range referencedArgs.Keys() {
		if _, exists := argsInfo[argName]; !exists {
			argsInfo[argName] = api.MAPPINGFIELDTYPE_STRING
		}
	}

	return argsInfo
}

// EnrichFormatterFieldArgsInfo enriches the ArgsInfo of a FormatterField with
// any $ARGS.named references from its SetExpr and GetExpr that are not already
// present. Missing arguments are added with type STRING.
func EnrichFormatterFieldArgsInfo(field *api.FormatterField) {
	if field == nil {
		return
	}

	var setExpr, getExpr string
	if field.SetExpr != nil {
		setExpr = *field.SetExpr
	}
	if field.GetExpr != nil {
		getExpr = *field.GetExpr
	}

	field.ArgsInfo = EnrichArgsInfoFromExpressions(field.ArgsInfo, setExpr, getExpr)
}

// EnrichSpecSelectorArgsInfo enriches the ArgsInfo of a SpecSelector with
// any $ARGS.named references from its Query that are not already present.
// Missing arguments are added with type STRING.
func EnrichSpecSelectorArgsInfo(sel *api.SpecSelector) {
	if sel == nil || sel.Query == "" {
		return
	}

	sel.ArgsInfo = EnrichArgsInfoFromExpressions(sel.ArgsInfo, sel.Query)
}

// EnrichItemSelectorArgsInfo enriches the ArgsInfo of an ItemSelector with
// any $ARGS.named references from its Query that are not already present.
// Missing arguments are added with type STRING.
func EnrichItemSelectorArgsInfo(sel *api.ItemSelector) {
	if sel == nil || sel.Query == "" {
		return
	}

	sel.ArgsInfo = EnrichArgsInfoFromExpressions(sel.ArgsInfo, sel.Query)
}
