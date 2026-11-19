/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"k8s.io/utils/ptr"
)

const (
	sqlSelectDeptNodePoolResources           = `SELECT * FROM department_node_pool_resources_view WHERE tenant_id = ? ORDER BY id DESC LIMIT ? OFFSET ?`
	sqlSelectProjNodePoolResourcesByGpuLimit = `SELECT * FROM project_node_pool_resources_view WHERE tenant_id = ? ORDER BY COALESCE(gpu_limit, -1) ASC NULLS FIRST,id DESC LIMIT ? OFFSET ?`
	sqlSelectProjNodePoolResources           = `SELECT * FROM project_node_pool_resources_view WHERE tenant_id = ? ORDER BY id DESC LIMIT ? OFFSET ?`
	sqlCountDeptNodePoolResources            = `SELECT count(*) FROM department_node_pool_resources_view WHERE tenant_id = ?`
	sqlCountProjNodePoolResources            = `SELECT count(*) FROM project_node_pool_resources_view WHERE tenant_id = ?`
	sqlSumProjGPUWeights                     = `SELECT COALESCE(SUM(gpu_over_quota_weight), 0) FROM project_node_pool_resources_view WHERE tenant_id = ? AND node_pool_id = ?`
	sqlSumProjGPUDeserved                    = `SELECT COALESCE(SUM(CASE WHEN gpu_deserved >= 0 THEN gpu_deserved ELSE 0 END), 0) FROM project_node_pool_resources_view WHERE tenant_id = ? AND node_pool_id = ?`
	sqlSumDeptResourcesAdjustable            = `SELECT CASE WHEN COUNT(CASE WHEN gpu_deserved IS NULL OR gpu_deserved < 0 THEN 1 END) > 0 THEN -1 ELSE COALESCE(SUM(gpu_deserved), 0) END AS deserved_sum, COALESCE(SUM(gpu_over_quota_weight), 0) AS weights_sum FROM department_node_pool_resources_view WHERE tenant_id = ? AND node_pool_id = ?`
	sqlSumDeptResourcesFixed                 = `SELECT CASE WHEN COUNT(CASE WHEN gpu_deserved IS NULL OR gpu_deserved < 0 THEN 1 END) > 0 THEN -1 ELSE COALESCE(SUM(gpu_deserved), 0) END AS deserved_sum, COALESCE(SUM(CASE WHEN gpu_deserved >= 0 THEN gpu_deserved ELSE 0 END), 0) AS weights_sum FROM department_node_pool_resources_view WHERE tenant_id = ? AND node_pool_id = ?`
	sqlSelectDeptNodePoolPerPair             = `SELECT * FROM department_node_pool_resources_view WHERE tenant_id = ? AND org_unit_id = ? AND node_pool_id = ? ORDER BY department_node_pool_resources_view.id LIMIT ?`
	sqlSelectProjNodePoolPerPair             = `SELECT * FROM project_node_pool_resources_view WHERE tenant_id = ? AND org_unit_id = ? AND node_pool_id = ? ORDER BY project_node_pool_resources_view.id LIMIT ?`
	sqlPatchResources                        = `UPDATE resources SET gpu_deserved=? WHERE id = ? AND node_pool_id = ?`
	sqlPatchDepartmentMetadata               = `UPDATE department SET updated_at=?,updated_by=? WHERE tenant_id = ? AND cluster_id = ? AND id = ?`
)

var _ = Describe("Node pool resources DAO SQL", func() {
	Describe("GetDepartmentNodePoolResources", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartmentNodePoolResources()
			mocker.ExpectQuery(sqlSelectDeptNodePoolResources, testTenantID, 10, 5).
				Returns([]models.DepartmentNodePoolResources{expected})

			got, err := GetDepartmentNodePoolResources(context.Background(), testTenantID, nil, "", "", 5, 10, nil, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
			Expect((*got)[0].OrgUnitName).To(Equal(expected.OrgUnitName))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDeptNodePoolResources, testTenantID, 10, 5).
				ReturnsError(errors.New("db error"))

			got, err := GetDepartmentNodePoolResources(context.Background(), testTenantID, nil, "", "", 5, 10, nil, "")
			Expect(err).To(MatchError("db error"))
			Expect(got).To(BeNil())
		})
	})
	Describe("GetProjectNodePoolResources", func() {
		It("sorts by gpuLimit with a NULL limit counted as -1", func() {
			mocker := setupSQLMock()
			sortColumn, err := queryapi.NewParser(models.ProjectNodePoolResources{}.SortByFieldsToColumn()).GetMapping("gpuLimit")
			Expect(err).NotTo(HaveOccurred())
			mocker.ExpectQuery(sqlSelectProjNodePoolResourcesByGpuLimit, testTenantID, 10, 5).
				Returns([]models.ProjectNodePoolResources{testProjectNodePoolResources()})

			_, err = GetProjectNodePoolResources(context.Background(), testTenantID, nil, sortColumn, "ASC", 5, 10, nil, "")
			Expect(err).NotTo(HaveOccurred())
		})
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProjectNodePoolResources()
			mocker.ExpectQuery(sqlSelectProjNodePoolResources, testTenantID, 10, 5).
				Returns([]models.ProjectNodePoolResources{expected})

			got, err := GetProjectNodePoolResources(context.Background(), testTenantID, nil, "", "", 5, 10, nil, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
			Expect((*got)[0].OrgUnitName).To(Equal(expected.OrgUnitName))
		})
	})
	Describe("CountDepartmentNodePoolResources", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlCountDeptNodePoolResources, testTenantID).Returns([]countResult{{Count: 2}})

			got, err := CountDepartmentNodePoolResources(context.Background(), testTenantID, nil, "", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(Equal(int64(2)))
		})
	})
	Describe("CountProjectNodePoolResources", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlCountProjNodePoolResources, testTenantID).Returns([]countResult{{Count: 5}})

			got, err := CountProjectNodePoolResources(context.Background(), testTenantID, nil, "", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(Equal(int64(5)))
		})
	})
	DescribeTable("filters node pool resources by gpuLimit, counting a NULL limit as -1",
		func(view string, mapping queryapi.APIFieldsMapping,
			count func(context.Context, string, []queryapi.Filter, string, *openapi.PermittedScopes) (*int64, error),
			filter, operator, value string,
		) {
			mocker := setupSQLMock()
			filters, err := models.ParseNodePoolResourcesFilters(mapping, []string{filter})
			Expect(err).NotTo(HaveOccurred())
			mocker.ExpectQuery(fmt.Sprintf("SELECT count(*) FROM %s WHERE tenant_id = ? AND COALESCE(gpu_limit, -1) %s ?", view, operator),
				testTenantID, value).Returns([]countResult{{Count: 1}})

			got, err := count(context.Background(), testTenantID, filters, "", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(Equal(int64(1)))
		},
		Entry("departments ==", models.DepartmentNodePoolResourcesView, models.DepartmentNodePoolResources{}.FilterByFieldsToColumn(), CountDepartmentNodePoolResources, "gpuLimit==-1", "=", "-1"),
		Entry("departments !=", models.DepartmentNodePoolResourcesView, models.DepartmentNodePoolResources{}.FilterByFieldsToColumn(), CountDepartmentNodePoolResources, "gpuLimit!=-1", "is distinct from", "-1"),
		Entry("departments <=", models.DepartmentNodePoolResourcesView, models.DepartmentNodePoolResources{}.FilterByFieldsToColumn(), CountDepartmentNodePoolResources, "gpuLimit<=2", "<=", "2"),
		Entry("departments >=", models.DepartmentNodePoolResourcesView, models.DepartmentNodePoolResources{}.FilterByFieldsToColumn(), CountDepartmentNodePoolResources, "gpuLimit>=4", ">=", "4"),
		Entry("projects ==", models.ProjectNodePoolResourcesView, models.ProjectNodePoolResources{}.FilterByFieldsToColumn(), CountProjectNodePoolResources, "gpuLimit==-1", "=", "-1"),
		Entry("projects !=", models.ProjectNodePoolResourcesView, models.ProjectNodePoolResources{}.FilterByFieldsToColumn(), CountProjectNodePoolResources, "gpuLimit!=-1", "is distinct from", "-1"),
		Entry("projects <=", models.ProjectNodePoolResourcesView, models.ProjectNodePoolResources{}.FilterByFieldsToColumn(), CountProjectNodePoolResources, "gpuLimit<=2", "<=", "2"),
		Entry("projects >=", models.ProjectNodePoolResourcesView, models.ProjectNodePoolResources{}.FilterByFieldsToColumn(), CountProjectNodePoolResources, "gpuLimit>=4", ">=", "4"),
	)
	Describe("SumProjectNodePoolResourcesGpuWeights", func() {
		It("adjustable", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSumProjGPUWeights, testTenantID, testNodePoolID).Returns([]sumRow{{Sum: 8}})

			got, err := SumProjectNodePoolResourcesGpuWeights(context.Background(), testTenantID, testNodePoolID, nil, true, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(Equal(float64(8)))
		})
		It("deserved", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSumProjGPUDeserved, testTenantID, testNodePoolID).Returns([]sumRow{{Sum: 3}})

			got, err := SumProjectNodePoolResourcesGpuWeights(context.Background(), testTenantID, testNodePoolID, nil, false, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(Equal(float64(3)))
		})
	})
	Describe("SumDepartmentNodePoolResources", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSumDeptResourcesAdjustable, testTenantID, testNodePoolID).
				Returns([]DepartmentNodePoolResourcesSums{{DeservedSum: 10, WeightsSum: 4}})

			got, err := SumDepartmentNodePoolResources(context.Background(), testTenantID, testNodePoolID, nil, true, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.DeservedSum).To(Equal(float64(10)))
			Expect(got.WeightsSum).To(Equal(float64(4)))
		})
		It("fixed weights", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSumDeptResourcesFixed, testTenantID, testNodePoolID).
				Returns([]DepartmentNodePoolResourcesSums{{DeservedSum: 6, WeightsSum: 6}})

			got, err := SumDepartmentNodePoolResources(context.Background(), testTenantID, testNodePoolID, nil, false, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.WeightsSum).To(Equal(float64(6)))
		})
	})
	Describe("GetDepartmentNodePoolResourcesPerDepartmentNodePool", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartmentNodePoolResources()
			mocker.ExpectQuery(sqlSelectDeptNodePoolPerPair, testTenantID, "11", testNodePoolID, 1).
				Returns([]models.DepartmentNodePoolResources{expected})

			got, err := GetDepartmentNodePoolResourcesPerDepartmentNodePool(context.Background(), testTenantID, "11", testNodePoolID, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.OrgUnitName).To(Equal(expected.OrgUnitName))
		})
	})
	Describe("GetProjectNodePoolResourcesPerDepartmentNodePool", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProjectNodePoolResources()
			mocker.ExpectQuery(sqlSelectProjNodePoolPerPair, testTenantID, "22", testNodePoolID, 1).
				Returns([]models.ProjectNodePoolResources{expected})

			got, err := GetProjectNodePoolResourcesPerDepartmentNodePool(context.Background(), testTenantID, "22", testNodePoolID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.OrgUnitName).To(Equal(expected.OrgUnitName))
		})
	})
	Describe("PatchDepartmentNodePoolResources", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			subject := "user@run.ai"
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlPatchResources, float64(2), testResourcesID().String(), testNodePoolID).Returns(1)
			mocker.ExpectExec(sqlPatchDepartmentMetadata, dbtesting.AnyTime(), subject, testTenantID, testClusterID().String(), "11").Returns(1)
			mocker.ExpectTransaction().Commit()

			err := PatchDepartmentNodePoolResources(
				context.Background(),
				testTenantID,
				testClusterID().String(),
				"11",
				testNodePoolID,
				testResourcesID().String(),
				ptr.To(subject),
				map[string]interface{}{"gpu_deserved": float64(2)},
			)
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlPatchResources).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			err := PatchDepartmentNodePoolResources(
				context.Background(),
				testTenantID,
				testClusterID().String(),
				"11",
				testNodePoolID,
				testResourcesID().String(),
				nil,
				map[string]interface{}{"gpu_deserved": float64(2)},
			)
			Expect(err).To(MatchError("update failed"))
		})
	})
})
