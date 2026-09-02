package dao

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"gorm.io/gorm"
)

const (
	sqlSelectDepartmentViewByID           = `SELECT * FROM department_view WHERE (department_view.id = ? AND department_view.tenant_id = ?) AND department_view.deleted_at IS NULL ORDER BY department_view.id LIMIT ?`
	sqlSelectDefaultDepartment            = `SELECT * FROM department_view WHERE (department_view.tenant_id = ? AND department_view.cluster_id = ? AND department_view.name = ?) AND department_view.deleted_at IS NULL ORDER BY department_view.id LIMIT ?`
	sqlSelectDepartmentsByParent          = `SELECT * FROM department WHERE (department.tenant_id = ? AND department.parent_id = ?) AND department.deleted_at IS NULL`
	sqlSelectDepartmentsByIDs             = `SELECT * FROM department WHERE tenant_id = ? AND id in (?,?) AND department.deleted_at IS NULL`
	sqlSelectDepartmentNames              = `SELECT name FROM department WHERE id IN (?,?)`
	sqlSelectDeletedDepartments           = `SELECT * FROM department WHERE tenant_id = ? AND deleted_at is not null AND deleted_at > ?`
	sqlSelectDepartmentsByClusters        = `SELECT * FROM department WHERE tenant_id = ? AND cluster_id in (?,?) AND department.deleted_at IS NULL`
	sqlCountDepartments                   = `SELECT count(*) FROM department WHERE deleted_at IS NULL AND tenant_id = ?`
	sqlCountUnlimitedGPUDept              = `SELECT count(*) FROM department LEFT JOIN resources ON department.org_unit_data_id = resources.org_unit_data_id WHERE resources.gpu_deserved = -1 AND department.deleted_at is null AND resources.deleted_at is null AND department.cluster_id = ? AND department.tenant_id = ?`
	sqlSelectDepartments                  = `SELECT * FROM department WHERE department.tenant_id = ? AND department.deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`
	sqlSelectDepartmentsByCluster         = `SELECT * FROM department_view WHERE (department_view.tenant_id = ? AND department_view.cluster_id = ?) AND department_view.deleted_at IS NULL`
	sqlSelectDepartmentsMetadataByCluster = `SELECT * FROM department WHERE (department.tenant_id = ? AND department.cluster_id = ?) AND department.deleted_at IS NULL`
	sqlSelectDepartmentByClusterAndIDs    = `SELECT * FROM department_view WHERE (department_view.tenant_id = ? AND department_view.cluster_id = ?) AND id in (?) AND department_view.deleted_at IS NULL`
	sqlSelectSyncDepartmentsView          = `SELECT * FROM (SELECT * FROM get_sync_departments(?, ?)) AS sync_view ORDER BY status_last_updated_time desc,calculated_updated_at,id`
	sqlUpdateDepartmentStatus             = `UPDATE department SET status_last_updated_time=? WHERE id = ? AND tenant_id = ?`
)

const sqlSelectDepartmentClusterID = `SELECT cluster_id FROM department WHERE id = ? AND tenant_id = ? AND department.deleted_at IS NULL ORDER BY department.id LIMIT ?`

const sqlSelectDepartmentByID = `SELECT * FROM department WHERE (department.id = ? AND department.tenant_id = ?) AND department.deleted_at IS NULL ORDER BY department.id LIMIT ?`

const (
	sqlDeleteTrainNodeTypesByOrgUnit     = `DELETE FROM org_unit_train_node_type WHERE org_unit_train_node_type.org_unit_data_id = ?`
	sqlDeleteWorkspaceNodeTypesByOrgUnit = `DELETE FROM org_unit_workspace_node_type WHERE org_unit_workspace_node_type.org_unit_data_id = ?`
	sqlUpdateDepartment                  = `UPDATE department SET updated_at=?,tenant_id=?,cluster_id=?,name=?,org_unit_data_id=? WHERE (department.id = ? AND department.tenant_id = ? AND department.name = ?) AND department.deleted_at IS NULL AND id = ?`
	sqlUpdateDepartmentChildrenEffective = `UPDATE org_unit_data SET default_node_pool_list_effective=?,default_node_pool_list_effective_origin=?,interactive_job_max_idle_duration_effective=?,interactive_job_max_idle_duration_effective_origin=?,interactive_job_preempt_idle_duration_effective=?,interactive_job_preempt_idle_duration_effective_origin=?,interactive_job_time_limit_effective=?,interactive_job_time_limit_effective_origin=?,training_job_max_idle_duration_effective=?,training_job_max_idle_duration_effective_origin=?,training_job_time_limit_effective=?,training_job_time_limit_effective_origin=?,training_node_types_effective_origin=?,updated_at=?,workspace_node_types_effective_origin=? WHERE id = ?`
	sqlSelectDepartmentStatusViewByID    = `SELECT * FROM department_status_view WHERE (department_status_view.id = ? AND department_status_view.tenant_id = ?) AND department_status_view.deleted_at IS NULL ORDER BY department_status_view.id LIMIT ?`
)

func expectDeleteDepartmentOnly(mocker dbtesting.Mocker) {
	mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE org_unit_data_id = ? AND resources.deleted_at IS NULL`).Returns(1)
	mocker.ExpectExec(`UPDATE org_unit_data SET deleted_at=? WHERE id = ? AND org_unit_data.deleted_at IS NULL`).Returns(1)
	mocker.ExpectExec(`UPDATE overtime_data SET deleted_at=? WHERE org_unit_data_id = ? AND overtime_data.deleted_at IS NULL`).Returns(0)
	mocker.ExpectExec(`UPDATE department SET deleted_at=? WHERE department.id = ? AND department.id = ? AND department.deleted_at IS NULL`).Returns(1)
}

func departmentWithProjectChild() models.Department {
	dept := testDepartment()
	project := testProject()
	project.OrgUnitData = models.OrgUnitData{
		BaseModel: dao_utils.BaseModel{ID: testOrgUnitDataID()},
		Resources: []models.Resources{testResources()},
	}
	dept.ProjectChildren = []models.Project{project}
	return dept
}

type departmentStatusViewRow struct {
	ID            int64
	TenantId      string
	ClusterId     uuid.UUID `gorm:"type:uuid"`
	Name          string
	OrgUnitDataID uuid.UUID
	GpuAllocated  *float64
}

func deleteClusterDepartmentsView() models.DepartmentView {
	expected := testDepartmentView()
	expected.ResourcesJson = []interface{}{
		map[string]interface{}{
			"id":           testResourcesID().String(),
			"node_pool_id": testNodePoolID,
		},
	}
	return expected
}

func expectDeleteClusterDepartmentsLookup(mocker dbtesting.Mocker, view models.DepartmentView) {
	mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
		Returns([]models.DepartmentView{view})
	mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
		Returns([]models.DepartmentView{view})
}

var _ = Describe("Department DAO SQL", func() {
	Describe("GetClusterIdByDepartment", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectDepartmentClusterID, testDepartmentID, testTenantID, 1).
				Returns([]clusterIDRow{{ClusterId: clusterID}})

			got, err := GetClusterIdByDepartment(context.Background(), testDepartmentID, testTenantID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(clusterID))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDepartmentClusterID, testDepartmentID, testTenantID, 1).
				ReturnsError(gorm.ErrRecordNotFound)

			got, err := GetClusterIdByDepartment(context.Background(), testDepartmentID, testTenantID)
			Expect(err).To(MatchError(gorm.ErrRecordNotFound))
			Expect(got).To(BeEmpty())
		})
	})
	Describe("GetDepartmentById", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartment()
			mocker.ExpectQuery(sqlSelectDepartmentByID, testDepartmentID, testTenantID, 1).
				Returns(departmentRows(expected))

			got, err := GetDepartmentById(context.Background(), testDepartmentID, testTenantID, models.MetaDataVerbosity, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(expected.Name))
			Expect(got.ID).To(Equal(expected.ID))
		})
		It("with slim verbosity", func() {
			mocker := setupSQLMock()
			expected := testDepartmentView()
			mocker.ExpectQuery(sqlSelectDepartmentViewByID, testDepartmentID, testTenantID, 1).
				Returns([]models.DepartmentView{expected})

			got, err := GetDepartmentById(context.Background(), testDepartmentID, testTenantID, models.SlimVerbosity, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(expected.Name))
		})
		It("returns not found", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDepartmentByID, testDepartmentID, testTenantID, 1).
				ReturnsError(gorm.ErrRecordNotFound)

			got, err := GetDepartmentById(context.Background(), testDepartmentID, testTenantID, models.MetaDataVerbosity, nil)
			Expect(err).To(MatchError(gorm.ErrRecordNotFound))
			Expect(got).To(BeNil())
		})
		It("with verbose verbosity", func() {
			mocker := setupSQLMock()
			gpu := 2.0
			mocker.ExpectQuery(sqlSelectDepartmentStatusViewByID, testDepartmentID, testTenantID, 1).
				Returns([]departmentStatusViewRow{{
					ID:            testDepartmentID,
					TenantId:      testTenantID,
					ClusterId:     testClusterID(),
					Name:          "research",
					OrgUnitDataID: testOrgUnitDataID(),
					GpuAllocated:  &gpu,
				}})
			mocker.ExpectQuery(sqlSelectProjectsViewByParent, testTenantID, testDepartmentID).
				Returns(projectViewRows(testProjectView()))

			got, err := GetDepartmentById(context.Background(), testDepartmentID, testTenantID, models.VerboseVerbosity, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal("research"))
			Expect(got.ProjectChildren).To(HaveLen(1))
			Expect(*got.DepartmentStatus.QuotaStatus.GpuAllocated).To(Equal(gpu))
		})
		It("verbose projects error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDepartmentStatusViewByID, testDepartmentID, testTenantID, 1).
				Returns([]departmentStatusViewRow{{
					ID:            testDepartmentID,
					TenantId:      testTenantID,
					ClusterId:     testClusterID(),
					Name:          "research",
					OrgUnitDataID: testOrgUnitDataID(),
				}})
			mocker.ExpectQuery(sqlSelectProjectsViewByParent, testTenantID, testDepartmentID).
				ReturnsError(errors.New("projects failed"))

			got, err := GetDepartmentById(context.Background(), testDepartmentID, testTenantID, models.VerboseVerbosity, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal("research"))
		})
		It("verbose error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDepartmentStatusViewByID, testDepartmentID, testTenantID, 1).
				ReturnsError(errors.New("db error"))

			got, err := GetDepartmentById(context.Background(), testDepartmentID, testTenantID, models.VerboseVerbosity, nil)
			Expect(err).To(MatchError("db error"))
			Expect(got).To(BeNil())
		})
		It("slim error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDepartmentViewByID, testDepartmentID, testTenantID, 1).
				ReturnsError(errors.New("db error"))

			got, err := GetDepartmentById(context.Background(), testDepartmentID, testTenantID, models.SlimVerbosity, nil)
			Expect(err).To(MatchError("db error"))
			Expect(got).To(BeNil())
		})
		It("unsupported verbosity", func() {
			setupSQLMock()
			got, err := GetDepartmentById(context.Background(), testDepartmentID, testTenantID, models.Verbosity(99), nil)
			Expect(err).To(MatchError("verbosity 99 is not supported"))
			Expect(got).To(BeNil())
		})
	})
	Describe("GetDefaultDepartment", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartmentView()
			expected.Name = defaultDepartmentName
			mocker.ExpectQuery(sqlSelectDefaultDepartment, testTenantID, testClusterID(), defaultDepartmentName, 1).
				Returns([]models.DepartmentView{expected})

			got, err := GetDefaultDepartment(context.Background(), testTenantID, testClusterID(), nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(defaultDepartmentName))
		})
	})
	Describe("GetDepartmentsByParent", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			parentID := int64(1)
			expected := testDepartment()
			expected.ParentID = &parentID
			mocker.ExpectQuery(sqlSelectDepartmentsByParent, testTenantID, parentID).
				Returns(departmentRows(expected))

			got, err := GetDepartmentsByParent(context.Background(), parentID, testTenantID)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
			Expect((*got)[0].Name).To(Equal(expected.Name))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDepartmentsByParent, testTenantID, int64(1)).
				ReturnsError(errors.New("db error"))

			got, err := GetDepartmentsByParent(context.Background(), 1, testTenantID)
			Expect(err).To(MatchError("db error"))
			Expect(got).To(BeNil())
		})
	})
	Describe("GetDepartmentsByIds", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartment()
			mocker.ExpectQuery(sqlSelectDepartmentsByIDs, testTenantID, int64(11), int64(12)).
				Returns(departmentRows(expected))

			got, err := GetDepartmentsByIds(context.Background(), testTenantID, []int64{11, 12}, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
		})
		It("permitted scopes", func() {
			mocker := setupSQLMock()
			expected := testDepartment()
			mocker.ExpectQuery(
				`SELECT * FROM department WHERE tenant_id = ? AND id in (?) AND ((false OR cluster_id IN ( '11111111-1111-1111-1111-111111111111') OR id IN ( '11'))) AND department.deleted_at IS NULL`,
				testTenantID, testDepartmentID,
			).Returns(departmentRows(expected))

			got, err := GetDepartmentsByIds(context.Background(), testTenantID, []int64{testDepartmentID}, &openapi.PermittedScopes{
				Clusters:    []string{testClusterID().String()},
				Departments: []string{"11"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
		})
	})
	Describe("GetDepartmentNamesByIds", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDepartmentNames, "11", "12").
				Returns([]nameRow{{Name: "research"}, {Name: "engineering"}})

			got, err := GetDepartmentNamesByIds(context.Background(), []string{"11", "12"})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal([]string{"research", "engineering"}))
		})
	})
	Describe("GetDeletedDepartments", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			deletedAfter := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			expected := testDepartment()
			mocker.ExpectQuery(sqlSelectDeletedDepartments, testTenantID, deletedAfter).
				Returns(departmentRows(expected))

			got, err := GetDeletedDepartments(context.Background(), testTenantID, deletedAfter)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
		})
	})
	Describe("GetDepartmentsByClusters", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			cluster1 := testClusterID().String()
			cluster2 := "22222222-2222-2222-2222-222222222222"
			expected := testDepartment()
			mocker.ExpectQuery(sqlSelectDepartmentsByClusters, testTenantID, cluster1, cluster2).
				Returns(departmentRows(expected))

			got, err := GetDepartmentsByClusters(context.Background(), testTenantID, []string{cluster1, cluster2})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
		})
	})
	Describe("CountDepartments", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlCountDepartments, testTenantID).Returns([]countResult{{Count: 3}})

			got, err := CountDepartments(context.Background(), testTenantID, nil, nil, models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(int64(3)))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlCountDepartments, testTenantID).ReturnsError(errors.New("db error"))

			got, err := CountDepartments(context.Background(), testTenantID, nil, nil, models.MetaDataVerbosity)
			Expect(err).To(MatchError("db error"))
			Expect(got).To(Equal(int64(0)))
		})
	})
	Describe("HasUnlimitedGpuDepartment", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlCountUnlimitedGPUDept, clusterID, testTenantID).Returns([]countResult{{Count: 1}})

			got, err := HasUnlimitedGpuDepartment(context.Background(), clusterID, testTenantID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeTrue())
		})
		It("none", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlCountUnlimitedGPUDept, clusterID, testTenantID).Returns([]countResult{{Count: 0}})

			got, err := HasUnlimitedGpuDepartment(context.Background(), clusterID, testTenantID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeFalse())
		})
	})
	Describe("GetDepartments", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartment()
			mocker.ExpectQuery(sqlSelectDepartments, testTenantID, 10, 5).
				Returns(departmentRows(expected))

			got, err := GetDepartments(context.Background(), testTenantID, nil, "", "", 5, 10, nil, false, models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
		})
		It("with verbose verbosity", func() {
			mocker := setupSQLMock()
			gpu := 2.0
			mocker.ExpectQuery(
				`SELECT * FROM department_status_view WHERE department_status_view.tenant_id = ? AND department_status_view.deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`,
				testTenantID, 10, 5,
			).Returns([]departmentStatusViewRow{{
				ID:            testDepartmentID,
				TenantId:      testTenantID,
				ClusterId:     testClusterID(),
				Name:          "research",
				OrgUnitDataID: testOrgUnitDataID(),
				GpuAllocated:  &gpu,
			}})

			got, err := GetDepartments(context.Background(), testTenantID, nil, "", "", 5, 10, nil, false, models.VerboseVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
			Expect((*got)[0].Name).To(Equal("research"))
		})
	})
	Describe("GetDepartmentsByCluster", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartmentView()
			// Cache miss with nil permittedScopes runs the query twice because the
			// cache-path assignment shadows the outer departments variable.
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]models.DepartmentView{expected})
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]models.DepartmentView{expected})

			got, err := GetDepartmentsByCluster(context.Background(), testTenantID, testClusterID(), nil, models.SlimVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
			Expect((*got)[0].Name).To(Equal(expected.Name))
		})
		It("metadata", func() {
			mocker := setupSQLMock()
			expected := testDepartment()
			// Same double-query as TestGetDepartmentsByCluster (cache-path variable shadowing).
			mocker.ExpectQuery(sqlSelectDepartmentsMetadataByCluster, testTenantID, testClusterID()).
				Returns(departmentRows(expected))
			mocker.ExpectQuery(sqlSelectDepartmentsMetadataByCluster, testTenantID, testClusterID()).
				Returns(departmentRows(expected))

			got, err := GetDepartmentsByCluster(context.Background(), testTenantID, testClusterID(), nil, models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
		})
	})
	Describe("GetDepartmentByClusterAndIds", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartmentView()
			mocker.ExpectQuery(sqlSelectDepartmentByClusterAndIDs, testTenantID, testClusterID(), "11").
				Returns([]models.DepartmentView{expected})

			got, err := GetDepartmentByClusterAndIds(context.Background(), testTenantID, testClusterID(), []string{"11"}, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
		})
	})
	Describe("GetSyncDepartmentsView", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := models.DepartmentSyncView{
				IntBaseModel:  models.IntBaseModel{ID: testDepartmentID},
				TenantId:      testTenantID,
				ClusterId:     testClusterID(),
				Name:          "research",
				OrgUnitDataID: testOrgUnitDataID(),
			}
			mocker.ExpectQuery(sqlSelectSyncDepartmentsView, testTenantID, testClusterID()).
				Returns([]models.DepartmentSyncView{expected})

			got, err := GetSyncDepartmentsView(context.Background(), testTenantID, testClusterID(), time.Time{}, 0, 0, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
			Expect((*got)[0].Name).To(Equal(expected.Name))
		})
		It("updated after", func() {
			mocker := setupSQLMock()
			after := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
			mocker.ExpectQuery(
				`SELECT * FROM (SELECT * FROM get_sync_departments(?, ?)) AS sync_view WHERE calculated_updated_at > ? OR status_last_updated_time is null OR calculated_updated_at > status_last_updated_time OR (deleted_at is not null AND deleted_at > ?) ORDER BY status_last_updated_time desc,calculated_updated_at,id`,
				testTenantID, testClusterID(), after, after,
			).Returns([]models.DepartmentSyncView{})

			got, err := GetSyncDepartmentsView(context.Background(), testTenantID, testClusterID(), after, 0, 0, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(BeEmpty())
		})
	})
	Describe("UpdateDepartmentStatus", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			now := time.Now()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateDepartmentStatus, now, "11", testTenantID).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := UpdateDepartmentStatus(context.Background(), testClusterID(), testTenantID, "11", &now)
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			now := time.Now()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateDepartmentStatus).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			err := UpdateDepartmentStatus(context.Background(), testClusterID(), testTenantID, "11", &now)
			Expect(err).To(MatchError("update failed"))
		})
	})
	Describe("GetDepartmentByIdAndCluster", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartment()
			mocker.ExpectQuery(
				`SELECT * FROM department WHERE (department.id = ? AND department.tenant_id = ? AND department.cluster_id = ?) AND department.deleted_at IS NULL ORDER BY department.id LIMIT ?`,
				testDepartmentID, testTenantID, testClusterID(), 1,
			).Returns(departmentRows(expected))

			got, err := GetDepartmentByIdAndCluster(context.Background(), testDepartmentID, testTenantID, testClusterID(), models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(expected.Name))
		})
	})
	Describe("CreateDepartment", func() {
		It("returns error", func() {
			mocker := setupSQLMock()
			dept := testDepartment()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectQuery(`INSERT INTO department (created_at,updated_at,deleted_at,created_by,updated_by,deleted_by,tenant_id,cluster_id,name,description,parent_id,org_unit_data_id,status_last_updated_time,id) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING id`).ReturnsError(errors.New("insert failed"))
			mocker.ExpectTransaction().Rollback()

			got, err := CreateDepartment(context.Background(), &dept)
			Expect(err).To(MatchError("insert failed"))
			Expect(got).To(BeNil())
		})
	})
	Describe("CreateDepartments", func() {
		It("returns error", func() {
			mocker := setupSQLMock()
			dept := testDepartment()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectQuery(`INSERT INTO department (created_at,updated_at,deleted_at,created_by,updated_by,deleted_by,tenant_id,cluster_id,name,description,parent_id,org_unit_data_id,status_last_updated_time,id) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT (id) DO UPDATE SET updated_at=?,deleted_at=excluded.deleted_at,created_by=excluded.created_by,updated_by=excluded.updated_by,deleted_by=excluded.deleted_by,tenant_id=excluded.tenant_id,cluster_id=excluded.cluster_id,name=excluded.name,description=excluded.description,parent_id=excluded.parent_id,org_unit_data_id=excluded.org_unit_data_id,status_last_updated_time=excluded.status_last_updated_time RETURNING id`).ReturnsError(errors.New("insert failed"))
			mocker.ExpectTransaction().Rollback()

			err := CreateDepartments(context.Background(), []*models.Department{&dept})
			Expect(err).To(MatchError("insert failed"))
		})
	})
	Describe("DeleteDepartment", func() {
		It("default name", func() {
			mocker := setupSQLMock()
			expected := testDepartment()
			expected.Name = defaultDepartmentName
			mocker.ExpectQuery(sqlSelectDepartmentByID, testDepartmentID, testTenantID, 1).
				Returns(departmentRows(expected))
			mocker.ExpectTransaction().Begin()
			mocker.ExpectTransaction().Rollback()

			name, err := DeleteDepartment(context.Background(), testDepartmentID, testTenantID, nil, defaultDepartmentName)
			Expect(err).To(MatchError("cannot delete default department"))
			Expect(name).To(BeEmpty())
		})
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartment()
			mocker.ExpectQuery(sqlSelectDepartmentByID, testDepartmentID, testTenantID, 1).
				Returns(departmentRows(expected))
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE org_unit_data_id = ? AND resources.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE org_unit_data SET deleted_at=? WHERE id = ? AND org_unit_data.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE overtime_data SET deleted_at=? WHERE org_unit_data_id = ? AND overtime_data.deleted_at IS NULL`).Returns(0)
			mocker.ExpectExec(`UPDATE department SET deleted_at=? WHERE department.id = ? AND department.id = ? AND department.deleted_at IS NULL`).Returns(1)
			mocker.ExpectTransaction().Commit()

			name, err := DeleteDepartment(context.Background(), testDepartmentID, testTenantID, nil, defaultDepartmentName)
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("research"))
		})
	})
	Describe("DeleteDepartmentWithChildren", func() {
		It("default name", func() {
			setupSQLMock()
			dept := testDepartment()
			dept.Name = defaultDepartmentName
			err := DeleteDepartmentWithChildren(context.Background(), &dept, nil, defaultDepartmentName)
			Expect(err).To(MatchError("cannot delete default department"))
		})
		It("succeeds", func() {
			mocker := setupSQLMock()
			dept := testDepartment()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE org_unit_data_id = ? AND resources.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE org_unit_data SET deleted_at=? WHERE id = ? AND org_unit_data.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE overtime_data SET deleted_at=? WHERE org_unit_data_id = ? AND overtime_data.deleted_at IS NULL`).Returns(0)
			mocker.ExpectExec(`UPDATE department SET deleted_at=? WHERE department.id = ? AND department.id = ? AND department.deleted_at IS NULL`).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := DeleteDepartmentWithChildren(context.Background(), &dept, nil, defaultDepartmentName)
			Expect(err).NotTo(HaveOccurred())
		})
		It("deletes children", func() {
			mocker := setupSQLMock()
			dept := departmentWithProjectChild()
			mocker.ExpectTransaction().Begin()
			expectDeleteDepartmentOnly(mocker)
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE id IN (?) AND resources.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE org_unit_data SET deleted_at=? WHERE id IN (?) AND org_unit_data.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE project SET deleted_at=? WHERE id IN (?) AND project.deleted_at IS NULL`).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := DeleteDepartmentWithChildren(context.Background(), &dept, nil, defaultDepartmentName)
			Expect(err).NotTo(HaveOccurred())
		})
		It("children resources error", func() {
			mocker := setupSQLMock()
			dept := departmentWithProjectChild()
			mocker.ExpectTransaction().Begin()
			expectDeleteDepartmentOnly(mocker)
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE id IN (?) AND resources.deleted_at IS NULL`).
				ReturnsError(errors.New("resources delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteDepartmentWithChildren(context.Background(), &dept, nil, defaultDepartmentName)
			Expect(err).To(MatchError("resources delete failed"))
		})
		It("children org unit error", func() {
			mocker := setupSQLMock()
			dept := departmentWithProjectChild()
			mocker.ExpectTransaction().Begin()
			expectDeleteDepartmentOnly(mocker)
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE id IN (?) AND resources.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE org_unit_data SET deleted_at=? WHERE id IN (?) AND org_unit_data.deleted_at IS NULL`).
				ReturnsError(errors.New("org unit delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteDepartmentWithChildren(context.Background(), &dept, nil, defaultDepartmentName)
			Expect(err).To(MatchError("org unit delete failed"))
		})
		It("children project error", func() {
			mocker := setupSQLMock()
			dept := departmentWithProjectChild()
			mocker.ExpectTransaction().Begin()
			expectDeleteDepartmentOnly(mocker)
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE id IN (?) AND resources.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE org_unit_data SET deleted_at=? WHERE id IN (?) AND org_unit_data.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE project SET deleted_at=? WHERE id IN (?) AND project.deleted_at IS NULL`).
				ReturnsError(errors.New("project delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteDepartmentWithChildren(context.Background(), &dept, nil, defaultDepartmentName)
			Expect(err).To(MatchError("project delete failed"))
		})
	})
	Describe("GetSyncDepartments", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := models.DepartmentSyncView{
				IntBaseModel:  models.IntBaseModel{ID: testDepartmentID},
				TenantId:      testTenantID,
				ClusterId:     testClusterID(),
				Name:          "research",
				OrgUnitDataID: testOrgUnitDataID(),
			}
			mocker.ExpectQuery(sqlSelectSyncDepartmentsView, testTenantID, testClusterID()).
				Returns([]models.DepartmentSyncView{expected})

			got, err := GetSyncDepartments(context.Background(), testTenantID, testClusterID(), time.Time{}, 0, 0, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
			Expect((*got)[0].Name).To(Equal(expected.Name))
		})
	})
	Describe("MapProjectsToDepartments", func() {
		It("succeeds", func() {
			parentID := testDepartmentID
			departments := []models.Department{testDepartment()}
			project := testProject()
			orphan := testProject()
			orphan.ParentID = nil
			got := MapProjectsToDepartments(&departments, []*models.Project{&project, &orphan, nil})
			Expect(*got).To(HaveLen(1))
			Expect(*(*got)[0].ProjectChildren[0].ParentID).To(Equal(parentID))
		})
	})
	Describe("DeleteClusterDepartments", func() {
		It("when empty", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]models.DepartmentView{})
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]models.DepartmentView{})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectTransaction().Rollback()

			err := DeleteClusterDepartments(context.Background(), testTenantID, testClusterID())
			Expect(err).To(MatchError("WHERE conditions required"))
		})
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartmentView()
			expected.ResourcesJson = []interface{}{
				map[string]interface{}{
					"id":           testResourcesID().String(),
					"node_pool_id": testNodePoolID,
				},
			}
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]models.DepartmentView{expected})
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				Returns([]models.DepartmentView{expected})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE resources.id = ? AND resources.id = ? AND resources.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE overtime_data SET deleted_at=? WHERE org_unit_data_id in (?) AND overtime_data.deleted_at IS NULL`).Returns(0)
			mocker.ExpectExec(`UPDATE org_unit_data SET deleted_at=? WHERE org_unit_data.id = ? AND org_unit_data.id = ? AND org_unit_data.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE department SET deleted_at=? WHERE department.id = ? AND department.id = ? AND department.deleted_at IS NULL`).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := DeleteClusterDepartments(context.Background(), testTenantID, testClusterID())
			Expect(err).NotTo(HaveOccurred())
		})
		It("get error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectDepartmentsByCluster, testTenantID, testClusterID()).
				ReturnsError(errors.New("db error"))

			err := DeleteClusterDepartments(context.Background(), testTenantID, testClusterID())
			Expect(err).To(MatchError("db error"))
		})
		It("resources error", func() {
			mocker := setupSQLMock()
			expectDeleteClusterDepartmentsLookup(mocker, deleteClusterDepartmentsView())
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE resources.id = ? AND resources.id = ? AND resources.deleted_at IS NULL`).
				ReturnsError(errors.New("resources delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteClusterDepartments(context.Background(), testTenantID, testClusterID())
			Expect(err).To(MatchError("resources delete failed"))
		})
		It("overtime error", func() {
			mocker := setupSQLMock()
			expectDeleteClusterDepartmentsLookup(mocker, deleteClusterDepartmentsView())
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE resources.id = ? AND resources.id = ? AND resources.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE overtime_data SET deleted_at=? WHERE org_unit_data_id in (?) AND overtime_data.deleted_at IS NULL`).
				ReturnsError(errors.New("overtime delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteClusterDepartments(context.Background(), testTenantID, testClusterID())
			Expect(err).To(MatchError("overtime delete failed"))
		})
		It("org unit error", func() {
			mocker := setupSQLMock()
			expectDeleteClusterDepartmentsLookup(mocker, deleteClusterDepartmentsView())
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE resources.id = ? AND resources.id = ? AND resources.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE overtime_data SET deleted_at=? WHERE org_unit_data_id in (?) AND overtime_data.deleted_at IS NULL`).Returns(0)
			mocker.ExpectExec(`UPDATE org_unit_data SET deleted_at=? WHERE org_unit_data.id = ? AND org_unit_data.id = ? AND org_unit_data.deleted_at IS NULL`).
				ReturnsError(errors.New("org unit delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteClusterDepartments(context.Background(), testTenantID, testClusterID())
			Expect(err).To(MatchError("org unit delete failed"))
		})
		It("department error", func() {
			mocker := setupSQLMock()
			expectDeleteClusterDepartmentsLookup(mocker, deleteClusterDepartmentsView())
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`UPDATE resources SET deleted_at=? WHERE resources.id = ? AND resources.id = ? AND resources.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE overtime_data SET deleted_at=? WHERE org_unit_data_id in (?) AND overtime_data.deleted_at IS NULL`).Returns(0)
			mocker.ExpectExec(`UPDATE org_unit_data SET deleted_at=? WHERE org_unit_data.id = ? AND org_unit_data.id = ? AND org_unit_data.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE department SET deleted_at=? WHERE department.id = ? AND department.id = ? AND department.deleted_at IS NULL`).
				ReturnsError(errors.New("department delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteClusterDepartments(context.Background(), testTenantID, testClusterID())
			Expect(err).To(MatchError("department delete failed"))
		})
	})
	Describe("UpdateDepartment", func() {
		It("returns error", func() {
			mocker := setupSQLMock()
			dept := testDepartment()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`DELETE FROM org_unit_train_node_type WHERE org_unit_train_node_type.org_unit_data_id = ?`, testOrgUnitDataID()).
				ReturnsError(errors.New("clear failed"))
			mocker.ExpectTransaction().Rollback()

			got, err := UpdateDepartment(context.Background(), &dept)
			Expect(err).To(MatchError("clear failed"))
			Expect(got).To(BeNil())
		})
	})
	Describe("UpdateDepartmentWithChildren", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			dept := testDepartment()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlDeleteTrainNodeTypesByOrgUnit, testOrgUnitDataID()).Returns(0)
			mocker.ExpectExec(sqlDeleteWorkspaceNodeTypesByOrgUnit, testOrgUnitDataID()).Returns(0)
			mocker.ExpectExec(sqlUpdateDepartment).Returns(1)
			mocker.ExpectExec(sqlUpdateDepartmentChildrenEffective).Returns(1)
			mocker.ExpectTransaction().Commit()

			got, err := UpdateDepartmentWithChildren(context.Background(), &dept, &[]models.Project{project})
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal("research"))
		})
		It("workspace clear error", func() {
			mocker := setupSQLMock()
			dept := testDepartment()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlDeleteTrainNodeTypesByOrgUnit, testOrgUnitDataID()).Returns(0)
			mocker.ExpectExec(sqlDeleteWorkspaceNodeTypesByOrgUnit, testOrgUnitDataID()).
				ReturnsError(errors.New("clear workspace failed"))
			mocker.ExpectTransaction().Rollback()

			got, err := UpdateDepartment(context.Background(), &dept)
			Expect(err).To(MatchError("clear workspace failed"))
			Expect(got).To(BeNil())
		})
		It("update error", func() {
			mocker := setupSQLMock()
			dept := testDepartment()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlDeleteTrainNodeTypesByOrgUnit, testOrgUnitDataID()).Returns(0)
			mocker.ExpectExec(sqlDeleteWorkspaceNodeTypesByOrgUnit, testOrgUnitDataID()).Returns(0)
			mocker.ExpectExec(sqlUpdateDepartment).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			got, err := UpdateDepartment(context.Background(), &dept)
			Expect(err).To(MatchError("update failed"))
			Expect(got).To(BeNil())
		})
		It("children error", func() {
			mocker := setupSQLMock()
			dept := testDepartment()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlDeleteTrainNodeTypesByOrgUnit, testOrgUnitDataID()).Returns(0)
			mocker.ExpectExec(sqlDeleteWorkspaceNodeTypesByOrgUnit, testOrgUnitDataID()).Returns(0)
			mocker.ExpectExec(sqlUpdateDepartment).Returns(1)
			mocker.ExpectExec(sqlUpdateDepartmentChildrenEffective).ReturnsError(errors.New("children update failed"))
			mocker.ExpectTransaction().Rollback()

			got, err := UpdateDepartmentWithChildren(context.Background(), &dept, &[]models.Project{project})
			Expect(err).To(MatchError("children update failed"))
			Expect(got).To(BeNil())
		})
	})
})
