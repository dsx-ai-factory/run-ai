package dao

import (
	"errors"

	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

type lockRow struct {
	TenantId int32
}

var _ = Describe("MappingLocking", func() {
	Describe("AcquireTenantMappingLock", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_locking WHERE tenant_id = ? AND locked_at <= ?`).Returns(0)
			tx.Commit()

			tx2 := m.ExpectTransaction()
			tx2.Begin()
			m.ExpectQuery(`INSERT INTO mapping_locking (locked_at,tenant_id) VALUES (?,?) RETURNING tenant_id`).
				Returns([]lockRow{{TenantId: 1}})
			tx2.Commit()

			release, apiErr := AcquireTenantMappingLock(ctxWithTenant(1))
			Expect(apiErr).To(BeNil())
			Expect(release).ToNot(BeNil())
		})

		It("handles stale delete error", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_locking WHERE tenant_id = ? AND locked_at <= ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			release, apiErr := AcquireTenantMappingLock(ctxWithTenant(1))
			Expect(apiErr).ToNot(BeNil())
			Expect(release).To(BeNil())
		})

		It("handles conflict", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_locking WHERE tenant_id = ? AND locked_at <= ?`).Returns(0)
			tx.Commit()

			tx2 := m.ExpectTransaction()
			tx2.Begin()
			m.ExpectQuery(`INSERT INTO mapping_locking (locked_at,tenant_id) VALUES (?,?) RETURNING tenant_id`).ReturnsError(errors.New("duplicate key value violates unique constraint"))
			tx2.Rollback()

			release, apiErr := AcquireTenantMappingLock(ctxWithTenant(1))
			Expect(apiErr).ToNot(BeNil())
			Expect(release).To(BeNil())
			Expect(int(apiErr.Code)).To(Equal(409))
		})
	})

	Describe("isUniqueViolation", func() {
		It("detects unique violations", func() {
			Expect(isUniqueViolation(gorm.ErrDuplicatedKey)).To(BeTrue())
			Expect(isUniqueViolation(errors.New("duplicate key value"))).To(BeTrue())
			Expect(isUniqueViolation(errors.New("unique constraint failed"))).To(BeTrue())
			Expect(isUniqueViolation(errors.New("some other error"))).To(BeFalse())
		})
	})

	Describe("AcquireTenantMappingLock", func() {
		It("handles create error", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_locking WHERE tenant_id = ? AND locked_at <= ?`).Returns(0)
			tx.Commit()

			tx2 := m.ExpectTransaction()
			tx2.Begin()
			m.ExpectQuery(`INSERT INTO mapping_locking (locked_at,tenant_id) VALUES (?,?) RETURNING tenant_id`).ReturnsError(errors.New("boom"))
			tx2.Rollback()

			release, apiErr := AcquireTenantMappingLock(ctxWithTenant(1))
			Expect(apiErr).ToNot(BeNil())
			Expect(release).To(BeNil())
			Expect(int(apiErr.Code)).ToNot(Equal(409))
		})

		It("handles release error", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_locking WHERE tenant_id = ? AND locked_at <= ?`).Returns(0)
			tx.Commit()

			tx2 := m.ExpectTransaction()
			tx2.Begin()
			m.ExpectQuery(`INSERT INTO mapping_locking (locked_at,tenant_id) VALUES (?,?) RETURNING tenant_id`).
				Returns([]lockRow{{TenantId: 1}})
			tx2.Commit()

			release, apiErr := AcquireTenantMappingLock(ctxWithTenant(1))
			Expect(apiErr).To(BeNil())
			Expect(release).ToNot(BeNil())

			tx3 := m.ExpectTransaction()
			tx3.Begin()
			m.ExpectExec(`DELETE FROM mapping_locking WHERE tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx3.Rollback()

			release()
		})
	})

	Describe("isUniqueViolation", func() {
		It("handles pq error", func() {
			Expect(isUniqueViolation(&pq.Error{Code: "23505"})).To(BeTrue())
			Expect(isUniqueViolation(&pq.Error{Code: "12345", Message: "other"})).To(BeFalse())
		})
	})
})
