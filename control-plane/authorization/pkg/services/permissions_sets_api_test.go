/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"

	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
)

const (
	minimalsResourceTypeMigrationVersion = "02240009_add_minimal_resource_permissions"
)

var _ = Describe("Permissions sets API tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	ctx := setContext(tenantId, TestUserId, []string{dao.SuperAdmins})
	PermissionSetMap := map[string]uuid.UUID{}
	permissionSetLength := 57
	workspaceEditPermissionsLen := 20

	permissionsApiService := NewPermissionsApiService()

	// create a role to be used by tests
	BeforeAll(func() {
		dao.InitPermissionSetsMap(&PermissionSetMap)
	})

	AfterAll(func() {
		TestsCleanupAndInit()
	})

	Context("init", func() {
		It("test static permission sets", func() {
			permissionSets, err := dao.DbGetPermissionSets(ctx, true)
			Expect(err).To(BeNil())
			Expect(len(*permissionSets)).To(Not(Equal(0)))
			var workspaceEditPS models.PermissionSet
			for i := range *permissionSets {
				if (*permissionSets)[i].Name == "Workspace edit" {
					workspaceEditPS = (*permissionSets)[i]
				}
			}
			Expect(len(workspaceEditPS.Permissions)).To(Equal(workspaceEditPermissionsLen))
		})

		It("validate permission sets have minimal resource types when needed", func() {
			var violations []string

			// Check all static permission sets
			for _, permissionSet := range dao.PermissionSets {
				hasCluster := false
				hasClustersMinimal := false
				hasNodePools := false
				hasNodePoolsMinimal := false
				hasUsers := false
				hasUsersMinimal := false

				// Check what resource types this permission set has
				for _, permission := range permissionSet.Permissions {
					switch permission.ResourceType {
					case api.RESOURCETYPE_CLUSTER:
						hasCluster = true
					case api.RESOURCETYPE_CLUSTERS_MINIMAL:
						hasClustersMinimal = true
					case api.RESOURCETYPE_NODEPOOLS:
						hasNodePools = true
					case api.RESOURCETYPE_NODEPOOLS_MINIMAL:
						hasNodePoolsMinimal = true
					case api.RESOURCETYPE_USERS:
						hasUsers = true
					case api.RESOURCETYPE_USERS_MINIMAL:
						hasUsersMinimal = true
					}
				}

				// Rule: If a permission set has the "full" resource type (CLUSTER, NODEPOOLS, or USERS),
				// it MUST also have the corresponding "minimal" resource type.
				// Note: Having only the minimal version without the full version is allowed.

				if hasCluster && !hasClustersMinimal {
					violations = append(violations, fmt.Sprintf("Permission set '%s' has RESOURCETYPE_CLUSTER but missing RESOURCETYPE_CLUSTERS_MINIMAL", permissionSet.Name))
				}

				if hasNodePools && !hasNodePoolsMinimal {
					violations = append(violations, fmt.Sprintf("Permission set '%s' has RESOURCETYPE_NODEPOOLS but missing RESOURCETYPE_NODEPOOLS_MINIMAL", permissionSet.Name))
				}

				if hasUsers && !hasUsersMinimal {
					violations = append(violations, fmt.Sprintf("Permission set '%s' has RESOURCETYPE_USERS but missing RESOURCETYPE_USERS_MINIMAL", permissionSet.Name))
				}
			}

			// If there are violations, fail the test with all violations listed
			if len(violations) > 0 {
				errorMessage := fmt.Sprintf("Found %d permission set(s) with missing minimal resource types:\n", len(violations))
				for _, violation := range violations {
					errorMessage += fmt.Sprintf("  - %s\n", violation)
				}
				Fail(errorMessage)
			}
		})
	})

	Context("GET api", func() {
		It("get all and get by id", func() {
			response, err := permissionsApiService.GetPermissionSets(ctx)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			permissionSets := response.Body.(api.GetPermissionSets200Response).PermissionSets
			Expect(len(permissionSets)).To(Equal(permissionSetLength))
			var workspaceEditPS api.PermissionSet
			catalogNames := map[string]bool{}
			for i := range permissionSets {
				catalogNames[(permissionSets)[i].Name] = true
				if (permissionSets)[i].Name == "Workspace edit" {
					workspaceEditPS = (permissionSets)[i]
				}
			}
			// the cluster-config permission sets are internal and must not
			// be exposed in the role-builder catalog.
			Expect(catalogNames[static_permission_sets.ClusterConfigEdit.Name]).To(BeFalse())
			Expect(catalogNames[static_permission_sets.ClusterConfigRead.Name]).To(BeFalse())
			Expect(len(workspaceEditPS.Permissions)).To(Equal(workspaceEditPermissionsLen))

			response, err = permissionsApiService.GetPermissionSet(ctx, workspaceEditPS.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			permissionSetGet := response.Body.(api.PermissionSet)
			Expect(len(permissionSetGet.Permissions)).To(Equal(workspaceEditPermissionsLen))
			Expect(permissionSetGet.Name).To(Equal("Workspace edit"))
		})
	})

	Context("granular asset permission sets", func() {
		type expectedPermission struct {
			resourceType api.ResourceType
			create       bool
			read         bool
			update       bool
			delete       bool
		}

		minimals := []expectedPermission{
			{resourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL, read: true},
			{resourceType: api.RESOURCETYPE_NODEPOOLS_MINIMAL, read: true},
		}
		granularSet := func(ps models.PermissionSet, resourceType api.ResourceType, editable bool) (models.PermissionSet, []expectedPermission) {
			return ps, append([]expectedPermission{
				{resourceType: resourceType, create: editable, read: true, update: editable, delete: editable},
			}, minimals...)
		}

		granularSets := map[string][]expectedPermission{}
		granularDefinitions := []models.PermissionSet{}
		for _, entry := range []struct {
			ps           models.PermissionSet
			resourceType api.ResourceType
			editable     bool
		}{
			{static_permission_sets.TemplatesOnlyRead, api.RESOURCETYPE_TEMPLATES, false},
			{static_permission_sets.TemplatesOnlyEdit, api.RESOURCETYPE_TEMPLATES, true},
			{static_permission_sets.EnvironmentsOnlyRead, api.RESOURCETYPE_ENVIRONMENTS, false},
			{static_permission_sets.EnvironmentsOnlyEdit, api.RESOURCETYPE_ENVIRONMENTS, true},
			{static_permission_sets.ComputeResourcesOnlyRead, api.RESOURCETYPE_COMPUTE_RESOURCES, false},
			{static_permission_sets.ComputeResourcesOnlyEdit, api.RESOURCETYPE_COMPUTE_RESOURCES, true},
			{static_permission_sets.CredentialsOnlyRead, api.RESOURCETYPE_CREDENTIALS, false},
			{static_permission_sets.CredentialsOnlyEdit, api.RESOURCETYPE_CREDENTIALS, true},
		} {
			ps, expected := granularSet(entry.ps, entry.resourceType, entry.editable)
			granularSets[ps.Name] = expected
			granularDefinitions = append(granularDefinitions, ps)
		}

		It("grants exactly one asset resource type plus the minimal resource types", func() {
			for _, ps := range granularDefinitions {
				expected := granularSets[ps.Name]
				Expect(ps.Permissions).To(HaveLen(len(expected)), "permission set %q", ps.Name)
				for i, permission := range ps.Permissions {
					Expect(permission.ResourceType).To(Equal(expected[i].resourceType), "permission set %q", ps.Name)
					Expect(permission.Create).To(Equal(expected[i].create), "permission set %q, resource %q", ps.Name, permission.ResourceType)
					Expect(permission.Read).To(Equal(expected[i].read), "permission set %q, resource %q", ps.Name, permission.ResourceType)
					Expect(permission.Update).To(Equal(expected[i].update), "permission set %q, resource %q", ps.Name, permission.ResourceType)
					Expect(permission.Delete).To(Equal(expected[i].delete), "permission set %q, resource %q", ps.Name, permission.ResourceType)
				}
			}
		})

		It("exposes every granular set in the role-builder catalog", func() {
			response, err := permissionsApiService.GetPermissionSets(ctx)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			catalogNames := map[string]bool{}
			for _, ps := range response.Body.(api.GetPermissionSets200Response).PermissionSets {
				catalogNames[ps.Name] = true
			}
			for name := range granularSets {
				Expect(catalogNames[name]).To(BeTrue(), "permission set %q is missing from the catalog", name)
			}
		})

		It("leaves the bundled templates sets untouched", func() {
			Expect(static_permission_sets.TemplatesRead.Permissions).To(HaveLen(13))
			Expect(static_permission_sets.TemplatesEdit.Permissions).To(HaveLen(13))
		})
	})

	Context("Migration 02240009 - add minimal resource permissions", func() {
		It("validates migration adds minimal permissions to custom roles with cluster/nodepool read access", func() {
			// Create several custom roles without minimal permissions but with cluster/nodepool read permissions
			customRoles := []struct {
				name        string
				description string
				permissions []models.Permission
			}{
				{
					name:        "Test Role With Cluster Read",
					description: "Role with cluster read permission but no clusters-minimal",
					permissions: []models.Permission{
						{ResourceType: api.RESOURCETYPE_CLUSTER, Read: true},
						{ResourceType: api.RESOURCETYPE_PROJECT, Read: true, Create: true},
					},
				},
				{
					name:        "Test Role With NodePools Read",
					description: "Role with nodepools read permission but no nodepools-minimal",
					permissions: []models.Permission{
						{ResourceType: api.RESOURCETYPE_NODEPOOLS, Read: true},
						{ResourceType: api.RESOURCETYPE_WORKLOADS, Read: true},
					},
				},
				{
					name:        "Test Role With Both Cluster And NodePools",
					description: "Role with both cluster and nodepools read permissions",
					permissions: []models.Permission{
						{ResourceType: api.RESOURCETYPE_CLUSTER, Read: true, Update: true},
						{ResourceType: api.RESOURCETYPE_NODEPOOLS, Read: true, Create: true},
						{ResourceType: api.RESOURCETYPE_DEPARTMENT, Read: true},
					},
				},
				{
					name:        "Test Role With Existing Clusters Minimal",
					description: "Role with cluster read and already has clusters-minimal - should not create duplicate",
					permissions: []models.Permission{
						{ResourceType: api.RESOURCETYPE_CLUSTER, Read: true},
						{ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL, Read: true},
						{ResourceType: api.RESOURCETYPE_PROJECT, Read: true},
					},
				},
			}

			// Create the test roles in the database
			createdRoleIds := []uint{}
			for _, roleSpec := range customRoles {
				role := models.Role{
					Name:        roleSpec.name,
					Description: roleSpec.description,
					Custom:      true,
					ScopeType:   models.ROLE_SCOPE_TENANT,
					Enabled:     true,
				}

				// Add permissions
				for _, perm := range roleSpec.permissions {
					permission := perm
					permission.RoleID = 0 // Will be set by GORM
					role.Permissions = append(role.Permissions, &permission)
				}

				// Create role in DB
				_, err := runaidb.DB().Create(&role)
				Expect(err).To(BeNil())
				createdRoleIds = append(createdRoleIds, role.ID)
			}

			// Verify minimal permissions DO NOT exist before migration (except for role that already has it)
			for idx, roleId := range createdRoleIds {
				var permissions []models.Permission
				err := runaidb.DB().Where("role_id = ?", roleId).Find(&permissions)
				Expect(err).To(BeNil())

				hasClustersMinimal := false
				hasNodePoolsMinimal := false
				hasCluster := false
				hasNodePools := false
				clustersMinimalCount := 0

				for _, perm := range permissions {
					switch perm.ResourceType {
					case api.RESOURCETYPE_CLUSTER:
						hasCluster = true
					case api.RESOURCETYPE_CLUSTERS_MINIMAL:
						hasClustersMinimal = true
						clustersMinimalCount++
					case api.RESOURCETYPE_NODEPOOLS:
						hasNodePools = true
					case api.RESOURCETYPE_NODEPOOLS_MINIMAL:
						hasNodePoolsMinimal = true
					}
				}

				// Validate based on what the role should have
				if idx == 0 { // First role: has cluster
					Expect(hasCluster).To(BeTrue(), fmt.Sprintf("Role %s should have CLUSTER permission", customRoles[idx].name))
					Expect(hasClustersMinimal).To(BeFalse(), fmt.Sprintf("Role %s should NOT have CLUSTERS_MINIMAL before migration", customRoles[idx].name))
				} else if idx == 1 { // Second role: has nodepools
					Expect(hasNodePools).To(BeTrue(), fmt.Sprintf("Role %s should have NODEPOOLS permission", customRoles[idx].name))
					Expect(hasNodePoolsMinimal).To(BeFalse(), fmt.Sprintf("Role %s should NOT have NODEPOOLS_MINIMAL before migration", customRoles[idx].name))
				} else if idx == 2 { // Third role: has both
					Expect(hasCluster).To(BeTrue(), fmt.Sprintf("Role %s should have CLUSTER permission", customRoles[idx].name))
					Expect(hasNodePools).To(BeTrue(), fmt.Sprintf("Role %s should have NODEPOOLS permission", customRoles[idx].name))
					Expect(hasClustersMinimal).To(BeFalse(), fmt.Sprintf("Role %s should NOT have CLUSTERS_MINIMAL before migration", customRoles[idx].name))
					Expect(hasNodePoolsMinimal).To(BeFalse(), fmt.Sprintf("Role %s should NOT have NODEPOOLS_MINIMAL before migration", customRoles[idx].name))
				} else if idx == 3 { // Fourth role: already has clusters-minimal
					Expect(hasCluster).To(BeTrue(), fmt.Sprintf("Role %s should have CLUSTER permission", customRoles[idx].name))
					Expect(hasClustersMinimal).To(BeTrue(), fmt.Sprintf("Role %s should already have CLUSTERS_MINIMAL before migration", customRoles[idx].name))
					Expect(clustersMinimalCount).To(Equal(1), fmt.Sprintf("Role %s should have exactly 1 CLUSTERS_MINIMAL permission before migration", customRoles[idx].name))
				}
			}

			// Load and execute the migration UP SQL file
			migrationUpFile := fmt.Sprintf("%s.up.sql", minimalsResourceTypeMigrationVersion)
			migrationUpPath := filepath.Join("..", "database", "migrations", "up", migrationUpFile)
			migrationUpSQL, err := os.ReadFile(migrationUpPath)
			Expect(err).To(BeNil(), fmt.Sprintf("Failed to read migration up file: %v", err))

			_, err = runaidb.DB().Exec(string(migrationUpSQL))
			Expect(err).To(BeNil(), fmt.Sprintf("Failed to execute migration up: %v", err))

			// Verify minimal permissions now EXIST after migration
			for idx, roleId := range createdRoleIds {
				var permissions []models.Permission
				err := runaidb.DB().Where("role_id = ?", roleId).Find(&permissions)
				Expect(err).To(BeNil())

				hasClustersMinimal := false
				hasNodePoolsMinimal := false
				hasCluster := false
				hasNodePools := false
				clustersMinimalCount := 0
				nodePoolsMinimalCount := 0

				for _, perm := range permissions {
					switch perm.ResourceType {
					case api.RESOURCETYPE_CLUSTER:
						hasCluster = true
					case api.RESOURCETYPE_CLUSTERS_MINIMAL:
						hasClustersMinimal = true
						clustersMinimalCount++
						// Validate it has read permission
						Expect(perm.Read).To(BeTrue(), fmt.Sprintf("CLUSTERS_MINIMAL for role %s should have read=true", customRoles[idx].name))
						Expect(perm.Create).To(BeFalse())
						Expect(perm.Update).To(BeFalse())
						Expect(perm.Delete).To(BeFalse())
						Expect(perm.Sync).To(BeFalse())
					case api.RESOURCETYPE_NODEPOOLS:
						hasNodePools = true
					case api.RESOURCETYPE_NODEPOOLS_MINIMAL:
						hasNodePoolsMinimal = true
						nodePoolsMinimalCount++
						// Validate it has read permission
						Expect(perm.Read).To(BeTrue(), fmt.Sprintf("NODEPOOLS_MINIMAL for role %s should have read=true", customRoles[idx].name))
						Expect(perm.Create).To(BeFalse())
						Expect(perm.Update).To(BeFalse())
						Expect(perm.Delete).To(BeFalse())
						Expect(perm.Sync).To(BeFalse())
					}
				}

				// Validate based on what the role should have after migration
				if idx == 0 { // First role: has cluster, should now have clusters-minimal
					Expect(hasCluster).To(BeTrue())
					Expect(hasClustersMinimal).To(BeTrue(), fmt.Sprintf("Role %s should have CLUSTERS_MINIMAL after migration", customRoles[idx].name))
					Expect(clustersMinimalCount).To(Equal(1), fmt.Sprintf("Role %s should have exactly 1 CLUSTERS_MINIMAL after migration", customRoles[idx].name))
				} else if idx == 1 { // Second role: has nodepools, should now have nodepools-minimal
					Expect(hasNodePools).To(BeTrue())
					Expect(hasNodePoolsMinimal).To(BeTrue(), fmt.Sprintf("Role %s should have NODEPOOLS_MINIMAL after migration", customRoles[idx].name))
					Expect(nodePoolsMinimalCount).To(Equal(1), fmt.Sprintf("Role %s should have exactly 1 NODEPOOLS_MINIMAL after migration", customRoles[idx].name))
				} else if idx == 2 { // Third role: has both, should now have both minimal
					Expect(hasCluster).To(BeTrue())
					Expect(hasNodePools).To(BeTrue())
					Expect(hasClustersMinimal).To(BeTrue(), fmt.Sprintf("Role %s should have CLUSTERS_MINIMAL after migration", customRoles[idx].name))
					Expect(hasNodePoolsMinimal).To(BeTrue(), fmt.Sprintf("Role %s should have NODEPOOLS_MINIMAL after migration", customRoles[idx].name))
					Expect(clustersMinimalCount).To(Equal(1), fmt.Sprintf("Role %s should have exactly 1 CLUSTERS_MINIMAL after migration", customRoles[idx].name))
					Expect(nodePoolsMinimalCount).To(Equal(1), fmt.Sprintf("Role %s should have exactly 1 NODEPOOLS_MINIMAL after migration", customRoles[idx].name))
				} else if idx == 3 { // Fourth role: already had clusters-minimal, should still have exactly 1
					Expect(hasCluster).To(BeTrue())
					Expect(hasClustersMinimal).To(BeTrue(), fmt.Sprintf("Role %s should still have CLUSTERS_MINIMAL after migration", customRoles[idx].name))
					Expect(clustersMinimalCount).To(Equal(1), fmt.Sprintf("Role %s should still have exactly 1 CLUSTERS_MINIMAL after migration (no duplicate created)", customRoles[idx].name))
				}
			}
		})
	})
})
