package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/google/uuid"

	static_permission_sets "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-permission-sets"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"
	static_legacy_roles "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	static_roles "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	testing "github.com/run-ai/runai/control-plane/authorization/pkg/test"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"golang.org/x/exp/slices"
	"k8s.io/utils/ptr"
)

var _ = Describe("Roles API tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	mgmtTenantId := fmt.Sprintf("%d", testing.ManagementTestTenant)
	createdRoles := []api.Role{}
	createdAccessRules := []api.AccessRule{}
	var tenantScopedRole, tenantScopedRolePS, systemScopedRole api.Role

	ctx := context.WithValue(setContext(tenantId, TestUserId, []string{dao.SuperAdmins}), authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
		Tenants: []string{tenantId},
	})
	mgmtCtx := context.WithValue(context.WithValue(setContext(mgmtTenantId, TestUserId, []string{dao.SuperAdmins}), runaictx.CfIsManagementTenant, "true"), authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
		System:  ptr.To(true),
		Tenants: []string{mgmtTenantId},
	})

	tenantSettings := NewMockTenantSettingsService(map[string]bool{})

	clustersClient := cluster_service_client.NewAPIClient(&cluster_service_client.Configuration{
		Servers: cluster_service_client.ServerConfigurations{
			{
				URL: config.Get[*authorization_config.Server]().ClusterServiceUrl,
			},
		},
	})

	rolesApiService := NewRolesApiService(clustersClient, tenantSettings)
	accessRulesApiService := NewAccessRulesApiService(tenantSettings)

	AfterAll(func() {
		for _, accessRule := range createdAccessRules {
			_, err := accessRulesApiService.DeleteAccessRule(ctx, accessRule.Id)
			Expect(err).To(BeNil())
		}
		createdAccessRules = []api.AccessRule{}

		for _, role := range createdRoles {
			// Use the appropriate context based on role scope type
			deleteCtx := ctx
			if role.ScopeType == systemScopeType {
				deleteCtx = mgmtCtx
			}
			_, err := rolesApiService.DeleteRole(deleteCtx, role.Id)
			Expect(err).To(BeNil())
		}
		createdRoles = []api.Role{}

		TestsCleanupAndInit()
	})

	BeforeAll(func() {
		TestsCleanupAndInit()
	})

	It("Each resource types should have a display name", func() {
		Expect(len(ResourceTypesDisplayName)).To(Equal(len(api.AllowedResourceTypeEnumValues)))
	})

	It("Each resource type should have a group id", func() {
		for _, resourceType := range api.AllowedResourceTypeEnumValues {
			Expect(ResourceTypesGroupId[resourceType]).ToNot(BeEmpty())
		}
	})

	It("validate static roles have minimal resource types when needed", func() {
		var violations []string

		// Collect all static roles to validate
		allStaticRoles := []models.Role{}
		allStaticRoles = append(allStaticRoles, dao.LegacyRoles...)
		allStaticRoles = append(allStaticRoles, dao.RolesMigrated...)
		allStaticRoles = append(allStaticRoles, dao.RolesNew...)
		allStaticRoles = append(allStaticRoles, static_roles.ApplicationAdmin)
		allStaticRoles = append(allStaticRoles, static_roles.CloudSupport)

		// Check all static roles
		for _, role := range allStaticRoles {
			if role.Permissions == nil {
				continue
			}

			hasCluster := false
			hasClustersMinimal := false
			hasNodePools := false
			hasNodePoolsMinimal := false

			// Check what resource types this role has in its direct Permissions
			for _, permission := range role.Permissions {
				switch permission.ResourceType {
				case api.RESOURCETYPE_CLUSTER:
					hasCluster = true
				case api.RESOURCETYPE_CLUSTERS_MINIMAL:
					hasClustersMinimal = true
				case api.RESOURCETYPE_NODEPOOLS:
					hasNodePools = true
				case api.RESOURCETYPE_NODEPOOLS_MINIMAL:
					hasNodePoolsMinimal = true
				}
			}

			// Rule: If a role has the "full" resource type (CLUSTER or NODEPOOLS),
			// it MUST also have the corresponding "minimal" resource type.
			// Note: Having only the minimal version without the full version is allowed.

			if hasCluster && !hasClustersMinimal {
				violations = append(violations, fmt.Sprintf("Role '%s' has RESOURCETYPE_CLUSTER but missing RESOURCETYPE_CLUSTERS_MINIMAL", role.Name))
			}

			if hasNodePools && !hasNodePoolsMinimal {
				violations = append(violations, fmt.Sprintf("Role '%s' has RESOURCETYPE_NODEPOOLS but missing RESOURCETYPE_NODEPOOLS_MINIMAL", role.Name))
			}
		}

		// If there are violations, fail the test with all violations listed
		if len(violations) > 0 {
			errorMessage := fmt.Sprintf("Found %d static role(s) with missing minimal resource types:\n", len(violations))
			for _, violation := range violations {
				errorMessage += fmt.Sprintf("  - %s\n", violation)
			}
			Fail(errorMessage)
		}
	})

	It("Post tenant role", func() {
		body := api.RoleCreationFieldsInternal{
			Name:        "tenant role",
			Description: "first tenant test role, can fully manage projects",
			Permissions: []api.PostPermission{
				{
					ResourceType: api.RESOURCETYPE_PROJECT,
					Actions: []api.Action{
						api.ACTION_CREATE,
						api.ACTION_READ,
						api.ACTION_UPDATE,
						api.ACTION_DELETE,
					},
				},
			},
			ScopeType: tenantScopeType,
			ScopeId:   ptr.To(tenantId),
			Enabled:   ptr.To(false),
		}

		role := createTestRole(ctx, rolesApiService, body, &createdRoles, false, http.StatusCreated)
		createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusConflict)
		Expect(role).ToNot(BeNil())
		Expect(role.Id).ToNot(BeNil())
		Expect(role.Name).To(Equal(body.Name))
		Expect(role.Description).To(Equal(body.Description))
		Expect(role.ScopeType).To(Equal(body.ScopeType))
		Expect(role.ScopeId).ToNot(BeNil())
		Expect(role.ScopeId).ToNot(Equal(tenantId))
		Expect(role.Enabled).To(Equal(body.Enabled))
		Expect(role.Custom).To(Equal(true))
		Expect(role.KubernetesPermissions.PredefinedRole).To(BeNil())
		Expect(len(role.Permissions)).To(Equal(len(api.AllowedResourceTypeEnumValues)))
		Expect(role.Permissions[0].ResourceType).To(Equal(body.Permissions[0].ResourceType))
		Expect(permissionActionsToActions(role.Permissions[0].Actions)).To(Equal(body.Permissions[0].Actions))

		tenantScopedRole = *role

		for _, permission := range role.Permissions {
			if permission.ResourceType == body.Permissions[0].ResourceType {
				Expect(permissionActionsToActions(permission.Actions)).To(Equal(body.Permissions[0].Actions))
			} else {
				Expect(len(permission.Actions)).To(Equal(0))
			}
		}

		response, err := rolesApiService.RoleDisable(ctx, tenantScopedRole.Id, api.RoleEnabledDisabledFields{})
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusNoContent))

		roleFromServer, err := rolesApiService.GetRole(ctx, tenantScopedRole.Id)
		Expect(err).To(BeNil())
		Expect(roleFromServer.Code).To(Equal(http.StatusOK))

		apiRole := httpResponseBodyToRole(roleFromServer)
		Expect(*apiRole.Enabled).To(Equal(false))
		Expect(apiRole.EffectiveEnabled).To(Equal(false))

		response, err = rolesApiService.RoleEnable(ctx, tenantScopedRole.Id, api.RoleEnabledDisabledFields{})
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusNoContent))

		roleFromServer, err = rolesApiService.GetRole(ctx, tenantScopedRole.Id)
		Expect(err).To(BeNil())
		Expect(roleFromServer.Code).To(Equal(http.StatusOK))

		apiRole = httpResponseBodyToRole(roleFromServer)
		Expect(*apiRole.Enabled).To(Equal(true))
		Expect(apiRole.EffectiveEnabled).To(Equal(true))

		tenantScopedRole.Enabled = ptr.To(true)
		tenantScopedRole.EffectiveEnabled = true
	})

	It("Post system role", func() {
		body := api.RoleCreationFieldsInternal{
			Name:        "system role",
			Description: "first system test role, can fully manage projects",
			Permissions: []api.PostPermission{
				{
					ResourceType: api.RESOURCETYPE_PROJECT,
					Actions: []api.Action{
						api.ACTION_CREATE,
						api.ACTION_READ,
						api.ACTION_UPDATE,
						api.ACTION_DELETE,
					},
				},
			},
			ScopeType: "system",
			Enabled:   ptr.To(false),
		}

		createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusForbidden)
		role := createTestRole(mgmtCtx, rolesApiService, body, &createdRoles, false, http.StatusCreated)

		Expect(role).ToNot(BeNil())
		Expect(role.Id).ToNot(BeNil())
		Expect(role.Name).To(Equal(body.Name))
		Expect(role.Description).To(Equal(body.Description))
		Expect(role.ScopeType).To(Equal(body.ScopeType))
		Expect(role.ScopeId).To(BeNil())
		Expect(role.Enabled).To(Equal(body.Enabled))
		Expect(role.Custom).To(Equal(true))
		Expect(role.KubernetesPermissions.PredefinedRole).To(BeNil())
		Expect(len(role.Permissions)).To(Equal(len(api.AllowedResourceTypeEnumValues)))
		Expect(role.Permissions[0].ResourceType).To(Equal(body.Permissions[0].ResourceType))
		Expect(permissionActionsToActions(role.Permissions[0].Actions)).To(Equal(body.Permissions[0].Actions))

		systemScopedRole = *role

		for _, permission := range role.Permissions {
			if permission.ResourceType == body.Permissions[0].ResourceType {
				Expect(permissionActionsToActions(permission.Actions)).To(Equal(body.Permissions[0].Actions))
			} else {
				Expect(len(permission.Actions)).To(Equal(0))
			}
		}
	})

	It("Post tenant role with permission set", func() {
		if !config.Get[*authorization_config.General]().UsePermissionSets {
			Skip("no permission sets")
		}
		body := api.RoleCreationFieldsInternal{
			Name:        "tenant role ps",
			Description: "first tenant test role, can read workloads",
			PermissionSets: []api.RolePermissionSet{{
				Id:   dao.MapStaticPermissionSetNameToId[static_permission_sets.WorkloadRead.Name].String(),
				Name: static_permission_sets.WorkloadRead.Name,
			}},
			ScopeType: tenantScopeType,
			ScopeId:   ptr.To(tenantId),
			Enabled:   ptr.To(false),
		}

		role := createTestRole(ctx, rolesApiService, body, &createdRoles, false, http.StatusCreated)
		createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusConflict)
		Expect(role).ToNot(BeNil())
		Expect(role.Id).ToNot(BeNil())
		Expect(role.Name).To(Equal(body.Name))
		Expect(role.Description).To(Equal(body.Description))
		Expect(role.ScopeType).To(Equal(body.ScopeType))
		Expect(role.ScopeId).ToNot(BeNil())
		Expect(role.ScopeId).ToNot(Equal(tenantId))
		Expect(role.Enabled).To(Equal(body.Enabled))
		Expect(role.Custom).To(Equal(true))
		Expect(role.KubernetesPermissions.PredefinedRole).To(BeNil())
		Expect(len(role.Permissions)).To(Equal(len(api.AllowedResourceTypeEnumValues)))
		for _, permission := range role.Permissions {
			switch permission.ResourceType {
			case api.RESOURCETYPE_PROJECT, api.RESOURCETYPE_WORKLOAD_PROPERTIES, api.RESOURCETYPE_INFERENCES, api.RESOURCETYPE_TRAININGS, api.RESOURCETYPE_WORKSPACES, api.RESOURCETYPE_WORKLOADS, api.RESOURCETYPE_NODEPOOLS_MINIMAL, api.RESOURCETYPE_CLUSTERS_MINIMAL:
				Expect(len(permission.Actions)).To(Equal(1))
				Expect(string(permission.Actions[0])).To(Equal("read"))
			default:
				Expect(len(permission.Actions)).To(Equal(0))
			}
		}

		tenantScopedRolePS = *role

		response, err := rolesApiService.RoleEnable(ctx, tenantScopedRolePS.Id, api.RoleEnabledDisabledFields{})
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusNoContent))

		tenantScopedRolePS.Enabled = ptr.To(true)
		tenantScopedRolePS.EffectiveEnabled = true
	})

	It("Post tenant role with permission set - invalid permission sets", func() {
		if !config.Get[*authorization_config.General]().UsePermissionSets {
			Skip("no permission sets")
		}
		body := api.RoleCreationFieldsInternal{
			Name:        "tenant role ps invalid",
			Description: "first tenant test role, can read workloads",
			PermissionSets: []api.RolePermissionSet{{
				Id:   dao.MapStaticPermissionSetNameToId[static_permission_sets.WorkloadRead.Name].String(),
				Name: static_permission_sets.WorkloadRead.Name,
			}},
			Permissions: []api.PostPermission{
				{
					ResourceType: api.RESOURCETYPE_NODEPOOLS,
					Actions:      []api.Action{api.ACTION_READ},
				},
			},
			ScopeType: tenantScopeType,
			ScopeId:   ptr.To(tenantId),
			Enabled:   ptr.To(false),
		}

		createTestRole(ctx, rolesApiService, body, &createdRoles, false, http.StatusBadRequest)

		body = api.RoleCreationFieldsInternal{
			Name:        "tenant role ps invalid",
			Description: "first tenant test role, can read workloads",
			PermissionSets: []api.RolePermissionSet{{
				Id:   dao.MapStaticPermissionSetNameToId[static_permission_sets.WorkloadRead.Name].String(),
				Name: "lalala",
			}},
			ScopeType: tenantScopeType,
			ScopeId:   ptr.To(tenantId),
			Enabled:   ptr.To(false),
		}

		createTestRole(ctx, rolesApiService, body, &createdRoles, false, http.StatusBadRequest)

		newUUID, _ := uuid.NewUUID()
		body = api.RoleCreationFieldsInternal{
			Name:        "tenant role ps invalid",
			Description: "first tenant test role, can read workloads",
			PermissionSets: []api.RolePermissionSet{{
				Id:   newUUID.String(),
				Name: static_permission_sets.WorkloadRead.Name,
			}},
			ScopeType: tenantScopeType,
			ScopeId:   ptr.To(tenantId),
			Enabled:   ptr.To(false),
		}

		createTestRole(ctx, rolesApiService, body, &createdRoles, false, http.StatusBadRequest)

		body = api.RoleCreationFieldsInternal{
			Name:        "tenant role ps invalid",
			Description: "first tenant test role, can read workloads",
			PermissionSets: []api.RolePermissionSet{{
				Id:   newUUID.String(),
				Name: "lalala",
			}},
			ScopeType: tenantScopeType,
			ScopeId:   ptr.To(tenantId),
			Enabled:   ptr.To(false),
		}

		createTestRole(ctx, rolesApiService, body, &createdRoles, false, http.StatusBadRequest)
	})

	It("Post system role with permission set", func() {
		if !config.Get[*authorization_config.General]().UsePermissionSets {
			Skip("no permission sets")
		}
		body := api.RoleCreationFieldsInternal{
			Name:        "system role ps",
			Description: "first system test role, can read workloads",
			PermissionSets: []api.RolePermissionSet{{
				Id:   dao.MapStaticPermissionSetNameToId[static_permission_sets.WorkloadRead.Name].String(),
				Name: static_permission_sets.WorkloadRead.Name,
			}},
			ScopeType: "system",
			Enabled:   ptr.To(false),
		}

		createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusForbidden)
		role := createTestRole(mgmtCtx, rolesApiService, body, &createdRoles, false, http.StatusCreated)

		Expect(role).ToNot(BeNil())
		Expect(role.Id).ToNot(BeNil())
		Expect(role.Name).To(Equal(body.Name))
		Expect(role.Description).To(Equal(body.Description))
		Expect(role.ScopeType).To(Equal(body.ScopeType))
		Expect(role.ScopeId).To(BeNil())
		Expect(role.Enabled).To(Equal(body.Enabled))
		Expect(role.Custom).To(Equal(true))
		Expect(role.KubernetesPermissions.PredefinedRole).To(BeNil())
		Expect(len(role.Permissions)).To(Equal(len(api.AllowedResourceTypeEnumValues)))
		for _, permission := range role.Permissions {
			switch permission.ResourceType {
			case api.RESOURCETYPE_PROJECT, api.RESOURCETYPE_WORKLOAD_PROPERTIES, api.RESOURCETYPE_INFERENCES, api.RESOURCETYPE_TRAININGS, api.RESOURCETYPE_WORKSPACES, api.RESOURCETYPE_WORKLOADS, api.RESOURCETYPE_NODEPOOLS_MINIMAL, api.RESOURCETYPE_CLUSTERS_MINIMAL:
				Expect(len(permission.Actions)).To(Equal(1))
				Expect(string(permission.Actions[0])).To(Equal("read"))
			default:
				Expect(len(permission.Actions)).To(Equal(0))
			}
		}
	})

	It("Post tenant role with KubernetesPermissions", func() {
		permissions := []api.PostPermission{}
		editorRoleId, ok := dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name]
		Expect(ok).To(BeTrue())
		editorRes, _ := rolesApiService.GetRole(ctx, int32(editorRoleId))
		Expect(editorRes.Code).To(Equal(http.StatusOK))
		editorPermissions := httpResponseBodyToRole(editorRes).Permissions
		permissionMap := map[api.ResourceType][]api.Action{}
		for _, editorPermission := range editorPermissions {
			if editorPermission.ResourceType == api.RESOURCETYPE_PROJECT {
				continue
			}
			actions := []api.Action{}
			for _, action := range editorPermission.Actions {
				actions = append(actions, api.Action(action))
			}
			permissionMap[editorPermission.ResourceType] = actions
			permissions = append(permissions, api.PostPermission{
				ResourceType: editorPermission.ResourceType,
				Actions:      actions,
			})
		}
		permissions = append(permissions, api.PostPermission{
			ResourceType: api.RESOURCETYPE_PROJECT,
			Actions: []api.Action{
				api.ACTION_CREATE,
				api.ACTION_READ,
				api.ACTION_UPDATE,
				api.ACTION_DELETE,
			},
		})
		permissionMap[api.RESOURCETYPE_PROJECT] = []api.Action{
			api.ACTION_CREATE,
			api.ACTION_READ,
			api.ACTION_UPDATE,
			api.ACTION_DELETE,
		}
		body := api.RoleCreationFieldsInternal{
			Name:        "tenant role kubernetes permissions",
			Description: "first tenant test role, can fully manage projects",
			Permissions: permissions,
			ScopeType:   "tenant",
			ScopeId:     ptr.To(tenantId),
			Enabled:     ptr.To(true),
			KubernetesPermissions: api.KubernetesPermissions{
				ptr.To(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name])),
			},
		}

		role := createTestRole(ctx, rolesApiService, body, &createdRoles, false, http.StatusCreated)
		createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusConflict)
		Expect(role).ToNot(BeNil())
		Expect(role.Id).ToNot(BeNil())
		Expect(role.Name).To(Equal(body.Name))
		Expect(role.Description).To(Equal(body.Description))
		Expect(role.ScopeType).To(Equal(body.ScopeType))
		Expect(role.ScopeId).ToNot(BeNil())
		Expect(role.ScopeId).ToNot(Equal(tenantId))
		Expect(role.Enabled).To(Equal(body.Enabled))
		Expect(role.Custom).To(Equal(true))
		Expect(role.KubernetesPermissions.PredefinedRole).To(Not(BeNil()))
		Expect(*role.KubernetesPermissions.PredefinedRole).To(Equal(*body.KubernetesPermissions.PredefinedRole))
		Expect(len(role.Permissions)).To(Equal(len(api.AllowedResourceTypeEnumValues)))

		for _, permission := range role.Permissions {
			actions := permissionMap[permission.ResourceType]
			Expect(len(permission.Actions)).To(Equal(len(actions)))
			Expect(permissionActionsToActions(permission.Actions)).To(ContainElements(actions))
		}
	})

	It("Post system role with KubernetesPermissions", func() {
		permissions := []api.PostPermission{}
		editorRoleId, ok := dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name]
		Expect(ok).To(BeTrue())
		editorRes, _ := rolesApiService.GetRole(ctx, int32(editorRoleId))
		Expect(editorRes.Code).To(Equal(http.StatusOK))
		editorPermissions := httpResponseBodyToRole(editorRes).Permissions
		permissionMap := map[api.ResourceType][]api.Action{}
		for _, editorPermission := range editorPermissions {
			if editorPermission.ResourceType == api.RESOURCETYPE_PROJECT {
				continue
			}
			actions := []api.Action{}
			for _, action := range editorPermission.Actions {
				actions = append(actions, api.Action(action))
			}
			permissionMap[editorPermission.ResourceType] = actions
			permissions = append(permissions, api.PostPermission{
				ResourceType: editorPermission.ResourceType,
				Actions:      actions,
			})
		}
		permissions = append(permissions, api.PostPermission{
			ResourceType: api.RESOURCETYPE_PROJECT,
			Actions: []api.Action{
				api.ACTION_CREATE,
				api.ACTION_READ,
				api.ACTION_UPDATE,
				api.ACTION_DELETE,
			},
		})
		permissionMap[api.RESOURCETYPE_PROJECT] = []api.Action{
			api.ACTION_CREATE,
			api.ACTION_READ,
			api.ACTION_UPDATE,
			api.ACTION_DELETE,
		}
		body := api.RoleCreationFieldsInternal{
			Name:        "system role kubernetes permissions",
			Description: "first system test role, can fully manage projects",
			Permissions: permissions,
			ScopeType:   "system",
			Enabled:     ptr.To(false),
			KubernetesPermissions: api.KubernetesPermissions{
				ptr.To(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name])),
			},
		}

		createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusForbidden)
		role := createTestRole(mgmtCtx, rolesApiService, body, &createdRoles, false, http.StatusCreated)

		Expect(role).ToNot(BeNil())
		Expect(role.Id).ToNot(BeNil())
		Expect(role.Name).To(Equal(body.Name))
		Expect(role.Description).To(Equal(body.Description))
		Expect(role.ScopeType).To(Equal(body.ScopeType))
		Expect(role.ScopeId).To(BeNil())
		Expect(role.Enabled).To(Equal(body.Enabled))
		Expect(role.Custom).To(Equal(true))
		Expect(role.KubernetesPermissions.PredefinedRole).To(Not(BeNil()))
		Expect(*role.KubernetesPermissions.PredefinedRole).To(Equal(*body.KubernetesPermissions.PredefinedRole))
		Expect(len(role.Permissions)).To(Equal(len(api.AllowedResourceTypeEnumValues)))

		for _, permission := range role.Permissions {
			actions := permissionMap[permission.ResourceType]
			Expect(len(permission.Actions)).To(Equal(len(actions)))
			Expect(permissionActionsToActions(permission.Actions)).To(ContainElements(actions))
		}

		response, err := rolesApiService.RoleEnable(mgmtCtx, systemScopedRole.Id, api.RoleEnabledDisabledFields{
			ScopeType: systemScopeType,
		})
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusNoContent))

		systemScopedRole.Enabled = ptr.To(true)
		systemScopedRole.EffectiveEnabled = true
	})

	It("Post tenant role with permission set and KubernetesPermissions", func() {
		if !config.Get[*authorization_config.General]().UsePermissionSets {
			Skip("no permission sets")
		}
		body := api.RoleCreationFieldsInternal{
			Name:        "tenant role ps kubernetes permissions",
			Description: "first tenant test role, can fully manage projects",
			PermissionSets: []api.RolePermissionSet{
				{Name: static_permission_sets.WorkloadRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.WorkloadRead.Name].String()},
				{Name: static_permission_sets.YamlWorkloadEdit.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.YamlWorkloadEdit.Name].String()},
				{Name: static_permission_sets.ApplicationEdit.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.ApplicationEdit.Name].String()},
				{Name: static_permission_sets.WorkloadPropertiesRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.WorkloadPropertiesRead.Name].String()},
				{Name: static_permission_sets.TemplatesRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.TemplatesRead.Name].String()},
				{Name: static_permission_sets.DataStorageRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.DataStorageRead.Name].String()},
				{Name: static_permission_sets.ClusterRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.ClusterRead.Name].String()},
				{Name: static_permission_sets.NetworkTopologyRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.NetworkTopologyRead.Name].String()},
				{Name: static_permission_sets.NodePoolsRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.NodePoolsRead.Name].String()},
				{Name: static_permission_sets.NodesRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.NodesRead.Name].String()},
				{Name: static_permission_sets.PoliciesRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.PoliciesRead.Name].String()},
				{Name: static_permission_sets.DepartmentRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.DepartmentRead.Name].String()},
				{Name: static_permission_sets.ProjectRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.ProjectRead.Name].String()},
				{Name: static_permission_sets.AccountRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.AccountRead.Name].String()},
				{Name: static_permission_sets.SettingsRead.Name, Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.SettingsRead.Name].String()},
				{Name: "", Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.BrandingSettingsRead.Name].String()},
				{Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.SecuritySettingsRead.Name].String()},
				{Id: dao.MapStaticPermissionSetNameToId[static_permission_sets.NotificationSettingsRead.Name].String()},
			},
			ScopeType: "tenant",
			ScopeId:   ptr.To(tenantId),
			Enabled:   ptr.To(true),
			KubernetesPermissions: api.KubernetesPermissions{
				ptr.To(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_roles.Viewer.Name])),
			},
		}

		role := createTestRole(ctx, rolesApiService, body, &createdRoles, false, http.StatusCreated)
		createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusConflict)
		Expect(role).ToNot(BeNil())
		Expect(role.Id).ToNot(BeNil())
		Expect(role.Name).To(Equal(body.Name))
		Expect(role.Description).To(Equal(body.Description))
		Expect(role.ScopeType).To(Equal(body.ScopeType))
		Expect(role.ScopeId).ToNot(BeNil())
		Expect(role.ScopeId).ToNot(Equal(tenantId))
		Expect(role.Enabled).To(Equal(body.Enabled))
		Expect(role.Custom).To(Equal(true))
		Expect(role.KubernetesPermissions.PredefinedRole).To(Not(BeNil()))
		Expect(*role.KubernetesPermissions.PredefinedRole).To(Equal(*body.KubernetesPermissions.PredefinedRole))
		Expect(len(role.Permissions)).To(Equal(len(api.AllowedResourceTypeEnumValues)))
	})

	Context("LegacyRoles", func() {
		It("Get V1", func() {
			getRolesResponse, err := rolesApiService.GetRolesV1(ctx, false)
			Expect(err).To(BeNil())
			Expect(getRolesResponse.Code).To(Equal(http.StatusOK))

			roles := httpResponseBodyToRoles(getRolesResponse)

			var createdTenantRoleFromServer api.Role
			var createdSystemRoleFromServer api.Role
			var internalRoleRetrieved bool = false
			for _, role := range roles {
				if role.Name == tenantScopedRole.Name {
					createdTenantRoleFromServer = role
					continue
				}
				if role.Name == systemScopedRole.Name {
					createdSystemRoleFromServer = role
					continue
				}
				if role.Name == static_roles.IntegrationMetricsCollector.Name {
					internalRoleRetrieved = true
				}
			}

			Expect(internalRoleRetrieved).To(BeFalse())
			validateRole(createdTenantRoleFromServer, tenantScopedRole)
			validateRole(createdSystemRoleFromServer, systemScopedRole)
		})

		It("Get V1 include internal", func() {
			getRolesResponse, err := rolesApiService.GetRolesV1(ctx, true)
			Expect(err).To(BeNil())
			Expect(getRolesResponse.Code).To(Equal(http.StatusOK))

			roles := httpResponseBodyToRoles(getRolesResponse)

			var createdTenantRoleFromServer api.Role
			var createdSystemRoleFromServer api.Role
			var internalRoleRetrieved bool = false
			for _, role := range roles {
				if role.Name == tenantScopedRole.Name {
					createdTenantRoleFromServer = role
					continue
				}
				if role.Name == systemScopedRole.Name {
					createdSystemRoleFromServer = role
					continue
				}
				if role.Name == static_roles.IntegrationMetricsCollector.Name {
					internalRoleRetrieved = true
				}
			}

			Expect(internalRoleRetrieved).To(BeTrue())
			validateRole(createdTenantRoleFromServer, tenantScopedRole)
			validateRole(createdSystemRoleFromServer, systemScopedRole)
		})

		It("Get V2", func() {
			getRolesResponse, err := rolesApiService.GetRoles(ctx, 0, 0, "", "", []string{}, "", false)
			Expect(err).To(BeNil())
			Expect(getRolesResponse.Code).To(Equal(http.StatusOK))

			roles := getRolesResponse.Body.(api.GetRoles200Response).Roles

			var createdTenantRoleFromServer api.Role
			var createdSystemRoleFromServer api.Role
			internalRoleRetrieved := false
			for _, role := range roles {
				if role.Name == tenantScopedRole.Name {
					createdTenantRoleFromServer = role
					continue
				}
				if role.Name == systemScopedRole.Name {
					createdSystemRoleFromServer = role
					continue
				}
				if role.Name == static_roles.IntegrationMetricsCollector.Name {
					internalRoleRetrieved = true
				}
			}

			Expect(internalRoleRetrieved).To(BeFalse())
			validateRole(createdTenantRoleFromServer, tenantScopedRole)
			validateRole(createdSystemRoleFromServer, systemScopedRole)
		})

		It("Get V2 include internal", func() {
			getRolesResponse, err := rolesApiService.GetRoles(ctx, 0, 0, "", "", []string{}, "", true)
			Expect(err).To(BeNil())
			Expect(getRolesResponse.Code).To(Equal(http.StatusOK))

			roles := getRolesResponse.Body.(api.GetRoles200Response).Roles
			internalRoleRetrieved := false
			var createdTenantRoleFromServer api.Role
			var createdSystemRoleFromServer api.Role
			for _, role := range roles {
				if isLegacyRole(role.Name) {
					Expect(role.Deprecated).To(BeTrue())
				} else {
					Expect(role.Deprecated).To(BeFalse())
				}
				if role.Name == tenantScopedRole.Name {
					createdTenantRoleFromServer = role
					continue
				}
				if role.Name == systemScopedRole.Name {
					createdSystemRoleFromServer = role
					continue
				}
				if role.Name == static_roles.IntegrationMetricsCollector.Name {
					internalRoleRetrieved = true
				}
			}
			Expect(internalRoleRetrieved).To(BeTrue())
			validateRole(createdTenantRoleFromServer, tenantScopedRole)
			validateRole(createdSystemRoleFromServer, systemScopedRole)
		})

		It("Get V2 search", func() {
			getRolesResponse, err := rolesApiService.GetRoles(ctx, 0, 0, "", "", []string{}, "tenant", false)
			Expect(err).To(BeNil())
			Expect(getRolesResponse.Code).To(Equal(http.StatusOK))

			roles := getRolesResponse.Body.(api.GetRoles200Response).Roles

			expectedRoles := 4
			if !config.Get[*authorization_config.General]().UsePermissionSets {
				expectedRoles = 2
			}
			Expect(len(roles)).To(Equal(expectedRoles))
			for i := range roles {
				Expect(roles[i].Name[:11]).To(Equal("tenant role"))
			}
		})

		It("Get V2 search", func() {
			getRolesResponse, err := rolesApiService.GetRoles(ctx, 0, 0, "", "", []string{}, "role", false)
			Expect(err).To(BeNil())
			Expect(getRolesResponse.Code).To(Equal(http.StatusOK))

			roles := getRolesResponse.Body.(api.GetRoles200Response).Roles

			expectedRoles := 7
			if !config.Get[*authorization_config.General]().UsePermissionSets {
				expectedRoles = 4
			}
			Expect(len(roles)).To(Equal(expectedRoles))
			for i := range roles {
				Expect(roles[i].Name[7:11]).To(Equal("role"))
			}
		})

		It("Upsert", func() {
			role := models.Role{
				Name: "test",
				Permissions: []*models.Permission{
					{
						ResourceType: "some-resource-type",
					},
					{
						ResourceType: api.RESOURCETYPE_TENANT,
						Read:         true,
					},
				},
				Description: "test",
				ScopeType:   "system",
				ScopeId:     nil,
				Custom:      true,
				Enabled:     true,
				TenantId:    ptr.To(tenantId),
			}
			created, err := dao.DbCreateRole(ctx, "", role)
			Expect(err).To(BeNil())

			roleForUpsert := models.Role{
				Name:        "test",
				Description: "test",
				ScopeType:   "system",
				Enabled:     true,
				Permissions: []*models.Permission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Read:         true,
					},
				},
			}

			initializer := dao.RoleDaoInitializer{}
			updateId, upsertErr := initializer.DbUpsertLegacyRole(ctx, roleForUpsert)
			Expect(upsertErr).To(BeNil())
			Expect(created.ID).To(Equal(updateId))

			dbRole, apiErr := dao.DbGetRoleById(ctx, tenantId, updateId, nil)
			Expect(apiErr).To(BeNil())
			for _, permission := range dbRole.Permissions {
				Expect(permission.Create).To(Equal(false))
				Expect(permission.Update).To(Equal(false))
				Expect(permission.Delete).To(Equal(false))
				Expect(permission.Sync).To(Equal(false))
				Expect(permission.Read).To(Equal(permission.ResourceType == api.RESOURCETYPE_PROJECT))
			}

			apiErr = dao.DbDeleteRole(ctx, tenantId, dbRole.ID)
			Expect(apiErr).To(BeNil())
		})

		It("Post should fail, multiple definitions for the same resource", func() {
			body := api.RoleCreationFieldsInternal{
				Name:        "test multiple definitions for the same resource",
				Description: "test multiple definitions for the same resource",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_READ,
						},
					},
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_CREATE,
						},
					},
				},
			}

			createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusBadRequest)
		})

		It("Post should fail, name already exist", func() {
			body := api.RoleCreationFieldsInternal{
				Name:        systemScopedRole.Name,
				Description: "test name uniqueness",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_CREATE,
						},
					},
				},
			}

			createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusConflict)
			body.Name = tenantScopedRole.Name
			body.ScopeType = "system"
			createTestRole(mgmtCtx, rolesApiService, body, &createdRoles, true, http.StatusConflict)
		})

		It("Post should fail, no permissions", func() {
			body := api.RoleCreationFieldsInternal{
				Name:        "empty permissions test",
				Description: "empty permissions test",
				Permissions: []api.PostPermission{},
			}

			createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusBadRequest)
		})

		It("Post should fail, no actions", func() {
			body := api.RoleCreationFieldsInternal{
				Name:        "empty actions test",
				Description: "empty actions test",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions:      []api.Action{},
					},
				},
			}

			createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusBadRequest)
		})

		It("Post should fail, kubernetes role privilege escalation - succeed", func() {
			permissions := []api.PostPermission{}
			l1ResearcherPermissions := static_legacy_roles.ResearcherL1.Permissions
			for _, l1ResearcherPermission := range l1ResearcherPermissions {
				if l1ResearcherPermission.ResourceType == api.RESOURCETYPE_CLUSTER {
					continue
				}
				actions := []api.Action{}
				if l1ResearcherPermission.Read {
					actions = append(actions, api.ACTION_READ)
				}
				if l1ResearcherPermission.Create {
					actions = append(actions, api.ACTION_CREATE)
				}
				if l1ResearcherPermission.Update {
					actions = append(actions, api.ACTION_UPDATE)
				}
				if l1ResearcherPermission.Delete {
					actions = append(actions, api.ACTION_DELETE)
				}
				permissions = append(permissions, api.PostPermission{
					ResourceType: l1ResearcherPermission.ResourceType,
					Actions:      actions,
				})
			}
			permissions = append(permissions, api.PostPermission{
				ResourceType: api.RESOURCETYPE_CLUSTER,
				Actions:      []api.Action{api.ACTION_CREATE, api.ACTION_UPDATE, api.ACTION_DELETE, api.ACTION_READ},
			})

			l1ResearcherId := int32(dao.MapStaticRoleNameToRoleId[static_legacy_roles.ResearcherL1.Name])
			body := api.RoleCreationFieldsInternal{
				Name:                  "empty actions test",
				Description:           "empty actions test",
				Permissions:           permissions,
				KubernetesPermissions: api.KubernetesPermissions{PredefinedRole: ptr.To(fmt.Sprintf("%d", l1ResearcherId))},
			}

			createTestRole(ctx, rolesApiService, body, &createdRoles, false, http.StatusCreated)
		})

		It("Post should fail, kubernetes role privilege escalation - fail", func() {
			permissions := []api.PostPermission{}
			resRoleId, ok := dao.MapStaticRoleNameToRoleId[static_legacy_roles.ResearcherL1.Name]
			Expect(ok).To(BeTrue())
			roleRes, _ := rolesApiService.GetRole(ctx, int32(resRoleId))
			Expect(roleRes.Code).To(Equal(http.StatusOK))
			l1ResearcherPermissions := httpResponseBodyToRole(roleRes).Permissions
			for _, l1ResearcherPermission := range l1ResearcherPermissions {
				if l1ResearcherPermission.ResourceType == api.RESOURCETYPE_PROJECT {
					continue
				}
				actions := []api.Action{}
				for _, action := range l1ResearcherPermission.Actions {
					actions = append(actions, api.Action(action))
				}

				permissions = append(permissions, api.PostPermission{
					ResourceType: l1ResearcherPermission.ResourceType,
					Actions:      actions,
				})
			}

			l1ResearcherId := int32(dao.MapStaticRoleNameToRoleId[static_legacy_roles.ResearcherL1.Name])
			body := api.RoleCreationFieldsInternal{
				Name:                  "empty actions test",
				Description:           "empty actions test",
				Permissions:           permissions,
				KubernetesPermissions: api.KubernetesPermissions{PredefinedRole: ptr.To(fmt.Sprintf("%d", l1ResearcherId))},
			}

			createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusBadRequest)
		})

		Describe("Post should fail, cluster version does not support predefined roles", func() {
			BeforeAll(func() {
				testing.ClusterVersion = "2.20.0"
			})
			AfterAll(func() {
				testing.ClusterVersion = "2.21.0"
			})

			It("Post should fail, cluster version does not support predefined roles", func() {
				permissions := []api.PostPermission{}
				editorRoleId, ok := dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name]
				Expect(ok).To(BeTrue())
				editorRes, _ := rolesApiService.GetRole(ctx, int32(editorRoleId))
				Expect(editorRes.Code).To(Equal(http.StatusOK))
				editorPermissions := httpResponseBodyToRole(editorRes).Permissions
				permissionMap := map[api.ResourceType][]api.Action{}
				for _, editorPermission := range editorPermissions {
					if editorPermission.ResourceType == api.RESOURCETYPE_PROJECT {
						continue
					}
					actions := []api.Action{}
					for _, action := range editorPermission.Actions {
						actions = append(actions, api.Action(action))
					}
					permissionMap[editorPermission.ResourceType] = actions
					permissions = append(permissions, api.PostPermission{
						ResourceType: editorPermission.ResourceType,
						Actions:      actions,
					})
				}
				permissions = append(permissions, api.PostPermission{
					ResourceType: api.RESOURCETYPE_PROJECT,
					Actions: []api.Action{
						api.ACTION_CREATE,
						api.ACTION_READ,
						api.ACTION_UPDATE,
						api.ACTION_DELETE,
					},
				})
				permissionMap[api.RESOURCETYPE_PROJECT] = []api.Action{
					api.ACTION_CREATE,
					api.ACTION_READ,
					api.ACTION_UPDATE,
					api.ACTION_DELETE,
				}
				body := api.RoleCreationFieldsInternal{
					Name:        "tenant role kubernetes permissions",
					Description: "first tenant test role, can fully manage projects",
					Permissions: permissions,
					ScopeType:   "tenant",
					ScopeId:     ptr.To(tenantId),
					Enabled:     ptr.To(true),
					KubernetesPermissions: api.KubernetesPermissions{
						ptr.To(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name])),
					},
				}
				createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusBadRequest)
			})
		})

		Describe("Post should succeed, cluster version empty", func() {
			BeforeAll(func() {
				testing.ClusterVersion = ""
			})
			AfterAll(func() {
				testing.ClusterVersion = "2.21.0"
			})

			It("Post should succeed, cluster version empty", func() {
				permissions := []api.PostPermission{}
				roleId, ok := dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name]
				Expect(ok).To(BeTrue())
				roleRes, _ := rolesApiService.GetRole(ctx, int32(roleId))
				Expect(roleRes.Code).To(Equal(http.StatusOK))
				editorPermissions := httpResponseBodyToRole(roleRes).Permissions
				permissionMap := map[api.ResourceType][]api.Action{}
				for _, editorPermission := range editorPermissions {
					if editorPermission.ResourceType == api.RESOURCETYPE_PROJECT {
						continue
					}
					actions := []api.Action{}
					for _, action := range editorPermission.Actions {
						actions = append(actions, api.Action(action))
					}
					permissionMap[editorPermission.ResourceType] = actions
					permissions = append(permissions, api.PostPermission{
						ResourceType: editorPermission.ResourceType,
						Actions:      actions,
					})
				}
				permissions = append(permissions, api.PostPermission{
					ResourceType: api.RESOURCETYPE_PROJECT,
					Actions: []api.Action{
						api.ACTION_CREATE,
						api.ACTION_READ,
						api.ACTION_UPDATE,
						api.ACTION_DELETE,
					},
				})
				permissionMap[api.RESOURCETYPE_PROJECT] = []api.Action{
					api.ACTION_CREATE,
					api.ACTION_READ,
					api.ACTION_UPDATE,
					api.ACTION_DELETE,
				}
				body := api.RoleCreationFieldsInternal{
					Name:        "tenant role kubernetes permissions cluster empty",
					Description: "first tenant test role, can fully manage projects",
					Permissions: permissions,
					ScopeType:   "tenant",
					ScopeId:     ptr.To(tenantId),
					Enabled:     ptr.To(true),
					KubernetesPermissions: api.KubernetesPermissions{
						ptr.To(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name])),
					},
				}
				createTestRole(ctx, rolesApiService, body, &createdRoles, false, http.StatusCreated)
			})
		})
	})

	Context("LegacyRoles by ID", func() {
		It("Get V1", func() {
			getRoleResponse, err := rolesApiService.GetRoleV1(ctx, tenantScopedRole.Id)
			Expect(err).To(BeNil())
			Expect(getRoleResponse.Code).To(Equal(http.StatusOK))

			role := httpResponseBodyToRole(getRoleResponse)
			validateRole(role, tenantScopedRole)
		})

		It("Get V2", func() {
			getRoleResponse, err := rolesApiService.GetRole(ctx, tenantScopedRole.Id)
			Expect(err).To(BeNil())
			Expect(getRoleResponse.Code).To(Equal(http.StatusOK))

			role := httpResponseBodyToRole(getRoleResponse)
			validateRole(role, tenantScopedRole)
		})

		It("Get should fail, not found", func() {
			getRoleResponse, err := rolesApiService.GetRoleV1(ctx, math.MaxInt32)
			Expect(err).To(BeNil())
			Expect(getRoleResponse.Code).To(Equal(http.StatusNotFound))

			getRoleResponse, err = rolesApiService.GetRole(ctx, math.MaxInt32)
			Expect(err).To(BeNil())
			Expect(getRoleResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("Tenant level disable system role", func() {
			testRoleActive(ctx, rolesApiService, systemScopedRole, api.RoleActiveFields{
				ScopeType: systemScopeType,
				Enabled:   false,
			}, http.StatusForbidden, false, false)

			testRoleActive(ctx, rolesApiService, systemScopedRole, api.RoleActiveFields{
				Enabled: false,
			}, http.StatusNoContent, true, false)
		})

		It("Tenant level enable system role", func() {
			testRoleActive(ctx, rolesApiService, systemScopedRole, api.RoleActiveFields{
				ScopeType: systemScopeType,
				Enabled:   true,
			}, http.StatusForbidden, false, false)

			testRoleActive(ctx, rolesApiService, systemScopedRole, api.RoleActiveFields{
				Enabled: true,
			}, http.StatusNoContent, true, true)
		})

		It("Tenant level disable tenant role", func() {
			testRoleActive(ctx, rolesApiService, tenantScopedRole, api.RoleActiveFields{
				ScopeType: systemScopeType,
				Enabled:   false,
			}, http.StatusForbidden, false, false)

			testRoleActive(ctx, rolesApiService, tenantScopedRole, api.RoleActiveFields{
				Enabled: false,
			}, http.StatusNoContent, false, false)
		})

		It("Tenant level enable tenant role", func() {
			testRoleActive(ctx, rolesApiService, tenantScopedRole, api.RoleActiveFields{
				ScopeType: systemScopeType,
				Enabled:   true,
			}, http.StatusForbidden, false, false)

			testRoleActive(ctx, rolesApiService, tenantScopedRole, api.RoleActiveFields{
				Enabled: true,
			}, http.StatusNoContent, true, true)
		})

		It("System level disable system role", func() {
			testRoleActive(mgmtCtx, rolesApiService, systemScopedRole, api.RoleActiveFields{
				ScopeType: systemScopeType,
				Enabled:   false,
			}, http.StatusNoContent, false, false)

			testRoleActive(mgmtCtx, rolesApiService, systemScopedRole, api.RoleActiveFields{
				ScopeType: systemScopeType,
				Enabled:   true,
			}, http.StatusNoContent, true, true)

			testRoleActive(mgmtCtx, rolesApiService, systemScopedRole, api.RoleActiveFields{
				Enabled: false,
			}, http.StatusNoContent, true, false)
		})

		It("System level enable system role", func() {
			testRoleActive(mgmtCtx, rolesApiService, systemScopedRole, api.RoleActiveFields{
				Enabled: true,
			}, http.StatusNoContent, true, true)
		})

		It("System level disable tenant role", func() {
			testRoleActive(mgmtCtx, rolesApiService, tenantScopedRole, api.RoleActiveFields{
				ScopeType: systemScopeType,
				Enabled:   false,
			}, http.StatusNotFound, false, false)

			testRoleActive(mgmtCtx, rolesApiService, tenantScopedRole, api.RoleActiveFields{
				Enabled: false,
			}, http.StatusNotFound, false, false)
		})

		It("System level enable tenant role", func() {
			testRoleActive(mgmtCtx, rolesApiService, tenantScopedRole, api.RoleActiveFields{
				ScopeType: systemScopeType,
				Enabled:   true,
			}, http.StatusNotFound, false, false)

			testRoleActive(mgmtCtx, rolesApiService, tenantScopedRole, api.RoleActiveFields{
				Enabled: true,
			}, http.StatusNotFound, false, false)
		})

		It("Update tenant role", func() {
			body := api.RoleCreationFieldsInternal{
				Name:        "tenant role",
				Description: "first tenant test role, can fully manage projects",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_DEPARTMENT,
						Actions: []api.Action{
							api.ACTION_CREATE,
							api.ACTION_READ,
							api.ACTION_UPDATE,
							api.ACTION_DELETE,
						},
					},
				},
				Enabled:               ptr.To(true),
				KubernetesPermissions: api.KubernetesPermissions{},
			}

			updateResponse, err := rolesApiService.UpdateRole(ctx, tenantScopedRole.Id, body)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))

			getRoleResponse, err := rolesApiService.GetRole(ctx, tenantScopedRole.Id)
			role := httpResponseBodyToRole(getRoleResponse)

			tenantScopedRole.KubernetesPermissions = api.KubernetesPermissions{}
			for i := range tenantScopedRole.Permissions {
				if tenantScopedRole.Permissions[i].ResourceType == api.RESOURCETYPE_DEPARTMENT {
					tenantScopedRole.Permissions[i].ResourceType = api.RESOURCETYPE_PROJECT
					tenantScopedRole.Permissions[i].DisplayName = ResourceTypesDisplayName[api.RESOURCETYPE_PROJECT]
					continue
				}
				if tenantScopedRole.Permissions[i].ResourceType == api.RESOURCETYPE_PROJECT {
					tenantScopedRole.Permissions[i].ResourceType = api.RESOURCETYPE_DEPARTMENT
					tenantScopedRole.Permissions[i].DisplayName = ResourceTypesDisplayName[api.RESOURCETYPE_DEPARTMENT]
					continue
				}
			}
			validateRole(role, tenantScopedRole)
		})

		It("Update tenant role PS", func() {
			if !config.Get[*authorization_config.General]().UsePermissionSets {
				Skip("no permission sets")
			}
			body := api.RoleCreationFieldsInternal{
				Name:        tenantScopedRolePS.Name,
				Description: tenantScopedRolePS.Description,
				PermissionSets: []api.RolePermissionSet{{
					Name: static_permission_sets.WorkloadPropertiesEdit.Name,
					Id:   dao.MapStaticPermissionSetNameToId[static_permission_sets.WorkloadPropertiesEdit.Name].String(),
				}},
				Enabled:               ptr.To(true),
				KubernetesPermissions: api.KubernetesPermissions{},
			}

			updateResponse, err := rolesApiService.UpdateRole(ctx, tenantScopedRolePS.Id, body)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))

			role := updateResponse.Body.(api.Role)

			tenantScopedRolePS.KubernetesPermissions = api.KubernetesPermissions{}
			for _, permission := range role.Permissions {
				switch permission.ResourceType {
				case api.RESOURCETYPE_WORKLOAD_PROPERTIES:
					Expect(len(permission.Actions)).To(Equal(4))
				case api.RESOURCETYPE_CLUSTERS_MINIMAL:
					Expect(len(permission.Actions)).To(Equal(1))
				default:
					Expect(len(permission.Actions)).To(Equal(0))
				}
			}
			tenantScopedRolePS.KubernetesPermissions = api.KubernetesPermissions{}
			for i := range tenantScopedRolePS.Permissions {
				if tenantScopedRolePS.Permissions[i].ResourceType == api.RESOURCETYPE_WORKLOAD_PROPERTIES {
					tenantScopedRolePS.Permissions[i].Actions = []api.PermissionAction{
						api.PERMISSIONACTION_CREATE, api.PERMISSIONACTION_READ, api.PERMISSIONACTION_UPDATE, api.PERMISSIONACTION_DELETE,
					}
				} else if tenantScopedRolePS.Permissions[i].ResourceType == api.RESOURCETYPE_CLUSTERS_MINIMAL {
					tenantScopedRolePS.Permissions[i].Actions = []api.PermissionAction{
						api.PERMISSIONACTION_READ,
					}
				} else {
					// Clear all other permissions that were in WorkloadRead but not in WorkloadPropertiesEdit
					tenantScopedRolePS.Permissions[i].Actions = []api.PermissionAction{}
				}
			}
			validateRole(role, tenantScopedRolePS)
		})

		It("Update tenant permission set role with permissions", func() {
			if !config.Get[*authorization_config.General]().UsePermissionSets {
				Skip("no permission sets")
			}
			body := api.RoleCreationFieldsInternal{
				Name:        "tenant role ps",
				Description: "first tenant test role, can fully manage projects",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_DEPARTMENT,
						Actions: []api.Action{
							api.ACTION_CREATE,
							api.ACTION_READ,
							api.ACTION_UPDATE,
							api.ACTION_DELETE,
						},
					},
				},
				Enabled:               ptr.To(true),
				KubernetesPermissions: api.KubernetesPermissions{},
			}

			updateResponse, err := rolesApiService.UpdateRole(ctx, tenantScopedRolePS.Id, body)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("Update should fail, system role", func() {
			body := api.RoleCreationFieldsInternal{
				Name:        "tenant role",
				Description: "first tenant test role, can fully manage projects",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_DEPARTMENT,
						Actions: []api.Action{
							api.ACTION_CREATE,
							api.ACTION_READ,
							api.ACTION_UPDATE,
							api.ACTION_DELETE,
						},
					},
				},
				KubernetesPermissions: api.KubernetesPermissions{},
			}

			updateResponse, err := rolesApiService.UpdateRole(ctx, systemScopedRole.Id, body)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusForbidden))
		})

		It("Update should fail, change scopeType", func() {
			body := api.RoleCreationFieldsInternal{
				Name:        "tenant role",
				Description: "first tenant test role, can fully manage projects",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_DEPARTMENT,
						Actions: []api.Action{
							api.ACTION_CREATE,
							api.ACTION_READ,
							api.ACTION_UPDATE,
							api.ACTION_DELETE,
						},
					},
				},
				ScopeType:             tenantScopeType,
				ScopeId:               ptr.To(tenantId),
				KubernetesPermissions: api.KubernetesPermissions{},
			}

			updateResponse, err := rolesApiService.UpdateRole(mgmtCtx, systemScopedRole.Id, body)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("Update should fail, pre-defined role", func() {
			response, err := rolesApiService.GetRoles(ctx, 0, 0, "", "", []string{}, "", false)
			Expect(err).To(BeNil())
			roles := response.Body.(api.GetRoles200Response).Roles
			var roleToUpdate api.Role
			for i := range roles {
				if roles[i].Name == "L1 researcher" {
					roleToUpdate = roles[i]
				}
			}

			body := api.RoleCreationFieldsInternal{
				Name:        roleToUpdate.Name,
				Description: roleToUpdate.Description,
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_DEPARTMENT,
						Actions: []api.Action{
							api.ACTION_CREATE,
							api.ACTION_READ,
							api.ACTION_UPDATE,
							api.ACTION_DELETE,
						},
					},
				},
				ScopeType:             roleToUpdate.ScopeType,
				ScopeId:               roleToUpdate.ScopeId,
				Enabled:               roleToUpdate.Enabled,
				KubernetesPermissions: roleToUpdate.KubernetesPermissions,
			}

			updateResponse, err := rolesApiService.UpdateRole(ctx, roleToUpdate.Id, body)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusForbidden))
		})

		It("Update should fail, pre-defined role PS", func() {
			if !config.Get[*authorization_config.General]().UsePermissionSets {
				Skip("no permission sets")
			}
			body := api.RoleCreationFieldsInternal{
				Name:                  tenantScopedRolePS.Name,
				Description:           tenantScopedRolePS.Description,
				PermissionSets:        tenantScopedRolePS.PermissionSets,
				ScopeType:             tenantScopedRolePS.ScopeType,
				ScopeId:               tenantScopedRolePS.ScopeId,
				Enabled:               tenantScopedRolePS.Enabled,
				KubernetesPermissions: api.KubernetesPermissions{PredefinedRole: ptr.To(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_legacy_roles.ResearcherL1.Name]))},
			}

			updateResponse, err := rolesApiService.UpdateRole(ctx, tenantScopedRolePS.Id, body)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("Update system role, role in use", func() {
			accessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     tenantId,
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      systemScopedRole.Id,
			}
			createdAccessRuleResponse, err := accessRulesApiService.CreateAccessRule(ctx, accessRuleBody)
			Expect(err).To(BeNil())
			accessRule := HttpResponseBodyToAccessRule(createdAccessRuleResponse)
			createdAccessRules = append(createdAccessRules, accessRule)

			body := api.RoleCreationFieldsInternal{
				Name:        "system role",
				Description: "first system test role, can fully manage projects",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_DEPARTMENT,
						Actions: []api.Action{
							api.ACTION_CREATE,
							api.ACTION_READ,
							api.ACTION_UPDATE,
							api.ACTION_DELETE,
						},
					},
				},
				ScopeType:             "system",
				KubernetesPermissions: api.KubernetesPermissions{},
			}

			updateResponse, err := rolesApiService.UpdateRole(mgmtCtx, systemScopedRole.Id, body)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))

			_, err = accessRulesApiService.DeleteAccessRule(ctx, accessRule.Id)
			Expect(err).To(BeNil())

			updateResponse, err = rolesApiService.UpdateRole(mgmtCtx, systemScopedRole.Id, body)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))

			getRoleResponse, err := rolesApiService.GetRole(mgmtCtx, systemScopedRole.Id)
			role := httpResponseBodyToRole(getRoleResponse)
			for _, permissions := range role.Permissions {
				if permissions.ResourceType == api.RESOURCETYPE_DEPARTMENT {
					Expect(len(permissions.Actions)).To(Equal(4))
				} else {
					Expect(len(permissions.Actions)).To(Equal(0))
				}
			}

		})

		It("Delete tenant role", func() {
			deletionResponse, err := rolesApiService.DeleteRole(ctx, tenantScopedRole.Id)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			createdRoles = createdRoles[1:]
		})

		It("Delete should fail, system role", func() {
			deletionResponse, err := rolesApiService.DeleteRole(ctx, systemScopedRole.Id)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusForbidden))
		})

		It("Delete should fail, predefined role", func() {
			response, err := rolesApiService.GetRoles(ctx, 0, 0, "", "", []string{}, "", false)
			Expect(err).To(BeNil())
			roles := response.Body.(api.GetRoles200Response).Roles
			var roleToDelete api.Role
			for i := range roles {
				if roles[i].Name == "L1 researcher" {
					roleToDelete = roles[i]
				}
			}
			deletionResponse, err := rolesApiService.DeleteRole(ctx, roleToDelete.Id)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusForbidden))
		})

		It("Delete system role, role in use", func() {
			localTenantSettings := NewMockTenantSettingsService(map[string]bool{})
			accessRulesApiService := NewAccessRulesApiService(localTenantSettings)

			accessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     tenantId,
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      systemScopedRole.Id,
			}
			createdAccessRuleResponse, err := accessRulesApiService.CreateAccessRule(ctx, accessRuleBody)
			Expect(err).To(BeNil())
			accessRule := HttpResponseBodyToAccessRule(createdAccessRuleResponse)
			createdAccessRules = append(createdAccessRules, accessRule)

			deletionResponse, err := rolesApiService.DeleteRole(mgmtCtx, systemScopedRole.Id)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusBadRequest))

			_, err = accessRulesApiService.DeleteAccessRule(ctx, accessRule.Id)
			Expect(err).To(BeNil())

			deletionResponse, err = rolesApiService.DeleteRole(mgmtCtx, systemScopedRole.Id)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))
		})

		It("Delete should fail, not found", func() {
			deletionResponse, err := rolesApiService.DeleteRole(ctx, math.MaxInt32)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("Custom roles disabled", func() {
		BeforeAll(func() {
			testing.CustomRolesEnabled = false
		})
		AfterAll(func() {
			testing.CustomRolesEnabled = true
		})
		It("Post role", func() {
			body := api.RoleCreationFieldsInternal{
				Name:        "tenant role",
				Description: "first tenant test role, can fully manage projects",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_CREATE,
							api.ACTION_READ,
							api.ACTION_UPDATE,
							api.ACTION_DELETE,
						},
					},
				},
				ScopeType: "tenant",
				ScopeId:   ptr.To(tenantId),
				Enabled:   ptr.To(true),
				KubernetesPermissions: api.KubernetesPermissions{
					ptr.To(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name])),
				},
			}
			createTestRole(ctx, rolesApiService, body, &createdRoles, true, http.StatusBadRequest)
		})
	})

	Context("Predefined roles", func() {
		AfterAll(func() {
			editorRoleId := dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name]
			getRoleResponse, err := rolesApiService.GetRole(ctx, int32(editorRoleId))
			Expect(err).To(BeNil())
			role := getRoleResponse.Body.(api.Role)

			testRoleActive(ctx, rolesApiService, role, api.RoleActiveFields{
				ScopeType: tenantScopeType,
				Enabled:   true,
			}, http.StatusNoContent, true, true)

			viewerRoleId := dao.MapStaticRoleNameToRoleId[static_roles.Viewer.Name]
			getRole2Response, err := rolesApiService.GetRole(ctx, int32(viewerRoleId))
			Expect(err).To(BeNil())
			role2 := getRole2Response.Body.(api.Role)

			testRoleActive(ctx, rolesApiService, role2, api.RoleActiveFields{
				ScopeType: tenantScopeType,
				Enabled:   true,
			}, http.StatusNoContent, true, true)
		})

		It("Tenant level disable predefined role", func() {
			editorRoleId := dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name]
			getRoleResponse, err := rolesApiService.GetRole(ctx, int32(editorRoleId))
			Expect(err).To(BeNil())
			Expect(getRoleResponse.Code).To(Equal(http.StatusOK))
			role := getRoleResponse.Body.(api.Role)

			testRoleActive(ctx, rolesApiService, role, api.RoleActiveFields{
				ScopeType: tenantScopeType,
				Enabled:   false,
			}, http.StatusNoContent, true, false)

			viewerRoleId := dao.MapStaticRoleNameToRoleId[static_roles.Viewer.Name]
			getRole2Response, err := rolesApiService.GetRole(ctx, int32(viewerRoleId))
			Expect(err).To(BeNil())
			Expect(getRole2Response.Code).To(Equal(http.StatusOK))
			role2 := getRole2Response.Body.(api.Role)

			testRoleActive(ctx, rolesApiService, role2, api.RoleActiveFields{
				ScopeType: tenantScopeType,
				Enabled:   false,
			}, http.StatusNoContent, true, false)
		})
	})

	Context("Role name scoped uniqueness constraints", func() {
		tenant1Id := fmt.Sprintf("%d", testing.TestTenant)
		tenant2Id := fmt.Sprintf("%d", testing.ManagementTestTenant)

		ctx1 := context.WithValue(setContext(tenant1Id, TestUserId, []string{dao.SuperAdmins}), authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
			Tenants: []string{tenant1Id},
		})

		ctx2 := context.WithValue(setContext(tenant2Id, TestUserId, []string{dao.SuperAdmins}), authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
			Tenants: []string{tenant2Id},
		})

		mgmtCtxForConstraints := context.WithValue(context.WithValue(setContext(tenant2Id, TestUserId, []string{dao.SuperAdmins}), runaictx.CfIsManagementTenant, "true"), authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
			System:  ptr.To(true),
			Tenants: []string{tenant2Id},
		})

		var createdConstraintTestRoles []api.Role

		AfterAll(func() {
			// Clean up all roles created during constraint tests
			for _, role := range createdConstraintTestRoles {
				if role.ScopeType == systemScopeType {
					rolesApiService.DeleteRole(mgmtCtxForConstraints, role.Id)
				} else {
					if role.ScopeId != nil && *role.ScopeId == tenant1Id {
						rolesApiService.DeleteRole(ctx1, role.Id)
					} else {
						rolesApiService.DeleteRole(ctx2, role.Id)
					}
				}
			}
		})

		It("Two tenants can add same role name", func() {
			roleName := "shared-tenant-role"

			// Create role in tenant 1
			body1 := api.RoleCreationFieldsInternal{
				Name:        roleName,
				Description: "role for tenant 1",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_READ,
						},
					},
				},
				ScopeType: tenantScopeType,
				ScopeId:   ptr.To(tenant1Id),
				Enabled:   ptr.To(true),
			}

			role1 := createTestRole(ctx1, rolesApiService, body1, &createdConstraintTestRoles, false, http.StatusCreated)
			Expect(role1).ToNot(BeNil())
			Expect(role1.Name).To(Equal(roleName))
			Expect(role1.ScopeType).To(Equal(tenantScopeType))

			// Create role with same name in tenant 2 - should succeed
			body2 := api.RoleCreationFieldsInternal{
				Name:        roleName,
				Description: "role for tenant 2",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_READ,
						},
					},
				},
				ScopeType: tenantScopeType,
				ScopeId:   ptr.To(tenant2Id),
				Enabled:   ptr.To(true),
			}

			role2 := createTestRole(ctx2, rolesApiService, body2, &createdConstraintTestRoles, false, http.StatusCreated)
			Expect(role2).ToNot(BeNil())
			Expect(role2.Name).To(Equal(roleName))
			Expect(role2.ScopeType).To(Equal(tenantScopeType))
			Expect(role2.Id).ToNot(Equal(role1.Id))
		})

		It("Tenant can't add same name as system role", func() {
			roleName := "system-protected-role"

			// Create system role first
			systemBody := api.RoleCreationFieldsInternal{
				Name:        roleName,
				Description: "system role that should prevent tenant usage",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_READ,
						},
					},
				},
				ScopeType: systemScopeType,
				Enabled:   ptr.To(true),
			}

			systemRole := createTestRole(mgmtCtxForConstraints, rolesApiService, systemBody, &createdConstraintTestRoles, false, http.StatusCreated)
			Expect(systemRole).ToNot(BeNil())
			Expect(systemRole.Name).To(Equal(roleName))
			Expect(systemRole.ScopeType).To(Equal(systemScopeType))

			// Try to create tenant role with same name - should fail with conflict
			tenantBody := api.RoleCreationFieldsInternal{
				Name:        roleName,
				Description: "tenant role that should fail",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_READ,
						},
					},
				},
				ScopeType: tenantScopeType,
				ScopeId:   ptr.To(tenant1Id),
				Enabled:   ptr.To(true),
			}

			createTestRole(ctx1, rolesApiService, tenantBody, &createdConstraintTestRoles, true, http.StatusConflict)
		})

		It("System can't add same name as existing tenant role", func() {
			roleName := "tenant-protected-role"

			// Create tenant role first
			tenantBody := api.RoleCreationFieldsInternal{
				Name:        roleName,
				Description: "tenant role that should prevent system usage",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_READ,
						},
					},
				},
				ScopeType: tenantScopeType,
				ScopeId:   ptr.To(tenant1Id),
				Enabled:   ptr.To(true),
			}

			tenantRole := createTestRole(ctx1, rolesApiService, tenantBody, &createdConstraintTestRoles, false, http.StatusCreated)
			Expect(tenantRole).ToNot(BeNil())
			Expect(tenantRole.Name).To(Equal(roleName))
			Expect(tenantRole.ScopeType).To(Equal(tenantScopeType))

			// Try to create system role with same name - should fail with conflict (409)
			systemBody := api.RoleCreationFieldsInternal{
				Name:        roleName,
				Description: "system role that should fail",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_READ,
						},
					},
				},
				ScopeType: systemScopeType,
				Enabled:   ptr.To(true),
			}

			createTestRole(mgmtCtxForConstraints, rolesApiService, systemBody, &createdConstraintTestRoles, true, http.StatusConflict)
		})

		It("Same tenant can't add duplicate role names", func() {
			roleName := "duplicate-in-tenant"

			// Create first role in tenant
			body1 := api.RoleCreationFieldsInternal{
				Name:        roleName,
				Description: "first role in tenant",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_READ,
						},
					},
				},
				ScopeType: tenantScopeType,
				ScopeId:   ptr.To(tenant1Id),
				Enabled:   ptr.To(true),
			}

			role1 := createTestRole(ctx1, rolesApiService, body1, &createdConstraintTestRoles, false, http.StatusCreated)
			Expect(role1).ToNot(BeNil())
			Expect(role1.Name).To(Equal(roleName))

			// Try to create another role with same name in same tenant - should fail
			body2 := api.RoleCreationFieldsInternal{
				Name:        roleName,
				Description: "duplicate role in same tenant",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_CREATE,
						},
					},
				},
				ScopeType: tenantScopeType,
				ScopeId:   ptr.To(tenant1Id),
				Enabled:   ptr.To(true),
			}

			createTestRole(ctx1, rolesApiService, body2, &createdConstraintTestRoles, true, http.StatusConflict)
		})

		It("System scope can't have duplicate role names", func() {
			roleName := "duplicate-system-role"

			// Create first system role
			body1 := api.RoleCreationFieldsInternal{
				Name:        roleName,
				Description: "first system role",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_READ,
						},
					},
				},
				ScopeType: systemScopeType,
				Enabled:   ptr.To(true),
			}

			systemRole1 := createTestRole(mgmtCtxForConstraints, rolesApiService, body1, &createdConstraintTestRoles, false, http.StatusCreated)
			Expect(systemRole1).ToNot(BeNil())
			Expect(systemRole1.Name).To(Equal(roleName))

			// Try to create another system role with same name - should fail
			body2 := api.RoleCreationFieldsInternal{
				Name:        roleName,
				Description: "duplicate system role",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions: []api.Action{
							api.ACTION_CREATE,
						},
					},
				},
				ScopeType: systemScopeType,
				Enabled:   ptr.To(true),
			}

			createTestRole(mgmtCtxForConstraints, rolesApiService, body2, &createdConstraintTestRoles, true, http.StatusConflict)
		})
	})
})

func testRoleActive(ctx context.Context, rolesApiService api.RolesAPIServicer, role api.Role, fields api.RoleActiveFields, expectedResponse int, expectedEnabled bool, expectedEffective bool) {
	var response api.ImplResponse
	var err error
	if fields.Enabled {
		response, err = rolesApiService.RoleEnable(ctx, role.Id, api.RoleEnabledDisabledFields{ScopeType: fields.ScopeType})
	} else {
		response, err = rolesApiService.RoleDisable(ctx, role.Id, api.RoleEnabledDisabledFields{ScopeType: fields.ScopeType})
	}
	Expect(err).To(BeNil())
	Expect(response.Code).To(Equal(expectedResponse))

	if expectedResponse >= 300 {
		return
	}

	roleFromServer, err := rolesApiService.GetRole(ctx, role.Id)
	Expect(err).To(BeNil())
	Expect(roleFromServer.Code).To(Equal(http.StatusOK))

	role = httpResponseBodyToRole(roleFromServer)
	Expect(*role.Enabled).To(Equal(expectedEnabled))
	Expect(role.EffectiveEnabled).To(Equal(expectedEffective))
}

func validateRole(serverRole, createdRole api.Role) {
	Expect(serverRole).ToNot(BeNil())
	Expect(serverRole.Id).To(Equal(createdRole.Id))
	Expect(serverRole.Name).To(Equal(createdRole.Name))
	Expect(serverRole.Description).To(Equal(createdRole.Description))
	Expect(serverRole.ScopeType).To(Equal(createdRole.ScopeType))
	Expect(serverRole.ScopeId).To(Equal(createdRole.ScopeId))
	Expect(serverRole.Enabled).To(Equal(createdRole.Enabled))
	Expect(serverRole.EffectiveEnabled).To(Equal(createdRole.EffectiveEnabled))
	if serverRole.KubernetesPermissions.PredefinedRole == nil {
		Expect(createdRole.KubernetesPermissions.PredefinedRole).To(BeNil())
	} else {
		Expect(createdRole.KubernetesPermissions.PredefinedRole).To(Not(BeNil()))
		Expect(*serverRole.KubernetesPermissions.PredefinedRole).To(Equal(*createdRole.KubernetesPermissions.PredefinedRole))
	}
	Expect(len(serverRole.Permissions)).To(Equal(len(createdRole.Permissions)))

	comparer := func(a, b api.RolePermission) int {
		return strings.Compare(string(a.ResourceType), string(b.ResourceType))
	}
	slices.SortFunc(serverRole.Permissions, comparer)
	slices.SortFunc(createdRole.Permissions, comparer)

	for i := range serverRole.Permissions {
		Expect(serverRole.Permissions[i].DisplayName).To(Equal(createdRole.Permissions[i].DisplayName))
		Expect(serverRole.Permissions[i].ResourceType).To(Equal(createdRole.Permissions[i].ResourceType))
		Expect(serverRole.Permissions[i].Actions).To(Equal(createdRole.Permissions[i].Actions))
	}
}

func createTestRole(ctx context.Context, rolesApiService api.RolesAPIServicer, creationRequest api.RoleCreationFieldsInternal, createdRoles *[]api.Role, shouldFail bool, expectedResponse int) *api.Role {
	createdRoleResponse, err := rolesApiService.CreateRole(ctx, creationRequest)
	Expect(err).To(BeNil())

	Expect(createdRoleResponse.Code).To(Equal(expectedResponse), fmt.Sprintf("error: %v", createdRoleResponse.Body))

	if shouldFail {
		return nil
	}

	role := httpResponseBodyToRole(createdRoleResponse)

	*createdRoles = append(*createdRoles, role)
	return &role
}

func httpResponseBodyToRole(httpResponse api.ImplResponse) api.Role {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	role := &api.Role{}
	err = json.Unmarshal(bytes, role)
	Expect(err).To(BeNil())
	return *role
}

func httpResponseBodyToRoles(httpResponse api.ImplResponse) []api.Role {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	roles := &[]api.Role{}
	err = json.Unmarshal(bytes, roles)
	Expect(err).To(BeNil())
	return *roles
}

func permissionActionsToActions(permissionActions []api.PermissionAction) []api.Action {
	actions := []api.Action{}
	for i := range permissionActions {
		actions = append(actions, api.Action(permissionActions[i]))
	}
	return actions
}

func isLegacyRole(roleName string) bool {
	for _, role := range dao.LegacyRoles {
		if role.Name == roleName {
			return true
		}
	}
	for _, role := range dao.RolesMigrated {
		if role.Name == roleName && role.Deprecated {
			return true
		}
	}
	for _, role := range dao.RolesNew {
		if role.Name == roleName && role.Deprecated {
			return true
		}
	}
	return false
}
