package dao

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MappingFormatters", func() {
	Describe("CreateFormatter", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO mapping_formatters (id,name,description,scope_type,tenant_id,fields,created_by,created_at,updated_by,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`).Returns(1)
			tx.Commit()

			apiErr := CreateFormatter(ctxWithTenant(1), &dbmodel.MappingFormatters{ID: uuid.New(), Name: "f1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO mapping_formatters (id,name,description,scope_type,tenant_id,fields,created_by,created_at,updated_by,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := CreateFormatter(ctxWithTenant(1), &dbmodel.MappingFormatters{ID: uuid.New(), Name: "f1"})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("GetFormatterByName", func() {
		It("finds the formatter", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM mapping_formatters WHERE mapping_formatters.name = ? AND ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)) ORDER BY mapping_formatters.id LIMIT ?`).
				Returns([]dbmodel.MappingFormatters{{ID: uuid.New(), Name: "f1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1}})

			res, apiErr := GetFormatterByName(ctxWithTenant(1), runaidb.DB(), "f1")
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM mapping_formatters WHERE mapping_formatters.name = ? AND ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)) ORDER BY mapping_formatters.id LIMIT ?`).
				Returns([]dbmodel.MappingFormatters{})

			res, apiErr := GetFormatterByName(ctxWithTenant(1), runaidb.DB(), "f1")
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
			Expect(int(apiErr.Code)).To(Equal(404))
		})
	})

	Describe("GetFormattersByNames", func() {
		It("returns empty for no names", func() {
			setupDaoMock(GinkgoT())
			res, apiErr := GetFormattersByNames(ctxWithTenant(1), nil)
			Expect(apiErr).To(BeNil())
			Expect(res).To(BeEmpty())
		})

		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM mapping_formatters WHERE name IN (?) AND ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?))`).
				Returns([]dbmodel.MappingFormatters{{ID: uuid.New(), Name: "f1"}})

			res, apiErr := GetFormattersByNames(ctxWithTenant(1), []string{"f1"})
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})
	})

	Describe("ListFormatters", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM mapping_formatters WHERE (scope_type = ?) OR (scope_type = ? AND tenant_id = ?) ORDER BY name LIMIT ?`).
				Returns([]dbmodel.MappingFormatters{{ID: uuid.New(), Name: "f1"}})

			res, apiErr := ListFormatters(ctxWithTenant(1), 0, 10)
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM mapping_formatters WHERE (scope_type = ?) OR (scope_type = ? AND tenant_id = ?) ORDER BY name LIMIT ?`).ReturnsError(errors.New("boom"))

			res, apiErr := ListFormatters(ctxWithTenant(1), 0, 10)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("SaveFormatter", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE mapping_formatters SET id=?,name=?,description=?,scope_type=?,tenant_id=?,fields=?,created_by=?,created_at=?,updated_by=?,updated_at=? WHERE (mapping_formatters.id = ? AND mapping_formatters.name = ? AND mapping_formatters.scope_type = ? AND mapping_formatters.tenant_id = ?) AND id = ?`).Returns(1)
			tx.Commit()

			apiErr := SaveFormatter(ctxWithTenant(1), runaidb.DB(), &dbmodel.MappingFormatters{ID: uuid.New(), Name: "f1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("inserts when the update matches nothing", func() {
			m := setupDaoMock(GinkgoT())

			txu := m.ExpectTransaction()
			txu.Begin()
			m.ExpectExec(`UPDATE mapping_formatters SET id=?,name=?,description=?,scope_type=?,tenant_id=?,fields=?,created_by=?,created_at=?,updated_by=?,updated_at=? WHERE (mapping_formatters.id = ? AND mapping_formatters.name = ? AND mapping_formatters.scope_type = ? AND mapping_formatters.tenant_id = ?) AND id = ?`).Returns(0)
			txu.Commit()

			txi := m.ExpectTransaction()
			txi.Begin()
			m.ExpectExec(`INSERT INTO mapping_formatters (id,name,description,scope_type,tenant_id,fields,created_by,created_at,updated_by,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?) ON CONFLICT (id) DO UPDATE SET updated_at=?,name=excluded.name,description=excluded.description,scope_type=excluded.scope_type,tenant_id=excluded.tenant_id,fields=excluded.fields,created_by=excluded.created_by,updated_by=excluded.updated_by`).Returns(1)
			txi.Commit()

			apiErr := SaveFormatter(ctxWithTenant(1), runaidb.DB(), &dbmodel.MappingFormatters{ID: uuid.New(), Name: "f1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1})
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteFormatterByName", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_formatters WHERE mapping_formatters.name = ? AND ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?))`).Returns(1)
			tx.Commit()

			apiErr := DeleteFormatterByName(ctxWithTenant(1), runaidb.DB(), "f1")
			Expect(apiErr).To(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_formatters WHERE mapping_formatters.name = ? AND ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?))`).Returns(0)
			tx.Commit()

			apiErr := DeleteFormatterByName(ctxWithTenant(1), runaidb.DB(), "f1")
			Expect(apiErr).ToNot(BeNil())
			Expect(int(apiErr.Code)).To(Equal(404))
		})
	})

	Describe("GetFormatterByName", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_formatters WHERE mapping_formatters.name = ? AND ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)) ORDER BY mapping_formatters.id LIMIT ?`).ReturnsError(errors.New("boom"))

			res, apiErr := GetFormatterByName(ctxWithTenant(1), runaidb.DB(), "f1")
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("GetFormattersByNames", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_formatters WHERE name IN (?) AND ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?))`).ReturnsError(errors.New("boom"))

			res, apiErr := GetFormattersByNames(ctxWithTenant(1), []string{"f1"})
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("SaveFormatter", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE mapping_formatters SET id=?,name=?,description=?,scope_type=?,tenant_id=?,fields=?,created_by=?,created_at=?,updated_by=?,updated_at=? WHERE (mapping_formatters.id = ? AND mapping_formatters.name = ? AND mapping_formatters.scope_type = ? AND mapping_formatters.tenant_id = ?) AND id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := SaveFormatter(ctxWithTenant(1), runaidb.DB(), &dbmodel.MappingFormatters{ID: uuid.New(), Name: "f1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteFormatterByName", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_formatters WHERE mapping_formatters.name = ? AND ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?))`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := DeleteFormatterByName(ctxWithTenant(1), runaidb.DB(), "f1")
			Expect(apiErr).ToNot(BeNil())
		})
	})
})
