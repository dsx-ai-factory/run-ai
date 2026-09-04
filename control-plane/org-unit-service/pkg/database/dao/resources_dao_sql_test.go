package dao

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clusteropenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
)

const (
	sqlInsertResources                    = `INSERT INTO resources (id,created_at,updated_at,deleted_at,created_by,updated_by,deleted_by,node_pool_id,node_pool_name,gpu_deserved,gpu_limit,gpu_over_quota_weight,cpu_deserved,cpu_limit,cpu_over_quota_weight,memory_deserved,memory_limit,memory_over_quota_weight,memory_units,org_unit_data_id,priority,allocated_gpu,allocated_cpu,allocated_memory,allocated_non_preemptible_gpu,allocated_non_preemptible_cpu,allocated_non_preemptible_memory,requested_gpu,requested_cpu,requested_memory) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	sqlSoftDeleteResources                = `UPDATE resources SET deleted_at=? WHERE resources.id = ? AND resources.deleted_at IS NULL`
	sqlResetProjectQuotas                 = `UPDATE resources SET gpu_deserved = 0, gpu_limit = NULL, cpu_deserved = NULL, cpu_limit = NULL, memory_deserved = NULL, memory_limit = NULL, updated_at=now() FROM project WHERE project.org_unit_data_id = resources.org_unit_data_id AND ((gpu_deserved <> 0 OR gpu_limit <> -1 OR cpu_deserved <> -1 OR cpu_limit <> -1 OR memory_deserved <> -1 OR memory_limit <> -1) AND resources.node_pool_id IN (?)) AND project.tenant_id = ? AND project.cluster_id = ? AND project.deleted_at is NULL and resources.deleted_at is NULL;`
	sqlOrgUnitsNodepoolSummaryProjects    = `SELECT o.id, o.name FROM project AS o JOIN resources on o.org_unit_data_id = resources.org_unit_data_id WHERE o.tenant_id = ? AND o.deleted_at IS NULL AND resources.deleted_at IS NULL AND resources.node_pool_id = ? AND ((gpu_deserved <> 0 OR gpu_limit <> -1 OR cpu_deserved <> -1 OR cpu_limit <> -1 OR memory_deserved <> -1 OR memory_limit <> -1)) ORDER BY o.id ASC LIMIT ?`
	sqlOrgUnitsNodepoolSummaryDepartments = `SELECT o.id, o.name FROM department AS o JOIN resources on o.org_unit_data_id = resources.org_unit_data_id WHERE o.tenant_id = ? AND o.deleted_at IS NULL AND resources.deleted_at IS NULL AND resources.node_pool_id = ? AND ((gpu_deserved <> 0 OR gpu_limit <> -1 OR cpu_deserved <> -1 OR cpu_limit <> -1 OR memory_deserved <> -1 OR memory_limit <> -1)) ORDER BY o.id ASC LIMIT ?`
	sqlDisableCPUQuotaProjects            = `UPDATE resources SET cpu_deserved = NULL, cpu_limit = NULL, cpu_over_quota_weight = NULL, memory_deserved = NULL, memory_limit = NULL, memory_over_quota_weight = NULL, updated_at=now() FROM project WHERE project.org_unit_data_id = resources.org_unit_data_id AND (cpu_deserved IS NOT NULL OR cpu_limit IS NOT NULL OR cpu_over_quota_weight IS NOT NULL OR memory_deserved IS NOT NULL OR memory_limit IS NOT NULL OR memory_over_quota_weight IS NOT NULL) AND project.tenant_id = ? AND project.cluster_id = ? AND project.deleted_at is NULL and resources.deleted_at is NULL;`
	sqlDisableCPUQuotaDepartments         = `UPDATE resources SET cpu_deserved = NULL, cpu_limit = NULL, cpu_over_quota_weight = NULL, memory_deserved = NULL, memory_limit = NULL, memory_over_quota_weight = NULL, updated_at=now() FROM department WHERE department.org_unit_data_id = resources.org_unit_data_id AND (cpu_deserved IS NOT NULL OR cpu_limit IS NOT NULL OR cpu_over_quota_weight IS NOT NULL OR memory_deserved IS NOT NULL OR memory_limit IS NOT NULL OR memory_over_quota_weight IS NOT NULL) AND department.tenant_id = ? AND department.cluster_id = ? AND department.deleted_at is NULL and resources.deleted_at is NULL;`
)

var _ = Describe("Resources DAO SQL", func() {
	Describe("AddResources", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			resource := testResources()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlInsertResources).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := AddResources(context.Background(), []*models.Resources{&resource})
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			resource := testResources()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlInsertResources).ReturnsError(errors.New("insert failed"))
			mocker.ExpectTransaction().Rollback()

			err := AddResources(context.Background(), []*models.Resources{&resource})
			Expect(err).To(MatchError("insert failed"))
		})
	})
	Describe("DeleteResources", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			resource := testResources()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlSoftDeleteResources).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := DeleteResources(context.Background(), []models.Resources{resource})
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			resource := testResources()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlSoftDeleteResources).ReturnsError(errors.New("delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteResources(context.Background(), []models.Resources{resource})
			Expect(err).To(MatchError("delete failed"))
		})
	})
	Describe("DisableCpuQuota", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			clusterUUID := testClusterID().String()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlDisableCPUQuotaProjects, testTenantID, clusterUUID).Returns(1)
			mocker.ExpectExec(sqlDisableCPUQuotaDepartments, testTenantID, clusterUUID).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := DisableCpuQuota(context.Background(), testTenantID, []clusteropenapi.ClustersMinimalData{{Uuid: clusterUUID}})
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			clusterUUID := testClusterID().String()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlDisableCPUQuotaProjects, testTenantID, clusterUUID).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			err := DisableCpuQuota(context.Background(), testTenantID, []clusteropenapi.ClustersMinimalData{{Uuid: clusterUUID}})
			Expect(err).To(MatchError("update failed"))
		})
	})
	Describe("ResetResources", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectExec(sqlResetProjectQuotas).Returns(2)

			err := ResetResources(context.Background(), testTenantID, testClusterID().String(), "", []string{testNodePoolID}, api.RESOURCESRESETACTION_QUOTAS_TO_DEFAULT, true)
			Expect(err).NotTo(HaveOccurred())
		})
		It("invalid action", func() {
			setupSQLMock()
			err := ResetResources(context.Background(), testTenantID, testClusterID().String(), "", []string{testNodePoolID}, "not-an-action", true)
			Expect(err).To(MatchError("invalid reset resources action: not-an-action"))
		})
		It("departments", func() {
			mocker := setupSQLMock()
			mocker.ExpectExec(`UPDATE resources SET gpu_deserved = 0, gpu_limit = NULL, cpu_deserved = NULL, cpu_limit = NULL, memory_deserved = NULL, memory_limit = NULL, updated_at=now() FROM department WHERE department.org_unit_data_id = resources.org_unit_data_id AND ((gpu_deserved <> 0 OR gpu_limit <> -1 OR cpu_deserved <> -1 OR cpu_limit <> -1 OR memory_deserved <> -1 OR memory_limit <> -1) AND resources.node_pool_id IN (?)) AND department.tenant_id = ? AND department.cluster_id = ? AND department.deleted_at is NULL and resources.deleted_at is NULL;`).Returns(1)

			err := ResetResources(context.Background(), testTenantID, testClusterID().String(), "", []string{testNodePoolID}, api.RESOURCESRESETACTION_QUOTAS_TO_DEFAULT, false)
			Expect(err).NotTo(HaveOccurred())
		})
		It("with department filter", func() {
			mocker := setupSQLMock()
			mocker.ExpectExec(`UPDATE resources SET gpu_deserved = 0, gpu_limit = NULL, cpu_deserved = NULL, cpu_limit = NULL, memory_deserved = NULL, memory_limit = NULL, updated_at=now() FROM project WHERE project.org_unit_data_id = resources.org_unit_data_id AND ((gpu_deserved <> 0 OR gpu_limit <> -1 OR cpu_deserved <> -1 OR cpu_limit <> -1 OR memory_deserved <> -1 OR memory_limit <> -1) AND resources.node_pool_id IN (?) AND project.parent_id = ?) AND project.tenant_id = ? AND project.cluster_id = ? AND project.deleted_at is NULL and resources.deleted_at is NULL;`).Returns(1)

			err := ResetResources(context.Background(), testTenantID, testClusterID().String(), "11", []string{testNodePoolID}, api.RESOURCESRESETACTION_QUOTAS_TO_DEFAULT, true)
			Expect(err).NotTo(HaveOccurred())
		})
		It("ranks to default", func() {
			mocker := setupSQLMock()
			mocker.ExpectExec(`UPDATE resources SET priority = 'MediumLow', updated_at=now() FROM project WHERE project.org_unit_data_id = resources.org_unit_data_id AND ((resources.priority <> 'MediumLow') AND resources.node_pool_id IN (?)) AND project.tenant_id = ? AND project.cluster_id = ? AND project.deleted_at is NULL and resources.deleted_at is NULL;`).Returns(1)

			err := ResetResources(context.Background(), testTenantID, testClusterID().String(), "", []string{testNodePoolID}, api.RESOURCESRESETACTION_RANKS_TO_DEFAULT, true)
			Expect(err).NotTo(HaveOccurred())
		})
		It("weights to default", func() {
			mocker := setupSQLMock()
			mocker.ExpectExec(`UPDATE resources SET gpu_over_quota_weight = 2.000000, cpu_over_quota_weight = CASE WHEN cpu_over_quota_weight IS NOT NULL THEN 2.000000 ELSE NULL END, memory_over_quota_weight = CASE WHEN memory_over_quota_weight IS NOT NULL THEN 2.000000 ELSE NULL END, updated_at=now() FROM project WHERE project.org_unit_data_id = resources.org_unit_data_id AND ((gpu_over_quota_weight IS DISTINCT FROM 2.000000 OR cpu_over_quota_weight <> 2.000000 OR memory_over_quota_weight <> 2.000000) AND resources.node_pool_id IN (?)) AND project.tenant_id = ? AND project.cluster_id = ? AND project.deleted_at is NULL and resources.deleted_at is NULL;`).Returns(1)

			err := ResetResources(context.Background(), testTenantID, testClusterID().String(), "", []string{testNodePoolID}, api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT, true)
			Expect(err).NotTo(HaveOccurred())
		})
		It("weights to deserved", func() {
			mocker := setupSQLMock()
			mocker.ExpectExec(`UPDATE resources SET gpu_over_quota_weight = gpu_deserved, cpu_over_quota_weight = cpu_deserved, memory_over_quota_weight = memory_deserved, updated_at=now() FROM project WHERE project.org_unit_data_id = resources.org_unit_data_id AND ((gpu_over_quota_weight IS DISTINCT FROM gpu_deserved OR cpu_over_quota_weight IS DISTINCT FROM cpu_deserved OR memory_over_quota_weight IS DISTINCT FROM memory_deserved) AND resources.node_pool_id IN (?)) AND project.tenant_id = ? AND project.cluster_id = ? AND project.deleted_at is NULL and resources.deleted_at is NULL;`).Returns(1)

			err := ResetResources(context.Background(), testTenantID, testClusterID().String(), "", []string{testNodePoolID}, api.RESOURCESRESETACTION_WEIGHTS_TO_DESERVED, true)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("GetOrgUnitsNodepoolSummary", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlOrgUnitsNodepoolSummaryProjects, testTenantID, testNodePoolID, 1000).
				Returns([]api.OrgUnitsNodepoolSummaryInner{{Id: "22", Name: "proj-a"}})
			mocker.ExpectQuery(sqlOrgUnitsNodepoolSummaryDepartments, testTenantID, testNodePoolID, 1000).
				Returns([]api.OrgUnitsNodepoolSummaryInner{{Id: "11", Name: "research"}})

			got, err := GetOrgUnitsNodepoolSummary(context.Background(), testTenantID, testNodePoolID, api.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(2))
			Expect(got[0].Type).To(Equal(models.ProjectTable))
			Expect(got[1].Type).To(Equal(models.DepartmentTable))
		})
		It("invalid type", func() {
			setupSQLMock()
			got, err := GetOrgUnitsNodepoolSummary(context.Background(), testTenantID, testNodePoolID, "invalid")
			Expect(err.Error()).To(ContainSubstring("invalid nodepool summary type"))
			Expect(got).To(BeNil())
		})
	})
	Describe("NodepoolResetResources", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			sqlProjects := `UPDATE resources SET cpu_deserved = NULL, cpu_limit = NULL, cpu_over_quota_weight = NULL, memory_deserved = NULL, memory_limit = NULL, memory_over_quota_weight = NULL, updated_at=now() FROM project WHERE project.org_unit_data_id = resources.org_unit_data_id AND ((cpu_deserved IS NOT NULL OR cpu_limit IS NOT NULL OR cpu_over_quota_weight IS NOT NULL OR memory_deserved IS NOT NULL OR memory_limit IS NOT NULL OR memory_over_quota_weight IS NOT NULL) AND resources.node_pool_id IN (?)) AND project.tenant_id = ? AND project.cluster_id = ? AND project.deleted_at is NULL and resources.deleted_at is NULL;`
			sqlDepartments := `UPDATE resources SET cpu_deserved = NULL, cpu_limit = NULL, cpu_over_quota_weight = NULL, memory_deserved = NULL, memory_limit = NULL, memory_over_quota_weight = NULL, updated_at=now() FROM department WHERE department.org_unit_data_id = resources.org_unit_data_id AND ((cpu_deserved IS NOT NULL OR cpu_limit IS NOT NULL OR cpu_over_quota_weight IS NOT NULL OR memory_deserved IS NOT NULL OR memory_limit IS NOT NULL OR memory_over_quota_weight IS NOT NULL) AND resources.node_pool_id IN (?)) AND department.tenant_id = ? AND department.cluster_id = ? AND department.deleted_at is NULL and resources.deleted_at is NULL;`
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlProjects).Returns(1)
			mocker.ExpectExec(sqlDepartments).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := NodepoolResetResources(context.Background(), testTenantID, testClusterID().String(), testNodePoolID, api.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT)
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			sqlProjects := `UPDATE resources SET cpu_deserved = NULL, cpu_limit = NULL, cpu_over_quota_weight = NULL, memory_deserved = NULL, memory_limit = NULL, memory_over_quota_weight = NULL, updated_at=now() FROM project WHERE project.org_unit_data_id = resources.org_unit_data_id AND ((cpu_deserved IS NOT NULL OR cpu_limit IS NOT NULL OR cpu_over_quota_weight IS NOT NULL OR memory_deserved IS NOT NULL OR memory_limit IS NOT NULL OR memory_over_quota_weight IS NOT NULL) AND resources.node_pool_id IN (?)) AND project.tenant_id = ? AND project.cluster_id = ? AND project.deleted_at is NULL and resources.deleted_at is NULL;`
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlProjects).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			err := NodepoolResetResources(context.Background(), testTenantID, testClusterID().String(), testNodePoolID, api.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT)
			Expect(err).To(MatchError("update failed"))
		})
	})
})
