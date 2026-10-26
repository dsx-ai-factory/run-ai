/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorizationApi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

type UnifiedListResponse struct {
	Entries    []*api.SupersetTemplate `json:"entries,omitempty"`
	Pagination api.PaginationInfo      `json:"pagination"`
}

const nonPrivate = false

// getFirstFilterValue returns the first value from a filter if it matches the field name, otherwise returns empty string
func getFirstFilterValue(filter queryapi.Filter, fieldName string) string {
	if filter.Field == fieldName && len(filter.Values) > 0 {
		return filter.Values[0]
	}
	return ""
}

// getAncestorsForIncludeAncestors handles the logic for both projectId and clusterId filtering
// when includeAncestors=true. It detects which type of filter is present and calls the appropriate handler.
func getAncestorsForIncludeAncestors(ctx context.Context, filters []queryapi.Filter) ([]queryapi.Filter, *helpers.Ancestors, *api.Error) {
	clusterId := ""
	projectId := ""

	// Single pass to detect both filter types
	for _, filter := range filters {
		if clusterId == "" {
			if value := getFirstFilterValue(filter, "cluster_id"); value != "" {
				clusterId = value
				continue
			}
		}
		if projectId == "" {
			if value := getFirstFilterValue(filter, "project_id"); value != "" {
				projectId = value
			}
		}
		// If we found both, we can break early
		if clusterId != "" && projectId != "" {
			break
		}
	}

	// Return error if both filters are present
	if clusterId != "" && projectId != "" {
		return nil, nil, api.NewValidationError("cannot use both projectId and clusterId filters with includeAncestors=true")
	}

	// Handle clusterId filter
	if clusterId != "" {
		return getClusterAncestors(ctx, filters)
	}

	// Handle projectId filter
	if projectId != "" {
		return getProjectAncestors(ctx, filters)
	}

	// No valid filter found
	return nil, nil, api.NewValidationError("filter for specific projectId or clusterId is mandatory for includeAncestors=true")
}

// resolveRoutingMethodFromFilters extracts clusterId from the parsed filters (or derives it
// from projectId) and resolves the cluster's routing method based on SubdomainRouting config.
// Returns empty string if neither clusterId nor projectId filter is present.
func resolveRoutingMethodFromFilters(ctx context.Context, filters []queryapi.Filter) string {
	var clusterId, projectId string
	for _, filter := range filters {
		if clusterId == "" {
			if value := getFirstFilterValue(filter, "cluster_id"); value != "" {
				clusterId = value
				continue
			}
		}
		if projectId == "" {
			if value := getFirstFilterValue(filter, "project_id"); value != "" {
				projectId = value
			}
		}
	}

	if clusterId == "" && projectId != "" {
		project, err := org_unit.Get().GetProjectById(ctx, projectId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to resolve cluster from project %s: %v", projectId, err)
			return ""
		}
		clusterId = project.ClusterId
	}

	if clusterId == "" {
		return ""
	}

	clusterConfig, err := checker.NewClusterCapabilitiesChecker(clusterId).ClusterConfiguration(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to resolve cluster routing method for cluster %s: %v", clusterId, err)
		return ""
	}

	if clusterConfig.SubdomainRouting {
		return models_crud.TemplatesColumnRoutingMethodHostBase
	}
	return models_crud.TemplatesColumnRoutingMethodPathBase
}

func (ts *TemplatesService) CreateTemplate(ctx context.Context, request *UnifiedRequest) *api.Error {
	audit.AuditFromCtx(ctx).WithEntityName(request.template.Meta.Name).WithEntityType(request.EntityKind())

	template := request.template

	// NOTE! in order to be aligned with ValidateAccessByResourceTypeAndAction I load the ancestors from the backend
	// client, but this needs to be changed from backend client to org unit client. we will do it when we refactor
	// the ValidateAccessByResourceTypeAndAction itself.
	var ancestors *orgUnitApi.Ancestors

	if request.template.Meta.ScopeType == api.SCOPETYPE_DEPARTMENT || request.template.Meta.ScopeType == api.SCOPETYPE_PROJECT {
		var getErr *orgUnitApi.Error

		ancestors, getErr = orgunit.Get().Client().GetAncestors(ctx, string(request.template.Meta.ScopeType), request.template.Meta.ScopeId)
		if getErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to obtain ancestors of %s/%s", request.template.Meta.ScopeType, request.template.Meta.ScopeId)
			return api.NewError(int(auth.ForbiddenError.Code), auth.ForbiddenError.Message, nil)
		}
	}

	if !request.migrating {
		authErr := auth.ValidateAccessByResourceTypeAndAction(auth.ContextWithAncestors(ctx, ancestors), request.ScopeInfo(), authorizationApi.RESOURCETYPE_TEMPLATES, authorizationApi.ACTION_CREATE, nonPrivate)
		if authErr != nil {
			return api.NewError(int(authErr.Code), authErr.Message, nil)
		}
	}

	switch template.Meta.ScopeType {
	case api.SCOPETYPE_SYSTEM:
		if !request.migrating {
			return api.NewValidationError("system scope is not supported for workload templates")
		}

	case api.SCOPETYPE_TENANT:
		request.template.Meta.ScopeId = fmt.Sprintf("%d", request.tenantId)

	case api.SCOPETYPE_CLUSTER:
		template.Meta.ClusterId = convutils.Ref(template.Meta.ScopeId)

	case api.SCOPETYPE_DEPARTMENT:
		template.Meta.ClusterId = ancestors.ClusterUuid.Get()
		template.Meta.DepartmentId = convutils.Ref(template.Meta.ScopeId)

	case api.SCOPETYPE_PROJECT:
		template.Meta.ClusterId = ancestors.ClusterUuid.Get()
		template.Meta.DepartmentId = ancestors.DepartmentId.Get()

		projectIdNum, convErr := strconv.ParseInt(template.Meta.ScopeId, 10, 32)
		if convErr != nil {
			return api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to parse project id '%v'", convErr), convErr)
		}
		template.Meta.ProjectId = convutils.Ref(int32(projectIdNum))

	default:
		return api.NewValidationError(fmt.Sprintf("scope '%s' is not supported", template.Meta.ScopeType))
	}

	assetSpecAsFlat, linkedAssets, specErr := request.ValidateSpecs(ctx)
	if specErr != nil {
		return specErr
	}

	template.Meta.Id = uuid.NewString()

	template.Meta.WorkloadType = request.workloadType

	template.Meta.CreatedAt = time.Now()
	template.Meta.UpdatedAt = &template.Meta.CreatedAt

	// When migrating respect the provided CreatedBy/UpdatedBy values
	// Otherwise, use values from context
	if request.migrating {
		// Keep the values that were set in the template preset
		if template.Meta.CreatedBy == "" {
			template.Meta.CreatedBy = convutils.Deref(runaictx.GetString(ctx, runaictx.CfSubjectID), "unknown")
		}
		if template.Meta.UpdatedBy == nil || *template.Meta.UpdatedBy == "" {
			template.Meta.UpdatedBy = &template.Meta.CreatedBy
		}
	} else {
		template.Meta.CreatedBy = convutils.Deref(runaictx.GetString(ctx, runaictx.CfSubjectID), "unknown")
		template.Meta.UpdatedBy = &template.Meta.CreatedBy
	}

	//
	//   calculate combined spec asset + template spec
	//
	conflictErr := validateSpecAndAssetsDoNoConflict(template.Spec, template.Assets)
	if conflictErr != nil {
		return conflictErr
	}

	if template.Master != nil {
		conflictErr := validateSpecAndAssetsDoNoConflict(template.Master.Spec, template.Master.Assets)
		if conflictErr != nil {
			return conflictErr
		}
	}

	if assetsGiven(template) {
		mutateErr := mutateTemplate(template, request.workloadType, assetSpecAsFlat)
		if mutateErr != nil {
			return mutateErr
		}
	} else {
		template.CombinedSpec = template.Spec
		if template.Master != nil {
			template.CombinedMasterSpec = template.Master.Spec
		}
	}

	//
	//   some fields are forbidden for templates. make sure they are not included in the spec or master spec
	//
	err := validateForbiddenFieldsForTemplate(template.CombinedSpec.EnvironmentVariables, template.CombinedSpec.Storage)
	if err != nil {
		return err
	}

	if template.CombinedMasterSpec != nil {
		err = validateForbiddenFieldsForTemplate(template.CombinedMasterSpec.EnvironmentVariables, template.CombinedMasterSpec.Storage)
		if err != nil {
			return err
		}
	}

	template.Runnable = request.RunnableInfo()

	rec, dbErr := request.ToDbRecord()
	if dbErr != nil {
		return dbErr
	}

	templateAssets := linkedAssetsToTemplateAssetsRec(request.tenantId, template.Meta.Id, linkedAssets)

	dbErr = dao_crud.CreateTemplate(ctx, rec, templateAssets)
	if dbErr != nil {
		return dbErr
	}

	audit.AuditFromCtx(ctx).WithEntityId(request.template.Meta.Id)

	return nil
}

type updateOptions struct {
	patchMode bool

	// for patch, there is an option to specify list of json path style expression of fields to clear
	// like spec.image, masterSpec.distributedFramework, etc.
	fieldsToClear []string

	// for patch, we need the original template update because it allows us to know if value for
	// meta were provided or not (the unified has the values as strings not as pointer)
	patchMetaFields *api.TemplatePatchMetaFields
}

func (ts *TemplatesService) UpdateTemplate(ctx context.Context, templateId string, request *UnifiedRequest, options updateOptions) (*UnifiedRequest, *api.Error) {
	audit.AuditFromCtx(ctx).WithEntityId(templateId)

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	// Fetch existing template first to get its scope for authorization
	existingRec, dbErr := dao_crud.GetTemplateById(ctx, tenantId, templateId, nil)
	if dbErr != nil {
		return nil, dbErr
	}

	// Skip authorization check for migrating requests (like preset template updates)
	if !request.migrating {
		// Convert existing record to get scope info for auth validation
		existingTemplate, convErr := NewUnifiedRequestFromRec(existingRec)
		if convErr != nil {
			return nil, convErr
		}
		if existingTemplate.IsSystem() {
			return nil, api.NewError(http.StatusForbidden, "system templates cannot be updated", nil)
		}
		// Use existing template's scope for authorization (not the request's empty scope)
		authErr := auth.ValidateAccessByResourceTypeAndAction(ctx, existingTemplate.ScopeInfoForAuthValidation(tenantId), authorizationApi.RESOURCETYPE_TEMPLATES, authorizationApi.ACTION_UPDATE, nonPrivate)
		if authErr != nil {
			return nil, api.NewError(int(authErr.Code), authErr.Message, nil)
		}
	}

	updatedTemplate, err := updateTemplateInt(ctx, existingRec, request, options)
	if err != nil {
		return nil, err
	}

	audit.AuditFromCtx(ctx).WithEntityName(existingRec.Name).WithEntityType(request.EntityKind())

	return updatedTemplate, nil
}

func updateTemplateInt(ctx context.Context, existingRec *models_crud.LoadedTemplate, request *UnifiedRequest, options updateOptions) (*UnifiedRequest, *api.Error) {
	existingTemplate, convErr := NewUnifiedRequestFromRec(existingRec)
	if convErr != nil {
		return nil, convErr
	}

	if request.workloadType != existingTemplate.workloadType {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s template '%s' does not exist", request.workloadType, existingRec.ID.String()), nil)
	}

	var linkedAssets map[string]replicas
	var updErr *api.Error

	if options.patchMode {
		linkedAssets, updErr = existingTemplate.PatchTemplateFrom(ctx, request, options)
	} else {
		linkedAssets, updErr = existingTemplate.UpdateTemplateFrom(ctx, request)
	}

	if updErr != nil {
		return nil, updErr
	}

	rec, dbErr := existingTemplate.ToDbRecord()
	if dbErr != nil {
		return nil, dbErr
	}

	templateAssets := linkedAssetsToTemplateAssetsRec(existingRec.TenantId, existingRec.ID.String(), linkedAssets)

	dbErr = dao_crud.UpdateTemplate(ctx, rec, templateAssets)
	if dbErr != nil {
		return nil, dbErr
	}

	return existingTemplate, nil
}

func (ts *TemplatesService) DeleteTemplate(ctx context.Context, workloadType api.WorkloadType, templateId string) *api.Error {
	audit.AuditFromCtx(ctx).WithEntityType(string(workloadType)).WithEntityId(templateId)

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return apiErr
	}

	templateUuid, err := uuid.Parse(templateId)
	if err != nil {
		return api.NewValidationError(fmt.Sprintf("invalid uuid: %v", err))
	}

	existingRec, dbErr := dao_crud.GetTemplateById(ctx, tenantId, templateId, nil)
	if dbErr != nil {
		return dbErr
	}

	existingTemplate, convErr := NewUnifiedRequestFromRec(existingRec)
	if convErr != nil {
		return convErr
	}

	authErr := auth.ValidateAccessByResourceTypeAndAction(
		ctx,
		existingTemplate.ScopeInfoForAuthValidation(tenantId),
		authorizationApi.RESOURCETYPE_TEMPLATES,
		authorizationApi.ACTION_DELETE,
		nonPrivate,
	)
	if authErr != nil {
		return api.NewError(int(authErr.Code), authErr.Message, nil)
	}

	if workloadType != existingTemplate.workloadType {
		return api.NewError(http.StatusNotFound, fmt.Sprintf("%s template '%s' does not exist", workloadType, templateId), nil)
	}

	if existingTemplate.IsSystem() {
		dbErr = dao_crud.DeletePresetForTenant(ctx, tenantId, templateUuid)
	} else {
		dbErr = dao_crud.DeleteTemplateById(ctx, tenantId, templateUuid)
	}
	if dbErr != nil {
		return dbErr
	}

	audit.AuditFromCtx(ctx).WithEntityName(existingTemplate.template.Meta.Name)

	return nil
}

func (ts *TemplatesService) GetTemplate(ctx context.Context, workloadType *api.WorkloadType, templateId string) (*UnifiedRequest, *api.Error) {
	audit.AuditFromCtx(ctx).WithEntityId(templateId)
	if workloadType != nil {
		audit.AuditFromCtx(ctx).WithEntityType(string(*workloadType))
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	existingRec, dbErr := dao_crud.GetTemplateById(ctx, tenantId, templateId, nil)
	if dbErr != nil {
		return nil, dbErr
	}

	existingTemplate, convErr := NewUnifiedRequestFromRec(existingRec)
	if convErr != nil {
		return nil, convErr
	}

	_, _, authErr := validateGetByIdPermission(ctx, &existingTemplate.template.Meta)
	if authErr != nil {
		return nil, api.NewError(int(authErr.Code), authErr.Message, nil)
	}

	if workloadType != nil && *workloadType != existingTemplate.workloadType {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s template '%s' does not exist", *workloadType, templateId), nil)
	}

	audit.AuditFromCtx(ctx).WithEntityName(existingTemplate.template.Meta.Name)

	return existingTemplate, nil
}

func (ts *TemplatesService) ListTemplatesUnified(ctx context.Context, workloadType *api.WorkloadType, offsetParam int32, limitParam int32, sortOrderParam string, sortByParam string,
	filterByParam []string, searchParam string, complianceParams *ComplianceParams, includeAncestors bool,
) (*UnifiedListResponse, *api.Error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	migrateAssetTemplatesIfNeeded(ctx, ts.tenantSettingsClient)

	limitParam = setDefaultLimit(limitParam)

	// for list, we need to allow adding filtering for all the org units the user is allowed to see.
	// so for example if user is a department admin, he can see not only department scoped templates
	// but also all the projects under this department.
	permittedScopes, err := auth.LoadAllowedOrgUnits(ctx)
	if err != nil {
		return nil, api.NewError(int(err.Code), err.Message, errors.New(err.Details))
	}

	filters, filterErr := queryapi.NewParser(models_crud.Template{}.FilterFieldsToColumn()).Filters(filterByParam)
	if filterErr != nil {
		return nil, api.NewValidationError(filterErr.Error())
	}

	routingMethod := resolveRoutingMethodFromFilters(ctx, filters)

	var templateAncestors *helpers.Ancestors

	if includeAncestors {
		var ancestorsErr *api.Error
		filters, templateAncestors, ancestorsErr = getAncestorsForIncludeAncestors(ctx, filters)
		if ancestorsErr != nil {
			return nil, ancestorsErr
		}
	}

	sortColumn, sortError := queryapi.NewParser(models_crud.Template{}.SortFieldsToColumn()).GetMapping(sortByParam)
	if sortError != nil {
		return nil, api.NewValidationError(sortError.Error())
	}

	templates, dbErr := dao_crud.ListTemplates(ctx, tenantId, permittedScopes.Raw(), offsetParam, limitParam, workloadType,
		sortOrderParam, sortColumn, filters, searchParam, templateAncestors, routingMethod)
	if dbErr != nil {
		return nil, dbErr
	}

	response := UnifiedListResponse{
		Entries: make([]*api.SupersetTemplate, 0, len(templates)),
	}
	for _, template := range templates {
		unifiedTemplate, convErr := NewUnifiedRequestFromRec(&template)
		if convErr != nil {
			return nil, convErr
		}

		response.Entries = append(response.Entries, unifiedTemplate.template)
	}

	if complianceParams != nil {
		err := patchPolicyCompliance(ctx, complianceParams, response.Entries)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to patch compliance policy compliance policy: %v", err)
		}
	}

	response.Pagination = api.PaginationInfo{
		Limit:  float32(limitParam),
		Offset: float32(offsetParam),
	}

	recordCount := int32(len(response.Entries))
	if recordCount == limitParam {
		response.Pagination.NextOffset = float32(offsetParam + recordCount)
	}

	for i, template := range response.Entries {
		projected, err := legacyCompatibleTemplate(template)
		if err != nil {
			return nil, err
		}
		response.Entries[i] = projected
	}

	return &response, nil
}

func (ts *TemplatesService) CountTemplatesUnified(ctx context.Context, filterByParam []string, searchParam string, includeAncestors bool,
) (int64, *api.Error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return 0, apiErr
	}

	permittedScopes, err := authorizationClient.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	filters, filterErr := queryapi.NewParser(models_crud.Template{}.FilterFieldsToColumn()).Filters(filterByParam)
	if filterErr != nil {
		return 0, api.NewValidationError(filterErr.Error())
	}

	routingMethod := resolveRoutingMethodFromFilters(ctx, filters)

	var templateAncestors *helpers.Ancestors

	if includeAncestors {
		var ancestorsErr *api.Error
		filters, templateAncestors, ancestorsErr = getAncestorsForIncludeAncestors(ctx, filters)
		if ancestorsErr != nil {
			return 0, ancestorsErr
		}
	}

	countResult, dbErr := dao_crud.CountTemplates(ctx, tenantId, permittedScopes, filters, searchParam, templateAncestors, routingMethod)
	if dbErr != nil {
		return 0, dbErr
	}

	return countResult, nil
}

func checkTemplateNameAvailability(ctx context.Context, templateName string, scopeType string, scopeId string) *api.Error {
	audit.AuditFromCtx(ctx).WithEntityName(templateName).WithEntityType(scopeType).WithEntityId(scopeId)

	missingParams := []string{}
	if templateName == "" {
		missingParams = append(missingParams, "templateName")
	}
	if scopeType == "" {
		missingParams = append(missingParams, "scopeType")
	}
	if scopeId == "" && api.ScopeType(scopeType) != api.SCOPETYPE_SYSTEM && api.ScopeType(scopeType) != api.SCOPETYPE_TENANT && api.ScopeType(scopeType) != api.SCOPETYPE_CLUSTER {
		missingParams = append(missingParams, "scopeId")
	}

	if len(missingParams) == 1 {
		return api.NewValidationError(fmt.Sprintf("mandatory parameter is not provided: %s", missingParams[0]))
	}
	if len(missingParams) > 1 {
		return api.NewValidationError(fmt.Sprintf("mandatory parameters are not provided: %s", strings.Join(missingParams, ", ")))
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return apiErr
	}

	// Pass nil for templateId since this is checking availability for a new template (not updating)
	isAvailable, dbErr := dao_crud.CheckTemplateNameAvailability(ctx, tenantId, templateName, scopeType, scopeId, nil)
	if dbErr != nil {
		return dbErr
	}

	if !isAvailable {
		return api.NewError(http.StatusConflict, "the given template name is not available", nil)
	}

	return nil
}

func setDefaultLimit(limit int32) int32 {
	if limit == 0 {
		limit = 50
	}
	limit = int32(math.Min(float64(limit), 500))
	limit = int32(math.Max(float64(limit), 1))

	return limit
}

// getProjectAncestors is invoked when includeAncestors flag is set to true. it expects projectId
// filter to be provided, but in dao the simple project_id=xxx query should be replaced with a complex
// query based on Ancestors. to do so, it removes the original filter and returns the ancestors. the
// dao will take care for the rest.
func getProjectAncestors(ctx context.Context, filters []queryapi.Filter) ([]queryapi.Filter, *helpers.Ancestors, *api.Error) {
	updatedFilters := make([]queryapi.Filter, 0, len(filters))
	projectId := ""

	for _, filter := range filters {
		if value := getFirstFilterValue(filter, "project_id"); value != "" {
			projectId = value
		} else {
			updatedFilters = append(updatedFilters, filter)
		}
	}

	if projectId == "" {
		return nil, nil, api.NewValidationError("filter for specific projectId is mandatory for includeAncestors=true")
	}

	//
	//   try to load the ancestors from the cache
	//

	ancestors, cacheErr := helpers.GetAncestorsCache(ctx, services.ProjectsResourceType, projectId)

	if cacheErr == nil && ancestors != nil {
		return updatedFilters, ancestors, nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Could not find project <%s> ancestors in cache. err: %v, continuing", projectId, cacheErr)

	project, orgUnitErr := org_unit.Get().GetProjectById(ctx, projectId)
	if orgUnitErr != nil {
		return nil, nil, api.NewError(int(orgUnitErr.Code), orgUnitErr.Message, errors.New(convutils.Deref(orgUnitErr.Details, "failed to obtain project ancestors")))
	}

	ancestors = &helpers.Ancestors{
		TenantId:     strconv.Itoa(int(auth.TenantIdOf(ctx))),
		ClusterUuid:  convutils.Ref(project.ClusterId),
		DepartmentId: convutils.Ref(project.ParentId),
		ProjectId:    convutils.Ref(projectId),
	}
	helpers.SetAncestorsCache(ctx, services.ProjectsResourceType, projectId, *ancestors)

	return updatedFilters, ancestors, nil
}

// getClusterAncestors is invoked when includeAncestors flag is set to true and clusterId filter is provided.
// it expects clusterId filter to be provided, but in dao the simple clusterId=xxx query should be replaced with a complex
// query that includes system, tenant, department, and project scoped templates that belong to the cluster. to do so, it removes the original filter and returns the ancestors.
// the dao will take care for the rest.
func getClusterAncestors(ctx context.Context, filters []queryapi.Filter) ([]queryapi.Filter, *helpers.Ancestors, *api.Error) {
	updatedFilters := make([]queryapi.Filter, 0, len(filters))
	clusterId := ""

	for _, filter := range filters {
		if value := getFirstFilterValue(filter, "cluster_id"); value != "" {
			clusterId = value
		} else {
			updatedFilters = append(updatedFilters, filter)
		}
	}

	// For clusterId filtering with includeAncestors=true, we want to include:
	// 1. Templates with the specific clusterId (original behavior)
	// 2. System scoped templates
	// 3. Tenant scoped templates
	// 4. Department and project templates that belong to this cluster
	// We need to fetch the cluster's department and project information
	ancestors := &helpers.Ancestors{
		TenantId:    strconv.Itoa(int(auth.TenantIdOf(ctx))),
		ClusterUuid: convutils.Ref(clusterId),
		// DepartmentId and ProjectId are nil since we want to include ALL department and project templates
		// that belong to this cluster, not just specific ones
	}

	return updatedFilters, ancestors, nil
}

func validateForbiddenFieldsForTemplate(envVars *[]api.EnvironmentVariable, storage *api.WorkspaceSpecSpecAllOfStorage) *api.Error {
	if envVars != nil {
		for _, envVar := range *envVars {
			if envVar.UserCredential != nil {
				name := convutils.Deref(envVar.Name, "environmentVariables")
				return api.NewValidationError(fmt.Sprintf("%v: cannot use user credentials as the value source of an environmentVariable within a template", name))
			}
		}
	}

	if storage != nil && storage.Git != nil {
		for _, gitItem := range *storage.Git {
			if gitItem.SecretRef != nil {
				name := convutils.Deref(gitItem.Name, "git")
				return api.NewValidationError(fmt.Sprintf("%v: cannot reference user credentials within a template", name))
			}
		}
	}

	return nil
}

// permissions of reading a template is different from permissions of create/update/delete (CUD)
// they are similar to asset read permissions in which cluster scoped asset, for example, can be
// read by anyone having permission of cluster, department or project belonging to that cluster
// this is up-side-down from usual permission in which cluster asset/template can be deleted
// only by someone having cluster, tenant or system permissions, but not department/project admin
// this function is similar to the logic used in asset service.
func validateGetByIdPermission(ctx context.Context, meta *api.TemplateMetaFields) (*auth.PermittedScopes, []string, *api.Error) {
	tenantId := auth.TenantIdOf(ctx)

	permittedScopes, err := auth.LoadAllowedOrgUnits(ctx)
	if err != nil {
		return nil, nil, api.NewError(int(err.Code), err.Message, errors.New(err.Details))
	}

	switch meta.ScopeType {
	case api.SCOPETYPE_PROJECT:
		projectIdStr := fmt.Sprintf("%d", convutils.Deref(meta.ProjectId, 0))
		if permittedScopes.AllowedProjects().Has(projectIdStr) {
			return permittedScopes, permittedScopes.AllowedProjectsArr(), nil
		}

	case api.SCOPETYPE_DEPARTMENT:
		if permittedScopes.AllowedDepartments().Has(convutils.Deref(meta.DepartmentId, "")) {
			return permittedScopes, permittedScopes.AllowedProjectsArr(), nil
		}

	case api.SCOPETYPE_CLUSTER:
		if permittedScopes.AllowedCluster().Has(convutils.Deref(meta.ClusterId, "")) {
			return permittedScopes, permittedScopes.AllowedProjectsArr(), nil
		}

	case api.SCOPETYPE_TENANT:
		if auth.TenantIdOf(ctx) == tenantId {
			return permittedScopes, permittedScopes.AllowedProjectsArr(), nil
		}

	case api.SCOPETYPE_SYSTEM:
		return permittedScopes, permittedScopes.AllowedProjectsArr(), nil
	}

	return permittedScopes, nil, api.NewError(int(auth.ForbiddenError.Code), auth.ForbiddenError.Message, errors.New(auth.ForbiddenError.Details))
}
