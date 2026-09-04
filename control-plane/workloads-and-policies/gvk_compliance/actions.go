/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gvk_compliance

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/conjungo"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/validation"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Enforcement is the runtime representation of a single PolicyRuleEnforcement.
//
// selectorPath is the JQ path of v in the workload — enforcements use it to
// pinpoint the offending field in errors (e.g.
// ".spec.template.spec.containers[0].image"). Nil when the selector
// chain isn't path-eligible; enforcements that
// don't need it pass _ []any.
type Enforcement interface {
	Exec(fieldName string, rule *ResolvedRule, selectorPath []any, v any) *policyApi.Error
}

// FinalizableEnforcement is an optional interface implemented by enforcements that need
// to perform validation after all items produced by the rule's selector have
// been iterated (e.g. MinItems, which can only be decided once the total count
// is known).
type FinalizableEnforcement interface {
	Finalize(fieldName string, rule *ResolvedRule) *policyApi.Error
}

// newEnforcement constructs the matching enforcement for whichever variant is populated on the given
// PolicyRuleEnforcement. Returns nil if none of the known variants is set.
func newEnforcement(a *policyApi.PolicyRuleEnforcement) Enforcement {
	switch {
	case a.PolicyEnforcementSet != nil:
		return &SetEnforcement{value: a.PolicyEnforcementSet.GetSet(), overwrite: true}
	case a.PolicyEnforcementDefault != nil:
		return &SetEnforcement{value: a.PolicyEnforcementDefault.GetDefault(), overwrite: false}
	case a.PolicyEnforcementMatch != nil:
		return &MatchEnforcement{enforcement: a.PolicyEnforcementMatch}
	case a.PolicyEnforcementElements != nil:
		return &ElementsEnforcement{enforcement: a.PolicyEnforcementElements}
	case a.PolicyEnforcementItems != nil:
		return &ItemsEnforcement{enforcement: a.PolicyEnforcementItems}
	case a.PolicyEnforcementLength != nil:
		return &LengthEnforcement{enforcement: a.PolicyEnforcementLength}
	case a.PolicyEnforcementOptions != nil:
		return &OptionsEnforcement{enforcement: a.PolicyEnforcementOptions}
	case a.PolicyEnforcementQuantityRange != nil:
		return &QuantityRangeEnforcement{enforcement: a.PolicyEnforcementQuantityRange}
	case a.PolicyEnforcementRange != nil:
		return &RangeEnforcement{enforcement: a.PolicyEnforcementRange}
	case a.PolicyEnforcementRequired != nil:
		return &RequiredEnforcement{enforcement: a.PolicyEnforcementRequired}
	case a.PolicyEnforcementBlocked != nil:
		return &BlockedEnforcement{enforcement: a.PolicyEnforcementBlocked}
	}
	return nil
}

type MatchEnforcement struct {
	enforcement *policyApi.PolicyEnforcementMatch
}

func (ac *MatchEnforcement) Exec(fieldName string, rule *ResolvedRule, _ []any, v any) *policyApi.Error {
	pattern := ac.enforcement.GetMatch()
	re, err := regexp.Compile(pattern)
	if err != nil {
		LogBrokenRule(rule.rule, "field %s has invalid match pattern %q: %v", fieldName, pattern, err)
		return nil
	}

	return runGetExpression(fieldName, "match", rule, v, func(calculated any) *policyApi.Error {
		for _, element := range asSlice(calculated) {
			s, ok := element.(string)
			if !ok {
				LogBrokenRule(rule.rule, "field %s value for 'match' enforcement is not a string (got %T)", fieldName, element)
				continue
			}
			if !re.MatchString(s) {
				return validationError(fieldName, "match",
					fmt.Errorf("value %q does not match pattern %q", s, pattern))
			}
		}
		return nil
	})
}

// LengthEnforcement enforces the value's length to lie within an inclusive
// [min, max] interval. validation.Length interprets length based on the
// value's kind: chars for strings, element count for slices/arrays/maps.
type LengthEnforcement struct {
	enforcement *policyApi.PolicyEnforcementLength
}

func (ac *LengthEnforcement) Exec(fieldName string, rule *ResolvedRule, _ []any, v any) *policyApi.Error {
	l := ac.enforcement.Length.Get()
	if l == nil {
		return nil
	}

	var minLen, maxLen *int
	if l.Min.IsSet() {
		mv := int(*l.Min.Get())
		minLen = &mv
	}
	if l.Max.IsSet() {
		mv := int(*l.Max.Get())
		maxLen = &mv
	}

	if !validateIntMinMax(fieldName, "length", rule, minLen, maxLen) {
		return nil
	}

	return runGetExpression(fieldName, "length", rule, v, func(calculated any) *policyApi.Error {
		if minLen != nil {
			lengthRule := validation.Length(*minLen, 0).Error(lengthRuleMessage(calculated, "no less than", *minLen))
			if err := lengthRule.Validate(calculated); err != nil {
				return validationError(fieldName, "length", err)
			}
		}
		if maxLen != nil {
			lengthRule := validation.Length(0, *maxLen).Error(lengthRuleMessage(calculated, "no more than", *maxLen))
			if err := lengthRule.Validate(calculated); err != nil {
				return validationError(fieldName, "length", err)
			}
		}
		return nil
	})
}

// RangeEnforcement enforces a numeric field to lie within an inclusive [min, max]
// interval, optionally constrained to increments of `step` starting from
// `min` (or from 0 when `min` is absent).
type RangeEnforcement struct {
	enforcement *policyApi.PolicyEnforcementRange
}

func (ac *RangeEnforcement) Exec(fieldName string, rule *ResolvedRule, _ []any, v any) *policyApi.Error {
	r := ac.enforcement.Range.Get()
	if r == nil {
		return nil
	}

	var minVal, maxVal, stepVal *float32
	if r.Min.IsSet() {
		minVal = r.Min.Get()
	}
	if r.Max.IsSet() {
		maxVal = r.Max.Get()
	}
	if r.Step.IsSet() {
		stepVal = r.Step.Get()
	}

	if !validateFloatRangeBounds(fieldName, "range", rule, minVal, maxVal, stepVal) {
		return nil
	}

	return runGetExpression(fieldName, "range", rule, v, func(calculated any) *policyApi.Error {
		for _, element := range asSlice(calculated) {
			if minVal != nil {
				if err := validation.Min(*minVal).Validate(element); err != nil {
					return validationError(fieldName, "range", err)
				}
			}
			if maxVal != nil {
				if err := validation.Max(*maxVal).Validate(element); err != nil {
					return validationError(fieldName, "range", err)
				}
			}
			if stepVal != nil {
				numeric, err := validation.ToFloat(element)
				if err != nil {
					return validationError(fieldName, "range", err)
				}
				anchor := float64(0)
				if minVal != nil {
					anchor = float64(*minVal)
				}
				if !OnStep(numeric, float64(*stepVal), anchor) {
					return validationError(fieldName, "range",
						fmt.Errorf("value %v must be a multiple of %v from %v",
							numeric, *stepVal, anchor))
				}
			}
		}
		return nil
	})
}

// QuantityRangeEnforcement is the k8s-quantity counterpart of RangeEnforcement. Each
// bound, plus the step, must be a valid Kubernetes resource quantity (e.g.
// 100Mi, 1Gi).
type QuantityRangeEnforcement struct {
	enforcement *policyApi.PolicyEnforcementQuantityRange
}

func (ac *QuantityRangeEnforcement) Exec(fieldName string, rule *ResolvedRule, _ []any, v any) *policyApi.Error {
	r := ac.enforcement.QuantityRange.Get()
	if r == nil {
		return nil
	}

	var minQty, maxQty, stepQty *resource.Quantity
	for _, qb := range []struct {
		name  string
		value *string
		out   **resource.Quantity
	}{
		{"quantityRange.min", r.Min.Get(), &minQty},
		{"quantityRange.max", r.Max.Get(), &maxQty},
		{"quantityRange.step", r.Step.Get(), &stepQty},
	} {
		if qb.value == nil {
			continue
		}
		q, err := resource.ParseQuantity(*qb.value)
		if err != nil {
			LogBrokenRule(rule.rule, "field %s has invalid %s %q: %v", fieldName, qb.name, *qb.value, err)
			return nil
		}
		*qb.out = &q
	}

	if !validateQuantityRangeBounds(fieldName, rule, minQty, maxQty, stepQty) {
		return nil
	}

	return runGetExpression(fieldName, "quantityRange", rule, v, func(calculated any) *policyApi.Error {
		for _, element := range asSlice(calculated) {
			q, apiErr := parseQuantity(fieldName, "quantityRange", element)
			if apiErr != nil {
				return apiErr
			}
			if minQty != nil && q.Cmp(*minQty) < 0 {
				return validationError(fieldName, "quantityRange",
					fmt.Errorf("value %s is below minimum %s", q.String(), minQty.String()))
			}
			if maxQty != nil && q.Cmp(*maxQty) > 0 {
				return validationError(fieldName, "quantityRange",
					fmt.Errorf("value %s exceeds maximum %s", q.String(), maxQty.String()))
			}
			if stepQty != nil {
				stepFloat := stepQty.AsApproximateFloat64()
				anchor := float64(0)
				if minQty != nil {
					anchor = minQty.AsApproximateFloat64()
				}
				if !OnStep(q.AsApproximateFloat64(), stepFloat, anchor) {
					return validationError(fieldName, "quantityRange",
						fmt.Errorf("value %s must be a multiple of %s from %s",
							q.String(), stepQty.String(), QuantityAnchorString(minQty)))
				}
			}
		}
		return nil
	})
}

// ElementsEnforcement enforces the number of elements of an array field to lie
// within an inclusive [min, max] interval.
type ElementsEnforcement struct {
	enforcement *policyApi.PolicyEnforcementElements
}

func (ac *ElementsEnforcement) Exec(fieldName string, rule *ResolvedRule, _ []any, v any) *policyApi.Error {
	e := ac.enforcement.Elements.Get()
	if e == nil {
		return nil
	}

	var minElements, maxElements *int
	if e.Min.IsSet() {
		mv := int(*e.Min.Get())
		minElements = &mv
	}
	if e.Max.IsSet() {
		mv := int(*e.Max.Get())
		maxElements = &mv
	}

	if !validateIntMinMax(fieldName, "elements", rule, minElements, maxElements) {
		return nil
	}

	return runGetExpression(fieldName, "elements", rule, v, func(calculated any) *policyApi.Error {
		if calculated == nil {
			return nil
		}
		slice, ok := calculated.([]any)
		if !ok {
			LogBrokenRule(rule.rule, "field %s value for 'elements' enforcement is not a slice (got %T)", fieldName, calculated)
			return nil
		}
		if minElements != nil && len(slice) < *minElements {
			return validationError(fieldName, "elements",
				fmt.Errorf("the number of elements must be no less than %d", *minElements))
		}
		if maxElements != nil && len(slice) > *maxElements {
			return validationError(fieldName, "elements",
				fmt.Errorf("the number of elements must be no more than %d", *maxElements))
		}
		return nil
	})
}

type OptionsEnforcement struct {
	enforcement *policyApi.PolicyEnforcementOptions
}

func (ac *OptionsEnforcement) Exec(fieldName string, rule *ResolvedRule, _ []any, v any) *policyApi.Error {
	options := ac.enforcement.GetOptions()
	values := make([]any, 0, len(options))
	for _, o := range options {
		values = append(values, o)
	}

	allowedList := strings.Join(options, ", ")

	return runGetExpression(fieldName, "options", rule, v, func(calculated any) *policyApi.Error {
		for _, element := range asSlice(calculated) {
			if err := validation.In(values...).Validate(element); err != nil {
				return validationError(fieldName, "options",
					fmt.Errorf("value %q is not one of [%s]", element, allowedList))
			}
		}
		return nil
	})
}

// ItemsEnforcement counts the items produced by the rule's itemSelector during
// the iteration and enforces the count to lie within an inclusive
// [min, max] interval. Max is checked as the iteration proceeds so a rule
// with an overflow fails fast; min can only be decided once the iteration
// has finished and is validated in Finalize.
//
// Note the difference between items (counts items produced by an
// itemSelector) and elements (counts the entries of an array-valued field).
type ItemsEnforcement struct {
	enforcement *policyApi.PolicyEnforcementItems
	count       int

	// broken is set by the first Exec call that discovers an inconsistent
	// bound specification (e.g. min > max). Subsequent Exec / Finalize calls
	// become no-ops so the broken-rule message is only emitted once and the
	// rule does not contribute any validation errors.
	boundsChecked bool
	broken        bool
}

func (ac *ItemsEnforcement) Exec(fieldName string, rule *ResolvedRule, _ []any, _ any) *policyApi.Error {
	if ac.checkBounds(fieldName, rule) {
		return nil
	}

	ac.count++

	i := ac.enforcement.Items.Get()
	if i == nil {
		return nil
	}
	if i.Max.IsSet() {
		maxItems := int(*i.Max.Get())
		if ac.count > maxItems {
			return validationError(fieldName, "items",
				fmt.Errorf("got at least %d items, maximum is %d", ac.count, maxItems))
		}
	}
	return nil
}

func (ac *ItemsEnforcement) Finalize(fieldName string, rule *ResolvedRule) *policyApi.Error {
	if ac.checkBounds(fieldName, rule) {
		return nil
	}
	i := ac.enforcement.Items.Get()
	if i == nil {
		return nil
	}
	if i.Min.IsSet() {
		minItems := int(*i.Min.Get())
		if ac.count < minItems {
			return validationError(fieldName, "items",
				fmt.Errorf("got %d items, minimum is %d", ac.count, minItems))
		}
	}
	return nil
}

// checkBounds validates the items min/max relation once per ItemsEnforcement
// lifetime and caches the outcome on the struct. Returns true when the
// rule is broken (caller should skip its work).
func (ac *ItemsEnforcement) checkBounds(fieldName string, rule *ResolvedRule) bool {
	if ac.boundsChecked {
		return ac.broken
	}
	ac.boundsChecked = true

	i := ac.enforcement.Items.Get()
	if i == nil {
		return false
	}
	var minItems, maxItems *int
	if i.Min.IsSet() {
		mv := int(*i.Min.Get())
		minItems = &mv
	}
	if i.Max.IsSet() {
		mv := int(*i.Max.Get())
		maxItems = &mv
	}
	ac.broken = !validateIntMinMax(fieldName, "items", rule, minItems, maxItems)
	return ac.broken
}

type RequiredEnforcement struct {
	enforcement *policyApi.PolicyEnforcementRequired
}

func (ac *RequiredEnforcement) Exec(fieldName string, rule *ResolvedRule, selectorPath []any, v any) *policyApi.Error {
	if !ac.enforcement.GetRequired() {
		return nil
	}
	formatter := NewRuntimeFormatter(fieldName, rule)
	if formatter == nil {
		return nil
	}
	if formatter.GetExpr == nil {
		LogBrokenRule(rule.rule, "field %s formatter has no get expression (required by 'required' enforcement)", fieldName)
		return nil
	}

	iter := newPathTrackingIter(formatter.GetExpr, formatter.GetPathExpr, v)

	// foundAny distinguishes "at least one occurrence was evaluated" from "the
	// iterator never yielded a value" (e.g. selector matches nothing, or the
	// get expression produced no results). In the latter case we never enter
	// the body below, so the per-occurrence nil check alone cannot report required.
	foundAny := false
	for {
		calculated, suffix, ok := iter.Next()
		if !ok {
			break
		}
		foundAny = true
		if err, ok := calculated.(error); ok {
			return policyApi.NewError(http.StatusInternalServerError,
				fmt.Sprintf("failed to evaluate get expression for field %s: %v", fieldName, err))
		}
		if calculated == nil {
			return validationError(fieldName, "required", errors.New(requiredMissingMessage(rule.rule, v, selectorPath, suffix)))
		}
	}

	if !foundAny {
		// Same validation outcome as a missing value on an occurrence, but for the
		// zero-occurrence case (empty iteration).
		return validationError(fieldName, "required", errors.New(requiredMissingMessage(rule.rule, v, selectorPath, nil)))
	}
	return nil
}

// BlockedEnforcement is the inverse of RequiredEnforcement: it fails as soon as the
// field has any value in the spec. blocked: false never fails.
type BlockedEnforcement struct {
	enforcement *policyApi.PolicyEnforcementBlocked
}

func (ac *BlockedEnforcement) Exec(fieldName string, rule *ResolvedRule, selectorPath []any, v any) *policyApi.Error {
	if !ac.enforcement.GetBlocked() {
		return nil
	}
	formatter := NewRuntimeFormatter(fieldName, rule)
	if formatter == nil {
		return nil
	}
	if formatter.GetExpr == nil {
		LogBrokenRule(rule.rule, "field %s formatter has no get expression (required by blocked enforcement)", fieldName)
		return nil
	}

	iter := newPathTrackingIter(formatter.GetExpr, formatter.GetPathExpr, v)
	for {
		calculated, suffix, ok := iter.Next()
		if !ok {
			return nil
		}
		if err, ok := calculated.(error); ok {
			return policyApi.NewError(http.StatusInternalServerError,
				fmt.Sprintf("failed to evaluate get expression for field %s: %v", fieldName, err))
		}
		if calculated != nil {
			return validationError(fieldName, "blocked", errors.New(blockedPresentMessage(rule.rule, v, selectorPath, suffix)))
		}
	}
}

// blockedPresentMessage describes why blocked validation failed.
func blockedPresentMessage(erp *policyApi.EffectiveRulesEffectiveRulesInner, v any, selectorPath, suffix []any) string {
	return withRuleOccurrenceDetails("must not be set", erp, v, selectorPath, suffix)
}

// SetEnforcement assigns a value to the field. When overwrite is true the existing
// value is replaced (PolicyEnforcementSet); when false the value is only assigned if
// the field is currently absent (PolicyEnforcementDefault).
type SetEnforcement struct {
	value     map[string]any
	overwrite bool
}

func (ac *SetEnforcement) Exec(fieldName string, rule *ResolvedRule, _ []any, v any) *policyApi.Error {
	formatter := NewRuntimeFormatter(fieldName, rule)
	if formatter == nil {
		return nil
	}
	if formatter.SetExpr == nil {
		LogBrokenRule(rule.rule, "field %s formatter has no 'set' expression (required by 'set' enforcement)", fieldName)
		return nil
	}

	if ak := formatter.ArrayKey; ak != nil {
		if _, keyGiven := ac.value[ak.Arg]; !keyGiven {
			return policyApi.NewError(http.StatusBadRequest,
				fmt.Sprintf("field %s requires key %q to be set", fieldName, ak.Arg))
		}
	}

	argsValue := map[string]any{"named": ac.value}
	iter := formatter.SetExpr.Run(map[string]any{}, argsValue)

	for {
		calculated, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := calculated.(error); ok {
			return policyApi.NewError(http.StatusInternalServerError,
				fmt.Sprintf("failed to evaluate set expression for field %s: %v", fieldName, err))
		}

		srcValue, ok := calculated.(map[string]any)
		if !ok {
			LogBrokenRule(rule.rule, "field %s set expression did not produce a map (got %T)", fieldName, calculated)
			continue
		}

		gojsonq.CleanMap(srcValue)

		// The gojq set expression emits plain float64/int for numbers, whereas
		// the target spec was loaded with decoder.UseNumber() and holds
		// json.Number. conjungo.Merge below bails on that type mismatch
		// ("types do not match: json.Number, float64"), so re-canonicalize the
		// gojq output back to json.Number before merging.
		gojsonq.NormalizeNumbers(srcValue)

		// KeyValueMap workaround: when no explicit "key" arg is provided, the set
		// expression produces a "value" placeholder that must be re-keyed to each
		// of the target map's existing keys. When a "key" arg IS provided, the JQ
		// expression already uses it directly, so CopyTargetKeys is skipped.
		//
		// When the formatter has a specRef, the set expression nests the
		// placeholder under SpecPath (e.g., specRef "labels" produces
		// `{ "labels": { "value": <v> } }`). Both srcValue and the target must
		// be walked the same way before fanning, otherwise the placeholder
		// would be looked up at the wrong level (it would be nil) and the
		// fan-out would clobber the parent of the actual map (e.g., setting
		// `metadata.labels = nil`). When SpecPath is empty (no specRef) the
		// placeholder lives at srcValue's root and v is the target map, which
		// is the original flat behavior.
		if formatter.CopyTargetKeys {
			if _, hasKey := ac.value["key"]; !hasKey {
				target, srcInner, ok := navigateForKVMapFanOut(v, srcValue, formatter.SpecPath)
				if !ok {
					LogBrokenRule(rule.rule, "field %s KeyValueMap set requires the target to be a map (got %T)", fieldName, v)
					continue
				}
				placeholder := srcInner["value"]
				for k := range target {
					srcInner[k] = placeholder
				}
				delete(srcInner, "value")
			}
		}

		options := newSetMergeOptions(ac.overwrite, formatter.ArrayKey)
		if err := conjungo.Merge(&v, srcValue, options); err != nil {
			return policyApi.NewError(http.StatusInternalServerError,
				fmt.Sprintf("failed to merge set value for field %s: %v", fieldName, err))
		}
	}

	return nil
}

// navigateForKVMapFanOut walks both srcValue and the target value v down the
// formatter's spec path so the KeyValueMap CopyTargetKeys fan-out lands at the
// correct map. When path is empty the inputs are returned as-is (the original
// flat behavior, used when the spec/item selector already narrowed to the
// labels map and no specRef was applied).
//
// Returns:
//   - target: the target map at path (whose existing keys we fan onto). If
//     v isn't a map at all, or any intermediate segment is missing /
//     non-map, target is an empty map — the placeholder is still cleared so
//     the merge becomes a no-op rather than silently writing junk like
//     `metadata.labels = nil`.
//   - srcInner: the inner srcValue map at path (where the "value" placeholder
//     currently lives). Returned only if every segment exists on srcValue;
//     otherwise the second return is false and the caller logs a broken-rule.
//   - ok: false only when the JQ-produced srcValue is structurally not what
//     the runtime expected (i.e. the set expression didn't nest under path).
func navigateForKVMapFanOut(v any, srcValue map[string]any, path []string) (map[string]any, map[string]any, bool) {
	if len(path) == 0 {
		target, tOk := v.(map[string]any)
		if !tOk {
			return nil, nil, false
		}
		return target, srcValue, true
	}

	// Walk srcValue down path. The set expression always builds this nesting,
	// so any failure here is a contract violation — surface it.
	srcInner := srcValue
	for _, seg := range path {
		next, ok := srcInner[seg].(map[string]any)
		if !ok {
			return nil, nil, false
		}
		srcInner = next
	}

	// Walk v down path. v may not be a map at all, or the path may not exist
	// yet (e.g., metadata exists but metadata.labels does not). Treat any of
	// these as "no existing keys to fan onto" and return an empty target so
	// the placeholder is still cleared and the merge becomes a no-op.
	target := map[string]any{}
	cur, vOk := v.(map[string]any)
	if !vOk {
		return target, srcInner, true
	}
	for _, seg := range path {
		next, ok := cur[seg].(map[string]any)
		if !ok {
			return target, srcInner, true
		}
		cur = next
	}
	return cur, srcInner, true
}

// runGetExpression resolves the rule's formatter and runs its get expression against spec section(s) v.
// For each value produced, onValue is invoked. Mapping issues (missing formatter/get expr)
// are logged via LogBrokenRule and the call becomes a no-op. JQ runtime errors are
// propagated as internal errors.
func runGetExpression(fieldName, enforcementType string, rule *ResolvedRule, v any, onValue func(any) *policyApi.Error) *policyApi.Error {
	formatter := NewRuntimeFormatter(fieldName, rule)
	if formatter == nil {
		return nil
	}
	if formatter.GetExpr == nil {
		LogBrokenRule(rule.rule, "field %s formatter has no get expression (required by '%s' enforcement)", fieldName, enforcementType)
		return nil
	}

	iter := formatter.GetExpr.Run(v)
	for {
		calculated, ok := iter.Next()

		// we do not perform validation if a value is not found, no value = no validation. for mandating value
		// in the spec there is the required rule that the policy can specify.
		if !ok {
			break
		}

		if err, ok := calculated.(error); ok {
			return policyApi.NewError(http.StatusInternalServerError,
				fmt.Sprintf("failed to evaluate get expression for field %s: %v", fieldName, err))
		}

		if apiErr := onValue(calculated); apiErr != nil {
			return apiErr
		}
	}
	return nil
}

// lengthRuleMessage returns the phrasing used for length validation errors.
// For strings the underlying "length" wording is used (characters); for all
// other value kinds that LengthOfValue supports (slices, maps, arrays) the
// message talks about "the number of elements".
func lengthRuleMessage(value any, comparison string, bound int) string {
	noun := "the number of elements"
	if _, isString := value.(string); isString {
		noun = "the length"
	}
	return fmt.Sprintf("%s must be %s %d", noun, comparison, bound)
}

// parseQuantity parses val as a k8s resource.Quantity string. Returns a bad-request *policyApi.Error
// if the value is not a string or is not a valid quantity.
func parseQuantity(fieldName, enforcementType string, val any) (resource.Quantity, *policyApi.Error) {
	s, ok := val.(string)
	if !ok {
		return resource.Quantity{}, policyApi.NewError(http.StatusBadRequest,
			fmt.Sprintf("field %s value for %s is not a string quantity: %v", fieldName, enforcementType, val))
	}
	q, err := resource.ParseQuantity(s)
	if err != nil {
		return resource.Quantity{}, policyApi.NewError(http.StatusBadRequest,
			fmt.Sprintf("field %s value %q is not a valid quantity: %v", fieldName, s, err))
	}
	return q, nil
}

// asSlice normalizes v into a []any for uniform iteration of scalar and slice
// in rules supporting both (for example range, which can validate range of a
// scalar field but also range on array of integers).
// JQ "null" values represent "no value" and should never be validated — policies
// use the required rule to enforce presence.
func asSlice(v any) []any {
	if v == nil {
		return nil
	}
	if s, ok := v.([]any); ok {
		result := make([]any, 0, len(s))
		for _, e := range s {
			if e != nil {
				result = append(result, e)
			}
		}
		return result
	}
	return []any{v}
}

// requiredMissingMessage describes why required validation failed.
func requiredMissingMessage(erp *policyApi.EffectiveRulesEffectiveRulesInner, v any, selectorPath, suffix []any) string {
	return withRuleOccurrenceDetails("field is required but missing from the spec", erp, v, selectorPath, suffix)
}

// withRuleOccurrenceDetails decorates base with a workload-relative location.
// When the rule's selector chain and the formatter's get expression are both
// path-eligible, emits ' (path=".spec.template.spec.containers[0].image")' —
// a JQ path the workload submitter can paste into yq/jq to find the offending
// node. Otherwise falls back to today's ' (specSelector=..., itemSelector=...,
// <param>=<value>..., name=...)' detail listing.
func withRuleOccurrenceDetails(base string, erp *policyApi.EffectiveRulesEffectiveRulesInner, v any, selectorPath, suffix []any) string {
	if path := composeJQPath(selectorPath, suffix); path != "" {
		return base + ` (path="` + path + `")`
	}
	if erp == nil {
		return base
	}
	var details []string
	if s := erp.GetSpecSelector(); s != "" {
		details = append(details, fmt.Sprintf("specSelector=%q", s))
	}
	if s := erp.GetItemSelector(); s != "" {
		details = append(details, fmt.Sprintf("itemSelector=%q", s))
	}
	if p := erp.SelectorParams; len(p) > 0 {
		keys := make([]string, 0, len(p))
		for k := range p {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			details = append(details, fmt.Sprintf("%s=%v", k, p[k]))
		}
	}
	if id := occurrenceNameHint(v); id != "" {
		details = append(details, id)
	}
	if len(details) == 0 {
		return base
	}
	return base + " (" + strings.Join(details, ", ") + ")"
}

// composeJQPath formats selectorPath followed by suffix as a dotted JQ path
// (e.g. ".spec.template.spec.containers[0].image"). Either input being nil
// signals "path unavailable" and triggers the caller's fallback to non-path
// details — emitting a partial path would mislead the workload submitter
// about where to look in the spec.
func composeJQPath(selectorPath, suffix []any) string {
	if selectorPath == nil || suffix == nil {
		return ""
	}
	var sb strings.Builder
	for _, p := range slices.Concat(selectorPath, suffix) {
		writeJQPathSegment(&sb, p)
	}
	if sb.Len() == 0 {
		return "."
	}
	return sb.String()
}

// writeJQPathSegment appends one JQ path segment for p: ".name" for
// identifier-safe object keys, '.["key"]' for other strings, and "[N]" for
// numeric array indices.
func writeJQPathSegment(sb *strings.Builder, p any) {
	if s, ok := p.(string); ok {
		if jqIdentRe.MatchString(s) {
			fmt.Fprintf(sb, ".%s", s)
		} else {
			fmt.Fprintf(sb, ".[%q]", s)
		}
		return
	}
	fmt.Fprintf(sb, "[%v]", p)
}

// jqIdentRe matches strings safe to use as a bare JQ identifier (.foo).
// Anything else is bracket-quoted (.["foo-bar"]).
var jqIdentRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// occurrenceNameHint returns a short identity for v when it is object-shaped
// and carries a name (common for container / env / metadata-tagged objects).
func occurrenceNameHint(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	if s, ok := m["name"].(string); ok && s != "" {
		return fmt.Sprintf("name=%q", s)
	}
	if meta, ok := m["metadata"].(map[string]any); ok {
		if s, ok := meta["name"].(string); ok && s != "" {
			return fmt.Sprintf("metadata.name=%q", s)
		}
	}
	return ""
}

// validationError wraps a validation error as a bad-request *policyApi.Error with the given context.
func validationError(fieldName, enforcementType string, err error) *policyApi.Error {
	return policyApi.NewError(http.StatusBadRequest,
		fmt.Sprintf("field %s failed %s validation: %v", fieldName, enforcementType, err))
}

// validateIntMinMax mirrors the policy-service-side consistency check on
// length/elements/items at runtime. If the rule is malformed (negative
// bound or min > max) the rule is logged as broken and the caller should
// skip applying it (returns false). Otherwise returns true. These checks
// are defensive: policy-service already rejects malformed rules on
// submission, but rules may reach the engine through other paths (tests,
// replay, mocks) and a broken rule should degrade gracefully rather than
// produce a nonsensical validation error.
func validateIntMinMax(fieldName, enforcementType string, rule *ResolvedRule, minVal, maxVal *int) bool {
	if minVal != nil && *minVal < 0 {
		LogBrokenRule(rule.rule, "field %s %s.min must be non-negative, got %d", fieldName, enforcementType, *minVal)
		return false
	}
	if maxVal != nil && *maxVal < 0 {
		LogBrokenRule(rule.rule, "field %s %s.max must be non-negative, got %d", fieldName, enforcementType, *maxVal)
		return false
	}
	if minVal != nil && maxVal != nil && *minVal > *maxVal {
		LogBrokenRule(rule.rule, "field %s %s.min (%d) must be less than or equal to %s.max (%d)",
			fieldName, enforcementType, *minVal, enforcementType, *maxVal)
		return false
	}
	return true
}

// validateFloatRangeBounds mirrors validate range in policy-service:
// min must be <= max when both are present, step must be > 0, and step
// must be <= max-min when all three are present. Logs the rule as broken
// and returns false if any constraint is violated.
func validateFloatRangeBounds(fieldName, enforcementType string, rule *ResolvedRule, minVal, maxVal, stepVal *float32) bool {
	if err := ValidateFloat32RangeBounds(minVal, maxVal, stepVal); err != nil {
		LogBrokenRule(rule.rule, "field %s %s: %v", fieldName, enforcementType, err)
		return false
	}
	return true
}

// validateQuantityRangeBounds is the k8s-quantity counterpart of
// validateFloatRangeBounds. It runs after the bound strings have been
// parsed so it only compares Quantity values.
func validateQuantityRangeBounds(fieldName string, rule *ResolvedRule, minQty, maxQty, stepQty *resource.Quantity) bool {
	if err := ValidateQuantityRangeBounds(minQty, maxQty, stepQty); err != nil {
		LogBrokenRule(rule.rule, "field %s quantityRange: %v", fieldName, err)
		return false
	}
	return true
}
