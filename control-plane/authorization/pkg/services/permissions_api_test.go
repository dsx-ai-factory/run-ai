/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	static_legacy_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/exp/slices"
)

var (
	ClusterId     = "a418ed33-9399-48c0-a890-122cadd13bfd"
	DepartmentId  = "1"
	DepartmentId2 = "2"
	ProjectId     = "1"
	ProjectId2    = "2"

	TestProjectId1    = "461001"
	TestProjectId2    = "461002"
	TestProjectId3    = "461003"
	TestDepartmentId1 = "450001"
	TestDepartmentId2 = "450002"
)

var _ = Describe("Permissions API tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	createdAccessRules := []api.AccessRule{}
	testAdminsGroups := []string{dao.SuperAdmins}
	ctx := setContext(tenantId, TestUserId, testAdminsGroups)

	tenantSettings := NewMockTenantSettingsService(map[string]bool{})

	accessRulesApiService := NewAccessRulesApiService(tenantSettings)
	permissionsApiService := NewPermissionsApiService()

	RunPermissionsTest := func(assignRoles []models.Role) {
		for _, role := range assignRoles {
			roleId := RolesMap[role.Name]
			createTestAccessRule(ctx, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     tenantId,
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(roleId),
			}, &createdAccessRules, false, http.StatusCreated)
		}

		noGroupCtx := setContext(tenantId, TestUserId, []string{})
		response, err := permissionsApiService.GetPermissions(noGroupCtx)
		Expect(err).To(BeNil())

		responsePermissions := httpResponseBodyToPermissions(response)
		responsePermissionsMap := createPermissionsMapping(responsePermissions)

		rolesPermissionsMap := make(map[api.ResourceType][]api.Action)
		for _, role := range assignRoles {
			if role.PermissionSets != nil && len(*role.PermissionSets) > 0 {
				dbRole, apiErr := dao.DbGetRoleById(ctx, "", dao.MapStaticRoleNameToRoleId[role.Name], nil)
				Expect(apiErr).To(BeNil())
				role = *dbRole
			}
			permissionsDaoToApi := PermissionsDaoToApi(RemoveDGXCReadOnlyPermissions(role.Permissions))
			rolePermissionsMap := createPermissionsMapping(permissionsDaoToApi)
			rolesPermissionsMap = concatMaps(rolesPermissionsMap, rolePermissionsMap)
		}

		validateMapsAreEqual(responsePermissionsMap, rolesPermissionsMap)
	}

	RunPermittedScopesTest := func(assignRole models.Role, scope api.ScopeType, scopeId string) {
		roleId := RolesMap[assignRole.Name]
		createTestAccessRule(ctx, accessRulesApiService, api.AccessRuleCreationFields{
			ScopeId:     scopeId,
			ScopeType:   scope,
			SubjectId:   TestUserId,
			SubjectType: api.SUBJECTTYPE_USER,
			RoleId:      int32(roleId),
		}, &createdAccessRules, false, http.StatusCreated)

		noGroupCtx := setContext(tenantId, TestUserId, []string{})
		if assignRole.PermissionSets != nil && len(*assignRole.PermissionSets) > 0 {
			dbRole, apiErr := dao.DbGetRoleById(ctx, "", dao.MapStaticRoleNameToRoleId[assignRole.Name], nil)
			Expect(apiErr).To(BeNil())
			assignRole = *dbRole
		}
		for _, permission := range RemoveDGXCReadOnlyPermissions(assignRole.Permissions) {
			for _, action := range []api.Action{api.ACTION_CREATE, api.ACTION_UPDATE, api.ACTION_DELETE, api.ACTION_READ} {
				actionStr := string(action)
				response, err := permissionsApiService.GetPermittedScopes(noGroupCtx,
					api.GetPermittedScopesRequest{
						ResourceType: permission.ResourceType,
						Action:       &actionStr,
					})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				permittedScopesActions := httpResponseBodyToPermittedScopesActions(response)
				switch action {
				case api.ACTION_CREATE:
					validatePermittedScopes(scope, scopeId, permission.Create, permittedScopesActions.Create)
					validatePermittedScopesAreEmpty(permittedScopesActions.Update, permittedScopesActions.Delete, permittedScopesActions.Read)
				case api.ACTION_UPDATE:
					validatePermittedScopes(scope, scopeId, permission.Update, permittedScopesActions.Update)
					validatePermittedScopesAreEmpty(permittedScopesActions.Create, permittedScopesActions.Delete, permittedScopesActions.Read)
				case api.ACTION_DELETE:
					validatePermittedScopes(scope, scopeId, permission.Delete, permittedScopesActions.Delete)
					validatePermittedScopesAreEmpty(permittedScopesActions.Update, permittedScopesActions.Create, permittedScopesActions.Read)
				case api.ACTION_READ:
					if permission.ResourceType == api.RESOURCETYPE_TENANT {
						validateScope(permittedScopesActions.Read.Tenants, permission.Read, tenantId)
					} else {
						validatePermittedScopes(scope, scopeId, permission.Read, permittedScopesActions.Read)
					}
					validatePermittedScopesAreEmpty(permittedScopesActions.Update, permittedScopesActions.Delete, permittedScopesActions.Create)
				}
			}

			// get permitted scopes for all actions
			response, err := permissionsApiService.GetPermittedScopes(noGroupCtx,
				api.GetPermittedScopesRequest{
					ResourceType: permission.ResourceType,
				})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			permittedScopesActions := httpResponseBodyToPermittedScopesActions(response)
			validatePermittedScopes(scope, scopeId, permission.Create, permittedScopesActions.Create)
			validatePermittedScopes(scope, scopeId, permission.Update, permittedScopesActions.Update)
			validatePermittedScopes(scope, scopeId, permission.Delete, permittedScopesActions.Delete)

			if permission.ResourceType == api.RESOURCETYPE_TENANT {
				validateScope(permittedScopesActions.Read.Tenants, permission.Read, tenantId)
			} else {
				validatePermittedScopes(scope, scopeId, permission.Read, permittedScopesActions.Read)
			}
		}
	}

	BeforeAll(func() {
		TestsCleanupAndInit()

		SetAncestorsCache(ctx, ProjectId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
			TenantId:     tenantId,
			ClusterUuid:  &ClusterId,
			DepartmentId: &DepartmentId,
		})

		SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
			TenantId:    tenantId,
			ClusterUuid: &ClusterId,
		})

		SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
			TenantId: tenantId,
		})
	})

	AfterAll(func() {
		TestsCleanupAndInit()
	})

	BeforeEach(func() {
		CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
		createdAccessRules = []api.AccessRule{}
	})

	DescribeTable("Running tests", RunPermissionsTest,
		Entry("single role", []models.Role{static_legacy_roles.Researcher}),
		Entry("multiple roles", []models.Role{static_legacy_roles.Researcher, static_legacy_roles.TemplatesAdmin, static_legacy_roles.ComputeResourcesAdmin}),
		Entry("LegacyRoles that are containing each other", []models.Role{static_legacy_roles.Researcher, static_roles.DepartmentAdmin, static_roles.SystemAdmin}),
	)

	DescribeTable("Running tests permitted scopes", RunPermittedScopesTest,
		Entry("Validate researcher role in tenant scope permitted scopes", static_legacy_roles.Researcher, api.SCOPETYPE_TENANT, tenantId),
		Entry("Validate admin role in cluster scope permitted scopes", static_roles.SystemAdmin, api.SCOPETYPE_CLUSTER, ClusterId),
		Entry("Validate environment admin role in department scope permitted scopes", static_legacy_roles.EnvironmentAdmin, api.SCOPETYPE_DEPARTMENT, DepartmentId),
		Entry("Validate editor role in project scope permitted scopes", static_roles.Editor, api.SCOPETYPE_PROJECT, ProjectId),
	)
})

type EvaluationTestArgs struct {
	SubjectId               string
	Groups                  []string
	RequestedResourceType   *api.ResourceType
	RequestedAction         *api.Action
	RequestedId             *string
	ExpectedStatusCode      int
	ExpectedPermittedScopes api.PermittedScopes
}

func calculateHasPermission(permission *models.Permission, action api.Action) bool {
	if testing.TestTenant == testing.DGXCTestTenant && permission.ReadOnlyForDGXCManaged && action != api.ACTION_READ {
		return false
	}

	switch action {
	case api.ACTION_CREATE:
		return permission.Create
	case api.ACTION_READ:
		return permission.Read
	case api.ACTION_UPDATE:
		return permission.Update
	case api.ACTION_SYNC:
		return permission.Sync
	default:
		return permission.Delete
	}
}

func expectedStatusCode(resourceType api.ResourceType, hasPermission bool) int {
	if !hasPermission {
		return http.StatusForbidden
	}
	if dao.IsGlobalResource(resourceType) {
		return http.StatusNoContent
	}
	return http.StatusOK
}

func httpResponseBodyToPermittedScopes(httpResponse api.ImplResponse) api.PermittedScopes {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	permittedScopes := &api.PermittedScopes{}
	err = json.Unmarshal(bytes, permittedScopes)
	Expect(err).To(BeNil())
	return *permittedScopes
}

func httpResponseBodyToPermittedScopesActions(httpResponse api.ImplResponse) api.PermittedScopesActions {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	permittedScopesActions := &api.PermittedScopesActions{}
	err = json.Unmarshal(bytes, permittedScopesActions)
	Expect(err).To(BeNil())
	return *permittedScopesActions
}

var _ = Describe("Evaluation API tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	createdRoles := []api.Role{}
	createdAccessRules := []api.AccessRule{}
	ctx := setContext(tenantId, TestUserId, []string{dao.SuperAdmins})
	managementTenant := fmt.Sprintf("%d", testing.ManagementTestTenant)
	ctxManagement := context.WithValue(setContext(managementTenant, TestUserId, []string{dao.SuperAdmins}), runaictx.CfIsManagementTenant, "true")

	evalTenantSettings := NewMockTenantSettingsService(map[string]bool{})

	clustersClient := cluster_service_client.NewAPIClient(&cluster_service_client.Configuration{
		Servers: cluster_service_client.ServerConfigurations{
			{
				URL: config.Get[*authorization_config.Server]().ClusterServiceUrl,
			},
		},
	})

	rolesApiService := NewRolesApiService(clustersClient, evalTenantSettings)
	accessRulesApiService := NewAccessRulesApiService(evalTenantSettings)
	evaluationApiService := NewPermissionsApiService()

	BeforeAll(func() {
		TestsCleanupAndInit()
	})

	AfterAll(func() {
		TestsCleanupAndInit()
	})

	RunEvaluationTest := func(args EvaluationTestArgs) {
		response, err := evaluationApiService.EvaluatePermission(setContext(tenantId, args.SubjectId, args.Groups), api.BackendEvaluationRequest{
			ResourceType:  args.RequestedResourceType,
			RequestAction: args.RequestedAction,
			ResourceId:    args.RequestedId,
		})
		Expect(err).To(BeNil())
		if args.ExpectedPermittedScopes.System {
			// We asked for evaluation in the general tenant. This tenant should never return a system scope, even if the user has permission he will get them only after authenticating to the system tenant 0
			Expect(response.Code).To(Equal(http.StatusForbidden))
			response, err = evaluationApiService.EvaluatePermission(context.WithValue(setContext(managementTenant, args.SubjectId, args.Groups), runaictx.CfIsManagementTenant, "true"), api.BackendEvaluationRequest{
				ResourceType:  args.RequestedResourceType,
				RequestAction: args.RequestedAction,
				ResourceId:    args.RequestedId,
			})
			Expect(err).To(BeNil())
		}

		Expect(response.Code).To(Equal(args.ExpectedStatusCode))
		if args.ExpectedStatusCode == http.StatusOK {
			permittedScopes := httpResponseBodyToPermittedScopes(response)
			Expect(permittedScopes.Tenants).To(Equal(args.ExpectedPermittedScopes.Tenants))
			Expect(permittedScopes.Clusters).To(Equal(args.ExpectedPermittedScopes.Clusters))
			Expect(permittedScopes.Departments).To(Equal(args.ExpectedPermittedScopes.Departments))
			Expect(permittedScopes.Projects).To(Equal(args.ExpectedPermittedScopes.Projects))
		}
	}

	Context("Test global resources permissions", func() {
		const TestGroup = "group-a"

		BeforeAll(func() {
			accessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     fmt.Sprintf("%d", testing.TestTenant),
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),
			}

			createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
			// TestUser is now a department admin

			accessRuleBody = api.AccessRuleCreationFields{
				ScopeId:     fmt.Sprintf("%d", testing.TestTenant),
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestGroup,
				SubjectType: api.SUBJECTTYPE_GROUP,
				RoleId:      int32(RolesMap[static_roles.Viewer.Name]),
			}

			createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
			// Every user in TestGroup is now a viewer
		})

		AfterAll(func() {
			CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
			createdAccessRules = []api.AccessRule{}

			for _, role := range createdRoles {
				_, err := rolesApiService.DeleteRole(ctx, role.Id)
				Expect(err).To(BeNil())
			}
			createdRoles = []api.Role{}
		})

		entries := []TableEntry{}

		resourceTypes := []api.ResourceType{
			api.RESOURCETYPE_USERS,
			api.RESOURCETYPE_APPS,
			api.RESOURCETYPE_EVENTS_HISTORY,
		}

		requestActions := []api.Action{
			api.ACTION_READ,
			api.ACTION_CREATE,
			api.ACTION_UPDATE,
			api.ACTION_DELETE,
		}

		for _, resourceType := range resourceTypes {
			for _, requestAction := range requestActions {
				expectedStatus := http.StatusNoContent
				if resourceType == api.RESOURCETYPE_EVENTS_HISTORY {
					expectedStatus = http.StatusForbidden
				}

				if testing.TestTenant == testing.DGXCTestTenant && requestAction != api.ACTION_READ && resourceType != api.RESOURCETYPE_EVENTS_HISTORY {
					expectedStatus = http.StatusForbidden
				}

				entries = append(entries, Entry(fmt.Sprintf("User asking to %s resource %s", string(requestAction), string(resourceType)), EvaluationTestArgs{
					SubjectId:             TestUserId,
					Groups:                nil,
					RequestedResourceType: &resourceType,
					RequestedAction:       &requestAction,
					ExpectedStatusCode:    expectedStatus,
				}))
			}
		}

		DescribeTable("Running tests", RunEvaluationTest, entries)
	})

	Context("Scoped permissions", func() {
		const (
			SystemResearchers     = "group-e"
			TenantResearchers     = "group-a"
			ClusterResearchers    = "group-b"
			DepartmentResearchers = "group-c"
			ProjectResearchers    = "group-d"
			OtherProjectId        = "1234"
		)
		BeforeAll(func() {
			SetAncestorsCache(ctx, ProjectId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:     tenantId,
				ClusterUuid:  &ClusterId,
				DepartmentId: &DepartmentId,
			})

			SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &ClusterId,
			})

			SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, OtherProjectId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{})

			createTestAccessRule(ctxManagement, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     "",
				ScopeType:   api.SCOPETYPE_SYSTEM,
				SubjectId:   SystemResearchers,
				SubjectType: api.SUBJECTTYPE_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.Researcher.Name]),
			}, &createdAccessRules, false, http.StatusCreated)

			createTestAccessRule(ctx, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     tenantId,
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TenantResearchers,
				SubjectType: api.SUBJECTTYPE_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.Researcher.Name]),
			}, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   ClusterResearchers,
				SubjectType: api.SUBJECTTYPE_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.Researcher.Name]),
			}, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   DepartmentResearchers,
				SubjectType: api.SUBJECTTYPE_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.Researcher.Name]),
			}, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   ProjectResearchers,
				SubjectType: api.SUBJECTTYPE_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.Researcher.Name]),
			}, &createdAccessRules, false, http.StatusCreated)
		})

		AfterAll(func() {
			// Clean up access rules with appropriate context based on tenant
			for _, accessRule := range createdAccessRules {
				deleteCtx := ctx
				if accessRule.TenantId == int32(testing.ManagementTestTenant) {
					deleteCtx = ctxManagement
				}
				_, err := accessRulesApiService.DeleteAccessRule(deleteCtx, accessRule.Id)
				Expect(err).To(BeNil())
			}
			createdAccessRules = []api.AccessRule{}

			for _, role := range createdRoles {
				_, err := rolesApiService.DeleteRole(ctx, role.Id)
				Expect(err).To(BeNil())
			}
			createdRoles = []api.Role{}

			Invalidate(ctx, ProjectId, api.RESOURCETYPE_PROJECT)

			Invalidate(ctx, OtherProjectId, api.RESOURCETYPE_PROJECT)
		})

		projectId := ProjectId           // can not take address of const therefore copying the value to a local param
		otherProjectId := OtherProjectId // can not take address of const therefore copying the value to a local param

		existingScopes := []api.ScopeType{
			api.SCOPETYPE_SYSTEM,
			api.SCOPETYPE_TENANT,
			api.SCOPETYPE_CLUSTER,
			api.SCOPETYPE_DEPARTMENT,
			api.SCOPETYPE_PROJECT,
		}

		entries := []TableEntry{
			Entry("User in all groups should get all scopes", EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{TenantResearchers, ClusterResearchers, DepartmentResearchers, ProjectResearchers},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_PROJECT),
				RequestedAction:       ptr.To(api.ACTION_READ),
				ExpectedStatusCode:    http.StatusOK,
				ExpectedPermittedScopes: api.PermittedScopes{
					Tenants:     []string{tenantId},
					Clusters:    []string{ClusterId},
					Departments: []string{DepartmentId},
					Projects:    []string{ProjectId},
				},
			}),
			Entry("User in no group is not permitted", EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{TenantResearchers, ClusterResearchers, DepartmentResearchers, ProjectResearchers},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_PROJECT),
				RequestedAction:       ptr.To(api.ACTION_READ),
				ExpectedStatusCode:    http.StatusOK,
				ExpectedPermittedScopes: api.PermittedScopes{
					Tenants:     []string{tenantId},
					Clusters:    []string{ClusterId},
					Departments: []string{DepartmentId},
					Projects:    []string{ProjectId},
				},
			}),
			Entry("Get for project not under the permitted scopes should fail", EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{TenantResearchers, ClusterResearchers, DepartmentResearchers, ProjectResearchers},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_PROJECT),
				RequestedAction:       ptr.To(api.ACTION_READ),
				RequestedId:           &otherProjectId,
				ExpectedStatusCode:    http.StatusForbidden,
			}),
		}

		for _, scope := range existingScopes {
			permittedScopes := api.PermittedScopes{}
			var group string
			switch scope {
			case api.SCOPETYPE_SYSTEM:
				permittedScopes.System = true
				group = SystemResearchers
				break
			case api.SCOPETYPE_TENANT:
				permittedScopes.Tenants = []string{tenantId}
				group = TenantResearchers
				break
			case api.SCOPETYPE_CLUSTER:
				permittedScopes.Clusters = []string{ClusterId}
				group = ClusterResearchers
				break
			case api.SCOPETYPE_DEPARTMENT:
				permittedScopes.Departments = []string{DepartmentId}
				group = DepartmentResearchers
				break
			case api.SCOPETYPE_PROJECT:
				permittedScopes.Projects = []string{ProjectId}
				group = ProjectResearchers
				break
			}

			entries = append(entries,
				Entry(fmt.Sprintf("User in group %v should get permitted scope %s", group, string(scope)), EvaluationTestArgs{
					SubjectId:               TestUserId,
					Groups:                  []string{group},
					RequestedResourceType:   ptr.To(api.RESOURCETYPE_PROJECT),
					RequestedAction:         ptr.To(api.ACTION_READ),
					ExpectedStatusCode:      http.StatusOK,
					ExpectedPermittedScopes: permittedScopes,
				}),
				Entry(fmt.Sprintf("User in group %v should get permitted scope %s when asking for specific resource under that scope.", group, string(scope)), EvaluationTestArgs{
					SubjectId:               TestUserId,
					Groups:                  []string{group},
					RequestedResourceType:   ptr.To(api.RESOURCETYPE_PROJECT),
					RequestedAction:         ptr.To(api.ACTION_READ),
					RequestedId:             &projectId,
					ExpectedStatusCode:      http.StatusOK,
					ExpectedPermittedScopes: permittedScopes,
				}),
				Entry(fmt.Sprintf("User in group %v should get permitted scope %s when submitting a workload.", group, string(scope)), EvaluationTestArgs{
					SubjectId:               TestUserId,
					Groups:                  []string{group},
					RequestedResourceType:   ptr.To(api.RESOURCETYPE_WORKLOADS),
					RequestedAction:         ptr.To(api.ACTION_CREATE),
					ExpectedStatusCode:      http.StatusOK,
					ExpectedPermittedScopes: permittedScopes,
				}),
				Entry(fmt.Sprintf("Researcher in group %s can not create an inference even in scope %s", group, string(scope)), EvaluationTestArgs{
					SubjectId:             TestUserId,
					Groups:                []string{group},
					RequestedResourceType: ptr.To(api.RESOURCETYPE_INFERENCES),
					RequestedAction:       ptr.To(api.ACTION_CREATE),
					ExpectedStatusCode:    http.StatusForbidden,
				}))
		}

		DescribeTable("Running tests", RunEvaluationTest, entries)
	})

	Context("Test all roles permissions", func() {
		AfterEach(func() {
			CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
			createdAccessRules = []api.AccessRule{}
		})

		RunEvaluationTestForRole := func(role models.Role) {
			createTestAccessRule(ctx, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     tenantId,
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[role.Name]),
			}, &createdAccessRules, false, http.StatusCreated)

			actions := []api.Action{
				api.ACTION_CREATE,
				api.ACTION_READ,
				api.ACTION_UPDATE,
				api.ACTION_DELETE,
			}

			if role.PermissionSets != nil && len(*role.PermissionSets) > 0 {
				dbRole, apiErr := dao.DbGetRoleById(ctx, "", dao.MapStaticRoleNameToRoleId[role.Name], nil)
				Expect(apiErr).To(BeNil())
				role = *dbRole
			}
			for _, permission := range role.Permissions {
				for _, action := range actions {
					hasPermission := calculateHasPermission(permission, action)
					args := EvaluationTestArgs{
						SubjectId:             TestUserId,
						Groups:                []string{},
						RequestedResourceType: &permission.ResourceType,
						RequestedAction:       &action,
						ExpectedStatusCode:    expectedStatusCode(permission.ResourceType, hasPermission),
						ExpectedPermittedScopes: api.PermittedScopes{
							Tenants: []string{tenantId},
						},
					}
					RunEvaluationTest(args)
				}
			}
		}

		entries := []TableEntry{}

		for _, role := range dao.LegacyRoles {
			if !role.InternalOnly {
				entries = append(entries, Entry(role.Name, role))
			}
		}

		DescribeTable("Running tests", RunEvaluationTestForRole, entries)
	})

	Context("Test compute resources access in org scoped permissions", func() {
		AfterEach(func() {
			CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
			createdAccessRules = []api.AccessRule{}
		})
		BeforeEach(func() {
			SetAncestorsCache(ctx, ProjectId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:     tenantId,
				ClusterUuid:  &ClusterId,
				DepartmentId: &DepartmentId,
			})
			accessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_legacy_roles.Researcher.Name]),
			}

			createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
		})

		It("Researcher in project can get clusters", func() {
			args := EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_CLUSTER),
				RequestedAction:       ptr.To(api.ACTION_READ),
				ExpectedStatusCode:    http.StatusOK,
				ExpectedPermittedScopes: api.PermittedScopes{
					Clusters: []string{ClusterId},
					Projects: []string{ProjectId},
				},
			}
			RunEvaluationTest(args)
		})

		It("Researcher in project can get specific cluster", func() {
			args := EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_CLUSTER),
				RequestedId:           &ClusterId,
				RequestedAction:       ptr.To(api.ACTION_READ),
				ExpectedStatusCode:    http.StatusOK,
				ExpectedPermittedScopes: api.PermittedScopes{
					Clusters: []string{ClusterId},
					Projects: []string{ProjectId},
				},
			}
			RunEvaluationTest(args)
		})

		It("Researcher in project can get tenants", func() {
			args := EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_TENANT),
				RequestedAction:       ptr.To(api.ACTION_READ),
				ExpectedStatusCode:    http.StatusOK,
				ExpectedPermittedScopes: api.PermittedScopes{
					Tenants:  []string{tenantId},
					Clusters: []string{ClusterId},
					Projects: []string{ProjectId},
				},
			}
			RunEvaluationTest(args)
		})

		It("Researcher in project can get node pools in cluster", func() {
			args := EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_NODEPOOLS),
				RequestedAction:       ptr.To(api.ACTION_READ),
				ExpectedStatusCode:    http.StatusOK,
				ExpectedPermittedScopes: api.PermittedScopes{
					Clusters: []string{ClusterId},
					Projects: []string{ProjectId},
				},
			}
			RunEvaluationTest(args)
		})
		It("Researcher in project can not create node pools", func() {
			args := EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_NODEPOOLS),
				RequestedAction:       ptr.To(api.ACTION_CREATE),
				ExpectedStatusCode:    http.StatusForbidden,
			}
			RunEvaluationTest(args)
		})
		It("Researcher in project has any read permission", func() {
			args := EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{},
				RequestedResourceType: nil,
				RequestedAction:       ptr.To(api.ACTION_READ),
				ExpectedStatusCode:    http.StatusNoContent,
			}
			RunEvaluationTest(args)
		})
		It("Researcher in project has any project permission", func() {
			args := EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_PROJECT),
				RequestedAction:       nil,
				ExpectedStatusCode:    http.StatusOK,
				ExpectedPermittedScopes: api.PermittedScopes{
					Projects: []string{ProjectId},
				},
			}
			RunEvaluationTest(args)
		})
		It("Researcher in project has any permission", func() {
			args := EvaluationTestArgs{
				SubjectId:             TestUserId,
				Groups:                []string{},
				RequestedResourceType: nil,
				RequestedAction:       nil,
				ExpectedStatusCode:    http.StatusNoContent,
			}
			RunEvaluationTest(args)
		})
		It("Unknown user has no permissions", func() {
			args := EvaluationTestArgs{
				SubjectId:             "unknown-user",
				Groups:                []string{},
				RequestedResourceType: nil,
				RequestedAction:       nil,
				ExpectedStatusCode:    http.StatusForbidden,
			}
			RunEvaluationTest(args)
		})
	})

	Context("All authenticated users", func() {
		const (
			allAuthUser         = "no-rule-user@run.ai"
			allAuthOtherProject = "9999"
			unrelatedGroup      = "unrelated-group-x"
		)
		var (
			allAuthRuleId int32
			ruleDeleted   bool
		)

		BeforeAll(func() {
			ruleDeleted = false

			SetAncestorsCache(ctx, ProjectId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:     tenantId,
				ClusterUuid:  &ClusterId,
				DepartmentId: &DepartmentId,
			})
			SetAncestorsCache(ctx, allAuthOtherProject, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:     tenantId,
				ClusterUuid:  &ClusterId,
				DepartmentId: &DepartmentId,
			})

			resp, err := accessRulesApiService.CreateAuthenticatedUsersAccessRule(ctx, api.AuthenticatedUsersAccessRuleCreationFields{
				RoleId:    int32(RolesMap[static_legacy_roles.Researcher.Name]),
				ScopeId:   ProjectId,
				ScopeType: string(api.SCOPETYPE_PROJECT),
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			allAuthRuleId = HttpResponseBodyToAccessRule(resp).Id
		})

		AfterAll(func() {
			if !ruleDeleted {
				_, err := accessRulesApiService.DeleteAccessRule(ctx, allAuthRuleId)
				Expect(err).To(BeNil())
			}
			Invalidate(ctx, ProjectId, api.RESOURCETYPE_PROJECT)
			Invalidate(ctx, allAuthOtherProject, api.RESOURCETYPE_PROJECT)
		})

		It("grants a user with no rules and no groups access on the scoped project", func() {
			RunEvaluationTest(EvaluationTestArgs{
				SubjectId:             allAuthUser,
				Groups:                []string{},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_PROJECT),
				RequestedAction:       ptr.To(api.ACTION_READ),
				ExpectedStatusCode:    http.StatusOK,
				ExpectedPermittedScopes: api.PermittedScopes{
					Projects: []string{ProjectId},
				},
			})
		})

		It("denies access on a different project", func() {
			otherProject := allAuthOtherProject
			RunEvaluationTest(EvaluationTestArgs{
				SubjectId:             allAuthUser,
				Groups:                []string{},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_PROJECT),
				RequestedAction:       ptr.To(api.ACTION_READ),
				RequestedId:           &otherProject,
				ExpectedStatusCode:    http.StatusForbidden,
			})
		})

		It("grants access even when the user belongs to unrelated groups", func() {
			RunEvaluationTest(EvaluationTestArgs{
				SubjectId:             allAuthUser,
				Groups:                []string{unrelatedGroup},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_PROJECT),
				RequestedAction:       ptr.To(api.ACTION_READ),
				ExpectedStatusCode:    http.StatusOK,
				ExpectedPermittedScopes: api.PermittedScopes{
					Projects: []string{ProjectId},
				},
			})
		})

		It("does not leak to a user in a different tenant", func() {
			otherTenantProject := "8888"
			SetAncestorsCache(ctx, otherTenantProject, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:     managementTenant,
				ClusterUuid:  &ClusterId,
				DepartmentId: &DepartmentId,
			})
			defer Invalidate(ctx, otherTenantProject, api.RESOURCETYPE_PROJECT)

			response, err := evaluationApiService.EvaluatePermission(
				setContext(managementTenant, allAuthUser, []string{}),
				api.BackendEvaluationRequest{
					ResourceType:  ptr.To(api.RESOURCETYPE_PROJECT),
					RequestAction: ptr.To(api.ACTION_READ),
					ResourceId:    &otherTenantProject,
				})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})

		It("lists the rule for a user with none of their own", func() {
			rules, apiErr := dao.DbGetMeAccessRules(ctx, testing.TestTenant, allAuthUser, nil, 0, 0, "")
			Expect(apiErr).To(BeNil())
			Expect(*rules).To(HaveLen(1))
			Expect((*rules)[0].SubjectType).To(BeEquivalentTo(models.AllAuthenticatedUsers))

			count, apiErr := dao.DbCountMeAccessRules(ctx, testing.TestTenant, allAuthUser, nil, "")
			Expect(apiErr).To(BeNil())
			Expect(count).To(Equal(int64(1)))
		})

		It("returns the rule from GetMeAccessRules with the extended subject type", func() {
			meApiService := NewMeApiService()
			testCtx := setContext(tenantId, allAuthUser, []string{})

			resp, err := meApiService.GetMeAccessRules(testCtx, 0, 0, "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			meAccessRules := httpResponseBodyToMeAccessRules(resp).AccessRules.AccessRules
			Expect(meAccessRules).To(HaveLen(1))
			Expect(meAccessRules[0].Id).To(Equal(allAuthRuleId))
			Expect(meAccessRules[0].SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_ALL_AUTHENTICATED_USERS))
			Expect(meAccessRules[0].SubjectId).To(BeEquivalentTo(models.AllAuthenticatedUsers))
			Expect(meAccessRules[0].ScopeId).To(Equal(ProjectId))
		})

		It("no longer grants once soft-deleted", func() {
			_, err := accessRulesApiService.DeleteAccessRule(ctx, allAuthRuleId)
			Expect(err).To(BeNil())
			ruleDeleted = true

			RunEvaluationTest(EvaluationTestArgs{
				SubjectId:             allAuthUser,
				Groups:                []string{},
				RequestedResourceType: ptr.To(api.RESOURCETYPE_PROJECT),
				RequestedAction:       ptr.To(api.ACTION_READ),
				ExpectedStatusCode:    http.StatusForbidden,
			})

			count, apiErr := dao.DbCountMeAccessRules(ctx, testing.TestTenant, allAuthUser, nil, "")
			Expect(apiErr).To(BeNil())
			Expect(count).To(Equal(int64(0)))
		})
	})
})

func validatePermittedScopes(scope api.ScopeType, scopeId string, actionIsAllowed bool, permittedScopes api.PermittedScopes) {
	if actionIsAllowed {
		validateScope(permittedScopes.Tenants, scope == api.SCOPETYPE_TENANT, scopeId)
		validateScope(permittedScopes.Clusters, scope == api.SCOPETYPE_CLUSTER, scopeId)
		validateScope(permittedScopes.Departments, scope == api.SCOPETYPE_DEPARTMENT, scopeId)
		validateScope(permittedScopes.Projects, scope == api.SCOPETYPE_PROJECT, scopeId)
	} else {
		Expect(permittedScopes).To(Equal(api.PermittedScopes{}))
	}
}

func validateScope(scopes []string, shouldBeInside bool, id string) {
	if shouldBeInside {
		Expect(scopes).To(ContainElements(id))
	} else {
		Expect(scopes).To(Not(ContainElements(id)))
	}
}

func validatePermittedScopesAreEmpty(permittedScopess ...api.PermittedScopes) {
	for _, permittedScopes := range permittedScopess {
		Expect(permittedScopes.Tenants).To(BeEmpty())
		Expect(permittedScopes.Clusters).To(BeEmpty())
		Expect(permittedScopes.Departments).To(BeEmpty())
		Expect(permittedScopes.Projects).To(BeEmpty())
	}
}

func validateMapsAreEqual(a, b map[api.ResourceType][]api.Action) {
	a = removeEmptyPermissions(a)
	b = removeEmptyPermissions(b)
	Expect(len(a)).To(Equal(len(b)))

	for keyA, valA := range a {
		valB, ok := b[keyA]
		Expect(ok).To(BeTrue())

		slices.Sort(valA)
		slices.Sort(valB)
		Expect(slices.Equal(valA, valB)).To(BeTrue())
	}
}

func removeEmptyPermissions(permissionsMap map[api.ResourceType][]api.Action) map[api.ResourceType][]api.Action {
	newMap := map[api.ResourceType][]api.Action{}

	for key, value := range permissionsMap {
		if len(value) > 0 {
			newMap[key] = value
		}
	}

	return newMap
}

func concatMaps(map1, map2 map[api.ResourceType][]api.Action) map[api.ResourceType][]api.Action {
	for key, val := range map2 {
		if _, ok := map1[key]; ok {
			for _, v := range val {
				if !slices.Contains(map1[key], v) {
					map1[key] = append(map1[key], v)
				}
			}
		} else {
			map1[key] = val
		}
	}

	return map1
}

func createPermissionsMapping(permissions []api.Permission) map[api.ResourceType][]api.Action {
	resultMap := make(map[api.ResourceType][]api.Action)
	for _, permission := range permissions {
		_, keyExists := resultMap[permission.ResourceType]
		Expect(keyExists).To(BeFalse())
		resultMap[permission.ResourceType] = permission.Actions
	}

	return resultMap
}

func httpResponseBodyToPermissions(httpResponse api.ImplResponse) []api.Permission {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	permissions := &[]api.Permission{}
	err = json.Unmarshal(bytes, permissions)
	Expect(err).To(BeNil())
	return *permissions
}

func RemoveDGXCReadOnlyPermissions(permissions []*models.Permission) []*models.Permission {
	result := []*models.Permission{}
	for _, permission := range permissions {
		if testing.DGXCTestTenant == testing.TestTenant && permission.ReadOnlyForDGXCManaged {
			result = append(result, &models.Permission{
				ResourceType: permission.ResourceType,
				Read:         permission.Read,
			})
		} else {
			result = append(result, permission)
		}
	}

	return result
}
