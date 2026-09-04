/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/validation"
	"k8s.io/apimachinery/pkg/api/resource"
)

func PolicyRuleEnforcementKind(a *api.PolicyRuleEnforcement) string {
	if a == nil {
		return ""
	}
	switch {
	case a.Set != nil:
		return EnforcementSet
	case a.Default != nil:
		return EnforcementDefault
	case a.Required != nil:
		return EnforcementRequired
	case a.Blocked != nil:
		return EnforcementBlocked
	case a.Match != nil:
		return EnforcementMatch
	case a.Options != nil:
		return EnforcementOptions
	case a.Length != nil:
		return EnforcementLength
	case a.Range != nil:
		return EnforcementRange
	case a.QuantityRange != nil:
		return EnforcementQuantityRange
	case a.Elements != nil:
		return EnforcementElements
	case a.Items != nil:
		return EnforcementItems
	default:
		return ""
	}
}

// runPostValidateRuleEnforcements cross-checks string-shaped literals in the rule against
// peer constraints: each options list entry and each set/default value against match
// and length (then set/default against the full constraint set including options, etc.).
func runPostValidateRuleEnforcements(prefix string, fieldType api.MappingFieldType, formatter api.BreakdownFormatter, byType map[string][]*api.PolicyRuleEnforcement) *api.Error {
	if err := validateBlockedConflicts(prefix, byType); err != nil {
		return err
	}
	if err := validateOptionsListAgainstPeerStringConstraints(prefix, fieldType, byType); err != nil {
		return err
	}
	for _, a := range byType[EnforcementSet] {
		if err := postValidateSetOrDefaultAgainstPeers(prefix, fieldType, formatter, a, byType, EnforcementSet); err != nil {
			return err
		}
	}
	for _, a := range byType[EnforcementDefault] {
		if err := postValidateSetOrDefaultAgainstPeers(prefix, fieldType, formatter, a, byType, EnforcementDefault); err != nil {
			return err
		}
	}
	return nil
}

// validateBlockedConflicts rejects rules that pair blocked=true with any
// other enforcement type. blocked=true asserts that the field must be empty,
// so every other enforcement kind is either directly contradictory
// (required=true, set, default) or vacuous (match/options/length/range/…
// have nothing to constrain once the field is absent). Treating the
// combination as a validation error surfaces the misconfiguration at
// create/update time rather than letting it survive as dead code.
//
// The check is intentionally strict: only benign no-op peers are tolerated
// alongside blocked=true, namely required=false (an explicit assertion that
// the field is not required, which agrees with blocked) and additional
// blocked entries (handled elsewhere as a duplicate-enforcement error).
// blocked=false is a no-op and is not treated as "blocked given" here at
// all — it may coexist with any other enforcement.
func validateBlockedConflicts(prefix string, byType map[string][]*api.PolicyRuleEnforcement) *api.Error {
	hasBlocked := slices.ContainsFunc(byType[EnforcementBlocked], func(a *api.PolicyRuleEnforcement) bool {
		return a.Blocked != nil && *a.Blocked
	})
	if !hasBlocked {
		return nil
	}

	for kind, peers := range byType {
		if kind == EnforcementBlocked {
			continue
		}
		if kind == EnforcementRequired {
			// required=false is a no-op and semantically agrees with
			// blocked=true; only reject when at least one required=true
			// peer is present.
			active := slices.ContainsFunc(peers, func(a *api.PolicyRuleEnforcement) bool {
				return a.Required != nil && *a.Required
			})
			if !active {
				continue
			}
		}
		if len(peers) == 0 {
			continue
		}
		return api.NewValidationError(fmt.Sprintf(
			"%s: blocked is incompatible with %s in the same rule", prefix, kind))
	}
	return nil
}

// validateOptionsListAgainstPeerStringConstraints is the options half of
// runPostValidateRuleEnforcements: each options entry must satisfy peer match and length
// (same checks as set/default against those peers, without re-applying the options list itself).
func validateOptionsListAgainstPeerStringConstraints(prefix string, fieldType api.MappingFieldType, byType map[string][]*api.PolicyRuleEnforcement) *api.Error {
	switch fieldType {
	case api.MAPPINGFIELDTYPE_STRING, api.MAPPINGFIELDTYPE_NAMES_ARRAY, api.MAPPINGFIELDTYPE_STRINGS:
	default:
		return nil
	}

	optActs := byType[EnforcementOptions]
	if len(optActs) == 0 || optActs[0].Options == nil || len(*optActs[0].Options) == 0 {
		return nil
	}

	c, err := buildRuleEnforcementConstraints(prefix, byType)
	if err != nil {
		return err
	}
	shapeOnly := &ruleEnforcementConstraints{
		matchRes:     c.matchRes,
		effLengthMin: c.effLengthMin,
		effLengthMax: c.effLengthMax,
	}
	if len(shapeOnly.matchRes) == 0 && shapeOnly.effLengthMin == nil && shapeOnly.effLengthMax == nil {
		return nil
	}

	for _, o := range *optActs[0].Options {
		ctx := fmt.Sprintf("%s options value %q", prefix, o)
		if e := validateStringLikeAgainstConstraints(ctx, o, shapeOnly); e != nil {
			return e
		}
	}
	return nil
}

func postValidateSetOrDefaultAgainstPeers(prefix string, fieldType api.MappingFieldType, formatter api.BreakdownFormatter,
	enforcement *api.PolicyRuleEnforcement, byType map[string][]*api.PolicyRuleEnforcement, kind string,
) *api.Error {
	values := enforcement.Set
	enforcementType := EnforcementSet
	if kind == EnforcementDefault {
		values = enforcement.Default
		enforcementType = EnforcementDefault
	}
	if values == nil {
		return nil
	}

	constraints, err := buildRuleEnforcementConstraints(prefix, byType)
	if err != nil {
		return err
	}
	if !constraints.hasAnyPeerConstraint() {
		return nil
	}

	if convutils.IsStringHasValue(formatter.Builtin) {
		// validateRuleFormatter has already verified this is one of the
		// known builtins (builtinFormatters), all of which use a single
		// "value" arg on the set side; cross-enforcement checks treat them
		// uniformly.
		if len(values) != 1 {
			return nil
		}
		val, ok := values[DefaultArgName]
		if !ok {
			return nil
		}
		return validateLiteralAgainstConstraints(fmt.Sprintf("%s %s", prefix, enforcementType), val, fieldType, constraints)
	}

	if formatter.SetExpr != nil {
		setExprArgs := datastructs.NewSet[string]()
		if q, err := gojq.Parse(*formatter.SetExpr); err == nil {
			CollectArgsNamedReferences(q, setExprArgs)
		}
		for paramName, paramValue := range values {
			if !setExprArgs.Has(paramName) {
				continue
			}
			argType, hasType := formatter.ArgsInfo[paramName]
			if !hasType {
				argType = api.MAPPINGFIELDTYPE_STRING
			}
			ctx := fmt.Sprintf("%s %s parameter %q", prefix, enforcementType, paramName)
			if err := validateLiteralAgainstConstraints(ctx, paramValue, argType, constraints); err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

type ruleEnforcementConstraints struct {
	intersectedOptions datastructs.Set[string]
	matchRes           []*regexp.Regexp
	effLengthMin       *int32
	effLengthMax       *int32
	ranges             []*api.PolicyEnforcementRangeRange
	quantityRanges     []*api.PolicyEnforcementQuantityRangeQuantityRange
	elements           []*api.PolicyEnforcementElementsElements
}

func (c *ruleEnforcementConstraints) hasAnyPeerConstraint() bool {
	if len(c.intersectedOptions) > 0 {
		return true
	}
	if len(c.matchRes) > 0 {
		return true
	}
	if c.effLengthMin != nil || c.effLengthMax != nil {
		return true
	}
	if len(c.ranges) > 0 {
		return true
	}
	if len(c.quantityRanges) > 0 {
		return true
	}
	if len(c.elements) > 0 {
		return true
	}
	return false
}

func buildRuleEnforcementConstraints(prefix string, byType map[string][]*api.PolicyRuleEnforcement) (*ruleEnforcementConstraints, *api.Error) {
	out := &ruleEnforcementConstraints{
		intersectedOptions: datastructs.NewSet[string](),
	}

	for _, a := range byType[EnforcementMatch] {
		if a.Match == nil {
			continue
		}
		re, err := regexp.Compile(*a.Match)
		if err != nil {
			return nil, api.NewValidationError(fmt.Sprintf("%s: match %q: %v", prefix, *a.Match, err))
		}
		out.matchRes = append(out.matchRes, re)
	}

	for _, a := range byType[EnforcementOptions] {
		if a.Options != nil && len(*a.Options) > 0 {
			out.intersectedOptions.Add(*a.Options...)
			break
		}
	}

	var effMin, effMax *int32
	for _, a := range byType[EnforcementLength] {
		if a.Length == nil {
			continue
		}
		if a.Length.Min != nil {
			if effMin == nil || *a.Length.Min > *effMin {
				m := *a.Length.Min
				effMin = &m
			}
		}
		if a.Length.Max != nil {
			if effMax == nil || *a.Length.Max < *effMax {
				m := *a.Length.Max
				effMax = &m
			}
		}
	}
	out.effLengthMin = effMin
	out.effLengthMax = effMax
	if effMin != nil && effMax != nil && *effMin > *effMax {
		return nil, api.NewValidationError(fmt.Sprintf("%s: conflicting bounds for length enforcement (effective min %d > effective max %d)", prefix, *effMin, *effMax))
	}

	for _, a := range byType[EnforcementRange] {
		if a.Range != nil {
			out.ranges = append(out.ranges, a.Range)
		}
	}
	for _, a := range byType[EnforcementQuantityRange] {
		if a.QuantityRange != nil {
			out.quantityRanges = append(out.quantityRanges, a.QuantityRange)
		}
	}
	for _, a := range byType[EnforcementElements] {
		if a.Elements != nil {
			out.elements = append(out.elements, a.Elements)
		}
	}

	return out, nil
}

// stringOnlyPeerConstraints returns options/match/length peers for validating a
// scalar or element using validateStringLikeAgainstConstraints (string form).
func stringOnlyPeerConstraints(c *ruleEnforcementConstraints) *ruleEnforcementConstraints {
	if c == nil {
		return nil
	}
	out := &ruleEnforcementConstraints{
		intersectedOptions: datastructs.NewSet[string](),
		matchRes:           c.matchRes,
		effLengthMin:       c.effLengthMin,
		effLengthMax:       c.effLengthMax,
	}
	if len(c.intersectedOptions) > 0 {
		out.intersectedOptions.Add(c.intersectedOptions.Keys()...)
	}
	return out
}

func hasPeerStringShapeChecksForNumericLiteral(c *ruleEnforcementConstraints) bool {
	s := stringOnlyPeerConstraints(c)
	return len(s.intersectedOptions) > 0 || len(s.matchRes) > 0 || s.effLengthMin != nil || s.effLengthMax != nil
}

func numericLiteralStringForm(value interface{}) string {
	return strings.TrimSpace(fmt.Sprintf("%v", value))
}

func validateLiteralAgainstConstraints(ctx string, value interface{}, fieldType api.MappingFieldType, c *ruleEnforcementConstraints) *api.Error {
	switch fieldType {
	case api.MAPPINGFIELDTYPE_STRING, api.MAPPINGFIELDTYPE_NAMES_ARRAY:
		s := fmt.Sprintf("%v", value)
		return validateStringLikeAgainstConstraints(ctx, s, c)
	case api.MAPPINGFIELDTYPE_STRINGS:
		elems, err := arrayElementsFromPolicyValue(ctx, value, api.MAPPINGFIELDTYPE_STRING)
		if err != nil {
			return err
		}
		if err := validateElementsCount(ctx, len(elems), c); err != nil {
			return err
		}
		for i, el := range elems {
			if err := validateStringLikeAgainstConstraints(fmt.Sprintf("%s[%d]", ctx, i), fmt.Sprintf("%v", el), c); err != nil {
				return err
			}
		}
		return nil
	case api.MAPPINGFIELDTYPE_INTEGER, api.MAPPINGFIELDTYPE_NUMBER:
		for _, r := range c.ranges {
			if err := validatePolicyNumericElementAgainstRange(value, r.Min, r.Max, r.Step); err != nil {
				return api.NewValidationError(fmt.Sprintf("%s: value inconsistent with range: %v", ctx, err))
			}
		}
		if hasPeerStringShapeChecksForNumericLiteral(c) {
			s := numericLiteralStringForm(value)
			if err := validateStringLikeAgainstConstraints(ctx, s, stringOnlyPeerConstraints(c)); err != nil {
				return err
			}
		}
		if len(c.quantityRanges) > 0 || len(c.elements) > 0 {
			return api.NewValidationError(fmt.Sprintf("%s: field type %q cannot be checked against quantityRange or elements on a scalar numeric value in this rule", ctx, fieldType))
		}
		return nil
	case api.MAPPINGFIELDTYPE_INTEGERS, api.MAPPINGFIELDTYPE_NUMBERS:
		elemType := api.MAPPINGFIELDTYPE_INTEGER
		if fieldType == api.MAPPINGFIELDTYPE_NUMBERS {
			elemType = api.MAPPINGFIELDTYPE_NUMBER
		}
		elems, err := arrayElementsFromPolicyValue(ctx, value, elemType)
		if err != nil {
			return err
		}
		if err := validateElementsCount(ctx, len(elems), c); err != nil {
			return err
		}
		for i, el := range elems {
			ix := fmt.Sprintf("%s[%d]", ctx, i)
			for _, r := range c.ranges {
				if err := validatePolicyNumericElementAgainstRange(el, r.Min, r.Max, r.Step); err != nil {
					return api.NewValidationError(fmt.Sprintf("%s: value inconsistent with range: %v", ix, err))
				}
			}
		}
		if hasPeerStringShapeChecksForNumericLiteral(c) {
			sc := stringOnlyPeerConstraints(c)
			for i, el := range elems {
				ix := fmt.Sprintf("%s[%d]", ctx, i)
				s := numericLiteralStringForm(el)
				if err := validateStringLikeAgainstConstraints(ix, s, sc); err != nil {
					return err
				}
			}
		}
		if len(c.quantityRanges) > 0 {
			return api.NewValidationError(fmt.Sprintf("%s: field type %q cannot be checked against quantityRange on an array numeric value in this rule", ctx, fieldType))
		}
		return nil
	case api.MAPPINGFIELDTYPE_QUANTITY:
		strVal := fmt.Sprintf("%v", value)
		q, err := resource.ParseQuantity(strVal)
		if err != nil {
			return api.NewValidationError(fmt.Sprintf("%s: expected quantity, got %q: %v", ctx, strVal, err))
		}
		for _, qr := range c.quantityRanges {
			minQ, maxQ, stepQ, err := parseQuantityRangeBounds(qr)
			if err != nil {
				return api.NewValidationError(fmt.Sprintf("%s: %v", ctx, err))
			}
			if err := validatePolicyQuantityRangeBounds(minQ, maxQ, stepQ); err != nil {
				return api.NewValidationError(fmt.Sprintf("%s: %v", ctx, err))
			}
			if err := validatePolicyQuantityValueAgainstRange(q, minQ, maxQ, stepQ); err != nil {
				return api.NewValidationError(fmt.Sprintf("%s: value inconsistent with quantityRange: %v", ctx, err))
			}
		}
		if len(c.intersectedOptions) > 0 || len(c.matchRes) > 0 || c.effLengthMin != nil || c.effLengthMax != nil || len(c.ranges) > 0 || len(c.elements) > 0 {
			return api.NewValidationError(fmt.Sprintf("%s: field type %q is not aligned with the given enforcements in this rule", ctx, fieldType))
		}
		return nil
	default:
		if c.hasAnyPeerConstraint() {
			return api.NewValidationError(fmt.Sprintf("%s: field type %q has no cross-enforcement consistency checks for the given constraints", ctx, fieldType))
		}
		return nil
	}
}

func parseQuantityRangeBounds(r *api.PolicyEnforcementQuantityRangeQuantityRange) (minQ, maxQ, stepQ *resource.Quantity, err error) {
	if r.Min != nil {
		q, e := resource.ParseQuantity(*r.Min)
		if e != nil {
			return nil, nil, nil, fmt.Errorf("quantityRange.min: %w", e)
		}
		minQ = &q
	}
	if r.Max != nil {
		q, e := resource.ParseQuantity(*r.Max)
		if e != nil {
			return nil, nil, nil, fmt.Errorf("quantityRange.max: %w", e)
		}
		maxQ = &q
	}
	if r.Step != nil {
		q, e := resource.ParseQuantity(*r.Step)
		if e != nil {
			return nil, nil, nil, fmt.Errorf("quantityRange.step: %w", e)
		}
		stepQ = &q
	}
	return minQ, maxQ, stepQ, nil
}

// formatAllowedOptionsListForError renders the allowed-options list for an error
// message. The input is a Set's Keys() (map-backed, non-deterministic), so we sort
// to keep the rendered list stable across runs.
func formatAllowedOptionsListForError(allowed []string) string {
	if len(allowed) == 0 {
		return "[]"
	}
	sorted := make([]string, len(allowed))
	copy(sorted, allowed)
	sort.Strings(sorted)
	return "[" + strings.Join(sorted, ", ") + "]"
}

func validateStringLikeAgainstConstraints(ctx, s string, c *ruleEnforcementConstraints) *api.Error {
	if len(c.intersectedOptions) > 0 {
		if !c.intersectedOptions.Has(s) {
			return api.NewValidationError(fmt.Sprintf(
				"%s: value %q is inconsistent with options: must be one of %s",
				ctx, s, formatAllowedOptionsListForError(c.intersectedOptions.Keys())))
		}
	}
	for _, re := range c.matchRes {
		if !re.MatchString(s) {
			return api.NewValidationError(fmt.Sprintf("%s: value %q does not match pattern %q", ctx, s, re.String()))
		}
	}
	if c.effLengthMin != nil {
		rule := validation.Length(int(*c.effLengthMin), 0).Error(lengthRuleMessageForPolicy(s, "no less than", int(*c.effLengthMin)))
		if err := rule.Validate(s); err != nil {
			return api.NewValidationError(fmt.Sprintf("%s: %v", ctx, err))
		}
	}
	if c.effLengthMax != nil {
		rule := validation.Length(0, int(*c.effLengthMax)).Error(lengthRuleMessageForPolicy(s, "no more than", int(*c.effLengthMax)))
		if err := rule.Validate(s); err != nil {
			return api.NewValidationError(fmt.Sprintf("%s: %v", ctx, err))
		}
	}
	if len(c.ranges) > 0 || len(c.quantityRanges) > 0 || len(c.elements) > 0 {
		return api.NewValidationError(fmt.Sprintf("%s: string value cannot satisfy range, quantityRange, or elements constraints in the same rule", ctx))
	}
	return nil
}

func lengthRuleMessageForPolicy(value string, comparison string, bound int) string {
	return fmt.Sprintf("the length must be %s %d", comparison, bound)
}

func validateElementsCount(ctx string, n int, c *ruleEnforcementConstraints) *api.Error {
	for _, e := range c.elements {
		if e.Min != nil && n < int(*e.Min) {
			return api.NewValidationError(fmt.Sprintf("%s: array length %d is below elements.min %d", ctx, n, *e.Min))
		}
		if e.Max != nil && n > int(*e.Max) {
			return api.NewValidationError(fmt.Sprintf("%s: array length %d exceeds elements.max %d", ctx, n, *e.Max))
		}
	}
	return nil
}

func arrayElementsFromPolicyValue(ctx string, value interface{}, elemType api.MappingFieldType) ([]interface{}, *api.Error) {
	arr, ok := value.([]interface{})
	if !ok {
		strVal := fmt.Sprintf("%v", value)
		if strings.Contains(strVal, ",") {
			parts := strings.Split(strVal, ",")
			out := make([]interface{}, len(parts))
			for i, part := range parts {
				tok := strings.TrimSpace(part)
				if err := validateValueForType(fmt.Sprintf("%s[%d]", ctx, i), tok, elemType); err != nil {
					return nil, err
				}
				out[i] = tok
			}
			return out, nil
		}
		return nil, api.NewValidationError(fmt.Sprintf("%s: expected array, got %T", ctx, value))
	}
	for i, elem := range arr {
		if err := validateValueForType(fmt.Sprintf("%s[%d]", ctx, i), elem, elemType); err != nil {
			return nil, err
		}
	}
	return arr, nil
}
