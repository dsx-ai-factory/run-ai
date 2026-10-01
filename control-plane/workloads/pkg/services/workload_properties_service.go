/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/google/uuid"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	config "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

const (
	RunaiSource    = "Run.ai"
	StatusCreating = "Creating"
	StatusFailed   = "Failed"
	StatusDeleting = "Deleting"
	StatusUpdating = "Updating"
	StatusReady    = "Ready"
	DeletedPhase   = "Deleted"
)

type WorkloadPropertiesService struct {
	Conf                *config.ServerConfig
	WorkloadsDao        *dao.WorkloadsDao
	AuthorizationClient authorization_client.IAuthorizationClient
}

func NewWorkloadPropertiesService() api.WorkloadPropertiesAPIServicer {
	conf := runai_config.Get[*config.ServerConfig]()
	return &WorkloadPropertiesService{
		Conf:                conf,
		WorkloadsDao:        dao.NewWorkloadsDao(conf),
		AuthorizationClient: authorization_client.NewAuthorizationClient(),
	}
}

func (s *WorkloadPropertiesService) ListCategories(ctx context.Context) (api.ImplResponse, error) {
	_, apiErr := s.isReadWorkloadPropertiesAllowed(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	categories, apiErr := dao.DbListCategories(ctx, tenantId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	response := api.ListCategories200Response{
		Categories: categories,
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *WorkloadPropertiesService) GetCategoryById(ctx context.Context, categoryId string) (api.ImplResponse, error) {
	_, apiErr := s.isReadWorkloadPropertiesAllowed(ctx)
	if apiErr != nil {
		// if the user has no permissions to read workload properties or workloads,
		// we return a 404 instead of 403, to avoid leaking information about the existence of these resources.
		apiErr.Code = http.StatusNotFound
		apiErr.Message = fmt.Sprintf("category with id <%s> not found", categoryId)
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	category, apiErr := dao.DbGetCategoryById(ctx, tenantId, categoryId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	return api.Response(http.StatusOK, category), nil
}

func (s *WorkloadPropertiesService) ListWorkloadTypes(ctx context.Context, externalTypesOnly bool, filterBy []string, noStatus bool) (api.ImplResponse, error) {
	permittedScopes, apiErr := s.isReadWorkloadPropertiesAllowed(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	filters, filterErr := queryapi.NewParser(models.ExternalWorkloadTypeConfigView{}.FieldsConfiguration()).Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	workloadTypeConfigs, dbErr := dao.DbListWorkloadTypeConfigs(ctx, tenantId, permittedScopes, filters, noStatus)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	apiWorkloadTypes := WorkloadTypeConfigsDbToApi(workloadTypeConfigs)
	if !externalTypesOnly {
		apiWorkloadTypes, apiErr = AppendRunaiWorkloadTypeConfigs(ctx, tenantId, filters, apiWorkloadTypes)
		if apiErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to append runai workload types. err: %v", apiErr)
			return NewErrorResponse(apiErr), nil
		}
	}

	response := api.ListWorkloadTypes200Response{
		Types: apiWorkloadTypes,
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *WorkloadPropertiesService) CreateWorkloadType(ctx context.Context, createFields api.WorkloadTypeCreateFields) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	parsedParams, validationErr := tryParseWorkloadTypeCreateParams(createFields)
	if validationErr != nil {
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}

	apiErr = validateWorkloadTypeCreateAllowed(ctx, tenantId, *parsedParams)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	externalWorkloadTypeConfig, apiErr := dao.DbCreateWorkloadTypeWithConfig(ctx, tenantId, *parsedParams)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	workloadTypeConfig, apiErr := dao.DbGetWorkloadTypeConfigById(ctx, tenantId, externalWorkloadTypeConfig.WorkloadTypeId.String())
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully created workload type <%s>", workloadTypeConfig.Name)
	response := WorkloadTypeConfigDbToApi(*workloadTypeConfig)
	return api.Response(http.StatusCreated, response), nil
}

func (s *WorkloadPropertiesService) GetWorkloadType(ctx context.Context, typeId string) (api.ImplResponse, error) {
	permittedScopes, apiErr := s.isReadWorkloadPropertiesAllowed(ctx)
	if apiErr != nil {
		// if the user has no permissions to read workload properties or workloads,
		// we return a 404 instead of 403, to avoid leaking information about the existence of these resources.
		apiErr.Code = http.StatusNotFound
		apiErr.Message = fmt.Sprintf("workload-type with id <%s> not found", typeId)
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	workloadTypeUuid, parseErr := uuid.Parse(typeId)
	if parseErr != nil {
		apiErr = api.NewError(http.StatusBadRequest, fmt.Sprintf("type Id must be a valid uuid. <%s>", typeId), nil)
		return NewErrorResponse(apiErr), nil
	}

	workloadType, dbErr := dao.DbGetWorkloadTypeById(ctx, workloadTypeUuid)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	if workloadType.Source == models.WorkloadTypeSourceNative {
		var workloadTypeConfig *api.WorkloadTypeConfig
		workloadTypeConfig, apiErr = GetRunaiWorkloadTypeConfig(ctx, tenantId, workloadType)
		if apiErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get native runai workload type config. err: %v", apiErr)
			return NewErrorResponse(apiErr), nil
		}
		return api.Response(http.StatusOK, *workloadTypeConfig), nil
	}

	workloadTypeConfig, dbErr := dao.DbGetWorkloadTypeConfigByIdWithStatuses(ctx, tenantId, typeId, permittedScopes)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	response := WorkloadTypeConfigDbToApi(*workloadTypeConfig)
	return api.Response(http.StatusOK, response), nil
}

func (s *WorkloadPropertiesService) UpdateWorkloadType(ctx context.Context, workloadTypeId string, updateFields api.WorkloadTypeUpdateFields) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	parsedParams, validationErr := tryParseWorkloadTypeUpdateParams(workloadTypeId, updateFields)
	if validationErr != nil {
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}
	workloadType, apiErr := dao.DbGetWorkloadTypeById(ctx, parsedParams.WorkloadTypeId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	if workloadType.Source == models.WorkloadTypeSourceNative {
		hasRIs := updateFields.ResourceInterfaces != nil && len(*updateFields.ResourceInterfaces) > 0
		hasKartas := updateFields.Kartas != nil && len(*updateFields.Kartas) > 0
		if hasRIs || hasKartas {
			fieldName := "kartas"
			if hasRIs && !hasKartas {
				fieldName = "resource interfaces"
			}
			err := fmt.Errorf("cannot set %s for native workload types", fieldName)
			return api.Response(http.StatusBadRequest, err.Error()), nil
		}

		parsedParams.WorkloadTypeName = workloadType.Name
		_, apiErr = UpdateNativeWorkloadType(ctx, tenantId, *parsedParams)
		if apiErr != nil {
			return NewErrorResponse(apiErr), nil
		}

		// Fetch the complete workload type config from DB to return a full response
		var workloadTypeConfig *api.WorkloadTypeConfig
		workloadTypeConfig, apiErr = GetRunaiWorkloadTypeConfig(ctx, tenantId, workloadType)
		if apiErr != nil {
			return NewErrorResponse(apiErr), nil
		}
		return api.Response(http.StatusOK, *workloadTypeConfig), nil
	}

	validationErr = validateGroupAndKindInKartas(workloadType.Name, workloadType.Group, parsedParams.Kartas)
	if validationErr != nil {
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}

	apiErr = dao.DbUpdateWorkloadTypeWithConfig(ctx, tenantId, *parsedParams)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	workloadTypeConfig, apiErr := dao.DbGetWorkloadTypeConfigById(ctx, tenantId, workloadTypeId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully updated workload type <%s> with categoryId <%s>, priorityId <%s>, preemptibility <%v>; current category name: <%s>, current priority name: <%s>",
		workloadType.Name, updateFields.CategoryId, updateFields.PriorityId, updateFields.Preemptibility, workloadTypeConfig.CategoryName, workloadTypeConfig.PriorityName)

	response := WorkloadTypeConfigDbToApi(*workloadTypeConfig)
	return api.Response(http.StatusOK, response), nil
}

func (s *WorkloadPropertiesService) DeleteWorkloadType(ctx context.Context, workloadTypeId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	workloadTypeUuid, err := uuid.Parse(workloadTypeId)
	if err != nil {
		return api.Response(http.StatusBadRequest, fmt.Sprintf("invalid workload type id %s", workloadTypeId)), nil
	}

	apiErr = validateWorkloadTypeDeleteAllowed(ctx, tenantId, workloadTypeUuid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiErr = dao.DbDeleteWorkloadTypeWithConfig(ctx, tenantId, workloadTypeUuid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully deleted workload type config for id <%s>", workloadTypeId)
	return api.Response(http.StatusNoContent, "Workload type deleted successfully"), nil
}

func (s *WorkloadPropertiesService) SyncWorkloadTypes(ctx context.Context, clusterId string, workloadTypesForSync api.WorkloadTypesForSync) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Syncing workload type statuses status for: %+v", workloadTypesForSync)
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return api.Response(http.StatusBadRequest, fmt.Sprintf("invalid cluster id %s", clusterId)), nil
	}

	workloadTypeStatuses, workloadTypeStatusesToDelete := getWorkloadTypesForStatusUpdate(strconv.Itoa(int(tenantId)), clusterUuid, workloadTypesForSync)

	apiErr = dao.DbSyncWorkloadTypeStatuses(ctx, strconv.Itoa(int(tenantId)), clusterUuid, workloadTypeStatuses, workloadTypeStatusesToDelete, workloadTypesForSync.FullSync)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync workload type statuses. err: %v", apiErr)
		return NewErrorResponse(apiErr), nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (s *WorkloadPropertiesService) GetWorkloadPriorities(ctx context.Context, sortOrder string, sortBy string, filterBy []string) (api.ImplResponse, error) {
	tenantIdInt, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	tenantId := fmt.Sprint(tenantIdInt)
	if permittedScopes != nil && (permittedScopes.System == nil || !*permittedScopes.System) && !slices.Contains(permittedScopes.Tenants, tenantId) {
		return api.Response(http.StatusForbidden, "permission denied"), nil
	}
	filters, filterErr := queryapi.NewParser(models.Priority{}.FilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}
	sortColumn, sortErr := queryapi.NewParser(models.Priority{}.FilterFieldsToColumn()).GetMapping(sortBy)
	if sortErr != nil {
		return ErrorToApiError(sortErr, http.StatusBadRequest), nil
	}

	priorities, err := dao.GetPriorities(ctx, tenantId, sortOrder, sortColumn, filters)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to get priorities", err)), nil
	}

	apiPriorities := []api.WorkloadPriority{}
	if priorities != nil {
		for _, priority := range *priorities {
			apiPriorities = append(apiPriorities, PriorityModelToApi(priority))
		}
	}

	return api.Response(http.StatusOK, api.GetWorkloadPriorities200Response{Priorities: apiPriorities}), nil
}

func (s *WorkloadPropertiesService) CreateWorkloadPriorities(ctx context.Context, priorityForCreate api.PriorityForCreate) (api.ImplResponse, error) {
	return api.Response(http.StatusNotImplemented, nil), nil
}

func (s *WorkloadPropertiesService) GetSyncWorkloadPriorities(ctx context.Context, clusterId string) (api.ImplResponse, error) {
	tenantIdInt, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantId := fmt.Sprint(tenantIdInt)

	priorities, err := dao.GetPrioritiesSync(ctx, tenantId, RunaiSource)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to get priorities", err)), nil
	}
	apiPriorities := []api.SyncPriority{}
	if priorities != nil {
		for _, priority := range *priorities {
			priorityApi := PriorityModelToSyncApi(priority)
			apiPriorities = append(apiPriorities, priorityApi)
		}
	}

	return api.Response(http.StatusOK, apiPriorities), nil
}

func (s *WorkloadPropertiesService) WorkloadPrioritiesStatus(ctx context.Context, clusterId string, prioritiesForStatus api.PrioritiesForStatus) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Syncing priority class status for: %+v", prioritiesForStatus)
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return api.Response(http.StatusBadRequest, fmt.Sprintf("invalid cluster id %s", clusterId)), nil
	}

	prioritiesForUpsert, prioritiesForDelete, failedGet := getPrioritiesForStatusUpdate(ctx, fmt.Sprint(tenantId), clusterUuid, prioritiesForStatus)

	synced, failed, err := dao.SyncRunaiPrioritiesStatus(ctx, clusterUuid, prioritiesForUpsert, prioritiesForDelete, prioritiesForStatus.FullSync)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to sync priorities", err)), nil
	}
	failed = append(failed, failedGet...)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Synced status for: %v", synced)
	if len(failed) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to synced status for: %v", failed)
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (s *WorkloadPropertiesService) WorkloadClusterPrioritiesStatus(ctx context.Context, clusterId string, prioritiesForStatus api.PrioritiesForStatus) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Syncing priority class status for: %+v", prioritiesForStatus)
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return api.Response(http.StatusBadRequest, fmt.Sprintf("invalid cluster id %s", clusterId)), nil
	}

	clusterPrioritiesForUpsert, clusterPrioritiesForDelete := getClusterPrioritiesForStatusUpdate(fmt.Sprint(tenantId), clusterUuid, prioritiesForStatus)

	synced, failed, err := dao.SyncClusterPrioritiesStatus(ctx, fmt.Sprint(tenantId), clusterUuid, clusterPrioritiesForUpsert, clusterPrioritiesForDelete, prioritiesForStatus.FullSync)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to sync priorities", err)), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Synced status for: %v", synced)
	if len(failed) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to synced status for: %v", failed)
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (s *WorkloadPropertiesService) GetWorkloadClusterPriorities(ctx context.Context, clusterId string, sortOrder string, sortBy string, filterBy []string) (api.ImplResponse, error) {
	tenantIdInt, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	tenantId := fmt.Sprint(tenantIdInt)
	if permittedScopes != nil && (permittedScopes.System == nil || !*permittedScopes.System) && !slices.Contains(permittedScopes.Tenants, tenantId) && !slices.Contains(permittedScopes.Clusters, clusterId) {
		return api.Response(http.StatusForbidden, "permission denied"), nil
	}
	filters, filterErr := queryapi.NewParser(models.PriorityClusterView{}.FilterFieldsToColumn()).Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}
	sortColumn, sortErr := queryapi.NewParser(models.PriorityClusterView{}.FilterFieldsToColumn()).GetMapping(sortBy)
	if sortErr != nil {
		return ErrorToApiError(sortErr, http.StatusBadRequest), nil
	}

	priorities, err := dao.GetClusterPriorities(ctx, tenantId, clusterId, sortOrder, sortColumn, filters)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to get priorities", err)), nil
	}

	apiPriorities := []api.ClusterPriority{}
	if priorities != nil {
		for _, priority := range *priorities {
			apiPriorities = append(apiPriorities, PriorityModelToClusterApi(priority))
		}
	}

	return api.Response(http.StatusOK, api.GetWorkloadClusterPriorities200Response{Priorities: apiPriorities}), nil
}

// validateWorkloadTypeCreateAllowed
// the logic for being able to create is complicated -
// create should fail if:
// - trying to create in tenant scope, but exists in system scope (should use "Update" in this case)
// - trying to create in a scope, but it exists in the exact same scope (system or same tenant)
// create should succeed if:
// - trying to create in tenant scope, and it does not exist in any other scope
// - trying to create in tenant scope, while it exists in a different tenant scope
// - trying to create in system scope, and it does not exist in any other scope
// - trying to create in system scope, while it exists in tenant/s scope
//
// so trying to count the workload type configs from view by name/group/version is the best way to validate.
func validateWorkloadTypeCreateAllowed(ctx context.Context, tenantId uint, createFields dao.WorkloadTypeCreateFieldsParsed) *api.Error {
	count, apiErr := dao.DbCountWorkloadTypeConfigsByNameGroup(ctx, tenantId, createFields.WorkloadTypeName, createFields.Group)
	if apiErr != nil {
		return apiErr
	}
	if count > 0 {
		errorMsg := fmt.Sprintf("workload type config with name/group <%s/%s> already exists",
			createFields.WorkloadTypeName, createFields.Group)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errorMsg)
		return api.NewError(http.StatusConflict, errorMsg, nil)
	}

	// for native types - it will not exist in the external_workload_type_config table,
	// so checking this case separately by trying to fetch from workload_types table.
	workloadType, apiErr := dao.DbGetWorkloadTypeByNameGroup(ctx, createFields.WorkloadTypeName, createFields.Group)
	if apiErr != nil {
		if apiErr.Code != http.StatusNotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload type by name <%s>. err: %v", createFields.WorkloadTypeName, apiErr)
			return apiErr
		}
		// NotFound is okay, it means it does not exist in the workload_types table.
	}
	if workloadType != nil && workloadType.Source == models.WorkloadTypeSourceNative {
		// for native types - it will not exist in the external workload type config table,
		// so checking this case separately.
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Workload type config with name <%s> already exists", createFields.WorkloadTypeName)
		return api.NewError(http.StatusConflict, fmt.Sprintf("workload type config with name <%s> already exists", createFields.WorkloadTypeName), nil)
	}

	return nil
}

// validateWorkloadTypeDeleteAllowed
// DELETE workload type is only allowed for:
//   - system scope, only for user_external source (not allowed for external source)
//   - tenant scope, only for user_external and external sources
//   - for either scope, only if a config exists for the specific scope
//     (for example, cannot be deleted for system scope if it exists only in tenant scope)
func validateWorkloadTypeDeleteAllowed(ctx context.Context, tenantId uint, workloadTypeId uuid.UUID) *api.Error {
	workloadType, apiErr := dao.DbGetWorkloadTypeById(ctx, workloadTypeId)
	if apiErr != nil {
		return apiErr
	}

	if workloadType.Source != models.WorkloadTypeSourceCustom {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("workload type <%s> is not a user-defined type and cannot be deleted", workloadTypeId), nil)
	}

	// get the workload config in the given scope, to see if it even can be deleted in the scope
	_, apiErr = dao.DbGetWorkloadTypeConfigByIdForScope(ctx, tenantId, workloadTypeId.String())
	if apiErr != nil {
		if apiErr.Code == http.StatusNotFound {
			return api.NewError(http.StatusForbidden, fmt.Sprintf("workload type <%s> cannot be deleted for the requested scope", workloadTypeId), nil)
		}
		return apiErr
	}

	isMgmtTenant := runaictx.IsManagementTenant(ctx)
	if !isMgmtTenant {
		systemConfigExists, dbErr := dao.DbDoesExternalWorkloadTypeConfigExistForIdForSystemScope(ctx, workloadTypeId.String())
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to check if workload type config exists for id <%s> in system scope. err: %v",
				workloadTypeId, apiErr)
			return dbErr
		}

		if systemConfigExists {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Workload type config with id <%s> exists in system scope, cannot delete", workloadTypeId)
			return api.NewError(http.StatusForbidden, fmt.Sprintf("workload type config with id <%s> exists in system scope, cannot delete in tenant scope", workloadTypeId), nil)
		}
	}

	return nil
}

func tryParseWorkloadTypeParamsInner(requestedCategoryId, requestedPriorityId string,
	requestedPreemptibility *api.WorkloadTypePreemptibility,
	requestedKartas, requestedResourceInterfaces *[]map[string]interface{},
	isUpdate bool) (*dao.WorkloadTypeCreateUpdateFieldsParsed, error) {
	if requestedKartas != nil && requestedResourceInterfaces != nil {
		return nil, errors.New("only one of kartas or resourceInterfaces may be set; resourceInterfaces is deprecated, use kartas")
	}
	// Prefer kartas; fall back to deprecated resourceInterfaces.
	selectedKartas := requestedKartas
	if selectedKartas == nil {
		selectedKartas = requestedResourceInterfaces
	}

	var err error
	result := &dao.WorkloadTypeCreateUpdateFieldsParsed{}

	if isUpdate {
		if requestedCategoryId == "" && requestedPriorityId == "" && !requestedPreemptibility.IsValid() {
			err = errors.Join(err, errors.New("at least one of categoryId, priorityId or preemptibility must be provided"))
		}
	} else {
		if requestedCategoryId == "" || requestedPriorityId == "" {
			err = errors.Join(err, errors.New("both categoryId and priorityId must be provided"))
		}
	}

	if requestedCategoryId != "" {
		categoryUuid, parseErr := uuid.Parse(requestedCategoryId)
		if parseErr != nil {
			err = errors.Join(err, fmt.Errorf("category Id must be a valid uuid. <%s>", requestedCategoryId))
		} else {
			result.CategoryId = ptr.To(categoryUuid)
		}
	}
	if requestedPriorityId != "" {
		priorityUuid, parseErr := uuid.Parse(requestedPriorityId)
		if parseErr != nil {
			err = errors.Join(err, fmt.Errorf("priority Id must be a valid uuid. <%s>", requestedPriorityId))
		} else {
			result.PriorityId = ptr.To(priorityUuid)
		}
	}

	if requestedPreemptibility != nil && (*requestedPreemptibility).IsValid() {
		result.Preemptibility = ptr.To(string(*requestedPreemptibility))
	}

	kartas, parseErr := parseWorkloadTypeKartasApiToDb(selectedKartas)
	if parseErr != nil {
		err = errors.Join(err, parseErr)
	} else {
		result.Kartas = kartas
	}

	if err != nil {
		return nil, err
	}
	return result, nil
}

func tryParseWorkloadTypeUpdateParams(workloadTypeId string, updateFields api.WorkloadTypeUpdateFields) (*dao.WorkloadTypeUpdateFieldsParsed, error) {
	var err error
	var result dao.WorkloadTypeUpdateFieldsParsed

	innerParsed, parseErr := tryParseWorkloadTypeParamsInner(updateFields.CategoryId, updateFields.PriorityId, updateFields.Preemptibility,
		updateFields.Kartas, updateFields.ResourceInterfaces, true)
	if parseErr != nil {
		err = errors.Join(err, parseErr)
	} else {
		result.WorkloadTypeCreateUpdateFieldsParsed = *innerParsed
	}

	workloadTypeUuid, parseErr := uuid.Parse(workloadTypeId)
	if parseErr != nil {
		err = errors.Join(err, fmt.Errorf("workload Id must be a valid uuid. <%s>", workloadTypeId))
	} else {
		result.WorkloadTypeId = workloadTypeUuid
	}

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func tryParseWorkloadTypeCreateParams(createFields api.WorkloadTypeCreateFields) (*dao.WorkloadTypeCreateFieldsParsed, error) {
	var err error
	var result dao.WorkloadTypeCreateFieldsParsed

	var kartasPtr, risPtr *[]map[string]interface{}
	if createFields.Kartas != nil {
		kartasPtr = &createFields.Kartas
	}
	if createFields.ResourceInterfaces != nil {
		risPtr = &createFields.ResourceInterfaces
	}

	innerParsed, parseErr := tryParseWorkloadTypeParamsInner(createFields.CategoryId, createFields.PriorityId, createFields.Preemptibility,
		kartasPtr, risPtr, false)
	if parseErr != nil {
		err = errors.Join(err, parseErr)
	} else {
		result.WorkloadTypeCreateUpdateFieldsParsed = *innerParsed
	}

	if createFields.Name == "" {
		err = errors.Join(err, errors.New("name must be provided"))
	} else {
		result.WorkloadTypeName = createFields.Name
	}

	if createFields.Group == "" {
		err = errors.Join(err, errors.New("group must be provided"))
	} else {
		result.Group = createFields.Group
	}

	groupErr := validateGroupAndKindInKartas(createFields.Name, createFields.Group, result.Kartas)
	if groupErr != nil {
		err = errors.Join(err, groupErr)
	}

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func validateGroupAndKindInKartas(kind, group string, kartas *[]models.Karta) error {
	if kartas == nil {
		return nil
	}

	for _, rd := range *kartas {
		gvk := rd.GroupVersionKind()
		if gvk == nil {
			return errors.New("kartas must include kind in root component")
		}
		if gvk.Group != group {
			return errors.New("group in kartas must match the group of the workload type")
		}
		if gvk.Kind != kind {
			return errors.New("kind in kartas must match the name of the workload type")
		}
	}
	return nil
}

func getPrioritiesForStatusUpdate(ctx context.Context, tenantId string, clusterId uuid.UUID, prioritiesForStatus api.PrioritiesForStatus) ([]models.Priority, []string, []string) {
	prioritiesForUpsert := []models.Priority{}
	prioritiesForDelete := []string{}
	failed := []string{}
	for _, priorityForStatus := range prioritiesForStatus.Priorities {
		if priorityForStatus.Source != RunaiSource {
			continue
		}
		if priorityForStatus.Status.Phase == DeletedPhase {
			prioritiesForDelete = append(prioritiesForDelete, priorityForStatus.Id)
			continue
		}
		id, err := uuid.Parse(priorityForStatus.Id)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get priority %s. err %v", priorityForStatus.Id, err)
			failed = append(failed, priorityForStatus.Id)
			continue
		}
		priority := models.Priority{
			BaseModel: dbmodels.BaseModel{
				ID: id,
			},
			Name:     priorityForStatus.Name,
			Value:    int(priorityForStatus.Value),
			Source:   priorityForStatus.Source,
			TenantId: tenantId,
			Status: []models.PriorityStatus{{
				PriorityID: id,
				ClusterId:  clusterId,
				Phase:      priorityForStatus.Status.Phase,
			}},
		}
		prioritiesForUpsert = append(prioritiesForUpsert, priority)
	}
	return prioritiesForUpsert, prioritiesForDelete, failed
}

func getClusterPrioritiesForStatusUpdate(tenantId string, clusterId uuid.UUID, prioritiesForStatus api.PrioritiesForStatus) ([]models.Priority, []string) {
	clusterPriorities := []models.Priority{}
	deletedClusterPriorities := []string{}
	for _, priorityForStatus := range prioritiesForStatus.Priorities {
		if priorityForStatus.Source == RunaiSource {
			continue
		}
		if priorityForStatus.Status.Phase == DeletedPhase {
			deletedClusterPriorities = append(deletedClusterPriorities, priorityForStatus.Name)
			continue
		}
		priority := models.Priority{
			Name:      priorityForStatus.Name,
			Value:     int(priorityForStatus.Value),
			Source:    priorityForStatus.Source,
			TenantId:  tenantId,
			ScopeType: dao.ClusterScopeType,
			ScopeId:   clusterId.String(),
			Status: []models.PriorityStatus{{
				ClusterId: clusterId,
				Phase:     priorityForStatus.Status.Phase,
			}},
		}
		clusterPriorities = append(clusterPriorities, priority)
	}
	return clusterPriorities, deletedClusterPriorities
}

func getWorkloadTypesForStatusUpdate(tenantId string, clusterId uuid.UUID, workloadTypesForSync api.WorkloadTypesForSync) ([]models.WorkloadTypeStatus, []models.WorkloadTypeStatus) {
	workloadTypeStatuses := []models.WorkloadTypeStatus{}
	workloadTypeIdVersionToDelete := []models.WorkloadTypeStatus{}
	for _, workloadType := range workloadTypesForSync.WorkloadTypes {
		workloadTypeId, err := uuid.Parse(workloadType.Id)
		if err != nil {
			// shouldn't happen, but just in case
			continue
		}

		if workloadType.Phase == DeletedPhase {
			workloadTypeIdVersionToDelete = append(workloadTypeIdVersionToDelete, models.WorkloadTypeStatus{
				WorkloadTypeId: workloadTypeId,
				Version:        workloadType.Version,
			})
			continue
		}

		workloadTypeStatus := models.WorkloadTypeStatus{
			ID:             uuid.New(),
			WorkloadTypeId: workloadTypeId,
			ClusterId:      clusterId,
			TenantId:       tenantId,
			UpdatedAt:      time.Now(),
			Phase:          workloadType.Phase,
			Conditions:     convutils.ConvertSliceTypes(workloadType.Conditions, conditionApiToModel),
			Version:        workloadType.Version,
		}

		workloadTypeStatuses = append(workloadTypeStatuses, workloadTypeStatus)
	}
	return workloadTypeStatuses, workloadTypeIdVersionToDelete
}

// isReadWorkloadPropertiesAllowed -
// checks if the user has permissions to read workload properties (in permitted tenant scope),
// or has permission to read workloads (in any scope that's in the tenant).
func (s *WorkloadPropertiesService) isReadWorkloadPropertiesAllowed(ctx context.Context) (*authorization_api.PermittedScopes, *api.Error) {
	// first, validate workload-properties permissions
	permittedScopes, apiErr := s.evaluateReadPermission(ctx, authorization_api.RESOURCETYPE_WORKLOAD_PROPERTIES)
	if apiErr == nil {
		return permittedScopes, nil
	}

	// now validate workloads permissions
	_, apiErr = s.evaluateReadPermission(ctx, authorization_api.RESOURCETYPE_WORKLOADS)
	return &authorization_api.PermittedScopes{}, apiErr
}

func (s *WorkloadPropertiesService) evaluateReadPermission(ctx context.Context, resourceType authorization_api.ResourceType) (*authorization_api.PermittedScopes, *api.Error) {
	readRequestAction := *authorization_api.NewNullableAction(ptr.To(authorization_api.ACTION_READ))
	evaluateRequest := s.AuthorizationClient.EvaluatePermission(ctx)
	evaluateRequest = evaluateRequest.BackendEvaluationRequest(authorization_api.BackendEvaluationRequest{
		ResourceType:  *authorization_api.NewNullableResourceType(ptr.To(resourceType)),
		RequestAction: readRequestAction,
	})
	permittedScopes, response, err := s.AuthorizationClient.EvaluatePermissionExecute(evaluateRequest)
	if err != nil || (response != nil && response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusOK) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to evaluate permissions for reading %v. err: %v, response: %+v",
			resourceType, err, response)
		return nil, api.NewError(http.StatusForbidden, fmt.Sprintf("Failed to evaluate permissions for reading %v", resourceType), err)
	}

	return permittedScopes, nil
}

func (s *WorkloadPropertiesService) GetAvailablePrioritiesAndCategories(ctx context.Context, clusterId string, sortOrder string, sortBy string, filterBy []string) (api.ImplResponse, error) {
	var combinedResp api.GetAvailablePrioritiesAndCategories200Response

	prioritiesResp, err := s.GetWorkloadClusterPriorities(ctx, clusterId, sortOrder, sortBy, filterBy)
	if err != nil || prioritiesResp.Code != http.StatusOK {
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, "failed to list available priorities", err)), err
	}

	prioritiesRespBody := prioritiesResp.Body.(api.GetWorkloadClusterPriorities200Response)
	combinedResp.Priorities = prioritiesRespBody.Priorities

	categoriesResp, err := s.ListCategories(ctx)
	if err != nil {
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, "failed to list available categories", err)), err
	}

	categoriesRespBody := categoriesResp.Body.(api.ListCategories200Response)
	combinedResp.Categories = categoriesRespBody.Categories

	return api.Response(http.StatusOK, combinedResp), nil
}
