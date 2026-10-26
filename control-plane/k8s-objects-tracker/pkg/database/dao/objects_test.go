/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Objects", func() {
	Describe("ListObjects", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM objects WHERE objects.tenant_id = ?`).Returns([]models.Object{})
			_, apiErr := ListObjects(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM objects WHERE objects.tenant_id = ?`).ReturnsError(errors.New("boom"))
			_, apiErr := ListObjects(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("ListStorageClassWithOptionalConfig", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM objects LEFT JOIN storage_class_configurations ON storage_class_configurations.tenant_id = objects.tenant_id AND storage_class_configurations.scope_type = objects.scope_type AND storage_class_configurations.scope_id = objects.cluster_id::text AND storage_class_configurations.name = objects.name WHERE objects.tenant_id = ? AND objects.type = ?`).
				Returns([]models.StorageClassWithOptionalConfig{})
			_, apiErr := ListStorageClassWithOptionalConfig(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM objects LEFT JOIN storage_class_configurations ON storage_class_configurations.tenant_id = objects.tenant_id AND storage_class_configurations.scope_type = objects.scope_type AND storage_class_configurations.scope_id = objects.cluster_id::text AND storage_class_configurations.name = objects.name WHERE objects.tenant_id = ? AND objects.type = ?`).
				ReturnsError(errors.New("boom"))
			_, apiErr := ListStorageClassWithOptionalConfig(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("UpsertObjects", func() {
		It("succeeds with an empty list", func() {
			setupDaoMock(GinkgoT())
			apiErr := UpsertObjects(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeStorageClass}, []*models.Object{})
			Expect(apiErr).To(BeNil())
		})

		It("fails fast for an unknown object type", func() {
			setupDaoMock(GinkgoT())
			apiErr := UpsertObjects(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1},
				[]*models.Object{{ObjectFilters: models.ObjectFilters{TenantId: 1, Name: ptr.To("o1")}}})
			Expect(apiErr).ToNot(BeNil())
		})

		It("succeeds for a cluster-scoped kind (StorageClass / DeviceClass)", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO objects (id,tenant_id,type,project_id,cluster_id,department_id,scope_type,name,spec) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT (tenant_id,type,cluster_id,name) WHERE type IN ('StorageClass','DeviceClass') DO UPDATE SET tenant_id=excluded.tenant_id,type=excluded.type,project_id=excluded.project_id,cluster_id=excluded.cluster_id,department_id=excluded.department_id,scope_type=excluded.scope_type,name=excluded.name,spec=excluded.spec`).Returns(1)
			tx.Commit()
			apiErr := UpsertObjects(ctxWithTenant(1),
				&models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeStorageClass},
				[]*models.Object{{ObjectFilters: models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeStorageClass, Name: ptr.To("o1")}}})
			Expect(apiErr).To(BeNil())
		})

		It("succeeds for ResourceClaimTemplate (scope-aware target)", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO objects (id,tenant_id,type,project_id,cluster_id,department_id,scope_type,name,spec) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT (tenant_id,cluster_id,scope_type,name,project_id,department_id) WHERE type = 'ResourceClaimTemplate' DO UPDATE SET tenant_id=excluded.tenant_id,type=excluded.type,project_id=excluded.project_id,cluster_id=excluded.cluster_id,department_id=excluded.department_id,scope_type=excluded.scope_type,name=excluded.name,spec=excluded.spec`).Returns(1)
			tx.Commit()
			apiErr := UpsertObjects(ctxWithTenant(1),
				&models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeResourceClaimTemplate},
				[]*models.Object{{ObjectFilters: models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeResourceClaimTemplate, Name: ptr.To("o1")}}})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO objects (id,tenant_id,type,project_id,cluster_id,department_id,scope_type,name,spec) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT (tenant_id,type,cluster_id,name) WHERE type IN ('StorageClass','DeviceClass') DO UPDATE SET tenant_id=excluded.tenant_id,type=excluded.type,project_id=excluded.project_id,cluster_id=excluded.cluster_id,department_id=excluded.department_id,scope_type=excluded.scope_type,name=excluded.name,spec=excluded.spec`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := UpsertObjects(ctxWithTenant(1),
				&models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeStorageClass},
				[]*models.Object{{ObjectFilters: models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeStorageClass, Name: ptr.To("o1")}}})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("Delete", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM objects WHERE objects.tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := Delete(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM objects WHERE objects.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := Delete(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteObjectsNotInCluster", func() {
		It("succeeds with empty objects", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM objects WHERE objects.tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteObjectsNotInCluster(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1}, []*models.Object{})
			Expect(apiErr).To(BeNil())
		})

		It("succeeds with objects", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM objects WHERE objects.tenant_id = ? AND NOT name in (?)`).Returns(1)
			tx.Commit()
			apiErr := DeleteObjectsNotInCluster(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1}, []*models.Object{{ObjectFilters: models.ObjectFilters{TenantId: 1, Name: ptr.To("o1")}}})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM objects WHERE objects.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := DeleteObjectsNotInCluster(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1}, []*models.Object{})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteObjectsNotInScopedKeys", func() {
		// Row shape for Pluck("id", ...): a one-column struct whose gorm tag
		// pins the column name to "id" so sqlmock emits the right column.
		type idRow struct {
			Id string `gorm:"column:id"`
		}

		It("deletes objects without touching resource_issues when nothing is pruned", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			// Pluck returns no ids - the issues delete must be skipped.
			m.ExpectQuery(`SELECT id FROM objects WHERE objects.tenant_id = ? AND objects.type = ?`).
				Returns([]idRow{})
			m.ExpectExec(`DELETE FROM objects WHERE objects.tenant_id = ? AND objects.type = ?`).Returns(0)
			tx.Commit()
			apiErr := DeleteObjectsNotInScopedKeys(ctxWithTenant(1),
				&models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeResourceClaimTemplate}, nil)
			Expect(apiErr).To(BeNil())
		})

		It("deletes resource_issues for pruned ids inside the same transaction", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT id FROM objects WHERE objects.tenant_id = ? AND objects.type = ?`).
				Returns([]idRow{{Id: "11111111-1111-1111-1111-111111111111"}, {Id: "22222222-2222-2222-2222-222222222222"}})
			// Tenant-only scope (ClusterId left zero), so no cluster_id predicate on the issues delete.
			m.ExpectExec(`DELETE FROM resource_issues WHERE tenant_id = ? AND resource_id IN (?,?)`).Returns(2)
			m.ExpectExec(`DELETE FROM objects WHERE objects.tenant_id = ? AND objects.type = ?`).Returns(2)
			tx.Commit()
			apiErr := DeleteObjectsNotInScopedKeys(ctxWithTenant(1),
				&models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeResourceClaimTemplate}, nil)
			Expect(apiErr).To(BeNil())
		})

		It("rolls back when the resource_issues delete fails", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT id FROM objects WHERE objects.tenant_id = ? AND objects.type = ?`).
				Returns([]idRow{{Id: "11111111-1111-1111-1111-111111111111"}})
			m.ExpectExec(`DELETE FROM resource_issues WHERE tenant_id = ? AND resource_id IN (?)`).
				ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := DeleteObjectsNotInScopedKeys(ctxWithTenant(1),
				&models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeResourceClaimTemplate}, nil)
			Expect(apiErr).ToNot(BeNil())
		})

		It("narrows resource_issues by cluster_id when the filter sets one", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			// Column order mirrors the ObjectFilters struct: tenant_id,
			// type, cluster_id - that is the order GORM emits them in.
			m.ExpectQuery(`SELECT id FROM objects WHERE objects.tenant_id = ? AND objects.type = ? AND objects.cluster_id = ?`).
				Returns([]idRow{{Id: "11111111-1111-1111-1111-111111111111"}})
			m.ExpectExec(`DELETE FROM resource_issues WHERE tenant_id = ? AND cluster_id = ? AND resource_id IN (?)`).Returns(1)
			m.ExpectExec(`DELETE FROM objects WHERE objects.tenant_id = ? AND objects.type = ? AND objects.cluster_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteObjectsNotInScopedKeys(ctxWithTenant(1),
				&models.ObjectFilters{
					TenantId:  1,
					ClusterId: uuid.MustParse("7daba6ee-2ea9-4af3-b735-a364bcdff72f"),
					Type:      models.ObjectTypeResourceClaimTemplate,
				}, nil)
			Expect(apiErr).To(BeNil())
		})

		It("appends the composite NOT IN predicate when keptKeys is non-empty", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			// GORM parenthesises the struct-based WHERE as a group before
			// AND-ing on the composite NOT IN - the SELECT and DELETE
			// expectations have to carry the same parentheses.
			m.ExpectQuery(`SELECT id FROM objects WHERE (objects.tenant_id = ? AND objects.type = ?) AND (scope_type, department_id, project_id, name) NOT IN ((?,?,?,?),(?,?,?,?))`).
				Returns([]idRow{})
			m.ExpectExec(`DELETE FROM objects WHERE (objects.tenant_id = ? AND objects.type = ?) AND (scope_type, department_id, project_id, name) NOT IN ((?,?,?,?),(?,?,?,?))`).Returns(0)
			tx.Commit()
			apiErr := DeleteObjectsNotInScopedKeys(ctxWithTenant(1),
				&models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeResourceClaimTemplate},
				[][]interface{}{
					{"cluster", "", "", "rct-a"},
					{"project", "", "4500001", "rct-b"},
				})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on an initial Pluck failure and rolls back", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT id FROM objects WHERE objects.tenant_id = ? AND objects.type = ?`).
				ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := DeleteObjectsNotInScopedKeys(ctxWithTenant(1),
				&models.ObjectFilters{TenantId: 1, Type: models.ObjectTypeResourceClaimTemplate}, nil)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteProjectObjects", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM objects WHERE tenant_id = ? AND scope_type = ? AND project_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteProjectObjects(ctxWithTenant(1), 1, "p1")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteClusterObjects", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM objects WHERE tenant_id = ? AND cluster_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteClusterObjects(ctxWithTenant(1), 1, "c1")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteTenantObjects", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM objects WHERE tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteTenantObjects(ctxWithTenant(1), 1)
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("extractNames", func() {
		It("extracts names from objects", func() {
			names := extractNames([]*models.Object{
				{ObjectFilters: models.ObjectFilters{Name: ptr.To("a")}},
				{ObjectFilters: models.ObjectFilters{Name: ptr.To("b")}},
			})
			Expect(names).To(Equal([]string{"a", "b"}))
		})
	})
})
