package dao

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("EntitiesBreakdownCache", func() {
	Describe("WipeAllEntityCaches", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache`).Returns(1)
			txc.Commit()

			apiErr := WipeAllEntityCaches(ctxWithTenant(1))
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency`).ReturnsError(errors.New("boom"))
			txd.Rollback()

			apiErr := WipeAllEntityCaches(ctxWithTenant(1))
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteEntityCachesOfTenant", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.tenant_id = ?`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache WHERE entities_cache.tenant_id = ?`).Returns(1)
			txc.Commit()

			apiErr := DeleteEntityCachesOfTenant(ctxWithTenant(1), runaidb.DB(), 1)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.tenant_id = ?`).ReturnsError(errors.New("boom"))
			txd.Rollback()

			apiErr := DeleteEntityCachesOfTenant(ctxWithTenant(1), runaidb.DB(), 1)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("GetCachedBreakdown", func() {
		It("returns the record", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).
				Returns([]dbmodel.EntitiesCache{
					{ID: uuid.New(), TenantId: 1, EntityType: dbmodel.EntityTypeBreakdown, EntityId: "p1", CachedInfo: dbmodels.JsonBMap{}},
				})

			res, apiErr := GetCachedBreakdown(ctxWithTenant(1), "p1")
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).
				Returns([]dbmodel.EntitiesCache{})

			res, apiErr := GetCachedBreakdown(ctxWithTenant(1), "p1")
			Expect(apiErr).To(BeNil())
			Expect(res).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).ReturnsError(errors.New("boom"))

			res, apiErr := GetCachedBreakdown(ctxWithTenant(1), "p1")
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("CacheBreakdown", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectQuery(`INSERT INTO entities_cache (id,tenant_id,entity_type,entity_id,cached_info) VALUES (?,?,?,?,?) ON CONFLICT (tenant_id,entity_type,entity_id) DO UPDATE SET cached_info=excluded.cached_info RETURNING id`).
				Returns([]struct{ ID uuid.UUID }{{ID: cacheId}})
			txc.Commit()

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`INSERT INTO entities_dependency (cache_id,tenant_id,component_type,component_id) VALUES (?,?,?,?) ON CONFLICT DO NOTHING`).Returns(1)
			txd.Commit()

			deps := []dbmodel.EntitiesDependency{
				{ComponentType: dbmodel.ComponentTypeMapping, ComponentId: "m1"},
			}
			id, apiErr := CacheBreakdown(ctxWithTenant(1), runaidb.DB(), "p1", &api.BreakdownBreakdown{}, deps)
			Expect(apiErr).To(BeNil())
			Expect(id).To(Equal(cacheId))
		})

		It("returns an error on cache upsert failure", func() {
			m := setupDaoMock(GinkgoT())

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectQuery(`INSERT INTO entities_cache (id,tenant_id,entity_type,entity_id,cached_info) VALUES (?,?,?,?,?) ON CONFLICT (tenant_id,entity_type,entity_id) DO UPDATE SET cached_info=excluded.cached_info RETURNING id`).
				ReturnsError(errors.New("boom"))
			txc.Rollback()

			id, apiErr := CacheBreakdown(ctxWithTenant(1), runaidb.DB(), "p1", &api.BreakdownBreakdown{}, nil)
			Expect(apiErr).ToNot(BeNil())
			Expect(id).To(Equal(uuid.Nil))
		})
	})

	Describe("DeleteBreakdownCachesOfSubtree", func() {
		It("returns nil for empty input", func() {
			setupDaoMock(GinkgoT())
			apiErr := DeleteBreakdownCachesOfSubtree(ctxWithTenant(1), runaidb.DB(), nil)
			Expect(apiErr).To(BeNil())
		})

		It("returns nil when there are no caches", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND entity_id IN (?)`).
				Returns([]struct{ ID uuid.UUID }{})

			apiErr := DeleteBreakdownCachesOfSubtree(ctxWithTenant(1), runaidb.DB(), []string{"p1"})
			Expect(apiErr).To(BeNil())
		})

		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND entity_id IN (?)`).
				Returns([]struct{ ID uuid.UUID }{{ID: cacheId}})

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND cache_id IN (?)`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache WHERE entities_cache.tenant_id = ? AND id IN (?)`).Returns(1)
			txc.Commit()

			apiErr := DeleteBreakdownCachesOfSubtree(ctxWithTenant(1), runaidb.DB(), []string{"p1"})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on pluck failure", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND entity_id IN (?)`).ReturnsError(errors.New("boom"))

			apiErr := DeleteBreakdownCachesOfSubtree(ctxWithTenant(1), runaidb.DB(), []string{"p1"})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("WipeAllEntityCaches", func() {
		It("returns an error on cache delete failure", func() {
			m := setupDaoMock(GinkgoT())

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache`).ReturnsError(errors.New("boom"))
			txc.Rollback()

			apiErr := WipeAllEntityCaches(ctxWithTenant(1))
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteEntityCachesOfTenant", func() {
		It("returns an error on cache delete failure", func() {
			m := setupDaoMock(GinkgoT())

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.tenant_id = ?`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache WHERE entities_cache.tenant_id = ?`).ReturnsError(errors.New("boom"))
			txc.Rollback()

			apiErr := DeleteEntityCachesOfTenant(ctxWithTenant(1), runaidb.DB(), 1)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("GetCachedBreakdown", func() {
		It("returns nil on corrupt data", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM entities_cache WHERE entities_cache.tenant_id = ? AND entities_cache.entity_type = ? AND entities_cache.entity_id = ? ORDER BY entities_cache.id LIMIT ?`).
				Returns([]dbmodel.EntitiesCache{
					{ID: uuid.New(), TenantId: 1, EntityType: dbmodel.EntityTypeBreakdown, EntityId: "p1", CachedInfo: dbmodels.JsonBMap{"fields": "not-a-map"}},
				})

			res, apiErr := GetCachedBreakdown(ctxWithTenant(1), "p1")
			Expect(apiErr).To(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("CacheBreakdown", func() {
		It("returns an error on dependency upsert failure", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectQuery(`INSERT INTO entities_cache (id,tenant_id,entity_type,entity_id,cached_info) VALUES (?,?,?,?,?) ON CONFLICT (tenant_id,entity_type,entity_id) DO UPDATE SET cached_info=excluded.cached_info RETURNING id`).
				Returns([]struct{ ID uuid.UUID }{{ID: cacheId}})
			txc.Commit()

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`INSERT INTO entities_dependency (cache_id,tenant_id,component_type,component_id) VALUES (?,?,?,?) ON CONFLICT DO NOTHING`).ReturnsError(errors.New("boom"))
			txd.Rollback()

			deps := []dbmodel.EntitiesDependency{
				{ComponentType: dbmodel.ComponentTypeMapping, ComponentId: "m1"},
			}
			id, apiErr := CacheBreakdown(ctxWithTenant(1), runaidb.DB(), "p1", &api.BreakdownBreakdown{}, deps)
			Expect(apiErr).ToNot(BeNil())
			Expect(id).To(Equal(uuid.Nil))
		})
	})

	Describe("DeleteBreakdownCachesOfSubtree", func() {
		It("returns an error on dependency delete failure", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND entity_id IN (?)`).
				Returns([]struct{ ID uuid.UUID }{{ID: cacheId}})

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND cache_id IN (?)`).ReturnsError(errors.New("boom"))
			txd.Rollback()

			apiErr := DeleteBreakdownCachesOfSubtree(ctxWithTenant(1), runaidb.DB(), []string{"p1"})
			Expect(apiErr).ToNot(BeNil())
		})

		It("returns an error on cache delete failure", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND entity_id IN (?)`).
				Returns([]struct{ ID uuid.UUID }{{ID: cacheId}})

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND cache_id IN (?)`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache WHERE entities_cache.tenant_id = ? AND id IN (?)`).ReturnsError(errors.New("boom"))
			txc.Rollback()

			apiErr := DeleteBreakdownCachesOfSubtree(ctxWithTenant(1), runaidb.DB(), []string{"p1"})
			Expect(apiErr).ToNot(BeNil())
		})
	})
})
