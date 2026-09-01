package dao

import (
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-application-packages/dbmodels"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

var _ = Describe("EntitiesCache", func() {
	Describe("getCachedEntity", func() {
		It("returns the record", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).
				Returns([]dbmodel.EntitiesCache{
					{ID: uuid.New(), TenantId: 1, EntityType: dbmodel.EntityTypeBreakdown, EntityId: "p1", CachedInfo: dbmodels.JsonBMap{}},
				})

			row, apiErr := getCachedEntity(ctxWithTenant(1), dbmodel.EntityTypeBreakdown, "p1")
			Expect(apiErr).To(BeNil())
			Expect(row).ToNot(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).
				Returns([]dbmodel.EntitiesCache{})

			row, apiErr := getCachedEntity(ctxWithTenant(1), dbmodel.EntityTypeBreakdown, "p1")
			Expect(apiErr).To(BeNil())
			Expect(row).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).ReturnsError(errors.New("boom"))

			row, apiErr := getCachedEntity(ctxWithTenant(1), dbmodel.EntityTypeBreakdown, "p1")
			Expect(apiErr).ToNot(BeNil())
			Expect(row).To(BeNil())
		})
	})

	Describe("UpsertEntityCache", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			id := uuid.New()
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`INSERT INTO entities_cache (id,tenant_id,entity_type,entity_id,cached_info) VALUES (?,?,?,?,?) ON CONFLICT (tenant_id,entity_type,entity_id) DO UPDATE SET cached_info=excluded.cached_info RETURNING id`).Returns([]struct{ ID uuid.UUID }{{ID: id}})
			tx.Commit()

			row := &dbmodel.EntitiesCache{EntityType: dbmodel.EntityTypeBreakdown, EntityId: "p1"}
			apiErr := UpsertEntityCache(ctxWithTenant(1), runaidb.DB(), row)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`INSERT INTO entities_cache (id,tenant_id,entity_type,entity_id,cached_info) VALUES (?,?,?,?,?) ON CONFLICT (tenant_id,entity_type,entity_id) DO UPDATE SET cached_info=excluded.cached_info RETURNING id`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			row := &dbmodel.EntitiesCache{EntityType: dbmodel.EntityTypeBreakdown, EntityId: "p1"}
			apiErr := UpsertEntityCache(ctxWithTenant(1), runaidb.DB(), row)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("OnEntityDeleted", func() {
		It("handles no cache", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).
				Returns([]dbmodel.EntitiesCache{})

			apiErr := OnEntityDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.EntityTypeBreakdown, "p1")
			Expect(apiErr).To(BeNil())
		})

		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).
				Returns([]dbmodel.EntitiesCache{{ID: cacheId, TenantId: 1, EntityType: dbmodel.EntityTypeBreakdown, EntityId: "p1"}})

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.cache_id = ? AND entities_dependency.tenant_id = ?`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache WHERE entities_cache.id = ? AND entities_cache.tenant_id = ?`).Returns(1)
			txc.Commit()

			apiErr := OnEntityDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.EntityTypeBreakdown, "p1")
			Expect(apiErr).To(BeNil())
		})

		It("handles find error", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).ReturnsError(errors.New("boom"))

			apiErr := OnEntityDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.EntityTypeBreakdown, "p1")
			Expect(apiErr).ToNot(BeNil())
		})

		It("handles dep delete error", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).
				Returns([]dbmodel.EntitiesCache{{ID: cacheId, TenantId: 1, EntityType: dbmodel.EntityTypeBreakdown, EntityId: "p1"}})

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.cache_id = ? AND entities_dependency.tenant_id = ?`).ReturnsError(errors.New("boom"))
			txd.Rollback()

			apiErr := OnEntityDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.EntityTypeBreakdown, "p1")
			Expect(apiErr).ToNot(BeNil())
		})

		It("handles cache delete error", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).
				Returns([]dbmodel.EntitiesCache{{ID: cacheId, TenantId: 1, EntityType: dbmodel.EntityTypeBreakdown, EntityId: "p1"}})

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.cache_id = ? AND entities_dependency.tenant_id = ?`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache WHERE entities_cache.id = ? AND entities_cache.tenant_id = ?`).ReturnsError(errors.New("boom"))
			txc.Rollback()

			apiErr := OnEntityDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.EntityTypeBreakdown, "p1")
			Expect(apiErr).ToNot(BeNil())
		})
	})
})
