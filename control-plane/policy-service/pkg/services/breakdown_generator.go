/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

// BreakdownResult is the output of BreakdownGenerator.Generate. it pairs the
// computed breakdown with the set of components it depended on so callers can
// record cache dependencies in one step.
type BreakdownResult struct {
	Breakdown    *api.BreakdownBreakdown
	Dependencies []dbmodel.EntitiesDependency
}

// FieldRef is a reference to a field in a policy or mapping, extracted from effective rules or mappings.
type FieldRef struct {
	Field        string
	SpecSelector string
	ItemSelector string

	// RuleIndex is the position of the source rule in the policy's Rules
	// slice. Used to attribute breakdown-resolution issues back to a
	// specific rule when reporting via *ruleIssues. In non-policy
	// contexts (e.g. getMappingFieldInfo) the value is informational
	// only — callers there pass a fresh, soon-to-be-discarded collector
	// to Generate.
	RuleIndex int
}

// BreakdownGenerator resolves field references against the mapping to produce
// breakdown information (field formatting, spec selectors, item selectors).
// It is decoupled from the caller: policies pass refs from effective rules,
// future mapping breakdown passes refs derived from the mapping fields.
type BreakdownGenerator struct {
	gvk     api.Gvk
	mapping *api.Mapping
	tx      orm.DB
}

func NewBreakdownGenerator(ctx context.Context, tx orm.DB, gvk api.Gvk) (*BreakdownGenerator, *api.Error) {
	byScope, err := dao.GetMappingByGvk(ctx, gvk)
	if err != nil {
		return nil, err
	}

	merged, err := mergeMappingRows(gvk, byScope)
	if err != nil {
		return nil, err
	}

	return &BreakdownGenerator{
		gvk:     gvk,
		mapping: merged,
		tx:      tx}, nil
}

// FindSpecSelectorsForField returns a list of spec selector names that expose
// the given field. This is used when the caller wants to query all spec
// selectors without specifying one explicitly.
func (bg *BreakdownGenerator) FindSpecSelectorsForField(ctx context.Context, fieldName string) ([]string, *api.Error) {
	fgNames := datastructs.NewSet[string]()
	for _, sel := range bg.mapping.SpecSelectors {
		for _, fgName := range sel.FieldGroups {
			fgNames.Add(fgName)
		}
	}

	fgByNameScope, err := dao.GetFieldGroupsByNames(ctx, bg.tx, fgNames.Keys())
	if err != nil {
		return nil, err
	}

	mergedFGs, err := bg.mergeFieldGroups(fgByNameScope)
	if err != nil {
		return nil, err
	}

	fgWithField := datastructs.NewSet[string]()
	for fgName, fg := range mergedFGs {
		if _, hasField := fg.Fields[fieldName]; hasField {
			fgWithField.Add(fgName)
		}
	}

	var result []string
	for selName, sel := range bg.mapping.SpecSelectors {
		for _, fgName := range sel.FieldGroups {
			if fgWithField.Has(fgName) {
				result = append(result, selName)
				break
			}
		}
	}
	return result, nil
}

// Generate resolves the given field references against the loaded mapping
// and builds a BreakdownResult. Refs whose specSelector or field cannot be
// resolved are skipped, and the reason is appended to the issues collector
// so the caller (typically validateRules' shared collector) surfaces the
// brokenness instead of silently dropping the rule.
//
// issues is required: every caller must own a *ruleIssues instance and
// pass it in (use newRuleIssues(0) when running in a non-policy context
// where the findings will be discarded). A nil collector is treated as a
// programming bug and returned as an internal error rather than being
// silently tolerated, since otherwise we'd be back to the silently-skipping
// behavior that this refactor exists to remove.
func (bg *BreakdownGenerator) Generate(ctx context.Context, fieldRefs []FieldRef, issues *ruleIssues) (*BreakdownResult, *api.Error) {
	if issues == nil {
		return nil, api.NewError(http.StatusInternalServerError, "BreakdownGenerator.Generate called without an issues collector", nil)
	}

	if len(fieldRefs) == 0 {
		return nil, nil
	}

	var deps []dbmodel.EntitiesDependency

	// the mapping is identified by its GVK so that any change to the mapping
	// for this workload type invalidates the cache.
	gvkId := fmt.Sprintf("%s/%s/%s", convutils.Deref(bg.gvk.Group, ""), bg.gvk.Kind, bg.gvk.Version)

	deps = append(deps, dbmodel.EntitiesDependency{
		ComponentType: dbmodel.ComponentTypeMapping,
		ComponentId:   gvkId,
	})

	selectorLookup := make(map[string]api.SpecSelectorsWithScopeValue, len(bg.mapping.SpecSelectors))
	for name, sel := range bg.mapping.SpecSelectors {
		selectorLookup[name] = sel
	}

	fgNamesSet := datastructs.NewSet[string]()
	for _, ref := range fieldRefs {
		sel, ok := selectorLookup[ref.SpecSelector]
		if ok {
			for _, fgName := range sel.FieldGroups {
				fgNamesSet.Add(fgName)
			}
		}
	}

	//
	// load the field groups from the DB. merge the system and tenant groups to
	// get the unified set of fields for each group.
	//
	fgByNameScope, err := dao.GetFieldGroupsByNames(ctx, bg.tx, fgNamesSet.Keys())
	if err != nil {
		return nil, err
	}

	// append a dependency for every field group row that was loaded (both
	// system and tenant variants), identified by its UUID.
	for _, byScope := range fgByNameScope {
		for _, row := range byScope {
			deps = append(deps, dbmodel.EntitiesDependency{
				ComponentType: dbmodel.ComponentTypeFieldGroup,
				ComponentId:   row.ID.String(),
			})
		}
	}

	mergedFGs, err := bg.mergeFieldGroups(fgByNameScope)
	if err != nil {
		return nil, err
	}

	type variationKey struct {
		fieldName string
		fgName    string
	}

	//
	//  build the variations of the fields. the reason for variation is that two specSelectors
	//  can import to variations of the field, for example:
	//     allContainers -> fieldGroupA -> image
	//     specContainers -> fieldGroupB -> image
	//  this cannot happen in the same spec (the mapping consistency validates that), but it
	//  is perfectly ok for two spec selectors to have the same field name because the same
	//  name can repeat in different parts of the spec, each having different characteristics.
	//
	variationMap := map[variationKey]*api.BreakdownFieldVariation{}
	selectorsInVariation := map[variationKey]datastructs.Set[string]{}

	referencedSelectors := datastructs.NewSet[string]()
	referencedFormatters := datastructs.NewSet[string]()

	for _, ref := range fieldRefs {
		if issues.full() {
			break
		}

		sel, ok := selectorLookup[ref.SpecSelector]
		if !ok {
			issues.addf("%s: specSelector %q is not declared in the mapping for %s/%s/%s",
				breakdownIssuePrefix(ref), ref.SpecSelector, convutils.Deref(bg.gvk.Group, ""), bg.gvk.Kind, bg.gvk.Version)
			continue
		}

		fgName, fieldDef := bg.findFieldInFieldGroups(ref.Field, sel.FieldGroups, mergedFGs)
		if fieldDef == nil {
			issues.addf("%s: field %q is not part of any field group reachable from specSelector %q",
				breakdownIssuePrefix(ref), ref.Field, ref.SpecSelector)
			continue
		}

		referencedSelectors.Add(ref.SpecSelector)

		key := variationKey{fieldName: ref.Field, fgName: fgName}

		//
		//   if the variation already exists, just add the spec selector to it to have
		//   a complete set of spec selectors that are related to this variation
		//
		if _, exists := variationMap[key]; exists {
			selectorsInVariation[key].Add(ref.SpecSelector)
			continue
		}

		if !builtinFormatters.Has(fieldDef.Formatter) {
			referencedFormatters.Add(fieldDef.Formatter)
		}

		itemSelectors := bg.resolveItemSelectors(fieldDef.ItemSelectors, mergedFGs[fgName])

		variationMap[key] = &api.BreakdownFieldVariation{
			SpecRef:       fieldDef.SpecRef,
			Type:          fieldDef.Type,
			ItemSelectors: itemSelectors,
		}

		selectorsInVariation[key] = datastructs.NewSet([]string{ref.SpecSelector}...)
	}

	formatters, err := dao.GetFormattersByNames(ctx, referencedFormatters.Keys())
	if err != nil {
		return nil, err
	}

	// append a dependency for every non-builtin formatter that was referenced,
	// identified by its UUID.
	for _, fm := range formatters {
		deps = append(deps, dbmodel.EntitiesDependency{
			ComponentType: dbmodel.ComponentTypeFormatter,
			ComponentId:   fm.ID.String(),
		})
	}

	for key, variation := range variationMap {
		fg := mergedFGs[key.fgName]
		if fg == nil {
			continue
		}
		fieldDef, ok := fg.Fields[key.fieldName]
		if !ok {
			continue
		}
		variation.Formatter = bg.buildBreakdownFormatter(key.fieldName, fieldDef.Formatter, formatters)
		variation.PossibleEnforcements = possibleEnforcementsForField(fieldDef.Type, len(variation.ItemSelectors) > 0, variation.Formatter)
	}

	fields := make(map[string][]api.BreakdownFieldVariation)
	for key, v := range variationMap {
		v.SpecSelectors = selectorsInVariation[key].Keys()
		fields[key.fieldName] = append(fields[key.fieldName], *v)
	}

	// Expand the set of referenced selectors to include their parent dependency
	// chains. The policy engine needs the parent selectors in the breakdown to
	// resolve dependency paths (i.e. to compile the parent's query and ensure
	// the dot-separated path exists before set enforcements run). Without this,
	// EnsureDependencyPaths in gvk_compliance silently no-ops on missing
	// parents.
	selectorsForBreakdown := datastructs.NewSet[string]()
	for _, selName := range referencedSelectors.Keys() {
		bg.addSelectorWithParents(selName, selectorLookup, selectorsForBreakdown)
	}

	specSelectors := make(map[string]api.BreakdownSpecSelector, len(selectorsForBreakdown))
	for _, selName := range selectorsForBreakdown.Keys() {
		sel := selectorLookup[selName]

		specSelectors[selName] = api.BreakdownSpecSelector{
			Description: sel.Description,
			Query:       sel.Query,
			ArgsInfo:    sel.ArgsInfo,
			ArgsEnum:    sel.ArgsEnum,
			Dependency:  sel.Dependency,
		}
	}

	return &BreakdownResult{
		Breakdown: &api.BreakdownBreakdown{
			Fields:        fields,
			SpecSelectors: specSelectors,
		},
		Dependencies: deps,
	}, nil
}

func (bg *BreakdownGenerator) mergeFieldGroups(fgByNameScope dao.FieldGroupsByNameAndScope) (map[string]*api.FieldGroup, *api.Error) {
	out := make(map[string]*api.FieldGroup, len(fgByNameScope))

	for name, byScope := range fgByNameScope {
		merged, err := mergeFieldGroupRows(name, byScope)
		if err != nil {
			return nil, err
		}
		out[name] = merged
	}

	return out, nil
}

func (bg *BreakdownGenerator) findFieldInFieldGroups(
	fieldName string,
	fgNames []string,
	mergedFGs map[string]*api.FieldGroup,
) (string, *api.FieldGroupFieldsWithScopeValue) {
	for _, fgName := range fgNames {
		fg, ok := mergedFGs[fgName]
		if !ok {
			continue
		}
		if fieldDef, found := fg.Fields[fieldName]; found {
			return fgName, &fieldDef
		}
	}
	return "", nil
}

func (bg *BreakdownGenerator) resolveItemSelectors(
	selectorNames []string,
	fg *api.FieldGroup,
) map[string]api.BreakdownItemSelector {
	if len(selectorNames) == 0 || fg == nil {
		return nil
	}

	out := make(map[string]api.BreakdownItemSelector, len(selectorNames))

	for _, name := range selectorNames {
		if sel, ok := fg.ItemSelectors[name]; ok {
			out[name] = api.BreakdownItemSelector{
				Query:    sel.Query,
				ArgsInfo: sel.ArgsInfo,
				ArgsEnum: sel.ArgsEnum,
			}
		}
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

// addSelectorWithParents adds selName plus every selector in its dependency
// chain to out. Cycles are guarded by checking out before recursing — the
// mapping consistency validator should reject cyclic dependencies upstream,
// but the loop here is defensive so a bad mapping cannot hang the generator.
// Missing parents (selector declares a dependency on a name that does not
// exist in the mapping) are silently ignored here; gvk_compliance also logs
// a warning when it sees this case, and the mapping consistency validator
// is the canonical place to surface the inconsistency.
func (bg *BreakdownGenerator) addSelectorWithParents(
	selName string,
	selectorLookup map[string]api.SpecSelectorsWithScopeValue,
	out datastructs.Set[string],
) {
	if selName == "" || out.Has(selName) {
		return
	}
	sel, ok := selectorLookup[selName]
	if !ok {
		return
	}
	out.Add(selName)

	if sel.Dependency == nil || sel.Dependency.SpecSelector == nil {
		return
	}
	bg.addSelectorWithParents(*sel.Dependency.SpecSelector, selectorLookup, out)
}

// breakdownIssuePrefix renders the rule context for a *ruleIssues entry
// produced by the breakdown generator, matching the shape used by
// validateOneRule so the caller sees a consistent "rule[N] (field=…,
// specSelector=…)" prefix regardless of whether the issue comes from
// breakdown resolution or rule validation.
func breakdownIssuePrefix(ref FieldRef) string {
	return fmt.Sprintf("rule[%d] (field=%s, specSelector=%s)", ref.RuleIndex, ref.Field, ref.SpecSelector)
}

func (bg *BreakdownGenerator) buildBreakdownFormatter(fieldName string, formatterName string, formatters map[string]dbmodel.MappingFormatters) api.BreakdownFormatter {
	fm, ok := formatters[formatterName]
	if !ok {
		return api.BreakdownFormatter{Builtin: &formatterName}
	}

	fmField, ok := fm.Fields[fieldName]
	if !ok {
		return api.BreakdownFormatter{Builtin: &formatterName}
	}

	bf := api.BreakdownFormatter{
		SetExpr:  fmField.SetExpr,
		GetExpr:  fmField.GetExpr,
		ArrayKey: fmField.ArrayKey,
		ArgsInfo: fmField.ArgsInfo,
		ArgsEnum: fmField.ArgsEnum,
	}
	if fmField.SetExpr == nil && fmField.GetExpr == nil && fmField.ArrayKey == nil {
		bf.Builtin = &formatterName
	}
	return bf
}
