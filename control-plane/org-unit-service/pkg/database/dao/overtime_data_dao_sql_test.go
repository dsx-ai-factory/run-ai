package dao

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"k8s.io/utils/ptr"
)

const (
	sqlSelectDepartmentRanking = `SELECT id, name, avg_gpu_allocation24h as allocation_value, avg_gpu_utilization24h as utilization_value, avg_gpu_memory_utilization24h as memory_utilization_value, total_gpu_quota as gpu_quota FROM department_view WHERE tenant_id = ? AND cluster_id = ? AND deleted_at IS NULL ORDER BY avg_gpu_allocation24h DESC NULLS LAST, name ASC LIMIT ?`
	sqlSelectProjectRanking    = `SELECT id, name, parent_name, avg_gpu_allocation24h as allocation_value, avg_gpu_utilization24h as utilization_value, avg_gpu_memory_utilization24h as memory_utilization_value, total_gpu_quota as gpu_quota FROM project_view WHERE tenant_id = ? AND cluster_id = ? AND deleted_at IS NULL ORDER BY avg_gpu_allocation24h DESC NULLS LAST, name ASC LIMIT ?`
	sqlSelectDeptAllocRatio    = `SELECT d.name, d.id::text as id, CASE WHEN dept_r.gpu_deserved > 0 THEN child_alloc.total_allocated / dept_r.gpu_deserved ELSE 0 END as allocation_ratio, dept_r.gpu_deserved as gpu_quota FROM department d JOIN org_unit_data dept_oud ON d.org_unit_data_id = dept_oud.id JOIN resources dept_r ON dept_r.org_unit_data_id = dept_oud.id AND dept_r.deleted_at IS NULL AND dept_r.node_pool_name = ? LEFT JOIN LATERAL ( SELECT COALESCE(SUM(r.allocated_gpu), 0) as total_allocated FROM project p JOIN org_unit_data oud ON p.org_unit_data_id = oud.id JOIN resources r ON r.org_unit_data_id = oud.id AND r.node_pool_name = ? AND r.deleted_at IS NULL WHERE p.parent_id = d.id AND p.deleted_at IS NULL ) child_alloc ON true WHERE d.tenant_id = ? AND d.cluster_id = ? AND d.deleted_at IS NULL ORDER BY allocation_ratio DESC NULLS LAST, name ASC LIMIT 5`
	sqlSelectProjAllocRatio    = `SELECT p.name, p.id::text as id, p.parent_name, CASE WHEN r.gpu_deserved > 0 THEN r.allocated_gpu / r.gpu_deserved ELSE 0 END as allocation_ratio, r.gpu_deserved as gpu_quota FROM project p JOIN org_unit_data oud ON p.org_unit_data_id = oud.id JOIN resources r ON r.org_unit_data_id = oud.id AND r.deleted_at IS NULL WHERE p.tenant_id = ? AND p.cluster_id = ? AND p.deleted_at IS NULL AND r.node_pool_name = ? ORDER BY allocation_ratio DESC NULLS LAST, name ASC LIMIT 5`
)

type rankingRow struct {
	Name                   string   `gorm:"column:name"`
	Id                     string   `gorm:"column:id"`
	ParentName             string   `gorm:"column:parent_name"`
	AllocationValue        *float64 `gorm:"column:allocation_value"`
	UtilizationValue       *float64 `gorm:"column:utilization_value"`
	MemoryUtilizationValue *float64 `gorm:"column:memory_utilization_value"`
	GpuQuota               *float64 `gorm:"column:gpu_quota"`
}

type allocationRatioRow struct {
	Name            string   `gorm:"column:name"`
	Id              string   `gorm:"column:id"`
	ParentName      string   `gorm:"column:parent_name"`
	AllocationRatio *float64 `gorm:"column:allocation_ratio"`
	GpuQuota        *float64 `gorm:"column:gpu_quota"`
}

type overtimeOrgUnitViewRow struct {
	ID             int64
	TenantId       string
	ClusterId      uuid.UUID
	Name           string
	OrgUnitDataID  uuid.UUID
	OvertimeDataID uuid.UUID `gorm:"column:overtime_data_id"`
}

func overtimeProjectViewRow(overtimeID uuid.UUID) overtimeOrgUnitViewRow {
	return overtimeOrgUnitViewRow{
		ID:             testProjectID,
		TenantId:       testTenantID,
		ClusterId:      testClusterID(),
		Name:           "proj-a",
		OrgUnitDataID:  testOrgUnitDataID(),
		OvertimeDataID: overtimeID,
	}
}

func overtimeDepartmentViewRow(overtimeID uuid.UUID) overtimeOrgUnitViewRow {
	return overtimeOrgUnitViewRow{
		ID:             testDepartmentID,
		TenantId:       testTenantID,
		ClusterId:      testClusterID(),
		Name:           "research",
		OrgUnitDataID:  testOrgUnitDataID(),
		OvertimeDataID: overtimeID,
	}
}

func testOvertimeMetrics() *models.OvertimeData {
	v := 1.5
	return &models.OvertimeData{
		AvgGpuAllocation24h:        &v,
		AvgGpuUtilization24h:       &v,
		AvgGpuMemoryUtilization24h: &v,
		AvgGpuAllocation7d:         &v,
		AvgGpuUtilization7d:        &v,
		AvgGpuMemoryUtilization7d:  &v,
		AvgGpuAllocation30d:        &v,
		AvgGpuUtilization30d:       &v,
		AvgGpuMemoryUtilization30d: &v,
	}
}

const (
	sqlInsertOvertimeData  = `INSERT INTO overtime_data (id,deleted_at,org_unit_data_id,avg_gpu_allocation24h,avg_gpu_utilization24h,avg_gpu_memory_utilization24h,avg24h_range_updated_at,avg_gpu_allocation7d,avg_gpu_utilization7d,avg_gpu_memory_utilization7d,avg7d_range_updated_at,avg_gpu_allocation30d,avg_gpu_utilization30d,avg_gpu_memory_utilization30d,avg30d_range_updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	sqlUpdateOvertimeDay   = `UPDATE overtime_data SET avg24h_range_updated_at=NOW(),avg_gpu_allocation24h=?,avg_gpu_memory_utilization24h=?,avg_gpu_utilization24h=? WHERE org_unit_data_id = ?`
	sqlUpdateOvertimeWeek  = `UPDATE overtime_data SET avg7d_range_updated_at=NOW(),avg_gpu_allocation7d=?,avg_gpu_memory_utilization7d=?,avg_gpu_utilization7d=? WHERE org_unit_data_id = ?`
	sqlUpdateOvertimeMonth = `UPDATE overtime_data SET avg30d_range_updated_at=NOW(),avg_gpu_allocation30d=?,avg_gpu_memory_utilization30d=?,avg_gpu_utilization30d=? WHERE org_unit_data_id = ?`
)

var _ = Describe("Overtime data DAO SQL", func() {
	Describe("GetOrgUnitsRanking", func() {
		It("departments", func() {
			mocker := setupSQLMock()
			alloc, util, mem, quota := 1.5, 0.8, 0.4, 8.0
			mocker.ExpectQuery(sqlSelectDepartmentRanking, testTenantID, testClusterID().String(), 5).
				Returns([]rankingRow{{
					Name: "research", Id: "11",
					AllocationValue: &alloc, UtilizationValue: &util, MemoryUtilizationValue: &mem, GpuQuota: &quota,
				}})

			cols, err := GetColumnNames(api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H)
			Expect(err).NotTo(HaveOccurred())
			got, err := GetOrgUnitsRanking(context.Background(), testTenantID, testClusterID().String(), nil, cols, AvgGpuAllocation24Column, api.RANKINGORDERTYPE_TOP, 5, true, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal("research"))
		})
		It("projects", func() {
			mocker := setupSQLMock()
			alloc, quota := 2.0, 4.0
			mocker.ExpectQuery(sqlSelectProjectRanking, testTenantID, testClusterID().String(), 5).
				Returns([]rankingRow{{Name: "proj-a", Id: "22", ParentName: "research", AllocationValue: &alloc, GpuQuota: &quota}})

			cols, err := GetColumnNames(api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H)
			Expect(err).NotTo(HaveOccurred())
			got, err := GetOrgUnitsRanking(context.Background(), testTenantID, testClusterID().String(), nil, cols, AvgGpuAllocation24Column, api.RANKINGORDERTYPE_TOP, 5, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal("proj-a"))
			Expect(got[0].ParentName).To(Equal("research"))
		})
		It("invalid columns", func() {
			setupSQLMock()
			got, err := GetOrgUnitsRanking(context.Background(), testTenantID, testClusterID().String(), nil, RankingColumnNames{}, "", api.RANKINGORDERTYPE_TOP, 5, true, "")
			Expect(err).To(MatchError("invalid column names provided"))
			Expect(got).To(BeNil())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDepartmentRanking, testTenantID, testClusterID().String(), 5).
				ReturnsError(errors.New("db error"))

			cols, err := GetColumnNames(api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H)
			Expect(err).NotTo(HaveOccurred())
			got, err := GetOrgUnitsRanking(context.Background(), testTenantID, testClusterID().String(), nil, cols, AvgGpuAllocation24Column, api.RANKINGORDERTYPE_TOP, 5, true, "")
			Expect(err).To(MatchError("db error"))
			Expect(got).To(BeNil())
		})
	})
	Describe("GetAllocationRatioRankingByNodepool", func() {
		It("departments", func() {
			mocker := setupSQLMock()
			ratio, quota := 0.5, 8.0
			mocker.ExpectQuery(sqlSelectDeptAllocRatio, testNodePoolName, testNodePoolName, testTenantID, testClusterID().String()).
				Returns([]allocationRatioRow{{Name: "research", Id: "11", AllocationRatio: &ratio, GpuQuota: &quota}})

			got, err := GetAllocationRatioRankingByNodepool(context.Background(), testTenantID, testClusterID().String(), testNodePoolName, nil, api.RANKINGORDERTYPE_TOP, 5, true, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal("research"))
			Expect(*got[0].AllocationValue).To(Equal(ratio))
		})
		It("projects", func() {
			mocker := setupSQLMock()
			ratio, quota := 0.25, 4.0
			mocker.ExpectQuery(sqlSelectProjAllocRatio, testTenantID, testClusterID().String(), testNodePoolName).
				Returns([]allocationRatioRow{{Name: "proj-a", Id: "22", ParentName: "research", AllocationRatio: &ratio, GpuQuota: &quota}})

			got, err := GetAllocationRatioRankingByNodepool(context.Background(), testTenantID, testClusterID().String(), testNodePoolName, nil, api.RANKINGORDERTYPE_TOP, 5, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal("proj-a"))
			Expect(got[0].ParentName).To(Equal("research"))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectProjAllocRatio, testTenantID, testClusterID().String(), testNodePoolName).
				ReturnsError(errors.New("db error"))

			got, err := GetAllocationRatioRankingByNodepool(context.Background(), testTenantID, testClusterID().String(), testNodePoolName, nil, api.RANKINGORDERTYPE_TOP, 5, false, "")
			Expect(err).To(MatchError("db error"))
			Expect(got).To(BeNil())
		})
		It("departments with scopes", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			ratio, quota := 0.5, 8.0
			mocker.ExpectQuery(
				sqlSelectDeptAllocRatio[:len(sqlSelectDeptAllocRatio)-len(" ORDER BY allocation_ratio DESC NULLS LAST, name ASC LIMIT 5")]+
					` AND (false OR d.cluster_id IN ( '`+clusterID+`') OR d.id IN ( '11')) ORDER BY allocation_ratio DESC NULLS LAST, name ASC LIMIT 5`,
				testNodePoolName, testNodePoolName, testTenantID, clusterID,
			).Returns([]allocationRatioRow{{Name: "research", Id: "11", AllocationRatio: &ratio, GpuQuota: &quota}})

			got, err := GetAllocationRatioRankingByNodepool(context.Background(), testTenantID, clusterID, testNodePoolName, &openapi.PermittedScopes{
				Clusters:    []string{clusterID},
				Departments: []string{"11"},
			}, api.RANKINGORDERTYPE_TOP, 5, true, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
		})
		It("projects with scopes", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			ratio, quota := 0.25, 4.0
			mocker.ExpectQuery(
				sqlSelectProjAllocRatio[:len(sqlSelectProjAllocRatio)-len(" ORDER BY allocation_ratio DESC NULLS LAST, name ASC LIMIT 5")]+
					` AND (false OR p.cluster_id IN ( '`+clusterID+`') OR p.parent_id IN ( '11') OR p.id IN ( '22')) ORDER BY allocation_ratio DESC NULLS LAST, name ASC LIMIT 5`,
				testTenantID, clusterID, testNodePoolName,
			).Returns([]allocationRatioRow{{Name: "proj-a", Id: "22", ParentName: "research", AllocationRatio: &ratio, GpuQuota: &quota}})

			got, err := GetAllocationRatioRankingByNodepool(context.Background(), testTenantID, clusterID, testNodePoolName, &openapi.PermittedScopes{
				Clusters:    []string{clusterID},
				Departments: []string{"11"},
				Projects:    []string{"22"},
			}, api.RANKINGORDERTYPE_TOP, 5, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
		})
	})
	Describe("BuildNodepoolRankingScopeFilters", func() {
		It("succeeds", func() {
			clusterID := testClusterID().String()

			dept, proj := buildNodepoolRankingScopeFilters(testTenantID, nil)
			Expect(dept).To(BeEmpty())
			Expect(proj).To(BeEmpty())

			dept, proj = buildNodepoolRankingScopeFilters(testTenantID, &openapi.PermittedScopes{Tenants: []string{testTenantID}})
			Expect(dept).To(BeEmpty())
			Expect(proj).To(BeEmpty())

			dept, proj = buildNodepoolRankingScopeFilters(testTenantID, &openapi.PermittedScopes{})
			Expect(dept).To(Equal("AND (false)"))
			Expect(proj).To(Equal("AND (false)"))

			dept, proj = buildNodepoolRankingScopeFilters(testTenantID, &openapi.PermittedScopes{Clusters: []string{clusterID}})
			Expect(dept).To(Equal("AND (false OR d.cluster_id IN ( '" + clusterID + "'))"))
			Expect(proj).To(Equal("AND (false OR p.cluster_id IN ( '" + clusterID + "'))"))

			dept, proj = buildNodepoolRankingScopeFilters(testTenantID, &openapi.PermittedScopes{Departments: []string{"11"}})
			Expect(dept).To(Equal("AND (false OR d.id IN ( '11'))"))
			Expect(proj).To(Equal("AND (false OR p.parent_id IN ( '11'))"))

			dept, proj = buildNodepoolRankingScopeFilters(testTenantID, &openapi.PermittedScopes{Projects: []string{"22"}})
			Expect(dept).To(Equal("AND (false)"))
			Expect(proj).To(Equal("AND (false OR p.id IN ( '22'))"))

			dept, proj = buildNodepoolRankingScopeFilters(testTenantID, &openapi.PermittedScopes{
				Clusters:    []string{clusterID},
				Departments: []string{"11"},
				Projects:    []string{"22"},
			})
			Expect(dept).To(Equal("AND (false OR d.cluster_id IN ( '" + clusterID + "') OR d.id IN ( '11'))"))
			Expect(proj).To(Equal("AND (false OR p.cluster_id IN ( '" + clusterID + "') OR p.parent_id IN ( '11') OR p.id IN ( '22'))"))
		})
	})
	Describe("GetDepartmentsRanking", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			alloc := 1.0
			mocker.ExpectQuery(sqlSelectDepartmentRanking, testTenantID, testClusterID().String(), 5).
				Returns([]rankingRow{{Name: "research", Id: "11", AllocationValue: &alloc}})

			cols, err := GetColumnNames(api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H)
			Expect(err).NotTo(HaveOccurred())
			got, err := GetDepartmentsRanking(context.Background(), testTenantID, testClusterID().String(), nil, cols, AvgGpuAllocation24Column, api.RANKINGORDERTYPE_TOP, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal("research"))
		})
	})
	Describe("GetProjectsRanking", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			alloc := 2.0
			mocker.ExpectQuery(sqlSelectProjectRanking, testTenantID, testClusterID().String(), 5).
				Returns([]rankingRow{{Name: "proj-a", Id: "22", ParentName: "research", AllocationValue: &alloc}})

			cols, err := GetColumnNames(api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H)
			Expect(err).NotTo(HaveOccurred())
			got, err := GetProjectsRanking(context.Background(), testTenantID, testClusterID().String(), nil, cols, AvgGpuAllocation24Column, api.RANKINGORDERTYPE_TOP, 5, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal("proj-a"))
		})
	})
	Describe("GetColumnNames", func() {
		It("succeeds", func() {
			_, err := GetColumnNames(api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D)
			Expect(err).NotTo(HaveOccurred())
			_, err = GetColumnNames(api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D)
			Expect(err).NotTo(HaveOccurred())
			_, err = GetColumnNames(api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO)
			Expect(err).NotTo(HaveOccurred())
			_, err = GetColumnNames("not-a-metric")
			Expect(err).To(MatchError("invalid metric type not-a-metric"))
		})
	})
	Describe("GetOrderBy", func() {
		It("succeeds", func() {
			got, err := GetOrderBy(api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(AvgGpuAllocation24Column))
			_, err = GetOrderBy(api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO, true)
			Expect(err).NotTo(HaveOccurred())
			_, err = GetOrderBy("not-a-metric", false)
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("GetColumnNamesForRanking", func() {
		It("department ratio", func() {
			cols, err := GetColumnNamesForRanking(api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(cols.AllocationColumn).To(Equal(GpuAllocationRatioDepartmentExpression))
		})
	})
	Describe("SetProjectsOvertimeDataForTenant", func() {
		It("invalid cluster", func() {
			setupSQLMock()
			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				"not-a-uuid": {"proj-a": {}},
			}, RangeDay)
			Expect(err).To(HaveOccurred())
		})
		It("get projects error", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, clusterID).
				ReturnsError(errors.New("db error"))

			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"proj-a": {}},
			}, RangeDay)
			Expect(err.Error()).To(ContainSubstring("db error"))
		})
		It("create", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, clusterID).
				Returns([]overtimeOrgUnitViewRow{overtimeProjectViewRow(uuid.Nil)})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlInsertOvertimeData).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"proj-a": testOvertimeMetrics()},
			}, RangeDay)
			Expect(err).NotTo(HaveOccurred())
		})
		It("update day", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, clusterID).
				Returns([]overtimeOrgUnitViewRow{overtimeProjectViewRow(testOvertimeDataID())})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateOvertimeDay).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"proj-a": testOvertimeMetrics()},
			}, RangeDay)
			Expect(err).NotTo(HaveOccurred())
		})
		It("update week", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, clusterID).
				Returns([]overtimeOrgUnitViewRow{overtimeProjectViewRow(testOvertimeDataID())})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateOvertimeWeek).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"proj-a": testOvertimeMetrics()},
			}, RangeWeek)
			Expect(err).NotTo(HaveOccurred())
		})
		It("update month", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, clusterID).
				Returns([]overtimeOrgUnitViewRow{overtimeProjectViewRow(testOvertimeDataID())})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateOvertimeMonth).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"proj-a": testOvertimeMetrics()},
			}, RangeMonth)
			Expect(err).NotTo(HaveOccurred())
		})
		It("missing name", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, clusterID).
				Returns([]overtimeOrgUnitViewRow{overtimeProjectViewRow(uuid.Nil)})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectTransaction().Commit()

			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"other-project": testOvertimeMetrics()},
			}, RangeDay)
			Expect(err).NotTo(HaveOccurred())
		})
		It("skip missing metrics", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, clusterID).
				Returns([]overtimeOrgUnitViewRow{overtimeProjectViewRow(testOvertimeDataID())})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectTransaction().Commit()

			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"proj-a": {}},
			}, RangeDay)
			Expect(err).NotTo(HaveOccurred())
		})
		It("invalid range", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, clusterID).
				Returns([]overtimeOrgUnitViewRow{overtimeProjectViewRow(testOvertimeDataID())})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectTransaction().Commit()

			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"proj-a": testOvertimeMetrics()},
			}, "not-a-range")
			Expect(err).NotTo(HaveOccurred())
		})
		It("create error", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, clusterID).
				Returns([]overtimeOrgUnitViewRow{overtimeProjectViewRow(uuid.Nil)})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlInsertOvertimeData).ReturnsError(errors.New("insert failed"))
			mocker.ExpectTransaction().Commit()

			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"proj-a": testOvertimeMetrics()},
			}, RangeDay)
			Expect(err).NotTo(HaveOccurred())
		})
		It("update error", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, clusterID).
				Returns([]overtimeOrgUnitViewRow{overtimeProjectViewRow(testOvertimeDataID())})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateOvertimeDay).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Commit()

			err := SetProjectsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"proj-a": testOvertimeMetrics()},
			}, RangeDay)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("SetDepartmentsOvertimeDataForTenant", func() {
		It("invalid cluster", func() {
			setupSQLMock()
			err := SetDepartmentsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				"not-a-uuid": {"research": {}},
			}, RangeDay)
			Expect(err).To(HaveOccurred())
		})
		It("get departments error", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				ReturnsError(errors.New("db error"))

			err := SetDepartmentsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"research": {}},
			}, RangeDay)
			Expect(err.Error()).To(ContainSubstring("db error"))
		})
		It("create", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			row := overtimeDepartmentViewRow(uuid.Nil)
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]overtimeOrgUnitViewRow{row})
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]overtimeOrgUnitViewRow{row})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlInsertOvertimeData).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := SetDepartmentsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"research": testOvertimeMetrics()},
			}, RangeWeek)
			Expect(err).NotTo(HaveOccurred())
		})
		It("update", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			row := overtimeDepartmentViewRow(testOvertimeDataID())
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]overtimeOrgUnitViewRow{row})
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]overtimeOrgUnitViewRow{row})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateOvertimeMonth).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := SetDepartmentsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"research": testOvertimeMetrics()},
			}, RangeMonth)
			Expect(err).NotTo(HaveOccurred())
		})
		It("missing name", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			row := overtimeDepartmentViewRow(uuid.Nil)
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]overtimeOrgUnitViewRow{row})
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]overtimeOrgUnitViewRow{row})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectTransaction().Commit()

			err := SetDepartmentsOvertimeDataForTenant(context.Background(), testTenantID, map[string]map[string]*models.OvertimeData{
				clusterID: {"other-dept": testOvertimeMetrics()},
			}, RangeDay)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("SetOvertimeDataUpdatedAtIfNeeded", func() {
		It("succeeds", func() {
			day := &models.OvertimeData{}
			setOvertimeDataUpdatedAtIfNeeded(day, RangeDay)
			Expect(day.Avg24hRangeUpdatedAt).NotTo(BeNil())
			Expect(day.Avg7dRangeUpdatedAt).To(BeNil())
			Expect(day.Avg30dRangeUpdatedAt).To(BeNil())

			week := &models.OvertimeData{}
			setOvertimeDataUpdatedAtIfNeeded(week, RangeWeek)
			Expect(week.Avg7dRangeUpdatedAt).NotTo(BeNil())

			month := &models.OvertimeData{}
			setOvertimeDataUpdatedAtIfNeeded(month, RangeMonth)
			Expect(month.Avg30dRangeUpdatedAt).NotTo(BeNil())

			other := &models.OvertimeData{}
			setOvertimeDataUpdatedAtIfNeeded(other, "unknown")
			Expect(other.Avg24hRangeUpdatedAt).To(BeNil())
			Expect(other.Avg7dRangeUpdatedAt).To(BeNil())
			Expect(other.Avg30dRangeUpdatedAt).To(BeNil())
		})
	})
	Describe("GetOvertimeDataQueryUpdates", func() {
		ctx := context.Background()
		alloc, util, mem := 1.0, 0.5, 0.25
		recent := models.OrgUnitData{BaseModel: dbmodels.BaseModel{CreatedAt: time.Now()}}
		old := models.OrgUnitData{BaseModel: dbmodels.BaseModel{CreatedAt: time.Now().Add(-48 * time.Hour)}}
		oldWeek := models.OrgUnitData{BaseModel: dbmodels.BaseModel{CreatedAt: time.Now().Add(-8 * 24 * time.Hour)}}
		oldMonth := models.OrgUnitData{BaseModel: dbmodels.BaseModel{CreatedAt: time.Now().Add(-31 * 24 * time.Hour)}}
		It("nil overtime data", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, nil, RangeDay, recent)
			Expect(err).To(MatchError("missing overtime data"))
			Expect(got).To(BeNil())
		})
		It("day all metrics", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{
				AvgGpuAllocation24h:        &alloc,
				AvgGpuUtilization24h:       &util,
				AvgGpuMemoryUtilization24h: &mem,
			}, RangeDay, recent)
			Expect(err).NotTo(HaveOccurred())
			Expect(got[AvgGpuAllocation24Column]).To(Equal(alloc))
			Expect(got[AvgGpuUtilization24Column]).To(Equal(util))
			Expect(got[AvgGpuMemoryUtilization24Column]).To(Equal(mem))
			Expect(got).To(HaveKey("avg24h_range_updated_at"))
		})
		It("day allocation only", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{AvgGpuAllocation24h: &alloc}, RangeDay, recent)
			Expect(err).NotTo(HaveOccurred())
			Expect(got[AvgGpuAllocation24Column]).To(Equal(alloc))
			Expect(got).NotTo(HaveKey(AvgGpuUtilization24Column))
		})
		It("day missing recent", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{}, RangeDay, recent)
			Expect(err).To(MatchError("missing 24h metric data"))
			Expect(got).To(BeNil())
		})
		It("day missing old", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{}, RangeDay, old)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeNil())
		})
		It("week all metrics", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{
				AvgGpuAllocation7d:        &alloc,
				AvgGpuUtilization7d:       &util,
				AvgGpuMemoryUtilization7d: &mem,
			}, RangeWeek, recent)
			Expect(err).NotTo(HaveOccurred())
			Expect(got[AvgGpuAllocation7dColumn]).To(Equal(alloc))
			Expect(got[AvgGpuUtilization7dColumn]).To(Equal(util))
			Expect(got[AvgGpuMemoryUtilization7dColumn]).To(Equal(mem))
			Expect(got).To(HaveKey("avg7d_range_updated_at"))
		})
		It("week utilization only", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{AvgGpuUtilization7d: &util}, RangeWeek, recent)
			Expect(err).NotTo(HaveOccurred())
			Expect(got[AvgGpuUtilization7dColumn]).To(Equal(util))
		})
		It("week missing recent", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{}, RangeWeek, recent)
			Expect(err).To(MatchError("missing 7d metric data"))
			Expect(got).To(BeNil())
		})
		It("week missing old", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{}, RangeWeek, oldWeek)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeNil())
		})
		It("month all metrics", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{
				AvgGpuAllocation30d:        &alloc,
				AvgGpuUtilization30d:       &util,
				AvgGpuMemoryUtilization30d: &mem,
			}, RangeMonth, recent)
			Expect(err).NotTo(HaveOccurred())
			Expect(got[AvgGpuAllocation30dColumn]).To(Equal(alloc))
			Expect(got[AvgGpuUtilization30dColumn]).To(Equal(util))
			Expect(got[AvgGpuMemoryUtilization30dColumn]).To(Equal(mem))
			Expect(got).To(HaveKey("avg30d_range_updated_at"))
		})
		It("month memory only", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{AvgGpuMemoryUtilization30d: &mem}, RangeMonth, recent)
			Expect(err).NotTo(HaveOccurred())
			Expect(got[AvgGpuMemoryUtilization30dColumn]).To(Equal(mem))
		})
		It("month missing recent", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{}, RangeMonth, recent)
			Expect(err).To(MatchError("missing 30d metric data"))
			Expect(got).To(BeNil())
		})
		It("month missing old", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{}, RangeMonth, oldMonth)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeNil())
		})
		It("invalid range", func() {
			got, err := getOvertimeDataQueryUpdates(ctx, &models.OvertimeData{AvgGpuAllocation24h: ptr.To(1.0)}, "bad", recent)
			Expect(err).To(MatchError("invalid overtime metric range: bad"))
			Expect(got).To(BeNil())
		})
	})
})
