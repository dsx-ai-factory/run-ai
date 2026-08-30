package gvk_compliance

import (
	"encoding/json"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	policyApi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

// SpecSelectorDependency represents a spec selector with its dependency info for path creation.
type SpecSelectorDependency struct {
	Name           string
	ParentName     string   // empty if no parent (relative to workload root)
	ElementPath    []string // path segments to create
	ParentQuery    *gojq.Code
	ParentArgsInfo map[string]policyApi.MappingFieldType
	// SelectorParams are the params from the rule that uses this selector.
	// These are passed to the parent query when ensuring paths.
	SelectorParams map[string]any
}

// selectorParamsKey creates a unique key for a (selector, params) pair for deduplication.
func selectorParamsKey(selectorName string, params map[string]any) string {
	if len(params) == 0 {
		return selectorName
	}
	// JSON-serialize params for a stable key. Marshal should never fail for
	// map[string]any from policy rules, but if it does, fall back to selector
	// name only (less deduplication, but still correct behavior).
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return selectorName
	}
	return selectorName + ":" + string(paramsJSON)
}

// CollectDependencies extracts spec selector dependencies from the breakdown for selectors
// that are used by rules with SET or DEFAULT enforcements. Each unique (selector, params)
// combination is collected separately to ensure paths are created for all param variations.
// This function recursively collects the entire dependency chain - if X depends on Y and
// Y depends on Z, all three are collected.
func CollectDependencies(
	rules []ResolvedRule,
	specSelectors map[string]policyApi.BreakdownSpecSelector,
) []SpecSelectorDependency {
	// Track unique (selector, params) pairs
	seen := datastructs.NewSet[string]()
	var deps []SpecSelectorDependency

	for _, rule := range rules {
		if !hasSetOrDefaultEnforcement(rule.rule.Enforcements) {
			continue
		}

		// Get params from this specific rule
		var params map[string]any
		if rule.rule.SelectorParams != nil {
			params = rule.rule.SelectorParams
		}

		// Recursively collect dependencies for this selector and its parent chain
		collectDependencyChain(rule.rule.SpecSelector, params, specSelectors, seen, &deps)
	}

	return deps
}

// collectDependencyChain recursively collects dependencies for a selector and its entire
// parent chain. For example, if X depends on Y and Y depends on Z, this collects X, Y, and Z.
func collectDependencyChain(
	selectorName string,
	params map[string]any,
	specSelectors map[string]policyApi.BreakdownSpecSelector,
	seen datastructs.Set[string],
	deps *[]SpecSelectorDependency,
) {
	selector, found := specSelectors[selectorName]
	if !found {
		return
	}
	if !selector.HasDependency() {
		return
	}

	dep := selector.GetDependency()
	if !dep.HasPath() {
		return
	}

	// Deduplicate by (selector, params) - same selector with same params only processed once
	key := selectorParamsKey(selectorName, params)
	if seen.Has(key) {
		return
	}
	seen.Add(key)

	ssd := SpecSelectorDependency{
		Name:           selectorName,
		ElementPath:    strings.Split(dep.GetPath(), "."),
		SelectorParams: params,
	}

	if dep.HasSpecSelector() && dep.GetSpecSelector() != "" {
		ssd.ParentName = dep.GetSpecSelector()
		// Compile parent's query for iteration
		parentSel, parentFound := specSelectors[ssd.ParentName]
		if !parentFound {
			logging.Warnf("spec selector %q declares dependency on parent %q which does not exist in mapping",
				selectorName, ssd.ParentName)
		} else if parentSel.Query == nil || *parentSel.Query == "" {
			logging.Warnf("spec selector %q declares dependency on parent %q which has no query",
				selectorName, ssd.ParentName)
		} else {
			code, err := compileJQWithArgs(*parentSel.Query)
			if err != nil {
				logging.Warnf("spec selector %q: failed to compile parent %q query: %v",
					selectorName, ssd.ParentName, err)
			} else {
				ssd.ParentQuery = code
				ssd.ParentArgsInfo = parentSel.GetArgsInfo()
			}
		}

		// Recursively collect the parent's dependencies (with the same params)
		collectDependencyChain(ssd.ParentName, params, specSelectors, seen, deps)
	}

	*deps = append(*deps, ssd)
}

// SortDependencies returns dependencies in topological order so that parent selectors
// are processed before their dependents. Dependencies with the same selector name but
// different params are treated as separate entries but share the same topological order.
func SortDependencies(deps []SpecSelectorDependency) ([]SpecSelectorDependency, error) {
	if len(deps) == 0 {
		return nil, nil
	}

	// Build dependency graph and group deps by name
	g := datastructs.NewDependenciesGraph[string]()
	depsByName := make(map[string][]SpecSelectorDependency)
	for _, dep := range deps {
		g.AddDependency(dep.Name, dep.ParentName)
		depsByName[dep.Name] = append(depsByName[dep.Name], dep)
	}

	order, err := g.FindOrder()
	if err != nil {
		return nil, err
	}

	// Collect deps in topological order
	sorted := make([]SpecSelectorDependency, 0, len(deps))
	for _, name := range order {
		sorted = append(sorted, depsByName[name]...)
	}

	return sorted, nil
}

// EnsureDependencyPaths ensures that all dependency paths exist in the manifest.
// For each dependency, it runs the parent selector (or uses manifest root if no parent),
// and ensures the element path exists under each result. Each dependency carries its
// own selector params which are passed to the parent query.
func EnsureDependencyPaths(manifest map[string]any, deps []SpecSelectorDependency) {
	for _, dep := range deps {
		ensureSingleDependency(manifest, dep)
	}
}

// ensureSingleDependency ensures paths exist for one dependency.
func ensureSingleDependency(manifest map[string]any, dep SpecSelectorDependency) {
	if dep.ParentName != "" && dep.ParentQuery == nil {
		// Parent exists but cannot be evaluated; safety measure, we do not expect this to happen
		return
	}
	if dep.ParentQuery == nil {
		// No parent - create path from manifest root
		ensurePath(manifest, dep.ElementPath)
		return
	}

	// Run parent selector to find all parent locations
	// Use the params from this specific dependency
	argsValue := map[string]any{}
	if dep.SelectorParams != nil {
		argsValue["named"] = dep.SelectorParams
	}

	iter := dep.ParentQuery.Run(manifest, argsValue)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if _, isErr := v.(error); isErr {
			continue
		}

		parent, isMap := v.(map[string]any)
		if !isMap {
			continue
		}

		ensurePath(parent, dep.ElementPath)
	}
}

// ensurePath creates all intermediate maps along pathSegments if they don't exist.
func ensurePath(parent map[string]any, pathSegments []string) {
	current := parent
	for _, segment := range pathSegments {
		if current[segment] == nil {
			current[segment] = make(map[string]any)
		}
		next, ok := current[segment].(map[string]any)
		if !ok {
			// Path exists but is not a map - don't overwrite
			return
		}
		current = next
	}
}

// hasSetOrDefaultEnforcement returns true if any enforcement is a SET or DEFAULT.
func hasSetOrDefaultEnforcement(enforcements []policyApi.PolicyRuleEnforcement) bool {
	for _, enforcement := range enforcements {
		if enforcement.PolicyEnforcementSet != nil || enforcement.PolicyEnforcementDefault != nil {
			return true
		}
	}
	return false
}
