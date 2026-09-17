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
	"runtime"
	"slices"
	"strconv"
	"sync"

	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/cache"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	static_legacy_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/structs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
)

const (
	MaxBatchDeletionSize         = 100
	MaxBatchCreationScopesSize   = 10
	MaxBatchCreationSubjectsSize = 10
	MaxDeprecatedMigrationSize   = 100
	MinimalCountOfChars          = 2
	AuditEntityType              = "access rule"

	// authenticatedUsersK8sGroup is the built-in k8s group that covers every authenticated user
	// and service account; old clusters that can't parse all-authenticated-users get this instead.
	authenticatedUsersK8sGroup = "system:authenticated"
)

// AccessRulesApiService is a service that implements the logic for the AccessRuleApiServicer
// This service should implement the business logic for every endpoint for the AccessRulesApi API.
// Include any external packages or services that will be required by this service.
type AccessRulesApiService struct {
	OrgHierarchyManager *cache.OrgHierarchyManager
	TenantSettings      TenantSettingsService
}

type AccessRuleFailure struct {
	accessRule models.AccessRule
	failure    *api.Error
}

type AccessRulesApiServicerInner interface {
	api.AccessRulesAPIServicer
	WithTenantSettings(tenantSettings TenantSettingsService) api.AccessRulesAPIServicer
}

var NewToLegacyRolesMap = map[string]string{
	static_roles.AiPractitioner.Name:      static_legacy_roles.ResearcherL1.Name,
	static_roles.DataAndStorageAdmin.Name: static_legacy_roles.DataSourcesAdmin.Name,
	static_roles.ProjectAdmin.Name:        static_legacy_roles.ResearcherL1.Name,
}

// NewAccessRulesApiService creates a default api service
func NewAccessRulesApiService(tenantSettings TenantSettingsService) api.AccessRulesAPIServicer {
	return &AccessRulesApiService{
		OrgHierarchyManager: cache.GetOrgHierarchyManager(),
		TenantSettings:      tenantSettings,
	}
}

func (s *AccessRulesApiService) WithTenantSettings(tenantSettings TenantSettingsService) api.AccessRulesAPIServicer {
	s.TenantSettings = tenantSettings
	return s
}

func (s *AccessRulesApiService) AccessRulesMigrateDeprecated(ctx context.Context, request api.AccessRulesMigrateDeprecatedFields) (api.ImplResponse, error) {
	if request.OldRoleId == request.NewRoleId {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "oldRoleId and newRoleId must differ", nil)), nil
	}

	tenantId, subjectId, groups, apiErr := parseAuthData(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	tenantIdStr := fmt.Sprintf("%d", tenantId)

	// Old role: must exist and be deprecated. Enabled check is intentionally skipped — a
	// deprecated role may also be disabled, and migrating away from it is still valid.
	oldRole, apiErr := dao.DbGetRoleByIdWithInternal(ctx, tenantIdStr, uint(request.OldRoleId), nil)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	if !oldRole.Deprecated {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Old role %d is not deprecated", request.OldRoleId), nil)), nil
	}

	// New role: must exist, be enabled, and NOT be deprecated.
	newRole, apiErr := validateCreationRequestRole(ctx, request.NewRoleId, tenantId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	if newRole.Deprecated {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("New role %d is deprecated", request.NewRoleId), nil)), nil
	}

	// Bulk privilege "discount": require caller to be able to create and delete access rules at the
	// tenant scope. This avoids per-AccessRule privilege validation, relying on the caller's tenant-wide
	// authority to act on any access rule in the tenant.
	if apiErr = validateCallerCanMigrateAccessRules(ctx, tenantId, subjectId, groups); apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	oldRules, apiErr := dao.DbGetDeprecatedAccessRulesForMigration(ctx, tenantId, uint(request.OldRoleId), MaxDeprecatedMigrationSize)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	created := []api.AccessRule{}
	preExisting := []api.AccessRule{}
	deleted := []string{}
	failed := []api.AccessRulesMigrateDeprecatedFailureItem{}

	for i := range oldRules {
		oldRule := &oldRules[i]
		migratedAccessRule, preExisted, step, migrateErr := dao.DbMigrateSingleAccessRule(ctx, tenantId, oldRule, uint(request.NewRoleId), subjectId)
		if migrateErr != nil {
			failed = append(failed, api.AccessRulesMigrateDeprecatedFailureItem{
				OldAccessRuleId: int32(oldRule.ID),
				Step:            step,
				Code:            migrateErr.Code,
				Message:         migrateErr.Message,
			})
			continue
		}

		migratedAccessRule.Role = *newRole

		if !preExisted {
			if auditErr := audit.AuditFromAudit(audit.AuditFromCtx(ctx)).
				WithEntityType(AuditEntityType).
				WithEntityId(strconv.FormatUint(uint64(migratedAccessRule.ID), 10)).
				WithStatusCode(http.StatusCreated).
				Write(); auditErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to write create audit for migrated access rule %d: %v", migratedAccessRule.ID, auditErr)
			}
		}
		if auditErr := audit.AuditFromAudit(audit.AuditFromCtx(ctx)).
			WithAction("Delete").
			WithHttpMethod(http.MethodDelete).
			WithEntityType(AuditEntityType).
			WithEntityId(strconv.FormatUint(uint64(oldRule.ID), 10)).
			WithStatusCode(http.StatusNoContent).
			Write(); auditErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to write delete audit for migrated access rule %d: %v", oldRule.ID, auditErr)
		}

		apiRule := AccessRuleDaoToApi(migratedAccessRule)
		if preExisted {
			preExisting = append(preExisting, apiRule)
		} else {
			created = append(created, apiRule)
		}
		deleted = append(deleted, strconv.FormatUint(uint64(oldRule.ID), 10))
	}
	audit.AuditFromCtx(ctx).Drop()

	return api.Response(http.StatusOK, api.AccessRulesMigrateDeprecatedResponse{
		Succeeded: api.AccessRulesMigrateDeprecatedResponseSucceeded{
			Created:     created,
			PreExisting: preExisting,
			Deleted:     deleted,
		},
		Failed: failed,
	}), nil
}

func (s *AccessRulesApiService) CreateAccessRule(ctx context.Context, accessRuleCreationRequest api.AccessRuleCreationFields) (api.ImplResponse, error) {
	tenantId, subjectId, groups, err := parseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	role, err := validateCreationRequestRole(ctx, accessRuleCreationRequest.RoleId, tenantId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	err = validateCurrentUserHasPrivilege(ctx, tenantId, subjectId, groups, models.AccessRule{
		Role:      *role,
		ScopeId:   accessRuleCreationRequest.ScopeId,
		ScopeType: accessRuleCreationRequest.ScopeType,
	})
	if err != nil {
		return NewErrorResponse(err), nil
	}

	accessRule, err := s.createSingleAccessRule(ctx, accessRuleCreationRequest, tenantId, subjectId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusCreated, AccessRuleDaoToApi(accessRule)), nil
}

func (s *AccessRulesApiService) createSingleAccessRule(ctx context.Context, accessRuleCreationRequest api.AccessRuleCreationFields, tenantId uint, subjectId string) (*models.AccessRule, *api.Error) {
	isAllAuthenticatedUsers := accessRuleCreationRequest.SubjectType == models.AllAuthenticatedUsers
	if !isAllAuthenticatedUsers && !accessRuleCreationRequest.SubjectType.IsValid() {
		return nil, api.NewError(http.StatusBadRequest,
			fmt.Sprintf("invalid subjectType %q: valid values are %v", accessRuleCreationRequest.SubjectType, api.AllowedSubjectTypeEnumValues), nil)
	}
	if apiErr := validateSubjectId(accessRuleCreationRequest); apiErr != nil {
		return nil, apiErr
	}
	if isAllAuthenticatedUsers {
		accessRuleCreationRequest.SubjectId = models.AllAuthenticatedUsers
	}
	if !accessRuleCreationRequest.ScopeType.IsValid() {
		return nil, api.NewError(http.StatusBadRequest,
			fmt.Sprintf("invalid scopeType %q: valid values are %v", accessRuleCreationRequest.ScopeType, api.AllowedScopeTypeEnumValues), nil)
	}

	// A system-scoped all-authenticated-users rule would bind system:authenticated cluster-wide across
	// every cluster, granting every in-cluster ServiceAccount — too broad to allow.
	if isAllAuthenticatedUsers && accessRuleCreationRequest.ScopeType == api.SCOPETYPE_SYSTEM {
		return nil, api.NewError(http.StatusBadRequest,
			"all-authenticated-users access rules are not allowed at system scope", nil)
	}

	// All system scoped access rules should be created under the management tenant and not under some costumer tenant
	if accessRuleCreationRequest.ScopeType == api.SCOPETYPE_SYSTEM && !runaictx.IsManagementTenant(ctx) {
		return nil, api.NewError(http.StatusForbidden, "Subject is not authorized to create system scoped access rules", nil)
	}

	scopeName, err := s.OrgHierarchyManager.GetScopeName(ctx, accessRuleCreationRequest.ScopeType, accessRuleCreationRequest.ScopeId)
	if err != nil {
		return nil, err
	}

	clusterId := getClusterUuid(ctx, accessRuleCreationRequest.ScopeType, accessRuleCreationRequest.ScopeId)
	usePhase := shouldUsePhaseWhenCreatingAccessRule(ctx, clusterId, accessRuleCreationRequest.ScopeType)

	accessRule, err := dao.DbCreateAccessRule(ctx, tenantId, subjectId, scopeName, clusterId, usePhase, accessRuleCreationRequest)
	if err != nil {
		return nil, err
	}

	audit.AuditFromCtx(ctx).WithEntityType(AuditEntityType).WithEntityId(strconv.FormatUint(uint64(accessRule.ID), 10))
	return accessRule, nil
}

func validateSubjectId(req api.AccessRuleCreationFields) *api.Error {
	if req.SubjectType == models.AllAuthenticatedUsers {
		if req.SubjectId != "" {
			return api.NewError(http.StatusBadRequest, "subjectId must not be set for the all-authenticated-users subject type", nil)
		}
		return nil
	}
	if req.SubjectId == "" {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("subjectId is required for subjectType %q", req.SubjectType), nil)
	}
	if req.SubjectId == models.AllAuthenticatedUsers {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("subjectId %q is reserved for the all-authenticated-users subject type", models.AllAuthenticatedUsers), nil)
	}
	return nil
}

func validateCreationRequestRole(ctx context.Context, roleId int32, tenantId uint) (*models.Role, *api.Error) {
	role, err := dao.DbGetRoleByIdWithInternal(ctx, fmt.Sprintf("%d", tenantId), uint(roleId), nil)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get the role %d data. err: %v", roleId, err)
		return nil, err
	}

	err = validateRoleEnabled(role)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func validateRoleEnabled(role *models.Role) *api.Error {
	roleEnabled := role.Enabled
	if role.EffectiveEnabled != nil {
		roleEnabled = *role.EffectiveEnabled
	}
	if !roleEnabled {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("Cannot create access rule, role %s is disabled", role.Name), nil)
	}
	return nil
}

func (s *AccessRulesApiService) DeleteAccessRule(ctx context.Context, accessRuleId int32) (api.ImplResponse, error) {
	tenantId, subjectId, groups, err := parseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	accessRule, err := dao.DbGetAccessRuleById(ctx, uint(accessRuleId), tenantId, nil, true, false)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	lastSystemAdminDeleted, err := deletionOfLastSystemAdmin(ctx, tenantId, []string{fmt.Sprint(accessRule.ID)})
	if err != nil {
		return NewErrorResponse(err), nil
	}
	if lastSystemAdminDeleted {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "Cannot delete last system admin access rule", nil)), nil
	}

	err = validateCurrentUserHasPrivilege(ctx, tenantId, subjectId, groups, *accessRule)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	err = dao.DbDeleteAccessRule(ctx, tenantId, accessRule.ID)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	audit.AuditFromCtx(ctx).WithEntityType(AuditEntityType).WithEntityId(strconv.FormatUint(uint64(accessRule.ID), 10))

	return NoContentResponse(), nil
}

func (s *AccessRulesApiService) GetAccessRule(ctx context.Context, accessRuleId int32) (api.ImplResponse, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	permittedScopes, scopesErr := s.calculatePermittedScopes(ctx)
	if scopesErr != nil {
		return ErrorToApiError(scopesErr, http.StatusBadRequest), nil
	}

	accessRule, err := dao.DbGetAccessRuleById(ctx, uint(accessRuleId), tenantId, permittedScopes, false, false)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	path, getError := s.getAccessRulePath(ctx, accessRule.ScopeName, accessRule.ScopeType, accessRule.ScopeId)
	if getError != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get access rule path of %s/%s. Err: %v", accessRule.ScopeType, accessRule.ScopeId, getError)
		path = nil
	}

	return api.Response(http.StatusOK, ConvertToAccessRuleById(AccessRuleDaoToApi(accessRule), path)), nil
}

func filterByBackwardsCompatibility(filterBy []string, subjectType string) []string {
	// Convert legacy subjectType parameter to filterBy for unified handling
	if subjectType != "" {
		filterBy = append(filterBy, "subjectType=="+subjectType)
	}

	// Apply subject type backward compatibility for app/service-account
	// Transform filters to treat 'app' and 'service-account' as equivalent
	filterBy = applySubjectTypeFilterBackwardCompatibility(filterBy)

	return filterBy
}

// applySubjectTypeFilterBackwardCompatibility transforms subject type filters for backward compatibility.
// For == operator: Uses OR syntax (field==value1||value2) to include both app and service-account
// For != operator: Adds the complementary != filter (e.g., !=app also adds !=service-account)
func applySubjectTypeFilterBackwardCompatibility(filterBy []string) []string {
	result := []string{}
	hasAppOrServiceAccountEqualFilter := false
	hasAppOrServiceAccountNonEqualFilter := false

	for _, filter := range filterBy {
		// Handle subjectType==app or subjectType==service-account
		switch filter {
		case "subjectType==app", "subjectType==service-account":
			if !hasAppOrServiceAccountEqualFilter {
				// Use OR syntax like type==Training||Distributed
				result = append(result, "subjectType==app||service-account")
				hasAppOrServiceAccountEqualFilter = true
			}
			// Skip duplicate if already processed
		case "subjectType!=app", "subjectType!=service-account":
			if !hasAppOrServiceAccountNonEqualFilter {
				// Add !=app and also !=service-account
				result = append(result, "subjectType!=app", "subjectType!=service-account")
				hasAppOrServiceAccountNonEqualFilter = true
			}
			// Skip duplicate if already processed
		default:
			// Keep all other filters as-is
			result = append(result, filter)
		}
	}

	return result
}

func (s *AccessRulesApiService) GetAccessRules(
	ctx context.Context,
	subjectType string,
	subjectIds []string,
	limit int32,
	offset int32,
	lastUpdated string,
	includeDeleted bool,
	clusterId string,
	scopeId string,
	sortOrder string,
	sortBy api.AccessRulesSortFilterFields,
	filterBy []string,
	search string,
) (api.ImplResponse, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	filterBy = filterByBackwardsCompatibility(filterBy, subjectType)

	parser := queryapi.NewParser(models.AccessRule{}.SortAndFilterFieldsToColumn())
	filters, filterErr := parser.Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	if len(subjectIds) == 0 {
		subjectIds = nil
	}

	permittedScopes, scopesErr := s.calculatePermittedScopes(ctx)
	if scopesErr != nil {
		return ErrorToApiError(scopesErr, http.StatusBadRequest), nil
	}
	sortColumn, sortError := parser.GetMapping(string(sortBy))
	if sortError != nil {
		return ErrorToApiError(sortError, http.StatusBadRequest), nil
	}

	accessRules, err := dao.DbGetAccessRules(ctx, tenantId, permittedScopes, "", subjectIds, limit,
		offset, lastUpdated, includeDeleted, clusterId, scopeId, sortOrder, sortColumn, filters, search)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	countAccessRules := int32(len(*accessRules))
	accessRulesResponse := api.GetAccessRules200Response{
		TotalRecords:   countAccessRules,
		DisplayRecords: countAccessRules,
		AccessRules:    []api.AccessRule{},
	}

	shouldMapServiceAccount := shouldMapServiceAccountToApp(ctx, clusterId)

	for _, accessRule := range *accessRules {
		rule := AccessRuleDaoToApi(&accessRule)
		if shouldMapServiceAccount && rule.SubjectType == api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT {
			rule.SubjectType = api.SUBJECTTYPEEXTENDED_APP
		}
		accessRulesResponse.AccessRules = append(accessRulesResponse.AccessRules, rule)
	}
	return api.Response(http.StatusOK, accessRulesResponse), nil
}

func (s *AccessRulesApiService) calculatePermittedScopes(ctx context.Context) (*authorization_api.PermittedScopes, error) {
	enabled := s.TenantSettings.IsScopedAuthorizationEnabled(ctx)
	if !enabled {
		return nil, nil
	}
	permittedScopes, ok := helpers.GetPermittedScopesFromCtx(ctx)
	if !ok {
		permittedScopes = nil
	}

	permittedScopes, scopesErr := s.expendPermittedScopesToDescendants(ctx, permittedScopes)
	if scopesErr != nil {
		return nil, scopesErr
	}
	return permittedScopes, nil
}

func (s *AccessRulesApiService) expendPermittedScopesToDescendants(ctx context.Context, permittedScopes *authorization_api.PermittedScopes) (*authorization_api.PermittedScopes, error) {
	if permittedScopes == nil {
		return nil, nil
	}
	if len(permittedScopes.Departments) == 0 {
		return permittedScopes, nil
	}

	projectIds, err := s.OrgHierarchyManager.GetDepartmentsDescendants(ctx, permittedScopes.Departments)
	if err != nil {
		return nil, err
	}
	permittedScopes.Projects = append(permittedScopes.Projects, projectIds...)
	return permittedScopes, nil
}

func (s *AccessRulesApiService) GetAccessRulesSync(ctx context.Context, limit int32, offset int32, lastUpdated string, clusterId string) (api.ImplResponse, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	accessRules, err := dao.DbGetAccessRules(ctx, tenantId, nil, "", nil, limit, offset, lastUpdated, true, clusterId, "", "", "", nil, "")
	if err != nil {
		return NewErrorResponse(err), nil
	}

	// Check if cluster supports service-account subject type (v2.24+)
	supportsServiceAccount := true
	if clusterId != "" {
		supported, versionErr := isServiceAccountSupportedInCluster(ctx, clusterId)
		if versionErr != nil {
			// If version check fails, default to old behavior (app)
			if errors.Is(versionErr, checker.ErrClusterVersionEmpty) {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to check cluster version for service-account support, defaulting to 'app'. err: %v", versionErr)
			} else {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to check cluster version for service-account support, defaulting to 'app'. err: %v", versionErr)
			}
			supportsServiceAccount = false
		} else {
			supportsServiceAccount = supported
		}
	}

	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)

	newRolesSupported, capabilitiesErr := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionPermissionSetRoles)
	if capabilitiesErr != nil {
		return ErrorToApiError(capabilitiesErr, http.StatusServiceUnavailable), nil
	}

	// Below v3.0 the cluster can't parse all-authenticated-users; the loop remaps it to a
	// system:authenticated group. The cluster version is already validated above (that check 503s
	// on the same unfetchable-version condition), so a residual error here is treated as
	// "unsupported" — the safe group remap.
	supportsAllAuthenticatedUsers, versionErr := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfAllAuthenticatedUsersSubjectType)
	if versionErr != nil {
		supportsAllAuthenticatedUsers = false
	}

	accessRulesResponse := api.GetAccessRulesSync200Response{
		AccessRules: []api.AccessRuleSync{},
	}
	for _, accessRule := range *accessRules {
		predefinedRole, err := getPredefinedRoleForSync(ctx, accessRule, tenantId, newRolesSupported)
		if err != nil {
			return NewErrorResponse(err), nil
		}
		syncRule := AccessRuleDaoToSyncApi(&accessRule, predefinedRole)
		syncRule = remapSyncSubjectForClusterSupport(syncRule, supportsServiceAccount, supportsAllAuthenticatedUsers)
		accessRulesResponse.AccessRules = append(accessRulesResponse.AccessRules, syncRule)
	}
	countAccessRules := int32(len(*accessRules))
	if countAccessRules == limit {
		accessRulesResponse.Next = countAccessRules + offset
	}
	return api.Response(http.StatusOK, accessRulesResponse), nil
}

// remapSyncSubjectForClusterSupport rewrites subject types a below-minimum-version cluster can't
// parse into equivalents it understands: service-account → app, and all-authenticated-users → a
// group bound to system:authenticated. Newer clusters receive the native types unchanged.
func remapSyncSubjectForClusterSupport(syncRule api.AccessRuleSync, supportsServiceAccount, supportsAllAuthenticatedUsers bool) api.AccessRuleSync {
	if !supportsServiceAccount && syncRule.SubjectType == api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT {
		syncRule.SubjectType = api.SUBJECTTYPEEXTENDED_APP
	}
	if !supportsAllAuthenticatedUsers && syncRule.SubjectType == api.SUBJECTTYPEEXTENDED_ALL_AUTHENTICATED_USERS {
		syncRule.SubjectType = api.SUBJECTTYPEEXTENDED_GROUP
		syncRule.SubjectId = authenticatedUsersK8sGroup
	}
	return syncRule
}

func getPredefinedRoleForSync(ctx context.Context, accessRule models.AccessRule, tenantId uint, newRolesSupported bool) (*models.Role, *api.Error) {
	if newRolesSupported {
		// all roles supported - return the original predefined role
		if accessRule.Role.ClusterRolePredefinedId != nil {
			return dao.DbGetRoleById(ctx, fmt.Sprint(tenantId), uint(*accessRule.Role.ClusterRolePredefinedId), nil)
		}
		return nil, nil
	}
	var predefinedRole *models.Role = nil
	var err *api.Error
	if accessRule.Role.ClusterRolePredefinedId != nil {
		// not all roles are supported and there is predefined role - return it or the supported one from NewToLegacyRolesMap
		predefinedRole, err = dao.DbGetRoleById(ctx, fmt.Sprint(tenantId), uint(*accessRule.Role.ClusterRolePredefinedId), nil)
		if err != nil {
			return nil, err
		}
		mappedRoleName, found := NewToLegacyRolesMap[predefinedRole.Name]
		if !found {
			return predefinedRole, nil
		}
		return dao.DbGetRoleById(ctx, fmt.Sprint(tenantId), dao.MapStaticRoleNameToRoleId[mappedRoleName], nil)
	}
	// not all roles are supported and there is no predefined role - return the role itself or the supported one from NewToLegacyRolesMap
	mappedRoleName, found := NewToLegacyRolesMap[accessRule.Role.Name]
	if found {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("role %s is not supported in cluster, mapping to role %s", accessRule.Role.Name, mappedRoleName)
		return dao.DbGetRoleById(ctx, fmt.Sprint(tenantId), dao.MapStaticRoleNameToRoleId[mappedRoleName], nil)
	}
	return nil, nil
}

func (s *AccessRulesApiService) SyncAccessRules(ctx context.Context, clusterId string, accessRuleForSync []api.AccessRuleForSync) (api.ImplResponse, error) {
	return NoContentResponse(), nil
}

func (s *AccessRulesApiService) UpdateAccessRuleStatus(ctx context.Context, accessRuleId int32, clusterId string, accessRulesStatusRequest api.AccessRuleStatusRequest) (api.ImplResponse, error) {
	if accessRulesStatusRequest.Phase == api.ACCESSRULECLUSTERPHASE_DELETED {
		// at this phase, we are not reporting deletion
		return NoContentResponse(), nil
	}

	tenantId, tenantErr := parseTenantId(ctx)
	if tenantErr != nil {
		return NewErrorResponse(tenantErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting access rule %d update status request", accessRuleId)

	accessRule, err := dao.DbGetAccessRuleById(ctx, uint(accessRuleId), tenantId, nil, true, false)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	err = dao.DbUpdateAccessRulePhase(ctx, uint(accessRuleId), tenantId, clusterId, accessRulesStatusRequest, accessRule)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully updated status of access rule %d to: %s/%s",
		accessRuleId, accessRulesStatusRequest.Phase, accessRulesStatusRequest.Message)
	return NoContentResponse(), nil
}

func (s *AccessRulesApiService) CountAccessRules(ctx context.Context, deleted bool, filterBy []string, search string) (api.ImplResponse, error) {
	tenantId, tenantErr := parseTenantId(ctx)
	if tenantErr != nil {
		return NewErrorResponse(tenantErr), nil
	}

	// Apply backward compatibility for subject type filters
	filterBy = applySubjectTypeFilterBackwardCompatibility(filterBy)

	parser := queryapi.NewParser(models.AccessRule{}.SortAndFilterFieldsToColumn())
	filters, filterErr := parser.Filters(filterBy)
	if filterErr != nil {
		return ErrorToApiError(filterErr, http.StatusBadRequest), nil
	}

	count, dbErr := dao.DbCountAccessRules(ctx, tenantId, deleted, filters, search)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	response := api.CountAccessRules200Response{
		Count: count,
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *AccessRulesApiService) getAccessRulePath(ctx context.Context, scopeName string, scopeType api.ScopeType, scopeId string) (*string, error) {
	orgTree, mapScopeToName, err := s.OrgHierarchyManager.GetOrgTree(ctx)
	if err != nil {
		return nil, err
	}

	path := constructPath(scopeName, scopeType, scopeId, mapScopeToName, orgTree)

	return &path, nil
}

func constructPath(scopeName string, scopeType api.ScopeType, scopeId string, mapScopeToName map[structs.OrgUnitKey]string, mapChildToParent structs.OrgUnitMap) string {
	current := structs.OrgUnitKey{
		Scope: string(scopeType),
		Id:    scopeId,
	}
	parent, found := mapChildToParent[current]
	path := scopeName
	for found {
		current = parent
		path = fmt.Sprintf("%s/%s", mapScopeToName[parent], path)
		parent, found = mapChildToParent[current]
	}

	return path
}

// AccessRulesBatch - AccessRules batch operations.
func (s *AccessRulesApiService) AccessRulesBatch(ctx context.Context, request api.AccessRulesBatchFields) (api.ImplResponse, error) {
	ids := removeDuplicates(request.Ids)
	err := validateBatchDeleteRequest(ids, request.Action)
	if err != nil {
		return api.Response(http.StatusBadRequest, nil), err
	}

	switch request.Action {
	case api.ACCESSRULESBATCHACTION_DELETE:
		return s.deleteAccessRulesBatch(ctx, ids)
	case api.ACCESSRULESBATCHACTION_VALIDATE_DELETE:
		return s.validateDeleteAccessRulesBatch(ctx, ids)
	default:
		return api.Response(http.StatusNotImplemented, nil), fmt.Errorf("AccessRules Batch method <%s> not implemented", request.Action)
	}
}

// CreateAuthenticatedUsersAccessRule sends no subjectId; createSingleAccessRule stores the
// reserved sentinel for the all-authenticated-users type.
func (s *AccessRulesApiService) CreateAuthenticatedUsersAccessRule(ctx context.Context, req api.AuthenticatedUsersAccessRuleCreationFields) (api.ImplResponse, error) {
	return s.CreateAccessRule(ctx, api.AccessRuleCreationFields{
		SubjectType: models.AllAuthenticatedUsers,
		RoleId:      req.RoleId,
		ScopeId:     req.ScopeId,
		ScopeType:   api.ScopeType(req.ScopeType),
		ClusterId:   req.ClusterId,
	})
}

func (s *AccessRulesApiService) AccessRulesBatchCreate(ctx context.Context, accessRulesBatchCreateFields api.AccessRulesBatchCreateFields) (api.ImplResponse, error) {
	err := validateBatchCreateRequest(accessRulesBatchCreateFields.Payload)
	if err != nil {
		return api.Response(http.StatusBadRequest, nil), err
	}
	tenantId, subjectId, groups, apiErr := parseAuthData(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	role, apiErr := validateCreationRequestRole(ctx, accessRulesBatchCreateFields.Payload.RoleId, tenantId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	accessRulesToCreate := []api.AccessRuleCreationFields{}

	for _, subject := range accessRulesBatchCreateFields.Payload.Subjects {
		for _, scope := range accessRulesBatchCreateFields.Payload.Scopes {
			accessRulesToCreate = append(accessRulesToCreate, api.AccessRuleCreationFields{
				SubjectId:   subject.SubjectId,
				SubjectType: subject.SubjectType,
				RoleId:      accessRulesBatchCreateFields.Payload.RoleId,
				ScopeId:     scope.ScopeId,
				ScopeType:   scope.ScopeType,
				ClusterId:   accessRulesBatchCreateFields.Payload.ClusterId,
			})
		}
	}

	orgTree, _, orgErr := cache.GetOrgHierarchyManager().GetOrgTree(ctx)
	if orgErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant org tree. err: %v", orgErr.Error())
		return NewErrorResponse(api.NewError(http.StatusBadRequest, orgErr.Error(), orgErr)), nil
	}

	succeededAccessRules, failedAccessRules, apiErr := validateAccessRulesBatchPrivilege(ctx, accessRulesToCreate, *role, tenantId, subjectId, groups, &orgTree)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	createdAccessRules := []api.AccessRule{}

	for _, creationFields := range succeededAccessRules {
		accessRule, apiErr := s.createSingleAccessRule(ctx, creationFields, tenantId, subjectId)
		if apiErr != nil {
			failedAccessRules = append(failedAccessRules, api.AccessRulesBatchCreateResponseFailedInner{
				Payload: creationFields,
				Code:    apiErr.Code,
				Message: apiErr.Message,
			})
		} else {
			createdAccessRules = append(createdAccessRules, AccessRuleDaoToApi(accessRule))
			auditLogError := audit.AuditFromAudit(audit.AuditFromCtx(ctx)).
				WithEntityType(AuditEntityType).
				WithEntityId(strconv.FormatUint(uint64(accessRule.ID), 10)).
				WithStatusCode(http.StatusCreated).
				Write()
			if auditLogError != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to write audit log for access rule %d creation: %v", accessRule.ID, auditLogError)
			}
		}

	}
	audit.AuditFromCtx(ctx).Drop()

	return api.Response(http.StatusOK, api.AccessRulesBatchCreateResponse{
		Succeeded: createdAccessRules,
		Failed:    failedAccessRules,
	}), nil
}

func buildBatchResponse(valid []uint, invalid []api.BatchResponseFailureItem) api.BatchResponse {
	var validIdsStr []string
	for _, validId := range valid {
		validIdsStr = append(validIdsStr, strconv.Itoa(int(validId)))
	}
	return api.BatchResponse{
		Succeeded: validIdsStr,
		Failed:    invalid,
	}
}

func (s *AccessRulesApiService) deleteAccessRulesBatch(ctx context.Context, accessRulesIds []string) (api.ImplResponse, error) {
	tenantId, _, _, err := parseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	validIds, invalidItems, err := s.getValidatedDeleteAccessRulesBatch(ctx, accessRulesIds)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	err = dao.DbBatchDeleteAccessRules(ctx, tenantId, validIds)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	for _, accessRuleId := range validIds {
		auditLogError := audit.AuditFromAudit(audit.AuditFromCtx(ctx)).
			WithAction("Delete").
			WithHttpMethod(http.MethodDelete).
			WithEntityType(AuditEntityType).
			WithStatusCode(http.StatusNoContent).
			WithEntityId(strconv.FormatUint(uint64(accessRuleId), 10)).
			Write()
		if auditLogError != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to write audit log for access rule %d deletion: %v", accessRuleId, auditLogError)
		}
	}

	audit.AuditFromCtx(ctx).Drop()
	result := buildBatchResponse(validIds, invalidItems)

	return api.Response(http.StatusOK, result), nil
}

func validateDeletionOfLastSystemAdminBatch(ctx context.Context, tenantId uint, accessRulesIds []string) (validIds []string, invalid []api.BatchResponseFailureItem, err *api.Error) {
	lastSystemAdminDeleted, err := validateDeletionOfLastSystemAdmin(ctx, tenantId, accessRulesIds)
	if err != nil {
		return nil, nil, err
	}
	nonSysAdminIds := []string{}
	invalidSysAdminItems := []api.BatchResponseFailureItem{}
	invalidMap := map[string]bool{}
	for _, id := range lastSystemAdminDeleted {
		invalidMap[id] = true
	}
	for _, id := range accessRulesIds {
		found := invalidMap[id]
		if found {
			invalidSysAdminItems = append(invalidSysAdminItems, api.BatchResponseFailureItem{
				Id:      id,
				Code:    http.StatusBadRequest,
				Message: "Cannot delete last system admin access rule",
			})
		} else {
			nonSysAdminIds = append(nonSysAdminIds, id)
		}
	}
	return nonSysAdminIds, invalidSysAdminItems, nil
}

func (s *AccessRulesApiService) getValidatedDeleteAccessRulesBatch(ctx context.Context, accessRulesIds []string) (validIds []uint, invalid []api.BatchResponseFailureItem, err *api.Error) {
	tenantId, subjectId, groups, err := parseAuthData(ctx)
	if err != nil {
		return nil, nil, err
	}

	nonSysAdminIds, invalidSysAdminItems, err := validateDeletionOfLastSystemAdminBatch(ctx, tenantId, accessRulesIds)
	if err != nil {
		return nil, nil, err
	}

	orgTree, _, orgErr := cache.GetOrgHierarchyManager().GetOrgTree(ctx)
	if orgErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant org tree. err: %v", orgErr.Error())
		return nil, nil, api.NewError(http.StatusBadRequest, orgErr.Error(), orgErr)
	}

	validIds, invalidItems, err := validateAccessRulesIds(ctx, nonSysAdminIds, tenantId, subjectId, groups, &orgTree)
	if err != nil {
		return nil, nil, err
	}

	audit.AuditFromCtx(ctx).Drop()
	return validIds, append(invalidItems, invalidSysAdminItems...), err
}

func (s *AccessRulesApiService) validateDeleteAccessRulesBatch(ctx context.Context, accessRulesIds []string) (api.ImplResponse, error) {
	validIds, invalidItems, err := s.getValidatedDeleteAccessRulesBatch(ctx, accessRulesIds)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result := buildBatchResponse(validIds, invalidItems)

	return api.Response(http.StatusOK, result), nil
}

func validateAccessRulesIds(ctx context.Context,
	accessRulesIds []string,
	tenantId uint,
	subjectId string,
	groups []string,
	orgTree *structs.OrgUnitMap,
) (validIds []uint, invalid []api.BatchResponseFailureItem, err *api.Error) {
	accessRules, getErr := dao.DbGetAccessRulesByIds(ctx, tenantId, accessRulesIds)
	if getErr != nil {
		return nil, nil, getErr
	}

	succeeded, failed, err := validateAccessRulesPrivilege(ctx, *accessRules, tenantId, subjectId, groups, orgTree)
	if err != nil {
		return nil, nil, err
	}

	succeededIds := []uint{}
	failedIds := []api.BatchResponseFailureItem{}
	for _, accessRuleSucceeded := range succeeded {
		succeededIds = append(succeededIds, accessRuleSucceeded.ID)
	}
	for _, accessRuleFailed := range failed {
		failedIds = append(failedIds, api.BatchResponseFailureItem{
			Id:      fmt.Sprint(accessRuleFailed.accessRule.ID),
			Code:    accessRuleFailed.failure.Code,
			Message: accessRuleFailed.failure.Message,
		})
	}
	return succeededIds, failedIds, nil
}

func validateAccessRulesBatchPrivilege(ctx context.Context,
	accessRules []api.AccessRuleCreationFields,
	role models.Role,
	tenantId uint,
	subjectId string,
	groups []string,
	orgTree *structs.OrgUnitMap,
) (valid []api.AccessRuleCreationFields, invalid []api.AccessRulesBatchCreateResponseFailedInner, err *api.Error) {
	modelAccessRules := []models.AccessRule{}
	for _, ar := range accessRules {
		modelAccessRules = append(modelAccessRules, models.AccessRule{
			SubjectId:   ar.SubjectId,
			SubjectType: api.SubjectTypeExtended(ar.SubjectType),
			RoleID:      uint(ar.RoleId),
			Role:        role,
			ScopeId:     ar.ScopeId,
			ScopeType:   ar.ScopeType,
			TenantId:    tenantId,
			ClusterUuid: ar.ClusterId,
		})
	}
	succeeded, failed, err := validateAccessRulesPrivilege(ctx, modelAccessRules, tenantId, subjectId, groups, orgTree)
	if err != nil {
		return nil, nil, err
	}

	succeededAr := []api.AccessRuleCreationFields{}
	failedAr := []api.AccessRulesBatchCreateResponseFailedInner{}
	for _, accessRuleSucceeded := range succeeded {
		succeededAr = append(succeededAr, api.AccessRuleCreationFields{
			SubjectId:   accessRuleSucceeded.SubjectId,
			SubjectType: api.SubjectType(accessRuleSucceeded.SubjectType),
			RoleId:      int32(accessRuleSucceeded.RoleID),
			ScopeId:     accessRuleSucceeded.ScopeId,
			ScopeType:   accessRuleSucceeded.ScopeType,
			ClusterId:   accessRuleSucceeded.ClusterUuid,
		})
	}
	for _, accessRuleFailed := range failed {
		failedAr = append(failedAr, api.AccessRulesBatchCreateResponseFailedInner{
			Payload: api.AccessRuleCreationFields{
				SubjectId:   accessRuleFailed.accessRule.SubjectId,
				SubjectType: api.SubjectType(accessRuleFailed.accessRule.SubjectType),
				RoleId:      int32(accessRuleFailed.accessRule.RoleID),
				ScopeId:     accessRuleFailed.accessRule.ScopeId,
				ScopeType:   accessRuleFailed.accessRule.ScopeType,
				ClusterId:   accessRuleFailed.accessRule.ClusterUuid,
			},
			Code:    accessRuleFailed.failure.Code,
			Message: accessRuleFailed.failure.Message,
		})
	}
	return succeededAr, failedAr, nil
}

func validateAccessRulesPrivilege(ctx context.Context,
	accessRules []models.AccessRule,
	tenantId uint,
	subjectId string,
	groups []string,
	orgTree *structs.OrgUnitMap,
) (valid []models.AccessRule, invalid []AccessRuleFailure, err *api.Error) {
	requesterAccessRules, getSubjectErr := dao.DbGetSubjectPermissions(ctx, tenantId, subjectId, groups)
	if getSubjectErr != nil {
		return nil, nil, getSubjectErr
	}

	numberOfWorkers := min(runtime.NumCPU(), len(accessRules))
	workChan := make(chan models.AccessRule, len(accessRules))
	resultChan := make(chan struct{}, len(accessRules))

	succeeded := SyncSlice[models.AccessRule]{}
	failed := SyncSlice[AccessRuleFailure]{}

	// Start worker goroutines
	var wg sync.WaitGroup
	for range numberOfWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for accessRule := range workChan {
				permissionErr := validateSufficientPrivilegesForRequester(ctx, requesterAccessRules, ptr.To([]models.AccessRule{accessRule}), orgTree)
				if permissionErr != nil {
					failed.Append(AccessRuleFailure{
						accessRule: accessRule,
						failure:    permissionErr,
					})
				} else {
					succeeded.Append(accessRule)
				}
				resultChan <- struct{}{}
			}
		}()
	}

	// Send work to workers
	go func() {
		for _, accessRule := range accessRules {
			workChan <- accessRule
		}
		close(workChan)
	}()

	// Wait for all work to complete
	for range accessRules {
		<-resultChan
	}

	// Wait for all workers to finish
	wg.Wait()

	succeededArr := (&succeeded).GetAll()
	failedArr := (&failed).GetAll()
	return succeededArr, failedArr, nil
}

func validateBatchDeleteRequest(ids []string, action api.AccessRulesBatchAction) error {
	if len(ids) == 0 {
		return errors.New("amount of ids cannot be 0")
	}
	switch action {
	case api.ACCESSRULESBATCHACTION_DELETE:

		return validateDeletionBatchRequest(ids)
	case api.ACCESSRULESBATCHACTION_VALIDATE_DELETE:
		return validateDeletionBatchRequest(ids)
	}
	return nil
}

func validateBatchCreateRequest(payload api.AccessRulesBatchCreateFieldsPayload) error {
	subjectsLength := len(payload.Subjects)
	scopesLength := len(payload.Scopes)
	if subjectsLength == 0 || scopesLength == 0 {
		return errors.New("subjects or scopes length cannot be 0")
	}
	if subjectsLength > MaxBatchCreationSubjectsSize {
		return fmt.Errorf("access rules subjects length cannot be more than %d", MaxBatchCreationSubjectsSize)
	}
	if scopesLength > MaxBatchCreationScopesSize {
		return fmt.Errorf("access rules scopes length cannot be more than %d", MaxBatchCreationScopesSize)
	}
	return nil
}

func validateDeletionBatchRequest(ids []string) error {
	if len(ids) > MaxBatchDeletionSize {
		return fmt.Errorf("ids cannot be more than %d", MaxBatchDeletionSize)
	}
	return nil
}

func deletionOfLastSystemAdmin(ctx context.Context, tenantId uint, accessRuleIds []string) (bool, *api.Error) {
	accessRulesCount, err := dao.DbCountSystemAdminAccessRules(ctx, tenantId, accessRuleIds)
	if err != nil {
		return false, api.NewError(http.StatusInternalServerError, err.Error(), err)
	}
	return accessRulesCount == 0, nil
}

func validateDeletionOfLastSystemAdmin(ctx context.Context, tenantId uint, accessRuleIds []string) ([]string, *api.Error) {
	lastSystemAdmins, err := dao.DbGetIdsForLastSystemAdminDeletion(ctx, fmt.Sprint(tenantId), accessRuleIds)
	if err != nil {
		return []string{}, api.NewError(http.StatusInternalServerError, err.Error(), err)
	}
	return lastSystemAdmins, nil
}

func shouldUsePhaseWhenCreatingAccessRule(ctx context.Context, clusterId string, scopeType api.ScopeType) bool {
	if scopeType == api.SCOPETYPE_SYSTEM ||
		scopeType == api.SCOPETYPE_TENANT {
		return false
	}

	if clusterId == "" {
		return false
	}

	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	doesClusterSupportStatuses, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionAccessRulesStatuses)
	if err != nil {
		if errors.Is(err, checker.ErrClusterVersionEmpty) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to validate cluster %s capability. err: %v", clusterId, err)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to validate cluster %s capability. err: %v", clusterId, err)
		}
		return false
	}

	return doesClusterSupportStatuses
}

// isServiceAccountSupportedInCluster checks if the cluster version supports service-account subject type (v2.24+)
// Returns true if cluster >= v2.24, false otherwise
func isServiceAccountSupportedInCluster(ctx context.Context, clusterId string) (bool, error) {
	if clusterId == "" {
		return false, nil
	}

	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	supportsServiceAccount, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfServiceAccountSubjectType)
	if err != nil {
		return false, fmt.Errorf("failed to validate cluster %s capability for service-account support: %w", clusterId, err)
	}

	return supportsServiceAccount, nil
}

// validateCallerCanMigrateAccessRules checks that the calling subject can create AND delete
// access_rules at (at least) the given tenant's scope. Used as a bulk privilege check in place of
// per-access-rule privilege validation on the migrate-deprecated endpoint.
func validateCallerCanMigrateAccessRules(ctx context.Context, tenantId uint, subjectId string, groups []string) *api.Error {
	tenantIdStr := fmt.Sprintf("%d", tenantId)
	resourceType := api.RESOURCETYPE_ACCESS_RULES
	for _, action := range []api.Action{api.ACTION_CREATE, api.ACTION_DELETE} {
		scopes, apiErr := dao.EvaluatePermission(ctx, tenantId, subjectId, groups, api.BackendEvaluationRequest{
			ResourceType:  &resourceType,
			RequestAction: &action,
		}, false)
		if apiErr != nil {
			return apiErr
		}
		if scopes == nil || (!scopes.System && !slices.Contains(scopes.Tenants, tenantIdStr)) {
			return api.NewError(http.StatusForbidden, fmt.Sprintf("Insufficient permissions: caller must be able to %s access_rules at tenant scope", action), nil)
		}
	}
	return nil
}
