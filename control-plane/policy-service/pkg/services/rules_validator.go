/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/itchyny/gojq"
	"k8s.io/apimachinery/pkg/api/resource"
)

// validateRules walks the policy and accumulates structural problems into the
// caller-provided collector. The collector is returned for inspection
// convenience (e.g. tests that chain `.toError()`); callers that supply a
// shared collector with the breakdown generator just keep using their own
// reference.
//
// Set issues.maxErrors > 0 to short-circuit after that many findings — the
// equivalent of the runtime's "log and continue, but cap how many warnings
// we keep" pattern. A maxErrors of 0 means "collect everything".
func (pa *PolicyApiService) validateRules(_ context.Context, breakdown *api.BreakdownBreakdown, rules []api.PolicyRule, issues *ruleIssues) *ruleIssues {
	if breakdown == nil {
		return issues
	}

	// a policy cannot have duplicate rules for the same (field+selectors) combo. keep the rule we had
	// so far in this map to validate no duplicate rules are specified
	rulesSeen := make(map[string]int)

	for i, rule := range rules {
		if issues.full() {
			break
		}
		validateOneRule(issues, i, rule, breakdown, rulesSeen)
	}

	return issues
}

// validateOneRule runs every per-rule check, appending each issue to the
// shared collector instead of returning on the first failure. Mirrors the
// runtime's gvk_compliance.LogBrokenRule pattern: a finding does not abort
// the rest of the work, so a single call to validateRules surfaces every
// problem the policy has at once.
//
// Structural prerequisites that prevent further checks from being meaningful
// (field/variation/specSelector existence) cause an early return for *this*
// rule only — the outer loop keeps validating the remaining rules.
func validateOneRule(issues *ruleIssues, idx int, rule api.PolicyRule, breakdown *api.BreakdownBreakdown, rulesSeen map[string]int) {
	prefix := fmt.Sprintf("rule[%d] (field=%s, specSelector=%s)", idx, rule.Field, rule.SpecSelector)

	ref := PolicyRuleRefFromRule(rule)
	key := dbmodel.PolicyRuleRefKey(ref)
	if prevIdx, dup := rulesSeen[key]; dup {
		issues.addf("%s: duplicate rule, same fields/selectors combination already defined at rule[%d]", prefix, prevIdx)
		// keep the original first-seen index in the map so that any further
		// duplicates of the same key are still reported relative to the
		// original rule rather than relative to this duplicate.
	} else {
		rulesSeen[key] = idx
	}

	// look for the variation of the field that takes part in this rule
	variations, ok := breakdown.Fields[rule.Field]
	if !ok {
		issues.addf("%s: field %q not found under %s", prefix, rule.Field, ref.SpecSelector)
		return
	}

	variation := findVariationForRule(rule, variations)
	if variation == nil {
		issues.addWithDetails(
			fmt.Sprintf("%s: field %q not found under specSelector %q", prefix, rule.Field, rule.SpecSelector),
			fmt.Sprintf("valid specSelectors for this field: %v", collectSpecSelectors(variations)),
		)
		return
	}

	specSel, hasSel := breakdown.SpecSelectors[rule.SpecSelector]
	if !hasSel {
		issues.addWithDetails(
			fmt.Sprintf("%s: specSelector %q not found in the mapping", prefix, rule.SpecSelector),
			fmt.Sprintf("valid specSelectors: %v", sortedKeys(breakdown.SpecSelectors)),
		)
		return
	}

	// the runtime resolver (gvk_compliance.ResolveRules) drops the rule when
	// the spec selector has no query; reject it explicitly here so the policy
	// is rejected at create/update time instead of going silent at apply time.
	specQueryEmpty := strings.TrimSpace(specSel.Query) == ""
	if specQueryEmpty {
		issues.addf("%s: specSelector %q has an empty query", prefix, rule.SpecSelector)
	}

	// validate itemSelector. when invalid we still continue with an empty
	// itemSelQuery so that downstream checks (formatter, enforcement values) can
	// run; only the JQ-composition check is skipped to avoid cascading noise.
	itemSelQuery := ""
	itemSelOK := true
	if convutils.IsStringHasValue(rule.ItemSelector) {
		variants := variation.ItemSelectors
		if len(variants) == 0 {
			issues.addf("%s: field does not accept an itemSelector, but %q was provided", prefix, *rule.ItemSelector)
			itemSelOK = false
		} else if q, found := variants[*rule.ItemSelector]; !found {
			issues.addWithDetails(
				fmt.Sprintf("%s: itemSelector %q is not valid for this field", prefix, *rule.ItemSelector),
				fmt.Sprintf("valid itemSelectors: %v", sortedKeys(variants)),
			)
			itemSelOK = false
		} else {
			itemSelQuery = q.Query
		}
	} else if len(variation.ItemSelectors) > 0 {
		issues.addWithDetails(
			fmt.Sprintf("%s: itemSelector is required for this field", prefix),
			fmt.Sprintf("valid itemSelectors: %v", sortedKeys(variation.ItemSelectors)),
		)
		itemSelOK = false
	}

	// match the gvk_compliance resolver: it composes spec | item and compiles
	// the combined query. catching the failure here surfaces the same
	// brokenness as a validation error at create/update time. skip when the
	// spec query was already flagged empty or the itemSelector was rejected:
	// either condition would cascade into a misleading parse error here.
	if !specQueryEmpty && itemSelOK {
		issues.add(validateComposedSelectorQuery(prefix, specSel.Query, itemSelQuery))
	}

	// selector params depend on the spec query for arg discovery; skip when
	// the query is empty so every user-supplied param doesn't get flagged as
	// "unrecognized" purely because the args list could not be extracted.
	if !specQueryEmpty {
		issues.add(validateSelectorParams(prefix, rule.SelectorParams, specSel, itemSelQuery))
	}

	fieldType := effectiveFieldType(variation.Type)

	formatter := variation.Formatter

	issues.add(validateRuleFormatter(prefix, formatter))

	hasItemSelectors := len(variation.ItemSelectors) > 0

	byEnforcementType := make(map[string][]*api.PolicyRuleEnforcement)

	for j := range rule.Enforcements {
		enforcement := rule.Enforcements[j]
		enforcementPrefix := fmt.Sprintf("%s enforcement[%d]", prefix, j)

		if err := validateSingleEnforcement(enforcementPrefix, enforcement); err != nil {
			// can't validate further without a single populated enforcement variant
			issues.add(err)
			continue
		}

		issues.add(validateEnforcementsForType(enforcementPrefix, enforcement, fieldType))

		if enforcement.Items != nil && !hasItemSelectors {
			issues.addf("%s: items is only applicable to fields with itemSelectors", enforcementPrefix)
		}

		issues.add(validateEnforcementValues(enforcementPrefix, enforcement, formatter, fieldType))

		if k := PolicyRuleEnforcementKind(&rule.Enforcements[j]); k != "" {
			if k != EnforcementSet && k != EnforcementDefault && len(byEnforcementType[k]) > 0 {
				issues.addf(
					"%s: duplicate %s enforcement is not allowed in the same rule",
					enforcementPrefix, k)
			}
			byEnforcementType[k] = append(byEnforcementType[k], &rule.Enforcements[j])
		}
	}

	validateSetDefaultUniqueness(prefix, formatter, rule.Enforcements, issues)

	issues.add(runPostValidateRuleEnforcements(prefix, fieldType, formatter, byEnforcementType))
}

// validateSetDefaultUniqueness enforces the per-rule cardinality of set and
// default enforcements on a given policy:
//
//   - Flat field: at most one set OR default enforcement per rule.
//   - Itemized field: at most one set OR default enforcement per arrayKey
//     value (see setDefaultGroupingArg for how the arrayKey argument is
//     derived).
//
// Applies to a single given rule only. The effective policy may still carry
// both a set and a default for the same field/arrayKey value when they come
// from different scopes; that layering is disambiguated at effective time
// (the set wins over the default).
func validateSetDefaultUniqueness(prefix string, formatter api.BreakdownFormatter, enforcements []api.PolicyRuleEnforcement, issues *ruleIssues) {
	arrayKeyArg := setDefaultGroupingArg(formatter)

	// Each set/default enforcement writes to one "group" on the target:
	//
	//   - Flat field (arrayKeyArg == ""): a single implicit group; every
	//     set/default collides with the first one.
	//   - Keyed field, concrete arrayKey value: one group per distinct value.
	//     Enforcements targeting different values are independent.
	//   - Keyed field, arrayKey argument omitted: the KeyValueMap fan-out
	//     form ({ value: <v> }). The runtime applies it to every existing
	//     key of the target map, so it conflicts with every other
	//     set/default on this field.
	//
	// The flat case and the fan-out case behave the same way once seen: any
	// subsequent set/default is a duplicate. We track them together in
	// `broadEntry`. Concrete-key entries live in `byArrayKey` so that
	// enforcements against different keys can coexist.
	//
	// `broadEntry` and a non-empty `byArrayKey` never coexist: once one
	// regime is established, entries from the other regime are reported as
	// duplicates and dropped.
	type ref struct {
		idx  int
		kind string
	}
	var broadEntry *ref
	byArrayKey := map[string]ref{}

	reportFlatDup := func(idx int, prev ref) {
		issues.addf(
			"%s enforcement[%d]: only one set or default is allowed per rule for this field; the same rule already has %s at enforcement[%d]",
			prefix, idx, prev.kind, prev.idx)
	}
	reportSameArrayKeyDup := func(idx int, key string, prev ref) {
		issues.addf(
			"%s enforcement[%d]: only one set or default is allowed per arrayKey value; the same rule already has %s for %s=%q at enforcement[%d]",
			prefix, idx, prev.kind, arrayKeyArg, key, prev.idx)
	}
	reportFanOutDup := func(idx int, prev ref) {
		issues.addf(
			"%s enforcement[%d]: only one set or default is allowed per arrayKey value; a set/default that omits %q fans out to every key and conflicts with %s at enforcement[%d]",
			prefix, idx, arrayKeyArg, prev.kind, prev.idx)
	}

	earliestByArrayKey := func() ref {
		var earliest ref
		first := true
		for _, r := range byArrayKey {
			if first || r.idx < earliest.idx {
				earliest, first = r, false
			}
		}
		return earliest
	}

	for idx := range enforcements {
		// The switch both filters out enforcement kinds we don't care about
		// (via the default `continue`) and picks the arguments map that
		// belongs to the kind we did match, so the two can't drift apart.
		kind := PolicyRuleEnforcementKind(&enforcements[idx])
		var values map[string]interface{}
		switch kind {
		case EnforcementSet:
			values = enforcements[idx].Set
		case EnforcementDefault:
			values = enforcements[idx].Default
		default:
			continue
		}

		this := ref{
			idx:  idx,
			kind: kind,
		}

		// Flat field: a single implicit group.
		if arrayKeyArg == "" {
			if broadEntry == nil {
				broadEntry = &this
			} else {
				reportFlatDup(idx, *broadEntry)
			}
			continue
		}

		if key, hasKey := concreteArrayKey(values, arrayKeyArg); hasKey {
			// Concrete arrayKey value: collides with the same key, or with a
			// prior fan-out that also targets it.
			switch {
			case broadEntry != nil:
				reportFanOutDup(idx, *broadEntry)
			default:
				if prev, dup := byArrayKey[key]; dup {
					reportSameArrayKeyDup(idx, key, prev)
				} else {
					byArrayKey[key] = this
				}
			}
			continue
		}

		// Fan-out form: no arrayKey, so it targets every key. Collides with
		// any earlier set/default on this field.
		switch {
		case broadEntry != nil:
			reportFanOutDup(idx, *broadEntry)
		case len(byArrayKey) > 0:
			reportFanOutDup(idx, earliestByArrayKey())
		default:
			broadEntry = &this
		}
	}
}

// concreteArrayKey extracts the arrayKey value from a set/default's arguments
// map. Returns (value, true) when the argument is present and non-nil;
// otherwise the enforcement is the KeyValueMap fan-out form (see
// validateSetDefaultUniqueness).
func concreteArrayKey(values map[string]interface{}, arrayKeyArg string) (string, bool) {
	v, ok := values[arrayKeyArg]
	if !ok || v == nil {
		return "", false
	}
	return fmt.Sprintf("%v", v), true
}

// setDefaultGroupingArg returns the argument name that identifies an
// individual entry inside a set/default enforcement, or the empty string
// when the field is flat. Two enforcements sharing the same value of this
// argument in the same rule are duplicates; two enforcements with
// different values are independent.
//
//   - Formatters that declare ArrayKey (setExpr-based, e.g. env vars) use
//     the declared argument.
//   - The KeyValueMap builtin uses "key" as its implicit arrayKey — the
//     runtime set expression is
//     { ($ARGS.named.key // "value"): $ARGS.named.value }
//     which is functionally equivalent to a formatter that keys entries
//     by the "key" argument.
//   - Everything else (KeyValue builtin, setExpr formatter without
//     arrayKey) is flat.
func setDefaultGroupingArg(formatter api.BreakdownFormatter) string {
	if formatter.ArrayKey != nil {
		return formatter.ArrayKey.ArgName
	}
	if convutils.IsStringHasValue(formatter.Builtin) && *formatter.Builtin == dbmodel.BuiltinFormatterKeyValueMap {
		return KeyValueMapKeyArg
	}
	return ""
}

func findVariationForRule(rule api.PolicyRule, variations []api.BreakdownFieldVariation) *api.BreakdownFieldVariation {
	for i, v := range variations {
		for _, sel := range v.SpecSelectors {
			if sel == rule.SpecSelector {
				return &variations[i]
			}
		}
	}
	return nil
}

func validateSelectorParams(prefix string, params map[string]interface{}, specSel api.BreakdownSpecSelector, itemSelQuery string) *api.Error {
	validArgs := collectQueryArgs(specSel.Query, itemSelQuery)

	for paramName := range params {
		if !validArgs.Has(paramName) {
			return validationErrorWithDetails(
				fmt.Sprintf("%s: selectorParam %q is not a recognized argument", prefix, paramName),
				fmt.Sprintf("valid arguments: %v", validArgs.Keys()),
			)
		}
	}

	for paramName, paramValue := range params {
		if enumValues, hasEnum := specSel.ArgsEnum[paramName]; hasEnum {
			strVal := fmt.Sprintf("%v", paramValue)
			found := false
			for _, e := range enumValues {
				if e == strVal {
					found = true
					break
				}
			}
			if !found {
				return api.NewValidationError(fmt.Sprintf("%s: selectorParam %q value %q is not one of the allowed values %v", prefix, paramName, strVal, enumValues))
			}
			continue
		}

		argType, hasType := specSel.ArgsInfo[paramName]
		if !hasType {
			argType = api.MAPPINGFIELDTYPE_STRING
		}

		if err := validateValueForType(fmt.Sprintf("%s selectorParam %q", prefix, paramName), paramValue, argType); err != nil {
			return err
		}
	}

	return nil
}

func collectQueryArgs(specSelQuery string, itemSelQuery string) datastructs.Set[string] {
	args := datastructs.NewSet[string]()

	if q, err := gojq.Parse(specSelQuery); err == nil {
		CollectArgsNamedReferences(q, args)
	}

	if itemSelQuery != "" {
		if q, err := gojq.Parse(itemSelQuery); err == nil {
			CollectArgsNamedReferences(q, args)
		}
	}

	return args
}

// validateSingleEnforcement rejects rule enforcement that don't specify exactly one
// enforcement kind (e.g. set, required, blocked, match).
func validateSingleEnforcement(prefix string, enforcement api.PolicyRuleEnforcement) *api.Error {
	count := 0
	if enforcement.Set != nil {
		count++
	}
	if enforcement.Default != nil {
		count++
	}
	if enforcement.Required != nil {
		count++
	}
	if enforcement.Blocked != nil {
		count++
	}
	if enforcement.Length != nil {
		count++
	}
	if enforcement.Options != nil {
		count++
	}
	if enforcement.Match != nil {
		count++
	}
	if enforcement.Range != nil {
		count++
	}
	if enforcement.Elements != nil {
		count++
	}
	if enforcement.Items != nil {
		count++
	}
	if enforcement.QuantityRange != nil {
		count++
	}

	if count == 0 {
		return api.NewValidationError(fmt.Sprintf("%s: enforcement must have exactly one non-nil field", prefix))
	}
	if count > 1 {
		return api.NewValidationError(fmt.Sprintf("%s: enforcement must have exactly one non-nil field, found %d", prefix, count))
	}

	return nil
}

func newTypeSet(types ...api.MappingFieldType) datastructs.Set[api.MappingFieldType] {
	return datastructs.NewSet[api.MappingFieldType](types...)
}

var (
	stringTypes  = newTypeSet(api.MAPPINGFIELDTYPE_STRING, api.MAPPINGFIELDTYPE_NAMES_ARRAY, api.MAPPINGFIELDTYPE_STRINGS)
	numericTypes = newTypeSet(api.MAPPINGFIELDTYPE_INTEGER, api.MAPPINGFIELDTYPE_NUMBER, api.MAPPINGFIELDTYPE_INTEGERS, api.MAPPINGFIELDTYPE_NUMBERS)
	quantityType = newTypeSet(api.MAPPINGFIELDTYPE_QUANTITY)
	arrayTypes   = newTypeSet(api.MAPPINGFIELDTYPE_STRINGS, api.MAPPINGFIELDTYPE_INTEGERS, api.MAPPINGFIELDTYPE_NUMBERS)
	allTypes     = func() datastructs.Set[api.MappingFieldType] {
		s := datastructs.NewSet[api.MappingFieldType]()
		for _, t := range api.AllowedMappingFieldTypeEnumValues {
			s.Add(t)
		}
		return s
	}()
)

const (
	EnforcementSet           = "set"
	EnforcementDefault       = "default"
	EnforcementRequired      = "required"
	EnforcementBlocked       = "blocked"
	EnforcementMatch         = "match"
	EnforcementOptions       = "options"
	EnforcementLength        = "length"
	EnforcementRange         = "range"
	EnforcementQuantityRange = "quantityRange"
	EnforcementElements      = "elements"
	EnforcementItems         = "items"

	DefaultArgName = "value"

	// KeyValueMapKeyArg is the argument name that carries the map key on a
	// set/default enforcement targeting a KeyValueMap-formatted field. The
	// runtime set expression for the KeyValueMap builtin is
	//     { ($ARGS.named.key // "value"): $ARGS.named.value }
	// so this argument acts as the implicit arrayKey when detecting
	// duplicate set/default enforcements against a KeyValueMap field.
	KeyValueMapKeyArg = "key"
)

type enforcementCheck struct {
	name    string
	present func(api.PolicyRuleEnforcement) bool
	allowed datastructs.Set[api.MappingFieldType]
}

var enforcementsTable = []enforcementCheck{
	{EnforcementSet, func(a api.PolicyRuleEnforcement) bool { return a.Set != nil }, allTypes},
	{EnforcementDefault, func(a api.PolicyRuleEnforcement) bool { return a.Default != nil }, allTypes},
	{EnforcementRequired, func(a api.PolicyRuleEnforcement) bool { return a.Required != nil }, allTypes},
	{EnforcementBlocked, func(a api.PolicyRuleEnforcement) bool { return a.Blocked != nil }, allTypes},
	{EnforcementMatch, func(a api.PolicyRuleEnforcement) bool { return a.Match != nil }, stringTypes},
	{EnforcementOptions, func(a api.PolicyRuleEnforcement) bool { return a.Options != nil }, stringTypes},
	{EnforcementLength, func(a api.PolicyRuleEnforcement) bool { return a.Length != nil }, stringTypes},
	{EnforcementRange, func(a api.PolicyRuleEnforcement) bool { return a.Range != nil }, numericTypes},
	{EnforcementQuantityRange, func(a api.PolicyRuleEnforcement) bool { return a.QuantityRange != nil }, quantityType},
	{EnforcementElements, func(a api.PolicyRuleEnforcement) bool { return a.Elements != nil }, arrayTypes},
	// items is validated separately — it requires itemSelectors, not a specific field type
}

// effectiveFieldType returns the field type to use when reasoning about
// applicable enforcements. An empty type is normalized to STRING — this
// matches the convention used elsewhere for missing types on the value side
// (selector args and formatter args also default to STRING when no type is
// declared, see validateSelectorParams / validateEnforcementValues). In
// well-formed data the type is always set explicitly (the schema enforces a
// non-empty enum), so the fallback is purely defensive.
func effectiveFieldType(fieldType api.MappingFieldType) api.MappingFieldType {
	if fieldType == "" {
		return api.MAPPINGFIELDTYPE_STRING
	}
	return fieldType
}

// possibleEnforcementsForField returns the enforcement kinds that may be
// applied to a field with the given type, itemSelectors presence, and
// formatter. The result is derived from enforcementsTable (the same allow-list
// used at rule validation time), with the same formatter-capability checks
// validateEnforcementValues applies: set/default need a builtin formatter or
// setExpr, while read-based constraints need a builtin formatter or getExpr.
// The type-independent `items` kind is appended when hasItemSelectors is true.
// The order follows enforcementsTable so the output is deterministic and
// matches how the validator iterates.
func possibleEnforcementsForField(fieldType api.MappingFieldType, hasItemSelectors bool, formatter api.BreakdownFormatter) []api.MappingPossibleEnforcement {
	fieldType = effectiveFieldType(fieldType)
	out := make([]api.MappingPossibleEnforcement, 0, len(enforcementsTable)+1)
	canSet := formatterCanSet(formatter)
	canValidate := formatterCanValidate(formatter)
	for _, check := range enforcementsTable {
		if !check.allowed.Has(fieldType) {
			continue
		}

		switch check.name {
		case EnforcementSet, EnforcementDefault:
			if canSet {
				out = append(out, api.MappingPossibleEnforcement(check.name))
			}
		default:
			if canValidate {
				out = append(out, api.MappingPossibleEnforcement(check.name))
			}
		}
	}
	if hasItemSelectors && canValidate {
		out = append(out, api.MappingPossibleEnforcement(EnforcementItems))
	}
	return out
}

// formatterCanSet reports whether set/default enforcements can be applied to a
// field with the given formatter.
func formatterCanSet(formatter api.BreakdownFormatter) bool {
	return convutils.IsStringHasValue(formatter.Builtin) || formatter.SetExpr != nil
}

// formatterCanValidate reports whether the non-set enforcements can be applied
// to a field with the given formatter.
func formatterCanValidate(formatter api.BreakdownFormatter) bool {
	return formatter.Builtin != nil || formatter.GetExpr != nil
}

func validateEnforcementsForType(prefix string, enforcement api.PolicyRuleEnforcement, fieldType api.MappingFieldType) *api.Error {
	for _, check := range enforcementsTable {
		if check.present(enforcement) && !check.allowed.Has(fieldType) {
			return api.NewValidationError(fmt.Sprintf("%s: %q is not applicable to field type %q", prefix, check.name, fieldType))
		}
	}
	return nil
}

func validateEnforcementValues(prefix string, enforcement api.PolicyRuleEnforcement, formatter api.BreakdownFormatter, fieldType api.MappingFieldType) *api.Error {
	if enforcement.Match != nil {
		if _, err := regexp.Compile(*enforcement.Match); err != nil {
			return api.NewValidationError(fmt.Sprintf("%s: match %q is not a valid regular expression: %v", prefix, *enforcement.Match, err))
		}
	}

	if enforcement.Range != nil {
		if err := validateRangeEnforcement(prefix, enforcement.Range); err != nil {
			return err
		}
	}

	if enforcement.QuantityRange != nil {
		if err := validateQuantityRangeEnforcement(prefix, enforcement.QuantityRange); err != nil {
			return err
		}
	}

	if enforcement.Length != nil {
		if err := validateIntRangeEnforcement(prefix, EnforcementLength, enforcement.Length.Min, enforcement.Length.Max); err != nil {
			return err
		}
	}

	if enforcement.Elements != nil {
		if err := validateIntRangeEnforcement(prefix, EnforcementElements, enforcement.Elements.Min, enforcement.Elements.Max); err != nil {
			return err
		}
	}

	if enforcement.Items != nil {
		if err := validateIntRangeEnforcement(prefix, EnforcementItems, enforcement.Items.Min, enforcement.Items.Max); err != nil {
			return err
		}
	}

	if enforcement.Set != nil || enforcement.Default != nil {
		values := enforcement.Set
		enforcementName := "set"
		if values == nil {
			values = enforcement.Default
			enforcementName = "default"
		}

		switch {
		case convutils.IsStringHasValue(formatter.Builtin):
			if err := validateBuiltinSetValues(prefix, enforcementName, *formatter.Builtin, values, fieldType); err != nil {
				return err
			}
		case formatter.SetExpr != nil:
			setExprArgs := datastructs.NewSet[string]()
			if q, err := gojq.Parse(*formatter.SetExpr); err == nil {
				CollectArgsNamedReferences(q, setExprArgs)
			}

			for paramName := range values {
				if !setExprArgs.Has(paramName) {
					return validationErrorWithDetails(
						fmt.Sprintf("%s: %s parameter %q is not a recognized argument in the formatter", prefix, enforcementName, paramName),
						fmt.Sprintf("valid arguments: %v", setExprArgs.Keys()),
					)
				}
			}

			for paramName, paramValue := range values {
				if enumValues, hasEnum := formatter.ArgsEnum[paramName]; hasEnum {
					strVal := fmt.Sprintf("%v", paramValue)
					found := false
					for _, e := range enumValues {
						if e == strVal {
							found = true
							break
						}
					}
					if !found {
						return api.NewValidationError(fmt.Sprintf("%s: %s parameter %q value %q is not one of the allowed values %v", prefix, enforcementName, paramName, strVal, enumValues))
					}
					continue
				}

				argType, hasType := formatter.ArgsInfo[paramName]
				if !hasType {
					argType = api.MAPPINGFIELDTYPE_STRING
				}

				if err := validateValueForType(fmt.Sprintf("%s %s parameter %q", prefix, enforcementName, paramName), paramValue, argType); err != nil {
					return err
				}
			}
		default:
			return api.NewValidationError(fmt.Sprintf(
				"%s: %s is only valid when the field uses the KeyValue built-in formatter or a formatter with setExpr; this field's formatter has neither",
				prefix, enforcementName))
		}
		return nil
	}

	//
	//  all other enforcements require get expression or builtin formatter
	//
	if formatter.GetExpr == nil && formatter.Builtin == nil {
		return api.NewValidationError(fmt.Sprintf(
			"%s: constraint enforcements require a builtin formatter or a formatter with getExpr", prefix))
	}

	return nil
}

// validateRangeEnforcement checks that the numeric range is internally
// consistent: at least one bound must be specified (otherwise the rule is a
// no-op), min must not exceed max when both are present, step must be
// positive, and step must not exceed max-min when all three are present
// (otherwise only `min` itself would match, which is almost certainly a
// misconfiguration and is better expressed as an `options`/exact rule).
func validateRangeEnforcement(prefix string, r *api.PolicyEnforcementRangeRange) *api.Error {
	if r.Min == nil && r.Max == nil && r.Step == nil {
		return api.NewValidationError(fmt.Sprintf("%s: range must specify at least one of min/max/step", prefix))
	}
	if r.Min != nil && r.Max != nil && *r.Min > *r.Max {
		return api.NewValidationError(fmt.Sprintf("%s: range.min (%v) must be less than or equal to range.max (%v)", prefix, *r.Min, *r.Max))
	}
	if r.Step != nil && *r.Step <= 0 {
		return api.NewValidationError(fmt.Sprintf("%s: range.step must be greater than 0, got %v", prefix, *r.Step))
	}
	if r.Min != nil && r.Max != nil && r.Step != nil {
		if span := *r.Max - *r.Min; *r.Step > span {
			return api.NewValidationError(fmt.Sprintf("%s: range.step (%v) must be less than or equal to range.max-range.min (%v)", prefix, *r.Step, span))
		}
	}
	return nil
}

// validateIntRangeEnforcement checks that an int32 min/max rule is internally
// consistent: at least one of min/max must be present, each must be
// non-negative, and min must not exceed max.
func validateIntRangeEnforcement(prefix, enforcementName string, minVal, maxVal *int32) *api.Error {
	if minVal == nil && maxVal == nil {
		return api.NewValidationError(fmt.Sprintf("%s: %s must specify at least one of min/max", prefix, enforcementName))
	}
	if minVal != nil && *minVal < 0 {
		return api.NewValidationError(fmt.Sprintf("%s: %s.min must be non-negative, got %d", prefix, enforcementName, *minVal))
	}
	if maxVal != nil && *maxVal < 0 {
		return api.NewValidationError(fmt.Sprintf("%s: %s.max must be non-negative, got %d", prefix, enforcementName, *maxVal))
	}
	if minVal != nil && maxVal != nil && *minVal > *maxVal {
		return api.NewValidationError(fmt.Sprintf("%s: %s.min (%d) must be less than or equal to %s.max (%d)", prefix, enforcementName, *minVal, enforcementName, *maxVal))
	}
	return nil
}

// validateQuantityRangeEnforcement checks that every specified bound parses
// as a k8s resource quantity, that min does not exceed max, that step is
// strictly positive, and that step does not exceed max-min when all three
// are present (see validateRangeEnforcement for the rationale). At least
// one bound must be specified.
func validateQuantityRangeEnforcement(prefix string, r *api.PolicyEnforcementQuantityRangeQuantityRange) *api.Error {
	if r.Min == nil && r.Max == nil && r.Step == nil {
		return api.NewValidationError(fmt.Sprintf("%s: quantityRange must specify at least one of min/max/step", prefix))
	}

	parsed := map[string]resource.Quantity{}
	for _, qv := range []struct {
		name  string
		value *string
	}{
		{"min", r.Min},
		{"max", r.Max},
		{"step", r.Step},
	} {
		if qv.value == nil {
			continue
		}
		q, err := resource.ParseQuantity(*qv.value)
		if err != nil {
			return api.NewValidationError(fmt.Sprintf("%s: quantityRange.%s %q is not a valid Kubernetes resource quantity", prefix, qv.name, *qv.value))
		}
		parsed[qv.name] = q
	}

	minQ, hasMin := parsed["min"]
	maxQ, hasMax := parsed["max"]
	stepQ, hasStep := parsed["step"]

	if hasMin && hasMax && minQ.Cmp(maxQ) > 0 {
		return api.NewValidationError(fmt.Sprintf("%s: quantityRange.min (%s) must be less than or equal to quantityRange.max (%s)", prefix, minQ.String(), maxQ.String()))
	}

	if hasStep && stepQ.Sign() <= 0 {
		return api.NewValidationError(fmt.Sprintf("%s: quantityRange.step must be greater than 0, got %s", prefix, stepQ.String()))
	}

	if hasMin && hasMax && hasStep {
		span := maxQ.DeepCopy()
		span.Sub(minQ)
		if stepQ.Cmp(span) > 0 {
			return api.NewValidationError(fmt.Sprintf("%s: quantityRange.step (%s) must be less than or equal to quantityRange.max-quantityRange.min (%s)", prefix, stepQ.String(), span.String()))
		}
	}

	return nil
}

// validateBuiltinSetValues enforces the set/default values shape for the two
// builtin formatters:
//
//   - KeyValue: scalar single-leaf. The enforcement MUST contain exactly one
//     entry with key "value" and that value must conform to the field's
//     declared type.
//
//   - KeyValueMap: writes one or more entries into a map. The set expression
//     compiled at runtime is:
//
//     { ($ARGS.named.key // "value"): $ARGS.named.value }
//
//     Two shapes are accepted:
//
//   - { value: <v> }              — applied to every existing key of the
//     target map by the runtime's CopyTargetKeys
//     workaround (rules_action_post_validate /
//     gvk_compliance.enforcements).
//
//   - { key: <k>, value: <v> }    — sets the single key <k> to <v>.
//     "key" must be a string; <v> is a map leaf
//     and is not type-checked against the field
//     type (which describes the map itself, not
//     individual leaves).
func validateBuiltinSetValues(prefix, enforcementName, builtin string, values map[string]any, fieldType api.MappingFieldType) *api.Error {
	switch builtin {
	case dbmodel.BuiltinFormatterKeyValueMap:
		return validateKeyValueMapSet(prefix, enforcementName, builtin, values)
	case dbmodel.BuiltinFormatterKeyValue:
		return validateKeyValueSet(prefix, enforcementName, builtin, values, fieldType)
	default:
		// Should not happen — validateRuleFormatter rejects unknown builtins
		// before we get here.
		return nil
	}
}

func validateKeyValueMapSet(prefix, enforcementName, builtin string, values map[string]any) *api.Error {
	for paramName := range values {
		if paramName != "key" && paramName != DefaultArgName {
			return api.NewValidationError(fmt.Sprintf(
				"%s: %s for a %s field accepts only %q and %q keys; got unexpected key %q",
				prefix, enforcementName, builtin, "key", DefaultArgName, paramName))
		}
	}
	if _, ok := values[DefaultArgName]; !ok {
		return api.NewValidationError(fmt.Sprintf(
			"%s: %s for a %s field must include the %q key",
			prefix, enforcementName, builtin, DefaultArgName))
	}
	if keyVal, hasKey := values["key"]; hasKey {
		// The runtime set expression for KeyValueMap is
		//     { ($ARGS.named.key // "value"): $ARGS.named.value }
		// JQ object construction requires the key expression to evaluate to a
		// string. validateValueForType is too lax here (it returns nil for the
		// STRING field type, accepting numbers/bools/nulls), so enforce the
		// string-typed key explicitly and surface a clear API-level error.
		if _, isStr := keyVal.(string); !isStr {
			return api.NewValidationError(fmt.Sprintf(
				"%s: %s for a %s field requires the %q value to be a string; got %T",
				prefix, enforcementName, builtin, "key", keyVal))
		}
	}
	return nil
}

func validateKeyValueSet(prefix, enforcementName, builtin string, values map[string]any, fieldType api.MappingFieldType) *api.Error {
	if len(values) != 1 {
		return api.NewValidationError(fmt.Sprintf(
			"%s: %s for a %s field must contain exactly one entry with key %q, got %d entries",
			prefix, enforcementName, builtin, DefaultArgName, len(values)))
	}
	val, ok := values[DefaultArgName]
	if !ok {
		return api.NewValidationError(fmt.Sprintf(
			"%s: %s for a %s field must use key %q",
			prefix, enforcementName, builtin, DefaultArgName))
	}
	return validateValueForType(fmt.Sprintf("%s %s", prefix, enforcementName), val, fieldType)
}

func validateValueForType(prefix string, value interface{}, fieldType api.MappingFieldType) *api.Error {
	strVal := fmt.Sprintf("%v", value)

	switch fieldType {
	case api.MAPPINGFIELDTYPE_STRING, api.MAPPINGFIELDTYPE_NAMES_ARRAY, api.MAPPINGFIELDTYPE_OBJECT:
		return nil
	case api.MAPPINGFIELDTYPE_BOOLEAN:
		if strVal != "true" && strVal != "false" {
			return api.NewValidationError(fmt.Sprintf("%s: expected boolean, got %q", prefix, strVal))
		}
	case api.MAPPINGFIELDTYPE_INTEGER:
		if _, err := strconv.ParseInt(strVal, 10, 64); err != nil {
			return api.NewValidationError(fmt.Sprintf("%s: expected integer, got %q", prefix, strVal))
		}
	case api.MAPPINGFIELDTYPE_NUMBER:
		if _, err := strconv.ParseFloat(strVal, 64); err != nil {
			return api.NewValidationError(fmt.Sprintf("%s: expected number, got %q", prefix, strVal))
		}
	case api.MAPPINGFIELDTYPE_QUANTITY:
		if _, err := resource.ParseQuantity(strVal); err != nil {
			return api.NewValidationError(fmt.Sprintf("%s: expected quantity, got %q", prefix, strVal))
		}
	case api.MAPPINGFIELDTYPE_IPADDRESS:
		if net.ParseIP(strVal) == nil {
			return api.NewValidationError(fmt.Sprintf("%s: expected IP address, got %q", prefix, strVal))
		}
	case api.MAPPINGFIELDTYPE_INTEGERS:
		return validateArrayValue(prefix, value, api.MAPPINGFIELDTYPE_INTEGER)
	case api.MAPPINGFIELDTYPE_NUMBERS:
		return validateArrayValue(prefix, value, api.MAPPINGFIELDTYPE_NUMBER)
	case api.MAPPINGFIELDTYPE_STRINGS:
		return validateArrayValue(prefix, value, api.MAPPINGFIELDTYPE_STRING)
	}

	return nil
}

func validateArrayValue(prefix string, value interface{}, elementType api.MappingFieldType) *api.Error {
	arr, ok := value.([]interface{})
	if !ok {
		strVal := fmt.Sprintf("%v", value)
		if strings.Contains(strVal, ",") {
			parts := strings.Split(strVal, ",")
			for i, part := range parts {
				elemPrefix := fmt.Sprintf("%s[%d]", prefix, i)
				if err := validateValueForType(elemPrefix, strings.TrimSpace(part), elementType); err != nil {
					return err
				}
			}
			return nil
		}
		return api.NewValidationError(fmt.Sprintf("%s: expected array, got %T", prefix, value))
	}

	for i, elem := range arr {
		elemPrefix := fmt.Sprintf("%s[%d]", prefix, i)
		if err := validateValueForType(elemPrefix, elem, elementType); err != nil {
			return err
		}
	}

	return nil
}

func validationErrorWithDetails(message string, details string) *api.Error {
	err := api.NewValidationError(message)
	err.Details = details
	return err
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func collectSpecSelectors(variations []api.BreakdownFieldVariation) []string {
	seen := map[string]bool{}
	for _, v := range variations {
		for _, sel := range v.SpecSelectors {
			seen[sel] = true
		}
	}
	return sortedKeys(seen)
}

// validateRuleFormatter cross-checks the formatter declared on the rule's
// variation against what the runtime is willing to accept. It catches three
// failure modes that gvk_compliance otherwise only reports via LogBrokenRule
// at apply time:
//
//   - an unrecognized Builtin (the runtime falls into the default branch of
//     NewRuntimeFormatter and returns nil);
//   - a getExpr that fails to parse or compile (compileJQNoArgs at runtime);
//   - a setExpr that fails to parse or compile (compileJQWithArgs at runtime).
//
// validateEnforcementValues already parses the setExpr for argument
// extraction, but does not compile it; the compile step here closes that
// gap and adds the equivalent for getExpr.
func validateRuleFormatter(prefix string, formatter api.BreakdownFormatter) *api.Error {
	if convutils.IsStringHasValue(formatter.Builtin) {
		name := *formatter.Builtin
		if !builtinFormatters.Has(name) {
			allowed := builtinFormatters.Keys()
			sort.Strings(allowed)
			return validationErrorWithDetails(
				fmt.Sprintf("%s: formatter has unknown builtin %q", prefix, name),
				fmt.Sprintf("known builtin formatters: %v", allowed),
			)
		}
	}

	if convutils.IsStringHasValue(formatter.GetExpr) {
		if err := compileJQForRule(prefix, "formatter getExpr", *formatter.GetExpr, false); err != nil {
			return err
		}
	}

	if convutils.IsStringHasValue(formatter.SetExpr) {
		if err := compileJQForRule(prefix, "formatter setExpr", *formatter.SetExpr, true); err != nil {
			return err
		}
	}

	return nil
}

// validateComposedSelectorQuery composes the spec selector query with the
// optional item selector query (joined by `|`, exactly the way
// gvk_compliance.ResolveRules does it) and verifies that the result both
// parses and compiles. The per-piece gojq.Parse done by collectQueryArgs is
// not enough: a query with an empty body or one that only fails at compile
// time (e.g. an undefined function reference) slips through, and only the
// runtime would observe the failure.
func validateComposedSelectorQuery(prefix, specSelQuery, itemSelQuery string) *api.Error {
	composed := strings.TrimSpace(specSelQuery)
	if itemSelQuery != "" {
		composed += " | " + strings.TrimSpace(itemSelQuery)
	}
	return compileJQForRule(prefix, "composed selector query", composed, true)
}

// compileJQForRule parses and compiles the given JQ expression, returning a
// validation error on either failure. When withArgs is true, the $ARGS
// variable is declared (the runtime does this for selector queries and for
// formatter setExpr); when false, no variables are declared (the runtime
// does this for formatter getExpr).
func compileJQForRule(prefix, exprName, expr string, withArgs bool) *api.Error {
	parsed, err := gojq.Parse(expr)
	if err != nil {
		return api.NewValidationError(fmt.Sprintf("%s: %s failed to parse: %v", prefix, exprName, err))
	}
	if err := validateParsedJQSafety(parsed); err != nil {
		return api.NewValidationError(fmt.Sprintf("%s: %s is unsafe: %v", prefix, exprName, err))
	}
	var opts []gojq.CompilerOption
	if withArgs {
		opts = append(opts, gojq.WithVariables([]string{"$ARGS"}))
	}
	if _, err := gojq.Compile(parsed, opts...); err != nil {
		return api.NewValidationError(fmt.Sprintf("%s: %s failed to compile: %v", prefix, exprName, err))
	}
	return nil
}
