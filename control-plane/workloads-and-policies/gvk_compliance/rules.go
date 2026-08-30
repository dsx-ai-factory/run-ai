package gvk_compliance

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	policyApi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

// there are three cycles of running the policy rules
// 1st cycle (pre-Set) runs the maxItems/minItems to check the number of workload items before the policy adds/remove them.
// 2nd cycle (set) runs the set/default enforcements of the policy.
// 3rd cycle (post-Set) runs all the remaining validations that validate the combined policy+workload values.
const (
	CyclePreSet = iota
	CycleSetValues
	CyclePostSet
)

var AllCycles = []int{CyclePreSet, CycleSetValues, CyclePostSet}

type ResolvedRule struct {
	rule          *policyApi.EffectiveRulesEffectiveRulesInner
	variation     *policyApi.BreakdownFieldVariation
	compiledQuery *gojq.Code

	// pathQuery is the compiled path() form of compiledQuery. It runs in
	// lockstep with compiledQuery and yields each match's JQ path
	// (e.g. ["spec","template","spec","containers",0,"image"]) so validation
	// enforcements can point error messages at the offending workload field —
	// rendered as .spec.template.spec.containers[0].image.
	//
	// Nil when gojq's path() refuses to compile this rule's selector chain
	// (e.g. an item selector built on the // alternative operator, which
	// path() can't reason about). Enforcements then receive a nil selectorPath
	// and the error message falls back to naming the rule's selectors —
	// for example, (specSelector="allContainers", itemSelector="allPorts").
	pathQuery *gojq.Code
}

func ResolveRules(pol *policyApi.GetGvkPolicyResp) []ResolvedRule {
	breakdown := pol.Breakdown.Get()
	if breakdown == nil || breakdown.Fields == nil || breakdown.SpecSelectors == nil {
		logging.Warnf("policy %s has no breakdown, skipping all rules", pol.Id)
		return nil
	}

	specSelectors := *breakdown.SpecSelectors
	var resolved []ResolvedRule

	for i := range pol.EffectiveRules {
		rule := &pol.EffectiveRules[i]

		// find the (field + specSelector) variation in the breakdown
		variation := findFieldVariation(rule, breakdown)
		if variation == nil {
			continue
		}

		// WAIT_FOR_OFER need also to support selector aliases
		specSelector, found := specSelectors[rule.SpecSelector]
		if !found {
			LogBrokenRule(rule, "spec selector %s not found in breakdown", rule.SpecSelector)
			continue
		}

		if specSelector.Query == nil {
			LogBrokenRule(rule, "spec selector %s has no query", rule.SpecSelector)
			continue
		}

		// for itemized fields, we create a query composed of the comibination of spec selector + item selector queries
		var itemSelectorQuery string
		if convutils.IsStringHasValue(rule.ItemSelector.Get()) {
			itemSelectorName := *rule.ItemSelector.Get()
			if variation.ItemSelectors == nil {
				LogBrokenRule(rule, "field %s variation has no item selectors but rule requires %s", rule.Field, itemSelectorName)
				continue
			}
			itemSel, ok := (*variation.ItemSelectors)[itemSelectorName]
			if !ok {
				LogBrokenRule(rule, "item selector %s not found in field %s variation", itemSelectorName, rule.Field)
				continue
			}
			if itemSel.Query != nil {
				itemSelectorQuery = *itemSel.Query
			} else {
				LogBrokenRule(rule, "item selector %s in field %s variation has no query", itemSelectorName, rule.Field)
				continue
			}
		}

		combinedQuery := strings.TrimSpace(*specSelector.Query)
		if itemSelectorQuery != "" {
			combinedQuery += " | " + strings.TrimSpace(itemSelectorQuery)
		}

		compiledCode, compileErr := compileJQWithArgs(combinedQuery)
		if compileErr != nil {
			LogBrokenRule(rule, "failed to compile selector query: %v", compileErr)
			continue
		}

		resolved = append(resolved, ResolvedRule{
			rule:          rule,
			variation:     variation,
			compiledQuery: compiledCode,
			pathQuery:     compilePathExpr(combinedQuery, compileJQWithArgs),
		})
	}

	return resolved
}

func RunRules(_ context.Context, manifest map[string]any, rules []ResolvedRule, maxErrors int, cycles []int) []*policyApi.Error {
	return RunRulesWithDependencies(manifest, rules, nil, maxErrors, cycles)
}

// RunRulesWithDependencies runs policy rules with spec selector dependency resolution.
// When specSelectors is provided and cycles includes CycleSetValues, this ensures
// that dependency paths exist in the manifest before running SET/DEFAULT enforcements.
func RunRulesWithDependencies(manifest map[string]any, rules []ResolvedRule, specSelectors map[string]policyApi.BreakdownSpecSelector, maxErrors int, cycles []int) []*policyApi.Error {
	var errors []*policyApi.Error

	// Ensure dependency paths before the SET cycle
	dependenciesEnsured := false

	for _, cycle := range cycles {
		// Ensure dependencies just before the SET cycle
		if cycle == CycleSetValues && !dependenciesEnsured && specSelectors != nil {
			if depErr := ensureDependenciesForRules(manifest, rules, specSelectors); depErr != nil {
				errors = append(errors, depErr)
				// Continue with other cycles - dependency errors don't block validation
			}
			dependenciesEnsured = true
		}

		for i := range rules {
			if maxErrors > 0 && len(errors) >= maxErrors {
				return errors
			}
			rule := &rules[i]
			if apiErr := RunRuleCycle(manifest, rule, cycle); apiErr != nil {
				errors = append(errors, policyApi.NewError(apiErr.Code, apiErr.Message))
			}
		}
	}

	return errors
}

// ensureDependenciesForRules builds and ensures all spec selector dependency paths.
// Returns an error if there's a cycle in the dependency graph (mapping configuration error).
func ensureDependenciesForRules(manifest map[string]any, rules []ResolvedRule, specSelectors map[string]policyApi.BreakdownSpecSelector) *policyApi.Error {
	deps := CollectDependencies(rules, specSelectors)
	if len(deps) == 0 {
		return nil
	}

	sorted, err := SortDependencies(deps)
	if err != nil {
		return policyApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("spec selector dependency cycle detected: %v", err))
	}

	EnsureDependencyPaths(manifest, sorted)
	return nil
}

// RunRuleCycle runs a single rule for the given cycle against the workload.
// It returns the first *policyApi.Error encountered (if any); once an error is
// produced, the remaining items/enforcements of this rule in this cycle are
// skipped, and the caller moves on to the next rule.
func RunRuleCycle(manifest map[string]any, rule *ResolvedRule, cycle int) *policyApi.Error {
	relevantEnforcements := enforcementsOfCycle(cycle, *rule)
	if len(relevantEnforcements) == 0 {
		return nil
	}

	argsValue := map[string]any{}
	if rule.rule.SelectorParams != nil {
		argsValue["named"] = rule.rule.SelectorParams
	}

	fieldName := rule.rule.Field

	// Build concrete enforcements once per rule/cycle so stateful enforcements
	// (e.g. min, max) can accumulate across items.
	concreteEnforcements := make([]Enforcement, 0, len(relevantEnforcements))

	for _, enforcement := range relevantEnforcements {
		concrete := newEnforcement(enforcement)
		if concrete == nil {
			LogBrokenRule(rule.rule, "unknown enforcement type on field %s", fieldName)
			continue
		}
		concreteEnforcements = append(concreteEnforcements, concrete)
	}

	iter := newPathTrackingIter(rule.compiledQuery, rule.pathQuery, manifest, argsValue)
	for {
		v, selectorPath, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return policyApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to apply policy: %v", err.Error()))
		}

		for _, enforcement := range concreteEnforcements {
			if apiErr := enforcement.Exec(fieldName, rule, selectorPath, v); apiErr != nil {
				return apiErr
			}
		}
	}

	// After the item iteration each enforcement may finalize (e.g. MinItems
	// validates the accumulated count against its minimum).
	for _, enforcement := range concreteEnforcements {
		finalizable, ok := enforcement.(FinalizableEnforcement)
		if !ok {
			continue
		}
		if apiErr := finalizable.Finalize(fieldName, rule); apiErr != nil {
			return apiErr
		}
	}

	return nil
}

func cycleOf(enforcement *policyApi.PolicyRuleEnforcement) int {
	if enforcement.PolicyEnforcementItems != nil {
		return 0
	}
	if enforcement.PolicyEnforcementDefault != nil || enforcement.PolicyEnforcementSet != nil {
		return 1
	}
	return 2
}

func enforcementsOfCycle(cycle int, rule ResolvedRule) []*policyApi.PolicyRuleEnforcement {
	result := make([]*policyApi.PolicyRuleEnforcement, 0, len(rule.rule.Enforcements))

	for i := range rule.rule.Enforcements {
		enforcement := &rule.rule.Enforcements[i]
		cycleOfEnforcement := cycleOf(enforcement)

		if cycleOfEnforcement == cycle {
			result = append(result, enforcement)
		}
	}

	return result
}

func findFieldVariation(rule *policyApi.EffectiveRulesEffectiveRulesInner, breakdown *policyApi.BreakdownBreakdown) *policyApi.BreakdownFieldVariation {
	if breakdown.Fields == nil {
		LogBrokenRule(rule, "breakdown has no fields")
		return nil
	}

	variations, fieldFound := (*breakdown.Fields)[rule.Field]
	if !fieldFound {
		LogBrokenRule(rule, "field %s not found in breakdown", rule.Field)
		return nil
	}

	for i := range variations {
		for _, specSelector := range variations[i].SpecSelectors {
			if specSelector == rule.SpecSelector {
				return &variations[i]
			}
		}
	}

	LogBrokenRule(rule, "no variation of field %s matches spec selector %s", rule.Field, rule.SpecSelector)
	return nil
}

func compileJQWithArgs(jq string) (*gojq.Code, *policyApi.Error) {
	parsed, err := gojq.Parse(jq)
	if err != nil {
		return nil, policyApi.NewError(http.StatusInternalServerError, fmt.Sprintf("invalid JQ in composed selector: %v", err))

	}

	code, err := gojq.Compile(parsed, gojq.WithVariables([]string{"$ARGS"}))
	if err != nil {
		return nil, policyApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to compile composed selector: %v", err))
	}

	return code, nil
}

func compileJQNoArgs(jq string) (*gojq.Code, *policyApi.Error) {
	parsed, err := gojq.Parse(jq)
	if err != nil {
		return nil, policyApi.NewError(http.StatusInternalServerError, fmt.Sprintf("invalid JQ in composed selector: %v", err))
	}

	code, err := gojq.Compile(parsed)
	if err != nil {
		return nil, policyApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to compile composed selector: %v", err))
	}

	return code, nil
}

// compilePathExpr compiles path(jq) — the path-tracking variant of jq —
// using the supplied compiler (with or without $ARGS). Returns nil when
// jq isn't path-eligible; callers fall back to non-path output.
func compilePathExpr(jq string, compile func(string) (*gojq.Code, *policyApi.Error)) *gojq.Code {
	code, _ := compile("path(" + rewriteForPath(jq) + ")")
	return code
}

// rewriteForPath converts the safe-iter idiom (.foo // [])[] to the
// equivalent .foo[]? so gojq's path() can compile it.
func rewriteForPath(jq string) string {
	return safeArrayIterRe.ReplaceAllString(jq, `$1[]?`)
}

// safeArrayIterRe matches the jq idiom ( PATH // [] )[] — used to safely
// iterate a possibly-missing array. PATH is captured as group 1 so
// rewriteForPath can splice it into the .foo[]? replacement.
//
// Pattern breakdown:
//
//	\(\s*               opening (
//	(\.[^()]*?)         PATH — group 1, starts with . and contains no parens
//	\s*//\s*\[\s*\]     // [] alternative (default to empty array)
//	\s*\)\s*            closing )
//	\[\s*\]             [] iterator
var safeArrayIterRe = regexp.MustCompile(`\(\s*(\.[^()]*?)\s*//\s*\[\s*\]\s*\)\s*\[\s*\]`)

// pathTrackingIter pairs a value iterator with its path() twin so each Next
// returns an occurrence's value plus its workload-relative JQ path. path is
// nil when gojq can't compute it — compile failure (see ResolvedRule.pathQuery)
// or a runtime panic on idioms like to_entries[]; the latter disables only
// the path side so values keep flowing.
type pathTrackingIter struct {
	values        gojq.Iter
	paths         gojq.Iter
	pathsDisabled bool // set after a panic in paths.Next; from then on path is always nil
}

func newPathTrackingIter(value, path *gojq.Code, input any, args ...any) *pathTrackingIter {
	it := &pathTrackingIter{values: value.Run(input, args...)}
	if path != nil {
		it.paths = path.Run(input, args...)
	}
	return it
}

// Next returns the next occurrence's value alongside its workload-relative
// JQ path. path is nil when path tracking isn't available; ok is false
// when there are no more values.
func (it *pathTrackingIter) Next() (value any, path []any, ok bool) {
	value, ok = it.values.Next()
	if !ok {
		return
	}
	path = it.safeNextPath()
	return
}

// safeNextPath advances the path iter with panic recovery scoped to the
// path side only — a gojq path() panic (e.g. on to_entries[]) marks
// pathsDisabled and yields nil. Panics from the value iter, by contrast,
// must propagate so real engine bugs aren't masked.
func (it *pathTrackingIter) safeNextPath() (path []any) {
	if it.paths == nil || it.pathsDisabled {
		return
	}
	defer func() {
		if recover() != nil {
			it.pathsDisabled = true
		}
	}()
	v, _ := it.paths.Next()
	path, _ = v.([]any)
	return
}

// BrokenRuleHook is an optional observer for broken-rule events. When set,
// LogBrokenRule delegates to it instead of writing to the logger. Tests use
// this to capture broken-rule messages in memory; production leaves it nil.
var BrokenRuleHook func(rule *policyApi.EffectiveRulesEffectiveRulesInner, msg string)

func LogBrokenRule(rule *policyApi.EffectiveRulesEffectiveRulesInner, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if BrokenRuleHook != nil {
		BrokenRuleHook(rule, msg)
		return
	}
	logging.Warnf("broken policy rule [field=%s, specSelector=%s]: %s", rule.Field, rule.SpecSelector, msg)
}
