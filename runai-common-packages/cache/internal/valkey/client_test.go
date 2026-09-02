/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package valkey

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/internal/validator"

	"github.com/go-redis/redismock/v9"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	valkeyClient := NewClient()
	require.NotNil(t, valkeyClient)
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}

	err := valkeyClient.Store(storeConfig)
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}

	err := valkeyClient.Store(storeConfig)
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}

	err := valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	v1, ok := valkeyClient.storeValidators.Load(storeConfig.Name)
	require.True(t, ok)

	err = valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	v2, ok := valkeyClient.storeValidators.Load(storeConfig.Name)
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
	valkeyClient := &Client{
		clientRW: clientRW,
	}
	val, err := valkeyClient.getValToStore(storeConfig.Name, []byte("value"))
	require.NoError(t, err)
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)
	mockRW.ExpectHSet(storeConfig.Name, "key", val).SetVal(1)

	err = valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	err = valkeyClient.Set(context.Background(), storeConfig.Name, "key", []byte("value"))
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
	valkeyClient := &Client{
		clientRW: clientRW,
	}

	val, err := valkeyClient.getValToStore(storeConfig.Name, []byte("value"))
	require.NoError(t, err)
	mockRW.ExpectHLen(storeConfig.Name).SetVal(0)
	mockRW.ExpectHSet(storeConfig.Name, "key", val).SetErr(errors.New("error"))

	err = valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	err = valkeyClient.Set(context.Background(), storeConfig.Name, "key", []byte("value"))
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}

	err := valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	err = valkeyClient.Set(context.Background(), storeConfig.Name, "key", []byte("value"))
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}
	err := valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	acquired, err := valkeyClient.SetIfAbsent(context.Background(), storeConfig.Name, "key", []byte("value"), nil)
	require.True(t, acquired)
	require.NoError(t, err)
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}
	err := valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	acquired, err := valkeyClient.SetIfAbsent(context.Background(), storeConfig.Name, "key", []byte("value"), nil)
	require.False(t, acquired)
	require.Error(t, err)
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}
	err := valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	acquired, err := valkeyClient.SetIfAbsent(context.Background(), storeConfig.Name, "key", []byte("value"), nil)
	require.False(t, acquired)
	require.Error(t, err)
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}
	err := valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	acquired, err := valkeyClient.SetIfAbsent(context.Background(), storeConfig.Name, "key", []byte("value"), nil)
	require.False(t, acquired)
	require.NoError(t, err)
}

func TestGet(t *testing.T) {
	storeConfig := getMockStoreConfig()

	clientR, mockR := redismock.NewClientMock()
	valkeyClient := &Client{
		clientR: clientR,
	}

	val, err := valkeyClient.getValToStore(storeConfig.Name, []byte("value"))
	require.NoError(t, err)
	mockR.ExpectHGet(storeConfig.Name, "key").SetVal(string(val))

	val, err = valkeyClient.Get(context.Background(), storeConfig.Name, "key")
	require.NoError(t, err)
	require.Equal(t, []byte("value"), val)
}

func TestGet_Expired(t *testing.T) {
	storeConfig := options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl:             time.Microsecond,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
		},
	}

	clientR, mockR := redismock.NewClientMock()
	valkeyClient := &Client{
		clientR:  clientR,
		clientRW: clientR,
	}
	mockR.ExpectHLen(storeConfig.Name).SetVal(calculateStoreMaxKeys(storeConfig.Options) + 1)
	err := valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	val, err := valkeyClient.getValToStore(storeConfig.Name, []byte("value"))
	require.NoError(t, err)
	mockR.ExpectHGet(storeConfig.Name, "key").SetVal(string(val))
	mockR.ExpectHDel(storeConfig.Name, "key")

	val, err = valkeyClient.Get(context.Background(), storeConfig.Name, "key")
	require.NoError(t, err)
	require.Nil(t, val)
}

func TestGet_KeyNotFound(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"

	clientR, mockR := redismock.NewClientMock()
	mockR.ExpectHGet(storeConfig.Name, key).RedisNil()

	valkeyClient := &Client{
		clientR: clientR,
	}

	val, err := valkeyClient.Get(ctx, storeConfig.Name, key)
	require.NoError(t, err)
	require.Nil(t, val)
}

func TestGet_Failed(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"

	clientR, mockR := redismock.NewClientMock()
	mockR.ExpectHGet(storeConfig.Name, key).SetErr(errors.New("error"))

	valkeyClient := &Client{
		clientR: clientR,
	}

	val, err := valkeyClient.Get(ctx, storeConfig.Name, key)
	require.Equal(t, "failed to get key <testStore,key1> from cache: error", err.Error())
	require.Nil(t, val)
}

func TestDelete(t *testing.T) {
	storeConfig := getMockStoreConfig()

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHDel(storeConfig.Name, "key").SetVal(1)

	valkeyClient := &Client{
		clientRW: clientRW,
	}

	err := valkeyClient.Delete(context.Background(), storeConfig.Name, "key")
	require.NoError(t, err)
}

func TestDelete_Failed(t *testing.T) {
	storeConfig := getMockStoreConfig()

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHDel(storeConfig.Name, "key").SetErr(errors.New("error"))

	valkeyClient := &Client{
		clientRW: clientRW,
	}

	err := valkeyClient.Delete(context.Background(), storeConfig.Name, "key")
	require.Equal(t, "failed to delete key <testStore,key> from cache: error", err.Error())
}

func TestClearStore(t *testing.T) {
	storeConfig := getMockStoreConfig()

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectDel(storeConfig.Name).SetVal(1)

	valkeyClient := &Client{
		clientRW:        clientRW,
		storeValidators: sync.Map{},
	}
	v, err := validator.NewStoreLenValidator("storeName", 100, func(ctx context.Context) (int64, error) {
		return 0, nil
	})
	require.NoError(t, err)
	valkeyClient.storeValidators.Store(storeConfig.Name, v)

	err = valkeyClient.ClearStore(context.Background(), storeConfig.Name)
	require.NoError(t, err)
}

func TestClearStore_Failed(t *testing.T) {
	storeConfig := getMockStoreConfig()

	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectDel(storeConfig.Name).SetErr(errors.New("error"))

	valkeyClient := &Client{
		clientRW: clientRW,
	}

	err := valkeyClient.ClearStore(context.Background(), storeConfig.Name)
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}

	err := valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	err = valkeyClient.validateStoreNotFull(context.Background(), storeConfig.Name)
	require.NoError(t, err)
}

func TestValidateStoreNotFull_NotFoundInSyncMap(t *testing.T) {
	valkeyClient := NewClient()

	err := valkeyClient.validateStoreNotFull(context.Background(), "storeName")
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}

	err := valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	err = valkeyClient.validateStoreNotFull(context.Background(), storeConfig.Name)
	require.Equal(t, "store <testStore> exceeded max size (20 keys)", err.Error())
}

func TestCheckValkeyHealthy(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	clientR, mockR := redismock.NewClientMock()

	mockRW.ExpectSet("Validation", "Redis connected successfully", time.Minute).SetVal("")
	mockR.ExpectGet("Validation").SetVal("Redis connected successfully")

	err := checkValkeyHealthy(clientRW, clientR)
	require.NoError(t, err)
}

func TestCheckValkeyHealthy_Failed(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	clientR, mockR := redismock.NewClientMock()

	mockRW.ExpectSet("Validation", "Redis connected successfully", time.Minute).SetVal("")
	mockR.ExpectGet("Validation").SetErr(errors.New("error"))

	err := checkValkeyHealthy(clientRW, clientR)
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

	valkeyClient := &Client{
		clientRW: clientRW,
	}

	err := valkeyClient.Store(storeConfig)
	require.NoError(t, err)

	ttl := valkeyClient.getKeyTTL(storeConfig.Name)
	require.Equal(t, time.Minute, ttl)
}

func TestGetKeyTTL_NotFoundInSyncMap(t *testing.T) {
	valkeyClient := NewClient()

	ttl := valkeyClient.getKeyTTL("storeName")
	require.Equal(t, time.Duration(-1), ttl)
}

func TestGetKeyTTL_InvalidSettingsValue(t *testing.T) {
	valkeyClient := NewClient()

	valkeyClient.storeTTLs.Store("storeName", "invalidValue")

	ttl := valkeyClient.getKeyTTL("storeName")
	require.Equal(t, time.Duration(-1), ttl)
}

func TestGetUrl(t *testing.T) {
	require.Equal(t, "redis://localhost:6379", getUrl("localhost:6379"))
}

func getMockStoreConfig() options.Store {
	return options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl: 10 * time.Second,
		},
	}
}
