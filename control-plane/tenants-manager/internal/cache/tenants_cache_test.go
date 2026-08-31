/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	cache_testing "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func initSubject(t *testing.T) (*TenantCacheManager, *cache_testing.MockCacher) {
	cacheMock := cache_testing.SetMock(t)
	return NewTenantCacheManager(), cacheMock
}

func TestSetTenant(t *testing.T) {
	subject, cacheMock := initSubject(t)
	cacheMock.On("Set", mock.Anything, cache_testing.KeyFromType[models.Tenant](), subject.KeyBuilder.BuildTenantKey(1001), mock.Anything).Return(nil)
	subject.SetTenant(context.Background(), 1001, models.Tenant{Model: gorm.Model{ID: 1001}})
}

func TestGetTenant_Exists(t *testing.T) {
	subject, cacheMock := initSubject(t)
	cacheMock.On("Get", mock.Anything, cache_testing.KeyFromType[models.Tenant](), subject.KeyBuilder.BuildTenantKey(1001)).Return(&models.Tenant{Model: gorm.Model{ID: 1001}}, nil)
	tenant, err := subject.GetTenant(context.Background(), 1001, func() (*models.Tenant, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1001, int(tenant.ID))
}

func TestGetTenant_NotExists_Fetch(t *testing.T) {
	subject, cacheMock := initSubject(t)
	cacheMock.On("Get", mock.Anything, cache_testing.KeyFromType[models.Tenant](), subject.KeyBuilder.BuildTenantKey(1001)).Return(nil, nil)
	cacheMock.On("Set", mock.Anything, cache_testing.KeyFromType[models.Tenant](), subject.KeyBuilder.BuildTenantKey(1001), mock.Anything).Return(nil)
	tenant, err := subject.GetTenant(context.Background(), 1001, func() (*models.Tenant, error) {
		return &models.Tenant{Model: gorm.Model{ID: 1001}}, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1001, int(tenant.ID))
}

func TestGetTenant_NotExists_CacheGetError(t *testing.T) {
	subject, cacheMock := initSubject(t)
	cacheMock.On("Get", mock.Anything, cache_testing.KeyFromType[models.Tenant](), subject.KeyBuilder.BuildTenantKey(1001)).Return(nil, errors.New("cache get error"))
	cacheMock.On("Set", mock.Anything, cache_testing.KeyFromType[models.Tenant](), subject.KeyBuilder.BuildTenantKey(1001), mock.Anything).Return(nil)
	tenant, err := subject.GetTenant(context.Background(), 1001, func() (*models.Tenant, error) {
		return &models.Tenant{Model: gorm.Model{ID: 1001}}, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1001, int(tenant.ID))
}

func TestGetTenant_NotExists_CacheSetError(t *testing.T) {
	subject, cacheMock := initSubject(t)
	cacheMock.On("Get", mock.Anything, cache_testing.KeyFromType[models.Tenant](), subject.KeyBuilder.BuildTenantKey(1001)).Return(nil, nil)
	cacheMock.On("Set", mock.Anything, cache_testing.KeyFromType[models.Tenant](), subject.KeyBuilder.BuildTenantKey(1001), mock.Anything).Return(errors.New("cache set error"))
	tenant, err := subject.GetTenant(context.Background(), 1001, func() (*models.Tenant, error) {
		return &models.Tenant{Model: gorm.Model{ID: 1001}}, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1001, int(tenant.ID))
}

func TestGetTenant_NotExists_FetchError(t *testing.T) {
	subject, cacheMock := initSubject(t)
	cacheMock.On("Get", mock.Anything, cache_testing.KeyFromType[models.Tenant](), subject.KeyBuilder.BuildTenantKey(1001)).Return(nil, nil)
	tenant, err := subject.GetTenant(context.Background(), 1001, func() (*models.Tenant, error) {
		return nil, errors.New("fetch error")
	})
	assert.Error(t, err)
	assert.Nil(t, tenant)
}
