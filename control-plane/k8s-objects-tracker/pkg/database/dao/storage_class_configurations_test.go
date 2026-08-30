package dao

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"k8s.io/utils/ptr"
)

var _ = Describe("StorageClassConfigurations", func() {
	Describe("ListStorageClassConfigurations", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM storage_class_configurations ORDER BY id`).Returns([]models.StorageClassConfiguration{})
			_, apiErr := ListStorageClassConfigurations(ctxWithTenant(1), nil, "", "", 0, 0, nil)
			Expect(apiErr).To(BeNil())
		})

		It("applies filters, sort, scopes and paging", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM storage_class_configurations WHERE ((false OR tenant_id IN ( '5'))) AND name = ? ORDER BY name ASC NULLS FIRST,id LIMIT ? OFFSET ?`).
				Returns([]models.StorageClassConfiguration{})
			filters := []queryapi.Filter{{Field: "name", Operator: "=", Values: []string{"sc1"}}}
			scopes := &openapi.PermittedScopes{Tenants: []string{"5"}}
			_, apiErr := ListStorageClassConfigurations(ctxWithTenant(1), filters, "name", "asc", 10, 20, scopes)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM storage_class_configurations ORDER BY id`).ReturnsError(errors.New("boom"))
			_, apiErr := ListStorageClassConfigurations(ctxWithTenant(1), nil, "", "", 0, 0, nil)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("GetStorageClassConfigurationById", func() {
		It("returns the configuration when found", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			m.ExpectQuery(`SELECT * FROM storage_class_configurations WHERE storage_class_configurations.id = ? AND storage_class_configurations.tenant_id = ? ORDER BY storage_class_configurations.id LIMIT ?`).
				Returns([]models.StorageClassConfiguration{{ID: id, StorageClassConfigurationFilters: models.StorageClassConfigurationFilters{TenantId: 1}}})
			_, apiErr := GetStorageClassConfigurationById(ctxWithTenant(1), id.String(), 1)
			Expect(apiErr).To(BeNil())
		})

		It("returns not found when missing", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			m.ExpectQuery(`SELECT * FROM storage_class_configurations WHERE storage_class_configurations.id = ? AND storage_class_configurations.tenant_id = ? ORDER BY storage_class_configurations.id LIMIT ?`).
				Returns([]models.StorageClassConfiguration{})
			_, apiErr := GetStorageClassConfigurationById(ctxWithTenant(1), id.String(), 1)
			Expect(apiErr).ToNot(BeNil())
			Expect(int(apiErr.Code)).To(Equal(http.StatusNotFound))
		})

		It("returns an error on invalid uuid", func() {
			setupDaoMock(GinkgoT())
			_, apiErr := GetStorageClassConfigurationById(ctxWithTenant(1), "not-a-uuid", 1)
			Expect(apiErr).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			m.ExpectQuery(`SELECT * FROM storage_class_configurations WHERE storage_class_configurations.id = ? AND storage_class_configurations.tenant_id = ? ORDER BY storage_class_configurations.id LIMIT ?`).
				ReturnsError(errors.New("boom"))
			_, apiErr := GetStorageClassConfigurationById(ctxWithTenant(1), id.String(), 1)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("GetStorageClassConfigurationByName", func() {
		It("returns the configuration when found", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM storage_class_configurations WHERE storage_class_configurations.tenant_id = ? AND storage_class_configurations.scope_type = ? AND storage_class_configurations.scope_id = ? AND storage_class_configurations.name = ? ORDER BY storage_class_configurations.id LIMIT ?`).
				Returns([]models.StorageClassConfiguration{{ID: uuid.New()}})
			_, apiErr := GetStorageClassConfigurationByName(ctxWithTenant(1), "c1", "sc1", 1)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM storage_class_configurations WHERE storage_class_configurations.tenant_id = ? AND storage_class_configurations.scope_type = ? AND storage_class_configurations.scope_id = ? AND storage_class_configurations.name = ? ORDER BY storage_class_configurations.id LIMIT ?`).
				ReturnsError(errors.New("boom"))
			_, apiErr := GetStorageClassConfigurationByName(ctxWithTenant(1), "c1", "sc1", 1)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("CreateStorageClassConfiguration", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO storage_class_configurations (id,tenant_id,scope_type,scope_id,name,permissions,customization) VALUES (?,?,?,?,?,?,?)`).Returns(1)
			tx.Commit()
			_, apiErr := CreateStorageClassConfiguration(ctxWithTenant(1), &models.StorageClassConfiguration{
				StorageClassConfigurationFilters: models.StorageClassConfigurationFilters{TenantId: 1, Name: ptr.To("sc1")},
			})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO storage_class_configurations (id,tenant_id,scope_type,scope_id,name,permissions,customization) VALUES (?,?,?,?,?,?,?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			_, apiErr := CreateStorageClassConfiguration(ctxWithTenant(1), &models.StorageClassConfiguration{
				StorageClassConfigurationFilters: models.StorageClassConfigurationFilters{TenantId: 1, Name: ptr.To("sc1")},
			})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteStorageClassConfiguration", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM storage_class_configurations WHERE storage_class_configurations.id = ? AND storage_class_configurations.tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteStorageClassConfiguration(ctxWithTenant(1), id.String(), 1)
			Expect(apiErr).To(BeNil())
		})

		It("returns not found when nothing is deleted", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM storage_class_configurations WHERE storage_class_configurations.id = ? AND storage_class_configurations.tenant_id = ?`).Returns(0)
			tx.Commit()
			apiErr := DeleteStorageClassConfiguration(ctxWithTenant(1), id.String(), 1)
			Expect(apiErr).ToNot(BeNil())
			Expect(int(apiErr.Code)).To(Equal(http.StatusNotFound))
		})

		It("returns an error on invalid uuid", func() {
			setupDaoMock(GinkgoT())
			apiErr := DeleteStorageClassConfiguration(ctxWithTenant(1), "not-a-uuid", 1)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("UpdateStorageClassConfiguration", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE storage_class_configurations SET name=? WHERE id = ?`).Returns(1)
			tx.Commit()
			_, apiErr := UpdateStorageClassConfiguration(ctxWithTenant(1), id.String(), &models.StorageClassConfiguration{
				StorageClassConfigurationFilters: models.StorageClassConfigurationFilters{Name: ptr.To("sc2")},
			}, 1)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on invalid uuid", func() {
			setupDaoMock(GinkgoT())
			_, apiErr := UpdateStorageClassConfiguration(ctxWithTenant(1), "not-a-uuid", &models.StorageClassConfiguration{}, 1)
			Expect(apiErr).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE storage_class_configurations SET name=? WHERE id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			_, apiErr := UpdateStorageClassConfiguration(ctxWithTenant(1), id.String(), &models.StorageClassConfiguration{
				StorageClassConfigurationFilters: models.StorageClassConfigurationFilters{Name: ptr.To("sc2")},
			}, 1)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteProjectStorageClasses", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM storage_class_configurations WHERE tenant_id = ? AND scope_type = ? AND scope_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteProjectStorageClasses(ctxWithTenant(1), 1, "p1")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteClusterStorageClasses", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM storage_class_configurations WHERE tenant_id = ? AND scope_type = ? AND scope_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteClusterStorageClasses(ctxWithTenant(1), 1, "c1")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteTenantStorageClasses", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM storage_class_configurations WHERE tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteTenantStorageClasses(ctxWithTenant(1), 1)
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("getPermittedScopesQuery", func() {
		It("builds the permitted scopes query", func() {
			Expect(getPermittedScopesQuery(openapi.PermittedScopes{System: ptr.To(true)})).To(Equal("(true)"))
			Expect(getPermittedScopesQuery(openapi.PermittedScopes{})).To(Equal("(false )"))
			Expect(getPermittedScopesQuery(openapi.PermittedScopes{Tenants: []string{"5"}})).To(Equal("(false OR tenant_id IN ( '5'))"))
			Expect(getPermittedScopesQuery(openapi.PermittedScopes{Clusters: []string{"c1"}})).To(Equal("(false OR (scope_type = 'cluster' AND scope_id IN ( 'c1')))"))
		})
	})
})
