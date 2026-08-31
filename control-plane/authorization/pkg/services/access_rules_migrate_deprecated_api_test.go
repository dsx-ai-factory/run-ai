/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	auditmocks "github.com/run-ai/runai/control-plane/audit-service/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	authorizationutils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"
	static_legacy_roles "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	testing "github.com/run-ai/runai/control-plane/authorization/pkg/test"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

var _ = Describe("Access rules migrate deprecated API tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	testAdminsGroups := []string{dao.SuperAdmins}
	ctx := setContext(tenantId, TestUserId, testAdminsGroups)

	var createdRoles []api.Role
	var createdAccessRules []api.AccessRule
	var deprecatedRole *api.Role
	var targetRole *api.Role

	tenantSettings := NewMockTenantSettingsService(map[string]bool{})

	clustersClient := cluster_service_client.NewAPIClient(&cluster_service_client.Configuration{
		Servers: cluster_service_client.ServerConfigurations{
			{URL: config.Get[*authorization_config.Server]().ClusterServiceUrl},
		},
	})

	rolesApiService := NewRolesApiService(clustersClient, tenantSettings)
	accessRulesApiService := NewAccessRulesApiService(tenantSettings)

	setRoleDeprecated := func(roleId int32, deprecated bool) {
		_, err := runaidb.DB().Model(&models.Role{}).Where("id = ?", roleId).Update("deprecated", deprecated)
		Expect(err).To(BeNil())
	}

	setRoleEnabled := func(roleId int32, enabled bool) {
		_, err := runaidb.DB().Model(&models.Role{}).Where("id = ?", roleId).Update("enabled", enabled)
		Expect(err).To(BeNil())
	}

	createAccessRule := func(subjectId string, subjectType api.SubjectType, scopeId string, scopeType api.ScopeType, roleId int32) *api.AccessRule {
		body := api.AccessRuleCreationFields{
			ScopeId:     scopeId,
			ScopeType:   scopeType,
			SubjectId:   subjectId,
			SubjectType: subjectType,
			RoleId:      roleId,
		}
		return createTestAccessRule(ctx, accessRulesApiService, body, &createdAccessRules, false, http.StatusCreated)
	}

	// createRoleWithPermissions creates a role with explicit per-resource permissions.
	// Used by privilege-check tests to grant the caller a narrowly-scoped permission set.
	// PredefinedRole is intentionally unset — otherwise the role's permissions would have to
	// be a superset of the predefined role's permissions (see validateKubernetesPermissions).
	createRoleWithPermissions := func(name string, permissions []api.PostPermission) *api.Role {
		roleBody := api.RoleCreationFieldsInternal{
			Name:        name,
			Description: name,
			Permissions: permissions,
		}
		return createTestRole(ctx, rolesApiService, roleBody, &createdRoles, false, http.StatusCreated)
	}

	migrate := func() api.ImplResponse {
		resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(ctx, api.AccessRulesMigrateDeprecatedFields{
			OldRoleId: deprecatedRole.Id,
			NewRoleId: targetRole.Id,
		})
		Expect(err).To(BeNil())
		return resp
	}

	migrateOK := func() api.AccessRulesMigrateDeprecatedResponse {
		resp := migrate()
		Expect(resp.Code).To(Equal(http.StatusOK))
		return resp.Body.(api.AccessRulesMigrateDeprecatedResponse)
	}

	// After a migration, the old rules in `createdAccessRules` have been soft-deleted by the
	// service — re-deleting them via CleanupAccessRules would fail. Reset the tracked list to
	// only the new (live) rules returned by the migration.
	replaceCreatedWithMigrated := func(response api.AccessRulesMigrateDeprecatedResponse) {
		createdAccessRules = []api.AccessRule{}
		createdAccessRules = append(createdAccessRules, response.Succeeded.Created...)
		createdAccessRules = append(createdAccessRules, response.Succeeded.PreExisting...)
	}

	BeforeAll(func() {
		TestsCleanupAndInit()
	})

	AfterAll(func() {
		CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
		createdAccessRules = []api.AccessRule{}
		for _, role := range createdRoles {
			_, err := rolesApiService.DeleteRole(ctx, role.Id)
			Expect(err).To(BeNil())
		}
		createdRoles = []api.Role{}
		TestsCleanupAndInit()
	})

	BeforeEach(func() {
		auditMock := auditmocks.NewWriter(GinkgoT())
		auditMock.On("Write", mock.Anything).Return(nil).Maybe()
		audit.SetMock(auditMock)

		// The role marked "deprecated" — source of migration.
		deprecatedRoleBody := createRoleCreationRequest(static_legacy_roles.ResearcherL1.Permissions, "deprecated role", "to migrate from")
		deprecatedRole = createTestRole(ctx, rolesApiService, deprecatedRoleBody, &createdRoles, false, http.StatusCreated)
		setRoleDeprecated(deprecatedRole.Id, true)

		// The target (non-deprecated, enabled) role — destination of migration.
		targetRoleBody := createRoleCreationRequest(static_legacy_roles.ResearcherL1.Permissions, "target role", "to migrate to")
		targetRole = createTestRole(ctx, rolesApiService, targetRoleBody, &createdRoles, false, http.StatusCreated)

		SetAncestorsCache(ctx, ProjectId, api.RESOURCETYPE_PROJECT, &authorizationutils.Ancestors{TenantId: tenantId})
		SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorizationutils.Ancestors{TenantId: tenantId})
		SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorizationutils.Ancestors{TenantId: tenantId})
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

	Describe("Happy flows", func() {
		It("H1 — migrates rules at tenant / department / project / cluster scope types", func() {
			oldTenantRule := createAccessRule(TestUserId, api.SUBJECTTYPE_USER, tenantId, api.SCOPETYPE_TENANT, deprecatedRole.Id)
			oldDepartmentRule := createAccessRule(TestUserId, api.SUBJECTTYPE_USER, DepartmentId, api.SCOPETYPE_DEPARTMENT, deprecatedRole.Id)
			oldProjectRule := createAccessRule(TestUserId, api.SUBJECTTYPE_USER, ProjectId, api.SCOPETYPE_PROJECT, deprecatedRole.Id)
			oldClusterRule := createAccessRule(TestUserId, api.SUBJECTTYPE_USER, ClusterId, api.SCOPETYPE_CLUSTER, deprecatedRole.Id)
			oldRules := []*api.AccessRule{oldTenantRule, oldDepartmentRule, oldProjectRule, oldClusterRule}

			body := migrateOK()

			Expect(body.Succeeded.Created).To(HaveLen(4))
			Expect(body.Succeeded.PreExisting).To(HaveLen(0))
			Expect(body.Succeeded.Deleted).To(HaveLen(4))
			Expect(body.Failed).To(HaveLen(0))

			// Each new rule must preserve subject + scope fields + cluster, and adopt new roleId + ctx subject.
			newByScope := map[string]api.AccessRule{}
			for _, n := range body.Succeeded.Created {
				newByScope[fmt.Sprintf("%s:%s", n.ScopeType, n.ScopeId)] = n
			}
			for _, o := range oldRules {
				key := fmt.Sprintf("%s:%s", o.ScopeType, o.ScopeId)
				n, ok := newByScope[key]
				Expect(ok).To(BeTrue(), "missing migrated rule for "+key)
				Expect(n.SubjectId).To(Equal(o.SubjectId))
				Expect(n.SubjectType).To(Equal(o.SubjectType))
				Expect(n.ScopeId).To(Equal(o.ScopeId))
				Expect(n.ScopeType).To(Equal(o.ScopeType))
				Expect(n.ScopeName).To(Equal(o.ScopeName))
				Expect(n.ClusterId).To(Equal(o.ClusterId))
				Expect(n.RoleId).To(Equal(targetRole.Id))
				Expect(n.CreatedBy).To(Equal(TestUserId))
			}

			// All old rules appear in the deleted list.
			deletedSet := map[string]bool{}
			for _, d := range body.Succeeded.Deleted {
				deletedSet[d] = true
			}
			for _, o := range oldRules {
				Expect(deletedSet[fmt.Sprint(o.Id)]).To(BeTrue())
			}

			replaceCreatedWithMigrated(body)
		})

		It("H2 — migrates user, group, and service-account subjects", func() {
			createAccessRule(TestUser2Id, api.SUBJECTTYPE_GROUP, ProjectId, api.SCOPETYPE_PROJECT, deprecatedRole.Id)
			createAccessRule(TestUserId, api.SUBJECTTYPE_USER, DepartmentId, api.SCOPETYPE_DEPARTMENT, deprecatedRole.Id)
			createAccessRule("test-service-account-id", api.SUBJECTTYPE_SERVICE_ACCOUNT, ClusterId, api.SCOPETYPE_CLUSTER, deprecatedRole.Id)

			body := migrateOK()

			Expect(body.Succeeded.Created).To(HaveLen(3))
			Expect(body.Succeeded.PreExisting).To(HaveLen(0))
			Expect(body.Succeeded.Deleted).To(HaveLen(3))
			Expect(body.Failed).To(HaveLen(0))

			subjectTypes := map[api.SubjectTypeExtended]bool{}
			for _, n := range body.Succeeded.Created {
				subjectTypes[n.SubjectType] = true
				Expect(n.RoleId).To(Equal(targetRole.Id))
			}
			Expect(subjectTypes).To(HaveKey(api.SUBJECTTYPEEXTENDED_USER))
			Expect(subjectTypes).To(HaveKey(api.SUBJECTTYPEEXTENDED_GROUP))
			Expect(subjectTypes).To(HaveKey(api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT))

			replaceCreatedWithMigrated(body)
		})

		It("H3 — leaves rules on other roles untouched", func() {
			otherRoleBody := createRoleCreationRequest(static_legacy_roles.ResearcherL1.Permissions, "other role", "unrelated")
			otherRole := createTestRole(ctx, rolesApiService, otherRoleBody, &createdRoles, false, http.StatusCreated)
			untouched := createAccessRule(TestUserId, api.SUBJECTTYPE_USER, ProjectId, api.SCOPETYPE_PROJECT, otherRole.Id)
			createAccessRule(TestUserId, api.SUBJECTTYPE_USER, DepartmentId, api.SCOPETYPE_DEPARTMENT, deprecatedRole.Id)

			body := migrateOK()

			Expect(body.Succeeded.Created).To(HaveLen(1))
			Expect(body.Succeeded.PreExisting).To(HaveLen(0))
			Expect(body.Succeeded.Deleted).To(HaveLen(1))
			Expect(body.Failed).To(HaveLen(0))

			// Untouched rule still exists at its original role.
			var untouchedRow models.AccessRule
			err := runaidb.DB().Where("id = ?", untouched.Id).First(&untouchedRow)
			Expect(err).To(BeNil())
			Expect(untouchedRow.RoleID).To(Equal(uint(otherRole.Id)))

			// Rebuild tracking: untouched rule + migrated rule.
			createdAccessRules = []api.AccessRule{*untouched}
			createdAccessRules = append(createdAccessRules, body.Succeeded.Created...)
		})

		It("H4 — empty result when no rules exist on the deprecated role", func() {
			body := migrateOK()
			Expect(body.Succeeded.Created).To(HaveLen(0))
			Expect(body.Succeeded.PreExisting).To(HaveLen(0))
			Expect(body.Succeeded.Deleted).To(HaveLen(0))
			Expect(body.Failed).To(HaveLen(0))
		})

		It("H5 — pre-existing new-role rule: no create, old rule still deleted, reported in preExisting", func() {
			// A new-role rule for TestUserId + ProjectId already exists.
			preExisting := createAccessRule(TestUserId, api.SUBJECTTYPE_USER, ProjectId, api.SCOPETYPE_PROJECT, targetRole.Id)
			// A deprecated-role rule for the same (user + project) that should be reconciled.
			overlapOld := createAccessRule(TestUserId, api.SUBJECTTYPE_USER, ProjectId, api.SCOPETYPE_PROJECT, deprecatedRole.Id)
			// Two unrelated deprecated-role rules that should migrate normally.
			createAccessRule(TestUserId, api.SUBJECTTYPE_USER, DepartmentId, api.SCOPETYPE_DEPARTMENT, deprecatedRole.Id)
			createAccessRule(TestUserId, api.SUBJECTTYPE_USER, ClusterId, api.SCOPETYPE_CLUSTER, deprecatedRole.Id)

			body := migrateOK()

			Expect(body.Succeeded.Created).To(HaveLen(2))
			Expect(body.Succeeded.PreExisting).To(HaveLen(1))
			Expect(body.Succeeded.Deleted).To(HaveLen(3))
			Expect(body.Failed).To(HaveLen(0))

			// The rule reported in `preExisting` is the one we created ahead of time (same Id).
			Expect(body.Succeeded.PreExisting[0].Id).To(Equal(preExisting.Id))
			Expect(body.Succeeded.PreExisting[0].ScopeType).To(Equal(api.SCOPETYPE_PROJECT))
			Expect(body.Succeeded.PreExisting[0].ScopeId).To(Equal(ProjectId))

			// The overlapping old rule was deleted (via preExisting path).
			Expect(body.Succeeded.Deleted).To(ContainElement(fmt.Sprint(overlapOld.Id)))

			replaceCreatedWithMigrated(body)
		})
	})

	Describe("Validation errors", func() {
		It("400 when oldRoleId == newRoleId", func() {
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(ctx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: deprecatedRole.Id,
				NewRoleId: deprecatedRole.Id,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})

		It("400 when old role is not deprecated", func() {
			setRoleDeprecated(deprecatedRole.Id, false)
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(ctx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: deprecatedRole.Id,
				NewRoleId: targetRole.Id,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})

		It("400 when new role is deprecated", func() {
			setRoleDeprecated(targetRole.Id, true)
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(ctx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: deprecatedRole.Id,
				NewRoleId: targetRole.Id,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})

		It("400 when new role is disabled", func() {
			setRoleEnabled(targetRole.Id, false)
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(ctx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: deprecatedRole.Id,
				NewRoleId: targetRole.Id,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})

		It("404 when old role does not exist", func() {
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(ctx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: 999999999,
				NewRoleId: targetRole.Id,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})

		It("404 when new role does not exist", func() {
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(ctx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: deprecatedRole.Id,
				NewRoleId: 999999999,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})
	})

	Describe("Bulk privilege check", func() {
		It("403 when caller has no access_rules permissions at tenant scope", func() {
			noPrivCtx := setContext(tenantId, TestUser2Id, []string{})
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(noPrivCtx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: deprecatedRole.Id,
				NewRoleId: targetRole.Id,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusForbidden))
		})

		It("403 when caller has access_rules.create at tenant scope but no delete", func() {
			createOnlyRole := createRoleWithPermissions("access-rules-create-only", []api.PostPermission{
				{ResourceType: api.RESOURCETYPE_ACCESS_RULES, Actions: []api.Action{api.ACTION_CREATE, api.ACTION_READ}},
			})
			createAccessRule(TestUser2Id, api.SUBJECTTYPE_USER, tenantId, api.SCOPETYPE_TENANT, createOnlyRole.Id)

			callerCtx := setContext(tenantId, TestUser2Id, []string{})
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(callerCtx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: deprecatedRole.Id,
				NewRoleId: targetRole.Id,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusForbidden))
		})

		It("403 when caller has access_rules.create+delete but only at cluster scope (not tenant)", func() {
			fullRole := createRoleWithPermissions("access-rules-cluster-scope-full", []api.PostPermission{
				{ResourceType: api.RESOURCETYPE_ACCESS_RULES, Actions: []api.Action{api.ACTION_CREATE, api.ACTION_DELETE, api.ACTION_READ}},
			})
			createAccessRule(TestUser2Id, api.SUBJECTTYPE_USER, ClusterId, api.SCOPETYPE_CLUSTER, fullRole.Id)

			callerCtx := setContext(tenantId, TestUser2Id, []string{})
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(callerCtx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: deprecatedRole.Id,
				NewRoleId: targetRole.Id,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusForbidden))
		})

		It("403 when caller has access_rules.create+delete but only at department scope (not tenant)", func() {
			fullRole := createRoleWithPermissions("access-rules-department-scope-full", []api.PostPermission{
				{ResourceType: api.RESOURCETYPE_ACCESS_RULES, Actions: []api.Action{api.ACTION_CREATE, api.ACTION_DELETE, api.ACTION_READ}},
			})
			createAccessRule(TestUser2Id, api.SUBJECTTYPE_USER, DepartmentId, api.SCOPETYPE_DEPARTMENT, fullRole.Id)

			callerCtx := setContext(tenantId, TestUser2Id, []string{})
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(callerCtx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: deprecatedRole.Id,
				NewRoleId: targetRole.Id,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusForbidden))
		})

		It("succeeds when caller has access_rules.create+delete at system scope", func() {
			// Main ctx uses SuperAdmins which gives SystemAdmin at TENANT scope — already covered
			// by happy-flow tests. This test validates the other success path: `scopes.System == true`.
			// System-scope access rules can't be created via the public API from a non-management
			// tenant, so insert the rule directly.
			fullRole := createRoleWithPermissions("access-rules-system-scope-full", []api.PostPermission{
				{ResourceType: api.RESOURCETYPE_ACCESS_RULES, Actions: []api.Action{api.ACTION_CREATE, api.ACTION_DELETE, api.ACTION_READ}},
			})
			systemAR := &models.AccessRule{
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPEEXTENDED_USER,
				RoleID:      uint(fullRole.Id),
				ScopeId:     "",
				ScopeType:   api.SCOPETYPE_SYSTEM,
				ScopeName:   "system",
				TenantId:    testing.TestTenant,
			}
			_, err := runaidb.DB().Create(systemAR)
			Expect(err).To(BeNil())
			defer func() {
				_, err := runaidb.DB().Unscoped().Delete(systemAR)
				Expect(err).To(BeNil())
			}()

			oldRule := createAccessRule(TestUserId, api.SUBJECTTYPE_USER, DepartmentId, api.SCOPETYPE_DEPARTMENT, deprecatedRole.Id)

			callerCtx := setContext(tenantId, TestUser2Id, []string{})
			resp, err := accessRulesApiService.AccessRulesMigrateDeprecated(callerCtx, api.AccessRulesMigrateDeprecatedFields{
				OldRoleId: deprecatedRole.Id,
				NewRoleId: targetRole.Id,
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			body := resp.Body.(api.AccessRulesMigrateDeprecatedResponse)
			Expect(body.Succeeded.Created).To(HaveLen(1))
			Expect(body.Succeeded.PreExisting).To(HaveLen(0))
			Expect(body.Succeeded.Deleted).To(ContainElement(fmt.Sprint(oldRule.Id)))
			Expect(body.Failed).To(HaveLen(0))

			replaceCreatedWithMigrated(body)
		})
	})

	Describe("Limit", func() {
		It("caps migration at MaxDeprecatedMigrationSize; a follow-up call drains the remainder", func() {
			total := MaxDeprecatedMigrationSize + 5
			for i := 0; i < total; i++ {
				subjectId := fmt.Sprintf("limituser-%d@run.ai", i)
				createAccessRule(subjectId, api.SUBJECTTYPE_USER, ProjectId, api.SCOPETYPE_PROJECT, deprecatedRole.Id)
			}

			// First call migrates exactly MaxDeprecatedMigrationSize rules, leaves 5 behind.
			body := migrateOK()
			Expect(body.Succeeded.Created).To(HaveLen(MaxDeprecatedMigrationSize))
			Expect(body.Succeeded.PreExisting).To(HaveLen(0))
			Expect(body.Succeeded.Deleted).To(HaveLen(MaxDeprecatedMigrationSize))
			Expect(body.Failed).To(HaveLen(0))

			// Exactly 5 deprecated-role rules remain live.
			var remainingCount int64
			err := runaidb.DB().Model(&models.AccessRule{}).
				Where("role_id = ?", deprecatedRole.Id).
				Where("tenant_id = ?", testing.TestTenant).
				Count(&remainingCount)
			Expect(err).To(BeNil())
			Expect(remainingCount).To(Equal(int64(5)))

			// Second call drains the remaining 5.
			body2 := migrateOK()
			Expect(body2.Succeeded.Created).To(HaveLen(5))
			Expect(body2.Succeeded.PreExisting).To(HaveLen(0))
			Expect(body2.Succeeded.Deleted).To(HaveLen(5))
			Expect(body2.Failed).To(HaveLen(0))

			err = runaidb.DB().Model(&models.AccessRule{}).
				Where("role_id = ?", deprecatedRole.Id).
				Where("tenant_id = ?", testing.TestTenant).
				Count(&remainingCount)
			Expect(err).To(BeNil())
			Expect(remainingCount).To(Equal(int64(0)))

			// Third call is a no-op (empty succeeded/failed).
			body3 := migrateOK()
			Expect(body3.Succeeded.Created).To(HaveLen(0))
			Expect(body3.Succeeded.PreExisting).To(HaveLen(0))
			Expect(body3.Succeeded.Deleted).To(HaveLen(0))
			Expect(body3.Failed).To(HaveLen(0))

			// Track every live rule so AfterEach can clean up cleanly.
			createdAccessRules = []api.AccessRule{}
			createdAccessRules = append(createdAccessRules, body.Succeeded.Created...)
			createdAccessRules = append(createdAccessRules, body2.Succeeded.Created...)
		})
	})
})
