/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/itchyny/gojq"
)

// keep a list of builtin formatters, those that are pre-defined
var builtinFormatters = datastructs.NewSet[string]("KeyValue", "KeyValueMap")

func (ms MappingApiService) createFormatter(ctx context.Context, req *dbmodel.MappingFormatters) (*api.Formatter, *api.Error) {
	if builtinFormatters.Has(req.Name) {
		return nil, api.NewValidationError(fmt.Sprintf("formatter %s is a builtin formatter and cannot be created", req.Name))
	}

	db := runaidb.DB().WithContext(ctx)

	_, getErr := dao.GetFormatterByName(ctx, db, req.Name)
	if getErr == nil {
		return nil, api.NewError(http.StatusConflict, fmt.Sprintf("formatter %s already exists", req.Name), nil)
	}
	if getErr.Code != http.StatusNotFound {
		return nil, getErr
	}

	err := ms.validateFormatter(req.Name, req.Fields)
	if err != nil {
		return nil, err
	}

	err = dao.CreateFormatter(ctx, req)
	if err != nil {
		return nil, err
	}

	updated, err := dao.GetFormatterByName(ctx, db, req.Name)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to load newly created formatter %s: %v", req.Name, err.Message), err)
	}

	return updated.ToApiFormatter(), nil
}

func (ms MappingApiService) listFormatters(ctx context.Context, offset, limit int) (*api.FormatterListResponse, *api.Error) {
	formatters, err := dao.ListFormatters(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	response := api.FormatterListResponse{
		Items: make([]api.Formatter, 0, len(formatters)),
	}

	for _, formatter := range formatters {
		response.Items = append(response.Items, *formatter.ToApiFormatter())
	}

	return &response, nil
}

func (ms MappingApiService) getFormatter(ctx context.Context, formatterName string) (*api.Formatter, *api.Error) {
	if builtinFormatters.Has(formatterName) {
		return nil, api.NewValidationError(fmt.Sprintf("formatter %s is a builtin formatter", formatterName))
	}

	if strings.TrimSpace(formatterName) == "" {
		return nil, api.NewError(http.StatusBadRequest, "formatter name is required", nil)
	}

	formatter, err := dao.GetFormatterByName(ctx, runaidb.DB().WithContext(ctx), formatterName)
	if err != nil {
		return nil, err
	}

	return formatter.ToApiFormatter(), nil
}

func (ms MappingApiService) deleteFormatter(ctx context.Context, formatterName string) *api.Error {
	if builtinFormatters.Has(formatterName) {
		return api.NewValidationError(fmt.Sprintf("formatter %s is a builtin formatter", formatterName))
	}

	inUse, err := dao.CheckIfFormatterBeingUsed(ctx, formatterName)
	if err != nil {
		return err
	}

	if inUse {
		return api.NewError(http.StatusConflict, "formatter cannot be deleted while referenced by a field group", nil)
	}

	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	row, err := dao.GetFormatterByName(ctx, tx, formatterName)
	if err != nil {
		return err
	}

	if err := dao.DeleteFormatterByName(ctx, tx, formatterName); err != nil {
		return err
	}

	if err := dao.OnComponentDeleted(ctx, tx, dbmodel.ComponentTypeFormatter, row.ID.String()); err != nil {
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return database.Error(database.COMMIT_DELETIONS, database.MappingFormatters).WithDetails(commitErr)
	}
	committed = true

	return nil
}

// patchFormatter merges request.Fields into the stored formatter (by map key: each provided key replaces that entry),
// validates the full resulting fields map, and persists.
func (ms MappingApiService) patchFormatter(ctx context.Context, formatterName string, req *api.FormatterUpdateRequest) (*api.Formatter, *api.Error) {
	if builtinFormatters.Has(formatterName) {
		return nil, api.NewValidationError(fmt.Sprintf("formatter %s is a builtin formatter", formatterName))
	}

	if len(req.Fields) == 0 {
		return nil, api.NewValidationError("patch must include at least one field in fields")
	}

	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	row, err := dao.GetFormatterByName(ctx, tx, formatterName)
	if err != nil {
		return nil, err
	}

	if req.Description != nil {
		row.Description = req.Description
	}

	if row.Fields == nil {
		row.Fields = map[string]api.FormatterField{}
	}

	for key, field := range req.Fields {
		EnrichFormatterFieldArgsInfo(&field)
		row.Fields[key] = field
	}

	if err := ms.validateFormatter(row.Name, row.Fields); err != nil {
		return nil, err
	}

	row.UpdatedAt = time.Now().UTC()
	row.UpdatedBy = auth.SubjectIdOf(ctx)

	if err := dao.SaveFormatter(ctx, tx, row); err != nil {
		return nil, err
	}

	if err := dao.OnComponentDeleted(ctx, tx, dbmodel.ComponentTypeFormatter, row.ID.String()); err != nil {
		return nil, err
	}

	updated, err := dao.GetFormatterByName(ctx, tx, row.Name)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to load formatter %s after update: %v", row.Name, err.Message), err)
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, database.Error(database.COMMIT_UPDATES, database.MappingFormatters).WithDetails(commitErr)
	}
	committed = true

	return updated.ToApiFormatter(), nil
}

// deleteFormatterFields removes the given keys from the formatter's fields map.
// If a key is missing, it returns a validation error and does not call Save (DB unchanged).
func (ms MappingApiService) deleteFormatterFields(ctx context.Context, formatterName string, fieldKeys []string) *api.Error {
	if builtinFormatters.Has(formatterName) {
		return api.NewValidationError(fmt.Sprintf("formatter %s is a builtin formatter", formatterName))
	}

	committed := false
	tx := runaidb.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	row, err := dao.GetFormatterByName(ctx, tx, formatterName)
	if err != nil {
		return err
	}

	if row.Fields == nil {
		row.Fields = map[string]api.FormatterField{}
	}

	for _, key := range fieldKeys {
		if err := validateMappingNameKey(key, "field"); err != nil {
			return err
		}
		if _, ok := row.Fields[key]; !ok {
			return api.NewValidationError(fmt.Sprintf("field %s does not exist in formatter %s", key, formatterName))
		}
		delete(row.Fields, key)
	}

	row.UpdatedAt = time.Now().UTC()
	row.UpdatedBy = auth.SubjectIdOf(ctx)

	if err := dao.SaveFormatter(ctx, tx, row); err != nil {
		return err
	}

	if err := dao.OnComponentDeleted(ctx, tx, dbmodel.ComponentTypeFormatter, row.ID.String()); err != nil {
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return database.Error(database.COMMIT_UPDATES, database.MappingFormatters).WithDetails(commitErr)
	}
	committed = true

	return nil
}

func (ms MappingApiService) validateFormatter(formatterName string, fields map[string]api.FormatterField) *api.Error {
	if builtinFormatters.Has(formatterName) {
		return api.NewValidationError(fmt.Sprintf("formatter %s is a builtin formatter", formatterName))
	}

	for fieldName, field := range fields {
		if err := validateMappingNameKey(fieldName, "field"); err != nil {
			return err
		}

		jqReferencedArgs := datastructs.NewSet[string]()

		setExpr, err := ms.validateJQExpression(formatterName, fieldName, "setExpr", field.SetExpr)
		if err != nil {
			return err
		}

		CollectArgsNamedReferences(setExpr, jqReferencedArgs)

		getExpr, err := ms.validateJQExpression(formatterName, fieldName, "getExpr", field.GetExpr)
		if err != nil {
			return err
		}

		CollectArgsNamedReferences(getExpr, jqReferencedArgs)

		if setExpr == nil && getExpr == nil {
			return api.NewValidationError(fmt.Sprintf("formatter %s: field %s: both setExpr and getExpr are empty", formatterName, fieldName))
		}

		//
		//  check that all args in argInfo are referenced in the queries
		//
		stringArgs := datastructs.NewSet[string]()

		for argName, argType := range field.ArgsInfo {
			if !jqReferencedArgs.Has(argName) {
				return api.NewValidationError(fmt.Sprintf("formatter %s: field %s: argument %s is not referenced in the get/set queries", formatterName, fieldName, argName))
			}

			if argType == api.MAPPINGFIELDTYPE_STRING || argType == api.MAPPINGFIELDTYPE_STRINGS {
				stringArgs.Add(argName)
			}
		}

		if field.ArrayKey != nil {
			if !jqReferencedArgs.Has(field.ArrayKey.ArgName) {
				return api.NewValidationError(fmt.Sprintf("formatter %s: field %s: array key arg %s is not referenced in the get/set queries", formatterName, fieldName, field.ArrayKey.ArgName))
			}
		}

		//
		//  enumerations are only allowed for string fields
		//
		for argName := range field.ArgsEnum {
			isString := stringArgs.Has(argName)
			if !isString {
				// if the arg is not referenced in argsInfo, it is considered as a string
				isString = jqReferencedArgs.Has(argName)
			}

			if !isString {
				return api.NewValidationError(fmt.Sprintf("formatter %s: field %s: enumeration is not allowed for arg %s", formatterName, fieldName, argName))
			}
		}
	}

	return nil
}

func (ms MappingApiService) validateJQExpression(formatterName string, fieldName string, setOrGet string, expr *string) (*gojq.Query, *api.Error) {
	if expr == nil {
		return nil, nil
	}

	query := strings.TrimSpace(*expr)
	if query == "" {
		return nil, api.NewValidationError(fmt.Sprintf("formatter %s: empty %s for field %s", formatterName, setOrGet, fieldName))
	}

	parsedQuery, err := gojq.Parse(query)
	if err != nil {
		return nil, api.NewValidationError(fmt.Sprintf("formatter %s: invalid %s query for field %s: %v", formatterName, setOrGet, fieldName, err))
	}

	if err := validateParsedJQSafety(parsedQuery); err != nil {
		return nil, api.NewValidationError(fmt.Sprintf("formatter %s: unsafe %s query for field %s: %v", formatterName, setOrGet, fieldName, err))
	}

	return parsedQuery, nil
}
