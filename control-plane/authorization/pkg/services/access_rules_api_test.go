/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	auditmocks "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	static_legacy_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"golang.org/x/exp/slices"
)

const (
	TestScopeId             = "1234"
	TestScope2Id            = "5678"
	TestUserId              = "user@run.ai"
	TestUser2Id             = "user2@run.ai"
	TestUser3Id             = "user3@run.ai"
	InitialAccessRulesCount = int64(2)
)

var _ = Describe("Access rules API tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	testAdminsGroups := []string{dao.SuperAdmins}
	createdRoles := []api.Role{}
	createdAccessRules := []api.AccessRule{}
	ctx := setContext(tenantId, TestUserId, testAdminsGroups)
	managementTenant := fmt.Sprintf("%d", testing.ManagementTestTenant)
	ctxManagement := context.WithValue(setContext(managementTenant, TestUserId, testAdminsGroups), runaictx.CfIsManagementTenant, "true")

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
	internalAccessRulesApiService := NewInternalApiService(tenantSettings)

	AfterAll(func() {
		TestsCleanupAndInit()
	})

	BeforeAll(func() {
		TestsCleanupAndInit()
	})

	BeforeEach(func() {
		auditMock := auditmocks.NewWriter(GinkgoT())
		auditMock.On("Write", mock.Anything).Return(nil).Maybe()
		audit.SetMock(auditMock)

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
		permissions = append(permissions, api.PostPermission{
			ResourceType: api.RESOURCETYPE_PROJECT,
			Actions:      []api.Action{api.ACTION_CREATE, api.ACTION_UPDATE, api.ACTION_DELETE, api.ACTION_READ},
		})

		roleBody := api.RoleCreationFieldsInternal{
			Name:                  "test role",
			Description:           "can fully manage projects",
			Permissions:           permissions,
			KubernetesPermissions: api.KubernetesPermissions{PredefinedRole: ptr.To(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_legacy_roles.ResearcherL1.Name]))},
		}

		createdRole := createTestRole(ctx, rolesApiService, roleBody, &createdRoles, false, http.StatusCreated)

		accessRuleBody := api.AccessRuleCreationFields{
			ScopeId:     ProjectId,
			ScopeType:   api.SCOPETYPE_PROJECT,
			SubjectId:   TestUserId,
			SubjectType: api.SUBJECTTYPE_USER,
			RoleId:      createdRole.Id,
		}

		SetAncestorsCache(ctx, ProjectId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
			TenantId:     tenantId,
			ClusterUuid:  ptr.To(TestCluster1Id),
			DepartmentId: ptr.To(DepartmentId),
		})
		SetAncestorsCache(ctx, ProjectId2, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
			TenantId:     tenantId,
			ClusterUuid:  ptr.To(TestCluster2Id),
			DepartmentId: ptr.To(DepartmentId2),
		})

		createdAccessRule := createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
		Expect(createdAccessRule.SubjectId).To(Equal(accessRuleBody.SubjectId))
		Expect(createdAccessRule.SubjectType).To(Equal(api.SubjectTypeExtended(accessRuleBody.SubjectType)))
		Expect(createdAccessRule.ScopeId).To(Equal(accessRuleBody.ScopeId))
		Expect(createdAccessRule.ScopeType).To(Equal(accessRuleBody.ScopeType))
		Expect(createdAccessRule.ScopeName).To(Equal(ProjectName))
		Expect(createdAccessRule.RoleId).To(Equal(createdRole.Id))
		Expect(createdAccessRule.RoleName).To(Equal(createdRole.Name))

		SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
			TenantId:    tenantId,
			ClusterUuid: ptr.To(TestCluster1Id),
		})
		SetAncestorsCache(ctx, DepartmentId2, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
			TenantId:    tenantId,
			ClusterUuid: ptr.To(TestCluster2Id),
		})

		SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
			TenantId: tenantId,
		})
		SetAncestorsCache(ctx, TestCluster1Id, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
			TenantId: tenantId,
		})
		SetAncestorsCache(ctx, TestCluster2Id, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
			TenantId: tenantId,
		})
	})

	AfterEach(func() {
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
	})

	Context("AccessRules", func() {
		It("Get", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			Expect(response.TotalRecords).To(Equal(int32(2)))
			Expect(response.DisplayRecords).To(Equal(int32(2)))
			Expect(len(accessRules)).To(Equal(2))
			Expect(accessRules[0].Id).To(Equal(createdAccessRules[0].Id))
			Expect(accessRules[0].SubjectId).To(Equal(createdAccessRules[0].SubjectId))
			Expect(accessRules[0].SubjectType).To(Equal(createdAccessRules[0].SubjectType))
			Expect(accessRules[0].ScopeId).To(Equal(createdAccessRules[0].ScopeId))
			Expect(accessRules[0].ScopeType).To(Equal(createdAccessRules[0].ScopeType))
			Expect(accessRules[0].RoleId).To(Equal(createdRoles[0].Id))
			Expect(accessRules[0].RoleName).To(Equal(createdRoles[0].Name))
		})

		It("Get with filter not existing", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", []string{"some-user", "some-app"}, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			Expect(response.TotalRecords).To(Equal(int32(0)))
			Expect(response.DisplayRecords).To(Equal(int32(0)))
			Expect(len(accessRules)).To(Equal(0))
		})

		It("Get with filter existing", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", []string{"some-user", "some-app", createdAccessRules[0].SubjectId}, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			Expect(response.TotalRecords).To(Equal(int32(1)))
			Expect(response.DisplayRecords).To(Equal(int32(1)))
			Expect(len(accessRules)).To(Equal(1))
			Expect(accessRules[0].Id).To(Equal(createdAccessRules[0].Id))
			Expect(accessRules[0].SubjectId).To(Equal(createdAccessRules[0].SubjectId))
			Expect(accessRules[0].SubjectType).To(Equal(createdAccessRules[0].SubjectType))
			Expect(accessRules[0].ScopeId).To(Equal(createdAccessRules[0].ScopeId))
			Expect(accessRules[0].ScopeType).To(Equal(createdAccessRules[0].ScopeType))
			Expect(accessRules[0].RoleId).To(Equal(createdRoles[0].Id))
			Expect(accessRules[0].RoleName).To(Equal(createdRoles[0].Name))
		})

		It("Get with filter subjectType", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, string(api.SUBJECTTYPE_USER), []string{}, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			Expect(accessRules).ToNot(BeEmpty())
			for _, accessRule := range accessRules {
				Expect(accessRule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
			}

			getAccessRulesResponse, err = accessRulesApiService.GetAccessRules(ctx, "TYPO", []string{}, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))
			response = httpResponseBodyToAccessRules(getAccessRulesResponse)
			Expect(response.TotalRecords).To(Equal(int32(0)))
		})

		It("Rejects the deprecated app subject type on create with 415", func() {
			createTestAccessRule(ctx, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   "deprecated-app-subject-type",
				SubjectType: api.SubjectType("app"),
				RoleId:      createdRoles[0].Id,
			}, &createdAccessRules, true, http.StatusUnsupportedMediaType)
		})

		It("Get with filter subjectId", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", []string{}, 0, 0, "", false, "", "", "", "", []string{"subjectIdFilter=@" + createdAccessRules[0].SubjectId[2:6]}, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			for _, accessRule := range accessRules {
				Expect(strings.ToLower(accessRule.SubjectId)).To(ContainSubstring(strings.ToLower(createdAccessRules[0].SubjectId[2:6])))
			}
		})

		It("Get groups", func() {
			getGroupsResponse, err := internalAccessRulesApiService.GetGroups(ctx)
			Expect(err).To(BeNil())
			Expect(getGroupsResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToGroups(getGroupsResponse)
			groups := response.Names
			Expect(len(groups)).To(Equal(1))
			Expect(groups[0]).To(Equal(dao.SuperAdmins))
		})

		It("Post", func() {
			bodyDepartment := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			createdAccessRuleDepartment := createTestAccessRule(ctx, accessRulesApiService, bodyDepartment, &createdAccessRules, false, http.StatusCreated)
			Expect(createdAccessRuleDepartment.ScopeName).To(Equal(DepartmentName))

			bodyCluster := api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			createdAccessRuleCluster := createTestAccessRule(ctx, accessRulesApiService, bodyCluster, &createdAccessRules, false, http.StatusCreated)
			Expect(createdAccessRuleCluster.ScopeName).To(Equal(ClusterName))

			bodyTenant := api.AccessRuleCreationFields{
				ScopeId:     tenantId,
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			createdAccessRuleTenant := createTestAccessRule(ctx, accessRulesApiService, bodyTenant, &createdAccessRules, false, http.StatusCreated)
			Expect(createdAccessRuleTenant.ScopeName).To(Equal(TenantName))

			bodySystem := api.AccessRuleCreationFields{
				ScopeId:     "",
				ScopeType:   api.SCOPETYPE_SYSTEM,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			// should fail since the ctx is in a tenant other than management
			createTestAccessRule(ctx, accessRulesApiService, bodySystem, &createdAccessRules, true, http.StatusForbidden)

			// should fail since the role is in other tenant than management
			createTestAccessRule(ctxManagement, accessRulesApiService, bodySystem, &createdAccessRules, true, http.StatusNotFound)

			bodySystem.RoleId = int32(dao.MapStaticRoleNameToRoleId[static_roles.DepartmentAdmin.Name])
			createdAccessRuleSystem := createTestAccessRule(ctxManagement, accessRulesApiService, bodySystem, &createdAccessRules, false, http.StatusCreated)
			Expect(createdAccessRuleSystem.ScopeType).To(Equal(api.SCOPETYPE_SYSTEM))
		})

		It("Post should fail, scope resource does not exists", func() {
			someDeletedProject := "deleted-project-id"
			// The project ancestors are still in cache, but it does not exist on the server, therefore new access rule scoped to it should not be created
			SetAncestorsCache(ctx, someDeletedProject, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			body := api.AccessRuleCreationFields{
				ScopeId:     someDeletedProject,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			createTestAccessRule(ctx, accessRulesApiService, body, &createdAccessRules, true, http.StatusForbidden)
		})

		It("Post should fail, role does not exists", func() {
			body := api.AccessRuleCreationFields{
				ScopeId:     ProjectId + "some-string",
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      123456,
			}

			createTestAccessRule(ctx, accessRulesApiService, body, &createdAccessRules, true, http.StatusNotFound)
		})

		It("Post should fail, role is disabled", func() {
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
			permissions = append(permissions, api.PostPermission{
				ResourceType: api.RESOURCETYPE_PROJECT,
				Actions:      []api.Action{api.ACTION_CREATE, api.ACTION_UPDATE, api.ACTION_DELETE, api.ACTION_READ},
			})
			roleBody := api.RoleCreationFieldsInternal{
				Name:                  "test role 2",
				Description:           "can fully manage projects2",
				Permissions:           permissions,
				Enabled:               ptr.To(false),
				KubernetesPermissions: api.KubernetesPermissions{PredefinedRole: ptr.To(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_legacy_roles.ResearcherL1.Name]))},
			}

			createdRoleDisabled := createTestRole(ctx, rolesApiService, roleBody, &createdRoles, false, http.StatusCreated)

			body := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoleDisabled.Id,
			}

			createTestAccessRule(ctx, accessRulesApiService, body, &createdAccessRules, true, http.StatusBadRequest)
		})

		It("Post should fail, invalid subjectType", func() {
			body := api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUserId,
				SubjectType: api.SubjectType("users"),
				RoleId:      createdRoles[0].Id,
			}

			createTestAccessRule(ctx, accessRulesApiService, body, &createdAccessRules, true, http.StatusBadRequest)
		})

		It("Post should fail, invalid scopeType", func() {
			body := api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.ScopeType("projects"),
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			createTestAccessRule(ctx, accessRulesApiService, body, &createdAccessRules, true, http.StatusBadRequest)
		})

		It("Post should fail, all-authenticated-users at system scope", func() {
			_, apiErr := (&AccessRulesApiService{}).createSingleAccessRule(ctxManagement, api.AccessRuleCreationFields{
				SubjectType: models.AllAuthenticatedUsers,
				ScopeType:   api.SCOPETYPE_SYSTEM,
				ScopeId:     "",
				RoleId:      int32(dao.MapStaticRoleNameToRoleId[static_roles.DepartmentAdmin.Name]),
			}, testing.ManagementTestTenant, TestUserId)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("Post should succeed with internal role", func() {
			Expect(static_roles.ClusterServiceAccount.InternalOnly).To(BeTrue())
			metricsScraper := int32(dao.MapStaticRoleNameToRoleId[static_roles.IntegrationMetricsCollector.Name])

			body := api.AccessRuleCreationFields{
				ScopeId:     tenantId,
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      metricsScraper,
			}

			createdAccessRule := createTestAccessRule(ctx, accessRulesApiService, body, &createdAccessRules, false, http.StatusCreated)
			Expect(createdAccessRule.RoleId).To(Equal(metricsScraper))
			Expect(createdAccessRule.RoleName).To(Equal(static_roles.IntegrationMetricsCollector.Name))
		})
	})

	Context("Subject id validation", func() {
		It("accepts all-authenticated-users with no subjectId", func() {
			Expect(validateSubjectId(api.AccessRuleCreationFields{SubjectType: models.AllAuthenticatedUsers})).To(BeNil())
		})

		It("rejects all-authenticated-users with a subjectId", func() {
			apiErr := validateSubjectId(api.AccessRuleCreationFields{SubjectType: models.AllAuthenticatedUsers, SubjectId: "someone@run.ai"})
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("accepts a regular subject type with a subjectId", func() {
			Expect(validateSubjectId(api.AccessRuleCreationFields{SubjectType: api.SUBJECTTYPE_USER, SubjectId: "alice@run.ai"})).To(BeNil())
		})

		It("requires a subjectId for non-all-authenticated types", func() {
			for _, subjectType := range []api.SubjectType{api.SUBJECTTYPE_USER, api.SUBJECTTYPE_GROUP, api.SUBJECTTYPE_SERVICE_ACCOUNT} {
				apiErr := validateSubjectId(api.AccessRuleCreationFields{SubjectType: subjectType})
				Expect(apiErr).ToNot(BeNil(), "subjectId should be required for %s", subjectType)
				Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
			}
		})

		It("rejects the reserved sentinel for non-all-authenticated types", func() {
			for _, subjectType := range []api.SubjectType{api.SUBJECTTYPE_USER, api.SUBJECTTYPE_GROUP, api.SUBJECTTYPE_SERVICE_ACCOUNT} {
				apiErr := validateSubjectId(api.AccessRuleCreationFields{SubjectType: subjectType, SubjectId: models.AllAuthenticatedUsers})
				Expect(apiErr).ToNot(BeNil(), "reserved sentinel should be rejected for %s", subjectType)
				Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
			}
		})
	})

	Context("Access rules by ID", func() {
		It("Get", func() {
			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))

			accessRule := HttpResponseBodyToAccessRuleById(getAccessRuleResponse)
			Expect(accessRule.Id).To(Equal(createdAccessRules[0].Id))
			Expect(accessRule.SubjectId).To(Equal(createdAccessRules[0].SubjectId))
			Expect(accessRule.SubjectType).To(Equal(createdAccessRules[0].SubjectType))
			Expect(accessRule.ScopeId).To(Equal(createdAccessRules[0].ScopeId))
			Expect(accessRule.ScopeType).To(Equal(createdAccessRules[0].ScopeType))
			Expect(accessRule.RoleId).To(Equal(createdRoles[0].Id))
			Expect(accessRule.RoleName).To(Equal(createdRoles[0].Name))
			Expect(accessRule.ScopePath).To(Not(BeNil()))
			Expect(*accessRule.ScopePath).To(Equal(fmt.Sprintf("%s/%s/%s/%s", TenantName, ClusterName, DepartmentName, ProjectName)))
		})

		It("Get should fail, not found", func() {
			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, math.MaxInt32)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("Delete", func() {
			deletionResponse, err := accessRulesApiService.DeleteAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))
		})

		It("Delete role with access rule should fail on 400", func() {
			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))

			deletionResponse, err := rolesApiService.DeleteRole(ctx, createdRoles[0].Id)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("Delete should fail, not found", func() {
			deletionResponse, err := accessRulesApiService.DeleteAccessRule(ctx, math.MaxInt32)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNotFound))
		})
	})
	Context("Subject access rules", func() {
		BeforeEach(func() {
			// create another access-rule with the same subject-type and subject-id
			accessRuleToCreate := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   createdAccessRules[0].SubjectId,
				SubjectType: api.SubjectType(createdAccessRules[0].SubjectType),
				RoleId:      createdAccessRules[0].RoleId,
			}
			createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)
		})

		It("Get", func() {
			getResponse, err := internalAccessRulesApiService.GetSubjectAccessRules(ctx, string(createdAccessRules[0].SubjectType), []string{createdAccessRules[0].SubjectId})
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			accessRules := httpResponseBodyToAccessRulesSlice(getResponse)
			Expect(len(accessRules)).To(BeNumerically(">=", 1))
			for _, accessRule := range accessRules {
				Expect(accessRule.SubjectType).To(Equal(createdAccessRules[0].SubjectType))
				Expect(accessRule.SubjectId).To(Equal(createdAccessRules[0].SubjectId))
			}
		})

		It("Wrong subject id no deletion", func() {
			deletionResponse, err := internalAccessRulesApiService.DeleteSubjectAccessRules(ctx, string(createdAccessRules[0].SubjectType), "some-string", false)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))
		})

		It("Wrong subject type no deletion", func() {
			deletionResponse, err := internalAccessRulesApiService.DeleteSubjectAccessRules(ctx, string(api.SUBJECTTYPE_SERVICE_ACCOUNT), createdAccessRules[0].SubjectId, false)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))
		})

		It("Rejects the deprecated app subject type on internal create with 415", func() {
			createResponse, err := internalAccessRulesApiService.CreateInternalAccessRule(ctx, api.InternalCreateAccessRuleRequest{
				SubjectId:   "deprecated-app-internal",
				SubjectType: api.SubjectType("app"),
				RoleName:    api.ALLOWEDINTERNALCREATEROLES_CLUSTER_SERVICE_ACCOUNT,
			})
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusUnsupportedMediaType))
		})

		It("Delete by the deprecated app subject type removes the service-account rules", func() {
			rule := createTestAccessRule(ctx, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   "sa-delete-by-deprecated-app",
				SubjectType: api.SUBJECTTYPE_SERVICE_ACCOUNT,
				RoleId:      createdAccessRules[0].RoleId,
			}, &createdAccessRules, false, http.StatusCreated)

			deletionResponse, err := internalAccessRulesApiService.DeleteSubjectAccessRules(ctx, "app", "sa-delete-by-deprecated-app", false)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			getResponse, err := accessRulesApiService.GetAccessRule(ctx, rule.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("Delete subject access rules", func() {
			deletionResponse, err := internalAccessRulesApiService.DeleteSubjectAccessRules(ctx, string(createdAccessRules[0].SubjectType), createdAccessRules[0].SubjectId, false)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			// should delete both created access-rules, they had the same subject-type and subject-id
			for _, createdAccessRule := range createdAccessRules {
				getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRule.Id)
				Expect(err).To(BeNil())
				Expect(getAccessRuleResponse.Code).To(Equal(http.StatusNotFound))
			}
		})
	})
	Context("Subject access rules deletion privilege escalation", func() {
		var toDeleteAccessRule *api.AccessRule
		BeforeEach(func() {
			accessRuleToCreate := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(dao.MapStaticRoleNameToRoleId[static_roles.DepartmentAdmin.Name]),
			}
			toDeleteAccessRule = createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)
		})

		It("Delete with no permissions", func() {
			requesterCtx := setContext(tenantId, "some-user", []string{})
			deletionResponse, err := internalAccessRulesApiService.DeleteSubjectAccessRules(requesterCtx, string(api.SUBJECTTYPE_USER), TestUser2Id, false)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusForbidden))
		})

		It("Delete with missing permissions", func() {
			accessRuleToCreate := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUser3Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(dao.MapStaticRoleNameToRoleId[static_roles.Editor.Name]),
			}
			createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)
			requesterCtx := setContext(tenantId, TestUser3Id, []string{})
			deletionResponse, err := internalAccessRulesApiService.DeleteSubjectAccessRules(requesterCtx, string(api.SUBJECTTYPE_USER), TestUser2Id, false)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusForbidden))
		})

		It("Delete in lower scope", func() {
			accessRuleToCreate := api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser3Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(dao.MapStaticRoleNameToRoleId[static_roles.DepartmentAdmin.Name]),
			}
			createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)
			requesterCtx := setContext(tenantId, TestUser3Id, []string{})
			deletionResponse, err := internalAccessRulesApiService.DeleteSubjectAccessRules(requesterCtx, string(api.SUBJECTTYPE_USER), TestUser2Id, false)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusForbidden))
		})

		It("Delete upper scope should pass, with dry run", func() {
			accessRuleToCreate := api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUser3Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(dao.MapStaticRoleNameToRoleId[static_roles.DepartmentAdmin.Name]),
			}
			createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)
			requesterCtx := setContext(tenantId, TestUser3Id, []string{})
			deletionResponse, err := internalAccessRulesApiService.DeleteSubjectAccessRules(requesterCtx, string(api.SUBJECTTYPE_USER), TestUser2Id, true)
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, toDeleteAccessRule.Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))
		})
	})
	Context("Scope access rules deletion", func() {
		BeforeEach(func() {
			// create another access-rule with the same scope-type and scope-id
			accessRuleToCreate := api.AccessRuleCreationFields{
				ScopeId:     createdAccessRules[0].ScopeId,
				ScopeType:   createdAccessRules[0].ScopeType,
				SubjectId:   TestGroupId,
				SubjectType: api.SUBJECTTYPE_GROUP,
				RoleId:      createdAccessRules[0].RoleId,
			}
			createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)
			// create another access-rule for app
			accessRuleToCreate2 := api.AccessRuleCreationFields{
				ScopeId:     createdAccessRules[0].ScopeId,
				ScopeType:   createdAccessRules[0].ScopeType,
				SubjectId:   TestAppId,
				SubjectType: api.SUBJECTTYPE_SERVICE_ACCOUNT,
				RoleId:      createdAccessRules[0].RoleId,
			}
			createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate2, &createdAccessRules, false, http.StatusCreated)
		})

		It("delete last system admin access-rule", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			adminAccessRules := []string{}
			var adminAccessRule int32
			for i := range accessRules {
				if accessRules[i].RoleName == static_roles.SystemAdmin.Name {
					adminAccessRules = append(adminAccessRules, fmt.Sprint(accessRules[i].Id))
					adminAccessRule = accessRules[i].Id
				}
			}

			if len(adminAccessRules) == 1 {
				deletionResponse, err := accessRulesApiService.DeleteAccessRule(ctx, adminAccessRule)
				Expect(err).To(BeNil())
				Expect(deletionResponse.Code).To(Equal(http.StatusBadRequest))
			}

			deletionBatchValidateResponse, err := accessRulesApiService.AccessRulesBatch(ctx, api.AccessRulesBatchFields{
				Ids:    adminAccessRules,
				Action: api.ACCESSRULESBATCHACTION_VALIDATE_DELETE,
			})
			Expect(err).To(BeNil())
			Expect(deletionBatchValidateResponse.Code).To(Equal(http.StatusOK))
			batchRes := deletionBatchValidateResponse.Body.(api.BatchResponse)
			Expect(len(batchRes.Succeeded)).To(Equal(0))
			Expect(len(batchRes.Failed)).To(Equal(len(adminAccessRules)))
			for _, failed := range batchRes.Failed {
				Expect(failed.Code).To(Equal(int32(http.StatusBadRequest)))
			}

			deletionBatchResponse, err := accessRulesApiService.AccessRulesBatch(ctx, api.AccessRulesBatchFields{
				Ids:    adminAccessRules,
				Action: api.ACCESSRULESBATCHACTION_DELETE,
			})
			Expect(err).To(BeNil())
			batchRes = deletionBatchResponse.Body.(api.BatchResponse)
			Expect(len(batchRes.Succeeded)).To(Equal(0))
			Expect(len(batchRes.Failed)).To(Equal(len(adminAccessRules)))
			for _, failed := range batchRes.Failed {
				Expect(failed.Code).To(Equal(int32(http.StatusBadRequest)))
			}

			accessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     TestScopeId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),
			}
			// Track this access rule in case test fails before deletion
			tmpAccessRules := []api.AccessRule{}
			created := createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &tmpAccessRules, false, http.StatusCreated)

			mixedRules := []string{}
			mixedRules = append(mixedRules, adminAccessRules[0])
			createdId := fmt.Sprint(created.Id)
			mixedRules = append(mixedRules, createdId)
			deletionBatchValidateResponse, err = accessRulesApiService.AccessRulesBatch(ctx, api.AccessRulesBatchFields{
				Ids:    mixedRules,
				Action: api.ACCESSRULESBATCHACTION_VALIDATE_DELETE,
			})
			Expect(err).To(BeNil())
			Expect(deletionBatchValidateResponse.Code).To(Equal(http.StatusOK))
			batchRes = deletionBatchValidateResponse.Body.(api.BatchResponse)
			Expect(len(batchRes.Succeeded)).To(Equal(1))
			Expect(len(batchRes.Failed)).To(Equal(1))
			Expect(batchRes.Succeeded[0]).To(Equal(createdId))
			Expect(batchRes.Failed[0].Id).To(Equal(adminAccessRules[0]))

			deletionBatchResponse, err = accessRulesApiService.AccessRulesBatch(ctx, api.AccessRulesBatchFields{
				Ids:    mixedRules,
				Action: api.ACCESSRULESBATCHACTION_DELETE,
			})
			Expect(err).To(BeNil())
			batchRes = deletionBatchResponse.Body.(api.BatchResponse)
			Expect(len(batchRes.Succeeded)).To(Equal(1))
			Expect(len(batchRes.Failed)).To(Equal(1))
			Expect(batchRes.Succeeded[0]).To(Equal(createdId))
			Expect(batchRes.Failed[0].Id).To(Equal(adminAccessRules[0]))
		})

		It("Wrong scope id no deletion", func() {
			deletionResponse, err := internalAccessRulesApiService.DeleteScopeAccessRules(ctx, string(createdAccessRules[0].ScopeType), "some-string", "", "")
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))
		})

		It("Wrong scope type no deletion", func() {
			deletionResponse, err := internalAccessRulesApiService.DeleteScopeAccessRules(ctx, string(api.SUBJECTTYPE_USER), createdAccessRules[0].ScopeId, "", "")
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))
		})

		It("Delete scope access roles", func() {
			deletionResponse, err := internalAccessRulesApiService.DeleteScopeAccessRules(ctx, string(createdAccessRules[0].ScopeType), createdAccessRules[0].ScopeId, "", "")
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			// should delete both created access-rules, they had the same scope-type and scope-id
			for _, createdAccessRule := range createdAccessRules {
				getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRule.Id)
				Expect(err).To(BeNil())
				Expect(getAccessRuleResponse.Code).To(Equal(http.StatusNotFound))
			}
		})

		It("Delete with current timestamp should not delete subject access rules", func() {
			deletionResponse, err := internalAccessRulesApiService.DeleteScopeAccessRules(ctx, string(createdAccessRules[0].ScopeType), createdAccessRules[0].ScopeId, FormatTime(time.Now().AddDate(0, 0, -1)), "")
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))
		})
		It("Delete with old timestamp should delete subject access rules", func() {
			deletionResponse, err := internalAccessRulesApiService.DeleteScopeAccessRules(ctx, string(createdAccessRules[0].ScopeType), createdAccessRules[0].ScopeId, FormatTime(time.Now()), "")
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			// should delete both created access-rules, they had the same subject-type and subject-id
			for _, createdAccessRule := range createdAccessRules {
				getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRule.Id)
				Expect(err).To(BeNil())
				Expect(getAccessRuleResponse.Code).To(Equal(http.StatusNotFound))
			}
		})
		It("Delete with wrong roles should not delete subject access rules", func() {
			deletionResponse, err := internalAccessRulesApiService.DeleteScopeAccessRules(ctx, string(createdAccessRules[0].ScopeType), createdAccessRules[0].ScopeId, "", "not-existing-role,viewer,editor")
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))
		})
		It("Delete with roles should delete subject access rules", func() {
			deletionResponse, err := internalAccessRulesApiService.DeleteScopeAccessRules(ctx, string(createdAccessRules[0].ScopeType), createdAccessRules[0].ScopeId, "", fmt.Sprintf("not-existing-role,viewer,editor,%s", createdAccessRules[0].RoleName))
			Expect(err).To(BeNil())
			Expect(deletionResponse.Code).To(Equal(http.StatusNoContent))

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusNotFound))
		})
	})
	Context("Access rules by scope ids", func() {
		BeforeEach(func() {
			accessRule2Body := api.AccessRuleCreationFields{
				ScopeId:     TestScopeId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),
			}
			accessRule3Body := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}
			accessRule4Body := api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUser3Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			SetAncestorsCache(ctx, TestScopeId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			createTestAccessRule(ctx, accessRulesApiService, accessRule2Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule3Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule4Body, &createdAccessRules, false, http.StatusCreated)
		})

		It("Get ok", func() {
			getAccessRuleResponse, err := internalAccessRulesApiService.GetScopeAccessRules(ctx, "", static_roles.DepartmentAdmin.Name, "", TestScopeId)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))

			accessRulesBody := httpResponseBodyToAccessRules(getAccessRuleResponse)
			Expect(int(accessRulesBody.TotalRecords)).To(Equal(1))
			rule := accessRulesBody.AccessRules[0]
			Expect(rule.ScopeId).To(Equal(TestScopeId))
			Expect(rule.ScopeType).To(Equal(api.SCOPETYPE_PROJECT))
			Expect(rule.RoleName).To(Equal(static_roles.DepartmentAdmin.Name))
		})

		It("Get 4 scopes", func() {
			getAccessRuleResponse, err := internalAccessRulesApiService.GetScopeAccessRules(ctx, ClusterId, "", DepartmentId, ProjectId+","+TestScopeId)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))

			accessRulesBody := httpResponseBodyToAccessRules(getAccessRuleResponse)
			Expect(int(accessRulesBody.TotalRecords)).To(Equal(5))
			accessRules := accessRulesBody.AccessRules
			for _, rule := range accessRules {
				switch rule.ScopeType {
				case api.SCOPETYPE_TENANT:
					Expect(rule.ScopeId).To(Equal(tenantId))
				case api.SCOPETYPE_CLUSTER:
					Expect(rule.ScopeId).To(Equal(ClusterId))
				case api.SCOPETYPE_DEPARTMENT:
					Expect(rule.ScopeId).To(Equal(DepartmentId))
				case api.SCOPETYPE_PROJECT:
					Expect(slices.Contains([]string{TestScopeId, ProjectId}, rule.ScopeId)).To(BeTrue())
				default:
					Fail("got wrong accessRules")
				}
			}
		})

		It("Get should return 0 records no scopeId", func() {
			getAccessRuleResponse, err := internalAccessRulesApiService.GetScopeAccessRules(ctx, "", static_roles.DepartmentAdmin.Name, "", "lala")
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))
			accessRulesBody := httpResponseBodyToAccessRules(getAccessRuleResponse)
			Expect(int(accessRulesBody.TotalRecords)).To(Equal(0))
		})
	})
	Context("Access rules with subject type filter", func() {
		BeforeEach(func() {
			accessRule2Body := api.AccessRuleCreationFields{
				ScopeId:     TestScopeId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),
			}
			accessRule3Body := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}
			accessRule4Body := api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUser3Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			SetAncestorsCache(ctx, TestScopeId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			createTestAccessRule(ctx, accessRulesApiService, accessRule2Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule3Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule4Body, &createdAccessRules, false, http.StatusCreated)
		})

		It("Get one rule", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", "", "", []string{"subjectIdFilter=@2@run."}, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			Expect(len(accessRules)).To(Equal(1))
			Expect(accessRules[0].SubjectId).To(Equal(TestUser2Id))
			Expect(accessRules[0].SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
		})

		It("Get 2 rules", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", "", "", []string{"subjectIdFilter=@r@run"}, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			Expect(len(accessRules)).To(Equal(2))
			Expect(accessRules[0].SubjectId).To(Equal(TestUserId))
			Expect(accessRules[0].SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
			Expect(accessRules[1].SubjectId).To(Equal(TestUserId))
			Expect(accessRules[1].SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
		})

		It("Get 0 rules", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", "", "", []string{"subjectIdFilter=@lulu@run"}, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			Expect(len(accessRules)).To(Equal(0))
		})
	})
	Context("Access rules with scope filter", func() {
		BeforeEach(func() {
			accessRule2Body := api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),
			}
			accessRule3Body := api.AccessRuleCreationFields{
				ScopeId:     TestScopeId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),
			}
			accessRule4Body := api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),
			}

			SetAncestorsCache(ctx, ProjectId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, TestScopeId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			createTestAccessRule(ctx, accessRulesApiService, accessRule2Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule3Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule4Body, &createdAccessRules, false, http.StatusCreated)
		})

		It("Get one rule", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", TestScopeId, "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			Expect(len(accessRules)).To(Equal(1))
			Expect(accessRules[0].ScopeId).To(Equal(TestScopeId))
		})

		It("Get 2 rules", func() {
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", "", "", []string{"scopeType==" + string(api.SCOPETYPE_PROJECT)}, "")
			getAccessRulesResponseAll, err := accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			responseAll := httpResponseBodyToAccessRules(getAccessRulesResponseAll)
			Expect(responseAll).To(Not(BeNil()))
			accessRules := response.AccessRules
			Expect(len(accessRules)).To(Equal(3))
			Expect(accessRules[0].ScopeType).To(Equal(api.SCOPETYPE_PROJECT))
			Expect(accessRules[1].ScopeType).To(Equal(api.SCOPETYPE_PROJECT))
			Expect(accessRules[2].ScopeType).To(Equal(api.SCOPETYPE_PROJECT))
		})
	})
	Context("Scope access rules update", func() {
		It("Update scope access rules", func() {
			request := api.UpdateScopeAccessRulesRequest{ScopeName: "new-name"}
			response, err := internalAccessRulesApiService.UpdateScopeAccessRules(ctx, string(createdAccessRules[0].ScopeType), createdAccessRules[0].ScopeId, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRules[0].Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusOK))

			newAccessRuleFromServer := HttpResponseBodyToAccessRule(getAccessRuleResponse)
			Expect(newAccessRuleFromServer.ScopeName).To(Equal(request.ScopeName))
		})
	})
	Context("Get access rules filter by", func() {
		BeforeEach(func() {
			accessRule2Body := api.AccessRuleCreationFields{
				ScopeId:     TestScopeId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),
			}
			accessRule3Body := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}
			accessRule4Body := api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUser3Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			SetAncestorsCache(ctx, TestScopeId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			createTestAccessRule(ctx, accessRulesApiService, accessRule2Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule3Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule4Body, &createdAccessRules, false, http.StatusCreated)
		})
		It("Filters by roleId", func() {
			filters := []string{"roleId==" + strconv.Itoa(int(RolesMap[static_roles.DepartmentAdmin.Name]))}
			testAccessRulesFilter(ctx, accessRulesApiService, "", "", filters, 1, false, "")
		})

		It("Filters by scopeType", func() {
			filters := []string{string("scopeType==" + api.SCOPETYPE_PROJECT)}
			testAccessRulesFilter(ctx, accessRulesApiService, "", "", filters, 2, false, string(api.SCOPETYPE_PROJECT))
		})

		It("Filters by scopeType", func() {
			filters := []string{string("scopeType==" + api.SCOPETYPE_TENANT)}
			testAccessRulesFilter(ctx, accessRulesApiService, "asc", "scopeType", filters, 1, false, string(api.SCOPETYPE_TENANT))
		})

		It("Filters by subjectId", func() {
			filters := []string{"subjectIdFilter==" + TestUserId}
			testAccessRulesFilter(ctx, accessRulesApiService, "", "scopeName", filters, 2, false, TestUserId)
		})

		It("Filters by scopeName", func() {
			filters := []string{"scopeName==" + ClusterName}
			testAccessRulesFilter(ctx, accessRulesApiService, "desc", "scopeName", filters, 1, false, ClusterName)
		})

		It("Should fail when sortBy is invalid", func() {
			testAccessRulesFilter(ctx, accessRulesApiService, "desc", "role", nil, 1, true, "")
		})

		It("Should fail when filter is invalid", func() {
			filters := []string{"scope-name==" + ClusterName}
			testAccessRulesFilter(ctx, accessRulesApiService, "", "", filters, 0, true, ClusterName)
		})

		It("Should fail when operator is invalid", func() {
			filters := []string{"scopeName===" + ClusterName}
			testAccessRulesFilter(ctx, accessRulesApiService, "", "", filters, 0, true, ClusterName)
		})

		It("Should get access rules filter by created_at", func() {
			filters := []string{fmt.Sprintf("createdAt>=%s", FormatTime(time.Now().Add(1*time.Minute)))}
			testAccessRulesFilter(ctx, accessRulesApiService, "", "", filters, 0, false, "")
		})

		It("should get access rules filter by created_at", func() {
			filters := []string{fmt.Sprintf("createdAt<=%s", FormatTime(time.Now().Add(1*time.Minute)))}
			testAccessRulesFilter(ctx, accessRulesApiService, "", "", filters, 5, false, "")
		})
	})
	Context("Get access rules sync", func() {
		BeforeEach(func() {
			accessRule2Body := api.AccessRuleCreationFields{
				ScopeId:     TestScopeId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),
			}
			accessRule3Body := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}
			accessRule4Body := api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUser3Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}
			accessRule5Body := api.AccessRuleCreationFields{
				ScopeId:     TestScopeId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.AiPractitioner.Name]),
			}
			accessRule6Body := api.AccessRuleCreationFields{
				ScopeId:     TestScopeId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.ProjectAdmin.Name]),
			}
			accessRule7Body := api.AccessRuleCreationFields{
				ScopeId:     TestScopeId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DataAndStorageAdmin.Name]),
			}

			SetAncestorsCache(ctx, TestScopeId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			createTestAccessRule(ctx, accessRulesApiService, accessRule2Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule3Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule4Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule5Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule6Body, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRule7Body, &createdAccessRules, false, http.StatusCreated)
		})
		It("get sync version = 2.24", func() {
			apiResponse, _ := accessRulesApiService.GetAccessRulesSync(ctx, 0, 0, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			accessRuleResponse := httpResponseBodyToAccessRulesSync(apiResponse)
			accessRules := accessRuleResponse.AccessRules
			for _, rule := range accessRules {
				if rule.RoleId == createdRoles[0].Id {
					Expect(rule.PredefinedRole).NotTo(BeNil())
					Expect((*rule.PredefinedRole).Id).To(Equal(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_legacy_roles.ResearcherL1.Name])))
					Expect((*rule.PredefinedRole).Name).To(Equal(static_legacy_roles.ResearcherL1.Name))
				} else {
					Expect(rule.PredefinedRole).To(BeNil())
				}

				Expect(rule.IsDeleting).To(BeFalse())
			}
		})
		It("get sync version < 2.24", func() {
			apiResponse, _ := accessRulesApiService.GetAccessRulesSync(ctx, 0, 0, "", ClusterId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			accessRuleResponse := httpResponseBodyToAccessRulesSync(apiResponse)
			accessRules := accessRuleResponse.AccessRules
			for _, rule := range accessRules {
				if rule.RoleId == createdRoles[0].Id {
					Expect(rule.PredefinedRole).NotTo(BeNil())
					Expect((*rule.PredefinedRole).Id).To(Equal(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_legacy_roles.ResearcherL1.Name])))
					Expect((*rule.PredefinedRole).Name).To(Equal(static_legacy_roles.ResearcherL1.Name))
				} else {
					mappedRule, found := NewToLegacyRolesMap[rule.RoleName]
					if found {
						Expect(rule.PredefinedRole).NotTo(BeNil())
						Expect((*rule.PredefinedRole).Name).To(Equal(mappedRule))
						Expect((*rule.PredefinedRole).Id).To(Equal(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[mappedRule])))
					} else {
						Expect(rule.PredefinedRole).To(BeNil())
					}
				}

				Expect(rule.IsDeleting).To(BeFalse())
			}
		})
	})
	Context("Count access rules", func() {
		testCases := []struct {
			description   string
			subjectType   api.SubjectType
			expectedCount int64
		}{
			{
				description:   "initial count",
				subjectType:   api.SUBJECTTYPE_USER,
				expectedCount: InitialAccessRulesCount,
			},
			{
				description:   "count after adding a new access rule",
				subjectType:   api.SUBJECTTYPE_GROUP,
				expectedCount: 3,
			},
		}
		var createdAccessRule *api.AccessRule
		for _, tc := range testCases {
			It(fmt.Sprintf("should return correct count for %s", tc.description), func() {
				if tc.subjectType == api.SUBJECTTYPE_GROUP {
					accessRuleToCreate := api.AccessRuleCreationFields{
						ScopeId:     createdAccessRules[0].ScopeId,
						ScopeType:   createdAccessRules[0].ScopeType,
						SubjectId:   TestGroupId,
						SubjectType: tc.subjectType,
						RoleId:      createdAccessRules[0].RoleId,
					}
					createdAccessRule = createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)
				}
				getCountAccessRulesResponse, err := accessRulesApiService.CountAccessRules(ctx, false, nil, "")
				Expect(err).To(BeNil())
				Expect(getCountAccessRulesResponse.Code).To(Equal(http.StatusOK))

				response := HttpResponseBodyToCountAccessRules(getCountAccessRulesResponse)
				Expect(response.Count).To(Equal(tc.expectedCount))
			})
		}

		It("should delete access rule and count correctly", func() {
			_, err := accessRulesApiService.DeleteAccessRule(ctx, createdAccessRule.Id)
			Expect(err).To(BeNil())

			getCountAccessRulesResponse, err := accessRulesApiService.CountAccessRules(ctx, false, nil, "")
			Expect(err).To(BeNil())
			Expect(getCountAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := HttpResponseBodyToCountAccessRules(getCountAccessRulesResponse)
			Expect(response.Count).To(Equal(InitialAccessRulesCount))
		})

		It("should return bad request", func() {
			getCountAccessRulesResponse, err := accessRulesApiService.CountAccessRules(ctx, false, []string{"badFilter"}, "")
			Expect(err).To(BeNil())
			Expect(getCountAccessRulesResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("should create access rule and get count filter by subjectType==app", func() {
			accessRuleToCreate := api.AccessRuleCreationFields{
				ScopeId:     createdAccessRules[0].ScopeId,
				ScopeType:   createdAccessRules[0].ScopeType,
				SubjectId:   TestGroupId,
				SubjectType: api.SUBJECTTYPE_SERVICE_ACCOUNT,
				RoleId:      createdAccessRules[0].RoleId,
			}
			createdAccessRule = createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)

			getCountAccessRulesResponse, err := accessRulesApiService.CountAccessRules(ctx, false, []string{fmt.Sprintf("subjectType==%s", api.SUBJECTTYPE_SERVICE_ACCOUNT)}, "")
			Expect(err).To(BeNil())
			Expect(getCountAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := HttpResponseBodyToCountAccessRules(getCountAccessRulesResponse)
			Expect(response.Count).To(Equal(int64(1)))
		})

		It("should create access rules and get access rules include deleted", func() {
			accessRuleToCreate := api.AccessRuleCreationFields{
				ScopeId:     createdAccessRules[0].ScopeId,
				ScopeType:   createdAccessRules[0].ScopeType,
				SubjectId:   TestGroupId,
				SubjectType: api.SUBJECTTYPE_SERVICE_ACCOUNT,
				RoleId:      createdAccessRules[0].RoleId,
			}
			createdAccessRule = createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)

			_, err := accessRulesApiService.DeleteAccessRule(ctx, createdAccessRule.Id)
			Expect(err).To(BeNil())

			getCountAccessRulesResponse, err := accessRulesApiService.CountAccessRules(ctx, true, nil, "")
			Expect(err).To(BeNil())
			Expect(getCountAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := HttpResponseBodyToCountAccessRules(getCountAccessRulesResponse)
			Expect(response.Count).To(BeNumerically(">=", 28))
		})
	})
	Context("Create/Get/Delete access rules - test phase and status fields", func() {
		It("create/get/update phase - test cases", func() {
			accessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			createdAccessRule := createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
			Expect(createdAccessRule.Status.Phase).To(Equal(api.ACCESSRULEPHASE_CREATING))
			Expect(createdAccessRule.Status.Message).To(Equal(""))
			Expect(createdAccessRule.Status.Details).To(Equal(""))

			getAccessRuleAndValidateStatus(ctx, accessRulesApiService, createdAccessRule.Id,
				api.ACCESSRULEPHASE_CREATING, "", "")

			// Update phase and status
			updateStatusRequest := api.AccessRuleStatusRequest{
				Phase:   api.ACCESSRULECLUSTERPHASE_READY,
				Message: "Access rule is ready",
			}
			response, err := accessRulesApiService.UpdateAccessRuleStatus(ctx, createdAccessRule.Id, TestCluster1Id, updateStatusRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			// Verify the update
			getAccessRuleAndValidateStatus(ctx, accessRulesApiService, createdAccessRule.Id,
				api.ACCESSRULEPHASE_READY, "Access rule is ready", "")

			updateStatusRequest = api.AccessRuleStatusRequest{
				Phase:   api.ACCESSRULECLUSTERPHASE_NOT_READY,
				Message: "Failed creating RoleBinding in namespace 'runai-test' for department-a scope",
			}
			response, err = accessRulesApiService.UpdateAccessRuleStatus(ctx, createdAccessRule.Id, TestCluster1Id, updateStatusRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			getAccessRuleAndValidateStatus(ctx, accessRulesApiService, createdAccessRule.Id,
				api.ACCESSRULEPHASE_NOT_READY, "Failed creating RoleBinding in namespace 'runai-test' for department-a scope", "")

			// delete the access rule - validate deleted (no "Deleting" logic at this time)
			_, err = accessRulesApiService.DeleteAccessRule(ctx, createdAccessRule.Id)
			Expect(err).To(BeNil())

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRule.Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("create/get/update + GetAccessRules - test cases", func() {
			// create several accessrules, update their phase, then Get and validate
			accessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			users := []string{TestUserId, TestUser2Id, TestUser3Id, "user4@run.ai"}
			readyAccessRulesIds := []int32{}
			notReadyAccessRulesIds := []int32{}
			for i := 0; i < 4; i++ {
				accessRuleBody.SubjectId = users[i]
				createdAccessRule := createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
				Expect(createdAccessRule.Status.Phase).To(Equal(api.ACCESSRULEPHASE_CREATING))

				// Update phase and status
				phase := api.AccessRuleClusterPhase("")
				if i%2 == 0 {
					phase = api.ACCESSRULECLUSTERPHASE_NOT_READY
					notReadyAccessRulesIds = append(notReadyAccessRulesIds, createdAccessRule.Id)
				} else {
					phase = api.ACCESSRULECLUSTERPHASE_READY
					readyAccessRulesIds = append(readyAccessRulesIds, createdAccessRule.Id)
				}

				updateStatusRequest := api.AccessRuleStatusRequest{
					Phase:   phase,
					Message: fmt.Sprintf("Access rule is %v", phase),
				}
				response, err := accessRulesApiService.UpdateAccessRuleStatus(ctx, createdAccessRule.Id, TestCluster1Id, updateStatusRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))
			}

			// Get all access rules
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			// Validate
			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules
			readyFound := 0
			notReadyFound := 0
			for _, rule := range accessRules {
				if slices.Contains(readyAccessRulesIds, rule.Id) {
					Expect(rule.Status.Phase).To(Equal(api.ACCESSRULEPHASE_READY))
					Expect(rule.Status.Message).To(Equal("Access rule is Ready"))
					readyFound += 1
				} else if slices.Contains(notReadyAccessRulesIds, rule.Id) {
					Expect(rule.Status.Phase).To(Equal(api.ACCESSRULEPHASE_NOT_READY))
					Expect(rule.Status.Message).To(Equal("Access rule is NotReady"))
					notReadyFound += 1
				}
			}
			Expect(readyFound).To(Equal(2))
			Expect(notReadyFound).To(Equal(2))

			// GetAccessRules with phase filter and validate
			getAccessRulesResponse, err = accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", "", "", []string{"phase==Ready"}, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))
			response = httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules = response.AccessRules
			Expect(len(accessRules)).To(Equal(2))
			for _, rule := range accessRules {
				Expect(rule.Status.Phase).To(Equal(api.ACCESSRULEPHASE_READY))
				Expect(rule.Status.Message).To(Equal("Access rule is Ready"))
			}
		})

		It("create/get in old cluster - no phase", func() {
			accessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId2,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			createdAccessRule := createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
			Expect(createdAccessRule.Status.Phase).To(Equal(api.AccessRulePhase("")))
			Expect(createdAccessRule.Status.Message).To(Equal(""))
			Expect(createdAccessRule.Status.Details).To(Equal(""))

			getAccessRuleAndValidateStatus(ctx, accessRulesApiService, createdAccessRule.Id,
				api.AccessRulePhase(""), "", "")

			// delete the access rule - validate deleted (no "Deleting" logic at this time)
			_, err := accessRulesApiService.DeleteAccessRule(ctx, createdAccessRule.Id)
			Expect(err).To(BeNil())

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRule.Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("create/get/update phase with access rule in tenant scope", func() {
			accessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     tenantId,
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			createdAccessRule := createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
			Expect(createdAccessRule.Status.Phase).To(Equal(api.AccessRulePhase("")))
			Expect(createdAccessRule.Status.Message).To(Equal(""))
			Expect(createdAccessRule.Status.Details).To(Equal(""))

			getAccessRuleAndValidateStatus(ctx, accessRulesApiService, createdAccessRule.Id,
				api.AccessRulePhase(""), "", "")

			// Update phase and status
			updateStatusRequest := api.AccessRuleStatusRequest{
				Phase:   api.ACCESSRULECLUSTERPHASE_READY,
				Message: "Access rule is ready",
			}
			response, err := accessRulesApiService.UpdateAccessRuleStatus(ctx, createdAccessRule.Id, TestCluster1Id, updateStatusRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))
			getAccessRuleAndValidateStatus(ctx, accessRulesApiService, createdAccessRule.Id,
				api.ACCESSRULEPHASE_READY, "", fmt.Sprintf("Ready: [%s]", TestCluster1Id))

			// update from different cluster
			updateStatusRequest = api.AccessRuleStatusRequest{
				Phase:   api.ACCESSRULECLUSTERPHASE_NOT_READY,
				Message: "Access rule is not ready",
			}
			response, err = accessRulesApiService.UpdateAccessRuleStatus(ctx, createdAccessRule.Id, TestCluster2Id, updateStatusRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))
			getAccessRuleAndValidateStatus(ctx, accessRulesApiService, createdAccessRule.Id,
				api.ACCESSRULEPHASE_NOT_READY, "", fmt.Sprintf("NotReady: [%s]; Ready: [%s]", TestCluster2Id, TestCluster1Id))

			// validate GetAccessRules as well
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))
			accessRules := httpResponseBodyToAccessRules(getAccessRulesResponse).AccessRules
			found := false
			for _, rule := range accessRules {
				if rule.Id == createdAccessRule.Id {
					Expect(rule.Status.Phase).To(Equal(api.ACCESSRULEPHASE_NOT_READY))
					Expect(rule.Status.Message).To(Equal(""))
					Expect(rule.Status.Details).To(Equal(fmt.Sprintf("NotReady: [%s]; Ready: [%s]", TestCluster2Id, TestCluster1Id)))
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "Access rule not found in GetAccessRules response")

			// update from different cluster
			updateStatusRequest = api.AccessRuleStatusRequest{
				Phase:   api.ACCESSRULECLUSTERPHASE_READY,
				Message: "Access rule is ready",
			}
			response, err = accessRulesApiService.UpdateAccessRuleStatus(ctx, createdAccessRule.Id, ClusterId, updateStatusRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))
			getAccessRuleAndValidateStatus(ctx, accessRulesApiService, createdAccessRule.Id,
				api.ACCESSRULEPHASE_NOT_READY, "", fmt.Sprintf("NotReady: [%s]; Ready: [%s, %s]", TestCluster2Id, ClusterId, TestCluster1Id))

			// delete the access rule - validate deleted (no "Deleting" logic at this time)
			_, err = accessRulesApiService.DeleteAccessRule(ctx, createdAccessRule.Id)
			Expect(err).To(BeNil())

			getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, createdAccessRule.Id)
			Expect(err).To(BeNil())
			Expect(getAccessRuleResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("forward compatibility - isDeleting should be true in getSync for agent", func() {
			// at this point, manually update the phase to "Deleting"
			// and validate isDeleting is true

			accessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     TestCluster1Id,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRoles[0].Id,
			}

			createdAccessRule := createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
			Expect(createdAccessRule.Status.Phase).To(Equal(api.ACCESSRULEPHASE_CREATING))
			Expect(createdAccessRule.Status.Message).To(Equal(""))
			Expect(createdAccessRule.Status.Details).To(Equal(""))

			updateStatusRequest := api.AccessRuleStatusRequest{
				Phase:   api.ACCESSRULECLUSTERPHASE_READY,
				Message: "Access rule is ready",
			}
			response, err := accessRulesApiService.UpdateAccessRuleStatus(ctx, createdAccessRule.Id, TestCluster1Id, updateStatusRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			// "manually" update the phase to "Deleting"
			// (in the future it will happen when use sends "Delete" request)
			updateStatusRequest = api.AccessRuleStatusRequest{
				Phase: api.AccessRuleClusterPhase(api.ACCESSRULEPHASE_DELETING),
			}
			dbErr := dao.DbUpdateAccessRulePhaseAndMessage(ctx, uint(createdAccessRule.Id), testing.TestTenant, TestCluster1Id, updateStatusRequest)
			Expect(dbErr).To(BeNil())

			// validate isDeleting is true in get sync access rules
			apiResponse, _ := accessRulesApiService.GetAccessRulesSync(ctx, 0, 0, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			accessRuleResponse := httpResponseBodyToAccessRulesSync(apiResponse)
			accessRules := accessRuleResponse.AccessRules
			found := false
			for _, accessRule := range accessRules {
				if accessRule.Id == createdAccessRule.Id {
					Expect(accessRule.IsDeleting).To(BeTrue())
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "Access rule not found in get sync response")
		})
	})
})

func getAccessRuleAndValidateStatus(ctx context.Context, accessRulesApiService api.AccessRulesAPIServicer, accessRuleId int32,
	expectedPhase api.AccessRulePhase, expectedMessage string, expectedDetails string,
) {
	getAccessRuleResponse, err := accessRulesApiService.GetAccessRule(ctx, accessRuleId)
	ExpectWithOffset(1, err).To(BeNil(), "GetAccessRule error")
	ExpectWithOffset(1, getAccessRuleResponse.Code).To(Equal(http.StatusOK), "GetAccessRule code")

	updatedAccessRule := HttpResponseBodyToAccessRuleById(getAccessRuleResponse)
	ExpectWithOffset(1, updatedAccessRule.Status.Phase).To(Equal(expectedPhase), "Phase")
	ExpectWithOffset(1, updatedAccessRule.Status.Message).To(Equal(expectedMessage), "Phase Message")
	ExpectWithOffset(1, updatedAccessRule.Status.Details).To(Equal(expectedDetails), "Phase Details")
}

func createTestAccessRule(ctx context.Context, accessRulesApiService api.AccessRulesAPIServicer, creationRequest api.AccessRuleCreationFields, createdAccessRules *[]api.AccessRule, shouldFail bool, expectedResponse int) *api.AccessRule {
	createdAccessRuleResponse, err := accessRulesApiService.CreateAccessRule(ctx, creationRequest)
	Expect(err).To(BeNil())

	Expect(createdAccessRuleResponse.Code).To(Equal(expectedResponse))

	if shouldFail {
		return nil
	}

	accessRule := HttpResponseBodyToAccessRule(createdAccessRuleResponse)

	*createdAccessRules = append(*createdAccessRules, accessRule)
	return &accessRule
}

func CleanupAccessRules(accessRulesApiService api.AccessRulesAPIServicer, ctx context.Context, accessRules []api.AccessRule) {
	for _, accessRule := range accessRules {
		_, err := accessRulesApiService.DeleteAccessRule(ctx, accessRule.Id)
		Expect(err).To(BeNil())
	}
}

func HttpResponseBodyToAccessRule(httpResponse api.ImplResponse) api.AccessRule {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	accessRule := &api.AccessRule{}
	err = json.Unmarshal(bytes, accessRule)
	Expect(err).To(BeNil())
	return *accessRule
}

func HttpResponseBodyToAccessRuleById(httpResponse api.ImplResponse) api.AccessRuleById {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	accessRule := &api.AccessRuleById{}
	err = json.Unmarshal(bytes, accessRule)
	Expect(err).To(BeNil())
	return *accessRule
}

func httpResponseBodyToAccessRules(httpResponse api.ImplResponse) api.GetAccessRules200Response {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	accessRules := &api.GetAccessRules200Response{}
	err = json.Unmarshal(bytes, accessRules)
	Expect(err).To(BeNil())
	return *accessRules
}

func httpResponseBodyToAccessRulesSync(httpResponse api.ImplResponse) api.GetAccessRulesSync200Response {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	accessRules := &api.GetAccessRulesSync200Response{}
	err = json.Unmarshal(bytes, accessRules)
	Expect(err).To(BeNil())
	return *accessRules
}

func HttpResponseBodyToCountAccessRules(httpResponse api.ImplResponse) api.CountAccessRules200Response {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	count := &api.CountAccessRules200Response{}
	err = json.Unmarshal(bytes, count)
	Expect(err).To(BeNil())
	return *count
}

func httpResponseBodyToAccessRulesSlice(httpResponse api.ImplResponse) []api.AccessRule {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	accessRules := &[]api.AccessRule{}
	err = json.Unmarshal(bytes, accessRules)
	Expect(err).To(BeNil())
	return *accessRules
}

func httpResponseBodyToGroups(response api.ImplResponse) api.GetGroups200Response {
	bytes, err := json.Marshal(response.Body)
	Expect(err).To(BeNil())

	groups := &api.GetGroups200Response{}
	err = json.Unmarshal(bytes, groups)
	Expect(err).To(BeNil())
	return *groups
}

func testAccessRulesFilter(ctx context.Context, accessRulesService api.AccessRulesAPIServicer, sortOrder string, sortBy string, filters []string, expectedCount int, shouldFail bool, search string) {
	apiResponse, _ := accessRulesService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", sortOrder, api.AccessRulesSortFilterFields(sortBy), filters, "")

	if shouldFail {
		Expect(apiResponse.Code).NotTo(Equal(http.StatusOK))
		return
	}

	Expect(apiResponse.Code).To(Equal(http.StatusOK))
	accessRuleResponse := httpResponseBodyToAccessRules(apiResponse)
	Expect(len(accessRuleResponse.AccessRules)).To(Equal(expectedCount))

	apiResponse, _ = accessRulesService.CountAccessRules(ctx, false, filters, "")
	Expect(apiResponse.Code).To(Equal(http.StatusOK))
	countResponse := HttpResponseBodyToCountAccessRules(apiResponse)
	Expect(countResponse.Count).To(Equal(int64(expectedCount)))

	if search != "" {
		testAccessRulesSearch(ctx, accessRulesService, sortOrder, sortBy, expectedCount, shouldFail, search)
	}
}

func testAccessRulesSearch(ctx context.Context, accessRulesService api.AccessRulesAPIServicer, sortOrder string, sortBy string, expectedCount int, shouldFail bool, search string) {
	apiResponse, _ := accessRulesService.GetAccessRules(ctx, "", nil, 0, 0, "", false, "", "", sortOrder, api.AccessRulesSortFilterFields(sortBy), nil, search)

	if shouldFail {
		Expect(apiResponse.Code).NotTo(Equal(http.StatusOK))
		return
	}

	Expect(apiResponse.Code).To(Equal(http.StatusOK))
	accessRuleResponse := httpResponseBodyToAccessRules(apiResponse)
	Expect(len(accessRuleResponse.AccessRules)).To(BeNumerically(">=", expectedCount))

	apiResponse, _ = accessRulesService.CountAccessRules(ctx, false, nil, search)
	Expect(apiResponse.Code).To(Equal(http.StatusOK))
	countResponse := HttpResponseBodyToCountAccessRules(apiResponse)
	Expect(countResponse.Count).To(Equal(int64(len(accessRuleResponse.AccessRules))))
}

var _ = Describe("Legacy subjectType query parameter", func() {
	Describe("filterByBackwardsCompatibility", func() {
		It("should convert the legacy subjectType parameter to a filterBy entry", func() {
			result := filterByBackwardsCompatibility([]string{}, "service-account")
			Expect(result).To(ConsistOf("subjectType==service-account"))
		})

		It("should keep existing filters and append the legacy parameter", func() {
			result := filterByBackwardsCompatibility([]string{"scopeType==project"}, "user")
			Expect(result).To(ConsistOf("scopeType==project", "subjectType==user"))
		})

		It("should leave filters unchanged when there is no legacy parameter", func() {
			result := filterByBackwardsCompatibility([]string{"subjectType!=group"}, "")
			Expect(result).To(Equal([]string{"subjectType!=group"}))
		})
	})
})

var _ = Describe("isDeprecatedSubjectType", func() {
	It("matches only the deprecated app subject type", func() {
		Expect(isDeprecatedSubjectType(api.SubjectType("app"))).To(BeTrue())
		Expect(isDeprecatedSubjectType("app")).To(BeTrue())
		for _, subjectType := range []api.SubjectType{api.SUBJECTTYPE_USER, api.SUBJECTTYPE_SERVICE_ACCOUNT, api.SUBJECTTYPE_GROUP} {
			Expect(isDeprecatedSubjectType(subjectType)).To(BeFalse())
		}
	})

	It("is reported as 415 Unsupported Media Type pointing at service-account", func() {
		apiErr := deprecatedSubjectTypeError()
		Expect(apiErr.Code).To(Equal(int32(http.StatusUnsupportedMediaType)))
		Expect(apiErr.Message).To(ContainSubstring("service-account"))
	})
})
