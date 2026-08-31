/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"testing"

	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	manager := NewCacheManager[iamModels.Idp]()
	assert.NotNil(t, manager)
}

func TestNewIdpCacheManager(t *testing.T) {
	manager := NewIdpCacheManager()
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.Manager)
}

func TestNewIdpMappersCacheManager(t *testing.T) {
	manager := NewIdpMappersCacheManager()
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.Manager)
}

func TestKeyBuilder_BuildIdpKey(t *testing.T) {
	builder := NewKeyBuilder()

	key := builder.BuildIdpKey("test-realm", "test-idp")
	assert.Equal(t, "idp-test-realm-test-idp", key)
}

func TestKeyBuilder_BuildIdpMappersKey(t *testing.T) {
	builder := NewKeyBuilder()

	key := builder.BuildIdpMappersKey("test-realm", "test-idp")
	assert.Equal(t, "idp-mappers-test-realm-test-idp", key)
}

func TestKeyBuilder_BuildClientKey(t *testing.T) {
	builder := NewKeyBuilder()
	key := builder.BuildClientKey("test-realm", "test-client")
	assert.Equal(t, "client-test-realm-test-client", key)
}

func TestInitCacheStores(t *testing.T) {
	InitCacheStores(30)
	assert.True(t, true, "InitCacheStores should not panic")
}

func TestNewIdpCacheWrapper(t *testing.T) {
	wrapper := NewIdpCacheWrapper()
	assert.NotNil(t, wrapper)
	assert.NotNil(t, wrapper.IdpCache)
	assert.NotNil(t, wrapper.IdpMappersCache)
	assert.NotNil(t, wrapper.IdpsListCache)
}

func TestNewIdpsListCacheManager(t *testing.T) {
	manager := NewIdpsListCacheManager()
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.Manager)
}

func TestNewClientCacheManager(t *testing.T) {
	manager := NewClientCacheManager()
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.Manager)
}
