/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package keydb

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/validator"

	"github.com/go-redis/redismock/v9"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	keydbClient := NewClient()
	require.NotNil(t, keydbClient)
}

func TestStore(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)
}

func TestStore_Failed(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetErr(errors.New("error"))

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.Equal(t, "failed to get HLEN for store <testStore>: error", err.Error())
}

func TestStore_StoreAlreadyStored(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)

	v1, ok := keydbClient.storeValidators.Load(storeConfig.Name)
	require.True(t, ok)

	err = keydbClient.Store(storeConfig)
	require.NoError(t, err)

	v2, ok := keydbClient.storeValidators.Load(storeConfig.Name)
	require.True(t, ok)

	require.Equal(t, v1, v2)
}

func TestSet(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1, // HExpire mock not supported by redismock so set to -1, which means no expiration
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)
	mockRW.ExpectHSet(storeConfig.Name, "key", []byte("value")).SetVal(1)

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)

	err = keydbClient.Set(context.Background(), storeConfig.Name, "key", []byte("value"))
	require.NoError(t, err)
}

func TestSet_Failed(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)
	mockRW.ExpectHSet(storeConfig.Name, "key", []byte("value")).SetErr(errors.New("error"))

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)

	err = keydbClient.Set(context.Background(), storeConfig.Name, "key", []byte("value"))
	require.Equal(t, "failed to set key <testStore,key> in cache: error", err.Error())
}

func TestSet_FailedToGetStoreHLen(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	maxKeys := calculateStoreMaxKeys(storeConfig.Options)

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(maxKeys + 1)         // first call- choose a number greater than max store size for newValidator
	mockRW.ExpectHLen(storeConfig.Name).SetErr(errors.New("error")) // second call- from set

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)

	err = keydbClient.Set(context.Background(), storeConfig.Name, "key", []byte("value"))
	require.Equal(t, "failed to get HLEN for store <testStore>: error", err.Error())
}

func TestSetIfAbsent(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)
	mockRW.ExpectHSetNX(storeConfig.Name, "key", []byte("value")).SetVal(true)

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)
	acquired, err := keydbClient.SetIfAbsent(context.Background(), storeConfig.Name, "key", []byte("value"), nil)
	require.NoError(t, err)
	require.True(t, acquired)
}

func TestSetIfAbsent_Exists(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)
	mockRW.ExpectHSetNX(storeConfig.Name, "key", []byte("value")).SetVal(false)

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)
	acquired, err := keydbClient.SetIfAbsent(context.Background(), storeConfig.Name, "key", []byte("value"), nil)
	require.NoError(t, err)
	require.False(t, acquired)
}
func TestSetIfAbsent_Failed(t *testing.T) {

	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)
	mockRW.ExpectHSetNX(storeConfig.Name, "key", []byte("value")).SetErr(errors.New("error"))

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)
	acquired, err := keydbClient.SetIfAbsent(context.Background(), storeConfig.Name, "key", []byte("value"), nil)
	require.Error(t, err)
	require.False(t, acquired)
}

func TestSetIfAbsent_FailedToGetStoreHLen(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}
	maxKeys := calculateStoreMaxKeys(storeConfig.Options)

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(maxKeys + 1)         // first call- choose a number greater than max store size for newValidator
	mockRW.ExpectHLen(storeConfig.Name).SetErr(errors.New("error")) // second call- from set

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)
	acquired, err := keydbClient.SetIfAbsent(context.Background(), storeConfig.Name, "key", []byte("value"), nil)
	require.Equal(t, "failed to get HLEN for store <testStore>: error", err.Error())
	require.False(t, acquired)
}

func TestGet(t *testing.T) {
	storeConfig := getMockStoreConfig()

	clientR, mockR := redismock.NewClientMock()
	mockR.ExpectHGet(storeConfig.Name, "key").SetVal("value")

	keydbClient := &Client{
		clientR: clientR,
	}

	val, err := keydbClient.Get(context.Background(), storeConfig.Name, "key")
	require.NoError(t, err)
	require.Equal(t, []byte("value"), val)
}

func TestGet_KeyNotFound(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"

	clientR, mockR := redismock.NewClientMock()
	mockR.ExpectHGet(storeConfig.Name, key).RedisNil()

	keydbClient := &Client{
		clientR: clientR,
	}

	val, err := keydbClient.Get(ctx, storeConfig.Name, key)
	require.NoError(t, err)
	require.Nil(t, val)
}

func TestGet_Failed(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"

	clientR, mockR := redismock.NewClientMock()
	mockR.ExpectHGet(storeConfig.Name, key).SetErr(errors.New("error"))

	keydbClient := &Client{
		clientR: clientR,
	}

	val, err := keydbClient.Get(ctx, storeConfig.Name, key)
	require.Equal(t, "failed to get key <testStore,key1> from cache: error", err.Error())
	require.Nil(t, val)
}

func TestDelete(t *testing.T) {
	storeConfig := getMockStoreConfig()

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHDel(storeConfig.Name, "key").SetVal(1)

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Delete(context.Background(), storeConfig.Name, "key")
	require.NoError(t, err)
}

func TestDelete_Failed(t *testing.T) {
	storeConfig := getMockStoreConfig()

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHDel(storeConfig.Name, "key").SetErr(errors.New("error"))

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Delete(context.Background(), storeConfig.Name, "key")
	require.Equal(t, "failed to delete key <testStore,key> from cache: error", err.Error())
}

func TestClearStore(t *testing.T) {
	storeConfig := getMockStoreConfig()

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectDel(storeConfig.Name).SetVal(1)

	keydbClient := &Client{
		clientRW:        clientRW,
		storeValidators: sync.Map{},
	}
	v, err := validator.NewStoreLenValidator("storeName", 100, func(ctx context.Context) (int64, error) {
		return 0, nil
	})
	require.NoError(t, err)
	keydbClient.storeValidators.Store(storeConfig.Name, v)

	err = keydbClient.ClearStore(context.Background(), storeConfig.Name)
	require.NoError(t, err)
}

func TestClearStore_Failed(t *testing.T) {
	storeConfig := getMockStoreConfig()

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectDel(storeConfig.Name).SetErr(errors.New("error"))

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.ClearStore(context.Background(), storeConfig.Name)
	require.Equal(t, "failed to clear store <testStore>: error", err.Error())
}

func TestValidateStoreNotFull(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)

	err = keydbClient.validateStoreNotFull(context.Background(), storeConfig.Name)
	require.NoError(t, err)
}

func TestValidateStoreNotFull_NotFoundInSyncMap(t *testing.T) {
	keydbClient := NewClient()

	err := keydbClient.validateStoreNotFull(context.Background(), "storeName")
	require.Equal(t, "no validator found for store <storeName>", err.Error())
}

func TestValidateStoreNotFull_StoreFull(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             -1,
			MaxStoreSizeMB:  1,
			MaxRecordSizeKB: 100,
		},
	}

	maxStoreKeys := calculateStoreMaxKeys(storeConfig.Options)

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(maxStoreKeys + 1) // first call - called in the constructor
	mockRW.ExpectHLen(storeConfig.Name).SetVal(maxStoreKeys + 1) // second call - called in validateStoreNotFull

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)

	err = keydbClient.validateStoreNotFull(context.Background(), storeConfig.Name)
	require.Equal(t, "store <testStore> exceeded max size (20 keys)", err.Error())
}

func TestCheckRedisHealthy(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	clientR, mockR := redismock.NewClientMock()

	mockRW.ExpectSet("Validation", "Redis connected successfully", time.Minute).SetVal("")
	mockR.ExpectGet("Validation").SetVal("Redis connected successfully")

	err := checkKeyDBHealthy(clientRW, clientR)
	require.NoError(t, err)
}

func TestCheckRedisHealthy_Failed(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	clientR, mockR := redismock.NewClientMock()

	mockRW.ExpectSet("Validation", "Redis connected successfully", time.Minute).SetVal("")
	mockR.ExpectGet("Validation").SetErr(errors.New("error"))

	err := checkKeyDBHealthy(clientRW, clientR)
	require.Equal(t, "error", err.Error())
}

func TestGetKeyTTL(t *testing.T) {
	storeConfig := options.Store{
		Options: &options.StoreOptions{
			Ttl:             time.Minute,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)

	keydbClient := &Client{
		clientRW: clientRW,
	}

	err := keydbClient.Store(storeConfig)
	require.NoError(t, err)

	ttl := keydbClient.getKeyTTL(storeConfig.Name)
	require.Equal(t, time.Minute, ttl)
}

func TestGetKeyTTL_NotFoundInSyncMap(t *testing.T) {
	keydbClient := NewClient()

	ttl := keydbClient.getKeyTTL("storeName")
	require.Equal(t, time.Duration(-1), ttl)
}

func TestGetKeyTTL_InvalidSettingsValue(t *testing.T) {
	keydbClient := NewClient()

	keydbClient.storeTTLs.Store("storeName", "invalidValue")

	ttl := keydbClient.getKeyTTL("storeName")
	require.Equal(t, time.Duration(-1), ttl)
}

func TestGetRedisUrl(t *testing.T) {
	require.Equal(t, "redis://localhost:6379", getKeyDBUrl("localhost:6379"))
}

func getMockStoreConfig() options.Store {
	return options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl: 10 * time.Second,
		},
	}
}
