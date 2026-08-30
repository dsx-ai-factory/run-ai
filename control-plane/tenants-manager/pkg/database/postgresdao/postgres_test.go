package postgresdao_test

import (
	"context"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
	"k8s.io/utils/ptr"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	internalCache "github.com/run-ai/runai/control-plane/tenants-manager/internal/cache"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/postgresdao"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
)

var (
	tenant1   models.Tenant
	tenant2   models.Tenant
	tenantDao *postgresdao.Tenant
)

var _ = Describe("Database Client", Ordered, func() {
	Describe("Initiate db", Ordered, func() {
		BeforeAll(func() {
			tenant1 = models.Tenant{
				Name:        "test1",
				DisplayName: "TEST1",
				Uuid:        uuid.New().String(),
				MimirId:     uuid.New().String(),
				Status:      models.TenantCreating,
			}
			tenant2 = models.Tenant{
				Name:        "tenant2",
				DisplayName: "TENANT2",
				MimirId:     uuid.New().String(),
				Uuid:        uuid.New().String(),
				Status:      models.TenantCreating,
			}
			cache.Init(&cache_config.Cache{
				Enabled: ptr.To(false),
			})
			tenantDao = postgresdao.NewTenantDao(internalCache.NewTenantCacheManager())
		})
		Context("when tenant db is empty", func() {
			BeforeEach(func() {
				postgresdao.DB().Exec("DELETE FROM " + postgresdao.TableNameTenants)
			})
			It("should create a new tenant", func() {
				newTenant, err := tenantDao.Create(context.Background(), tenant1)
				Expect(err).To(BeNil())
				Expect(newTenant.ID).To(Equal(uint(1001)))
			})
			It("should get empty tenants list", func() {
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, nil)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(0))
			})
			It("should get error 404", func() {
				tenant, err := tenantDao.GetById(context.Background(), 1001)
				Expect(err).ToNot(BeNil())
				Expect(tenant).To(BeNil())
			})
		})
		Context("when tenant db is not empty", func() {
			BeforeEach(func() {
				postgresdao.DB().Exec("DELETE FROM " + postgresdao.TableNameTenants)
				newTenant, err := tenantDao.Create(context.Background(), tenant1)
				Expect(err).To(BeNil())
				Expect(newTenant.Status).To(Equal(tenant1.Status))
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, nil)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(1))
			})
			It("should get TenantByName", func() {
				tenant, err := tenantDao.GetByName(context.Background(), tenant1.Name)
				Expect(err).To(BeNil())
				Expect(tenant.ID).NotTo(Equal(1001))
				Expect(tenant.CreatedAt).NotTo(BeNil())
				Expect(tenant.UpdatedAt).NotTo(BeNil())
				Expect(tenant.CreatedAt).To(Equal(tenant.UpdatedAt))
				Expect(tenant).ShouldNot(HaveExistingField("DeleteAt"))
			})
			It("should create a new tenant also if running migration twice", func() {
				newTenant, err := tenantDao.Create(context.Background(), tenant2)
				Expect(err).To(BeNil())
				Expect(newTenant.Name).To(Equal(tenant2.Name))
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, nil)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(2))
			})
			It("should create a new tenant with excepted dates", func() {
				newTenant, err := tenantDao.Create(context.Background(), tenant2)
				Expect(err).To(BeNil())
				tenant, err := tenantDao.GetById(context.Background(), newTenant.ID)
				Expect(err).To(BeNil())
				Expect(tenant.CreatedAt).NotTo(BeNil())
				Expect(tenant.UpdatedAt).NotTo(BeNil())
				Expect(tenant.CreatedAt).To(Equal(tenant.UpdatedAt))
				Expect(tenant).ShouldNot(HaveExistingField("DeleteAt"))
			})
			It("should error when creating existing tenant name", func() {
				newTenant, err := tenantDao.Create(context.Background(), tenant1)
				Expect(err).NotTo(BeNil())
				Expect(err.Reason).To(Equal(dao.Conflict))
				Expect(newTenant).To(BeNil())
			})
			It("should update status", func() {
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, nil)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(1))
				tenant := (*tenants)[0]
				Expect(tenant.Status).To(Equal(models.TenantCreating))
				newTenant, err := tenantDao.UpdateStatus(context.Background(), tenant.ID, models.TenantReady)
				Expect(err).To(BeNil())
				Expect(newTenant).ToNot(BeNil())
				Expect(newTenant.Status).To(Equal(models.TenantReady))
				Expect(newTenant.Uuid).To(Equal(tenant.Uuid))
				Expect(newTenant.UpdatedAt).ToNot(Equal(newTenant.CreatedAt))
				Expect(newTenant.CreatedAt).To(Equal(tenant.CreatedAt))
			})
		})
		Context("GetAll with filters", func() {
			var tenantWithAlias models.Tenant
			BeforeEach(func() {
				postgresdao.DB().Exec("DELETE FROM " + postgresdao.TableNameTenants)
				alias := "tenant1-alias"
				tenantWithAlias = models.Tenant{
					Name:        "tenant-with-alias",
					DisplayName: "Tenant With Alias",
					Uuid:        uuid.New().String(),
					MimirId:     uuid.New().String(),
					Status:      models.TenantReady,
					Alias:       &alias,
				}
				_, err := tenantDao.Create(context.Background(), tenant1)
				Expect(err).To(BeNil())
				_, err = tenantDao.Create(context.Background(), tenant2)
				Expect(err).To(BeNil())
				_, err = tenantDao.Create(context.Background(), tenantWithAlias)
				Expect(err).To(BeNil())
			})
			It("should return all tenants when no filters provided", func() {
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, nil)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(3))
			})
			It("should filter tenants by name", func() {
				filters := &dao.TenantFilters{Name: tenant1.Name}
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, filters)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(1))
				Expect((*tenants)[0].Name).To(Equal(tenant1.Name))
			})
			It("should filter tenants by alias", func() {
				filters := &dao.TenantFilters{Alias: *tenantWithAlias.Alias}
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, filters)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(1))
				Expect((*tenants)[0].Name).To(Equal(tenantWithAlias.Name))
				Expect(*(*tenants)[0].Alias).To(Equal(*tenantWithAlias.Alias))
			})
			It("should filter tenants by both name and alias", func() {
				filters := &dao.TenantFilters{Name: tenantWithAlias.Name, Alias: *tenantWithAlias.Alias}
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, filters)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(1))
				Expect((*tenants)[0].Name).To(Equal(tenantWithAlias.Name))
			})
			It("should return empty list when name filter does not match", func() {
				filters := &dao.TenantFilters{Name: "non-existent-tenant"}
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, filters)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(0))
			})
			It("should return empty list when alias filter does not match", func() {
				filters := &dao.TenantFilters{Alias: "non-existent-alias"}
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, filters)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(0))
			})
			It("should return tenant when name matches but alias does not", func() {
				filters := &dao.TenantFilters{Name: tenant1.Name, Alias: "wrong-alias"}
				tenants, daoError := tenantDao.GetAll(context.Background(), nil, filters)
				Expect(daoError).To(BeNil())
				Expect(len(*tenants)).To(Equal(1))
			})
		})
	})
})
