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

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/conjungo"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/itchyny/gojq"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
)

// runAiMappingGroup is the GVK group used for the fictive workload mappings that
// back native policy translation (e.g. run.ai/Workspace, run.ai/Distributed).
// These are internal and excluded from the public list-mappings response.
const runAiMappingGroup = "run.ai"

func (ms MappingApiService) listMappings(ctx context.Context, filterBy []string) (*api.MappingListResponse, *api.Error) {
	filters, filterErr := queryapi.NewParser(dbmodel.MappingSelectors{}.FilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		return nil, api.NewValidationError(filterErr.Error())
	}

	byGvkAndScope, err := dao.ListMappings(ctx, filters)
	if err != nil {
		return nil, err
	}

	items := make([]api.Mapping, 0, len(byGvkAndScope))
	for key, byScope := range byGvkAndScope {
		gvk := byGvkAndScope.GvkFor(key)

		// The "run.ai" group holds fictive GVK mappings used internally for
		// native policy translation (e.g. run.ai/Workspace, run.ai/Distributed);
		// they are not real workload types and must not be listed.
		if convutils.Deref(gvk.Group, "") == runAiMappingGroup {
			continue
		}

		merged, err := mergeMappingRows(gvk, byScope)
		if err != nil {
			return nil, err
		}

		items = append(items, *merged)
	}

	return &api.MappingListResponse{Items: items}, nil
}

func (ms MappingApiService) getMapping(ctx context.Context, gvk api.Gvk) (*api.Mapping, *api.Error) {
	byScope, err := dao.GetMappingByGvk(ctx, gvk)
	if err != nil {
		return nil, err
	}

	return mergeMappingRows(gvk, byScope)
}

func (ms MappingApiService) createMapping(ctx context.Context, req *dbmodel.MappingSelectors, cv *MappingConsistencyValidator) (*api.Mapping, *api.Error) {
	gvk := api.Gvk{
		Group:   convutils.Ref(req.GvkGroup),
		Kind:    req.GvkKind,
		Version: req.GvkVersion,
	}

	byScope, err := dao.GetMappingByGvk(ctx, gvk)
	if err != nil {
		if err.Code != http.StatusNotFound {
			return nil, err
		}
		byScope = dao.MappingByScope{}
	}
	if _, exists := byScope[req.ScopeType]; exists {
		return nil, api.NewError(http.StatusConflict,
			fmt.Sprintf("mapping for workload %s/%s/%s already exists (%s scope)", convutils.Deref(gvk.Group, ""), gvk.Kind, gvk.Version, req.ScopeType), nil)
	}

	// System scope must not be validated against tenant definitions.
	if req.ScopeType == api.MAPPINGSCOPE_SYSTEM {
		delete(byScope, api.MAPPINGSCOPE_TENANT)
	}

	byScope[req.ScopeType] = *req

	merged, err := mergeMappingRows(gvk, byScope)
	if err != nil {
		return nil, err
	}

	if err := validateSpecSelectorMapKeys(merged.SpecSelectors); err != nil {
		return nil, err
	}

	if err := validateSpecSelectorJQExprs(req.SpecSelectors); err != nil {
		return nil, err
	}

	if cv != nil {
		err = cv.ValidateMappingChange(ctx, runaidb.DB().WithContext(ctx), req.SpecSelectors)
		if err != nil {
			return nil, err
		}
	}

	if err := dao.CreateMapping(ctx, req); err != nil {
		return nil, err
	}

	return ms.getMapping(ctx, gvk)
}

func (ms MappingApiService) deleteMapping(ctx context.Context, req *api.MappingDeletionRequest, scopeType api.MappingScope) *api.Error {
	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	if err := dao.DeleteMappingByGvkAndScope(ctx, tx, req.WorkloadType, scopeType); err != nil {
		return err
	}

	if err := dao.OnComponentDeleted(ctx, tx, dbmodel.ComponentTypeMapping, gvkComponentId(req.WorkloadType)); err != nil {
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return database.Error(database.COMMIT_DELETIONS, database.MappingSelectors).WithDetails(commitErr)
	}
	committed = true

	return nil
}

func (ms MappingApiService) updateMapping(ctx context.Context, req *api.MappingCreUpdRequest, scopeType api.MappingScope,
	cv *MappingConsistencyValidator,
) (*api.Mapping, *api.Error) {
	byScope, err := dao.GetMappingByGvk(ctx, req.WorkloadType)
	if err != nil {
		return nil, err
	}

	// tenant scope can use system scope elements, but not vice versa
	if scopeType == api.MAPPINGSCOPE_SYSTEM {
		delete(byScope, api.MAPPINGSCOPE_TENANT)
	}

	row, ok := byScope[scopeType]
	if !ok {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s scoped mapping %s/%s/%s not found",
			scopeType, convutils.Deref(req.WorkloadType.Group, ""), req.WorkloadType.Kind, req.WorkloadType.Version), nil)
	}

	if req.Description != nil {
		row.Description = req.Description
	}

	if row.SpecSelectors == nil {
		row.SpecSelectors = map[string]api.SpecSelectorsWithScopeValue{}
	}

	for key, value := range addScopeToSpecSelectors(req.SpecSelectors, scopeType) {
		row.SpecSelectors[key] = value
	}

	if err := validateSpecSelectorMapKeys(row.SpecSelectors); err != nil {
		return nil, err
	}

	if err := validateSpecSelectorJQExprs(row.SpecSelectors); err != nil {
		return nil, err
	}

	// update the field group
	byScope[scopeType] = row

	row.UpdatedAt = time.Now().UTC()
	row.UpdatedBy = auth.SubjectIdOf(ctx)

	if cv != nil {
		err = cv.ValidateMappingChange(ctx, runaidb.DB().WithContext(ctx), row.SpecSelectors)
		if err != nil {
			return nil, err
		}
	}

	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	if err := dao.SaveMapping(ctx, tx, &row); err != nil {
		return nil, err
	}

	if err := dao.OnComponentDeleted(ctx, tx, dbmodel.ComponentTypeMapping, gvkComponentId(req.WorkloadType)); err != nil {
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, database.Error(database.COMMIT_UPDATES, database.MappingSelectors).WithDetails(commitErr)
	}
	committed = true

	return ms.getMapping(ctx, req.WorkloadType)
}

func (ms MappingApiService) deleteMappingSpecSelectors(ctx context.Context, req *api.MappingDeleteSpecSelectorsReq,
	scopeType api.MappingScope,
) (*api.Mapping, *api.Error) {
	byScope, err := dao.GetMappingByGvk(ctx, req.WorkloadType)
	if err != nil {
		return nil, err
	}

	// tenant scope can use system scope elements, but not vice versa
	if scopeType == api.MAPPINGSCOPE_SYSTEM {
		delete(byScope, api.MAPPINGSCOPE_TENANT)
	}

	row, ok := byScope[scopeType]
	if !ok {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s scoped mapping %s/%s/%s not found", scopeType,
			convutils.Deref(req.WorkloadType.Group, ""), req.WorkloadType.Kind, req.WorkloadType.Version), nil)
	}

	if row.SpecSelectors == nil {
		row.SpecSelectors = map[string]api.SpecSelectorsWithScopeValue{}
	}

	for _, key := range req.SpecSelectors {
		if err := validateMappingNameKey(key, "specSelector"); err != nil {
			return nil, err
		}

		delete(row.SpecSelectors, key)
	}

	row.UpdatedAt = time.Now().UTC()
	row.UpdatedBy = auth.SubjectIdOf(ctx)

	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	if err := dao.SaveMapping(ctx, tx, &row); err != nil {
		return nil, err
	}

	if err := dao.OnComponentDeleted(ctx, tx, dbmodel.ComponentTypeMapping, gvkComponentId(req.WorkloadType)); err != nil {
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, database.Error(database.COMMIT_UPDATES, database.MappingSelectors).WithDetails(commitErr)
	}
	committed = true

	return ms.getMapping(ctx, req.WorkloadType)
}

func mergeMappingRows(gvk api.Gvk, byScope dao.MappingByScope) (*api.Mapping, *api.Error) {
	if len(byScope) == 0 {
		return nil, api.NewError(http.StatusNotFound,
			fmt.Sprintf("mapping for %s/%s/%s not found", convutils.Deref(gvk.Group, ""), gvk.Kind, gvk.Version), nil)
	}

	if len(byScope) == 1 {
		if _, ok := byScope[api.MAPPINGSCOPE_SYSTEM]; ok {
			return byScope[api.MAPPINGSCOPE_SYSTEM].DbmodelToApi(), nil
		} else {
			return byScope[api.MAPPINGSCOPE_TENANT].DbmodelToApi(), nil
		}
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
		return nil, api.NewError(http.StatusInternalServerError, "failed to merge spec selectors of system and tenant scope", err)
	}

	return systemValues.DbmodelToApi(), nil
}

func validateSpecSelectorMapKeys(m map[string]api.SpecSelectorsWithScopeValue) *api.Error {
	if m == nil {
		return nil
	}
	for k := range m {
		if err := validateMappingNameKey(k, "specSelector"); err != nil {
			return err
		}
	}
	return nil
}

// validateSpecSelectorJQExprs runs spec-selector validation that mirrors the formatter
// validation (see MappingApiService.validateFormatter):
//   - query (when non-empty) parse as JQ.
//   - every arg declared in argsInfo is referenced via $ARGS.named.<arg> in the query.
//   - every arg in argsEnum is string-like (declared as string/strings in argsInfo, or
//     referenced in the query — in which case it defaults to string).
//
// Required-ness of query is enforced upstream by OpenAPI (minLength=1).
func validateSpecSelectorJQExprs(m map[string]api.SpecSelectorsWithScopeValue) *api.Error {
	if m == nil {
		return nil
	}
	for name, sel := range m {
		queryQ, err := parseSpecSelectorJQ(name, "query", sel.Query)
		if err != nil {
			return err
		}

		jqReferencedArgs := datastructs.NewSet[string]()
		CollectArgsNamedReferences(queryQ, jqReferencedArgs)

		stringArgs := datastructs.NewSet[string]()
		for argName, argType := range sel.ArgsInfo {
			if !jqReferencedArgs.Has(argName) {
				return api.NewValidationError(fmt.Sprintf(
					"specSelector %q: argsInfo argument %q is not referenced in the query", name, argName))
			}
			if argType == api.MAPPINGFIELDTYPE_STRING || argType == api.MAPPINGFIELDTYPE_STRINGS {
				stringArgs.Add(argName)
			}
		}

		for argName := range sel.ArgsEnum {
			isString := stringArgs.Has(argName)
			if !isString {
				// if the arg is not declared in argsInfo, it is considered as a string
				isString = jqReferencedArgs.Has(argName)
			}
			if !isString {
				return api.NewValidationError(fmt.Sprintf(
					"specSelector %q: enumeration is not allowed for arg %q", name, argName))
			}
		}
	}
	return nil
}

func parseSpecSelectorJQ(specSelectorName, exprName, expr string) (*gojq.Query, *api.Error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return nil, nil
	}
	q, err := gojq.Parse(trimmed)
	if err != nil {
		return nil, api.NewValidationError(fmt.Sprintf("specSelector %q: invalid %s JQ expression: %v", specSelectorName, exprName, err))
	}
	if err := validateParsedJQSafety(q); err != nil {
		return nil, api.NewValidationError(fmt.Sprintf("specSelector %q: unsafe %s JQ expression: %v", specSelectorName, exprName, err))
	}
	return q, nil
}
