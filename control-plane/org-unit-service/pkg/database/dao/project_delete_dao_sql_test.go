package dao

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"k8s.io/utils/ptr"
)

const (
	sqlSetProjectDeleting           = `UPDATE project SET updated_at=?,deleted_by=?,tenant_id=?,cluster_id=?,name=?,parent_id=?,deletion_request=?,deletion_type=?,force_delete=?,org_unit_data_id=?,phase=?,quota_status_id=? WHERE (tenant_id = ? AND id = ?) AND project.deleted_at IS NULL AND id = ?`
	sqlSetProjectDeletingNoSubject  = `UPDATE project SET updated_at=?,tenant_id=?,cluster_id=?,name=?,parent_id=?,deletion_request=?,deletion_type=?,force_delete=?,org_unit_data_id=?,phase=?,quota_status_id=? WHERE (tenant_id = ? AND id = ?) AND project.deleted_at IS NULL AND id = ?`
	sqlSoftDeleteResourcesByOrgUnit = `UPDATE resources SET deleted_at=? WHERE org_unit_data_id = ? AND resources.deleted_at IS NULL`
	sqlDeleteTrainNodeTypes         = `DELETE FROM org_unit_train_node_type WHERE org_unit_data_id = ?`
	sqlDeleteWorkspaceNodeTypes     = `DELETE FROM org_unit_workspace_node_type WHERE org_unit_data_id = ?`
	sqlSoftDeleteOrgUnitData        = `UPDATE org_unit_data SET deleted_at=? WHERE id = ? AND org_unit_data.deleted_at IS NULL`
	sqlSoftDeleteOvertimeData       = `UPDATE overtime_data SET deleted_at=? WHERE org_unit_data_id = ? AND overtime_data.deleted_at IS NULL`
	sqlSoftDeleteQuotaStatus        = `UPDATE quota_status SET deleted_at=? WHERE id = ? AND quota_status.deleted_at IS NULL`
	sqlSoftDeleteProject            = `UPDATE project SET deleted_at=? WHERE project.id = ? AND project.id = ? AND project.deleted_at IS NULL`
)

var _ = Describe("Project delete DAO SQL", func() {
	Describe("SetProjectAsDeleting", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			project := testProject()
			subject := "user@run.ai"
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlSetProjectDeleting).Returns(1)
			mocker.ExpectTransaction().Commit()

			name, err := SetProjectAsDeleting(context.Background(), &project, ptr.To(subject), api.DELETIONTYPES_ORPHAN, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("proj-a"))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlSetProjectDeletingNoSubject).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			name, err := SetProjectAsDeleting(context.Background(), &project, nil, api.DELETIONTYPES_ORPHAN, false)
			Expect(err).To(MatchError("update failed"))
			Expect(name).To(BeEmpty())
		})
	})
	Describe("DeleteProjectInDb", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			project := testProject()
			subject := "user@run.ai"
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlSoftDeleteResourcesByOrgUnit, dbtesting.AnyTime(), testOrgUnitDataID()).Returns(1)
			mocker.ExpectExec(sqlDeleteTrainNodeTypes).Returns(0)
			mocker.ExpectExec(sqlDeleteWorkspaceNodeTypes).Returns(0)
			mocker.ExpectExec(sqlSoftDeleteOrgUnitData, dbtesting.AnyTime(), testOrgUnitDataID()).Returns(1)
			mocker.ExpectExec(sqlSoftDeleteOvertimeData, dbtesting.AnyTime(), testOrgUnitDataID()).Returns(1)
			mocker.ExpectExec(sqlSoftDeleteQuotaStatus, dbtesting.AnyTime(), testQuotaStatusID()).Returns(1)
			mocker.ExpectExec(sqlSoftDeleteProject, dbtesting.AnyTime(), testProjectID, testProjectID).Returns(1)
			mocker.ExpectTransaction().Commit()

			name, parentID, err := DeleteProjectInDb(context.Background(), &project, ptr.To(subject), api.DELETIONTYPES_ORPHAN, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("proj-a"))
			Expect(parentID).To(Equal(testDepartmentID))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlSoftDeleteResourcesByOrgUnit).ReturnsError(errors.New("delete failed"))
			mocker.ExpectTransaction().Rollback()

			name, parentID, err := DeleteProjectInDb(context.Background(), &project, nil, api.DELETIONTYPES_ORPHAN, false)
			Expect(err).To(MatchError("delete failed"))
			Expect(name).To(BeEmpty())
			Expect(parentID).To(Equal(int64(-1)))
		})
	})
	Describe("DeleteProjectsInDb", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			project := testProject()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`SAVEPOINT sp`).Returns(0)
			mocker.ExpectExec(sqlSoftDeleteResourcesByOrgUnit, dbtesting.AnyTime(), testOrgUnitDataID()).Returns(1)
			mocker.ExpectExec(sqlDeleteTrainNodeTypes).Returns(0)
			mocker.ExpectExec(sqlDeleteWorkspaceNodeTypes).Returns(0)
			mocker.ExpectExec(sqlSoftDeleteOrgUnitData, dbtesting.AnyTime(), testOrgUnitDataID()).Returns(1)
			mocker.ExpectExec(sqlSoftDeleteOvertimeData, dbtesting.AnyTime(), testOrgUnitDataID()).Returns(1)
			mocker.ExpectExec(sqlSoftDeleteQuotaStatus, dbtesting.AnyTime(), testQuotaStatusID()).Returns(1)
			mocker.ExpectExec(sqlSoftDeleteProject, dbtesting.AnyTime(), testProjectID, testProjectID).Returns(1)
			mocker.ExpectExec(`SAVEPOINT sp`).Returns(0)
			mocker.ExpectTransaction().Commit()

			got, err := DeleteProjectsInDb(context.Background(), []*models.Project{&project})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
		})
	})
})
