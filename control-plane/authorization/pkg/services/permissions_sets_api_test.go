/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/google/uuid"

	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
)

var _ = Describe("Permissions sets API tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	ctx := setContext(tenantId, TestUserId, []string{dao.SuperAdmins})
	PermissionSetMap := map[string]uuid.UUID{}
	permissionSetLength := 57
	workspaceEditPermissionsLen := 22

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
		// granularSet builds the expected shape of an asset-granular permission
		// set: one or more primary asset resource types (all sharing the same
		// editability), followed by the cluster/nodepool minimals. Compute
		// resources and resource-claim-templates are managed as a paired asset
		// grant, so ComputeResourcesOnly{Read,Edit} declares both primaries.
		granularSet := func(ps models.PermissionSet, resourceTypes []api.ResourceType, editable bool) (models.PermissionSet, []expectedPermission) {
			primaries := make([]expectedPermission, 0, len(resourceTypes))
			for _, rt := range resourceTypes {
				primaries = append(primaries, expectedPermission{
					resourceType: rt,
					create:       editable,
					read:         true,
					update:       editable,
					delete:       editable,
				})
			}
			return ps, append(primaries, minimals...)
		}

		granularSets := map[string][]expectedPermission{}
		granularDefinitions := []models.PermissionSet{}
		for _, entry := range []struct {
			ps            models.PermissionSet
			resourceTypes []api.ResourceType
			editable      bool
		}{
			{static_permission_sets.TemplatesOnlyRead, []api.ResourceType{api.RESOURCETYPE_TEMPLATES}, false},
			{static_permission_sets.TemplatesOnlyEdit, []api.ResourceType{api.RESOURCETYPE_TEMPLATES}, true},
			{static_permission_sets.EnvironmentsOnlyRead, []api.ResourceType{api.RESOURCETYPE_ENVIRONMENTS}, false},
			{static_permission_sets.EnvironmentsOnlyEdit, []api.ResourceType{api.RESOURCETYPE_ENVIRONMENTS}, true},
			{static_permission_sets.ComputeResourcesOnlyRead, []api.ResourceType{api.RESOURCETYPE_COMPUTE_RESOURCES, api.RESOURCETYPE_RESOURCE_CLAIM_TEMPLATES}, false},
			{static_permission_sets.ComputeResourcesOnlyEdit, []api.ResourceType{api.RESOURCETYPE_COMPUTE_RESOURCES, api.RESOURCETYPE_RESOURCE_CLAIM_TEMPLATES}, true},
			{static_permission_sets.CredentialsOnlyRead, []api.ResourceType{api.RESOURCETYPE_CREDENTIALS}, false},
			{static_permission_sets.CredentialsOnlyEdit, []api.ResourceType{api.RESOURCETYPE_CREDENTIALS}, true},
		} {
			ps, expected := granularSet(entry.ps, entry.resourceTypes, entry.editable)
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
			Expect(static_permission_sets.TemplatesRead.Permissions).To(HaveLen(14))
			Expect(static_permission_sets.TemplatesEdit.Permissions).To(HaveLen(14))
		})
	})
})
