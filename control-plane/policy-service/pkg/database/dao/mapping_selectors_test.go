package dao

import (
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

func gvk() api.Gvk {
	return api.Gvk{Group: convutils.Ref("g"), Kind: "K", Version: "v1"}
}

var _ = Describe("MappingSelectors", func() {
	Describe("CreateMapping", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO mapping_selectors (id,gvk_group,gvk_kind,gvk_version,description,scope_type,tenant_id,spec_selectors,created_by,created_at,updated_by,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`).Returns(1)
			tx.Commit()

			apiErr := CreateMapping(ctxWithTenant(1), &dbmodel.MappingSelectors{ID: uuid.New(), GvkKind: "K", GvkVersion: "v1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO mapping_selectors (id,gvk_group,gvk_kind,gvk_version,description,scope_type,tenant_id,spec_selectors,created_by,created_at,updated_by,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := CreateMapping(ctxWithTenant(1), &dbmodel.MappingSelectors{ID: uuid.New(), GvkKind: "K", GvkVersion: "v1"})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("ListMappings", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_selectors WHERE ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?))`).
				Returns([]dbmodel.MappingSelectors{{ID: uuid.New(), GvkGroup: "g", GvkKind: "K", GvkVersion: "v1", ScopeType: api.MAPPINGSCOPE_TENANT}})

			res, apiErr := ListMappings(ctxWithTenant(1), nil)
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_selectors WHERE ((scope_type = ?) OR (scope_type = ? AND tenant_id = ?))`).ReturnsError(errors.New("boom"))

			res, apiErr := ListMappings(ctxWithTenant(1), nil)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("ListMappingsReferencingFieldGroup", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_selectors WHERE EXISTS ( SELECT 1 FROM jsonb_each(spec_selectors) AS e(k, v) WHERE EXISTS ( SELECT 1 FROM jsonb_array_elements_text(COALESCE(v->'fieldGroups', '[]'::jsonb)) AS elem WHERE elem = ? ) ) AND (((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)))`).
				Returns([]dbmodel.MappingSelectors{{ID: uuid.New(), GvkKind: "K", GvkVersion: "v1"}})

			res, apiErr := ListMappingsReferencingFieldGroup(ctxWithTenant(1), runaidb.DB(), "g1")
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})
	})

	Describe("GetMappingByGvk", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_selectors WHERE (mapping_selectors.gvk_group = ? AND mapping_selectors.gvk_kind = ? AND mapping_selectors.gvk_version = ?) AND (((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)))`).
				Returns([]dbmodel.MappingSelectors{{ID: uuid.New(), GvkGroup: "g", GvkKind: "K", GvkVersion: "v1", ScopeType: api.MAPPINGSCOPE_TENANT}})

			res, apiErr := GetMappingByGvk(ctxWithTenant(1), gvk())
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})
	})

	Describe("GetMappingByGvkAndScope", func() {
		It("finds the mapping", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_selectors WHERE mapping_selectors.gvk_group = ? AND mapping_selectors.gvk_kind = ? AND mapping_selectors.gvk_version = ? AND mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ? ORDER BY mapping_selectors.id LIMIT ?`).
				Returns([]dbmodel.MappingSelectors{{ID: uuid.New(), GvkGroup: "g", GvkKind: "K", GvkVersion: "v1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1}})

			res, apiErr := GetMappingByGvkAndScope(ctxWithTenant(1), gvk(), api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_selectors WHERE mapping_selectors.gvk_group = ? AND mapping_selectors.gvk_kind = ? AND mapping_selectors.gvk_version = ? AND mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ? ORDER BY mapping_selectors.id LIMIT ?`).
				Returns([]dbmodel.MappingSelectors{})

			res, apiErr := GetMappingByGvkAndScope(ctxWithTenant(1), gvk(), api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
			Expect(int(apiErr.Code)).To(Equal(404))
		})
	})

	Describe("SaveMapping", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE mapping_selectors SET id=?,gvk_group=?,gvk_kind=?,gvk_version=?,description=?,scope_type=?,tenant_id=?,spec_selectors=?,created_by=?,created_at=?,updated_by=?,updated_at=? WHERE (mapping_selectors.id = ? AND mapping_selectors.gvk_group = ? AND mapping_selectors.gvk_kind = ? AND mapping_selectors.gvk_version = ? AND mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ?) AND id = ?`).Returns(1)
			tx.Commit()

			apiErr := SaveMapping(ctxWithTenant(1), runaidb.DB(), &dbmodel.MappingSelectors{ID: uuid.New(), GvkGroup: "g", GvkKind: "K", GvkVersion: "v1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE mapping_selectors SET id=?,gvk_group=?,gvk_kind=?,gvk_version=?,description=?,scope_type=?,tenant_id=?,spec_selectors=?,created_by=?,created_at=?,updated_by=?,updated_at=? WHERE (mapping_selectors.id = ? AND mapping_selectors.gvk_group = ? AND mapping_selectors.gvk_kind = ? AND mapping_selectors.gvk_version = ? AND mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ?) AND id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := SaveMapping(ctxWithTenant(1), runaidb.DB(), &dbmodel.MappingSelectors{ID: uuid.New(), GvkGroup: "g", GvkKind: "K", GvkVersion: "v1", ScopeType: api.MAPPINGSCOPE_TENANT, TenantId: 1})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteMappingByGvkAndScope", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_selectors WHERE mapping_selectors.gvk_group = ? AND mapping_selectors.gvk_kind = ? AND mapping_selectors.gvk_version = ? AND mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ?`).Returns(1)
			tx.Commit()

			apiErr := DeleteMappingByGvkAndScope(ctxWithTenant(1), runaidb.DB(), gvk(), api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).To(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_selectors WHERE mapping_selectors.gvk_group = ? AND mapping_selectors.gvk_kind = ? AND mapping_selectors.gvk_version = ? AND mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ?`).Returns(0)
			tx.Commit()

			apiErr := DeleteMappingByGvkAndScope(ctxWithTenant(1), runaidb.DB(), gvk(), api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).ToNot(BeNil())
			Expect(int(apiErr.Code)).To(Equal(404))
		})
	})

	Describe("GvkFor", func() {
		It("returns the gvk for a key", func() {
			key := makeGvkKey("g", "K", "v1")
			m := MappingsByGvkAndScope{
				key: MappingByScope{
					api.MAPPINGSCOPE_TENANT: dbmodel.MappingSelectors{GvkGroup: "g", GvkKind: "K", GvkVersion: "v1"},
				},
			}
			got := m.GvkFor(key)
			Expect(got.Kind).To(Equal("K"))
			Expect(got.Version).To(Equal("v1"))

			empty := m.GvkFor(makeGvkKey("x", "y", "z"))
			Expect(empty).To(Equal(api.Gvk{}))
		})
	})

	Describe("ListMappingsReferencingFieldGroup", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_selectors WHERE EXISTS ( SELECT 1 FROM jsonb_each(spec_selectors) AS e(k, v) WHERE EXISTS ( SELECT 1 FROM jsonb_array_elements_text(COALESCE(v->'fieldGroups', '[]'::jsonb)) AS elem WHERE elem = ? ) ) AND (((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)))`).ReturnsError(errors.New("boom"))

			res, apiErr := ListMappingsReferencingFieldGroup(ctxWithTenant(1), runaidb.DB(), "g1")
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("GetMappingByGvk", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_selectors WHERE (mapping_selectors.gvk_group = ? AND mapping_selectors.gvk_kind = ? AND mapping_selectors.gvk_version = ?) AND (((scope_type = ?) OR (scope_type = ? AND tenant_id = ?)))`).ReturnsError(errors.New("boom"))

			res, apiErr := GetMappingByGvk(ctxWithTenant(1), gvk())
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("GetMappingByGvkAndScope", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM mapping_selectors WHERE mapping_selectors.gvk_group = ? AND mapping_selectors.gvk_kind = ? AND mapping_selectors.gvk_version = ? AND mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ? ORDER BY mapping_selectors.id LIMIT ?`).ReturnsError(errors.New("boom"))

			res, apiErr := GetMappingByGvkAndScope(ctxWithTenant(1), gvk(), api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("DeleteMappingByGvkAndScope", func() {
		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM mapping_selectors WHERE mapping_selectors.gvk_group = ? AND mapping_selectors.gvk_kind = ? AND mapping_selectors.gvk_version = ? AND mapping_selectors.scope_type = ? AND mapping_selectors.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := DeleteMappingByGvkAndScope(ctxWithTenant(1), runaidb.DB(), gvk(), api.MAPPINGSCOPE_TENANT)
			Expect(apiErr).ToNot(BeNil())
		})
	})
})
