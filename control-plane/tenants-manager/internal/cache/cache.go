package cache

import (
	"context"
	"fmt"
	"time"

	tm_config "github.com/run-ai/runai/control-plane/tenants-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func InitCacheStores() {
	tenantCacheExpirationTime := time.Duration(config.Get[*tm_config.Tenant]().CacheTtlMinutes) * time.Minute
	if err := cache.RegisterStore[models.Tenant](options.Ttl(tenantCacheExpirationTime)); err != nil {
		logging.Errorf("Failed to register tenant cache: %v", err)
	}
}

//go:generate mockery
type CacheManagerInterface[T any] interface {
	Get(ctx context.Context, cacheKey string) (*T, error)
	Set(ctx context.Context, cacheKey string, value T) error
	Delete(ctx context.Context, cacheKey string) error
	GetOrFetch(ctx context.Context, key string, fetchFunc func() (*T, error)) (*T, error)
}

type CacheManager[T any] struct{}

func NewCacheManager[T any]() *CacheManager[T] {
	return &CacheManager[T]{}
}

func (m *CacheManager[T]) Get(ctx context.Context, cacheKey string) (*T, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Getting cached item for key: %s", cacheKey)

	item, err := cache.Get[T](ctx, cacheKey)

	if item != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache hit for key: %s", cacheKey)
		return item, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache miss for key: %s, error: %v", cacheKey, err)
	return nil, err
}

func (m *CacheManager[T]) Set(ctx context.Context, cacheKey string, value T) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Setting cached item for key: %s", cacheKey)

	return cache.Set(ctx, cacheKey, value)
}

func (m *CacheManager[T]) Delete(ctx context.Context, cacheKey string) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Deleting cached item for key: %s", cacheKey)

	return cache.Del[T](ctx, cacheKey)
}

func (m *CacheManager[T]) GetOrFetch(ctx context.Context, key string, fetchFunc func() (*T, error)) (*T, error) {
	item, err := m.Get(ctx, key)
	if item != nil {
		return item, nil
	}
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Cache error on get for key: %s, error: %v", key, err)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache miss for key: %s, fetching from backend", key)

	item, err = fetchFunc()
	if err != nil {
		return nil, err
	}

	if item != nil {
		if err := m.Set(ctx, key, *item); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to store item in cache for key: %s, error: %v", key, err)
		}
	}

	return item, nil
}

type KeyBuilder struct{}

func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{}
}

func (kb *KeyBuilder) BuildTenantKey(tenantId int32) string {
	return fmt.Sprintf("tenant-%d", tenantId)
}
