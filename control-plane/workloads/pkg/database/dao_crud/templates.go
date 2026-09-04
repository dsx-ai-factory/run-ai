/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"context"
	"errors"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"

	"github.com/google/uuid"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func GetTemplateById(ctx context.Context, tenantId uint, templateId string, permittedScopes *openapi.PermittedScopes) (*models_crud.LoadedTemplate, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to get template '%s'", templateId), nil)

	db := runaidb.DB().WithContext(ctx)

	templateUuid, uuidErr := uuid.Parse(templateId)
	if uuidErr != nil {
		return nil, api.NewValidationError(fmt.Sprintf("invalid uuid: %v", uuidErr))
	}

	rec := models_crud.LoadedTemplate{
		ID:     templateUuid,
		Assets: make([]models_crud.TemplateAssets, 0),
	}

	result := db.Model(rec).Where(rec)

	result = result.Where(fmt.Sprintf("(%s = 0 or %s = ?)", models_crud.TemplatesColumnTenantId, models_crud.TemplatesColumnTenantId), tenantId)

	result = filterOutForkedSystemTemplates(result, tenantId)

	if permittedScopes != nil {
		var err error
		result, err = applyFilters(ctx, result, tenantId, permittedScopes, nil, []queryapi.Filter{})
		if err != nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to load template", err)
		}
	}

	rows, err := result.Scan(&rec)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: load template: %v", dbErr.Message, err)
		return nil, dbErr.WithDetailsFromDbResult(rows, err)
	}
	if rows == 0 {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("template '%s' does not exist", templateId), nil)
	}

	//
	//   locate the assets related to this template
	//
	assetsLookup := models_crud.TemplateAssets{
		TenantId:   tenantId,
		TemplateId: templateId,
	}

	err = db.Model(assetsLookup).Where(assetsLookup).Find(&rec.Assets)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: load related assets: %v", dbErr.Message, err)
		return nil, dbErr.WithDetails(err)
	}

	return &rec, nil
}

func GetRawTemplateById(ctx context.Context, tenantId uint, templateId string, permittedScopes *openapi.PermittedScopes) (*models_crud.Template, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to get template '%s'", templateId), nil)

	db := runaidb.DB().WithContext(ctx)

	templateUuid, uuidErr := uuid.Parse(templateId)
	if uuidErr != nil {
		return nil, api.NewValidationError(fmt.Sprintf("invalid uuid: %v", uuidErr))
	}

	rec := models_crud.Template{
		ID: templateUuid,
	}

	result := db.Model(rec).Where(rec)

	result = result.Where(fmt.Sprintf("(%s = 0 or %s = ?)", models_crud.TemplatesColumnTenantId, models_crud.TemplatesColumnTenantId), tenantId)

	result = filterOutForkedSystemTemplates(result, tenantId)

	if permittedScopes != nil {
		var err error
		result, err = applyFilters(ctx, result, tenantId, permittedScopes, nil, []queryapi.Filter{})
		if err != nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to load template", err)
		}
	}

	rows, err := result.Scan(&rec)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return nil, dbErr.WithDetailsFromDbResult(rows, err)
	}
	if rows == 0 {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("template '%s' does not exist", templateId), nil)
	}

	return &rec, nil
}

func CreateTemplate(ctx context.Context, rec *models_crud.Template, templateAssets []models_crud.TemplateAssets) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "failed to create template", nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	rows, err := tx.Model(rec).Create(rec)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: create templates record: %v", dbErr.Message, err)
		return dbErr.WithDetailsFromDbResult(rows, err)
	}

	dbErr = updateTemplatesAssets(tx, rec.TenantId, rec.ID.String(), templateAssets)
	if dbErr != nil {
		return dbErr
	}

	err = tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: commit transaction: %v", dbErr.Message, err)
		return dbErr.WithDetails(err)
	}

	committed = true

	return nil
}

func UpdateTemplate(ctx context.Context, rec *models_crud.Template, templateAssets []models_crud.TemplateAssets) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to update template '%s'", rec.ID.String()), nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	recId := models_crud.Template{
		ID:           rec.ID,
		TemplateMeta: models_crud.TemplateMeta{},
	}

	rows, err := tx.Model(rec).Where(recId).Save(&rec)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: save templates record: %v", dbErr.Message, err)
		return dbErr.WithDetailsFromDbResult(rows, err)
	}

	dbErr = updateTemplatesAssets(tx, rec.TenantId, rec.ID.String(), templateAssets)
	if dbErr != nil {
		return dbErr
	}

	err = tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: commit transaction: %v", dbErr.Message, err)
		return dbErr.WithDetails(err)
	}
	committed = true

	return nil
}

func DeleteTemplateById(ctx context.Context, tenantId uint, templateId uuid.UUID) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to delete template '%s'", templateId), nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	templateKey := &models_crud.Template{
		ID: templateId,
		TemplateMeta: models_crud.TemplateMeta{
			TenantId: tenantId,
		},
	}

	rows, err := tx.Model(&templateKey).Where(templateKey).Delete(allRecords)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: delete templates record: %v", dbErr.Message, err)
		return dbErr.WithDetailsFromDbResult(rows, err)
	} else if rows == 0 {
		return api.NewError(http.StatusNotFound, fmt.Sprintf("template '%s' does not exist", templateId), nil)
	}

	dbErr = updateTemplatesAssets(tx, tenantId, templateId.String(), nil)
	if dbErr != nil {
		return dbErr
	}

	dbErr = deletePolicyComplianceOfTemplate(ctx, tx, templateId.String())
	if dbErr != nil {
		return dbErr
	}

	err = tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: commit transaction: %v", dbErr.Message, err)
		return dbErr.WithDetails(err)
	}
	committed = true

	return nil
}

func deleteTemplatesByScope(ctx context.Context, tenantId string, scopeColumn string, scopeValue string) error {
	db := runaidb.DB().WithContext(ctx)

	tx := db.Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	var templateIds []string
	err := tx.Table(models_crud.TemplatesTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.TemplatesColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", scopeColumn), scopeValue).
		Pluck(models_crud.TemplatesColumnId, &templateIds)
	if err != nil {
		return fmt.Errorf("failed to find templates for %s=%s in tenant %s: %w", scopeColumn, scopeValue, tenantId, err)
	}

	if len(templateIds) == 0 {
		switch scopeColumn {
		case models_crud.TemplatesColumnProjectId, models_crud.TemplatesColumnTenantId:
			// proceed — the deleted scope may still have compliance cache rows for ancestor-scoped templates
		default:
			return nil
		}
	}

	if len(templateIds) > 0 {
		if _, err := tx.Table(models_crud.TableNamePolicyCompliance).
			Where(fmt.Sprintf("%s = ?", models_crud.PolicyComplianceColumnTenantId), tenantId).
			Where(fmt.Sprintf("%s IN ?", models_crud.PolicyComplianceColumnTemplateId), templateIds).
			Delete(allRecords); err != nil {
			return fmt.Errorf("failed to delete policy compliance for templates: %w", err)
		}
	}

	// For project/tenant scopes, also delete compliance by scope to clean cached
	// records for ancestor-scoped templates (not covered by the template_id cleanup above).
	switch scopeColumn {
	case models_crud.TemplatesColumnProjectId:
		if _, err := tx.Table(models_crud.TableNamePolicyCompliance).
			Where(fmt.Sprintf("%s = ? AND %s = ?",
				models_crud.PolicyComplianceColumnTenantId, models_crud.PolicyComplianceColumnProjectId),
				tenantId, scopeValue).
			Delete(allRecords); err != nil {
			return fmt.Errorf("failed to delete policy compliance for project %s: %w", scopeValue, err)
		}
	case models_crud.TemplatesColumnTenantId:
		if _, err := tx.Table(models_crud.TableNamePolicyCompliance).
			Where(fmt.Sprintf("%s = ?", models_crud.PolicyComplianceColumnTenantId), tenantId).
			Delete(allRecords); err != nil {
			return fmt.Errorf("failed to delete policy compliance for tenant %s: %w", tenantId, err)
		}
	}

	if len(templateIds) > 0 {
		if _, err := tx.Table(models_crud.TemplatesAssetsTableName).
			Where(fmt.Sprintf("%s = ?", models_crud.TemplatesAssetsColumnTenantId), tenantId).
			Where(fmt.Sprintf("%s IN ?", models_crud.TemplatesAssetsColumnTemplateId), templateIds).
			Delete(allRecords); err != nil {
			return fmt.Errorf("failed to delete template assets for templates: %w", err)
		}

		// When scopeColumn is cluster_id, this also removes department-scoped and project-scoped
		// templates under that cluster, since they store their ancestor cluster_id.
		if _, err := tx.Table(models_crud.TemplatesTableName).
			Where(fmt.Sprintf("%s = ?", models_crud.TemplatesColumnTenantId), tenantId).
			Where(fmt.Sprintf("%s = ?", scopeColumn), scopeValue).
			Delete(allRecords); err != nil {
			return fmt.Errorf("failed to delete templates for %s=%s in tenant %s: %w", scopeColumn, scopeValue, tenantId, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit template deletion: %w", err)
	}
	committed = true
	return nil
}

func DeleteTemplatesOfProjectId(ctx context.Context, tenantId string, projectId string) error {
	return deleteTemplatesByScope(ctx, tenantId, models_crud.TemplatesColumnProjectId, projectId)
}

func DeleteTemplatesOfClusterId(ctx context.Context, tenantId string, clusterId string) error {
	return deleteTemplatesByScope(ctx, tenantId, models_crud.TemplatesColumnClusterId, clusterId)
}

func DeleteTemplatesOfTenantId(ctx context.Context, tenantId string) error {
	return deleteTemplatesByScope(ctx, tenantId, models_crud.TemplatesColumnTenantId, tenantId)
}

func DeletePresetForTenant(ctx context.Context, tenantId uint, presetTemplateId uuid.UUID) *api.Error {
	db := runaidb.DB().WithContext(ctx)

	//
	//   mark the preset as deleted for this tenant
	//
	dbErr := saveTemplatesPresetsOfTenant(ctx, db, &models_crud.TemplatesPresets{
		ID:               uuid.New(),
		TenantId:         tenantId,
		PresetTemplateId: presetTemplateId.String(),
		IsDeleted:        true,
	})
	if dbErr != nil {
		return dbErr
	}

	return nil
}

func ListTemplates(
	ctx context.Context,
	tenantId uint,
	permittedScopes *openapi.PermittedScopes,
	offset int32,
	limit int32,
	workloadType *api.WorkloadType,
	sortOrder string,
	sortBy string,
	filterBy []queryapi.Filter,
	search string,
	ancestors *authorization_utils.Ancestors,
	routingMethod string,
) ([]models_crud.LoadedTemplate, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list templates", nil)

	db := runaidb.DB().WithContext(ctx)

	query, err := applyFilters(ctx, db, tenantId, permittedScopes, workloadType, filterBy)
	if err != nil {
		return nil, dbErr.WithDetails(err)
	}

	query = applyAncestorsFilter(query, ancestors)

	query = filterOutForkedSystemTemplates(query, tenantId)

	query = applyRoutingMethodFilter(query, routingMethod)

	if sortBy != "" {
		query = query.Order(querydbenrich.FormatSortClause(sortBy, sortOrder))
	}

	query = query.Order("id DESC")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, models_crud.TemplatesColumnsSearchVector))
	}

	templates := make([]models_crud.LoadedTemplate, 0, limit)

	err = query.Find(&templates)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: list templates: %v", dbErr.Message, err)
		return nil, dbErr.WithDetails(err)
	}

	//
	//   once we found the templates we need to get their ids in order to find assets
	//
	idToRec := make(map[string]*models_crud.LoadedTemplate)
	allIds := datastructs.Set[string]{}

	for idx, template := range templates {
		idToRec[template.ID.String()] = &templates[idx]
		allIds.Add(template.ID.String())
	}

	assets := make([]models_crud.TemplateAssets, 0)

	err = db.Model(models_crud.TemplateAssets{}).Where(fmt.Sprintf("%s in ?", models_crud.TemplatesAssetsColumnTemplateId), allIds.Keys()).Find(&assets)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: list templates assets: %v", dbErr.Message, err)
		return nil, dbErr.WithDetails(err)
	}

	for _, asset := range assets {
		template := idToRec[asset.TemplateId]
		if template != nil {
			template.Assets = append(template.Assets, asset)
		}
	}

	return templates, nil
}

type TemplateNameAndType struct {
	Name         string
	WorkloadType string
}

func MakeTemplateNameAndType(Name string, WorkloadType string) TemplateNameAndType {
	return TemplateNameAndType{
		Name:         Name,
		WorkloadType: WorkloadType,
	}
}

func ListPresetTemplates(ctx context.Context) (map[TemplateNameAndType]*models_crud.LoadedTemplate, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list preset templates", nil)

	db := runaidb.DB().WithContext(ctx)

	templates := make([]models_crud.LoadedTemplate, 0)

	filter := models_crud.LoadedTemplate{
		TemplateMeta: models_crud.TemplateMeta{
			ScopeType: string(api.SCOPETYPE_SYSTEM),
		},
	}

	err := db.Model(&filter).Where(&filter).Find(&templates)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: list templates: %v", dbErr.Message, err)
		return nil, dbErr.WithDetails(err)
	}

	nameToMeta := make(map[TemplateNameAndType]*models_crud.LoadedTemplate)

	for _, template := range templates {
		key := MakeTemplateNameAndType(template.Name, template.WorkloadType)
		nameToMeta[key] = &template
	}

	return nameToMeta, nil
}

func CountTemplates(
	ctx context.Context,
	tenantId uint,
	permittedScopes *openapi.PermittedScopes,
	filterBy []queryapi.Filter,
	search string,
	ancestors *authorization_utils.Ancestors,
	routingMethod string,
) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count templates", nil)

	db := runaidb.DB().WithContext(ctx)

	query, err := applyFilters(ctx, db, tenantId, permittedScopes, nil, filterBy)
	if err != nil {
		return 0, dbErr.WithDetails(err)
	}

	query = applyAncestorsFilter(query, ancestors)

	query = filterOutForkedSystemTemplates(query, tenantId)

	query = applyRoutingMethodFilter(query, routingMethod)

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, models_crud.TemplatesColumnsSearchVector))
	}

	var count int64

	err = query.Count(&count)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count templates, err: %v", err)
		return 0, dbErr.WithDetails(err)
	}

	return count, nil
}

func applyFilters(ctx context.Context, db orm.DB,
	tenantId uint,
	permittedScopes *openapi.PermittedScopes,
	workloadType *api.WorkloadType,
	filterBy []queryapi.Filter,
) (orm.DB, error) {
	query := db.Model(models_crud.Template{})

	query = query.Where(fmt.Sprintf("(%s = 0 or %s = ?)", models_crud.TemplatesColumnTenantId, models_crud.TemplatesColumnTenantId), tenantId)

	if workloadType != nil {
		query = query.Where(fmt.Sprintf("%s = ?", models_crud.TemplatesColumnWorkloadType), string(*workloadType))
	}

	if permittedScopes != nil {
		if permittedScopes.System == nil || !*permittedScopes.System {
			query = query.Where(getPermittedScopesFilter(*permittedScopes))
		}
	}

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	return query, nil
}

// applyRoutingMethodFilter filters templates by routing method when a cluster's routing
// method is known. Templates with matching routing method, empty routing method (agnostic),
// or explicit 'agnostic' value are included.
func applyRoutingMethodFilter(query orm.DB, routingMethod string) orm.DB {
	if routingMethod == "" {
		return query
	}
	return query.Where(
		fmt.Sprintf("%s IN (?, '', ?)", models_crud.TemplatesColumnRoutingMethod),
		routingMethod,
		models_crud.TemplatesColumnRoutingMethodAgnostic,
	)
}

func getPermittedScopesFilter(permittedScopes openapi.PermittedScopes) string {
	return fmt.Sprintf("(%s = 0 or %s)", models_crud.TemplatesColumnTenantId, authorization_utils.GetPermittedScopesDbFilter(permittedScopes))
}

func updateTemplatesAssets(tx orm.DB, tenantId uint, templateId string, templateAssets []models_crud.TemplateAssets) *api.Error {
	//
	//  delete all the existing records related to this template
	//
	deleteCriteria := models_crud.TemplateAssets{
		TenantId:   tenantId,
		TemplateId: templateId,
	}

	rows, err := tx.Model(deleteCriteria).Where(deleteCriteria).Delete(allRecords)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to clean template_assets of template %s", deleteCriteria.TemplateId),
			err).WithDetailsFromDbResult(rows, err)
	}

	//
	//   for add/update, we have to add the updated set of assets
	//
	for _, toAdd := range templateAssets {
		toAdd.TemplateId = templateId

		rows, err = tx.Model(toAdd).Create(toAdd)
		if err != nil {
			return api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to add template_assets of asset %s for template %s",
				toAdd.AssetId, toAdd.TemplateId), err).WithDetailsFromDbResult(rows, err)
		}
	}

	return nil
}

func ListTemplatesByAssetId(ctx context.Context, tenantId uint, assetId string) ([]models_crud.LoadedTemplate, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to list templates using asset %s", assetId), nil)

	templates := make([]models_crud.LoadedTemplate, 0)

	db := runaidb.DB().WithContext(ctx)

	joinQuery := fmt.Sprintf("JOIN %s ON %s.%s = %s.%s AND %s.%s = %s.%s",
		models_crud.TemplatesAssetsTableName,
		models_crud.TemplatesTableName, models_crud.TemplatesColumnTenantId,
		models_crud.TemplatesAssetsTableName, models_crud.TemplatesAssetsColumnTenantId,
		models_crud.TemplatesTableName, models_crud.TemplatesColumnId,
		models_crud.TemplatesAssetsTableName, models_crud.TemplatesAssetsColumnTemplateId)

	err := db.
		Model(&models_crud.Template{}).
		Joins(joinQuery).
		Where(fmt.Sprintf("%s.%s = ?", models_crud.TemplatesAssetsTableName, models_crud.TemplatesAssetsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s.%s = ?", models_crud.TemplatesAssetsTableName, models_crud.TemplatesAssetsColumnAssetId), assetId).
		Distinct().
		Preload("Assets").
		Find(&templates)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: join templates and templates assets: %v", dbErr.Message, err)
		return nil, dbErr.WithDetails(err)
	}

	return templates, nil
}

func applyAncestorsFilter(query orm.DB, ancestors *authorization_utils.Ancestors) orm.DB {
	if ancestors == nil {
		return query
	}

	ancestorsQuery := database.NewParametrizedQuery()

	ancestorsQuery.Add(fmt.Sprintf("((%s = ? or %s = ?)",
		models_crud.TemplatesColumnScopeType,
		models_crud.TemplatesColumnScopeType), api.SCOPETYPE_SYSTEM, api.SCOPETYPE_TENANT)

	// we either want cluster ancestors or project ancestors
	projectAncestors := (ancestors.ProjectId != nil)

	if projectAncestors {
		ancestorsQuery.Add(fmt.Sprintf(" or (%s = ? and %s = ?)",
			models_crud.TemplatesColumnScopeType,
			models_crud.TemplatesColumnProjectId), api.SCOPETYPE_PROJECT, *ancestors.ProjectId)

		ancestorsQuery.Add(fmt.Sprintf(" or (%s = ? and %s = ?)",
			models_crud.TemplatesColumnScopeType,
			models_crud.TemplatesColumnDepartmentId), api.SCOPETYPE_DEPARTMENT, *ancestors.DepartmentId)

		ancestorsQuery.Add(fmt.Sprintf(" or (%s = ? and %s = ?)",
			models_crud.TemplatesColumnScopeType,
			models_crud.TemplatesColumnClusterId), api.SCOPETYPE_CLUSTER, *ancestors.ClusterUuid)

	} else {
		// If we want ancestors of clusters, we need proj + dep + cluster of the given cluster
		ancestorsQuery.Add(fmt.Sprintf(" or (%s IN ? and %s = ?)",
			models_crud.TemplatesColumnScopeType,
			models_crud.TemplatesColumnClusterId),
			[]api.ScopeType{api.SCOPETYPE_PROJECT, api.SCOPETYPE_DEPARTMENT, api.SCOPETYPE_CLUSTER},
			*ancestors.ClusterUuid)
	}

	ancestorsQuery.Add(")")

	query = query.Where(ancestorsQuery.Query, ancestorsQuery.Params...)

	return query
}

func filterOutForkedSystemTemplates(query orm.DB, tenantId uint) orm.DB {
	//
	//  the query we want to build is:
	//    (scope_type != 'system' or id not in (select preset_template_id from crud_templates_presets where tenant_id = <our-tenant-id>))
	//  this leaves in the query all the records that are either non-system templates
	//  or system templates that their id is not marked as forked for this tenant
	//
	sql := fmt.Sprintf("(%s != ? or %s not in (select %s from %s where %s = ?))",
		models_crud.TemplatesColumnScopeType,
		models_crud.TemplatesColumnId,
		models_crud.TemplatesPresetsColumnPresetsTemplateId,
		models_crud.TemplatesPresetsTableName,
		models_crud.TemplatesColumnTenantId)

	return query.Where(sql, api.SCOPETYPE_SYSTEM, tenantId)
}

func CheckTemplateNameAvailability(ctx context.Context, tenantId uint, templateName string, scopeType string, scopeId string, templateId *string) (bool, *api.Error) {
	db := runaidb.DB().WithContext(ctx)

	nameValidationQuery, err := getTemplateNameValidationQuery(ctx, tenantId, scopeType, scopeId)
	if err != nil {
		return false, err
	}

	var count int64
	countErr := db.Model(models_crud.Template{}).Where(fmt.Sprintf("%s = ?", models_crud.TemplatesColumnName), templateName).
		Where(nameValidationQuery.Query, nameValidationQuery.Params...).
		Count(&count)
	if countErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to check template name availability: %v", countErr)
		return false, api.NewError(http.StatusInternalServerError, "failed to check template name availability", countErr)
	}

	// If count > 0, the name is taken (not available)
	// If count == 0, the name is available
	return count == 0, nil
}

// getTemplateNameValidationQuery builds a query to check if a template name is already taken in the org hierarchy.
// The validation follows the organizational hierarchy: system -> tenant -> cluster -> department -> project
func getTemplateNameValidationQuery(ctx context.Context, tenantId uint, scopeType string, scopeId string) (*database.ParametrizedQuery, *api.Error) {
	nameValidationQuery := database.NewParametrizedQuery()

	nameValidationQuery.Add("(")

	// Checking system scope templates
	nameValidationQuery.Add(fmt.Sprintf("(%s = ?)", models_crud.TemplatesColumnScopeType), api.SCOPETYPE_SYSTEM)

	// Checking all templates in the same tenant
	if scopeType == string(api.SCOPETYPE_TENANT) {
		nameValidationQuery.Add(fmt.Sprintf(" OR (%s = ?)", models_crud.TemplatesColumnTenantId), tenantId)
		nameValidationQuery.Add(")")
		return nameValidationQuery, nil
	}

	// Checking tenant scope templates in the same tenant
	nameValidationQuery.Add(fmt.Sprintf(" OR (%s = ? AND %s = ?)", models_crud.TemplatesColumnScopeType, models_crud.TemplatesColumnTenantId), api.SCOPETYPE_TENANT, tenantId)

	// Checking cluster scope templates
	if scopeType == string(api.SCOPETYPE_CLUSTER) {
		nameValidationQuery.Add(fmt.Sprintf(" OR (%s = ?)", models_crud.TemplatesColumnClusterId), scopeId)
		nameValidationQuery.Add(")")
		return nameValidationQuery, nil
	}

	// Checking department scope templates
	if scopeType == string(api.SCOPETYPE_DEPARTMENT) {
		ancestors, err := org_unit.Get().GetAncestorsOfScope(ctx, orgUnitApi.ScopeType(api.SCOPETYPE_DEPARTMENT), scopeId)
		if err != nil {
			return nil, api.NewError(int(err.GetCode()), "failed to get ancestors of department", errors.New(err.Message))
		}
		clusterId := ancestors.ClusterUuid.Get()
		if clusterId == nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to get ancestor cluster id for department", errors.New("cluster id is nil"))
		}
		// Checking all cluster scope templates in the same cluster
		nameValidationQuery.Add(fmt.Sprintf(" OR (%s = ? AND %s = ?)", models_crud.TemplatesColumnScopeType, models_crud.TemplatesColumnClusterId), api.SCOPETYPE_CLUSTER, *clusterId)

		// Checking all department and project scope templates in the same department
		nameValidationQuery.Add(fmt.Sprintf(" OR (%s = ?)", models_crud.TemplatesColumnDepartmentId), scopeId)
		nameValidationQuery.Add(")")
		return nameValidationQuery, nil
	}

	// Checking project scope templates
	if scopeType == string(api.SCOPETYPE_PROJECT) {
		ancestors, err := org_unit.Get().GetAncestorsOfScope(ctx, orgUnitApi.ScopeType(api.SCOPETYPE_PROJECT), scopeId)
		if err != nil {
			return nil, api.NewError(int(err.GetCode()), "failed to get ancestors of project", errors.New(err.Message))
		}
		departmentId := ancestors.DepartmentId.Get()
		clusterId := ancestors.ClusterUuid.Get()
		if clusterId == nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to get ancestor cluster id for project", errors.New("cluster id is nil"))
		}
		if departmentId == nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to get ancestor department id for project", errors.New("department id is nil"))
		}
		// Checking all templates in ancestors scopes and the same project
		nameValidationQuery.Add(fmt.Sprintf(" OR (%s = ? AND %s = ?)", models_crud.TemplatesColumnScopeType, models_crud.TemplatesColumnClusterId), api.SCOPETYPE_CLUSTER, *clusterId)
		nameValidationQuery.Add(fmt.Sprintf(" OR (%s = ? AND %s = ?)", models_crud.TemplatesColumnScopeType, models_crud.TemplatesColumnDepartmentId), api.SCOPETYPE_DEPARTMENT, *departmentId)
		nameValidationQuery.Add(fmt.Sprintf(" OR (%s = ? AND %s = ?)", models_crud.TemplatesColumnScopeType, models_crud.TemplatesColumnProjectId), api.SCOPETYPE_PROJECT, scopeId)
		nameValidationQuery.Add(")")
		return nameValidationQuery, nil
	}

	return nil, api.NewValidationError(fmt.Sprintf("unsupported scope type: %s", scopeType))
}
