/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/versions"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

// RolesApiService is a service that implements the logic for the RolesApiServicer
// This service should implement the business logic for every endpoint for the RolesApi API.
// Include any external packages or services that will be required by this service.

const (
	systemScopeType                     = "system"
	tenantScopeType                     = "tenant"
	minimumPredefinedRoleClusterVersion = "2.21"
	featureDisabledErrorMsg             = "This feature is currently disabled. To use this endpoint, please enable the corresponding feature flag in your account settings."
)

type RolesApiService struct {
	ClustersClient *cluster_service_client.APIClient
	TenantSettings TenantSettingsService
}

// NewRolesApiService creates a default api service
func NewRolesApiService(clustersClient *cluster_service_client.APIClient, tenantSettings TenantSettingsService) api.RolesAPIServicer {
	return &RolesApiService{
		ClustersClient: clustersClient,
		TenantSettings: tenantSettings,
	}
}

func (s *RolesApiService) GetRole(ctx context.Context, roleId int32) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	role, apiErr := dao.DbGetRoleById(ctx, tenantId, uint(roleId))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	return api.Response(http.StatusOK, RoleDaoToApi(role)), nil
}

func (s *RolesApiService) GetRoles(ctx context.Context, limit int32, offset int32, sortOrder string, sortBy api.RolesSortFilterFields, filterBy []string, search string, includeInternal bool) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	parser := queryapi.NewParser(models.Role{}.SortAndFilterFieldsToColumn())
	filters, filterErr := parser.Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	sortColumn, sortError := parser.GetMapping(string(sortBy))

	if sortError != nil {
		return ErrorToApiError(sortError, http.StatusBadRequest), nil
	}

	roles, apiErr := dao.DbGetRoles(ctx, tenantId, limit, offset, sortOrder, sortColumn, filters, search, includeInternal)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	response := api.GetRoles200Response{
		Roles: []api.Role{},
	}
	for _, role := range *roles {
		response.Roles = append(response.Roles, RoleDaoToApi(&role))
	}

	recordsCount := int32(len(response.Roles))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *RolesApiService) RoleEnable(ctx context.Context, roleId int32, roleEnabledDisabledFields api.RoleEnabledDisabledFields) (api.ImplResponse, error) {
	return s.enableDisableRole(ctx, roleId, roleEnabledDisabledFields.ScopeType, true)
}

func (s *RolesApiService) RoleDisable(ctx context.Context, roleId int32, roleEnabledDisabledFields api.RoleEnabledDisabledFields) (api.ImplResponse, error) {
	return s.enableDisableRole(ctx, roleId, roleEnabledDisabledFields.ScopeType, false)
}

func (s *RolesApiService) RoleActive(ctx context.Context, roleId int32, roleActiveFields api.RoleActiveFields) (api.ImplResponse, error) {
	return s.enableDisableRole(ctx, roleId, roleActiveFields.ScopeType, roleActiveFields.Enabled)
}

func (s *RolesApiService) enableDisableRole(ctx context.Context, roleId int32, scopeType string, enabled bool) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	apiErr := validateActivateRequest(ctx, roleId, scopeType, enabled)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	existingRole, apiErr := dao.DbGetRoleById(ctx, tenantId, uint(roleId))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	// there are 4 options for this API:
	// 1. system role activate/deactivate in system scope
	// 2. tenant role activate/deactivate in system scope
	// 3. system role activate/deactivate in tenant scope
	// 4. tenant role activate/deactivate in tenant scope
	switch scopeType {
	case systemScopeType:
		if existingRole.ScopeType == tenantScopeType {
			// option #2 - tenant role activate/deactivate in system scope - not allowed:
			return NewErrorResponse(api.NewError(http.StatusBadRequest, "A tenant role can not be enabled/disabled in the system scope", nil)), nil
		}
		// option #1 - system role activate/deactivate in system scope:
		apiErr = validateSystemScopeRolePermissions(ctx, tenantId, uint(roleId), existingRole)
		if apiErr != nil {
			return NewErrorResponse(apiErr), nil
		}

		apiErr = dao.DbSetRoleEnabled(ctx, tenantId, roleId, enabled)
	default:
		if existingRole.ScopeType == tenantScopeType {
			// option #4 - tenant role activate/deactivate in tenant scope:
			apiErr = dao.DbSetRoleEnabled(ctx, tenantId, roleId, enabled)
			break
		}
		// option #3 - system role activate/deactivate in tenant scope:
		if enabled {
			if !existingRole.Enabled {
				return NewErrorResponse(api.NewError(http.StatusBadRequest, "Can not enable a disabled system role", nil)), nil
			}

			apiErr = dao.DbSetRoleEnabledPerTenant(ctx, tenantId, uint(roleId))
		} else {
			apiErr = dao.DbSetRoleDisabledPerTenant(ctx, tenantId, uint(roleId))
		}
	}

	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func validateActivateRequest(ctx context.Context, roleId int32, scopeType string, enabled bool) *api.Error {
	if scopeType == systemScopeType {
		apiErr := validateSystemUpdate(ctx)
		if apiErr != nil {
			return apiErr
		}
	}
	if !enabled {
		inUse, apiErr := dao.IsRoleInUse(ctx, uint(roleId))
		if apiErr != nil {
			return apiErr
		}
		if inUse {
			return api.NewError(http.StatusBadRequest, "Can not disable a role that is currently in use", nil)
		}
	}
	return nil
}

func validateSystemScopeRolePermissions(ctx context.Context, tenantId string, roleId uint, role *models.Role) (apiErr *api.Error) {
	permittedScopes, exists := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if permittedScopes == nil || !exists || (permittedScopes.System != nil && *permittedScopes.System) {
		return nil
	}

	if role == nil {
		role, apiErr = dao.DbGetRoleById(ctx, tenantId, roleId)
		if apiErr != nil {
			return apiErr
		}
	}
	if role.ScopeType == tenantScopeType {
		return api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
	}

	return nil
}

func (s *RolesApiService) UpdateRole(ctx context.Context, roleId int32, roleUpdateFieldsInternal api.RoleCreationFieldsInternal) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	apiErr := canEditRole(ctx, tenantId, uint(roleId))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	roleUpdateFieldsInternal, apiErr = validateRoleCreateUpdateAndSetRoleDefaults(ctx, s.ClustersClient.ClustersAPI, tenantId, roleUpdateFieldsInternal, ptr.To(uint(roleId)))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	roleSpec, err := creationRequestToRoleDao(roleUpdateFieldsInternal, ptr.To(tenantId))
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	role, apiErr := dao.DbUpdateRole(ctx, tenantId, roleId, roleSpec)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	return api.Response(http.StatusOK, RoleDaoToApi(role)), nil
}

// CreateRole - Create role
func (s *RolesApiService) CreateRole(ctx context.Context, roleCreationFieldsInternal api.RoleCreationFieldsInternal) (api.ImplResponse, error) {
	tenantId, subjectId, _, err := runaictx.ParseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	roleCreationFieldsInternal, apiErr := validateRoleCreateUpdateAndSetRoleDefaults(ctx, s.ClustersClient.ClustersAPI, tenantId, roleCreationFieldsInternal, nil)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	roleSpec, err := creationRequestToRoleDao(roleCreationFieldsInternal, ptr.To(tenantId))
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	role, apiErr := dao.DbCreateRole(ctx, subjectId, roleSpec)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	return api.Response(http.StatusCreated, RoleDaoToApi(role)), nil
}

func creationRequestToRoleDao(role api.RoleCreationFieldsInternal, tenantId *string) (models.Role, error) {
	enabled := true
	if role.Enabled != nil {
		enabled = *role.Enabled
	}
	modelRole := models.Role{
		Name:                    role.Name,
		Description:             role.Description,
		ScopeType:               apiRoleScopeToModel(role.ScopeType),
		ScopeId:                 role.ScopeId,
		TenantId:                tenantId,
		ScopeName:               nil,
		Custom:                  true,
		Enabled:                 enabled,
		ClusterRolePredefinedId: StringToInt64Ptr(role.KubernetesPermissions.PredefinedRole),
	}
	if len(role.PermissionSets) > 0 {
		modelRole.NoPermissionSet = ptr.To(false)
		permissionSetsDao, err := PermissionSetsApiToDao(role.PermissionSets)
		if err != nil {
			return modelRole, err
		}
		modelRole.PermissionSets = &permissionSetsDao
	} else {
		modelRole.NoPermissionSet = ptr.To(true)
		modelRole.Permissions = PermissionsApiToDao(role.Permissions)
	}
	return modelRole, nil
}

func apiRoleScopeToModel(apiRoleScope string) models.RoleScope {
	if apiRoleScope == systemScopeType {
		return models.ROLE_SCOPE_SYSTEM
	}
	return models.ROLE_SCOPE_TENANT
}

func StringToInt64Ptr(s *string) *int64 {
	if s == nil {
		return nil
	}

	val, err := strconv.ParseInt(*s, 10, 64)
	if err != nil {
		return nil
	}

	return &val
}

func PermissionsApiToDao(permissions []api.PostPermission, roleId ...uint) []*models.Permission {
	result := []*models.Permission{}
	for _, permission := range permissions {
		actionMap := map[string]bool{}
		for _, action := range permission.Actions {
			actionMap[strings.ToLower(string(action))] = true
		}

		permissionDao := &models.Permission{
			ResourceType: permission.ResourceType,
			Create:       actionMap[string(api.ACTION_CREATE)],
			Read:         actionMap[string(api.ACTION_READ)],
			Update:       actionMap[string(api.ACTION_UPDATE)],
			Delete:       actionMap[string(api.ACTION_DELETE)],
			Sync:         actionMap[string(api.ACTION_SYNC)],
		}

		// No optional parameters in golang, this is the general workaround
		if len(roleId) > 0 {
			permissionDao.RoleID = roleId[0]
		}

		result = append(result, permissionDao)
	}

	return result
}

func PermissionSetsApiToDao(permissionSets []api.RolePermissionSet) ([]models.PermissionSet, error) {
	result := []models.PermissionSet{}
	for _, permissionSet := range permissionSets {
		permissionSetId, err := uuid.Parse(permissionSet.Id)
		if err != nil {
			return result, err
		}
		result = append(result, models.PermissionSet{
			DeleteBaseModel: dbmodels.DeleteBaseModel{
				ID: permissionSetId,
			},
			Name: permissionSet.Name,
		})
	}

	return result, nil
}

type PermissionActions struct {
	ResourceType api.ResourceType
	Read         bool
	Create       bool
	Delete       bool
	Update       bool
}

func createPermissionsActionsFromPermissions(permissions []api.PostPermission) map[string]PermissionActions {
	customRolePermissionsMap := map[string]PermissionActions{}

	for _, customRolePermission := range permissions {
		permissionActions := PermissionActions{
			ResourceType: customRolePermission.ResourceType,
		}
		for _, action := range customRolePermission.Actions {
			switch action {
			case api.ACTION_READ:
				permissionActions.Read = true
			case api.ACTION_CREATE:
				permissionActions.Create = true
			case api.ACTION_UPDATE:
				permissionActions.Update = true
			case api.ACTION_DELETE:
				permissionActions.Delete = true
			}
		}
		customRolePermissionsMap[string(customRolePermission.ResourceType)] = permissionActions
	}
	return customRolePermissionsMap
}

func createPermissionsActionsFromPermissionSets(ctx context.Context, apiPermissionSets []api.RolePermissionSet) (map[string]PermissionActions, *api.Error) {
	customRolePermissionsMap := map[string]PermissionActions{}
	permissionSetsIds := []string{}
	for _, ps := range apiPermissionSets {
		permissionSetsIds = append(permissionSetsIds, ps.Id)
	}
	permissionSets, err := dao.DbGetPermissionSetsByIds(ctx, permissionSetsIds)
	if err != nil {
		return customRolePermissionsMap, err
	}
	for _, ps := range *permissionSets {
		for _, permission := range ps.Permissions {
			actions, found := customRolePermissionsMap[string(permission.ResourceType)]
			if !found {
				actions = PermissionActions{ResourceType: permission.ResourceType}
			}
			if permission.Create {
				actions.Create = true
			}
			if permission.Update {
				actions.Update = true
			}
			if permission.Read {
				actions.Read = true
			}
			if permission.Delete {
				actions.Delete = true
			}
			customRolePermissionsMap[string(permission.ResourceType)] = actions
		}
	}

	return customRolePermissionsMap, nil
}

func validateKubernetesPermissions(ctx context.Context, kubernetesPermissions api.KubernetesPermissions, permissionSets []api.RolePermissionSet, permissions []api.PostPermission, tenantId string) *api.Error {
	predefinedRoleId := ptr.Deref(kubernetesPermissions.PredefinedRole, "")
	if predefinedRoleId != "" {
		intValue, err := strconv.ParseInt(predefinedRoleId, 10, 64)
		if err != nil {
			return api.NewError(http.StatusBadRequest, "KubernetesPermissions predefined role id invalid", nil)
		}
		predefinedRole, apiErr := dao.DbGetRoleById(ctx, tenantId, uint(intValue))
		if apiErr != nil {
			return apiErr
		}

		var customRolePermissionsMap map[string]PermissionActions

		if len(permissionSets) > 0 && len(permissions) == 0 {
			customRolePermissionsMap, apiErr = createPermissionsActionsFromPermissionSets(ctx, permissionSets)
			if apiErr != nil {
				return apiErr
			}
		} else {
			customRolePermissionsMap = createPermissionsActionsFromPermissions(permissions)
		}
		predefinedRolePermissionsMap := map[string]PermissionActions{}
		for _, predefinedRolePermission := range predefinedRole.Permissions {
			if predefinedRolePermission != nil {
				predefinedRolePermissionsMap[string(predefinedRolePermission.ResourceType)] = PermissionActions{
					ResourceType: predefinedRolePermission.ResourceType,
					Read:         predefinedRolePermission.Read,
					Create:       predefinedRolePermission.Create,
					Update:       predefinedRolePermission.Update,
					Delete:       predefinedRolePermission.Delete,
				}
			}
		}
		for _, resourceType := range api.AllowedResourceTypeEnumValues {
			customPermissions := customRolePermissionsMap[string(resourceType)]
			predefinedPermission := predefinedRolePermissionsMap[string(resourceType)]
			apiErr = validateResourceTypeKubernetesPermissions(predefinedPermission, customPermissions)
			if apiErr != nil {
				return apiErr
			}
		}
	}
	return nil
}

func validateResourceTypeKubernetesPermissions(preDefinedPermission PermissionActions, customRolePermission PermissionActions) *api.Error {
	errMsg := "custom role invalid since KubernetesPermissions predefined role has %s permission on %s and custom role does not"
	if preDefinedPermission.Read && !customRolePermission.Read {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf(errMsg, "read", string(preDefinedPermission.ResourceType)), nil)
	}
	if preDefinedPermission.Create && !customRolePermission.Create {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf(errMsg, "create", string(preDefinedPermission.ResourceType)), nil)
	}
	if preDefinedPermission.Update && !customRolePermission.Update {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf(errMsg, "update", string(preDefinedPermission.ResourceType)), nil)
	}
	if preDefinedPermission.Delete && !customRolePermission.Delete {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf(errMsg, "delete", string(preDefinedPermission.ResourceType)), nil)
	}
	return nil
}

func validateInputPermissions(ctx context.Context, permissions []api.PostPermission, permissionSets []api.RolePermissionSet) *api.Error {
	hasPermissions := len(permissions) > 0
	hasPermissionSets := len(permissionSets) > 0

	if !hasPermissions && !hasPermissionSets {
		return api.NewValidationError("A role must contain at least one permission or permission set")
	}
	if hasPermissions && hasPermissionSets {
		return api.NewValidationError("A role must contain permission sets or permissions")
	}

	if hasPermissionSets {
		return validatePermissionSetsExist(ctx, permissionSets)
	}
	// else - hasPermissions
	nonEmptyPermissions := 0

	seenResources := map[api.ResourceType]bool{}
	for _, permission := range permissions {
		if seenResources[permission.ResourceType] {
			return api.NewValidationError("Got multiple definitions of permission for the same resource type")
		}
		seenResources[permission.ResourceType] = true

		if len(permission.Actions) > 0 {
			nonEmptyPermissions++
		}
	}
	if nonEmptyPermissions == 0 {
		return api.NewValidationError("Permissions must contain at least one action")
	}
	return nil
}

func setRoleDefaults(tenantId string, roleCreationRequest api.RoleCreationFieldsInternal) api.RoleCreationFieldsInternal {
	if roleCreationRequest.ScopeType == systemScopeType {
		roleCreationRequest.ScopeId = nil
		return roleCreationRequest
	}
	if roleCreationRequest.ScopeType == "" {
		roleCreationRequest.ScopeType = tenantScopeType
	}
	if roleCreationRequest.ScopeId == nil {
		roleCreationRequest.ScopeId = ptr.To(tenantId)
	}
	if roleCreationRequest.Enabled == nil {
		roleCreationRequest.Enabled = ptr.To(true)
	}
	return roleCreationRequest
}

func validateRoleCreateUpdate(ctx context.Context, clustersClient cluster_service_client.ClustersAPI, tenantId string, roleCreationRequest api.RoleCreationFieldsInternal) *api.Error {
	if roleCreationRequest.ScopeType == systemScopeType {
		return validateSystemUpdate(ctx)
	}
	// validate tenant scope role
	if roleCreationRequest.ScopeType != tenantScopeType {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("ScopeType %s invalid", roleCreationRequest.ScopeType), nil)
	}
	if *roleCreationRequest.ScopeId != tenantId {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("ScopeId %s is different than subject tenant id", *roleCreationRequest.ScopeId), nil)
	}
	predefinedRoleId := ptr.Deref(roleCreationRequest.KubernetesPermissions.PredefinedRole, "")
	if predefinedRoleId != "" {
		return validateClusterVersions(ctx, clustersClient)
	}
	return nil
}

func validateClusterVersions(ctx context.Context, clustersClient cluster_service_client.ClustersAPI) *api.Error {
	clustersGetRequest := clustersClient.GetClusters(ctx)
	execute, response, err := clustersClient.GetClustersExecute(clustersGetRequest)
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to retrieve clusters version, error: %s", err.Error()), err)
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to retrieve clusters version, response %d", response.StatusCode), nil)
	}
	clusterVersions := []string{}
	for _, cluster := range execute {
		version := cluster.Version.Get()
		if version != nil && *version != "" {
			clusterVersions = append(clusterVersions, *version)
		}
	}
	if len(clusterVersions) == 0 {
		return nil
	}
	clustersSupported, err := versions.AreAllVersionsSupported(clusterVersions, minimumPredefinedRoleClusterVersion)
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to retrieve clusters version, error: %s", err.Error()), err)
	}
	if !clustersSupported {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("Predifined kubernetes role is supported for cluster version 2.21 and above. Some versions in %v are not supported", clusterVersions), err)
	}
	return nil
}

func validateSystemUpdate(ctx context.Context) *api.Error {
	if !runaictx.IsManagementTenant(ctx) {
		return api.NewError(http.StatusForbidden, "Subject is not authorized to create system scoped roles", nil)
	}

	permittedScopes, _ := ctx.Value(authorization_utils.PermittedScopesKey).(*authorization_api.PermittedScopes)
	if permittedScopes != nil && (permittedScopes.System == nil || !*permittedScopes.System) {
		return api.NewError(http.StatusForbidden, "Subject is not authorized to create system scoped roles", nil)
	}
	return nil
}

func validateRoleCreateUpdateAndSetRoleDefaults(
	ctx context.Context,
	clustersClient cluster_service_client.ClustersAPI,
	tenantId string,
	roleCreationRequest api.RoleCreationFieldsInternal,
	existingRoleId *uint,
) (api.RoleCreationFieldsInternal, *api.Error) {
	roleCreationRequest = setRoleDefaults(tenantId, roleCreationRequest)
	err := validateRoleCreateUpdate(ctx, clustersClient, tenantId, roleCreationRequest)
	if err != nil {
		return roleCreationRequest, err
	}
	err = validateInputPermissions(ctx, roleCreationRequest.Permissions, roleCreationRequest.PermissionSets)
	if err != nil {
		return roleCreationRequest, err
	}
	err = validateKubernetesPermissions(ctx, roleCreationRequest.KubernetesPermissions, roleCreationRequest.PermissionSets, roleCreationRequest.Permissions, tenantId)
	if err != nil {
		return roleCreationRequest, err
	}
	if existingRoleId != nil {
		scopeTypeNotChanged, apiErr := dao.IsRoleInScopeType(ctx, *existingRoleId, roleCreationRequest.ScopeType)
		if apiErr != nil {
			return roleCreationRequest, apiErr
		}
		if !scopeTypeNotChanged {
			return roleCreationRequest, api.NewError(http.StatusBadRequest, "Can not update role's scope type", nil)
		}
	}
	return roleCreationRequest, nil
}

func validatePermissionSetsExist(ctx context.Context, permissionSets []api.RolePermissionSet) *api.Error {
	if len(permissionSets) == 0 {
		return nil
	}

	ids := make([]string, len(permissionSets))
	for i, ps := range permissionSets {
		ids[i] = ps.Id
	}

	existingPermissionSets, err := dao.DbGetPermissionSetsByIds(ctx, ids)
	if err != nil {
		return err
	}
	permMap := make(map[string]string)

	for _, p := range *existingPermissionSets {
		// internal permission sets are created and owned by the service on behalf of a single
		// role, so they are not attachable by id. Skipping them here makes them indistinguishable
		// from a non-existing id, which also avoids leaking whether one exists.
		if p.Internal {
			continue
		}
		permMap[p.ID.String()] = p.Name
	}

	invalidPermissionSets := []api.RolePermissionSet{}
	for _, ps := range permissionSets {
		name, found := permMap[ps.Id]
		if !found || (ps.Name != "" && name != ps.Name) {
			invalidPermissionSets = append(invalidPermissionSets, ps)
		}
	}

	if len(invalidPermissionSets) > 0 {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("Permission sets not found: %v", invalidPermissionSets), nil)
	}

	return nil
}

// DeleteRole - Delete role
func (s *RolesApiService) DeleteRole(ctx context.Context, roleId int32) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	apiErr := canEditRole(ctx, tenantId, uint(roleId))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiErr = dao.DbDeleteRole(ctx, tenantId, uint(roleId))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	return NoContentResponse(), nil
}

func canEditRole(ctx context.Context, tenantId string, roleId uint) *api.Error {
	permittedScopes, exists := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if permittedScopes != nil && exists && (permittedScopes.System == nil || !*permittedScopes.System) {
		role, apiErr := dao.DbGetRoleById(ctx, tenantId, roleId)
		if apiErr != nil {
			return apiErr
		}

		if !role.Custom || role.ScopeType == models.ROLE_SCOPE_SYSTEM {
			return api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
		}
	}

	inUse, apiErr := dao.IsRoleInUse(ctx, roleId)
	if apiErr != nil {
		return apiErr
	}
	if inUse {
		return api.NewError(http.StatusBadRequest, "Can not update or delete a role that is currently in use", nil)
	}

	return nil
}

// GetRoleV1 - Get role
func (s *RolesApiService) GetRoleV1(ctx context.Context, roleId int32) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	role, apiErr := dao.DbGetRoleById(ctx, tenantId, uint(roleId))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	return api.Response(http.StatusOK, RoleDaoToV1Api(role)), nil
}

// GetRolesV1 - Get roles list
func (s *RolesApiService) GetRolesV1(ctx context.Context, includeInternal bool) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	roles, apiErr := dao.DbGetRoles(ctx, tenantId, 0, 0, "", "", nil, "", includeInternal)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	resultRoles := []api.RoleV1{}
	for _, role := range *roles {
		resultRoles = append(resultRoles, RoleDaoToV1Api(&role))
	}
	return api.Response(http.StatusOK, resultRoles), nil
}
