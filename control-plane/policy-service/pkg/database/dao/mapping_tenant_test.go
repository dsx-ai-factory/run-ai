package dao

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

var _ = Describe("MappingTenant", func() {
	Describe("DeleteMappingsOfTenant", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			tx1 := m.ExpectTransaction()
			tx1.Begin()
			m.ExpectExec(`DELETE FROM mapping_selectors WHERE mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ?`).Returns(1)
			tx1.Commit()

			tx2 := m.ExpectTransaction()
			tx2.Begin()
			m.ExpectExec(`DELETE FROM mapping_field_groups WHERE mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ?`).Returns(1)
			tx2.Commit()

			tx3 := m.ExpectTransaction()
			tx3.Begin()
			m.ExpectExec(`DELETE FROM mapping_formatters WHERE mapping_formatters.scope_type = ? AND mapping_formatters.tenant_id = ?`).Returns(1)
			tx3.Commit()

			tx4 := m.ExpectTransaction()
			tx4.Begin()
			m.ExpectExec(`DELETE FROM mapping_locking WHERE mapping_locking.tenant_id = ?`).Returns(1)
			tx4.Commit()

			apiErr := DeleteMappingsOfTenant(ctxWithTenant(1), runaidb.DB(), 1)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			tx1 := m.ExpectTransaction()
			tx1.Begin()
			m.ExpectExec(`DELETE FROM mapping_selectors WHERE mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx1.Rollback()

			apiErr := DeleteMappingsOfTenant(ctxWithTenant(1), runaidb.DB(), 1)
			Expect(apiErr).ToNot(BeNil())
		})

		It("handles field groups error", func() {
			m := setupDaoMock(GinkgoT())

			tx1 := m.ExpectTransaction()
			tx1.Begin()
			m.ExpectExec(`DELETE FROM mapping_selectors WHERE mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ?`).Returns(1)
			tx1.Commit()

			tx2 := m.ExpectTransaction()
			tx2.Begin()
			m.ExpectExec(`DELETE FROM mapping_field_groups WHERE mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx2.Rollback()

			apiErr := DeleteMappingsOfTenant(ctxWithTenant(1), runaidb.DB(), 1)
			Expect(apiErr).ToNot(BeNil())
		})

		It("handles formatters error", func() {
			m := setupDaoMock(GinkgoT())

			tx1 := m.ExpectTransaction()
			tx1.Begin()
			m.ExpectExec(`DELETE FROM mapping_selectors WHERE mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ?`).Returns(1)
			tx1.Commit()

			tx2 := m.ExpectTransaction()
			tx2.Begin()
			m.ExpectExec(`DELETE FROM mapping_field_groups WHERE mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ?`).Returns(1)
			tx2.Commit()

			tx3 := m.ExpectTransaction()
			tx3.Begin()
			m.ExpectExec(`DELETE FROM mapping_formatters WHERE mapping_formatters.scope_type = ? AND mapping_formatters.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx3.Rollback()

			apiErr := DeleteMappingsOfTenant(ctxWithTenant(1), runaidb.DB(), 1)
			Expect(apiErr).ToNot(BeNil())
		})

		It("handles locking error", func() {
			m := setupDaoMock(GinkgoT())

			tx1 := m.ExpectTransaction()
			tx1.Begin()
			m.ExpectExec(`DELETE FROM mapping_selectors WHERE mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ?`).Returns(1)
			tx1.Commit()

			tx2 := m.ExpectTransaction()
			tx2.Begin()
			m.ExpectExec(`DELETE FROM mapping_field_groups WHERE mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ?`).Returns(1)
			tx2.Commit()

			tx3 := m.ExpectTransaction()
			tx3.Begin()
			m.ExpectExec(`DELETE FROM mapping_formatters WHERE mapping_formatters.scope_type = ? AND mapping_formatters.tenant_id = ?`).Returns(1)
			tx3.Commit()

			tx4 := m.ExpectTransaction()
			tx4.Begin()
			m.ExpectExec(`DELETE FROM mapping_locking WHERE mapping_locking.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx4.Rollback()

			apiErr := DeleteMappingsOfTenant(ctxWithTenant(1), runaidb.DB(), 1)
			Expect(apiErr).ToNot(BeNil())
		})
	})
})
