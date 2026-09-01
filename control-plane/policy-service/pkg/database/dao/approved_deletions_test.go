package dao

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

var _ = Describe("ApprovedDeletions", func() {
	Describe("ListApprovedDeletionsByTenant", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM approved_deletions WHERE approved_deletions.tenant_id = ?`).
				Returns([]dbmodel.ApprovedDeletions{
					{ApprovedDeletionsKey: dbmodel.ApprovedDeletionsKey{TenantId: 1, ClusterId: convutils.Ref("c1")}, AvoidDeletionsOfObsoletePolicies: true},
				})

			res, apiErr := ListApprovedDeletionsByTenant(ctxWithTenant(1))
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
			Expect(res["c1"].AvoidDeletionsOfObsoletePolicies).To(BeTrue())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM approved_deletions WHERE approved_deletions.tenant_id = ?`).ReturnsError(errors.New("boom"))

			res, apiErr := ListApprovedDeletionsByTenant(ctxWithTenant(1))
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("GetApprovedDeletionsInfoOfCluster", func() {
		It("returns the record", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM approved_deletions WHERE approved_deletions.tenant_id = ? AND approved_deletions.cluster_id = ? LIMIT ?`).
				Returns([]dbmodel.ApprovedDeletions{
					{ApprovedDeletionsKey: dbmodel.ApprovedDeletionsKey{TenantId: 1, ClusterId: convutils.Ref("c1")}, AvoidDeletionsOfObsoletePolicies: true},
				})

			res, apiErr := GetApprovedDeletionsInfoOfCluster(ctxWithTenant(1), "c1")
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
			Expect(res.AvoidDeletionsOfObsoletePolicies).To(BeTrue())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM approved_deletions WHERE approved_deletions.tenant_id = ? AND approved_deletions.cluster_id = ? LIMIT ?`).
				Returns([]dbmodel.ApprovedDeletions{})

			res, apiErr := GetApprovedDeletionsInfoOfCluster(ctxWithTenant(1), "c1")
			Expect(apiErr).To(BeNil())
			Expect(res).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM approved_deletions WHERE approved_deletions.tenant_id = ? AND approved_deletions.cluster_id = ? LIMIT ?`).ReturnsError(errors.New("boom"))

			res, apiErr := GetApprovedDeletionsInfoOfCluster(ctxWithTenant(1), "c1")
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("UpdateApprovedDeletionOfCluster", func() {
		It("updates the existing record", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM approved_deletions WHERE approved_deletions.tenant_id = ? AND approved_deletions.cluster_id = ? ORDER BY approved_deletions.tenant_id LIMIT ?`).
				Returns([]dbmodel.ApprovedDeletions{
					{ApprovedDeletionsKey: dbmodel.ApprovedDeletionsKey{TenantId: 1, ClusterId: convutils.Ref("c1")}, AvoidDeletionsOfObsoletePolicies: false},
				})

			txu := m.ExpectTransaction()
			txu.Begin()
			m.ExpectExec(`UPDATE approved_deletions SET avoid_deletions_of_obsolete_policies=? WHERE approved_deletions.tenant_id = ? AND approved_deletions.cluster_id = ?`).Returns(1)
			txu.Commit()

			res, apiErr := UpdateApprovedDeletionOfCluster(ctxWithTenant(1), api.SyncConfig{ClusterId: "c1", AvoidDeletionsOfObsoletePolicies: true})
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM approved_deletions WHERE approved_deletions.tenant_id = ? AND approved_deletions.cluster_id = ? ORDER BY approved_deletions.tenant_id LIMIT ?`).ReturnsError(errors.New("boom"))

			res, apiErr := UpdateApprovedDeletionOfCluster(ctxWithTenant(1), api.SyncConfig{ClusterId: "c1"})
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("UpdateApprovedDeletionByIds", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO approved_deletions (tenant_id,cluster_id,avoid_deletions_of_obsolete_policies) VALUES (?,?,?)`).Returns(1)
			tx.Commit()

			UpdateApprovedDeletionByIds(ctxWithTenant(1), []string{"c1"}, true)
		})
	})

	Describe("DeleteApprovedDeletionByIds", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM approved_deletions WHERE approved_deletions.tenant_id = ? AND cluster_id in (?)`).Returns(1)
			tx.Commit()

			DeleteApprovedDeletionByIds(ctxWithTenant(1), []string{"c1"})
		})
	})

	Describe("UpdateApprovedDeletionByIds", func() {
		It("handles a non-conflict error", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO approved_deletions (tenant_id,cluster_id,avoid_deletions_of_obsolete_policies) VALUES (?,?,?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			UpdateApprovedDeletionByIds(ctxWithTenant(1), []string{"c1"}, true)
		})

		It("returns a conflict error", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO approved_deletions (tenant_id,cluster_id,avoid_deletions_of_obsolete_policies) VALUES (?,?,?)`).ReturnsError(errors.New("duplicate key value violates unique constraint"))
			tx.Rollback()

			UpdateApprovedDeletionByIds(ctxWithTenant(1), []string{"c1"}, true)
		})
	})

	Describe("DeleteApprovedDeletionByIds", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM approved_deletions WHERE approved_deletions.tenant_id = ? AND cluster_id in (?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			DeleteApprovedDeletionByIds(ctxWithTenant(1), []string{"c1"})
		})
	})
})
