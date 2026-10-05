/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
)

type ServiceAccountsCache interface {
	GetTenantServiceAccountsForRealm(ctx context.Context, realm string, fetchFunc func() (*TenantServiceAccounts, error)) (*TenantServiceAccounts, error)
	SetTenantServiceAccountsForRealm(ctx context.Context, realm string, accounts TenantServiceAccounts) error
	DeleteTenantServiceAccountsFromRealm(ctx context.Context, realm string) error
}

type TenantServiceAccounts struct {
	ServiceAccounts []models.Client
}

type ServiceAccountsCacheManager struct {
	KeyBuilder *KeyBuilder
	Manager    CacheManagerInterface[TenantServiceAccounts]
}

func NewServiceAccountsCacheManager() *ServiceAccountsCacheManager {
	return &ServiceAccountsCacheManager{KeyBuilder: NewKeyBuilder(), Manager: NewCacheManager[TenantServiceAccounts]()}
}

func (acm *ServiceAccountsCacheManager) GetTenantServiceAccountsForRealm(ctx context.Context, realm string, fetchFunc func() (*TenantServiceAccounts, error)) (*TenantServiceAccounts, error) {
	cacheKey := acm.KeyBuilder.BuildTenantServiceAccountsKey(realm)
	return acm.Manager.GetOrFetch(ctx, cacheKey, fetchFunc)
}

func (acm *ServiceAccountsCacheManager) SetTenantServiceAccountsForRealm(ctx context.Context, realm string, accounts TenantServiceAccounts) error {
	return acm.Manager.Set(ctx, acm.KeyBuilder.BuildTenantServiceAccountsKey(realm), accounts)
}

func (acm *ServiceAccountsCacheManager) DeleteTenantServiceAccountsFromRealm(ctx context.Context, realm string) error {
	return acm.Manager.Delete(ctx, acm.KeyBuilder.BuildTenantServiceAccountsKey(realm))
}
