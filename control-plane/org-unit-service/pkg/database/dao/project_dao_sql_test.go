package dao

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	telemetry "github.com/run-ai/runai/runai-application-packages/telemetry"
	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"
)

const (
	sqlSelectProjectClusterID      = `SELECT cluster_id FROM project WHERE id = ? AND tenant_id = ? AND project.deleted_at IS NULL ORDER BY project.id LIMIT ?`
	sqlSelectProjectByID           = `SELECT * FROM project WHERE (project.id = ? AND project.tenant_id = ?) AND project.deleted_at IS NULL AND project.id = ? ORDER BY project.id LIMIT ?`
	sqlSelectProjectViewByID       = `SELECT * FROM project_view WHERE (project_view.id = ? AND project_view.tenant_id = ?) AND project_view.deleted_at IS NULL ORDER BY project_view.id LIMIT ?`
	sqlSelectProjectViewByIDFirst  = `SELECT * FROM project_view WHERE (project_view.id = ? AND project_view.tenant_id = ?) AND project_view.deleted_at IS NULL AND project_view.id = ? ORDER BY project_view.id LIMIT ?`
	sqlSelectProjectByName         = `SELECT * FROM project WHERE (project.tenant_id = ? AND project.cluster_id = ? AND project.name = ?) AND project.deleted_at IS NULL ORDER BY project.id LIMIT ?`
	sqlSelectProjectsByParent      = `SELECT * FROM project WHERE (project.tenant_id = ? AND project.parent_id = ?) AND project.deleted_at IS NULL`
	sqlSelectProjectsViewByParent  = `SELECT * FROM project_view WHERE (project_view.tenant_id = ? AND project_view.parent_id = ?) AND project_view.deleted_at IS NULL`
	sqlSelectProjectIDsByDept      = `SELECT id,parent_id FROM project WHERE tenant_id = ? AND parent_id IN (?) AND deleted_at IS NULL`
	sqlCountProjectByNamespace     = `SELECT count(*) FROM project WHERE cluster_id = ? AND (requested_namespace = ? OR status_namespace = ?) AND deleted_at is null`
	sqlSelectSyncProjectsView      = `SELECT * FROM (SELECT * FROM get_sync_projects(?, ?)) AS sync_view ORDER BY status_last_updated_time desc,calculated_updated_at,id`
	sqlUpdateProjectStatus         = `UPDATE project SET conditions=?,phase=?,phase_message=?,status_last_updated_time=?,status_namespace=? WHERE id = ? AND tenant_id = ?`
	sqlCountProjects               = `SELECT count(*) FROM project WHERE deleted_at IS NULL AND tenant_id = ?`
	sqlSelectDeletedProjects       = `SELECT * FROM project WHERE tenant_id = ? AND deleted_at is not null AND deleted_at > ?`
	sqlSelectProjectsByClusters    = `SELECT * FROM project WHERE tenant_id = ? AND cluster_id in (?,?) AND project.deleted_at IS NULL`
	sqlSelectProjects              = `SELECT * FROM project WHERE project.tenant_id = ? AND project.deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`
	sqlSelectProjectsViewByCluster = `SELECT * FROM project_view WHERE project_view.tenant_id = ? AND cluster_id = ? AND project_view.deleted_at IS NULL ORDER BY id DESC`
	sqlSelectProjectsByIDs         = `SELECT * FROM project_view WHERE project_view.tenant_id = ? AND id in (?) AND cluster_id = ? AND project_view.deleted_at IS NULL`
	sqlUpdateQuotaStatus           = `UPDATE quota_status SET cpu_allocated=?,cpu_allocated_non_preemptible=?,cpu_requested=?,gpu_allocated=?,gpu_allocated_non_preemptible=?,gpu_requested=?,memory_allocated=?,memory_allocated_non_preemptible=?,memory_requested=? WHERE id = ?`
	sqlSelectProjectTelemetry      = `SELECT SUM(gpu_deserved) as data, cluster_id as group1_main_value FROM project JOIN resources on project.org_unit_data_id = resources.org_unit_data_id WHERE tenant_id = ? AND project.deleted_at IS NULL AND resources.deleted_at IS NULL GROUP BY cluster_id`
)

const sqlUpdateResourcesStatus = `UPDATE resources SET allocated_cpu=?,allocated_gpu=?,allocated_memory=?,allocated_non_preemptible_cpu=?,allocated_non_preemptible_gpu=?,allocated_non_preemptible_memory=?,requested_cpu=?,requested_gpu=?,requested_memory=? WHERE org_unit_data_id = ? AND node_pool_name = ?`

var _ = Describe("Project DAO SQL", func() {
	Describe("GetClusterIdByProject", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			clusterID := testClusterID().String()
			mocker.ExpectQuery(sqlSelectProjectClusterID, testProjectID, testTenantID, 1).
				Returns([]clusterIDRow{{ClusterId: clusterID}})

			got, err := GetClusterIdByProject(context.Background(), testProjectID, testTenantID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(clusterID))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectProjectClusterID, testProjectID, testTenantID, 1).
				ReturnsError(gorm.ErrRecordNotFound)

			got, err := GetClusterIdByProject(context.Background(), testProjectID, testTenantID)
			Expect(err).To(MatchError(gorm.ErrRecordNotFound))
			Expect(got).To(BeEmpty())
		})
	})
	Describe("GetProjectById", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProject()
			mocker.ExpectQuery(sqlSelectProjectByID, testProjectID, testTenantID, testProjectID, 1).
				Returns(projectRows(expected))

			got, err := GetProjectById(context.Background(), testProjectID, testTenantID, models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(expected.Name))
			Expect(got.ID).To(Equal(expected.ID))
		})
		It("with slim verbosity", func() {
			mocker := setupSQLMock()
			expected := testProjectView()
			mocker.ExpectQuery(sqlSelectProjectViewByID, testProjectID, testTenantID, 1).
				Returns(projectViewRows(expected))

			got, err := GetProjectById(context.Background(), testProjectID, testTenantID, models.SlimVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(expected.Name))
		})
		It("returns not found", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectProjectByID, testProjectID, testTenantID, testProjectID, 1).
				ReturnsError(gorm.ErrRecordNotFound)

			got, err := GetProjectById(context.Background(), testProjectID, testTenantID, models.MetaDataVerbosity)
			Expect(err).To(MatchError(gorm.ErrRecordNotFound))
			Expect(got).To(BeNil())
		})
	})
	Describe("GetProjectByIdWithPermittedScopes", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProjectView()
			mocker.ExpectQuery(sqlSelectProjectViewByIDFirst, testProjectID, testTenantID, testProjectID, 1).
				Returns(projectViewRows(expected))

			got, err := GetProjectByIdWithPermittedScopes(context.Background(), testProjectID, testTenantID, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(expected.Name))
		})
	})
	Describe("GetProjectByNameAndCluster", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProject()
			mocker.ExpectQuery(sqlSelectProjectByName, testTenantID, testClusterID(), "proj-a", 1).
				Returns(projectRows(expected))

			got, err := GetProjectByNameAndCluster(context.Background(), "proj-a", testClusterID(), testTenantID, models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(expected.Name))
		})
	})
	Describe("GetProjectsByParent", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProject()
			mocker.ExpectQuery(sqlSelectProjectsByParent, testTenantID, testDepartmentID).
				Returns(projectRows(expected))

			got, err := GetProjectsByParent(context.Background(), testDepartmentID, testTenantID, models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
		})
		It("with slim verbosity", func() {
			mocker := setupSQLMock()
			expected := testProjectView()
			mocker.ExpectQuery(sqlSelectProjectsViewByParent, testTenantID, testDepartmentID).
				Returns(projectViewRows(expected))

			got, err := GetProjectsByParent(context.Background(), testDepartmentID, testTenantID, models.SlimVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
		})
	})
	Describe("GetProjectsIdsByDepartmentIds", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProject()
			mocker.ExpectQuery(sqlSelectProjectIDsByDept, testTenantID, testDepartmentID).
				Returns(projectRows(expected))

			got, err := GetProjectsIdsByDepartmentIds(context.Background(), testTenantID, []int64{testDepartmentID})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].ID).To(Equal(testProjectID))
		})
	})
	Describe("CountProjectByNamespace", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlCountProjectByNamespace, testClusterID().String(), testNamespace, testNamespace).
				Returns([]countResult{{Count: 1}})

			got, err := CountProjectByNamespace(context.Background(), testClusterID(), testNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(int64(1)))
		})
	})
	Describe("CountProjects", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlCountProjects, testTenantID).Returns([]countResult{{Count: 4}})

			got, err := CountProjects(context.Background(), testTenantID, nil, nil, "", models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(int64(4)))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlCountProjects, testTenantID).ReturnsError(errors.New("db error"))

			got, err := CountProjects(context.Background(), testTenantID, nil, nil, "", models.MetaDataVerbosity)
			Expect(err).To(MatchError("db error"))
			Expect(got).To(Equal(int64(0)))
		})
	})
	Describe("GetDeletedProjects", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			deletedAfter := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			expected := testProject()
			mocker.ExpectQuery(sqlSelectDeletedProjects, testTenantID, deletedAfter).
				Returns(projectRows(expected))

			got, err := GetDeletedProjects(context.Background(), testTenantID, deletedAfter)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
		})
	})
	Describe("GetProjectsByClusters", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			cluster1 := testClusterID().String()
			cluster2 := "22222222-2222-2222-2222-222222222222"
			expected := testProject()
			mocker.ExpectQuery(sqlSelectProjectsByClusters, testTenantID, cluster1, cluster2).
				Returns(projectRows(expected))

			got, err := GetProjectsByClusters(context.Background(), testTenantID, []string{cluster1, cluster2})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
		})
	})
	Describe("GetProjects", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProject()
			mocker.ExpectQuery(sqlSelectProjects, testTenantID, 10, 5).
				Returns(projectRows(expected))

			got, err := GetProjects(context.Background(), testTenantID, nil, "", "", 5, 10, "", nil, false, models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
		})
		It("with filter", func() {
			mocker := setupSQLMock()
			expected := testProject()
			mocker.ExpectQuery(
				`SELECT * FROM project WHERE project.tenant_id = ? AND cluster_id = ? AND project.deleted_at IS NULL ORDER BY id DESC`,
				testTenantID, testClusterID().String(),
			).Returns(projectRows(expected))

			got, err := GetProjects(context.Background(), testTenantID, []validators.ParameterizedFilter{
				{Field: "cluster_id", Operator: "=", FormattedValues: []string{testClusterID().String()}},
			}, "", "", 0, 0, "", nil, false, models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
		})
	})
	Describe("GetProjectsByCluster", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProjectView()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, testClusterID().String()).
				Returns(projectViewRows(expected))

			got, err := GetProjectsByCluster(context.Background(), testTenantID, testClusterID(), nil, models.SlimVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal(expected.Name))
		})
	})
	Describe("GetProjectsByIds", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProjectView()
			clusterID := testClusterID()
			mocker.ExpectQuery(sqlSelectProjectsByIDs, testTenantID, "22", clusterID).
				Returns(projectViewRows(expected))

			got, err := GetProjectsByIds(context.Background(), testTenantID, &clusterID, []string{"22"}, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
		})
	})
	Describe("GetSyncProjectsView", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := models.ProjectSyncView{
				IntBaseModel:  models.IntBaseModel{ID: testProjectID},
				TenantId:      testTenantID,
				ClusterId:     testClusterID(),
				Name:          "proj-a",
				OrgUnitDataID: testOrgUnitDataID(),
			}
			mocker.ExpectQuery(sqlSelectSyncProjectsView, testTenantID, testClusterID()).
				Returns([]models.ProjectSyncView{expected})

			got, err := GetSyncProjectsView(context.Background(), testTenantID, testClusterID(), time.Time{}, 0, 0, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
			Expect((*got)[0].Name).To(Equal(expected.Name))
		})
	})
	Describe("UpdateProjectStatus", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			now := time.Now()
			project := testProject()
			status := models.ProjectStatus{
				Phase:                 ptr.To("Ready"),
				PhaseMessage:          ptr.To("ok"),
				StatusNamespace:       ptr.To(testNamespace),
				StatusLastUpdatedTime: &now,
				QuotaStatusID:         testQuotaStatusID(),
			}
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateProjectStatus).Returns(1)
			mocker.ExpectExec(sqlUpdateQuotaStatus).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := UpdateProjectStatus(context.Background(), testTenantID, testClusterID(), &project, status, nil)
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateProjectStatus).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			err := UpdateProjectStatus(context.Background(), testTenantID, testClusterID(), &project, models.ProjectStatus{}, nil)
			Expect(err).To(MatchError("update failed"))
		})
		It("quota error", func() {
			mocker := setupSQLMock()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateProjectStatus).Returns(1)
			mocker.ExpectExec(sqlUpdateQuotaStatus).ReturnsError(errors.New("quota update failed"))
			mocker.ExpectTransaction().Rollback()

			err := UpdateProjectStatus(context.Background(), testTenantID, testClusterID(), &project, models.ProjectStatus{}, nil)
			Expect(err).To(MatchError("quota update failed"))
		})
		It("with resources", func() {
			mocker := setupSQLMock()
			now := time.Now()
			project := testProject()
			gpu := 1.0
			status := models.ProjectStatus{
				Phase:                 ptr.To("Ready"),
				PhaseMessage:          ptr.To("ok"),
				StatusNamespace:       ptr.To(testNamespace),
				StatusLastUpdatedTime: &now,
				QuotaStatusID:         testQuotaStatusID(),
			}
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateProjectStatus).Returns(1)
			mocker.ExpectExec(sqlUpdateQuotaStatus).Returns(1)
			mocker.ExpectExec(sqlUpdateResourcesStatus).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := UpdateProjectStatus(context.Background(), testTenantID, testClusterID(), &project, status, map[string]models.ResourcesStatus{
				testNodePoolName: {AllocatedGpu: &gpu},
			})
			Expect(err).NotTo(HaveOccurred())
		})
		It("resources error", func() {
			mocker := setupSQLMock()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateProjectStatus).Returns(1)
			mocker.ExpectExec(sqlUpdateQuotaStatus).Returns(1)
			mocker.ExpectExec(sqlUpdateResourcesStatus).ReturnsError(errors.New("resources update failed"))
			mocker.ExpectTransaction().Rollback()

			err := UpdateProjectStatus(context.Background(), testTenantID, testClusterID(), &project, models.ProjectStatus{}, map[string]models.ResourcesStatus{
				testNodePoolName: {},
			})
			Expect(err).To(MatchError("resources update failed"))
		})
	})
	Describe("GetProjectResourcesTelemetry", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectProjectTelemetry, testTenantID).
				Returns([]telemetry.TelemetryGroupValue{{Data: "4", Group1MainValue: "cluster-1"}})

			got, err := GetProjectResourcesTelemetry(
				context.Background(),
				"gpu_deserved",
				testTenantID,
				nil,
				"",
				"",
				"",
				"",
				[][]string{{"cluster_id"}},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectProjectTelemetry, testTenantID).ReturnsError(errors.New("db error"))

			got, err := GetProjectResourcesTelemetry(
				context.Background(),
				"gpu_deserved",
				testTenantID,
				nil,
				"",
				"",
				"",
				"",
				[][]string{{"cluster_id"}},
			)
			Expect(err).To(MatchError("db error"))
			Expect(got).NotTo(BeNil())
		})
	})
	Describe("GetProjectByIdAndCluster", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProject()
			mocker.ExpectQuery(
				`SELECT * FROM project WHERE (project.id = ? AND project.tenant_id = ? AND project.cluster_id = ?) AND project.deleted_at IS NULL AND project.id = ? ORDER BY project.id LIMIT ?`,
				testProjectID, testTenantID, testClusterID(), testProjectID, 1,
			).Returns(projectRows(expected))

			got, err := GetProjectByIdAndCluster(context.Background(), testProjectID, testTenantID, testClusterID(), models.MetaDataVerbosity)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(expected.Name))
		})
	})
	Describe("CreateProject", func() {
		It("returns error", func() {
			mocker := setupSQLMock()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectQuery(`INSERT INTO project (created_at,updated_at,deleted_at,created_by,updated_by,deleted_by,tenant_id,cluster_id,name,description,requested_namespace,enforce_runai_scheduler,parent_id,parent_name,deletion_request,deletion_type,force_delete,org_unit_data_id,status_last_updated_time,phase,phase_message,status_namespace,conditions,quota_status_id,id) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING id`).ReturnsError(errors.New("insert failed"))
			mocker.ExpectTransaction().Rollback()

			got, err := CreateProject(context.Background(), &project)
			Expect(err).To(MatchError("insert failed"))
			Expect(got).To(BeNil())
		})
	})
	Describe("CreateProjects", func() {
		It("returns error", func() {
			mocker := setupSQLMock()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectQuery(`INSERT INTO project (created_at,updated_at,deleted_at,created_by,updated_by,deleted_by,tenant_id,cluster_id,name,description,requested_namespace,enforce_runai_scheduler,parent_id,parent_name,deletion_request,deletion_type,force_delete,org_unit_data_id,status_last_updated_time,phase,phase_message,status_namespace,conditions,quota_status_id,id) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT (id) DO UPDATE SET updated_at=?,deleted_at=excluded.deleted_at,created_by=excluded.created_by,updated_by=excluded.updated_by,deleted_by=excluded.deleted_by,tenant_id=excluded.tenant_id,cluster_id=excluded.cluster_id,name=excluded.name,description=excluded.description,requested_namespace=excluded.requested_namespace,enforce_runai_scheduler=excluded.enforce_runai_scheduler,parent_id=excluded.parent_id,parent_name=excluded.parent_name,deletion_request=excluded.deletion_request,deletion_type=excluded.deletion_type,force_delete=excluded.force_delete,org_unit_data_id=excluded.org_unit_data_id,status_last_updated_time=excluded.status_last_updated_time,phase=excluded.phase,phase_message=excluded.phase_message,status_namespace=excluded.status_namespace,conditions=excluded.conditions,quota_status_id=excluded.quota_status_id RETURNING id`).ReturnsError(errors.New("insert failed"))
			mocker.ExpectTransaction().Rollback()

			err := CreateProjects(context.Background(), []*models.Project{&project})
			Expect(err).To(MatchError("insert failed"))
		})
	})
	Describe("DeleteClusterProjects", func() {
		It("when empty", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, testClusterID().String()).
				Returns([]projectRow{})

			err := DeleteClusterProjects(context.Background(), testTenantID, testClusterID())
			Expect(err).NotTo(HaveOccurred())
		})
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProjectView()
			mocker.ExpectQuery(sqlSelectProjectsViewByCluster, testTenantID, testClusterID().String()).
				Returns(projectViewRows(expected))
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`UPDATE overtime_data SET deleted_at=? WHERE org_unit_data_id in (?) AND overtime_data.deleted_at IS NULL`).Returns(0)
			mocker.ExpectExec(`UPDATE org_unit_data SET deleted_at=? WHERE org_unit_data.id = ? AND org_unit_data.deleted_at IS NULL`).Returns(1)
			mocker.ExpectExec(`UPDATE project SET deleted_at=? WHERE project.id = ? AND project.deleted_at IS NULL`).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := DeleteClusterProjects(context.Background(), testTenantID, testClusterID())
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("GetSyncProjects", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := models.ProjectSyncView{
				IntBaseModel:  models.IntBaseModel{ID: testProjectID},
				TenantId:      testTenantID,
				ClusterId:     testClusterID(),
				Name:          "proj-a",
				OrgUnitDataID: testOrgUnitDataID(),
			}
			mocker.ExpectQuery(sqlSelectSyncProjectsView, testTenantID, testClusterID()).
				Returns([]models.ProjectSyncView{expected})

			got, err := GetSyncProjects(context.Background(), testTenantID, testClusterID(), time.Time{}, 0, 0, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal(expected.Name))
		})
	})
	Describe("UpdateProject", func() {
		It("returns error", func() {
			mocker := setupSQLMock()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`UPDATE project SET created_at=?,updated_at=?,created_by=?,updated_by=?,tenant_id=?,cluster_id=?,name=?,description=?,requested_namespace=?,enforce_runai_scheduler=?,parent_id=?,parent_name=?,org_unit_data_id=?,phase=?,phase_message=? WHERE (project.id = ? AND project.tenant_id = ? AND project.name = ?) AND project.deleted_at IS NULL AND id = ?`).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			got, err := UpdateProject(context.Background(), &project, false, false)
			Expect(err).To(MatchError("update failed"))
			Expect(got).To(BeNil())
		})
		It("succeeds", func() {
			mocker := setupSQLMock()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`UPDATE project SET created_at=?,updated_at=?,created_by=?,updated_by=?,tenant_id=?,cluster_id=?,name=?,description=?,requested_namespace=?,enforce_runai_scheduler=?,parent_id=?,parent_name=?,org_unit_data_id=?,phase=?,phase_message=? WHERE (project.id = ? AND project.tenant_id = ? AND project.name = ?) AND project.deleted_at IS NULL AND id = ?`).Returns(1)
			mocker.ExpectTransaction().Commit()

			got, err := UpdateProject(context.Background(), &project, false, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal("proj-a"))
		})
	})
})
