/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	cache_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var storeKey = reflect.TypeOf((*testType)(nil)).Elem()

func TestRegisterStore(t *testing.T) {
	config.Set(&cache_config.Cache{
		ServiceName: "test",
	})
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)

	cacheClientMock.On("Store", mock.Anything).Return(nil)
	expectedStore := &options.Store{
		Name:    "testStore",
		Options: options.NewStoreOptions(options.Ttl(time.Hour), options.MaxRecordSizeKB(100), options.MaxStoreSizeMB(100)),
	}
	storeRegistryMock.On("Get", mock.Anything).Return(nil)
	storeRegistryMock.On("Set", storeKey.String(), expectedStore).Return(nil)

	assert.NoError(t, cacher.RegisterStore(storeKey, "testStore", options.Ttl(time.Hour), options.MaxRecordSizeKB(100), options.MaxStoreSizeMB(100)))
}

func TestRegisterStore_ReadOnly_Error(t *testing.T) {
	cacher, _, _, _ := initCacherTest(t)
	assert.Error(t, cacher.RegisterStore(storeKey, "storeName", options.ReadOnly()))
}

func TestRegisterSharedStore(t *testing.T) {
	config.Set(&cache_config.Cache{
		ServiceName: "test",
	})
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)

	cacheClientMock.On("Store", mock.Anything).Return(nil)

	expectedStore := &options.Store{
		Name:    "shared-name-cache_testType",
		Options: options.NewStoreOptions(options.Ttl(time.Hour), options.MaxRecordSizeKB(100), options.MaxStoreSizeMB(100), options.ReadOnly()),
	}
	storeRegistryMock.On("Get", mock.Anything).Return(nil)
	storeRegistryMock.On("Set", storeKey.String(), expectedStore).Return(nil)

	storeKey := getStoreKeyByType[testType]()
	assert.NoError(t, cacher.RegisterSharedStore(getStoreKeyByType[testType](), buildSharedStoreName("name", storeKey), options.Ttl(time.Hour), options.MaxRecordSizeKB(100), options.MaxStoreSizeMB(100), options.ReadOnly()))
}

func TestRegisterSharedStore_InvalidName_Error(t *testing.T) {
	cacher, _, _, _ := initCacherTest(t)
	assert.Error(t, cacher.RegisterSharedStore(getStoreKeyByType[testType](), "name.invalid"))
}

func TestClearStore_StoreNotRegistered(t *testing.T) {
	cacher, _, storeRegistryMock, _ := initCacherTest(t)

	storeRegistryMock.On("Get", mock.Anything).Return(nil)

	err := cacher.ClearStore(context.Background(), getStoreKeyByType[testType]())
	assert.Equal(t, "store not found", err.Error())
}

func TestClearStore_ReadOnlyStoreNotCleared(t *testing.T) {
	cacher, _, storeRegistryMock, _ := initCacherTest(t)

	storeRegistryMock.On("Get", mock.Anything).Return(&options.Store{
		Name: "conf-1",
		Options: &options.StoreOptions{
			ReadOnly: true,
		},
	})

	err := cacher.ClearStore(context.Background(), getStoreKeyByType[testType]())
	assert.Equal(t, "cannot clear read only store", err.Error())
}

func TestClearStore(t *testing.T) {
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)

	storeRegistryMock.On("Get", mock.Anything).Return(&options.Store{
		Name:    "conf-1",
		Options: &options.StoreOptions{},
	})

	cacheClientMock.On("ClearStore", mock.Anything, "conf-1").Return(nil)

	err := cacher.ClearStore(context.Background(), getStoreKeyByType[testType]())
	assert.NoError(t, err)
}

func TestClearStore_Error(t *testing.T) {
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)

	storeRegistryMock.On("Get", mock.Anything).Return(&options.Store{
		Name:    "conf-1",
		Options: &options.StoreOptions{},
	})

	cacheClientMock.On("ClearStore", mock.Anything, "conf-1").Return(fmt.Errorf("error"))

	err := cacher.ClearStore(context.Background(), getStoreKeyByType[testType]())
	assert.Equal(t, "failed to clear store conf-1: error", err.Error())
}

func TestClearCache(t *testing.T) {
	cacher, cacheClientMock, storeRegistryMock, defaultStore := initCacherTest(t)

	storeRegistryMock.On("List").Return([]*options.Store{
		{
			Name:    "conf-1",
			Options: &options.StoreOptions{},
		},
		{
			Name:    "conf-2",
			Options: &options.StoreOptions{},
		},
		defaultStore,
	})

	cacheClientMock.On("ClearStore", mock.Anything, "conf-1").Return(nil)
	cacheClientMock.On("ClearStore", mock.Anything, "conf-2").Return(nil)
	cacheClientMock.On("ClearStore", mock.Anything, "default-store").Return(nil)

	err := cacher.ClearCache(context.Background())
	assert.NoError(t, err)
}

func TestClearCache_Error(t *testing.T) {
	cacher, cacheClientMock, storeRegistryMock, defaultStore := initCacherTest(t)

	storeRegistryMock.On("List").Return([]*options.Store{
		{
			Name:    "conf-1",
			Options: &options.StoreOptions{},
		},
		defaultStore,
	})

	cacheClientMock.On("ClearStore", mock.Anything, "conf-1").Return(fmt.Errorf("error"))
	cacheClientMock.On("ClearStore", mock.Anything, "default-store").Return(nil)

	err := cacher.ClearCache(context.Background())
	assert.Equal(t, "failed to clear stores: [failed to clear store conf-1: error]", err.Error())
}

func TestClearCache_ReadOnlyStoreNotCleared(t *testing.T) {
	cacher, cacheClientMock, storeRegistryMock, defaultStore := initCacherTest(t)

	storeRegistryMock.On("List").Return([]*options.Store{
		{
			Name: "conf-1",
			Options: &options.StoreOptions{
				ReadOnly: true,
			},
		},
		defaultStore,
	})

	cacheClientMock.On("ClearStore", mock.Anything, "default-store").Return(nil)
	err := cacher.ClearCache(context.Background())
	assert.NoError(t, err)
	cacheClientMock.AssertNotCalled(t, "ClearStore", mock.Anything, "conf-1")
}

func TestGetStoreKeyByType(t *testing.T) {
	assert.Equal(t, storeKey, getStoreKeyByType[testType]())
}

func TestGet_NoValue(t *testing.T) {
	ctx := context.Background()
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)

	storeRegistryMock.On("GetOrDefault", storeKey.String()).Return(&options.Store{
		Name:    "store-name",
		Options: options.NewStoreOptions()})
	cacheClientMock.On("Get", ctx, "store-name", "key").Return(nil, nil)

	val, err := cacher.Get(ctx, storeKey, "key")
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestGet_SucceedWithValue(t *testing.T) {
	ctx := context.Background()
	v, _ := json.Marshal("some-value")
	stringStoreKey := reflect.TypeOf((*string)(nil)).Elem()
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)
	storeRegistryMock.On("GetOrDefault", stringStoreKey.String()).Return(&options.Store{
		Name:    "store-name",
		Options: options.NewStoreOptions()})
	cacheClientMock.On("Get", ctx, "store-name", "key").Return(v, nil)

	val, err := cacher.Get(ctx, stringStoreKey, "key")
	assert.NoError(t, err)
	assert.Equal(t, "some-value", *(val.(*string)))
}

func TestGet_Failed(t *testing.T) {
	ctx := context.Background()
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)
	storeRegistryMock.On("GetOrDefault", storeKey.String()).Return(&options.Store{
		Name:    "store-name",
		Options: options.NewStoreOptions()})
	cacheClientMock.On("Get", ctx, "store-name", "key").Return(nil, fmt.Errorf("failed"))

	val, err := cacher.Get(ctx, storeKey, "key")
	assert.EqualError(t, err, "failed")
	assert.Nil(t, val)
}

func TestSet(t *testing.T) {
	ctx := context.Background()
	v, _ := json.Marshal("some-value")
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)
	store := &options.Store{
		Name:    "store-name",
		Options: options.NewStoreOptions()}
	store.Options.MaxRecordSizeKB = 1
	storeRegistryMock.On("GetOrDefault", storeKey.String()).Return(store)
	cacheClientMock.On("Set", ctx, "store-name", "key", v).Return(nil)

	err := cacher.Set(ctx, storeKey, "key", "some-value")
	assert.NoError(t, err)
}

func TestSet_Failed(t *testing.T) {
	ctx := context.Background()
	v, _ := json.Marshal("some-value")
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)
	store := &options.Store{
		Name:    "store-name",
		Options: options.NewStoreOptions()}
	store.Options.MaxRecordSizeKB = 1
	storeRegistryMock.On("GetOrDefault", storeKey.String()).Return(store)
	cacheClientMock.On("Set", ctx, "store-name", "key", v).Return(fmt.Errorf("failed"))

	err := cacher.Set(ctx, storeKey, "key", "some-value")
	assert.Error(t, err)
}

func TestSet_SetToSharedReadOnlyStore(t *testing.T) {
	storeName := "shared-store"
	opt := options.NewStoreOptions()
	opt.ReadOnly = true
	cacher, _, storeRegistryMock, _ := initCacherTest(t)
	storeRegistryMock.On("GetOrDefault", storeKey.String()).Return(&options.Store{
		Name:    storeName,
		Options: opt})

	err := cacher.Set(context.Background(), storeKey, "key", "some-value")
	assert.EqualError(t, err, "store is read only")
}

func TestDel(t *testing.T) {
	ctx := context.Background()
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)
	storeRegistryMock.On("GetOrDefault", storeKey.String()).Return(&options.Store{
		Name:    "store-name",
		Options: options.NewStoreOptions()})
	cacheClientMock.On("Delete", ctx, "store-name", "key").Return(nil)

	err := cacher.Del(ctx, storeKey, "key")
	assert.NoError(t, err)
}

func TestDel_Failed(t *testing.T) {
	ctx := context.Background()
	cacher, cacheClientMock, storeRegistryMock, _ := initCacherTest(t)
	storeRegistryMock.On("GetOrDefault", storeKey.String()).Return(&options.Store{
		Name:    "store-name",
		Options: options.NewStoreOptions()})
	cacheClientMock.On("Delete", ctx, "store-name", "key").Return(fmt.Errorf("delete failed"))

	err := cacher.Del(ctx, storeKey, "key")
	assert.EqualError(t, err, "delete failed")
}

func TestDel_SharedReadOnlyStore(t *testing.T) {
	storeName := "shared-store"
	opt := options.NewStoreOptions()
	opt.ReadOnly = true
	cacher, _, storeRegistryMock, _ := initCacherTest(t)
	storeRegistryMock.On("GetOrDefault", storeKey.String()).Return(&options.Store{
		Name:    storeName,
		Options: opt})

	err := cacher.Del(context.Background(), storeKey, "key")
	assert.EqualError(t, err, "store is read only")
}

func TestValidateMaxRecordSize(t *testing.T) {
	err := validateMaxRecordSize(1, "key", []byte("value"))
	assert.NoError(t, err)
}

func TestValidateMaxRecordSize_Error(t *testing.T) {
	value := make([]byte, 1000) // 1KB
	err := validateMaxRecordSize(1, "key", value)
	assert.Equal(t, "record size 1.003 KB exceeds max allowed size of 1 KB", err.Error())
}

func initCacherTest(t *testing.T) (Cacher, *mocks.CacheClienter, *mocks.StoreRegistryer, *options.Store) {
	cacheClientMock := mocks.NewCacheClienter(t)
	storeRegistryMock := mocks.NewStoreRegistryer(t)
	defaultStore := &options.Store{
		Name:    "default-store",
		Options: options.NewStoreOptions(),
	}
	cacher := newCacher(cacheClientMock, storeRegistryMock)

	return cacher, cacheClientMock, storeRegistryMock, defaultStore
}
