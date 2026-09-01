package dao

import (
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

var _ = Describe("MappingFieldGroups", func() {
	Describe("CreateFieldGroup", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO mapping_field_groups (id,name,description,scope_type,tenant_id,fields,item_selectors,created_by,created_at,updated_by,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`).Returns(1)
			tx.Commit()

			apiErr := CreateFieldGroup(ctxWithTenant(1), runaidb.DB(), &dbmodel.MappingFieldGroups{ID: uuid.New(), Name: "g1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO mapping_field_groups (id,name,description,scope_type,tenant_id,fields,item_selectors,created_by,created_at,updated_by,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := CreateFieldGroup(ctxWithTenant(1), runaidb.DB(), &dbmodel.MappingFieldGroups{ID: uuid.New(), Name: "g1"})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("GetFieldGroupByName", func() {
		It("returns the matching field groups", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_field_groups WHERE mapping_field_groups.name = ? AND (((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)))`).
				Returns([]dbmodel.MappingFieldGroups{{ID: uuid.New(), Name: "g1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1}})

			res, apiErr := GetFieldGroupByName(ctxWithTenant(1), runaidb.DB(), "g1")
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_field_groups WHERE mapping_field_groups.name = ? AND (((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)))`).
				Returns([]dbmodel.MappingFieldGroups{})

			res, apiErr := GetFieldGroupByName(ctxWithTenant(1), runaidb.DB(), "g1")
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
			Expect(int(apiErr.Code)).To(Equal(404))
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_field_groups WHERE mapping_field_groups.name = ? AND (((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)))`).ReturnsError(errors.New("boom"))

			res, apiErr := GetFieldGroupByName(ctxWithTenant(1), runaidb.DB(), "g1")
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("GetFieldGroupByNameAndScope", func() {
		It("finds a tenant-scoped field group", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_field_groups WHERE mapping_field_groups.name = ? AND mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ? LIMIT ?`).
				Returns([]dbmodel.MappingFieldGroups{{ID: uuid.New(), Name: "g1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1}})

			res, apiErr := GetFieldGroupByNameAndScope(ctxWithTenant(1), runaidb.DB(), "g1", api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns not found for a system scope", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_field_groups WHERE mapping_field_groups.name = ? AND mapping_field_groups.scope_type = ? LIMIT ?`).
				Returns([]dbmodel.MappingFieldGroups{})

			res, apiErr := GetFieldGroupByNameAndScope(ctxWithTenant(1), runaidb.DB(), "g1", api.MAPPINGSCOPE_SYSTEM)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
			Expect(int(apiErr.Code)).To(Equal(404))
		})
	})

	Describe("ListFieldGroupNamesPage", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT DISTINCT name FROM mapping_field_groups WHERE ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)) ORDER BY name LIMIT ? OFFSET ?`).
				Returns([]struct{ Name string }{{Name: "g1"}})

			res, apiErr := ListFieldGroupNamesPage(ctxWithTenant(1), runaidb.DB(), 1, 3)
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})
	})

	Describe("GetFieldGroupsByNames", func() {
		It("returns empty for no names", func() {
			setupDaoMock(GinkgoT())
			res, apiErr := GetFieldGroupsByNames(ctxWithTenant(1), runaidb.DB(), nil)
			Expect(apiErr).To(BeNil())
			Expect(res).To(BeEmpty())
		})

		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_field_groups WHERE name IN (?) AND (((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)))`).
				Returns([]dbmodel.MappingFieldGroups{{ID: uuid.New(), Name: "g1", ScopeType: api.MAPPINGSCOPE_TENANT}})

			res, apiErr := GetFieldGroupsByNames(ctxWithTenant(1), runaidb.DB(), []string{"g1"})
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})
	})

	Describe("SaveFieldGroup", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE mapping_field_groups SET id=?,name=?,description=?,scope_type=?,tenant_id=?,fields=?,item_selectors=?,created_by=?,created_at=?,updated_by=?,updated_at=? WHERE (mapping_field_groups.id = ? AND mapping_field_groups.name = ? AND mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ?) AND id = ?`).Returns(1)
			tx.Commit()

			apiErr := SaveFieldGroup(ctxWithTenant(1), runaidb.DB(), &dbmodel.MappingFieldGroups{ID: uuid.New(), Name: "g1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE mapping_field_groups SET id=?,name=?,description=?,scope_type=?,tenant_id=?,fields=?,item_selectors=?,created_by=?,created_at=?,updated_by=?,updated_at=? WHERE (mapping_field_groups.id = ? AND mapping_field_groups.name = ? AND mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ?) AND id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := SaveFieldGroup(ctxWithTenant(1), runaidb.DB(), &dbmodel.MappingFieldGroups{ID: uuid.New(), Name: "g1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteFieldGroupByNameAndScope", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_field_groups WHERE mapping_field_groups.name = ? AND mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ?`).Returns(1)
			tx.Commit()

			apiErr := DeleteFieldGroupByNameAndScope(ctxWithTenant(1), runaidb.DB(), "g1", api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).To(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_field_groups WHERE mapping_field_groups.name = ? AND mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ?`).Returns(0)
			tx.Commit()

			apiErr := DeleteFieldGroupByNameAndScope(ctxWithTenant(1), runaidb.DB(), "g1", api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).ToNot(BeNil())
			Expect(int(apiErr.Code)).To(Equal(404))
		})
	})

	Describe("CheckIfFormatterBeingUsed", func() {
		It("returns true when the formatter is used", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT COUNT(*) FROM mapping_field_groups mfg WHERE ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)) AND EXISTS ( SELECT 1 FROM jsonb_each(mfg.fields) AS e(k, v) WHERE (v->>'formatter') = ? )`).
				Returns([]countRow{{Count: 2}})

			used, apiErr := CheckIfFormatterBeingUsed(ctxWithTenant(1), "fmt1")
			Expect(apiErr).To(BeNil())
			Expect(used).To(BeTrue())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT COUNT(*) FROM mapping_field_groups mfg WHERE ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)) AND EXISTS ( SELECT 1 FROM jsonb_each(mfg.fields) AS e(k, v) WHERE (v->>'formatter') = ? )`).
				ReturnsError(errors.New("boom"))

			used, apiErr := CheckIfFormatterBeingUsed(ctxWithTenant(1), "fmt1")
			Expect(apiErr).ToNot(BeNil())
			Expect(used).To(BeFalse())
		})
	})

	Describe("GetFieldGroupByNameAndScope", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_field_groups WHERE mapping_field_groups.name = ? AND mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ? LIMIT ?`).ReturnsError(errors.New("boom"))

			res, apiErr := GetFieldGroupByNameAndScope(ctxWithTenant(1), runaidb.DB(), "g1", api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("ListFieldGroupNamesPage", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT DISTINCT name FROM mapping_field_groups WHERE ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)) ORDER BY name LIMIT ? OFFSET ?`).ReturnsError(errors.New("boom"))

			res, apiErr := ListFieldGroupNamesPage(ctxWithTenant(1), runaidb.DB(), 1, 3)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("GetFieldGroupsByNames", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_field_groups WHERE name IN (?) AND (((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)))`).ReturnsError(errors.New("boom"))

			res, apiErr := GetFieldGroupsByNames(ctxWithTenant(1), runaidb.DB(), []string{"g1"})
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("DeleteFieldGroupByNameAndScope", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_field_groups WHERE mapping_field_groups.name = ? AND mapping_field_groups.scope_type = ? AND mapping_field_groups.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := DeleteFieldGroupByNameAndScope(ctxWithTenant(1), runaidb.DB(), "g1", api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).ToNot(BeNil())
		})
	})
})
