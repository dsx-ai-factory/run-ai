package dao

import (
	"context"
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"k8s.io/utils/ptr"
)

func pvcFilters() *models.PvcFilters {
	return &models.PvcFilters{
		TenantId:   1,
		ClusterId:  ptr.To("c1"),
		ScopeQuery: &models.ScopeQueryParams{ScopeType: "project", ScopeId: "p1"},
	}
}

var _ = Describe("Pvcs", func() {
	Describe("ListPvcs", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM pvcs WHERE pvcs.tenant_id = ?`).Returns([]models.Pvc{})
			_, apiErr := ListPvcs(ctxWithTenant(1), &models.PvcFilters{TenantId: 1}, nil)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM pvcs WHERE pvcs.tenant_id = ?`).ReturnsError(errors.New("boom"))
			_, apiErr := ListPvcs(ctxWithTenant(1), &models.PvcFilters{TenantId: 1}, &models.ListOptions{})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("CreatePvc", func() {
		It("creates a new pvc", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM pvcs WHERE pvcs.tenant_id = ? AND pvcs.pvc_name = ? ORDER BY pvcs.id LIMIT ?`).Returns([]models.Pvc{})
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO pvcs (id,tenant_id,cluster_id,pvc_name,origin,scope_type,scope_id,phase,storage_class_name,requested_capacity,actual_capacity,data_sharing) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`).Returns(1)
			tx.Commit()
			_, isNew, apiErr := CreatePvc(ctxWithTenant(1), &models.Pvc{PvcFilters: models.PvcFilters{TenantId: 1, PvcName: ptr.To("p1")}})
			Expect(apiErr).To(BeNil())
			Expect(*isNew).To(BeTrue())
		})

		It("updates an existing pvc", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM pvcs WHERE pvcs.tenant_id = ? AND pvcs.pvc_name = ? ORDER BY pvcs.id LIMIT ?`).Returns([]models.Pvc{{ID: uuid.New(), PvcFilters: models.PvcFilters{TenantId: 1, PvcName: ptr.To("p1")}}})
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE pvcs SET tenant_id=?,cluster_id=?,pvc_name=?,origin=?,scope_type=?,scope_id=?,phase=?,storage_class_name=?,requested_capacity=?,actual_capacity=?,data_sharing=? WHERE id = ?`).Returns(1)
			tx.Commit()
			_, isNew, apiErr := CreatePvc(ctxWithTenant(1), &models.Pvc{PvcFilters: models.PvcFilters{TenantId: 1, PvcName: ptr.To("p1")}})
			Expect(apiErr).To(BeNil())
			Expect(*isNew).To(BeFalse())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM pvcs WHERE pvcs.tenant_id = ? AND pvcs.pvc_name = ? ORDER BY pvcs.id LIMIT ?`).ReturnsError(errors.New("boom"))
			_, _, apiErr := CreatePvc(ctxWithTenant(1), &models.Pvc{PvcFilters: models.PvcFilters{TenantId: 1, PvcName: ptr.To("p1")}})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeletePvc", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM pvcs WHERE pvcs.tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeletePvc(ctxWithTenant(1), &models.PvcFilters{TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM pvcs WHERE pvcs.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := DeletePvc(ctxWithTenant(1), &models.PvcFilters{TenantId: 1})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeletePvcsNotInCluster", func() {
		It("succeeds with ids", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM pvcs WHERE tenant_id = ? AND pvcs.tenant_id = ? AND NOT id in (?)`).Returns(1)
			tx.Commit()
			apiErr := DeletePvcsNotInCluster(ctxWithTenant(1), &models.PvcFilters{TenantId: 1}, []string{"id1"})
			Expect(apiErr).To(BeNil())
		})

		It("succeeds with no ids", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM pvcs WHERE tenant_id = ? AND pvcs.tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeletePvcsNotInCluster(ctxWithTenant(1), &models.PvcFilters{TenantId: 1}, []string{})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error when the tenant is missing", func() {
			setupDaoMock(GinkgoT())
			apiErr := DeletePvcsNotInCluster(context.Background(), &models.PvcFilters{TenantId: 1}, nil)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("ApplyPvcs", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			// upsertPvc: First lookup on db
			m.ExpectQuery(`SELECT * FROM pvcs WHERE pvcs.tenant_id = ? AND pvcs.cluster_id = ? AND pvcs.pvc_name = ? AND pvcs.origin = ? AND pvcs.scope_type = ? AND pvcs.scope_id = ? ORDER BY pvcs.id LIMIT ?`).Returns([]models.Pvc{})
			// Create (new pvc) in its own transaction
			innerTx := m.ExpectTransaction()
			innerTx.Begin()
			m.ExpectExec(`INSERT INTO pvcs (id,tenant_id,cluster_id,pvc_name,origin,scope_type,scope_id,phase,storage_class_name,requested_capacity,actual_capacity,data_sharing) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`).Returns(1)
			innerTx.Commit()
			// delete not-in
			m.ExpectExec(`DELETE FROM pvcs WHERE (pvcs.tenant_id = ? AND pvcs.cluster_id = ? AND pvcs.scope_type = ? AND pvcs.scope_id = ?) AND pvc_name not in (?)`).Returns(0)
			tx.Commit()
			_, _, apiErr := ApplyPvcs(ctxWithTenant(1), pvcFilters(), []api.Pvc{{Name: "p1", Phase: ptr.To(api.PVCPHASEOPTIONAL_BOUND)}})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT * FROM pvcs WHERE pvcs.tenant_id = ? AND pvcs.cluster_id = ? AND pvcs.pvc_name = ? AND pvcs.origin = ? AND pvcs.scope_type = ? AND pvcs.scope_id = ? ORDER BY pvcs.id LIMIT ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			_, _, apiErr := ApplyPvcs(ctxWithTenant(1), pvcFilters(), []api.Pvc{{Name: "p1"}})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteProjectPvcs", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM pvcs WHERE tenant_id = ? AND scope_type = ? AND scope_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteProjectPvcs(ctxWithTenant(1), 1, "p1")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteClusterPvcs", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM pvcs WHERE tenant_id = ? AND cluster_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteClusterPvcs(ctxWithTenant(1), 1, "c1")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteTenantPvcs", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM pvcs WHERE tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteTenantPvcs(ctxWithTenant(1), 1)
			Expect(apiErr).To(BeNil())
		})
	})
})
