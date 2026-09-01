/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/conjungo"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

const (
	fieldGroupListDefaultLimit = 50
	fieldGroupListMaxLimit     = 500

	// selectorValidationWorkerPoolSize bounds concurrent in-memory ValidateSpecSelector runs after field groups are loaded.
	selectorValidationWorkerPoolSize = 8
)

func (ms MappingApiService) createFieldGroup(ctx context.Context, req *dbmodel.MappingFieldGroups) (*api.FieldGroup, *api.Error) {
	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	byScope, err := dao.GetFieldGroupByName(ctx, tx, req.Name)
	if err != nil {
		if err.Code != http.StatusNotFound {
			return nil, err
		}
		byScope = dao.FieldGroupByScope{}
	} else if _, exists := byScope[req.ScopeType]; exists {
		return nil, api.NewError(http.StatusConflict,
			fmt.Sprintf("field group %s already exists (%s scope)", req.Name, req.ScopeType), nil)
	}

	// System scope must not be validated against tenant definitions.
	if req.ScopeType == api.MAPPINGSCOPE_SYSTEM {
		delete(byScope, api.MAPPINGSCOPE_TENANT)
	}

	byScope[req.ScopeType] = *req

	merged, err := mergeFieldGroupRows(req.Name, byScope)
	if err != nil {
		return nil, err
	}

	if err := ms.validateFieldGroupShape(req.Name, merged.Fields, merged.ItemSelectors); err != nil {
		return nil, err
	}

	if err := ms.validateFieldGroupFormatters(ctx, req.Fields); err != nil {
		return nil, err
	}

	if err := dao.CreateFieldGroup(ctx, tx, req); err != nil {
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, database.MappingFieldGroups).WithDetails(commitErr)
	}
	committed = true

	return merged, nil
}

func (ms MappingApiService) listFieldGroups(ctx context.Context, offset, limit int) ([]api.FieldGroup, *api.Error) {
	db := runaidb.DB().WithContext(ctx)

	names, err := dao.ListFieldGroupNamesPage(ctx, db, offset, limit)
	if err != nil {
		return nil, err
	}

	byNameScope, err := dao.GetFieldGroupsByNames(ctx, db, names)
	if err != nil {
		return nil, err
	}

	items := make([]api.FieldGroup, 0, len(names))

	for _, name := range names {
		byScope := byNameScope[name]

		// it is possible that the field group has been removed b/w the first and 2nd query
		if len(byScope) == 0 {
			continue
		}

		merged, err := mergeFieldGroupRows(name, byScope)
		if err != nil {
			return nil, err
		}

		items = append(items, *merged)
	}

	return items, nil
}

func (ms MappingApiService) getFieldGroup(ctx context.Context, name string) (*api.FieldGroup, *api.Error) {
	db := runaidb.DB().WithContext(ctx)

	if strings.TrimSpace(name) == "" {
		return nil, api.NewError(http.StatusBadRequest, "field group name is required", nil)
	}

	byScope, err := dao.GetFieldGroupByName(ctx, db, name)
	if err != nil {
		return nil, err
	}

	merged, err := mergeFieldGroupRows(name, byScope)
	if err != nil {
		return nil, err
	}

	return merged, nil
}

func (ms MappingApiService) patchFieldGroup(ctx context.Context, groupName string, scope api.MappingScope,
	req *api.FieldGroupUpdateRequest, cv *MappingConsistencyValidator,
) (*api.FieldGroup, *api.Error) {
	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	byScope, err := dao.GetFieldGroupByName(ctx, tx, groupName)
	if err != nil {
		return nil, err
	}

	// tenant scope can use system scope elements, but not vice versa
	if scope == api.MAPPINGSCOPE_SYSTEM {
		delete(byScope, api.MAPPINGSCOPE_TENANT)
	}

	row, ok := byScope[scope]
	if !ok {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s scoped field group %s not found", scope, groupName), nil)
	}

	if req.Description != nil {
		row.Description = req.Description
	}

	if row.Fields == nil {
		row.Fields = map[string]api.FieldGroupFieldsWithScopeValue{}
	}

	if row.ItemSelectors == nil {
		row.ItemSelectors = map[string]api.ItemSelectorsWithScopeValue{}
	}

	for key, value := range addScopeToFields(req.Fields, scope) {
		row.Fields[key] = value
	}

	for key, value := range addScopeToItemSelectors(req.ItemSelectors, scope) {
		row.ItemSelectors[key] = value
	}

	// update the field group and get a merged set of fields and item selectors for validation
	byScope[scope] = row

	merged, err := mergeFieldGroupRows(groupName, byScope)
	if err != nil {
		return nil, err
	}

	if err := ms.validateFieldGroupShape(groupName, merged.Fields, merged.ItemSelectors); err != nil {
		return nil, err
	}

	if err := ms.validateFieldGroupFormatters(ctx, merged.Fields); err != nil {
		return nil, err
	}

	row.UpdatedAt = time.Now().UTC()
	row.UpdatedBy = auth.SubjectIdOf(ctx)

	if err := dao.SaveFieldGroup(ctx, tx, &row); err != nil {
		return nil, err
	}

	if cv != nil {
		err = cv.ValidateFieldGroupChange(ctx, tx, row.Name)
		if err != nil {
			return nil, err
		}
	}

	if err := dao.OnComponentDeleted(ctx, tx, dbmodel.ComponentTypeFieldGroup, row.ID.String()); err != nil {
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, database.Error(database.COMMIT_UPDATES, database.MappingFieldGroups).WithDetails(commitErr)
	}
	committed = true

	return ms.getFieldGroup(ctx, groupName)
}

func (ms MappingApiService) deleteFieldGroup(ctx context.Context, name string, scope api.MappingScope) *api.Error {
	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	// fetch the UUID before deleting so we can invalidate the cache afterwards
	row, err := dao.GetFieldGroupByNameAndScope(ctx, tx, name, scope)
	if err != nil {
		return err
	}

	if err := dao.DeleteFieldGroupByNameAndScope(ctx, tx, name, scope); err != nil {
		return err
	}

	if err := dao.OnComponentDeleted(ctx, tx, dbmodel.ComponentTypeFieldGroup, row.ID.String()); err != nil {
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return database.Error(database.COMMIT_DELETIONS, database.MappingFieldGroups).WithDetails(commitErr)
	}
	committed = true

	return nil
}

func (ms MappingApiService) deleteFieldGroupFields(ctx context.Context, groupName string,
	scope api.MappingScope, req *api.FieldGroupDeleteFieldsRequest,
) (*api.FieldGroup, *api.Error) {
	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	byScope, err := dao.GetFieldGroupByName(ctx, tx, groupName)
	if err != nil {
		return nil, err
	}

	// tenant scope can use system scope elements, but not vice versa
	if scope == api.MAPPINGSCOPE_SYSTEM {
		delete(byScope, api.MAPPINGSCOPE_TENANT)
	}

	row, ok := byScope[scope]
	if !ok {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s scoped field group %s not found", scope, groupName), nil)
	}

	if row.Fields == nil {
		row.Fields = map[string]api.FieldGroupFieldsWithScopeValue{}
	}

	for _, key := range req.Fields {
		if err := validateMappingNameKey(key, "field"); err != nil {
			return nil, err
		}

		if _, ok := row.Fields[key]; !ok {
			return nil, api.NewValidationError(fmt.Sprintf("field %s does not exist in field group %s", key, groupName))
		}

		delete(row.Fields, key)
	}

	byScope[scope] = row

	row.UpdatedAt = time.Now().UTC()
	row.UpdatedBy = auth.SubjectIdOf(ctx)

	if err := dao.SaveFieldGroup(ctx, tx, &row); err != nil {
		return nil, err
	}

	if err := dao.OnComponentDeleted(ctx, tx, dbmodel.ComponentTypeFieldGroup, row.ID.String()); err != nil {
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, database.Error(database.COMMIT_UPDATES, database.MappingFieldGroups).WithDetails(commitErr)
	}
	committed = true

	return ms.getFieldGroup(ctx, groupName)
}

func (ms MappingApiService) deleteFieldGroupItemSelectors(ctx context.Context, groupName string,
	scope api.MappingScope, req *api.FieldGroupDeleteItemSelectorsRequest,
) (*api.FieldGroup, *api.Error) {
	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	byScope, err := dao.GetFieldGroupByName(ctx, tx, groupName)
	if err != nil {
		return nil, err
	}

	// tenant scope can use system scope elements, but not vice versa
	if scope == api.MAPPINGSCOPE_SYSTEM {
		delete(byScope, api.MAPPINGSCOPE_TENANT)
	}

	row, ok := byScope[scope]
	if !ok {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s scoped field group %s not found", scope, groupName), nil)
	}

	if row.ItemSelectors == nil {
		row.ItemSelectors = map[string]api.ItemSelectorsWithScopeValue{}
	}

	if row.Fields == nil {
		row.Fields = map[string]api.FieldGroupFieldsWithScopeValue{}
	}

	for _, key := range req.ItemSelectors {
		if err := validateMappingNameKey(key, "itemSelector"); err != nil {
			return nil, err
		}

		if _, ok := row.ItemSelectors[key]; !ok {
			return nil, api.NewValidationError(fmt.Sprintf("item selector %s does not exist in field group %s", key, groupName))
		}
	}

	for _, key := range req.ItemSelectors {
		delete(row.ItemSelectors, key)
	}

	byScope[scope] = row

	merged, err := mergeFieldGroupRows(groupName, byScope)
	if err != nil {
		return nil, err
	}

	itemSelectorsToDelete := datastructs.NewSet(req.ItemSelectors...)

	// we must check that the field group is ok after this change (we did not remove item selectros that are in use)
	if err := validateFieldGroupShapeAfterDeletion(groupName, merged.Fields, merged.ItemSelectors, itemSelectorsToDelete); err != nil {
		return nil, err
	}

	row.UpdatedAt = time.Now().UTC()
	row.UpdatedBy = auth.SubjectIdOf(ctx)

	if err := dao.SaveFieldGroup(ctx, tx, &row); err != nil {
		return nil, err
	}

	if err := dao.OnComponentDeleted(ctx, tx, dbmodel.ComponentTypeFieldGroup, row.ID.String()); err != nil {
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, database.Error(database.COMMIT_UPDATES, database.MappingFieldGroups).WithDetails(commitErr)
	}
	committed = true

	return ms.getFieldGroup(ctx, groupName)
}

func (ms MappingApiService) validateFieldGroupShape(groupName string,
	fields map[string]api.FieldGroupFieldsWithScopeValue,
	itemSelectors map[string]api.ItemSelectorsWithScopeValue,
) *api.Error {
	return validateFieldGroupShapeAfterDeletion(groupName, fields, itemSelectors, nil)
}

// validateFieldGroupShapeAfterDeletion and validateFieldGroupShape are the same, the main different
// is specific error messages for the case that we are trying to delete an itemSelector
func validateFieldGroupShapeAfterDeletion(groupName string,
	fields map[string]api.FieldGroupFieldsWithScopeValue,
	itemSelectors map[string]api.ItemSelectorsWithScopeValue,
	itemSelectorsToDelete datastructs.Set[string],
) *api.Error {
	if fields == nil {
		fields = map[string]api.FieldGroupFieldsWithScopeValue{}
	}
	if itemSelectors == nil {
		itemSelectors = map[string]api.ItemSelectorsWithScopeValue{}
	}

	for fname, f := range fields {
		if err := validateMappingNameKey(fname, "field"); err != nil {
			return err
		}

		for _, itSelName := range f.ItemSelectors {
			if err := validateMappingNameKey(itSelName, "itemSelector"); err != nil {
				return err
			}

			if _, ok := itemSelectors[itSelName]; !ok {
				if itemSelectorsToDelete != nil && itemSelectorsToDelete.Has(itSelName) {
					return api.NewValidationError(fmt.Sprintf("cannot delete item selector %q, it is being used by field %s", itSelName, fname))
				}
				return api.NewValidationError(fmt.Sprintf("field %s references item selector %q, which is not defined in field group %s", fname, itSelName, groupName))
			}
		}
	}

	for itSelName := range itemSelectors {
		if err := validateMappingNameKey(itSelName, "itemSelector"); err != nil {
			return err
		}
	}

	return nil
}

func (ms MappingApiService) validateFieldGroupFormatters(ctx context.Context, fields map[string]api.FieldGroupFieldsWithScopeValue) *api.Error {
	if len(fields) == 0 {
		return nil
	}

	needFormatters := datastructs.NewSet[string]()

	for fname, f := range fields {
		formatterName := strings.TrimSpace(f.Formatter)
		if formatterName == "" {
			return api.NewValidationError(fmt.Sprintf("field %s: formatter name is required", fname))
		}

		if builtinFormatters.Has(formatterName) {
			continue
		}

		needFormatters.Add(formatterName)
	}

	existingFormatters, err := dao.GetFormattersByNames(ctx, needFormatters.Keys())
	if err != nil {
		return err
	}

	// Field group map keys must match formatter field keys: e.g. fields["ports"].formatter "podSpec" requires
	// formatter podSpec to define a "ports" entry in its fields map.
	for fname, f := range fields {
		formatterName := strings.TrimSpace(f.Formatter)

		if builtinFormatters.Has(formatterName) {
			continue
		}

		formatterInfo, ok := existingFormatters[formatterName]
		if !ok {
			return api.NewValidationError(fmt.Sprintf("field %s: formatter %q not found", fname, formatterName))
		}

		fmFields := formatterInfo.Fields
		if fmFields == nil {
			fmFields = map[string]api.FormatterField{}
		}

		if _, ok := fmFields[fname]; !ok {
			return api.NewValidationError(fmt.Sprintf(
				"field %q uses formatter %q, which has no field definition named %q",
				fname, formatterName, fname,
			))
		}
	}

	return nil
}

// merge system and tenant scope field group with a given name
func mergeFieldGroupRows(name string, byScope dao.FieldGroupByScope) (*api.FieldGroup, *api.Error) {
	if len(byScope) == 0 {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("field group %s not found", name), nil)
	}

	// if only one field group with the given name just return it
	if len(byScope) == 1 {
		result, ok := byScope[api.MAPPINGSCOPE_SYSTEM]
		if !ok {
			result = byScope[api.MAPPINGSCOPE_TENANT]
		}
		return result.ToApiFieldGroup(), nil
	}

	//
	// override the system values with tenant values. result will be in systemValues
	//
	systemValues := byScope[api.MAPPINGSCOPE_SYSTEM]
	tenantValues := byScope[api.MAPPINGSCOPE_TENANT]

	mergeOpts := conjungo.NewOptions()
	mergeOpts.SetKindMergeFunc(reflect.Slice, overrideSliceWhileMerging)

	err := conjungo.Merge(&systemValues, &tenantValues, mergeOpts)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to merge system and tenant field group values", err)
	}

	return systemValues.ToApiFieldGroup(), nil
}

// overrideSliceWhileMerging merges slices when merging system and tenant field-groups. Conjungo's
// default slice merge appends the source to the target, which duplicates entries when both scopes list the
// same item selectors. in our case, we can simply return the source as if tenant and source both have
// the same field, we only care about the tenant (source) fields/itemSelectors
func overrideSliceWhileMerging(target, source reflect.Value, _ *conjungo.Options) (reflect.Value, error) {
	if target.Type() != source.Type() {
		return reflect.Value{}, fmt.Errorf("slices must have same type: T: %v S: %v", target.Type(), source.Type())
	}

	return source, nil
}
