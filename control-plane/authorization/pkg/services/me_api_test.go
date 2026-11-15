/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	static_legacy_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
)

var _ = Describe("MeApi", Ordered, func() {
	const TestGroup2 = "group-b"
	const TestGroup3 = "group-c"

	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	createdAccessRules := []api.AccessRule{}
	ctx := setContext(tenantId, TestUserId, []string{dao.SuperAdmins})

	tenantSettings := NewMockTenantSettingsService(map[string]bool{})

	meApiService := NewMeApiService()
	accessRulesApiService := NewAccessRulesApiService(tenantSettings)
	internalApiService := NewInternalApiService(tenantSettings)

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

		accessRulesToCreate := []api.AccessRuleCreationFields{
			{
				ScopeId:     fmt.Sprintf("%d", testing.TestTenant),
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUser3Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),
			},
			{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestGroup3,
				SubjectType: api.SUBJECTTYPE_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.ResearcherL1.Name]),
			},

			{
				ScopeId:     fmt.Sprintf("%d", testing.TestTenant),
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_legacy_roles.MLEngineer.Name]),
			},
			{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestGroup2,
				SubjectType: api.SUBJECTTYPE_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.EnvironmentAdmin.Name]),
			},
			{
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestGroup2,
				SubjectType: api.SUBJECTTYPE_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.ResearchManager.Name]),
			},
		}

		for _, accessRuleBody := range accessRulesToCreate {
			createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
		}

		// create an internal role for user2, just to validate we don't see it in the "get access rules" response.
		response, err := internalApiService.CreateInternalAccessRule(ctx, api.InternalCreateAccessRuleRequest{
			SubjectId:   TestUser2Id,
			SubjectType: api.SUBJECTTYPE_USER,
			RoleName:    api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT,
		})
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusCreated))
	})

	AfterAll(func() {
		_, err := internalApiService.DeleteSubjectAccessRules(ctx, string(api.SUBJECTTYPE_USER), TestUser2Id, false)
		Expect(err).To(BeNil())

		CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
		createdAccessRules = []api.AccessRule{}
		TestsCleanupAndInit()
	})

	It("Test Get me/accessrules for user 3", func() {
		// user 3 should have 2 access rules
		testCtx := setContext(tenantId, TestUser3Id, []string{TestGroup3})

		getAccessRulesResponse, err := meApiService.GetMeAccessRules(testCtx, 0, 0, "")
		Expect(err).To(BeNil())
		Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

		response := httpResponseBodyToMeAccessRules(getAccessRulesResponse)
		meAccessRules := response.AccessRules.AccessRules
		Expect(len(meAccessRules)).To(Equal(2))
		Expect(response.AccessRules.TotalRecords).To(BeEquivalentTo(2))
		Expect(response.AccessRules.DisplayRecords).To(BeEquivalentTo(2))

		expectedMeAccessRules := map[int32]api.MeAccessRule{
			createdAccessRules[0].Id: {
				ScopeId:     fmt.Sprintf("%d", testing.TestTenant),
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUser3Id,
				SubjectType: api.SUBJECTTYPEEXTENDED_USER,
				RoleId:      int32(RolesMap[static_roles.DepartmentAdmin.Name]),

				RoleName:  static_roles.DepartmentAdmin.Name,
				ScopeName: "Tenant-1",
				Id:        createdAccessRules[0].Id,
				TenantId:  int32(testing.TestTenant),
			},
			createdAccessRules[1].Id: {
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestGroup3,
				SubjectType: api.SUBJECTTYPEEXTENDED_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.ResearcherL1.Name]),

				ClusterId: ClusterId,
				RoleName:  static_legacy_roles.ResearcherL1.Name,
				ScopeName: "Project-1",
				Id:        createdAccessRules[1].Id,
				TenantId:  int32(testing.TestTenant),
			},
		}

		for _, meAccessRule := range meAccessRules {
			Expect(expectedMeAccessRules).To(HaveKey(meAccessRule.Id))
			Expect(meAccessRule).To(Equal(expectedMeAccessRules[meAccessRule.Id]))
		}
	})

	It("Test Get me/accessrules for user 2", func() {
		// user 2 should have 3 access rules
		testCtx := setContext(tenantId, TestUser2Id, []string{TestGroup2})

		getAccessRulesResponse, err := meApiService.GetMeAccessRules(testCtx, 0, 0, "")
		Expect(err).To(BeNil())
		Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

		response := httpResponseBodyToMeAccessRules(getAccessRulesResponse)
		meAccessRules := response.AccessRules.AccessRules
		Expect(len(meAccessRules)).To(Equal(3))
		Expect(response.AccessRules.TotalRecords).To(BeEquivalentTo(3))
		Expect(response.AccessRules.DisplayRecords).To(BeEquivalentTo(3))

		expectedMeAccessRules := map[int32]api.MeAccessRule{
			createdAccessRules[2].Id: {
				ScopeId:     fmt.Sprintf("%d", testing.TestTenant),
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPEEXTENDED_USER,
				RoleId:      int32(RolesMap[static_legacy_roles.MLEngineer.Name]),

				RoleName:  static_legacy_roles.MLEngineer.Name,
				ScopeName: "Tenant-1",
				Id:        createdAccessRules[2].Id,
				TenantId:  int32(testing.TestTenant),
			},
			createdAccessRules[3].Id: {
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestGroup2,
				SubjectType: api.SUBJECTTYPEEXTENDED_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.EnvironmentAdmin.Name]),

				ClusterId: ClusterId,
				RoleName:  static_legacy_roles.EnvironmentAdmin.Name,
				ScopeName: "Department-1",
				Id:        createdAccessRules[3].Id,
				TenantId:  int32(testing.TestTenant),
			},
			createdAccessRules[4].Id: {
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestGroup2,
				SubjectType: api.SUBJECTTYPEEXTENDED_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.ResearchManager.Name]),

				ClusterId: ClusterId,
				RoleName:  static_legacy_roles.ResearchManager.Name,
				ScopeName: "Cluster-1",
				Id:        createdAccessRules[4].Id,
				TenantId:  int32(testing.TestTenant),
			},
		}

		for _, meAccessRule := range meAccessRules {
			Expect(expectedMeAccessRules).To(HaveKey(meAccessRule.Id))
			Expect(meAccessRule).To(Equal(expectedMeAccessRules[meAccessRule.Id]))
		}
	})

	It("Test Get me/accessrules with limit and offset", func() {
		// user 2 should have 3 access rules
		testCtx := setContext(tenantId, TestUser2Id, []string{TestGroup2})

		getAccessRulesResponse, err := meApiService.GetMeAccessRules(testCtx, 1, 1, "")
		Expect(err).To(BeNil())
		Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

		response := httpResponseBodyToMeAccessRules(getAccessRulesResponse)
		meAccessRules := response.AccessRules.AccessRules
		Expect(len(meAccessRules)).To(Equal(1))
		Expect(response.AccessRules.TotalRecords).To(BeEquivalentTo(1))
		Expect(response.AccessRules.DisplayRecords).To(BeEquivalentTo(1))

		expectedMeAccessRules := map[int32]api.MeAccessRule{
			createdAccessRules[3].Id: {
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestGroup2,
				SubjectType: api.SUBJECTTYPEEXTENDED_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.EnvironmentAdmin.Name]),

				ClusterId: ClusterId,
				RoleName:  static_legacy_roles.EnvironmentAdmin.Name,
				ScopeName: "Department-1",
				Id:        createdAccessRules[3].Id,
				TenantId:  int32(testing.TestTenant),
			},
		}

		for _, meAccessRule := range meAccessRules {
			Expect(expectedMeAccessRules).To(HaveKey(meAccessRule.Id))
			Expect(meAccessRule).To(Equal(expectedMeAccessRules[meAccessRule.Id]))
		}
	})

	It("Get me/accessrules with search", func() {
		testCtx := setContext(tenantId, TestUser2Id, []string{TestGroup2})

		getAccessRulesResponse, err := meApiService.GetMeAccessRules(testCtx, 0, 0, "Cluster-1")
		Expect(err).To(BeNil())
		Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

		response := httpResponseBodyToMeAccessRules(getAccessRulesResponse)
		meAccessRules := response.AccessRules.AccessRules
		Expect(len(meAccessRules)).To(Equal(1))
		Expect(response.AccessRules.TotalRecords).To(BeEquivalentTo(1))
		Expect(response.AccessRules.DisplayRecords).To(BeEquivalentTo(1))

		expectedMeAccessRules := map[int32]api.MeAccessRule{
			createdAccessRules[4].Id: {
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestGroup2,
				SubjectType: api.SUBJECTTYPEEXTENDED_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.ResearchManager.Name]),

				ClusterId: ClusterId,
				RoleName:  static_legacy_roles.ResearchManager.Name,
				ScopeName: "Cluster-1",
				Id:        createdAccessRules[4].Id,
				TenantId:  int32(testing.TestTenant),
			},
		}

		for _, meAccessRule := range meAccessRules {
			Expect(expectedMeAccessRules).To(HaveKey(meAccessRule.Id))
			Expect(meAccessRule).To(Equal(expectedMeAccessRules[meAccessRule.Id]))
		}
	})

	It("Get me/accessrules with search on role name", func() {
		testCtx := setContext(tenantId, TestUser2Id, []string{TestGroup2})

		getAccessRulesResponse, err := meApiService.GetMeAccessRules(testCtx, 0, 0, "research")
		Expect(err).To(BeNil())
		Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

		// validate we only get 1 access rule,
		// and that we don't get the other access rule that is assigned to a different user -
		// with role "L2 researcher".
		response := httpResponseBodyToMeAccessRules(getAccessRulesResponse)
		meAccessRules := response.AccessRules.AccessRules
		Expect(len(meAccessRules)).To(Equal(1))
		Expect(response.AccessRules.TotalRecords).To(BeEquivalentTo(1))
		Expect(response.AccessRules.DisplayRecords).To(BeEquivalentTo(1))

		expectedMeAccessRules := map[int32]api.MeAccessRule{
			createdAccessRules[4].Id: {
				ScopeId:     ClusterId,
				ScopeType:   api.SCOPETYPE_CLUSTER,
				SubjectId:   TestGroup2,
				SubjectType: api.SUBJECTTYPEEXTENDED_GROUP,
				RoleId:      int32(RolesMap[static_legacy_roles.ResearchManager.Name]),

				ClusterId: ClusterId,
				RoleName:  static_legacy_roles.ResearchManager.Name,
				ScopeName: "Cluster-1",
				Id:        createdAccessRules[4].Id,
				TenantId:  int32(testing.TestTenant),
			},
		}

		for _, meAccessRule := range meAccessRules {
			Expect(expectedMeAccessRules).To(HaveKey(meAccessRule.Id))
			Expect(meAccessRule).To(Equal(expectedMeAccessRules[meAccessRule.Id]))
		}
	})

	It("Test Get me/accessrules/count for user 3", func() {
		// user 3 should have 2 access rules
		testCtx := setContext(tenantId, TestUser3Id, []string{TestGroup3})

		countAccessRulesResponse, err := meApiService.CountMeAccessRules(testCtx, "")
		Expect(err).To(BeNil())
		Expect(countAccessRulesResponse.Code).To(Equal(http.StatusOK))

		response := HttpResponseBodyToCountAccessRules(countAccessRulesResponse)
		Expect(response.Count).To(BeEquivalentTo(2))
	})

	It("Test Get me/accessrules/count for user 2", func() {
		// user 2 should have 3 access rules
		testCtx := setContext(tenantId, TestUser2Id, []string{TestGroup2})

		countAccessRulesResponse, err := meApiService.CountMeAccessRules(testCtx, "")
		Expect(err).To(BeNil())
		Expect(countAccessRulesResponse.Code).To(Equal(http.StatusOK))

		response := HttpResponseBodyToCountAccessRules(countAccessRulesResponse)
		Expect(response.Count).To(BeEquivalentTo(3))
	})

	It("Test Get me/accessrules/count with search", func() {
		// user 2 should have 3 access rules
		testCtx := setContext(tenantId, TestUser2Id, []string{TestGroup2})

		countAccessRulesResponse, err := meApiService.CountMeAccessRules(testCtx, "Cluster-1")
		Expect(err).To(BeNil())
		Expect(countAccessRulesResponse.Code).To(Equal(http.StatusOK))

		response := HttpResponseBodyToCountAccessRules(countAccessRulesResponse)
		Expect(response.Count).To(BeEquivalentTo(1))
	})
})

func httpResponseBodyToMeAccessRules(httpResponse api.ImplResponse) api.MeAccessRulesResponse {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	accessRules := &api.MeAccessRulesResponse{}
	err = json.Unmarshal(bytes, accessRules)
	Expect(err).To(BeNil())
	return *accessRules
}
