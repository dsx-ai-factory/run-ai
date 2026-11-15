/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	static_legacy_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
	sharedCache "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"k8s.io/utils/ptr"
)

type ValidateCurrentUserHasPrivilegeTestArgs struct {
	SubjectId        string
	AssignedRoleName string
	ScopeId          string
	ScopeType        api.ScopeType
	ShouldAllow      bool
	Tenant           uint
	AssignedRole     *models.Role
}

const SuperAdmin = "super-admin"

var _ = Describe("Privilege escalation tests", Ordered, func() {
	const (
		OrgManager        = "org-manager"
		Researcher        = "researcher"
		ClusterAdmin      = "cluster-admin"
		OtherClusterAdmin = "other-cluster-admin"
		DepartmentAdmin   = "department-admin"
	)
	var (
		ClusterId        = "a418ed33-9399-48c0-a890-122cadd13bfd"
		ClusterName      = "Cluster-a"
		OtherClusterId   = "122cadd1-9399-48c0-a890-3bfda418ed33"
		OtherClusterName = "Cluster-b"
		DepartmentId     = "1"
		DepartmentName   = "Department-a"
		ProjectId        = "1"
	)

	ctx := context.Background()
	testTenant := fmt.Sprintf("%d", testing.TestTenant)
	createdAccessRules := []*models.AccessRule{}
	createdAccessRulesForNewUser := []*models.AccessRule{}
	tenantSettings := NewMockTenantSettingsService(map[string]bool{})
	internalApiService := NewInternalApiService(tenantSettings)

	createInitialAccessRule := func(roleId uint, name string, scopeId string, scopeName string, clusterId string, scopeType api.ScopeType, tenant uint) {
		accessRule, err := dao.DbCreateAccessRule(ctx, tenant, SuperAdmin, scopeName, clusterId, false, api.AccessRuleCreationFields{
			ScopeId:     scopeId,
			ScopeType:   scopeType,
			SubjectId:   name,
			SubjectType: api.SUBJECTTYPE_USER,
			RoleId:      int32(roleId),
		})
		Expect(err).To(BeNil())
		createdAccessRules = append(createdAccessRules, accessRule)
	}

	runValidateCurrentUserHasPrivilegeTest := func(args ValidateCurrentUserHasPrivilegeTestArgs) {
		role := args.AssignedRole
		var err *api.Error
		if role == nil {
			role, err = dao.DbGetRoleByIdWithInternal(ctx, fmt.Sprintf("%d", args.Tenant), dao.MapStaticRoleNameToRoleId[args.AssignedRoleName])
			Expect(err).To(BeNil())
		} else {
			// roles are override (in DescribeTable) so we want the new id
			filters, _ := queryapi.NewParser(models.Role{}.SortAndFilterFieldsToColumn()).
				Filters([]string{fmt.Sprintf("name=^%s", args.AssignedRoleName)})
			roles, err := dao.DbGetRoles(ctx, fmt.Sprintf("%d", testing.TestTenant), 0, 0, "", "", filters, "", false)
			Expect(err).To(BeNil())
			Expect(roles).NotTo(BeNil())
			Expect(len(*roles)).To(BeNumerically(">=", 1)) // at least one role with the name, could be more due legacy and non-legacy roles
			// Find exact match when prefix filter returns multiple results
			role = nil
			for i := range *roles {
				if (*roles)[i].Name == args.AssignedRoleName {
					role = &(*roles)[i]
					break
				}
			}
			Expect(role).NotTo(BeNil())
		}

		uuid, _ := uuid.NewUUID()
		subjectWithAccessRule := fmt.Sprintf("%s@run.ai", uuid.String())
		accessRule := models.AccessRule{
			Role:        *role,
			ScopeId:     args.ScopeId,
			ScopeType:   args.ScopeType,
			SubjectId:   subjectWithAccessRule,
			SubjectType: api.SUBJECTTYPEEXTENDED_USER,
		}
		err = validateCurrentUserHasPrivilege(ctx, args.Tenant, args.SubjectId, []string{}, accessRule)
		Expect(err == nil).To(Equal(args.ShouldAllow))

		ar, arErr := dao.DbCreateAccessRule(ctx, args.Tenant, "", "", "", false, api.AccessRuleCreationFields{
			SubjectId:   accessRule.SubjectId,
			SubjectType: api.SubjectType(accessRule.SubjectType),
			RoleId:      int32(role.ID),
			ScopeId:     accessRule.ScopeId,
			ScopeType:   accessRule.ScopeType,
			ClusterId:   accessRule.ClusterUuid,
		})
		Expect(arErr).To(BeNil())
		createdAccessRulesForNewUser = append(createdAccessRulesForNewUser, ar)
		apiCtx := setContext(fmt.Sprint(args.Tenant), args.SubjectId, []string{})
		evaluate, apiErr := internalApiService.PrivilegeEscalationEvaluate(apiCtx, subjectWithAccessRule, []string{})
		Expect(apiErr).To(BeNil())
		if args.ShouldAllow {
			Expect(evaluate.Code).To(Equal(http.StatusNoContent))
		} else {
			Expect(evaluate.Code).To(Equal(http.StatusForbidden))
		}
	}

	BeforeAll(func() {
		TestsCleanupAndInit()

		SetAncestorsCache(ctx, ProjectId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
			TenantId:     testTenant,
			ClusterUuid:  &ClusterId,
			DepartmentId: &DepartmentId,
		})

		SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
			TenantId:    testTenant,
			ClusterUuid: &ClusterId,
		})

		SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
			TenantId: testTenant,
		})

		SetAncestorsCache(ctx, OtherClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
			TenantId: testTenant,
		})

		createInitialAccessRule(dao.MapStaticRoleNameToRoleId[static_roles.SystemAdmin.Name], ClusterAdmin, ClusterId, ClusterName, ClusterId, api.SCOPETYPE_CLUSTER, testing.TestTenant)

		createInitialAccessRule(dao.MapStaticRoleNameToRoleId[static_roles.SystemAdmin.Name], OtherClusterAdmin, OtherClusterId, OtherClusterName, OtherClusterId, api.SCOPETYPE_CLUSTER, testing.TestTenant)

		createInitialAccessRule(dao.MapStaticRoleNameToRoleId[static_roles.DepartmentAdmin.Name], DepartmentAdmin, DepartmentId, DepartmentName, ClusterId, api.SCOPETYPE_DEPARTMENT, testing.TestTenant)

		createInitialAccessRule(dao.MapStaticRoleNameToRoleId[static_legacy_roles.Researcher.Name], Researcher, testTenant, testTenant, "", api.SCOPETYPE_TENANT, testing.TestTenant)

		createInitialAccessRule(dao.MapStaticRoleNameToRoleId[static_roles.SystemAdmin.Name], OrgManager, testTenant, testTenant, "", api.SCOPETYPE_TENANT, testing.TestTenant)
	})

	AfterAll(func() {
		for _, accessRule := range createdAccessRules {
			err := dao.DbDeleteAccessRule(ctx, testing.TestTenant, accessRule.ID)
			Expect(err).To(BeNil())
		}
		createdAccessRules = []*models.AccessRule{}
		for _, accessRule := range createdAccessRulesForNewUser {
			err := dao.DbDeleteAccessRule(ctx, testing.TestTenant, accessRule.ID)
			Expect(err).To(BeNil())
		}
		createdAccessRulesForNewUser = []*models.AccessRule{}

		sharedCache.ClearCache()

		TestsCleanupAndInit()
	})

	entries := []TableEntry{}
	unScopedUsers := []string{
		Researcher,
		OrgManager,
	}
	unScopedRoles := []string{
		static_legacy_roles.Researcher.Name,
		static_roles.SystemAdmin.Name,
	}

	for _, assigner := range unScopedUsers {
		for _, roleAssigned := range unScopedRoles {
			entries = append(entries,
				Entry(fmt.Sprintf("%s assign %s role", assigner, roleAssigned), ValidateCurrentUserHasPrivilegeTestArgs{
					SubjectId:        assigner,
					AssignedRoleName: roleAssigned,
					ShouldAllow:      assigner == OrgManager || roleAssigned == static_legacy_roles.Researcher.Name,
					ScopeId:          testTenant,
					ScopeType:        api.SCOPETYPE_TENANT,
					Tenant:           testing.TestTenant,
				}))
		}
	}

	DescribeTable("Un-scoped", runValidateCurrentUserHasPrivilegeTest, entries)
	entries = []TableEntry{}

	scopedUsers := []string{
		OrgManager,
		ClusterAdmin,
		DepartmentAdmin,
		Researcher,
	}
	scopedRoles := []string{
		static_legacy_roles.Researcher.Name,
		static_legacy_roles.MLEngineer.Name,
		static_roles.DepartmentAdmin.Name,
		static_roles.SystemAdmin.Name,
	}

	entries = append(entries,
		Entry(fmt.Sprintf("cluster admin cant assign tenant admin"), ValidateCurrentUserHasPrivilegeTestArgs{
			SubjectId:        ClusterAdmin,
			AssignedRoleName: static_roles.SystemAdmin.Name,
			ShouldAllow:      false,
			ScopeId:          testTenant,
			ScopeType:        api.SCOPETYPE_TENANT,
			Tenant:           testing.TestTenant,
		}))

	for _, roleAssigned := range scopedRoles {
		for _, assigner := range scopedUsers {
			entries = append(entries,
				Entry(fmt.Sprintf("%s assign %s role", assigner, roleAssigned), ValidateCurrentUserHasPrivilegeTestArgs{
					SubjectId:        assigner,
					AssignedRoleName: roleAssigned,
					ShouldAllow: assigner == OrgManager ||
						assigner == ClusterAdmin ||
						(assigner == DepartmentAdmin && roleAssigned != static_roles.SystemAdmin.Name &&
							roleAssigned != static_legacy_roles.Researcher.Name && roleAssigned != static_legacy_roles.MLEngineer.Name) ||
						(assigner == Researcher && roleAssigned == static_legacy_roles.Researcher.Name),
					ScopeId:   ProjectId,
					ScopeType: api.SCOPETYPE_PROJECT,
					Tenant:    testing.TestTenant,
				}))
		}
		entries = append(entries,
			Entry(fmt.Sprintf("Other cluster admin cant assign role %s", roleAssigned), ValidateCurrentUserHasPrivilegeTestArgs{
				SubjectId:        OtherClusterAdmin,
				AssignedRoleName: roleAssigned,
				ShouldAllow:      false,
				ScopeId:          ProjectId,
				ScopeType:        api.SCOPETYPE_PROJECT,
				Tenant:           testing.TestTenant,
			}))
	}

	for _, assigner := range scopedUsers {
		entries = append(entries,
			Entry("No one can assign ClusterServiceAccount role", ValidateCurrentUserHasPrivilegeTestArgs{
				SubjectId:        assigner,
				AssignedRoleName: static_roles.ClusterServiceAccount.Name,
				ShouldAllow:      false,
				ScopeId:          ClusterId,
				ScopeType:        api.SCOPETYPE_CLUSTER,
				Tenant:           testing.TestTenant,
			}))
		entries = append(entries,
			Entry("No one can assign InstallerServiceAccount role", ValidateCurrentUserHasPrivilegeTestArgs{
				SubjectId:        assigner,
				AssignedRoleName: static_roles.InstallerServiceAccount.Name,
				ShouldAllow:      false,
				ScopeId:          testTenant,
				ScopeType:        api.SCOPETYPE_TENANT,
				Tenant:           testing.TestTenant,
			}))
	}
	DescribeTable("Scoped", runValidateCurrentUserHasPrivilegeTest, entries)
	entries = []TableEntry{}

	// A System administrator must be able to assign a custom role that
	// grants cluster-config (e.g. a role cloned from an L2-researcher template in
	// 2.22, which carried cluster-config:read). After the permission-sets refactor
	// the System administrator role no longer grants cluster-config, so the
	// privilege-escalation guard wrongly blocks the assignment with 403.
	It("System administrator can assign a custom role that grants cluster-config", func() {
		// A custom tenant role granting cluster-config:read. In permission-set mode
		// DbCreateRole wraps the inline permission into an internal permission set.
		customRole, createErr := dao.DbCreateRole(ctx, "", models.Role{
			Name:        "custom-cluster-config-role",
			Description: "custom role granting cluster-config read",
			ScopeType:   models.ROLE_SCOPE_TENANT,
			ScopeId:     ptr.To(testTenant),
			Custom:      true,
			Enabled:     true,
			TenantId:    ptr.To(testTenant),
			Permissions: []*models.Permission{
				{ResourceType: api.RESOURCETYPE_CLUSTER_CONFIG, Read: true},
			},
		})
		Expect(createErr).To(BeNil())
		DeferCleanup(func() {
			Expect(dao.DbDeleteRole(ctx, testTenant, customRole.ID)).To(BeNil())
		})

		// OrgManager holds System administrator at tenant scope (seeded in BeforeAll).
		accessRule := models.AccessRule{
			Role:        *customRole,
			RoleID:      customRole.ID,
			ScopeId:     testTenant,
			ScopeType:   api.SCOPETYPE_TENANT,
			SubjectId:   "assignee@run.ai",
			SubjectType: api.SUBJECTTYPEEXTENDED_USER,
		}
		err := validateCurrentUserHasPrivilege(ctx, testing.TestTenant, OrgManager, []string{}, accessRule)
		Expect(err).To(BeNil(), "System administrator should be able to assign a custom role that grants cluster-config")
	})
})
