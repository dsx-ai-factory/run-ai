/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
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
			apiErr := UpsertObjects(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1}, []*models.Object{})
			Expect(apiErr).To(BeNil())
		})

		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO objects (tenant_id,type,project_id,cluster_id,department_id,scope_type,name,spec) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT ON CONSTRAINT unique_tenant_type_cluster_name_project DO UPDATE SET tenant_id=excluded.tenant_id,type=excluded.type,project_id=excluded.project_id,cluster_id=excluded.cluster_id,department_id=excluded.department_id,scope_type=excluded.scope_type,name=excluded.name,spec=excluded.spec`).Returns(1)
			tx.Commit()
			apiErr := UpsertObjects(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1}, []*models.Object{{ObjectFilters: models.ObjectFilters{TenantId: 1, Name: ptr.To("o1")}}})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO objects (tenant_id,type,project_id,cluster_id,department_id,scope_type,name,spec) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT ON CONSTRAINT unique_tenant_type_cluster_name_project DO UPDATE SET tenant_id=excluded.tenant_id,type=excluded.type,project_id=excluded.project_id,cluster_id=excluded.cluster_id,department_id=excluded.department_id,scope_type=excluded.scope_type,name=excluded.name,spec=excluded.spec`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := UpsertObjects(ctxWithTenant(1), &models.ObjectFilters{TenantId: 1}, []*models.Object{{ObjectFilters: models.ObjectFilters{TenantId: 1, Name: ptr.To("o1")}}})
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
