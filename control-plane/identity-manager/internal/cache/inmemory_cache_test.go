/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	cache_testing "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func beforeEach(t *testing.T) {
	cache_testing.EnableCache()
	InitCacheStores(30)
}

func TestInMemoryIdpCache_SetGetDelete(t *testing.T) {
	beforeEach(t)
	subject := NewIdpCacheManager()
	err := subject.SetIdpForRealm(context.Background(), "realm", "idp", &iamModels.Idp{Alias: ptr.To("idp")})
	assert.NoError(t, err)
	idp, err := subject.GetIdpForRealm(context.Background(), "realm", "idp", func() (*iamModels.Idp, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "idp", *idp.Alias)
	err = subject.DeleteIdpFromRealm(context.Background(), "realm", "idp")
	assert.NoError(t, err)
	idp, err = subject.GetIdpForRealm(context.Background(), "realm", "idp", func() (*iamModels.Idp, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Nil(t, idp)
}

func TestInMemoryIdpMappersCache_SetGetDelete(t *testing.T) {
	beforeEach(t)
	subject := NewIdpMappersCacheManager()
	err := subject.SetIdpMappersForRealm(context.Background(), "realm", "idp", IdpMappers{Mappers: []*iamModels.IdpMapper{{Name: ptr.To("idp")}}})
	assert.NoError(t, err)
	idpMappers, err := subject.GetIdpMappersForRealm(context.Background(), "realm", "idp", func() (*IdpMappers, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "idp", *idpMappers.Mappers[0].Name)
	err = subject.DeleteIdpMappersFromRealm(context.Background(), "realm", "idp")
	assert.NoError(t, err)
	idpMappers, err = subject.GetIdpMappersForRealm(context.Background(), "realm", "idp", func() (*IdpMappers, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Nil(t, idpMappers)
}

func TestInMemoryIdpsListCache_SetGetDelete(t *testing.T) {
	beforeEach(t)
	subject := NewIdpsListCacheManager()
	err := subject.SetIdpsListForRealm(context.Background(), "realm", IdpsList{Idps: []*iamModels.Idp{{Alias: ptr.To("idp")}}})
	assert.NoError(t, err)
	idps, err := subject.GetIdpsListForRealm(context.Background(), "realm", func() (*IdpsList, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "idp", *idps.Idps[0].Alias)
	err = subject.DeleteIdpsListFromRealm(context.Background(), "realm")
	assert.NoError(t, err)
	idps, err = subject.GetIdpsListForRealm(context.Background(), "realm", func() (*IdpsList, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Nil(t, idps)
}

func TestInMemoryClientCache_SetGetDelete(t *testing.T) {
	beforeEach(t)
	subject := NewClientCacheManager()
	err := subject.SetClientForRealm(context.Background(), "realm1", "client", iamModels.Client{ClientID: ptr.To("client")})
	assert.NoError(t, err)
	client, err := subject.GetClientForRealm(context.Background(), "realm1", "client", func() (*iamModels.Client, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "client", *client.ClientID)
	err = subject.DeleteClientFromRealm(context.Background(), "realm1", "client")
	assert.NoError(t, err)
	client, err = subject.GetClientForRealm(context.Background(), "realm1", "client", func() (*iamModels.Client, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Nil(t, client)
}

func TestInMemoryServiceAccountCache_SetGetDelete(t *testing.T) {
	beforeEach(t)
	subject := NewServiceAccountsCacheManager()
	err := subject.SetTenantServiceAccountsForRealm(context.Background(), "realm", TenantServiceAccounts{ServiceAccounts: []models.Client{models.Client{ID: ptr.To("apps")}}})
	assert.NoError(t, err)
	accounts, err := subject.GetTenantServiceAccountsForRealm(context.Background(), "realm", func() (*TenantServiceAccounts, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "apps", *accounts.ServiceAccounts[0].ID)
	err = subject.DeleteTenantServiceAccountsFromRealm(context.Background(), "realm")
	assert.NoError(t, err)
	accounts, err = subject.GetTenantServiceAccountsForRealm(context.Background(), "realm", func() (*TenantServiceAccounts, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	assert.Nil(t, accounts)
}
