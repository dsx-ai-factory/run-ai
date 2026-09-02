/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/common"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
)

func createTestTemplate(tenantId uint, name string, scopeType api.ScopeType, scopeId string, clusterId, departmentId, projectId *string) string {
	templateId := uuid.New()
	now := time.Now()

	rec := &models_crud.Template{
		ID: templateId,
		TemplateMeta: models_crud.TemplateMeta{
			TenantId:     tenantId,
			WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
			Name:         name,
			ScopeType:    string(scopeType),
			ScopeId:      scopeId,
			ClusterId:    clusterId,
			DepartmentId: departmentId,
			ProjectId:    projectId,
			CreatedBy:    testing.TestUser,
			CreatedAt:    now,
			UpdatedBy:    testing.TestUser,
			UpdatedAt:    now,
		},
		Spec:               dao_utils.JsonBMap{"image": "ubuntu"},
		CombinedSpec:       dao_utils.JsonBMap{"image": "ubuntu"},
		CombinedMasterSpec: dao_utils.JsonBMap{},
		Runnable:           true,
	}

	err := dao_crud.CreateTemplate(context.Background(), rec, nil)
	Expect(err).To(BeNil())
	return templateId.String()
}

func createTestTemplateAsset(tenantId uint, templateId string, assetId string) {
	rec := models_crud.TemplateAssets{
		ID:          uuid.New(),
		TenantId:    tenantId,
		TemplateId:  templateId,
		AssetId:     assetId,
		ReplicaType: string(api.REPLICATYPE_WORKER),
		Category:    "environment",
	}
	result := dao_utils.DB().Create(&rec)
	Expect(result.Error).To(BeNil())
}

func createTestPolicyCompliance(tenantId int32, projectId int32, templateId string) {
	rec := models_crud.PolicyComplianceRecord{
		TenantId:     tenantId,
		ProjectId:    projectId,
		TemplateId:   templateId,
		WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
		PolicyTime:   time.Now(),
		Aligned:      true,
	}
	result := dao_utils.DB().Create(&rec)
	Expect(result.Error).To(BeNil())
}

func countTemplates(tenantId uint) int64 {
	var count int64
	dao_utils.DB().Table(models_crud.TemplatesTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.TemplatesColumnTenantId), tenantId).
		Count(&count)
	return count
}

func countTemplateAssets(tenantId uint) int64 {
	var count int64
	dao_utils.DB().Table(models_crud.TemplatesAssetsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.TemplatesAssetsColumnTenantId), tenantId).
		Count(&count)
	return count
}

func countPolicyCompliance(tenantId int32) int64 {
	var count int64
	dao_utils.DB().Model(models_crud.PolicyComplianceRecord{}).
		Where("tenant_id = ?", tenantId).
		Count(&count)
	return count
}

func countPolicyComplianceForProject(tenantId int32, projectId int32) int64 {
	var count int64
	dao_utils.DB().Model(models_crud.PolicyComplianceRecord{}).
		Where("tenant_id = ? AND project_id = ?", tenantId, projectId).
		Count(&count)
	return count
}

var _ = Describe("pubsub - template cascade cleanup on scope deletion events", func() {
	var randoms *testutils.RandomValues
	var ctx context.Context

	var clusterId1, clusterId2 string
	var departmentId1, departmentId2 string
	var projectId1, projectId2 string

	AfterEach(func() {
		testing.CleanTables()
	})

	BeforeEach(func() {
		// Use a unique tenant id per spec; NewRandomValues alone shares one across the process.
		randoms = testutils.NewRandomValues().CloneWithDifferentTenant()
		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		clusterId1 = uuid.NewString()
		clusterId2 = uuid.NewString()
		departmentId1 = "dept-1"
		departmentId2 = "dept-2"
		projectId1 = "100001"
		projectId2 = "100002"
	})

	Context("project deletion", func() {
		It("removes project templates and cleans up asset mappings and compliance records", func() {
			t1 := createTestTemplate(randoms.TenantId, "tmpl-proj1-a", api.SCOPETYPE_PROJECT, projectId1,
				&clusterId1, &departmentId1, &projectId1)
			t2 := createTestTemplate(randoms.TenantId, "tmpl-proj1-b", api.SCOPETYPE_PROJECT, projectId1,
				&clusterId1, &departmentId1, &projectId1)
			createTestTemplate(randoms.TenantId, "tmpl-proj2", api.SCOPETYPE_PROJECT, projectId2,
				&clusterId1, &departmentId1, &projectId2)
			createTestTemplate(randoms.TenantId, "tmpl-cluster", api.SCOPETYPE_CLUSTER, clusterId1,
				&clusterId1, nil, nil)
			createTestTemplate(randoms.TenantId, "tmpl-dept", api.SCOPETYPE_DEPARTMENT, departmentId1,
				&clusterId1, &departmentId1, nil)

			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(5)))

			createTestTemplateAsset(randoms.TenantId, t1, uuid.NewString())
			createTestTemplateAsset(randoms.TenantId, t1, uuid.NewString())
			createTestTemplateAsset(randoms.TenantId, t1, uuid.NewString())
			createTestTemplateAsset(randoms.TenantId, t2, uuid.NewString())
			createTestPolicyCompliance(int32(randoms.TenantId), 100001, t1)
			createTestPolicyCompliance(int32(randoms.TenantId), 100002, t1)
			createTestPolicyCompliance(int32(randoms.TenantId), 100003, t1)

			err := common.HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ClusterId: clusterId1,
				ProjectId: projectId1,
			})
			Expect(err).To(BeNil())

			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(3)))
			Expect(countTemplateAssets(randoms.TenantId)).To(Equal(int64(0)))
			Expect(countPolicyCompliance(int32(randoms.TenantId))).To(Equal(int64(0)))

			_, getErr := dao_crud.GetTemplateById(ctx, randoms.TenantId, t1, nil)
			Expect(getErr).NotTo(BeNil())
			Expect(getErr.Code).To(Equal(int32(404)))
		})

		It("skips cleanup when LeaveOrphans is true", func() {
			createTestTemplate(randoms.TenantId, "tmpl-orphan", api.SCOPETYPE_PROJECT, projectId1,
				&clusterId1, &departmentId1, &projectId1)

			err := common.HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ClusterId:    clusterId1,
				ProjectId:    projectId1,
				LeaveOrphans: true,
			})
			Expect(err).To(BeNil())
			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(1)))
		})

		It("cleans up compliance records for ancestor-scoped templates when project has no own templates", func() {
			tenantTmpl := createTestTemplate(randoms.TenantId, "tmpl-tenant", api.SCOPETYPE_TENANT, randoms.TenantIdStr,
				nil, nil, nil)
			clusterTmpl := createTestTemplate(randoms.TenantId, "tmpl-cluster", api.SCOPETYPE_CLUSTER, clusterId1,
				&clusterId1, nil, nil)

			projIdInt, _ := strconv.Atoi(projectId1)
			createTestPolicyCompliance(int32(randoms.TenantId), int32(projIdInt), tenantTmpl)
			createTestPolicyCompliance(int32(randoms.TenantId), int32(projIdInt), clusterTmpl)

			otherProjIdInt, _ := strconv.Atoi(projectId2)
			createTestPolicyCompliance(int32(randoms.TenantId), int32(otherProjIdInt), tenantTmpl)

			Expect(countPolicyCompliance(int32(randoms.TenantId))).To(Equal(int64(3)))

			err := common.HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ClusterId: clusterId1,
				ProjectId: projectId1,
			})
			Expect(err).To(BeNil())

			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(2)))
			Expect(countPolicyCompliance(int32(randoms.TenantId))).To(Equal(int64(1)))
		})

		It("does not affect other tenants and succeeds when project has no templates", func() {
			tenant2Id := randoms.TenantId + 1

			createTestTemplate(tenant2Id, "tmpl-other-tenant", api.SCOPETYPE_PROJECT, projectId1,
				&clusterId1, &departmentId1, &projectId1)
			createTestTemplate(randoms.TenantId, "tmpl-other-project", api.SCOPETYPE_PROJECT, projectId2,
				&clusterId1, &departmentId1, &projectId2)

			err := common.HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ClusterId: clusterId1,
				ProjectId: projectId1,
			})
			Expect(err).To(BeNil())

			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(1)))
			Expect(countTemplates(tenant2Id)).To(Equal(int64(1)))
		})
	})

	Context("cluster deletion", func() {
		It("removes all cluster, department, and project scoped templates under the cluster", func() {
			createTestTemplate(randoms.TenantId, "tmpl-cluster-scope", api.SCOPETYPE_CLUSTER, clusterId1,
				&clusterId1, nil, nil)
			createTestTemplate(randoms.TenantId, "tmpl-dept-scope", api.SCOPETYPE_DEPARTMENT, departmentId1,
				&clusterId1, &departmentId1, nil)
			t3 := createTestTemplate(randoms.TenantId, "tmpl-proj-scope", api.SCOPETYPE_PROJECT, projectId1,
				&clusterId1, &departmentId1, &projectId1)
			createTestTemplate(randoms.TenantId, "tmpl-proj2-scope", api.SCOPETYPE_PROJECT, projectId2,
				&clusterId1, &departmentId2, &projectId2)

			createTestTemplate(randoms.TenantId, "tmpl-other-cluster", api.SCOPETYPE_CLUSTER, clusterId2,
				&clusterId2, nil, nil)
			createTestTemplate(randoms.TenantId, "tmpl-tenant-scope", api.SCOPETYPE_TENANT, randoms.TenantIdStr,
				nil, nil, nil)

			createTestTemplateAsset(randoms.TenantId, t3, uuid.NewString())
			createTestPolicyCompliance(int32(randoms.TenantId), 100001, t3)

			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(6)))

			err := common.HandleClusterDeleteRequest(ctx, &events.ClusterDelete{
				ClusterId: clusterId1,
			})
			Expect(err).To(BeNil())

			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(2)))
			Expect(countTemplateAssets(randoms.TenantId)).To(Equal(int64(0)))
			Expect(countPolicyCompliance(int32(randoms.TenantId))).To(Equal(int64(0)))
		})

		It("preserves templates belonging to other clusters", func() {
			createTestTemplate(randoms.TenantId, "tmpl-p1-c1", api.SCOPETYPE_PROJECT, projectId1,
				&clusterId1, &departmentId1, &projectId1)
			createTestTemplate(randoms.TenantId, "tmpl-p2-c1", api.SCOPETYPE_PROJECT, projectId2,
				&clusterId1, &departmentId2, &projectId2)
			createTestTemplate(randoms.TenantId, "tmpl-p1-c2", api.SCOPETYPE_PROJECT, projectId1,
				&clusterId2, &departmentId1, &projectId1)

			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(3)))

			err := common.HandleClusterDeleteRequest(ctx, &events.ClusterDelete{
				ClusterId: clusterId1,
			})
			Expect(err).To(BeNil())

			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(1)))
		})
	})

	Context("tenant deletion", func() {
		It("removes all templates across all scopes for the tenant", func() {
			tenant2Id := randoms.TenantId + 1

			createTestTemplate(randoms.TenantId, "tmpl-tenant-a", api.SCOPETYPE_TENANT, randoms.TenantIdStr,
				nil, nil, nil)
			createTestTemplate(randoms.TenantId, "tmpl-cluster-a", api.SCOPETYPE_CLUSTER, clusterId1,
				&clusterId1, nil, nil)
			t3 := createTestTemplate(randoms.TenantId, "tmpl-proj-a", api.SCOPETYPE_PROJECT, projectId1,
				&clusterId1, &departmentId1, &projectId1)

			createTestTemplate(tenant2Id, "tmpl-other-tenant", api.SCOPETYPE_TENANT, strconv.Itoa(int(tenant2Id)),
				nil, nil, nil)

			createTestTemplateAsset(randoms.TenantId, t3, uuid.NewString())
			createTestPolicyCompliance(int32(randoms.TenantId), 100001, t3)

			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(3)))
			Expect(countTemplates(tenant2Id)).To(Equal(int64(1)))

			err := common.HandleTenantDeleteRequest(ctx, &events.TenantDelete{
				TenantId: randoms.TenantIdStr,
			})
			Expect(err).To(BeNil())

			Expect(countTemplates(randoms.TenantId)).To(Equal(int64(0)))
			Expect(countTemplateAssets(randoms.TenantId)).To(Equal(int64(0)))
			Expect(countPolicyCompliance(int32(randoms.TenantId))).To(Equal(int64(0)))
			Expect(countTemplates(tenant2Id)).To(Equal(int64(1)))
		})
	})
})
