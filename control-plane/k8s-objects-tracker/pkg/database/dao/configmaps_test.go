/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"k8s.io/utils/ptr"
)

var _ = Describe("ConfigMaps", func() {
	Describe("ListConfigMaps", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM configmaps WHERE configmaps.tenant_id = ?`).Returns([]models.ConfigMap{})
			_, apiErr := (&configMapsDBAccessor{}).ListConfigMaps(ctxWithTenant(1), &models.ConfigMapFilters{}, nil)
			Expect(apiErr).To(BeNil())
		})

		It("succeeds with filter and options", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM configmaps WHERE configmaps.tenant_id = ? AND configmaps.cluster_id = ?`).Returns([]models.ConfigMap{})
			filter := &models.ConfigMapFilters{ClusterId: ptr.To("c1")}
			opts := &models.ListOptions{IncludeAncestors: true, IncludeDescendants: true}
			_, apiErr := (&configMapsDBAccessor{}).ListConfigMaps(ctxWithTenant(1), filter, opts)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error when the tenant is missing", func() {
			setupDaoMock(GinkgoT())
			_, apiErr := (&configMapsDBAccessor{}).ListConfigMaps(context.Background(), &models.ConfigMapFilters{}, nil)
			Expect(apiErr).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM configmaps WHERE configmaps.tenant_id = ?`).ReturnsError(errors.New("boom"))
			_, apiErr := (&configMapsDBAccessor{}).ListConfigMaps(ctxWithTenant(1), &models.ConfigMapFilters{}, nil)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("CreateConfigMap", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM configmaps WHERE configmaps.tenant_id = ? AND configmaps.name = ? ORDER BY configmaps.id LIMIT ?`).
				Returns([]models.ConfigMap{{ConfigMapFilters: models.ConfigMapFilters{TenantId: 1, Name: ptr.To("cm")}}})
			_, apiErr := (&configMapsDBAccessor{}).CreateConfigMap(ctxWithTenant(1), &models.ConfigMap{ConfigMapFilters: models.ConfigMapFilters{Name: ptr.To("cm")}})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error when the tenant is missing", func() {
			setupDaoMock(GinkgoT())
			_, apiErr := (&configMapsDBAccessor{}).CreateConfigMap(context.Background(), &models.ConfigMap{ConfigMapFilters: models.ConfigMapFilters{Name: ptr.To("cm")}})
			Expect(apiErr).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM configmaps WHERE configmaps.tenant_id = ? AND configmaps.name = ? ORDER BY configmaps.id LIMIT ?`).
				ReturnsError(errors.New("boom"))
			_, apiErr := (&configMapsDBAccessor{}).CreateConfigMap(ctxWithTenant(1), &models.ConfigMap{ConfigMapFilters: models.ConfigMapFilters{Name: ptr.To("cm")}})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteConfigMap", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE configmaps.tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := (&configMapsDBAccessor{}).DeleteConfigMap(ctxWithTenant(1), &models.ConfigMapFilters{})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error when the tenant is missing", func() {
			setupDaoMock(GinkgoT())
			apiErr := (&configMapsDBAccessor{}).DeleteConfigMap(context.Background(), &models.ConfigMapFilters{})
			Expect(apiErr).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE configmaps.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := (&configMapsDBAccessor{}).DeleteConfigMap(ctxWithTenant(1), &models.ConfigMapFilters{})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("ApplyConfigMaps", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			// upsert (FirstOrCreate with OnConflict DoNothing) per configmap
			m.ExpectQuery(`SELECT * FROM configmaps WHERE configmaps.tenant_id = ? AND configmaps.name = ? ORDER BY configmaps.id LIMIT ?`).
				Returns([]models.ConfigMap{{ConfigMapFilters: models.ConfigMapFilters{TenantId: 1, Name: ptr.To("cm-a")}}})
			// delete not-in
			m.ExpectExec(`DELETE FROM configmaps WHERE configmaps.tenant_id = ? AND name not in (?)`).Returns(0)
			tx.Commit()
			_, apiErr := (&configMapsDBAccessor{}).ApplyConfigMaps(ctxWithTenant(1), &models.ConfigMapFilters{TenantId: 1}, []api.ConfigMap{{Name: "cm-a"}})
			Expect(apiErr).To(BeNil())
		})

		It("succeeds with an empty list", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE configmaps.tenant_id = ? AND name not in (NULL)`).Returns(0)
			tx.Commit()
			_, apiErr := (&configMapsDBAccessor{}).ApplyConfigMaps(ctxWithTenant(1), &models.ConfigMapFilters{TenantId: 1}, []api.ConfigMap{})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT * FROM configmaps WHERE configmaps.tenant_id = ? AND configmaps.name = ? ORDER BY configmaps.id LIMIT ?`).
				ReturnsError(errors.New("boom"))
			tx.Rollback()
			_, apiErr := (&configMapsDBAccessor{}).ApplyConfigMaps(ctxWithTenant(1), &models.ConfigMapFilters{TenantId: 1}, []api.ConfigMap{{Name: "cm-a"}})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteProjectConfigMaps", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE tenant_id = ? AND scope_type = ? AND scope_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteProjectConfigMaps(ctxWithTenant(1), 1, "p1")
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE tenant_id = ? AND scope_type = ? AND scope_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := DeleteProjectConfigMaps(ctxWithTenant(1), 1, "p1")
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteClusterConfigMaps", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE tenant_id = ? AND cluster_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteClusterConfigMaps(ctxWithTenant(1), 1, "c1")
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE tenant_id = ? AND cluster_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := DeleteClusterConfigMaps(ctxWithTenant(1), 1, "c1")
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteTenantConfigMaps", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteTenantConfigMaps(ctxWithTenant(1), 1)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := DeleteTenantConfigMaps(ctxWithTenant(1), 1)
			Expect(apiErr).ToNot(BeNil())
		})
	})
})
