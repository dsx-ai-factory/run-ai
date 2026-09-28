/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	org_cache "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/cache"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var ResourceTypesDisplayName = map[api.ResourceType]string{
	api.RESOURCETYPE_TENANT:                       "Account",
	api.RESOURCETYPE_DEPARTMENT:                   "Departments",
	api.RESOURCETYPE_PROJECT:                      "Projects",
	api.RESOURCETYPE_SETTINGS:                     "Settings",
	api.RESOURCETYPE_SECURITY_SETTINGS:            "Security settings",
	api.RESOURCETYPE_NOTIFICATION_SETTINGS:        "Notification settings",
	api.RESOURCETYPE_BRANDING_SETTINGS:            "Branding settings",
	api.RESOURCETYPE_POLICIES:                     "Policies",
	api.RESOURCETYPE_EVENTS_HISTORY:               "Event history",
	api.RESOURCETYPE_CLUSTER:                      "Clusters",
	api.RESOURCETYPE_CLUSTER_CONFIG:               "Cluster configuration",
	api.RESOURCETYPE_NODEPOOLS:                    "Node pools",
	api.RESOURCETYPE_NODES:                        "Nodes",
	api.RESOURCETYPE_USERS:                        "Users",
	api.RESOURCETYPE_APPS:                         "Applications",
	api.RESOURCETYPE_SERVICE_ACCOUNT:              "Service accounts",
	api.RESOURCETYPE_ROLES:                        "Roles",
	api.RESOURCETYPE_ACCESS_RULES:                 "Access rules",
	api.RESOURCETYPE_WORKLOADS:                    "Workloads",
	api.RESOURCETYPE_APPLICATIONS:                 "AI Applications",
	api.RESOURCETYPE_INFERENCES:                   "Inferences",
	api.RESOURCETYPE_WORKSPACES:                   "Workspaces",
	api.RESOURCETYPE_TRAININGS:                    "Trainings",
	api.RESOURCETYPE_ENVIRONMENTS:                 "Environments",
	api.RESOURCETYPE_PVC_ASSETS:                   "Data sources",
	api.RESOURCETYPE_GIT_ASSETS:                   "Data sources",
	api.RESOURCETYPE_HOST_PATH_ASSETS:             "Data sources",
	api.RESOURCETYPE_NFS_ASSETS:                   "Data sources",
	api.RESOURCETYPE_S3_ASSETS:                    "Data sources",
	api.RESOURCETYPE_CM_VOLUME_ASSETS:             "Data sources",
	api.RESOURCETYPE_SECRET_VOLUME_ASSETS:         "Data sources",
	api.RESOURCETYPE_COMPUTE_RESOURCES:            "Compute resources",
	api.RESOURCETYPE_RESOURCE_CLAIM_TEMPLATES:     "Resource claim templates",
	api.RESOURCETYPE_TEMPLATES:                    "Templates",
	api.RESOURCETYPE_CREDENTIALS:                  "Credentials",
	api.RESOURCETYPE_DATAVOLUMES:                  "Data volumes",
	api.RESOURCETYPE_STORAGE_CLASS_CONFIGURATION:  "Storage class configuration",
	api.RESOURCETYPE_DEVICE_CLASSES:               "Device classes",
	api.RESOURCETYPE_ACCESS_KEYS:                  "Access Keys",
	api.RESOURCETYPE_WORKLOAD_PROPERTIES:          "Workload properties",
	api.RESOURCETYPE_NETWORK_TOPOLOGIES:           "Network topologies",
	api.RESOURCETYPE_REGISTRIES:                   "Registries",
	api.RESOURCETYPE_WORKLOAD_INTEGRATION_METRICS: "Workload integration metrics",
	api.RESOURCETYPE_CLUSTERS_MINIMAL:             "Clusters minimal",
	api.RESOURCETYPE_NODEPOOLS_MINIMAL:            "Nodepools minimal",
	api.RESOURCETYPE_USERS_MINIMAL:                "Users minimal",
}

var ResourceTypesGroupId = map[api.ResourceType]api.ResourceTypeGroupId{
	api.RESOURCETYPE_TENANT:                       api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_DEPARTMENT:                   api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_PROJECT:                      api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_SETTINGS:                     api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_SECURITY_SETTINGS:            api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_NOTIFICATION_SETTINGS:        api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_BRANDING_SETTINGS:            api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_POLICIES:                     api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_EVENTS_HISTORY:               api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_NETWORK_TOPOLOGIES:           api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_WORKLOAD_INTEGRATION_METRICS: api.RESOURCETYPEGROUPID_ORGANIZATION,
	api.RESOURCETYPE_CLUSTER:                      api.RESOURCETYPEGROUPID_PHYSICAL_RESOURCE,
	api.RESOURCETYPE_CLUSTER_CONFIG:               api.RESOURCETYPEGROUPID_PHYSICAL_RESOURCE,
	api.RESOURCETYPE_NODEPOOLS:                    api.RESOURCETYPEGROUPID_PHYSICAL_RESOURCE,
	api.RESOURCETYPE_NODES:                        api.RESOURCETYPEGROUPID_PHYSICAL_RESOURCE,
	api.RESOURCETYPE_CLUSTERS_MINIMAL:             api.RESOURCETYPEGROUPID_PHYSICAL_RESOURCE,
	api.RESOURCETYPE_NODEPOOLS_MINIMAL:            api.RESOURCETYPEGROUPID_PHYSICAL_RESOURCE,
	api.RESOURCETYPE_USERS:                        api.RESOURCETYPEGROUPID_IAM,
	api.RESOURCETYPE_USERS_MINIMAL:                api.RESOURCETYPEGROUPID_IAM,
	api.RESOURCETYPE_APPS:                         api.RESOURCETYPEGROUPID_IAM,
	api.RESOURCETYPE_SERVICE_ACCOUNT:              api.RESOURCETYPEGROUPID_IAM,
	api.RESOURCETYPE_ROLES:                        api.RESOURCETYPEGROUPID_IAM,
	api.RESOURCETYPE_ACCESS_RULES:                 api.RESOURCETYPEGROUPID_IAM,
	api.RESOURCETYPE_ACCESS_KEYS:                  api.RESOURCETYPEGROUPID_IAM,
	api.RESOURCETYPE_WORKLOADS:                    api.RESOURCETYPEGROUPID_WORKLOAD,
	api.RESOURCETYPE_INFERENCES:                   api.RESOURCETYPEGROUPID_WORKLOAD,
	api.RESOURCETYPE_WORKSPACES:                   api.RESOURCETYPEGROUPID_WORKLOAD,
	api.RESOURCETYPE_WORKLOAD_PROPERTIES:          api.RESOURCETYPEGROUPID_WORKLOAD,
	api.RESOURCETYPE_TRAININGS:                    api.RESOURCETYPEGROUPID_WORKLOAD,
	api.RESOURCETYPE_APPLICATIONS:                 api.RESOURCETYPEGROUPID_WORKLOAD,
	api.RESOURCETYPE_ENVIRONMENTS:                 api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_PVC_ASSETS:                   api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_GIT_ASSETS:                   api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_HOST_PATH_ASSETS:             api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_NFS_ASSETS:                   api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_S3_ASSETS:                    api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_CM_VOLUME_ASSETS:             api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_SECRET_VOLUME_ASSETS:         api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_COMPUTE_RESOURCES:            api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_RESOURCE_CLAIM_TEMPLATES:     api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_TEMPLATES:                    api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_CREDENTIALS:                  api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_DATAVOLUMES:                  api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_STORAGE_CLASS_CONFIGURATION:  api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_DEVICE_CLASSES:               api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
	api.RESOURCETYPE_REGISTRIES:                   api.RESOURCETYPEGROUPID_WORKLOAD_ASSET,
}

func RoleDaoToApi(role *models.Role) api.Role {
	k8sPermissions := api.KubernetesPermissions{}
	if role.ClusterRolePredefinedId != nil {
		k8sPermissions.PredefinedRole = ptr.To(fmt.Sprintf("%d", *role.ClusterRolePredefinedId))
	}
	effectiveEnable := role.Enabled
	if role.EffectiveEnabled != nil {
		effectiveEnable = *role.EffectiveEnabled
	}
	permissionSets, permissions := PermissionsSetDaoToRolePermApi(role.PermissionSets)
	return api.Role{
		Name:                  role.Name,
		Description:           role.Description,
		PermissionSets:        permissionSets,
		Permissions:           permissions,
		ScopeType:             string(role.ScopeType),
		ScopeId:               role.ScopeId,
		Enabled:               ptr.To(role.Enabled),
		KubernetesPermissions: k8sPermissions,
		Id:                    int32(role.ID),
		CreatedAt:             role.CreatedAt,
		UpdatedAt:             role.UpdatedAt,
		CreatedBy:             role.CreatedBy,
		Custom:                role.Custom,
		EffectiveEnabled:      effectiveEnable,
		Deprecated:            role.Deprecated,
	}
}

func RoleDaoToV1Api(role *models.Role) api.RoleV1 {
	k8sPermissions := api.KubernetesPermissions{}
	if role.ClusterRolePredefinedId != nil {
		k8sPermissions.PredefinedRole = ptr.To(fmt.Sprintf("%d", *role.ClusterRolePredefinedId))
	}
	effectiveEnable := role.Enabled
	if role.EffectiveEnabled != nil {
		effectiveEnable = *role.EffectiveEnabled
	}
	permissions := PermissionsSetDaoToPermissionApi(role.PermissionSets)
	return api.RoleV1{
		Name:                  role.Name,
		Description:           role.Description,
		Permissions:           permissions,
		ScopeType:             string(role.ScopeType),
		ScopeId:               role.ScopeId,
		Enabled:               ptr.To(role.Enabled),
		KubernetesPermissions: k8sPermissions,
		Id:                    int32(role.ID),
		CreatedAt:             role.CreatedAt,
		UpdatedAt:             role.UpdatedAt,
		CreatedBy:             role.CreatedBy,
		Custom:                role.Custom,
		EffectiveEnabled:      effectiveEnable,
	}
}

func AccessRuleDaoToApi(accessRule *models.AccessRule) api.AccessRule {
	return api.AccessRule{
		ScopeId:     accessRule.ScopeId,
		ScopeType:   accessRule.ScopeType,
		ScopeName:   accessRule.ScopeName,
		SubjectId:   accessRule.SubjectId,
		SubjectType: accessRule.SubjectType,
		RoleId:      int32(accessRule.RoleID),
		RoleName:    accessRule.Role.Name,
		Id:          int32(accessRule.ID),
		CreatedAt:   accessRule.CreatedAt,
		UpdatedAt:   accessRule.UpdatedAt,
		DeletedAt:   dbmodels.GormDeletedAtToNullableTime(accessRule.DeletedAt),
		CreatedBy:   accessRule.CreatedBy,
		TenantId:    int32(accessRule.TenantId),
		ClusterId:   accessRule.ClusterUuid,
		Status:      accessRuleStatusResponseDaoToApi(accessRule),
	}
}

func AccessRuleDaoToSyncApi(accessRule *models.AccessRule, predefinedRole *models.Role) api.AccessRuleSync {
	var preDefinedRoleApi *api.AccessRuleSyncAllOfPredefinedRole = nil
	if predefinedRole != nil {
		preDefinedRoleApi = ptr.To(api.AccessRuleSyncAllOfPredefinedRole{
			Id:   fmt.Sprint(predefinedRole.ID),
			Name: predefinedRole.Name,
		})
	}

	return api.AccessRuleSync{
		ScopeId:        accessRule.ScopeId,
		ScopeType:      accessRule.ScopeType,
		ScopeName:      accessRule.ScopeName,
		SubjectId:      accessRule.SubjectId,
		SubjectType:    accessRule.SubjectType,
		RoleId:         int32(accessRule.RoleID),
		RoleName:       accessRule.Role.Name,
		Id:             int32(accessRule.ID),
		CreatedAt:      accessRule.CreatedAt,
		UpdatedAt:      accessRule.UpdatedAt,
		DeletedAt:      dbmodels.GormDeletedAtToNullableTime(accessRule.DeletedAt),
		CreatedBy:      accessRule.CreatedBy,
		TenantId:       int32(accessRule.TenantId),
		ClusterId:      accessRule.ClusterUuid,
		PredefinedRole: preDefinedRoleApi,
		IsDeleting:     accessRule.Phase == string(api.ACCESSRULEPHASE_DELETING),
	}
}

func AccessRuleDaoToMeApi(accessRule *models.AccessRule) api.MeAccessRule {
	return api.MeAccessRule{
		ScopeId:     accessRule.ScopeId,
		ScopeType:   accessRule.ScopeType,
		ScopeName:   accessRule.ScopeName,
		SubjectId:   accessRule.SubjectId,
		SubjectType: accessRule.SubjectType,
		RoleId:      int32(accessRule.RoleID),
		RoleName:    accessRule.Role.Name,
		Id:          int32(accessRule.ID),
		ClusterId:   accessRule.ClusterUuid,
		TenantId:    int32(accessRule.TenantId),
	}
}

func ConvertToAccessRuleById(accessRule api.AccessRule, scopePath *string) api.AccessRuleById {
	return api.AccessRuleById{
		SubjectId:   accessRule.SubjectId,
		SubjectType: accessRule.SubjectType,
		RoleId:      accessRule.RoleId,
		ScopeId:     accessRule.ScopeId,
		ScopeType:   accessRule.ScopeType,
		ClusterId:   accessRule.ClusterId,
		RoleName:    accessRule.RoleName,
		ScopeName:   accessRule.ScopeName,
		Id:          accessRule.Id,
		CreatedAt:   accessRule.CreatedAt,
		UpdatedAt:   accessRule.UpdatedAt,
		DeletedAt:   accessRule.DeletedAt,
		TenantId:    accessRule.TenantId,
		CreatedBy:   accessRule.CreatedBy,
		Status:      accessRule.Status,
		ScopePath:   scopePath,
	}
}

func accessRuleStatusResponseDaoToApi(accessRule *models.AccessRule) api.AccessRuleStatusResponse {
	if accessRule.Phase == "" {
		return api.AccessRuleStatusResponse{}
	}

	return api.AccessRuleStatusResponse{
		Phase:   api.AccessRulePhase(accessRule.Phase),
		Message: accessRule.PhaseMessage,
		Details: getAccessRuleStatusResponseDetails(accessRule.PhaseByCluster),
	}
}

func getAccessRuleStatusResponseDetails(phaseByCluster models.PhaseByCluster) string {
	if len(phaseByCluster) == 0 {
		return ""
	}

	notReadyClusterIds := []string{}
	readyClusterIds := []string{}
	for clusterId, clusterPhase := range phaseByCluster {
		if clusterPhase == string(api.ACCESSRULECLUSTERPHASE_NOT_READY) {
			notReadyClusterIds = append(notReadyClusterIds, clusterId)
		} else if clusterPhase == string(api.ACCESSRULECLUSTERPHASE_READY) {
			readyClusterIds = append(readyClusterIds, clusterId)
		}
		// phases should be only Ready/NotReady here
	}

	details := []string{}
	if len(notReadyClusterIds) > 0 {
		sort.Strings(notReadyClusterIds)
		details = append(details, fmt.Sprintf("%s: [%s]", api.ACCESSRULECLUSTERPHASE_NOT_READY, strings.Join(notReadyClusterIds, ", ")))
	}
	if len(readyClusterIds) > 0 {
		sort.Strings(readyClusterIds)
		details = append(details, fmt.Sprintf("%s: [%s]", api.ACCESSRULECLUSTERPHASE_READY, strings.Join(readyClusterIds, ", ")))
	}

	return strings.Join(details, "; ")
}

func PermissionsDaoToApi(permissions []*models.Permission) []api.Permission {
	result := []api.Permission{}
	for _, permission := range permissions {
		// Filtering out unknown resource types.
		// This case could happen when running in gradual rollout and the new code adds a new resource type but the old services are getting it from the DB but cant parse it into the enum
		if !api.ResourceType.IsValid(permission.ResourceType) {
			continue
		}

		actions := []api.Action{}
		actions = AppendAction(actions, permission.Create, api.ACTION_CREATE)
		actions = AppendAction(actions, permission.Read, api.ACTION_READ)
		actions = AppendAction(actions, permission.Update, api.ACTION_UPDATE)
		actions = AppendAction(actions, permission.Delete, api.ACTION_DELETE)
		actions = AppendAction(actions, permission.Sync, api.ACTION_SYNC)

		result = append(result, api.Permission{
			ResourceType: permission.ResourceType,
			DisplayName:  GetResourceTypeDisplayName(permission.ResourceType),
			GroupId:      GetResourceTypeGroupId(permission.ResourceType),
			Actions:      actions,
		})
	}

	return result
}

func RolePermissionsDaoToApi(permissions []*models.RolePermission) []api.Permission {
	result := []api.Permission{}
	for _, permission := range permissions {
		// Filtering out unknown resource types.
		// This case could happen when running in gradual rollout and the new code adds a new resource type but the old services are getting it from the DB but cant parse it into the enum
		if !api.ResourceType.IsValid(permission.ResourceType) {
			continue
		}

		actions := []api.Action{}
		actions = AppendAction(actions, permission.Create, api.ACTION_CREATE)
		actions = AppendAction(actions, permission.Read, api.ACTION_READ)
		actions = AppendAction(actions, permission.Update, api.ACTION_UPDATE)
		actions = AppendAction(actions, permission.Delete, api.ACTION_DELETE)
		actions = AppendAction(actions, permission.Sync, api.ACTION_SYNC)

		result = append(result, api.Permission{
			ResourceType: permission.ResourceType,
			DisplayName:  GetResourceTypeDisplayName(permission.ResourceType),
			GroupId:      GetResourceTypeGroupId(permission.ResourceType),
			Actions:      actions,
		})
	}
	result = addMissingResourceTypesV1(result)
	return result
}

func PermissionsSetDaoToPermissionApi(permissionSets *[]models.PermissionSet) []api.Permission {
	apiPermissions := []api.Permission{}
	apiPermissionsMap := map[string]api.Permission{}
	for _, ps := range *permissionSets {
		for _, permission := range ps.Permissions {
			// Filtering out unknown resource types.
			// This case could happen when running in gradual rollout and the new code adds a new resource type but the old services are getting it from the DB but cant parse it into the enum
			if !api.ResourceType.IsValid(permission.ResourceType) {
				continue
			}

			apiPermission, found := apiPermissionsMap[string(permission.ResourceType)]
			if !found {
				apiPermission = api.Permission{
					ResourceType: permission.ResourceType,
					DisplayName:  GetResourceTypeDisplayName(permission.ResourceType),
					GroupId:      GetResourceTypeGroupId(permission.ResourceType),
					Actions:      []api.Action{},
				}
			}

			actions := apiPermission.Actions
			actions = AppendAction(actions, permission.Create, api.ACTION_CREATE)
			actions = AppendAction(actions, permission.Read, api.ACTION_READ)
			actions = AppendAction(actions, permission.Update, api.ACTION_UPDATE)
			actions = AppendAction(actions, permission.Delete, api.ACTION_DELETE)
			apiPermission.Actions = actions

			apiPermissionsMap[string(permission.ResourceType)] = apiPermission
		}
	}
	for _, apiPermission := range apiPermissionsMap {
		apiPermissions = append(apiPermissions, apiPermission)
	}

	apiPermissions = addMissingResourceTypesV1(apiPermissions)

	return apiPermissions
}

func PermissionsSetDaoToRolePermApi(permissionSets *[]models.PermissionSet) ([]api.RolePermissionSet, []api.RolePermission) {
	apiPermissionsMap := map[string]api.RolePermission{}
	apiPermissions := []api.RolePermission{}
	apiPermissionSets := []api.RolePermissionSet{}
	for _, ps := range *permissionSets {
		if !ps.Internal {
			apiPermissionSets = append(apiPermissionSets, api.RolePermissionSet{
				Id:   ps.ID.String(),
				Name: ps.Name,
			})
		}
		for _, permission := range ps.Permissions {
			// Filtering out unknown resource types.
			// This case could happen when running in gradual rollout and the new code adds a new resource type but the old services are getting it from the DB but cant parse it into the enum
			if !api.ResourceType.IsValid(permission.ResourceType) {
				continue
			}

			rolePermission, found := apiPermissionsMap[string(permission.ResourceType)]
			if !found {
				rolePermission = api.RolePermission{
					ResourceType: permission.ResourceType,
					DisplayName:  GetResourceTypeDisplayName(permission.ResourceType),
					Actions:      []api.PermissionAction{},
				}
			}

			actions := rolePermission.Actions
			actions = AppendPermissionAction(actions, permission.Create, api.PERMISSIONACTION_CREATE)
			actions = AppendPermissionAction(actions, permission.Read, api.PERMISSIONACTION_READ)
			actions = AppendPermissionAction(actions, permission.Update, api.PERMISSIONACTION_UPDATE)
			actions = AppendPermissionAction(actions, permission.Delete, api.PERMISSIONACTION_DELETE)
			rolePermission.Actions = actions

			apiPermissionsMap[string(permission.ResourceType)] = rolePermission
		}
	}
	for _, apiPermission := range apiPermissionsMap {
		apiPermissions = append(apiPermissions, apiPermission)
	}

	apiPermissions = addMissingResourceTypes(apiPermissions)

	return apiPermissionSets, apiPermissions
}

func addMissingResourceTypes(permissions []api.RolePermission) []api.RolePermission {
	permissionsMap := map[api.ResourceType]bool{}
	for _, permission := range permissions {
		permissionsMap[permission.ResourceType] = true
	}

	for _, resourceType := range api.AllowedResourceTypeEnumValues {
		if _, found := permissionsMap[resourceType]; !found {
			permissions = append(permissions, api.RolePermission{
				ResourceType: resourceType,
				DisplayName:  GetResourceTypeDisplayName(resourceType),
				Actions:      []api.PermissionAction{},
			})
		}
	}

	return permissions
}

// duplicated - deprecated
func addMissingResourceTypesV1(permissions []api.Permission) []api.Permission {
	permissionsMap := map[api.ResourceType]bool{}
	for _, permission := range permissions {
		permissionsMap[permission.ResourceType] = true
	}

	for _, resourceType := range api.AllowedResourceTypeEnumValues {
		if _, found := permissionsMap[resourceType]; !found {
			permissions = append(permissions, api.Permission{
				ResourceType: resourceType,
				DisplayName:  GetResourceTypeDisplayName(resourceType),
				GroupId:      GetResourceTypeGroupId(resourceType),
				Actions:      []api.Action{},
			})
		}
	}

	return permissions
}

func GetResourceTypeDisplayName(resourceType api.ResourceType) string {
	if displayName, found := ResourceTypesDisplayName[resourceType]; found {
		return displayName
	}

	return ""
}

func GetResourceTypeGroupId(resourceType api.ResourceType) api.ResourceTypeGroupId {
	if groupId, found := ResourceTypesGroupId[resourceType]; found {
		return groupId
	}
	return ""
}

func AppendAction(actions []api.Action, toAppend bool, value api.Action) []api.Action {
	if toAppend && !Contains(actions, value) {
		actions = append(actions, value)
	}
	return actions
}

func AppendPermissionAction(actions []api.PermissionAction, toAppend bool, value api.PermissionAction) []api.PermissionAction {
	if toAppend && !Contains(actions, value) {
		actions = append(actions, value)
	}
	return actions
}

func Contains[T comparable](arr []T, value T) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

func NoContentResponse() api.ImplResponse {
	return api.ImplResponse{
		Code: http.StatusNoContent,
	}
}

func NewErrorResponse(apiError *api.Error) api.ImplResponse {
	apiError.Details = ""
	return api.ImplResponse{
		Code: int(apiError.Code),
		Body: *apiError,
	}
}

func ErrorToApiError(err error, errorCode int) api.ImplResponse {
	return api.ImplResponse{
		Code: errorCode,
		Body: err.Error(),
	}
}

func parseAuthData(ctx context.Context) (uint, string, []string, *api.Error) {
	var ok bool
	var subjectId string
	if subjectId, ok = ctx.Value(runaictx.CfSubjectID).(string); !ok {
		return 0, "", nil, api.NewError(http.StatusBadRequest, "No subject id provided", nil)
	}

	var groups []string
	if groups, ok = ctx.Value(runaictx.CfGroups).([]string); !ok {
		groups = []string{}
	}
	groups = decodeGroups(groups)

	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return 0, "", nil, err
	}

	return tenantId, subjectId, groups, nil
}

func decodeGroups(groups []string) []string {
	decoded := make([]string, len(groups))
	for i, g := range groups {
		if unescaped, err := url.QueryUnescape(g); err == nil {
			decoded[i] = unescaped
		} else {
			decoded[i] = g
		}
	}
	return decoded
}

func parseTenantId(ctx context.Context) (uint, *api.Error) {
	var ok bool

	var tenantId string
	if tenantId, ok = ctx.Value(runaictx.CfTenantId).(string); !ok {
		return 0, api.NewError(http.StatusBadRequest, "No tenant id provided", nil)
	}

	tenantIdUint, parseErr := strconv.Atoi(tenantId)
	if parseErr != nil {
		return 0, api.NewError(http.StatusBadRequest, "Invalid tenant id", nil)
	}

	return uint(tenantIdUint), nil
}

func convertMapKeysToSlice[T comparable, D any](m map[T]D) []T {
	var arr []T
	for k := range m {
		arr = append(arr, k)
	}
	return arr
}

func removeDuplicates[T comparable](arr []T) []T {
	set := map[T]bool{}

	for _, item := range arr {
		set[item] = true
	}
	return convertMapKeysToSlice(set)
}

func getClusterUuid(ctx context.Context, scopeType api.ScopeType, scopeId string) string {
	if scopeType == api.SCOPETYPE_TENANT || scopeType == api.SCOPETYPE_SYSTEM {
		return ""
	}
	if scopeType == api.SCOPETYPE_CLUSTER {
		return scopeId
	}

	ancestors, _, err := org_cache.GetOrgHierarchyManager().GetAncestors(ctx, scopeId, api.ResourceType(scopeType))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get the resource %s %s ancestors.", scopeId, string(scopeType))
		return ""
	}
	if ancestors != nil && ancestors.ClusterUuid != nil {
		return *ancestors.ClusterUuid
	}

	return ""
}

type SyncSlice[T any] struct {
	mu    sync.RWMutex
	items []T
}

func (s *SyncSlice[T]) Append(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
}

func (s *SyncSlice[T]) GetAll() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]T(nil), s.items...)
}

// shouldMapServiceAccountToApp returns true if service-account should be mapped to app for this cluster
func shouldMapServiceAccountToApp(ctx context.Context, clusterId string) bool {
	if clusterId == "" {
		return false
	}

	capabilitiesChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	supported, _ := capabilitiesChecker.ValidateClusterCapability(ctx, checker.MinVersionOfServiceAccountInGetAccessRulesApi)
	return !supported
}
