package dao

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("EntitiesDependency", func() {
	Describe("UpsertEntityDependencies", func() {
		It("handles empty", func() {
			setupDaoMock(GinkgoT())
			apiErr := UpsertEntityDependencies(ctxWithTenant(1), runaidb.DB(), nil)
			Expect(apiErr).To(BeNil())
		})

		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO entities_dependency (cache_id,tenant_id,component_type,component_id) VALUES (?,?,?,?) ON CONFLICT DO NOTHING`).Returns(1)
			tx.Commit()

			rows := []dbmodel.EntitiesDependency{
				{CacheId: uuid.New(), ComponentType: dbmodel.ComponentTypeMapping, ComponentId: "m1"},
			}
			apiErr := UpsertEntityDependencies(ctxWithTenant(1), runaidb.DB(), rows)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO entities_dependency (cache_id,tenant_id,component_type,component_id) VALUES (?,?,?,?) ON CONFLICT DO NOTHING`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			rows := []dbmodel.EntitiesDependency{
				{CacheId: uuid.New(), ComponentType: dbmodel.ComponentTypeMapping, ComponentId: "m1"},
			}
			apiErr := UpsertEntityDependencies(ctxWithTenant(1), runaidb.DB(), rows)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("OnComponentDeleted", func() {
		It("handles no caches", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT DISTINCT cache_id FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND entities_dependency.component_type = ? AND entities_dependency.component_id = ?`).
				Returns([]struct{ CacheId uuid.UUID }{})

			apiErr := OnComponentDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.ComponentTypeMapping, "m1")
			Expect(apiErr).To(BeNil())
		})

		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT DISTINCT cache_id FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND entities_dependency.component_type = ? AND entities_dependency.component_id = ?`).
				Returns([]struct{ CacheId uuid.UUID }{{CacheId: cacheId}})

			m.ExpectQuery(`SELECT entity_id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND id IN (?)`).
				Returns([]struct{ EntityId string }{{EntityId: uuid.New().String()}})

			txu := m.ExpectTransaction()
			txu.Begin()
			m.ExpectExec(`UPDATE policies SET validation_issues=?,validation_status=? WHERE policies.tenant_id = ? AND policies.id IN (?)`).Returns(1)
			txu.Commit()

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND cache_id IN (?)`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache WHERE entities_cache.tenant_id = ? AND id IN (?)`).Returns(1)
			txc.Commit()

			apiErr := OnComponentDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.ComponentTypeMapping, "m1")
			Expect(apiErr).To(BeNil())
		})

		It("handles step 1 error", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT DISTINCT cache_id FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND entities_dependency.component_type = ? AND entities_dependency.component_id = ?`).ReturnsError(errors.New("boom"))

			apiErr := OnComponentDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.ComponentTypeMapping, "m1")
			Expect(apiErr).ToNot(BeNil())
		})

		It("handles step 2 error", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT DISTINCT cache_id FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND entities_dependency.component_type = ? AND entities_dependency.component_id = ?`).
				Returns([]struct{ CacheId uuid.UUID }{{CacheId: cacheId}})

			m.ExpectQuery(`SELECT entity_id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND id IN (?)`).ReturnsError(errors.New("boom"))

			apiErr := OnComponentDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.ComponentTypeMapping, "m1")
			Expect(apiErr).ToNot(BeNil())
		})

		It("handles no policy ids", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT DISTINCT cache_id FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND entities_dependency.component_type = ? AND entities_dependency.component_id = ?`).
				Returns([]struct{ CacheId uuid.UUID }{{CacheId: cacheId}})

			m.ExpectQuery(`SELECT entity_id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND id IN (?)`).
				Returns([]struct{ EntityId string }{})

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND cache_id IN (?)`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache WHERE entities_cache.tenant_id = ? AND id IN (?)`).Returns(1)
			txc.Commit()

			apiErr := OnComponentDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.ComponentTypeMapping, "m1")
			Expect(apiErr).To(BeNil())
		})

		It("handles status update error", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT DISTINCT cache_id FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND entities_dependency.component_type = ? AND entities_dependency.component_id = ?`).
				Returns([]struct{ CacheId uuid.UUID }{{CacheId: cacheId}})

			m.ExpectQuery(`SELECT entity_id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND id IN (?)`).
				Returns([]struct{ EntityId string }{{EntityId: uuid.New().String()}})

			txu := m.ExpectTransaction()
			txu.Begin()
			m.ExpectExec(`UPDATE policies SET validation_issues=?,validation_status=? WHERE policies.tenant_id = ? AND policies.id IN (?)`).ReturnsError(errors.New("boom"))
			txu.Rollback()

			apiErr := OnComponentDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.ComponentTypeMapping, "m1")
			Expect(apiErr).ToNot(BeNil())
		})

		It("handles dep delete error", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT DISTINCT cache_id FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND entities_dependency.component_type = ? AND entities_dependency.component_id = ?`).
				Returns([]struct{ CacheId uuid.UUID }{{CacheId: cacheId}})

			m.ExpectQuery(`SELECT entity_id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND id IN (?)`).
				Returns([]struct{ EntityId string }{{EntityId: uuid.New().String()}})

			txu := m.ExpectTransaction()
			txu.Begin()
			m.ExpectExec(`UPDATE policies SET validation_issues=?,validation_status=? WHERE policies.tenant_id = ? AND policies.id IN (?)`).Returns(1)
			txu.Commit()

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND cache_id IN (?)`).ReturnsError(errors.New("boom"))
			txd.Rollback()

			apiErr := OnComponentDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.ComponentTypeMapping, "m1")
			Expect(apiErr).ToNot(BeNil())
		})

		It("handles cache delete error", func() {
			m := setupDaoMock(GinkgoT())

			cacheId := uuid.New()
			m.ExpectQuery(`SELECT DISTINCT cache_id FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND entities_dependency.component_type = ? AND entities_dependency.component_id = ?`).
				Returns([]struct{ CacheId uuid.UUID }{{CacheId: cacheId}})

			m.ExpectQuery(`SELECT entity_id FROM entities_cache WHERE (entities_cache.tenant_id = ? AND entities_cache.entity_type = ?) AND id IN (?)`).
				Returns([]struct{ EntityId string }{{EntityId: uuid.New().String()}})

			txu := m.ExpectTransaction()
			txu.Begin()
			m.ExpectExec(`UPDATE policies SET validation_issues=?,validation_status=? WHERE policies.tenant_id = ? AND policies.id IN (?)`).Returns(1)
			txu.Commit()

			txd := m.ExpectTransaction()
			txd.Begin()
			m.ExpectExec(`DELETE FROM entities_dependency WHERE entities_dependency.tenant_id = ? AND cache_id IN (?)`).Returns(1)
			txd.Commit()

			txc := m.ExpectTransaction()
			txc.Begin()
			m.ExpectExec(`DELETE FROM entities_cache WHERE entities_cache.tenant_id = ? AND id IN (?)`).ReturnsError(errors.New("boom"))
			txc.Rollback()

			apiErr := OnComponentDeleted(ctxWithTenant(1), runaidb.DB(), dbmodel.ComponentTypeMapping, "m1")
			Expect(apiErr).ToNot(BeNil())
		})
	})
})
