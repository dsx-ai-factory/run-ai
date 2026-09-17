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

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/cache"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/pubsub"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/structs"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

// InternalApiService is a service that implements the logic for the InternalApiServicer
// This service should implement the business logic for every endpoint for the SubjectAccessRulesApi API.
// Include any external packages or services that will be required by this service.
type InternalApiService struct {
	OrgHierarchyManager *cache.OrgHierarchyManager
	TenantSettings      TenantSettingsService
}

func NewInternalApiService(tenantSettings TenantSettingsService) api.InternalAPIServicer {
	internalService := &InternalApiService{
		OrgHierarchyManager: cache.GetOrgHierarchyManager(),
		TenantSettings:      tenantSettings,
	}

	err := pubsub.Register(internalService)
	if err != nil {
		logging.Warnf("Failed to register internal service to pubsub server. error: %v", err)
	}

	return internalService
}

func (i *InternalApiService) GetGroups(ctx context.Context) (api.ImplResponse, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	groups, err := dao.DbGetGroupNames(ctx, tenantId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	response := api.GetGroups200Response{
		Names: groups,
	}

	return api.Response(http.StatusOK, response), nil
}

func (i *InternalApiService) GetSubjectAccessRules(ctx context.Context, subjectType string, subjectIds []string) (api.ImplResponse, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	accessRules, err := dao.DbGetAccessRules(ctx, tenantId, nil, subjectType, subjectIds, 0, 0, "", false, "",
		"", "", "", nil, "")
	if err != nil {
		return NewErrorResponse(err), nil
	}

	accessRulesResponse := []api.AccessRule{}
	for _, accessRule := range *accessRules {
		accessRulesResponse = append(accessRulesResponse, AccessRuleDaoToApi(&accessRule))
	}
	return api.Response(http.StatusOK, accessRulesResponse), nil
}

func (i *InternalApiService) DeleteSubjectAccessRules(ctx context.Context, subjectType string, subjectId string, dryRun bool) (api.ImplResponse, error) {
	tenantId, requesterId, groups, err := parseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	accessRules, err := dao.DbGetSubjectPermissions(
		ctx,
		tenantId,
		subjectId,
		[]string{},
	)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	err = validateCurrentUserHasSufficientPrivileges(ctx, tenantId, requesterId, groups, accessRules)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	if !dryRun {
		err = dao.DbDeleteSubjectAccessRules(ctx, tenantId, subjectType, subjectId)
		if err != nil {
			return NewErrorResponse(err), nil
		}
	}

	return NoContentResponse(), nil
}

func (i *InternalApiService) DeleteScopeAccessRules(ctx context.Context, scopeType string, scopeId string, lastUpdated string, roles string) (api.ImplResponse, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	rolesList := normalizeList(roles)

	isDGXC := i.TenantSettings.IsDGXCManaged(ctx)

	err = dao.DbDeleteScopeAccessRules(ctx, isDGXC, tenantId, scopeType, scopeId, lastUpdated, rolesList)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return NoContentResponse(), nil
}

func (i *InternalApiService) UpdateScopeAccessRules(ctx context.Context, scopeType string, scopeId string, updateRequest api.UpdateScopeAccessRulesRequest) (api.ImplResponse, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	err = dao.DbUpdateScopeAccessRules(ctx, tenantId, scopeType, scopeId, updateRequest)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return NoContentResponse(), nil
}

func (i *InternalApiService) CreateInternalAccessRule(ctx context.Context, request api.InternalCreateAccessRuleRequest) (api.ImplResponse, error) {
	return i.createInternalAccessRule(ctx, request, dao.MapStaticRoleNameToRoleId[string(request.RoleName)])
}

func (i *InternalApiService) GetScopeAccessRules(ctx context.Context, clusterId string, roles string, departmentIds string, projectIds string) (api.ImplResponse, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("running GetScopeAccessRules for roles %v tenant %d cluster %s departments %v projects %v", roles, tenantId, clusterId, departmentIds, projectIds)

	rolesList := normalizeList(roles)
	departmentIdsList := normalizeList(departmentIds)
	projectIdsList := normalizeList(projectIds)

	isDGXC := i.TenantSettings.IsDGXCManaged(ctx)

	accessRules, err := dao.DbGetAccessRulesByScopes(ctx, isDGXC, rolesList, tenantId, clusterId, departmentIdsList, projectIdsList)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	countAccessRules := int32(len(*accessRules))
	accessRulesResponse := api.GetAccessRules200Response{
		TotalRecords:   countAccessRules,
		DisplayRecords: countAccessRules,
	}

	for _, accessRule := range *accessRules {
		accessRulesResponse.AccessRules = append(accessRulesResponse.AccessRules, AccessRuleDaoToApi(&accessRule))
	}
	return api.Response(http.StatusOK, accessRulesResponse), nil
}

func (i *InternalApiService) createInternalAccessRule(ctx context.Context, request api.InternalCreateAccessRuleRequest, roleId uint) (api.ImplResponse, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	var scopeType api.ScopeType
	var scopeId, clusterUuid, scopeName string
	if request.Scope != nil {
		switch request.Scope.Type {
		case api.SCOPETYPE_CLUSTER:
			clusterUuid = request.Scope.Id
		case api.SCOPETYPE_TENANT, api.SCOPETYPE_SYSTEM:
		default:
			// At the moment we have no use case for other scope types for internal access rules
			return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Unsupported cluster app scope type %s", request.Scope.Type), nil)), nil
		}

		scopeId = request.Scope.Id
		scopeType = request.Scope.Type
		scopeName = request.Scope.Name
	} else {
		scopeType = api.SCOPETYPE_TENANT
		scopeId = fmt.Sprintf("%d", tenantId)
	}

	if scopeName == "" {
		scopeName, err = i.OrgHierarchyManager.GetScopeName(ctx, scopeType, scopeId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get scope name for scope type %s scope id %s: %v", scopeType, scopeId, err)
			scopeName = ""
		}
	}

	accessRule, err := dao.DbCreateInternalUserAccessRule(ctx, request, tenantId, roleId, scopeType, scopeId, scopeName, clusterUuid)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusCreated, AccessRuleDaoToApi(accessRule)), nil
}

// PrivilegeEscalationEvaluate - Should only be internal
func (i *InternalApiService) PrivilegeEscalationEvaluate(ctx context.Context, subjectId string, groupIds []string) (api.ImplResponse, error) {
	tenantId, requesterId, requesterGroups, err := parseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	targetAccessRules, err := dao.DbGetSubjectPermissions(ctx, tenantId, subjectId, groupIds)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	if targetAccessRules == nil || len(*targetAccessRules) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("PrivilegeEscalationEvaluate: no access rules found for subject %s", subjectId)
	}

	err = validateCurrentUserHasSufficientPrivileges(ctx, tenantId, requesterId, requesterGroups, targetAccessRules)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return NoContentResponse(), nil
}

func normalizeList(stringList string) []string {
	list := strings.Split(stringList, ",")
	if list != nil && (len(list) == 0 || (len(list) == 1 && list[0] == "")) {
		list = nil
	}
	return list
}

func validateCurrentUserHasPrivilege(ctx context.Context, tenantId uint, subjectId string, groups []string, accessRule models.AccessRule) *api.Error {
	return validateCurrentUserHasSufficientPrivileges(ctx, tenantId, subjectId, groups, ptr.To([]models.AccessRule{accessRule}))
}

func validateCurrentUserHasSufficientPrivileges(ctx context.Context, tenantId uint, subjectId string, groups []string, accessRules *[]models.AccessRule) *api.Error {
	requesterAccessRules, err := dao.DbGetSubjectPermissions(ctx, tenantId, subjectId, groups)
	if err != nil {
		return err
	}

	orgTree, _, orgErr := cache.GetOrgHierarchyManager().GetOrgTree(ctx)
	if orgErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant org tree. err: %v", orgErr.Error())
		return api.NewError(http.StatusInternalServerError, orgErr.Error(), orgErr)
	}
	return validateSufficientPrivilegesForRequester(ctx, requesterAccessRules, accessRules, &orgTree)
}

func validateSufficientPrivilegesForRequester(ctx context.Context, requesterAccessRules *[]models.AccessRule, accessRules *[]models.AccessRule, orgTree *structs.OrgUnitMap) *api.Error {
	requesterPermissionsMap, err := constructPermissionMap(ctx, *requesterAccessRules)
	if err != nil {
		return err
	}
	targetPermissionsMap, err := constructPermissionMap(ctx, *accessRules)
	if err != nil {
		return err
	}
	return validatePermissionsMapContains(ctx, requesterPermissionsMap, targetPermissionsMap, orgTree)
}

func validatePermissionsMapContains(
	ctx context.Context,
	requesterPermissionsMap,
	targetPermissionsMap map[api.ResourceType]map[api.Action]api.PermittedScopes,
	mapChildToParent *structs.OrgUnitMap,
) *api.Error {
	for _, resourceType := range api.AllowedResourceTypeEnumValues {
		requesterPermissions, requesterOk := requesterPermissionsMap[resourceType]
		targetPermissions, targetOk := targetPermissionsMap[resourceType]
		if !targetOk {
			continue
		}
		if !requesterOk && len(targetPermissions) > 0 {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Requester does'nt have any permission on resource: %s", resourceType)
			return api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
		}

		for _, action := range api.AllowedActionEnumValues {
			requesterScopes, requesterOk := requesterPermissions[action]
			targetScopes, targetOk := targetPermissions[action]
			if !targetOk {
				continue
			}
			if !requesterOk {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Requester does'nt have permission on resource and action: %s-%s", resourceType, action)
				return api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
			}
			if dao.IsGlobalResource(resourceType) {
				continue
			}

			err := validateScopesContains(ctx, requesterScopes, targetScopes, mapChildToParent)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func validateScopesContains(ctx context.Context, requesterScopes, targetScopes api.PermittedScopes, mapChildToParent *structs.OrgUnitMap) *api.Error {
	if requesterScopes.System {
		return nil
	}
	if targetScopes.System {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Requester does'nt have permission on system scope")
		return api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
	}

	requesterTenants := map[string]bool{}
	for _, tenant := range requesterScopes.Tenants {
		requesterTenants[tenant] = true
	}
	requesterClusters := map[string]bool{}
	for _, cluster := range requesterScopes.Clusters {
		requesterClusters[cluster] = true
	}
	requesterDepartments := map[string]bool{}
	for _, department := range requesterScopes.Departments {
		requesterDepartments[department] = true
	}
	requesterProjects := map[string]bool{}
	for _, proj := range requesterScopes.Projects {
		requesterProjects[proj] = true
	}

	for _, tenant := range targetScopes.Tenants {
		if !requesterTenants[tenant] {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Requester does'nt have permission on tenant %s", tenant)
			return api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
		}
	}

	for _, cluster := range targetScopes.Clusters {
		current := structs.OrgUnitKey{Scope: string(openapi.SCOPETYPE_CLUSTER), Id: cluster}
		parent, found := (*mapChildToParent)[current]
		if !requesterClusters[cluster] &&
			(!found || !requesterTenants[parent.Id]) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Requester does'nt have permission on cluster %s", cluster)
			return api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
		}
	}

	for _, department := range targetScopes.Departments {
		current := structs.OrgUnitKey{Scope: string(openapi.SCOPETYPE_DEPARTMENT), Id: department}
		clusterParent, clusterFound := (*mapChildToParent)[current]
		tenantParent, tenantFound := (*mapChildToParent)[clusterParent]
		if !requesterDepartments[department] &&
			(!clusterFound || !requesterClusters[clusterParent.Id]) &&
			(!tenantFound || !requesterTenants[tenantParent.Id]) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Requester does'nt have permission on department %s or above", department)
			return api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
		}
	}

	for _, project := range targetScopes.Projects {
		current := structs.OrgUnitKey{Scope: string(openapi.SCOPETYPE_PROJECT), Id: project}
		depParent, depFound := (*mapChildToParent)[current]
		clusterParent, clusterFound := (*mapChildToParent)[depParent]
		tenantParent, tenantFound := (*mapChildToParent)[clusterParent]
		if !requesterProjects[project] &&
			(!depFound || !requesterDepartments[depParent.Id]) &&
			(!clusterFound || !requesterClusters[clusterParent.Id]) &&
			(!tenantFound || !requesterTenants[tenantParent.Id]) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Requester does'nt have permission on project %s or above", project)
			return api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
		}
	}

	return nil
}

func getAccessRulePermissionSets(ctx context.Context, accessRule models.AccessRule) (*[]models.PermissionSet, *api.Error) {
	if accessRule.Role.PermissionSets != nil && len(*accessRule.Role.PermissionSets) > 0 {
		return accessRule.Role.PermissionSets, nil
	}
	// fetch from db
	return dao.DbGetPermissionSetsByRoleId(ctx, accessRule.RoleID)
}

func constructPermissionMap(ctx context.Context, accessRules []models.AccessRule) (map[api.ResourceType]map[api.Action]api.PermittedScopes, *api.Error) {
	permissionsMap := map[api.ResourceType]map[api.Action]api.PermittedScopes{}
	for _, accessRule := range accessRules {
		permissionSets, err := getAccessRulePermissionSets(ctx, accessRule)
		if err != nil {
			return permissionsMap, err
		}
		for _, permissionSet := range *permissionSets {
			for _, permission := range permissionSet.Permissions {
				actionToPermittedScopes, found := permissionsMap[permission.ResourceType]
				if !found {
					actionToPermittedScopes = map[api.Action]api.PermittedScopes{}
					permissionsMap[permission.ResourceType] = actionToPermittedScopes
				}
				if permission.Create {
					actionToPermittedScopes[api.ACTION_CREATE] = appendPermittedScope(accessRule.ScopeId, accessRule.ScopeType, actionToPermittedScopes[api.ACTION_CREATE])
				}
				if permission.Read {
					actionToPermittedScopes[api.ACTION_READ] = appendPermittedScope(accessRule.ScopeId, accessRule.ScopeType, actionToPermittedScopes[api.ACTION_READ])
				}
				if permission.Update {
					actionToPermittedScopes[api.ACTION_UPDATE] = appendPermittedScope(accessRule.ScopeId, accessRule.ScopeType, actionToPermittedScopes[api.ACTION_UPDATE])
				}
				if permission.Delete {
					actionToPermittedScopes[api.ACTION_DELETE] = appendPermittedScope(accessRule.ScopeId, accessRule.ScopeType, actionToPermittedScopes[api.ACTION_DELETE])
				}
				if permission.Sync {
					actionToPermittedScopes[api.ACTION_SYNC] = appendPermittedScope(accessRule.ScopeId, accessRule.ScopeType, actionToPermittedScopes[api.ACTION_SYNC])
				}

				permissionsMap[permission.ResourceType] = actionToPermittedScopes
			}
		}
	}
	// remove duplicates
	for resourceType, actionMap := range permissionsMap {
		for action, permittedScopes := range actionMap {
			newPermittedScopes := api.PermittedScopes{
				System:      permittedScopes.System,
				Tenants:     dao.RemoveDuplicates(permittedScopes.Tenants),
				Clusters:    dao.RemoveDuplicates(permittedScopes.Clusters),
				Departments: dao.RemoveDuplicates(permittedScopes.Departments),
				Projects:    dao.RemoveDuplicates(permittedScopes.Projects),
			}
			permissionsMap[resourceType][action] = newPermittedScopes
		}
	}

	return permissionsMap, nil
}

func appendPermittedScope(scopeId string, scopeType api.ScopeType, permittedScopes api.PermittedScopes) api.PermittedScopes {
	switch scopeType {
	case api.SCOPETYPE_SYSTEM:
		permittedScopes.System = true
	case api.SCOPETYPE_TENANT:
		permittedScopes.Tenants = append(permittedScopes.Tenants, scopeId)
	case api.SCOPETYPE_CLUSTER:
		permittedScopes.Clusters = append(permittedScopes.Clusters, scopeId)
	case api.SCOPETYPE_DEPARTMENT:
		permittedScopes.Departments = append(permittedScopes.Departments, scopeId)
	case api.SCOPETYPE_PROJECT:
		permittedScopes.Projects = append(permittedScopes.Projects, scopeId)
	}

	return permittedScopes
}
