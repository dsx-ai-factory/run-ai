package dao

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"gorm.io/gorm"
)

const (
	sqlSelectNodeType                 = `SELECT * FROM node_type WHERE (node_type.id = ? AND node_type.cluster_id = ? AND node_type.tenant_id = ?) AND node_type.deleted_at IS NULL AND node_type.id = ? ORDER BY node_type.id LIMIT ?`
	sqlSelectNodeTypesByCluster       = `SELECT * FROM node_type WHERE (node_type.cluster_id = ? AND node_type.tenant_id = ?) AND node_type.deleted_at IS NULL`
	sqlInsertNodeType                 = `INSERT INTO node_type (created_at,updated_at,deleted_at,created_by,updated_by,deleted_by,name,cluster_id,tenant_id) VALUES (?,?,?,?,?,?,?,?,?) RETURNING id`
	sqlInsertNodeTypeOnConflict       = `INSERT INTO node_type (created_at,updated_at,deleted_at,created_by,updated_by,deleted_by,name,cluster_id,tenant_id) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT (id) DO UPDATE SET updated_at=?,deleted_at=excluded.deleted_at,created_by=excluded.created_by,updated_by=excluded.updated_by,deleted_by=excluded.deleted_by,name=excluded.name,cluster_id=excluded.cluster_id,tenant_id=excluded.tenant_id RETURNING id`
	sqlUpdateNodeType                 = `UPDATE node_type SET updated_at=?,name=?,cluster_id=?,tenant_id=? WHERE (node_type.id = ? AND node_type.cluster_id = ? AND node_type.tenant_id = ?) AND node_type.deleted_at IS NULL AND id = ?`
	sqlSoftDeleteNodeTypesByCluster   = `UPDATE node_type SET deleted_at=? WHERE (node_type.cluster_id = ? AND node_type.tenant_id = ?) AND node_type.deleted_at IS NULL`
	sqlSoftDeleteNodeTypeFromCluster  = `UPDATE node_type SET deleted_at=? WHERE (node_type.id = ? AND node_type.cluster_id = ? AND node_type.tenant_id = ?) AND node_type.id = ? AND node_type.deleted_at IS NULL`
	sqlDeleteOrgUnitTrainNodeType     = `DELETE FROM org_unit_train_node_type WHERE node_type_id = ?`
	sqlDeleteOrgUnitWorkspaceNodeType = `DELETE FROM org_unit_workspace_node_type WHERE node_type_id = ?`
)

var _ = Describe("Node type DAO SQL", func() {
	Describe("GetNodeType", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testNodeType()
			mocker.ExpectQuery(sqlSelectNodeType, testNodeTypeID, testClusterID(), testTenantID, testNodeTypeID, 1).
				Returns([]models.NodeType{expected})

			got, err := GetNodeType(context.Background(), testTenantID, testClusterID(), testNodeTypeID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(expected.Name))
			Expect(got.ID).To(Equal(expected.ID))
		})
		It("returns not found", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectNodeType, testNodeTypeID, testClusterID(), testTenantID, testNodeTypeID, 1).
				ReturnsError(gorm.ErrRecordNotFound)

			got, err := GetNodeType(context.Background(), testTenantID, testClusterID(), testNodeTypeID)
			Expect(err).To(MatchError(gorm.ErrRecordNotFound))
			Expect(got).To(BeNil())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectNodeType, testNodeTypeID, testClusterID(), testTenantID, testNodeTypeID, 1).
				ReturnsError(errors.New("db error"))

			got, err := GetNodeType(context.Background(), testTenantID, testClusterID(), testNodeTypeID)
			Expect(err).To(MatchError("db error"))
			Expect(got).To(BeNil())
		})
	})
	Describe("GetNodeTypes", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			expected := testNodeType()
			mocker.ExpectQuery(sqlSelectNodeTypesByCluster, testClusterID(), testTenantID).
				Returns([]models.NodeType{expected})

			got, err := GetNodeTypes(context.Background(), testTenantID, testClusterID(), nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*got).To(HaveLen(1))
			Expect((*got)[0].Name).To(Equal(expected.Name))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectQuery(sqlSelectNodeTypesByCluster, testClusterID(), testTenantID).
				ReturnsError(errors.New("db error"))

			got, err := GetNodeTypes(context.Background(), testTenantID, testClusterID(), nil)
			Expect(err).To(MatchError("db error"))
			Expect(got).To(BeNil())
		})
	})
	Describe("CreateNodeType", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectQuery(sqlInsertNodeType).Returns([]models.NodeType{{IntBaseModel: models.IntBaseModel{ID: testNodeTypeID}, Name: "h100"}})
			mocker.ExpectTransaction().Commit()

			nt := &models.NodeType{Name: "h100", ClusterId: testClusterID(), TenantId: testTenantID}
			got, err := CreateNodeType(context.Background(), nt)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal("h100"))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectQuery(sqlInsertNodeType).ReturnsError(errors.New("insert failed"))
			mocker.ExpectTransaction().Rollback()

			nt := &models.NodeType{Name: "h100", ClusterId: testClusterID(), TenantId: testTenantID}
			got, err := CreateNodeType(context.Background(), nt)
			Expect(err).To(MatchError("insert failed"))
			Expect(got).To(BeNil())
		})
	})
	Describe("UpdateNodeType", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateNodeType).Returns(1)
			mocker.ExpectTransaction().Commit()

			nt := &models.NodeType{
				IntBaseModel: models.IntBaseModel{ID: testNodeTypeID},
				Name:         "h200",
				ClusterId:    testClusterID(),
				TenantId:     testTenantID,
			}
			got, err := UpdateNodeType(context.Background(), nt)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal("h200"))
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlUpdateNodeType).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			nt := &models.NodeType{
				IntBaseModel: models.IntBaseModel{ID: testNodeTypeID},
				Name:         "h200",
				ClusterId:    testClusterID(),
				TenantId:     testTenantID,
			}
			got, err := UpdateNodeType(context.Background(), nt)
			Expect(err).To(MatchError("update failed"))
			Expect(got).To(BeNil())
		})
	})
	Describe("DeleteClusterNodeTypes", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlSoftDeleteNodeTypesByCluster, dbtesting.AnyTime(), testClusterID(), testTenantID).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := DeleteClusterNodeTypes(context.Background(), testTenantID, testClusterID())
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlSoftDeleteNodeTypesByCluster).ReturnsError(errors.New("delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteClusterNodeTypes(context.Background(), testTenantID, testClusterID())
			Expect(err).To(MatchError("delete failed"))
		})
	})
	Describe("DeleteNodeTypeFromCluster", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlDeleteOrgUnitTrainNodeType, testNodeTypeID).Returns(1)
			mocker.ExpectExec(sqlDeleteOrgUnitWorkspaceNodeType, testNodeTypeID).Returns(1)
			mocker.ExpectExec(sqlSoftDeleteNodeTypeFromCluster, dbtesting.AnyTime(), testNodeTypeID, testClusterID(), testTenantID, testNodeTypeID).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := DeleteNodeTypeFromCluster(context.Background(), testNodeTypeID, testTenantID, testClusterID())
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(sqlDeleteOrgUnitTrainNodeType, testNodeTypeID).ReturnsError(errors.New("delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteNodeTypeFromCluster(context.Background(), testNodeTypeID, testTenantID, testClusterID())
			Expect(err).To(MatchError("delete failed"))
		})
	})
	Describe("CreateNodeTypes", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectQuery(sqlInsertNodeTypeOnConflict).Returns([]models.NodeType{{IntBaseModel: models.IntBaseModel{ID: testNodeTypeID}, Name: "h100"}})
			mocker.ExpectTransaction().Commit()

			err := CreateNodeTypes(context.Background(), []*models.NodeType{{Name: "h100", ClusterId: testClusterID(), TenantId: testTenantID}})
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectQuery(sqlInsertNodeTypeOnConflict).ReturnsError(errors.New("insert failed"))
			mocker.ExpectTransaction().Rollback()

			err := CreateNodeTypes(context.Background(), []*models.NodeType{{Name: "h100", ClusterId: testClusterID(), TenantId: testTenantID}})
			Expect(err).To(MatchError("insert failed"))
		})
	})
	Describe("DeleteNodeType", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectQuery(`SELECT org_unit_data_id FROM org_unit_train_node_type WHERE node_type_id = ?`, testNodeTypeID).
				Returns([]struct{ OrgUnitDataId string }{})
			mocker.ExpectQuery(`SELECT org_unit_data_id FROM org_unit_workspace_node_type WHERE node_type_id = ?`, testNodeTypeID).
				Returns([]struct{ OrgUnitDataId string }{})
			mocker.ExpectExec(sqlDeleteOrgUnitTrainNodeType, testNodeTypeID).Returns(0)
			mocker.ExpectExec(sqlDeleteOrgUnitWorkspaceNodeType, testNodeTypeID).Returns(0)
			mocker.ExpectExec(`UPDATE org_unit_data SET updated_at=?,updated_by=? WHERE id in (NULL)`).Returns(0)
			mocker.ExpectExec(`UPDATE node_type SET deleted_at=? WHERE (node_type.id = ? AND node_type.tenant_id = ?) AND node_type.id = ? AND node_type.deleted_at IS NULL`).Returns(1)
			mocker.ExpectTransaction().Commit()

			err := DeleteNodeType(context.Background(), testNodeTypeID, testTenantID, "user@run.ai")
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			mocker.ExpectTransaction().Begin()
			mocker.ExpectQuery(`SELECT org_unit_data_id FROM org_unit_train_node_type WHERE node_type_id = ?`, testNodeTypeID).
				ReturnsError(errors.New("lookup failed"))
			mocker.ExpectTransaction().Rollback()

			err := DeleteNodeType(context.Background(), testNodeTypeID, testTenantID, "user@run.ai")
			Expect(err).To(MatchError("lookup failed"))
		})
	})
})
