package dao

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lib/pq"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"k8s.io/utils/ptr"
)

const (
	sqlUpdateOrgUnitDefaultNodePools = `UPDATE org_unit_data SET default_node_pool_list=?,default_node_pool_list_effective=?,default_node_pool_list_effective_origin=? WHERE id = ?`
)

var _ = Describe("Org unit data DAO SQL", func() {
	Describe("SaveOrgUnitDataDefaultNodepoolLists", func() {
		It("succeeds", func() {
			mocker := setupSQLMock()
			list := pq.StringArray{testNodePoolName}
			orgUnit := models.OrgUnitData{
				BaseModel: dbmodels.BaseModel{ID: testOrgUnitDataID()},
				EffectiveData: models.EffectiveData{
					DefaultNodePoolListEffective:       &list,
					DefaultNodePoolListEffectiveOrigin: ptr.To(testOrgUnitDataID()),
				},
				DefaultNodePoolList: &list,
			}
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`SAVEPOINT sp`).Returns(0)
			mocker.ExpectExec(sqlUpdateOrgUnitDefaultNodePools, dbtesting.AnyArg(), dbtesting.AnyArg(), testOrgUnitDataID(), testOrgUnitDataID()).Returns(1)
			mocker.ExpectExec(`SAVEPOINT sp`).Returns(0)
			mocker.ExpectTransaction().Commit()

			err := SaveOrgUnitDataDefaultNodepoolLists(context.Background(), []models.OrgUnitData{orgUnit})
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error", func() {
			mocker := setupSQLMock()
			orgUnit := models.OrgUnitData{BaseModel: dbmodels.BaseModel{ID: testOrgUnitDataID()}}
			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(`SAVEPOINT sp`).Returns(0)
			mocker.ExpectExec(sqlUpdateOrgUnitDefaultNodePools).ReturnsError(errors.New("update failed"))
			mocker.ExpectExec(`ROLLBACK TO SAVEPOINT sp`).Returns(0)
			mocker.ExpectTransaction().Commit()

			err := SaveOrgUnitDataDefaultNodepoolLists(context.Background(), []models.OrgUnitData{orgUnit})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
