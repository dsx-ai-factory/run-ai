/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scopes_sync

import (
	"context"
	"fmt"
	ginkgoTesting "testing"
	"time"

	static_roles "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-roles"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	testing "github.com/run-ai/runai/control-plane/authorization/pkg/test"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

const (
	TestUserId = "user@run.ai"
)

func TestResourcesSync(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	testing.SetupPostgres(t, "file://../database/migrations")
	if err := authorization_config.Init(); err != nil {
		t.Fatalf("failed to initialize config: %v", err)
	}
	dao.InitStaticPermissionSets()
	dao.InitStaticLegacyRoles()
	dao.InitStaticRoles()

	RunSpecs(t, "scopes sync test")
}

var _ = Describe("test sync job", Ordered, func() {
	Context("test sync job - delete ghost access rules", Ordered, func() {
		var (
			deletedProjAr,
			keptProjAr,
			deletedDeptAr,
			keptDeptAr,
			deletedClusterAr,
			keptClusterAr uint

			tenant2ProjAr, tenant2DeptAr, tenant2ClusterAr uint

			roleId             = int32(dao.MapStaticRoleNameToRoleId[static_roles.Viewer.Name])
			clusterUUID        = uuid.NewString()
			projectToDelete    = "proj-del-1"
			projectToKeep      = "proj-keep-1"
			departmentToDelete = "dept-del-1"
			departmentToKeep   = "dept-keep-1"
			clusterToDelete    = uuid.NewString()
			clusterToKeep      = uuid.NewString()

			projectTenant2    = "proj-tenant2"
			departmentTenant2 = "dept-tenant2"
			clusterTenant2    = uuid.NewString()
		)
		ctx := setContext(fmt.Sprintf("%d", testing.TestTenant), TestUserId, []string{})

		BeforeAll(func() {
			deletedProjAr = createAccessRule(ctx, clusterUUID, projectToDelete, api.SCOPETYPE_PROJECT, roleId, testing.TestTenant)
			keptProjAr = createAccessRule(ctx, clusterUUID, projectToKeep, api.SCOPETYPE_PROJECT, roleId, testing.TestTenant)
			deletedDeptAr = createAccessRule(ctx, clusterUUID, departmentToDelete, api.SCOPETYPE_DEPARTMENT, roleId, testing.TestTenant)
			keptDeptAr = createAccessRule(ctx, clusterUUID, departmentToKeep, api.SCOPETYPE_DEPARTMENT, roleId, testing.TestTenant)
			deletedClusterAr = createAccessRule(ctx, clusterToDelete, clusterToDelete, api.SCOPETYPE_CLUSTER, roleId, testing.TestTenant)
			keptClusterAr = createAccessRule(ctx, clusterToKeep, clusterToKeep, api.SCOPETYPE_CLUSTER, roleId, testing.TestTenant)
			tenant2ProjAr = createAccessRule(ctx, clusterTenant2, projectTenant2, api.SCOPETYPE_PROJECT, roleId, 2)
			tenant2DeptAr = createAccessRule(ctx, clusterTenant2, departmentTenant2, api.SCOPETYPE_DEPARTMENT, roleId, 2)
			tenant2ClusterAr = createAccessRule(ctx, clusterTenant2, clusterTenant2, api.SCOPETYPE_CLUSTER, roleId, 2)
		})

		It("marks access rules deleted for provided org unit ids", func() {
			deletedAfter := time.Now()
			deleteGhostProjectsAccessRules(ctx, fmt.Sprintf("%d", testing.TestTenant), []string{projectToDelete, projectTenant2}, deletedAfter)
			deleteGhostDepartmentsAccessRules(ctx, fmt.Sprintf("%d", testing.TestTenant), []string{departmentToDelete, departmentTenant2}, deletedAfter)
			deleteGhostClustersAccessRules(ctx, fmt.Sprintf("%d", testing.TestTenant), []string{clusterToDelete, clusterTenant2}, deletedAfter)

			// Delete only for requested org-units
			assertDeleted(deletedProjAr, true)
			assertDeleted(keptProjAr, false)
			assertDeleted(deletedDeptAr, true)
			assertDeleted(keptDeptAr, false)
			assertDeleted(deletedClusterAr, true)
			assertDeleted(keptClusterAr, false)

			// Delete only tenant's access rules
			assertDeleted(tenant2ProjAr, false)
			assertDeleted(tenant2DeptAr, false)
			assertDeleted(tenant2ClusterAr, false)
		})
	})
})

func setContext(tenantId string, subjectId string, groups []string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	ctx = context.WithValue(ctx, runaictx.CfGroups, groups)
	return ctx
}

func createAccessRule(ctx context.Context, clusterId string, scopeId string, scopeType api.ScopeType, roleId int32, tenantId uint) uint {
	ar, err := dao.DbCreateAccessRule(ctx, tenantId, TestUserId, "name-"+scopeId, clusterId, false, api.AccessRuleCreationFields{
		ScopeId:     scopeId,
		ScopeType:   scopeType,
		SubjectId:   TestUserId,
		SubjectType: api.SUBJECTTYPE_USER,
		RoleId:      roleId,
	})
	Expect(err).To(BeNil())
	return ar.ID
}

func assertDeleted(id uint, expectDeleted bool) {
	var ar models.AccessRule
	err := runaidb.DB().Unscoped().Where("id = ?", id).First(&ar)
	Expect(err).To(BeNil())
	if expectDeleted {
		Expect(ar.DeletedAt.Valid).To(BeTrue())
	} else {
		Expect(ar.DeletedAt.Valid).To(BeFalse())
	}
}
