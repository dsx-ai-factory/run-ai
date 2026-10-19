/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	static_legacy_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
)

var _ = Describe("Internal create access rules API tests", Ordered, func() {
	testSubjectId := "some-unique-id"
	createdAccessRules := []api.AccessRule{}
	adminCtx := setContext(fmt.Sprintf("%d", testing.TestTenant), TestUserId, []string{dao.SuperAdmins})
	userCtx := setContext(fmt.Sprintf("%d", testing.TestTenant), testSubjectId, []string{})

	tenantSettings := NewMockTenantSettingsService(map[string]bool{})

	accessRulesApiService := NewAccessRulesApiService(tenantSettings)
	internalApiService := NewInternalApiService(tenantSettings)

	AfterAll(func() {
		TestsCleanupAndInit()
	})

	Context("Admin user", func() {
		AfterEach(func() {
			CleanupAccessRules(accessRulesApiService, adminCtx, createdAccessRules)
			createdAccessRules = []api.AccessRule{}
		})

		It("Post", func() {
			response, err := internalApiService.CreateInternalAccessRule(userCtx, api.InternalCreateAccessRuleRequest{
				SubjectId:   testSubjectId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleName:    api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			accessRule := HttpResponseBodyToAccessRule(response)
			createdAccessRules = append(createdAccessRules, accessRule)

			Expect(accessRule.TenantId).To(Equal(int32(testing.TestTenant)))
			Expect(accessRule.SubjectId).To(Equal(testSubjectId))
			Expect(accessRule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
			Expect(accessRule.ScopeType).To(Equal(api.SCOPETYPE_TENANT))
			Expect(accessRule.RoleId).To(Equal(int32(dao.MapStaticRoleNameToRoleId[static_roles.SystemAdmin.Name])))
			Expect(accessRule.ScopeName).To(Equal(TenantName))
		})

		It("Returns 409 on duplicate", func() {
			request := api.InternalCreateAccessRuleRequest{
				SubjectId:   testSubjectId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleName:    api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR,
			}

			response, err := internalApiService.CreateInternalAccessRule(userCtx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			createdAccessRules = append(createdAccessRules, HttpResponseBodyToAccessRule(response))

			response, err = internalApiService.CreateInternalAccessRule(userCtx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusConflict))
		})

		It("Admin user has permissions", func() {
			response, err := internalApiService.CreateInternalAccessRule(userCtx, api.InternalCreateAccessRuleRequest{
				SubjectId:   testSubjectId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleName:    api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			accessRule := HttpResponseBodyToAccessRule(response)
			createdAccessRules = append(createdAccessRules, accessRule)

			createTestAccessRule(userCtx, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     fmt.Sprintf("%d", testing.TestTenant),
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      int32(RolesMap[static_legacy_roles.Researcher.Name]),
			}, &createdAccessRules, false, http.StatusCreated)
		})
	})
	Context("Cluster app", func() {
		AfterEach(func() {
			TestsCleanupAndInit()
		})

		It("Post without scope", func() {
			response, err := internalApiService.CreateInternalAccessRule(userCtx, api.InternalCreateAccessRuleRequest{
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleName:    api.ALLOWEDINTERNALCREATEROLES_CLUSTER_SERVICE_ACCOUNT,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			accessRule := HttpResponseBodyToAccessRule(response)

			Expect(accessRule.TenantId).To(Equal(int32(testing.TestTenant)))
			Expect(accessRule.SubjectId).To(Equal(TestUserId))
			Expect(accessRule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
			Expect(accessRule.ScopeType).To(Equal(api.SCOPETYPE_TENANT))
			Expect(accessRule.ScopeName).To(Equal(TenantName))
			Expect(accessRule.RoleId).To(Equal(int32(dao.MapStaticRoleNameToRoleId[static_roles.ClusterServiceAccount.Name])))
		})

		It("Post with cluster scope", func() {
			response, err := internalApiService.CreateInternalAccessRule(userCtx, api.InternalCreateAccessRuleRequest{
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleName:    api.ALLOWEDINTERNALCREATEROLES_CLUSTER_SERVICE_ACCOUNT,
				Scope: &api.InternalCreateAccessRuleRequestScope{
					Type: api.SCOPETYPE_CLUSTER,
					Id:   TestCluster1Id,
				},
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			accessRule := HttpResponseBodyToAccessRule(response)

			Expect(accessRule.TenantId).To(Equal(int32(testing.TestTenant)))
			Expect(accessRule.SubjectId).To(Equal(TestUserId))
			Expect(accessRule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
			Expect(accessRule.ScopeType).To(Equal(api.SCOPETYPE_CLUSTER))
			Expect(accessRule.ScopeId).To(Equal(TestCluster1Id))
			Expect(accessRule.RoleId).To(Equal(int32(dao.MapStaticRoleNameToRoleId[static_roles.ClusterServiceAccount.Name])))
		})

		It("Post with tenant scope", func() {
			response, err := internalApiService.CreateInternalAccessRule(userCtx, api.InternalCreateAccessRuleRequest{
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleName:    api.ALLOWEDINTERNALCREATEROLES_CLUSTER_SERVICE_ACCOUNT,
				Scope: &api.InternalCreateAccessRuleRequestScope{
					Type: api.SCOPETYPE_TENANT,
					Id:   fmt.Sprintf("%d", testing.TestTenant),
				},
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			accessRule := HttpResponseBodyToAccessRule(response)

			Expect(accessRule.TenantId).To(Equal(int32(testing.TestTenant)))
			Expect(accessRule.SubjectId).To(Equal(TestUserId))
			Expect(accessRule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
			Expect(accessRule.ScopeType).To(Equal(api.SCOPETYPE_TENANT))
			Expect(accessRule.ScopeId).To(Equal(fmt.Sprintf("%d", testing.TestTenant)))
			Expect(accessRule.RoleId).To(Equal(int32(dao.MapStaticRoleNameToRoleId[static_roles.ClusterServiceAccount.Name])))
		})

		When("Post with project scope", func() {
			It("Fails", func() {
				response, err := internalApiService.CreateInternalAccessRule(userCtx, api.InternalCreateAccessRuleRequest{
					SubjectId:   TestUserId,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleName:    api.ALLOWEDINTERNALCREATEROLES_CLUSTER_SERVICE_ACCOUNT,
					Scope: &api.InternalCreateAccessRuleRequestScope{
						Type: api.SCOPETYPE_PROJECT,
						Id:   TestProject2Id,
					},
				})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Context("Installer app", func() {
		AfterEach(func() {
			TestsCleanupAndInit()
		})

		It("Post without scope", func() {
			response, err := internalApiService.CreateInternalAccessRule(userCtx, api.InternalCreateAccessRuleRequest{
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleName:    api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			accessRule := HttpResponseBodyToAccessRule(response)

			Expect(accessRule.TenantId).To(Equal(int32(testing.TestTenant)))
			Expect(accessRule.SubjectId).To(Equal(TestUserId))
			Expect(accessRule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
			Expect(accessRule.ScopeType).To(Equal(api.SCOPETYPE_TENANT))
			Expect(accessRule.ScopeName).To(Equal(TenantName))
			Expect(accessRule.RoleId).To(Equal(int32(dao.MapStaticRoleNameToRoleId[static_roles.InstallerServiceAccount.Name])))
		})

		It("Post with cluster scope", func() {
			response, err := internalApiService.CreateInternalAccessRule(userCtx, api.InternalCreateAccessRuleRequest{
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleName:    api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT,
				Scope: &api.InternalCreateAccessRuleRequestScope{
					Type: api.SCOPETYPE_CLUSTER,
					Id:   TestCluster1Id,
				},
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			accessRule := HttpResponseBodyToAccessRule(response)

			Expect(accessRule.TenantId).To(Equal(int32(testing.TestTenant)))
			Expect(accessRule.SubjectId).To(Equal(TestUserId))
			Expect(accessRule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
			Expect(accessRule.ScopeType).To(Equal(api.SCOPETYPE_CLUSTER))
			Expect(accessRule.ScopeId).To(Equal(TestCluster1Id))
			Expect(accessRule.RoleId).To(Equal(int32(dao.MapStaticRoleNameToRoleId[static_roles.InstallerServiceAccount.Name])))
		})

		It("Post with tenant scope", func() {
			response, err := internalApiService.CreateInternalAccessRule(userCtx, api.InternalCreateAccessRuleRequest{
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleName:    api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT,
				Scope: &api.InternalCreateAccessRuleRequestScope{
					Type: api.SCOPETYPE_TENANT,
					Id:   fmt.Sprintf("%d", testing.TestTenant),
				},
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			accessRule := HttpResponseBodyToAccessRule(response)

			Expect(accessRule.TenantId).To(Equal(int32(testing.TestTenant)))
			Expect(accessRule.SubjectId).To(Equal(TestUserId))
			Expect(accessRule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
			Expect(accessRule.ScopeType).To(Equal(api.SCOPETYPE_TENANT))
			Expect(accessRule.ScopeId).To(Equal(fmt.Sprintf("%d", testing.TestTenant)))
			Expect(accessRule.RoleId).To(Equal(int32(dao.MapStaticRoleNameToRoleId[static_roles.InstallerServiceAccount.Name])))
		})

		When("Post with project scope", func() {
			It("Fails", func() {
				response, err := internalApiService.CreateInternalAccessRule(userCtx, api.InternalCreateAccessRuleRequest{
					SubjectId:   TestUserId,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleName:    api.ALLOWEDINTERNALCREATEROLES_CLUSTER_SERVICE_ACCOUNT,
					Scope: &api.InternalCreateAccessRuleRequestScope{
						Type: api.SCOPETYPE_PROJECT,
						Id:   TestProject2Id,
					},
				})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
