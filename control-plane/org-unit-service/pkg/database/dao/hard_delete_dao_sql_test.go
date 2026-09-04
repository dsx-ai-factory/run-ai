package dao

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
)

const (
	sqlSelectAllTenantNodeTypes      = `SELECT * FROM node_type WHERE node_type.tenant_id = ?`
	sqlHardDeleteOvertimeData        = `DELETE FROM overtime_data WHERE org_unit_data_id in (?,?)`
	sqlHardDeleteTrainByNodeType     = `DELETE FROM org_unit_train_node_type WHERE node_type_id IN (?)`
	sqlHardDeleteWorkspaceByNodeType = `DELETE FROM org_unit_workspace_node_type WHERE node_type_id IN (?)`
	sqlHardDeleteNodeTypes           = `DELETE FROM node_type WHERE node_type.id = ?`
)

var _ = Describe("Hard delete DAO SQL", func() {
	Describe("GetAllTenantNodeTypes", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testNodeType()
			mocker.ExpectQuery(sqlSelectAllTenantNodeTypes, testTenantID).
				Returns([]models.NodeType{expected})

			got, err := GetAllTenantNodeTypes(context.Background(), testTenantID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal(expected.Name))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectAllTenantNodeTypes, testTenantID).
				ReturnsError(errors.New("db error"))

			got, err := GetAllTenantNodeTypes(context.Background(), testTenantID)
			Expect(err).To(MatchError("db error"))
			Expect(got).To(BeNil())
		})
	})
	Describe("HardDeleteOvertimeData", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			project := testProject()
			department := testDepartment()
			deptOrgID := testOrgUnitDataID()
			projOrgID := testResourcesID()
			department.OrgUnitData.ID = deptOrgID
			project.OrgUnitData.ID = projOrgID
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlHardDeleteOvertimeData, deptOrgID, projOrgID).Returns(2)
			mocker.ExpectTransaction().Commit()

			err := HardDeleteOvertimeData(context.Background(), testTenantID, []*models.Project{&project}, []*models.Department{&department})
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			project := testProject()
			department := testDepartment()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlHardDeleteOvertimeData).ReturnsError(errors.New("delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := HardDeleteOvertimeData(context.Background(), testTenantID, []*models.Project{&project}, []*models.Department{&department})
			Expect(err).To(MatchError("delete failed"))
		})
	})
	Describe("HardDeleteTenantNodeTypes", func() {
		It("when empty", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectAllTenantNodeTypes, testTenantID).
				Returns([]models.NodeType{})

			err := HardDeleteTenantNodeTypes(context.Background(), testTenantID)
			Expect(err).NotTo(HaveOccurred())
		})
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testNodeType()
			mocker.ExpectQuery(sqlSelectAllTenantNodeTypes, testTenantID).
				Returns([]models.NodeType{expected})
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlHardDeleteTrainByNodeType, testNodeTypeID).Returns(1)
			mocker.ExpectExec(sqlHardDeleteWorkspaceByNodeType, testNodeTypeID).Returns(1)
			mocker.ExpectExec(sqlHardDeleteNodeTypes, testNodeTypeID).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := HardDeleteTenantNodeTypes(context.Background(), testTenantID)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("HardDeleteTenantProjects", func() {
		It("when empty", func() {
			setupSQLMock()
			err := HardDeleteTenantProjects(context.Background(), testTenantID, nil)
			Expect(err).NotTo(HaveOccurred())
		})
		It("succeeds", func() {
			mocker := setupSQLMock()
			project := testProject()
			project.OrgUnitData.ID = testOrgUnitDataID()
			project.OrgUnitData.Resources = []models.Resources{testResources()}
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`DELETE FROM project WHERE project.id = ?`, testProjectID).Returns(1)
			mocker.ExpectExec(`DELETE FROM resources WHERE resources.id = ?`, testResourcesID()).Returns(1)
			mocker.ExpectExec(`DELETE FROM org_unit_data WHERE id = ?`, testOrgUnitDataID()).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := HardDeleteTenantProjects(context.Background(), testTenantID, []*models.Project{&project})
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("HardDeleteTenantDepartments", func() {
		It("when empty", func() {
			setupSQLMock()
			err := HardDeleteTenantDepartments(context.Background(), testTenantID, nil)
			Expect(err).NotTo(HaveOccurred())
		})
		It("succeeds", func() {
			mocker := setupSQLMock()
			department := testDepartment()
			department.OrgUnitData.ID = testOrgUnitDataID()
			department.OrgUnitData.Resources = []models.Resources{testResources()}
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`DELETE FROM resources WHERE resources.id = ?`, testResourcesID()).Returns(1)
			mocker.ExpectExec(`DELETE FROM department WHERE department.id = ?`, testDepartmentID).Returns(1)
			mocker.ExpectExec(`DELETE FROM org_unit_data WHERE org_unit_data.id = ?`, testOrgUnitDataID()).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := HardDeleteTenantDepartments(context.Background(), testTenantID, []*models.Department{&department})
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("GetAllTenantProjects", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testProjectView()
			mocker.ExpectQuery(`SELECT * FROM project_view WHERE project_view.tenant_id = ? ORDER BY id DESC`, testTenantID).
				Returns(projectViewRows(expected))

			got, err := GetAllTenantProjects(context.Background(), testTenantID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal(expected.Name))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(`SELECT * FROM project_view WHERE project_view.tenant_id = ? ORDER BY id DESC`, testTenantID).
				ReturnsError(errors.New("db error"))

			got, err := GetAllTenantProjects(context.Background(), testTenantID)
			Expect(err).To(MatchError("db error"))
			Expect(got).To(BeNil())
		})
	})
	Describe("GetAllTenantDepartments", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testDepartmentView()
			mocker.ExpectQuery(`SELECT * FROM department_view WHERE department_view.tenant_id = ? ORDER BY id DESC`, testTenantID).
				Returns([]models.DepartmentView{expected})

			got, err := GetAllTenantDepartments(context.Background(), testTenantID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal(expected.Name))
		})
	})
})
