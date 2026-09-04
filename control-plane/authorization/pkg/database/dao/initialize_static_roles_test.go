/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	static_legacy_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-legacy-roles"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"gorm.io/gorm"
)

var _ = Describe("Initialize static roles and permission sets tests", Ordered, func() {
	Context("init roles and permission set on upgrade", func() {
		// this is a test for permission sets - if UsePermissionSets = true, for running init roles twice
		BeforeAll(func() {
			newLegacyRoles := []models.Role{}
			for i := range LegacyRoles {
				if LegacyRoles[i].Name == "L1 researcher" {
					role := LegacyRoles[i]
					newPermissions := []*models.Permission{}
					for _, resourceType := range api.AllowedResourceTypeEnumValues {
						if resourceType != api.RESOURCETYPE_TENANT {
							newPermissions = append(newPermissions, &models.Permission{
								ResourceType: resourceType,
								Sync:         true,
							})
						}
					}
					role.Permissions = newPermissions
					newLegacyRoles = append(newLegacyRoles, role)
				} else {
					newLegacyRoles = append(newLegacyRoles, LegacyRoles[i])
				}
			}
			LegacyRoles = newLegacyRoles

			newRoles := []models.Role{}
			for i := range RolesMigrated {
				if RolesMigrated[i].Name == "System administrator" {
					role := RolesMigrated[i]
					role.PermissionSets = &[]models.PermissionSet{
						{Name: static_permission_sets.AccountEdit.Name},
						{Name: static_permission_sets.ClusterEdit.Name},
						{Name: static_permission_sets.WorkloadRead.Name},
					}
					newRoles = append(newRoles, role)
				} else {
					newRoles = append(newRoles, RolesMigrated[i])
				}
			}
			RolesMigrated = newRoles
		})

		It("test after init that legacy roles have permission set", func() {
			InitStaticLegacyRoles()
			roles, err := DbGetRoles(context.Background(), "", false, 100, 0, "", "", nil, "", false)
			Expect(err).To(BeNil())
			for _, role := range *roles {
				if role.Name == "L1 researcher" {
					for _, perm := range role.Permissions {
						if perm.ResourceType == api.RESOURCETYPE_TENANT {
							Expect(perm.Create).To(BeFalse())
							Expect(perm.Read).To(BeFalse())
							Expect(perm.Update).To(BeFalse())
							Expect(perm.Delete).To(BeFalse())
							Expect(perm.Sync).To(BeFalse())
						} else {
							Expect(perm.Create).To(BeFalse())
							Expect(perm.Read).To(BeFalse())
							Expect(perm.Update).To(BeFalse())
							Expect(perm.Delete).To(BeFalse())
							Expect(perm.Sync).To(BeTrue())
						}
					}
				}
			}
		})

		It("test after init that new roles have permission set", func() {
			// before second run - System administrator has 30 permission sets
			roles, err := DbGetRoles(context.Background(), "", false, 100, 0, "", "", nil, "", false)
			Expect(err).To(BeNil())
			for _, role := range *roles {
				if role.Name == "System administrator" {
					Expect(len(*role.PermissionSets)).To(Equal(30))
					for _, permission := range role.Permissions {
						switch permission.ResourceType {
						case api.RESOURCETYPE_WORKLOADS,
							api.RESOURCETYPE_WORKSPACES,
							api.RESOURCETYPE_TRAININGS,
							api.RESOURCETYPE_INFERENCES,
							api.RESOURCETYPE_WORKLOAD_PROPERTIES,
							api.RESOURCETYPE_PROJECT:
							Expect(permission.Create).To(BeTrue())
							Expect(permission.Update).To(BeTrue())
							Expect(permission.Read).To(BeTrue())
							Expect(permission.Delete).To(BeTrue())
							Expect(permission.Sync).To(BeFalse())
							break
						case api.RESOURCETYPE_CLUSTER:
							Expect(permission.Create).To(BeTrue())
							Expect(permission.Update).To(BeTrue())
							Expect(permission.Read).To(BeTrue())
							Expect(permission.Delete).To(BeTrue())
							Expect(permission.Sync).To(BeFalse())
							break
						case api.RESOURCETYPE_TENANT:
							Expect(permission.Create).To(BeFalse())
							Expect(permission.Update).To(BeFalse())
							Expect(permission.Read).To(BeTrue())
							Expect(permission.Delete).To(BeFalse())
							Expect(permission.Sync).To(BeFalse())
							break
						}
					}
				}
			}
			InitStaticRoles()
			// before second run - System administrator has 3 permission sets
			roles, err = DbGetRoles(context.Background(), "", false, 100, 0, "", "", nil, "", false)
			Expect(err).To(BeNil())
			for _, role := range *roles {
				if role.Name == "System administrator" {
					Expect(len(*role.PermissionSets)).To(Equal(3))
					for _, permission := range role.Permissions {
						switch permission.ResourceType {
						case api.RESOURCETYPE_WORKLOADS,
							api.RESOURCETYPE_WORKSPACES,
							api.RESOURCETYPE_TRAININGS,
							api.RESOURCETYPE_INFERENCES,
							api.RESOURCETYPE_WORKLOAD_PROPERTIES,
							api.RESOURCETYPE_PROJECT:
							Expect(permission.Create).To(BeFalse())
							Expect(permission.Update).To(BeFalse())
							Expect(permission.Read).To(BeTrue())
							Expect(permission.Delete).To(BeFalse())
							Expect(permission.Sync).To(BeFalse())
							break
						case api.RESOURCETYPE_CLUSTER,
							api.RESOURCETYPE_TENANT:
							Expect(permission.Create).To(BeTrue())
							Expect(permission.Update).To(BeTrue())
							Expect(permission.Read).To(BeTrue())
							Expect(permission.Delete).To(BeTrue())
							Expect(permission.Sync).To(BeFalse())
							break
						case api.RESOURCETYPE_CLUSTERS_MINIMAL:
							Expect(permission.Create).To(BeFalse())
							Expect(permission.Update).To(BeFalse())
							Expect(permission.Read).To(BeTrue())
							Expect(permission.Delete).To(BeFalse())
							Expect(permission.Sync).To(BeFalse())
							break
						case api.RESOURCETYPE_NODEPOOLS_MINIMAL:
							Expect(permission.Create).To(BeFalse())
							Expect(permission.Update).To(BeFalse())
							Expect(permission.Read).To(BeTrue())
							Expect(permission.Delete).To(BeFalse())
							Expect(permission.Sync).To(BeFalse())
							break
						default:
							Expect(permission.Create).To(BeFalse())
							Expect(permission.Update).To(BeFalse())
							Expect(permission.Read).To(BeFalse())
							Expect(permission.Delete).To(BeFalse())
							Expect(permission.Sync).To(BeFalse())
						}
					}
				}
			}
		})
	})

	Context("init roles and permission set on upgrade from no permission set", func() {
		// this is a test for permission sets - if UsePermissionSets = true, for running init roles first time after permissionSet change added
		BeforeAll(func() {
			var legacyForUpdate models.Role
			var migratedForUpdate models.Role
			roles, err := DbGetRoles(context.Background(), "", false, 100, 0, "", "", nil, "", false)
			Expect(err).To(BeNil())
			for _, role := range *roles {
				if role.Name == "L1 researcher" {
					legacyForUpdate = role
				}
				if role.Name == "Viewer" {
					migratedForUpdate = role
				}
			}
			legacyForUpdate.PermissionSets = nil
			legacyForUpdate.Permissions = []*models.Permission{{
				ResourceType: api.RESOURCETYPE_TENANT,
				Create:       true,
			}}
			migratedForUpdate.PermissionSets = nil
			migratedForUpdate.Permissions = []*models.Permission{{
				ResourceType: api.RESOURCETYPE_TENANT,
				Create:       true,
			}}
			var rowsAffected int64
			var dbErr error
			rowsAffected, dbErr = runaidb.DB().Session(&gorm.Session{FullSaveAssociations: true}).Save(&legacyForUpdate)
			Expect(dbErr).To(BeNil())
			Expect(rowsAffected).NotTo(Equal(int64(0)))
			rowsAffected, dbErr = runaidb.DB().Session(&gorm.Session{FullSaveAssociations: true}).Save(&migratedForUpdate)
			Expect(dbErr).To(BeNil())
			Expect(rowsAffected).NotTo(Equal(int64(0)))
			rowsAffected, dbErr = runaidb.DB().Exec(fmt.Sprintf("delete from role_to_permission_set where role_id in ('%v','%v')", legacyForUpdate.ID, migratedForUpdate.ID))
			Expect(dbErr).To(BeNil())
			Expect(rowsAffected).NotTo(Equal(int64(0)))
			rowsAffected, dbErr = runaidb.DB().Where("name = ?", "L1 researcher-internal-ps").Delete(&models.PermissionSet{})
			Expect(dbErr).To(BeNil())
			Expect(rowsAffected).NotTo(Equal(int64(0)))
		})

		It("test legacy update after roles don't have permission set", func() {
			roles, err := DbGetRoles(context.Background(), "", false, 100, 0, "", "", nil, "", false)
			Expect(err).To(BeNil())
			for _, role := range *roles {
				if role.Name == "L1 researcher" {
					Expect(len(*role.PermissionSets)).To(Equal(0))
					for _, perm := range role.Permissions {
						if perm.ResourceType == api.RESOURCETYPE_TENANT {
							Expect(perm.Create).To(BeTrue())
							Expect(perm.Read).To(BeFalse())
							Expect(perm.Update).To(BeFalse())
							Expect(perm.Delete).To(BeFalse())
							Expect(perm.Sync).To(BeFalse())
						} else {
							Expect(perm.Create).To(BeFalse())
							Expect(perm.Read).To(BeFalse())
							Expect(perm.Update).To(BeFalse())
							Expect(perm.Delete).To(BeFalse())
							Expect(perm.Sync).To(BeFalse())
						}
					}
				}
			}
			InitStaticLegacyRoles()
			roles, err = DbGetRoles(context.Background(), "", false, 100, 0, "", "", nil, "", false)
			Expect(err).To(BeNil())
			for _, role := range *roles {
				if role.Name == "L1 researcher" {
					Expect(len(*role.PermissionSets)).To(Equal(1))
					for _, perm := range role.Permissions {
						if perm.ResourceType == api.RESOURCETYPE_TENANT {
							Expect(perm.Create).To(BeFalse())
							Expect(perm.Read).To(BeTrue())
							Expect(perm.Update).To(BeFalse())
							Expect(perm.Delete).To(BeFalse())
							Expect(perm.Sync).To(BeFalse())
						}
					}
				}
			}
		})

		It("test upgraded update after roles don't have permission set", func() {
			roles, err := DbGetRoles(context.Background(), "", false, 100, 0, "", "", nil, "", false)
			Expect(err).To(BeNil())
			for _, role := range *roles {
				if role.Name == "Viewer" {
					Expect(len(*role.PermissionSets)).To(Equal(0))
				}
			}
			InitStaticRoles()
			roles, err = DbGetRoles(context.Background(), "", false, 100, 0, "", "", nil, "", false)
			Expect(err).To(BeNil())
			for _, role := range *roles {
				if role.Name == "Viewer" {
					Expect(len(*role.PermissionSets)).To(Equal(18))
				}
			}
		})
	})

	Context("init permission sets on upgrade", func() {
		BeforeAll(func() {
			newPermissionSets := []models.PermissionSet{}
			for i := range PermissionSets {
				if PermissionSets[i].Name == static_permission_sets.WorkspaceEdit.Name {
					ps := PermissionSets[i]
					newPermissions := []models.RolePermission{}
					for _, resourceType := range api.AllowedResourceTypeEnumValues {
						if resourceType != api.RESOURCETYPE_PROJECT {
							newPermissions = append(newPermissions, models.RolePermission{
								ResourceType: resourceType,
								Sync:         true,
							})
						}
					}
					ps.Permissions = newPermissions
					newPermissionSets = append(newPermissionSets, ps)
				} else {
					newPermissionSets = append(newPermissionSets, PermissionSets[i])
				}
			}
			PermissionSets = newPermissionSets
		})

		It("test after init permission sets updated", func() {
			InitStaticPermissionSets()
			permissionSets, err := DbGetPermissionSets(context.Background(), true)
			Expect(err).To(BeNil())
			for _, ps := range *permissionSets {
				if ps.Name == static_permission_sets.WorkspaceEdit.Name {
					for _, perm := range ps.Permissions {
						if perm.ResourceType == api.RESOURCETYPE_PROJECT {
							Expect(perm.Create).To(BeFalse())
							Expect(perm.Read).To(BeFalse())
							Expect(perm.Update).To(BeFalse())
							Expect(perm.Delete).To(BeFalse())
							Expect(perm.Sync).To(BeFalse())
						} else {
							Expect(perm.Create).To(BeFalse())
							Expect(perm.Read).To(BeFalse())
							Expect(perm.Update).To(BeFalse())
							Expect(perm.Delete).To(BeFalse())
							Expect(perm.Sync).To(BeTrue())
						}
					}
				}
			}
		})
	})

	Context("validate nodepools-minimal permission", func() {
		It("should have nodepools-minimal read if role has department or project read permission", func() {
			// Collect all static roles
			allRoles := []models.Role{}
			allRoles = append(allRoles, LegacyRoles...)
			allRoles = append(allRoles, RolesMigrated...)
			allRoles = append(allRoles, RolesNew...)
			allRoles = append(allRoles, static_roles.ApplicationAdmin)
			allRoles = append(allRoles, static_roles.CloudSupport)
			// Include commented-out roles that might be enabled later
			allRoles = append(allRoles, static_roles.AiPractitioner)
			allRoles = append(allRoles, static_roles.DataAndStorageAdmin)
			allRoles = append(allRoles, static_roles.ProjectAdmin)

			for _, role := range allRoles {
				hasDepartmentRead := false
				hasProjectRead := false
				hasNodepoolsMinimalRead := false

				// Check direct permissions
				for _, perm := range role.Permissions {
					if perm.ResourceType == api.RESOURCETYPE_DEPARTMENT && perm.Read {
						hasDepartmentRead = true
					}
					if perm.ResourceType == api.RESOURCETYPE_PROJECT && perm.Read {
						hasProjectRead = true
					}
					if perm.ResourceType == api.RESOURCETYPE_NODEPOOLS_MINIMAL && perm.Read {
						hasNodepoolsMinimalRead = true
					}
				}

				// If role has department or project read, it must have nodepools-minimal read
				if hasDepartmentRead || hasProjectRead {
					Expect(hasNodepoolsMinimalRead).To(BeTrue(),
						fmt.Sprintf("Role '%s' has department or project read permission but is missing nodepools-minimal read permission", role.Name))
				}
			}
		})
	})
})

var _ = Describe("Mark legacy roles deprecated on upgrade", func() {
	// Simulates upgrading a tenant whose legacy role already exists in the DB but
	// predates the Deprecated flag (so it is stored with deprecated=false). Re-running
	// the static-role initialization must reconcile the flag to true.
	getDeprecated := func(roleName string) bool {
		role := &models.Role{}
		err := runaidb.DB().
			Where("name = ?", roleName).
			Where("scope_type = ?", models.ROLE_SCOPE_SYSTEM).
			First(role)
		Expect(err).To(BeNil())
		return role.Deprecated
	}

	It("re-marks an existing legacy role as deprecated", func() {
		// The role exists from suite init; force the pre-upgrade state where it is not deprecated.
		rowsAffected, err := runaidb.DB().Model(&models.Role{}).
			Where("name = ?", static_legacy_roles.ResearchManager.Name).
			Where("scope_type = ?", models.ROLE_SCOPE_SYSTEM).
			Update("deprecated", false)
		Expect(err).To(BeNil())
		Expect(rowsAffected).NotTo(BeZero())
		Expect(getDeprecated(static_legacy_roles.ResearchManager.Name)).To(BeFalse())
		Expect(getDeprecated(static_roles.SystemAdmin.Name)).To(BeFalse())

		// Upgrade reconciliation.
		InitStaticLegacyRoles()

		Expect(getDeprecated(static_legacy_roles.ResearchManager.Name)).To(BeTrue())
		Expect(getDeprecated(static_roles.SystemAdmin.Name)).To(BeFalse())
	})
})

var _ = Describe("System administrator cluster-config under CSP", func() {
	// In CSP deployments the System administrator's cluster-config permission must be
	// downgraded to read-only, mirroring the pre-refactor ReadOnlyForCSP behavior. This
	// is achieved by swapping the "Cluster config edit" permission set for the read-only
	// "Cluster config read" set via CspRoleChange.
	It("swaps Cluster config edit for Cluster config read", func() {
		changeMap := CspRoleChange[static_roles.SystemAdmin.Name]
		Expect(changeMap[static_permission_sets.ClusterConfigEdit.Name]).To(Equal(static_permission_sets.ClusterConfigRead.Name))

		cspPermissionSets := getCspPermissionSets(static_roles.SystemAdmin.PermissionSets, changeMap)

		names := map[string]bool{}
		for _, ps := range cspPermissionSets {
			names[ps.Name] = true
		}
		Expect(names[static_permission_sets.ClusterConfigRead.Name]).To(BeTrue(), "CSP System administrator should hold Cluster config read")
		Expect(names[static_permission_sets.ClusterConfigEdit.Name]).To(BeFalse(), "CSP System administrator should not hold Cluster config edit")
	})

	It("keeps the cluster-config permission sets internal (not user-selectable)", func() {
		Expect(static_permission_sets.ClusterConfigEdit.Internal).To(BeTrue())
		Expect(static_permission_sets.ClusterConfigRead.Internal).To(BeTrue())
	})

	It("grants only read on cluster-config in the Cluster config read set", func() {
		var found bool
		for _, permission := range static_permission_sets.ClusterConfigRead.Permissions {
			if permission.ResourceType == api.RESOURCETYPE_CLUSTER_CONFIG {
				found = true
				Expect(permission.Read).To(BeTrue())
				Expect(permission.Create).To(BeFalse())
				Expect(permission.Update).To(BeFalse())
				Expect(permission.Delete).To(BeFalse())
			}
		}
		Expect(found).To(BeTrue(), "Cluster config read should grant a cluster-config permission")
	})
})
