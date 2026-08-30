package cache

import (
	"context"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
)

//go:generate mockery
type TenantsCache interface {
	SetTenant(ctx context.Context, tenantId int32, tenant models.Tenant) error
	GetTenant(ctx context.Context, tenantId int32, fetchFunc func() (*models.Tenant, error)) (*models.Tenant, error)
	DeleteTenant(ctx context.Context, tenantId int32) error
}

type TenantCacheManager struct {
	KeyBuilder *KeyBuilder
	Tenant     CacheManagerInterface[models.Tenant]
}

func NewTenantCacheManager() *TenantCacheManager {
	return &TenantCacheManager{
		KeyBuilder: NewKeyBuilder(),
		Tenant:     NewCacheManager[models.Tenant](),
	}
}

func (im *TenantCacheManager) SetTenant(ctx context.Context, tenantId int32, tenant models.Tenant) error {
	key := im.KeyBuilder.BuildTenantKey(tenantId)
	return im.Tenant.Set(ctx, key, tenant)
}

func (im *TenantCacheManager) GetTenant(ctx context.Context, tenantId int32, fetchFunc func() (*models.Tenant, error)) (*models.Tenant, error) {
	key := im.KeyBuilder.BuildTenantKey(tenantId)
	return im.Tenant.GetOrFetch(ctx, key, fetchFunc)
}

func (im *TenantCacheManager) DeleteTenant(ctx context.Context, tenantId int32) error {
	key := im.KeyBuilder.BuildTenantKey(tenantId)
	return im.Tenant.Delete(ctx, key)
}
