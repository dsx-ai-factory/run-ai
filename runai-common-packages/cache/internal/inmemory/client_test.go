/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inmemory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/maypok86/otter/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	err := (&Client{}).Init()
	require.NoError(t, err)
}

func TestStore(t *testing.T) {
	client := &Client{}
	cfg := getStoreConfig("test-store")

	err := client.Store(cfg)
	require.NoError(t, err)

	_, exists := client.stores.Load(cfg.Name)
	require.True(t, exists)
}

func TestSetAndGet(t *testing.T) {
	client := &Client{}
	cfg := getStoreConfig("test-store")

	err := client.Store(cfg)
	require.NoError(t, err)

	err = client.Set(context.Background(), cfg.Name, "some-key", []byte("some-value"))
	require.NoError(t, err)

	val, err := client.Get(context.Background(), cfg.Name, "some-key")
	require.Equal(t, "some-value", string(val))
	require.NoError(t, err)
}

func TestSet_ResetTheSameKey(t *testing.T) {
	client := &Client{}
	cfg := getStoreConfig("test-store")

	err := client.Store(cfg)
	require.NoError(t, err)

	err = client.Set(context.Background(), cfg.Name, "some-key", []byte("some-value"))
	require.NoError(t, err)

	val, err := client.Get(context.Background(), cfg.Name, "some-key")
	require.Equal(t, "some-value", string(val))
	require.NoError(t, err)

	err = client.Set(context.Background(), cfg.Name, "some-key", []byte("new-value"))
	require.NoError(t, err)

	val, err = client.Get(context.Background(), cfg.Name, "some-key")
	require.Equal(t, "new-value", string(val))
	require.NoError(t, err)
}

func TestSet_ExceedsMaxRecordSize(t *testing.T) {
	client := &Client{}
	cfg := options.Store{
		Name: "test-store",
		Options: &options.StoreOptions{
			MaxStoreSizeMB: 1,
			Ttl:            5 * time.Second,
		},
	}

	err := client.Store(cfg)
	require.NoError(t, err)

	value := make([]byte, 2_000_000)
	client.Set(context.Background(), cfg.Name, "large-key", value)

	// time to async eviction that triggered by Set operation to complete (W-TinyLFU - not stable yet)
	<-time.After(time.Millisecond)

	value, err = client.Get(context.Background(), cfg.Name, "large-key")
	require.NoError(t, err)
	require.Equal(t, value, []byte(nil))
}

func TestSet_SomeEntryEvicted(t *testing.T) {
	client := &Client{}
	cfg := options.Store{
		Name: "test-store",
		Options: &options.StoreOptions{
			MaxStoreSizeMB: 6,
			Ttl:            5 * time.Minute,
		},
	}

	err := client.Store(cfg)
	require.NoError(t, err)

	value := make([]byte, 500*1000)

	// insert 12 records to fill the cache
	for i := 0; i < 13; i++ {
		key := fmt.Sprintf("key_%d", i)
		err = client.Set(context.Background(), cfg.Name, key, value)
		require.NoError(t, err)
	}

	// time for W-TinyLFU to stabilize
	<-time.After(5 * time.Millisecond)

	// Verify that eviction occurred - one key should be missing
	require.Eventually(t, func() (existRemovedKey bool) {
		for i := 0; i < 13; i++ {
			v, err := client.Get(context.Background(), cfg.Name, fmt.Sprintf("key_%d", i))
			if v == nil && err == nil {
				existRemovedKey = true
				break
			}
		}
		return
	}, 2*time.Second, 100*time.Millisecond, "some key should be evicted")
}

func TestSetIfAbsent(t *testing.T) {
	client := &Client{}
	cfg := getStoreConfig("test-store")

	err := client.Store(cfg)
	require.NoError(t, err)

	acquired, err := client.SetIfAbsent(context.Background(), cfg.Name, "key", []byte("value"), nil)
	require.True(t, acquired)
	require.NoError(t, err)
}

func TestSetIfAbsent_Exists(t *testing.T) {
	client := &Client{}
	cfg := getStoreConfig("test-store")

	err := client.Store(cfg)
	require.NoError(t, err)

	acquired, err := client.SetIfAbsent(context.Background(), cfg.Name, "key", []byte("value"), nil)
	require.True(t, acquired)
	require.NoError(t, err)
	acquired, err = client.SetIfAbsent(context.Background(), cfg.Name, "key", []byte("value"), nil)
	require.False(t, acquired)
	require.NoError(t, err)
}

func TestSetIfAbsent_TTL(t *testing.T) {
	client := &Client{}
	cfg := getStoreConfig("test-store")

	err := client.Store(cfg)
	require.NoError(t, err)
	ttl := time.Duration(1 * time.Second)
	acquired, err := client.SetIfAbsent(context.Background(), cfg.Name, "key", []byte("value"), &ttl)
	require.True(t, acquired)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)
	acquired, err = client.SetIfAbsent(context.Background(), cfg.Name, "key", []byte("value"), nil)
	require.True(t, acquired)
	require.NoError(t, err)
}

func TestDelete(t *testing.T) {
	client := &Client{}
	cfg := getStoreConfig("test-store")

	err := client.Store(cfg)
	require.NoError(t, err)

	err = client.Set(context.Background(), cfg.Name, "some-key", []byte("some-value"))
	require.NoError(t, err)

	val, err := client.Get(context.Background(), cfg.Name, "some-key")
	require.Equal(t, "some-value", string(val))
	require.NoError(t, err)

	err = client.Delete(context.Background(), cfg.Name, "some-key")
	require.NoError(t, err)

	val, err = client.Get(context.Background(), cfg.Name, "some-key")
	require.Nil(t, val)
	require.NoError(t, err)
}

func TestTtlDeletion(t *testing.T) {
	client := &Client{}
	cfg := options.Store{
		Name: "test-store",
		Options: &options.StoreOptions{
			MaxStoreSizeMB: 10,
			Ttl:            1 * time.Second,
		},
	}

	err := client.Store(cfg)
	require.NoError(t, err)

	err = client.Set(context.Background(), cfg.Name, "some-key", []byte("some-value"))
	require.NoError(t, err)

	val, err := client.Get(context.Background(), cfg.Name, "some-key")
	require.Equal(t, "some-value", string(val))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		val, err = client.Get(context.Background(), cfg.Name, "some-key")
		return val == nil && err == nil
	}, 2*time.Second, 100*time.Millisecond)
}

func TestClearStore(t *testing.T) {
	client := &Client{}
	cfg := getStoreConfig("store-to-clear")

	err := client.Store(cfg)
	require.NoError(t, err)

	err = client.Set(context.Background(), cfg.Name, "some-key", []byte("some-value"))
	require.NoError(t, err)

	store, ok := client.stores.Load(cfg.Name)
	require.True(t, ok)
	s := store.(*otter.Cache[string, []byte])
	require.Equal(t, 1, s.EstimatedSize())

	err = client.ClearStore(context.Background(), cfg.Name)
	require.NoError(t, err)

	store, ok = client.stores.Load(cfg.Name)
	require.True(t, ok)
	s = store.(*otter.Cache[string, []byte])
	require.Equal(t, 0, s.EstimatedSize())
}

func TestClose(t *testing.T) {
	client := &Client{}

	err := client.Close()
	require.NoError(t, err)
}

func getStoreConfig(name string) options.Store {
	return options.Store{
		Name: name,
		Options: &options.StoreOptions{
			MaxStoreSizeMB: 10,
			Ttl:            5 * time.Minute,
		},
	}
}
