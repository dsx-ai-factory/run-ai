/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"
	"fmt"
	"time"

	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

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
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache miss for key: %s, error: %v", cacheKey, err)
		return nil, err
	}

	if item != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache hit for key: %s", cacheKey)
	}

	return item, nil
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
	if item, err := m.Get(ctx, key); err == nil && item != nil {
		return item, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache miss for key: %s, fetching from backend", key)

	item, err := fetchFunc()
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

func (kb *KeyBuilder) BuildIdpKey(realm, idpAlias string) string {
	return fmt.Sprintf("idp-%s-%s", realm, idpAlias)
}

func (kb *KeyBuilder) BuildIdpMappersKey(realm, idpAlias string) string {
	return fmt.Sprintf("idp-mappers-%s-%s", realm, idpAlias)
}

func (kb *KeyBuilder) BuildIdpsListKey(realm string) string {
	return fmt.Sprintf("idps-list-%s", realm)
}

func (kb *KeyBuilder) BuildClientKey(realm, clientId string) string {
	return fmt.Sprintf("client-%s-%s", realm, clientId)
}

func (kb *KeyBuilder) BuildTenantServiceAccountsKey(realm string) string {
	return fmt.Sprintf("tenant-service-accounts-%s", realm)
}

func InitCacheStores(cacheExpirationMinutes int) {
	cacheExpirationTime := time.Duration(cacheExpirationMinutes) * time.Minute
	if err := cache.RegisterStore[iamModels.Idp](options.Ttl(cacheExpirationTime), options.MaxRecordSizeKB(10)); err != nil {
		logging.Errorf("failed to register store for idp. err: %v", err)
	}
	if err := cache.RegisterStore[IdpMappers](options.Ttl(cacheExpirationTime), options.MaxRecordSizeKB(10)); err != nil {
		logging.Errorf("failed to register store for idp mapper. err: %v", err)
	}
	if err := cache.RegisterStore[IdpsList](options.Ttl(cacheExpirationTime), options.MaxRecordSizeKB(50)); err != nil {
		logging.Errorf("failed to register store for idp. err: %v", err)
	}
	if err := cache.RegisterStore[iamModels.Client](options.Ttl(cacheExpirationTime), options.MaxRecordSizeKB(10)); err != nil {
		logging.Errorf("failed to register store for client. err: %v", err)
	}
	if err := cache.RegisterStore[TenantServiceAccounts](options.Ttl(cacheExpirationTime), options.MaxRecordSizeKB(200), options.MaxStoreSizeMB(200)); err != nil {
		logging.Errorf("failed to register store for tenant service accounts. err: %v", err)
	}
}
