package services

import (
	"context"
	"fmt"
	"net/http"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"
	testing "github.com/run-ai/runai/control-plane/authorization/pkg/test"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
)

const (
	TestUser1Id = "user1@run.ai"
)

var _ = Describe("GetAccessRules scope hierarchy tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	ctx := setContext(tenantId, TestUser1Id, []string{dao.SuperAdmins})

	tenantSettingsService := NewMockTenantSettingsService(map[string]bool{
		EnableScopedAuthorizationKey: true,
	})

	clustersClient := cluster_service_client.NewAPIClient(&cluster_service_client.Configuration{
		Servers: cluster_service_client.ServerConfigurations{
			{
				URL: config.Get[*authorization_config.Server]().ClusterServiceUrl,
			},
		},
	})

	rolesApiService := NewRolesApiService(clustersClient, tenantSettingsService)
	accessRulesApiService := NewAccessRulesApiService(tenantSettingsService)

	service := accessRulesApiService.(AccessRulesApiServicerInner)
	service.WithTenantSettings(tenantSettingsService)

	var createdRoleId int32
	var createdAccessRules []api.AccessRule
	var createdRoles []api.Role

	// Create a role with ARReadPermission
	BeforeAll(func() {
		roleBody := api.RoleCreationFieldsInternal{
			Name:        "ARReadPermission Role",
			Description: "Role with permission to read access rules",
			Permissions: []api.PostPermission{
				{
					ResourceType: api.RESOURCETYPE_ACCESS_RULES,
					Actions: []api.Action{
						api.ACTION_READ,
					},
				},
			},
		}

		createdRoles = []api.Role{}
		createdRole := createTestRole(ctx, rolesApiService, roleBody, &createdRoles, false, http.StatusCreated)
		createdRoleId = createdRole.Id
	})

	AfterAll(func() {
		CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
		createdAccessRules = []api.AccessRule{}
		for _, role := range createdRoles {
			deleteRole, err := rolesApiService.DeleteRole(ctx, role.Id)
			Expect(err).To(BeNil())
			Expect(deleteRole.Code).To(Equal(http.StatusNoContent))
		}
		TestsCleanupAndInit()
	})

	Context("Scope hierarchy validation", func() {
		BeforeEach(func() {
			// Set up ancestors cache for all scopes using existing cluster IDs
			cluster1Id := TestCluster1Id
			cluster2Id := TestCluster2Id
			dept1Id := TestDepartmentId1
			dept2Id := TestDepartmentId2

			SetAncestorsCache(ctx, TestProjectId1, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:     tenantId,
				ClusterUuid:  &cluster1Id,
				DepartmentId: &dept1Id,
			})
			SetAncestorsCache(ctx, TestProjectId2, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:     tenantId,
				ClusterUuid:  &cluster1Id,
				DepartmentId: &dept1Id,
			})
			SetAncestorsCache(ctx, TestProjectId3, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:     tenantId,
				ClusterUuid:  &cluster2Id,
				DepartmentId: &dept2Id,
			})
			SetAncestorsCache(ctx, TestDepartmentId1, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &cluster1Id,
			})
			SetAncestorsCache(ctx, TestDepartmentId2, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &cluster2Id,
			})
			SetAncestorsCache(ctx, TestCluster1Id, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})
			SetAncestorsCache(ctx, TestCluster2Id, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			accessRulesToCreate := []api.AccessRuleCreationFields{
				{
					ScopeId:     TestProjectId1,
					ScopeType:   api.SCOPETYPE_PROJECT,
					SubjectId:   TestUser1Id,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestProjectId2,
					ScopeType:   api.SCOPETYPE_PROJECT,
					SubjectId:   TestUser1Id,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestDepartmentId1,
					ScopeType:   api.SCOPETYPE_DEPARTMENT,
					SubjectId:   TestUser1Id,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestCluster1Id,
					ScopeType:   api.SCOPETYPE_CLUSTER,
					SubjectId:   TestUser1Id,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     tenantId,
					ScopeType:   api.SCOPETYPE_TENANT,
					SubjectId:   TestUser1Id,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestProjectId3,
					ScopeType:   api.SCOPETYPE_PROJECT,
					SubjectId:   TestUser1Id,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestDepartmentId2,
					ScopeType:   api.SCOPETYPE_DEPARTMENT,
					SubjectId:   TestUser1Id,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestCluster2Id,
					ScopeType:   api.SCOPETYPE_CLUSTER,
					SubjectId:   TestUser1Id,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     tenantId,
					ScopeType:   api.SCOPETYPE_TENANT,
					SubjectId:   TestUser1Id,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
			}

			// Create all access rules
			for _, accessRuleToCreate := range accessRulesToCreate {
				createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)
			}

			SetDescendantsCache(ctx, TestDepartmentId1, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Descendants{
				ProjectIds: []string{TestProjectId1, TestProjectId2},
			})
			SetDescendantsCache(ctx, TestDepartmentId2, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Descendants{
				ProjectIds: []string{TestProjectId3},
			})
		})

		AfterEach(func() {
			CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
			createdAccessRules = []api.AccessRule{}
		})

		It("should return access rules for all scopes", func() {
			testAdminsGroups := []string{dao.SuperAdmins}
			user1Ctx := context.WithValue(setContext(tenantId, TestUser1Id, testAdminsGroups), authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
				Clusters: []string{TestCluster1Id},
			})

			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(user1Ctx, "", nil, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules

			Expect(len(accessRules)).To(Equal(4), "TestCluster1Id should have 4 access rules")
		})

		It("should return access rules for cluster department and project", func() {
			user2Ctx := context.WithValue(setContext(tenantId, TestUser1Id, []string{}), authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
				Tenants:     []string{},
				Clusters:    []string{TestCluster2Id},
				Departments: []string{TestDepartmentId2},
				Projects:    []string{},
			})

			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(user2Ctx, "", nil, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules

			Expect(len(accessRules)).To(Equal(3), "User2 should have 3 access rules")

			scopeTypes := []api.ScopeType{api.SCOPETYPE_PROJECT, api.SCOPETYPE_DEPARTMENT, api.SCOPETYPE_CLUSTER}
			scopeIds := []string{TestProjectId3, TestDepartmentId2, TestCluster2Id}

			for i, scopeType := range scopeTypes {
				found := false
				for _, rule := range accessRules {
					if rule.ScopeType == scopeType && rule.ScopeId == scopeIds[i] {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), "User2 should have access rule for scope %s with id %s", scopeType, scopeIds[i])
			}
		})

		It("should allow department-level user to see access rules for projects under their department", func() {
			user1DeptCtx := context.WithValue(setContext(tenantId, TestUser1Id, []string{}), authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
				Departments: []string{TestDepartmentId1},
			})

			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(user1DeptCtx, "", nil, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			Expect(len(accessRules)).To(Equal(3), "Should see 3 access rules for projects and department1")

			// Verify we can see both project access rules
			project1_1Found := false
			project1_2Found := false
			for _, rule := range accessRules {
				if rule.ScopeType == api.SCOPETYPE_PROJECT && rule.ScopeId == TestProjectId1 {
					project1_1Found = true
				}
				if rule.ScopeType == api.SCOPETYPE_PROJECT && rule.ScopeId == TestProjectId2 {
					project1_2Found = true
				}
			}

			Expect(project1_1Found).To(BeTrue(), "Should see access rule for 4610008")
			Expect(project1_2Found).To(BeTrue(), "Should see access rule for 4610009")
		})

		It("should allow department-level user to see access rules for projects under their department", func() {
			user1DeptCtx := context.WithValue(setContext(tenantId, TestUser1Id, []string{}), authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
				Departments: []string{TestDepartmentId2},
			})

			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(user1DeptCtx, "", nil, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			Expect(len(accessRules)).To(Equal(2), "Should see 2 access rules for projects under department1")

			project3Found := false
			department2Found := false
			for _, rule := range accessRules {
				if rule.ScopeType == api.SCOPETYPE_PROJECT && rule.ScopeId == TestProjectId3 {
					project3Found = true
				}
				if rule.ScopeType == api.SCOPETYPE_DEPARTMENT && rule.ScopeId == TestDepartmentId2 {
					department2Found = true
				}
			}

			Expect(project3Found).To(BeTrue(), "Should see access rule for TestProjectId3")
			Expect(department2Found).To(BeTrue(), "Should see access rule for TestDepartmentId2")
		})

		It("should return access rule exactly for one project", func() {
			user2DeptCtx := context.WithValue(setContext(tenantId, TestUser1Id, []string{}), authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{TestProjectId1},
			})

			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(user2DeptCtx, "", nil, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules

			Expect(len(accessRules)).To(Equal(1), "Should see access rule for project TestProjectId1")
			Expect(accessRules[0].ScopeId).To(Equal(TestProjectId1))
			Expect(accessRules[0].ScopeType).To(Equal(api.SCOPETYPE_PROJECT))
		})
	})
})
