/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

// MappingConsistencyValidator validates that the spec selectors are consistent.
// The rule is: each field/item selector is sourced from a single tenant OR system scoped field group.
// If a field/item selector is sourced from two groups of the same scope - the mapping is not valid.
// Note that system consistency is checked while updating system mappings. The API will only
// test tenant scope elements.
type MappingConsistencyValidator struct {
	// scope is the mapping row being validated; only field-group rows for this scope are loaded.
	scope api.MappingScope

	// relevantFieldGroups holds at most one row per logical name — the row for fv.scope only.
	relevantFieldGroups map[string]dbmodel.MappingFieldGroups

	// if we have two selectors:
	//   selector1 -> fg1,fg2,fg3
	//   selector2 -> fg1,fg2,fg3
	// we only need to validate one of them, because both have the same combination of field groups.
	// we could take it to more extreme and skip validating fg1,fg2 if we already validated fg1,fg2,fg3 combo,
	// but it's not worth the complexity for now.
	fgComboSeen datastructs.Set[string]
}

func NewMappingConsistencyValidator(_ context.Context, scope api.MappingScope) *MappingConsistencyValidator {
	return &MappingConsistencyValidator{
		scope: scope,

		relevantFieldGroups: make(map[string]dbmodel.MappingFieldGroups),

		fgComboSeen: datastructs.NewSet[string](),
	}
}

// mappingSelectorToValidate is one deduplicated selector check; ref is used only in validation errors.
type mappingSelectorToValidate struct {
	specSelRef  string
	specSelInfo api.SpecSelectorsWithScopeValue
}

// ValidateMappingChange needs to validate all the selectors that are included in this mapping.
func (fv *MappingConsistencyValidator) ValidateMappingChange(ctx context.Context, tx orm.DB, specSelectors map[string]api.SpecSelectorsWithScopeValue) *api.Error {
	fieldGroupToResolve := datastructs.NewSet[string]()

	for selectorName, specSelector := range specSelectors {
		fieldGroups := datastructs.NewSet[string]()

		for _, fgName := range specSelector.FieldGroups {
			if fieldGroups.Has(fgName) {
				return api.NewValidationError(fmt.Sprintf("field group %s is multiply defined in selector %s", fgName, selectorName))
			}
			fieldGroups.Add(fgName)

			if _, ok := fv.relevantFieldGroups[fgName]; !ok {
				fieldGroupToResolve.Add(fgName)
			}
		}
	}

	err := fv.loadRelevantFieldGroups(ctx, tx, fieldGroupToResolve, "mapping consistency validation")
	if err != nil {
		return err
	}

	// make a list of all the selectors that we need to validate. no need to re-validate combinations of field
	// groups that are already validated.
	jobs := make([]mappingSelectorToValidate, 0, len(specSelectors))

	for selectorName, specSelector := range specSelectors {
		fgCombo := canonicalFieldGroupsCombo(specSelector.FieldGroups)

		if !fv.fgComboSeen.Has(fgCombo) {
			jobs = append(jobs, mappingSelectorToValidate{
				specSelRef:  selectorName,
				specSelInfo: specSelector,
			})

			fv.fgComboSeen.Add(fgCombo)
		}
	}

	// finally, validate all the necessary selectors
	return fv.validateSelectors(ctx, jobs)
}

// ValidateFieldGroupChange should validate all the selectors that reference this field group.
func (fv *MappingConsistencyValidator) ValidateFieldGroupChange(ctx context.Context, tx orm.DB, fgName string) *api.Error {
	mappingRows, err := dao.ListMappingsReferencingFieldGroup(ctx, tx, fgName)
	if err != nil {
		return err
	}
	if len(mappingRows) == 0 {
		return nil
	}

	fgToLoad := datastructs.NewSet[string]()

	for _, row := range mappingRows {
		for _, specSel := range row.SpecSelectors {
			for _, fgName := range specSel.FieldGroups {
				if _, ok := fv.relevantFieldGroups[fgName]; !ok {
					fgToLoad.Add(fgName)
				}
			}
		}
	}

	validationKind := fmt.Sprintf("consistency check after change to field group %q", fgName)

	if err := fv.loadRelevantFieldGroups(ctx, tx, fgToLoad, validationKind); err != nil {
		return err
	}

	// make a list of all the selectors that we need to validate. no need to re-validate selectors
	// that their field group combination was already validated.
	jobs := make([]mappingSelectorToValidate, 0)

	for _, row := range mappingRows {
		for selKey, specSel := range row.SpecSelectors {
			fgCombo := canonicalFieldGroupsCombo(specSel.FieldGroups)

			if !fv.fgComboSeen.Has(fgCombo) {
				jobs = append(jobs, mappingSelectorToValidate{
					specSelRef:  fmt.Sprintf("selector %s in %s/%s/%s mapping", selKey, row.GvkGroup, row.GvkKind, row.GvkVersion),
					specSelInfo: specSel,
				})

				fv.fgComboSeen.Add(fgCombo)
			}
		}
	}

	return fv.validateSelectors(ctx, jobs)
}

// creates a consistent signature for the field groups that a selector references, to  identify selectors
// that repeat the same field group combinations, which allows us to validate only one representative of them.
func canonicalFieldGroupsCombo(fieldGroups []string) string {
	if len(fieldGroups) == 0 {
		return ""
	}
	copied := slices.Clone(fieldGroups)
	slices.Sort(copied)
	return strings.Join(copied, ",")
}

func (fv *MappingConsistencyValidator) loadRelevantFieldGroups(ctx context.Context, tx orm.DB, toLoad datastructs.Set[string], validationKind string) *api.Error {
	if len(toLoad) == 0 {
		return nil
	}

	fieldGroupsInfo, err := dao.GetFieldGroupsByNames(ctx, tx, toLoad.Keys())
	if err != nil {
		return api.NewError(http.StatusInternalServerError, fmt.Sprintf("%s: fail to load field groups", validationKind), err)
	}

	for fgName, fgInfo := range fieldGroupsInfo {
		if row, ok := fgInfo[fv.scope]; ok {
			fv.relevantFieldGroups[fgName] = row
		}
		toLoad.Delete(fgName)
	}

	if len(toLoad) > 0 {
		return api.NewValidationError(fmt.Sprintf("field group(s) %s does not exist", strings.Join(toLoad.Keys(), ",")))
	}

	return nil
}

// validateSelectors runs ValidateSpecSelector for each job. Multiple jobs run in parallel (bounded pool); only
// in-memory work is concurrent—callers must finish all DB loads before invoking this.
// On the first validation error, remaining work is cancelled: goroutines waiting on the pool exit, and workers
// that have not started ValidateSpecSelector yet return without running it.
func (fv *MappingConsistencyValidator) validateSelectors(ctx context.Context, jobs []mappingSelectorToValidate) *api.Error {
	if len(jobs) == 0 {
		return nil
	}

	if len(jobs) == 1 {
		return fv.ValidateSpecSelector(jobs[0].specSelRef, &jobs[0].specSelInfo)
	}

	jobCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		firstError *api.Error

		mu sync.Mutex

		wg sync.WaitGroup
	)

	validationSlots := make(chan struct{}, selectorValidationWorkerPoolSize)

	for jobIdx := range jobs {
		job := jobs[jobIdx]

		wg.Add(1)

		go func() {
			defer wg.Done()
			select {
			case validationSlots <- struct{}{}:
			case <-jobCtx.Done():
				return
			}
			defer func() { <-validationSlots }()

			if jobCtx.Err() != nil {
				return
			}

			err := fv.ValidateSpecSelector(job.specSelRef, &job.specSelInfo)
			if err != nil {
				mu.Lock()
				if firstError == nil {
					firstError = err
					cancel()
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	return firstError
}

func (fv *MappingConsistencyValidator) ValidateSpecSelector(selectorName string, selector *api.SpecSelectorsWithScopeValue) *api.Error {
	fieldSources := make(map[string]string)
	iselSources := make(map[string]string)

	for _, fgName := range selector.FieldGroups {
		fgRow, ok := fv.relevantFieldGroups[fgName]
		if !ok {
			// field group is from different scope, no need to test it
			continue
		}

		if fgRow.Fields != nil {
			for fieldName := range fgRow.Fields {
				if fieldSources[fieldName] != "" && fieldSources[fieldName] != fgName {
					return api.NewValidationError(fmt.Sprintf("field %s is sourced from both %s and %s in specSelector %s",
						fieldName, fieldSources[fieldName], fgName, selectorName))
				}
				fieldSources[fieldName] = fgName
			}
		}

		if fgRow.ItemSelectors != nil {
			for iselName := range fgRow.ItemSelectors {
				if iselSources[iselName] != "" && iselSources[iselName] != fgName {
					return api.NewValidationError(fmt.Sprintf("itemSelector %s is sourced from both %s and %s in specSelector %s",
						iselName, iselSources[iselName], fgName, selectorName))
				}
				iselSources[iselName] = fgName
			}
		}
	}

	return nil
}
