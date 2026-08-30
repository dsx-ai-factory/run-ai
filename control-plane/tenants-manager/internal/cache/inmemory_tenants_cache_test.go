package cache

import (
	"context"
	"testing"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	cache_testing "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func beforeEach() {
	cache_testing.EnableCache()
	InitCacheStores()
}

func TestInMemoryTenantsCache_SetAndGet(t *testing.T) {
	beforeEach()
	subject := NewTenantCacheManager()
	err := subject.SetTenant(context.Background(), 1001, models.Tenant{Model: gorm.Model{ID: 1001}})
	assert.NoError(t, err)
	tenant, err := subject.GetTenant(context.Background(), 1001, func() (*models.Tenant, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1001, int(tenant.ID))
}

func TestInMemoryTenantsCache_GetNotExists_Fetch(t *testing.T) {
	beforeEach()
	subject := NewTenantCacheManager()
	tenant, err := subject.GetTenant(context.Background(), 1002, func() (*models.Tenant, error) {
		return &models.Tenant{Model: gorm.Model{ID: 1002}}, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1002, int(tenant.ID))
}

func TestInMemoryTenantsCache_SetAndDelete(t *testing.T) {
	beforeEach()
	subject := NewTenantCacheManager()
	err := subject.SetTenant(context.Background(), 1003, models.Tenant{Model: gorm.Model{ID: 1003}})
	assert.NoError(t, err)
	tenant, err := subject.GetTenant(context.Background(), 1003, func() (*models.Tenant, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1003, int(tenant.ID))
	err = subject.DeleteTenant(context.Background(), 1003)
	assert.NoError(t, err)
	tenant, err = subject.GetTenant(context.Background(), 1003, func() (*models.Tenant, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Nil(t, tenant)
}
