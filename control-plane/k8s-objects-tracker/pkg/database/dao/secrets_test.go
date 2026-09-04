/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

func secretFilters() *models.SecretFilters {
	return &models.SecretFilters{
		TenantId:   1,
		ClusterId:  ptr.To("c1"),
		ScopeQuery: &models.ScopeQueryParams{ScopeType: "project", ScopeId: "p1"},
	}
}

var _ = Describe("Secrets", func() {
	Describe("ListSecrets", func() {
		It("excludes private secrets", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM secrets WHERE secrets.tenant_id = ? AND (private IS NULL OR private = ?)`).Returns([]models.Secret{})
			_, apiErr := ListSecrets(ctxWithTenant(1), &models.SecretFilters{TenantId: 1}, nil)
			Expect(apiErr).To(BeNil())
		})

		It("includes private secrets", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM secrets WHERE secrets.tenant_id = ?`).Returns([]models.Secret{})
			_, apiErr := ListSecrets(ctxWithTenant(1), &models.SecretFilters{TenantId: 1}, &models.ListOptions{IncludePrivate: true})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM secrets WHERE secrets.tenant_id = ? AND (private IS NULL OR private = ?)`).ReturnsError(errors.New("boom"))
			_, apiErr := ListSecrets(ctxWithTenant(1), &models.SecretFilters{TenantId: 1}, nil)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("CreateSecret", func() {
		It("creates a new secret", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM secrets WHERE secrets.tenant_id = ? AND secrets.secret_name = ? ORDER BY secrets.id LIMIT ?`).Returns([]models.Secret{})
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO secrets (id,tenant_id,cluster_id,secret_name,secret_type,origin,scope_type,scope_id,private) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`).Returns(1)
			tx.Commit()
			_, apiErr := CreateSecret(ctxWithTenant(1), &models.Secret{SecretFilters: models.SecretFilters{TenantId: 1, SecretName: ptr.To("s1")}})
			Expect(apiErr).To(BeNil())
		})

		It("updates an existing secret", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM secrets WHERE secrets.tenant_id = ? AND secrets.secret_name = ? ORDER BY secrets.id LIMIT ?`).
				Returns([]models.Secret{{ID: uuid.New(), SecretFilters: models.SecretFilters{TenantId: 1, SecretName: ptr.To("s1")}}})
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE secrets SET tenant_id=?,cluster_id=?,secret_name=?,secret_type=?,origin=?,scope_type=?,scope_id=?,private=? WHERE id = ?`).Returns(1)
			tx.Commit()
			_, apiErr := CreateSecret(ctxWithTenant(1), &models.Secret{SecretFilters: models.SecretFilters{TenantId: 1, SecretName: ptr.To("s1")}})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM secrets WHERE secrets.tenant_id = ? AND secrets.secret_name = ? ORDER BY secrets.id LIMIT ?`).ReturnsError(errors.New("boom"))
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO secrets (id,tenant_id,cluster_id,secret_name,secret_type,origin,scope_type,scope_id,private) VALUES (?,?,?,?,?,?,?,?,?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			_, apiErr := CreateSecret(ctxWithTenant(1), &models.Secret{SecretFilters: models.SecretFilters{TenantId: 1, SecretName: ptr.To("s1")}})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteSecret", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM secrets WHERE secrets.tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteSecret(ctxWithTenant(1), &models.SecretFilters{TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM secrets WHERE secrets.tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := DeleteSecret(ctxWithTenant(1), &models.SecretFilters{TenantId: 1})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("ApplySecrets", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT * FROM secrets WHERE secrets.tenant_id = ? AND secrets.cluster_id = ? AND secrets.secret_name = ? AND secrets.origin = ? AND secrets.scope_type = ? AND secrets.scope_id = ? ORDER BY secrets.id LIMIT ?`).Returns([]models.Secret{})
			m.ExpectExec(`INSERT INTO secrets (id,tenant_id,cluster_id,secret_name,secret_type,origin,scope_type,scope_id,private) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`).Returns(1)
			m.ExpectExec(`DELETE FROM secrets WHERE (secrets.tenant_id = ? AND secrets.cluster_id = ? AND secrets.scope_type = ? AND secrets.scope_id = ?) AND secret_name not in (?)`).Returns(0)
			tx.Commit()
			_, apiErr := ApplySecrets(ctxWithTenant(1), secretFilters(), []api.Secret{{Name: "s1", Type: api.SECRETTYPE_PASSWORD}})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT * FROM secrets WHERE secrets.tenant_id = ? AND secrets.cluster_id = ? AND secrets.secret_name = ? AND secrets.origin = ? AND secrets.scope_type = ? AND secrets.scope_id = ? ORDER BY secrets.id LIMIT ?`).Returns([]models.Secret{})
			m.ExpectExec(`INSERT INTO secrets (id,tenant_id,cluster_id,secret_name,secret_type,origin,scope_type,scope_id,private) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			_, apiErr := ApplySecrets(ctxWithTenant(1), secretFilters(), []api.Secret{{Name: "s1", Type: api.SECRETTYPE_PASSWORD}})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteSecretsNotInCluster", func() {
		It("succeeds with ids", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM secrets WHERE tenant_id = ? AND secrets.tenant_id = ? AND NOT id in (?)`).Returns(1)
			tx.Commit()
			apiErr := DeleteSecretsNotInCluster(ctxWithTenant(1), &models.SecretFilters{TenantId: 1}, []string{"id1"})
			Expect(apiErr).To(BeNil())
		})

		It("succeeds with no ids", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM secrets WHERE tenant_id = ? AND secrets.tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteSecretsNotInCluster(ctxWithTenant(1), &models.SecretFilters{TenantId: 1}, nil)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error when the tenant is missing", func() {
			setupDaoMock(GinkgoT())
			apiErr := DeleteSecretsNotInCluster(context.Background(), &models.SecretFilters{TenantId: 1}, nil)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("OriginPtrToStringPtr", func() {
		It("maps origin pointers to string pointers", func() {
			Expect(*OriginPtrToStringPtr(nil)).To(Equal(string(api.ORIGIN_EXISTING)))
			empty := api.Origin("")
			Expect(*OriginPtrToStringPtr(&empty)).To(Equal(string(api.ORIGIN_EXISTING)))
			assets := api.ORIGIN_ASSETS
			Expect(*OriginPtrToStringPtr(&assets)).To(Equal(string(api.ORIGIN_ASSETS)))
		})
	})

	Describe("OriginFilterToStringPtr", func() {
		It("maps origin filters to string pointers", func() {
			Expect(OriginFilterToStringPtr(AllOrigins)).To(BeNil())
			Expect(*OriginFilterToStringPtr("")).To(Equal(string(api.ORIGIN_EXISTING)))
			Expect(*OriginFilterToStringPtr("custom")).To(Equal("custom"))
		})
	})

	Describe("DeleteProjectSecrets", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM secrets WHERE tenant_id = ? AND scope_type = ? AND scope_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteProjectSecrets(ctxWithTenant(1), 1, "p1")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteClusterSecrets", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM secrets WHERE tenant_id = ? AND cluster_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteClusterSecrets(ctxWithTenant(1), 1, "c1")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteTenantSecrets", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM secrets WHERE tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteTenantSecrets(ctxWithTenant(1), 1)
			Expect(apiErr).To(BeNil())
		})
	})
})
