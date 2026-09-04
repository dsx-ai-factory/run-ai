/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

var _ = Describe("Hard deletion tests", Ordered, func() {
	tenantId := testing.TestTenant
	tenantIdStr := fmt.Sprintf("%d", tenantId)
	testAdminsGroups := []string{dao.SuperAdmins}
	createdRoles := []api.Role{}
	createdAccessRules := []api.AccessRule{}
	ctx := setContext(tenantIdStr, TestUserId, testAdminsGroups)

	anotherTenantId := fmt.Sprintf("%d", testing.ManagementTestTenant)
	anotherCtx := setContext(anotherTenantId, TestUserId, testAdminsGroups)

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

	getContextByTenantId := func(tId int32) context.Context {
		if uint(tId) == testing.TestTenant {
			return ctx
		}
		if uint(tId) == testing.ManagementTestTenant {
			return anotherCtx
		}
		return nil
	}

	AfterAll(func() {
		TestsCleanupAndInit()
	})

	BeforeAll(func() {
		TestsCleanupAndInit()
	})

	AfterEach(func() {
		CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
		createdAccessRules = []api.AccessRule{}

		for _, role := range createdRoles {
			_, err := rolesApiService.DeleteRole(ctx, role.Id)
			Expect(err).To(BeNil())
		}
		createdRoles = []api.Role{}
	})

	Context("Hard access rules deletion", func() {
		BeforeEach(func() {
			accessRuleToCreate1 := api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUser3Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(dao.MapStaticRoleNameToRoleId[static_roles.DepartmentAdmin.Name]),
			}

			accessRuleToCreate2 := api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(dao.MapStaticRoleNameToRoleId[static_roles.DepartmentAdmin.Name]),
			}
			createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate1, &createdAccessRules, false, http.StatusCreated)
			createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate2, &createdAccessRules, false, http.StatusCreated)
			anotherTenantAccessRuleToCreate := api.AccessRuleCreationFields{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(dao.MapStaticRoleNameToRoleId[static_roles.DepartmentAdmin.Name]),
			}

			createTestAccessRule(anotherCtx, accessRulesApiService, anotherTenantAccessRuleToCreate, &createdAccessRules, false, http.StatusCreated)
			accessRuleEntry, _ := dao.DbGetAccessRuleById(ctx, uint(createdAccessRules[2].Id), uint(createdAccessRules[2].TenantId), nil, false, true)
			Expect(accessRuleEntry).NotTo(BeNil())
		})

		It("Hard tenant deletion", func() {
			for _, accessRule := range createdAccessRules {
				c := getContextByTenantId(accessRule.TenantId)
				accessRuleEntry, err := dao.DbGetAccessRuleById(c, uint(accessRule.Id), uint(accessRule.TenantId), nil, false, true)
				Expect(err).To(BeNil())
				Expect(accessRuleEntry).To(Not(BeNil()))
			}

			err := dao.HardDeleteTenantAccessRules(ctx, tenantIdStr)
			Expect(err).To(BeNil())

			for _, accessRule := range createdAccessRules {
				c := getContextByTenantId(accessRule.TenantId)
				accessRuleEntry, err := dao.DbGetAccessRuleById(c, uint(accessRule.Id), uint(accessRule.TenantId), nil, false, true)
				if uint(accessRule.TenantId) == testing.TestTenant {
					Expect(accessRuleEntry).To(BeNil())
					Expect(err).To(Not(BeNil()))
				} else {
					Expect(err).To(BeNil())
					Expect(accessRuleEntry).To(Not(BeNil()))
					Expect(accessRuleEntry.TenantId).To(Equal(testing.ManagementTestTenant))
				}
			}
		})
	})
})
